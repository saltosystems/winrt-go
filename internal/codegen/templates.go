package codegen

import (
	"embed"
	"strings"
	"text/template"

	"github.com/saltosystems/winrt-go/internal/winmd"
)

type genDataFile struct {
	Filename string
	Data     genData
}

type genData struct {
	Package         string
	Imports         []string
	Classes         []genClass
	Enums           []genEnum
	Interfaces      []genInterface
	Structs         []genStruct
	Delegates       []genDelegate
	DelegateExports []genDelegate
}

func (g *genData) ComputeImports(typeDef *winmd.TypeDef) {
	// gather all imports
	imports := make([]genImport, 0)
	if g.Classes != nil {
		for _, c := range g.Classes {
			imports = append(imports, c.GetRequiredImports()...)
		}
	}
	if g.Interfaces != nil {
		for _, i := range g.Interfaces {
			imports = append(imports, i.GetRequiredImports()...)
		}
	}

	for _, i := range imports {
		if typeDef.TypeNamespace != i.Namespace {
			g.Imports = append(g.Imports, i.ToGoImport())
		}
	}
}

type genInterface struct {
	Name  string
	GUID  string
	Funcs []genFunc
}

func (g *genInterface) GetRequiredImports() []genImport {
	imports := make([]genImport, 0)
	for _, f := range g.Funcs {
		imports = append(imports, f.RequiresImports...)
	}
	return imports
}

type genClass struct {
	Name                string
	RequiresImports     []genImport
	FullyQualifiedName  string
	ImplInterfaces      []string
	ExclusiveInterfaces []genInterface
	HasEmptyConstructor bool
}

func (g *genClass) GetRequiredImports() []genImport {
	imports := make([]genImport, 0)
	if g.RequiresImports != nil {
		imports = append(imports, g.RequiresImports...)
	}
	if g.ExclusiveInterfaces != nil {
		for _, i := range g.ExclusiveInterfaces {
			imports = append(imports, i.GetRequiredImports()...)
		}
	}

	return imports
}

type genDelegate struct {
	Name        string
	GUID        string
	InParams    []*genParam
	ReturnParam *genParam // this may be nil
}

type genEnum struct {
	Name   string
	Type   string
	Values []genEnumValue
}
type genEnumValue struct {
	Name  string
	Value string
}

type genFunc struct {
	Name            string
	RequiresImports []genImport
	Implement       bool
	FuncOwner       string
	InParams        []*genParam
	ReturnParam     *genParam // this may be nil

	// ExclusiveTo is the name of the class that this function is exclusive to.
	// The funcion will be called statically using the RoGetActivationFactory function.
	ExclusiveTo string

	RequiresActivation bool
}

type genImport struct {
	Namespace, Name string
}

func (i genImport) ToGoImport() string {
	folder := typeToFolder(i.Namespace, i.Name)
	return "github.com/saltosystems/winrt-go/" + folder
}

type genParam struct {
	Name         string
	Type         string
	IsPointer    bool
	DefaultValue string

	genType         *genParamReference
	genDefaultValue *genParamReference
}

type genParamReference struct {
	Namespace   string
	Name        string
	IsPointer   bool
	IsPrimitive bool
}

func (g genParamReference) GoParamString(callerPackage string) string {
	if g.IsPrimitive {
		return g.Name
	}

	name := typeNameToGoName(g.Name, true) // assume all are public

	pkg := typePackage(g.Namespace, g.Name)
	if callerPackage != pkg {
		name = pkg + "." + name
	}

	if g.IsPointer {
		name = "*" + name
	}

	return name
}

type genStruct struct {
	Name   string
	Fields []*genParam
}

//go:embed templates/*
var templatesFS embed.FS

func getTemplates() (*template.Template, error) {
	return template.New("").
		Funcs(funcs()).
		ParseFS(templatesFS, "templates/*")
}

func funcs() template.FuncMap {
	return template.FuncMap{
		"funcName": funcName,
	}
}

// funcName is used to generate the name of a function.
func funcName(m genFunc) string {
	switch {
	case strings.HasPrefix(m.Name, "get_"):
		return strings.Replace(m.Name, "get_", "Get", 1)
	case strings.HasPrefix(m.Name, "put_"):
		return strings.Replace(m.Name, "put_", "Set", 1)
	case strings.HasPrefix(m.Name, "add_"):
		return strings.Replace(m.Name, "add_", "Add", 1)
	case strings.HasPrefix(m.Name, "remove_"):
		return strings.Replace(m.Name, "remove_", "Remove", 1)
	}
	return m.Name
}

func typeToFolder(ns, name string) string {
	fullName := ns
	return strings.ToLower(strings.Replace(fullName, ".", "/", -1))
}

func typePackage(ns, name string) string {
	sns := strings.Split(ns, ".")
	return strings.ToLower(sns[len(sns)-1])
}

func enumName(typeName string, enumName string) string {
	return typeName + enumName
}

func typeDefGoName(typeName string, public bool) string {
	name := typeName

	if isParameterizedName(typeName) {
		name = strings.Split(name, "`")[0]
	}

	if !public {
		name = strings.ToLower(name[0:1]) + name[1:]
	}
	return name
}

func isParameterizedName(typeName string) bool {
	// parameterized types contain a '`' followed by the amount of generic parameters in their name.
	return strings.Contains(typeName, "`")
}

func typeFilename(typeName string) string {
	// public boolean is not relevant, we are going to lower everything
	goname := typeDefGoName(typeName, true)
	return strings.ToLower(goname)
}

// removes Go reserved words from param names
func cleanReservedWords(name string) string {
	switch name {
	case "type":
		return "mType"
	}
	return name
}

func typeNameToGoName(typeName string, public bool) string {
	name := typeName

	if isParameterizedName(typeName) {
		name = strings.Split(name, "`")[0]
	}

	if !public {
		name = strings.ToLower(name[0:1]) + name[1:]
	}
	return name
}
