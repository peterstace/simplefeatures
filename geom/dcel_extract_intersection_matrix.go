package geom

func (d *doublyConnectedEdgeList) extractIntersectionMatrix() matrix {
	im := newMatrix()
	for _, v := range d.vertices {
		locA := v.location(operandA)
		locB := v.location(operandB)
		im.set(locA, locB, '0')
	}
	for _, e := range d.halfEdges {
		locA := e.location(operandA)
		locB := e.location(operandB)
		im.set(locA, locB, '1')
	}
	for _, f := range d.faces {
		locA := f.location(operandA)
		locB := f.location(operandB)
		im.set(locA, locB, '2')
	}
	return im
}

func (f *faceRecord) location(operand operand) imLocation {
	if !f.inSet[operand] {
		return imExterior
	}
	return imInterior
}

func (e *halfEdgeRecord) location(operand operand) imLocation {
	face1Present := e.incident.inSet[operand]
	face2Present := e.twin.incident.inSet[operand]
	if face1Present && face2Present {
		return imInterior
	}
	if face1Present != face2Present {
		return imBoundary
	}
	if e.inSet[operand] {
		return imInterior
	}
	return imExterior
}

func (v *vertexRecord) location(operand operand) imLocation {
	// NOTE: It's important that we check the Boundary flag before the Interior
	// flag, since both might be set. In that case, we want to treat the
	// location as a Boundary, since the boundary is a more specific case.
	if v.locations[operand].boundary {
		return imBoundary
	}
	if v.locations[operand].interior {
		return imInterior
	}

	// We don't know the location of the point. But it must be either Exterior
	// or Interior because if it were Boundary, then we would know that due to
	// an explicit flag. We can just use the location of one of the incident
	// edges, since that would have the same location.
	for e := range v.incidents {
		return e.location(operand)
	}
	panic("point has no incidents") // Can't happen, due to ghost edges.
}
