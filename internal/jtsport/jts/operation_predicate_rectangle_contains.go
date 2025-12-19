package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationPredicate_RectangleContains is an optimized implementation of the
// contains spatial predicate for cases where the first Geometry is a
// rectangle. This class works for all input geometries, including
// GeometryCollections.
//
// As a further optimization, this class can be used to test many geometries
// against a single rectangle in a slightly more efficient way.
type OperationPredicate_RectangleContains struct {
	rectEnv *Geom_Envelope
}

// OperationPredicate_RectangleContains_Contains tests whether a rectangle
// contains a given geometry.
func OperationPredicate_RectangleContains_Contains(rectangle *Geom_Polygon, b *Geom_Geometry) bool {
	rc := OperationPredicate_NewRectangleContains(rectangle)
	return rc.Contains(b)
}

// OperationPredicate_NewRectangleContains creates a new contains computer for
// two geometries.
func OperationPredicate_NewRectangleContains(rectangle *Geom_Polygon) *OperationPredicate_RectangleContains {
	return &OperationPredicate_RectangleContains{
		rectEnv: rectangle.GetEnvelopeInternal(),
	}
}

func (rc *OperationPredicate_RectangleContains) Contains(geom *Geom_Geometry) bool {
	// The test geometry must be wholly contained in the rectangle envelope.
	if !rc.rectEnv.ContainsEnvelope(geom.GetEnvelopeInternal()) {
		return false
	}

	// Check that geom is not contained entirely in the rectangle boundary.
	// According to the somewhat odd spec of the SFS, if this is the case the
	// geometry is NOT contained.
	if rc.isContainedInBoundary(geom) {
		return false
	}
	return true
}

func (rc *OperationPredicate_RectangleContains) isContainedInBoundary(geom *Geom_Geometry) bool {
	// Polygons can never be wholly contained in the boundary.
	if java.InstanceOf[*Geom_Polygon](geom) {
		return false
	}
	if java.InstanceOf[*Geom_Point](geom) {
		return rc.isPointContainedInBoundary(java.Cast[*Geom_Point](geom))
	}
	if java.InstanceOf[*Geom_LineString](geom) {
		return rc.isLineStringContainedInBoundary(java.Cast[*Geom_LineString](geom))
	}

	for i := 0; i < geom.GetNumGeometries(); i++ {
		comp := geom.GetGeometryN(i)
		if !rc.isContainedInBoundary(comp) {
			return false
		}
	}
	return true
}

func (rc *OperationPredicate_RectangleContains) isPointContainedInBoundary(point *Geom_Point) bool {
	return rc.isCoordContainedInBoundary(point.GetCoordinate())
}

// isCoordContainedInBoundary tests if a point is contained in the boundary of
// the target rectangle.
func (rc *OperationPredicate_RectangleContains) isCoordContainedInBoundary(pt *Geom_Coordinate) bool {
	// contains = false if the point is properly contained in the rectangle.
	// This code assumes that the point lies in the rectangle envelope.
	return pt.GetX() == rc.rectEnv.GetMinX() ||
		pt.GetX() == rc.rectEnv.GetMaxX() ||
		pt.GetY() == rc.rectEnv.GetMinY() ||
		pt.GetY() == rc.rectEnv.GetMaxY()
}

// isLineStringContainedInBoundary tests if a linestring is completely
// contained in the boundary of the target rectangle.
func (rc *OperationPredicate_RectangleContains) isLineStringContainedInBoundary(line *Geom_LineString) bool {
	seq := line.GetCoordinateSequence()
	p0 := Geom_NewCoordinate()
	p1 := Geom_NewCoordinate()
	for i := 0; i < seq.Size()-1; i++ {
		seq.GetCoordinateInto(i, p0)
		seq.GetCoordinateInto(i+1, p1)

		if !rc.isLineSegmentContainedInBoundary(p0, p1) {
			return false
		}
	}
	return true
}

// isLineSegmentContainedInBoundary tests if a line segment is contained in the
// boundary of the target rectangle.
func (rc *OperationPredicate_RectangleContains) isLineSegmentContainedInBoundary(p0, p1 *Geom_Coordinate) bool {
	if p0.Equals(p1) {
		return rc.isCoordContainedInBoundary(p0)
	}

	// We already know that the segment is contained in the rectangle envelope.
	if p0.GetX() == p1.GetX() {
		if p0.GetX() == rc.rectEnv.GetMinX() || p0.GetX() == rc.rectEnv.GetMaxX() {
			return true
		}
	} else if p0.GetY() == p1.GetY() {
		if p0.GetY() == rc.rectEnv.GetMinY() || p0.GetY() == rc.rectEnv.GetMaxY() {
			return true
		}
	}
	// Either both x and y values are different or one of x and y are the same,
	// but the other ordinate is not the same as a boundary ordinate.
	// In either case, the segment is not wholly in the boundary.
	return false
}
