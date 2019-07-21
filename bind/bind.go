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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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
		data := convertToData(contractName, contractAbi, "adapter")
		file := filepath.Join(path, contractName+".go")
		if err = RenderFile(file, data); err != nil {
			return err
		}
	}

	return nil
}

// convertToData generates a Go wrapper around a tmplContract ABI. This wrapper isn't meant
// to be used as is in client code, but rather as an intermediate struct which
// enforces compile time type safety and naming convention opposed to having to
// manually maintain hard coded strings that break on runtime.
func convertToData(name string, evmABI abi.ABI, pkg string) *tmplData {
	log.SetFlags(log.Llongfile)

	// Process each individual tmplContract requested binding
	contracts := make(map[string]*tmplContract)

	abiByte, err := json.Marshal(evmABI)
	if err != nil {
		return nil
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
	return &tmplData{
		Package:   pkg,
		Contracts: contracts,
	}
}
