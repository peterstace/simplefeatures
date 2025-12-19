package jts

import "testing"

func TestRectangleLineIntersector_300Points(t *testing.T) {
	// TRANSLITERATION NOTE: This test requires buffer operation to generate
	// test points. Buffer is not yet ported, so this test is skipped.
	t.Skip("buffer operation not yet ported")

	validator := algorithm_newRectangleLineIntersectorValidator()
	validator.init(300)
	if !validator.validate() {
		t.Error("RectangleLineIntersector validation failed")
	}
}

// algorithm_rectangleLineIntersectorValidator tests optimized
// RectangleLineIntersector against a brute force approach (which is assumed to
// be correct).
type algorithm_rectangleLineIntersectorValidator struct {
	geomFact *Geom_GeometryFactory
	baseX    float64
	baseY    float64
	rectSize float64
	rectEnv  *Geom_Envelope
	pts      []*Geom_Coordinate
	isValid  bool
}

func algorithm_newRectangleLineIntersectorValidator() *algorithm_rectangleLineIntersectorValidator {
	return &algorithm_rectangleLineIntersectorValidator{
		geomFact: Geom_NewGeometryFactoryDefault(),
		baseX:    0,
		baseY:    0,
		rectSize: 100,
		isValid:  true,
	}
}

func (v *algorithm_rectangleLineIntersectorValidator) init(nPts int) {
	v.rectEnv = v.createRectangle()
	v.pts = v.createTestPoints(nPts)
}

func (v *algorithm_rectangleLineIntersectorValidator) validate() bool {
	v.run(true, true)
	return v.isValid
}

func (v *algorithm_rectangleLineIntersectorValidator) run(useSegInt, useSideInt bool) {
	rectSegIntersector := Algorithm_NewRectangleLineIntersector(v.rectEnv)
	rectSideIntersector := algorithm_newSimpleRectangleIntersector(v.rectEnv)

	for i := 0; i < len(v.pts); i++ {
		for j := 0; j < len(v.pts); j++ {
			if i == j {
				continue
			}

			segResult := false
			if useSegInt {
				segResult = rectSegIntersector.Intersects(v.pts[i], v.pts[j])
			}
			sideResult := false
			if useSideInt {
				sideResult = rectSideIntersector.intersects(v.pts[i], v.pts[j])
			}

			if useSegInt && useSideInt {
				if segResult != sideResult {
					v.isValid = false
				}
			}
		}
	}
}

func (v *algorithm_rectangleLineIntersectorValidator) createTestPoints(nPts int) []*Geom_Coordinate {
	pt := v.geomFact.CreatePointFromCoordinate(Geom_NewCoordinateWithXY(v.baseX, v.baseY))
	circle := pt.BufferWithQuadrantSegments(2*v.rectSize, nPts/4)
	return circle.GetCoordinates()
}

func (v *algorithm_rectangleLineIntersectorValidator) createRectangle() *Geom_Envelope {
	return Geom_NewEnvelopeFromCoordinates(
		Geom_NewCoordinateWithXY(v.baseX, v.baseY),
		Geom_NewCoordinateWithXY(v.baseX+v.rectSize, v.baseY+v.rectSize))
}

// algorithm_simpleRectangleIntersector is a brute force rectangle intersector
// for testing purposes.
type algorithm_simpleRectangleIntersector struct {
	li      *Algorithm_RobustLineIntersector
	rectEnv *Geom_Envelope
	// The corners of the rectangle, in the order:
	//  10
	//  23
	corner [4]*Geom_Coordinate
}

func algorithm_newSimpleRectangleIntersector(rectEnv *Geom_Envelope) *algorithm_simpleRectangleIntersector {
	s := &algorithm_simpleRectangleIntersector{
		li:      Algorithm_NewRobustLineIntersector(),
		rectEnv: rectEnv,
	}
	s.initCorners(rectEnv)
	return s
}

func (s *algorithm_simpleRectangleIntersector) initCorners(rectEnv *Geom_Envelope) {
	s.corner[0] = Geom_NewCoordinateWithXY(rectEnv.GetMaxX(), rectEnv.GetMaxY())
	s.corner[1] = Geom_NewCoordinateWithXY(rectEnv.GetMinX(), rectEnv.GetMaxY())
	s.corner[2] = Geom_NewCoordinateWithXY(rectEnv.GetMinX(), rectEnv.GetMinY())
	s.corner[3] = Geom_NewCoordinateWithXY(rectEnv.GetMaxX(), rectEnv.GetMinY())
}

func (s *algorithm_simpleRectangleIntersector) intersects(p0, p1 *Geom_Coordinate) bool {
	segEnv := Geom_NewEnvelopeFromCoordinates(p0, p1)
	if !s.rectEnv.IntersectsEnvelope(segEnv) {
		return false
	}

	s.li.ComputeIntersection(p0, p1, s.corner[0], s.corner[1])
	if s.li.HasIntersection() {
		return true
	}
	s.li.ComputeIntersection(p0, p1, s.corner[1], s.corner[2])
	if s.li.HasIntersection() {
		return true
	}
	s.li.ComputeIntersection(p0, p1, s.corner[2], s.corner[3])
	if s.li.HasIntersection() {
		return true
	}
	s.li.ComputeIntersection(p0, p1, s.corner[3], s.corner[0])
	if s.li.HasIntersection() {
		return true
	}

	return false
}
