// package main implements a command-line utility for manually setting Discord's
// rich presence activity.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
)

import "github.com/p00ya/chrome-discord-bridge/internal/discord"

const (
	activityGameType   = 0
	activityStreaming  = 1
	activityListening  = 2
	activityWatching   = 3
	activityCustom     = 4
	acitivityCompeting = 5
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n"+
		"%s [-d DETAILS] [-p PID] CLIENT_ID ACTIVITY_STATE\n\n", os.Args[0])
	flag.PrintDefaults()
}

const (
	exitSuccess      = 0
	exitInvalidUsage = 1
	exitFailure      = 2
)

const (
	argClientId      = 0
	argActivityState = 1
)

// Maps symbolic client names to client IDs.
var clientMap = map[string]string{
	// https://github.com/PreMiD/Presences/blob/main/websites/M/Monkeytype/presence.ts
	"monkeytype": "798272335035498557",
	// https://github.com/PreMiD/Presences/blob/main/websites/W/WaniKani/presence.ts
	"wanikani": "800166344023867443",
}

func clientKeys() []string {
	keys := make([]string, len(clientMap))
	i := 0
	for k, _ := range clientMap {
		keys[i] = k
		i++
	}
	return keys
}

func getClientId(clientIdOrName string) (clientId string, err error) {
	if id, ok := clientMap[clientIdOrName]; ok {
		clientId = id
	} else if ok, _ := regexp.MatchString(`\d+`, clientIdOrName); !ok {
		return "", fmt.Errorf(
			"invalid CLIENT_ID '%s'; must be number or one of: [%s]",
			clientIdOrName, strings.Join(clientKeys(), ", "))
	} else {
		clientId = clientIdOrName
	}
	return clientId, nil
}

func main() {
	var detailsFlag = flag.String("d", "", "Activity details")
	var pidFlag = flag.Int("p", -1, "PID of the activity")
	flag.Usage = printUsage
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Error: expected 2 arguments, got %d\n", flag.NArg())
		printUsage()
		os.Exit(exitInvalidUsage)
	}
	var clientId string
	var err error
	clientId, err = getClientId(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		printUsage()
		os.Exit(exitInvalidUsage)
	}

	activityState := flag.Arg(1)
	pid := *pidFlag
	if pid < 0 {
		// Use our own PID by default
		pid = os.Getpid()
	}

	tmpDir := os.TempDir()
	var discordClient *discord.Client
	discordClient, err = discord.Dial(tmpDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to Discord socket: %v", err)
		os.Exit(exitFailure)
	}
	go discordClient.Start()

	if res, err := sendHandshake(discordClient, clientId); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending HANDSHAKE: %v", err)
		os.Exit(exitFailure)
	} else {
		fmt.Println(string(res))
	}
	if res, err := sendSetActivity(discordClient, pid, activityState, *detailsFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending SET_ACTIVITY: %v", err)
		os.Exit(exitFailure)
	} else {
		fmt.Println(string(res))
	}

	// Wait for a user interrupt before exiting.
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	<-sigint
	discordClient.Close()
	os.Exit(exitSuccess)
}
