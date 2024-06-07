package rawgeos

/*
#include "geos_c.h"
#include <stdlib.h>

#define MAKE_VALID_MIN_VERSION "3.8.0"
#define MAKE_VALID_MISSING ( \
	GEOS_VERSION_MAJOR < 3 || \
	(GEOS_VERSION_MAJOR == 3 && GEOS_VERSION_MINOR < 8) \
)
#if MAKE_VALID_MISSING
// This stub implementation always fails:
GEOSGeometry *GEOSMakeValid_r(GEOSContextHandle_t handle, const GEOSGeometry* g) { return NULL; }
#endif

#define COVERAGE_UNION_MIN_VERSION "3.8.0"
#define COVERAGE_UNION_MISSING ( \
	GEOS_VERSION_MAJOR < 3 || \
	(GEOS_VERSION_MAJOR == 3 && GEOS_VERSION_MINOR < 8) \
)
#if COVERAGE_UNION_MISSING
// This stub implementation always fails:
GEOSGeometry *GEOSCoverageUnion_r(GEOSContextHandle_t handle, const GEOSGeometry* g) { return NULL; }
#endif

#define COVERAGE_SIMPLIFY_VW_MIN_VERSION "3.12.0"
#define COVERAGE_SIMPLIFY_VW_MISSING ( \
	GEOS_VERSION_MAJOR < 3 || \
	(GEOS_VERSION_MAJOR == 3 && GEOS_VERSION_MINOR < 12) \
)
#if COVERAGE_SIMPLIFY_VW_MISSING
// This stub implementation always fails:
GEOSGeometry *GEOSCoverageSimplifyVW_r(GEOSContextHandle_t handle, const GEOSGeometry* g, double tolerance, int preserveTopology) { return NULL; }
#endif

#define CONCAVE_HULL_MIN_VERSION "3.11.0"
#define CONCAVE_HULL_MISSING ( \
	GEOS_VERSION_MAJOR < 3 || \
	(GEOS_VERSION_MAJOR == 3 && GEOS_VERSION_MINOR < 11) \
)
#if CONCAVE_HULL_MISSING
// This stub implementation always fails:
GEOSGeometry* GEOSConcaveHull_r(GEOSContextHandle_t handle, const GEOSGeometry *g, double ratio, unsigned int allowHoles) { return NULL; }
#endif

*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/peterstace/simplefeatures/geom"
)

func Equals(a, b geom.Geometry) (bool, error) {
	result, err := binaryOpB(a, b, func(h C.GEOSContextHandle_t, a, b *C.GEOSGeometry) C.char {
		return C.GEOSEquals_r(h, a, b)
	})
	return result, wrap(err, "executing GEOSEquals_r")
}

func Disjoint(a, b geom.Geometry) (bool, error) {
	result, err := binaryOpB(a, b, func(h C.GEOSContextHandle_t, a, b *C.GEOSGeometry) C.char {
		return C.GEOSDisjoint_r(h, a, b)
	})
	return result, wrap(err, "executing GEOSDisjoint_r")
}

func Touches(a, b geom.Geometry) (bool, error) {
	result, err := binaryOpB(a, b, func(h C.GEOSContextHandle_t, a, b *C.GEOSGeometry) C.char {
		return C.GEOSTouches_r(h, a, b)
	})
	return result, wrap(err, "executing GEOSTouches_r")
}

func Contains(a, b geom.Geometry) (bool, error) {
	result, err := binaryOpB(a, b, func(h C.GEOSContextHandle_t, a, b *C.GEOSGeometry) C.char {
		return C.GEOSContains_r(h, a, b)
	})
	return result, wrap(err, "executing GEOSContains_r")
}

func Covers(a, b geom.Geometry) (bool, error) {
	result, err := binaryOpB(a, b, func(h C.GEOSContextHandle_t, a, b *C.GEOSGeometry) C.char {
		return C.GEOSCovers_r(h, a, b)
	})
	return result, wrap(err, "executing GEOSCovers_r")
}

func Intersects(a, b geom.Geometry) (bool, error) {
	result, err := binaryOpB(a, b, func(h C.GEOSContextHandle_t, a, b *C.GEOSGeometry) C.char {
		return C.GEOSIntersects_r(h, a, b)
	})
	return result, wrap(err, "executing GEOSIntersects_r")
}

func Within(a, b geom.Geometry) (bool, error) {
	result, err := binaryOpB(a, b, func(h C.GEOSContextHandle_t, a, b *C.GEOSGeometry) C.char {
		return C.GEOSWithin_r(h, a, b)
	})
	return result, wrap(err, "executing GEOSWithin_r")
}

func CoveredBy(a, b geom.Geometry) (bool, error) {
	result, err := binaryOpB(a, b, func(h C.GEOSContextHandle_t, a, b *C.GEOSGeometry) C.char {
		return C.GEOSCoveredBy_r(h, a, b)
	})
	return result, wrap(err, "executing GEOSCoveredBy_r")
}

func Crosses(a, b geom.Geometry) (bool, error) {
	result, err := binaryOpB(a, b, func(h C.GEOSContextHandle_t, a, b *C.GEOSGeometry) C.char {
		return C.GEOSCrosses_r(h, a, b)
	})
	return result, wrap(err, "executing GEOSCrosses_r")
}

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

func Overlaps(a, b geom.Geometry) (bool, error) {
	result, err := binaryOpB(a, b, func(h C.GEOSContextHandle_t, a, b *C.GEOSGeometry) C.char {
		return C.GEOSOverlaps_r(h, a, b)
	})
	return result, wrap(err, "executing GEOSOverlaps_r")
}

func EqualsExact(a, b geom.Geometry) (bool, error) {
	result, err := binaryOpB(a, b, func(h C.GEOSContextHandle_t, a, b *C.GEOSGeometry) C.char {
		return C.GEOSEqualsExact_r(h, a, b, C.double(0.0))
	})
	return result, wrap(err, "executing GEOSEqualsExact_r")
}

func Union(a, b geom.Geometry) (geom.Geometry, error) {
	g, err := binaryOpG(a, b, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSUnion_r(ctx, a, b)
	})
	return g, wrap(err, "executing GEOSUnion_r")
}

func Intersection(a, b geom.Geometry) (geom.Geometry, error) {
	g, err := binaryOpG(a, b, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSIntersection_r(ctx, a, b)
	})
	return g, wrap(err, "executing GEOSIntersection_r")
}

type BufferEndCapStyle string

const (
	BufferEndCapStyleRound  BufferEndCapStyle = "round"
	BufferEndCapStyleFlat   BufferEndCapStyle = "flat"
	BufferEndCapStyleSquare BufferEndCapStyle = "square"
)

type BufferJoinStyle string

const (
	BufferJoinStyleRound BufferJoinStyle = "round"
	BufferJoinStyleMitre BufferJoinStyle = "mitre"
	BufferJoinStyleBevel BufferJoinStyle = "bevel"
)

type BufferOptions struct {
	Radius       float64
	QuadSegments int
	EndCapStyle  BufferEndCapStyle
	JoinStyle    BufferJoinStyle
	MitreLimit   float64 // Only applicable with BufferJoinStyleMitre.
}

func Buffer(g geom.Geometry, opts BufferOptions) (geom.Geometry, error) {
	var endCapStyle int
	switch opts.EndCapStyle {
	case BufferEndCapStyleRound:
		endCapStyle = int(C.GEOSBUF_CAP_ROUND)
	case BufferEndCapStyleFlat:
		endCapStyle = int(C.GEOSBUF_CAP_FLAT)
	case BufferEndCapStyleSquare:
		endCapStyle = int(C.GEOSBUF_CAP_SQUARE)
	default:
		return geom.Geometry{}, fmt.Errorf("invalid end cap style: '%s'", opts.EndCapStyle)
	}

	var joinStyle int
	var mitreLimit float64
	switch opts.JoinStyle {
	case BufferJoinStyleRound:
		joinStyle = int(C.GEOSBUF_JOIN_ROUND)
		mitreLimit = 0.0
	case BufferJoinStyleMitre:
		joinStyle = int(C.GEOSBUF_JOIN_MITRE)
		mitreLimit = opts.MitreLimit
	case BufferJoinStyleBevel:
		joinStyle = int(C.GEOSBUF_JOIN_BEVEL)
		mitreLimit = 0.0
	default:
		return geom.Geometry{}, fmt.Errorf("invalid join style: '%s'", opts.JoinStyle)
	}

	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, gh *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSBufferWithStyle_r(
			ctx, gh,
			C.double(opts.Radius),
			C.int(opts.QuadSegments),
			C.int(endCapStyle),
			C.int(joinStyle),
			C.double(mitreLimit),
		)
	})
	return result, wrap(err, "executing GEOSBufferWithStyle_r")
}

func Simplify(g geom.Geometry, tolerance float64) (geom.Geometry, error) {
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, gh *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSSimplify_r(ctx, gh, C.double(tolerance))
	})
	return result, wrap(err, "executing GEOSSimplify_r")
}

func TopologyPreserveSimplify(g geom.Geometry, tolerance float64) (geom.Geometry, error) {
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, gh *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSTopologyPreserveSimplify_r(ctx, gh, C.double(tolerance))
	})
	return result, wrap(err, "executing GEOSSimplifyPreserveTopology_r")
}

func Difference(a, b geom.Geometry) (geom.Geometry, error) {
	result, err := binaryOpG(a, b, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSDifference_r(ctx, a, b)
	})
	return result, wrap(err, "executing GEOSDifference_r")
}

func SymmetricDifference(a, b geom.Geometry) (geom.Geometry, error) {
	result, err := binaryOpG(a, b, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSSymDifference_r(ctx, a, b)
	})
	return result, wrap(err, "executing GEOSSymDifference_r")
}

func MakeValid(g geom.Geometry) (geom.Geometry, error) {
	if C.MAKE_VALID_MISSING != 0 {
		return geom.Geometry{}, UnsupportedGEOSVersionError{
			C.MAKE_VALID_MIN_VERSION, "MakeValid",
		}
	}
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSMakeValid_r(ctx, g)
	})
	return result, wrap(err, "executing GEOSMakeValid_r")
}

func CoverageUnion(g geom.Geometry) (geom.Geometry, error) {
	if C.COVERAGE_UNION_MISSING != 0 {
		return geom.Geometry{}, UnsupportedGEOSVersionError{
			C.COVERAGE_UNION_MIN_VERSION, "CoverageUnion",
		}
	}
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSCoverageUnion_r(ctx, g)
	})
	return result, wrap(err, "executing GEOSCoverageUnion_r")
}

func CoverageSimplifyVW(g geom.Geometry, tolerance float64, preserveBoundary bool) (geom.Geometry, error) {
	if C.COVERAGE_SIMPLIFY_VW_MISSING != 0 {
		return geom.Geometry{}, UnsupportedGEOSVersionError{
			C.COVERAGE_SIMPLIFY_VW_MIN_VERSION, "CoverageSimplifyVW",
		}
	}
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSCoverageSimplifyVW_r(ctx, g, C.double(tolerance), goBoolToCInt(preserveBoundary))
	})
	return result, wrap(err, "executing GEOSCoverageSimplifyVW_r")
}

func UnaryUnion(g geom.Geometry) (geom.Geometry, error) {
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSUnaryUnion_r(ctx, g)
	})
	return result, wrap(err, "executing GEOSUnaryUnion_r")
}

func Reverse(g geom.Geometry) (geom.Geometry, error) {
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSReverse_r(ctx, g)
	})
	return result, wrap(err, "executing GEOSReverse_r")
}

func Boundary(g geom.Geometry) (geom.Geometry, error) {
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSBoundary_r(ctx, g)
	})
	return result, wrap(err, "executing GEOSBoundary_r")
}

func MinimumRotatedRectangle(g geom.Geometry) (geom.Geometry, error) {
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSMinimumRotatedRectangle_r(ctx, g)
	})
	return result, wrap(err, "executing GEOSMinimumRotatedRectangle_r")
}

func AsText(g geom.Geometry) (string, error) {
	var result string
	if err := unaryOpE(g, func(h *handle, gh *C.GEOSGeometry) error {
		writer := C.GEOSWKTWriter_create_r(h.context)
		if writer == nil {
			return wrap(h.err(), "executing GEOSWKTWriter_create_r")
		}
		defer C.GEOSWKTWriter_destroy_r(h.context, writer)
		C.GEOSWKTWriter_setTrim_r(h.context, writer, 1)

		cstr := C.GEOSWKTWriter_write_r(h.context, writer, gh)
		if cstr == nil {
			return wrap(h.err(), "executing GEOSWKTWriter_write_r")
		}
		defer C.GEOSFree_r(h.context, unsafe.Pointer(cstr))
		result = C.GoString(cstr)

		return nil
	}); err != nil {
		return "", err
	}
	return result, nil
}

func AsBinary(g geom.Geometry) ([]byte, error) {
	var result []byte
	if err := unaryOpE(g, func(h *handle, gh *C.GEOSGeometry) error {
		var size C.size_t
		cptr := C.GEOSGeomToWKB_buf_r(h.context, gh, &size)
		if cptr == nil {
			return wrap(h.err(), "executing GEOSGeomToWKB_buf_r")
		}
		defer C.GEOSFree_r(h.context, unsafe.Pointer(cptr))
		result = C.GoBytes(unsafe.Pointer(cptr), C.int(size))
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func IsValid(g geom.Geometry) (bool, error) {
	result, err := unaryOpB(g, func(h C.GEOSContextHandle_t, g *C.GEOSGeometry) C.char {
		return C.GEOSisValid_r(h, g)
	})
	return result, wrap(err, "executing GEOSisValid_r")
}

func IsEmpty(g geom.Geometry) (bool, error) {
	result, err := unaryOpB(g, func(h C.GEOSContextHandle_t, g *C.GEOSGeometry) C.char {
		return C.GEOSisEmpty_r(h, g)
	})
	return result, wrap(err, "executing GEOSisEmpty_r")
}

func IsRing(g geom.Geometry) (bool, error) {
	result, err := unaryOpB(g, func(h C.GEOSContextHandle_t, g *C.GEOSGeometry) C.char {
		return C.GEOSisRing_r(h, g)
	})
	return result, wrap(err, "executing GEOSisRing_r")
}

func IsSimple(g geom.Geometry) (bool, error) {
	result, err := unaryOpB(g, func(h C.GEOSContextHandle_t, g *C.GEOSGeometry) C.char {
		return C.GEOSisSimple_r(h, g)
	})
	return result, wrap(err, "executing GEOSisSimple_r")
}

func ConvexHull(g geom.Geometry) (geom.Geometry, error) {
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSConvexHull_r(ctx, g)
	})
	return result, wrap(err, "executing GEOSConvexHull_r")
}

func ConcaveHull(g geom.Geometry, ratio float64, allowHoles bool) (geom.Geometry, error) {
	if C.CONCAVE_HULL_MISSING != 0 {
		return geom.Geometry{}, UnsupportedGEOSVersionError{
			C.CONCAVE_HULL_MIN_VERSION, "ConcaveHull",
		}
	}
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSConcaveHull_r(ctx, g, C.double(ratio), goBoolToCUint(allowHoles))
	})
	return result, wrap(err, "executing GEOSConcaveHull_r")
}

func Centroid(g geom.Geometry) (geom.Geometry, error) {
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSGetCentroid_r(ctx, g)
	})
	return result, wrap(err, "executing GEOSGetCentroid_r")
}

func Envelope(g geom.Geometry) (geom.Geometry, error) {
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSEnvelope_r(ctx, g)
	})
	return result, wrap(err, "executing GEOSEnvelope_r")
}

func Area(g geom.Geometry) (float64, error) {
	result, err := unaryOpF(g, func(h C.GEOSContextHandle_t, g *C.GEOSGeometry, d *C.double) C.int {
		return C.GEOSArea_r(h, g, d)
	})
	return result, wrap(err, "executing GEOSArea_r")
}

func Length(g geom.Geometry) (float64, error) {
	result, err := unaryOpF(g, func(h C.GEOSContextHandle_t, g *C.GEOSGeometry, d *C.double) C.int {
		return C.GEOSLength_r(h, g, d)
	})
	return result, wrap(err, "executing GEOSLength_r")
}

func Dimension(g geom.Geometry) (int, error) {
	result, err := unaryOpI(g, func(h C.GEOSContextHandle_t, g *C.GEOSGeometry) C.int {
		return C.GEOSGeom_getDimensions_r(h, g)
	})
	return result, wrap(err, "executing GEOSGeom_getDimensions_r")
}

func Distance(a, b geom.Geometry) (float64, error) {
	result, err := binaryOpF(a, b, func(h C.GEOSContextHandle_t, a, b *C.GEOSGeometry, d *C.double) C.int {
		return C.GEOSDistance_r(h, a, b, d)
	})
	return result, wrap(err, "executing GEOSDistance_r")
}

func RelatePatternMatch(mat, pat string) (bool, error) {
	h, err := newHandle()
	if err != nil {
		return false, err
	}
	defer h.release()

	cMat := C.CString(mat)
	cPat := C.CString(pat)
	defer C.free(unsafe.Pointer(cMat))
	defer C.free(unsafe.Pointer(cPat))

	return h.boolErr(C.GEOSRelatePatternMatch_r(h.context, cMat, cPat))
}

func FromText(wkt string) (geom.Geometry, error) {
	h, err := newHandle()
	if err != nil {
		return geom.Geometry{}, err
	}
	defer h.release()

	cwkt := C.CString(wkt)
	defer C.free(unsafe.Pointer(cwkt))

	reader := C.GEOSWKTReader_create_r(h.context)
	if reader == nil {
		return geom.Geometry{}, fmt.Errorf("creating wkt reader: %w", h.err())
	}
	defer C.GEOSWKTReader_destroy_r(h.context, reader)

	gh := C.GEOSWKTReader_read_r(h.context, reader, cwkt)
	if gh == nil {
		return geom.Geometry{}, h.err()
	}
	return h.decode(gh)
}

func FromBinary(wkb []byte) (geom.Geometry, error) {
	h, err := newHandle()
	if err != nil {
		return geom.Geometry{}, err
	}
	defer h.release()

	gh, err := h.createGeometryHandleFromWKB(wkb)
	if err != nil {
		return geom.Geometry{}, err
	}
	defer C.GEOSGeom_destroy_r(h.context, gh)

	return h.decode(gh)
}
