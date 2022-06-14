package codegen

import (
	"embed"
	"strings"
	"text/template"

	"github.com/tdakkota/win32metadata/types"
)

type genData struct {
	Package    string
	Imports    []string
	Classes    []genClass
	Interfaces []genInterface
}

type genInterface struct {
	Name  string
	GUID  string
	Funcs []genFunc
}

type genClass struct {
	Name             string
	ImplInterfaces   []string
	StaticInterfaces []genInterface
}

type genFunc struct {
	Name      string
	Implement bool

	FuncOwner   string
	InParams    []genParam
	ReturnParam *genParam // this may be nil

	ExclusiveTo *types.TypeDef
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
