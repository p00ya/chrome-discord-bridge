package discord

import (
	"fmt"
	"net"
	"time"
)

import winio "github.com/Microsoft/go-winio"

func getDiscordNamedPipe(prefix string, n int) string {
	// Local named pipes usually have a "\\." prefix:
	// https://docs.microsoft.com/en-us/windows/win32/ipc/pipe-names
	//
	// However, the official discord-rpc library uses a "\\?" prefix.  There
	// doesn't seem to be a good reason for this, since it just disables parsing
	// and size limitations that wouldn't apply anyway:
	// https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file#win32-file-namespaces
	return fmt.Sprintf(`\\?\pipe\%sdiscord-ipc-%d`, prefix, n)
}

// DialPrefix opens the Discord socket and returns a client for sending
// messages.
//
// Instead of using the normal name pipe that Discord uses, it prepends the
// prefix.  This is useful for testing purposes (to not collide with the real
// pipe).
func dialPrefix(prefix string) (*Client, error) {
	var err error

	// Socket may be numbered from 0 to 9.
	for i := 0; i < 10; i++ {
		addr := getDiscordNamedPipe(prefix, i)

		var conn net.Conn
		timeout := time.Second
		if conn, err = winio.DialPipe(addr, &timeout); err != nil {
			continue
		}

		return newClient(conn), nil
	}

	return nil, fmt.Errorf("got errors opening Discord named pipe, last was: %w", err)
}

// Dial opens the Discord socket and returns a client for sending messages.
func Dial() (*Client, error) {
	return dialPrefix("")
}
