package geom

import (
	"math"
	"sort"
)

type doublyConnectedEdgeList struct {
	faces     []*faceRecord
	halfEdges []*halfEdgeRecord // TODO: I don't think this is a great way of tracking the half edges.
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

func (d *doublyConnectedEdgeList) overlay(other *doublyConnectedEdgeList) {
	d.overlayVertices(other)
	d.overlayEdges(other)
	d.fixVertices()
	d.reAssignFaces()
}

func (d *doublyConnectedEdgeList) overlayVertices(other *doublyConnectedEdgeList) {
	for _, face := range other.faces {
		if cmp := face.outerComponent; cmp != nil {
			d.overlayVerticesInComponent(cmp)
		}
		for _, cmp := range face.innerComponents {
			d.overlayVerticesInComponent(cmp)
		}
	}
}

func (d *doublyConnectedEdgeList) overlayVerticesInComponent(start *halfEdgeRecord) {
	forEachEdge(start, func(e *halfEdgeRecord) {
		if existing, ok := d.vertices[e.origin.coords]; ok {
			e.origin = existing
		} else {
			d.vertices[e.origin.coords] = e.origin
		}
	})
}

func forEachEdge(start *halfEdgeRecord, fn func(*halfEdgeRecord)) {
	e := start
	for {
		fn(e)
		e = e.next
		if e == start {
			break
		}
	}
}

func (d *doublyConnectedEdgeList) overlayEdges(other *doublyConnectedEdgeList) {
	for _, face := range other.faces {
		if cmp := face.outerComponent; cmp != nil {
			d.overlayEdgesInComponent(cmp)
		}
		for _, cmp := range face.innerComponents {
			d.overlayEdgesInComponent(cmp)
		}
	}
}

func (d *doublyConnectedEdgeList) overlayEdgesInComponent(start *halfEdgeRecord) {
	// TODO: should handle the case where some half edges overlap with existing ones.
	forEachEdge(start, func(e *halfEdgeRecord) {
		d.halfEdges = append(d.halfEdges, e)
	})
}

func (d *doublyConnectedEdgeList) fixVertices() {
	for xy := range d.vertices {
		d.fixVertex(xy)
	}
}

func (d *doublyConnectedEdgeList) fixVertex(v XY) {
	// Find edges that start at v.
	//
	// TODO: This is not efficient, we should use an acceleration structure
	// rather than a linear search.
	var incident []*halfEdgeRecord
	for _, e := range d.halfEdges {
		if e.origin.coords == v {
			incident = append(incident, e)
		}
	}

	// Sort the edges radially.
	//
	// TODO: Might be able to use regular vector operations rather than
	// trigonometry here.
	sort.Slice(incident, func(i, j int) bool {
		ei := incident[i]
		ej := incident[j]
		di := ei.twin.origin.coords.Sub(ei.origin.coords)
		dj := ej.twin.origin.coords.Sub(ej.origin.coords)
		aI := math.Atan2(di.Y, di.X)
		aJ := math.Atan2(dj.Y, dj.X)
		return aI < aJ
	})

	// Fix pointers.
	for i := range incident {
		ei := incident[i]
		ej := incident[(i+1)%len(incident)]
		ei.prev = ej.twin
		ej.twin.next = ei
	}
}

// reAssignFaces clears the DCEL face list and creates new faces based on the
// half edge loops.
//
// TODO: We currently make the assumption that there is a 1-1 mapping between
// faces and half edge loops. This assumption only holds for faces that have a
// single inner component (and no outer component), or just an outer component
// (with no inner components).
func (d *doublyConnectedEdgeList) reAssignFaces() {
	d.faces = nil
	seen := make(map[*halfEdgeRecord]bool)
	for _, e := range d.halfEdges {
		if seen[e] {
			continue
		}

		// A new face record is created for each edge loop.
		// mark each edge in the loop as visited, we will only create a single
		// face per loop.
		f := new(faceRecord)
		d.faces = append(d.faces, f)

		leftmostLowest := edgeLoopLeftmostLowest(e)
		if edgeLoopIsOuterComponent(leftmostLowest) {
			f.outerComponent = leftmostLowest
		} else {
			f.innerComponents = append(f.innerComponents, leftmostLowest)
		}

		forEachEdge(e, func(e *halfEdgeRecord) {
			seen[e] = true
			e.incident = f
		})
	}
}

// edgeLoopLeftmostLowest finds the edge whose origin is the leftmost (or
// lowest for a tie) point in the loop.
func edgeLoopLeftmostLowest(start *halfEdgeRecord) *halfEdgeRecord {
	var best *halfEdgeRecord
	forEachEdge(start, func(e *halfEdgeRecord) {
		if best == nil || e.origin.coords.Less(best.origin.coords) {
			best = e
		}
	})
	return best
}

// edgeLoopIsOuterComponent checks to see if an edge loop is an outer edge loop
// or an inner edge loop. It does this by examining the edge whose origin is
// the leftmost (or lowest for ties) in the loop.
func edgeLoopIsOuterComponent(leftmostLowest *halfEdgeRecord) bool {
	// We can look at the next and prev points relative to the leftmost (then
	// lowest) point in the cycle. Then we can use orientation of the triplet
	// to determine if we're looking at an outer or inner component. This works
	// because outer components are wound CCW and inner components are wound CW.
	prev := leftmostLowest.prev.origin.coords
	here := leftmostLowest.origin.coords
	next := leftmostLowest.next.origin.coords
	return orientation(prev, here, next) == leftTurn
}
