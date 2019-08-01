package klaytn

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

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

func bindSimpleType(strKind string) string {
	innerLen, innerMapping := bindUnnestedType(strKind)
	return arrayBinding(wrapArray(strKind, innerLen, innerMapping))
}

// bindTypeGo converts a Solidity type to a Go one. Since there is no clear mapping
// from all Solidity types to Go ones (e.g. uint17), those that cannot be exactly
// mapped will use an upscaled type (e.g. *big.Int).
func bindType(kind abi.Type, typeOption option) string {
	if name, ok := typeOption[kind.String()]; ok {
		return name
	}

	if kind.T == abi.TupleTy {
		var builder strings.Builder

		builder.WriteString("struct{\n")
		for index := range kind.TupleElems {
			tupleName := capitalise(kind.TupleRawNames[index])
			tupleElem := bindType(*kind.TupleElems[index], typeOption)

			builder.WriteString(fmt.Sprintln(tupleName, tupleElem))
		}
		builder.WriteString("}")
		return builder.String()
	} else {
		return bindSimpleType(kind.String())
	}
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
func bindTopicType(kind abi.Type, typeOption option) string {
	bound := bindType(kind, typeOption)
	if bound == "string" || bound == "[]byte" {
		bound = "common.Hash"
	}
	return bound
}
