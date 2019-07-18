package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"unicode"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type Rpc struct {
	Name   string
	Input  string
	Output string
}

type Service struct {
	Name string
	Rpcs []Rpc
}

type Arg struct {
	Name  string
	Type  string
	Count int
}

type Message struct {
	Comment string
	Name    string
	Args    []Arg
}

type contract struct {
	PackageName string
	Services    []Service
	Messages    []Message
}

const tmpl = `syntax = "proto3"
package airbloc.{{.PackageName}}

import "google/protobuf/empty.proto"

{{range .Services}}
service {{.Name}} {
	{{range .Rpcs}}
		rpc {{.Name}}({{.Input}}) returns ({{.Output}});
	{{end}}
}

{{range .Messages}}
// {{.Comment}}
message {{.Name}} {
	{{range .Args}}
		{{.Type}} {{.Name}} = {{.Index}};
	{{end}}
}
{{end}}
{{end}}
`

func parseType(typ string) string {
	return typ
}

func parseName(name string) string {
	return name
}

func main() {
	res, err := http.Get("http://localhost:8500")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	deployments := make(map[string]map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&deployments); err != nil {
		panic(err)
	}

	contracts := make(map[string]abi.ABI, len(deployments))
	for contractName, contractInfo := range deployments {
		rawAbi, err := json.Marshal(contractInfo["abi"])
		if err != nil {
			panic(err)
		}

		parsedAbi, err := abi.JSON(bytes.NewReader(rawAbi))
		if err != nil {
			panic(err)
		}

		contracts[contractName] = parsedAbi
	}

	for contractName, contractAbi := range contracts {
		var c contract

		// first letter lower case
		c.PackageName = string(unicode.ToLower(([]rune(contractName))[0]))
		c.Services = []Service{{
			Name: contractName,
			Rpcs: make([]Rpc, len(contractAbi.Methods)),
		}}
		c.Messages = []Message{}

		var msgIndex = 0
		var rpcIndex = 0
		for methodName, method := range contractAbi.Methods {
			rpc := Rpc{
				Name:   methodName,
				Input:  fmt.Sprintf("Request%s", methodName),
				Output: fmt.Sprintf("Response%s", methodName),
			}

			if len(method.Inputs) == 0 {
				rpc.Input = "google.protobuf.Empty"
			}
			if len(method.Outputs) == 0 {
				rpc.Output = "google.protobuf.Empty"
			}

			msg := Message{
				Comment: method.Sig(),
				Args:    make([]Arg, len(method.Inputs)),
				Name:    rpc.Input,
			}
			for index, input := range method.Inputs {
				msg.Args[index] = Arg{
					Count: index,
					Name:  parseName(input.Name),
					Type:  parseType(input.Type.String()),
				}
			}

			c.Messages[msgIndex] = msg
			c.Services[0].Rpcs[rpcIndex] = rpc
		}
	}

}
