package jts

// OperationOverlayng_RingClipper clips rings of points to a rectangle. Uses a
// variant of Cohen-Sutherland clipping.
//
// In general the output is not topologically valid. In particular, the output
// may contain coincident non-noded line segments along the clip rectangle
// sides. However, the output is sufficiently well-structured that it can be
// used as input to the OverlayNG algorithm (which is able to process coincident
// linework due to the need to handle topology collapse under precision
// reduction).
//
// Because of the likelihood of creating extraneous line segments along the
// clipping rectangle sides, this class is not suitable for clipping linestrings.
type OperationOverlayng_RingClipper struct {
	clipEnv     *Geom_Envelope
	clipEnvMinY float64
	clipEnvMaxY float64
	clipEnvMinX float64
	clipEnvMaxX float64
}

const (
	operationOverlayng_RingClipper_BOX_BOTTOM = 0
	operationOverlayng_RingClipper_BOX_RIGHT  = 1
	operationOverlayng_RingClipper_BOX_TOP    = 2
	operationOverlayng_RingClipper_BOX_LEFT   = 3
)

// OperationOverlayng_NewRingClipper creates a new clipper for the given envelope.
func OperationOverlayng_NewRingClipper(clipEnv *Geom_Envelope) *OperationOverlayng_RingClipper {
	return &OperationOverlayng_RingClipper{
		clipEnv:     clipEnv,
		clipEnvMinY: clipEnv.GetMinY(),
		clipEnvMaxY: clipEnv.GetMaxY(),
		clipEnvMinX: clipEnv.GetMinX(),
		clipEnvMaxX: clipEnv.GetMaxX(),
	}
}

// Clip clips a list of points to the clipping rectangle box.
func (rc *OperationOverlayng_RingClipper) Clip(pts []*Geom_Coordinate) []*Geom_Coordinate {
	for edgeIndex := 0; edgeIndex < 4; edgeIndex++ {
		closeRing := edgeIndex == 3
		pts = rc.clipToBoxEdge(pts, edgeIndex, closeRing)
		if len(pts) == 0 {
			return pts
		}
	}
	return pts
}

// clipToBoxEdge clips line to the axis-parallel line defined by a single box
// edge.
func (rc *OperationOverlayng_RingClipper) clipToBoxEdge(pts []*Geom_Coordinate, edgeIndex int, closeRing bool) []*Geom_Coordinate {
	ptsClip := Geom_NewCoordinateList()

	p0 := pts[len(pts)-1]
	for i := 0; i < len(pts); i++ {
		p1 := pts[i]
		if rc.isInsideEdge(p1, edgeIndex) {
			if !rc.isInsideEdge(p0, edgeIndex) {
				intPt := rc.intersection(p0, p1, edgeIndex)
				ptsClip.AddCoordinate(intPt, false)
			}
			ptsClip.AddCoordinate(p1.Copy(), false)
		} else if rc.isInsideEdge(p0, edgeIndex) {
			intPt := rc.intersection(p0, p1, edgeIndex)
			ptsClip.AddCoordinate(intPt, false)
		}
		// else p0-p1 is outside box, so it is dropped

		p0 = p1
	}

	// Add closing point if required.
	if closeRing && ptsClip.Size() > 0 {
		start := ptsClip.GetCoordinate(0)
		if !start.Equals2D(ptsClip.GetCoordinate(ptsClip.Size() - 1)) {
			ptsClip.AddCoordinate(start.Copy(), true)
		}
	}
	return ptsClip.ToCoordinateArray()
}

// intersection computes the intersection point of a segment with an edge of
// the clip box.
func (rc *OperationOverlayng_RingClipper) intersection(a, b *Geom_Coordinate, edgeIndex int) *Geom_Coordinate {
	var intPt *Geom_Coordinate
	switch edgeIndex {
	case operationOverlayng_RingClipper_BOX_BOTTOM:
		intPt = Geom_NewCoordinateXY2DWithXY(rc.intersectionLineY(a, b, rc.clipEnvMinY), rc.clipEnvMinY).Geom_Coordinate
	case operationOverlayng_RingClipper_BOX_RIGHT:
		intPt = Geom_NewCoordinateXY2DWithXY(rc.clipEnvMaxX, rc.intersectionLineX(a, b, rc.clipEnvMaxX)).Geom_Coordinate
	case operationOverlayng_RingClipper_BOX_TOP:
		intPt = Geom_NewCoordinateXY2DWithXY(rc.intersectionLineY(a, b, rc.clipEnvMaxY), rc.clipEnvMaxY).Geom_Coordinate
	case operationOverlayng_RingClipper_BOX_LEFT:
		intPt = Geom_NewCoordinateXY2DWithXY(rc.clipEnvMinX, rc.intersectionLineX(a, b, rc.clipEnvMinX)).Geom_Coordinate
	default:
		intPt = Geom_NewCoordinateXY2DWithXY(rc.clipEnvMinX, rc.intersectionLineX(a, b, rc.clipEnvMinX)).Geom_Coordinate
	}
	return intPt
}

func (rc *OperationOverlayng_RingClipper) intersectionLineY(a, b *Geom_Coordinate, y float64) float64 {
	m := (b.GetX() - a.GetX()) / (b.GetY() - a.GetY())
	intercept := (y - a.GetY()) * m
	return a.GetX() + intercept
}

func (rc *OperationOverlayng_RingClipper) intersectionLineX(a, b *Geom_Coordinate, x float64) float64 {
	m := (b.GetY() - a.GetY()) / (b.GetX() - a.GetX())
	intercept := (x - a.GetX()) * m
	return a.GetY() + intercept
}

func (rc *OperationOverlayng_RingClipper) isInsideEdge(p *Geom_Coordinate, edgeIndex int) bool {
	isInside := false
	switch edgeIndex {
	case operationOverlayng_RingClipper_BOX_BOTTOM:
		isInside = p.GetY() > rc.clipEnvMinY
	case operationOverlayng_RingClipper_BOX_RIGHT:
		isInside = p.GetX() < rc.clipEnvMaxX
	case operationOverlayng_RingClipper_BOX_TOP:
		isInside = p.GetY() < rc.clipEnvMaxY
	case operationOverlayng_RingClipper_BOX_LEFT:
		isInside = p.GetX() > rc.clipEnvMinX
	default:
		isInside = p.GetX() > rc.clipEnvMinX
	}
	return isInside
}
