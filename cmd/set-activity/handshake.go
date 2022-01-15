package main

import (
	"encoding/json"
)

import "github.com/p00ya/chrome-discord-bridge/internal/discord"

// HandshakeRequest represents the Discord handshake JSON.
// It doesn't appear to be officially documented.
type HandshakeRequest struct {
	// Undocumented API version, 1 is the only known value.
	Version int `json:"v"`
	// Application ID from the Discord developer portal.
	ClientId string `json:"client_id"`
	// A token or sequence number for matching the response.
	Nonce string `json:"nonce,omitifempty"`
}

// sendHandshake sends a handshake to Discord.
// It returns the JSON response from Discord, or an error.
func sendHandshake(discordClient *discord.Client, clientId string) ([]byte, error) {
	request := HandshakeRequest{
		Version:  1,
		ClientId: clientId,
		Nonce:    "0",
	}
	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	return discordClient.Send(requestJson)
}
