package jts

// OperationOverlayng_RobustClipEnvelopeComputer computes a robust clipping
// envelope for a pair of polygonal geometries. The envelope is computed to be
// large enough to include the full length of all geometry line segments which
// intersect a given target envelope.
type OperationOverlayng_RobustClipEnvelopeComputer struct {
	targetEnv *Geom_Envelope
	clipEnv   *Geom_Envelope
}

// OperationOverlayng_RobustClipEnvelopeComputer_GetEnvelope computes the clip
// envelope for two geometries and a target envelope.
func OperationOverlayng_RobustClipEnvelopeComputer_GetEnvelope(a, b *Geom_Geometry, targetEnv *Geom_Envelope) *Geom_Envelope {
	cec := OperationOverlayng_NewRobustClipEnvelopeComputer(targetEnv)
	cec.Add(a)
	cec.Add(b)
	return cec.GetEnvelope()
}

// OperationOverlayng_NewRobustClipEnvelopeComputer creates a new
// RobustClipEnvelopeComputer for the given target envelope.
func OperationOverlayng_NewRobustClipEnvelopeComputer(targetEnv *Geom_Envelope) *OperationOverlayng_RobustClipEnvelopeComputer {
	return &OperationOverlayng_RobustClipEnvelopeComputer{
		targetEnv: targetEnv,
		clipEnv:   targetEnv.Copy(),
	}
}

// GetEnvelope returns the computed clip envelope.
func (rcec *OperationOverlayng_RobustClipEnvelopeComputer) GetEnvelope() *Geom_Envelope {
	return rcec.clipEnv
}

// Add adds a geometry to the envelope computation.
func (rcec *OperationOverlayng_RobustClipEnvelopeComputer) Add(g *Geom_Geometry) {
	if g == nil || g.IsEmpty() {
		return
	}

	if poly, ok := g.GetChild().(*Geom_Polygon); ok {
		rcec.addPolygon(poly)
	} else if gc, ok := g.GetChild().(*Geom_GeometryCollection); ok {
		rcec.addCollection(gc)
	}
}

func (rcec *OperationOverlayng_RobustClipEnvelopeComputer) addCollection(gc *Geom_GeometryCollection) {
	for i := 0; i < gc.GetNumGeometries(); i++ {
		g := gc.GetGeometryN(i)
		rcec.Add(g)
	}
}

func (rcec *OperationOverlayng_RobustClipEnvelopeComputer) addPolygon(poly *Geom_Polygon) {
	shell := poly.GetExteriorRing()
	rcec.addPolygonRing(shell)

	for i := 0; i < poly.GetNumInteriorRing(); i++ {
		hole := poly.GetInteriorRingN(i)
		rcec.addPolygonRing(hole)
	}
}

// addPolygonRing adds a polygon ring to the graph. Empty rings are ignored.
func (rcec *OperationOverlayng_RobustClipEnvelopeComputer) addPolygonRing(ring *Geom_LinearRing) {
	if ring.IsEmpty() {
		return
	}

	seq := ring.GetCoordinateSequence()
	for i := 1; i < seq.Size(); i++ {
		rcec.addSegment(seq.GetCoordinate(i-1), seq.GetCoordinate(i))
	}
}

func (rcec *OperationOverlayng_RobustClipEnvelopeComputer) addSegment(p1, p2 *Geom_Coordinate) {
	if operationOverlayng_RobustClipEnvelopeComputer_intersectsSegment(rcec.targetEnv, p1, p2) {
		rcec.clipEnv.ExpandToIncludeCoordinate(p1)
		rcec.clipEnv.ExpandToIncludeCoordinate(p2)
	}
}

func operationOverlayng_RobustClipEnvelopeComputer_intersectsSegment(env *Geom_Envelope, p1, p2 *Geom_Coordinate) bool {
	// This is a crude test of whether segment intersects envelope. It could be
	// refined by checking exact intersection.
	return env.IntersectsCoordinates(p1, p2)
}
