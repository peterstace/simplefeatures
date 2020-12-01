package geom

import (
	"fmt"
	"math"
	"sort"
)

// appendNewNodesFromLineLineIntersection finds the new nodes that would be
// created on a line when it is intersected with another line.
func appendNewNodesFromLineLineIntersection(dst []XY, ln, other line, eps float64) []XY {
	if distBetweenXYAndLine(other.a, ln) < eps {
		dst = append(dst, other.a)
	}
	if distBetweenXYAndLine(other.b, ln) < eps {
		dst = append(dst, other.b)
	}
	inter := ln.intersectLine(other)
	if !inter.empty {
		dst = append(dst, inter.ptA, inter.ptB)
	}
	return dst
}

// newNodeFromLinePointIntersection finds the new node that might be created on
// a line when it is intersected with a point.
func newNodeFromLinePointIntersection(ln line, pt XY, eps float64) (XY, bool) {
	return pt, distBetweenXYAndLine(pt, ln) < eps
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
func reNodeGeometries(g1, g2 Geometry, mls MultiLineString) (Geometry, Geometry, MultiLineString, error) {
	// Calculate the maximum ULP size over all control points in the input
	// geometries. This size is a good indication of the precision that we
	// should use when node merging.
	var maxULPSize float64
	all := NewGeometryCollection([]Geometry{g1, g2, mls.AsGeometry()}).AsGeometry()
	walk(all, func(xy XY) {
		maxULPSize = math.Max(maxULPSize, math.Max(
			ulpSize(math.Abs(xy.X)),
			ulpSize(math.Abs(xy.Y)),
		))
	})

	nodes := newNodeSet(maxULPSize)
	cut := newCutSet(all)
	walk(all, func(xy XY) {
		nodes.insertOrGet(xy)
	})

	a, err := reNodeGeometry(g1, cut, nodes)
	if err != nil {
		return Geometry{}, Geometry{}, MultiLineString{}, err
	}
	b, err := reNodeGeometry(g2, cut, nodes)
	if err != nil {
		return Geometry{}, Geometry{}, MultiLineString{}, err
	}
	c, err := reNodeMultiLineString(mls, cut, nodes)
	if err != nil {
		return Geometry{}, Geometry{}, MultiLineString{}, err
	}
	return a, b, c, nil
}

// reNodeGeometry re-nodes a single geometry, using a common cut set and node map.
func reNodeGeometry(g Geometry, cut cutSet, nodes nodeSet) (Geometry, error) {
	switch g.Type() {
	case TypeGeometryCollection:
		gc, err := reNodeGeometryCollection(g.AsGeometryCollection(), cut, nodes)
		return gc.AsGeometry(), err
	case TypeLineString:
		ls, err := reNodeLineString(g.AsLineString(), cut, nodes)
		return ls.AsGeometry(), err
	case TypePolygon:
		poly, err := reNodePolygon(g.AsPolygon(), cut, nodes)
		return poly.AsGeometry(), err
	case TypeMultiLineString:
		mls, err := reNodeMultiLineString(g.AsMultiLineString(), cut, nodes)
		return mls.AsGeometry(), err
	case TypeMultiPolygon:
		mp, err := reNodeMultiPolygonString(g.AsMultiPolygon(), cut, nodes)
		return mp.AsGeometry(), err
	case TypePoint:
		return reNodeMultiPoint(g.AsPoint().AsMultiPoint(), nodes).AsGeometry(), nil
	case TypeMultiPoint:
		return reNodeMultiPoint(g.AsMultiPoint(), nodes).AsGeometry(), nil
	default:
		panic(fmt.Sprintf("unknown geometry type %v", g.Type()))
	}
}

type cutSet struct {
	lnIndex indexedLines
	ptIndex indexedPoints
}

func newCutSet(g Geometry) cutSet {
	lines := appendLines(nil, g)
	points := appendPoints(nil, g)
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

func reNodeLineString(ls LineString, cut cutSet, nodes nodeSet) (LineString, error) {
	var (
		tmp       []XY
		newCoords []float64
	)
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
			tmp = appendNewNodesFromLineLineIntersection(tmp[:0], ln, other, eps)
			for _, xy := range tmp {
				xys = append(xys, nodes.insertOrGet(xy))
			}
			return nil
		})
		cut.ptIndex.tree.RangeSearch(ln.envelope().box(), func(i int) error {
			other := cut.ptIndex.points[i]
			if xy, ok := newNodeFromLinePointIntersection(ln, other, eps); ok {
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
		return LineString{}, err
	}
	return newLS, nil
}

func reNodeMultiLineString(mls MultiLineString, cut cutSet, nodes nodeSet) (MultiLineString, error) {
	n := mls.NumLineStrings()
	lss := make([]LineString, n)
	for i := 0; i < n; i++ {
		var err error
		lss[i], err = reNodeLineString(mls.LineStringN(i), cut, nodes)
		if err != nil {
			return MultiLineString{}, err
		}
	}
	return NewMultiLineStringFromLineStrings(lss, DisableAllValidations), nil
}

func reNodePolygon(poly Polygon, cut cutSet, nodes nodeSet) (Polygon, error) {
	reNodedBoundary, err := reNodeMultiLineString(poly.Boundary(), cut, nodes)
	if err != nil {
		return Polygon{}, err
	}
	n := reNodedBoundary.NumLineStrings()
	rings := make([]LineString, n)
	for i := 0; i < n; i++ {
		rings[i] = reNodedBoundary.LineStringN(i)
	}
	reNodedPoly, err := NewPolygonFromRings(rings, DisableAllValidations)
	if err != nil {
		return Polygon{}, err
	}
	return reNodedPoly, nil
}

func reNodeMultiPolygonString(mp MultiPolygon, cut cutSet, nodes nodeSet) (MultiPolygon, error) {
	n := mp.NumPolygons()
	polys := make([]Polygon, n)
	for i := 0; i < n; i++ {
		var err error
		polys[i], err = reNodePolygon(mp.PolygonN(i), cut, nodes)
		if err != nil {
			return MultiPolygon{}, err
		}
	}
	reNodedMP, err := NewMultiPolygonFromPolygons(polys, DisableAllValidations)
	if err != nil {
		return MultiPolygon{}, err
	}
	return reNodedMP, nil
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

func reNodeGeometryCollection(gc GeometryCollection, cut cutSet, nodes nodeSet) (GeometryCollection, error) {
	n := gc.NumGeometries()
	geoms := make([]Geometry, n)
	for i := 0; i < n; i++ {
		var err error
		geoms[i], err = reNodeGeometry(gc.GeometryN(i), cut, nodes)
		if err != nil {
			return GeometryCollection{}, err
		}
	}
	return NewGeometryCollection(geoms, DisableAllValidations), nil
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
