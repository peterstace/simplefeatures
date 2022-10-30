package geom

import (
	"fmt"
)

type doublyConnectedEdgeList struct {
	faces     []*faceRecord // only populated in the overlay
	halfEdges []*halfEdgeRecord
	vertices  map[XY]*vertexRecord
}

type faceRecord struct {
	cycle     *halfEdgeRecord
	labels    [2]label
	extracted bool
}

func (f *faceRecord) String() string {
	if f == nil {
		return "nil"
	}
	return "[" + f.cycle.String() + "]"
}

type halfEdgeRecord struct {
	origin       *vertexRecord
	twin         *halfEdgeRecord
	incident     *faceRecord // only populated in the overlay
	next, prev   *halfEdgeRecord
	intermediate []XY
	edgeLabels   [2]label
	faceLabels   [2]label
	extracted    bool
}

// secondXY gives the second (1-indexed) XY in the edge. This is either the
// first intermediate XY, or the origin of the next/twin edge in the case where
// there are no intermediates.
func (e *halfEdgeRecord) secondXY() XY {
	if len(e.intermediate) == 0 {
		return e.twin.origin.coords
	}
	return e.intermediate[0]
}

// String shows the origin and destination of the edge (for debugging
// purposes). We can remove this once DCEL active development is completed.
func (e *halfEdgeRecord) String() string {
	if e == nil {
		return "nil"
	}
	return fmt.Sprintf("%v->%v->%v", e.origin.coords, e.intermediate, e.twin.origin.coords)
}

type vertexRecord struct {
	coords    XY
	incidents []*halfEdgeRecord
	labels    [2]label
	locations [2]location
	extracted bool
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

func newDCELFromGeometry(g Geometry, ghosts MultiLineString, operand operand, interactions map[XY]struct{}) *doublyConnectedEdgeList {
	dcel := &doublyConnectedEdgeList{
		vertices: make(map[XY]*vertexRecord),
	}
	dcel.addGeometry(g, operand, interactions)
	dcel.addGhosts(ghosts, operand, interactions)
	return dcel
}

func (d *doublyConnectedEdgeList) addGeometry(g Geometry, operand operand, interactions map[XY]struct{}) {
	switch g.Type() {
	case TypePolygon:
		poly := g.MustAsPolygon()
		d.addMultiPolygon(poly.AsMultiPolygon(), operand, interactions)
	case TypeMultiPolygon:
		mp := g.MustAsMultiPolygon()
		d.addMultiPolygon(mp, operand, interactions)
	case TypeLineString:
		mls := g.MustAsLineString().AsMultiLineString()
		d.addMultiLineString(mls, operand, interactions)
	case TypeMultiLineString:
		mls := g.MustAsMultiLineString()
		d.addMultiLineString(mls, operand, interactions)
	case TypePoint:
		mp := g.MustAsPoint().AsMultiPoint()
		d.addMultiPoint(mp, operand)
	case TypeMultiPoint:
		mp := g.MustAsMultiPoint()
		d.addMultiPoint(mp, operand)
	case TypeGeometryCollection:
		gc := g.MustAsGeometryCollection()
		d.addGeometryCollection(gc, operand, interactions)
	default:
		panic(fmt.Sprintf("unknown geometry type: %v", g.Type()))
	}
}

func (d *doublyConnectedEdgeList) addMultiPolygon(mp MultiPolygon, operand operand, interactions map[XY]struct{}) {
	mp = mp.ForceCCW()

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
				if _, ok := interactions[xy]; !ok {
					continue
				}
				if _, ok := d.vertices[xy]; !ok {
					vr := &vertexRecord{
						coords:    xy,
						incidents: nil, // populated later
						labels:    newHalfPopulatedLabels(operand, true),
						locations: newLocationsOnBoundary(operand),
					}
					d.vertices[xy] = vr
				}
			}
		}

		for _, ring := range rings {
			var newEdges []*halfEdgeRecord
			forEachNonInteractingSegment(ring, interactions, func(segment []XY) {
				// Construct the internal points slices.
				intermediateFwd := segment[1 : len(segment)-1]
				intermediateRev := reverseXYs(intermediateFwd)

				// Build the edges (fwd and rev).
				vertA := d.vertices[segment[0]]
				vertB := d.vertices[segment[len(segment)-1]]
				internalEdge := &halfEdgeRecord{
					origin:       vertA,
					twin:         nil, // populated later
					incident:     nil, // only populated in the overlay
					next:         nil, // populated later
					prev:         nil, // populated later
					intermediate: intermediateFwd,
					edgeLabels:   newHalfPopulatedLabels(operand, true),
					faceLabels:   newHalfPopulatedLabels(operand, true),
				}
				externalEdge := &halfEdgeRecord{
					origin:       vertB,
					twin:         internalEdge,
					incident:     nil, // only populated in the overlay
					next:         nil, // populated later
					prev:         nil, // populated later
					intermediate: intermediateRev,
					edgeLabels:   newHalfPopulatedLabels(operand, true),
					faceLabels:   newHalfPopulatedLabels(operand, false),
				}
				internalEdge.twin = externalEdge
				vertA.incidents = append(vertA.incidents, internalEdge)
				vertB.incidents = append(vertB.incidents, externalEdge)
				newEdges = append(newEdges, internalEdge, externalEdge)
			})

			// Link together next/prev pointers.
			numEdges := len(newEdges)
			for i := 0; i < numEdges/2; i++ {
				newEdges[i*2+0].next = newEdges[(2*i+2)%numEdges]
				newEdges[i*2+1].next = newEdges[(2*i-1+numEdges)%numEdges]
				newEdges[i*2+0].prev = newEdges[(2*i-2+numEdges)%numEdges]
				newEdges[i*2+1].prev = newEdges[(2*i+3)%numEdges]
			}
			d.halfEdges = append(d.halfEdges, newEdges...)
		}
	}
}

func (d *doublyConnectedEdgeList) addMultiLineString(mls MultiLineString, operand operand, interactions map[XY]struct{}) {
	// Add vertices.
	for i := 0; i < mls.NumLineStrings(); i++ {
		ls := mls.LineStringN(i)
		seq := ls.Coordinates()
		n := seq.Length()
		for j := 0; j < n; j++ {
			xy := seq.GetXY(j)
			if _, ok := interactions[xy]; !ok {
				continue
			}

			onBoundary := (j == 0 || j == n-1) && !ls.IsClosed()
			if v, ok := d.vertices[xy]; !ok {
				var locs [2]location
				if onBoundary {
					locs[operand].boundary = true
				} else {
					locs[operand].interior = true
				}
				d.vertices[xy] = &vertexRecord{
					xy,
					nil, // populated later
					newHalfPopulatedLabels(operand, true),
					locs,
					false,
				}
			} else {
				if onBoundary {
					if v.locations[operand].boundary {
						// Handle mod-2 rule (the boundary passes through the point
						// an even number of times, then it should be treated as an
						// interior point).
						v.locations[operand].boundary = false
						v.locations[operand].interior = true
					} else {
						v.locations[operand].boundary = true
						v.locations[operand].interior = false
					}
				} else {
					v.locations[operand].interior = true
				}
			}
		}
	}

	edges := make(edgeSet)

	// Add edges.
	for i := 0; i < mls.NumLineStrings(); i++ {
		seq := mls.LineStringN(i).Coordinates()
		forEachNonInteractingSegment(seq, interactions, func(segment []XY) {
			startXY := segment[0]
			endXY := segment[len(segment)-1]

			intermediateFwd := segment[1 : len(segment)-1]
			intermediateRev := reverseXYs(intermediateFwd)

			if edges.containsStartIntermediateEnd(startXY, intermediateFwd, endXY) {
				return
			}

			vOrigin := d.vertices[startXY]
			vDestin := d.vertices[endXY]

			fwd := &halfEdgeRecord{
				origin:       vOrigin,
				twin:         nil, // set later
				incident:     nil, // only populated in overlay
				next:         nil, // set later
				prev:         nil, // set later
				intermediate: intermediateFwd,
				edgeLabels:   newHalfPopulatedLabels(operand, true),
				faceLabels:   newUnpopulatedLabels(),
			}
			rev := &halfEdgeRecord{
				origin:       vDestin,
				twin:         fwd,
				incident:     nil, // only populated in overlay
				next:         fwd,
				prev:         fwd,
				intermediate: intermediateRev,
				edgeLabels:   newHalfPopulatedLabels(operand, true),
				faceLabels:   newUnpopulatedLabels(),
			}
			fwd.twin = rev
			fwd.next = rev
			fwd.prev = rev

			edges.insertEdge(fwd)
			edges.insertEdge(rev)

			vOrigin.incidents = append(vOrigin.incidents, fwd)
			vDestin.incidents = append(vDestin.incidents, rev)

			d.halfEdges = append(d.halfEdges, fwd, rev)
		})
	}
}

func (d *doublyConnectedEdgeList) addMultiPoint(mp MultiPoint, operand operand) {
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		xy, ok := mp.PointN(i).XY()
		if !ok {
			continue
		}
		record, ok := d.vertices[xy]
		if !ok {
			record = &vertexRecord{
				coords:    xy,
				incidents: nil,
				labels:    [2]label{},    // set below
				locations: [2]location{}, // set below
			}
			d.vertices[xy] = record
		}
		record.labels[operand] = label{populated: true, inSet: true}
		record.locations[operand].interior = true
	}
}

func (d *doublyConnectedEdgeList) addGeometryCollection(gc GeometryCollection, operand operand, interactions map[XY]struct{}) {
	n := gc.NumGeometries()
	for i := 0; i < n; i++ {
		d.addGeometry(gc.GeometryN(i), operand, interactions)
	}
}

func (d *doublyConnectedEdgeList) addGhosts(mls MultiLineString, operand operand, interactions map[XY]struct{}) {
	edges := make(edgeSet)
	for _, e := range d.halfEdges {
		edges.insertEdge(e)
	}

	for i := 0; i < mls.NumLineStrings(); i++ {
		seq := mls.LineStringN(i).Coordinates()
		forEachNonInteractingSegment(seq, interactions, func(segment []XY) {
			startXY := segment[0]
			endXY := segment[len(segment)-1]
			intermediateFwd := segment[1 : len(segment)-1]
			intermediateRev := reverseXYs(intermediateFwd)

			if _, ok := d.vertices[startXY]; !ok {
				d.vertices[startXY] = &vertexRecord{coords: startXY, incidents: nil, labels: [2]label{}}
			}
			if _, ok := d.vertices[endXY]; !ok {
				d.vertices[endXY] = &vertexRecord{coords: endXY, incidents: nil, labels: [2]label{}}
			}

			if edges.containsStartIntermediateEnd(startXY, intermediateFwd, endXY) {
				// Already exists, so shouldn't add.
				return
			}

			fwd, rev := d.addGhostLine(startXY, intermediateFwd, intermediateRev, endXY, operand)
			edges.insertEdge(fwd)
			edges.insertEdge(rev)
		})
	}
}

func (d *doublyConnectedEdgeList) addGhostLine(startXY XY, intermediateFwd, intermediateRev []XY, endXY XY, operand operand) (*halfEdgeRecord, *halfEdgeRecord) {
	vertA := d.vertices[startXY]
	vertB := d.vertices[endXY]

	fwd := &halfEdgeRecord{
		origin:       vertA,
		twin:         nil, // populated later
		incident:     nil, // only populated in the overlay
		next:         nil, // popluated later
		prev:         nil, // populated later
		intermediate: intermediateFwd,
		edgeLabels:   newHalfPopulatedLabels(operand, false),
		faceLabels:   [2]label{},
	}
	rev := &halfEdgeRecord{
		origin:       vertB,
		twin:         fwd,
		incident:     nil, // only populated in the overlay
		next:         fwd,
		prev:         fwd,
		intermediate: intermediateRev,
		edgeLabels:   newHalfPopulatedLabels(operand, false),
		faceLabels:   [2]label{},
	}
	fwd.twin = rev
	fwd.next = rev
	fwd.prev = rev

	vertA.incidents = append(vertA.incidents, fwd)
	vertB.incidents = append(vertB.incidents, rev)

	d.halfEdges = append(d.halfEdges, fwd, rev)

	d.fixVertex(vertA)
	d.fixVertex(vertB)

	return fwd, rev
}

func forEachNonInteractingSegment(seq Sequence, interactions map[XY]struct{}, fn func([]XY)) {
	n := seq.Length()
	i := 0
	for i < n-1 {
		// Find the next interaction point after i. This will be the
		// end of the next non-interacting segment.
		start := i
		var end int
		for j := i + 1; j < n; j++ {
			if _, ok := interactions[seq.GetXY(j)]; ok {
				end = j
				break
			}
		}

		// Construct the segment.
		segment := make([]XY, end-start+1)
		for j := range segment {
			segment[j] = seq.GetXY(start + j)
		}

		// Execute the callback with the segment.
		fn(segment)

		// On the next iteration, start the next edge at the end of
		// this one.
		i = end
	}
}
