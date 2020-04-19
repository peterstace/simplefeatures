package geom

import "github.com/peterstace/simplefeatures/rtree"

// TODO: make this a method on Envelope
func toBox(env Envelope) rtree.Box {
	return rtree.Box{
		MinX: env.min.X,
		MinY: env.min.Y,
		MaxX: env.max.X,
		MaxY: env.max.Y,
	}
}

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
			Box:      toBox(ln.envelope()),
			RecordID: i,
		}
	}
	return indexedLines{lines, rtree.BulkLoad(bulk)}
}
