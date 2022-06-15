package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/saltosystems/winrt-go/internal/metadata"
	"github.com/tdakkota/win32metadata/types"
	"golang.org/x/tools/imports"
)

const (
	attributeTypeGUID                 = "Windows.Foundation.Metadata.GuidAttribute"
	attributeTypeExclusiveTo          = "Windows.Foundation.Metadata.ExclusiveToAttribute"
	attributeTypeStaticAttribute      = "Windows.Foundation.Metadata.StaticAttribute"
	attributeTypeActivatableAttribute = "Windows.Foundation.Metadata.ActivatableAttribute"
)

type generator struct {
	class        string
	methodFilter *MethodFilter

	logger log.Logger

	genData *genData

	mdStore *metadata.Store
}

// Generate generates the code for the given config.
func Generate(cfg *Config, logger log.Logger) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	mdStore, err := metadata.NewStore(logger)
	if err != nil {
		return err
	}

	g := &generator{
		class:        cfg.Class,
		methodFilter: cfg.MethodFilter(),
		logger:       logger,
		mdStore:      mdStore,
	}
	return g.run()
}

func (g *generator) run() error {
	_ = level.Debug(g.logger).Log("msg", "starting code generation", "class", g.class)

	typeDef, err := g.mdStore.TypeDefByName(g.class)
	if err != nil {
		return err
	}

	return g.generate(typeDef)
}

func (g *generator) generate(typeDef *metadata.TypeDef) error {

	// we only support WinRT types: check the tdWindowsRuntime flag (0x4000)
	// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#runtime-classes
	if typeDef.Flags&0x4000 == 0 {
		return fmt.Errorf("%s.%s is not a WinRT class", typeDef.TypeNamespace, typeDef.TypeName)
	}

	_ = level.Info(g.logger).Log("msg", "generating class", "class", typeDef.TypeNamespace+"."+typeDef.TypeName)

	// get templates
	tmpl, err := getTemplates()
	if err != nil {
		return err
	}

	// get data & execute templates

	if err := g.loadCodeGenData(typeDef); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "file.tmpl", g.genData); err != nil {
		return err
	}

	// create file & write contents
	folder := typeToFolder(typeDef.TypeNamespace, typeDef.TypeName)
	filename := folder + "/" + strings.ToLower(typeDef.TypeName) + ".go"
	err = os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(filepath.Clean(filename))
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	// use go imports to cleanup imports
	goimported, err := imports.Process(filename, buf.Bytes(), nil)
	if err != nil {
		// write unimported  source code to file as a debugging mechanism
		_, _ = file.Write(buf.Bytes())
		return err
	}

	// format the output source code
	formatted, err := format.Source(goimported)
	if err != nil {
		// write unformatted source code to file as a debugging mechanism
		_, _ = file.Write(goimported)
		return err
	}

	// and write it to file
	_, err = file.Write(formatted)

	return err
}

func (g *generator) loadCodeGenData(typeDef *metadata.TypeDef) error {
	g.genData = &genData{
		Package: typePackage(typeDef.TypeNamespace, typeDef.TypeName),
	}

	switch {
	case g.isInterface(typeDef):
		if err := g.validateInterface(typeDef); err != nil {
			return err
		}

		iface, err := g.createGenInterface(typeDef, false)
		if err != nil {
			return err
		}
		g.genData.Interfaces = append(g.genData.Interfaces, *iface)
	case g.isEnum(typeDef):
		enum, err := g.createGenEnum(typeDef)
		if err != nil {
			return err
		}
		g.genData.Enums = append(g.genData.Enums, *enum)
	default:
		class, err := g.createGenClass(typeDef)
		if err != nil {
			return err
		}
		g.genData.Classes = append(g.genData.Classes, *class)
	}

	return nil
}

func (g *generator) isInterface(typeDef *metadata.TypeDef) bool {
	return typeDef.Flags.Interface()
}

func (g *generator) isEnum(typeDef *metadata.TypeDef) bool {
	ns, name, err := typeDef.Ctx().ResolveTypeDefOrRefName(typeDef.Extends)
	if err != nil {
		_ = level.Error(g.logger).Log("msg", "error resolving type extends, all classes should extend at least System.Object", "err", err)
		return false
	}
	return ns == "System" && name == "Enum"
}

func (g *generator) validateInterface(typeDef *metadata.TypeDef) error {
	// Any WinRT interface with private visibility must have a single ExclusiveToAttribute.
	// the ExclusiveToAttribute must reference a runtime class.

	// we do not support generating these types of classes, they are exclusive to a runtime class,
	// and thus will be generated when the runtime class is generated.

	if typeDef.Flags.NotPublic() {
		return fmt.Errorf("interface %s is not public", typeDef.TypeNamespace+"."+typeDef.TypeName)
	}
	return nil
}

// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#interfaces
func (g *generator) createGenInterface(typeDef *metadata.TypeDef, requiresActivation bool) (*genInterface, error) {
	// Any WinRT interface with public visibility must not have an ExclusiveToAttribute.

	funcs, err := g.getGenFuncs(typeDef, requiresActivation)
	if err != nil {
		return nil, err
	}

	// Interfaces' TypeDef rows must have a GuidAttribute as well as a VersionAttribute.
	guid, err := g.typeGUID(typeDef)
	if err != nil {
		return nil, err
	}

	return &genInterface{
		Name:  typeDefGoName(typeDef),
		GUID:  guid,
		Funcs: funcs,
	}, nil
}

// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#runtime-classes
func (g *generator) createGenClass(typeDef *metadata.TypeDef) (*genClass, error) {
	exclusiveInterfaceTypes := make([]*metadata.TypeDef, 0)
	// true => interface requires activation, false => interface is implemented by this class
	activatedInterfaces := make(map[string]bool)

	// get all the interfaces this class implements
	interfaces, err := typeDef.GetImplementedInterfaces()
	if err != nil {
		return nil, err
	}
	implInterfaces := make([]string, 0, len(interfaces))
	for _, iface := range interfaces {
		// the interface needs to be implemented by this class
		pkg := ""
		if typeDef.TypeNamespace != iface.Namespace {
			pkg = typePackage(iface.Namespace, iface.Name) + "."
			g.addImportFor(iface.Namespace, iface.Name)
		}

		ifaceTypeDef, err := g.mdStore.TypeDefByName(iface.Namespace + "." + iface.Name)
		if err != nil {
			return nil, err
		}
		implInterfaces = append(implInterfaces, pkg+typeDefGoName(ifaceTypeDef))

		// The interface we implement may be exclusive to this class, in which case we need to generate it.
		// An exclusive (private) interface should always belong to the same winmd file. So even if this is
		// a TypeRef, the class should be found using its name.
		if td, err := g.mdStore.TypeDefByName(iface.Namespace + "." + iface.Name); err == nil {
			if _, ok := g.interfaceIsExclusiveTo(td); ok {
				exclusiveInterfaceTypes = append(exclusiveInterfaceTypes, td)
				activatedInterfaces[td.TypeNamespace+"."+td.TypeName] = false // implemented interfaces do not require activation
			}
		}
	}

	// Runtime classes have zero or more StaticAttribute custom attributes
	// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#static-interfaces
	staticAttributeBlobs := typeDef.GetTypeDefAttributesWithType(attributeTypeStaticAttribute)
	for _, blob := range staticAttributeBlobs {
		class := extractClassFromBlob(blob)
		_ = level.Debug(g.logger).Log("msg", "found static interface", "class", class)
		staticClass, err := g.mdStore.TypeDefByName(class)
		if err != nil {
			_ = level.Error(g.logger).Log("msg", "static class defined in StaticAttribute not found", "class", class, "err", err)
			return nil, err
		}

		exclusiveInterfaceTypes = append(exclusiveInterfaceTypes, staticClass)
		activatedInterfaces[staticClass.TypeNamespace+"."+staticClass.TypeName] = true // static interfaces require activation
	}

	// Runtime classes have zero or more ActivatableAttribute custom attributes
	// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#activation
	activatableAttributeBlobs := typeDef.GetTypeDefAttributesWithType(attributeTypeActivatableAttribute)
	hasEmptyConstructor := false
	for _, blob := range activatableAttributeBlobs {
		// check for empty constructor
		if activatableAttrIsEmpty(blob) {
			// this activatable attribute is empty, so the class has an empty constructor
			hasEmptyConstructor = true
			continue
		}

		// check for an activation interface
		class := extractClassFromBlob(blob)
		_ = level.Debug(g.logger).Log("msg", "found activatable interface", "class", class)
		activatableClass, err := g.mdStore.TypeDefByName(class)
		if err != nil {
			// the activatable class may be empty in some cases, example:
			// https://github.com/tpn/winsdk-10/blob/9b69fd26ac0c7d0b83d378dba01080e93349c2ed/Include/10.0.14393.0/winrt/windows.devices.bluetooth.advertisement.idl#L518
			_ = level.Error(g.logger).Log("msg", "activatable class defined in ActivatableAttribute not found", "class", class, "err", err)

			// so do not fail
			continue
		}
		exclusiveInterfaceTypes = append(exclusiveInterfaceTypes, activatableClass)
		activatedInterfaces[activatableClass.TypeNamespace+"."+activatableClass.TypeName] = true // activatable interfaces require activation
	}

	// generate exclusive interfaces
	exclusiveGenInterfaces := make([]genInterface, 0)
	for _, iface := range exclusiveInterfaceTypes {
		requiresActivation := activatedInterfaces[iface.TypeNamespace+"."+iface.TypeName]
		isExtendedInterface := !requiresActivation

		ifaceGen, err := g.createGenInterface(iface, requiresActivation)
		if err != nil {
			return nil, err
		}

		// if all methods from the exclusive interface have been filtered, and the interface
		// is an activated interface, then we can skip it.
		impl := false
		for _, m := range ifaceGen.Funcs {
			if m.Implement {
				impl = true
			}
		}

		// Extended interfaces always need to be generated (even if they have no method).
		if isExtendedInterface || impl {
			exclusiveGenInterfaces = append(exclusiveGenInterfaces, *ifaceGen)
		}
	}

	return &genClass{
		Name:                typeDefGoName(typeDef),
		FullyQualifiedName:  typeDef.TypeNamespace + "." + typeDef.TypeName,
		ImplInterfaces:      implInterfaces,
		ExclusiveInterfaces: exclusiveGenInterfaces,
		HasEmptyConstructor: hasEmptyConstructor,
	}, nil
}

// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#enums
func (g *generator) createGenEnum(typeDef *metadata.TypeDef) (*genEnum, error) {
	fields, err := typeDef.ResolveFieldList(typeDef.Ctx())
	if err != nil {
		return nil, err
	}

	// An enum has a single instance field that specifies the underlying integer type for the enum,
	// as well as zero or more static fields; one for each enum value defined by the enum type.
	if len(fields) == 0 {
		return nil, fmt.Errorf("enum %s has no fields", typeDef.TypeName)
	}

	// the first row should be the underlying integer type of the enum. It must have the following flags:
	if !(fields[0].Flags.Private() && fields[0].Flags.SpecialName() && fields[0].Flags.RTSpecialName()) {
		return nil, fmt.Errorf("enum %s has more than one instance field, expected 1", typeDef.TypeNamespace+"."+typeDef.TypeName)
	}

	fieldSig, err := fields[0].Signature.Reader().Field(typeDef.Ctx())
	if err != nil {
		return nil, err
	}
	elType, err := g.elementType(typeDef.Ctx(), fieldSig.Field)
	if err != nil {
		return nil, err
	}
	// this will always be a primitive type, so we can just use the name
	enumType := elType.Name

	// After the enum value definition comes a field definition for each of the values in the enumeration.
	enumValues := make([]genEnumValue, 0, len(fields[1:]))
	for i, field := range fields[1:] {
		if !(field.Flags.Public() && field.Flags.Static() && field.Flags.Literal() && field.Flags.HasDefault()) {
			return nil,
				fmt.Errorf(
					"enum %s field value does not comply with the spec. Checkout https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#enums",
					typeDef.TypeNamespace+"."+typeDef.TypeName,
				)
		}

		var fieldIndex uint32 = typeDef.FieldList.Start() + 1 + uint32(i)
		enumRawValue, err := typeDef.GetValueForEnumField(fieldIndex)
		if err != nil {
			return nil, err
		}

		enumValues = append(enumValues, genEnumValue{
			Name:  enumName(typeDef.TypeName, field.Name),
			Value: enumRawValue,
		})
	}

	return &genEnum{
		Name:   typeDefGoName(typeDef),
		Type:   enumType,
		Values: enumValues,
	}, nil
}

func (g *generator) interfaceIsExclusiveTo(typeDef *metadata.TypeDef) (string, bool) {
	exclusiveToBlob, err := typeDef.GetTypeDefAttributeWithType(attributeTypeExclusiveTo)
	// an error here is fine, we just won't have the ExclusiveTo attribute
	if err != nil {
		return "", false
	}
	exclusiveToClass := extractClassFromBlob(exclusiveToBlob)
	return exclusiveToClass, true
}

func typeDefGoName(typeDef *metadata.TypeDef) string {
	name := typeDef.TypeName
	if !typeDef.Flags.Public() {
		name = strings.ToLower(name[0:1]) + name[1:]
	}
	return name
}

func activatableAttrIsEmpty(blob []byte) bool {
	// the activatable attribute is empty if the size is 0
	// 01 00 - header
	// 00 - size
	return len(blob) >= 3 && blob[0] == 0x01 && blob[1] == 0x00 && blob[2] == 0x00
}

func extractClassFromBlob(blob []byte) string {
	// the blob contains a two byte header
	// 01 00
	// followed by a byte with the size
	// XX
	// followed by the type name

	// so we need at least 4 bytes
	if len(blob) < 4 {
		return ""
	}
	size := blob[2]
	class := blob[3 : 3+size]
	return string(class)
}

func (g *generator) addImportFor(ns, name string) {
	folder := typeToFolder(ns, name)
	i := "github.com/saltosystems/winrt-go/" + folder
	g.genData.Imports = append(g.genData.Imports, i)
}

func (g *generator) getGenFuncs(typeDef *metadata.TypeDef, requiresActivation bool) ([]genFunc, error) {
	var genFuncs []genFunc

	methods, err := typeDef.ResolveMethodList(typeDef.Ctx())
	if err != nil {
		return nil, err
	}

	var exclusiveToType string
	if ex, ok := g.interfaceIsExclusiveTo(typeDef); ok {
		exclusiveToType = ex
	}

	for _, m := range methods {
		methodDef := m
		generatedFunc, err := g.genFuncFromMethod(typeDef, &methodDef, exclusiveToType, requiresActivation)
		if err != nil {
			return nil, err
		}
		genFuncs = append(genFuncs, *generatedFunc)
	}

	return genFuncs, nil
}

func (g *generator) genFuncFromMethod(typeDef *metadata.TypeDef, methodDef *types.MethodDef, exclusiveTo string, requiresActivation bool) (*genFunc, error) {
	// add the type imports to the top of the file
	// only if the method is going to be implemented
	implement := g.shouldImplementMethod(methodDef)
	var params []*genParam
	var retParam *genParam
	if implement {
		paramsCandidate, err := g.getInParameters(typeDef, methodDef)
		if err != nil {
			return nil, err
		}
		params = paramsCandidate

		retParamCandidate, err := g.getReturnParameters(typeDef, methodDef)
		if err != nil {
			return nil, err
		}
		retParam = retParamCandidate

		curPackage := typePackage(typeDef.TypeNamespace, typeDef.TypeName)

		implementedParams := make([]*genParam, 0)
		implementedParams = append(implementedParams, params...)
		if retParam != nil {
			implementedParams = append(implementedParams, retParam)
		}

		// for each param
		for _, p := range implementedParams {
			// compute the reference in the code (with vs without package, pointer, ...)
			if p.genType != nil {
				p.Type = p.genType.GoParamString(curPackage)
			}
			if p.genDefaultValue != nil {
				p.DefaultValue = p.genDefaultValue.GoParamString(curPackage)
			}

			if p.genType.Namespace == "" {
				// no import required
				continue
			}

			typePkg := typePackage(p.genType.Namespace, p.genType.Name)
			if curPackage != typePkg {
				// imports are addded globally
				g.addImportFor(p.genType.Namespace, p.genType.Name)
			}
		}
	}

	return &genFunc{
		Name:               methodDef.Name,
		Implement:          implement,
		InParams:           params,
		ReturnParam:        retParam,
		FuncOwner:          typeDefGoName(typeDef),
		ExclusiveTo:        exclusiveTo,
		RequiresActivation: requiresActivation,
	}, nil
}

func (g *generator) shouldImplementMethod(methodDef *types.MethodDef) bool {
	return g.methodFilter.Filter(methodDef.Name)
}

func (g *generator) typeGUID(typeDef *metadata.TypeDef) (string, error) {
	blob, err := typeDef.GetTypeDefAttributeWithType(attributeTypeGUID)
	if err != nil {
		return "", err
	}
	return guidBlobToString(blob)
}

// guidBlobToString converts an array into the textual representation of a GUID
func guidBlobToString(b types.Blob) (string, error) {
	// the guid is a blob of 20 bytes
	if len(b) != 20 {
		return "", fmt.Errorf("invalid GUID blob length: %d", len(b))
	}

	// that starts with 0100
	if b[0] != 0x01 || b[1] != 0x00 {
		return "", fmt.Errorf("invalid GUID blob header, expected '0x01 0x00' but found '0x%02x 0x%02x'", b[0], b[1])
	}

	// and ends with 0000
	if b[18] != 0x00 || b[19] != 0x00 {
		return "", fmt.Errorf("invalid GUID blob footer, expected '0x00 0x00' but found '0x%02x 0x%02x'", b[18], b[19])
	}

	guid := b[2 : len(b)-2]
	// the string version has 5 parts separated by '-'
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%04x%08x",
		// The first 3 are encoded as little endian
		uint32(guid[0])|uint32(guid[1])<<8|uint32(guid[2])<<16|uint32(guid[3])<<24,
		uint16(guid[4])|uint16(guid[5])<<8,
		uint16(guid[6])|uint16(guid[7])<<8,
		//the rest is not
		uint16(guid[8])<<8|uint16(guid[9]),
		uint16(guid[10])<<8|uint16(guid[11]),
		uint32(guid[12])<<24|uint32(guid[13])<<16|uint32(guid[14])<<8|uint32(guid[15])), nil
}

func (g *generator) getInParameters(typeDef *metadata.TypeDef, methodDef *types.MethodDef) ([]*genParam, error) {

	params, err := methodDef.ResolveParamList(typeDef.Ctx())
	if err != nil {
		return nil, err
	}

	// the signature contains the parameter
	// types and return type of the method
	r := methodDef.Signature.Reader()
	mr, err := r.Method(typeDef.Ctx())
	if err != nil {
		return nil, err
	}

	genParams := make([]*genParam, 0)
	for i, e := range mr.Params {
		elType, err := g.elementType(typeDef.Ctx(), e)
		if err != nil {
			return nil, err
		}
		genParams = append(genParams, &genParam{
			Name:    getParamName(params, uint16(i+1)),
			Type:    "",
			genType: elType,
		})
	}

	return genParams, nil
}

func (g *generator) getReturnParameters(typeDef *metadata.TypeDef, methodDef *types.MethodDef) (*genParam, error) {
	// the signature contains the parameter
	// types and return type of the method
	r := methodDef.Signature.Reader()
	methodSignature, err := r.Method(typeDef.Ctx())
	if err != nil {
		return nil, err
	}

	// ignore void types
	if methodSignature.Return.Type.Kind == types.ELEMENT_TYPE_VOID {
		return nil, nil
	}

	elType, err := g.elementType(typeDef.Ctx(), methodSignature.Return)
	if err != nil {
		return nil, err
	}

	defValue, err := g.elementDefaultValue(typeDef.Ctx(), methodSignature.Return)
	if err != nil {
		return nil, err
	}

	return &genParam{
		Name:            "",
		Type:            "",
		genType:         elType,
		DefaultValue:    "",
		genDefaultValue: defValue,
	}, nil
}

func getParamName(params []types.Param, i uint16) string {
	for _, p := range params {
		if p.Flags.In() && p.Sequence == i {
			return p.Name
		}
	}
	return "__ERROR__"
}

func (g *generator) elementType(ctx *types.Context, e types.Element) (*genParamReference, error) {
	switch e.Type.Kind {
	case types.ELEMENT_TYPE_BOOLEAN:
		return &genParamReference{
			Name:      "bool",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_CHAR:
		return &genParamReference{
			Name:      "byte",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_I1:
		return &genParamReference{
			Name:      "int8",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_U1:
		return &genParamReference{
			Name:      "uint8",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_I2:
		return &genParamReference{
			Name:      "int16",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_U2:
		return &genParamReference{
			Name:      "uint16",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_I4:
		return &genParamReference{
			Name:      "int32",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_U4:
		return &genParamReference{
			Name:      "uint32",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_I8:
		return &genParamReference{
			Name:      "int64",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_U8:
		return &genParamReference{
			Name:      "uint64",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_R4:
		return &genParamReference{
			Name:      "float32",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_R8:
		return &genParamReference{
			Name:      "float64",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_STRING:
		return &genParamReference{
			Name:      "string",
			Namespace: "",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_CLASS:
		// return class name
		namespace, name, err := ctx.ResolveTypeDefOrRefName(e.Type.TypeDef.Index)
		if err != nil {
			return nil, err
		}
		return &genParamReference{
			Name:      name,
			Namespace: namespace,
			IsPointer: true,
		}, nil
	case types.ELEMENT_TYPE_VALUETYPE:
		namespace, name, err := ctx.ResolveTypeDefOrRefName(e.Type.TypeDef.Index)
		if err != nil {
			return nil, err
		}
		return &genParamReference{
			Name:      name,
			Namespace: namespace,
			IsPointer: false,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported element type: %v", e.Type.Kind)
	}
}

func (g *generator) elementDefaultValue(ctx *types.Context, e types.Element) (*genParamReference, error) {
	switch e.Type.Kind {
	case types.ELEMENT_TYPE_BOOLEAN:
		return &genParamReference{
			Namespace: "",
			Name:      "false",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_CHAR:
		fallthrough
	case types.ELEMENT_TYPE_I1:
		fallthrough
	case types.ELEMENT_TYPE_U1:
		fallthrough
	case types.ELEMENT_TYPE_I2:
		fallthrough
	case types.ELEMENT_TYPE_U2:
		fallthrough
	case types.ELEMENT_TYPE_I4:
		fallthrough
	case types.ELEMENT_TYPE_U4:
		fallthrough
	case types.ELEMENT_TYPE_I8:
		fallthrough
	case types.ELEMENT_TYPE_U8:
		return &genParamReference{
			Namespace: "",
			Name:      "0",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_R4:
		fallthrough
	case types.ELEMENT_TYPE_R8:
		return &genParamReference{
			Namespace: "",
			Name:      "0.0",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_STRING:
		return &genParamReference{
			Namespace: "",
			Name:      "\"\"",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_CLASS:
		return &genParamReference{
			Namespace: "",
			Name:      "nil",
			IsPointer: false,
		}, nil
	case types.ELEMENT_TYPE_VALUETYPE:
		// we need to get the underlying type (enum, struct, etc...)
		namespace, name, err := ctx.ResolveTypeDefOrRefName(e.Type.TypeDef.Index)
		if err != nil {
			return nil, err
		}
		elementTypeDef, err := g.mdStore.TypeDefByName(namespace + "." + name)
		if err != nil {
			return nil, err
		}

		if g.isEnum(elementTypeDef) {
			// return the first enum value
			fields, err := elementTypeDef.ResolveFieldList(ctx)
			if err != nil {
				return nil, err
			}
			// the first field defines the enum type, the second is the first value
			if len(fields) < 2 {
				return nil, fmt.Errorf("enum %v has no fields", namespace+"."+name)
			}

			return &genParamReference{
				Namespace: elementTypeDef.TypeNamespace,
				Name:      enumName(elementTypeDef.TypeName, fields[1].Name),
				IsPointer: false,
			}, nil
		}

		// TODO: handle structs, etc...
		return &genParamReference{
			Namespace: "",
			Name:      "nil",
			IsPointer: false,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported element type: %v", e.Type.Kind)
	}
}
