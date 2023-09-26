package geom

import (
	"fmt"
	"math"
	"sort"
)

// appendNewNodesFromLineLineIntersection finds the new nodes that would be
// created on a line when it is intersected with another line.
func appendNewNodesFromLineLineIntersection(dst []XY, ln, other line, eps float64, nodes nodeSet) []XY {
	if !ln.hasEndpoint(other.a) && distBetweenXYAndLine(other.a, ln) < eps {
		dst = appendNewNode(dst, nodes, ln, other.a)
	}
	if !ln.hasEndpoint(other.b) && distBetweenXYAndLine(other.b, ln) < eps {
		dst = appendNewNode(dst, nodes, ln, other.b)
	}
	inter := ln.intersectLine(other)
	if !inter.empty {
		if !ln.hasEndpoint(inter.ptA) {
			dst = appendNewNode(dst, nodes, ln, inter.ptA)
		}
		if inter.ptA != inter.ptB && !ln.hasEndpoint(inter.ptB) {
			dst = appendNewNode(dst, nodes, ln, inter.ptB)
		}
	}
	return dst
}

// appendNewNodesFromLinePointIntersection finds the new nodes that would be
// created on a line when it is intersected with a point.
func appendNewNodesFromLinePointIntersection(dst []XY, ln line, pt XY, eps float64, nodes nodeSet) []XY {
	if !ln.hasEndpoint(pt) && distBetweenXYAndLine(pt, ln) < eps {
		dst = appendNewNode(dst, nodes, ln, pt)
	}
	return dst
}

// appendNewNode appends xy to dst (and returns dst) after creating it as a
// node. But it only does so if the node is *not* already an endpoint of ln
// (since those nodes already exist).
func appendNewNode(dst []XY, nodes nodeSet, ln line, xy XY) []XY {
	xy = nodes.insertOrGet(xy)
	if xy == ln.a || xy == ln.b {
		return dst
	}
	return append(dst, xy)
}

// ulpSizeForLine finds the maximum ULP out of the 4 float64s that make a line.
func ulpSizeForLine(ln line) float64 {
	return fastMax(fastMax(fastMax(
		ulpSize(ln.a.X),
		ulpSize(ln.a.Y)),
		ulpSize(ln.b.X)),
		ulpSize(ln.b.Y))
}

// reNodeGeometries returns the input geometries, but with additional
// intermediate nodes (i.e. control points). The additional nodes are created
// such that when the two geometries are overlaid the only interactions
// (including self-interactions) between geometries are at nodes. Nodes that
// are close to each other are also snapped together.
func reNodeGeometries(g1, g2 Geometry, mls MultiLineString) (Geometry, Geometry, MultiLineString) {
	// Calculate the maximum ULP size over all control points in the input
	// geometries. This size is a good indication of the precision that we
	// should use when node merging.
	var maxULPSize float64
	all := NewGeometryCollection([]Geometry{g1, g2, mls.AsGeometry()}).AsGeometry()
	var xyCount int
	walk(all, func(xy XY) {
		xyCount++
		maxULPSize = fastMax(maxULPSize, fastMax(
			ulpSize(math.Abs(xy.X)),
			ulpSize(math.Abs(xy.Y)),
		))
	})

	// Snap vertices together if they are very close.
	nodes := newNodeSet(maxULPSize, xyCount)
	g1 = g1.TransformXY(nodes.insertOrGet)
	g2 = g2.TransformXY(nodes.insertOrGet)
	mls = mls.TransformXY(nodes.insertOrGet)

	// Create additional nodes for crossings.
	cut := newCutSet(all)
	g1 = reNodeGeometry(g1, cut, nodes)
	g2 = reNodeGeometry(g2, cut, nodes)
	mls = reNodeMultiLineString(mls, cut, nodes)
	return g1, g2, mls
}

// reNodeGeometry re-nodes a single geometry, using a common cut set and node
// map. The cut set is already noded.
func reNodeGeometry(g Geometry, cut cutSet, nodes nodeSet) Geometry {
	switch g.Type() {
	case TypeGeometryCollection:
		return reNodeGeometryCollection(g.MustAsGeometryCollection(), cut, nodes).AsGeometry()
	case TypeLineString:
		return reNodeLineString(g.MustAsLineString(), cut, nodes).AsGeometry()
	case TypePolygon:
		return reNodePolygon(g.MustAsPolygon(), cut, nodes).AsGeometry()
	case TypeMultiLineString:
		return reNodeMultiLineString(g.MustAsMultiLineString(), cut, nodes).AsGeometry()
	case TypeMultiPolygon:
		return reNodeMultiPolygonString(g.MustAsMultiPolygon(), cut, nodes).AsGeometry()
	case TypePoint, TypeMultiPoint:
		return g
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
		seq := g.MustAsLineString().Coordinates()
		n := seq.Length()
		for i := 0; i < n; i++ {
			ln, ok := getLine(seq, i)
			if ok {
				lines = append(lines, ln)
			}
		}
	case TypeMultiLineString:
		mls := g.MustAsMultiLineString()
		for i := 0; i < mls.NumLineStrings(); i++ {
			ls := mls.LineStringN(i)
			lines = appendLines(lines, ls.AsGeometry())
		}
	case TypePolygon:
		lines = appendLines(lines, g.MustAsPolygon().Boundary().AsGeometry())
	case TypeMultiPolygon:
		lines = appendLines(lines, g.MustAsMultiPolygon().Boundary().AsGeometry())
	case TypeGeometryCollection:
		gc := g.MustAsGeometryCollection()
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
		coords, ok := g.MustAsPoint().Coordinates()
		if ok {
			points = append(points, coords.XY)
		}
	case TypeMultiPoint:
		mp := g.MustAsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			points = appendPoints(points, mp.PointN(i).AsGeometry())
		}
	case TypeGeometryCollection:
		gc := g.MustAsGeometryCollection()
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

		// Copy over first point of line. We don't copy the final point of the
		// LineString until the end.
		newCoords = append(newCoords, ln.a.X, ln.a.Y)

		// Collect cut locations that are *interior* to ln.
		eps := 0xFF * ulpSizeForLine(ln)
		var xys []XY
		cut.lnIndex.tree.RangeSearch(ln.box(), func(i int) error {
			other := cut.lnIndex.lines[i]
			xys = appendNewNodesFromLineLineIntersection(xys, ln, other, eps, nodes)
			return nil
		})
		cut.ptIndex.tree.RangeSearch(ln.box(), func(i int) error {
			other := cut.ptIndex.points[i]
			xys = appendNewNodesFromLinePointIntersection(xys, ln, other, eps, nodes)
			return nil
		})

		// Uniquify and sort cut locations.
		xys = sortAndUniquifyXYs(xys)
		sort.Slice(xys, func(i, j int) bool {
			distI := ln.a.distanceSquaredTo(xys[i])
			distJ := ln.a.distanceSquaredTo(xys[j])
			return distI < distJ
		})

		// Copy cut locations into output.
		for _, xy := range xys {
			newCoords = append(newCoords, xy.X, xy.Y)
		}
	}

	// Copy over final point.
	if n > 0 {
		last := seq.GetXY(n - 1)
		newCoords = append(newCoords, last.X, last.Y)
	}

	return NewLineString(NewSequence(newCoords, DimXY))
}

func reNodeMultiLineString(mls MultiLineString, cut cutSet, nodes nodeSet) MultiLineString {
	n := mls.NumLineStrings()
	lss := make([]LineString, n)
	for i := 0; i < n; i++ {
		lss[i] = reNodeLineString(mls.LineStringN(i), cut, nodes)
	}
	return NewMultiLineString(lss)
}

func reNodePolygon(poly Polygon, cut cutSet, nodes nodeSet) Polygon {
	reNodedBoundary := reNodeMultiLineString(poly.Boundary(), cut, nodes)
	n := reNodedBoundary.NumLineStrings()
	rings := make([]LineString, n)
	for i := 0; i < n; i++ {
		rings[i] = reNodedBoundary.LineStringN(i)
	}
	return NewPolygon(rings)
}

func reNodeMultiPolygonString(mp MultiPolygon, cut cutSet, nodes nodeSet) MultiPolygon {
	n := mp.NumPolygons()
	polys := make([]Polygon, n)
	for i := 0; i < n; i++ {
		polys[i] = reNodePolygon(mp.PolygonN(i), cut, nodes)
	}
	return NewMultiPolygon(polys)
}

func reNodeGeometryCollection(gc GeometryCollection, cut cutSet, nodes nodeSet) GeometryCollection {
	n := gc.NumGeometries()
	geoms := make([]Geometry, n)
	for i := 0; i < n; i++ {
		geoms[i] = reNodeGeometry(gc.GeometryN(i), cut, nodes)
	}
	return NewGeometryCollection(geoms)
}
