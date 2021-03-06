package contracts

const Caller = `
{{define "Caller"}}{{$contract := .}}{{$structs := .Structs}}
    // {{$contract.Type}}Caller is an auto generated read-only Go binding around an Ethereum contract.
    type {{$contract.Type}}Caller interface { {{range $contract.Calls}}
        {{.Normalized.Name}}(
            ctx context.Context, {{range .Normalized.Inputs}}
            {{.Name}} {{bindtype .Type $structs}},{{end}}
        ) ({{if .Structured}}struct{
        {{range .Normalized.Outputs}}{{.Name}} {{bindtype .Type $structs}};
        {{end}}
        },{{else}}{{range .Normalized.Outputs}}
            {{bindtype .Type $structs}},{{end}}
        {{end}} error,
        ){{end}}
    }

    type {{decapitalise $contract.Type}}Caller struct {
        contract *ablbind.BoundContract // Generic contract wrapper for the low level calls
    }

    {{range $contract.Calls}}
        // {{.Normalized.Name}} is a free data retrieval call binding the contract method 0x{{printf "%x" .Original.ID}}.
        //
        // Solidity: {{formatmethod .Original $structs}}
        func (_{{$contract.Type}} *{{decapitalise $contract.Type}}Caller) {{.Normalized.Name}}(ctx context.Context {{range .Normalized.Inputs}}, {{.Name}} {{bindtype .Type $structs}} {{end}}) ({{if .Structured}}struct{ {{range .Normalized.Outputs}}{{.Name}} {{bindtype .Type $structs}};{{end}} },{{else}}{{range .Normalized.Outputs}}{{bindtype .Type $structs}},{{end}}{{end}} error) {
            {{if .Structured}}ret := new(struct{
                {{range .Normalized.Outputs}}{{.Name}} {{bindtype .Type $structs}}
                {{end}}
            }){{else}}var (
                {{range $i, $_ := .Normalized.Outputs}}ret{{$i}} = new({{bindtype .Type $structs}})
                {{end}}
            ){{end}}
            out := {{if .Structured}}ret{{else}}{{if eq (len .Normalized.Outputs) 1}}ret0{{else}}&[]interface{}{
                {{range $i, $_ := .Normalized.Outputs}}ret{{$i}},
                {{end}}
            }{{end}}{{end}}

            err := _{{$contract.Type}}.contract.Call(&bind.CallOpts{Context: ctx}, out, "{{.Original.Name}}" {{range .Normalized.Inputs}}, {{.Name}}{{end}})
            return {{if .Structured}}*ret,{{else}}{{range $i, $_ := .Normalized.Outputs}}*ret{{$i}},{{end}}{{end}} err
        }
    {{end}}
{{end}}
`
