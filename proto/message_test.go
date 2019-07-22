package proto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessage_PrintMessage(t *testing.T) {
	msg := Message{
		Comment: "frostornge",
		Name:    "airbloc",
		Msgs: []Message{
			{
				Comment: "test struct",
				Name:    "struct",
				Args: []argument{
					{
						Name:     "structArg1",
						Repeated: true,
						Type:     "string",
						Count:    1,
					},
				},
			},
		},
		Args: []argument{
			{
				Name:     "messageArg3",
				Repeated: true,
				Type:     "uint64",
				Count:    1,
			},
		},
	}

	expected := "// frostornge\n" +
		"message airbloc {\n\t" +
		"// test struct\n\t" +
		"message struct {\n\t\t" +
		"repeated string structArg1 = 1;\n\t}\n\t" +
		"repeated uint64 messageArg3 = 1;\n}\n"
	assert.Equal(t, expected, msg.PrintMessage())
}
