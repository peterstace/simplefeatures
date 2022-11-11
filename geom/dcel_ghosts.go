package geom

import (
	"fmt"

	"github.com/peterstace/simplefeatures/rtree"
)

// createGhosts creates a MultiLineString that connects all components of the
// input Geometries.
func createGhosts(a, b Geometry) MultiLineString {
	var points []XY
	points = appendComponentPoints(points, a)
	points = appendComponentPoints(points, b)
	ghosts := spanningTree(points)
	return ghosts
}

// spanningTree creates a near-minimum spanning tree (using the euclidean
// distance metric) over the supplied points. The tree will consist of N-1
// lines, where N is the number of _distinct_ xys supplied.
//
// It's a 'near' minimum spanning tree rather than a spanning tree, because we
// use a simple greedy algorithm rather than a proper minimum spanning tree
// algorithm.
func spanningTree(xys []XY) MultiLineString {
	if len(xys) <= 1 {
		return MultiLineString{}
	}

	// Load points into r-tree.
	xys = sortAndUniquifyXYs(xys)
	items := make([]rtree.BulkItem, len(xys))
	for i, xy := range xys {
		items[i] = rtree.BulkItem{Box: xy.box(), RecordID: i}
	}
	tree := rtree.BulkLoad(items)

	// The disjoint set keeps track of which points have been joined together
	// so far. Two entries in dset are in the same set iff they are connected
	// in the incrementally-built spanning tree.
	dset := newDisjointSet(len(xys))
	lss := make([]LineString, 0, len(xys)-1)

	for i, xyi := range xys {
		if i == len(xys)-1 {
			// Skip the last point, since a tree is formed from N-1 edges
			// rather than N edges. The last point will be included by virtue
			// of being the closest to another point.
			continue
		}
		tree.PrioritySearch(xyi.box(), func(j int) error {
			// We don't want to include a new edge in the spanning tree if it
			// would cause a cycle (i.e. the two endpoints are already in the
			// same tree). This is checked via dset.
			if i == j || dset.find(i) == dset.find(j) {
				return nil
			}
			dset.union(i, j)
			xyj := xys[j]
			lss = append(lss, line{xyi, xyj}.asLineString())
			return rtree.Stop
		})
	}

	return NewMultiLineString(lss)
}

func appendXYForPoint(xys []XY, pt Point) []XY {
	if xy, ok := pt.XY(); ok {
		xys = append(xys, xy)
	}
	return xys
}

func appendXYForLineString(xys []XY, ls LineString) []XY {
	return appendXYForPoint(xys, ls.StartPoint())
}

func appendXYsForPolygon(xys []XY, poly Polygon) []XY {
	xys = appendXYForLineString(xys, poly.ExteriorRing())
	n := poly.NumInteriorRings()
	for i := 0; i < n; i++ {
		xys = appendXYForLineString(xys, poly.InteriorRingN(i))
	}
	return xys
}

func appendComponentPoints(xys []XY, g Geometry) []XY {
	switch g.Type() {
	case TypePoint:
		return appendXYForPoint(xys, g.MustAsPoint())
	case TypeMultiPoint:
		mp := g.MustAsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			xys = appendXYForPoint(xys, mp.PointN(i))
		}
		return xys
	case TypeLineString:
		ls := g.MustAsLineString()
		return appendXYForLineString(xys, ls)
	case TypeMultiLineString:
		mls := g.MustAsMultiLineString()
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			ls := mls.LineStringN(i)
			xys = appendXYForLineString(xys, ls)
		}
		return xys
	case TypePolygon:
		poly := g.MustAsPolygon()
		return appendXYsForPolygon(xys, poly)
	case TypeMultiPolygon:
		mp := g.MustAsMultiPolygon()
		n := mp.NumPolygons()
		for i := 0; i < n; i++ {
			poly := mp.PolygonN(i)
			xys = appendXYsForPolygon(xys, poly)
		}
		return xys
	case TypeGeometryCollection:
		gc := g.MustAsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			xys = appendComponentPoints(xys, gc.GeometryN(i))
		}
		return xys
	default:
		panic(fmt.Sprintf("unknown geometry type: %v", g.Type()))
	}
}
