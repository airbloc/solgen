package managers

const Managers = `
{{define "managers"}}
package {{.Package}}

import (
    "math/big"
    "strings"

    {{range $name, $import := .Imports}}{{$name}} "{{$import}}"
    {{end}})

{{$contract := .Contract}}
{{$structs := .Contract.Structs}}

//go:generate mockgen -source {{toSnakeCase $contract.Type}}.go -destination ./mocks/mock_{{toSnakeCase $contract.Type}}.go -package mocks I{{$contract.Type}}Manager

type {{$contract.Type}}Manager interface {
    Address() common.Address
    TxHash() common.Hash
    CreatedAt() *big.Int

    contracts.{{$contract.Type}}Caller

    {{range $contract.Transacts}}{{.Normalized.Name}}(
        ctx context.Context,
        opts *ablbind.TransactOpts, {{range .Normalized.Inputs}}
        {{.Name}} {{bindtype .Type $structs}},{{end}}
    ) ({{if .Structured}}
        struct{ {{range .Normalized.Outputs}}
            {{.Name}} {{bindtype .Type $structs}};{{end}}
        },
        {{else}}{{range .Normalized.Outputs}}
            {{bindtype .Type $structs}},
        {{end}}{{end}} error,
    )
    {{end}}

    contracts.{{$contract.Type}}EventFilterer
    contracts.{{$contract.Type}}EventWatcher
}

// {{decapitalise $contract.Type}}Manager is contract wrapper struct
type {{decapitalise $contract.Type}}Manager struct {
    *contracts.{{$contract.Type}}Contract
    client ablbind.ContractBackend
    log    logger.Logger
}

// New{{$contract.Type}}Manager makes new *{{decapitalise $contract.Type}}Manager struct
func New{{$contract.Type}}Manager(backend ablbind.ContractBackend) ({{$contract.Type}}Manager, error) {
    contract, err := contracts.New{{$contract.Type}}Contract(backend)
    if err != nil {
        return nil, err
    }

    return &{{decapitalise $contract.Type}}Manager{
        {{$contract.Type}}Contract: contract,
        client:                     backend,
        log:                        logger.New("{{toSnakeCase $contract.Type}}"),
    }, nil
}

{{range $contract.Transacts}}
// {{.Normalized.Name}} is a paid mutator transaction binding the contract method 0x{{printf "%x" .Original.ID}}.
//
// Solidity: {{.Original.String}}
func (manager *{{decapitalise $contract.Type}}Manager) {{.Normalized.Name}}(
    ctx context.Context,
    opts *ablbind.TransactOpts,
    {{range .Normalized.Inputs}}{{.Name}} {{bindtype .Type $structs}},
    {{end}}) ({{if .Structured}}struct{
    {{range .Normalized.Outputs}}{{.Name}} {{bindtype .Type $structs}};
    {{end}}
},{{else}}{{range .Normalized.Outputs}}
    {{bindtype .Type $structs}},{{end}}
    {{end}} error,
) {
    return {{if .Structured}}nil,{{else}}{{range .Normalized.Outputs}}nil,{{end}}{{end}} nil
}
{{end}}{{end}}
`
