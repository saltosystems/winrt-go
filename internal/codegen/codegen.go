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
	attributeTypeGUID = "Windows.Foundation.Metadata.GuidAttribute"
)

type classNotFoundError struct {
	class string
}

func (e *classNotFoundError) Error() string {
	return fmt.Sprintf("class %s was not found", e.class)
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
		return g.loadInterfaceData(typeDef)
	}

	return g.loadClassData(typeDef)
}

// https://docs.microsoft.com/en-us/uwp/winrt-cref/winmd-files#interfaces
func (g *generator) loadInterfaceData(typeDef types.TypeDef) error {
	// Any WinRT interface with private visibility must have a single ExclusiveToAttribute.
	// the ExclusiveToAttribute must reference a runtime class.

	// we do not support generating these types of classes, they are exclusive to a runtime class,
	// and thus will be generated when the runtime class is generated.

	if typeDef.Flags.NotPublic() {
		return fmt.Errorf("interface %s is not public", typeDef.TypeNamespace+"."+typeDef.TypeName)

	}

	// Any WinRT interface with public visibility must not have an ExclusiveToAttribute.
	// Adding it as a documentation, there's no point on checking if this is true.

	funcs, err := g.getGenFuncs(typeDef, typeDef, false)
	if err != nil {
		return err
	}

	// Interfaces' TypeDef rows must have a GuidAttribute as well as a VersionAttribute.
	guid, err := g.typeGUID(typeDef)
	if err != nil {
		return err
	}

	g.addType(genType{
		Name:  typeDef.TypeName,
		GUID:  guid,
		Funcs: funcs,
	})
	return nil
}

func (g *generator) loadClassData(typeDef types.TypeDef) error {
	// class interface
	funcs, err := g.getGenFuncs(typeDef, typeDef, false)
	if err != nil {
		return err
	}

	g.addType(genType{
		Name:  typeDef.TypeName,
		Funcs: funcs,
	})
	return nil
}

func (g *generator) getCustomAttributeForClassWithTypeClass(typeDef types.TypeDef, lookupAttrTypeClass string) ([]byte, error) {
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
			return cAttr.Value, nil
		}
	}

	return nil, fmt.Errorf("could not find CustomAttribute for %s.%s with type %s", typeDef.TypeNamespace, typeDef.TypeName, lookupAttrTypeClass)
}

func (g *generator) addType(gt genType) {
	g.genData.Types = append(g.genData.Types, gt)
}

func (g *generator) addImportFor(ns, name string) {
	folder := typeToFolder(ns, name)
	i := "github.com/saltosystems/winrt-go/" + folder
	g.genData.Imports = append(g.genData.Imports, i)
}

func (g *generator) getGenFuncs(typeDef, runtimeClass types.TypeDef, activatable bool) ([]genFunc, error) {
	var genFuncs []genFunc

	if activatable {
		var parentGUID string
		parentGUID, err := g.typeGUID(typeDef)
		if err != nil {
			// the type may not contain this information, just ignore it
			_ = level.Warn(g.logger).Log("msg", "failed to get type GUID", "type", typeDef.TypeNamespace+"."+typeDef.TypeName, "err", err)
		}

		// add the constructor
		genFuncs = append(genFuncs, genFunc{
			Name:          "Activate" + typeDef.TypeName,
			IsConstructor: true,
			InParams:      []genParam{},
			ReturnParam: &genParam{
				Type:         "*" + typeDef.TypeName,
				DefaultValue: "nil",
			},
			Signature:      nil,
			ParentType:     typeDef,
			ParentTypeGUID: parentGUID,
			RuntimeClass:   runtimeClass,
			FuncOwner:      "",
		})
	}

	methods, err := typeDef.ResolveMethodList(g.winmdCtx)
	if err != nil {
		return nil, err
	}

	for _, m := range methods {
		generatedFunc, err := g.genFuncFromMethod(typeDef, runtimeClass, m)
		if err != nil {
			return nil, err
		}
		genFuncs = append(genFuncs, *generatedFunc)
	}

	return genFuncs, nil
}

func (g *generator) genFuncFromMethod(typeDef, runtimeClass types.TypeDef, m types.MethodDef) (*genFunc, error) {
	params, err := g.getInParameters(typeDef, runtimeClass, m)
	if err != nil {
		return nil, err
	}

	retParam, err := g.getReturnParameters(typeDef, runtimeClass, m)
	if err != nil {
		return nil, err
	}

	var parentGUID string
	parentGUID, err = g.typeGUID(typeDef)
	if err != nil {
		// the type may not contain this information, just ignore it
		_ = level.Warn(g.logger).Log("msg", "failed to get type GUID", "type", typeDef.TypeNamespace+"."+typeDef.TypeName, "err", err)
	}

	return &genFunc{
		Name:           m.Name,
		Implement:      g.shouldImplementMethod(m),
		IsConstructor:  false,
		InParams:       params,
		ReturnParam:    retParam,
		Signature:      m.Signature,
		ParentType:     typeDef,
		ParentTypeGUID: parentGUID,
		RuntimeClass:   runtimeClass,
		FuncOwner:      typeDef.TypeName,
	}, nil
}

func (g *generator) shouldImplementMethod(m types.MethodDef) bool {
	return g.methodFilter.Filter(m.Name)
}

func (g *generator) typeGUID(typeDef types.TypeDef) (string, error) {
	blob, err := g.getCustomAttributeForClassWithTypeClass(typeDef, attributeTypeGUID)
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

func (g *generator) getInParameters(t, rt types.TypeDef, m types.MethodDef) ([]genParam, error) {

	params, err := m.ResolveParamList(g.winmdCtx)
	if err != nil {
		return nil, err
	}

	// the signature contains the parameter
	// types and return type of the method
	r := m.Signature.Reader()
	mr, err := r.Method(g.winmdCtx)
	if err != nil {
		return nil, err
	}

	genParams := []genParam{}
	for i, e := range mr.Params {
		genParams = append(genParams, genParam{
			Name: getParamName(params, uint16(i+1)),
			Type: g.elementType(e, typePackage(rt.TypeNamespace, rt.TypeName)),
		})
	}

	return genParams, nil
}

func (g *generator) getReturnParameters(t, rt types.TypeDef, m types.MethodDef) (*genParam, error) {
	// the signature contains the parameter
	// types and return type of the method
	r := m.Signature.Reader()
	methodSignature, err := r.Method(g.winmdCtx)
	if err != nil {
		return nil, err
	}

	// ignore void types
	if methodSignature.Return.Type.Kind == types.ELEMENT_TYPE_VOID {
		return nil, nil
	}

	return &genParam{
		Name:         "",
		Type:         g.elementType(methodSignature.Return, typePackage(rt.TypeNamespace, rt.TypeName)),
		DefaultValue: g.elementDefaultValue(methodSignature.Return),
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

func (g *generator) elementType(e types.Element, currentPkg string) string {
	switch e.Type.Kind {
	case types.ELEMENT_TYPE_BOOLEAN:
		return "bool"
	case types.ELEMENT_TYPE_CHAR:
		return "byte"
	case types.ELEMENT_TYPE_I1:
		return "int8"
	case types.ELEMENT_TYPE_U1:
		return "uint8"
	case types.ELEMENT_TYPE_I2:
		return "int16"
	case types.ELEMENT_TYPE_U2:
		return "uint16"
	case types.ELEMENT_TYPE_I4:
		return "int32"
	case types.ELEMENT_TYPE_U4:
		return "uint32"
	case types.ELEMENT_TYPE_I8:
		return "int64"
	case types.ELEMENT_TYPE_U8:
		return "uint64"
	case types.ELEMENT_TYPE_R4:
		return "float32"
	case types.ELEMENT_TYPE_R8:
		return "float64"
	case types.ELEMENT_TYPE_STRING:
		return "string"
	case types.ELEMENT_TYPE_CLASS:
		// return class name
		namespace, name, err := g.winmdCtx.ResolveTypeDefOrRefName(e.Type.TypeDef.Index)
		if err != nil {
			return "__ERROR_ELEMENT_TYPE_CLASS__"
		}

		// this may be the runtime class itself, but we only know the interfaces
		// TODO: make this smarter
		if !strings.HasPrefix(name, "I") {
			name = "I" + name
		}

		// name is always an interface, so we need to remove the initial I
		typePkg := typePackage(namespace, name[1:])
		if currentPkg != typePkg {
			g.addImportFor(namespace, name[1:])
			name = typePkg + "." + name
		}

		return "*" + name
	default:
		return "__ERROR_" + e.Type.Kind.String() + "__"
	}
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
