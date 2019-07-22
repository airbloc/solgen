package bind

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type event struct {
	Original   abi.Event
	Normalized abi.Event
}

type events map[string]*event

func parseEvents(evmEvents map[string]abi.Event) events {
	events := make(events)

	for _, original := range evmEvents {
		// Skip anonymous events as they don't support explicit filtering
		if original.Anonymous {
			continue
		}
		// Normalize the tmplEvent for capital cases and non-anonymous outputs
		normalized := original
		normalized.Name = capitalise(original.Name)

		normalized.Inputs = make([]abi.Argument, len(original.Inputs))
		copy(normalized.Inputs, original.Inputs)
		for j, input := range normalized.Inputs {
			// Indexed fields are input, non-indexed ones are outputs
			if input.Indexed {
				if input.Name == "" {
					normalized.Inputs[j].Name = fmt.Sprintf("arg%d", j)
				}
			}
		}
		// Append the tmplEvent to the accumulator list
		events[original.Name] = &event{Original: original, Normalized: normalized}
	}

	return events
}
