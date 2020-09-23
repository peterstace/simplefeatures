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
	case TypeLineString:
		mls := g.AsLineString().AsMultiLineString()
		return newDCELFromMultiLineString(mls, mask)
	case TypeMultiLineString:
		mls := g.AsMultiLineString()
		return newDCELFromMultiLineString(mls, mask)
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
		label:           populatedMask & mask,
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
					label:           populatedMask & mask,
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

func newDCELFromMultiLineString(mls MultiLineString, mask uint8) *doublyConnectedEdgeList {
	dcel := &doublyConnectedEdgeList{
		vertices: make(map[XY]*vertexRecord),
	}

	// Add vertices.
	for i := 0; i < mls.NumLineStrings(); i++ {
		ls := mls.LineStringN(i)
		seq := ls.Coordinates()
		for j := 0; j < seq.Length(); j++ {
			xy := seq.GetXY(j)
			dcel.vertices[xy] = &vertexRecord{coords: xy, label: mask}
		}
	}

	// Linear elements have no face structure, so everything just points to the
	// infinite face.
	infFace := &faceRecord{
		outerComponent:  nil,
		innerComponents: nil,
		label:           mask & populatedMask,
	}
	dcel.faces = []*faceRecord{infFace}

	// Add edges.
	for i := 0; i < mls.NumLineStrings(); i++ {
		var newEdges []*halfEdgeRecord
		ls := mls.LineStringN(i)
		seq := ls.Coordinates()
		for j := 0; j < seq.Length(); j++ {
			ln, ok := getLine(seq, j)
			if !ok {
				continue
			}
			vOrigin := dcel.vertices[ln.a]
			vDestin := dcel.vertices[ln.b]
			fwd := &halfEdgeRecord{
				origin:   vOrigin,
				twin:     nil, // set later
				incident: infFace,
				next:     nil, // set later
				prev:     nil, // set later
				label:    mask,
			}
			rev := &halfEdgeRecord{
				origin:   vDestin,
				twin:     fwd,
				incident: infFace,
				next:     nil, // set later
				prev:     nil, // set later
				label:    mask,
			}
			fwd.twin = rev
			newEdges = append(newEdges, fwd, rev)
			vOrigin.incident = fwd
			vDestin.incident = rev
		}
		n := len(newEdges)
		for j, e := range newEdges {
			if j%2 == 0 {
				if j+2 < n {
					e.next = newEdges[j+2]
				}
				if j-2 >= 0 {
					e.prev = newEdges[j-2]
				}
			} else {
				if j-2 >= 0 {
					e.next = newEdges[j-2]
				}
				if j+2 < n {
					e.prev = newEdges[j+2]
				}
			}
		}
		newEdges[0].prev = newEdges[1]
		newEdges[1].next = newEdges[0]
		newEdges[n-2].next = newEdges[n-1]
		newEdges[n-1].prev = newEdges[n-2]

		dcel.halfEdges = append(dcel.halfEdges, newEdges...)
		infFace.innerComponents = append(infFace.innerComponents, newEdges[0])
	}

	return dcel
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
	if dst.label&inputAPopulated == 0 {
		dst.label |= (src.label & inputAMask)
	}
	if dst.label&inputBPopulated == 0 {
		dst.label |= (src.label & inputBMask)
	}
}

// edgeLoopLeftmostLowest finds the edge in the cycle whose origin is the
// leftmost (or lowest for a tie) point in the loop.
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
			twins[e] = struct{}{}
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

	// TODO: do we need to set the presence bits for all vertices? They might not be set yet in some cases.
}

// extractGeometry converts the DECL into a Geometry that represents it.
func (d *doublyConnectedEdgeList) extractGeometry(include func(uint8) bool) Geometry {
	areals := d.extractPolygons(include)
	linears := d.extractLineStrings(include)
	points := d.extractPoints(include)

	switch {
	case len(areals) > 0 && len(linears) == 0 && len(points) == 0:
		if len(areals) == 1 {
			return areals[0].AsGeometry()
		}
		mp, err := NewMultiPolygonFromPolygons(areals)
		if err != nil {
			panic(fmt.Sprintf("could not create MultiPolygon: %v", err))
		}
		return mp.AsGeometry()
	case len(areals) == 0 && len(linears) > 0 && len(points) == 0:
		if len(linears) == 1 {
			return linears[0].AsGeometry()
		}
		return NewMultiLineStringFromLineStrings(linears).AsGeometry()
	case len(areals) == 0 && len(linears) == 0 && len(points) > 0:
		if len(points) == 1 {
			return NewPointFromXY(points[0]).AsGeometry()
		}
		coords := make([]float64, 2*len(points))
		for i, xy := range points {
			coords[i*2+0] = xy.X
			coords[i*2+1] = xy.Y
		}
		return NewMultiPoint(NewSequence(coords, DimXY)).AsGeometry()
	default:
		geoms := make([]Geometry, 0, len(areals)+len(linears)+len(points))
		for _, poly := range areals {
			geoms = append(geoms, poly.AsGeometry())
		}
		for _, ls := range linears {
			geoms = append(geoms, ls.AsGeometry())
		}
		for _, xy := range points {
			geoms = append(geoms, NewPointFromXY(xy).AsGeometry())
		}
		return NewGeometryCollection(geoms).AsGeometry()
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
		for f := range facesInPoly {
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

				// Mark all edges and vertices intersecting with the polygon as
				// being extracted.  This will prevent them being considered
				// during linear and point geometry extraction.
				edge.label |= extracted
				edge.twin.label |= extracted
				edge.origin.label |= extracted

				if seen[edge] {
					return
				}
				if include(edge.twin.incident.label) {
					// Adjacent face is in the polygon, so this edge cannot be part
					// of the boundary.
					seen[edge] = true
					return
				}
				seq := extractPolygonBoundary(facesInPoly, edge, seen)
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

func extractPolygonBoundary(faceSet map[*faceRecord]bool, start *halfEdgeRecord, seen map[*halfEdgeRecord]bool) Sequence {
	var coords []float64
	e := start
	for {
		seen[e] = true
		xy := e.origin.coords
		coords = append(coords, xy.X, xy.Y)

		// Sweep through the edges around the vertex (in a counter-clockwise
		// order) until we find the next edge that is part of the polygon
		// boundary.
		e = e.twin.prev.twin
		for !faceSet[e.incident] {
			e = e.prev.twin
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
func findFacesMakingPolygon(include func(uint8) bool, start *faceRecord) map[*faceRecord]bool {
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
	return expanded
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

// TODO: Line extraction isn't working too well at the moment. It's currently
// extracting each line individually, which isn't intended. It might be better
// to return a []line here, and then construct back into LineString and
// MultiLineString as a separate logical step since it seems tricky to do
// inline.

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
		e.origin.label |= extracted
		e.twin.origin.label |= extracted

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

// extractPoints extracts any vertices in the DCEL that should be part of the
// output geometry, but aren't yet represented as part of any previously
// extracted geometries.
func (d *doublyConnectedEdgeList) extractPoints(include func(uint8) bool) []XY {
	var xys []XY
	for xy, vert := range d.vertices {
		if include(vert.label) && vert.label&extracted == 0 {
			vert.label |= extracted
			xys = append(xys, xy)
		}
	}
	return xys
}
