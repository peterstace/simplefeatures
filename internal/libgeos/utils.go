package libgeos

import "C"
import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

func boolErr(c C.char) (bool, error) {
	switch c {
	case 0:
		return false, nil
	case 1:
		return true, nil
	case 2:
		return false, errors.New("an exception occurred")
	default:
		return false, fmt.Errorf("illegal result from libgeos: %v", c)
	}
}

func intToErr(i C.int) error {
	switch i {
	case 0:
		return errors.New("an exception occurred")
	case 1:
		return nil
	default:
		return fmt.Errorf("illegal result from libgeos: %v", i)
	}
}

func copyBytes(byts *C.uchar, size C.size_t) []byte {
	src := cBytesAsSlice(byts, size)
	dest := make([]byte, size)
	copy(dest, src)
	return dest
}

func cBytesAsSlice(byts *C.uchar, size C.size_t) []byte {
	var slice []byte
	ptr := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	ptr.Data = uintptr(unsafe.Pointer(byts))
	ptr.Len = int(size)
	ptr.Cap = int(size)
	return slice
}
