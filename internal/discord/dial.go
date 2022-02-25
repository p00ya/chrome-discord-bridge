//go:build !windows

package discord

import (
	"fmt"
	"net"
	"os"
)

// getDiscordSocket constructs a path to a Discord IPC socket.
//
// A socket may not actually exist at the returned path.
func getDiscordSocket(tmpDir string, n int) string {
	return fmt.Sprintf("%s/discord-ipc-%d", tmpDir, n)
}

// Dial opens the Discord socket and returns a client for sending messages.
func dialIn(tmpDir string) (*Client, error) {
	var err error

	// Socket may be numbered from 0 to 9.
	for i := 0; i < 10; i++ {
		addr := getDiscordSocket(tmpDir, i)

		var conn net.Conn
		// Go's "unix" network is equivalent to AF_UNIX/SOCK_STREAM.
		if conn, err = net.Dial("unix", addr); err != nil {
			continue
		}

		return newClient(conn), nil
	}

	return nil, fmt.Errorf("got errors opening Discord sockets, last was: %w", err)
}

func Dial() (*Client, error) {
	return dialIn(os.TempDir())
}
