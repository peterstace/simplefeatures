package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationPredicate_RectangleIntersects is an implementation of the
// intersects spatial predicate optimized for the case where one Geometry is a
// rectangle. This class works for all input geometries, including
// GeometryCollections.
//
// As a further optimization, this class can be used in batch style to test
// many geometries against a single rectangle.
type OperationPredicate_RectangleIntersects struct {
	rectangle *Geom_Polygon
	rectEnv   *Geom_Envelope
}

// OperationPredicate_RectangleIntersects_Intersects tests whether a rectangle
// intersects a given geometry.
func OperationPredicate_RectangleIntersects_Intersects(rectangle *Geom_Polygon, b *Geom_Geometry) bool {
	rp := OperationPredicate_NewRectangleIntersects(rectangle)
	return rp.Intersects(b)
}

// OperationPredicate_NewRectangleIntersects creates a new intersects computer
// for a rectangle.
func OperationPredicate_NewRectangleIntersects(rectangle *Geom_Polygon) *OperationPredicate_RectangleIntersects {
	return &OperationPredicate_RectangleIntersects{
		rectangle: rectangle,
		rectEnv:   rectangle.GetEnvelopeInternal(),
	}
}

// Intersects tests whether the given Geometry intersects the query rectangle.
func (r *OperationPredicate_RectangleIntersects) Intersects(geom *Geom_Geometry) bool {
	if !r.rectEnv.IntersectsEnvelope(geom.GetEnvelopeInternal()) {
		return false
	}

	// Test if rectangle envelope intersects any component envelope. This
	// handles Point components as well.
	visitor := operationPredicate_newEnvelopeIntersectsVisitor(r.rectEnv)
	visitor.ApplyTo(geom, visitor)
	if visitor.Intersects() {
		return true
	}

	// Test if any rectangle vertex is contained in the target geometry.
	ecpVisitor := operationPredicate_newGeometryContainsPointVisitor(r.rectangle)
	ecpVisitor.ApplyTo(geom, ecpVisitor)
	if ecpVisitor.ContainsPoint() {
		return true
	}

	// Test if any target geometry line segment intersects the rectangle.
	riVisitor := operationPredicate_newRectangleIntersectsSegmentVisitor(r.rectangle)
	riVisitor.ApplyTo(geom, riVisitor)
	if riVisitor.Intersects() {
		return true
	}

	return false
}

// operationPredicate_envelopeIntersectsVisitor tests whether it can be
// concluded that a rectangle intersects a geometry, based on the relationship
// of the envelope(s) of the geometry.
type operationPredicate_envelopeIntersectsVisitor struct {
	GeomUtil_ShortCircuitedGeometryVisitor
	rectEnv    *Geom_Envelope
	intersects bool
}

func operationPredicate_newEnvelopeIntersectsVisitor(rectEnv *Geom_Envelope) *operationPredicate_envelopeIntersectsVisitor {
	return &operationPredicate_envelopeIntersectsVisitor{
		rectEnv: rectEnv,
	}
}

// Intersects reports whether it can be concluded that an intersection occurs,
// or whether further testing is required.
func (v *operationPredicate_envelopeIntersectsVisitor) Intersects() bool {
	return v.intersects
}

func (v *operationPredicate_envelopeIntersectsVisitor) Visit(element *Geom_Geometry) {
	elementEnv := element.GetEnvelopeInternal()

	// Disjoint => no intersection.
	if !v.rectEnv.IntersectsEnvelope(elementEnv) {
		return
	}
	// Rectangle contains target env => must intersect.
	if v.rectEnv.ContainsEnvelope(elementEnv) {
		v.intersects = true
		return
	}
	// Since the envelopes intersect and the test element is connected, if the
	// test envelope is completely bisected by an edge of the rectangle the
	// element and the rectangle must touch (This is basically an application
	// of the Jordan Curve Theorem). The alternative situation is that the test
	// envelope is "on a corner" of the rectangle envelope, i.e. is not
	// completely bisected. In this case it is not possible to make a
	// conclusion about the presence of an intersection.
	if elementEnv.GetMinX() >= v.rectEnv.GetMinX() && elementEnv.GetMaxX() <= v.rectEnv.GetMaxX() {
		v.intersects = true
		return
	}
	if elementEnv.GetMinY() >= v.rectEnv.GetMinY() && elementEnv.GetMaxY() <= v.rectEnv.GetMaxY() {
		v.intersects = true
		return
	}
}

func (v *operationPredicate_envelopeIntersectsVisitor) IsDone() bool {
	return v.intersects
}

// operationPredicate_geometryContainsPointVisitor is a visitor which tests
// whether it can be concluded that a geometry contains a vertex of a query
// geometry.
type operationPredicate_geometryContainsPointVisitor struct {
	GeomUtil_ShortCircuitedGeometryVisitor
	rectSeq       Geom_CoordinateSequence
	rectEnv       *Geom_Envelope
	containsPoint bool
}

func operationPredicate_newGeometryContainsPointVisitor(rectangle *Geom_Polygon) *operationPredicate_geometryContainsPointVisitor {
	return &operationPredicate_geometryContainsPointVisitor{
		rectSeq: rectangle.GetExteriorRing().GetCoordinateSequence(),
		rectEnv: rectangle.GetEnvelopeInternal(),
	}
}

// ContainsPoint reports whether it can be concluded that a corner point of the
// rectangle is contained in the geometry, or whether further testing is
// required.
func (v *operationPredicate_geometryContainsPointVisitor) ContainsPoint() bool {
	return v.containsPoint
}

func (v *operationPredicate_geometryContainsPointVisitor) Visit(geom *Geom_Geometry) {
	// If test geometry is not polygonal this check is not needed.
	if !java.InstanceOf[*Geom_Polygon](geom) {
		return
	}

	// Skip if envelopes do not intersect.
	elementEnv := geom.GetEnvelopeInternal()
	if !v.rectEnv.IntersectsEnvelope(elementEnv) {
		return
	}

	// Test each corner of rectangle for inclusion.
	rectPt := Geom_NewCoordinate()
	for i := 0; i < 4; i++ {
		v.rectSeq.GetCoordinateInto(i, rectPt)
		if !elementEnv.ContainsCoordinate(rectPt) {
			continue
		}
		// Check rect point in poly (rect is known not to touch polygon at this point).
		if AlgorithmLocate_SimplePointInAreaLocator_ContainsPointInPolygon(rectPt, java.Cast[*Geom_Polygon](geom)) {
			v.containsPoint = true
			return
		}
	}
}

func (v *operationPredicate_geometryContainsPointVisitor) IsDone() bool {
	return v.containsPoint
}

// operationPredicate_rectangleIntersectsSegmentVisitor is a visitor to test
// for intersection between the query rectangle and the line segments of the
// geometry.
type operationPredicate_rectangleIntersectsSegmentVisitor struct {
	GeomUtil_ShortCircuitedGeometryVisitor
	rectEnv         *Geom_Envelope
	rectIntersector *Algorithm_RectangleLineIntersector
	hasIntersection bool
}

func operationPredicate_newRectangleIntersectsSegmentVisitor(rectangle *Geom_Polygon) *operationPredicate_rectangleIntersectsSegmentVisitor {
	rectEnv := rectangle.GetEnvelopeInternal()
	return &operationPredicate_rectangleIntersectsSegmentVisitor{
		rectEnv:         rectEnv,
		rectIntersector: Algorithm_NewRectangleLineIntersector(rectEnv),
	}
}

// Intersects reports whether any segment intersection exists.
func (v *operationPredicate_rectangleIntersectsSegmentVisitor) Intersects() bool {
	return v.hasIntersection
}

func (v *operationPredicate_rectangleIntersectsSegmentVisitor) Visit(geom *Geom_Geometry) {
	// It may be the case that the rectangle and the envelope of the geometry
	// component are disjoint, so it is worth checking this simple condition.
	elementEnv := geom.GetEnvelopeInternal()
	if !v.rectEnv.IntersectsEnvelope(elementEnv) {
		return
	}

	// Check segment intersections. Get all lines from geometry component
	// (there may be more than one if it's a multi-ring polygon).
	lines := GeomUtil_LinearComponentExtracter_GetLines(geom)
	v.checkIntersectionWithLineStrings(lines)
}

func (v *operationPredicate_rectangleIntersectsSegmentVisitor) checkIntersectionWithLineStrings(lines []*Geom_LineString) {
	for _, testLine := range lines {
		v.checkIntersectionWithSegments(testLine)
		if v.hasIntersection {
			return
		}
	}
}

func (v *operationPredicate_rectangleIntersectsSegmentVisitor) checkIntersectionWithSegments(testLine *Geom_LineString) {
	seq1 := testLine.GetCoordinateSequence()
	p0 := seq1.CreateCoordinate()
	p1 := seq1.CreateCoordinate()
	for j := 1; j < seq1.Size(); j++ {
		seq1.GetCoordinateInto(j-1, p0)
		seq1.GetCoordinateInto(j, p1)

		if v.rectIntersector.Intersects(p0, p1) {
			v.hasIntersection = true
			return
		}
	}
}

func (v *operationPredicate_rectangleIntersectsSegmentVisitor) IsDone() bool {
	return v.hasIntersection
}
