package geom

import "github.com/peterstace/simplefeatures/rtree"

// indexedLines is a simple container to hold a list of lines, and a r-tree
// structure indexing those lines. The record IDs in the rtree correspond to
// the indices of the lines slice.
type indexedLines struct {
	lines []line
	tree  rtree.RTree
}

func newIndexedLines(lines []line) indexedLines {
	bulk := make([]rtree.BulkItem, len(lines))
	for i, ln := range lines {
		bulk[i] = rtree.BulkItem{
			Box:      ln.envelope().box(),
			RecordID: i,
		}
	}
	return indexedLines{lines, rtree.BulkLoad(bulk)}
}
