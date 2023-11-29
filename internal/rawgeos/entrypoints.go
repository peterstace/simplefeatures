package rawgeos

/*
#cgo LDFLAGS: -lgeos_c
#cgo CFLAGS: -Wall
#include "geos_c.h"

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
		return geom.Geometry{}, unsupportedGEOSVersionError{
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
		return geom.Geometry{}, unsupportedGEOSVersionError{
			C.COVERAGE_UNION_MIN_VERSION, "CoverageUnion",
		}
	}
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSCoverageUnion_r(ctx, g)
	})
	return result, wrap(err, "executing GEOSCoverageUnion_r")
}

func UnaryUnion(g geom.Geometry) (geom.Geometry, error) {
	result, err := unaryOpG(g, func(ctx C.GEOSContextHandle_t, g *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSUnaryUnion_r(ctx, g)
	})
	return result, wrap(err, "executing GEOSUnaryUnion_r")
}
