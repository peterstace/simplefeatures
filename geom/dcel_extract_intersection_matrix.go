package geom

import (
	"fmt"

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
		aI, aB, aE := v.isOnIBE(inputAMask)
		bI, bB, bE := v.isOnIBE(inputBMask)
		for _, pair := range []struct {
			update     bool
			locA, locB de9im.Location
		}{
			{aI && bI, de9im.Interior, de9im.Interior},
			{aI && bB, de9im.Interior, de9im.Boundary},
			{aI && bE, de9im.Interior, de9im.Exterior},
			{aB && bI, de9im.Boundary, de9im.Interior},
			{aB && bB, de9im.Boundary, de9im.Boundary},
			{aB && bE, de9im.Boundary, de9im.Exterior},
			{aE && bI, de9im.Exterior, de9im.Interior},
			{aE && bB, de9im.Exterior, de9im.Boundary},
			{aE && bE, de9im.Exterior, de9im.Exterior},
		} {
			if pair.update {
				oldDim := m.Get(pair.locA, pair.locB)
				newDim := de9im.MaxDimension(de9im.Dim0, oldDim)
				m = m.With(pair.locA, pair.locB, newDim)
			}
		}
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

// isOnIBE calculates if the vertex is on the Interior, Boundary, or Exterior
// of the geometry on the given side. Note that the three locations are not
// mutually exclusive.
func (v *vertexRecord) isOnIBE(sideMask uint8) (bool, bool, bool) {
	if (v.locLabel & sideMask) == 0 {
		// We don't know the location of the point. But it must be either
		// Exterior or Interior because if it were Boundary, then we would know
		// that. We can just use the location of one of the incident edges,
		// since that would have the same location.
		if len(v.incidents) == 0 {
			// Can't happen, due to ghost edges.
			panic("point has no incidents")
		}
		switch loc := v.incidents[0].location(sideMask); loc {
		case de9im.Interior:
			return true, false, false
		case de9im.Boundary:
			return false, true, false
		case de9im.Exterior:
			return false, false, true
		default:
			panic(fmt.Sprintf("invalid location: %v", loc))
		}
	}

	isOnI := (v.locLabel & sideMask & locInterior) != 0
	isOnB := (v.locLabel & sideMask & locBoundary) != 0
	return isOnI, isOnB, false
}
