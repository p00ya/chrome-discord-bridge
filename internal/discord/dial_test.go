//go:build !windows

package discord

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

// fakeServer creates and listens on a UNIX domain socket, like Discord would.
type fakeServer struct {
	TmpDir string
	addr   string
}

func newFakeServer() (fake fakeServer, err error) {
	fake = fakeServer{}

	fake.TmpDir, err = os.MkdirTemp("", "discord_test")
	if err != nil {
		return fake, err
	}
	fake.addr = fmt.Sprintf("%s/discord-ipc-0", fake.TmpDir)
	return fake, err
}

func (f fakeServer) Listener() (net.Listener, error) {
	return net.Listen("unix", f.addr)
}

func (f fakeServer) Close() {
	os.RemoveAll(f.TmpDir)
}

func TestDial(t *testing.T) {
	fake, err := newFakeServer()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(fake.Close)
	listener, err := fake.Listener()
	if err != nil {
		t.Fatal(err)
	}

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
		client, err := dialIn(fake.TmpDir)
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

func TestGetDiscordSocket(t *testing.T) {
	var tests = []struct {
		tmpDir string
		n      int
		want   string
	}{
		{"/tmp/foo", 0, "/tmp/foo/discord-ipc-0"},
		{"/tmp/foo", 1, "/tmp/foo/discord-ipc-1"},
		{"/tmp/bar", 0, "/tmp/bar/discord-ipc-0"},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s,%d", tt.tmpDir, tt.n)
		t.Run(testname, func(t *testing.T) {
			ans := getDiscordSocket(tt.tmpDir, tt.n)
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}
