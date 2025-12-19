package jts

// coord2DKey is a 2D coordinate key for map lookups.
// This is needed because Geom_Coordinate includes Z which may be NaN,
// and NaN != NaN in Go, breaking map lookups.
// Java's Coordinate.equals() and hashCode() only use X and Y.
type coord2DKey struct {
	x, y float64
}

// OperationRelateng_LinearBoundary determines the boundary points of a linear
// geometry, using a BoundaryNodeRule.
type OperationRelateng_LinearBoundary struct {
	vertexDegree     map[coord2DKey]int
	hasBoundary      bool
	boundaryNodeRule Algorithm_BoundaryNodeRule
}

// OperationRelateng_NewLinearBoundary creates a new LinearBoundary for the
// given lines using the specified boundary node rule.
func OperationRelateng_NewLinearBoundary(lines []*Geom_LineString, bnRule Algorithm_BoundaryNodeRule) *OperationRelateng_LinearBoundary {
	lb := &OperationRelateng_LinearBoundary{
		boundaryNodeRule: bnRule,
	}
	lb.vertexDegree = operationRelateng_LinearBoundary_computeBoundaryPoints(lines)
	lb.hasBoundary = lb.checkBoundary(lb.vertexDegree)
	return lb
}

func (lb *OperationRelateng_LinearBoundary) checkBoundary(vertexDegree map[coord2DKey]int) bool {
	for _, degree := range vertexDegree {
		if lb.boundaryNodeRule.IsInBoundary(degree) {
			return true
		}
	}
	return false
}

// HasBoundary reports whether this linear geometry has any boundary points.
func (lb *OperationRelateng_LinearBoundary) HasBoundary() bool {
	return lb.hasBoundary
}

// IsBoundary tests whether a point is a boundary point of the linear geometry.
func (lb *OperationRelateng_LinearBoundary) IsBoundary(pt *Geom_Coordinate) bool {
	key := coord2DKey{x: pt.X, y: pt.Y}
	degree, exists := lb.vertexDegree[key]
	if !exists {
		return false
	}
	return lb.boundaryNodeRule.IsInBoundary(degree)
}

func operationRelateng_LinearBoundary_computeBoundaryPoints(lines []*Geom_LineString) map[coord2DKey]int {
	vertexDegree := make(map[coord2DKey]int)
	for _, line := range lines {
		if line.IsEmpty() {
			continue
		}
		operationRelateng_LinearBoundary_addEndpoint(line.GetCoordinateN(0), vertexDegree)
		operationRelateng_LinearBoundary_addEndpoint(line.GetCoordinateN(line.GetNumPoints()-1), vertexDegree)
	}
	return vertexDegree
}

func operationRelateng_LinearBoundary_addEndpoint(p *Geom_Coordinate, degree map[coord2DKey]int) {
	key := coord2DKey{x: p.X, y: p.Y}
	dim := degree[key]
	dim++
	degree[key] = dim
}
