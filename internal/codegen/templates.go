package codegen

import (
	"embed"
	"strings"
	"text/template"

	"github.com/tdakkota/win32metadata/types"
)

type genData struct {
	Package string
	Imports []string
	Types   []genType
}

type genType struct {
	Name  string
	Funcs []genFunc
}

type genFunc struct {
	Name          string
	IsConstructor bool

	FuncOwner      string
	ParentType     types.TypeDef
	ParentTypeGUID string
	RuntimeClass   types.TypeDef

	Signature types.Blob

	InParams    []genParam
	ReturnParam *genParam // this may be nil
}

type genParam struct {
	Name         string
	Type         string
	DefaultValue string
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
	}
	return m.Name
}
