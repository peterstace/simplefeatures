package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestDirectedEdgeComparator(t *testing.T) {
	d1 := jts.Planargraph_NewDirectedEdge(
		jts.Planargraph_NewNode(jts.Geom_NewCoordinateWithXY(0, 0)),
		jts.Planargraph_NewNode(jts.Geom_NewCoordinateWithXY(10, 10)),
		jts.Geom_NewCoordinateWithXY(10, 10),
		true,
	)
	d2 := jts.Planargraph_NewDirectedEdge(
		jts.Planargraph_NewNode(jts.Geom_NewCoordinateWithXY(0, 0)),
		jts.Planargraph_NewNode(jts.Geom_NewCoordinateWithXY(20, 20)),
		jts.Geom_NewCoordinateWithXY(20, 20),
		false,
	)
	junit.AssertEquals(t, 0, d2.CompareTo(d1))
}

func TestDirectedEdgeToEdges(t *testing.T) {
	d1 := jts.Planargraph_NewDirectedEdge(
		jts.Planargraph_NewNode(jts.Geom_NewCoordinateWithXY(0, 0)),
		jts.Planargraph_NewNode(jts.Geom_NewCoordinateWithXY(10, 10)),
		jts.Geom_NewCoordinateWithXY(10, 10),
		true,
	)
	d2 := jts.Planargraph_NewDirectedEdge(
		jts.Planargraph_NewNode(jts.Geom_NewCoordinateWithXY(20, 0)),
		jts.Planargraph_NewNode(jts.Geom_NewCoordinateWithXY(20, 10)),
		jts.Geom_NewCoordinateWithXY(20, 10),
		false,
	)
	edges := jts.Planargraph_DirectedEdge_ToEdges([]*jts.Planargraph_DirectedEdge{d1, d2})
	junit.AssertEquals(t, 2, len(edges))
	junit.AssertNull(t, edges[0])
	junit.AssertNull(t, edges[1])
}
