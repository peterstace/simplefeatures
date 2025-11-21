package geom

import (
	"fmt"
	"math"
	"sort"
)

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

// reNodeGeometries returns the input geometries, but with additional
// intermediate nodes (i.e. control points). The additional nodes are created
// such that when the two geometries are overlaid the only interactions
// (including self-interactions) between geometries are at nodes. Nodes that
// are close to each other are also snapped together.
func reNodeGeometries(g1, g2 Geometry, mls MultiLineString) (Geometry, Geometry, MultiLineString) {
	// Calculate the maximum ULP size over all control points in the input
	// geometries. This size is a good indication of the precision that we
	// should use when node merging.
	var ulp float64
	var xyCount int
	all := func() Geometry {
		return NewGeometryCollection([]Geometry{g1, g2, mls.AsGeometry()}).AsGeometry()
	}
	walk(all(), func(xy XY) {
		xyCount++
		ulp = fastMax(ulp, fastMax(
			ulpSize(math.Abs(xy.X)),
			ulpSize(math.Abs(xy.Y)),
		))
	})
	nodes := newNodeSet(ulp, xyCount)

	// Snap vertices together if they are very close.
	g1 = g1.TransformXY(nodes.insertOrGet)
	g2 = g2.TransformXY(nodes.insertOrGet)
	mls = mls.TransformXY(nodes.insertOrGet)

	// Create new nodes for point/line intersections.
	ptIndex := newIndexedPoints(nodes.list())
	appendCutsForPointXLine := func(ln line, cuts []XY) []XY {
		ptIndex.tree.RangeSearch(ln.box(), func(i int) error {
			xy := ptIndex.points[i]
			if !ln.hasEndpoint(xy) && distBetweenXYAndLine(xy, ln) < ulp*0x200 {
				cuts = append(cuts, xy)
			}
			return nil
		})
		return cuts
	}
	g1 = reNodeGeometry(g1, appendCutsForPointXLine)
	g2 = reNodeGeometry(g2, appendCutsForPointXLine)
	mls = reNodeMultiLineString(mls, appendCutsForPointXLine)

	// Create new nodes for line/line intersections.
	lnIndex := newIndexedLines(appendLines(nil, all()))
	appendCutsLineXLine := func(ln line, cuts []XY) []XY {
		lnIndex.tree.RangeSearch(ln.box(), func(i int) error {
			other := lnIndex.lines[i]

			// TODO: This is a hacky approach (re-orders inputs, rather than
			// making the operation truly symmetric). Instead, it would be
			// better to use "solution 2" described in
			// https://github.com/peterstace/simplefeatures/issues/574.
			inter := symmetricLineIntersection(ln, other)

			if !inter.empty {
				if !ln.hasEndpoint(inter.ptA) {
					cuts = appendNewNode(cuts, nodes, ln, inter.ptA)
				}
				if inter.ptA != inter.ptB && !ln.hasEndpoint(inter.ptB) {
					cuts = appendNewNode(cuts, nodes, ln, inter.ptB)
				}
			}
			return nil
		})
		return cuts
	}

	g1 = reNodeGeometry(g1, appendCutsLineXLine)
	g2 = reNodeGeometry(g2, appendCutsLineXLine)
	mls = reNodeMultiLineString(mls, appendCutsLineXLine)

	return g1, g2, mls
}

func reNodeGeometry(g Geometry, appendCuts func(line, []XY) []XY) Geometry {
	switch g.Type() {
	case TypeGeometryCollection:
		return reNodeGeometryCollection(g.MustAsGeometryCollection(), appendCuts).AsGeometry()
	case TypeLineString:
		return reNodeLineString(g.MustAsLineString(), appendCuts).AsGeometry()
	case TypePolygon:
		return reNodePolygon(g.MustAsPolygon(), appendCuts).AsGeometry()
	case TypeMultiLineString:
		return reNodeMultiLineString(g.MustAsMultiLineString(), appendCuts).AsGeometry()
	case TypeMultiPolygon:
		return reNodeMultiPolygon(g.MustAsMultiPolygon(), appendCuts).AsGeometry()
	case TypePoint, TypeMultiPoint:
		return g
	default:
		panic(fmt.Sprintf("unknown geometry type %v", g.Type()))
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

func reNodeLineString(ls LineString, appendCuts func(line, []XY) []XY) LineString {
	var newCoords []float64
	var cuts []XY
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

		// Collect cut locations.
		cuts = cuts[:0]
		cuts = appendCuts(ln, cuts)
		sort.Slice(cuts, func(i, j int) bool {
			distI := ln.a.distanceSquaredTo(cuts[i])
			distJ := ln.a.distanceSquaredTo(cuts[j])
			return distI < distJ
		})
		cuts = uniquifyGroupedXYs(cuts)

		// Copy cut locations into output.
		for _, xy := range cuts {
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

func reNodeMultiLineString(mls MultiLineString, appendCuts func(line, []XY) []XY) MultiLineString {
	n := mls.NumLineStrings()
	lss := make([]LineString, n)
	for i := 0; i < n; i++ {
		lss[i] = reNodeLineString(mls.LineStringN(i), appendCuts)
	}
	return NewMultiLineString(lss)
}

func reNodePolygon(poly Polygon, appendCuts func(line, []XY) []XY) Polygon {
	reNodedBoundary := reNodeMultiLineString(poly.Boundary(), appendCuts)
	n := reNodedBoundary.NumLineStrings()
	rings := make([]LineString, n)
	for i := 0; i < n; i++ {
		rings[i] = reNodedBoundary.LineStringN(i)
	}
	return NewPolygon(rings)
}

func reNodeMultiPolygon(mp MultiPolygon, appendCuts func(line, []XY) []XY) MultiPolygon {
	n := mp.NumPolygons()
	polys := make([]Polygon, n)
	for i := 0; i < n; i++ {
		polys[i] = reNodePolygon(mp.PolygonN(i), appendCuts)
	}
	return NewMultiPolygon(polys)
}

func reNodeGeometryCollection(gc GeometryCollection, appendCuts func(line, []XY) []XY) GeometryCollection {
	n := gc.NumGeometries()
	geoms := make([]Geometry, n)
	for i := 0; i < n; i++ {
		geoms[i] = reNodeGeometry(gc.GeometryN(i), appendCuts)
	}
	return NewGeometryCollection(geoms)
}
