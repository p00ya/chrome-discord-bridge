// Package chrome provides a Chrome native messaging host.
package chrome

import (
	"fmt"
	"io"
)

// Host manages the I/O for a Chrome native messaging host.
//
// Create a new Host with NewHost().  Run Start() to start the event loop.
// Read and respond to messages from Chrome, one at a time, with Receive().
// Terminate the connection with Close().
type Host struct {
	// in receives payloads that were received from Chrome.
	// Only written to by Start(), only read by Receive().
	in chan []byte

	// out receives payloads that should be sent to Chrome.
	// Only written to by Respond(), only read by Start().
	out chan []byte

	// closed receives a value when the connection to Chrome should be shut down.
	// Only written to by Close(), only read by Start().
	closed chan bool

	// inFile is the file for reading from Chrome (typically stdin).
	// Only read by (an anonymous goroutine spawned by) Start().
	reader io.Reader

	// outFile is the file for writing to Chrome (typically stdout).
	// Only written to by Start().
	writer io.WriteCloser
}

// headerLen is the number of bytes in the Chrome native messaging header.
const headerLen = 4

// maxPayloadBytes in the maximum length in bytes of a message from Chrome (not
// including the 4-byte header).
const maxPayloadBytes = 4096

// NewHost returns a Chrome native messaging host that will read requests from
// the given reader, and send responses on the given writer.  The I/O must
// not be buffered.
func NewHost(in io.Reader, out io.WriteCloser) *Host {
	return &Host{
		in:     make(chan []byte),
		out:    make(chan []byte),
		closed: make(chan bool),
		reader: in,
		writer: out,
	}
}

// Start begins listening for messages from Chrome, and will return them via
// Read().  It then waits for a response via Send(), and will send the response
// to Chrome.
//
// Start returns if there was an error or if Close() was called.  It will close
// the writer the host was created with, but not the reader.
func (host *Host) Start() (err error) {
	readerCh := make(chan []byte)

	// Read messages from host.reader and send them to readerCh.
	// This goroutine has exclusive access to host.reader.
	go func(reader io.Reader) {
		for {
			buf, err := readPayload(reader)
			if err != nil || len(buf) == 0 {
				break
			}
			readerCh <- buf
		}
		close(readerCh)
	}(host.reader)

EventLoop:
	for {
		select {
		case request, ok := <-readerCh:
			if !ok {
				// Chrome-initiated shutdown.
				break EventLoop
			}
			host.in <- request
		case <-host.closed:
			// Client-initiated shutdown.
			break EventLoop
		}

		// Don't read more messages from Chrome until we've responded.
		response := <-host.out
		if err = writePayload(response, host.writer); err != nil {
			break EventLoop
		}
	}
	close(host.in)
	host.writer.Close()
	return
}

// readPayload returns a Chrome native messaging payload read from the given
// file.
func readPayload(in io.Reader) (payload []byte, err error) {
	header := make([]byte, headerLen)
	var n int
	switch n, err = in.Read(header); {
	case n == 0 || err == io.EOF:
		// Clean shutdown from Chrome's end.
		return nil, io.EOF
	case err != nil:
		return
	case n != headerLen:
		err = fmt.Errorf("Wanted %d-byte header, read %d bytes", headerLen, n)
		return
	}

	payloadLen := nativeEndian.Uint32(header)
	if payloadLen > maxPayloadBytes {
		err = fmt.Errorf("Want at most %d-byte payload, got %d", maxPayloadBytes, payloadLen)
	}

	payload = make([]byte, payloadLen)
	if n, err = in.Read(payload); n != int(payloadLen) {
		err = fmt.Errorf("Wanted %d-byte payload, read %d bytes", payloadLen, n)
	}
	return
}

// writePayload sends a Chrome native messaging payload to Chrome.
func writePayload(payload []byte, out io.Writer) (err error) {
	buf := make([]byte, headerLen+len(payload))
	nativeEndian.PutUint32(buf[0:headerLen], uint32(len(payload)))
	copy(buf[headerLen:], payload)

	var n int
	if n, err = out.Write(buf); n != len(buf) {
		err = fmt.Errorf("Wanted to write %d bytes, wrote %d bytes", len(buf), n)
	}
	return
}

// Responder is an abstraction for responding to a request from Chrome.
type Responder struct {
	// response receives a payload for sending to Chrome.
	response chan []byte
}

// Respond sends the given response payload to Chrome.  Must be called exactly
// once.
func (r Responder) Respond(response []byte) {
	r.response <- response
}

// Receive blocks on receiving one message from Chrome, and then returns the
// request payload and a Responder object.  The caller must call the Respond()
// method on the returned object exactly once (and before calling Receive
// again), which will forward the response to Chrome.
func (host *Host) Receive() (request []byte, responder *Responder) {
	request = <-host.in
	responder = &Responder{response: host.out}
	return
}

// Close terminates the event loop and indicates that no more messages will
// be processed.
func (host *Host) Close() {
	host.closed <- true
}
