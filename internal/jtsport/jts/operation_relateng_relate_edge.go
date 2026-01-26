package jts

import "fmt"

// OperationRelateng_RelateEdge represents an edge in a RelateNode graph.
// Constants for edge direction.
const (
	OperationRelateng_RelateEdge_IS_FORWARD = true
	OperationRelateng_RelateEdge_IS_REVERSE = false
)

// The dimension of an input geometry which is not known.
const OperationRelateng_RelateEdge_DIM_UNKNOWN = -1

// Indicates that the location is currently unknown.
const operationRelateng_RelateEdge_LOC_UNKNOWN = Geom_Location_None

// OperationRelateng_RelateEdge_Create creates a new RelateEdge.
func OperationRelateng_RelateEdge_Create(node *OperationRelateng_RelateNode, dirPt *Geom_Coordinate, isA bool, dim int, isForward bool) *OperationRelateng_RelateEdge {
	if dim == Geom_Dimension_A {
		// Create an area edge.
		return OperationRelateng_NewRelateEdgeArea(node, dirPt, isA, isForward)
	}
	// Create line edge.
	return OperationRelateng_NewRelateEdgeLine(node, dirPt, isA)
}

// OperationRelateng_RelateEdge_FindKnownEdgeIndex finds the index of the first
// edge with a known location for the given geometry.
func OperationRelateng_RelateEdge_FindKnownEdgeIndex(edges []*OperationRelateng_RelateEdge, isA bool) int {
	for i, e := range edges {
		if e.isKnown(isA) {
			return i
		}
	}
	return -1
}

// OperationRelateng_RelateEdge_SetAreaInteriorAll sets all edges to area interior.
func OperationRelateng_RelateEdge_SetAreaInteriorAll(edges []*OperationRelateng_RelateEdge, isA bool) {
	for _, e := range edges {
		e.SetAreaInterior(isA)
	}
}

// OperationRelateng_RelateEdge represents an edge in the RelateNode topology
// graph.
type OperationRelateng_RelateEdge struct {
	node  *OperationRelateng_RelateNode
	dirPt *Geom_Coordinate

	aDim      int
	aLocLeft  int
	aLocRight int
	aLocLine  int

	bDim      int
	bLocLeft  int
	bLocRight int
	bLocLine  int
}

// OperationRelateng_NewRelateEdgeArea creates an area edge.
func OperationRelateng_NewRelateEdgeArea(node *OperationRelateng_RelateNode, pt *Geom_Coordinate, isA bool, isForward bool) *OperationRelateng_RelateEdge {
	e := &OperationRelateng_RelateEdge{
		node:      node,
		dirPt:     pt,
		aDim:      OperationRelateng_RelateEdge_DIM_UNKNOWN,
		aLocLeft:  operationRelateng_RelateEdge_LOC_UNKNOWN,
		aLocRight: operationRelateng_RelateEdge_LOC_UNKNOWN,
		aLocLine:  operationRelateng_RelateEdge_LOC_UNKNOWN,
		bDim:      OperationRelateng_RelateEdge_DIM_UNKNOWN,
		bLocLeft:  operationRelateng_RelateEdge_LOC_UNKNOWN,
		bLocRight: operationRelateng_RelateEdge_LOC_UNKNOWN,
		bLocLine:  operationRelateng_RelateEdge_LOC_UNKNOWN,
	}
	e.setLocationsArea(isA, isForward)
	return e
}

// OperationRelateng_NewRelateEdgeLine creates a line edge.
func OperationRelateng_NewRelateEdgeLine(node *OperationRelateng_RelateNode, pt *Geom_Coordinate, isA bool) *OperationRelateng_RelateEdge {
	e := &OperationRelateng_RelateEdge{
		node:      node,
		dirPt:     pt,
		aDim:      OperationRelateng_RelateEdge_DIM_UNKNOWN,
		aLocLeft:  operationRelateng_RelateEdge_LOC_UNKNOWN,
		aLocRight: operationRelateng_RelateEdge_LOC_UNKNOWN,
		aLocLine:  operationRelateng_RelateEdge_LOC_UNKNOWN,
		bDim:      OperationRelateng_RelateEdge_DIM_UNKNOWN,
		bLocLeft:  operationRelateng_RelateEdge_LOC_UNKNOWN,
		bLocRight: operationRelateng_RelateEdge_LOC_UNKNOWN,
		bLocLine:  operationRelateng_RelateEdge_LOC_UNKNOWN,
	}
	e.setLocationsLine(isA)
	return e
}

// OperationRelateng_NewRelateEdgeFull creates an edge with explicit locations.
func OperationRelateng_NewRelateEdgeFull(node *OperationRelateng_RelateNode, pt *Geom_Coordinate, isA bool, locLeft, locRight, locLine int) *OperationRelateng_RelateEdge {
	e := &OperationRelateng_RelateEdge{
		node:      node,
		dirPt:     pt,
		aDim:      OperationRelateng_RelateEdge_DIM_UNKNOWN,
		aLocLeft:  operationRelateng_RelateEdge_LOC_UNKNOWN,
		aLocRight: operationRelateng_RelateEdge_LOC_UNKNOWN,
		aLocLine:  operationRelateng_RelateEdge_LOC_UNKNOWN,
		bDim:      OperationRelateng_RelateEdge_DIM_UNKNOWN,
		bLocLeft:  operationRelateng_RelateEdge_LOC_UNKNOWN,
		bLocRight: operationRelateng_RelateEdge_LOC_UNKNOWN,
		bLocLine:  operationRelateng_RelateEdge_LOC_UNKNOWN,
	}
	e.setLocations(isA, locLeft, locRight, locLine)
	return e
}

func (e *OperationRelateng_RelateEdge) setLocations(isA bool, locLeft, locRight, locLine int) {
	if isA {
		e.aDim = 2
		e.aLocLeft = locLeft
		e.aLocRight = locRight
		e.aLocLine = locLine
	} else {
		e.bDim = 2
		e.bLocLeft = locLeft
		e.bLocRight = locRight
		e.bLocLine = locLine
	}
}

func (e *OperationRelateng_RelateEdge) setLocationsLine(isA bool) {
	if isA {
		e.aDim = 1
		e.aLocLeft = Geom_Location_Exterior
		e.aLocRight = Geom_Location_Exterior
		e.aLocLine = Geom_Location_Interior
	} else {
		e.bDim = 1
		e.bLocLeft = Geom_Location_Exterior
		e.bLocRight = Geom_Location_Exterior
		e.bLocLine = Geom_Location_Interior
	}
}

func (e *OperationRelateng_RelateEdge) setLocationsArea(isA bool, isForward bool) {
	locLeft := Geom_Location_Interior
	locRight := Geom_Location_Exterior
	if isForward {
		locLeft = Geom_Location_Exterior
		locRight = Geom_Location_Interior
	}
	if isA {
		e.aDim = 2
		e.aLocLeft = locLeft
		e.aLocRight = locRight
		e.aLocLine = Geom_Location_Boundary
	} else {
		e.bDim = 2
		e.bLocLeft = locLeft
		e.bLocRight = locRight
		e.bLocLine = Geom_Location_Boundary
	}
}

// CompareToEdge compares this edge's direction point angle to another edge
// direction point.
func (e *OperationRelateng_RelateEdge) CompareToEdge(edgeDirPt *Geom_Coordinate) int {
	return Algorithm_PolygonNodeTopology_CompareAngle(e.node.GetCoordinate(), e.dirPt, edgeDirPt)
}

// Merge merges another edge's locations into this edge.
func (e *OperationRelateng_RelateEdge) Merge(isA bool, dirPt *Geom_Coordinate, dim int, isForward bool) {
	locEdge := Geom_Location_Interior
	locLeft := Geom_Location_Exterior
	locRight := Geom_Location_Exterior
	if dim == Geom_Dimension_A {
		locEdge = Geom_Location_Boundary
		if isForward {
			locLeft = Geom_Location_Exterior
			locRight = Geom_Location_Interior
		} else {
			locLeft = Geom_Location_Interior
			locRight = Geom_Location_Exterior
		}
	}

	if !e.isKnown(isA) {
		e.setDimension(isA, dim)
		e.setOn(isA, locEdge)
		e.setLeft(isA, locLeft)
		e.setRight(isA, locRight)
		return
	}

	// Assert: node-dirpt is collinear with node-pt.
	e.mergeDimEdgeLoc(isA, locEdge)
	e.mergeSideLocation(isA, Geom_Position_Left, locLeft)
	e.mergeSideLocation(isA, Geom_Position_Right, locRight)
}

// Area edges override Line edges. Merging edges of same dimension is a no-op
// for the dimension and on location. But merging an area edge into a line edge
// sets the dimension to A and the location to BOUNDARY.
func (e *OperationRelateng_RelateEdge) mergeDimEdgeLoc(isA bool, locEdge int) {
	// TODO: this logic needs work - ie handling A edges marked as Interior.
	dim := Geom_Dimension_L
	if locEdge == Geom_Location_Boundary {
		dim = Geom_Dimension_A
	}
	if dim == Geom_Dimension_A && e.dimension(isA) == Geom_Dimension_L {
		e.setDimension(isA, dim)
		e.setOn(isA, Geom_Location_Boundary)
	}
}

func (e *OperationRelateng_RelateEdge) mergeSideLocation(isA bool, pos, loc int) {
	currLoc := e.Location(isA, pos)
	// INTERIOR takes precedence over EXTERIOR.
	if currLoc != Geom_Location_Interior {
		e.SetLocation(isA, pos, loc)
	}
}

func (e *OperationRelateng_RelateEdge) setDimension(isA bool, dimension int) {
	if isA {
		e.aDim = dimension
	} else {
		e.bDim = dimension
	}
}

// SetLocation sets the location for a given position.
func (e *OperationRelateng_RelateEdge) SetLocation(isA bool, pos, loc int) {
	switch pos {
	case Geom_Position_Left:
		e.setLeft(isA, loc)
	case Geom_Position_Right:
		e.setRight(isA, loc)
	case Geom_Position_On:
		e.setOn(isA, loc)
	}
}

// SetAllLocations sets all locations to the given value.
func (e *OperationRelateng_RelateEdge) SetAllLocations(isA bool, loc int) {
	e.setLeft(isA, loc)
	e.setRight(isA, loc)
	e.setOn(isA, loc)
}

// SetUnknownLocations sets unknown locations to the given value.
func (e *OperationRelateng_RelateEdge) SetUnknownLocations(isA bool, loc int) {
	if !e.isKnownPos(isA, Geom_Position_Left) {
		e.SetLocation(isA, Geom_Position_Left, loc)
	}
	if !e.isKnownPos(isA, Geom_Position_Right) {
		e.SetLocation(isA, Geom_Position_Right, loc)
	}
	if !e.isKnownPos(isA, Geom_Position_On) {
		e.SetLocation(isA, Geom_Position_On, loc)
	}
}

func (e *OperationRelateng_RelateEdge) setLeft(isA bool, loc int) {
	if isA {
		e.aLocLeft = loc
	} else {
		e.bLocLeft = loc
	}
}

func (e *OperationRelateng_RelateEdge) setRight(isA bool, loc int) {
	if isA {
		e.aLocRight = loc
	} else {
		e.bLocRight = loc
	}
}

func (e *OperationRelateng_RelateEdge) setOn(isA bool, loc int) {
	if isA {
		e.aLocLine = loc
	} else {
		e.bLocLine = loc
	}
}

// Location returns the location for a given position.
func (e *OperationRelateng_RelateEdge) Location(isA bool, position int) int {
	if isA {
		switch position {
		case Geom_Position_Left:
			return e.aLocLeft
		case Geom_Position_Right:
			return e.aLocRight
		case Geom_Position_On:
			return e.aLocLine
		}
	} else {
		switch position {
		case Geom_Position_Left:
			return e.bLocLeft
		case Geom_Position_Right:
			return e.bLocRight
		case Geom_Position_On:
			return e.bLocLine
		}
	}
	panic("should never reach here")
}

func (e *OperationRelateng_RelateEdge) dimension(isA bool) int {
	if isA {
		return e.aDim
	}
	return e.bDim
}

func (e *OperationRelateng_RelateEdge) isKnown(isA bool) bool {
	if isA {
		return e.aDim != OperationRelateng_RelateEdge_DIM_UNKNOWN
	}
	return e.bDim != OperationRelateng_RelateEdge_DIM_UNKNOWN
}

func (e *OperationRelateng_RelateEdge) isKnownPos(isA bool, pos int) bool {
	return e.Location(isA, pos) != operationRelateng_RelateEdge_LOC_UNKNOWN
}

// IsInterior tests if the given position is in the interior.
func (e *OperationRelateng_RelateEdge) IsInterior(isA bool, position int) bool {
	return e.Location(isA, position) == Geom_Location_Interior
}

// SetDimLocations sets the dimension and all locations for a geometry.
func (e *OperationRelateng_RelateEdge) SetDimLocations(isA bool, dim, loc int) {
	if isA {
		e.aDim = dim
		e.aLocLeft = loc
		e.aLocRight = loc
		e.aLocLine = loc
	} else {
		e.bDim = dim
		e.bLocLeft = loc
		e.bLocRight = loc
		e.bLocLine = loc
	}
}

// SetAreaInterior sets all locations to interior for an area.
func (e *OperationRelateng_RelateEdge) SetAreaInterior(isA bool) {
	if isA {
		e.aLocLeft = Geom_Location_Interior
		e.aLocRight = Geom_Location_Interior
		e.aLocLine = Geom_Location_Interior
	} else {
		e.bLocLeft = Geom_Location_Interior
		e.bLocRight = Geom_Location_Interior
		e.bLocLine = Geom_Location_Interior
	}
}

// String returns a string representation of this edge.
func (e *OperationRelateng_RelateEdge) String() string {
	return Io_WKTWriter_ToLineStringFromTwoCoords(e.node.GetCoordinate(), e.dirPt) +
		" - " + e.labelString()
}

func (e *OperationRelateng_RelateEdge) labelString() string {
	return fmt.Sprintf("A:%s/B:%s",
		e.locationString(OperationRelateng_RelateGeometry_GEOM_A),
		e.locationString(OperationRelateng_RelateGeometry_GEOM_B))
}

func (e *OperationRelateng_RelateEdge) locationString(isA bool) string {
	return fmt.Sprintf("%c%c%c",
		Geom_Location_ToLocationSymbol(e.Location(isA, Geom_Position_Left)),
		Geom_Location_ToLocationSymbol(e.Location(isA, Geom_Position_On)),
		Geom_Location_ToLocationSymbol(e.Location(isA, Geom_Position_Right)))
}
