package geom

func (d *doublyConnectedEdgeList) extractIntersectionMatrix() IntersectionMatrix {
	var m IntersectionMatrix
	for _, f := range d.faces {
		locA := f.location(inputAMask)
		locB := f.location(inputBMask)
		m = m.upgradeEntry(locA, locB, imEntry2)
	}
	for _, e := range d.halfEdges {
		locA := e.location(inputAMask)
		locB := e.location(inputBMask)
		m = m.upgradeEntry(locA, locB, imEntry1)
	}
	for _, v := range d.vertices {
		locA := v.location(inputAMask)
		locB := v.location(inputBMask)
		m = m.upgradeEntry(locA, locB, imEntry0)
	}
	return m
}

func (f *faceRecord) location(sideMask uint8) imLocation {
	if (f.label & inSetMask & sideMask) == 0 {
		return imExterior
	}
	return imInterior
}

func (e *halfEdgeRecord) location(sideMask uint8) imLocation {
	if (e.edgeLabel & inSetMask & sideMask) == 0 {
		return imExterior
	}
	face1Present := (e.incident.label & inSetMask & sideMask) != 0
	face2Present := (e.twin.incident.label & inSetMask & sideMask) != 0
	if face1Present != face2Present {
		return imBoundary
	}
	return imInterior
}

func (v *vertexRecord) location(sideMask uint8) imLocation {
	// NOTE: It's important that we check the Boundary flag before the Interior
	// flag, since both might be set. In that case, we want to treat the
	// location as a Boundary, since the boundary is a more specific case.
	switch {
	case (v.locLabel & sideMask & locBoundary) != 0:
		return imBoundary
	case (v.locLabel & sideMask & locInterior) != 0:
		return imInterior
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
