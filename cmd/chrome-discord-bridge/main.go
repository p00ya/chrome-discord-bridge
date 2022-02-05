// package main implements a command-line utility for manually setting Discord's
// rich presence activity.
package main

import (
	"log"
	"os"
)

import (
	"github.com/p00ya/chrome-discord-bridge/internal/chrome"
	"github.com/p00ya/chrome-discord-bridge/internal/discord"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Error: wanted origin URL argument, got %d args", len(os.Args)-1)
	}

	origin := os.Args[1]
	if !IsValidOrigin(origin) {
		log.Fatalf("Error: invalid origin %s", origin)
	}

	tmpDir := os.TempDir()
	var discordClient *discord.Client
	var err error
	discordClient, err = discord.Dial(tmpDir)
	if err != nil {
		log.Fatalf("Error connecting to Discord socket: %v\n", err)
	}
	go discordClient.Start()
	defer discordClient.Close()

	host := chrome.NewHost(os.Stdin, os.Stdout)
	go host.Start()
	defer host.Close()

	for {
		req, responder := host.Receive()
		if req == nil {
			// Clean exit - Chrome destroyed native messaging port.
			break
		}
		res, err := discordClient.Send(req)
		if err != nil {
			log.Fatalf("Error receiving from Discord: %v\n", err)
		}
		responder.Respond(res)
	}
}
