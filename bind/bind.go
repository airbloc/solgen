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
	"github.com/airbloc/solgen/bind/utils"
)

type Mode string

const (
	Contract Mode = "contract"
	Manager  Mode = "manager"
	Wrapper  Mode = "wrapper"
)

func getInternalFuncs(mode Mode, lang language.Lang) map[string]interface{} {
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
		return map[string]interface{}{}
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

func Bind(
	mode Mode,
	name, abi string,
	customs Customs,
	plat platform.Platform,
	lang language.Lang,
) (string, error) {
	contract, err := getContract(abi, customs, lang)
	if err != nil {
		return "", err
	}
	contract.Type = utils.Capitalise(name)

	data := &template.Data{
		Package:  string(mode),
		Imports:  platform.MergeImports(platform.Imports[plat], customs.Imports),
		Contract: contract,
	}

	tmplPath, err := filepath.Abs(path.Join("./bind", "template", string(lang), string(mode), "*"))
	if err != nil {
		return "", err
	}

	buffer := new(bytes.Buffer)
	funcs := getInternalFuncs(mode, lang)
	t := tmpl.Must(tmpl.New(string(mode)).Funcs(funcs).ParseGlob(tmplPath))
	if err := t.ExecuteTemplate(buffer, string(mode), data); err != nil {
		return "", err
	}

	var code string

	switch lang {
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
