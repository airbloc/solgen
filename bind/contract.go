package bind

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
)

type contract struct {
	Type        string
	TypeOption  option
	InputABI    string
	Constructor abi.Method
	Calls       map[string]*method
	Transacts   map[string]*method
	Events      map[string]*event
}

func parseContract(evmABI abi.ABI, contractName string, typeOption option) (*contract, error) {
	inputABI, err := stripABI(evmABI)
	if err != nil {
		return nil, err
	}

	contract := &contract{
		Type:        capitalise(contractName),
		TypeOption:  typeOption,
		InputABI:    inputABI,
		Constructor: evmABI.Constructor,
		Calls:       make(map[string]*method),
		Transacts:   make(map[string]*method),
		Events:      parseEvents(evmABI),
	}

	contract.Calls, contract.Transacts, err = parseMethods(evmABI, typeOption)
	if err != nil {
		return nil, err
	}

	return contract, nil
}
