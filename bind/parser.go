package bind

import (
	"bytes"
	"errors"
	"fmt"
	"unicode"

	"github.com/airbloc/solgen/bind/language"
	"github.com/airbloc/solgen/bind/template"
	"github.com/airbloc/solgen/deployment"
	"github.com/airbloc/solgen/utils"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func parseContract(deployment deployment.Deployment, customs Customs, lang language.Language) (*template.Contract, error) {
	evmABI := deployment.EvmABI

	// Extract the call and transact methods; events, struct definitions; and sort them alphabetically
	var (
		calls     = make(map[string]*template.Method)
		transacts = make(map[string]*template.Method)
		events    = make(map[string]*template.Event)
		structs   = make(map[string]*template.Struct)
	)
	for methodName, original := range evmABI.Methods {
		if ok, exists := customs.Methods[methodName]; !exists || !ok {
			continue
		}

		// Normalize the method for capital cases and non-anonymous inputs/outputs
		normalized := original
		normalized.Name = language.MethodNormalizer[lang](original.Name)

		normalized.Inputs = make([]abi.Argument, len(original.Inputs))
		copy(normalized.Inputs, original.Inputs)
		for j, input := range normalized.Inputs {
			if input.Name == "" {
				normalized.Inputs[j].Name = fmt.Sprintf("arg%d", j)
			}
			if _, exist := structs[input.Type.String()]; input.Type.T == abi.TupleTy && !exist {
				language.BindStructType[lang](input.Type, structs)
			}
		}
		normalized.Outputs = make([]abi.Argument, len(original.Outputs))
		copy(normalized.Outputs, original.Outputs)
		for j, output := range normalized.Outputs {
			if output.Name != "" {
				normalized.Outputs[j].Name = utils.Capitalise(output.Name)
			}
			if _, exist := structs[output.Type.String()]; output.Type.T == abi.TupleTy && !exist {
				language.BindStructType[lang](output.Type, structs)
			}
		}
		// Append the methods to the call or transact lists
		if original.Const {
			calls[original.Name] = &template.Method{
				Original:   original,
				Normalized: normalized,
				Structured: utils.Structured(original.Outputs),
			}
		} else {
			transacts[original.Name] = &template.Method{
				Original:   original,
				Normalized: normalized,
				Structured: utils.Structured(original.Outputs),
			}
		}
	}
	for _, original := range evmABI.Events {
		// Skip anonymous events as they don't support explicit filtering
		if original.Anonymous {
			continue
		}
		// Normalize the event for capital cases and non-anonymous outputs
		normalized := original
		normalized.Name = language.MethodNormalizer[lang](original.Name)

		normalized.Inputs = make([]abi.Argument, len(original.Inputs))
		copy(normalized.Inputs, original.Inputs)
		for j, input := range normalized.Inputs {
			// Indexed fields are input, non-indexed ones are outputs
			if input.Indexed {
				if input.Name == "" {
					normalized.Inputs[j].Name = fmt.Sprintf("arg%d", j)
				}
				if _, exist := structs[input.Type.String()]; input.Type.T == abi.TupleTy && !exist {
					language.BindStructType[lang](input.Type, structs)
				}
			}
		}
		// Append the event to the accumulator list
		events[original.Name] = &template.Event{Original: original, Normalized: normalized}
	}

	// There is no easy way to pass arbitrary java objects to the Go side.
	if len(structs) > 0 && lang == language.Java {
		return nil, errors.New("java binding for tuple arguments is not supported yet")
	}

	for exp, strt := range structs {
		if n, ok := customs.Structs[exp]; ok {
			strt.Name = n
		}
	}

	contract := &template.Contract{
		Address:     deployment.Address.Hex(),
		TxHash:      deployment.TxHash.Hex(),
		CreatedAt:   common.BytesToHash(deployment.CreatedAt.Bytes()).Hex(),
		Constructor: evmABI.Constructor,
		Calls:       calls,
		Transacts:   transacts,
		Events:      events,
		Structs:     structs,
	}

	return contract, nil
}

func getContract(
	deployment deployment.Deployment,
	customs Customs,
	lang language.Language,
) (*template.Contract, error) {
	// ABI
	strippedABI := bytes.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, deployment.RawABI)
	strippedABI = bytes.ReplaceAll(strippedABI, []byte("\""), []byte("\\\""))

	contract, err := parseContract(deployment, customs, lang)
	if err != nil {
		return nil, err
	}
	contract.InputABI = string(strippedABI)

	return contract, nil
}
