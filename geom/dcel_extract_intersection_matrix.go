package geom

import (
	"github.com/peterstace/simplefeatures/de9im"
)

func (d *doublyConnectedEdgeList) extractIntersectionMatrix() de9im.Matrix {
	var m de9im.Matrix

	for _, f := range d.faces {
		assertPresenceBits(f.label)
		locA, locB := de9im.Exterior, de9im.Exterior
		if f.label&inputAInSet != 0 {
			locA = de9im.Interior
		}
		if f.label&inputBInSet != 0 {
			locB = de9im.Interior
		}
		if m.Get(locA, locB) == de9im.Empty {
			m = m.With(locA, locB, de9im.Dim2)
		}
	}

	for _, e := range d.halfEdges {
		locA := e.location(inputAMask)
		locB := e.location(inputBMask)
		newDim := de9im.MaxDimension(de9im.Dim1, m.Get(locA, locB))
		m = m.With(locA, locB, newDim)
	}

	for _, v := range d.vertices {
		locA := v.location(inputAMask)
		locB := v.location(inputBMask)
		newDim := de9im.MaxDimension(de9im.Dim0, m.Get(locA, locB))
		m = m.With(locA, locB, newDim)
	}
	return m
}

func (e *halfEdgeRecord) location(sideMask uint8) de9im.Location {
	if (e.edgeLabel & inSetMask & sideMask) == 0 {
		return de9im.Exterior
	}
	face1Present := (e.incident.label & inSetMask & sideMask) != 0
	face2Present := (e.twin.incident.label & inSetMask & sideMask) != 0
	if face1Present != face2Present {
		return de9im.Boundary
	}
	return de9im.Interior
}

func (v *vertexRecord) location(sideMask uint8) de9im.Location {
	switch {
	case (v.locLabel & sideMask & locInterior) != 0:
		return de9im.Interior
	case (v.locLabel & sideMask & locBoundary) != 0:
		return de9im.Boundary
	default:
		// We don't know the location of the point. But it must be either
		// Exterior or Interior because if it were Boundary, then we would know
		// that. We can just use the location of one of the incident edges,
		// since that would have the same location.
		for _, e := range v.incidents {
			return e.location(sideMask)
		}
		panic("point has no incidents") // Can't happen, due to ghost edges.
	}
}
