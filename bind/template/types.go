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

package template

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
)

// Data is the data structure required to fill the binding template.
type Data struct {
	Package  string            // Name of the package to place the generated file in
	Imports  map[string]string // List of custom imports to push into this file
	Contract *Contract         // List of contracts to generate into this file
}

// Contract contains the data needed to generate an individual contract binding.
type Contract struct {
	Type        string // Type name of the main contract binding
	Address     string
	TxHash      string
	CreatedAt   string
	InputABI    string             // JSON ABI used as the input to generate the binding from
	Constructor abi.Method         // Contract constructor for deploy parametrization
	Calls       map[string]*Method // Contract calls that only read state data
	Transacts   map[string]*Method // Contract calls that write state data
	Events      map[string]*Event  // Contract events accessors
	Libraries   map[string]string  // Same as Data, but filtered to only keep what the contract needs
	Structs     map[string]*Struct // Contract struct type definitions
	Library     bool
}

// Method is a wrapper around an abi.Method that contains a few preprocessed
// and cached data fields.
type Method struct {
	Original   abi.Method // Original method as parsed by the abi package
	Normalized abi.Method // Normalized version of the parsed method (capitalized names, non-anonymous args/returns)
	Structured bool       // Whether the returns should be accumulated into a struct
}

// Event is a wrapper around an a
type Event struct {
	Original   abi.Event // Original event as parsed by the abi package
	Normalized abi.Event // Normalized version of the parsed fields
}

// Field is a wrapper around a struct field with binding language
// struct type definition and relative filed name.
type Field struct {
	Type    string   // Field type representation depends on target binding language
	Name    string   // Field name converted from the raw user-defined field name
	SolKind abi.Type // Raw abi type information
}

// Struct is a wrapper around an abi.tuple contains a auto-generated
// struct name.
type Struct struct {
	Name   string   // Auto-generated struct name(We can't obtain the raw struct name through abi)
	Fields []*Field // Struct fields definition depends on the binding language.
}
