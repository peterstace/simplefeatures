package geom

import (
	"fmt"
	"math"
	"sort"
)

type doublyConnectedEdgeList struct {
	faces     []*faceRecord
	halfEdges []*halfEdgeRecord
	vertices  map[XY]*vertexRecord
}

type faceRecord struct {
	outerComponent  *halfEdgeRecord
	innerComponents []*halfEdgeRecord
	label           uint8
}

type halfEdgeRecord struct {
	origin     *vertexRecord
	twin       *halfEdgeRecord
	incident   *faceRecord
	next, prev *halfEdgeRecord
	label      uint8
}

type vertexRecord struct {
	coords   XY
	incident *halfEdgeRecord
	label    uint8
}

func newDCELFromGeometry(g Geometry, mask uint8) *doublyConnectedEdgeList {
	switch g.Type() {
	case TypePolygon:
		poly := g.AsPolygon()
		return newDCELFromMultiPolygon(poly.AsMultiPolygon(), mask)
	case TypeMultiPolygon:
		mp := g.AsMultiPolygon()
		return newDCELFromMultiPolygon(mp, mask)
	default:
		// TODO: support all other input geometry types.
		panic(fmt.Sprintf("binary op not implemented for type %s", g.Type()))
	}
}

func newDCELFromMultiPolygon(mp MultiPolygon, mask uint8) *doublyConnectedEdgeList {
	mp = mp.ForceCCW()

	dcel := &doublyConnectedEdgeList{vertices: make(map[XY]*vertexRecord)}

	infFace := &faceRecord{
		outerComponent:  nil, // left nil
		innerComponents: nil, // populated later
		label:           presenceMask & mask,
	}
	dcel.faces = append(dcel.faces, infFace)

	for polyIdx := 0; polyIdx < mp.NumPolygons(); polyIdx++ {
		poly := mp.PolygonN(polyIdx)

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
					dcel.vertices[xy] = &vertexRecord{xy, nil /* populated later */, mask}
				}
			}
		}

		polyFace := &faceRecord{
			outerComponent:  nil, // populated later
			innerComponents: nil, // populated later
			label:           mask,
		}
		dcel.faces = append(dcel.faces, polyFace)

		for ringIdx, ring := range rings {
			interiorFace := polyFace
			exteriorFace := infFace
			if ringIdx > 0 {
				holeFace := &faceRecord{
					outerComponent:  nil, // left nil
					innerComponents: nil, // populated later
					label:           presenceMask & mask,
				}
				// For inner rings, the exterior face is a hole rather than the
				// infinite face.
				exteriorFace = holeFace
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
					label:    mask,
				}
				externalEdge := &halfEdgeRecord{
					origin:   dcel.vertices[ln.b],
					twin:     internalEdge,
					incident: exteriorFace,
					next:     nil, // populated later
					prev:     nil, // populated later
					label:    mask,
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
	}
	return dcel
}

func (d *doublyConnectedEdgeList) reNodeGraph(other []line) {
	indexed := newIndexedLines(other)
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
				cutVert = &vertexRecord{xy, nil /* populated later */, e.label}
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
		label:    e.label,
	}
	ePrimeTwin := &halfEdgeRecord{
		origin:   dest,
		twin:     ePrime,
		incident: e.twin.incident,
		next:     e.twin,
		prev:     next.twin,
		label:    e.label,
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
	faceLabels := d.overlayEdges(other)
	d.fixVertices()
	d.reAssignFaces(faceLabels)
	d.fixLabels()
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
			existing.label |= e.origin.label
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
	edgeRecords := make(map[line]*halfEdgeRecord)
	faceLabels := make(map[line]uint8)
	for _, e := range d.halfEdges {
		ln := line{e.origin.coords, e.next.origin.coords}
		edgeRecords[ln] = e
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
		}
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
func (d *doublyConnectedEdgeList) reAssignFaces(faceLabels map[line]uint8) {
	// Find all boundary cycles, and categorise them as either outer components
	// or inner components.
	var innerComponents, outerComponents []*halfEdgeRecord
	seen := make(map[*halfEdgeRecord]bool)
	for _, e := range d.halfEdges {
		if seen[e] {
			continue
		}
		leftmostLowest := edgeLoopLeftmostLowest(e)
		if edgeLoopIsOuterComponent(leftmostLowest) {
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
		nextLeft := d.findNextDownEdgeToTheLeft(leftmostLowest)
		if nextLeft != nil {
			// When there is no next left edge, then this indicates that the
			// current component is the inner component of the infinite face.
			// In this case, we *don't* want to find the lowest (or leftmost
			// for tie) edge, since there is no actual loop.
			nextLeft = edgeLoopLeftmostLowest(nextLeft)
		}
		graph.union(leftmostLowest, nextLeft)
	}

	// Construct new faces.
	d.faces = nil
	for _, set := range graph.sets {
		f := new(faceRecord)
		d.faces = append(d.faces, f)
		for _, e := range set {
			if e == nil || edgeLoopIsOuterComponent(e) {
				if f.outerComponent != nil {
					panic("double outer component")
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
		return (face.label & presenceMask) == presenceMask
	}
	expanded := make(map[*faceRecord]bool)
	stack := []*faceRecord{face}
	for len(stack) > 0 {
		popped := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		adjacent := adjacentFaces(popped)
		expanded[popped] = true
		for _, adj := range adjacent {
			completeFaceLabel(face, adj)
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

func completeFaceLabel(dst, src *faceRecord) {
	// TODO: is there a way to do this without treating the two halfs of the label separately?
	if dst.label&inputAPresent == 0 {
		dst.label |= (src.label & inputAMask)
	}
	if dst.label&inputBPresent == 0 {
		dst.label |= (src.label & inputBMask)
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

func (d *doublyConnectedEdgeList) findNextDownEdgeToTheLeft(edge *halfEdgeRecord) *halfEdgeRecord {
	var bestEdge *halfEdgeRecord
	var bestDist float64

	for _, e := range d.halfEdges {
		origin := e.origin.coords
		destin := e.next.origin.coords
		pt := edge.origin.coords
		if !(destin.Y <= pt.Y && pt.Y <= origin.Y) {
			// We only want to consider edges that go "down" (or horizontal)
			// and overlap vertically with pt.
			continue
		}
		if origin.Y == destin.Y && origin.X < destin.X {
			// For horizontal lines, we only want to consider edges that go
			// from the right to the left.
			continue
		}
		ln := line{origin, destin}
		dist := signedHorizontalDistanceBetweenXYAndLine(pt, ln)
		if dist <= 0 {
			// Edge is on the wrong side of pt (we only want edges to the left).
			continue
		}

		if bestEdge == nil || dist < bestDist {
			bestEdge = e
			bestDist = dist
		}
	}
	return bestEdge
}

func signedHorizontalDistanceBetweenXYAndLine(xy XY, ln line) float64 {
	// TODO: This is not robust in cases of an *almost* horizontal line. Need
	// to have a think about a better approach here.
	if ln.b.Y == ln.a.Y {
		return xy.X - math.Max(ln.a.X, ln.b.X)
	}
	rat := (xy.Y - ln.a.Y) / (ln.b.Y - ln.a.Y)
	x := (1-rat)*ln.a.X + rat*ln.b.X
	return xy.X - x
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

// union unions together the distinct sets containing e1 and e2.
func (s *disjointEdgeSet) union(e1, e2 *halfEdgeRecord) {
	idx1, idx2 := -1, -1
	for i, set := range s.sets {
		for _, e := range set {
			if e == e1 {
				if idx1 != -1 {
					panic(idx1)
				}
				idx1 = i
			}
			if e == e2 {
				if idx2 != -1 {
					panic(idx2)
				}
				idx2 = i
			}
		}
	}
	if idx1 == -1 || idx2 == -1 || idx1 == idx2 {
		panic(fmt.Sprintf("e1: %p e2: %p state: %v indexes: %d %d", e1, e2, s.sets, idx1, idx2))
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
}

// fixLabels updates edge and vertex labels after performing an overlay.
func (d *doublyConnectedEdgeList) fixLabels() {
	for _, e := range d.halfEdges {
		// Add edge presence if the two faces adjacent to the edge are both
		// present. The edge is part of that geometry (because it's "within"
		// it), even if it's not explicitly part of its boundary.
		face1 := e.incident.label
		face2 := e.twin.incident.label
		e.label |= face1 & face2 & valueMask

		// If we haven't seen an edge label yet for one of the two input
		// geometries, we can assume that we'll never see it. So we mark off
		// that side as having the bit populated.
		e.label |= presenceMask

		// Copy edge labels onto the labels of adjacent vertices. This is
		// because the vertices represent the endpoints of the edges, and
		// should have at least those bits set.
		e.origin.label |= e.label | e.prev.label
	}

	// TODO: do we need to set the presence bits for all vertices? They might not be set yet in some cases.
}

// extractGeometry converts the DECL into a Geometry that represents it.
//
// TODO: extract point geometries as well.
func (d *doublyConnectedEdgeList) extractGeometry(include func(uint8) bool) Geometry {
	areals := d.extractPolygons(include)
	linears := d.extractLineStrings(include)

	var arealGeom Geometry
	if len(areals) == 1 {
		arealGeom = areals[0].AsGeometry()
	} else if len(areals) > 1 {
		mp, err := NewMultiPolygonFromPolygons(areals)
		if err != nil {
			panic(fmt.Sprintf("could not create MultiPolygon: %v", err))
		}
		arealGeom = mp.AsGeometry()
	}

	var linearGeom Geometry
	if len(linears) == 1 {
		linearGeom = linears[0].AsGeometry()
	} else if len(linears) > 1 {
		linearGeom = NewMultiLineStringFromLineStrings(linears).AsGeometry()
	}

	switch {
	case !arealGeom.IsEmpty() && !linearGeom.IsEmpty():
		return NewGeometryCollection([]Geometry{arealGeom, linearGeom}).AsGeometry()
	case !arealGeom.IsEmpty():
		return arealGeom
	case !linearGeom.IsEmpty():
		return linearGeom
	default:
		return GeometryCollection{}.AsGeometry()
	}
}

func (d *doublyConnectedEdgeList) extractPolygons(include func(uint8) bool) []Polygon {
	var polys []Polygon
	for _, face := range d.faces {
		if (face.label & extracted) != 0 {
			continue
		}
		if !include(face.label) {
			continue
		}

		// Find all faces that make up the polygon.
		facesInPoly := findFacesMakingPolygon(include, face)

		// Find all edge cycles incident to the faces. Edges in these cycles
		// are are candidates to be part of the Polygon boundary.
		var components []*halfEdgeRecord
		for _, f := range facesInPoly {
			f.label |= extracted
			if cmp := f.outerComponent; cmp != nil {
				components = append(components, cmp)
			}
			components = append(components, f.innerComponents...)
		}

		// Extract the Polygon boundaries from the candidate edges.
		var rings []LineString
		seen := make(map[*halfEdgeRecord]bool)
		for _, cmp := range components {
			forEachEdge(cmp, func(edge *halfEdgeRecord) {
				if seen[edge] {
					return
				}
				if include(edge.twin.incident.label) {
					// Adjacent face is in the polygon, so this edge cannot be part
					// of the boundary.
					seen[edge] = true
					return
				}
				seq := extractPolygonBoundary(include, edge, seen)
				ring, err := NewLineString(seq)
				if err != nil {
					panic(fmt.Sprintf("could not create LineString: %v", err))
				}
				rings = append(rings, ring)
			})
		}

		// Construct the polygon.
		orderCCWRingFirst(rings)
		poly, err := NewPolygonFromRings(rings)
		if err != nil {
			panic(fmt.Sprintf("could not create Polygon: %v", err))
		}
		polys = append(polys, poly)
	}
	return polys
}

func extractPolygonBoundary(include func(uint8) bool, start *halfEdgeRecord, seen map[*halfEdgeRecord]bool) Sequence {
	var coords []float64
	e := start
	for {
		seen[e] = true
		xy := e.origin.coords
		coords = append(coords, xy.X, xy.Y)

		// Sweep through the edges around the vertex until we find the next
		// edge that is part of the polygon boundary.
		e = e.next
		for include(e.twin.incident.label) {
			e = e.twin.next
		}

		if e == start {
			break
		}
	}

	coords = append(coords, coords[:2]...)
	return NewSequence(coords, DimXY)
}

// findFacesMakingPolygon finds all faces that belong to the polygon that
// contains the start face (according to the given inclusion criteria).
func findFacesMakingPolygon(include func(uint8) bool, start *faceRecord) []*faceRecord {
	expanded := make(map[*faceRecord]bool)
	toExpand := make(map[*faceRecord]bool)
	toExpand[start] = true
	pop := func() *faceRecord {
		for f := range toExpand {
			delete(toExpand, f)
			return f
		}
		panic("could not pop")
	}

	for len(toExpand) > 0 {
		popped := pop()
		adj := adjacentFaces(popped)
		expanded[popped] = true
		for _, f := range adj {
			if !include(f.label) {
				continue
			}
			if expanded[f] {
				continue
			}
			if toExpand[f] {
				continue
			}
			toExpand[f] = true
		}
	}

	list := make([]*faceRecord, 0, len(expanded))
	for f := range expanded {
		list = append(list, f)
	}
	return list
}

// orderCCWRingFirst reorders rings such that if it contains at least one CCW
// ring, then a CCW ring is the first element.
func orderCCWRingFirst(rings []LineString) {
	for i, r := range rings {
		if ccw := signedAreaOfLinearRing(r, nil) > 0; ccw {
			rings[i], rings[0] = rings[0], rings[i]
			return
		}
	}
}

func (d *doublyConnectedEdgeList) extractLineStrings(include func(uint8) bool) []LineString {
	var lss []LineString
	for _, e := range d.halfEdges {
		if shouldExtractLine(e, include) {
			ls := extractLineString(e, include)
			lss = append(lss, ls)
		}
	}
	return lss
}

func extractLineString(e *halfEdgeRecord, include func(uint8) bool) LineString {
	u := e.origin.coords
	coords := []float64{u.X, u.Y}

	for {
		v := e.next.origin.coords
		coords = append(coords, v.X, v.Y)
		e.label |= extracted
		e.twin.label |= extracted

		e = nextNoBranch(e, include)
		if e == nil {
			break
		}
	}

	seq := NewSequence(coords, DimXY)
	ls, err := NewLineString(seq)
	if err != nil {
		// Shouldn't ever happen, since we have at least one edge.
		panic(fmt.Sprintf("could not construct line string using %v: %v", coords, err))
	}
	return ls
}

func shouldExtractLine(e *halfEdgeRecord, include func(uint8) bool) bool {
	return (e.label&extracted == 0) && include(e.label) && !include(e.incident.label) && !include(e.twin.incident.label)
}

// nextNoBranch checks to see if the given edge has multiple next edges that it
// could use for linear extraction. If there are multiple edges, then nil is
// returned (this is called a 'branch'). If there is just one possible next
// edge, then that next edge is returned.
func nextNoBranch(edge *halfEdgeRecord, include func(uint8) bool) *halfEdgeRecord {
	e := edge.next
	var nextEdge *halfEdgeRecord

	// Find the first next edge.
	for {
		if e == edge.twin {
			// There are no linear branches that could be extracted.
			return nil
		}
		if shouldExtractLine(e, include) {
			nextEdge = e
			break
		}
		e = e.twin.next
	}

	// Check to see if there are additional next edges (i.e. a branch scenario).
	for {
		if e == edge.twin {
			// There is no branching.
			return nextEdge
		}
		if shouldExtractLine(e, include) {
			// There is branching, so indicate this by returning nil.
			return nil
		}
		e = e.twin.next
	}
}
