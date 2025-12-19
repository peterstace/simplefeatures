package jts

// OperationOverlayng_EdgeSourceInfo records topological information about an
// edge representing a piece of linework (lineString or polygon ring) from a
// single source geometry. This information is carried through the noding
// process (which may result in many noded edges sharing the same information
// object). It is then used to populate the topology info fields in Edges
// (possibly via merging). That information is used to construct the topology
// graph OverlayLabels.
type OperationOverlayng_EdgeSourceInfo struct {
	index      int
	dim        int
	isHole     bool
	depthDelta int
}

// OperationOverlayng_NewEdgeSourceInfoForArea creates an EdgeSourceInfo for an
// area edge.
func OperationOverlayng_NewEdgeSourceInfoForArea(index, depthDelta int, isHole bool) *OperationOverlayng_EdgeSourceInfo {
	return &OperationOverlayng_EdgeSourceInfo{
		index:      index,
		dim:        Geom_Dimension_A,
		depthDelta: depthDelta,
		isHole:     isHole,
	}
}

// OperationOverlayng_NewEdgeSourceInfoForLine creates an EdgeSourceInfo for a
// line edge.
func OperationOverlayng_NewEdgeSourceInfoForLine(index int) *OperationOverlayng_EdgeSourceInfo {
	return &OperationOverlayng_EdgeSourceInfo{
		index: index,
		dim:   Geom_Dimension_L,
	}
}

// GetIndex returns the index of the parent geometry.
func (esi *OperationOverlayng_EdgeSourceInfo) GetIndex() int {
	return esi.index
}

// GetDimension returns the dimension of the edge.
func (esi *OperationOverlayng_EdgeSourceInfo) GetDimension() int {
	return esi.dim
}

// GetDepthDelta returns the depth delta of the edge.
func (esi *OperationOverlayng_EdgeSourceInfo) GetDepthDelta() int {
	return esi.depthDelta
}

// IsHole returns whether the edge is part of a hole.
func (esi *OperationOverlayng_EdgeSourceInfo) IsHole() bool {
	return esi.isHole
}

// String returns a string representation of the EdgeSourceInfo.
func (esi *OperationOverlayng_EdgeSourceInfo) String() string {
	return OperationOverlayng_Edge_InfoString(esi.index, esi.dim, esi.isHole, esi.depthDelta)
}
