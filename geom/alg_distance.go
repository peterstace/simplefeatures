package geom

import (
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/rtree"
)

// Distance calculates the shortest distance (using the Euclidean metric)
// between two geometries. If either geometry is empty, then false is returned
// and the distance is not calculated.
func Distance(g1, g2 Geometry) (float64, bool) {
	// If the geometries intersect with each other, then the distance between
	// them is trivially zero.
	if Intersects(g1, g2) {
		return 0, true
	}

	// The general approach of the distance algorithm is as follows:
	//
	// 1. Convert both geometries to lists of points and line segments.
	//
	// 2. Index the second geometry using an RTree.
	//
	// 3. Iterate over every part (point or line) in the first geometry. For
	//    each part, search in the RTree for the nearest part of the second
	//    geometry. We can stop searching if the bounding box in the RTree is
	//    further away than the best distance so far.

	xys1, lns1 := extractXYsAndLines(g1)
	xys2, lns2 := extractXYsAndLines(g2)

	// Swap order so that the larger geometry goes into the RTree.
	if len(xys1)+len(lns1) > len(xys2)+len(lns2) {
		xys1, xys2 = xys2, xys1
		lns1, lns2 = lns2, lns1
	}

	xyTree := loadXYTree(xys2)
	lnTree := loadLineTree(lns2)
	minDist := math.Inf(+1)

	for _, xy1 := range xys1 {
		xy1Env := xy1.uncheckedEnvelope()
		_ = xyTree.PrioritySearch(xy1.box(), func(xy2 XY) error {
			if d, ok := xy2.uncheckedEnvelope().Distance(xy1Env); ok && d > minDist {
				return rtree.Stop
			}
			minDist = fastMin(minDist, distBetweenXYs(xy1, xy2))
			return nil
		})
		_ = lnTree.PrioritySearch(xy1.box(), func(ln2 line) error {
			if d, ok := ln2.uncheckedEnvelope().Distance(xy1Env); ok && d > minDist {
				return rtree.Stop
			}
			minDist = fastMin(minDist, distBetweenXYAndLine(xy1, ln2))
			return nil
		})
	}
	for _, ln1 := range lns1 {
		ln1Env := ln1.uncheckedEnvelope()
		_ = xyTree.PrioritySearch(ln1.box(), func(xy2 XY) error {
			if d, ok := xy2.uncheckedEnvelope().Distance(ln1Env); ok && d > minDist {
				return rtree.Stop
			}
			minDist = fastMin(minDist, distBetweenXYAndLine(xy2, ln1))
			return nil
		})
		_ = lnTree.PrioritySearch(ln1.box(), func(ln2 line) error {
			if d, ok := ln2.uncheckedEnvelope().Distance(ln1Env); ok && d > minDist {
				return rtree.Stop
			}
			minDist = fastMin(minDist, distBetweenLineAndLine(ln1, ln2))
			return nil
		})
	}

	if math.IsInf(minDist, +1) {
		return 0, false
	}
	return minDist, true
}

func extractXYsAndLines(g Geometry) ([]XY, []line) {
	switch g.Type() {
	case TypePoint:
		return g.MustAsPoint().asXYs(), nil
	case TypeLineString:
		return nil, g.MustAsLineString().asLines()
	case TypePolygon:
		return nil, g.MustAsPolygon().Boundary().asLines()
	case TypeMultiPoint:
		return g.MustAsMultiPoint().asXYs(), nil
	case TypeMultiLineString:
		return nil, g.MustAsMultiLineString().asLines()
	case TypeMultiPolygon:
		return nil, g.MustAsMultiPolygon().Boundary().asLines()
	case TypeGeometryCollection:
		var allXYs []XY
		var allLines []line
		g.MustAsGeometryCollection().walk(func(child Geometry) {
			xys, lns := extractXYsAndLines(child)
			allXYs = append(allXYs, xys...)
			allLines = append(allLines, lns...)
		})
		return allXYs, allLines
	default:
		panic(fmt.Sprintf("implementation error: unhandled geometry types %s", g.Type()))
	}
}

func loadXYTree(xys []XY) *rtree.RTree[XY] {
	items := make([]rtree.BulkItem[XY], len(xys))
	for i, xy := range xys {
		items[i] = rtree.BulkItem[XY]{
			Box:    xy.box(),
			Record: xy,
		}
	}
	return rtree.BulkLoad(items)
}

func loadLineTree(lns []line) *rtree.RTree[line] {
	items := make([]rtree.BulkItem[line], len(lns))
	for i, ln := range lns {
		items[i] = rtree.BulkItem[line]{
			Box:    ln.box(),
			Record: ln,
		}
	}
	return rtree.BulkLoad(items)
}

func distBetweenXYs(xy1, xy2 XY) float64 {
	return xy1.Sub(xy2).Length()
}

func distBetweenXYAndLine(xy XY, ln line) float64 {
	ab := ln.b.Sub(ln.a)
	abLen := ab.Length()
	proj := xy.Sub(ln.a).Dot(ab) / abLen
	var closest XY
	switch {
	case proj < 0:
		closest = ln.a
	case proj > abLen:
		closest = ln.b
	default:
		scaled := ab.Scale(proj / abLen)
		closest = scaled.Add(ln.a)
	}
	return distBetweenXYs(xy, closest)
}

func distBetweenLineAndLine(ln1, ln2 line) float64 {
	minDist := math.Inf(+1)
	for _, dist := range [4]float64{
		distBetweenXYAndLine(ln1.a, ln2),
		distBetweenXYAndLine(ln1.b, ln2),
		distBetweenXYAndLine(ln2.a, ln1),
		distBetweenXYAndLine(ln2.b, ln1),
	} {
		minDist = fastMin(minDist, dist)
	}
	return minDist
}
