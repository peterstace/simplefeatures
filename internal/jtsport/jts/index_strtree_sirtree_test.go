package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestSIRtree(t *testing.T) {
	tree := jts.IndexStrtree_NewSIRtreeWithCapacity(2)
	tree.Insert(2, 6, "A")
	tree.Insert(2, 4, "B")
	tree.Insert(2, 3, "C")
	tree.Insert(2, 4, "D")
	tree.Insert(0, 1, "E")
	tree.Insert(2, 4, "F")
	tree.Insert(5, 6, "G")
	tree.Build()

	junit.AssertEquals(t, 2, tree.GetRoot().GetLevel())
	junit.AssertEquals(t, 4, len(tree.BoundablesAtLevel(0)))
	junit.AssertEquals(t, 2, len(tree.BoundablesAtLevel(1)))
	junit.AssertEquals(t, 1, len(tree.BoundablesAtLevel(2)))
	junit.AssertEquals(t, 1, len(tree.QueryRange(0.5, 0.5)))
	junit.AssertEquals(t, 0, len(tree.QueryRange(1.5, 1.5)))
	junit.AssertEquals(t, 2, len(tree.QueryRange(4.5, 5.5)))
}

func TestSIRtreeEmptyTree(t *testing.T) {
	tree := jts.IndexStrtree_NewSIRtreeWithCapacity(2)
	tree.Build()

	junit.AssertEquals(t, 0, tree.GetRoot().GetLevel())
	junit.AssertEquals(t, 1, len(tree.BoundablesAtLevel(0)))
	junit.AssertEquals(t, 0, len(tree.BoundablesAtLevel(1)))
	junit.AssertEquals(t, 0, len(tree.BoundablesAtLevel(-1)))
	junit.AssertEquals(t, 0, len(tree.QueryRange(0.5, 0.5)))
}
