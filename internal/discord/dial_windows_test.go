package discord

import (
	"fmt"
	"net"
	"testing"
	"time"
)

import winio "github.com/Microsoft/go-winio"

func fakeListener(prefix string) (net.Listener, error) {
	c := winio.PipeConfig{}
	name := fmt.Sprintf(`\\?\pipe\%sdiscord-ipc-0`, prefix)
	return winio.ListenPipe(name, &c)
}

func TestDial(t *testing.T) {
	// Prevent collisions with concurrent tests or real Discord.
	prefix := fmt.Sprintf("test-%d-", time.Now().UnixNano())

	listener, err := fakeListener(prefix)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = listener.Close()
	})

	serverDone := make(chan net.Conn)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Got %v while waiting for connections", err)
		}
		serverDone <- conn
	}()

	clientDone := make(chan *Client)
	go func() {
		client, err := dialPrefix(prefix)
		if err != nil {
			t.Errorf("Got %v while dialing", err)
		}
		clientDone <- client
	}()

	// Wait for both the client and server goroutines to finish.
	for n := 2; n > 0; {
		select {
		case <-serverDone:
			n--
		case <-clientDone:
			n--
		case <-time.After(timeoutSeconds * time.Second):
			t.Fatal("Timeout")
			n = 0
		}
	}
}
