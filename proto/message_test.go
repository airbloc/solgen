package proto

import (
	"log"
	"testing"
)

func TestMessage_PrintMessage(t *testing.T) {
	msg := Message{
		Comment: "frostornge",
		Name:    "airbloc",
		Msgs: []Message{
			{
				Comment: "test struct",
				Name:    "struct",
				Msgs: []Message{
					{
						Comment: "test struct in struct",
						Name:    "struct in struct",
						Args: []argument{
							{
								Name:     "structArg1",
								Repeated: true,
								Type:     "string",
								Count:    1,
							},
							{
								Name:     "structArg2",
								Repeated: false,
								Type:     "uint32",
								Count:    2,
							},
							{
								Name:     "structArg3",
								Repeated: false,
								Type:     "string",
								Count:    3,
							},
						},
					},
				},
				Args: []argument{
					{
						Name:     "structArg1",
						Repeated: true,
						Type:     "string",
						Count:    1,
					},
					{
						Name:     "structArg2",
						Repeated: false,
						Type:     "uint32",
						Count:    2,
					},
					{
						Name:     "structArg3",
						Repeated: false,
						Type:     "string",
						Count:    3,
					},
				},
			},
		},
		Args: []argument{
			{
				Name:     "messageArg1",
				Repeated: true,
				Type:     "struct",
				Count:    1,
			},
			{
				Name:     "messageArg2",
				Repeated: false,
				Type:     "string",
				Count:    2,
			},
			{
				Name:     "messageArg3",
				Repeated: true,
				Type:     "uint64",
				Count:    3,
			},
		},
	}

	log.Println(msg.PrintMessage())
}
