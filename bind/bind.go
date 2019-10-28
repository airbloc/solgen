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
	"strings"
	tmpl "text/template"

	"github.com/airbloc/solgen/bind/language"
	"github.com/airbloc/solgen/bind/platform"
	"github.com/airbloc/solgen/bind/template"
	"github.com/airbloc/solgen/utils"
)

type Mode string

const (
	Contract Mode = "contract"
	Manager  Mode = "manager"
	Wrapper  Mode = "wrapper"
)

var Modes = []Mode{
	Contract,
	Manager,
	Wrapper,
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
	case Wrapper:
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
			"toSnakeCase":  utils.ToSnakeCase,
		}
	default:
		return nil
	}
}

func Bind(name string, opt Option) (map[Mode]string, error) {
	contract, err := getContract(opt.ABI, opt.Customs, opt.Language)
	if err != nil {
		return nil, err
	}
	contract.Type = utils.Capitalise(name)

	data := &template.Data{
		Imports:  platform.MergeImports(platform.Imports[opt.Platform], opt.Customs.Imports),
		Contract: contract,
	}

	codes := make(map[Mode]string)
	for _, mode := range Modes {
		data.Package = string(mode)

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
) (string, error) {
	tmplPath, err := filepath.Abs(path.Join("./bind", "template", string(opt.Language), string(mode), "*"))
	if err != nil {
		return "", err
	}

	buffer := new(bytes.Buffer)
	functions := getInternalFuncs(mode, opt.Language)
	t := tmpl.Must(tmpl.New(string(mode)).Funcs(functions).ParseGlob(tmplPath))
	if err := t.ExecuteTemplate(buffer, string(mode), data); err != nil {
		return "", err
	}

	var code string

	switch opt.Language {
	case language.Go:
		codeBytes, err := format.Source(buffer.Bytes())
		if err != nil {
			return "", fmt.Errorf("%v\n%s", err, buffer)
		}
		code = string(codeBytes)
	default:
		code = buffer.String()
	}

	code = strings.ReplaceAll(code, "[8]byte", "types.ID")
	code = strings.ReplaceAll(code, "[32]byte", "common.Hash")
	code = strings.ReplaceAll(code, "[20]byte", "types.DataId")

	return code, nil
}
