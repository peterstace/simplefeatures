package geos

/*
#cgo LDFLAGS: -lgeos_c
#cgo CFLAGS: -Wall
#include "geos_c.h"
#include <stdlib.h>
#include <string.h>

void sf_error_handler(const char *message, void *userdata) {
	strncpy(userdata, message, 1024);
}

GEOSContextHandle_t sf_init(void *userdata) {
	GEOSContextHandle_t ctx = GEOS_init_r();
	GEOSContext_setErrorMessageHandler_r(ctx, sf_error_handler, userdata);
	return ctx;
}

char *marshal(GEOSContextHandle_t handle, const GEOSGeometry *g, size_t *size, char *isWKT);

GEOSGeometry const *noop(GEOSContextHandle_t handle, const GEOSGeometry *g) {
	return g;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/peterstace/simplefeatures/geom"
)

// noop returns the geometry unaltered, via conversion to and from GEOS. This
// function is only for benchmarking purposes, hence it is not exported or used
// outside of benchmark tests.
func noop(g geom.Geometry) (geom.Geometry, error) {
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.noop(ctx, g)
	})
	return result, wrap(err, "executing noop")
}

// handle is an opaque handle that can be used to invoke GEOS operations.
// Instances are not threadsafe and thus should only be used serially (e.g.
// protected by a mutex or similar).
type handle struct {
	context C.GEOSContextHandle_t
	reader  *C.GEOSWKBReader
	errBuf  *C.char
}

func newHandle() (*handle, error) {
	h := &handle{}

	h.errBuf = (*C.char)(C.malloc(1024))
	if h.errBuf == nil {
		h.release()
		return nil, errors.New("malloc failed")
	}
	C.memset((unsafe.Pointer)(h.errBuf), 0, 1024)

	h.context = C.sf_init(unsafe.Pointer(h.errBuf))
	if h.context == nil {
		h.release()
		return nil, errors.New("could not create GEOS context")
	}

	h.reader = C.GEOSWKBReader_create_r(h.context)
	if h.reader == nil {
		h.release()
		return nil, wrap(h.err(), "creating GEOS WKB reader")
	}

	return h, nil
}

// err gets the last error message reported by GEOS as an error type. It
// always returns a non-nil error. If no error message has been reported, then
// it returns a generic error message.
func (h *handle) err() error {
	msg := h.errMsg()
	if msg == "" {
		// No error stored, which indicates that the error handler didn't get
		// trigged. The best we can do is give a generic error.
		msg = "missing error message"
	}
	C.memset((unsafe.Pointer)(h.errBuf), 0, 1024) // Reset the buffer for the next error message.
	return fmt.Errorf("GEOS internal error: %v", strings.TrimSpace(msg))
}

// errMsg gets the textual representation of the last error message reported by
// GEOS.
func (h *handle) errMsg() string {
	// The error message is either NULL terminated, or fills the entire buffer.
	buf := C.GoBytes((unsafe.Pointer)(h.errBuf), 1024)
	for i, b := range buf {
		if b == 0 {
			return string(buf[:i])
		}
	}
	return string(buf[:])
}

// release releases any resources held by the handle. The handle should not be
// used after release is called.
func (h *handle) release() {
	if h.reader != nil {
		C.GEOSWKBReader_destroy_r(h.context, h.reader)
		h.reader = (*C.GEOSWKBReader)(C.NULL)
	}
	if h.context != nil {
		C.GEOS_finish_r(h.context)
		h.context = C.GEOSContextHandle_t(C.NULL)
	}
	if h.errBuf != nil {
		C.free((unsafe.Pointer)(h.errBuf))
		h.errBuf = (*C.char)(C.NULL)
	}
}

// createGeometryHandle converts a Geometry object into a GEOS geometry handle.
func (h *handle) createGeometryHandle(g geom.Geometry) (*C.GEOSGeometry, error) {
	wkb := g.AsBinary()
	gh := C.GEOSWKBReader_read_r(
		h.context,
		h.reader,
		(*C.uchar)(&wkb[0]),
		C.ulong(len(wkb)),
	)
	if gh == nil {
		return nil, wrap(h.err(), "executing GEOSWKBReader_read_r")
	}
	return gh, nil
}

// boolErr converts a char result from GEOS into a boolean result.
func (h *handle) boolErr(c C.char) (bool, error) {
	const (
		// From geos_c.h:
		// return 2 on exception, 1 on true, 0 on false.
		relateException = 2
		relateTrue      = 1
		relateFalse     = 0
	)
	switch c {
	case 0:
		return false, nil
	case 1:
		return true, nil
	case 2:
		return false, h.err()
	default:
		return false, wrap(h.err(), "illegal result %v from GEOS", c)
	}
}

func (h *handle) decode(gh *C.GEOSGeometry) (geom.Geometry, error) {
	var (
		isWKT C.char
		size  C.size_t
	)
	serialised := C.marshal(h.context, gh, &size, &isWKT)
	if serialised == nil {
		return geom.Geometry{}, wrap(h.err(), "marshalling result")
	}
	defer C.GEOSFree_r(h.context, unsafe.Pointer(serialised))

	if isWKT != 0 {
		wkt := C.GoStringN(serialised, C.int(size))
		g, err := geom.UnmarshalWKTWithoutValidation(wkt)
		return g, wrap(err, "failed to unmarshal GEOS WKT result")
	}
	wkb := C.GoBytes(unsafe.Pointer(serialised), C.int(size))
	g, err := geom.UnmarshalWKBWithoutValidation(wkb)
	return g, wrap(err, "failed to unmarshal GEOS WKB result")
}
