package geom

import "github.com/peterstace/simplefeatures/rtree"

// TODO: Use this instead of indexedLines/Points where possible.
func newLineRTree(lines []line) *rtree.RTree[line] {
	items := make([]rtree.BulkItem[line], len(lines))
	for i, ln := range lines {
		items[i] = rtree.BulkItem[line]{
			Box:    ln.box(),
			Record: ln,
		}
	}
	return rtree.BulkLoad(items)
}

// TODO: Use this instead of indexedLines/Points where possible.
func newPointRTree(points []XY) *rtree.RTree[XY] {
	items := make([]rtree.BulkItem[XY], len(points))
	for i, pt := range points {
		items[i] = rtree.BulkItem[XY]{
			Box:    pt.box(),
			Record: pt,
		}
	}
	return rtree.BulkLoad(items)
}

// indexedLines is a simple container to hold a list of lines, and a r-tree
// structure indexing those lines. The record IDs in the rtree correspond to
// the indices of the lines slice.
type indexedLines struct {
	lines []line
	tree  *rtree.RTree[int]
}

func newIndexedLines(lines []line) indexedLines {
	bulk := make([]rtree.BulkItem[int], len(lines))
	for i, ln := range lines {
		bulk[i] = rtree.BulkItem[int]{
			Box:    ln.box(),
			Record: i,
		}
	}
	return indexedLines{lines, rtree.BulkLoad(bulk)}
}

// indexedPoints is a simple container to hold a list of points, and a r-tree
// structure indexing those points. The record IDs in the rtree correspond to
// the indices of the points slice.
type indexedPoints struct {
	points []XY
	tree   *rtree.RTree[int]
}

func newIndexedPoints(points []XY) indexedPoints {
	bulk := make([]rtree.BulkItem[int], len(points))
	for i, pt := range points {
		bulk[i] = rtree.BulkItem[int]{
			Box:    rtree.Box{MinX: pt.X, MaxX: pt.X, MinY: pt.Y, MaxY: pt.Y},
			Record: i,
		}
	}
	return indexedPoints{points, rtree.BulkLoad(bulk)}
}
