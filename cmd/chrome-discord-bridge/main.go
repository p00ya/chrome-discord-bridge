// package main implements a command-line utility for manually setting Discord's
// rich presence activity.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

import (
	"github.com/p00ya/chrome-discord-bridge/internal/chrome"
	"github.com/p00ya/chrome-discord-bridge/internal/chrome/install"
	"github.com/p00ya/chrome-discord-bridge/internal/discord"
)

const usageMessage = `Usage:

    chrome-discord-bridge -install
    chrome-discord-bridge ORIGIN
`

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
}

const (
	exitSuccess      = 0
	exitInvalidUsage = 1
	exitFailure      = 2
)

func main() {
	install := flag.Bool("install", false, "Install Chrome manifest for current user")

	flag.Usage = usage
	flag.Parse()
	if *install {
		if flag.NArg() > 0 {
			fmt.Fprintf(os.Stderr, "No arguments expected with -install, got %d\n", flag.NArg())
			os.Exit(exitInvalidUsage)
		}
		runInstall()
	} else {
		serveChrome()
	}
}

// name is the host used to register chrome-discord-bridge with Chrome.
const name = "io.github.p00ya.cdb"

// description is used to register chrome-discord-bridge with Chrome.
const description = `Chrome/Discord bridge - see https://github.com/p00ya/chrome-discord-bridge`

func runInstall() {
	binary := os.Args[0]
	absPath, err := filepath.Abs(binary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving absolute path to %s\n", binary)
		os.Exit(exitFailure)
	}

	m := install.Manifest{
		Name:           name,
		Description:    description,
		Path:           absPath,
		AllowedOrigins: uniqueOrigins(),
	}

	err = install.CurrentUser(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitFailure)
	}

	fmt.Printf("Wrote manifest for %s\n", name)
}

func serveChrome() {
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
