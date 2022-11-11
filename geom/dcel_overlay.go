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
	d.populateInSetLabels()
}

func (d *doublyConnectedEdgeList) overlayVertices(other *doublyConnectedEdgeList) {
	for xy, otherVert := range other.vertices {
		vert, ok := d.vertices[xy]
		if ok {
			mergeBools(&vert.src, otherVert.src)
			mergeBools(&vert.inSet, otherVert.inSet)
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
			mergeBools(&existing.srcEdge, e.srcEdge)
			mergeBools(&existing.srcFace, e.srcFace)
			mergeBools(&existing.inSet, e.inSet)
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
			di := ei.seq.GetXY(1).Sub(ei.seq.GetXY(0))
			dj := ej.seq.GetXY(1).Sub(ej.seq.GetXY(0))
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
			cycle: cycle,
		}
		d.faces = append(d.faces, f)
		forEachEdge(cycle, func(e *halfEdgeRecord) {
			forEachOperand(func(operand operand) {
				if e.srcFace[operand] {
					f.inSet[operand] = true
				}
			})
			e.incident = f
		})
	}

	// Populate inSet for faces that did not have edges from their respective
	// input geometries.
	forEachOperand(func(operand operand) {
		visited := make(map[*faceRecord]bool)
		var dfs func(*faceRecord)
		dfs = func(f *faceRecord) {
			if visited[f] {
				return
			}
			visited[f] = true
			forEachEdge(f.cycle, func(e *halfEdgeRecord) {
				if !e.srcFace[operand] {
					e.twin.incident.inSet[operand] = true
					dfs(e.twin.incident)
				}
			})
		}
		for _, f := range d.faces {
			if f.inSet[operand] {
				dfs(f)
			}
		}
	})

	// If we couldn't find any cycles, then we wouldn't have constructed any
	// faces. This happens in the case where there are only point geometries.
	// We need to artificially create an infinite face.
	if len(d.faces) == 0 {
		d.faces = append(d.faces, &faceRecord{
			cycle: nil,
			inSet: [2]bool{},
		})
	}
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

// populateInSetLabels populates the inSet labels for edges and vertices.
func (d *doublyConnectedEdgeList) populateInSetLabels() {
	for _, v := range d.vertices {
		forEachOperand(func(op operand) {
			v.inSet[op] = v.src[op]
		})
	}
	for _, e := range d.halfEdges {
		forEachOperand(func(op operand) {
			// Copy labels from incident faces into edge since the edge
			// represents the (closed) border of the face.
			e.inSet[op] = e.srcEdge[op] || e.incident.inSet[op] || e.twin.incident.inSet[op]

			// Copy edge labels onto the labels of adjacent vertices. This is
			// because the vertices represent the endpoints of the edges, and
			// should have at least those bits set.
			e.origin.inSet[op] = e.origin.inSet[op] || e.inSet[op] || e.prev.inSet[op]
		})
	}
}
