package proto

import (
	"os"
	"text/template"
)

func Render(path string, c Contract) error {
	t := template.New("test")

	t, err := t.Parse(tmpl)
	if err != nil {
		return err
	}

	var out *os.File
	if _, err := os.Stat(path); os.IsNotExist(err) {
		out, err = os.Create(path)
		if err != nil {
			return err
		}
	} else {
		out, err = os.OpenFile(path, os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if err := t.Execute(out, c); err != nil {
		return err
	}

	return nil
}

const tmpl = `
// Auto Generated. DO NOT EDIT!
syntax = "proto3";
package airbloc.{{.PackageName}};

import "google/protobuf/empty.proto";

{{range .Services}}{{.PrintService}}
{{end}}
{{range .Messages}}{{.PrintMessage}}
{{end}}
`
