package main

import (
	"encoding/json"
)

import "github.com/p00ya/chrome-discord-bridge/internal/discord"

// FrameRequest contains the generic outer fields for Discord JSON requests.
type FrameRequest struct {
	Nonce string      `json:"nonce"`
	Args  interface{} `json:"args"`
	Cmd   string      `json:"cmd"`
}

// Activity represents a Discord activity.
// https://discord.com/developers/docs/topics/gateway#activity-object
// can be used as a guide, but not all fields and values are supported.
type Activity struct {
	State   string `json:"state"`
	Details string `json:"details,omitempty"`
}

// SetActivityArgs represents the "args" in a Discord SET_ACTIVITY request.
// Documented at:
// https://github.com/discord/discord-rpc/blob/master/documentation/hard-mode.md
//
// The description at:
// https://discord.com/developers/docs/topics/rpc#setactivity
// can also be used as a guide, but various fields from the websocket API are
// not supported and the values have different limitations.
type SetActivityArgs struct {
	Pid      int      `json:"pid"`
	Activity Activity `json:"activity"`
}

// sendActivity sends a SET_ACTIVITY request via IPC to Discord.
// It returns the JSON response from Discord, or an error.
func sendSetActivity(discordClient *discord.Client, pid int, state string, details string) ([]byte, error) {
	request := FrameRequest{
		Nonce: "1",
		Args: SetActivityArgs{
			Pid: pid,
			Activity: Activity{
				State:   state,
				Details: details,
			},
		},
		Cmd: "SET_ACTIVITY",
	}

	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	return discordClient.Send(requestJson)
}
