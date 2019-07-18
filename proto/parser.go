package proto

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

const EmptyMessage = "google.protobuf.Empty"

func parseType(t abi.Type) (string, bool) {
	repeated := t.T == abi.ArrayTy || t.T == abi.SliceTy

	switch t.T {
	case abi.UintTy:
		if t.Size == 8 {
			return "uint32", repeated
		} else if t.Size == 32 {
			return "uint32", repeated
		} else if t.Size == 64 {
			return "uint64", repeated
		}
	case abi.BoolTy:
		return "bool", repeated
	case abi.StringTy:
	case abi.SliceTy:
	case abi.TupleTy:
		for _, elem := range t.TupleElems {
			log.Println(elem)
			if elem.T == abi.TupleTy {
				for _, e := range elem.TupleElems {
					log.Println(e)
				}
			}
		}
	case abi.AddressTy:
	case abi.FixedBytesTy:
	case abi.BytesTy:
	case abi.HashTy:
	default:
		log.Println(t.String(), repeated)
	}

	return "string", repeated
}

func parseName(arg abi.Argument) string {
	name := arg.Name
	typ := arg.Type.String()

	switch name {
	case "":
		return typ[:len(typ)-1]
	default:
		return name
	}
}

func parseArguments(name string, sig string, args abi.Arguments) message {
	msg := message{
		Comment: sig,
		Args:    make([]arg, len(args)),
		Name:    name,
	}

	for index, argument := range args {
		argName := parseName(argument)
		typeName, repeated := parseType(argument.Type)

		msg.Args[index] = arg{
			Count:    index + 1,
			Name:     argName,
			Repeated: repeated,
			Type:     typeName,
		}
	}

	return msg
}

func Parse(deployments Deployments) (contracts []contract) {
	for contractName, contractAbi := range deployments {
		contract := contract{PackageName: toLowerCase(contractName, 0)}

		service := service{Name: contractName}
		for methodName, method := range contractAbi.Methods {
			// parse RPC
			inputMessage := fmt.Sprintf("Request%s", toUpperCase(methodName, 0))
			outputMessage := fmt.Sprintf("Response%s", toUpperCase(methodName, 0))

			if len(method.Inputs) == 0 {
				inputMessage = EmptyMessage
			}
			if len(method.Outputs) == 0 {
				outputMessage = EmptyMessage
			}

			service.Rpcs = append(
				service.Rpcs,
				rpc{
					Name:   methodName,
					Input:  inputMessage,
					Output: outputMessage,
				})

			// input
			if len(method.Inputs) != 0 {
				msg := parseArguments(inputMessage, method.Sig(), method.Inputs)
				contract.Messages = append(contract.Messages, msg)
			}

			if len(method.Outputs) != 0 {
				msg := parseArguments(outputMessage, method.Sig(), method.Outputs)
				contract.Messages = append(contract.Messages, msg)
			}
		}

		contract.Services = append(contract.Services, service)
		contracts = append(contracts, contract)
	}

	return
}
