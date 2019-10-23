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
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// Lang is a target programming language selector to generate bindings for.
type Lang int
type Platform string

const (
	LangGo Lang = iota
	LangJava
	LangObjC

	PlatformEth  Platform = "ethereum"
	PlatformKlay Platform = "klaytn"
)

var imports = map[Platform]map[string]string{
	PlatformEth: {
		"platform":   "github.com/ethereum/go-ethereum",
		"abi":        "github.com/ethereum/go-ethereum/accounts/abi",
		"bind":       "github.com/ethereum/go-ethereum/accounts/abi/bind",
		"common":     "github.com/ethereum/go-ethereum/common",
		"chainTypes": "github.com/ethereum/go-ethereum/core/types",
		"event":      "github.com/ethereum/go-ethereum/event",
		"types":      "github.com/airbloc/airbloc-go/shared/types",
		"blockchain": "github.com/airbloc/airbloc-go/shared/blockchain",
	},
	PlatformKlay: {
		"platform":   "github.com/klaytn/klaytn",
		"abi":        "github.com/klaytn/klaytn/accounts/abi",
		"bind":       "github.com/klaytn/klaytn/accounts/abi/bind",
		"common":     "github.com/klaytn/klaytn/common",
		"chainTypes": "github.com/klaytn/klaytn/blockchain/types",
		"event":      "github.com/klaytn/klaytn/event",
		"types":      "github.com/airbloc/airbloc-go/shared/types",
		"blockchain": "github.com/airbloc/airbloc-go/shared/blockchain",
	},
}

func apply(srcs ...map[string]string) map[string]string {
	o := make(map[string]string)
	for _, src := range srcs {
		for k, v := range src {
			o[k] = v
		}
	}
	return o
}

type Customs struct {
	Structs map[string]string
	Imports map[string]string
	Methods map[string]bool
}

func templateData(
	name, rawABI, pkg string,
	customs Customs, plat Platform, lang Lang,
) (*tmplData, error) {
	// Parse the actual ABI to generate the binding for
	evmABI, err := abi.JSON(strings.NewReader(rawABI))
	if err != nil {
		return nil, err
	}
	// Strip any whitespace from the JSON ABI
	strippedABI := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, rawABI)

	// Extract the call and transact methods; events, struct definitions; and sort them alphabetically
	var (
		calls     = make(map[string]*tmplMethod)
		transacts = make(map[string]*tmplMethod)
		events    = make(map[string]*tmplEvent)
		structs   = make(map[string]*tmplStruct)
	)
	for methodName, original := range evmABI.Methods {
		if ok, exists := customs.Methods[methodName]; !exists || !ok {
			continue
		}

		// Normalize the method for capital cases and non-anonymous inputs/outputs
		normalized := original
		normalized.Name = methodNormalizer[lang](original.Name)

		normalized.Inputs = make([]abi.Argument, len(original.Inputs))
		copy(normalized.Inputs, original.Inputs)
		for j, input := range normalized.Inputs {
			if input.Name == "" {
				normalized.Inputs[j].Name = fmt.Sprintf("arg%d", j)
			}
			if _, exist := structs[input.Type.String()]; input.Type.T == abi.TupleTy && !exist {
				bindStructType[lang](input.Type, structs)
			}
		}
		normalized.Outputs = make([]abi.Argument, len(original.Outputs))
		copy(normalized.Outputs, original.Outputs)
		for j, output := range normalized.Outputs {
			if output.Name != "" {
				normalized.Outputs[j].Name = capitalise(output.Name)
			}
			if _, exist := structs[output.Type.String()]; output.Type.T == abi.TupleTy && !exist {
				bindStructType[lang](output.Type, structs)
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
		// Normalize the event for capital cases and non-anonymous outputs
		normalized := original
		normalized.Name = methodNormalizer[lang](original.Name)

		normalized.Inputs = make([]abi.Argument, len(original.Inputs))
		copy(normalized.Inputs, original.Inputs)
		for j, input := range normalized.Inputs {
			// Indexed fields are input, non-indexed ones are outputs
			if input.Indexed {
				if input.Name == "" {
					normalized.Inputs[j].Name = fmt.Sprintf("arg%d", j)
				}
				if _, exist := structs[input.Type.String()]; input.Type.T == abi.TupleTy && !exist {
					bindStructType[lang](input.Type, structs)
				}
			}
		}
		// Append the event to the accumulator list
		events[original.Name] = &tmplEvent{Original: original, Normalized: normalized}
	}

	// There is no easy way to pass arbitrary java objects to the Go side.
	if len(structs) > 0 && lang == LangJava {
		return nil, errors.New("java binding for tuple arguments is not supported yet")
	}

	for exp, strt := range structs {
		if n, ok := customs.Structs[exp]; ok {
			strt.Name = n
		}
	}

	contract := &tmplContract{
		Type:        capitalise(name),
		InputABI:    strings.Replace(strippedABI, "\"", "\\\"", -1),
		Constructor: evmABI.Constructor,
		Calls:       calls,
		Transacts:   transacts,
		Events:      events,
		Structs:     structs,
	}

	// Generate the contract template data content and render it
	data := &tmplData{
		Package:  pkg,
		Imports:  apply(imports[plat], customs.Imports),
		Contract: contract,
	}

	return data, nil
}

// Bind generates a Go wrapper around a contract ABI. This wrapper isn't meant
// to be used as is in client code, but rather as an intermediate struct which
// enforces compile time type safety and naming convention opposed to having to
// manually maintain hard coded strings that break on runtime.
func BindContract(
	bindFile *os.File,
	contractName, contractABI, pkg string,
	customs Customs, plat Platform, lang Lang,
) error {
	data, err := templateData(contractName, contractABI, pkg, customs, plat, lang)
	if err != nil {
		return err
	}

	bindCode, err := RenderBind(data, lang)
	if err != nil {
		return err
	}
	bindCode = strings.ReplaceAll(bindCode, "[8]byte", "types.ID")
	bindCode = strings.ReplaceAll(bindCode, "[32]byte", "common.Hash")
	bindCode = strings.ReplaceAll(bindCode, "[20]byte", "types.DataId")

	_, err = io.Copy(bindFile, strings.NewReader(bindCode))
	if err != nil {
		return err
	}
	return nil
}

func BindWrapper(
	wrapFile *os.File,
	contractName, contractABI, pkg string,
	customs Customs, plat Platform, lang Lang,
) error {
	if customs.Imports == nil {
		customs.Imports = map[string]string{
			"contracts": "github.com/airbloc/airbloc-sdk-go/bind/contracts",
		}
	}

	data, err := templateData(contractName, contractABI, pkg, customs, plat, lang)
	if err != nil {
		return err
	}

	wrapCode, err := RenderWrap(data, lang)
	if err != nil {
		return err
	}
	wrapCode = strings.ReplaceAll(wrapCode, "[8]byte", "types.ID")
	wrapCode = strings.ReplaceAll(wrapCode, "[32]byte", "common.Hash")
	wrapCode = strings.ReplaceAll(wrapCode, "[20]byte", "types.DataId")

	_, err = io.Copy(wrapFile, strings.NewReader(wrapCode))
	if err != nil {
		return err
	}

	return nil
}

// bindType is a set of type binders that convert Solidity types to some supported
// programming language types.
var bindType = map[Lang]func(kind abi.Type, structs map[string]*tmplStruct) string{
	LangGo:   bindTypeGo,
	LangJava: bindTypeJava,
}

// bindBasicTypeGo converts basic solidity types(except array, slice and tuple) to Go one.
func bindBasicTypeGo(kind abi.Type) string {
	switch kind.T {
	case abi.AddressTy:
		return "common.Address"
	case abi.IntTy, abi.UintTy:
		parts := regexp.MustCompile(`(u)?int([0-9]*)`).FindStringSubmatch(kind.String())
		switch parts[2] {
		case "8", "16", "32", "64":
			return fmt.Sprintf("%sint%s", parts[1], parts[2])
		}
		return "*big.Int"
	case abi.FixedBytesTy:
		return fmt.Sprintf("[%d]byte", kind.Size)
	case abi.BytesTy:
		return "[]byte"
	case abi.FunctionTy:
		return "[24]byte"
	default:
		// string, bool types
		return kind.String()
	}
}

// bindTypeGo converts solidity types to Go ones. Since there is no clear mapping
// from all Solidity types to Go ones (e.g. uint17), those that cannot be exactly
// mapped will use an upscaled type (e.g. BigDecimal).
func bindTypeGo(kind abi.Type, structs map[string]*tmplStruct) string {
	switch kind.T {
	case abi.TupleTy:
		return structs[kind.String()].Name
	case abi.ArrayTy:
		return fmt.Sprintf("[%d]", kind.Size) + bindTypeGo(*kind.Elem, structs)
	case abi.SliceTy:
		return "[]" + bindTypeGo(*kind.Elem, structs)
	default:
		return bindBasicTypeGo(kind)
	}
}

// bindBasicTypeJava converts basic solidity types(except array, slice and tuple) to Java one.
func bindBasicTypeJava(kind abi.Type) string {
	switch kind.T {
	case abi.AddressTy:
		return "Address"
	case abi.IntTy, abi.UintTy:
		// Note that uint and int (without digits) are also matched,
		// these are size 256, and will translate to BigInt (the default).
		parts := regexp.MustCompile(`(u)?int([0-9]*)`).FindStringSubmatch(kind.String())
		if len(parts) != 3 {
			return kind.String()
		}
		// All unsigned integers should be translated to BigInt since gomobile doesn't
		// support them.
		if parts[1] == "u" {
			return "BigInt"
		}

		namedSize := map[string]string{
			"8":  "byte",
			"16": "short",
			"32": "int",
			"64": "long",
		}[parts[2]]

		// default to BigInt
		if namedSize == "" {
			namedSize = "BigInt"
		}
		return namedSize
	case abi.FixedBytesTy, abi.BytesTy:
		return "byte[]"
	case abi.BoolTy:
		return "boolean"
	case abi.StringTy:
		return "String"
	case abi.FunctionTy:
		return "byte[24]"
	default:
		return kind.String()
	}
}

// pluralizeJavaType explicitly converts multidimensional types to predefined
// type in go side.
func pluralizeJavaType(typ string) string {
	switch typ {
	case "boolean":
		return "Bools"
	case "String":
		return "Strings"
	case "Address":
		return "Addresses"
	case "byte[]":
		return "Binaries"
	case "BigInt":
		return "BigInts"
	}
	return typ + "[]"
}

// bindTypeJava converts a Solidity type to a Java one. Since there is no clear mapping
// from all Solidity types to Java ones (e.g. uint17), those that cannot be exactly
// mapped will use an upscaled type (e.g. BigDecimal).
func bindTypeJava(kind abi.Type, structs map[string]*tmplStruct) string {
	switch kind.T {
	case abi.TupleTy:
		return structs[kind.String()].Name
	case abi.ArrayTy, abi.SliceTy:
		return pluralizeJavaType(bindTypeJava(*kind.Elem, structs))
	default:
		return bindBasicTypeJava(kind)
	}
}

// bindTopicType is a set of type binders that convert Solidity types to some
// supported programming language topic types.
var bindTopicType = map[Lang]func(kind abi.Type, structs map[string]*tmplStruct) string{
	LangGo:   bindTopicTypeGo,
	LangJava: bindTopicTypeJava,
}

// bindTopicTypeGo converts a Solidity topic type to a Go one. It is almost the same
// funcionality as for simple types, but dynamic types get converted to hashes.
func bindTopicTypeGo(kind abi.Type, structs map[string]*tmplStruct) string {
	bound := bindTypeGo(kind, structs)
	if bound == "string" || bound == "[]byte" {
		bound = "common.Hash"
	}
	return bound
}

// bindTopicTypeJava converts a Solidity topic type to a Java one. It is almost the same
// funcionality as for simple types, but dynamic types get converted to hashes.
func bindTopicTypeJava(kind abi.Type, structs map[string]*tmplStruct) string {
	bound := bindTypeJava(kind, structs)
	if bound == "String" || bound == "byte[]" {
		bound = "Hash"
	}
	return bound
}

// bindStructType is a set of type binders that convert Solidity tuple types to some supported
// programming language struct definition.
var bindStructType = map[Lang]func(kind abi.Type, structs map[string]*tmplStruct) string{
	LangGo:   bindStructTypeGo,
	LangJava: bindStructTypeJava,
}

// bindStructTypeGo converts a Solidity tuple type to a Go one and records the mapping
// in the given map.
// Notably, this function will resolve and record nested struct recursively.
func bindStructTypeGo(kind abi.Type, structs map[string]*tmplStruct) string {
	switch kind.T {
	case abi.TupleTy:
		if s, exist := structs[kind.String()]; exist {
			return s.Name
		}
		var fields []*tmplField
		for i, elem := range kind.TupleElems {
			field := bindStructTypeGo(*elem, structs)
			fields = append(fields, &tmplField{Type: field, Name: capitalise(kind.TupleRawNames[i]), SolKind: *elem})
		}
		name := fmt.Sprintf("Struct%d", len(structs))
		structs[kind.String()] = &tmplStruct{
			Name:   name,
			Fields: fields,
		}
		return name
	case abi.ArrayTy:
		return fmt.Sprintf("[%d]", kind.Size) + bindStructTypeGo(*kind.Elem, structs)
	case abi.SliceTy:
		return "[]" + bindStructTypeGo(*kind.Elem, structs)
	default:
		return bindBasicTypeGo(kind)
	}
}

// bindStructTypeJava converts a Solidity tuple type to a Java one and records the mapping
// in the given map.
// Notably, this function will resolve and record nested struct recursively.
func bindStructTypeJava(kind abi.Type, structs map[string]*tmplStruct) string {
	switch kind.T {
	case abi.TupleTy:
		if s, exist := structs[kind.String()]; exist {
			return s.Name
		}
		var fields []*tmplField
		for i, elem := range kind.TupleElems {
			field := bindStructTypeJava(*elem, structs)
			fields = append(fields, &tmplField{Type: field, Name: decapitalise(kind.TupleRawNames[i]), SolKind: *elem})
		}
		name := fmt.Sprintf("Class%d", len(structs))
		structs[kind.String()] = &tmplStruct{
			Name:   name,
			Fields: fields,
		}
		return name
	case abi.ArrayTy, abi.SliceTy:
		return pluralizeJavaType(bindStructTypeJava(*kind.Elem, structs))
	default:
		return bindBasicTypeJava(kind)
	}
}

// namedType is a set of functions that transform language specific types to
// named versions that my be used inside method names.
var namedType = map[Lang]func(string, abi.Type) string{
	LangGo:   func(string, abi.Type) string { panic("this shouldn't be needed") },
	LangJava: namedTypeJava,
}

// namedTypeJava converts some primitive data types to named variants that can
// be used as parts of method names.
func namedTypeJava(javaKind string, solKind abi.Type) string {
	switch javaKind {
	case "byte[]":
		return "Binary"
	case "boolean":
		return "Bool"
	default:
		parts := regexp.MustCompile(`(u)?int([0-9]*)(\[[0-9]*\])?`).FindStringSubmatch(solKind.String())
		if len(parts) != 4 {
			return javaKind
		}
		switch parts[2] {
		case "8", "16", "32", "64":
			if parts[3] == "" {
				return capitalise(fmt.Sprintf("%sint%s", parts[1], parts[2]))
			}
			return capitalise(fmt.Sprintf("%sint%ss", parts[1], parts[2]))

		default:
			return javaKind
		}
	}
}

// methodNormalizer is a name transformer that modifies Solidity method names to
// conform to target language naming concentions.
var methodNormalizer = map[Lang]func(string) string{
	LangGo:   abi.ToCamelCase,
	LangJava: decapitalise,
}

// capitalise makes a camel-case string which starts with an upper case character.
func capitalise(input string) string {
	return abi.ToCamelCase(input)
}

// decapitalise makes a camel-case string which starts with a lower case character.
func decapitalise(input string) string {
	if len(input) == 0 {
		return input
	}

	goForm := abi.ToCamelCase(input)
	return strings.ToLower(goForm[:1]) + goForm[1:]
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

// resolveArgName converts a raw argument representation into a user friendly format.
func resolveArgName(arg abi.Argument, structs map[string]*tmplStruct) string {
	var (
		prefix   string
		embedded string
		typ      = &arg.Type
	)
loop:
	for {
		switch typ.T {
		case abi.SliceTy:
			prefix += "[]"
		case abi.ArrayTy:
			prefix += fmt.Sprintf("[%d]", typ.Size)
		default:
			embedded = typ.String()
			break loop
		}
		typ = typ.Elem
	}
	if s, exist := structs[embedded]; exist {
		return prefix + s.Name
	} else {
		return arg.Type.String()
	}
}

// formatMethod transforms raw method representation into a user friendly one.
func formatMethod(method abi.Method, structs map[string]*tmplStruct) string {
	inputs := make([]string, len(method.Inputs))
	for i, input := range method.Inputs {
		inputs[i] = fmt.Sprintf("%v %v", resolveArgName(input, structs), input.Name)
	}
	outputs := make([]string, len(method.Outputs))
	for i, output := range method.Outputs {
		outputs[i] = resolveArgName(output, structs)
		if len(output.Name) > 0 {
			outputs[i] += fmt.Sprintf(" %v", output.Name)
		}
	}
	constant := ""
	if method.Const {
		constant = "constant "
	}
	return fmt.Sprintf("function %v(%v) %sreturns(%v)", method.RawName, strings.Join(inputs, ", "), constant, strings.Join(outputs, ", "))
}

// formatEvent transforms raw event representation into a user friendly one.
func formatEvent(event abi.Event, structs map[string]*tmplStruct) string {
	inputs := make([]string, len(event.Inputs))
	for i, input := range event.Inputs {
		if input.Indexed {
			inputs[i] = fmt.Sprintf("%v indexed %v", resolveArgName(input, structs), input.Name)
		} else {
			inputs[i] = fmt.Sprintf("%v %v", resolveArgName(input, structs), input.Name)
		}
	}
	return fmt.Sprintf("event %v(%v)", event.RawName, strings.Join(inputs, ", "))
}
