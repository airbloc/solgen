package ethereum

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/frostornge/solgen/deployment"
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

func parseContract(abi deployment.Deployment, contractName string, typeOption option) (c *contract, err error) {
	inputABI, err := stripABI(abi.RawABI)
	if err != nil {
		return nil, err
	}

	c = &contract{
		Type:        capitalise(contractName),
		TypeOption:  typeOption,
		InputABI:    inputABI,
		Constructor: abi.Constructor,
		Calls:       make(map[string]*method),
		Transacts:   make(map[string]*method),
		Events:      parseEvents(abi.ABI),
	}

	c.Calls, c.Transacts, err = parseMethods(abi.ABI, typeOption)
	if err != nil {
		return nil, err
	}
	return
}
