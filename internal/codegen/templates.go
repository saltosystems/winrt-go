package codegen

import (
	"embed"
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

type genType struct{}
type genFunc struct{}

//go:embed templates/*
var templatesFS embed.FS

func getTemplates() (*template.Template, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*")
	if err != nil {
		return nil, err
	}

	return tmpl.Funcs(funcs()), nil
}

func funcs() template.FuncMap {
	return template.FuncMap{
		//"toLower": strings.ToLower,
		//"toUpper": strings.ToUpper,
	}
}
