package geom

import (
	"math"
	"sort"
)

func createOverlay(a, b Geometry) (*doublyConnectedEdgeList, error) {
	var points []XY
	points = appendComponentPoints(points, a)
	points = appendComponentPoints(points, b)
	ghosts := spanningTree(points)

	a, b, ghosts, err := reNodeGeometries(a, b, ghosts)
	if err != nil {
		return nil, wrap(err, "re-noding")
	}

	interactionPoints := findInteractionPoints([]Geometry{a, b, ghosts.AsGeometry()})

	dcelA := newDCELFromGeometry(a, ghosts, operandA, interactionPoints)
	dcelB := newDCELFromGeometry(b, ghosts, operandB, interactionPoints)

	dcelA.overlay(dcelB)
	return dcelA, nil
}

func (d *doublyConnectedEdgeList) overlay(other *doublyConnectedEdgeList) {
	d.overlayVertices(other)
	d.overlayEdges(other)
	d.fixVertices()
	d.reAssignFaces()
	d.fixLabels()
}

func (d *doublyConnectedEdgeList) overlayVertices(other *doublyConnectedEdgeList) {
	for xy, otherVert := range other.vertices {
		vert, ok := d.vertices[xy]
		if ok {
			mergeLabels(&vert.labels, otherVert.labels)
			mergeLocations(&vert.locations, otherVert.locations)
		} else {
			d.vertices[xy] = otherVert
		}
	}
	for _, e := range other.halfEdges {
		if existing, ok := d.vertices[e.origin.coords]; ok {
			e.origin = existing
		} else {
			d.vertices[e.origin.coords] = e.origin
		}
	}
}

func (d *doublyConnectedEdgeList) overlayEdges(other *doublyConnectedEdgeList) {
	// Clear incidents lists, since we're going to re-compute them.
	for _, vert := range d.vertices {
		vert.incidents = nil
	}
	for _, vert := range other.vertices {
		vert.incidents = nil
	}

	edges := make(edgeSet)
	for _, e := range d.halfEdges {
		edges.insertEdge(e)
		e.origin.incidents = append(e.origin.incidents, e)
	}

	for _, e := range other.halfEdges {
		if existing, ok := edges.lookupEdge(e); ok {
			mergeLabels(&existing.edgeLabels, e.edgeLabels)
			mergeLabels(&existing.faceLabels, e.faceLabels)
		} else {
			edges.insertEdge(e)
			e.origin = d.vertices[e.origin.coords]
			e.origin.incidents = append(e.origin.incidents, e)
			d.halfEdges = append(d.halfEdges, e)
		}
	}
}

func (d *doublyConnectedEdgeList) fixVertices() {
	for _, vert := range d.vertices {
		d.fixVertex(vert)
	}
}

func (d *doublyConnectedEdgeList) fixVertex(v *vertexRecord) {
	// Sort the edges radially.
	if len(v.incidents) >= 3 {
		sort.Slice(v.incidents, func(i, j int) bool {
			ei := v.incidents[i]
			ej := v.incidents[j]
			di := ei.secondXY().Sub(ei.origin.coords)
			dj := ej.secondXY().Sub(ej.origin.coords)
			aI := math.Atan2(di.Y, di.X)
			aJ := math.Atan2(dj.Y, dj.X)
			return aI < aJ
		})
	}

	// Fix pointers.
	for i := range v.incidents {
		ei := v.incidents[i]
		ej := v.incidents[(i+1)%len(v.incidents)]
		ei.prev = ej.twin
		ej.twin.next = ei
	}
}

// reAssignFaces clears the DCEL face list and creates new faces based on the
// half edge loops.
func (d *doublyConnectedEdgeList) reAssignFaces() {
	// Find all cycles.
	var cycles []*halfEdgeRecord
	seen := make(map[*halfEdgeRecord]bool)
	for _, e := range d.halfEdges {
		if seen[e] {
			continue
		}
		forEachEdge(e, func(e *halfEdgeRecord) {
			seen[e] = true
		})
		cycles = append(cycles, e)
	}

	// Construct new faces.
	d.faces = nil
	for _, cycle := range cycles {
		f := &faceRecord{
			cycle:  cycle,
			labels: [2]label{}, // populated below
		}
		d.faces = append(d.faces, f)
		forEachEdge(cycle, func(e *halfEdgeRecord) {
			mergeLabels(&f.labels, e.faceLabels)
			e.incident = f
		})
	}

	// If we couldn't find any cycles, then we wouldn't have constructed any
	// faces. This happens in the case where there are only point geometries.
	// We need to artificially create an infinite face.
	if len(d.faces) == 0 {
		d.faces = append(d.faces, &faceRecord{
			cycle:  nil,
			labels: newPopulatedLabels(false),
		})
	}

	for _, face := range d.faces {
		d.completePartialFaceLabel(face)
	}
}

// completePartialFaceLabel checks to see if the face label for the given face
// is complete (i.e. contains a part for both A and B). If it's not complete,
// then in searches adjacent faces until it finds a face that it can copy the
// missing part of the label from. This situation occurs whenever a face in the
// overlay DCEL doesn't have any edges from one of the original geometries.
func (d *doublyConnectedEdgeList) completePartialFaceLabel(face *faceRecord) {
	labelIsComplete := func() bool {
		return face.labels[0].populated && face.labels[1].populated
	}
	if labelIsComplete() {
		return
	}
	expanded := make(map[*faceRecord]bool)
	stack := []*faceRecord{face}
	for len(stack) > 0 {
		popped := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		adjacent := adjacentFaces(popped)
		expanded[popped] = true
		for _, adj := range adjacent {
			face.labels = completeLabels(face.labels, adj.labels)
			if labelIsComplete() {
				return
			}
			if !expanded[adj] {
				stack = append(stack, adj)
			}
		}
	}

	// It's possible that we're still missing part of the face label. This
	// could happen if one of the inputs is a Point/MultiPoint input because
	// its associated ghost lines would not add to the label pool. We can
	// safely fill in the presence bits for this case.
	face.labels[0].populated = true
	face.labels[1].populated = true
}

// adjacentFaces finds all of the faces that adjacent to f.
func adjacentFaces(f *faceRecord) []*faceRecord {
	var adjacent []*faceRecord
	set := make(map[*faceRecord]bool)
	forEachEdge(f.cycle, func(e *halfEdgeRecord) {
		adj := e.twin.incident
		if !set[adj] {
			set[adj] = true
			adjacent = append(adjacent, adj)
		}
	})
	return adjacent
}

// completeLabels copies any missing portion (part A or part B) of the label
// from donor to recipient, and then returns recipient.
func completeLabels(recipient, donor [2]label) [2]label {
	for i := 0; i < 2; i++ {
		if !recipient[i].populated && donor[i].populated {
			recipient[i].populated = true
			recipient[i].inSet = donor[i].inSet
		}
	}
	return recipient
}

// fixLabels updates edge and vertex labels after performing an overlay.
func (d *doublyConnectedEdgeList) fixLabels() {
	for _, e := range d.halfEdges {
		// Copy labels from incident faces into edge since the edge represents
		// the (closed) border of the face.
		mergeLabels(&e.edgeLabels, e.incident.labels)
		mergeLabels(&e.edgeLabels, e.twin.incident.labels)

		// If we haven't seen an edge label yet for one of the two input
		// geometries, we can assume that we'll never see it. So we mark off
		// that side as having the bit populated.
		e.edgeLabels[0].populated = true
		e.edgeLabels[1].populated = true

		// Copy edge labels onto the labels of adjacent vertices. This is
		// because the vertices represent the endpoints of the edges, and
		// should have at least those bits set.
		mergeLabels(&e.origin.labels, e.edgeLabels)
		mergeLabels(&e.origin.labels, e.prev.edgeLabels)
	}
}
