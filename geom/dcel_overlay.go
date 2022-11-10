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

	interactions := findInteractionPoints([]Geometry{a, b, ghosts.AsGeometry()})

	dcel := newDCEL()
	dcel.addVertices(interactions)
	dcel.addGhosts(ghosts, interactions)
	dcel.addGeometry(a, operandA, interactions)
	dcel.addGeometry(b, operandB, interactions)

	dcel.fixVertices()
	dcel.reAssignFaces()
	dcel.populateInSetLabels()

	//dumpDCEL(dcel)

	return dcel, nil
}

func (d *doublyConnectedEdgeList) fixVertices() {
	for _, vert := range d.vertices {
		d.fixVertex(vert)
	}
}

func (d *doublyConnectedEdgeList) fixVertex(v *vertexRecord) {
	// Sort the edges radially.
	var incidents []*halfEdgeRecord
	for e := range v.incidents {
		incidents = append(incidents, e)
	}
	if len(incidents) >= 3 {
		// TODO: consider using a solution like
		// https://stackoverflow.com/questions/6989100/sort-points-in-clockwise-order
		// instead of using trigonometry.
		sort.Slice(incidents, func(i, j int) bool {
			ei := incidents[i]
			ej := incidents[j]
			di := ei.seq.GetXY(1).Sub(ei.seq.GetXY(0))
			dj := ej.seq.GetXY(1).Sub(ej.seq.GetXY(0))
			aI := math.Atan2(di.Y, di.X)
			aJ := math.Atan2(dj.Y, dj.X)
			return aI < aJ
		})
	}

	// Fix pointers.
	for i := range incidents {
		ei := incidents[i]
		ej := incidents[(i+1)%len(v.incidents)]
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
	for _, e := range d.halfEdges {
		// Copy labels from incident faces into edge since the edge represents
		// the (closed) border of the face.
		e.inSet[0] = e.srcEdge[0] || e.incident.inSet[0] || e.twin.incident.inSet[0]
		e.inSet[1] = e.srcEdge[1] || e.incident.inSet[1] || e.twin.incident.inSet[1]

		// Copy edge labels onto the labels of adjacent vertices. This is
		// because the vertices represent the endpoints of the edges, and
		// should have at least those bits set.
		e.origin.inSet[0] = e.origin.src[0] || e.inSet[0] || e.prev.inSet[0]
		e.origin.inSet[1] = e.origin.src[1] || e.inSet[1] || e.prev.inSet[1]
	}
}
