package geoscpp

/*
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

func Union(g1, g2 geom.Geometry) (geom.Geometry, error) {
	g1WKB := g1.AsBinary()
	g2WKB := g2.AsBinary()
	var outputBuf, errBuf *C.char
	C.SF_GEOS_Union(
		(*C.char)(&g1WKB[0]),
		(*C.char)(&g2WKB[0]),
		&outputBuf,
		&errBuf,
	)

	// TODO: release memory for output and errBuf.

	if errBuf != nil {
		err := fmt.Errorf("%s", C.GoString(errBuf))
		return geom.Geometry{}, err
	}
	if outputBuf == nil {
		return geom.Geometry{}, fmt.Errorf("unexpected nil output from GEOS")
	}

	var outputLen C.int // TODO: populate
	outputWKB := C.GoBytes(unsafe.Pointer(outputBuf), outputLen)
	return geom.UnmarshalWKB(outputWKB)
}
