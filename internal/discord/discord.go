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

	// conn is the Discord IPC socket.
	//
	// conn is accessed exclusively by the Start() goroutine.
	conn net.Conn

	// Blocks until the last message sent with Send() has received an answer.
	waitForAnswer chan int
}

// Discord-RPC message opcodes.
const (
	Handshake = 0
	Frame     = 1
	Close     = 2
	Ping      = 3
	Pong      = 4
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
func newClient(conn net.Conn) *Client {
	return &Client{
		// A buffer size of 1 causes Send() to block until the previous message
		// has an answer.
		out:  make(chan Payload, 1),
		in:   make(chan message),
		conn: conn,
	}
}

// Dial opens the Discord socket and returns a client for sending messages.
func Dial(tmpDir string) (c *Client, err error) {
	var conn net.Conn

	// Socket may be numbered from 0 to 9.
	for i := 0; i < 10; i++ {
		addr := getDiscordSocket(tmpDir, i)

		// Go's "unix" network is equivalent to AF_UNIX/SOCK_STREAM.
		conn, err = net.Dial("unix", addr)

		if err == nil {
			break
		}
	}

	if err != nil {
		return
	}

	c = newClient(conn)
	return
}

// Start waits for messages written with Send(), and sends them to the socket.
// It will also listen for any messages initiated by Discord.
func (c *Client) Start() (err error) {
	// nextOpcode is the opcode to use for the next packet sent via Send().
	// The first packet sent will be marked as a Handshake, and subsequent packets
	// will be marked as Frames.
	var nextOpcode int32 = Handshake

	// We want to handle events coming from either direction (reads from Discord
	// or writes from Send()).  Each iteration of the loop reads one message from
	// Discord, either if there is an unsolicited one waiting, or as a response
	// to Send().
EventLoop:
	for err == nil {
		readCh := make(chan messageResult)
		go func() {
			msg, err := readMessage(c.conn)
			readCh <- messageResult{msg, err}
		}()

		var r messageResult
		select {
		case payload, ok := <-c.out:
			// Got a Send() or Close().
			if !ok {
				break EventLoop
			}

			// Note that we have a concurrent conn.Write() and conn.Read() here, which
			// is okay according to the net.Conn docs.
			if err = writeMessage(message{Opcode: nextOpcode, Payload: payload}, c.conn); err != nil {
				break EventLoop
			}
			nextOpcode = Frame
			// Block on an answer from Discord.
			r = <-readCh
			err = r.err
			if r.err == nil {
				c.in <- r.msg
			}

		case r = <-readCh:
			err = r.err
			// Unsolicited message from Discord.
			switch {
			case err != nil:
				// Terminates loop.
			case r.msg.Opcode == Ping:
				// Respond immediately to ping.
				r.msg.Opcode = Pong
				err = writeMessage(r.msg, c.conn)
			case r.msg.Opcode == Close:
				err = fmt.Errorf("Discord IPC connection terminated by Discord")
			default:
				err = fmt.Errorf("Got unexpected opcode: %d, payload: %v", r.msg.Opcode, r.msg.Payload)
			}
		}
	}
	close(c.in)
	c.conn.Close()
	c.conn = nil
	return
}

// Send sends the given payload to Discord and returns the answer payload. It
// blocks on both waiting to send the message, and waiting for the answer to be
// received.
//
// The returned answer does not include the header used by Discord's IPC
// protocol; it's just the payload (typically JSON).
func (c *Client) Send(payload Payload) (ans Payload, err error) {
	c.out <- payload

	if m, ok := <-c.in; !ok {
		return ans, fmt.Errorf("Socket closed while waiting for response")
	} else {
		return m.Payload, nil
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
func readMessage(conn io.Reader) (m message, err error) {
	header := make([]byte, headerLen)
	var n int
	switch n, err = conn.Read(header); {
	case err != nil:
		return
	case n != headerLen:
		err = fmt.Errorf("Wanted %d-byte header, read %d bytes", headerLen, n)
		return
	}

	reader := bytes.NewReader(header)
	binary.Read(reader, binary.LittleEndian, &m.Opcode)
	var payloadLen int32
	binary.Read(reader, binary.LittleEndian, &payloadLen)

	payload := make([]byte, payloadLen)
	switch n, err = conn.Read(payload); {
	case err != nil:
		return
	case n != int(payloadLen):
		return m, fmt.Errorf("Wanted %d-byte payload, read %d bytes", payloadLen, n)
	}
	m.Payload = payload
	return
}

// writeMessage writes a message to the socket.
func writeMessage(m message, conn io.Writer) (err error) {
	buf := m.encode()
	var n int
	if n, err = conn.Write(buf); n != len(buf) {
		err = fmt.Errorf("Wanted to write %d bytes, wrote %d bytes", len(buf), n)
	}
	return
}
