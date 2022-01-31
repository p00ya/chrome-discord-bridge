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

func TestHost(t *testing.T) {
	in, inPipe := io.Pipe()
	outPipe, out := io.Pipe()
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
			inPipe.Write(requestWire)
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
			buf := make([]byte, len(responseWire))
			_, err := io.ReadFull(outPipe, buf)

			if err != nil {
				errors <- fmt.Errorf("Wanted response, got %v", err)
			}
			if !bytes.Equal(buf, responseWire) {
				errors <- fmt.Errorf("Wanted write %v, got %v", responseWire, buf)
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
		bufSize := 4

		// Simulate Chrome writing a request.
		go func() {
			var i int
			for i = 0; i+bufSize < len(requestWire); i += bufSize {
				inPipe.Write(requestWire[i : i+bufSize])
			}
			inPipe.Write(requestWire[i:])

			done <- 1
		}()

		go func() {
			_, responder := host.Receive()
			responder.Respond(responsePayload)
			done <- 1
		}()

		// Simulate Chrome reading the response.
		go func() {
			buf := make([]byte, len(responseWire))
			switch _, err := io.ReadFull(outPipe, buf); {
			case err != nil:
				errors <- fmt.Errorf("Wanted response, got %v", err)
			case !bytes.Equal(buf, responseWire):
				errors <- fmt.Errorf("Wanted write %v, got %v", responseWire, buf)
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
		done := make(chan int)
		go func() {
			buf := make([]byte, 1)
			switch n, err := outPipe.Read(buf); {
			case n > 0:
				t.Errorf("Got unexpected buf %v", buf[:n])
			case err != io.EOF && err != io.ErrClosedPipe:
				t.Errorf("Got unexpected err %v, wanted EOF", err)
			}
			done <- 1
		}()

		host.Close()
		select {
		case <-done:
			// Good.
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
