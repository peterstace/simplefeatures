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

// Relate returns a 9-character DE9-IM string that describes the relationship
// between two geometries.
func Relate(g1, g2 geom.Geometry) (string, error) {
	var result string
	err := binaryOpE(g1, g2, func(h *handle, gh1, gh2 *C.GEOSGeometry) error {
		cstr := C.GEOSRelate_r(h.context, gh1, gh2)
		if cstr == nil {
			return wrap(h.err(), "executing GEOSRelate_r")
		}
		defer C.GEOSFree_r(h.context, unsafe.Pointer(cstr))
		result = C.GoString(cstr)
		return nil
	})
	return result, err
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
	return binaryOpG(a, b, opts, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSUnion_r(ctx, a, b)
	})
}

// Intersection returns a geometry that is the intersection of the input
// geometries. Formally, the returned geometry will contain a particular point
// X if and only if X is present in both geometries.
func Intersection(a, b geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	return binaryOpG(a, b, opts, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSIntersection_r(ctx, a, b)
	})
}

// BufferOption allows the behaviour of the Buffer operation to be modified.
type BufferOption func(*bufferOptionSet)

type bufferOptionSet struct {
	quadSegments int
	endCapStyle  int
	joinStyle    int
	mitreLimit   float64
	ctorOpts     []geom.ConstructorOption
}

func newBufferOptionSet(opts []BufferOption) bufferOptionSet {
	bos := bufferOptionSet{
		quadSegments: 8,
		endCapStyle:  int(C.GEOSBUF_CAP_ROUND),
		joinStyle:    int(C.GEOSBUF_JOIN_ROUND),
		mitreLimit:   0.0,
	}
	for _, opt := range opts {
		opt(&bos)
	}
	return bos
}

// BufferQuadSegments sets the number of segments used to approximate a quarter
// circle. It defaults to 8.
func BufferQuadSegments(quadSegments int) BufferOption {
	return func(bos *bufferOptionSet) {
		bos.quadSegments = quadSegments
	}
}

// BufferEndCapRound sets the end cap style to 'round'. It is 'round' by
// default.
func BufferEndCapRound() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.endCapStyle = int(C.GEOSBUF_CAP_ROUND)
	}
}

// BufferEndCapFlat sets the end cap style to 'flat'. It is 'round' by default.
func BufferEndCapFlat() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.endCapStyle = int(C.GEOSBUF_CAP_FLAT)
	}
}

// BufferEndCapSquare sets the end cap style to 'square'. It is 'round' by
// default.
func BufferEndCapSquare() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.endCapStyle = int(C.GEOSBUF_CAP_SQUARE)
	}
}

// BufferJoinStyleRound sets the join style to 'round'. It is 'round' by
// default.
func BufferJoinStyleRound() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.joinStyle = int(C.GEOSBUF_JOIN_ROUND)
		bos.mitreLimit = 0.0
	}
}

// BufferJoinStyleMitre sets the join style to 'mitre'. It is 'round' by
// default.
func BufferJoinStyleMitre(mitreLimit float64) BufferOption {
	return func(bos *bufferOptionSet) {
		bos.joinStyle = int(C.GEOSBUF_JOIN_MITRE)
		bos.mitreLimit = mitreLimit
	}
}

// BufferJoinStyleBevel sets the join style to 'bevel'. It is 'round' by
// default.
func BufferJoinStyleBevel() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.joinStyle = int(C.GEOSBUF_JOIN_BEVEL)
		bos.mitreLimit = 0.0
	}
}

// BufferConstructorOption sets constructor option that are used when
// reconstructing the buffered geometry that is returned from the GEOS lib.
func BufferConstructorOption(opts ...geom.ConstructorOption) BufferOption {
	return func(bos *bufferOptionSet) {
		bos.ctorOpts = append(bos.ctorOpts, opts...)
	}
}

// Buffer returns a geometry that contains all points within the given radius
// of the input geometry.
func Buffer(g geom.Geometry, radius float64, opts ...BufferOption) (geom.Geometry, error) {
	optSet := newBufferOptionSet(opts)
	return unaryOpG(g, optSet.ctorOpts, func(ctx C.GEOSContextHandle_t, gh *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSBufferWithStyle_r(
			ctx, gh, C.double(radius),
			C.int(optSet.quadSegments),
			C.int(optSet.endCapStyle),
			C.int(optSet.joinStyle),
			C.double(optSet.mitreLimit),
		)
	})
}

// Simplify creates a simplified version of a geometry using the
// Douglas-Peucker algorithm. Topological invariants may not be maintained,
// e.g. polygons can collapse into linestrings, and holes in polygons may be
// lost.
func Simplify(g geom.Geometry, tolerance float64, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	return unaryOpG(g, opts, func(ctx C.GEOSContextHandle_t, gh *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSSimplify_r(ctx, gh, C.double(tolerance))
	})
}

// Difference returns the geometry that represents the parts of input geometry
// A that are not part of input geometry B.
func Difference(a, b geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	return binaryOpG(a, b, opts, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSDifference_r(ctx, a, b)
	})
}

// SymmetricDifference returns the geometry that represents the parts of the
// input geometries that are not part of the other input geometry.
func SymmetricDifference(a, b geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	return binaryOpG(a, b, opts, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSSymDifference_r(ctx, a, b)
	})
}

// noop returns the geometry unaltered, via conversion to and from GEOS. This
// function is only for benchmarking purposes, hence it is not exported or used
// outside of benchmark tests.
func noop(g geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	return unaryOpG(g, opts, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
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

// relatesAny checks if the two geometries are related using any of the masks.
func relatesAny(g1, g2 geom.Geometry, masks ...string) (bool, error) {
	for i, m := range masks {
		r, err := relate(g1, g2, m)
		if err != nil {
			return false, wrap(err, "could not relate mask %d of %d", i+1, len(masks))
		}
		if r {
			return true, nil
		}
	}
	return false, nil
}

// relate invokes the GEOS GEOSRelatePattern function, which checks if two
// geometries are related according to a DE-9IM 'relates' mask.
func relate(g1, g2 geom.Geometry, mask string) (bool, error) {
	if g1.IsGeometryCollection() || g2.IsGeometryCollection() {
		return false, errors.New("GeometryCollection not supported")
	}
	if len(mask) != 9 {
		return false, fmt.Errorf("mask has invalid length %d (must be 9)", len(mask))
	}

	// Not all versions of GEOS can handle Z and M geometries correctly. For
	// Relates, we only need 2D geometries anyway.
	g1 = g1.Force2D()
	g2 = g2.Force2D()

	// The bytes in cmask represent a NULL terminated string, hence it's 10
	// chars long (rather than 9) and the last char is left as 0.
	var cmask [10]byte
	copy(cmask[:], mask)

	var result bool
	err := binaryOpE(g1, g2, func(h *handle, gh1, gh2 *C.GEOSGeometry) error {
		var err error
		result, err = h.boolErr(C.GEOSRelatePattern_r(
			h.context, gh1, gh2, (*C.char)(unsafe.Pointer(&cmask)),
		))
		return wrap(err, "executing GEOSRelatePattern_r")
	})
	return result, err
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

func binaryOpE(
	g1, g2 geom.Geometry,
	op func(*handle, *C.GEOSGeometry, *C.GEOSGeometry) error,
) error {
	h, err := newHandle()
	if err != nil {
		return err
	}
	defer h.release()

	gh1, err := h.createGeometryHandle(g1)
	if err != nil {
		return wrap(err, "converting first geom argument")
	}
	defer C.GEOSGeom_destroy(gh1)

	gh2, err := h.createGeometryHandle(g2)
	if err != nil {
		return wrap(err, "converting second geom argument")
	}
	defer C.GEOSGeom_destroy(gh2)

	return op(h, gh1, gh2)
}

func binaryOpG(
	g1, g2 geom.Geometry,
	opts []geom.ConstructorOption,
	op func(C.GEOSContextHandle_t, *C.GEOSGeometry, *C.GEOSGeometry) *C.GEOSGeometry,
) (geom.Geometry, error) {
	// Not all versions of GEOS can handle Z and M geometries correctly. For
	// binary operations, we only need 2D geometries anyway.
	g1 = g1.Force2D()
	g2 = g2.Force2D()

	var result geom.Geometry
	err := binaryOpE(g1, g2, func(h *handle, gh1, gh2 *C.GEOSGeometry) error {
		resultGH := op(h.context, gh1, gh2)
		if resultGH == nil {
			return h.err()
		}
		defer C.GEOSGeom_destroy(resultGH)
		var err error
		result, err = h.decode(resultGH, opts)
		return wrap(err, "decoding result")
	})
	return result, err
}

func unaryOpE(g geom.Geometry, op func(*handle, *C.GEOSGeometry) error) error {
	h, err := newHandle()
	if err != nil {
		return err
	}
	defer h.release()

	gh, err := h.createGeometryHandle(g)
	if err != nil {
		return wrap(err, "converting geom argument")
	}
	defer C.GEOSGeom_destroy(gh)

	return op(h, gh)
}

func unaryOpG(
	g geom.Geometry,
	opts []geom.ConstructorOption,
	op func(C.GEOSContextHandle_t, *C.GEOSGeometry) *C.GEOSGeometry,
) (geom.Geometry, error) {
	// Not all versions of libgeos can handle Z and M geometries correctly. For
	// unary operations, we only need 2D geometries anyway.
	g = g.Force2D()

	var result geom.Geometry
	err := unaryOpE(g, func(h *handle, gh *C.GEOSGeometry) error {
		resultGH := op(h.context, gh)
		if resultGH == nil {
			return h.err()
		}
		if gh != resultGH {
			// gh and resultGH will be the same if op is the noop function that
			// just returns its input. We need to avoid destroying resultGH in
			// that case otherwise we will do a double-free.
			defer C.GEOSGeom_destroy(resultGH)
		}
		var err error
		result, err = h.decode(resultGH, opts)
		return wrap(err, "decoding result")
	})
	return result, err
}

func (h *handle) decode(gh *C.GEOSGeometry, opts []geom.ConstructorOption) (geom.Geometry, error) {
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
		g, err := geom.UnmarshalWKT(wkt, opts...)
		return g, wrap(err, "failed to unmarshal GEOS WKT result")
	}
	wkb := C.GoBytes(unsafe.Pointer(serialised), C.int(size))
	g, err := geom.UnmarshalWKB(wkb, opts...)
	return g, wrap(err, "failed to unmarshal GEOS WKB result")
}

func wrap(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %v", append(args, err)...)
}
