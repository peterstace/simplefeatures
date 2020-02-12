package libgeos

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/peterstace/simplefeatures/geom"
)

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

// Handle is a handle into the libgeos C library. Handle is not threadsafe.  If
// libgeos needs to be used in a concurrent fashion, then multiple handles can
// be used.
type Handle struct {
	context   C.GEOSContextHandle_t
	wkbReader *C.GEOSWKBReader
	wkbWriter *C.GEOSWKBWriter
	wkbBuf    []byte
	errBuf    [1024]byte
}

// NewHandle creates a new handle.
func NewHandle() (*Handle, error) {
	h := &Handle{}
	h.context = C.sf_init(unsafe.Pointer(&h.errBuf))
	if h.context == nil {
		return nil, errors.New("could not create libgeos context")
	}
	h.wkbReader = C.GEOSWKBReader_create_r(h.context)
	if h.wkbReader == nil {
		return nil, h.err()
	}
	h.wkbWriter = C.GEOSWKBWriter_create_r(h.context)
	if h.wkbWriter == nil {
		return nil, h.err()
	}
	return h, nil
}

// Close cleans up memory resources associated with the handle. If Close is not
// called, then a memory leak will occurr.
func (h *Handle) Close() {
	C.GEOSWKBWriter_destroy_r(h.context, h.wkbWriter)
	C.GEOSWKBReader_destroy_r(h.context, h.wkbReader)
	C.GEOS_finish_r(h.context)
}

func (h *Handle) err() error {
	msg := h.errMsg()
	if msg == "" {
		// No error stored, which indicatse that the error handler didn't get
		// trigged. The best we can do is give a generic error.
		msg = "libgeos internal error"
	}
	h.errBuf = [1024]byte{} // Reset the buffer for the next error message.
	return errors.New(msg)
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

func (h *Handle) boolErr(c C.char) (bool, error) {
	switch c {
	case 0:
		return false, nil
	case 1:
		return true, nil
	case 2:
		return false, h.err()
	default:
		return false, fmt.Errorf("illegal result from libgeos: %v", c)
	}
}

func (h *Handle) intToErr(i C.int) error {
	switch i {
	case 0:
		return h.err()
	case 1:
		return nil
	default:
		return fmt.Errorf("illegal result from libgeos: %v", i)
	}
}

func (h *Handle) createGeomHandle(g geom.Geometry) (*C.GEOSGeometry, error) {
	wkb := bytes.NewBuffer(h.wkbBuf)
	if err := g.AsBinary(wkb); err != nil {
		return nil, err
	}
	h.wkbBuf = wkb.Bytes()
	gh := C.GEOSWKBReader_read_r(
		h.context,
		h.wkbReader,
		(*C.uchar)(&h.wkbBuf[0]),
		C.ulong(wkb.Len()),
	)
	h.wkbBuf = h.wkbBuf[:0]
	if gh == nil {
		return nil, h.err()
	}
	return gh, nil
}

func (h *Handle) decodeGeomHandle(gh *C.GEOSGeometry) (geom.Geometry, error) {
	hasEmptyPoint, err := h.containsEmptyPoint(gh)
	if err != nil {
		return geom.Geometry{}, err
	}
	if hasEmptyPoint {
		writer := C.GEOSWKTWriter_create_r(h.context)
		if writer == nil {
			return geom.Geometry{}, h.err()
		}
		defer C.GEOSWKTWriter_destroy_r(h.context, writer)
		C.GEOSWKTWriter_setTrim_r(h.context, writer, C.char(1))
		wkt := C.GEOSWKTWriter_write_r(h.context, writer, gh)
		if wkt == nil {
			return geom.Geometry{}, h.err()
		}
		defer C.GEOSFree_r(h.context, unsafe.Pointer(wkt))
		return geom.UnmarshalWKT(strings.NewReader(C.GoString(wkt)))
	} else {
		var size C.size_t
		wkb := C.GEOSWKBWriter_write_r(h.context, h.wkbWriter, gh, &size)
		if wkb == nil {
			return geom.Geometry{}, fmt.Errorf("writing wkb: %v", h.err())
		}
		defer C.GEOSFree_r(h.context, unsafe.Pointer(wkb))
		reader := bytes.NewReader(C.GoBytes(unsafe.Pointer(wkb), C.int(size)))
		return geom.UnmarshalWKB(reader)
	}
}

func (h *Handle) containsEmptyPoint(gh *C.GEOSGeometry) (bool, error) {
	geomType := C.GEOSGeomType_r(h.context, gh)
	if geomType == nil {
		return false, h.err()
	}
	defer C.free(unsafe.Pointer(geomType))
	switch C.GoString(geomType) {
	case "Point":
		isEmpty, err := h.boolErr(C.GEOSisEmpty_r(h.context, gh))
		if err != nil {
			return false, err
		}
		return isEmpty, nil
	case "MultiPoint", "GeometryCollection":
		n := C.GEOSGetNumGeometries_r(h.context, gh)
		if n == -1 {
			return false, h.err()
		}
		for i := 0; i < int(n); i++ {
			sub := C.GEOSGetGeometryN_r(h.context, gh, C.int(i))
			if sub == nil {
				return false, h.err()
			}
			if has, err := h.containsEmptyPoint(sub); err != nil {
				return false, err
			} else if has {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, nil
	}
}

func (h *Handle) AsText(g geom.Geometry) (string, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return "", err
	}
	defer C.GEOSGeom_destroy(gh)

	writer := C.GEOSWKTWriter_create_r(h.context)
	if writer == nil {
		return "", h.err()
	}
	defer C.GEOSWKTWriter_destroy_r(h.context, writer)
	C.GEOSWKTWriter_setTrim_r(h.context, writer, C.char(1))

	wkt := C.GEOSWKTWriter_write_r(h.context, writer, gh)
	if wkt == nil {
		return "", h.err()
	}
	defer C.GEOSFree_r(h.context, unsafe.Pointer(wkt))
	return C.GoString(wkt), nil
}

func (h *Handle) FromText(wkt string) (geom.Geometry, error) {
	reader := C.GEOSWKTReader_create_r(h.context)
	if reader == nil {
		return geom.Geometry{}, fmt.Errorf("creating wkt reader: %v", h.err())
	}
	defer C.GEOSWKTReader_destroy_r(h.context, reader)

	cwkt := C.CString(wkt)
	defer C.free(unsafe.Pointer(cwkt))

	gh := C.GEOSWKTReader_read_r(h.context, reader, cwkt)
	if gh == nil {
		return geom.Geometry{}, fmt.Errorf("reading: %v", h.err())
	}

	return h.decodeGeomHandle(gh)
}

func (h *Handle) AsBinary(g geom.Geometry) ([]byte, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return nil, err
	}
	defer C.GEOSGeom_destroy(gh)

	writer := C.GEOSWKBWriter_create_r(h.context)
	if writer == nil {
		return nil, h.err()
	}
	defer C.GEOSWKBWriter_destroy_r(h.context, writer)
	var size C.size_t
	wkb := C.GEOSWKBWriter_write_r(h.context, writer, gh, &size)
	if wkb == nil {
		return nil, h.err()
	}
	defer C.GEOSFree_r(h.context, unsafe.Pointer(wkb))
	return C.GoBytes(unsafe.Pointer(wkb), C.int(size)), nil
}

func (h *Handle) FromBinary(wkb []byte) (geom.Geometry, error) {
	gh := C.GEOSWKBReader_read_r(
		h.context,
		h.wkbReader,
		(*C.uchar)(&wkb[0]),
		C.ulong(len(wkb)),
	)
	if gh == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, gh)
	return h.decodeGeomHandle(gh)
}

func (h *Handle) IsEmpty(g geom.Geometry) (bool, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return false, err
	}
	defer C.GEOSGeom_destroy(gh)

	return h.boolErr(C.GEOSisEmpty_r(h.context, gh))
}

func (h *Handle) Dimension(g geom.Geometry) (int, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return 0, err
	}
	defer C.GEOSGeom_destroy(gh)

	dim := int(C.GEOSGeom_getDimensions_r(h.context, gh))
	if h.errMsg() != "" {
		return 0, h.err()
	}
	return dim, nil
}

func (h *Handle) Envelope(g geom.Geometry) (geom.Envelope, bool, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return geom.Envelope{}, false, err
	}
	defer C.GEOSGeom_destroy(gh)

	env := C.GEOSEnvelope_r(h.context, gh)
	if env == nil {
		return geom.Envelope{}, false, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, env)

	if isEmpty, err := h.boolErr(C.GEOSisEmpty_r(h.context, env)); err != nil {
		return geom.Envelope{}, false, err
	} else if isEmpty {
		return geom.Envelope{}, false, nil
	}

	// libgeos will return either a Point or a Polygon. In the case where the
	// envelope has a width but no height or a height but no width, then an
	// invalid Polygon is returned.
	geomType := C.GEOSGeomType_r(h.context, env)
	if geomType == nil {
		return geom.Envelope{}, false, h.err()
	}
	defer C.free(unsafe.Pointer(geomType))
	if C.GoString(geomType) == "Point" {
		var x C.double
		if C.GEOSGeomGetX_r(h.context, env, &x) == 0 {
			return geom.Envelope{}, false, h.err()
		}
		var y C.double
		if C.GEOSGeomGetY_r(h.context, env, &y) == 0 {
			return geom.Envelope{}, false, h.err()
		}
		return geom.NewEnvelope(geom.XY{X: float64(x), Y: float64(y)}), true, nil
	}

	ring := C.GEOSGetExteriorRing_r(h.context, env)
	if ring == nil {
		return geom.Envelope{}, false, h.err()
	}
	// ring belongs to env, so doesn't need to be destroyed.

	seq := C.GEOSGeom_getCoordSeq_r(h.context, ring)
	if seq == nil {
		return geom.Envelope{}, false, h.err()
	}
	// seq belongs to ring, so doesn't need to be destroyed.

	var size C.uint
	if C.GEOSCoordSeq_getSize_r(h.context, seq, &size) == 0 {
		return geom.Envelope{}, false, h.err()
	}
	if size == 0 {
		return geom.Envelope{}, false, errors.New(
			"coordinate sequence doesn't contain any points")
	}

	var sfEnv geom.Envelope
	for i := C.uint(0); i < size; i++ {
		var x C.double
		if C.GEOSCoordSeq_getX_r(h.context, seq, i, &x) == 0 {
			return geom.Envelope{}, false, h.err()
		}
		var y C.double
		if C.GEOSCoordSeq_getY_r(h.context, seq, i, &y) == 0 {
			return geom.Envelope{}, false, h.err()
		}
		xy := geom.XY{X: float64(x), Y: float64(y)}
		if i == 0 {
			sfEnv = geom.NewEnvelope(xy)
		} else {
			sfEnv = sfEnv.ExtendToIncludePoint(xy)
		}
	}

	return sfEnv, true, nil
}

func (h *Handle) IsSimple(g geom.Geometry) (isSimple bool, defined bool, err error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return false, false, err
	}
	defer C.GEOSGeom_destroy(gh)

	// IsSimple is not defined for GeometryCollections.
	geomType := C.GEOSGeomType_r(h.context, gh)
	if geomType == nil {
		return false, false, h.err()
	}
	defer C.free(unsafe.Pointer(geomType))
	if C.GoString(geomType) == "GeometryCollection" {
		return false, false, nil
	}

	isSimple, err = h.boolErr(C.GEOSisSimple_r(h.context, gh))
	return isSimple, true, err
}

func (h *Handle) Boundary(g geom.Geometry) (geom.Geometry, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return geom.Geometry{}, err
	}
	defer C.GEOSGeom_destroy(gh)

	env := C.GEOSBoundary_r(h.context, gh)
	if env == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, env)

	return h.decodeGeomHandle(env)
}

func (h *Handle) ConvexHull(g geom.Geometry) (geom.Geometry, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return geom.Geometry{}, err
	}
	defer C.GEOSGeom_destroy(gh)

	env := C.GEOSConvexHull_r(h.context, gh)
	if env == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, env)

	return h.decodeGeomHandle(env)
}

func (h *Handle) IsValid(g geom.Geometry) (bool, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return false, err
	}
	defer C.GEOSGeom_destroy(gh)

	return h.boolErr(C.GEOSisValid_r(h.context, gh))
}

func (h *Handle) IsRing(g geom.Geometry) (bool, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return false, err
	}
	defer C.GEOSGeom_destroy(gh)

	return h.boolErr(C.GEOSisRing_r(h.context, gh))
}

func (h *Handle) Length(g geom.Geometry) (float64, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return 0, err
	}
	defer C.GEOSGeom_destroy(gh)

	var length float64
	errInt := C.GEOSLength_r(h.context, gh, (*C.double)(&length))
	return length, h.intToErr(errInt)
}

func (h *Handle) Area(g geom.Geometry) (float64, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return 0, err
	}
	defer C.GEOSGeom_destroy(gh)

	var area float64
	errInt := C.GEOSArea_r(h.context, gh, (*C.double)(&area))
	return area, h.intToErr(errInt)
}

func (h *Handle) Centroid(g geom.Geometry) (geom.Geometry, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return geom.Geometry{}, err
	}
	defer C.GEOSGeom_destroy(gh)

	env := C.GEOSGetCentroid_r(h.context, gh)
	if env == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, env)

	return h.decodeGeomHandle(env)
}

func (h *Handle) Reverse(g geom.Geometry) (geom.Geometry, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return geom.Geometry{}, err
	}
	defer C.GEOSGeom_destroy(gh)

	env := C.GEOSReverse_r(h.context, gh)
	if env == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, env)

	return h.decodeGeomHandle(env)
}
