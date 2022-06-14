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
	"github.com/saltosystems/winrt-go/winmd"
	"github.com/tdakkota/win32metadata/md"
	"github.com/tdakkota/win32metadata/types"
)

const (
	attributeTypeGUID                 = "Windows.Foundation.Metadata.GuidAttribute"
	attributeTypeExclusiveTo          = "Windows.Foundation.Metadata.ExclusiveToAttribute"
	attributeTypeStaticAttribute      = "Windows.Foundation.Metadata.StaticAttribute"
	attributeTypeActivatableAttribute = "Windows.Foundation.Metadata.ActivatableAttribute"
)

type classNotFoundError struct {
	class string
}

func (e *classNotFoundError) Error() string {
	return fmt.Sprintf("class %s was not found", e.class)
}

type qualifiedID struct {
	Namespace string
	Name      string
}

type generator struct {
	class        string
	methodFilter *MethodFilter

	logger log.Logger

	genData  *genData
	winmdCtx *types.Context
}

// Generate generates the code for the given config.
func Generate(cfg *Config, logger log.Logger) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	g := &generator{
		class:        cfg.Class,
		methodFilter: cfg.MethodFilter(),
		logger:       logger,
	}
	return g.run()
}

func (g *generator) run() error {
	_ = level.Debug(g.logger).Log("msg", "starting code generation", "class", g.class)

	winmdFiles, err := winmd.AllFiles()
	if err != nil {
		return err
	}

	// we don't know which winmd file contains the class, so we have to iterate over all of them
	for _, f := range winmdFiles {
		winmdCtx, err := parseWinMDFile(f.Name())
		if err != nil {
			return err
		}
		g.winmdCtx = winmdCtx

		typeDef, err := g.typeDefByName(g.class)
		if err != nil {
			// class not found errors are ok
			if _, ok := err.(*classNotFoundError); ok {
				continue
			}

			return err
		}

		return g.generate(*typeDef)
	}

	return fmt.Errorf("class %s was not found", g.class)

}

func (g *generator) typeDefByName(class string) (*types.TypeDef, error) {
	typeDefTable := g.winmdCtx.Table(md.TypeDef)
	for i := uint32(0); i < typeDefTable.RowCount(); i++ {
		var typeDef types.TypeDef
		if err := typeDef.FromRow(typeDefTable.Row(i)); err != nil {
			return nil, err
		}

		if typeDef.TypeNamespace+"."+typeDef.TypeName == class {
			return &typeDef, nil
		}
	}

	return nil, &classNotFoundError{class: class}
}

func (g *generator) generate(typeDef types.TypeDef) error {

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

	// format the output source code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// write unformatted source code to file as a debugging mechanism
		_, _ = file.Write(buf.Bytes())
		return err
	}

	// and write it to file
	_, err = file.Write(formatted)

	return err
}

func typeToFolder(ns, name string) string {
	fullName := ns
	return strings.ToLower(strings.Replace(fullName, ".", "/", -1))
}

func typePackage(ns, name string) string {
	sns := strings.Split(ns, ".")
	return strings.ToLower(sns[len(sns)-1])
}

func parseWinMDFile(path string) (*types.Context, error) {
	f, err := winmd.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return types.FromPE(f)
}

func (g *generator) loadCodeGenData(typeDef types.TypeDef) error {
	g.genData = &genData{
		Package: typePackage(typeDef.TypeNamespace, typeDef.TypeName),
	}

	if typeDef.Flags.Interface() {
		if err := g.validateInterface(typeDef); err != nil {
			return err
		}

		iface, err := g.createGenInterface(typeDef)
		if err != nil {
			return err
		}
		g.genData.Interfaces = append(g.genData.Interfaces, *iface)
	} else {
		class, err := g.createGenClass(typeDef)
		if err != nil {
			return err
		}
		g.genData.Classes = append(g.genData.Classes, *class)
	}
	return nil
}

func (g *generator) validateInterface(typeDef types.TypeDef) error {
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
func (g *generator) createGenInterface(typeDef types.TypeDef) (*genInterface, error) {
	// Any WinRT interface with public visibility must not have an ExclusiveToAttribute.

	funcs, err := g.getGenFuncs(typeDef)
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
func (g *generator) createGenClass(typeDef types.TypeDef) (*genClass, error) {
	exclusiveInterfaces := make([]*types.TypeDef, 0)

	// get all the interfaces this class implements
	interfaces, err := g.getImplementedInterfaces(typeDef)
	if err != nil {
		return nil, err
	}
	implInterfaces := make([]string, 0, len(interfaces))
	for _, iface := range interfaces {
		pkg := ""
		ifaceNS, ifaceName, err := g.winmdCtx.ResolveTypeDefOrRefName(iface)
		if err != nil {
			return nil, err
		}
		if typeDef.TypeNamespace != ifaceNS {
			pkg = typePackage(ifaceNS, ifaceName) + "."
			g.addImportFor(ifaceNS, ifaceName)
		}
		implInterfaces = append(implInterfaces, pkg+ifaceName)
	}

	// Runtime classes have zero or more StaticAttribute custom attributes
	// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#static-interfaces
	staticAttributeBlobs := g.getTypeDefAttributesWithType(typeDef, attributeTypeStaticAttribute)
	staticInterfaces := make([]genInterface, 0, len(staticAttributeBlobs))
	for _, blob := range staticAttributeBlobs {
		class := extractClassFromBlob(blob)
		_ = level.Debug(g.logger).Log("msg", "found static interface", "class", class)
		staticClass, err := g.typeDefByName(class)
		if err != nil {
			_ = level.Error(g.logger).Log("msg", "static class defined in StaticAttribute not found", "class", class, "err", err)
			return nil, err
		}

		exclusiveInterfaces = append(exclusiveInterfaces, staticClass)
	}

	// Runtime classes have zero or more ActivatableAttribute custom attributes
	// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#activation
	activatableAttributeBlobs := g.getTypeDefAttributesWithType(typeDef, attributeTypeActivatableAttribute)
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
		activatableClass, err := g.typeDefByName(class)
		if err != nil {
			// the activatable class may be empty in some cases, example:
			// https://github.com/tpn/winsdk-10/blob/9b69fd26ac0c7d0b83d378dba01080e93349c2ed/Include/10.0.14393.0/winrt/windows.devices.bluetooth.advertisement.idl#L518
			_ = level.Error(g.logger).Log("msg", "activatable class defined in ActivatableAttribute not found", "class", class, "err", err)

			// so do not fail
			continue
		}
		exclusiveInterfaces = append(exclusiveInterfaces, activatableClass)
	}

	// generate exclusive interfaces
	for _, iface := range exclusiveInterfaces {
		iface, err := g.createGenInterface(*iface)
		if err != nil {
			return nil, err
		}
		// if all methods from the statics interface have been filtered, then we can skip it
		for _, m := range iface.Funcs {
			if m.Implement {
				// we only add the interface if it has at least one implemented method
				staticInterfaces = append(staticInterfaces, *iface)
				break
			}
		}
	}

	return &genClass{
		Name:                typeDefGoName(typeDef),
		FullyQualifiedName:  typeDef.TypeNamespace + "." + typeDef.TypeName,
		ImplInterfaces:      implInterfaces,
		StaticInterfaces:    staticInterfaces,
		HasEmptyConstructor: hasEmptyConstructor,
	}, nil
}

func typeDefGoName(typeDef types.TypeDef) string {
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

func (g *generator) getTypeDefAttributeWithType(typeDef types.TypeDef, lookupAttrTypeClass string) ([]byte, error) {
	result := g.getTypeDefAttributesWithType(typeDef, lookupAttrTypeClass)
	if len(result) == 0 {
		return nil, fmt.Errorf("type %s has no custom attribute %s", typeDef.TypeNamespace+"."+typeDef.TypeName, lookupAttrTypeClass)
	} else if len(result) > 1 {
		_ = level.Warn(g.logger).Log(
			"msg", "type has multiple custom attributes, returning the first one",
			"type", typeDef.TypeNamespace+"."+typeDef.TypeName,
			"attr", lookupAttrTypeClass,
		)
	}

	return result[0], nil
}

func (g *generator) getTypeDefAttributesWithType(typeDef types.TypeDef, lookupAttrTypeClass string) [][]byte {
	result := make([][]byte, 0)

	cAttrTable := g.winmdCtx.Table(md.CustomAttribute)
	for i := uint32(0); i < cAttrTable.RowCount(); i++ {
		var cAttr types.CustomAttribute
		if err := cAttr.FromRow(cAttrTable.Row(i)); err != nil {
			continue
		}

		// - Parent: The owner of the Attribute must be the given typeDef
		if cAttrParentTable, _ := cAttr.Parent.Table(); cAttrParentTable != md.TypeDef {
			continue
		}

		var parentTypeDef types.TypeDef
		row, ok := cAttr.Parent.Row(g.winmdCtx)
		if !ok {
			continue
		}
		if err := parentTypeDef.FromRow(row); err != nil {
			continue
		}

		// does the blob belong to the type we're looking for?
		if parentTypeDef.TypeNamespace != typeDef.TypeNamespace || parentTypeDef.TypeName != typeDef.TypeName {
			continue
		}

		// - Type: the attribute type must be the given type
		// the cAttr.Type table can be either a MemberRef or a MethodRef.
		// Since we are looking for a type, we will only consider the MemberRef.
		if cAttrTypeTable, _ := cAttr.Type.Table(); cAttrTypeTable != md.MemberRef {
			continue
		}

		var attrTypeMemberRef types.MemberRef
		row, ok = cAttr.Type.Row(g.winmdCtx)
		if !ok {
			continue
		}
		if err := attrTypeMemberRef.FromRow(row); err != nil {
			continue
		}

		// we need to check the MemberRef Class
		// the value can belong to several tables, but we are only going to check for TypeRef
		if classTable, _ := attrTypeMemberRef.Class.Table(); classTable != md.TypeRef {
			continue
		}

		var attrTypeRef types.TypeRef
		row, ok = attrTypeMemberRef.Class.Row(g.winmdCtx)
		if !ok {
			continue
		}
		if err := attrTypeRef.FromRow(row); err != nil {
			continue
		}

		if attrTypeRef.TypeNamespace+"."+attrTypeRef.TypeName == lookupAttrTypeClass {
			result = append(result, cAttr.Value)
		}
	}

	return result
}

func (g *generator) getImplementedInterfaces(typeDef types.TypeDef) ([]types.TypeDefOrRef, error) {
	interfaces := make([]types.TypeDefOrRef, 0)

	tableInterfaceImpl := g.winmdCtx.Table(md.InterfaceImpl)
	for i := uint32(0); i < tableInterfaceImpl.RowCount(); i++ {
		var interfaceImpl types.InterfaceImpl
		if err := interfaceImpl.FromRow(tableInterfaceImpl.Row(i)); err != nil {
			return nil, err
		}

		classTd, err := interfaceImpl.ResolveClass(g.winmdCtx)
		if err != nil {
			return nil, err
		}

		if classTd.TypeNamespace+"."+classTd.TypeName != typeDef.TypeNamespace+"."+typeDef.TypeName {
			// not the class we are looking for
			continue
		}

		if t, ok := interfaceImpl.Interface.Table(); ok && t == md.TypeSpec {
			// ignore type spec rows
			continue
		}

		interfaces = append(interfaces, interfaceImpl.Interface)
	}

	return interfaces, nil
}

func (g *generator) addImportFor(ns, name string) {
	folder := typeToFolder(ns, name)
	i := "github.com/saltosystems/winrt-go/" + folder
	g.genData.Imports = append(g.genData.Imports, i)
}

func (g *generator) getGenFuncs(typeDef types.TypeDef) ([]genFunc, error) {
	var genFuncs []genFunc

	methods, err := typeDef.ResolveMethodList(g.winmdCtx)
	if err != nil {
		return nil, err
	}

	exclusiveToBlob, err := g.getTypeDefAttributeWithType(typeDef, attributeTypeExclusiveTo)
	var exclusiveToType string
	// an error here is fine, we just won't have the ExclusiveTo attribute
	if err == nil {
		exclusiveToClass := extractClassFromBlob(exclusiveToBlob)
		exclusiveToTypeCandidate, err := g.typeDefByName(exclusiveToClass)
		if err != nil {
			return nil, err
		}
		exclusiveToType = exclusiveToTypeCandidate.TypeNamespace + "." + exclusiveToTypeCandidate.TypeName
	}

	for _, m := range methods {
		generatedFunc, err := g.genFuncFromMethod(typeDef, m, exclusiveToType)
		if err != nil {
			return nil, err
		}
		genFuncs = append(genFuncs, *generatedFunc)
	}

	return genFuncs, nil
}

func (g *generator) genFuncFromMethod(typeDef types.TypeDef, m types.MethodDef, exclusiveTo string) (*genFunc, error) {
	params, reqInImports, err := g.getInParameters(typeDef, m)
	if err != nil {
		return nil, err
	}

	retParam, reqOutImports, err := g.getReturnParameters(typeDef, m)
	if err != nil {
		return nil, err
	}

	// add the type imports to the top of the file
	// only if the method is going to be implemented
	implement := g.shouldImplementMethod(m)
	if implement {
		curPackage := typePackage(typeDef.TypeNamespace, typeDef.TypeName)

		var allImports []qualifiedID
		allImports = append(allImports, reqInImports...)
		allImports = append(allImports, reqOutImports...)

		for _, i := range allImports {
			pkg := typePackage(i.Namespace, i.Name)
			if curPackage != pkg {
				// imports are addded globally
				g.addImportFor(i.Namespace, i.Name)
			}
		}
	}

	return &genFunc{
		Name:        m.Name,
		Implement:   implement,
		InParams:    params,
		ReturnParam: retParam,
		FuncOwner:   typeDefGoName(typeDef),
		ExclusiveTo: exclusiveTo,
	}, nil
}

func (g *generator) shouldImplementMethod(m types.MethodDef) bool {
	return g.methodFilter.Filter(m.Name)
}

func (g *generator) typeGUID(typeDef types.TypeDef) (string, error) {
	blob, err := g.getTypeDefAttributeWithType(typeDef, attributeTypeGUID)
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

func (g *generator) getInParameters(typeDef types.TypeDef, m types.MethodDef) ([]genParam, []qualifiedID, error) {

	params, err := m.ResolveParamList(g.winmdCtx)
	if err != nil {
		return nil, nil, err
	}

	// the signature contains the parameter
	// types and return type of the method
	r := m.Signature.Reader()
	mr, err := r.Method(g.winmdCtx)
	if err != nil {
		return nil, nil, err
	}

	genParams := []genParam{}
	var reqImports []qualifiedID
	for i, e := range mr.Params {
		elType, requiredImports := g.elementType(e)
		reqImports = append(reqImports, requiredImports...)
		genParams = append(genParams, genParam{
			Name: getParamName(params, uint16(i+1)),
			Type: elType,
		})
	}

	return genParams, reqImports, nil
}

func (g *generator) getReturnParameters(typeDef types.TypeDef, m types.MethodDef) (*genParam, []qualifiedID, error) {
	// the signature contains the parameter
	// types and return type of the method
	r := m.Signature.Reader()
	methodSignature, err := r.Method(g.winmdCtx)
	if err != nil {
		return nil, nil, err
	}

	// ignore void types
	if methodSignature.Return.Type.Kind == types.ELEMENT_TYPE_VOID {
		return nil, nil, nil
	}

	elType, reqImports := g.elementType(methodSignature.Return)
	return &genParam{
		Name:         "",
		Type:         elType,
		DefaultValue: g.elementDefaultValue(methodSignature.Return),
	}, reqImports, nil
}

func getParamName(params []types.Param, i uint16) string {
	for _, p := range params {
		if p.Flags.In() && p.Sequence == i {
			return p.Name
		}
	}
	return "__ERROR__"
}

func (g *generator) elementType(e types.Element) (string, []qualifiedID) {
	var elType string
	var requiredImports []qualifiedID
	switch e.Type.Kind {
	case types.ELEMENT_TYPE_BOOLEAN:
		elType = "bool"
	case types.ELEMENT_TYPE_CHAR:
		elType = "byte"
	case types.ELEMENT_TYPE_I1:
		elType = "int8"
	case types.ELEMENT_TYPE_U1:
		elType = "uint8"
	case types.ELEMENT_TYPE_I2:
		elType = "int16"
	case types.ELEMENT_TYPE_U2:
		elType = "uint16"
	case types.ELEMENT_TYPE_I4:
		elType = "int32"
	case types.ELEMENT_TYPE_U4:
		elType = "uint32"
	case types.ELEMENT_TYPE_I8:
		elType = "int64"
	case types.ELEMENT_TYPE_U8:
		elType = "uint64"
	case types.ELEMENT_TYPE_R4:
		elType = "float32"
	case types.ELEMENT_TYPE_R8:
		elType = "float64"
	case types.ELEMENT_TYPE_STRING:
		elType = "string"
	case types.ELEMENT_TYPE_CLASS:
		// return class name
		namespace, name, err := g.winmdCtx.ResolveTypeDefOrRefName(e.Type.TypeDef.Index)
		if err != nil {
			elType = "__ERROR_ELEMENT_TYPE_CLASS__"
		} else {
			requiredImports = []qualifiedID{{namespace, name}}
			elType = "*" + name
		}
	default:
		elType = "__ERROR_" + e.Type.Kind.String() + "__"
	}

	return elType, requiredImports
}

func (g *generator) elementDefaultValue(e types.Element) string {
	switch e.Type.Kind {
	case types.ELEMENT_TYPE_BOOLEAN:
		return "false"
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
		return "0"
	case types.ELEMENT_TYPE_R4:
		fallthrough
	case types.ELEMENT_TYPE_R8:
		return "0.0"
	case types.ELEMENT_TYPE_STRING:
		return "\"\""
	case types.ELEMENT_TYPE_CLASS:
		return "nil"
	default:
		return "__ERROR_" + e.Type.Kind.String() + "__"
	}
}
