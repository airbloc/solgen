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

// Package bind generates Ethereum tmplContract Go bindings.
//
// Detailed usage document and tutorial available on the go-ethereum Wiki page:
// https://github.com/ethereum/go-ethereum/wiki/Native-DApps:-Go-bindings-to-Ethereum-contracts
package bind

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/frostornge/solgen/deployments"
)

func GenerateBind(path string, deployments deployments.Deployments) error {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	} else {
		if !stat.IsDir() {
			return errors.New("is not directory")
		}
	}

	for contractName, contractAbi := range deployments {
		bindData, err := Bind(contractName, contractAbi, "adapter")
		if err != nil {
			return err
		}

		file := filepath.Join(path, contractName+".go")
		if err = ioutil.WriteFile(file, bindData, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

// Bind generates a Go wrapper around a tmplContract ABI. This wrapper isn't meant
// to be used as is in client code, but rather as an intermediate struct which
// enforces compile time type safety and naming convention opposed to having to
// manually maintain hard coded strings that break on runtime.
func Bind(name string, evmABI abi.ABI, pkg string) ([]byte, error) {
	log.SetFlags(log.Llongfile)

	// Process each individual tmplContract requested binding
	contracts := make(map[string]*tmplContract)

	abiByte, err := json.Marshal(evmABI)
	if err != nil {
		return nil, err
	}
	abistr := string(abiByte)

	// Strip any whitespace from the JSON ABI
	strippedABI := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, abistr)

	// Extract the call and transact methods; events; and sort them alphabetically
	var (
		calls     = make(map[string]*tmplMethod)
		transacts = make(map[string]*tmplMethod)
		events    = make(map[string]*tmplEvent)
	)
	for _, original := range evmABI.Methods {
		// Normalize the tmplMethod for capital cases and non-anonymous inputs/outputs
		normalized := original
		normalized.Name = capitalise(original.Name)

		normalized.Inputs = make([]abi.Argument, len(original.Inputs))
		copy(normalized.Inputs, original.Inputs)
		for j, input := range normalized.Inputs {
			if input.Name == "" {
				normalized.Inputs[j].Name = fmt.Sprintf("arg%d", j)
			}
		}
		normalized.Outputs = make([]abi.Argument, len(original.Outputs))
		copy(normalized.Outputs, original.Outputs)
		for j, output := range normalized.Outputs {
			if output.Name != "" {
				normalized.Outputs[j].Name = capitalise(output.Name)
			}
		}
		// Append the methods to the call or transact lists
		if original.Const {
			calls[original.Name] = &tmplMethod{Original: original, Normalized: normalized, Structured: structured(original.Outputs)}
		} else {
			transacts[original.Name] = &tmplMethod{Original: original, Normalized: normalized, Structured: structured(original.Outputs)}
		}
	}
	for _, original := range evmABI.Events {
		// Skip anonymous events as they don't support explicit filtering
		if original.Anonymous {
			continue
		}
		// Normalize the tmplEvent for capital cases and non-anonymous outputs
		normalized := original
		normalized.Name = capitalise(original.Name)

		normalized.Inputs = make([]abi.Argument, len(original.Inputs))
		copy(normalized.Inputs, original.Inputs)
		for j, input := range normalized.Inputs {
			// Indexed fields are input, non-indexed ones are outputs
			if input.Indexed {
				if input.Name == "" {
					normalized.Inputs[j].Name = fmt.Sprintf("arg%d", j)
				}
			}
		}
		// Append the tmplEvent to the accumulator list
		events[original.Name] = &tmplEvent{Original: original, Normalized: normalized}
	}
	contracts[name] = &tmplContract{
		Type:        capitalise(name),
		InputABI:    strings.Replace(strippedABI, "\"", "\\\"", -1),
		Constructor: evmABI.Constructor,
		Calls:       calls,
		Transacts:   transacts,
		Events:      events,
	}
	// Generate the tmplContract template data content and render it
	data := &tmplData{
		Package:   pkg,
		Contracts: contracts,
	}
	buffer := new(bytes.Buffer)

	funcs := map[string]interface{}{
		"bindtype":      bindType,
		"bindtopictype": bindTopicType,
		"namedtype":     func(string, abi.Type) string { panic("this shouldn't be needed") },
		"capitalise":    capitalise,
		"decapitalise":  decapitalise,
	}
	tmpl := template.Must(template.New("").Funcs(funcs).Parse(tmplSource))
	if err := tmpl.Execute(buffer, data); err != nil {
		return []byte{}, err
	}

	// For Go bindings pass the code through gofmt to clean it up
	code, err := format.Source(buffer.Bytes())
	if err != nil {
		return []byte{}, fmt.Errorf("%v\n%s", err, buffer)
	}
	return code, nil
}

// Helper function for the binding generators.
// It reads the unmatched characters after the inner type-match,
//  (since the inner type is a prefix of the total type declaration),
//  looks for valid arrays (possibly a dynamic one) wrapping the inner type,
//  and returns the sizes of these arrays.
//
// Returned array sizes are in the same order as solidity signatures; inner array size first.
// Array sizes may also be "", indicating a dynamic array.
func wrapArray(stringKind string, innerLen int, innerMapping string) (string, []string) {
	remainder := stringKind[innerLen:]
	//find all the sizes
	matches := regexp.MustCompile(`\[(\d*)]`).FindAllStringSubmatch(remainder, -1)
	parts := make([]string, 0, len(matches))
	for _, match := range matches {
		//get group 1 from the regex match
		parts = append(parts, match[1])
	}
	return innerMapping, parts
}

// Translates the array sizes to a Go-lang declaration of a (nested) array of the inner type.
// Simply returns the inner type if arraySizes is empty.
func arrayBinding(inner string, arraySizes []string) string {
	out := ""
	//prepend all array sizes, from outer (end arraySizes) to inner (start arraySizes)
	for i := len(arraySizes) - 1; i >= 0; i-- {
		out += "[" + arraySizes[i] + "]"
	}
	out += inner
	return out
}

// bindTypeGo converts a Solidity type to a Go one. Since there is no clear mapping
// from all Solidity types to Go ones (e.g. uint17), those that cannot be exactly
// mapped will use an upscaled type (e.g. *big.Int).
func bindType(kind abi.Type) string {
	stringKind := kind.String()
	innerLen, innerMapping := bindUnnestedType(stringKind)
	return arrayBinding(wrapArray(stringKind, innerLen, innerMapping))
}

// The inner function of bindTypeGo, this finds the inner type of stringKind.
// (Or just the type itself if it is not an array or slice)
// The length of the matched part is returned, with the translated type.
func bindUnnestedType(stringKind string) (int, string) {

	switch {
	case strings.HasPrefix(stringKind, "address"):
		return len("address"), "common.Address"

	case strings.HasPrefix(stringKind, "bytes"):
		parts := regexp.MustCompile(`bytes([0-9]*)`).FindStringSubmatch(stringKind)
		return len(parts[0]), fmt.Sprintf("[%s]byte", parts[1])

	case strings.HasPrefix(stringKind, "int") || strings.HasPrefix(stringKind, "uint"):
		parts := regexp.MustCompile(`(u)?int([0-9]*)`).FindStringSubmatch(stringKind)
		switch parts[2] {
		case "8", "16", "32", "64":
			return len(parts[0]), fmt.Sprintf("%sint%s", parts[1], parts[2])
		}
		return len(parts[0]), "*big.Int"

	case strings.HasPrefix(stringKind, "bool"):
		return len("bool"), "bool"

	case strings.HasPrefix(stringKind, "string"):
		return len("string"), "string"

	default:
		return len(stringKind), stringKind
	}
}

// bindType converts a Solidity topic type to a Go one. It is almost the same
// funcionality as for simple types, but dynamic types get converted to hashes.
func bindTopicType(kind abi.Type) string {
	bound := bindType(kind)
	if bound == "string" || bound == "[]byte" {
		bound = "common.Hash"
	}
	return bound
}

// capitalise makes a camel-case string which starts with an upper case character.
func capitalise(input string) string {
	for len(input) > 0 && input[0] == '_' {
		input = input[1:]
	}
	if len(input) == 0 {
		return ""
	}
	return toCamelCase(strings.ToUpper(input[:1]) + input[1:])
}

// decapitalise makes a camel-case string which starts with a lower case character.
func decapitalise(input string) string {
	for len(input) > 0 && input[0] == '_' {
		input = input[1:]
	}
	if len(input) == 0 {
		return ""
	}
	return toCamelCase(strings.ToLower(input[:1]) + input[1:])
}

// toCamelCase converts an under-score string to a camel-case string
func toCamelCase(input string) string {
	toupper := false

	result := ""
	for k, v := range input {
		switch {
		case k == 0:
			result = strings.ToUpper(string(input[0]))

		case toupper:
			result += strings.ToUpper(string(v))
			toupper = false

		case v == '_':
			toupper = true

		default:
			result += string(v)
		}
	}
	return result
}

// structured checks whether a list of ABI data types has enough information to
// operate through a proper Go struct or if flat returns are needed.
func structured(args abi.Arguments) bool {
	if len(args) < 2 {
		return false
	}
	exists := make(map[string]bool)
	for _, out := range args {
		// If the name is anonymous, we can't organize into a struct
		if out.Name == "" {
			return false
		}
		// If the field name is empty when normalized or collides (var, Var, _var, _Var),
		// we can't organize into a struct
		field := capitalise(out.Name)
		if field == "" || exists[field] {
			return false
		}
		exists[field] = true
	}
	return true
}
