package geom

import (
	"sort"
)

func (d *doublyConnectedEdgeList) fixVertices() {
	for _, vert := range d.vertices {
		d.fixVertex(vert)
	}
}

func (d *doublyConnectedEdgeList) fixVertex(v *vertexRecord) {
	// Create slice of incident edges so that they can be sorted radially.
	incidents := make([]*halfEdgeRecord, 0, len(v.incidents))
	for e := range v.incidents {
		incidents = append(incidents, e)
	}

	// If there are 2 or less edges, then the edges are already trivially
	// sorted around the vertex with relation to each other.
	alreadySorted := len(incidents) <= 2

	// Perform the sort.
	if !alreadySorted {
		sort.Slice(incidents, func(i, j int) bool {
			ei := incidents[i]
			ej := incidents[j]
			di := ei.seq.GetXY(1).Sub(ei.seq.GetXY(0))
			dj := ej.seq.GetXY(1).Sub(ej.seq.GetXY(0))
			return radialLess(di, dj)
		})
	}

	// Fix pointers.
	for i := range incidents {
		ei := incidents[i]
		ej := incidents[(i+1)%len(incidents)]
		ei.prev = ej.twin
		ej.twin.next = ei
	}
}

// radialLess provides an ordering for sorting vectors radially around the origin.
// This solution is a reworking of
// https://stackoverflow.com/questions/6989100/sort-points-in-clockwise-order
// to avoid using trigonometry.
func radialLess(di, dj XY) bool {
	if di.X >= 0 && dj.X < 0 {
		return true
	}
	if di.X < 0 && dj.X >= 0 {
		return false
	}
	if di.X == 0 && dj.X == 0 {
		if di.Y >= 0 || dj.Y >= 0 {
			return di.Y < dj.Y
		}
		return dj.Y < di.Y
	}

	// Due to the previous checks, di and dj must be in different sides (LHS vs
	// RHS) of the XY plane. Therefore the sign of the cross product can
	// provide an ordering within each half.
	if det := di.Cross(dj); det != 0 {
		return det > 0
	}

	// Points are on the same line from the center.
	// Check which point is further from the center.
	li := di.lengthSq()
	lj := dj.lengthSq()
	return li < lj
}

// assignFaces populates the face list based on half edge loops.
func (d *doublyConnectedEdgeList) assignFaces() {
	// Find all cycles.
	var cycles []*halfEdgeRecord
	seen := make(map[*halfEdgeRecord]bool)
	for _, e := range d.halfEdges {
		if seen[e] {
			continue
		}
		forEachEdgeInCycle(e, func(e *halfEdgeRecord) {
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
		forEachEdgeInCycle(cycle, func(e *halfEdgeRecord) {
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
			forEachEdgeInCycle(f.cycle, func(e *halfEdgeRecord) {
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
	forEachEdgeInCycle(f.cycle, func(e *halfEdgeRecord) {
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
