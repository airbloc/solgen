package contracts

const Filterer = `
{{define "Filterer"}}{{$contract := .}}{{$structs := .Structs}}
    type {{$contract.Type}}Events interface {
        {{$contract.Type}}EventFilterer
        {{$contract.Type}}EventParser
        {{$contract.Type}}EventWatcher
    }

    // {{$contract.Type}}Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
    type {{$contract.Type}}EventFilterer interface { {{range $contract.Events}}
        // Filterer
        Filter{{.Normalized.Name}}(
            opts *bind.FilterOpts,
            {{range .Normalized.Inputs}}{{if .Indexed}}{{.Name}} []{{bindtype .Type $structs}},{{end}}{{end}}
        ) (ablbind.EventIterator, error)
    {{end}} }

    type {{$contract.Type}}EventParser interface { {{range $contract.Events}}
        // Parser
        Parse{{.Normalized.Name}}(log chainTypes.Log) (*{{$contract.Type}}{{.Normalized.Name}}, error)
        Parse{{.Normalized.Name}}FromReceipt(receipt *chainTypes.Receipt) ([]*{{$contract.Type}}{{.Normalized.Name}}, error)
    {{end}} }

    type {{$contract.Type}}EventWatcher interface { {{range $contract.Events}}
        // Watcher
        Watch{{.Normalized.Name}}(
            opts *bind.WatchOpts,
            sink chan<- *{{$contract.Type}}{{.Normalized.Name}},
            {{range .Normalized.Inputs}}{{if .Indexed}}{{.Name}} []{{bindtype .Type $structs}},{{end}}{{end}}
        ) (event.Subscription, error)
    {{end}} }

    type {{decapitalise $contract.Type}}Events struct {
        contract *ablbind.BoundContract // Generic contract wrapper for the low level calls
    }

    {{range $contract.Events}}
        // {{$contract.Type}}{{.Normalized.Name}}Iterator is returned from Filter{{.Normalized.Name}} and is used to iterate over the raw logs and unpacked data for {{.Normalized.Name}} events raised by the {{$contract.Type}} contract.
        type {{$contract.Type}}{{.Normalized.Name}}Iterator struct {
            Evt *{{$contract.Type}}{{.Normalized.Name}} // Event containing the contract specifics and raw log

            contract *ablbind.BoundContract // Generic contract to use for unpacking event data
            event    string              // Event name to use for unpacking event data

            logs chan chainTypes.Log        // Log channel receiving the found contract events
            sub  platform.Subscription // Subscription for errors, completion and termination
            done bool                  // Whether the subscription completed delivering logs
            fail error                 // Occurred error to stop iteration
        }
        // Next advances the iterator to the subsequent event, returning whether there
        // are any more events found. In case of a retrieval or parsing error, false is
        // returned and Error() can be queried for the exact failure.
        func (it *{{$contract.Type}}{{.Normalized.Name}}Iterator) Next() bool {
            // If the iterator failed, stop iterating
            if (it.fail != nil) {
                return false
            }
            // If the iterator completed, deliver directly whatever's available
            if (it.done) {
                select {
                case log := <-it.logs:
                    it.Evt = new({{$contract.Type}}{{.Normalized.Name}})
                    if err := it.contract.UnpackLog(it.Evt, it.event, log); err != nil {
                        it.fail = err
                        return false
                    }
                    it.Evt.Raw = log
                    return true

                default:
                    return false
                }
            }
            // Iterator still in progress, wait for either a data or an error event
            select {
            case log := <-it.logs:
                it.Evt = new({{$contract.Type}}{{.Normalized.Name}})
                if err := it.contract.UnpackLog(it.Evt, it.event, log); err != nil {
                    it.fail = err
                    return false
                }
                it.Evt.Raw = log
                return true

            case err := <-it.sub.Err():
                it.done = true
                it.fail = err
                return it.Next()
            }
        }
        // Error returns any retrieval or parsing error occurred during filtering.
        func (it *{{$contract.Type}}{{.Normalized.Name}}Iterator) Event() interface{} {
            return it.Evt
        }
        // Error returns any retrieval or parsing error occurred during filtering.
        func (it *{{$contract.Type}}{{.Normalized.Name}}Iterator) Error() error {
            return it.fail
        }
        // Close terminates the iteration process, releasing any pending underlying
        // resources.
        func (it *{{$contract.Type}}{{.Normalized.Name}}Iterator) Close() error {
            it.sub.Unsubscribe()
            return nil
        }

        // {{$contract.Type}}{{.Normalized.Name}} represents a {{.Normalized.Name}} event raised by the {{$contract.Type}} contract.
        type {{$contract.Type}}{{.Normalized.Name}} struct { {{range .Normalized.Inputs}}
            {{capitalise .Name}} {{if .Indexed}}{{bindtopictype .Type $structs}}{{else}}{{bindtype .Type $structs}}{{end}}; {{end}}
            Raw chainTypes.Log // Blockchain specific contextual infos
        }

        // Filter{{.Normalized.Name}} is a free log retrieval operation binding the contract event 0x{{printf "%x" .Original.ID}}.
        //
        // Solidity: {{formatevent .Original $structs}}
        func (_{{$contract.Type}} *{{decapitalise $contract.Type}}Events) Filter{{.Normalized.Name}}(opts *bind.FilterOpts{{range .Normalized.Inputs}}{{if .Indexed}}, {{.Name}} []{{bindtype .Type $structs}}{{end}}{{end}}) (ablbind.EventIterator, error) {
            {{range .Normalized.Inputs}}
            {{if .Indexed}}var {{.Name}}Rule []interface{}
            for _, {{.Name}}Item := range {{.Name}} {
                {{.Name}}Rule = append({{.Name}}Rule, {{.Name}}Item)
            }{{end}}{{end}}

            logs, sub, err := _{{$contract.Type}}.contract.FilterLogs(opts, "{{.Original.Name}}"{{range .Normalized.Inputs}}{{if .Indexed}}, {{.Name}}Rule{{end}}{{end}})
            if err != nil {
                return nil, err
            }
            return &{{$contract.Type}}{{.Normalized.Name}}Iterator{contract: _{{$contract.Type}}.contract, event: "{{.Original.Name}}", logs: logs, sub: sub}, nil
        }

        // Watch{{.Normalized.Name}} is a free log subscription operation binding the contract event 0x{{printf "%x" .Original.ID}}.
        //
        // Solidity: {{formatevent .Original $structs}}
        func (_{{$contract.Type}} *{{decapitalise $contract.Type}}Events) Watch{{.Normalized.Name}}(opts *bind.WatchOpts, sink chan<- *{{$contract.Type}}{{.Normalized.Name}}{{range .Normalized.Inputs}}{{if .Indexed}}, {{.Name}} []{{bindtype .Type $structs}}{{end}}{{end}}) (event.Subscription, error) {
            {{range .Normalized.Inputs}}
            {{if .Indexed}}var {{.Name}}Rule []interface{}
            for _, {{.Name}}Item := range {{.Name}} {
                {{.Name}}Rule = append({{.Name}}Rule, {{.Name}}Item)
            }{{end}}{{end}}

            logs, sub, err := _{{$contract.Type}}.contract.WatchLogs(opts, "{{.Original.Name}}"{{range .Normalized.Inputs}}{{if .Indexed}}, {{.Name}}Rule{{end}}{{end}})
            if err != nil {
                return nil, err
            }
            return event.NewSubscription(func(quit <-chan struct{}) error {
                defer sub.Unsubscribe()
                for {
                    select {
                    case log := <-logs:
                        // New log arrived, parse the event and forward to the user
                        evt := new({{$contract.Type}}{{.Normalized.Name}})
                        if err := _{{$contract.Type}}.contract.UnpackLog(evt, "{{.Original.Name}}", log); err != nil {
                            return err
                        }
                        evt.Raw = log

                        select {
                        case sink <- evt:
                        case err := <-sub.Err():
                            return err
                        case <-quit:
                            return nil
                        }
                    case err := <-sub.Err():
                        return err
                    case <-quit:
                        return nil
                    }
                }
            }), nil
        }

        // Parse{{.Normalized.Name}} is a log parse operation binding the contract event 0x{{printf "%x" .Original.ID}}.
        //
        // Solidity: {{.Original.String}}
        func (_{{$contract.Type}} *{{decapitalise $contract.Type}}Events) Parse{{.Normalized.Name}}(log chainTypes.Log) (*{{$contract.Type}}{{.Normalized.Name}}, error) {
            evt := new({{$contract.Type}}{{.Normalized.Name}})
            if err := _{{$contract.Type}}.contract.UnpackLog(evt, "{{.Original.Name}}", log); err != nil {
                return nil, err
            }
            return evt, nil
        }

        // Parse{{.Normalized.Name}}FromReceipt parses the event from given transaction receipt.
        //
        // Solidity: {{.Original.String}}
        func (_{{$contract.Type}} *{{decapitalise $contract.Type}}Events) Parse{{.Normalized.Name}}FromReceipt(receipt *chainTypes.Receipt) ([]*{{$contract.Type}}{{.Normalized.Name}}, error) {
            var evts []*{{$contract.Type}}{{.Normalized.Name}}
            for _, log := range receipt.Logs {
                if log.Topics[0] == common.HexToHash("0x{{printf "%x" .Original.ID}}") {
                    evt, err := _{{$contract.Type}}.Parse{{.Normalized.Name}}(*log)
                    if err != nil {
                        return nil, err
                    }
                    evts = append(evts, evt)
                }
            }

            if len(evts) == 0 {
                return nil, errors.New("{{.Original.Name}} event not found")
            }
            return evts, nil
        }
    {{end}}
{{end}}
`
