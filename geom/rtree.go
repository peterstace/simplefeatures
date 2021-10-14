package geom

import "github.com/peterstace/simplefeatures/rtree"

// indexedLines is a simple container to hold a list of lines, and a r-tree
// structure indexing those lines. The record IDs in the rtree correspond to
// the indices of the lines slice.
type indexedLines struct {
	lines []line
	tree  *rtree.RTree
}

func newIndexedLines(lines []line) indexedLines {
	bulk := make([]rtree.BulkItem, len(lines))
	for i, ln := range lines {
		bulk[i] = rtree.BulkItem{
			Box:      ln.box(),
			RecordID: i,
		}
	}
	return indexedLines{lines, rtree.BulkLoad(bulk)}
}

// indexedPoints is a simple container to hold a list of points, and a r-tree
// structure indexing those points. The record IDs in the rtree correspond to
// the indices of the points slice.
type indexedPoints struct {
	points []XY
	tree   *rtree.RTree
}

func newIndexedPoints(points []XY) indexedPoints {
	bulk := make([]rtree.BulkItem, len(points))
	for i, pt := range points {
		bulk[i] = rtree.BulkItem{
			Box:      rtree.Box{MinX: pt.X, MaxX: pt.X, MinY: pt.Y, MaxY: pt.Y},
			RecordID: i,
		}
	}
	return indexedPoints{points, rtree.BulkLoad(bulk)}
}
