// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package bind

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"os"
	"text/template"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type tmplMethod struct {
	Original   abi.Method
	Normalized abi.Method
	Structured bool
}

type tmplEvent struct {
	Original   abi.Event
	Normalized abi.Event
}

type tmplContract struct {
	Type        string
	InputABI    string
	Constructor abi.Method
	Calls       map[string]*tmplMethod
	Transacts   map[string]*tmplMethod
	Events      map[string]*tmplEvent
}

// tmplData is the data structure required to fill the binding template.
type tmplData struct {
	Package   string                   // Name of the package to place the generated file in
	Contracts map[string]*tmplContract // List of contracts to generate into this file
}

const templatePath = "./bind/templates/*"

func render(writer io.Writer, data *tmplData) error {
	funcs := map[string]interface{}{
		"bindtype":      bindType,
		"bindtopictype": bindTopicType,
		"capitalise":    capitalise,
		"decapitalise":  decapitalise,
	}

	tmpl := template.Must(template.New("Bind").Funcs(funcs).ParseGlob(templatePath))
	if err := tmpl.Execute(writer, data); err != nil {
		return err
	}
	return nil
}

func Render(data *tmplData) ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := render(buffer, data); err != nil {
		return nil, err
	}

	// For Go bindings pass the code through gofmt to clean it up
	code, err := format.Source(buffer.Bytes())
	if err != nil {
		return []byte{}, fmt.Errorf("%v\n%s", err, buffer)
	}
	return code, nil
}

func RenderFile(path string, data *tmplData) error {
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
	return render(out, data)
}
