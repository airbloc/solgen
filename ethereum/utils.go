package ethereum

import (
	"encoding/json"
	"strings"
	"unicode"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func stripABI(evmABI abi.ABI) (string, error) {
	abiByte, err := json.Marshal(evmABI)
	if err != nil {
		return "", nil
	}
	abistr := string(abiByte)

	// Strip any whitespace from the JSON ABI
	strippedABI := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, abistr)

	return strings.Replace(strippedABI, "\"", "\\\"", -1), nil
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
	if len(args) <= 2 {
		if len(args) > 0 {
			return args[0].Type.T == abi.TupleTy
		}
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
