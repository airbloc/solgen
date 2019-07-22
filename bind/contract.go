package bind

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
)

type contract struct {
	Type        string
	InputABI    string
	Constructor abi.Method
	Calls       methods
	Transacts   methods
	Events      events
}

func parseContract(name string, evmABI abi.ABI) *contract {
	inputABI, err := stripABI(evmABI)
	if err != nil {
		return nil
	}

	calls, transacts := parseMethods(evmABI.Methods)

	return &contract{
		Type:        capitalise(name),
		InputABI:    inputABI,
		Constructor: evmABI.Constructor,
		Calls:       calls,
		Transacts:   transacts,
		Events:      parseEvents(evmABI.Events),
	}
}
