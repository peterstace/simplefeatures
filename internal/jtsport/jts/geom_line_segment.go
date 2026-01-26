package jts

import (
	"fmt"
	"math"
)

// Geom_LineSegment represents a line segment defined by two Coordinates.
// Provides methods to compute various geometric properties and relationships
// of line segments.
//
// This class is designed to be easily mutable (to the extent of having its
// contained points public). This supports a common pattern of reusing a single
// LineSegment object as a way of computing segment properties on the segments
// defined by arrays or lists of Coordinates.
type Geom_LineSegment struct {
	P0 *Geom_Coordinate
	P1 *Geom_Coordinate
}

// Geom_NewLineSegment creates a new LineSegment with default coordinates (0, 0).
func Geom_NewLineSegment() *Geom_LineSegment {
	return &Geom_LineSegment{
		P0: Geom_NewCoordinate(),
		P1: Geom_NewCoordinate(),
	}
}

// Geom_NewLineSegmentFromCoordinates creates a new LineSegment from two coordinates.
func Geom_NewLineSegmentFromCoordinates(p0, p1 *Geom_Coordinate) *Geom_LineSegment {
	return &Geom_LineSegment{
		P0: p0,
		P1: p1,
	}
}

// Geom_NewLineSegmentFromXY creates a new LineSegment from coordinate values.
func Geom_NewLineSegmentFromXY(x0, y0, x1, y1 float64) *Geom_LineSegment {
	return Geom_NewLineSegmentFromCoordinates(
		Geom_NewCoordinateWithXY(x0, y0),
		Geom_NewCoordinateWithXY(x1, y1),
	)
}

// Geom_NewLineSegmentFromLineSegment creates a copy of another LineSegment.
func Geom_NewLineSegmentFromLineSegment(ls *Geom_LineSegment) *Geom_LineSegment {
	return Geom_NewLineSegmentFromCoordinates(ls.P0, ls.P1)
}

// GetCoordinate gets the coordinate at the given index (0 or 1).
func (ls *Geom_LineSegment) GetCoordinate(i int) *Geom_Coordinate {
	if i == 0 {
		return ls.P0
	}
	return ls.P1
}

// SetCoordinatesFromLineSegment sets the coordinates from another LineSegment.
func (ls *Geom_LineSegment) SetCoordinatesFromLineSegment(other *Geom_LineSegment) {
	ls.SetCoordinates(other.P0, other.P1)
}

// SetCoordinates sets the coordinates of this segment.
func (ls *Geom_LineSegment) SetCoordinates(p0, p1 *Geom_Coordinate) {
	ls.P0.X = p0.X
	ls.P0.Y = p0.Y
	ls.P1.X = p1.X
	ls.P1.Y = p1.Y
}

// MinX gets the minimum X ordinate.
func (ls *Geom_LineSegment) MinX() float64 {
	return math.Min(ls.P0.X, ls.P1.X)
}

// MaxX gets the maximum X ordinate.
func (ls *Geom_LineSegment) MaxX() float64 {
	return math.Max(ls.P0.X, ls.P1.X)
}

// MinY gets the minimum Y ordinate.
func (ls *Geom_LineSegment) MinY() float64 {
	return math.Min(ls.P0.Y, ls.P1.Y)
}

// MaxY gets the maximum Y ordinate.
func (ls *Geom_LineSegment) MaxY() float64 {
	return math.Max(ls.P0.Y, ls.P1.Y)
}

// GetLength computes the length of the line segment.
func (ls *Geom_LineSegment) GetLength() float64 {
	return ls.P0.Distance(ls.P1)
}

// IsHorizontal tests whether the segment is horizontal.
func (ls *Geom_LineSegment) IsHorizontal() bool {
	return ls.P0.Y == ls.P1.Y
}

// IsVertical tests whether the segment is vertical.
func (ls *Geom_LineSegment) IsVertical() bool {
	return ls.P0.X == ls.P1.X
}

// OrientationIndexSegment determines the orientation of a LineSegment relative
// to this segment.
//
// Returns:
//
//	 1 if seg is to the left of this segment
//	-1 if seg is to the right of this segment
//	 0 if seg is collinear to or crosses this segment
func (ls *Geom_LineSegment) OrientationIndexSegment(seg *Geom_LineSegment) int {
	orient0 := Algorithm_Orientation_Index(ls.P0, ls.P1, seg.P0)
	orient1 := Algorithm_Orientation_Index(ls.P0, ls.P1, seg.P1)
	// This handles the case where the points are L or collinear.
	if orient0 >= 0 && orient1 >= 0 {
		if orient0 > orient1 {
			return orient0
		}
		return orient1
	}
	// This handles the case where the points are R or collinear.
	if orient0 <= 0 && orient1 <= 0 {
		if orient0 < orient1 {
			return orient0
		}
		return orient1
	}
	// Points lie on opposite sides => indeterminate orientation.
	return 0
}

// OrientationIndex determines the orientation index of a Coordinate relative
// to this segment.
//
// Returns:
//
//	 1 (LEFT) if p is to the left of this segment
//	-1 (RIGHT) if p is to the right of this segment
//	 0 (COLLINEAR) if p is collinear with this segment
func (ls *Geom_LineSegment) OrientationIndex(p *Geom_Coordinate) int {
	return Algorithm_Orientation_Index(ls.P0, ls.P1, p)
}

// Reverse reverses the direction of the line segment.
func (ls *Geom_LineSegment) Reverse() {
	ls.P0, ls.P1 = ls.P1, ls.P0
}

// Normalize puts the line segment into a normalized form. This is useful for
// using line segments in maps and indexes when topological equality rather
// than exact equality is desired. A segment in normalized form has the first
// point smaller than the second (according to the standard ordering on
// Coordinate).
func (ls *Geom_LineSegment) Normalize() {
	if ls.P1.CompareTo(ls.P0) < 0 {
		ls.Reverse()
	}
}

// Angle computes the angle that the vector defined by this segment makes with
// the X-axis. The angle will be in the range [ -PI, PI ] radians.
func (ls *Geom_LineSegment) Angle() float64 {
	return math.Atan2(ls.P1.Y-ls.P0.Y, ls.P1.X-ls.P0.X)
}

// MidPoint computes the midpoint of the segment.
func (ls *Geom_LineSegment) MidPoint() *Geom_Coordinate {
	return Geom_LineSegment_MidPoint(ls.P0, ls.P1)
}

// Geom_LineSegment_MidPoint computes the midpoint of a segment.
func Geom_LineSegment_MidPoint(p0, p1 *Geom_Coordinate) *Geom_Coordinate {
	return Geom_NewCoordinateWithXY((p0.X+p1.X)/2, (p0.Y+p1.Y)/2)
}

// DistanceToLineSegment computes the distance between this line segment and
// another segment.
func (ls *Geom_LineSegment) DistanceToLineSegment(other *Geom_LineSegment) float64 {
	return Algorithm_Distance_SegmentToSegment(ls.P0, ls.P1, other.P0, other.P1)
}

// DistanceToPoint computes the distance between this line segment and a given point.
func (ls *Geom_LineSegment) DistanceToPoint(p *Geom_Coordinate) float64 {
	return Algorithm_Distance_PointToSegment(p, ls.P0, ls.P1)
}

// DistancePerpendicular computes the perpendicular distance between the
// (infinite) line defined by this line segment and a point. If the segment has
// zero length this returns the distance between the segment and the point.
func (ls *Geom_LineSegment) DistancePerpendicular(p *Geom_Coordinate) float64 {
	if ls.P0.Equals2D(ls.P1) {
		return ls.P0.Distance(p)
	}
	return Algorithm_Distance_PointToLinePerpendicular(p, ls.P0, ls.P1)
}

// DistancePerpendicularOriented computes the oriented perpendicular distance
// between the (infinite) line defined by this line segment and a point. The
// oriented distance is positive if the point is on the left of the line, and
// negative if it is on the right. If the segment has zero length this returns
// the distance between the segment and the point.
func (ls *Geom_LineSegment) DistancePerpendicularOriented(p *Geom_Coordinate) float64 {
	if ls.P0.Equals2D(ls.P1) {
		return ls.P0.Distance(p)
	}
	dist := ls.DistancePerpendicular(p)
	if ls.OrientationIndex(p) < 0 {
		return -dist
	}
	return dist
}

// PointAlong computes the Coordinate that lies a given fraction along the line
// defined by this segment. A fraction of 0.0 returns the start point of the
// segment; a fraction of 1.0 returns the end point of the segment. If the
// fraction is < 0.0 or > 1.0 the point returned will lie before the start or
// beyond the end of the segment.
func (ls *Geom_LineSegment) PointAlong(segmentLengthFraction float64) *Geom_Coordinate {
	coord := ls.P0.Create()
	coord.X = ls.P0.X + segmentLengthFraction*(ls.P1.X-ls.P0.X)
	coord.Y = ls.P0.Y + segmentLengthFraction*(ls.P1.Y-ls.P0.Y)
	return coord
}

// PointAlongOffset computes the Coordinate that lies a given fraction along
// the line defined by this segment and offset from the segment by a given
// distance. A fraction of 0.0 offsets from the start point of the segment; a
// fraction of 1.0 offsets from the end point of the segment. The computed
// point is offset to the left of the line if the offset distance is positive,
// to the right if negative.
//
// Panics if the segment has zero length and offsetDistance is not 0.0.
func (ls *Geom_LineSegment) PointAlongOffset(segmentLengthFraction, offsetDistance float64) *Geom_Coordinate {
	// The point on the segment line.
	segx := ls.P0.X + segmentLengthFraction*(ls.P1.X-ls.P0.X)
	segy := ls.P0.Y + segmentLengthFraction*(ls.P1.Y-ls.P0.Y)

	dx := ls.P1.X - ls.P0.X
	dy := ls.P1.Y - ls.P0.Y
	length := math.Hypot(dx, dy)
	ux := 0.0
	uy := 0.0
	if offsetDistance != 0.0 {
		if length <= 0.0 {
			panic("Cannot compute offset from zero-length line segment")
		}
		// u is the vector that is the length of the offset, in the direction of the segment.
		ux = offsetDistance * dx / length
		uy = offsetDistance * dy / length
	}

	// The offset point is the seg point plus the offset vector rotated 90 degrees CCW.
	offsetx := segx - uy
	offsety := segy + ux

	coord := ls.P0.Create()
	coord.SetX(offsetx)
	coord.SetY(offsety)
	return coord
}

// ProjectionFactor computes the Projection Factor for the projection of the
// point p onto this LineSegment. The Projection Factor is the constant r by
// which the vector for this segment must be multiplied to equal the vector for
// the projection of p on the line defined by this segment.
//
// The projection factor will lie in the range (-inf, +inf), or be NaN if the
// line segment has zero length.
func (ls *Geom_LineSegment) ProjectionFactor(p *Geom_Coordinate) float64 {
	if p.Equals(ls.P0) {
		return 0.0
	}
	if p.Equals(ls.P1) {
		return 1.0
	}
	// Otherwise, use comp.graphics.algorithms Frequently Asked Questions method.
	dx := ls.P1.X - ls.P0.X
	dy := ls.P1.Y - ls.P0.Y
	length := dx*dx + dy*dy

	// Handle zero-length segments.
	if length <= 0.0 {
		return math.NaN()
	}

	r := ((p.X-ls.P0.X)*dx + (p.Y-ls.P0.Y)*dy) / length
	return r
}

// SegmentFraction computes the fraction of distance (in [0.0, 1.0]) that the
// projection of a point occurs along this line segment. If the point is beyond
// either ends of the line segment, the closest fractional value (0.0 or 1.0)
// is returned.
//
// Essentially, this is the ProjectionFactor clamped to the range [0.0, 1.0].
// If the segment has zero length, 1.0 is returned.
func (ls *Geom_LineSegment) SegmentFraction(inputPt *Geom_Coordinate) float64 {
	segFrac := ls.ProjectionFactor(inputPt)
	if segFrac < 0.0 {
		segFrac = 0.0
	} else if segFrac > 1.0 || math.IsNaN(segFrac) {
		segFrac = 1.0
	}
	return segFrac
}

// Project computes the projection of a point onto the line determined by this
// line segment.
//
// Note that the projected point may lie outside the line segment. If this is
// the case, the projection factor will lie outside the range [0.0, 1.0].
func (ls *Geom_LineSegment) Project(p *Geom_Coordinate) *Geom_Coordinate {
	if p.Equals(ls.P0) || p.Equals(ls.P1) {
		return p.Copy()
	}
	r := ls.ProjectionFactor(p)
	return ls.project(p, r)
}

func (ls *Geom_LineSegment) project(p *Geom_Coordinate, projectionFactor float64) *Geom_Coordinate {
	coord := p.Copy()
	coord.X = ls.P0.X + projectionFactor*(ls.P1.X-ls.P0.X)
	coord.Y = ls.P0.Y + projectionFactor*(ls.P1.Y-ls.P0.Y)
	return coord
}

// ProjectLineSegment projects a line segment onto this line segment and
// returns the resulting line segment. The returned line segment will be a
// subset of the target line segment. This subset may be nil if the segments
// are oriented in such a way that there is no projection.
//
// Note that the returned line may have zero length (i.e. the same endpoints).
// This can happen for instance if the lines are perpendicular to one another.
func (ls *Geom_LineSegment) ProjectLineSegment(seg *Geom_LineSegment) *Geom_LineSegment {
	pf0 := ls.ProjectionFactor(seg.P0)
	pf1 := ls.ProjectionFactor(seg.P1)
	// Check if segment projects at all.
	if pf0 >= 1.0 && pf1 >= 1.0 {
		return nil
	}
	if pf0 <= 0.0 && pf1 <= 0.0 {
		return nil
	}

	newp0 := ls.project(seg.P0, pf0)
	if pf0 < 0.0 {
		newp0 = ls.P0
	}
	if pf0 > 1.0 {
		newp0 = ls.P1
	}

	newp1 := ls.project(seg.P1, pf1)
	if pf1 < 0.0 {
		newp1 = ls.P0
	}
	if pf1 > 1.0 {
		newp1 = ls.P1
	}

	return Geom_NewLineSegmentFromCoordinates(newp0, newp1)
}

// Offset computes the LineSegment that is offset from the segment by a given
// distance. The computed segment is offset to the left of the line if the
// offset distance is positive, to the right if negative.
//
// Panics if the segment has zero length.
func (ls *Geom_LineSegment) Offset(offsetDistance float64) *Geom_LineSegment {
	offset0 := ls.PointAlongOffset(0, offsetDistance)
	offset1 := ls.PointAlongOffset(1, offsetDistance)
	return Geom_NewLineSegmentFromCoordinates(offset0, offset1)
}

// Reflect computes the reflection of a point in the line defined by this line
// segment.
func (ls *Geom_LineSegment) Reflect(p *Geom_Coordinate) *Geom_Coordinate {
	// General line equation.
	A := ls.P1.GetY() - ls.P0.GetY()
	B := ls.P0.GetX() - ls.P1.GetX()
	C := ls.P0.GetY()*(ls.P1.GetX()-ls.P0.GetX()) - ls.P0.GetX()*(ls.P1.GetY()-ls.P0.GetY())

	// Compute reflected point.
	A2plusB2 := A*A + B*B
	A2subB2 := A*A - B*B

	x := p.GetX()
	y := p.GetY()
	rx := (-A2subB2*x - 2*A*B*y - 2*A*C) / A2plusB2
	ry := (A2subB2*y - 2*A*B*x - 2*B*C) / A2plusB2

	coord := p.Copy()
	coord.SetX(rx)
	coord.SetY(ry)
	return coord
}

// ClosestPoint computes the closest point on this line segment to another point.
func (ls *Geom_LineSegment) ClosestPoint(p *Geom_Coordinate) *Geom_Coordinate {
	factor := ls.ProjectionFactor(p)
	if factor > 0 && factor < 1 {
		return ls.project(p, factor)
	}
	dist0 := ls.P0.Distance(p)
	dist1 := ls.P1.Distance(p)
	if dist0 < dist1 {
		return ls.P0
	}
	return ls.P1
}

// ClosestPoints computes the closest points on two line segments.
// Returns a pair of Coordinates which are the closest points on the line segments.
func (ls *Geom_LineSegment) ClosestPoints(line *Geom_LineSegment) []*Geom_Coordinate {
	// Test for intersection.
	intPt := ls.Intersection(line)
	if intPt != nil {
		return []*Geom_Coordinate{intPt, intPt}
	}

	// If no intersection, closest pair contains at least one endpoint.
	// Test each endpoint in turn.
	closestPt := make([]*Geom_Coordinate, 2)
	minDistance := math.MaxFloat64

	close00 := ls.ClosestPoint(line.P0)
	minDistance = close00.Distance(line.P0)
	closestPt[0] = close00
	closestPt[1] = line.P0

	close01 := ls.ClosestPoint(line.P1)
	dist := close01.Distance(line.P1)
	if dist < minDistance {
		minDistance = dist
		closestPt[0] = close01
		closestPt[1] = line.P1
	}

	close10 := line.ClosestPoint(ls.P0)
	dist = close10.Distance(ls.P0)
	if dist < minDistance {
		minDistance = dist
		closestPt[0] = ls.P0
		closestPt[1] = close10
	}

	close11 := line.ClosestPoint(ls.P1)
	dist = close11.Distance(ls.P1)
	if dist < minDistance {
		closestPt[0] = ls.P1
		closestPt[1] = close11
	}

	return closestPt
}

// Intersection computes an intersection point between two line segments, if
// there is one. There may be 0, 1 or many intersection points between two
// segments. If there are 0, nil is returned. If there is 1 or more, exactly
// one of them is returned (chosen at the discretion of the algorithm). If more
// information is required about the details of the intersection, the
// RobustLineIntersector class should be used.
func (ls *Geom_LineSegment) Intersection(line *Geom_LineSegment) *Geom_Coordinate {
	li := Algorithm_NewRobustLineIntersector()
	li.ComputeIntersection(ls.P0, ls.P1, line.P0, line.P1)
	if li.HasIntersection() {
		return li.GetIntersection(0)
	}
	return nil
}

// LineIntersection computes the intersection point of the lines of infinite
// extent defined by two line segments (if there is one). There may be 0, 1 or
// an infinite number of intersection points between two lines. If there is a
// unique intersection point, it is returned. Otherwise, nil is returned. If
// more information is required about the details of the intersection, the
// RobustLineIntersector class should be used.
func (ls *Geom_LineSegment) LineIntersection(line *Geom_LineSegment) *Geom_Coordinate {
	return Algorithm_Intersection_Intersection(ls.P0, ls.P1, line.P0, line.P1)
}

// ToGeometry creates a LineString with the same coordinates as this segment.
func (ls *Geom_LineSegment) ToGeometry(geomFactory *Geom_GeometryFactory) *Geom_LineString {
	return geomFactory.CreateLineStringFromCoordinates([]*Geom_Coordinate{ls.P0, ls.P1})
}

// Equals returns true if other has the same values for its points.
func (ls *Geom_LineSegment) Equals(other *Geom_LineSegment) bool {
	return ls.P0.Equals(other.P0) && ls.P1.Equals(other.P1)
}

// HashCode gets a hashcode for this object.
func (ls *Geom_LineSegment) HashCode() int {
	hash := 17
	hash = hash*29 + geom_lineSegment_hashCodeFloat64(ls.P0.X)
	hash = hash*29 + geom_lineSegment_hashCodeFloat64(ls.P0.Y)
	hash = hash*29 + geom_lineSegment_hashCodeFloat64(ls.P1.X)
	hash = hash*29 + geom_lineSegment_hashCodeFloat64(ls.P1.Y)
	return hash
}

func geom_lineSegment_hashCodeFloat64(x float64) int {
	bits := math.Float64bits(x)
	return int(bits ^ (bits >> 32))
}

// CompareTo compares this object with the specified object for order. Uses the
// standard lexicographic ordering for the points in the LineSegment.
//
// Returns a negative integer, zero, or a positive integer as this LineSegment
// is less than, equal to, or greater than the specified LineSegment.
func (ls *Geom_LineSegment) CompareTo(other *Geom_LineSegment) int {
	comp0 := ls.P0.CompareTo(other.P0)
	if comp0 != 0 {
		return comp0
	}
	return ls.P1.CompareTo(other.P1)
}

// EqualsTopo returns true if other is topologically equal to this LineSegment
// (e.g. irrespective of orientation).
func (ls *Geom_LineSegment) EqualsTopo(other *Geom_LineSegment) bool {
	return (ls.P0.Equals(other.P0) && ls.P1.Equals(other.P1)) ||
		(ls.P0.Equals(other.P1) && ls.P1.Equals(other.P0))
}

// String returns a string representation of this LineSegment.
func (ls *Geom_LineSegment) String() string {
	return fmt.Sprintf("LINESTRING (%v %v, %v %v)", ls.P0.X, ls.P0.Y, ls.P1.X, ls.P1.Y)
}
