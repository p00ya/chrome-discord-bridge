package discord

import (
	"bytes"
	"testing"
	"time"
)

// Number of seconds to wait for things that should be near-instantaneous.
const timeoutSeconds = 2

func TestMessageEncode(t *testing.T) {
	m := message{
		Opcode:  1,
		Payload: []byte(`{}`),
	}
	want := []byte("\x01\x00\x00\x00\x02\x00\x00\x00{}")
	got := m.encode()
	if !bytes.Equal(want, got) {
		t.Errorf("Wanted %v, got %v", want, got)
	}
}

// fakeConn is a channel-driven implementation of io.ReadWriteCloser for
// testing.
type fakeConn struct {
	// WriteCh is sent packets from Write().
	WriteCh chan []byte
	// ReadCh is sent packets to be read with Read().
	ReadCh chan []byte
	// readBuf contains any unread bytes from the current packet.
	readBuf []byte
}

func NewFakeConn() *fakeConn {
	return &fakeConn{
		WriteCh: make(chan []byte),
		ReadCh:  make(chan []byte),
	}
}

func (f *fakeConn) Read(b []byte) (n int, err error) {
	if len(f.readBuf) == 0 {
		f.readBuf = <-f.ReadCh
	}
	n = copy(b, f.readBuf)
	f.readBuf = f.readBuf[n:]
	return
}

func (f *fakeConn) Write(b []byte) (int, error) {
	f.WriteCh <- b
	return len(b), nil
}

func (f *fakeConn) Close() error {
	close(f.WriteCh)
	return nil
}

func TestClientSend(t *testing.T) {
	fakeConn := NewFakeConn()
	client := newClient(fakeConn)

	startDone := make(chan error)
	go func() {
		startDone <- client.Start()
	}()

	type sendData struct {
		opcode                          int32
		requestPayload, responsePayload Payload
		wireRequest                     []byte
		wireResponse                    [][]byte
	}

	// testSend returns a test subcase function that verifies a message is
	// sent and the appropriate answer is received.
	testSend := func(data sendData) func(t *testing.T) {
		done := make(chan int)

		go func() {
			ans, err := client.Send(data.requestPayload)

			if err != nil {
				t.Error(err)
			}
			if !bytes.Equal(ans, data.responsePayload) {
				t.Errorf("client wanted answer `%s`, got `%s`", data.responsePayload, ans)
			}
			done <- 1
		}()

		go func() {
			buf := <-fakeConn.WriteCh

			if !bytes.Equal(buf, data.wireRequest) {
				t.Errorf("server wanted %v, got %v", data.wireRequest, buf)
			}

			for _, packet := range data.wireResponse {
				fakeConn.ReadCh <- packet
			}
			done <- 1
		}()

		return func(t *testing.T) {
			// Wait for both the client and server goroutines to finish.
			for n := 2; n > 0; {
				select {
				case <-done:
					n--
				case <-time.After(timeoutSeconds * time.Second):
					t.Fatal("Timeout")
					n = 0
				}
			}
		}
	}

	t.Run("SendHandshake", testSend(sendData{
		opcode:          0,
		requestPayload:  []byte(`{"cmd":"TEST"}`),
		responsePayload: []byte(`{"cmd":"TEST","reply":1}`),
		wireRequest:     []byte("\x00\x00\x00\x00\x0e\x00\x00\x00" + `{"cmd":"TEST"}`),
		wireResponse:    [][]byte{[]byte("\x00\x00\x00\x00\x18\x00\x00\x00" + `{"cmd":"TEST","reply":1}`)},
	}))

	t.Run("SendFrame", testSend(sendData{
		opcode:          1,
		requestPayload:  []byte(`{"cmd":"TEST"}`),
		responsePayload: []byte(`{"cmd":"TEST","reply":1}`),
		wireRequest:     []byte("\x01\x00\x00\x00\x0e\x00\x00\x00" + `{"cmd":"TEST"}`),
		wireResponse:    [][]byte{[]byte("\x01\x00\x00\x00\x18\x00\x00\x00" + `{"cmd":"TEST","reply":1}`)},
	}))

	t.Run("SplitResponse", testSend(sendData{
		opcode:          1,
		requestPayload:  []byte(`{"cmd":"TEST","nonce":2}`),
		responsePayload: []byte(`{"cmd":"TEST","reply":2}`),
		wireRequest:     []byte("\x01\x00\x00\x00\x18\x00\x00\x00" + `{"cmd":"TEST","nonce":2}`),
		wireResponse: [][]byte{
			[]byte("\x01\x00\x00\x00\x18\x00\x00\x00"),
			[]byte(`{"cmd":"TEST"`),
			[]byte(`,"reply":2}`),
		},
	}))

	t.Run("Close", func(t *testing.T) {
		client.Close()
		select {
		case err := <-startDone:
			if err != nil {
				t.Errorf("Start() returned error %v, wanted nil", err)
			}
		case <-time.After(timeoutSeconds * time.Second):
			t.Fatal("Timeout waiting for Start() to return")
		}
	})
}

func TestClientServerInitiated(t *testing.T) {
	fakeConn := NewFakeConn()
	client := newClient(fakeConn)

	startDone := make(chan error)
	go func() {
		startDone <- client.Start()
	}()

	t.Run("Ping", func(t *testing.T) {
		done := make(chan int)

		go func() {
			fakeConn.ReadCh <- []byte("\x03\x00\x00\x00\x0d\x00\x00\x00" + `{"nonce":"1"}`)
			response := []byte("\x04\x00\x00\x00\x0d\x00\x00\x00" + `{"nonce":"1"}`)

			buf := <-fakeConn.WriteCh
			if !bytes.Equal(buf, response) {
				t.Errorf("server wanted %v, got %v", response, buf)
			}

			done <- 1
		}()

		select {
		case <-done:
			// Good.
		case <-time.After((1 + timeoutSeconds) * time.Second):
			t.Fatal("Timeout")
		}
	})

	t.Run("Close", func(t *testing.T) {
		done := make(chan int)

		go func() {
			fakeConn.ReadCh <- []byte("\x02\x00\x00\x00\x17\x00\x00\x00" + `{"code":1,"message":""}`)
			done <- 1
		}()

		select {
		case <-done:
			// Good.
		case <-time.After(timeoutSeconds * time.Second):
			t.Fatal("Timeout waiting for Start() to return")
		}

		select {
		case err := <-startDone:
			if err == nil {
				t.Errorf("wanted Start() to return error, got nil")
			}
		case <-time.After(timeoutSeconds * time.Second):
			t.Fatal("Timeout waiting for Start() to return")
		}
	})
}
