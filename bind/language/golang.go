package language

import (
	"fmt"
	"regexp"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/frostornge/solgen/bind/template"
	"github.com/frostornge/solgen/bind/utils"
)

// bindTypeGo converts solidity types to Go ones. Since there is no clear mapping
// from all Solidity types to Go ones (e.g. uint17), those that cannot be exactly
// mapped will use an upscaled type (e.g. BigDecimal).
func bindTypeGo(kind abi.Type, structs map[string]*template.Struct) string {
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

// bindTopicTypeGo converts a Solidity topic type to a Go one. It is almost the same
// funcionality as for simple types, but dynamic types get converted to hashes.
func bindTopicTypeGo(kind abi.Type, structs map[string]*template.Struct) string {
	bound := bindTypeGo(kind, structs)
	if bound == "string" || bound == "[]byte" {
		bound = "common.Hash"
	}
	return bound
}

// bindStructTypeGo converts a Solidity tuple type to a Go one and records the mapping
// in the given map.
// Notably, this function will resolve and record nested struct recursively.
func bindStructTypeGo(kind abi.Type, structs map[string]*template.Struct) string {
	switch kind.T {
	case abi.TupleTy:
		if s, exist := structs[kind.String()]; exist {
			return s.Name
		}
		var fields []*template.Field
		for i, elem := range kind.TupleElems {
			field := bindTopicTypeGo(*elem, structs)
			fields = append(fields, &template.Field{Type: field, Name: utils.Capitalise(kind.TupleRawNames[i]), SolKind: *elem})
		}
		name := fmt.Sprintf("Struct%d", len(structs))
		structs[kind.String()] = &template.Struct{
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
