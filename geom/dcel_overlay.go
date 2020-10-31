package geom

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

func (d *doublyConnectedEdgeList) overlay(other *doublyConnectedEdgeList) error {
	d.overlayVertices(other)
	faceLabels := d.overlayEdges(other)
	d.fixVertices()
	if err := d.reAssignFaces(faceLabels); err != nil {
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

func (d *doublyConnectedEdgeList) overlayEdges(other *doublyConnectedEdgeList) map[line]uint8 {
	// Clear incidents lists, since we're going to re-compute them.
	for _, vert := range d.vertices {
		vert.incidents = nil
	}
	for _, vert := range other.vertices {
		vert.incidents = nil
	}

	edgeRecords := make(map[line]*halfEdgeRecord)
	faceLabels := make(map[line]uint8)
	for _, e := range d.halfEdges {
		ln := line{e.origin.coords, e.next.origin.coords}
		edgeRecords[ln] = e
		e.origin.incidents = append(e.origin.incidents, e)
		faceLabels[ln] = e.incident.label
	}

	for _, face := range other.faces {
		if cmp := face.outerComponent; cmp != nil {
			d.overlayEdgesInComponent(cmp, edgeRecords, faceLabels)
		}
		for _, cmp := range face.innerComponents {
			d.overlayEdgesInComponent(cmp, edgeRecords, faceLabels)
		}
	}
	return faceLabels
}

func (d *doublyConnectedEdgeList) overlayEdgesInComponent(start *halfEdgeRecord, edgeRecords map[line]*halfEdgeRecord, faceLabels map[line]uint8) {
	forEachEdge(start, func(e *halfEdgeRecord) {
		ln := line{e.origin.coords, e.next.origin.coords}

		label, ok := faceLabels[ln]
		if !ok {
			d.halfEdges = append(d.halfEdges, e)
		}
		label |= e.incident.label
		faceLabels[ln] = label

		if existing, ok := edgeRecords[ln]; ok {
			existing.label |= e.label
		} else {
			edgeRecords[ln] = e
			e.origin.incidents = append(e.origin.incidents, e)
		}
	})
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
func (d *doublyConnectedEdgeList) reAssignFaces(faceLabels map[line]uint8) error {
	// Find all boundary cycles, and categorise them as either outer components
	// or inner components.
	var innerComponents, outerComponents []*halfEdgeRecord
	seen := make(map[*halfEdgeRecord]bool)
	for _, e := range d.halfEdges {
		if seen[e] {
			continue
		}
		leftmostLowest := edgeLoopLeftmostLowest(e)
		if edgeLoopIsOuterComponent(e) {
			outerComponents = append(outerComponents, leftmostLowest)
		} else {
			innerComponents = append(innerComponents, leftmostLowest)
		}
		forEachEdge(e, func(e *halfEdgeRecord) {
			seen[e] = true
		})
	}

	// Group together boundary cycles that are for the same face.
	var graph disjointEdgeSet
	for _, e := range innerComponents {
		graph.addSingleton(e)
	}
	for _, e := range outerComponents {
		graph.addSingleton(e)
	}
	graph.addSingleton(nil) // nil represents the outer component of the infinite face

	for _, leftmostLowest := range innerComponents {
		nextLeft := d.findNextDownEdgeToTheLeft(leftmostLowest.origin.coords)
		if nextLeft != nil {
			// When there is no next left edge, then this indicates that the
			// current component is the inner component of the infinite face.
			// In this case, we *don't* want to find the lowest (or leftmost
			// for tie) edge, since there is no actual loop.
			nextLeft = edgeLoopLeftmostLowest(nextLeft)
		}
		if err := graph.union(leftmostLowest, nextLeft); err != nil {
			return err
		}
	}

	// Construct new faces.
	d.faces = nil
	for _, set := range graph.sets {
		f := new(faceRecord)
		d.faces = append(d.faces, f)
		for _, e := range set {
			if e == nil || edgeLoopIsOuterComponent(e) {
				if f.outerComponent != nil {
					return errors.New("double outer component")
				}
				f.outerComponent = e
			} else {
				f.innerComponents = append(f.innerComponents, e)
			}
			if e != nil {
				forEachEdge(e, func(e *halfEdgeRecord) {
					ln := line{e.origin.coords, e.next.origin.coords}
					f.label |= faceLabels[ln]
					e.incident = f
				})
			}
		}
		if f.outerComponent == nil {
			// The outer face never has geometries present on it, so we can
			// just mark it's label as being populated now.
			f.label |= populatedMask
		}
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
}

// adjacentFaces finds all of the faces that adjacent to f.
func adjacentFaces(f *faceRecord) []*faceRecord {
	set := make(map[*faceRecord]struct{})
	if cmp := f.outerComponent; cmp != nil {
		forEachEdge(cmp, func(e *halfEdgeRecord) {
			set[e.twin.incident] = struct{}{}
		})
	}
	for _, cmp := range f.innerComponents {
		forEachEdge(cmp, func(e *halfEdgeRecord) {
			set[e.twin.incident] = struct{}{}
		})
	}

	faces := make([]*faceRecord, 0, len(set))
	for face := range set {
		faces = append(faces, face)
	}
	return faces
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

// edgeLoopLeftmostLowest finds the edge in the cycle whose origin is the
// leftmost (or lowest for a tie) point in the loop. If there is a tie (i.e.
// two edges in the cycle with the same origin), then it uses the edge
// destination to choose between them.
func edgeLoopLeftmostLowest(start *halfEdgeRecord) *halfEdgeRecord {
	var best *halfEdgeRecord
	forEachEdge(start, func(e *halfEdgeRecord) {
		if best == nil ||
			e.origin.coords.Less(best.origin.coords) ||
			e.origin.coords == best.origin.coords && e.twin.origin.coords.Less(best.twin.origin.coords) {
			best = e
		}
	})
	return best
}

// edgeLoopIsOuterComponent checks to see if an edge loop is an outer edge loop
// or an inner edge loop. It does this by checking its orientation via the
// Shoelace Formula.
func edgeLoopIsOuterComponent(start *halfEdgeRecord) bool {
	// Check to see if the loop is degenerate (i.e. doesn't enclose any area).
	// Degenerate loops must always be inner components. We can't just use the
	// shoelace formula and check for an area of zero due to numerical
	// precision issues. Instead, we keep track of which twin edges we see
	// during edge cycle iteration -- if we never see a twin again then we can
	// be sure the cycle isn't degenerate.
	twins := make(map[*halfEdgeRecord]struct{})
	forEachEdge(start, func(e *halfEdgeRecord) {
		if _, ok := twins[e]; ok {
			delete(twins, e)
		} else {
			twins[e.twin] = struct{}{}
		}
	})
	if len(twins) == 0 {
		return false // degenerate
	}

	// Check the area of the loop using the shoelace formula. The sign of the
	// loop area implies an orientation, which in turn implies whether the loop
	// is an inner or outer component.
	var sum float64
	forEachEdge(start, func(e *halfEdgeRecord) {
		pt0 := e.origin.coords
		pt1 := e.next.origin.coords
		sum += (pt1.X + pt0.X) * (pt1.Y - pt0.Y)
	})
	return sum > 0
}

func (d *doublyConnectedEdgeList) findNextDownEdgeToTheLeft(pt XY) *halfEdgeRecord {
	var acc nextDownEdgeToTheLeftAccumulator
	var bestEdge *halfEdgeRecord
	for _, e := range d.halfEdges {
		ln := line{e.origin.coords, e.twin.origin.coords}
		if acc.accumulate(ln, pt) {
			bestEdge = e
		}
	}
	return bestEdge
}

type nextDownEdgeToTheLeftAccumulator struct {
	bestLine line
	bestDist float64
	tieBreak float64
}

func (a *nextDownEdgeToTheLeftAccumulator) accumulate(ln line, pt XY) bool {
	origin := ln.a
	destin := ln.b
	if !(destin.Y <= pt.Y && pt.Y <= origin.Y) {
		// We only want to consider edges that go "down" (or horizontal)
		// and overlap vertically with pt.
		return false
	}
	if origin.Y == destin.Y && origin.X < destin.X {
		// For horizontal lines, we only want to consider edges that go
		// from the right to the left.
		return false
	}

	// Calculate distance.
	dist := signedHorizontalDistanceBetweenXYAndLine(pt, ln)
	if dist <= 0 {
		// Edge is on the wrong side of pt (we only want edges to the left).
		return false
	}

	// Calculate tie-break.
	unitAwayFromHit := XY{1, 0}
	hit := XY{pt.X - dist, pt.Y}
	var other XY
	if hit.distanceTo(origin) > hit.distanceTo(destin) {
		other = origin
	} else {
		other = destin
	}
	edgeUnit := other.Sub(hit).Unit()
	tieBreak := unitAwayFromHit.Dot(edgeUnit)

	// Replace if best.
	if a.bestLine == (line{}) || dist < a.bestDist || (dist == a.bestDist && tieBreak > a.tieBreak) {
		a.bestLine = ln
		a.bestDist = dist
		a.tieBreak = tieBreak
		return true
	}
	return false
}

func signedHorizontalDistanceBetweenXYAndLine(xy XY, ln line) float64 {
	// TODO: This may not be robust in cases of an *almost* horizontal line.
	// Need to have a think about a better approach here.
	if ln.b.Y == ln.a.Y {
		return xy.X - math.Max(ln.a.X, ln.b.X)
	}
	rat := (xy.Y - ln.a.Y) / (ln.b.Y - ln.a.Y)
	x := (1-rat)*ln.a.X + rat*ln.b.X
	dist := xy.X - x
	return dist
}

// disjointEdgeSet is a disjoint set data structure where each element in the
// set is a *halfEdgeRecord (see
// https://en.wikipedia.org/wiki/Disjoint-set_data_structure).
//
// TODO: the implementation here is naive/brute-force. There are well known
// better disjoint edge set algorithms.
type disjointEdgeSet struct {
	sets [][]*halfEdgeRecord
}

// addSingleton adds a new set containing just e to the disjoint edge set.
func (s *disjointEdgeSet) addSingleton(e *halfEdgeRecord) {
	s.sets = append(s.sets, []*halfEdgeRecord{e})
}

// union unions together the distinct sets containing e1 and e2. It *shouldn't*
// ever return an error (however we want to be cautious not to panic).
func (s *disjointEdgeSet) union(e1, e2 *halfEdgeRecord) error {
	idx1, idx2 := -1, -1
	for i, set := range s.sets {
		for _, e := range set {
			if e == e1 {
				if idx1 != -1 {
					return fmt.Errorf("idx1 already set: %d", idx1)
				}
				idx1 = i
			}
			if e == e2 {
				if idx2 != -1 {
					return fmt.Errorf("idx2 already set: %d", idx2)
				}
				idx2 = i
			}
		}
	}
	if idx1 == -1 || idx2 == -1 || idx1 == idx2 {
		return fmt.Errorf("e1: %p e2: %p state: %v indexes: %d %d", e1, e2, s.sets, idx1, idx2)
	}

	set1 := s.sets[idx1]
	set2 := s.sets[idx2]
	n := len(s.sets)
	s.sets[idx1], s.sets[n-1] = s.sets[n-1], s.sets[idx1]
	if idx2 == n-1 {
		idx2 = idx1
	}
	s.sets[idx2], s.sets[n-2] = s.sets[n-2], s.sets[idx2]
	s.sets = s.sets[:n-2]
	s.sets = append(s.sets, append(set1, set2...))
	return nil
}

// fixLabels updates edge and vertex labels after performing an overlay.
func (d *doublyConnectedEdgeList) fixLabels() {
	for _, e := range d.halfEdges {
		// Add edge presence if the two faces adjacent to the edge are both
		// present. The edge is part of that geometry (because it's "within"
		// it), even if it's not explicitly part of its boundary.
		face1 := e.incident.label
		face2 := e.twin.incident.label
		e.label |= face1 & face2 & inSetMask

		// If we haven't seen an edge label yet for one of the two input
		// geometries, we can assume that we'll never see it. So we mark off
		// that side as having the bit populated.
		e.label |= populatedMask

		// Copy edge labels onto the labels of adjacent vertices. This is
		// because the vertices represent the endpoints of the edges, and
		// should have at least those bits set.
		e.origin.label |= e.label | e.prev.label
	}

	var infFace *faceRecord
	for _, f := range d.faces {
		if f.outerComponent == nil {
			infFace = f
			break
		}
	}

	// If there are any vertices that don't have populated labels, it's because
	// they are isolated (i.e. in the middle of a face). We need to work out
	// which face they are part of.
	for _, vert := range d.vertices {
		xy := vert.coords
		if (vert.label & populatedMask) == populatedMask {
			continue
		}
		var face *faceRecord
		e := d.findNextDownEdgeToTheLeft(xy)
		if e == nil {
			// There was no next edge to the left, so the face we're interested
			// in is the infinite face.
			face = infFace
		} else {
			face = e.incident
		}
		vert.label = completeLabel(vert.label, face.label)
		face.internalVertices = append(face.internalVertices, vert)
	}
}
