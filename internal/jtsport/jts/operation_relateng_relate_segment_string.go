package jts

// OperationRelateng_RelateSegmentString models a linear edge of a
// RelateGeometry.
type OperationRelateng_RelateSegmentString struct {
	*Noding_BasicSegmentString
	isA             bool
	dimension       int
	id              int
	ringId          int
	inputGeom       *OperationRelateng_RelateGeometry
	parentPolygonal *Geom_Geometry
}

// OperationRelateng_RelateSegmentString_CreateLine creates a RelateSegmentString
// for a line.
func OperationRelateng_RelateSegmentString_CreateLine(pts []*Geom_Coordinate, isA bool, elementId int, parent *OperationRelateng_RelateGeometry) *OperationRelateng_RelateSegmentString {
	return operationRelateng_createSegmentString(pts, isA, Geom_Dimension_L, elementId, -1, nil, parent)
}

// OperationRelateng_RelateSegmentString_CreateRing creates a RelateSegmentString
// for a polygon ring.
func OperationRelateng_RelateSegmentString_CreateRing(pts []*Geom_Coordinate, isA bool, elementId, ringId int, poly *Geom_Geometry, parent *OperationRelateng_RelateGeometry) *OperationRelateng_RelateSegmentString {
	return operationRelateng_createSegmentString(pts, isA, Geom_Dimension_A, elementId, ringId, poly, parent)
}

func operationRelateng_createSegmentString(pts []*Geom_Coordinate, isA bool, dim, elementId, ringId int, poly *Geom_Geometry, parent *OperationRelateng_RelateGeometry) *OperationRelateng_RelateSegmentString {
	pts = operationRelateng_removeRepeatedPoints(pts)
	return operationRelateng_NewRelateSegmentString(pts, isA, dim, elementId, ringId, poly, parent)
}

func operationRelateng_removeRepeatedPoints(pts []*Geom_Coordinate) []*Geom_Coordinate {
	if Geom_CoordinateArrays_HasRepeatedPoints(pts) {
		pts = Geom_CoordinateArrays_RemoveRepeatedPoints(pts)
	}
	return pts
}

func operationRelateng_NewRelateSegmentString(pts []*Geom_Coordinate, isA bool, dimension, id, ringId int, poly *Geom_Geometry, inputGeom *OperationRelateng_RelateGeometry) *OperationRelateng_RelateSegmentString {
	parent := Noding_NewBasicSegmentString(pts, nil)
	return &OperationRelateng_RelateSegmentString{
		Noding_BasicSegmentString: parent,
		isA:                      isA,
		dimension:                dimension,
		id:                       id,
		ringId:                   ringId,
		parentPolygonal:          poly,
		inputGeom:                inputGeom,
	}
}

// IsA returns true if this segment string is from geometry A.
func (rss *OperationRelateng_RelateSegmentString) IsA() bool {
	return rss.isA
}

// GetGeometry returns the parent RelateGeometry.
func (rss *OperationRelateng_RelateSegmentString) GetGeometry() *OperationRelateng_RelateGeometry {
	return rss.inputGeom
}

// GetPolygonal returns the parent polygonal geometry if this is a ring.
func (rss *OperationRelateng_RelateSegmentString) GetPolygonal() *Geom_Geometry {
	return rss.parentPolygonal
}

// CreateNodeSection creates a NodeSection for an intersection point on this
// segment string.
func (rss *OperationRelateng_RelateSegmentString) CreateNodeSection(segIndex int, intPt *Geom_Coordinate) *OperationRelateng_NodeSection {
	isNodeAtVertex := intPt.Equals2D(rss.GetCoordinate(segIndex)) ||
		intPt.Equals2D(rss.GetCoordinate(segIndex+1))
	prev := rss.prevVertex(segIndex, intPt)
	next := rss.nextVertex(segIndex, intPt)
	return OperationRelateng_NewNodeSection(rss.isA, rss.dimension, rss.id, rss.ringId,
		rss.parentPolygonal, isNodeAtVertex, prev, intPt, next)
}

func (rss *OperationRelateng_RelateSegmentString) prevVertex(segIndex int, pt *Geom_Coordinate) *Geom_Coordinate {
	segStart := rss.GetCoordinate(segIndex)
	if !segStart.Equals2D(pt) {
		return segStart
	}
	// pt is at segment start, so get previous vertex.
	if segIndex > 0 {
		return rss.GetCoordinate(segIndex - 1)
	}
	if rss.IsClosed() {
		return rss.prevInRing(segIndex)
	}
	return nil
}

func (rss *OperationRelateng_RelateSegmentString) nextVertex(segIndex int, pt *Geom_Coordinate) *Geom_Coordinate {
	segEnd := rss.GetCoordinate(segIndex + 1)
	if !segEnd.Equals2D(pt) {
		return segEnd
	}
	// pt is at seg end, so get next vertex.
	if segIndex < rss.Size()-2 {
		return rss.GetCoordinate(segIndex + 2)
	}
	if rss.IsClosed() {
		return rss.nextInRing(segIndex + 1)
	}
	// segstring is not closed, so there is no next segment.
	return nil
}

func (rss *OperationRelateng_RelateSegmentString) prevInRing(segIndex int) *Geom_Coordinate {
	// In a closed ring, the point before index 0 is at index n-2 (since first =
	// last).
	i := segIndex - 1
	if i < 0 {
		i = rss.Size() - 2
	}
	return rss.GetCoordinate(i)
}

func (rss *OperationRelateng_RelateSegmentString) nextInRing(segIndex int) *Geom_Coordinate {
	// In a closed ring, the point after index n-1 is at index 1 (since first =
	// last).
	i := segIndex + 1
	if i > rss.Size()-1 {
		i = 1
	}
	return rss.GetCoordinate(i)
}

// IsContainingSegment tests if a segment intersection point has that segment
// as its canonical containing segment. Segments are half-closed, and contain
// their start point but not the endpoint, except for the final segment in a
// non-closed segment string, which contains its endpoint as well.
func (rss *OperationRelateng_RelateSegmentString) IsContainingSegment(segIndex int, pt *Geom_Coordinate) bool {
	// Intersection is at segment start vertex - process it.
	if pt.Equals2D(rss.GetCoordinate(segIndex)) {
		return true
	}
	if pt.Equals2D(rss.GetCoordinate(segIndex + 1)) {
		isFinalSegment := segIndex == rss.Size()-2
		if rss.IsClosed() || !isFinalSegment {
			return false
		}
		// For final segment, process intersections with final endpoint.
		return true
	}
	// Intersection is interior - process it.
	return true
}
