package codegen

import (
	"embed"
	"strings"
	"text/template"
)

type genData struct {
	Imports []genImport
	Types   []genType
	Funcs   []genFunc
}

type genImport struct {
	Alias string
	Name  string
}

type genType struct {
	Name            string
	EmbeddedStructs []qualifiedID
	Methods         []genMethod
}

type genMethod struct {
	Name     string
	IsStatic bool
}

type qualifiedID struct {
	Package string
	Name    string
}

type genFunc struct{}

//go:embed templates/*
var templatesFS embed.FS

func getTemplates() (*template.Template, error) {
	return template.New("").
		Funcs(funcs()).
		ParseFS(templatesFS, "templates/*")
}

func funcs() template.FuncMap {
	return template.FuncMap{
		"methodName": methodName,
	}
}

func methodName(m genMethod) string {
	switch {
	case m.Name == ".ctor":
		return "New"
	case strings.HasPrefix(m.Name, "get_"):
		return strings.Replace(m.Name, "get_", "Get", 1)
	case strings.HasPrefix(m.Name, "put_"):
		return strings.Replace(m.Name, "put_", "Set", 1)
	}
	return m.Name
}
