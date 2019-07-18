package proto

import (
	"os"
	"text/template"
)

func Render(c contract) error {
	t := template.New("test")

	t, err := t.Parse(tmpl)
	if err != nil {
		return err
	}

	out, err := os.OpenFile("test.proto", os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	if err := t.Execute(out, c); err != nil {
		return err
	}

	return nil
}

const tmpl = `// Auto Generated. DO NOT EDIT!
syntax = "proto3";
package airbloc.{{.PackageName}};

import "google/protobuf/empty.proto";

{{range .Services}}service {{.Name}} {
	{{range .Rpcs}}rpc {{.Name}}({{.Input}}) returns ({{.Output}});
	{{end}}
}{{end}}
{{range .Messages}}// {{.Comment}}
message {{.Name}} {
	{{range .Args}}{{if .Repeated}}repeated{{end}} {{.Type}} {{.Name}} = {{.Count}};
	{{end}}
}
{{end}}
`
