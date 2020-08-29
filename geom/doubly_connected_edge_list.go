package geom

type doublyConnectedEdgeList struct {
	faces     []*faceRecord
	halfEdges []*halfEdgeRecord
	vertices  map[XY]*vertexRecord
}

type faceRecord struct {
	outerComponent  *halfEdgeRecord
	innerComponents []*halfEdgeRecord
}

type halfEdgeRecord struct {
	origin     *vertexRecord
	twin       *halfEdgeRecord
	incident   *faceRecord
	next, prev *halfEdgeRecord
}

type vertexRecord struct {
	coords   XY
	incident *halfEdgeRecord
}

func newDCELFromPolygon(poly Polygon) *doublyConnectedEdgeList {
	poly = poly.ForceCCW()

	dcel := &doublyConnectedEdgeList{vertices: make(map[XY]*vertexRecord)}

	// Extract rings.
	rings := make([]Sequence, 1+poly.NumInteriorRings())
	rings[0] = poly.ExteriorRing().Coordinates()
	for i := 0; i < poly.NumInteriorRings(); i++ {
		rings[i+1] = poly.InteriorRingN(i).Coordinates()
	}

	// Populate vertices.
	for _, ring := range rings {
		for i := 0; i < ring.Length(); i++ {
			xy := ring.GetXY(i)
			if _, ok := dcel.vertices[xy]; !ok {
				dcel.vertices[xy] = &vertexRecord{xy, nil /* populated later */}
			}
		}
	}

	infFace := &faceRecord{
		outerComponent:  nil, // left nil
		innerComponents: nil, // populated later
	}
	polyFace := &faceRecord{
		outerComponent:  nil, // populated later
		innerComponents: nil, // populated later
	}
	dcel.faces = append(dcel.faces, infFace, polyFace)

	for ringIdx, ring := range rings {
		interiorFace := polyFace
		exteriorFace := infFace
		if ringIdx > 0 {
			// For inner rings, the exterior face is a hole rather than the
			// infinite face.
			exteriorFace = &faceRecord{
				outerComponent:  nil, // left nil
				innerComponents: nil, // populated later
			}
			dcel.faces = append(dcel.faces, exteriorFace)
		}

		var newEdges []*halfEdgeRecord
		first := true
		for i := 0; i < ring.Length(); i++ {
			ln, ok := getLine(ring, i)
			if !ok {
				continue
			}
			internalEdge := &halfEdgeRecord{
				origin:   dcel.vertices[ln.a],
				twin:     nil, // populated later
				incident: interiorFace,
				next:     nil, // populated later
				prev:     nil, // populated later
			}
			externalEdge := &halfEdgeRecord{
				origin:   dcel.vertices[ln.b],
				twin:     internalEdge,
				incident: exteriorFace,
				next:     nil, // populated later
				prev:     nil, // populated later
			}
			internalEdge.twin = externalEdge
			dcel.vertices[ln.a].incident = internalEdge
			newEdges = append(newEdges, internalEdge, externalEdge)

			// Set interior/exterior face linkage.
			if first {
				// TODO: The logic here feels awkward. The might be a more general way to do this.
				first = false
				if ringIdx == 0 {
					exteriorFace.innerComponents = append(exteriorFace.innerComponents, externalEdge)
					if interiorFace.outerComponent == nil {
						interiorFace.outerComponent = internalEdge
					}
				} else {
					interiorFace.innerComponents = append(interiorFace.innerComponents, internalEdge)
					if exteriorFace.outerComponent == nil {
						exteriorFace.outerComponent = externalEdge
					}
				}
			}
		}

		numEdges := len(newEdges)
		for i := 0; i < numEdges/2; i++ {
			newEdges[i*2].next = newEdges[(2*i+2)%numEdges]
			newEdges[i*2+1].next = newEdges[(i*2-1+numEdges)%numEdges]
			newEdges[i*2].prev = newEdges[(2*i-2+numEdges)%numEdges]
			newEdges[i*2+1].prev = newEdges[(2*i+3)%numEdges]
		}
		dcel.halfEdges = append(dcel.halfEdges, newEdges...)
	}

	return dcel
}

func (d *doublyConnectedEdgeList) reNodeGraph(other Polygon) {
	indexed := newIndexedLines(other.Boundary().asLines())
	for _, face := range d.faces {
		d.reNodeFace(face, indexed)
	}
}

func (d *doublyConnectedEdgeList) reNodeFace(face *faceRecord, indexed indexedLines) {
	if face.outerComponent != nil { // nil for infinite face
		d.reNodeComponent(face.outerComponent, indexed)
	}
	for _, inner := range face.innerComponents {
		d.reNodeComponent(inner, indexed)
	}
}

func (d *doublyConnectedEdgeList) reNodeComponent(start *halfEdgeRecord, indexed indexedLines) {
	e := start
	for {
		// Gather cut locations.
		ln := line{
			e.origin.coords,
			e.twin.origin.coords,
		}
		xys := []XY{ln.a, ln.b}
		indexed.tree.RangeSearch(ln.envelope().box(), func(i int) error {
			other := indexed.lines[i]
			inter := ln.intersectLine(other)
			if inter.empty {
				return nil
			}
			xys = append(xys, inter.ptA, inter.ptB)
			return nil
		})
		xys = sortAndUniquifyXYs(xys) // TODO: make common function

		// Reverse order to match direction of edge.
		if xys[0] != ln.a {
			for i := 0; i < len(xys)/2; i++ {
				j := len(xys) - i - 1
				xys[i], xys[j] = xys[j], xys[i]
			}
		}

		// Perform cuts.
		cuts := len(xys) - 2
		for i := 0; i < cuts; i++ {
			xy := xys[i+1]
			cutVert, ok := d.vertices[xy]
			if !ok {
				cutVert = &vertexRecord{
					coords:   xy,
					incident: nil, /* populated later */
				}
				d.vertices[xy] = cutVert
			}
			d.reNodeEdge(e, cutVert)
			e = e.next
		}
		e = e.next

		if e == start {
			break
		}
	}
}

func (d *doublyConnectedEdgeList) reNodeEdge(e *halfEdgeRecord, cut *vertexRecord) {
	// Store original values we need later.
	dest := e.twin.origin
	next := e.next

	// Create new edges.
	ePrime := &halfEdgeRecord{
		origin:   cut,
		twin:     nil, // populated later
		incident: e.incident,
		next:     next,
		prev:     e,
	}
	ePrimeTwin := &halfEdgeRecord{
		origin:   dest,
		twin:     ePrime,
		incident: e.twin.incident,
		next:     e.twin,
		prev:     next.twin,
	}
	ePrime.twin = ePrimeTwin

	e.twin.origin = cut
	e.next = ePrime
	next.twin.next = ePrimeTwin
	next.prev = ePrime
	e.twin.prev = ePrimeTwin
	e.prev.twin.prev = e.twin
	cut.incident = ePrime
	dest.incident = ePrimeTwin

	d.halfEdges = append(d.halfEdges, ePrime, ePrimeTwin)
}
