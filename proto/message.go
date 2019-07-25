package proto

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type argument struct {
	Name     string
	Repeated bool
	Type     string
	Count    int
}

func (arg argument) printArg(prefix string) string {
	str := fmt.Sprintf("%s %s = %d;", arg.Type, arg.Name, arg.Count)
	if arg.Repeated {
		str = "repeated " + str
	}
	return prefix + str + "\n"
}

type message struct {
	Comment string
	Name    string
	Msgs    []message
	Args    []argument

	contractName string
	typeOptions  Options
}

func (msg *message) parseArguments(args []abi.Argument) {
	for index, arg := range args {
		m, a := parseType(index+1, arg, msg.typeOptions[msg.contractName])
		if m != nil {
			msg.Msgs = append(msg.Msgs, *m)
		}
		msg.Args[index] = *a
	}
}

func (msg message) PrintMessage() string {
	return msg.printMessage("")
}

func (msg message) printMessage(prefix string) string {
	var builder strings.Builder

	builder.WriteString(prefix + fmt.Sprintf("// %s", msg.Comment) + "\n")
	builder.WriteString(prefix + fmt.Sprintf("message %s {", msg.Name) + "\n")

	for _, m := range msg.Msgs {
		builder.WriteString(m.printMessage(prefix + "\t"))
	}

	for _, a := range msg.Args {
		builder.WriteString(a.printArg(prefix + "\t"))
	}

	builder.WriteString(prefix + "}" + "\n")

	return builder.String()
}
