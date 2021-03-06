package contracts

const Transactor = `
{{define "Transactor"}}{{$contract := .}}{{$structs := .Structs}}
    // {{$contract.Type}}Transactor is an auto generated write-only Go binding around an Ethereum contract.
    type {{$contract.Type}}Transactor interface { {{range $contract.Transacts}}
        {{.Normalized.Name}}(
            ctx context.Context,
            opts *ablbind.TransactOpts,
            {{range .Normalized.Inputs}}{{.Name}} {{bindtype .Type $structs}},
        {{end}}) (*chainTypes.Receipt, error){{end}}
    }

    type {{decapitalise $contract.Type}}Transactor struct {
        contract *ablbind.BoundContract // Generic contract wrapper for the low level calls
        backend ablbind.ContractBackend
    }

    {{range $contract.Transacts}}
        // {{.Normalized.Name}} is a paid mutator transaction binding the contract method 0x{{printf "%x" .Original.ID}}.
        //
        // Solidity: {{.Original.String}}
        func (_{{$contract.Type}} *{{decapitalise $contract.Type}}Transactor) {{.Normalized.Name}}(
            ctx context.Context,
            opts *ablbind.TransactOpts,
            {{range .Normalized.Inputs}}{{.Name}} {{bindtype .Type $structs}},
        {{end}}) (*chainTypes.Receipt, error) {
            if opts == nil {
                opts = &ablbind.TransactOpts{}
            }
            opts.Context = ctx

            return _{{$contract.Type}}.contract.Transact(opts, "{{.Original.Name}}" {{range .Normalized.Inputs}}, {{.Name}}{{end}})
        }
    {{end}}
{{end}}
`
