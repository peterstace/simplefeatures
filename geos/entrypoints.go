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

// BufferOption allows the behaviour of the [Buffer] operation to be modified.
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

// TopologyPreserveSimplify creates a simplified version of a geometry using
// the Douglas-Peucker algorithm. An attempt is made to preserve topological
// invariants, e.g.  ring collapse and intersection.
//
// The validity of the result is not checked.
func TopologyPreserveSimplify(g geom.Geometry, tolerance float64) (geom.Geometry, error) {
	return rawgeos.TopologyPreserveSimplify(g, tolerance)
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
// coverage, provided as a [geom.GeometryCollection] of [geom.Polygon]s and/or [geom.MultiPolygon]s.
// This method is much faster than other unioning methods, but it relies on the
// input being a valid coverage (see the [CoverageIsValid] function for details).
//
// CoverageUnion may detect that the input is not a coverage and return an
// error, but this is not guaranteed (causing an invalid result to be returned
// without an error). It is the responsibility of the caller to ensure that the
// is valid before using this function.
//
// The validity of the result is not checked.
func CoverageUnion(g geom.Geometry) (geom.Geometry, error) {
	return rawgeos.CoverageUnion(g)
}

// CoverageSimplifyVW simplifies a polygon coverage, provided as a
// [geom.GeometryCollection] of [geom.Polygon]s and/or [geom.MultiPolygon]s. It uses the
// Visvalingam–Whyatt algorithm and relies on the coverage being valid (see the
// [CoverageIsValid] function for details).
//
// It may not check that the input forms a valid coverage, so it's possible
// that an incorrect result is returned without an error.
//
// The validity of the result is not checked.
func CoverageSimplifyVW(g geom.Geometry, tolerance float64, preserveBoundary bool) (geom.Geometry, error) {
	return rawgeos.CoverageSimplifyVW(g, tolerance, preserveBoundary)
}

// CoverageIsValid checks if a coverage (provided as a [geom.GeometryCollection]) is
// valid. Coverage validity is indicated by the boolean return value. A valid
// coverage must have the following properties:
//
//  1. all input geometries are of type [geom.Polygon] or [geom.MultiPolygon],
//
//  2. the interiors of the inputs do not intersect, and
//
//  3. the common boundaries of adjacent [geom.Polygon]s or [geom.MultiPolygon]s have the
//     same set of vertices.
//
// If the coverage is not valid, then the returned geometry shows the invalid
// edges.
func CoverageIsValid(g geom.Geometry, gapWidth float64) (bool, geom.Geometry, error) {
	return rawgeos.CoverageIsValid(g, gapWidth)
}

// UnaryUnion is a single argument version of [Union]. It is most useful when
// supplied with a [geom.GeometryCollection], resulting in the union of the
// [geom.GeometryCollection]'s child geometries.
//
// The validity of the result is not checked.
func UnaryUnion(g geom.Geometry) (geom.Geometry, error) {
	return rawgeos.UnaryUnion(g)
}

// ConcaveHull returns a concave hull of the input. A concave hull is generally
// a [geom.Polygon], but could also be a 2-point [geom.LineString] or a [geom.Point] in degenerate
// cases. It will be made of vertices that are a subset of the input vertices.
// The concavenessRatio parameter controls the concaveness of the hull (a value
// of 1 will produce convex hulls, and a value of 0 will produce maximally
// concave hulls). The allowHoles parameter controls whether holes are allowed
// in the hull.
func ConcaveHull(g geom.Geometry, concavenessRatio float64, allowHoles bool) (geom.Geometry, error) {
	return rawgeos.ConcaveHull(g, concavenessRatio, allowHoles)
}
