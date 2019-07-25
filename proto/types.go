package proto

import (
	"log"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

const StructPrefix = "XYX__tmp"

func parseType(index int, a abi.Argument, typeOptions map[string]string) (msg *message, arg *argument) {
	argName := toLowerCase(parseName(index, a), 0)
	argType := a.Type

	repeated := argType.T == abi.ArrayTy || argType.T == abi.SliceTy

	arg = &argument{
		Count:    index,
		Name:     argName,
		Repeated: repeated,
		Type:     parseSimpleType(argType),
	}

	if arg.Type == "struct" {
		log.Println(a)
		log.Println(argType.String(), argType.TupleRawNames)
		msg = &message{
			Args:    make([]argument, len(argType.TupleElems)),
			Comment: argType.Type.String(),
			Name:    StructPrefix + toUpperCase(argType.Kind.String(), 0) + strconv.Itoa(index),
		}

		if !strings.HasPrefix(argName, "struct") {
			arg.Name = toLowerCase(argName, 0)
			msg.Name = toUpperCase(argName, 0)
		}

		if name, ok := typeOptions[argType.String()]; ok {
			arg.Name = toLowerCase(name, 0)
			msg.Name = toUpperCase(name, 0)
		}

		args := make([]abi.Argument, len(argType.TupleElems))
		for index := range argType.TupleElems {
			args[index] = abi.Argument{
				Name: argType.TupleRawNames[index],
				Type: *argType.TupleElems[index],
			}
		}

		msg.parseArguments(args)
		arg.Type = msg.Name
	}

	return
}

func parseSimpleType(t abi.Type) string {
	switch t.T {
	case abi.IntTy:
		if t.Size <= 32 {
			return "int32"
		} else if t.Size <= 64 {
			return "int64"
		}
	case abi.UintTy:
		if t.Size <= 32 {
			return "uint32"
		} else if t.Size <= 64 {
			return "uint64"
		}
	case abi.BoolTy:
		return "bool"
	case abi.StringTy:
	case abi.SliceTy:
	case abi.ArrayTy:
	case abi.AddressTy:
	case abi.TupleTy:
		return "struct"
	case abi.FixedBytesTy:
	case abi.BytesTy:
	case abi.HashTy:
	case abi.FixedPointTy:
	case abi.FunctionTy:
	}
	return "string"
}

func parseName(index int, arg abi.Argument) string {
	argName := arg.Name
	argType := arg.Type.Kind.String()

	switch argName {
	case "":
		return argType + strconv.Itoa(index)
	default:
		return argName
	}
}
