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

// Package bind generates Ethereum contract Go bindings.
//
// Detailed usage document and tutorial available on the go-ethereum Wiki page:
// https://github.com/ethereum/go-ethereum/wiki/Native-DApps:-Go-bindings-to-Ethereum-contracts
package bind

import (
	"bytes"
	"fmt"
	"go/format"
	"path"
	"path/filepath"
	tmpl "text/template"

	"github.com/airbloc/solgen/bind/language"
	"github.com/airbloc/solgen/bind/platform"
	"github.com/airbloc/solgen/bind/template"
	"github.com/airbloc/solgen/deployment"
	"github.com/airbloc/solgen/utils"
)

type Mode string

const (
	Contract Mode = "contracts"
	Manager  Mode = "managers"
	//Wrapper  Mode = "wrappers"
)

var Modes = []Mode{
	Contract,
	Manager,
	//Wrapper,
}

func getInternalFuncs(mode Mode, lang language.Language) map[string]interface{} {
	switch mode {
	case Contract:
		return map[string]interface{}{
			// from lang package
			"bindtype":      language.BindType[lang],
			"bindtopictype": language.BindTopicType[lang],
			"namedtype":     language.NamedType[lang],

			// from utils package
			"formatmethod": utils.FormatMethod,
			"formatevent":  utils.FormatEvent,
			"capitalise":   utils.Capitalise,
			"decapitalise": utils.Decapitalise,
		}
	case Manager:
		return map[string]interface{}{
			// from lang package
			"bindtype":      language.BindType[lang],
			"bindtopictype": language.BindTopicType[lang],

			// from utils package
			"decapitalise": utils.Decapitalise,
			"toSnakeCase":  utils.ToSnakeCase,
		}
	default:
		return nil
	}
}

func Bind(name string, deployment deployment.Deployment, opt Option) (map[Mode][]byte, error) {
	contract, err := getContract(deployment, opt.Customs, opt.Language)
	if err != nil {
		return nil, err
	}
	contract.Type = utils.Capitalise(name)

	codes := make(map[Mode][]byte)
	for _, mode := range Modes {
		data := &template.Data{
			Imports:  platform.MergeImports(platform.Imports[opt.Platform], opt.Customs.Imports),
			Contract: contract,
			Package:  string(mode),
		}
		//if mode == Wrapper {
		//	data.Imports = platform.MergeImports(data.Imports, map[string]string{
		//		"contracts": "github.com/airbloc/airbloc-go/bind/contracts",
		//	})
		//}
		if mode == Manager {
			data.Imports = platform.ManagerImports(opt.Platform)
		}

		code, err := bind(mode, data, opt)
		if err != nil {
			return nil, err
		}
		codes[mode] = code
	}

	return codes, nil
}

func bind(
	mode Mode,
	data *template.Data,
	opt Option,
) ([]byte, error) {
	tmplPath, err := filepath.Abs(path.Join("./bind", "template", string(opt.Language), string(mode), "*"))
	if err != nil {
		return nil, err
	}

	buffer := new(bytes.Buffer)
	functions := getInternalFuncs(mode, opt.Language)
	t := tmpl.Must(tmpl.New(string(mode)).Funcs(functions).ParseGlob(tmplPath))
	if err := t.ExecuteTemplate(buffer, string(mode), data); err != nil {
		return nil, err
	}

	var code []byte

	switch opt.Language {
	case language.Go:
		code, err = format.Source(buffer.Bytes())
		if err != nil {
			return nil, fmt.Errorf("%v\n%s", err, buffer)
		}
	default:
		code = buffer.Bytes()
	}

	code = bytes.ReplaceAll(code, []byte("[8]byte"), []byte("types.ID"))
	code = bytes.ReplaceAll(code, []byte("[32]byte"), []byte("common.Hash"))
	code = bytes.ReplaceAll(code, []byte("[20]byte"), []byte("types.DataId"))

	return code, nil
}
