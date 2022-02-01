// Package discord is an interface to Discord's IPC.
//
// Discord exposes an IPC interface over a UNIX domain socket; this is
// documented (with a reference implementation) at:
// https://github.com/discord/discord-rpc/blob/master/documentation/hard-mode.md
//
// Discord's IPC messages consist of a header and payload.  The header contains
// an opcode and the length of the payload.
//
// The opcode is handled automatically. The first message sent will be assigned
// a Handshake opcode. Subsequent messages will be assigned the Frame opcode.
// Pings and close messages (originating from Discord) are handled internally.
package discord

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// Payload is the type for Discord's message payload.
type Payload []byte

// Client is a wrapper for reading and writing to the Discord client via its
// IPC socket.
//
// Outgoing messages can be queued with Send().  They will be processed,
// one at a time, once Start() has been called.
type Client struct {
	// out receives payloads that should be sent to Discord.
	//
	// out is read exclusively by the Start() goroutine, and written to
	// exclusively by Send() and Close().
	out chan Payload

	// in receives messages that were read from Discord.
	//
	// in is written to exclusively by the Start() goroutine, and read
	// exclusively by Send().
	in chan message

	// rw is the Discord IPC socket.
	//
	// It need not be a net.Conn, but it must support concurrent calls to Read
	// and Write like net.Conn.
	rw io.ReadWriteCloser
}

// Discord-RPC message opcodes.
const (
	Handshake = iota
	Frame
	Close
	Ping
	Pong
)

// Message is the Discord RPC message.
//
// The wire format consists of the opcode, payload length, and payload.
// Here, the length is inferred from the payload string.
type message struct {
	// The Discord-RPC opcode.
	Opcode int32

	// Payload is the JSON string encoded as UTF-8.
	Payload Payload
}

// getDiscordSocket constructs a path to a Discord IPC socket.
//
// A socket may not actually exist at the returned path.
func getDiscordSocket(tmpDir string, n int) string {
	return fmt.Sprintf("%s/discord-ipc-%d", tmpDir, n)
}

// newClient creates a Client with the specified IPC socket.
func newClient(rw io.ReadWriteCloser) *Client {
	return &Client{
		// A buffer size of 1 causes Send() to block until the previous message
		// has an answer.
		out: make(chan Payload, 1),
		in:  make(chan message),
		rw:  rw,
	}
}

// Dial opens the Discord socket and returns a client for sending messages.
func Dial(tmpDir string) (*Client, error) {
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

func (c *Client) close() {
	close(c.in)
	c.rw.Close()
	c.rw = nil
}

// Start waits for messages written with Send(), and sends them to the socket.
// It will also listen for any messages initiated by Discord.
func (c *Client) Start() error {
	defer c.close()

	// nextOpcode is the opcode to use for the next packet sent via Send().
	// The first packet sent will be marked as a Handshake, and subsequent packets
	// will be marked as Frames.
	var nextOpcode int32 = Handshake

	// We want to handle events coming from either direction (reads from Discord
	// or writes from Send()).  Each iteration of the loop reads one message from
	// Discord, either if there is an unsolicited one waiting, or as a response
	// to Send().
	for {
		readCh := make(chan messageResult)
		// Will block the for loop from iterating until it returns, since one value
		// from readCh is always read.
		go func() {
			msg, err := readMessage(c.rw)
			readCh <- messageResult{msg, err}
		}()

		select {
		case payload, ok := <-c.out:
			// Got a Send() or Close().
			if !ok {
				return nil
			}

			if err := writeMessage(message{Opcode: nextOpcode, Payload: payload}, c.rw); err != nil {
				return err
			}
			nextOpcode = Frame
			// Block on an answer from Discord.
			var r messageResult
			if r = <-readCh; r.err != nil {
				return r.err
			}
			c.in <- r.msg

		case r := <-readCh:
			// Unsolicited message from Discord.
			switch {
			case r.err != nil:
				return r.err
			case r.msg.Opcode == Ping:
				// Respond immediately to ping.
				r.msg.Opcode = Pong
				if err := writeMessage(r.msg, c.rw); err != nil {
					return err
				}
			case r.msg.Opcode == Close:
				return fmt.Errorf("Discord IPC connection terminated by Discord")
			default:
				return fmt.Errorf("got unexpected opcode: %d, payload: %v", r.msg.Opcode, r.msg.Payload)
			}
		}
	}
}

// Send sends the given payload to Discord and returns the answer payload. It
// blocks on both waiting to send the message, and waiting for the answer to be
// received.
//
// The returned answer does not include the header used by Discord's IPC
// protocol; it's just the payload (typically JSON).
func (c *Client) Send(payload Payload) (Payload, error) {
	c.out <- payload

	if msg, ok := <-c.in; !ok {
		return nil, fmt.Errorf("socket closed while waiting for response")
	} else {
		return msg.Payload, nil
	}
}

// Close terminates the connection to the Discord socket.
//
// Send() calls made after Close() will have errors.
func (c *Client) Close() {
	close(c.out)
}

// headerLen is the number of bytes in the Discord IPC message header.
const headerLen = 8

// encode serializes the message in wire format.
func (m message) encode() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, len(m.Payload)+headerLen))
	binary.Write(buf, binary.LittleEndian, m.Opcode)
	binary.Write(buf, binary.LittleEndian, int32(len(m.Payload)))
	buf.Write(m.Payload)
	return buf.Bytes()
}

// messageResult contains readMessage's result list.
type messageResult struct {
	msg message
	err error
}

// readMessage reads a message from the socket.
func readMessage(r io.Reader) (message, error) {
	var msg message

	header := make([]byte, headerLen)
	switch n, err := r.Read(header); {
	case err != nil:
		return msg, err
	case n != headerLen:
		return msg, fmt.Errorf("wanted %d-byte header, read %d bytes", headerLen, n)
	}

	reader := bytes.NewReader(header)
	binary.Read(reader, binary.LittleEndian, &msg.Opcode)
	var payloadLen int32
	binary.Read(reader, binary.LittleEndian, &payloadLen)

	msg.Payload = make([]byte, payloadLen)
	_, err := io.ReadFull(r, msg.Payload)
	return msg, err
}

// writeMessage writes a message to the socket.
func writeMessage(msg message, w io.Writer) error {
	buf := msg.encode()
	switch n, err := w.Write(buf); {
	case err != nil:
		return err
	case n != len(buf):
		return fmt.Errorf("wanted to write %d bytes, wrote %d bytes", len(buf), n)
	}
	return nil
}
