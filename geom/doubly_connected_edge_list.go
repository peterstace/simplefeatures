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
