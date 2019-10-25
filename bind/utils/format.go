package utils

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/frostornge/solgen/bind/template"
)

// structured checks whether a list of ABI data types has enough information to
// operate through a proper Go struct or if flat returns are needed.
func Structured(args abi.Arguments) bool {
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
		field := Capitalise(out.Name)
		if field == "" || exists[field] {
			return false
		}
		exists[field] = true
	}
	return true
}

// resolveArgName converts a raw argument representation into a user friendly format.
func resolveArgName(arg abi.Argument, structs map[string]*template.Struct) string {
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
func FormatMethod(method abi.Method, structs map[string]*template.Struct) string {
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
func FormatEvent(event abi.Event, structs map[string]*template.Struct) string {
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
