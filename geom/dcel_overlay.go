package geom

import (
	"math"
	"sort"
)

func (d *doublyConnectedEdgeList) overlay(other *doublyConnectedEdgeList) error {
	d.overlayVertices(other)
	d.overlayEdges(other)
	d.fixVertices()
	if err := d.reAssignFaces(); err != nil {
		return err
	}
	d.fixLabels()
	return nil
}

func (d *doublyConnectedEdgeList) overlayVertices(other *doublyConnectedEdgeList) {
	for xy, otherVert := range other.vertices {
		vert, ok := d.vertices[xy]
		if ok {
			vert.label |= otherVert.label
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

	edgeRecords := make(map[line]*halfEdgeRecord)
	for _, e := range d.halfEdges {
		ln := line{e.origin.coords, e.next.origin.coords}
		edgeRecords[ln] = e
		e.origin.incidents = append(e.origin.incidents, e)
	}

	for _, e := range other.halfEdges {
		ln := line{e.origin.coords, e.next.origin.coords}
		if existing, ok := edgeRecords[ln]; ok {
			existing.edgeLabel |= e.edgeLabel
			existing.faceLabel |= e.faceLabel
		} else {
			edgeRecords[ln] = e
			e.origin = d.vertices[ln.a]
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
			di := ei.twin.origin.coords.Sub(ei.origin.coords)
			dj := ej.twin.origin.coords.Sub(ej.origin.coords)
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
func (d *doublyConnectedEdgeList) reAssignFaces() error {
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
			label: 0, // populated below
		}
		d.faces = append(d.faces, f)
		forEachEdge(cycle, func(e *halfEdgeRecord) {
			f.label |= e.faceLabel
			e.incident = f
		})
	}

	for _, face := range d.faces {
		d.completePartialFaceLabel(face)
	}
	return nil
}

// completePartialFaceLabel checks to see if the face label for the given face
// is complete (i.e. contains a part for both A and B). If it's not complete,
// then in searches adjacent faces until it finds a face that it can copy the
// missing part of the label from. This situation occurs whenever a face in the
// overlay DCEL doesn't have any edges from one of the original geometries.
func (d *doublyConnectedEdgeList) completePartialFaceLabel(face *faceRecord) {
	labelIsComplete := func() bool {
		return (face.label & populatedMask) == populatedMask
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
			face.label = completeLabel(face.label, adj.label)
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
	face.label |= populatedMask
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

// completeLabel copies any missing portion (part A or part B) of the label
// from donor to recipient, and then returns recipient.
func completeLabel(recipient, donor uint8) uint8 {
	if (recipient & inputAPopulated) == 0 {
		recipient |= (donor & inputAMask)
	}
	if (recipient & inputBPopulated) == 0 {
		recipient |= (donor & inputBMask)
	}
	return recipient
}

// fixLabels updates edge and vertex labels after performing an overlay.
func (d *doublyConnectedEdgeList) fixLabels() {
	for _, e := range d.halfEdges {
		// Copy labels from incident faces into edge since the edge represents
		// the (closed) border of the face.
		e.edgeLabel |= e.incident.label | e.twin.incident.label

		// If we haven't seen an edge label yet for one of the two input
		// geometries, we can assume that we'll never see it. So we mark off
		// that side as having the bit populated.
		e.edgeLabel |= populatedMask

		// Copy edge labels onto the labels of adjacent vertices. This is
		// because the vertices represent the endpoints of the edges, and
		// should have at least those bits set.
		e.origin.label |= e.edgeLabel | e.prev.edgeLabel
	}
}
