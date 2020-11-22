package geom

import (
	"errors"
	"fmt"
)

type doublyConnectedEdgeList struct {
	faces     []*faceRecord // only populated in the overlay
	halfEdges []*halfEdgeRecord
	vertices  map[XY]*vertexRecord
}

type faceRecord struct {
	cycle *halfEdgeRecord
	label uint8
}

func (f *faceRecord) String() string {
	if f == nil {
		return "nil"
	}
	return "[" + f.cycle.String() + "]"
}

type halfEdgeRecord struct {
	origin        *vertexRecord
	intermediates []XY
	twin          *halfEdgeRecord
	incident      *faceRecord // only populated in the overlay
	next, prev    *halfEdgeRecord
	edgeLabel     uint8
	faceLabel     uint8
}

// String shows the origin and destination of the edge (for debugging
// purposes). We can remove this once DCEL active development is completed.
func (e *halfEdgeRecord) String() string {
	if e == nil {
		return "nil"
	}
	return fmt.Sprintf("%v->%v", e.origin.coords, e.twin.origin.coords)
}

type vertexRecord struct {
	coords    XY
	incidents []*halfEdgeRecord
	label     uint8
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

func newDCELFromGeometry(g Geometry, ghosts MultiLineString, edgeColour map[line]byte, mask uint8) (*doublyConnectedEdgeList, error) {
	var dcel *doublyConnectedEdgeList
	switch g.Type() {
	case TypePolygon:
		poly := g.AsPolygon()
		dcel = newDCELFromMultiPolygon(poly.AsMultiPolygon(), edgeColour, mask)
	case TypeMultiPolygon:
		mp := g.AsMultiPolygon()
		dcel = newDCELFromMultiPolygon(mp, edgeColour, mask)
	case TypeLineString:
		mls := g.AsLineString().AsMultiLineString()
		dcel = newDCELFromMultiLineString(mls, edgeColour, mask)
	case TypeMultiLineString:
		mls := g.AsMultiLineString()
		dcel = newDCELFromMultiLineString(mls, edgeColour, mask)
	case TypePoint:
		mp := NewMultiPointFromPoints([]Point{g.AsPoint()})
		dcel = newDCELFromMultiPoint(mp, mask)
	case TypeMultiPoint:
		mp := g.AsMultiPoint()
		dcel = newDCELFromMultiPoint(mp, mask)
	case TypeGeometryCollection:
		// TODO: Add support for GeometryCollection inputs.
		return nil, errors.New("GeometryCollection argument not supported")
	default:
		panic(fmt.Sprintf("unknown geometry type: %v", g.Type()))
	}

	dcel.addGhosts(ghosts, mask)
	return dcel, nil
}

func newDCELFromMultiPolygon(mp MultiPolygon, edgeColour map[line]byte, mask uint8) *doublyConnectedEdgeList {
	mp = mp.ForceCCW()

	dcel := &doublyConnectedEdgeList{vertices: make(map[XY]*vertexRecord)}

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

		for _, ring := range rings {
			var newEdges []*halfEdgeRecord
			for i := 0; i < ring.Length(); i++ {
				ln, ok := getLine(ring, i)
				if !ok {
					continue
				}
				vertA := dcel.vertices[ln.a]
				vertB := dcel.vertices[ln.b]
				internalEdge := &halfEdgeRecord{
					origin:    vertA,
					twin:      nil, // populated later
					incident:  nil, // only populated in the overlay
					next:      nil, // populated later
					prev:      nil, // populated later
					edgeLabel: mask,
					faceLabel: mask,
				}
				externalEdge := &halfEdgeRecord{
					origin:    vertB,
					twin:      internalEdge,
					incident:  nil, // only populated in the overlay
					next:      nil, // populated later
					prev:      nil, // populated later
					edgeLabel: mask,
					faceLabel: mask & populatedMask,
				}
				internalEdge.twin = externalEdge
				vertA.incidents = append(vertA.incidents, internalEdge)
				vertB.incidents = append(vertB.incidents, externalEdge)
				newEdges = append(newEdges, internalEdge, externalEdge)
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

func newDCELFromMultiLineString(mls MultiLineString, edgeColour map[line]byte, mask uint8) *doublyConnectedEdgeList {
	dcel := &doublyConnectedEdgeList{
		vertices: make(map[XY]*vertexRecord),
	}

	// Add vertices.
	for i := 0; i < mls.NumLineStrings(); i++ {
		ls := mls.LineStringN(i)
		seq := ls.Coordinates()
		for j := 0; j < seq.Length(); j++ {
			xy := seq.GetXY(j)
			dcel.vertices[xy] = &vertexRecord{xy, nil /* populated later */, mask}
		}
	}

	// Linear elements have no face structure, so everything just points to the
	// infinite face.
	type vertPair struct {
		a, b *vertexRecord
	}
	edgeSet := make(map[vertPair]bool)

	// Add edges.
	for i := 0; i < mls.NumLineStrings(); i++ {
		seq := mls.LineStringN(i).Coordinates()
		for j := 0; j < seq.Length(); j++ {
			ln, ok := getLine(seq, j)
			if !ok {
				continue
			}

			vOrigin := dcel.vertices[ln.a]
			vDestin := dcel.vertices[ln.b]

			pair := vertPair{vOrigin, vDestin}
			if pair.a.coords.Less(pair.b.coords) {
				pair.a, pair.b = pair.b, pair.a
			}
			if edgeSet[pair] {
				continue
			}
			edgeSet[pair] = true

			fwd := &halfEdgeRecord{
				origin:    vOrigin,
				twin:      nil, // set later
				incident:  nil, // only populated in overlay
				next:      nil, // set later
				prev:      nil, // set later
				edgeLabel: mask,
				faceLabel: mask & populatedMask,
			}
			rev := &halfEdgeRecord{
				origin:    vDestin,
				twin:      fwd,
				incident:  nil, // only populated in overlay
				next:      fwd,
				prev:      fwd,
				edgeLabel: mask,
				faceLabel: mask & populatedMask,
			}
			fwd.twin = rev
			fwd.next = rev
			fwd.prev = rev

			vOrigin.incidents = append(vOrigin.incidents, fwd)
			vDestin.incidents = append(vDestin.incidents, rev)

			dcel.halfEdges = append(dcel.halfEdges, fwd, rev)
		}
	}
	return dcel
}

func newDCELFromMultiPoint(mp MultiPoint, mask uint8) *doublyConnectedEdgeList {
	dcel := &doublyConnectedEdgeList{vertices: make(map[XY]*vertexRecord)}
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		xy, ok := mp.PointN(i).XY()
		if !ok {
			continue
		}
		record, ok := dcel.vertices[xy]
		if !ok {
			record = &vertexRecord{
				coords:    xy,
				incidents: nil,
				label:     0,
			}
			dcel.vertices[xy] = record
		}
		record.label |= mask
	}
	return dcel
}

func (d *doublyConnectedEdgeList) addGhosts(mls MultiLineString, mask uint8) {
	edges := make(map[line]*halfEdgeRecord)
	for _, e := range d.halfEdges {
		ln := line{e.origin.coords, e.twin.origin.coords}
		edges[ln] = e
	}

	for i := 0; i < mls.NumLineStrings(); i++ {
		seq := mls.LineStringN(i).Coordinates()
		n := seq.Length()
		for j := 0; j < n; j++ {
			if ln, ok := getLine(seq, j); ok {
				if _, ok := d.vertices[ln.a]; !ok {
					d.vertices[ln.a] = &vertexRecord{coords: ln.a, incidents: nil, label: 0}
				}
				if _, ok := d.vertices[ln.b]; !ok {
					d.vertices[ln.b] = &vertexRecord{coords: ln.b, incidents: nil, label: 0}
				}
				d.addGhostLine(ln, mask, edges)
			}
		}
	}
}

func (d *doublyConnectedEdgeList) addGhostLine(ln line, mask uint8, edges map[line]*halfEdgeRecord) {
	if _, ok := edges[ln]; ok {
		// Already exists, so shouldn't add.
		return
	}

	vertA := d.vertices[ln.a]
	vertB := d.vertices[ln.b]

	e1 := &halfEdgeRecord{
		origin:    vertA,
		twin:      nil, // populated later
		incident:  nil, // only populated in the overlay
		next:      nil, // popluated later
		prev:      nil, // populated later
		edgeLabel: mask & populatedMask,
		faceLabel: 0,
	}
	e2 := &halfEdgeRecord{
		origin:    vertB,
		twin:      e1,
		incident:  nil, // only populated in the overlay
		next:      e1,
		prev:      e1,
		edgeLabel: mask & populatedMask,
		faceLabel: 0,
	}
	e1.twin = e2
	e1.next = e2
	e1.prev = e2

	vertA.incidents = append(vertA.incidents, e1)
	vertB.incidents = append(vertB.incidents, e2)

	d.halfEdges = append(d.halfEdges, e1, e2)

	edges[ln] = e1
	ln.a, ln.b = ln.b, ln.a
	edges[ln] = e2

	d.fixVertex(vertA)
	d.fixVertex(vertB)
}
