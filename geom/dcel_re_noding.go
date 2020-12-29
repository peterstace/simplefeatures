package geom

import (
	"fmt"
	"math"
	"sort"
)

// appendNewNodesFromLineLineIntersection finds the new nodes that would be
// created on a line when it is intersected with another line.
func appendNewNodesFromLineLineIntersection(dst []XY, ln, other line, eps float64, nodes nodeSet) []XY {
	if distBetweenXYAndLine(other.a, ln) < eps {
		dst = appendNewNode(dst, nodes, ln, other.a)
	}
	if distBetweenXYAndLine(other.b, ln) < eps {
		dst = appendNewNode(dst, nodes, ln, other.b)
	}
	inter := ln.intersectLine(other)
	if !inter.empty {
		dst = appendNewNode(dst, nodes, ln, inter.ptA)
		if inter.ptA != inter.ptB {
			dst = appendNewNode(dst, nodes, ln, inter.ptB)
		}
	}
	return dst
}

// appendNewNodesFromLinePointIntersection finds the new nodes that would be
// created on a line when it is intersected with a point.
func appendNewNodesFromLinePointIntersection(dst []XY, ln line, pt XY, eps float64, nodes nodeSet) []XY {
	if distBetweenXYAndLine(pt, ln) < eps {
		dst = appendNewNode(dst, nodes, ln, pt)
	}
	return dst
}

// appendNewNode appends xy to dst (and returns dst) after creating it as a
// node. But it only does so if the node is *not* already an endpoint of ln
// (since those nodes already exist).
func appendNewNode(dst []XY, nodes nodeSet, ln line, xy XY) []XY {
	if xy == ln.a || xy == ln.b {
		return dst
	}
	xy = nodes.insertOrGet(xy)
	if xy == ln.a || xy == ln.b {
		return dst
	}
	return append(dst, xy)
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
// such that when the two geometries are overlaid the only interactions
// (including self-interactions) between geometries are at nodes.
func reNodeGeometries(g1, g2 Geometry, mls MultiLineString) (Geometry, Geometry, MultiLineString, error) {
	// Calculate the maximum ULP size over all control points in the input
	// geometries. This size is a good indication of the precision that we
	// should use when node merging.
	var maxULPSize float64
	all := NewGeometryCollection([]Geometry{g1, g2, mls.AsGeometry()}).AsGeometry()
	var xyCount int
	walk(all, func(xy XY) {
		xyCount++
		maxULPSize = math.Max(maxULPSize, math.Max(
			ulpSize(math.Abs(xy.X)),
			ulpSize(math.Abs(xy.Y)),
		))
	})

	nodes := newNodeSet(maxULPSize, xyCount)
	cut := newCutSet(all, nodes)

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

// reNodeGeometry re-nodes a single geometry, using a common cut set and node
// map. The cut set is already noded.
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

// cutSet is an indexed set of lines and points from all input geometries
// (including ghosts). It is used to "cut" (i.e. split lines into multiple
// lines) geometries so that interactions only occur at nodes.
type cutSet struct {
	lnIndex indexedLines
	ptIndex indexedPoints
}

func newCutSet(g Geometry, nodes nodeSet) cutSet {
	lines := appendLines(nil, g, nodes)
	points := appendPoints(nil, g, nodes)
	return cutSet{
		lnIndex: newIndexedLines(lines),
		ptIndex: newIndexedPoints(points),
	}
}

func appendLines(lines []line, g Geometry, nodes nodeSet) []line {
	switch g.Type() {
	case TypeLineString:
		seq := g.AsLineString().Coordinates()
		n := seq.Length()
		for i := 0; i < n; i++ {
			ln, ok := getLine(seq, i)
			if ok {
				ln.a = nodes.insertOrGet(ln.a)
				ln.b = nodes.insertOrGet(ln.b)
				if ln.a != ln.b {
					lines = append(lines, ln)
				}
			}
		}
	case TypeMultiLineString:
		mls := g.AsMultiLineString()
		for i := 0; i < mls.NumLineStrings(); i++ {
			ls := mls.LineStringN(i)
			lines = appendLines(lines, ls.AsGeometry(), nodes)
		}
	case TypePolygon:
		lines = appendLines(lines, g.AsPolygon().Boundary().AsGeometry(), nodes)
	case TypeMultiPolygon:
		lines = appendLines(lines, g.AsMultiPolygon().Boundary().AsGeometry(), nodes)
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			lines = appendLines(lines, gc.GeometryN(i), nodes)
		}
	}
	return lines
}

func appendPoints(points []XY, g Geometry, nodes nodeSet) []XY {
	switch g.Type() {
	case TypePoint:
		coords, ok := g.AsPoint().Coordinates()
		if ok {
			points = append(points, nodes.insertOrGet(coords.XY))
		}
	case TypeMultiPoint:
		mp := g.AsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			points = appendPoints(points, mp.PointN(i).AsGeometry(), nodes)
		}
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			points = appendPoints(points, gc.GeometryN(i), nodes)
		}
	}
	return points
}

func reNodeLineString(ls LineString, cut cutSet, nodes nodeSet) (LineString, error) {
	var newCoords []float64
	seq := ls.Coordinates()
	n := seq.Length()
	for lnIdx := 0; lnIdx < n; lnIdx++ {
		ln, ok := getLine(seq, lnIdx)
		if !ok {
			continue
		}
		ln.a = nodes.insertOrGet(ln.a)
		ln.b = nodes.insertOrGet(ln.b)
		if ln.a == ln.b {
			continue
		}

		// Copy over first point of line. We don't copy the final point of the
		// LineString until the end.
		newCoords = append(newCoords, ln.a.X, ln.a.Y)

		// Collect cut locations that are *interior* to ln.
		eps := 0xFF * ulpSizeForLine(ln)
		var xys []XY
		cut.lnIndex.tree.RangeSearch(ln.envelope().box(), func(i int) error {
			other := cut.lnIndex.lines[i]
			xys = appendNewNodesFromLineLineIntersection(xys, ln, other, eps, nodes)
			return nil
		})
		cut.ptIndex.tree.RangeSearch(ln.envelope().box(), func(i int) error {
			other := cut.ptIndex.points[i]
			xys = appendNewNodesFromLinePointIntersection(xys, ln, other, eps, nodes)
			return nil
		})

		// Uniquify and sort cut locations.
		xys = sortAndUniquifyXYs(xys)
		sortOrigin := nodes.insertOrGet(ln.a)
		sort.Slice(xys, func(i, j int) bool {
			distI := sortOrigin.distanceSquaredTo(xys[i])
			distJ := sortOrigin.distanceSquaredTo(xys[j])
			return distI < distJ
		})

		// Copy cut locations into output.
		for _, xy := range xys {
			newCoords = append(newCoords, xy.X, xy.Y)
		}
	}

	// Copy over final point.
	if n > 0 {
		last := nodes.insertOrGet(seq.GetXY(n - 1))
		newCoords = append(newCoords, last.X, last.Y)
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

func newNodeSet(maxULPSize float64, sizeHint int) nodeSet {
	// The appropriate multiplication factor to use to calculate bucket size is
	// a bit of a guess.
	bucketSize := maxULPSize * 0xff
	return nodeSet{bucketSize, make(map[nodeBucket]XY, sizeHint)}
}

// nodeSet is a set of XY values (nodes). If an XY value is inserted, but it is
// "close" to an existing XY in the set, then the original XY is returned (and
// the new XY _not_ inserted). The two XYs essentially merge together.
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
