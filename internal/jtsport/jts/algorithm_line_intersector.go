package jts

import (
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Algorithm_LineIntersector is an algorithm that can both test whether two line
// segments intersect and compute the intersection point(s) if they do.
//
// There are three possible outcomes when determining whether two line segments
// intersect:
//   - NO_INTERSECTION - the segments do not intersect
//   - POINT_INTERSECTION - the segments intersect in a single point
//   - COLLINEAR_INTERSECTION - the segments are collinear and they intersect in
//     a line segment
//
// For segments which intersect in a single point, the point may be either an
// endpoint or in the interior of each segment. If the point lies in the
// interior of both segments, this is termed a proper intersection. The method
// IsProper() tests for this situation.
//
// The intersection point(s) may be computed in a precise or non-precise manner.
// Computing an intersection point precisely involves rounding it via a supplied
// PrecisionModel.
//
// LineIntersectors do not perform an initial envelope intersection test to
// determine if the segments are disjoint. This is because this class is likely
// to be used in a context where envelope overlap is already known to occur (or
// be likely).
type Algorithm_LineIntersector struct {
	child java.Polymorphic

	result         int
	inputLines     [2][2]*Geom_Coordinate
	intPt          [2]*Geom_Coordinate
	intLineIndex   [][]int
	isProper       bool
	pa             *Geom_Coordinate
	pb             *Geom_Coordinate
	precisionModel *Geom_PrecisionModel
}

// Deprecated constants (due to ambiguous naming).
const (
	Algorithm_LineIntersector_DontIntersect = 0
	Algorithm_LineIntersector_DoIntersect   = 1
	Algorithm_LineIntersector_Collinear     = 2
)

// Intersection result constants.
const (
	// Algorithm_LineIntersector_NoIntersection indicates that line segments do
	// not intersect.
	Algorithm_LineIntersector_NoIntersection = 0
	// Algorithm_LineIntersector_PointIntersection indicates that line segments
	// intersect in a single point.
	Algorithm_LineIntersector_PointIntersection = 1
	// Algorithm_LineIntersector_CollinearIntersection indicates that line
	// segments intersect in a line segment.
	Algorithm_LineIntersector_CollinearIntersection = 2
)

// Algorithm_LineIntersector_ComputeEdgeDistance computes the "edge distance" of
// an intersection point p along a segment. The edge distance is a metric of the
// point along the edge. The metric used is a robust and easy to compute metric
// function. It is not equivalent to the usual Euclidean metric. It relies on
// the fact that either the x or the y ordinates of the points in the edge are
// unique, depending on whether the edge is longer in the horizontal or vertical
// direction.
//
// NOTE: This function may produce incorrect distances for inputs where p is not
// precisely on p0-p2 (E.g. p = (139,9) p0 = (139,10), p1 = (280,1) produces
// distance 0.0, which is incorrect.
//
// My hypothesis is that the function is safe to use for points which are the
// result of rounding points which lie on the line, but not safe to use for
// truncated points.
func Algorithm_LineIntersector_ComputeEdgeDistance(p, p0, p1 *Geom_Coordinate) float64 {
	dx := math.Abs(p1.GetX() - p0.GetX())
	dy := math.Abs(p1.GetY() - p0.GetY())

	dist := -1.0 // sentinel value
	if p.Equals(p0) {
		dist = 0.0
	} else if p.Equals(p1) {
		if dx > dy {
			dist = dx
		} else {
			dist = dy
		}
	} else {
		pdx := math.Abs(p.GetX() - p0.GetX())
		pdy := math.Abs(p.GetY() - p0.GetY())
		if dx > dy {
			dist = pdx
		} else {
			dist = pdy
		}
		// Hack to ensure that non-endpoints always have a non-zero distance.
		if dist == 0.0 && !p.Equals(p0) {
			dist = math.Max(pdx, pdy)
		}
	}
	Util_Assert_IsTrueWithMessage(!(dist == 0.0 && !p.Equals(p0)), "Bad distance calculation")
	return dist
}

// Algorithm_LineIntersector_NonRobustComputeEdgeDistance computes edge distance
// using Euclidean distance.
func Algorithm_LineIntersector_NonRobustComputeEdgeDistance(p, p1, p2 *Geom_Coordinate) float64 {
	dx := p.GetX() - p1.GetX()
	dy := p.GetY() - p1.GetY()
	dist := math.Hypot(dx, dy)
	Util_Assert_IsTrueWithMessage(!(dist == 0.0 && !p.Equals(p1)), "Invalid distance calculation")
	return dist
}

// GetChild returns the immediate child in the type hierarchy chain.
func (li *Algorithm_LineIntersector) GetChild() java.Polymorphic {
	return li.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (li *Algorithm_LineIntersector) GetParent() java.Polymorphic {
	return nil
}

// Algorithm_NewLineIntersector creates a new LineIntersector.
func Algorithm_NewLineIntersector() *Algorithm_LineIntersector {
	li := &Algorithm_LineIntersector{}
	li.intPt[0] = Geom_NewCoordinate()
	li.intPt[1] = Geom_NewCoordinate()
	// Alias the intersection points for ease of reference.
	li.pa = li.intPt[0]
	li.pb = li.intPt[1]
	li.result = 0
	return li
}

// SetMakePrecise forces computed intersection to be rounded to a given
// precision model.
//
// Deprecated: use SetPrecisionModel instead.
func (li *Algorithm_LineIntersector) SetMakePrecise(precisionModel *Geom_PrecisionModel) {
	li.precisionModel = precisionModel
}

// SetPrecisionModel forces computed intersection to be rounded to a given
// precision model. No getter is provided, because the precision model is not
// required to be specified.
func (li *Algorithm_LineIntersector) SetPrecisionModel(precisionModel *Geom_PrecisionModel) {
	li.precisionModel = precisionModel
}

// GetEndpoint gets an endpoint of an input segment.
func (li *Algorithm_LineIntersector) GetEndpoint(segmentIndex, ptIndex int) *Geom_Coordinate {
	return li.inputLines[segmentIndex][ptIndex]
}

// ComputeIntersectionPointLine computes the intersection of a point p and the
// line p1-p2. This function computes the boolean value of the hasIntersection
// test. The actual value of the intersection (if there is one) is equal to the
// value of p.
func (li *Algorithm_LineIntersector) ComputeIntersectionPointLine(p, p1, p2 *Geom_Coordinate) {
	if impl, ok := java.GetLeaf(li).(interface {
		ComputeIntersectionPointLine_BODY(*Geom_Coordinate, *Geom_Coordinate, *Geom_Coordinate)
	}); ok {
		impl.ComputeIntersectionPointLine_BODY(p, p1, p2)
		return
	}
	panic("abstract method called")
}

func (li *Algorithm_LineIntersector) isCollinear() bool {
	return li.result == Algorithm_LineIntersector_CollinearIntersection
}

// ComputeIntersection computes the intersection of the lines p1-p2 and p3-p4.
// This function computes both the boolean value of the hasIntersection test and
// the (approximate) value of the intersection point itself (if there is one).
func (li *Algorithm_LineIntersector) ComputeIntersection(p1, p2, p3, p4 *Geom_Coordinate) {
	li.inputLines[0][0] = p1
	li.inputLines[0][1] = p2
	li.inputLines[1][0] = p3
	li.inputLines[1][1] = p4
	li.result = li.computeIntersect(p1, p2, p3, p4)
}

func (li *Algorithm_LineIntersector) computeIntersect(p1, p2, q1, q2 *Geom_Coordinate) int {
	if impl, ok := java.GetLeaf(li).(interface {
		computeIntersect_BODY(*Geom_Coordinate, *Geom_Coordinate, *Geom_Coordinate, *Geom_Coordinate) int
	}); ok {
		return impl.computeIntersect_BODY(p1, p2, q1, q2)
	}
	panic("abstract method called")
}

// String returns a string representation.
func (li *Algorithm_LineIntersector) String() string {
	return fmt.Sprintf("LINESTRING (%v %v, %v %v) - LINESTRING (%v %v, %v %v)%s",
		li.inputLines[0][0].GetX(), li.inputLines[0][0].GetY(),
		li.inputLines[0][1].GetX(), li.inputLines[0][1].GetY(),
		li.inputLines[1][0].GetX(), li.inputLines[1][0].GetY(),
		li.inputLines[1][1].GetX(), li.inputLines[1][1].GetY(),
		li.getTopologySummary())
}

func (li *Algorithm_LineIntersector) getTopologySummary() string {
	var result string
	if li.isEndPoint() {
		result += " endpoint"
	}
	if li.isProper {
		result += " proper"
	}
	if li.isCollinear() {
		result += " collinear"
	}
	return result
}

func (li *Algorithm_LineIntersector) isEndPoint() bool {
	return li.HasIntersection() && !li.isProper
}

// HasIntersection tests whether the input geometries intersect.
func (li *Algorithm_LineIntersector) HasIntersection() bool {
	return li.result != Algorithm_LineIntersector_NoIntersection
}

// GetIntersectionNum returns the number of intersection points found. This will
// be either 0, 1 or 2.
func (li *Algorithm_LineIntersector) GetIntersectionNum() int {
	return li.result
}

// GetIntersection returns the intIndex'th intersection point.
func (li *Algorithm_LineIntersector) GetIntersection(intIndex int) *Geom_Coordinate {
	return li.intPt[intIndex]
}

func (li *Algorithm_LineIntersector) computeIntLineIndex() {
	if li.intLineIndex == nil {
		li.intLineIndex = make([][]int, 2)
		li.intLineIndex[0] = make([]int, 2)
		li.intLineIndex[1] = make([]int, 2)
		li.computeIntLineIndex_segment(0)
		li.computeIntLineIndex_segment(1)
	}
}

// IsIntersection tests whether a point is an intersection point of two line
// segments. Note that if the intersection is a line segment, this method only
// tests for equality with the endpoints of the intersection segment. It does
// not return true if the input point is internal to the intersection segment.
func (li *Algorithm_LineIntersector) IsIntersection(pt *Geom_Coordinate) bool {
	for i := 0; i < li.result; i++ {
		if li.intPt[i].Equals2D(pt) {
			return true
		}
	}
	return false
}

// IsInteriorIntersection tests whether either intersection point is an interior
// point of one of the input segments.
func (li *Algorithm_LineIntersector) IsInteriorIntersection() bool {
	if li.IsInteriorIntersectionFor(0) {
		return true
	}
	if li.IsInteriorIntersectionFor(1) {
		return true
	}
	return false
}

// IsInteriorIntersectionFor tests whether either intersection point is an
// interior point of the specified input segment.
func (li *Algorithm_LineIntersector) IsInteriorIntersectionFor(inputLineIndex int) bool {
	for i := 0; i < li.result; i++ {
		if !(li.intPt[i].Equals2D(li.inputLines[inputLineIndex][0]) ||
			li.intPt[i].Equals2D(li.inputLines[inputLineIndex][1])) {
			return true
		}
	}
	return false
}

// IsProper tests whether an intersection is proper.
//
// The intersection between two line segments is considered proper if they
// intersect in a single point in the interior of both segments (e.g. the
// intersection is a single point and is not equal to any of the endpoints).
//
// The intersection between a point and a line segment is considered proper if
// the point lies in the interior of the segment (e.g. is not equal to either of
// the endpoints).
func (li *Algorithm_LineIntersector) IsProper() bool {
	return li.HasIntersection() && li.isProper
}

// GetIntersectionAlongSegment computes the intIndex'th intersection point in
// the direction of a specified input line segment.
func (li *Algorithm_LineIntersector) GetIntersectionAlongSegment(segmentIndex, intIndex int) *Geom_Coordinate {
	// Lazily compute int line array.
	li.computeIntLineIndex()
	return li.intPt[li.intLineIndex[segmentIndex][intIndex]]
}

// GetIndexAlongSegment computes the index (order) of the intIndex'th
// intersection point in the direction of a specified input line segment.
func (li *Algorithm_LineIntersector) GetIndexAlongSegment(segmentIndex, intIndex int) int {
	li.computeIntLineIndex()
	return li.intLineIndex[segmentIndex][intIndex]
}

func (li *Algorithm_LineIntersector) computeIntLineIndex_segment(segmentIndex int) {
	dist0 := li.GetEdgeDistance(segmentIndex, 0)
	dist1 := li.GetEdgeDistance(segmentIndex, 1)
	if dist0 > dist1 {
		li.intLineIndex[segmentIndex][0] = 0
		li.intLineIndex[segmentIndex][1] = 1
	} else {
		li.intLineIndex[segmentIndex][0] = 1
		li.intLineIndex[segmentIndex][1] = 0
	}
}

// GetEdgeDistance computes the "edge distance" of an intersection point along
// the specified input line segment.
func (li *Algorithm_LineIntersector) GetEdgeDistance(segmentIndex, intIndex int) float64 {
	dist := Algorithm_LineIntersector_ComputeEdgeDistance(li.intPt[intIndex],
		li.inputLines[segmentIndex][0], li.inputLines[segmentIndex][1])
	return dist
}
