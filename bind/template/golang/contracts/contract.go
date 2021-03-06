package contracts

const Contract = `
{{define "contracts"}}
package {{.Package}}

import (
    "math/big"
    "strings"

    {{range $name, $import := .Imports}}{{$name}} "{{$import}}"
    {{end}})

{{template "wrapper" .Contract}}
{{end}}

{{define "wrapper"}}
    {{$contract := .}}{{$structs := .Structs}}

    // {{.Type}}ABI is the input ABI used to generate the binding from.
    const (
        {{.Type}}Address = "{{.Address}}"
        {{.Type}}TxHash = "{{.TxHash}}"
        {{.Type}}CreatedAt = "{{.CreatedAt}}"
        {{.Type}}ABI = "{{.InputABI}}"
    )

    {{template "Caller" .}}
    {{template "Transactor" .}}
    {{template "Filterer" .}}

    // Manager is contract wrapper struct
    type {{$contract.Type}}Contract struct {
        ablbind.Deployment
        client    ablbind.ContractBackend

        {{$contract.Type}}Caller
        {{$contract.Type}}Transactor
        {{$contract.Type}}Events
    }

    func New{{$contract.Type}}Contract(backend ablbind.ContractBackend) (*{{$contract.Type}}Contract, error) {
        deployment, exist := backend.Deployment("{{$contract.Type}}")
        if !exist {
            evmABI, err := abi.JSON(strings.NewReader({{$contract.Type}}ABI))
            if err != nil {
                return nil, err
            }

            deployment = ablbind.NewDeployment(
                common.HexToAddress({{$contract.Type}}Address),
                common.HexToHash({{$contract.Type}}TxHash),
                new(big.Int).SetBytes(common.HexToHash({{$contract.Type}}CreatedAt).Bytes()),
                evmABI,
            )
        }

        base := ablbind.NewBoundContract(deployment.Address(), deployment.ParsedABI, "{{$contract.Type}}", backend)

        contract := &{{$contract.Type}}Contract{
            Deployment: deployment,
            client:    backend,

            {{$contract.Type}}Caller: &{{decapitalise $contract.Type}}Caller{base},
            {{$contract.Type}}Transactor: &{{decapitalise $contract.Type}}Transactor{base, backend},
            {{$contract.Type}}Events: &{{decapitalise $contract.Type}}Events{base},
        }

        return contract, nil
    }
{{end}}
`
