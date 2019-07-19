package proto

import (
	"fmt"

	"github.com/frostornge/solgen/deployments"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type Contract struct {
	PackageName string
	Services    []Service
	Messages    []Message
}

func parseContract(contractName string, contractAbi abi.ABI) Contract {
	service := Service{
		Comment: contractName,
		Name:    contractName,
		Methods: make([]method, len(contractAbi.Methods)),
	}
	var messages []Message

	methodIndex := 0
	for methodName, methodInfo := range contractAbi.Methods {
		inputMessage := fmt.Sprintf("Request%s", toUpperCase(methodName, 0))
		outputMessage := fmt.Sprintf("Response%s", toUpperCase(methodName, 0))

		if len(methodInfo.Inputs) == 0 {
			inputMessage = EmptyMessage
		} else {
			messages = append(messages, parseMessage(
				inputMessage,
				methodInfo.Sig(),
				methodInfo.Inputs),
			)
		}

		if len(methodInfo.Outputs) == 0 {
			outputMessage = EmptyMessage
		} else {
			messages = append(messages, parseMessage(
				outputMessage,
				methodInfo.Sig(),
				methodInfo.Outputs),
			)
		}

		service.Methods[methodIndex] = method{
			Name:   methodName,
			Input:  inputMessage,
			Output: outputMessage,
		}

		methodIndex += 1
	}

	return Contract{
		PackageName: toLowerCase(contractName, 0),
		Services:    []Service{service},
		Messages:    messages,
	}
}

func parseContracts(deployments deployments.Deployments) []Contract {
	contracts, contractsIndex := make([]Contract, len(deployments)), 0
	for name, deployment := range deployments {
		contracts[contractsIndex] = parseContract(name, deployment)
		contractsIndex += 1
	}
	return contracts
}
