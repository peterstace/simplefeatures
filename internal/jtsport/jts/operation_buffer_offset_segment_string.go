package jts

// operationBuffer_OffsetSegmentString is a dynamic list of the vertices in a constructed offset curve.
// Automatically removes adjacent vertices which are closer than a given tolerance.
type operationBuffer_OffsetSegmentString struct {
	ptList         []*Geom_Coordinate
	precisionModel *Geom_PrecisionModel

	// minimimVertexDistance is the distance below which two adjacent points on the curve
	// are considered to be coincident.
	// This is chosen to be a small fraction of the offset distance.
	minimimVertexDistance float64
}

// operationBuffer_newOffsetSegmentString creates a new OffsetSegmentString.
func operationBuffer_newOffsetSegmentString() *operationBuffer_OffsetSegmentString {
	return &operationBuffer_OffsetSegmentString{
		ptList: make([]*Geom_Coordinate, 0),
	}
}

// SetPrecisionModel sets the precision model for this offset segment string.
func (oss *operationBuffer_OffsetSegmentString) SetPrecisionModel(precisionModel *Geom_PrecisionModel) {
	oss.precisionModel = precisionModel
}

// SetMinimumVertexDistance sets the minimum vertex distance.
func (oss *operationBuffer_OffsetSegmentString) SetMinimumVertexDistance(minimimVertexDistance float64) {
	oss.minimimVertexDistance = minimimVertexDistance
}

// AddPt adds a point to the offset segment string.
func (oss *operationBuffer_OffsetSegmentString) AddPt(pt *Geom_Coordinate) {
	bufPt := Geom_NewCoordinateFromCoordinate(pt)
	oss.precisionModel.MakePreciseCoordinate(bufPt)
	// don't add duplicate (or near-duplicate) points
	if oss.isRedundant(bufPt) {
		return
	}
	oss.ptList = append(oss.ptList, bufPt)
}

// AddPts adds an array of points to the offset segment string.
func (oss *operationBuffer_OffsetSegmentString) AddPts(pt []*Geom_Coordinate, isForward bool) {
	if isForward {
		for i := 0; i < len(pt); i++ {
			oss.AddPt(pt[i])
		}
	} else {
		for i := len(pt) - 1; i >= 0; i-- {
			oss.AddPt(pt[i])
		}
	}
}

// isRedundant tests whether the given point is redundant
// relative to the previous point in the list (up to tolerance).
func (oss *operationBuffer_OffsetSegmentString) isRedundant(pt *Geom_Coordinate) bool {
	if len(oss.ptList) < 1 {
		return false
	}
	lastPt := oss.ptList[len(oss.ptList)-1]
	ptDist := pt.Distance(lastPt)
	if ptDist < oss.minimimVertexDistance {
		return true
	}
	return false
}

// CloseRing closes the ring by adding the first point at the end if needed.
func (oss *operationBuffer_OffsetSegmentString) CloseRing() {
	if len(oss.ptList) < 1 {
		return
	}
	startPt := Geom_NewCoordinateFromCoordinate(oss.ptList[0])
	lastPt := oss.ptList[len(oss.ptList)-1]
	if startPt.Equals(lastPt) {
		return
	}
	oss.ptList = append(oss.ptList, startPt)
}

// Reverse reverses the order of points in the list.
func (oss *operationBuffer_OffsetSegmentString) Reverse() {
	// The Java implementation is empty, so we keep it empty too
}

// GetCoordinates returns the coordinates as an array.
func (oss *operationBuffer_OffsetSegmentString) GetCoordinates() []*Geom_Coordinate {
	coord := make([]*Geom_Coordinate, len(oss.ptList))
	copy(coord, oss.ptList)
	return coord
}

// String returns a string representation of this offset segment string.
func (oss *operationBuffer_OffsetSegmentString) String() string {
	fact := Geom_NewGeometryFactoryDefault()
	line := fact.CreateLineStringFromCoordinates(oss.GetCoordinates())
	return Io_WKTWriter_ToLineStringFromCoords(line.GetCoordinates())
}
