package geos

/*
#cgo LDFLAGS: -lgeos_c
#cgo CFLAGS: -Wall
#include "geos_c.h"
*/
import "C"

import (
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
	g, err := binaryOpG(a, b, opts, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSUnion_r(ctx, a, b)
	})
	return g, wrap(err, "executing GEOSUnion_r")
}

// Intersection returns a geometry that is the intersection of the input
// geometries. Formally, the returned geometry will contain a particular point
// X if and only if X is present in both geometries.
func Intersection(a, b geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	g, err := binaryOpG(a, b, opts, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSIntersection_r(ctx, a, b)
	})
	return g, wrap(err, "executing GEOSIntersection_r")
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
	result, err := unaryOpG(g, optSet.ctorOpts, func(ctx C.GEOSContextHandle_t, gh *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSBufferWithStyle_r(
			ctx, gh, C.double(radius),
			C.int(optSet.quadSegments),
			C.int(optSet.endCapStyle),
			C.int(optSet.joinStyle),
			C.double(optSet.mitreLimit),
		)
	})
	return result, wrap(err, "executing GEOSBufferWithStyle_r")
}

// Simplify creates a simplified version of a geometry using the
// Douglas-Peucker algorithm. Topological invariants may not be maintained,
// e.g. polygons can collapse into linestrings, and holes in polygons may be
// lost.
func Simplify(g geom.Geometry, tolerance float64, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	result, err := unaryOpG(g, opts, func(ctx C.GEOSContextHandle_t, gh *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSSimplify_r(ctx, gh, C.double(tolerance))
	})
	return result, wrap(err, "executing GEOSSimplify_r")
}

// Difference returns the geometry that represents the parts of input geometry
// A that are not part of input geometry B.
func Difference(a, b geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	result, err := binaryOpG(a, b, opts, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSDifference_r(ctx, a, b)
	})
	return result, wrap(err, "executing GEOSDifference_r")
}

// SymmetricDifference returns the geometry that represents the parts of the
// input geometries that are not part of the other input geometry.
func SymmetricDifference(a, b geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error) {
	result, err := binaryOpG(a, b, opts, func(ctx C.GEOSContextHandle_t, a, b *C.GEOSGeometry) *C.GEOSGeometry {
		return C.GEOSSymDifference_r(ctx, a, b)
	})
	return result, wrap(err, "executing GEOSSymDifference_r")
}
