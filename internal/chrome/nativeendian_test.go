// go:build amd64

package chrome

import (
	"bytes"
	"fmt"
	"testing"
)

// These tests assume a little-endian architecture.
func TestNativeEndian(t *testing.T) {
	var tests = []struct {
		raw []byte
		n   uint32
	}{
		{[]byte("\x01\x00\x00\x00"), 1},
		{[]byte("\x00\x00\x00\x01"), 0x1000000},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.n)
		t.Run(testname, func(t *testing.T) {
			i := nativeEndian.Uint32(tt.raw)
			if i != tt.n {
				t.Errorf("ReadUint32() got %d, want %d", i, tt.n)
			}

			rt := make([]byte, 4)
			nativeEndian.PutUint32(rt, i)
			if !bytes.Equal(rt, tt.raw) {
				t.Errorf("PutUint32() got %v, want %v", rt, tt.raw)
			}
		})
	}
}
