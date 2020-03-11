package libgeos

/*
#cgo linux CFLAGS: -I/usr/include
#cgo linux LDFLAGS: -L/usr/lib -lgeos_c
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

*/
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/peterstace/simplefeatures/geom"
)

type Handle struct {
	context C.GEOSContextHandle_t
	errBuf  [1024]byte
}

func NewHandle() (*Handle, error) {
	h := &Handle{}
	h.context = C.sf_init(unsafe.Pointer(&h.errBuf))
	if h.context == nil {
		return nil, errors.New("could not create libgeos context")
	}
	return h, nil
}

func (h *Handle) err() error {
	msg := h.errMsg()
	if msg == "" {
		// No error stored, which indicates that the error handler didn't get
		// trigged. The best we can do is give a generic error.
		msg = "libgeos internal error"
	}
	h.errBuf = [1024]byte{} // Reset the buffer for the next error message.
	return errors.New(strings.TrimSpace(msg))
}

func (h *Handle) errMsg() string {
	// The error message is either NULL terminated, or fills the entire buffer.
	firstZero := len(h.errBuf)
	for i, b := range h.errBuf {
		if b == 0 {
			firstZero = i
			break
		}
	}
	return string(h.errBuf[:firstZero])
}

func (h *Handle) Release() {
	C.GEOS_finish_r(h.context)
}

func (h *Handle) createGeometryHandle(g geom.Geometry) (*C.GEOSGeometry, error) {
	wkbReader := C.GEOSWKBReader_create_r(h.context)
	if wkbReader == nil {
		return nil, h.err()
	}
	defer C.GEOSWKBReader_destroy_r(h.context, wkbReader)

	wkb := new(bytes.Buffer)
	if err := g.AsBinary(wkb); err != nil {
		return nil, err
	}
	gh := C.GEOSWKBReader_read_r(
		h.context,
		wkbReader,
		(*C.uchar)(&wkb.Bytes()[0]),
		C.ulong(wkb.Len()),
	)
	if gh == nil {
		return nil, h.err()
	}
	return gh, nil
}

func (h *Handle) Equals(g1, g2 geom.Geometry) (bool, error) {
	return h.relate(g1, g2, "T*F**FFF*")
}

func (h *Handle) relate(g1, g2 geom.Geometry, mask string) (bool, error) {
	gh1, err := h.createGeometryHandle(g1)
	if err != nil {
		return false, err
	}
	defer C.GEOSGeom_destroy(gh1)

	gh2, err := h.createGeometryHandle(g2)
	if err != nil {
		return false, err
	}

	defer C.GEOSGeom_destroy(gh2)
	cmask := C.CString(mask)
	defer C.free(unsafe.Pointer(cmask))

	switch ret := C.GEOSRelatePattern_r(h.context, gh1, gh2, cmask); ret {
	case 2:
		return false, h.err()
	case 1:
		return true, nil
	case 0:
		return false, nil
	default:
		return false, fmt.Errorf("unexpeted return code: %d", ret)
	}
}
