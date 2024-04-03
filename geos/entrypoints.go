package geos

import (
	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/rawgeos"
)

// Equals returns true if and only if the input geometries are spatially equal,
// i.e. they represent exactly the same set of points.
func Equals(a, b geom.Geometry) (bool, error) {
	return rawgeos.Equals(a, b)
}

// Disjoint returns true if and only if the input geometries have no points in
// common.
func Disjoint(a, b geom.Geometry) (bool, error) {
	return rawgeos.Disjoint(a, b)
}

// Touches returns true if and only if the geometries have at least 1 point in
// common, but their interiors don't intersect.
func Touches(a, b geom.Geometry) (bool, error) {
	return rawgeos.Touches(a, b)
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
	return rawgeos.Contains(a, b)
}

// Covers returns true if and only if geometry A covers geometry B. Formally,
// the following two conditions must hold:
//
// 1. No points of B lies on the exterior of geometry A. That is, B must only be
// in the exterior or boundary of A.
//
// 2. At least one point of B lies on A (either its interior or boundary).
func Covers(a, b geom.Geometry) (bool, error) {
	return rawgeos.Covers(a, b)
}

// Intersects returns true if and only if the geometries share at least one
// point in common.
func Intersects(a, b geom.Geometry) (bool, error) {
	return rawgeos.Intersects(a, b)
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
	return rawgeos.Within(a, b)
}

// CoveredBy returns true if and only if geometry A is covered by geometry B.
// Formally, the following two conditions must hold:
//
// 1. No points of A lies on the exterior of geometry B. That is, A must only be
// in the exterior or boundary of B.
//
// 2. At least one point of A lies on B (either its interior or boundary).
func CoveredBy(a, b geom.Geometry) (bool, error) {
	return rawgeos.CoveredBy(a, b)
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
	return rawgeos.Crosses(a, b)
}

// Relate returns a 9-character DE9-IM string that describes the relationship
// between two geometries.
func Relate(g1, g2 geom.Geometry) (string, error) {
	return rawgeos.Relate(g1, g2)
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
	return rawgeos.Overlaps(a, b)
}

// Union returns a geometry that is the union of the input geometries.
// Formally, the returned geometry will contain a particular point X if and
// only if X is present in either geometry (or both).
//
// The validity of the result is not checked.
func Union(a, b geom.Geometry) (geom.Geometry, error) {
	return rawgeos.Union(a, b)
}

// Intersection returns a geometry that is the intersection of the input
// geometries. Formally, the returned geometry will contain a particular point
// X if and only if X is present in both geometries.
//
// The validity of the result is not checked.
func Intersection(a, b geom.Geometry) (geom.Geometry, error) {
	return rawgeos.Intersection(a, b)
}

// BufferOption allows the behaviour of the Buffer operation to be modified.
type BufferOption func(*bufferOptionSet)

type bufferOptionSet struct {
	quadSegments int
	endCapStyle  rawgeos.BufferEndCapStyle
	joinStyle    rawgeos.BufferJoinStyle
	mitreLimit   float64
}

func newBufferOptionSet(opts []BufferOption) bufferOptionSet {
	bos := bufferOptionSet{
		quadSegments: 8,
		endCapStyle:  rawgeos.BufferEndCapStyleRound,
		joinStyle:    rawgeos.BufferJoinStyleRound,
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
		bos.endCapStyle = rawgeos.BufferEndCapStyleRound
	}
}

// BufferEndCapFlat sets the end cap style to 'flat'. It is 'round' by default.
func BufferEndCapFlat() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.endCapStyle = rawgeos.BufferEndCapStyleFlat
	}
}

// BufferEndCapSquare sets the end cap style to 'square'. It is 'round' by
// default.
func BufferEndCapSquare() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.endCapStyle = rawgeos.BufferEndCapStyleSquare
	}
}

// BufferJoinStyleRound sets the join style to 'round'. It is 'round' by
// default.
func BufferJoinStyleRound() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.joinStyle = rawgeos.BufferJoinStyleRound
		bos.mitreLimit = 0.0
	}
}

// BufferJoinStyleMitre sets the join style to 'mitre'. It is 'round' by
// default.
func BufferJoinStyleMitre(mitreLimit float64) BufferOption {
	return func(bos *bufferOptionSet) {
		bos.joinStyle = rawgeos.BufferJoinStyleMitre
		bos.mitreLimit = mitreLimit
	}
}

// BufferJoinStyleBevel sets the join style to 'bevel'. It is 'round' by
// default.
func BufferJoinStyleBevel() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.joinStyle = rawgeos.BufferJoinStyleBevel
		bos.mitreLimit = 0.0
	}
}

// Buffer returns a geometry that contains all points within the given radius
// of the input geometry.
//
// The validity of the result is not checked.
func Buffer(g geom.Geometry, radius float64, opts ...BufferOption) (geom.Geometry, error) {
	optSet := newBufferOptionSet(opts)
	return rawgeos.Buffer(g, rawgeos.BufferOptions{
		Radius:       radius,
		QuadSegments: optSet.quadSegments,
		EndCapStyle:  optSet.endCapStyle,
		JoinStyle:    optSet.joinStyle,
		MitreLimit:   optSet.mitreLimit,
	})
}

// Simplify creates a simplified version of a geometry using the
// Douglas-Peucker algorithm. Topological invariants may not be maintained,
// e.g. polygons can collapse into linestrings, and holes in polygons may be
// lost.
//
// The validity of the result is not checked.
func Simplify(g geom.Geometry, tolerance float64) (geom.Geometry, error) {
	return rawgeos.Simplify(g, tolerance)
}

// Difference returns the geometry that represents the parts of input geometry
// A that are not part of input geometry B.
//
// The validity of the result is not checked.
func Difference(a, b geom.Geometry) (geom.Geometry, error) {
	return rawgeos.Difference(a, b)
}

// SymmetricDifference returns the geometry that represents the parts of the
// input geometries that are not part of the other input geometry.
//
// The validity of the result is not checked.
func SymmetricDifference(a, b geom.Geometry) (geom.Geometry, error) {
	return rawgeos.SymmetricDifference(a, b)
}

// MakeValid can be used to convert an invalid geometry into a valid geometry.
// It does this by keeping the original control points and constructing a new
// geometry that is valid and similar (but not the same as) the original
// invalid geometry. If the input geometry is valid, then it is returned
// unaltered.
//
// The validity of the result is not checked.
func MakeValid(g geom.Geometry) (geom.Geometry, error) {
	return rawgeos.MakeValid(g)
}

// The CoverageUnion function is used to union polygonal inputs that form a
// coverage, which are typically provided as a GeometryCollection. This method
// is much faster than other unioning methods, but there are some constraints
// that must be met by the inputs to form a valid polygonal coverage. These
// constraints are:
//
//  1. all input geometries must be polygonal,
//  2. the interiors of the inputs must not intersect, and
//  3. the common boundaries of adjacent polygons have the same set of vertices in both polygons.
//
// It should be noted that while CoverageUnion may detect constraint violations
// and return an error, but this is not guaranteed, and an invalid result may
// be returned without an error. It is the responsibility of the caller to
// ensure that the constraints are met before using this function.
//
// The validity of the result is not checked.
func CoverageUnion(g geom.Geometry) (geom.Geometry, error) {
	return rawgeos.CoverageUnion(g)
}

// UnaryUnion is a single argument version of Union. It is most useful when
// supplied with a GeometryCollection, resulting in the union of the
// GeometryCollection's child geometries.
//
// The validity of the result is not checked.
func UnaryUnion(g geom.Geometry) (geom.Geometry, error) {
	return rawgeos.UnaryUnion(g)
}

// ConcaveHull returns concave hull of input geometry.
// pctconvex - ratio 0 to 1 (0 - max concaveness, 1 - convex hull)
// allowHoles - true to allow holes inside of polygons.
func ConcaveHull(g geom.Geometry, pctconvex float64, allowHoles bool) (geom.Geometry, error) {
	return rawgeos.ConcaveHull(g, pctconvex, allowHoles)
}

// ConvexHull returns convex hull of input geometry.
func ConvexHull(g geom.Geometry) (geom.Geometry, error) {
	return rawgeos.ConvexHull(g)
}
