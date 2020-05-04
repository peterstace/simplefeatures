package geos

// Package geos provides a cgo wrapper around the GEOS (Geometry Engine, Open
// Source) library.
//
// Its purpose is to provide functionality that has been implemented in GEOS,
// but is not yet available in the simplefeatures library.
//
// The operations in this package ignore Z and M values if they are present.
//
// To use this package, you will need to install the GEOS library.

/*
#cgo LDFLAGS: -lgeos_c
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

unsigned char *marshal(GEOSContextHandle_t handle, const GEOSGeometry *g, size_t *size, char *isWKT);

GEOSGeometry const *noop(GEOSContextHandle_t handle, const GEOSGeometry *g) {
	return g;
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

// Equals returns true if and only if the input geometries are spatially equal,
// i.e. they represent exactly the same set of points.
func Equals(g1, g2 geom.Geometry) (bool, error) {
	if g1.IsEmpty() && g2.IsEmpty() {
		// Part of the mask is 'dim(I(a) ∩ I(b)) = T'.  If both inputs are
		// empty, then their interiors will be empty, and thus
		// 'dim(I(a) ∩ I(b) = F'. However, we want to return 'true' for this
		// case. So we just return true manually rather than using DE-9IM.
		return true, nil
	}
	return relate(g1, g2, "T*F**FFF*")
}

// Disjoint returns true if and only if the input geometries have no points in
// common.
func Disjoint(g1, g2 geom.Geometry) (bool, error) {
	return relate(g1, g2, "FF*FF****")
}

// Touches returns true if and only if the geometries have at least 1 point in
// common, but their interiors don't intersect.
func Touches(g1, g2 geom.Geometry) (bool, error) {
	return relatesAny(
		g1, g2,
		"FT*******",
		"F**T*****",
		"F***T****",
	)
}

// Contains returns true if and only if geometry A contains geometry B.
// Formally, the following two conditions must hold:
//
// 1. No points of B lies on the exterior of geometry A. That is, B must only be
// in the exterior or boundary of A.
//
// 2 .At least one point of the interior of B lies on the interior of A. That
// is, they can't *only* intersect at their boundaries.
func Contains(a, b geom.Geometry) (bool, error) {
	return relate(a, b, "T*****FF*")
}

// Covers returns true if and only if geometry A covers geometry B. Formally,
// the following two conditions must hold:
//
// 1. No points of B lies on the exterior of geometry A. That is, B must only be
// in the exterior or boundary of A.
//
// 2. At least one point of B lies on A (either its interior or boundary).
func Covers(a, b geom.Geometry) (bool, error) {
	return relatesAny(
		a, b,
		"T*****FF*",
		"*T****FF*",
		"***T**FF*",
		"****T*FF*",
	)
}

// Intersects returns true if and only if the geometries share at least one
// point in common.
func Intersects(a, b geom.Geometry) (bool, error) {
	return relatesAny(
		a, b,
		"T********",
		"*T*******",
		"***T*****",
		"****T****",
	)
}

// Within returns true if and only if geometry A is completely within geometry
// B. Formally, the following two conditions must hold:
//
// 1. No points of A lies on the exterior of geometry B. That is, A must only be
// in the exterior or boundary of B.
//
// 2.At least one point of the interior of A lies on the interior of B. That
// is, they can't *only* intersect at their boundaries.
func Within(a, b geom.Geometry) (bool, error) {
	return relate(a, b, "T*F**F***")
}

// CoveredBy returns true if and only if geometry A is covered by geometry B.
// Formally, the following two conditions must hold:
//
// 1. No points of A lies on the exterior of geometry B. That is, A must only be
// in the exterior or boundary of B.
//
// 2. At least one point of A lies on B (either its interior or boundary).
func CoveredBy(a, b geom.Geometry) (bool, error) {
	return relatesAny(
		a, b,
		"T*F**F***",
		"*TF**F***",
		"**FT*F***",
		"**F*TF***",
	)
}

// Crosses returns true if and only if geometry A and B cross each other.
// Formally, the following conditions must hold:
//
// 1. The geometries must have some but not all interior points in common.
//
// 2. The dimensionality of the intersection must be less than the maximum
// dimension of the input geometries.
//
// 3. The intersection must not equal either of the input geometries.
func Crosses(a, b geom.Geometry) (bool, error) {
	dimA := a.Dimension()
	dimB := b.Dimension()
	switch {
	case dimA < dimB: // Point/Line, Point/Area, Line/Area
		return relate(a, b, "T*T******")
	case dimA > dimB: // Line/Point, Area/Point, Area/Line
		return relate(a, b, "T*****T**")
	case dimA == 1 && dimB == 1: // Line/Line
		return relate(a, b, "0********")
	default:
		return false, nil
	}
}

// Overlaps returns true if and only if geometry A and B overlap with each
// other. Formally, the following conditions must hold:
//
// 1. The geometries must have the same dimension.
//
// 2. The geometries must have some but not all points in common.
//
// 3. The intersection of the geometries must have the same dimension as the
// geometries themselves.
func Overlaps(a, b geom.Geometry) (bool, error) {
	dimA := a.Dimension()
	dimB := b.Dimension()
	switch {
	case (dimA == 0 && dimB == 0) || (dimA == 2 && dimB == 2):
		return relate(a, b, "T*T***T**")
	case (dimA == 1 && dimB == 1):
		return relate(a, b, "1*T***T**")
	default:
		return false, nil
	}

}

// Union returns a geometry that is the union of the input geometries.
// Formally, the returned geometry will contain a particular point X if and
// only if X is present in either geometry (or both).
func Union(a, b geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	return binaryOperation(a, b, opts, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSUnion_r(ctx, a, b)
	})
}

// Intersection returns a geometry that is the intersection of the input
// geometries. Formally, the returned geometry will contain a particular point
// X if and only if X is present in both geometries.
func Intersection(a, b geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	return binaryOperation(a, b, opts, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSIntersection_r(ctx, a, b)
	})
}

// Buffer returns a geometry that contains all points within the given radius
// of the input geometry.
func Buffer(g geom.Geometry, radius float64, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	return unaryOperation(g, opts, func(ctx C.GEOSContextHandle_t, gh *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSBufferWithStyle_r(ctx, gh, C.double(radius), 8, C.GEOSBUF_CAP_ROUND, C.GEOSBUF_JOIN_ROUND, 0.0)
	})
}

// Simplify creates a simplified version of a geometry using the
// Douglas-Peucker algorithm. Topological invariants may not be maintained,
// e.g. polygons can collapse into linestrings, and holes in polygons may be
// lost.
func Simplify(g geom.Geometry, tolerance float64, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	return unaryOperation(g, opts, func(ctx C.GEOSContextHandle_t, gh *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSSimplify_r(ctx, gh, C.double(tolerance))
	})
}

// Difference returns the geometry that represents the parts of input geometry
// A that are not part of input geometry B.
func Difference(a, b geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	return binaryOperation(a, b, opts, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSDifference_r(ctx, a, b)
	})
}

// SymmetricDifference returns the geometry that represents the parts of the
// input geometries that are not part of the other input geometry.
func SymmetricDifference(a, b geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	return binaryOperation(a, b, opts, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSSymDifference_r(ctx, a, b)
	})
}

// noop returns the geometry unaltered, via conversion to and from GEOS. This
// function is only for benchmarking purposes, hence it is not exported or used
// outside of benchmark tests.
func noop(g geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	return unaryOperation(g, opts, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.noop(ctx, g)
	})
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

	h.context = C.sf_init(unsafe.Pointer(h.errBuf))
	if h.context == nil {
		h.release()
		return nil, errors.New("could not create GEOS context")
	}

	h.reader = C.GEOSWKBReader_create_r(h.context)
	if h.reader == nil {
		h.release()
		return nil, h.err()
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
		msg = "GEOS internal error"
	}
	C.memset((unsafe.Pointer)(h.errBuf), 0, 1024) // Reset the buffer for the next error message.
	return errors.New(strings.TrimSpace(msg))
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
		return nil, h.err()
	}
	return gh, nil
}

// relatesAny checks if the two geometries are related using any of the masks.
func relatesAny(g1, g2 geom.Geometry, masks ...string) (bool, error) {
	for _, m := range masks {
		r, err := relate(g1, g2, m)
		if err != nil {
			return false, err
		}
		if r {
			return true, nil
		}
	}
	return false, nil
}

// ErrGeometryCollectionNotSupported indicates that a GeometryCollection was
// passed to a function that does not support GeometryCollections.
var ErrGeometryCollectionNotSupported = errors.New("GeometryCollection not supported")

// relate invokes the GEOS GEOSRelatePattern function, which checks if two
// geometries are related according to a DE-9IM 'relates' mask.
func relate(g1, g2 geom.Geometry, mask string) (bool, error) {
	if g1.IsGeometryCollection() || g2.IsGeometryCollection() {
		return false, ErrGeometryCollectionNotSupported
	}
	if len(mask) != 9 {
		return false, fmt.Errorf("mask has invalid length: %q", mask)
	}

	h, err := newHandle()
	if err != nil {
		return false, err
	}
	defer h.release()

	// Not all versions of GEOS can handle Z and M geometries correctly. For
	// Relates, we only need 2D geometries anyway.
	g1 = g1.Force2D()
	g2 = g2.Force2D()

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

	var cmask [10]byte
	copy(cmask[:], mask)

	return h.boolErr(C.GEOSRelatePattern_r(
		h.context, gh1, gh2,
		(*C.char)(unsafe.Pointer(&cmask)),
	))
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
		return false, fmt.Errorf("illegal result from GEOS: %v", c)
	}
}

func binaryOperation(
	g1, g2 geom.Geometry,
	opts []geom.ConstructorOption,
	op func(C.GEOSContextHandle_t, *C.GEOSGeometry, *C.GEOSGeometry) *C.GEOSGeometry,
) (geom.Geometry, error) {
	// Not all versions of GEOS can handle Z and M geometries correctly. For
	// binary operations, we only need 2D geometries anyway.
	g1 = g1.Force2D()
	g2 = g2.Force2D()

	h, err := newHandle()
	if err != nil {
		return geom.Geometry{}, err
	}
	defer h.release()

	gh1, err := h.createGeometryHandle(g1)
	if err != nil {
		return geom.Geometry{}, err
	}
	defer C.GEOSGeom_destroy(gh1)

	gh2, err := h.createGeometryHandle(g2)
	if err != nil {
		return geom.Geometry{}, err
	}
	defer C.GEOSGeom_destroy(gh2)

	resultGH := op(h.context, gh1, gh2)
	if resultGH == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy(resultGH)

	return h.decode(resultGH, opts)
}

func unaryOperation(
	g geom.Geometry,
	opts []geom.ConstructorOption,
	op func(C.GEOSContextHandle_t, *C.GEOSGeometry) *C.GEOSGeometry,
) (geom.Geometry, error) {
	// Not all versions of libgeos can handle Z and M geometries correctly. For
	// unary operations, we only need 2D geometries anyway.
	g = g.Force2D()

	h, err := newHandle()
	if err != nil {
		return geom.Geometry{}, err
	}
	defer h.release()

	gh, err := h.createGeometryHandle(g)
	if err != nil {
		return geom.Geometry{}, err
	}
	defer C.GEOSGeom_destroy(gh)

	resultGH := op(h.context, gh)
	if resultGH == nil {
		return geom.Geometry{}, h.err()
	}
	if gh != resultGH {
		// gh and resultGH will be the same if op is the noop function that
		// just returns its input.
		defer C.GEOSGeom_destroy(resultGH)
	}

	return h.decode(resultGH, opts)
}

func (h *handle) decode(gh *C.GEOSGeometry, opts []geom.ConstructorOption) (geom.Geometry, error) {
	var (
		isWKT C.char
		size  C.size_t
	)
	serialised := C.marshal(h.context, gh, &size, &isWKT)
	if serialised == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSFree_r(h.context, unsafe.Pointer(serialised))
	r := bytes.NewReader(C.GoBytes(unsafe.Pointer(serialised), C.int(size)))

	if isWKT != 0 {
		return geom.UnmarshalWKTFromReader(r, opts...)
	}
	return geom.UnmarshalWKB(r, opts...)
}
