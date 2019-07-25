package proto

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type contract struct {
	PackageName string
	Services    []Service
	Messages    []message

	contractName string
	typeOptions  Options
}

const EmptyMessage = "google.protobuf.Empty"

func (c *contract) parseMessage(name string, comment string, args abi.Arguments) {
	msg := &message{
		Args:         make([]argument, len(args)),
		Comment:      comment,
		Name:         name,
		contractName: c.contractName,
		typeOptions:  c.typeOptions,
	}
	msg.parseArguments(args)
	c.Messages = append(c.Messages, *msg)
}

func (c *contract) parseContract(contractAbi abi.ABI) {
	service := Service{
		Comment: c.contractName,
		Name:    c.contractName,
		Methods: make([]method, len(contractAbi.Methods)),
	}

	methodIndex := 0
	for methodName, methodInfo := range contractAbi.Methods {
		inputMessage := fmt.Sprintf("Request%s", toUpperCase(methodName, 0))
		outputMessage := fmt.Sprintf("Response%s", toUpperCase(methodName, 0))

		if len(methodInfo.Inputs) == 0 {
			inputMessage = EmptyMessage
		} else {
			c.parseMessage(inputMessage, methodInfo.Sig(), methodInfo.Inputs)
		}

		if len(methodInfo.Outputs) == 0 {
			outputMessage = EmptyMessage
		} else {
			c.parseMessage(outputMessage, methodInfo.Sig(), methodInfo.Outputs)
		}

		service.Methods[methodIndex] = method{
			Name:   methodName,
			Input:  inputMessage,
			Output: outputMessage,
		}

		methodIndex += 1
	}

	c.PackageName = toLowerCase(c.contractName, 0)
	c.Services = []Service{service}
}
