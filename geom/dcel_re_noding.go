package geom

import (
	"fmt"
	"math"
	"sort"
)

// infLineWithLineIntersection extends infinitely the two input lines and finds
// their intersection. In degenerate cases where the lines are parallel, then
// the result will be at infinity.
func infLineWithLineIntersection(ln1, ln2 line) XY {
	// TODO: I took this formula directly from Wikipedia. But it's definitely
	// possible to refactor it to be a bit easier to read and understand using
	// vector notation.
	a, b, c, d := ln1.a, ln1.b, ln2.a, ln2.b
	numerX := (a.X*b.Y-a.Y*b.X)*(c.X-d.X) - (a.X-b.X)*(c.X*d.Y-c.Y*d.X)
	denomX := (a.X-b.X)*(c.Y-d.Y) - (a.Y-b.Y)*(c.X-d.X)
	numerY := (a.X*b.Y-a.Y*b.X)*(c.Y-d.Y) - (a.Y-b.Y)*(c.X*d.Y-c.Y*d.X)
	denomY := (a.X-b.X)*(c.Y-d.Y) - (a.Y-b.Y)*(c.X-d.X)
	return XY{
		numerX / denomX,
		numerY / denomY,
	}
}

// newNodesFromLineLineIntersection finds the new nodes that would be created
// on a line when it is intersected with another line.
func newNodesFromLineLineIntersection(ln, other line, eps float64) []XY {
	var xys []XY
	if distBetweenXYAndLine(other.a, ln) < eps {
		xys = append(xys, other.a)
	}
	if distBetweenXYAndLine(other.b, ln) < eps {
		xys = append(xys, other.b)
	}
	e := infLineWithLineIntersection(ln, other)
	if ln.envelope().Contains(e) && other.envelope().Contains(e) {
		xys = append(xys, e)
	}
	return xys
}

// newNodesFromLinePointIntersection finds the new nodes that would be created
// on a line when it is intersected with a point.
func newNodesFromLinePointIntersection(ln line, pt XY, eps float64) []XY {
	if distBetweenXYAndLine(pt, ln) < eps {
		return []XY{pt}
	}
	return nil
}

// ulpSizeForLine finds the maximum ULP out of the 4 float64s that make a line.
func ulpSizeForLine(ln line) float64 {
	return math.Max(math.Max(math.Max(
		ulpSize(ln.a.X),
		ulpSize(ln.a.Y)),
		ulpSize(ln.b.X)),
		ulpSize(ln.b.Y))
}

// reNodeGeometries returns the input geometries, but with additional
// intermediate nodes (i.e. control points). The additional nodes are created
// such that when the two geometries are overlaid they only interact at nodes.
func reNodeGeometries(g1, g2 Geometry) (Geometry, Geometry) {
	// Calculate the maximum ULP size over all control points in the input
	// geometries. This size is a good indication of the precision that we
	// should use when node merging.
	var maxULPSize float64
	walk(NewGeometryCollection([]Geometry{g1, g2}).AsGeometry(), func(xy XY) {
		maxULPSize = math.Max(maxULPSize, math.Max(
			ulpSize(math.Abs(xy.X)),
			ulpSize(math.Abs(xy.Y)),
		))
	})

	nodes := newNodeSet(maxULPSize)
	cut := newCutSet(g1, g2)
	// TODO: We may want to insert vertices from both geometries first, since
	// it's first-in-best-dressed for the real vertex per cell grid. It
	// probably makes more sense to have input vertices rather than derived
	// vertices in the output.
	a := reNodeGeometry(g1, cut, nodes)
	b := reNodeGeometry(g2, cut, nodes)
	return a, b
}

// reNodeGeometry re-nodes a single geometry, using a common cut set and node map.
func reNodeGeometry(g Geometry, cut cutSet, nodes nodeSet) Geometry {
	switch g.Type() {
	case TypeGeometryCollection:
		return reNodeGeometryCollection(g.AsGeometryCollection(), cut, nodes).AsGeometry()
	case TypeLineString:
		return reNodeLineString(g.AsLineString(), cut, nodes).AsGeometry()
	case TypePolygon:
		return reNodePolygon(g.AsPolygon(), cut, nodes).AsGeometry()
	case TypeMultiLineString:
		return reNodeMultiLineString(g.AsMultiLineString(), cut, nodes).AsGeometry()
	case TypeMultiPolygon:
		return reNodeMultiPolygonString(g.AsMultiPolygon(), cut, nodes).AsGeometry()
	case TypePoint:
		return reNodeMultiPoint(g.AsPoint().AsMultiPoint(), nodes).AsGeometry()
	case TypeMultiPoint:
		return reNodeMultiPoint(g.AsMultiPoint(), nodes).AsGeometry()
	default:
		panic(fmt.Sprintf("unknown geometry type %v", g.Type()))
	}
}

type cutSet struct {
	lnIndex indexedLines
	ptIndex indexedPoints
}

func newCutSet(g1, g2 Geometry) cutSet {
	lines := appendLines(appendLines(nil, g1), g2)
	points := appendPoints(appendPoints(nil, g1), g2)
	return cutSet{
		lnIndex: newIndexedLines(lines),
		ptIndex: newIndexedPoints(points),
	}
}

func appendLines(lines []line, g Geometry) []line {
	switch g.Type() {
	case TypeLineString:
		seq := g.AsLineString().Coordinates()
		n := seq.Length()
		for i := 0; i < n; i++ {
			ln, ok := getLine(seq, i)
			if ok {
				lines = append(lines, ln)
			}
		}
	case TypeMultiLineString:
		mls := g.AsMultiLineString()
		for i := 0; i < mls.NumLineStrings(); i++ {
			ls := mls.LineStringN(i)
			lines = appendLines(lines, ls.AsGeometry())
		}
	case TypePolygon:
		lines = appendLines(lines, g.AsPolygon().Boundary().AsGeometry())
	case TypeMultiPolygon:
		lines = appendLines(lines, g.AsMultiPolygon().Boundary().AsGeometry())
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			lines = appendLines(lines, gc.GeometryN(i))
		}
	}
	return lines
}

func appendPoints(points []XY, g Geometry) []XY {
	switch g.Type() {
	case TypePoint:
		coords, ok := g.AsPoint().Coordinates()
		if ok {
			points = append(points, coords.XY)
		}
	case TypeMultiPoint:
		mp := g.AsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			points = appendPoints(points, mp.PointN(i).AsGeometry())
		}
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			points = appendPoints(points, gc.GeometryN(i))
		}
	}
	return points
}

func reNodeLineString(ls LineString, cut cutSet, nodes nodeSet) LineString {
	var newCoords []float64
	seq := ls.Coordinates()
	n := seq.Length()
	for lnIdx := 0; lnIdx < n; lnIdx++ {
		ln, ok := getLine(seq, lnIdx)
		if !ok {
			continue
		}

		// Collect cut locations.
		eps := 0xFF * ulpSizeForLine(ln)
		xys := []XY{nodes.insertOrGet(ln.a), nodes.insertOrGet(ln.b)}
		cut.lnIndex.tree.RangeSearch(ln.envelope().box(), func(i int) error {
			other := cut.lnIndex.lines[i]
			newXYs := newNodesFromLineLineIntersection(ln, other, eps)
			for _, xy := range newXYs {
				xys = append(xys, nodes.insertOrGet(xy))
			}
			return nil
		})
		cut.ptIndex.tree.RangeSearch(ln.envelope().box(), func(i int) error {
			other := cut.ptIndex.points[i]
			newXYs := newNodesFromLinePointIntersection(ln, other, eps)
			for _, xy := range newXYs {
				xys = append(xys, nodes.insertOrGet(xy))
			}
			return nil
		})

		// Uniquify and sort.
		xys = sortAndUniquifyXYs(xys) // TODO: make common function
		sortOrigin := nodes.insertOrGet(ln.a)
		sort.Slice(xys, func(i, j int) bool {
			distI := sortOrigin.distanceSquaredTo(xys[i])
			distJ := sortOrigin.distanceSquaredTo(xys[j])
			return distI < distJ
		})

		// Add coords related to this line segment. The end of the previous
		// line is the same as the first point of this line, so we skip it to
		// avoid doubling up.
		if len(newCoords) == 0 {
			newCoords = append(newCoords, xys[0].X, xys[0].Y)
		}
		for _, xy := range xys[1:] {
			newCoords = append(newCoords, xy.X, xy.Y)
		}
	}

	newLS, err := NewLineString(NewSequence(newCoords, DimXY), DisableAllValidations)
	if err != nil {
		panic(fmt.Sprintf("could not re-node LineString: %v", err))
	}
	return newLS
}

func reNodeMultiLineString(mls MultiLineString, cut cutSet, nodes nodeSet) MultiLineString {
	n := mls.NumLineStrings()
	lss := make([]LineString, n)
	for i := 0; i < n; i++ {
		lss[i] = reNodeLineString(mls.LineStringN(i), cut, nodes)
	}
	return NewMultiLineStringFromLineStrings(lss, DisableAllValidations)
}

func reNodePolygon(poly Polygon, cut cutSet, nodes nodeSet) Polygon {
	reNodedBoundary := reNodeMultiLineString(poly.Boundary(), cut, nodes)
	n := reNodedBoundary.NumLineStrings()
	rings := make([]LineString, n)
	for i := 0; i < n; i++ {
		rings[i] = reNodedBoundary.LineStringN(i)
	}
	reNodedPoly, err := NewPolygonFromRings(rings, DisableAllValidations)
	if err != nil {
		panic(err)
	}
	return reNodedPoly
}

func reNodeMultiPolygonString(mp MultiPolygon, cut cutSet, nodes nodeSet) MultiPolygon {
	n := mp.NumPolygons()
	polys := make([]Polygon, n)
	for i := 0; i < n; i++ {
		polys[i] = reNodePolygon(mp.PolygonN(i), cut, nodes)
	}
	reNodedMP, err := NewMultiPolygonFromPolygons(polys, DisableAllValidations)
	if err != nil {
		panic(err)
	}
	return reNodedMP
}

func reNodeMultiPoint(mp MultiPoint, nodes nodeSet) MultiPoint {
	n := mp.NumPoints()
	coords := make([]float64, 0, n*2)
	for i := 0; i < n; i++ {
		xy, ok := mp.PointN(i).XY()
		if ok {
			node := nodes.insertOrGet(xy)
			coords = append(coords, node.X, node.Y)
		}
	}
	return NewMultiPoint(NewSequence(coords, DimXY))
}

func reNodeGeometryCollection(gc GeometryCollection, cut cutSet, nodes nodeSet) GeometryCollection {
	n := gc.NumGeometries()
	geoms := make([]Geometry, n)
	for i := 0; i < n; i++ {
		geoms[i] = reNodeGeometry(gc.GeometryN(i), cut, nodes)
	}
	return NewGeometryCollection(geoms, DisableAllValidations)
}

func newNodeSet(maxULPSize float64) nodeSet {
	// The appropriate multiplication factor to use to calculate bucket size is
	// a bit of a guess.
	bucketSize := maxULPSize * 0xff
	return nodeSet{bucketSize, make(map[nodeBucket]XY)}
}

type nodeSet struct {
	bucketSize float64
	nodes      map[nodeBucket]XY
}

type nodeBucket struct {
	x, y int
}

func (s nodeSet) insertOrGet(xy XY) XY {
	bucket := nodeBucket{
		int(math.Floor(xy.X / s.bucketSize)),
		int(math.Floor(xy.Y / s.bucketSize)),
	}
	xNext := bucket.x + 1
	xPrev := bucket.x - 1
	yNext := bucket.y + 1
	yPrev := bucket.y - 1

	for _, bucket := range [...]nodeBucket{
		bucket, // the original bucket goes first, since it's the most likely entry
		nodeBucket{bucket.x, yNext},
		nodeBucket{bucket.x, yPrev},
		nodeBucket{xPrev, yPrev},
		nodeBucket{xPrev, bucket.y},
		nodeBucket{xPrev, yNext},
		nodeBucket{xNext, yPrev},
		nodeBucket{xNext, bucket.y},
		nodeBucket{xNext, yNext},
	} {
		node, ok := s.nodes[bucket]
		if ok {
			return node
		}
	}
	s.nodes[bucket] = xy
	return xy
}
