package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

// TestSortedPackedIntervalRTreeEmpty tests that querying an empty tree does
// not cause issues. See JTS GH Issue #19. Used to infinite-loop on empty
// geometries.
func TestSortedPackedIntervalRTreeEmpty(t *testing.T) {
	spitree := jts.IndexIntervalrtree_NewSortedPackedIntervalRTree()
	visitor := jts.Index_NewArrayListVisitor()
	spitree.Query(0, 1, visitor)
}
