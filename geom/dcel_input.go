package geom

import "fmt"

func (d *doublyConnectedEdgeList) addVertices(interactions map[XY]struct{}) {
	for xy := range interactions {
		d.vertices[xy] = &vertexRecord{
			coords:    xy,
			incidents: make(map[*halfEdgeRecord]struct{}),
		}
	}
}

func (d *doublyConnectedEdgeList) addGeometry(g Geometry, operand operand, interactions map[XY]struct{}) {
	switch g.Type() {
	case TypePolygon:
		poly := g.MustAsPolygon()
		d.addPolygon(poly, operand, interactions)
	case TypeMultiPolygon:
		mp := g.MustAsMultiPolygon()
		d.addMultiPolygon(mp, operand, interactions)
	case TypeLineString:
		ls := g.MustAsLineString()
		d.addLineString(ls, operand, interactions)
	case TypeMultiLineString:
		mls := g.MustAsMultiLineString()
		d.addMultiLineString(mls, operand, interactions)
	case TypePoint:
		pt := g.MustAsPoint()
		d.addPoint(pt, operand)
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
	for i := 0; i < mp.NumPolygons(); i++ {
		d.addPolygon(mp.PolygonN(i), operand, interactions)
	}
}

func (d *doublyConnectedEdgeList) addPolygon(poly Polygon, operand operand, interactions map[XY]struct{}) {
	poly = poly.ForceCCW()
	for _, ring := range poly.DumpRings() {
		forEachNonInteractingSegment(ring.Coordinates(), interactions, func(segment Sequence, _ int) {
			e := d.addOrGetEdge(segment)
			e.start.src[operand] = true
			e.end.src[operand] = true
			e.fwd.srcEdge[operand] = true
			e.rev.srcEdge[operand] = true
			e.fwd.srcFace[operand] = true

			// TODO: is this treatment of boundary correct? It may not follow
			// the odd-even rule if the node occurs multiple times due to
			// geometry collections.
			e.start.locations[operand].boundary = true
		})
	}
}

func (d *doublyConnectedEdgeList) addMultiLineString(mls MultiLineString, operand operand, interactions map[XY]struct{}) {
	for i := 0; i < mls.NumLineStrings(); i++ {
		d.addLineString(mls.LineStringN(i), operand, interactions)
	}
}

func (d *doublyConnectedEdgeList) addLineString(ls LineString, operand operand, interactions map[XY]struct{}) {
	seq := ls.Coordinates()
	forEachNonInteractingSegment(seq, interactions, func(segment Sequence, startIdx int) {
		edge := d.addOrGetEdge(segment)
		edge.start.src[operand] = true
		edge.end.src[operand] = true
		edge.fwd.srcEdge[operand] = true
		edge.rev.srcEdge[operand] = true

		// TODO: This modelling of location is complicated. Could it just be a
		// tri-value enum instead?

		for _, c := range [2]struct {
			v          *vertexRecord
			onBoundary bool
		}{
			{edge.start, startIdx == 0 && !ls.IsClosed()},
			{edge.end, startIdx+segment.Length() == seq.Length() && !ls.IsClosed()},
		} {
			if !c.v.locations[operand].boundary && !c.v.locations[operand].interior {
				if c.onBoundary {
					c.v.locations[operand].boundary = true
				} else {
					c.v.locations[operand].interior = true
				}
			} else {
				if c.onBoundary {
					if c.v.locations[operand].boundary {
						c.v.locations[operand].boundary = false
						c.v.locations[operand].interior = true
					} else {
						c.v.locations[operand].boundary = true
						c.v.locations[operand].interior = false
					}
				} else {
					c.v.locations[operand].interior = true
				}
			}
		}
	})
}

func (d *doublyConnectedEdgeList) addMultiPoint(mp MultiPoint, operand operand) {
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		d.addPoint(mp.PointN(i), operand)
	}
}

func (d *doublyConnectedEdgeList) addPoint(pt Point, operand operand) {
	xy, ok := pt.XY()
	if !ok {
		return
	}
	v := d.vertices[xy]
	v.src[operand] = true
	v.locations[operand].interior = true
}

func (d *doublyConnectedEdgeList) addGeometryCollection(gc GeometryCollection, operand operand, interactions map[XY]struct{}) {
	n := gc.NumGeometries()
	for i := 0; i < n; i++ {
		d.addGeometry(gc.GeometryN(i), operand, interactions)
	}
}

func (d *doublyConnectedEdgeList) addGhosts(mls MultiLineString, interactions map[XY]struct{}) {
	for i := 0; i < mls.NumLineStrings(); i++ {
		seq := mls.LineStringN(i).Coordinates()
		forEachNonInteractingSegment(seq, interactions, func(segment Sequence, _ int) {
			// No need to update labels/locations since ghosts since these only
			// apply to input geometries.
			_ = d.addOrGetEdge(segment)
		})
	}
}

type edge struct {
	start, end *vertexRecord
	fwd, rev   *halfEdgeRecord
}

func (d *doublyConnectedEdgeList) addOrGetEdge(segment Sequence) edge {
	if n := segment.Length(); n < 2 {
		panic(fmt.Sprintf("segment of length less than 2: %d", n))
	}

	reverseSegment := segment.Reverse()
	startXY := segment.GetXY(0)
	endXY := reverseSegment.GetXY(0)

	fwd := d.getOrAddHalfEdge(segment)
	rev := d.getOrAddHalfEdge(reverseSegment)

	startV := d.vertices[startXY]
	endV := d.vertices[endXY]

	startV.incidents[fwd] = struct{}{}
	endV.incidents[rev] = struct{}{}

	fwd.origin = startV
	rev.origin = endV
	fwd.twin = rev
	rev.twin = fwd
	fwd.next = rev
	fwd.prev = rev
	rev.next = fwd
	rev.prev = fwd

	return edge{
		start: startV,
		end:   endV,
		fwd:   fwd,
		rev:   rev,
	}
}

func (d *doublyConnectedEdgeList) getOrAddHalfEdge(segment Sequence) *halfEdgeRecord {
	k := [2]XY{
		segment.GetXY(0),
		segment.GetXY(1),
	}
	e, ok := d.halfEdges[k]
	if !ok {
		e = &halfEdgeRecord{seq: segment}
		d.halfEdges[k] = e
	}
	return e
}

func forEachNonInteractingSegment(seq Sequence, interactions map[XY]struct{}, fn func(Sequence, int)) {
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

		// Execute the callback with the segment.
		segment := seq.Slice(start, end+1)
		fn(segment, start)

		// On the next iteration, start the next edge at the end of
		// this one.
		i = end
	}
}
