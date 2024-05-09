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
	"github.com/saltosystems/winrt-go"
	"github.com/saltosystems/winrt-go/internal/winmd"
	"github.com/tdakkota/win32metadata/types"
	"golang.org/x/tools/imports"
)

// Delegate constants
const (
	invokeMethodName = "Invoke"
)

type generator struct {
	class        string
	validateOnly bool
	methodFilter *MethodFilter

	logger log.Logger

	genDataFiles []*genDataFile

	mdStore *winmd.Store
}

// Generate generates the code for the given config.
func Generate(cfg *Config, logger log.Logger) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	mdStore, err := winmd.NewStore(logger)
	if err != nil {
		return err
	}

	g := &generator{
		class:        cfg.Class,
		validateOnly: cfg.ValidateOnly,
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

func (g *generator) generate(typeDef *winmd.TypeDef) error {

	// we only support WinRT types: check the tdWindowsRuntime flag (0x4000)
	// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#runtime-classes
	if typeDef.Flags&0x4000 == 0 {
		return fmt.Errorf("%s.%s is not a WinRT class", typeDef.TypeNamespace, typeDef.TypeName)
	}

	// get data & execute templates
	if err := g.loadCodeGenData(typeDef); err != nil {
		return err
	}

	for _, fData := range g.genDataFiles {
		if err := g.generateDataFile(fData, typeDef); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) generateDataFile(fData *genDataFile, typeDef *winmd.TypeDef) error {
	// get templates
	tmpl, err := getTemplates()
	if err != nil {
		return err
	}

	fData.Data.ComputeImports(typeDef)

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "file.tmpl", fData.Data); err != nil {
		return err
	}

	// use go imports to cleanup imports
	goimported, err := imports.Process(fData.Filename, buf.Bytes(), nil)
	if err != nil {
		return err
	}

	// format the output source code
	formatted, err := format.Source(goimported)
	if err != nil {
		return err
	}

	if g.validateOnly {
		// validate existing file content
		return g.validateFileContent(fData, formatted)
	}

	// create file & write contents
	return g.writeFile(fData, formatted)
}

func (g *generator) validateFileContent(fData *genDataFile, genContent []byte) error {
	// validate existing content
	existingContent, err := os.ReadFile(fData.Filename)
	if err != nil {
		return err
	}

	// compare existing content to generated
	_ = level.Debug(g.logger).Log("msg", "validating generated code", "filename", fData.Filename)
	if string(existingContent) != string(genContent) {
		return fmt.Errorf("file %s does not contain expected content", fData.Filename)
	}
	return nil
}

func (g *generator) writeFile(fData *genDataFile, content []byte) error {
	parts := strings.Split(fData.Filename, "/")
	folder := strings.Join(parts[:len(parts)-1], "/")
	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(filepath.Clean(fData.Filename))
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	// and write it to file
	_, err = file.Write(content)
	if err != nil {
		return err
	}

	return nil
}

func (g *generator) loadCodeGenData(typeDef *winmd.TypeDef) error {
	f := g.addFile(typeDef, "")

	switch {
	case typeDef.IsInterface():
		_ = level.Info(g.logger).Log("msg", "generating interface", "interface", typeDef.TypeNamespace+"."+typeDef.TypeName)

		if err := g.validateInterface(typeDef); err != nil {
			return err
		}

		iface, err := g.createGenInterface(typeDef, false)
		if err != nil {
			return err
		}
		f.Data.Interfaces = append(f.Data.Interfaces, iface)
	case typeDef.IsEnum():
		_ = level.Info(g.logger).Log("msg", "generating enum", "enum", typeDef.TypeNamespace+"."+typeDef.TypeName)

		enum, err := g.createGenEnum(typeDef)
		if err != nil {
			return err
		}
		f.Data.Enums = append(f.Data.Enums, enum)
	case typeDef.IsStruct():
		_ = level.Info(g.logger).Log("msg", "generating struct", "struct", typeDef.TypeNamespace+"."+typeDef.TypeName)

		genStruct, err := g.createGenStruct(typeDef)
		if err != nil {
			return err
		}
		f.Data.Structs = append(f.Data.Structs, genStruct)
	case typeDef.IsDelegate():
		delegate, err := g.createGenDelegate(typeDef)
		if err != nil {
			return err
		}
		f.Data.Delegates = append(f.Data.Delegates, delegate)
	default:
		_ = level.Info(g.logger).Log("msg", "generating class", "class", typeDef.TypeNamespace+"."+typeDef.TypeName)

		class, err := g.createGenClass(typeDef)
		if err != nil {
			return err
		}
		f.Data.Classes = append(f.Data.Classes, class)
	}

	return nil
}

func (g *generator) addFile(typeDef *winmd.TypeDef, suffix string) *genDataFile {
	folder := typeToFolder(typeDef.TypeNamespace, typeDef.TypeName)
	filename := folder + "/" + typeFilename(typeDef.TypeName) + suffix + ".go"
	f := genDataFile{
		Filename: filename,
		Data: genData{
			Package: typePackage(typeDef.TypeNamespace, typeDef.TypeName),
		},
	}
	g.genDataFiles = append(g.genDataFiles, &f)
	return &f
}

func (g *generator) validateInterface(typeDef *winmd.TypeDef) error {
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
func (g *generator) createGenInterface(typeDef *winmd.TypeDef, requiresActivation bool) (*genInterface, error) {
	funcs, err := g.getGenFuncs(typeDef, requiresActivation)
	if err != nil {
		return nil, err
	}

	// Interfaces' TypeDef rows must have a GuidAttribute as well as a VersionAttribute.
	guid, err := typeDef.GUID()
	if err != nil {
		return nil, err
	}

	typeSig, err := g.Signature(typeDef)
	if err != nil {
		return nil, err
	}

	return &genInterface{
		Name:      typeDefGoName(typeDef.TypeName, typeDef.Flags.Public()),
		GUID:      guid,
		Signature: typeSig,
		Funcs:     funcs,
	}, nil
}

// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#runtime-classes
func (g *generator) createGenClass(typeDef *winmd.TypeDef) (*genClass, error) {
	var requiredImports []*genImport
	var exclusiveInterfaceTypes []*winmd.TypeDef

	// true => interface requires activation, false => interface is implemented by this class
	activatedInterfaces := make(map[string]bool)

	// get all the interfaces this class implements
	interfaces, err := typeDef.GetImplementedInterfaces()
	if err != nil {
		return nil, err
	}
	implInterfaces := make([]*genInterface, 0, len(interfaces))
	for _, iface := range interfaces {
		// the interface needs to be implemented by this class
		requiredImports = append(requiredImports, &genImport{iface.Namespace, iface.Name})

		ifaceTypeDef, err := g.mdStore.TypeDefByName(iface.Namespace + "." + iface.Name)
		if err != nil {
			return nil, err
		}

		itf, err := g.createGenInterface(ifaceTypeDef, false)
		if err != nil {
			return nil, err
		}

		pkg := ""
		if typeDef.TypeNamespace != ifaceTypeDef.TypeNamespace {
			pkg = typePackage(iface.Namespace, iface.Name)
		}
		for _, f := range itf.Funcs {
			f.InheritedFrom = winmd.QualifiedID{
				Namespace: pkg,
				Name:      typeDefGoName(ifaceTypeDef.TypeName, ifaceTypeDef.Flags.Public()),
			}
		}

		implInterfaces = append(implInterfaces, itf)

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
	staticAttributeBlobs := typeDef.GetTypeDefAttributesWithType(winmd.AttributeTypeStaticAttribute)
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
	activatableAttributeBlobs := typeDef.GetTypeDefAttributesWithType(winmd.AttributeTypeActivatableAttribute)
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
	var exclusiveGenInterfaces []*genInterface
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
			exclusiveGenInterfaces = append(exclusiveGenInterfaces, ifaceGen)
		}
	}

	typeSig, err := g.Signature(typeDef)
	if err != nil {
		return nil, err
	}

	return &genClass{
		Name:                typeDefGoName(typeDef.TypeName, typeDef.Flags.Public()),
		Signature:           typeSig,
		RequiresImports:     requiredImports,
		FullyQualifiedName:  typeDef.TypeNamespace + "." + typeDef.TypeName,
		ImplInterfaces:      implInterfaces,
		ExclusiveInterfaces: exclusiveGenInterfaces,
		HasEmptyConstructor: hasEmptyConstructor,
		IsAbstract:          typeDef.Flags.Abstract(),
	}, nil
}

// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#enums
func (g *generator) createGenEnum(typeDef *winmd.TypeDef) (*genEnum, error) {
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
	enumType := elType.name

	// After the enum value definition comes a field definition for each of the values in the enumeration.
	var enumValues []*genEnumValue
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

		enumValues = append(enumValues, &genEnumValue{
			Name:  enumName(typeDef.TypeName, field.Name),
			Value: enumRawValue,
		})
	}

	typeSig, err := g.Signature(typeDef)
	if err != nil {
		return nil, err
	}

	return &genEnum{
		Name:      typeDefGoName(typeDef.TypeName, typeDef.Flags.Public()),
		Type:      enumType,
		Signature: typeSig,
		Values:    enumValues,
	}, nil
}

// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#structs
func (g *generator) createGenStruct(typeDef *winmd.TypeDef) (*genStruct, error) {
	// structs do not have methods, only fields
	fields, err := typeDef.ResolveFieldList(typeDef.Ctx())
	if err != nil {
		return nil, err
	}

	curPkg := typePackage(typeDef.TypeNamespace, typeDef.TypeName)

	var genFields []*genParam
	for _, f := range fields {
		fSig, err := f.Signature.Reader().Field(typeDef.Ctx())
		if err != nil {
			return nil, err
		}

		fieldType, err := g.elementType(typeDef.Ctx(), fSig.Field)
		if err != nil {
			return nil, err
		}

		// Struct fields must be fundamental types, enums, or other structs
		genFields = append(genFields, &genParam{
			callerPackage: curPkg,
			varName:       cleanReservedWords(f.Name),
			IsOut:         false,
			Type:          fieldType,
		})
	}

	typeSig, err := g.Signature(typeDef)
	if err != nil {
		return nil, err
	}

	return &genStruct{
		Name:      typeDefGoName(typeDef.TypeName, typeDef.Flags.Public()),
		Signature: typeSig,
		Fields:    genFields,
	}, nil
}

// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#delegates
func (g *generator) createGenDelegate(typeDef *winmd.TypeDef) (*genDelegate, error) {
	// FieldList: must be empty
	// MethodList: An index into the MethodDef table (ECMA II.22.26), marking the first of a contiguous run of methods owned by this type.
	// Delegates' TypeDef rows must have a GuidAttribute
	guid, err := typeDef.GUID()
	if err != nil {
		return nil, err
	}

	// Delegates have exactly two MethodDef table entries. The first defines a constructor.
	methods, err := typeDef.ResolveMethodList(typeDef.Ctx())
	if err != nil {
		return nil, err
	}

	if len(methods) != 2 {
		return nil, fmt.Errorf("delegate %s has more than two methods", typeDef.TypeNamespace+"."+typeDef.TypeName)
	}

	// This constructor is a compatibility marker. WinRT Delegates have no such constructor method.

	// We only care about the invoke method
	invokeMethod := methods[1]
	if invokeMethod.Name != invokeMethodName {
		return nil, fmt.Errorf("found method '%s' on delegate %s but expected '%s'",
			invokeMethod.Name,
			typeDef.TypeNamespace+"."+typeDef.TypeName,
			invokeMethodName,
		)
	}

	// this is going to be used to define the callback type. We don't
	// really need the whole function, only its input parameters,
	// so we can reuse the logic used for getting them.
	f, err := g.genFuncFromMethod(typeDef, &invokeMethod, "", false)
	if err != nil {
		return nil, err
	}

	typeSig, err := g.Signature(typeDef)
	if err != nil {
		return nil, err
	}

	return &genDelegate{
		Name:      typeDefGoName(typeDef.TypeName, true),
		GUID:      guid,
		Signature: typeSig,
		InParams:  f.InParams,
	}, nil
}

func (g *generator) interfaceIsExclusiveTo(typeDef *winmd.TypeDef) (string, bool) {
	exclusiveToBlob, err := typeDef.GetAttributeWithType(winmd.AttributeTypeExclusiveTo)
	// an error here is fine, we just won't have the ExclusiveTo attribute
	if err != nil {
		return "", false
	}
	exclusiveToClass := extractClassFromBlob(exclusiveToBlob)
	return exclusiveToClass, true
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

func (g *generator) getGenFuncs(typeDef *winmd.TypeDef, requiresActivation bool) ([]*genFunc, error) {
	var genFuncs []*genFunc

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

		genFuncs = append(genFuncs, generatedFunc)
	}

	return genFuncs, nil
}

func (g *generator) genFuncFromMethod(typeDef *winmd.TypeDef, methodDef *types.MethodDef, exclusiveTo string, requiresActivation bool) (*genFunc, error) {
	// add the type imports to the top of the file
	// only if the method is going to be implemented

	overloadName := winmd.GetMethodOverloadName(typeDef.Ctx(), methodDef)
	implement := g.shouldImplementMethod(overloadName)
	if !implement {
		// if we don't implement the method, we don't need to gather
		// all the information, just the name of it is enough
		return &genFunc{
			Name:               overloadName,
			RequiresImports:    nil,
			Implement:          implement,
			InParams:           nil,
			ReturnParams:       nil,
			FuncOwner:          typeDefGoName(typeDef.TypeName, typeDef.Flags.Public()),
			ExclusiveTo:        exclusiveTo,
			RequiresActivation: requiresActivation,
		}, nil
	}

	curPackage := typePackage(typeDef.TypeNamespace, typeDef.TypeName)

	params, err := g.getInParameters(curPackage, typeDef, methodDef)
	if err != nil {
		return nil, err
	}

	retParams, err := g.getReturnParameters(curPackage, typeDef, methodDef)
	if err != nil {
		return nil, err
	}

	// iterate over all parameters (in or out) to gather the required imports
	// and calculate if we need the package name when referencing a type
	// based on the current package
	var allImplementedParams []*genParam
	allImplementedParams = append(allImplementedParams, params...)
	allImplementedParams = append(allImplementedParams, retParams...)

	var requiredImports []*genImport
	for _, p := range allImplementedParams {
		p.callerPackage = curPackage
		if !p.Type.IsPrimitive {
			requiredImports = append(requiredImports, &genImport{p.Type.namespace, p.Type.name})
		}
	}

	return &genFunc{
		Name:               overloadName,
		RequiresImports:    requiredImports,
		Implement:          implement,
		InParams:           params,
		ReturnParams:       retParams,
		FuncOwner:          typeDefGoName(typeDef.TypeName, typeDef.Flags.Public()),
		ExclusiveTo:        exclusiveTo,
		RequiresActivation: requiresActivation,
	}, nil
}

func (g *generator) shouldImplementMethod(methodName string) bool {
	return g.methodFilter.Filter(methodName)
}

func (g *generator) getInParameters(curPackage string, typeDef *winmd.TypeDef, methodDef *types.MethodDef) ([]*genParam, error) {

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

	var genParams []*genParam
	for i, e := range mr.Params {
		param := getParamByIndex(params, uint16(i+1))
		if param == nil {
			_ = level.Error(g.logger).Log("msg", "Parameter with index not found", "index", i+1)
			continue // do not fail
		}

		// When encoding an Array parameter for any interface member type, the array length
		// parameter that immediately precedes the array parameter is omitted from both the
		// MethodDefSig blob as well from as the params table. => so we need to add it manually.
		// Do not trust e.IsArray variable, it's only true for the ELEMENT_TYPE_ARRAY, it
		if e.Type.Kind == types.ELEMENT_TYPE_SZARRAY || e.Type.Kind == types.ELEMENT_TYPE_ARRAY {
			// The direction of the array parameter is directly encoded in metadata.The direction of
			// the array length parameter may be inferred as follows.
			//   - If the array parameter is an in parameter, the array length parameter must also
			//     be an IN PARAMETER.
			//   - If the array parameter is an out parameter and is not carrying the BYREF
			//     marker, the array length is an IN PARAMETER.

			//   - If the array parameter is an out parameter and carries the BYREF marker, the
			//     array length is an OUT PARAMETER.
			sizeIsOutParam := param.Flags.Out() && e.ByRef
			genParams = append(genParams, &genParam{
				callerPackage: curPackage,
				// Do not change this without also changing the code in the templates
				varName: cleanReservedWords(param.Name + "Size"),
				IsOut:   sizeIsOutParam,
				Type: &genParamType{
					namespace:    "",
					name:         "uint32",
					defaultValue: genDefaultValue{"0", true},
					IsPrimitive:  true,
					IsPointer:    false,
					IsArray:      false,
				},
			})
		}

		elType, err := g.elementType(typeDef.Ctx(), e)
		if err != nil {
			return nil, err
		}
		genParams = append(genParams, &genParam{
			callerPackage: curPackage,
			varName:       cleanReservedWords(getParamName(params, uint16(i+1))),
			IsOut:         param.Flags.Out(),
			Type:          elType,
		})
	}

	return genParams, nil
}

func (g *generator) getReturnParameters(curPackage string, typeDef *winmd.TypeDef, methodDef *types.MethodDef) ([]*genParam, error) {
	// the signature contains the parameter
	// types and return type of the method
	r := methodDef.Signature.Reader()
	methodSignature, err := r.Method(typeDef.Ctx())
	if err != nil {
		return nil, err
	}

	var genParams []*genParam

	// ignore void types
	if methodSignature.Return.Type.Kind == types.ELEMENT_TYPE_VOID {
		return genParams, nil
	}

	elType, err := g.elementType(typeDef.Ctx(), methodSignature.Return)
	if err != nil {
		return nil, err
	}

	genParams = append(genParams, &genParam{
		// return param always has an index of zero
		callerPackage: curPackage,
		varName:       "out",
		IsOut:         true,
		Type:          elType,
	})

	return genParams, nil
}

func getParamName(params []types.Param, i uint16) string {
	for _, p := range params {
		if p.Sequence == i {
			return p.Name
		}
	}
	return fmt.Sprintf("__ERROR_PARAM_%d_NOT_FOUND__", i)
}

func getParamByIndex(params []types.Param, i uint16) *types.Param {
	for _, p := range params {
		if p.Sequence == i {
			return &p
		}
	}
	return nil
}

func (g *generator) elementType(ctx *types.Context, e types.Element) (*genParamType, error) {
	switch e.Type.Kind {
	case types.ELEMENT_TYPE_BOOLEAN:
		return &genParamType{
			namespace:    "",
			name:         "bool",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_CHAR:
		return &genParamType{
			namespace:    "",
			name:         "byte",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_I1:
		return &genParamType{
			namespace:    "",
			name:         "int8",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_U1:
		return &genParamType{
			namespace:    "",
			name:         "uint8",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_I2:
		return &genParamType{
			namespace:    "",
			name:         "int16",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_U2:
		return &genParamType{
			namespace:    "",
			name:         "uint16",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_I4:
		return &genParamType{
			namespace:    "",
			name:         "int32",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_U4:
		return &genParamType{
			namespace:    "",
			name:         "uint32",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_I8:
		return &genParamType{
			namespace:    "",
			name:         "int64",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_U8:
		return &genParamType{
			namespace:    "",
			name:         "uint64",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_R4:
		return &genParamType{
			namespace:    "",
			name:         "float32",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_R8:
		return &genParamType{
			namespace:    "",
			name:         "float64",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_STRING:
		return &genParamType{
			namespace:    "",
			name:         "string",
			IsPointer:    false,
			IsPrimitive:  true,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_GENERICINST:
		fallthrough
	case types.ELEMENT_TYPE_CLASS:
		// return class name
		namespace, name, err := ctx.ResolveTypeDefOrRefName(e.Type.TypeDef.Index)
		if err != nil {
			return nil, err
		}
		return &genParamType{
			namespace:    namespace,
			name:         name,
			IsPointer:    true,
			IsPrimitive:  false,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_VALUETYPE:
		namespace, name, err := ctx.ResolveTypeDefOrRefName(e.Type.TypeDef.Index)
		if err != nil {
			return nil, err
		}

		// Check for system types
		if t, ok := isSystemType(namespace, name); ok {
			return t, nil
		}

		elementTypeDef, err := g.mdStore.TypeDefByName(namespace + "." + name)
		if err != nil {
			return nil, err
		}

		// if its an enum, we will need the underlying type
		isEnum := false
		enumType := ""
		if elementTypeDef.IsEnum() {
			enumData, err := g.createGenEnum(elementTypeDef)
			if err != nil {
				return nil, err
			}

			// Treat the enum as a primitive
			isEnum = true
			enumType = enumData.Type
		}
		return &genParamType{
			namespace:          namespace,
			name:               name,
			IsPointer:          false,
			IsPrimitive:        false,
			IsArray:            false,
			IsEnum:             isEnum,
			UnderlyingEnumType: enumType,
			defaultValue:       g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_VAR:
		// Generic types are not fully supported yet,
		// so we will just pass the raw unsafe.Pointer up to the user.
		return &genParamType{
			namespace:    "unsafe",
			name:         "Pointer",
			IsGeneric:    true,
			IsPointer:    false,
			IsPrimitive:  false,
			IsArray:      false,
			defaultValue: g.elementDefaultValue(ctx, e),
		}, nil
	case types.ELEMENT_TYPE_SZARRAY:
		//A single-dimensional, zero lower-bound array type modifier

		// e.Type.SZArray.Elem should be non-nil
		param, err := g.elementType(ctx, *e.Type.SZArray.Elem)
		if err != nil {
			return nil, err
		}

		param.IsArray = true
		// override default val
		param.defaultValue = genDefaultValue{"nil", true}

		return param, err
	case types.ELEMENT_TYPE_OBJECT:
		// This represents System.Object, so just use a pointer
		return &genParamType{
			namespace:    "unsafe",
			name:         "Pointer",
			IsPointer:    false,
			IsPrimitive:  false,
			IsArray:      false,
			defaultValue: genDefaultValue{"nil", true},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported element type: %v", e.Type.Kind)
	}
}

func isSystemType(namespace, name string) (*genParamType, bool) {
	if namespace != "System" {
		return nil, false
	}
	switch name {
	case "Guid":
		// System.Guid is a struct (value type), so we can just
		// use syscall.GUID which has the same structure
		return &genParamType{
			namespace:    "syscall",
			name:         "GUID",
			IsPointer:    false,
			IsPrimitive:  false,
			IsArray:      false,
			IsEnum:       false,
			defaultValue: genDefaultValue{value: "GUID{}", isPrimitive: false},
		}, true
	}

	return nil, false
}

func (g *generator) elementDefaultValue(ctx *types.Context, e types.Element) genDefaultValue {
	switch e.Type.Kind {
	case types.ELEMENT_TYPE_BOOLEAN:
		return genDefaultValue{"false", true}
	case types.ELEMENT_TYPE_CHAR,
		types.ELEMENT_TYPE_I1, types.ELEMENT_TYPE_U1,
		types.ELEMENT_TYPE_I2, types.ELEMENT_TYPE_U2,
		types.ELEMENT_TYPE_I4, types.ELEMENT_TYPE_U4,
		types.ELEMENT_TYPE_I8, types.ELEMENT_TYPE_U8:
		return genDefaultValue{"0", true}
	case types.ELEMENT_TYPE_R4, types.ELEMENT_TYPE_R8:
		return genDefaultValue{"0.0", true}
	case types.ELEMENT_TYPE_STRING:
		return genDefaultValue{"\"\"", true}
	case types.ELEMENT_TYPE_CLASS,
		types.ELEMENT_TYPE_GENERICINST, types.ELEMENT_TYPE_SZARRAY:
		return genDefaultValue{"nil", true}
	case types.ELEMENT_TYPE_VALUETYPE:
		// we need to get the underlying type (enum, struct, etc...)
		namespace, name, err := ctx.ResolveTypeDefOrRefName(e.Type.TypeDef.Index)
		if err != nil {
			return genDefaultValue{"__ERROR_" + err.Error(), true}
		}
		elementTypeDef, err := g.mdStore.TypeDefByName(namespace + "." + name)
		if err != nil {
			return genDefaultValue{"__ERROR_" + err.Error(), true}
		}

		if elementTypeDef.IsEnum() {
			// return the first enum value
			fields, err := elementTypeDef.ResolveFieldList(ctx)
			if err != nil {
				return genDefaultValue{"__ERROR_" + err.Error(), true}
			}
			// the first field defines the enum type, the second is the first value
			if len(fields) < 2 {
				return genDefaultValue{"__ERROR_" + fmt.Errorf("enum %v has no fields", namespace+"."+name).Error(), true}
			}

			return genDefaultValue{enumName(elementTypeDef.TypeName, fields[1].Name), false}
		} else if elementTypeDef.IsStruct() {
			return genDefaultValue{elementTypeDef.TypeName + "{}", false}
		}

		return genDefaultValue{"nil", true}
	case types.ELEMENT_TYPE_VAR:
		return genDefaultValue{"nil", true}
	case types.ELEMENT_TYPE_OBJECT:
		return genDefaultValue{"nil", true}
	default:
		return genDefaultValue{"__ERROR_" + fmt.Errorf("unsupported element type: %v", e.Type.Kind).Error(), true}
	}
}

func (g *generator) Signature(typeDef *winmd.TypeDef) (string, error) {
	// Signature generation defined in
	// https://docs.microsoft.com/en-us/uwp/winrt-cref/winrt-type-system#guid-generation-for-parameterized-types

	// type_signature => (only relevant types included)
	//   - interface_signature
	//   - delegate_signature
	//   - interface_group_signature
	//   - runtime_class_signature
	//   - struct_signature
	//   - enum_signature
	//   # these instance signatures cannot be generated until instantiated
	//   - pinterface_instance_signature
	//   - pdelegate_instance_signature
	//
	// Where:
	//   - interface_signature => guid
	//   - delegate_signature => "delegate(" guid ")"
	//   - interface_group_signature => "ig(" interface_group_name ";" default_interface ")"
	//   - runtime_class_signature => "rc(" runtime_class_name ";" default_interface ")"
	//   - struct_signature => "struct(" struct_name ";" args ")"
	//   - enum_signature => "enum(" enum_name ";" enum_underlying_type ")"

	switch {
	case typeDef.IsInterface():
		// interface_signature => guid
		guid, err := typeDef.GUID()
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("{%s}", guid), nil
	case typeDef.IsEnum():
		// enum_signature => "enum(" enum_name ";" enum_underlying_type ")"
		// the first field should be the underlying integer type of the enum. It must have the following flags:
		fields, err := typeDef.ResolveFieldList(typeDef.Ctx())
		if err != nil {
			return "", err
		}
		fieldSig, err := fields[0].Signature.Reader().Field(typeDef.Ctx())
		if err != nil {
			return "", err
		}

		enumType := primitiveTypeSignature(fieldSig.Field.Type.Kind)
		return fmt.Sprintf(`enum(%s;%s)`, typeDef.TypeNamespace+"."+typeDef.TypeName, enumType), nil
	case typeDef.IsStruct():
		// struct_signature => "struct(" struct_name ";" args ")"
		fields, err := typeDef.ResolveFieldList(typeDef.Ctx())
		if err != nil {
			return "", err
		}
		structArgs := []string{}
		for _, f := range fields {
			fSig, err := f.Signature.Reader().Field(typeDef.Ctx())
			if err != nil {
				return "", err
			}

			// Struct fields must be fundamental types, enums, or other structs
			if fSig.Field.Type.Kind == types.ELEMENT_TYPE_VALUETYPE {
				// this is an struct or an enum
				fieldType, err := g.elementType(typeDef.Ctx(), fSig.Field)
				if err != nil {
					return "", err
				}

				t, err := g.mdStore.TypeDefByName(fieldType.namespace + "." + fieldType.name)
				if err != nil {
					return "", err
				}

				sig, err := g.Signature(t)
				if err != nil {
					return "", err
				}
				structArgs = append(structArgs, sig)
			} else {
				// Assume everything else is a fundamental type
				structArgs = append(structArgs, primitiveTypeSignature(fSig.Field.Type.Kind))
			}
		}
		return fmt.Sprintf(`struct(%s;%s)`, typeDef.TypeNamespace+"."+typeDef.TypeName, strings.Join(structArgs, ";")), nil
	case typeDef.IsDelegate():
		//delegate_signature => "delegate(" guid ")"
		guid, err := typeDef.GUID()
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(`delegate({%s})`, guid), nil
	case typeDef.IsRuntimeClass():
		// Static only classes carry the abstract flag.
		// These cannot be instantiated so no signature needed.
		if typeDef.Flags.Abstract() {
			return "", nil
		}

		// runtime_class_signature => "rc(" runtime_class_name ";" default_interface ")"

		// Runtime classes must specify the DefaultAttribute on exactly one of their InterfaceImpl rows.
		defaultInterface, err := typeDef.GetAttributeWithType(winmd.AttributeTypeDefaultAttribute)
		if err != nil {
			// Some classes (Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementWatcher) do not
			// define a runtime class. I'm not sure if this is an error in the IDL or the documentation.
			// But we are not going to fail here. Just default to the first implemented interface
			ifs, ifserr := typeDef.GetImplementedInterfaces()
			if ifserr != nil {
				return "", err
			}

			if len(ifs) == 0 {
				return "", err
			}
			defaultInterface = []byte(ifs[0].Namespace + "." + ifs[0].Name)
		}

		td, err := g.mdStore.TypeDefByName(string(defaultInterface))
		if err != nil {
			return "", err
		}

		defaultInterfaceSignature, err := g.Signature(td)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`rc(%s;%s)`, typeDef.TypeNamespace+"."+typeDef.TypeName, defaultInterfaceSignature), nil
	default:
		return "", fmt.Errorf("unsupported type: %v", typeDef.TypeName)
	}
}

func primitiveTypeSignature(kind types.ElementTypeKind) string {
	switch kind {
	// Fundamental types
	case types.ELEMENT_TYPE_U1:
		return winrt.SignatureUInt8
	case types.ELEMENT_TYPE_U2:
		return winrt.SignatureUInt16
	case types.ELEMENT_TYPE_U4:
		return winrt.SignatureUInt32
	case types.ELEMENT_TYPE_U8:
		return winrt.SignatureUInt64
	case types.ELEMENT_TYPE_I1:
		return winrt.SignatureInt8
	case types.ELEMENT_TYPE_I2:
		return winrt.SignatureInt16
	case types.ELEMENT_TYPE_I4:
		return winrt.SignatureInt32
	case types.ELEMENT_TYPE_I8:
		return winrt.SignatureInt64
	case types.ELEMENT_TYPE_R4:
		return winrt.SignatureFloat32
	case types.ELEMENT_TYPE_R8:
		return winrt.SignatureFloat64
	case types.ELEMENT_TYPE_BOOLEAN:
		return winrt.SignatureBool
	case types.ELEMENT_TYPE_CHAR:
		return winrt.SignatureChar
	case types.ELEMENT_TYPE_STRING:
		return winrt.SignatureString
	}
	return ""
}
