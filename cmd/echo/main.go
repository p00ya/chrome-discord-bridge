// package main implements a Chrome native messaging host that echoes requests
// back to Chrome.
package main

import (
	"os"
)

import "github.com/p00ya/chrome-discord-bridge/internal/chrome"

func main() {
	host := chrome.NewHost(os.Stdin, os.Stdout)
	go host.Start()
	defer host.Close()

	for {
		req, responder := host.Receive()
		if req == nil {
			// Clean exit - Chrome destroyed native messaging port.
			break
		}
		responder.Respond(req)
	}
}
