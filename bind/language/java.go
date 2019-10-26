package language

import (
	"fmt"
	"regexp"

	"github.com/airbloc/solgen/bind/template"
	"github.com/airbloc/solgen/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

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
func bindTypeJava(kind abi.Type, structs map[string]*template.Struct) string {
	switch kind.T {
	case abi.TupleTy:
		return structs[kind.String()].Name
	case abi.ArrayTy, abi.SliceTy:
		return pluralizeJavaType(bindTypeJava(*kind.Elem, structs))
	default:
		return bindBasicTypeJava(kind)
	}
}

// bindTopicTypeJava converts a Solidity topic type to a Java one. It is almost the same
// funcionality as for simple types, but dynamic types get converted to hashes.
func bindTopicTypeJava(kind abi.Type, structs map[string]*template.Struct) string {
	bound := bindTypeJava(kind, structs)
	if bound == "String" || bound == "byte[]" {
		bound = "Hash"
	}
	return bound
}

// bindStructTypeJava converts a Solidity tuple type to a Java one and records the mapping
// in the given map.
// Notably, this function will resolve and record nested struct recursively.
func bindStructTypeJava(kind abi.Type, structs map[string]*template.Struct) string {
	switch kind.T {
	case abi.TupleTy:
		if s, exist := structs[kind.String()]; exist {
			return s.Name
		}
		var fields []*template.Field
		for i, elem := range kind.TupleElems {
			field := bindStructTypeJava(*elem, structs)
			fields = append(fields, &template.Field{Type: field, Name: utils.Decapitalise(kind.TupleRawNames[i]), SolKind: *elem})
		}
		name := fmt.Sprintf("Class%d", len(structs))
		structs[kind.String()] = &template.Struct{
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

// namedTypeJava converts some primitive data types to named variants that can
// be used as parts of method names.
func namedTypeJava(javaKind string, solKind abi.Type) string {
	switch javaKind {
	case "byte[]":
		return "Binary"
	case "boolean":
		return "Bool"
	default:
		parts := regexp.MustCompile(`(u)?int([0-9]*)(\[[0-9]*])?`).FindStringSubmatch(solKind.String())
		if len(parts) != 4 {
			return javaKind
		}
		switch parts[2] {
		case "8", "16", "32", "64":
			if parts[3] == "" {
				return utils.Capitalise(fmt.Sprintf("%sint%s", parts[1], parts[2]))
			}
			return utils.Capitalise(fmt.Sprintf("%sint%ss", parts[1], parts[2]))

		default:
			return javaKind
		}
	}
}
