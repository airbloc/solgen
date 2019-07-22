package bind

import (
	"fmt"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type method struct {
	Original   abi.Method
	Normalized abi.Method
	Structured bool
}

/*
if true
	{{range .Normalized.Inputs}} {{.Name}} {{bindtype .Type}} {{end}}
else
	{{range .Normalized.Inputs}}, {{.Name}}{{end}}
*/
func (mtd method) InputArgs(withType bool) string {
	var builder strings.Builder

	for _, input := range mtd.Normalized.Inputs {
		builder.WriteString(input.Name)
		if withType {
			builder.WriteString(" " + bindType(input.Type))
		}
		builder.WriteString(",")
	}

	argStr := builder.String()
	if len(argStr) == 0 {
		return ""
	}
	return argStr[:len(argStr)-1] // remove comma
}

/*
{{if .Structured}}
	struct{{{range .Normalized.Outputs}} {{.Name}} {{bindtype .Type}}; {{end}}},
{{else}}
	{{range .Normalized.Outputs}}{{bindtype .Type}}, {{end}}
{{end}}
*/
func (mtd method) OutputArgs() string {
	var builder strings.Builder

	for _, output := range mtd.Normalized.Outputs {
		builder.WriteString(bindType(output.Type) + ",")
	}

	argStr := builder.String()
	if len(argStr) == 0 {
		return ""
	}
	return argStr[:len(argStr)-1]
}

type methods map[string]*method

func parseMethods(evmMethods map[string]abi.Method) (methods, methods) {
	calls := make(methods)
	transacts := make(methods)

	for _, original := range evmMethods {
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

		if len(original.Outputs) < 2 {
			normalized.Outputs = make([]abi.Argument, len(original.Outputs))
			copy(normalized.Outputs, original.Outputs)
			for j, output := range normalized.Outputs {
				if output.Name != "" {
					normalized.Outputs[j].Name = capitalise(output.Name)
				}
			}
		} else {
			normalized.Outputs = []abi.Argument{}

			var args []abi.ArgumentMarshaling
			for _, output := range original.Outputs {
				if output.Name != "" {
					args = append(args, abi.ArgumentMarshaling{
						Type: output.Type.String(),
						Name: capitalise(output.Name),
					})
				}
			}

			outputType, err := abi.NewType("tuple", args)
			if err != nil {
				log.Println(err)
				return nil, nil
			}

			normalized.Outputs = append(normalized.Outputs, abi.Argument{Type: outputType})
		}

		// Append the methods to the call or transact lists
		if original.Const {
			calls[original.Name] = &method{
				Original:   original,
				Normalized: normalized,
				Structured: structured(original.Outputs),
			}
		} else {
			transacts[original.Name] = &method{
				Original:   original,
				Normalized: normalized,
				Structured: structured(original.Outputs),
			}
		}
	}

	return calls, transacts
}
