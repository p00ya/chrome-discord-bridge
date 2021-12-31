package chrome

import (
	"encoding/binary"
	"unsafe"
)

// nativeEndian is the runtime native byte order.
var nativeEndian binary.ByteOrder

func init() {
	// Initialize nativeEndian.
	var i int32 = 1
	b := (*byte)(unsafe.Pointer(&i))
	if *b == 0 {
		nativeEndian = binary.BigEndian
	} else {
		nativeEndian = binary.LittleEndian
	}
}
