package main

import (
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

void sf_geos_error_handler(const char *message, void *userdata) {
	strncpy(userdata, message, 1024);
}

GEOSContextHandle_t sf_geos_init(void *userdata) {
	GEOSContextHandle_t ctx = GEOS_init_r();
	GEOSContext_setErrorMessageHandler_r(ctx, sf_geos_error_handler, userdata);
	return ctx;
}

*/
import "C"

var (
	errNonEmptyGeometryCollectionNotSupported = errors.New(
		"non-empty GeometryCollection not supported")
	errLibgeosCrash = errors.New(
		"libgeos would crash with this input")
)

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
	h.context = C.sf_geos_init(unsafe.Pointer(&h.errBuf))
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
// called, then a memory leak will occur.
func (h *Handle) Close() {
	C.GEOSWKBWriter_destroy_r(h.context, h.wkbWriter)
	C.GEOSWKBReader_destroy_r(h.context, h.wkbReader)
	C.GEOS_finish_r(h.context)
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
	switch {
	case g.IsPoint():
		return h.createGeomHandleForPoint(g.MustAsPoint())
	case g.IsMultiPoint():
		return h.createGeomHandleForMultiPoint(g.MustAsMultiPoint())
	case g.IsMultiPolygon():
		return h.createGeomHandleForMultiPolygon(g.MustAsMultiPolygon())
	case g.IsGeometryCollection():
		return h.createGeomHandleForGeometryCollection(g.MustAsGeometryCollection())
	default:
		return h.createGeomHandleUsingWKB(g)
	}
}

func (h *Handle) createGeomHandleForPoint(pt geom.Point) (*C.GEOSGeometry, error) {
	if !pt.IsEmpty() {
		return h.createGeomHandleUsingWKB(pt.AsGeometry())
	}
	// Empty Points cannot officially be represented in WKB, so construct
	// manually.
	gh := C.GEOSGeom_createEmptyPoint_r(h.context)
	if gh == nil {
		return nil, h.err()
	}
	return gh, nil
}

func (h *Handle) createGeomHandleForMultiPoint(mp geom.MultiPoint) (*C.GEOSGeometry, error) {
	n := mp.NumPoints()
	points := make([]*C.GEOSGeometry, n)
	for i := 0; i < n; i++ {
		var err error
		points[i], err = h.createGeomHandleForPoint(mp.PointN(i))
		if err != nil {
			for _, gh := range points {
				if gh != nil {
					C.GEOSGeom_destroy_r(h.context, gh)
				}
			}
			return nil, err
		}
	}
	var geomsPtr **C.GEOSGeometry
	if len(points) > 0 {
		geomsPtr = &points[0]
	}
	gh := C.GEOSGeom_createCollection_r(
		h.context, C.GEOS_MULTIPOINT, geomsPtr, C.uint(n),
	)
	if gh == nil {
		return nil, h.err()
	}
	return gh, nil
}

func (h *Handle) createGeomHandleForMultiPolygon(mp geom.MultiPolygon) (*C.GEOSGeometry, error) {
	n := mp.NumPolygons()
	polys := make([]*C.GEOSGeometry, n)
	for i := 0; i < n; i++ {
		var err error
		polys[i], err = h.createGeomHandle(mp.PolygonN(i).AsGeometry())
		if err != nil {
			for _, gh := range polys {
				if gh != nil {
					C.GEOSGeom_destroy_r(h.context, gh)
				}
			}
			return nil, err
		}
	}
	var geomsPtr **C.GEOSGeometry
	if len(polys) > 0 {
		geomsPtr = &polys[0]
	}
	gh := C.GEOSGeom_createCollection_r(
		h.context, C.GEOS_MULTIPOLYGON, geomsPtr, C.uint(n),
	)
	if gh == nil {
		return nil, h.err()
	}
	return gh, nil
}

func (h *Handle) createGeomHandleForGeometryCollection(gc geom.GeometryCollection) (*C.GEOSGeometry, error) {
	n := gc.NumGeometries()
	geoms := make([]*C.GEOSGeometry, n)
	for i := 0; i < n; i++ {
		var err error
		geoms[i], err = h.createGeomHandle(gc.GeometryN(i))
		if err != nil {
			for _, gh := range geoms {
				if gh != nil {
					C.GEOSGeom_destroy_r(h.context, gh)
				}
			}
			return nil, err
		}
	}
	var geomsPtr **C.GEOSGeometry
	if len(geoms) > 0 {
		geomsPtr = &geoms[0]
	}
	gh := C.GEOSGeom_createCollection_r(
		h.context, C.GEOS_GEOMETRYCOLLECTION, geomsPtr, C.uint(n),
	)
	if gh == nil {
		return nil, h.err()
	}
	return gh, nil
}

func (h *Handle) createGeomHandleUsingWKB(g geom.Geometry) (*C.GEOSGeometry, error) {
	h.wkbBuf = g.AppendWKB(h.wkbBuf)
	gh := C.GEOSWKBReader_read_r(
		h.context,
		h.wkbReader,
		(*C.uchar)(&h.wkbBuf[0]),
		C.ulong(len(h.wkbBuf)),
	)
	h.wkbBuf = h.wkbBuf[:0]
	if gh == nil {
		return nil, h.err()
	}
	return gh, nil
}

func (h *Handle) decodeGeomHandle(gh *C.GEOSGeometry) (geom.Geometry, error) {
	geomType := C.GEOSGeomType_r(h.context, gh)
	if geomType == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.free(unsafe.Pointer(geomType))

	switch C.GoString(geomType) {
	case "Point":
		isEmpty, err := h.boolErr(C.GEOSisEmpty_r(h.context, gh))
		if err != nil {
			return geom.Geometry{}, err
		}
		if isEmpty {
			return geom.NewEmptyPoint(geom.DimXY).AsGeometry(), nil
		}
		return h.decodeGeomHandleUsingWKB(gh)
	case "MultiPoint":
		n := C.GEOSGetNumGeometries_r(h.context, gh)
		if n == -1 {
			return geom.Geometry{}, h.err()
		}
		subPoints := make([]geom.Point, n)
		for i := 0; i < int(n); i++ {
			sub := C.GEOSGetGeometryN_r(h.context, gh, C.int(i))
			if sub == nil {
				return geom.Geometry{}, h.err()
			}
			isEmpty, err := h.boolErr(C.GEOSisEmpty_r(h.context, sub))
			if err != nil {
				return geom.Geometry{}, err
			}
			if isEmpty {
				subPoints[i] = geom.NewEmptyPoint(geom.DimXY)
			} else {
				subPointAsGeom, err := h.decodeGeomHandleUsingWKB(sub)
				if err != nil {
					return geom.Geometry{}, nil
				}
				if !subPointAsGeom.IsPoint() {
					return geom.Geometry{}, errors.New(
						"internal error: expected point")
				}
				subPoints[i] = subPointAsGeom.MustAsPoint()
			}
		}
		return geom.NewMultiPoint(subPoints).AsGeometry(), nil
	case "MultiPolygon":
		n := C.GEOSGetNumGeometries_r(h.context, gh)
		if n == -1 {
			return geom.Geometry{}, h.err()
		}
		subPolys := make([]geom.Polygon, n)
		for i := 0; i < int(n); i++ {
			sub := C.GEOSGetGeometryN_r(h.context, gh, C.int(i))
			if sub == nil {
				return geom.Geometry{}, h.err()
			}
			subPolyAsGeom, err := h.decodeGeomHandleUsingWKB(sub)
			if err != nil {
				return geom.Geometry{}, err
			}
			if !subPolyAsGeom.IsPolygon() {
				return geom.Geometry{}, errors.New(
					"internal error: expected polygon")
			}
			subPolys[i] = subPolyAsGeom.MustAsPolygon()
		}
		mp := geom.NewMultiPolygon(subPolys)
		if err := mp.Validate(); err != nil {
			return geom.Geometry{}, err
		}
		return mp.AsGeometry(), nil
	case "GeometryCollection":
		n := C.GEOSGetNumGeometries_r(h.context, gh)
		if n == -1 {
			return geom.Geometry{}, h.err()
		}
		subGeoms := make([]geom.Geometry, n)
		for i := 0; i < int(n); i++ {
			sub := C.GEOSGetGeometryN_r(h.context, gh, C.int(i))
			if sub == nil {
				return geom.Geometry{}, h.err()
			}
			var err error
			subGeoms[i], err = h.decodeGeomHandle(sub)
			if err != nil {
				return geom.Geometry{}, err
			}
		}
		return geom.NewGeometryCollection(subGeoms).AsGeometry(), nil
	case "LineString", "Polygon", "MultiLineString":
		return h.decodeGeomHandleUsingWKB(gh)
	default:
		return geom.Geometry{}, fmt.Errorf("unexpected geometry type: %s", C.GoString(geomType))
	}
}

// errInvalidAccordingToGEOS indicates that the geometry or geometry resulting
// from an operation is invalid according to GEOS.
var errInvalidAccordingToGEOS = errors.New("invalid geometry according to GEOS")

func (h *Handle) decodeGeomHandleUsingWKB(gh *C.GEOSGeometry) (geom.Geometry, error) {
	// Check to see if GEOS thinks the geometry is invalid. Sometimes the
	// results of complex operations (e.g. Union) can be invalid due to bugs in
	// GEOS.
	isValid, err := h.boolErr(C.GEOSisValid_r(h.context, gh))
	if err != nil {
		// Geometry is so invalid that GEOS can't even tell if it's valid or not.
		return geom.Geometry{}, err
	}
	if !isValid {
		return geom.Geometry{}, errInvalidAccordingToGEOS
	}

	var size C.size_t
	wkb := C.GEOSWKBWriter_write_r(h.context, h.wkbWriter, gh, &size)
	if wkb == nil {
		return geom.Geometry{}, fmt.Errorf("writing wkb: %v", h.err())
	}
	defer C.GEOSFree_r(h.context, unsafe.Pointer(wkb))
	byts := C.GoBytes(unsafe.Pointer(wkb), C.int(size))

	return geom.UnmarshalWKB(byts)
}

func (h *Handle) asText(g geom.Geometry) (string, error) {
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

func (h *Handle) fromText(wkt string) (geom.Geometry, error) {
	reader := C.GEOSWKTReader_create_r(h.context)
	if reader == nil {
		return geom.Geometry{}, fmt.Errorf("creating wkt reader: %v", h.err())
	}
	defer C.GEOSWKTReader_destroy_r(h.context, reader)

	cwkt := C.CString(wkt)
	defer C.free(unsafe.Pointer(cwkt))

	gh := C.GEOSWKTReader_read_r(h.context, reader, cwkt)
	if gh == nil {
		return geom.Geometry{}, h.err()
	}

	return h.decodeGeomHandle(gh)
}

func (h *Handle) asBinary(g geom.Geometry) ([]byte, error) {
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

func (h *Handle) fromBinary(wkb []byte) (geom.Geometry, error) {
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

func (h *Handle) isEmpty(g geom.Geometry) (bool, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return false, err
	}
	defer C.GEOSGeom_destroy(gh)

	return h.boolErr(C.GEOSisEmpty_r(h.context, gh))
}

func (h *Handle) dimension(g geom.Geometry) (int, error) {
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

func (h *Handle) envelope(g geom.Geometry) (geom.Envelope, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return geom.Envelope{}, err
	}
	defer C.GEOSGeom_destroy(gh)

	env := C.GEOSEnvelope_r(h.context, gh)
	if env == nil {
		return geom.Envelope{}, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, env)

	if isEmpty, err := h.boolErr(C.GEOSisEmpty_r(h.context, env)); err != nil {
		return geom.Envelope{}, err
	} else if isEmpty {
		return geom.Envelope{}, nil
	}

	// libgeos will return either a Point or a Polygon. In the case where the
	// envelope has a width but no height or a height but no width, then an
	// invalid Polygon is returned.
	geomType := C.GEOSGeomType_r(h.context, env)
	if geomType == nil {
		return geom.Envelope{}, h.err()
	}
	defer C.free(unsafe.Pointer(geomType))
	switch gTypeStr := C.GoString(geomType); gTypeStr {
	case "Point":
		var x C.double
		if C.GEOSGeomGetX_r(h.context, env, &x) == 0 {
			return geom.Envelope{}, h.err()
		}
		var y C.double
		if C.GEOSGeomGetY_r(h.context, env, &y) == 0 {
			return geom.Envelope{}, h.err()
		}
		return geom.NewEnvelope([]geom.XY{{X: float64(x), Y: float64(y)}})
	case "Polygon":
		// Continues below
	default:
		return geom.Envelope{}, fmt.Errorf(
			"unexpected geometry type from GEOSEnvelope_r: %v", gTypeStr)
	}

	ring := C.GEOSGetExteriorRing_r(h.context, env)
	if ring == nil {
		return geom.Envelope{}, h.err()
	}
	// ring belongs to env, so doesn't need to be destroyed.

	seq := C.GEOSGeom_getCoordSeq_r(h.context, ring)
	if seq == nil {
		return geom.Envelope{}, h.err()
	}
	// seq belongs to ring, so doesn't need to be destroyed.

	var size C.uint
	if C.GEOSCoordSeq_getSize_r(h.context, seq, &size) == 0 {
		return geom.Envelope{}, h.err()
	}
	if size == 0 {
		return geom.Envelope{}, errors.New(
			"coordinate sequence doesn't contain any points")
	}

	var sfEnv geom.Envelope
	for i := C.uint(0); i < size; i++ {
		var x C.double
		if C.GEOSCoordSeq_getX_r(h.context, seq, i, &x) == 0 {
			return geom.Envelope{}, h.err()
		}
		var y C.double
		if C.GEOSCoordSeq_getY_r(h.context, seq, i, &y) == 0 {
			return geom.Envelope{}, h.err()
		}
		xy := geom.XY{X: float64(x), Y: float64(y)}
		sfEnv, err = sfEnv.ExtendToIncludeXY(xy)
		if err != nil {
			return geom.Envelope{}, err
		}
	}

	return sfEnv, nil
}

func (h *Handle) isSimple(g geom.Geometry) (isSimple bool, defined bool, err error) {
	// libgeos crashes when GEOSisSimple_r is called with MultiPoints
	// containing empty Points.
	if containsMultiPointWithEmptyPoint(g) {
		return false, false, errLibgeosCrash
	}

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

func (h *Handle) boundary(g geom.Geometry) (geom.Geometry, bool, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return geom.Geometry{}, false, err
	}
	defer C.GEOSGeom_destroy(gh)

	geomType := C.GEOSGeomType_r(h.context, gh)
	if geomType == nil {
		return geom.Geometry{}, false, h.err()
	}
	defer C.free(unsafe.Pointer(geomType))
	if C.GoString(geomType) == "GeometryCollection" {
		return geom.Geometry{}, false, nil
	}

	bound := C.GEOSBoundary_r(h.context, gh)
	if bound == nil {
		return geom.Geometry{}, false, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, bound)

	sfBound, err := h.decodeGeomHandle(bound)
	return sfBound, true, err
}

func (h *Handle) convexHull(g geom.Geometry) (geom.Geometry, error) {
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

func (h *Handle) isValid(g geom.Geometry) (bool, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return false, err
	}
	defer C.GEOSGeom_destroy(gh)

	return h.boolErr(C.GEOSisValid_r(h.context, gh))
}

func (h *Handle) isRing(g geom.Geometry) (bool, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return false, err
	}
	defer C.GEOSGeom_destroy(gh)

	return h.boolErr(C.GEOSisRing_r(h.context, gh))
}

func (h *Handle) length(g geom.Geometry) (float64, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return 0, err
	}
	defer C.GEOSGeom_destroy(gh)

	var length float64
	errInt := C.GEOSLength_r(h.context, gh, (*C.double)(&length))
	return length, h.intToErr(errInt)
}

func (h *Handle) area(g geom.Geometry) (float64, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return 0, err
	}
	defer C.GEOSGeom_destroy(gh)

	var area float64
	errInt := C.GEOSArea_r(h.context, gh, (*C.double)(&area))
	return area, h.intToErr(errInt)
}

func (h *Handle) centroid(g geom.Geometry) (geom.Geometry, error) {
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

func (h *Handle) reverse(g geom.Geometry) (geom.Geometry, error) {
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

func (h *Handle) simplify(g geom.Geometry, threshold float64) (geom.Geometry, error) {
	gh, err := h.createGeomHandle(g)
	if err != nil {
		return geom.Geometry{}, err
	}
	defer C.GEOSGeom_destroy(gh)

	simp := C.GEOSSimplify_r(h.context, gh, C.double(threshold))
	if simp == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, simp)

	return h.decodeGeomHandle(simp)
}

func (h *Handle) intersects(g1, g2 geom.Geometry) (bool, error) {
	if isNonEmptyGeometryCollection(g1) || isNonEmptyGeometryCollection(g2) {
		return false, errNonEmptyGeometryCollectionNotSupported
	}
	// libgeos crashes when GEOSIntersects_r is called with MultiPoints
	// containing empty Points.
	if containsMultiPointWithEmptyPoint(g1) || containsMultiPointWithEmptyPoint(g2) {
		return false, errLibgeosCrash
	}

	gh1, err := h.createGeomHandle(g1)
	if err != nil {
		return false, h.err()
	}
	defer C.GEOSGeom_destroy(gh1)
	gh2, err := h.createGeomHandle(g2)
	if err != nil {
		return false, h.err()
	}
	defer C.GEOSGeom_destroy(gh2)

	return h.boolErr(C.GEOSIntersects_r(h.context, gh1, gh2))
}

func (h *Handle) exactEquals(g1, g2 geom.Geometry) (bool, error) {
	if isNonEmptyGeometryCollection(g1) || isNonEmptyGeometryCollection(g2) {
		return false, errNonEmptyGeometryCollectionNotSupported
	}

	gh1, err := h.createGeomHandle(g1)
	if err != nil {
		return false, h.err()
	}
	defer C.GEOSGeom_destroy(gh1)
	gh2, err := h.createGeomHandle(g2)
	if err != nil {
		return false, h.err()
	}
	defer C.GEOSGeom_destroy(gh2)

	return h.boolErr(C.GEOSEqualsExact_r(h.context, gh1, gh2, 0.0))
}

func (h *Handle) distance(g1, g2 geom.Geometry) (float64, error) {
	if containsMultiLineStringWithEmptyLineString(g1) ||
		containsMultiLineStringWithEmptyLineString(g2) ||
		containsMultiPointWithEmptyPoint(g1) ||
		containsMultiPointWithEmptyPoint(g2) ||
		containsMultiPolygonWithEmptyPolygon(g1) ||
		containsMultiPolygonWithEmptyPolygon(g2) {
		// GEOS crashes on these inputs.
		return 0, errLibgeosCrash
	}

	gh1, err := h.createGeomHandle(g1)
	if err != nil {
		return 0, h.err()
	}
	defer C.GEOSGeom_destroy(gh1)
	gh2, err := h.createGeomHandle(g2)
	if err != nil {
		return 0, h.err()
	}
	defer C.GEOSGeom_destroy(gh2)

	var dist C.double
	err = h.intToErr(C.GEOSDistance_r(h.context, gh1, gh2, &dist))
	return float64(dist), err
}

func (h *Handle) union(g1, g2 geom.Geometry) (geom.Geometry, error) {
	gh1, err := h.createGeomHandle(g1)
	if err != nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy(gh1)
	gh2, err := h.createGeomHandle(g2)
	if err != nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy(gh2)

	union := C.GEOSUnion_r(h.context, gh1, gh2)
	if union == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, union)

	return h.decodeGeomHandle(union)
}

func (h *Handle) intersection(g1, g2 geom.Geometry) (geom.Geometry, error) {
	gh1, err := h.createGeomHandle(g1)
	if err != nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy(gh1)
	gh2, err := h.createGeomHandle(g2)
	if err != nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy(gh2)

	union := C.GEOSIntersection_r(h.context, gh1, gh2)
	if union == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, union)

	return h.decodeGeomHandle(union)
}

func (h *Handle) difference(g1, g2 geom.Geometry) (geom.Geometry, error) {
	gh1, err := h.createGeomHandle(g1)
	if err != nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy(gh1)
	gh2, err := h.createGeomHandle(g2)
	if err != nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy(gh2)

	union := C.GEOSDifference_r(h.context, gh1, gh2)
	if union == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, union)

	return h.decodeGeomHandle(union)
}

func (h *Handle) symmetricDifference(g1, g2 geom.Geometry) (geom.Geometry, error) {
	gh1, err := h.createGeomHandle(g1)
	if err != nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy(gh1)
	gh2, err := h.createGeomHandle(g2)
	if err != nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy(gh2)

	union := C.GEOSSymDifference_r(h.context, gh1, gh2)
	if union == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy_r(h.context, union)

	return h.decodeGeomHandle(union)
}

func (h *Handle) relate(g1, g2 geom.Geometry) (string, error) {
	if containsMultiPointWithEmptyPoint(g1) || containsMultiPointWithEmptyPoint(g2) {
		// GEOS crashes on these inputs.
		return "", errLibgeosCrash
	}

	gh1, err := h.createGeomHandle(g1)
	if err != nil {
		return "", h.err()
	}
	defer C.GEOSGeom_destroy(gh1)
	gh2, err := h.createGeomHandle(g2)
	if err != nil {
		return "", h.err()
	}
	defer C.GEOSGeom_destroy(gh2)

	matrix := C.GEOSRelate_r(h.context, gh1, gh2)
	if matrix == nil {
		return "", h.err()
	}
	defer C.GEOSFree_r(h.context, unsafe.Pointer(matrix))
	return C.GoString(matrix), nil
}

func (h *Handle) relateMatch(mat, pat string) (bool, error) {
	cMat := C.CString(mat)
	cPat := C.CString(pat)
	defer C.free(unsafe.Pointer(cMat))
	defer C.free(unsafe.Pointer(cPat))
	return h.boolErr(C.GEOSRelatePatternMatch_r(h.context, cMat, cPat))
}
