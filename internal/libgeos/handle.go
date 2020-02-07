package libgeos

import (
	"bytes"
	"errors"
	"unsafe"

	"github.com/peterstace/simplefeatures/geom"
)

/*
#cgo linux CFLAGS: -I/usr/include
#cgo linux LDFLAGS: -L/usr/lib -lgeos_c
#include "geos_c.h"
#include <stdlib.h>
*/
import "C"

// Handle is a handle into the libgeos C library. Handle is not threadsafe.  If
// libgeos needs to be used in a concurrent fashion, then multiple handles can
// be used.
type Handle struct {
	context   C.GEOSContextHandle_t
	wkbReader *C.GEOSWKBReader
	wkbBuf    []byte
}

// NewHandle creates a new handle.
func NewHandle() *Handle {
	ctx := C.GEOS_init_r()
	return &Handle{
		context:   ctx,
		wkbReader: C.GEOSWKBReader_create_r(ctx),
	}
}

// Close cleans up memory resources associated with the handle. If Close is not
// called, then a memory leak will occurr.
func (h *Handle) Close() {
	C.GEOSWKBReader_destroy_r(h.context, h.wkbReader)
	C.GEOS_finish_r(h.context)
}

func (h *Handle) createGeomHandle(g geom.Geometry) *C.GEOSGeometry {
	wkb := bytes.NewBuffer(h.wkbBuf)
	if err := g.AsBinary(wkb); err != nil {
		panic(err) // can't fail writing to a buffer
	}
	h.wkbBuf = wkb.Bytes()
	gh := C.GEOSWKBReader_read_r(
		h.context,
		h.wkbReader,
		(*C.uchar)(&h.wkbBuf[0]),
		C.ulong(wkb.Len()),
	)
	h.wkbBuf = h.wkbBuf[:0]
	return gh
}

func (h *Handle) decodeGeomHandle(gh *C.GEOSGeometry) (geom.Geometry, error) {
	// TODO: writer should be stored on the handle.
	writer := C.GEOSWKBWriter_create_r(h.context)
	defer C.GEOSWKBWriter_destroy_r(h.context, writer)

	var size C.size_t
	wkb := C.GEOSWKBWriter_write_r(h.context, writer, gh, &size)
	defer C.free(unsafe.Pointer(wkb))

	return geom.UnmarshalWKB(bytes.NewReader(cBytesAsSlice(wkb, size)))
}

func (h *Handle) AsText(g geom.Geometry) (string, error) {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)

	writer := C.GEOSWKTWriter_create_r(h.context)
	defer C.GEOSWKTWriter_destroy_r(h.context, writer)
	wkt := C.GEOSWKTWriter_write_r(h.context, writer, gh)
	defer C.free(unsafe.Pointer(wkt))
	return C.GoString(wkt), nil
}

func (h *Handle) AsBinary(g geom.Geometry) []byte {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)

	writer := C.GEOSWKBWriter_create_r(h.context)
	defer C.GEOSWKBWriter_destroy_r(h.context, writer)
	var size C.size_t
	wkb := C.GEOSWKBWriter_write_r(h.context, writer, gh, &size)
	defer C.free(unsafe.Pointer(wkb))
	return copyBytes(wkb, size)
}

func (h *Handle) IsEmpty(g geom.Geometry) (bool, error) {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)
	return boolErr(C.GEOSisEmpty_r(h.context, gh))
}

func (h *Handle) Dimension(g geom.Geometry) int {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)
	return int(C.GEOSGeom_getDimensions_r(h.context, gh))
}

func (h *Handle) Envelope(g geom.Geometry) (geom.Geometry, error) {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)

	env := C.GEOSEnvelope_r(h.context, gh)
	if env == nil {
		return geom.Geometry{}, errors.New("could not calculate envelope")
	}
	defer C.GEOSGeom_destroy_r(h.context, env)

	return h.decodeGeomHandle(env)
}

func (h *Handle) IsSimple(g geom.Geometry) (bool, error) {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)
	return boolErr(C.GEOSisSimple_r(h.context, gh))
}

func (h *Handle) Boundary(g geom.Geometry) (geom.Geometry, error) {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)

	env := C.GEOSBoundary_r(h.context, gh)
	if env == nil {
		return geom.Geometry{}, errors.New("could not calculate envelope")
	}
	defer C.GEOSGeom_destroy_r(h.context, env)

	return h.decodeGeomHandle(env)
}

func (h *Handle) ConvexHull(g geom.Geometry) (geom.Geometry, error) {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)

	env := C.GEOSConvexHull_r(h.context, gh)
	if env == nil {
		return geom.Geometry{}, errors.New("could not calculate envelope")
	}
	defer C.GEOSGeom_destroy_r(h.context, env)

	return h.decodeGeomHandle(env)
}

func (h *Handle) IsValid(g geom.Geometry) (bool, error) {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)
	return boolErr(C.GEOSisValid_r(h.context, gh))
}

func (h *Handle) IsRing(g geom.Geometry) (bool, error) {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)
	return boolErr(C.GEOSisRing_r(h.context, gh))
}

func (h *Handle) Length(g geom.Geometry) (float64, error) {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)
	var length float64
	errInt := C.GEOSLength_r(h.context, gh, (*C.double)(&length))
	return length, intToErr(errInt)
}

func (h *Handle) Area(g geom.Geometry) (float64, error) {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)
	var area float64
	errInt := C.GEOSArea_r(h.context, gh, (*C.double)(&area))
	return area, intToErr(errInt)
}

func (h *Handle) Centroid(g geom.Geometry) (geom.Geometry, error) {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)

	env := C.GEOSGetCentroid_r(h.context, gh)
	if env == nil {
		return geom.Geometry{}, errors.New("could not calculate envelope")
	}
	defer C.GEOSGeom_destroy_r(h.context, env)

	return h.decodeGeomHandle(env)
}

func (h *Handle) Reverse(g geom.Geometry) (geom.Geometry, error) {
	gh := h.createGeomHandle(g)
	defer C.GEOSGeom_destroy(gh)

	env := C.GEOSReverse_r(h.context, gh)
	if env == nil {
		return geom.Geometry{}, errors.New("could not calculate envelope")
	}
	defer C.GEOSGeom_destroy_r(h.context, env)

	return h.decodeGeomHandle(env)
}
