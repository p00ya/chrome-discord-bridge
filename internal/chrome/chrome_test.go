package chrome

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"
)

// Number of seconds to wait for things that should be near-instantaneous.
const timeoutSeconds = 2

// fakeReader is an implementation of io.Reader for testing.
type fakeReader struct {
	// ReadCh is packets to be read with Read().
	ReadCh chan []byte

	// BufSize is the maximum number of bytes that can be written at once, or 0
	// for no limit.
	BufSize int

	// readBuf contains any unread bytes from the current packet.
	readBuf []byte
}

func (fake *fakeReader) Read(b []byte) (n int, err error) {
	if len(fake.readBuf) == 0 {
		var ok bool
		fake.readBuf, ok = <-fake.ReadCh
		if !ok {
			return 0, io.EOF
		}
	}
	n = copy(b, fake.readBuf)
	fake.readBuf = fake.readBuf[n:]
	return
}

// fakeWriter is an implementation of io.WriteCloser for testing.
type fakeWriter struct {
	// WriteCh is packets that have been written with Write().
	WriteCh chan []byte

	// BufSize is the maximum number of bytes that can be written at once, or 0
	// for no limit.
	BufSize int
}

func (fake *fakeWriter) Write(p []byte) (n int, err error) {
	if fake.BufSize > 0 && fake.BufSize < len(p) {
		p = p[:fake.BufSize]
	}

	fake.WriteCh <- p
	return len(p), nil
}

func (fake *fakeWriter) Close() error {
	close(fake.WriteCh)
	return nil
}

func TestHost(t *testing.T) {
	in := &fakeReader{
		ReadCh: make(chan []byte),
	}
	out := &fakeWriter{
		WriteCh: make(chan []byte),
		BufSize: 0,
	}
	host := NewHost(in, out)

	startDone := make(chan error)
	go func() {
		startDone <- host.Start()
	}()

	requestPayload := []byte(`{"request":1}`)
	requestWire := []byte(`....{"request":1}`)
	nativeEndian.PutUint32(requestWire[:4], uint32(len(requestPayload)))

	responsePayload := []byte(`{"response":1}`)
	responseWire := []byte(`....{"response":1}`)
	nativeEndian.PutUint32(responseWire[:4], uint32(len(responsePayload)))

	t.Run("Receive", func(t *testing.T) {
		done := make(chan int)
		errors := make(chan error)

		// Simulate Chrome writing a request.
		go func() {
			in.ReadCh <- requestWire
			done <- 1
		}()

		// Test calling Receive() and Respond().
		go func() {
			request, responder := host.Receive()
			if !bytes.Equal(request, requestPayload) {
				errors <- fmt.Errorf("Wanted request %v, got %v", requestPayload, request)
			}
			responder.Respond(responsePayload)
			done <- 1
		}()

		// Simulate Chrome reading the response.
		go func() {
			response, ok := <-out.WriteCh

			if !ok {
				errors <- fmt.Errorf("Wanted response, got closed channel")
			}
			if !bytes.Equal(response, responseWire) {
				errors <- fmt.Errorf("Wanted write %v, got %v", responseWire, response)
			}
			done <- 1
		}()

		for n := 3; n > 0; {
			select {
			case <-done:
				n--
			case err := <-errors:
				t.Error(err)
			case <-time.After(timeoutSeconds * time.Second):
				t.Fatalf("Timeout, still waiting on %d goroutines", n)
			}
		}
	})

	// Test the writer having a buffer smaller than the response.
	t.Run("PartialReadWrites", func(t *testing.T) {
		done := make(chan int)
		errors := make(chan error)
		in.ReadCh = make(chan []byte, 1)
		in.BufSize = 4
		out.BufSize = 4

		// Simulate Chrome writing a request.
		go func() {
			var i int
			for i = 0; i+in.BufSize < len(requestWire); i += in.BufSize {
				in.ReadCh <- requestWire[i : i+in.BufSize]
			}
			in.ReadCh <- requestWire[i:]

			done <- 1
		}()

		go func() {
			_, responder := host.Receive()
			responder.Respond(responsePayload)
			done <- 1
		}()

		// Simulate Chrome reading the response.
		go func() {
			for i := 0; i < len(responseWire); i += out.BufSize {
				partialWrite, ok := <-out.WriteCh

				if !ok {
					errors <- fmt.Errorf("Wanted response, got closed channel")
					break
				}
				end := i + out.BufSize
				if end > len(responseWire) {
					end = len(responseWire)
				}
				partialExpected := responseWire[i:end]
				if !bytes.Equal(partialWrite, partialExpected) {
					errors <- fmt.Errorf("Wanted write %v, got %v", partialExpected, partialWrite)
					break
				}
			}
			done <- 1
		}()

		for n := 3; n > 0; {
			select {
			case <-done:
				n--
			case err := <-errors:
				t.Error(err)
			case <-time.After(timeoutSeconds * time.Second):
				t.Fatalf("Timeout, still waiting on %d goroutines", n)
			}
		}
	})

	t.Run("Close", func(t *testing.T) {
		host.Close()
		select {
		case _, ok := <-out.WriteCh:
			if ok {
				t.Errorf("Got response, expected output file to be closed")
			}
		case <-time.After(timeoutSeconds * time.Second):
			t.Fatal("Timeout")
		}
	})

	t.Run("Start", func(t *testing.T) {
		select {
		case err := <-startDone:
			if err != nil {
				t.Errorf("Start() returned error %v, expected nil", err)
			}
		case <-time.After(timeoutSeconds * time.Second):
			t.Fatal("Timeout waiting for Start() to return")
		}
	})
}
