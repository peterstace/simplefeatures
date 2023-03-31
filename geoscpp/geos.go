package geoscpp

/*
#cgo LDFLAGS: -lgeos
#cgo CFLAGS: -Wall
#include "bridge_geos.h"
#include "stdlib.h"
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/peterstace/simplefeatures/geom"
)

func Version() string {
	v := C.sf_version()
	defer C.sf_delete_char_buffer(v)
	return C.GoString(v)
}

func Union(g1, g2 geom.Geometry) (geom.Geometry, error) {
	g1WKB := g1.AsBinary()
	g2WKB := g2.AsBinary()
	var outputBuf *C.uchar
	var outputSize C.size_t
	var errStr *C.char
	C.sf_union(
		(*C.uchar)(&g1WKB[0]),
		C.size_t(len(g1WKB)),
		(*C.uchar)(&g2WKB[0]),
		C.size_t(len(g2WKB)),
		&outputBuf,
		&outputSize,
		&errStr,
	)
	defer func() {
		C.sf_delete_uchar_buffer(outputBuf)
		C.sf_delete_char_buffer(errStr)
	}()

	if errStr != nil {
		err := fmt.Errorf("%s", C.GoString(errStr))
		return geom.Geometry{}, err
	}
	if outputBuf == nil {
		return geom.Geometry{}, fmt.Errorf("unexpected nil output from GEOS")
	}

	outputWKB := C.GoBytes(unsafe.Pointer(outputBuf), C.int(outputSize))
	return geom.UnmarshalWKB(outputWKB)
}
