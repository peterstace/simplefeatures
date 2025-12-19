package jts

import "strconv"

// OperationOverlayng_Edge represents the linework for edges in the topology
// derived from (up to) two parent geometries. An edge may be the result of the
// merging of two or more edges which have the same linework (although possibly
// different orientations). In this case the topology information is derived
// from the merging of the information in the source edges. Merged edges can
// occur in the following situations:
//   - Due to coincident edges of polygonal or linear geometries.
//   - Due to topology collapse caused by snapping or rounding of polygonal
//     geometries.
//
// The source edges may have the same parent geometry, or different ones, or a
// mix of the two.
type OperationOverlayng_Edge struct {
	pts         []*Geom_Coordinate
	aDim        int
	aDepthDelta int
	aIsHole     bool
	bDim        int
	bDepthDelta int
	bIsHole     bool
}

// OperationOverlayng_Edge_IsCollapsed tests if the given point sequence is a
// collapsed line. A collapsed edge has fewer than two distinct points.
func OperationOverlayng_Edge_IsCollapsed(pts []*Geom_Coordinate) bool {
	if len(pts) < 2 {
		return true
	}
	// zero-length line
	if pts[0].Equals2D(pts[1]) {
		return true
	}
	// TODO: is pts > 2 with equal points ever expected?
	if len(pts) > 2 {
		if pts[len(pts)-1].Equals2D(pts[len(pts)-2]) {
			return true
		}
	}
	return false
}

// OperationOverlayng_NewEdge creates a new Edge from a coordinate sequence and
// edge source info.
func OperationOverlayng_NewEdge(pts []*Geom_Coordinate, info *OperationOverlayng_EdgeSourceInfo) *OperationOverlayng_Edge {
	e := &OperationOverlayng_Edge{
		pts:  pts,
		aDim: OperationOverlayng_OverlayLabel_DIM_UNKNOWN,
		bDim: OperationOverlayng_OverlayLabel_DIM_UNKNOWN,
	}
	e.copyInfo(info)
	return e
}

// GetCoordinates returns the coordinates of the edge.
func (e *OperationOverlayng_Edge) GetCoordinates() []*Geom_Coordinate {
	return e.pts
}

// GetCoordinate returns the coordinate at the given index.
func (e *OperationOverlayng_Edge) GetCoordinate(index int) *Geom_Coordinate {
	return e.pts[index]
}

// Size returns the number of coordinates in the edge.
func (e *OperationOverlayng_Edge) Size() int {
	return len(e.pts)
}

// Direction computes the direction of the edge based on its endpoint
// coordinates.
func (e *OperationOverlayng_Edge) Direction() bool {
	pts := e.GetCoordinates()
	if len(pts) < 2 {
		panic("Edge must have >= 2 points")
	}
	p0 := pts[0]
	p1 := pts[1]

	pn0 := pts[len(pts)-1]
	pn1 := pts[len(pts)-2]

	cmp := 0
	cmp0 := p0.CompareTo(pn0)
	if cmp0 != 0 {
		cmp = cmp0
	}

	if cmp == 0 {
		cmp1 := p1.CompareTo(pn1)
		if cmp1 != 0 {
			cmp = cmp1
		}
	}

	if cmp == 0 {
		panic("Edge direction cannot be determined because endpoints are equal")
	}

	return cmp == -1
}

// RelativeDirection compares two coincident edges to determine whether they
// have the same or opposite direction.
func (e *OperationOverlayng_Edge) RelativeDirection(edge2 *OperationOverlayng_Edge) bool {
	// assert: the edges match (have the same coordinates up to direction)
	if !e.GetCoordinate(0).Equals2D(edge2.GetCoordinate(0)) {
		return false
	}
	if !e.GetCoordinate(1).Equals2D(edge2.GetCoordinate(1)) {
		return false
	}
	return true
}

// CreateLabel creates an OverlayLabel for this edge.
func (e *OperationOverlayng_Edge) CreateLabel() *OperationOverlayng_OverlayLabel {
	lbl := OperationOverlayng_NewOverlayLabel()
	e.initLabel(lbl, 0, e.aDim, e.aDepthDelta, e.aIsHole)
	e.initLabel(lbl, 1, e.bDim, e.bDepthDelta, e.bIsHole)
	return lbl
}

// initLabel populates the label for an edge resulting from an input geometry.
func (e *OperationOverlayng_Edge) initLabel(lbl *OperationOverlayng_OverlayLabel, geomIndex, dim, depthDelta int, isHole bool) {
	dimLabel := operationOverlayng_Edge_labelDim(dim, depthDelta)

	switch dimLabel {
	case OperationOverlayng_OverlayLabel_DIM_NOT_PART:
		lbl.InitNotPart(geomIndex)
	case OperationOverlayng_OverlayLabel_DIM_BOUNDARY:
		lbl.InitBoundary(geomIndex, operationOverlayng_Edge_locationLeft(depthDelta), operationOverlayng_Edge_locationRight(depthDelta), isHole)
	case OperationOverlayng_OverlayLabel_DIM_COLLAPSE:
		lbl.InitCollapse(geomIndex, isHole)
	case OperationOverlayng_OverlayLabel_DIM_LINE:
		lbl.InitLine(geomIndex)
	}
}

func operationOverlayng_Edge_labelDim(dim, depthDelta int) int {
	if dim == Geom_Dimension_False {
		return OperationOverlayng_OverlayLabel_DIM_NOT_PART
	}
	if dim == Geom_Dimension_L {
		return OperationOverlayng_OverlayLabel_DIM_LINE
	}
	// assert: dim is A
	isCollapse := depthDelta == 0
	if isCollapse {
		return OperationOverlayng_OverlayLabel_DIM_COLLAPSE
	}
	return OperationOverlayng_OverlayLabel_DIM_BOUNDARY
}

// isShell tests whether the edge is part of a shell in the given geometry.
// This is only the case if the edge is a boundary.
func (e *OperationOverlayng_Edge) isShell(geomIndex int) bool {
	if geomIndex == 0 {
		return e.aDim == OperationOverlayng_OverlayLabel_DIM_BOUNDARY && !e.aIsHole
	}
	return e.bDim == OperationOverlayng_OverlayLabel_DIM_BOUNDARY && !e.bIsHole
}

func operationOverlayng_Edge_locationRight(depthDelta int) int {
	delSign := operationOverlayng_Edge_delSign(depthDelta)
	switch delSign {
	case 0:
		return OperationOverlayng_OverlayLabel_LOC_UNKNOWN
	case 1:
		return Geom_Location_Interior
	case -1:
		return Geom_Location_Exterior
	}
	return OperationOverlayng_OverlayLabel_LOC_UNKNOWN
}

func operationOverlayng_Edge_locationLeft(depthDelta int) int {
	delSign := operationOverlayng_Edge_delSign(depthDelta)
	switch delSign {
	case 0:
		return OperationOverlayng_OverlayLabel_LOC_UNKNOWN
	case 1:
		return Geom_Location_Exterior
	case -1:
		return Geom_Location_Interior
	}
	return OperationOverlayng_OverlayLabel_LOC_UNKNOWN
}

func operationOverlayng_Edge_delSign(depthDel int) int {
	if depthDel > 0 {
		return 1
	}
	if depthDel < 0 {
		return -1
	}
	return 0
}

func (e *OperationOverlayng_Edge) copyInfo(info *OperationOverlayng_EdgeSourceInfo) {
	if info.GetIndex() == 0 {
		e.aDim = info.GetDimension()
		e.aIsHole = info.IsHole()
		e.aDepthDelta = info.GetDepthDelta()
	} else {
		e.bDim = info.GetDimension()
		e.bIsHole = info.IsHole()
		e.bDepthDelta = info.GetDepthDelta()
	}
}

// Merge merges an edge into this edge, updating the topology info accordingly.
func (e *OperationOverlayng_Edge) Merge(edge *OperationOverlayng_Edge) {
	// Marks this as a shell edge if any contributing edge is a shell.
	// Update hole status first, since it depends on edge dim
	e.aIsHole = operationOverlayng_Edge_isHoleMerged(0, e, edge)
	e.bIsHole = operationOverlayng_Edge_isHoleMerged(1, e, edge)

	if edge.aDim > e.aDim {
		e.aDim = edge.aDim
	}
	if edge.bDim > e.bDim {
		e.bDim = edge.bDim
	}

	relDir := e.RelativeDirection(edge)
	flipFactor := 1
	if !relDir {
		flipFactor = -1
	}
	e.aDepthDelta += flipFactor * edge.aDepthDelta
	e.bDepthDelta += flipFactor * edge.bDepthDelta
}

func operationOverlayng_Edge_isHoleMerged(geomIndex int, edge1, edge2 *OperationOverlayng_Edge) bool {
	isShell1 := edge1.isShell(geomIndex)
	isShell2 := edge2.isShell(geomIndex)
	isShellMerged := isShell1 || isShell2
	// flip since isHole is stored
	return !isShellMerged
}

// String returns a string representation of the edge.
func (e *OperationOverlayng_Edge) String() string {
	ptsStr := operationOverlayng_Edge_toStringPts(e.pts)
	aInfo := OperationOverlayng_Edge_InfoString(0, e.aDim, e.aIsHole, e.aDepthDelta)
	bInfo := OperationOverlayng_Edge_InfoString(1, e.bDim, e.bIsHole, e.bDepthDelta)
	return "Edge( " + ptsStr + " ) " + aInfo + "/" + bInfo
}

// ToLineString returns a WKT representation of the edge as a LINESTRING.
func (e *OperationOverlayng_Edge) ToLineString() string {
	return IO_WKTWriter_ToLineStringFromCoords(e.pts)
}

func operationOverlayng_Edge_toStringPts(pts []*Geom_Coordinate) string {
	orig := pts[0]
	dest := pts[len(pts)-1]
	dirPtStr := ""
	if len(pts) > 2 {
		dirPtStr = ", " + IO_WKTWriter_Format(pts[1])
	}
	return IO_WKTWriter_Format(orig) + dirPtStr + " .. " + IO_WKTWriter_Format(dest)
}

// OperationOverlayng_Edge_InfoString returns a string representation of edge info.
func OperationOverlayng_Edge_InfoString(index, dim int, isHole bool, depthDelta int) string {
	prefix := "A:"
	if index != 0 {
		prefix = "B:"
	}
	return prefix +
		string(OperationOverlayng_OverlayLabel_DimensionSymbol(dim)) +
		operationOverlayng_Edge_ringRoleSymbol(dim, isHole) +
		strconv.Itoa(depthDelta)
}

func operationOverlayng_Edge_ringRoleSymbol(dim int, isHole bool) string {
	if operationOverlayng_Edge_hasAreaParent(dim) {
		return string(OperationOverlayng_OverlayLabel_RingRoleSymbol(isHole))
	}
	return ""
}

func operationOverlayng_Edge_hasAreaParent(dim int) bool {
	return dim == OperationOverlayng_OverlayLabel_DIM_BOUNDARY || dim == OperationOverlayng_OverlayLabel_DIM_COLLAPSE
}
