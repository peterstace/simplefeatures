package geom

import (
	"fmt"
)

type doublyConnectedEdgeList struct {
	faces     []*faceRecord // only populated in the overlay
	halfEdges edgeSet
	vertices  map[XY]*vertexRecord
}

func newDCEL() *doublyConnectedEdgeList {
	return &doublyConnectedEdgeList{
		faces:     nil,
		halfEdges: make(map[[2]XY]*halfEdgeRecord),
		vertices:  make(map[XY]*vertexRecord),
	}
}

type faceRecord struct {
	cycle     *halfEdgeRecord
	inSet     [2]bool
	extracted bool
}

type halfEdgeRecord struct {
	origin     *vertexRecord
	twin       *halfEdgeRecord
	incident   *faceRecord // only populated in the overlay
	next, prev *halfEdgeRecord
	seq        Sequence
	srcEdge    [2]bool
	srcFace    [2]bool
	inSet      [2]bool
	extracted  bool
}

// secondXY gives the second (1-indexed) XY in the edge. This is either the
// first intermediate XY, or the origin of the next/twin edge in the case where
// there are no intermediates.
func (e *halfEdgeRecord) secondXY() XY {
	return e.seq.GetXY(1)
}

func (e *halfEdgeRecord) xys() []XY {
	xys := make([]XY, e.seq.Length())
	for i := range xys {
		xys[i] = e.seq.GetXY(i)
	}
	return xys
}

type vertexRecord struct {
	coords    XY
	incidents map[*halfEdgeRecord]struct{}

	src       [2]bool
	inSet     [2]bool
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
			e.start.locations[operand].boundary = true
			e.fwd.srcEdge[operand] = true
			e.rev.srcEdge[operand] = true
			e.fwd.srcFace[operand] = true
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

		// TODO: do we need to do this when adding polygons as well?
		// TODO: is there a better way to model location? Could it just be a tri-value enum?

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
	record, ok := d.vertices[xy]
	if !ok {
		record = &vertexRecord{
			coords:    xy,
			incidents: make(map[*halfEdgeRecord]struct{}),
			src:       [2]bool{},     // set below
			locations: [2]location{}, // set below
		}
		d.vertices[xy] = record
	}
	record.src[operand] = true
	record.locations[operand].interior = true
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
			// No need to update labels/locations, since anything added is a "ghost" point.
			_ = d.addOrGetEdge(segment)
		})
	}
}

type edge struct {
	start, end *vertexRecord
	fwd, rev   *halfEdgeRecord
}

func (d *doublyConnectedEdgeList) addOrGetEdge(segment Sequence) edge {
	n := segment.Length()
	if n < 2 {
		panic(fmt.Sprintf("segment of length less than 2: %d", n))
	}

	startXY := segment.GetXY(0)
	endXY := segment.GetXY(n - 1)
	reverseSegment := reverseSequence(segment)

	fwd, addedFwd := d.halfEdges.addOrGet(segment)
	rev, addedRev := d.halfEdges.addOrGet(reverseSegment)
	if addedFwd != addedRev {
		panic(fmt.Sprintf("addedFwd != addedRev: %t vs %t", addedFwd, addedRev))
	}

	startV, ok := d.vertices[startXY]
	if !ok {
		startV = &vertexRecord{
			coords:    startXY,
			incidents: make(map[*halfEdgeRecord]struct{}),
		}
		d.vertices[startXY] = startV
	}
	endV, ok := d.vertices[endXY]
	if !ok {
		endV = &vertexRecord{
			coords:    endXY,
			incidents: make(map[*halfEdgeRecord]struct{}),
		}
		d.vertices[endXY] = endV
	}

	startV.incidents[fwd] = struct{}{}
	endV.incidents[rev] = struct{}{}

	if addedFwd {
		fwd.origin = startV
		rev.origin = endV
		fwd.twin = rev
		rev.twin = fwd
		fwd.next = rev
		fwd.prev = rev
		rev.next = fwd
		rev.prev = fwd
		fwd.seq = segment
		rev.seq = reverseSegment
	}

	return edge{
		start: startV,
		end:   endV,
		fwd:   fwd,
		rev:   rev,
	}
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

		// Construct the segment.
		segment := seq.Slice(start, end+1)

		// Execute the callback with the segment.
		fn(segment, start)

		// On the next iteration, start the next edge at the end of
		// this one.
		i = end
	}
}
