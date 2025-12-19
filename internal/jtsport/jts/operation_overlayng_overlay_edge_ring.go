package jts

// OperationOverlayng_OverlayEdgeRing represents a ring of overlay edges that
// form a polygonal ring in the overlay result.
type OperationOverlayng_OverlayEdgeRing struct {
	startEdge *OperationOverlayng_OverlayEdge
	ring      *Geom_LinearRing
	isHole    bool
	ringPts   []*Geom_Coordinate
	locator   AlgorithmLocate_PointOnGeometryLocator
	shell     *OperationOverlayng_OverlayEdgeRing
	holes     []*OperationOverlayng_OverlayEdgeRing
}

// OperationOverlayng_NewOverlayEdgeRing creates a new OverlayEdgeRing from a
// starting edge.
func OperationOverlayng_NewOverlayEdgeRing(start *OperationOverlayng_OverlayEdge, geometryFactory *Geom_GeometryFactory) *OperationOverlayng_OverlayEdgeRing {
	oer := &OperationOverlayng_OverlayEdgeRing{
		startEdge: start,
		holes:     make([]*OperationOverlayng_OverlayEdgeRing, 0),
	}
	oer.ringPts = oer.computeRingPts(start)
	oer.computeRing(oer.ringPts, geometryFactory)
	return oer
}

// GetRing returns the LinearRing for this edge ring.
func (oer *OperationOverlayng_OverlayEdgeRing) GetRing() *Geom_LinearRing {
	return oer.ring
}

func (oer *OperationOverlayng_OverlayEdgeRing) getEnvelope() *Geom_Envelope {
	return oer.ring.GetEnvelopeInternal()
}

// IsHole tests whether this ring is a hole.
func (oer *OperationOverlayng_OverlayEdgeRing) IsHole() bool {
	return oer.isHole
}

// SetShell sets the containing shell ring of a ring that has been determined
// to be a hole.
func (oer *OperationOverlayng_OverlayEdgeRing) SetShell(shell *OperationOverlayng_OverlayEdgeRing) {
	oer.shell = shell
	if shell != nil {
		shell.AddHole(oer)
	}
}

// HasShell tests whether this ring has a shell assigned to it.
func (oer *OperationOverlayng_OverlayEdgeRing) HasShell() bool {
	return oer.shell != nil
}

// GetShell gets the shell for this ring. The shell is the ring itself if it is
// not a hole, otherwise its parent shell.
func (oer *OperationOverlayng_OverlayEdgeRing) GetShell() *OperationOverlayng_OverlayEdgeRing {
	if oer.IsHole() {
		return oer.shell
	}
	return oer
}

// AddHole adds a hole to this ring.
func (oer *OperationOverlayng_OverlayEdgeRing) AddHole(ring *OperationOverlayng_OverlayEdgeRing) {
	oer.holes = append(oer.holes, ring)
}

func (oer *OperationOverlayng_OverlayEdgeRing) computeRingPts(start *OperationOverlayng_OverlayEdge) []*Geom_Coordinate {
	edge := start
	pts := Geom_NewCoordinateList()
	for {
		if edge.GetEdgeRing() == oer {
			panic(Geom_NewTopologyExceptionWithCoordinate("Edge visited twice during ring-building at "+edge.GetCoordinate().String(), edge.GetCoordinate()))
		}

		edge.AddCoordinates(pts)
		edge.SetEdgeRing(oer)
		if edge.NextResult() == nil {
			panic(Geom_NewTopologyExceptionWithCoordinate("Found null edge in ring", edge.Dest()))
		}

		edge = edge.NextResult()
		if edge == start {
			break
		}
	}
	pts.CloseRing()
	return pts.ToCoordinateArray()
}

func (oer *OperationOverlayng_OverlayEdgeRing) computeRing(ringPts []*Geom_Coordinate, geometryFactory *Geom_GeometryFactory) {
	if oer.ring != nil {
		return // don't compute more than once
	}
	oer.ring = geometryFactory.CreateLinearRingFromCoordinates(ringPts)
	oer.isHole = Algorithm_Orientation_IsCCW(oer.ring.GetCoordinates())
}

func (oer *OperationOverlayng_OverlayEdgeRing) getCoordinates() []*Geom_Coordinate {
	return oer.ringPts
}

// FindEdgeRingContaining finds the innermost enclosing shell OverlayEdgeRing
// containing this OverlayEdgeRing, if any. The innermost enclosing ring is the
// smallest enclosing ring.
func (oer *OperationOverlayng_OverlayEdgeRing) FindEdgeRingContaining(erList []*OperationOverlayng_OverlayEdgeRing) *OperationOverlayng_OverlayEdgeRing {
	var minContainingRing *OperationOverlayng_OverlayEdgeRing

	for _, edgeRing := range erList {
		if edgeRing.contains(oer) {
			if minContainingRing == nil ||
				minContainingRing.getEnvelope().ContainsEnvelope(edgeRing.getEnvelope()) {
				minContainingRing = edgeRing
			}
		}
	}
	return minContainingRing
}

func (oer *OperationOverlayng_OverlayEdgeRing) getLocator() AlgorithmLocate_PointOnGeometryLocator {
	if oer.locator == nil {
		oer.locator = AlgorithmLocate_NewIndexedPointInAreaLocator(oer.GetRing().Geom_LineString.Geom_Geometry)
	}
	return oer.locator
}

// Locate returns the location of a coordinate relative to this ring.
func (oer *OperationOverlayng_OverlayEdgeRing) Locate(pt *Geom_Coordinate) int {
	return oer.getLocator().Locate(pt)
}

// contains tests if an edgeRing is properly contained in this ring. Relies on
// property that edgeRings never overlap (although they may touch at single
// vertices).
func (oer *OperationOverlayng_OverlayEdgeRing) contains(ring *OperationOverlayng_OverlayEdgeRing) bool {
	// the test envelope must be properly contained (guards against testing
	// rings against themselves)
	env := oer.getEnvelope()
	testEnv := ring.getEnvelope()
	if !env.ContainsProperly(testEnv) {
		return false
	}
	return oer.isPointInOrOut(ring)
}

func (oer *OperationOverlayng_OverlayEdgeRing) isPointInOrOut(ring *OperationOverlayng_OverlayEdgeRing) bool {
	// in most cases only one or two points will be checked
	for _, pt := range ring.getCoordinates() {
		loc := oer.Locate(pt)
		if loc == Geom_Location_Interior {
			return true
		}
		if loc == Geom_Location_Exterior {
			return false
		}
		// pt is on BOUNDARY, so keep checking for a determining location
	}
	return false
}

// GetCoordinate returns the first coordinate of the ring.
func (oer *OperationOverlayng_OverlayEdgeRing) GetCoordinate() *Geom_Coordinate {
	return oer.ringPts[0]
}

// ToPolygon computes the Polygon formed by this ring and any contained holes.
func (oer *OperationOverlayng_OverlayEdgeRing) ToPolygon(factory *Geom_GeometryFactory) *Geom_Polygon {
	var holeLR []*Geom_LinearRing
	if oer.holes != nil {
		holeLR = make([]*Geom_LinearRing, len(oer.holes))
		for i := range oer.holes {
			holeLR[i] = oer.holes[i].GetRing()
		}
	}
	return factory.CreatePolygonWithLinearRingAndHoles(oer.ring, holeLR)
}

// GetEdge returns the starting edge of this ring.
func (oer *OperationOverlayng_OverlayEdgeRing) GetEdge() *OperationOverlayng_OverlayEdge {
	return oer.startEdge
}
