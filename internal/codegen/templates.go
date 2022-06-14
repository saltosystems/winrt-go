package codegen

import (
	"embed"
	"strings"
	"text/template"
)

type qualifiedID struct {
	Namespace string
	Name      string
}

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
	Name                string
	FullyQualifiedName  string
	ImplInterfaces      []string
	ExclusiveInterfaces []genInterface
	HasEmptyConstructor bool
}

type genFunc struct {
	Name        string
	Implement   bool
	FuncOwner   string
	InParams    []genParam
	ReturnParam *genParam // this may be nil

	// ExclusiveTo is the name of the class that this function is exclusive to.
	// The funcion will be called statically using the RoGetActivationFactory function.
	ExclusiveTo string

	RequiresActivation bool
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
	case strings.HasPrefix(m.Name, "add_"):
		return strings.Replace(m.Name, "add_", "Add", 1)
	case strings.HasPrefix(m.Name, "remove_"):
		return strings.Replace(m.Name, "remove_", "Remove", 1)
	}
	return m.Name
}
