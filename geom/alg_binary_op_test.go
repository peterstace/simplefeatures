package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

// Results for the following tests can be found using the following style of
// SQL query:
//
// WITH CONST AS (
//   SELECT
//     ST_GeomFromText('POLYGON((0 0,1 2,2 0,0 0))') AS input1,
//     ST_GeomFromText('POLYGON((0 1,2 1,1 3,0 1))') AS input2
// )
// SELECT
//   ST_AsText(input1) AS input2,
//   ST_AsText(input2) AS input1,
//   ST_AsText(ST_Union(input1, input2)) AS union,
//   ST_AsText(ST_Intersection(input1, input2)) AS inter,
//   ST_AsText(ST_Difference(input1, input2)) AS fwd_diff,
//   ST_AsText(ST_Difference(input2, input1)) AS rev_diff,
//   ST_AsText(ST_SymDifference(input2, input1)) AS sym_diff
//   FROM const;

func TestBinaryOp(t *testing.T) {
	for i, geomCase := range []struct {
		input1, input2                          string
		union, inter, fwdDiff, revDiff, symDiff string
	}{
		{
			/*
			          /\
			         /  \
			        /    \
			       /      \
			      /   /\   \
			     /   /  \   \
			    /   /    \   \
			   +---/------\---+
			      /        \
			     /          \
			    /            \
			   +--------------+
			*/
			input1:  "POLYGON((0 0,1 2,2 0,0 0))",
			input2:  "POLYGON((0 1,2 1,1 3,0 1))",
			union:   "POLYGON((0 0,0.5 1,0 1,1 3,2 1,1.5 1,2 0,0 0))",
			inter:   "POLYGON((0.5 1,1 2,1.5 1,0.5 1))",
			fwdDiff: "POLYGON((0 0,2 0,1.5 1,0.5 1,0 0))",
			revDiff: "POLYGON((1 3,2 1,1.5 1,1 2,0.5 1,0 1,1 3))",
			symDiff: "MULTIPOLYGON(((0 0,2 0,1.5 1,0.5 1,0 0)),((0 1,0.5 1,1 2,1.5 1,2 1,1 3,0 1)))",
		},
		{
			/*
			         +-----------+
			         |           |
			         |           |
			   +-----+-----+     |
			   |     |     |     |
			   |     |     |     |
			   |     +-----+-----+
			   |           |
			   |           |
			   +-----------+
			*/
			input1:  "POLYGON((0 0,2 0,2 2,0 2,0 0))",
			input2:  "POLYGON((1 1,3 1,3 3,1 3,1 1))",
			union:   "POLYGON((0 0,2 0,2 1,3 1,3 3,1 3,1 2,0 2,0 0))",
			inter:   "POLYGON((1 1,2 1,2 2,1 2,1 1))",
			fwdDiff: "POLYGON((0 0,2 0,2 1,1 1,1 2,0 2,0 0))",
			revDiff: "POLYGON((2 1,3 1,3 3,1 3,1 2,2 2,2 1))",
			symDiff: "MULTIPOLYGON(((0 0,2 0,2 1,1 1,1 2,0 2,0 0)),((2 1,3 1,3 3,1 3,1 2,2 2,2 1)))",
		},
		{
			/*
			               +-----+
			               |     |
			               |     |
			               +-----+


			   +-----+
			   |     |
			   |     |
			   +-----+
			*/
			input1:  "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			input2:  "POLYGON((2 2,3 2,3 3,2 3,2 2))",
			union:   "MULTIPOLYGON(((0 0,1 0,1 1,0 1,0 0)),((2 2,3 2,3 3,2 3,2 2)))",
			inter:   "POLYGON EMPTY", // TODO: should this be GEOMETRYCOLLECTION EMPTY?
			fwdDiff: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			revDiff: "POLYGON((2 2,3 2,3 3,2 3,2 2))",
			symDiff: "MULTIPOLYGON(((0 0,1 0,1 1,0 1,0 0)),((2 2,3 2,3 3,2 3,2 2)))",
		},
		{
			/*
			   +-----------------+
			   |                 |
			   |                 |
			   |     +-----+     |
			   |     |     |     |
			   |     |     |     |
			   |     +-----+     |
			   |                 |
			   |                 |
			   +-----------------+
			*/
			input1:  "POLYGON((0 0,3 0,3 3,0 3,0 0))",
			input2:  "POLYGON((1 1,2 1,2 2,1 2,1 1))",
			union:   "POLYGON((0 0,3 0,3 3,0 3,0 0))",
			inter:   "POLYGON((1 1,2 1,2 2,1 2,1 1))",
			fwdDiff: "POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))",
			revDiff: "POLYGON EMPTY", // TODO: should this be GEOMETRYCOLLECTION EMPTY?
			symDiff: "POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,2 1,2 2,1 2,1 1))",
		},
		{
			/*
			   +-----+
			   | A   |
			   |     |
			   +-----+


			   +-----------+
			   | A         |
			   |           |
			   |     +-----+-----+
			   |     | A&B |     |
			   |     |     |     |
			   +-----+-----+     |     +-----+
			         |           |     | B   |
			         |         B |     |     |
			   o     +-----------+     +-----+
			*/
			input1:  "MULTIPOLYGON(((0 4,0 5,1 5,1 4,0 4)),((0 1,0 3,2 3,2 1,0 1)))",
			input2:  "MULTIPOLYGON(((4 0,4 1,5 1,5 0,4 0)),((1 0,1 2,3 2,3 0,1 0)))",
			union:   "MULTIPOLYGON(((0 4,0 5,1 5,1 4,0 4)),((0 1,0 3,2 3,2 2,3 2,3 0,1 0,1 1,0 1)),((4 0,4 1,5 1,5 0,4 0)))",
			inter:   "POLYGON((2 2,2 1,1 1,1 2,2 2))",
			fwdDiff: "MULTIPOLYGON(((0 4,0 5,1 5,1 4,0 4)),((0 1,0 3,2 3,2 2,1 2,1 1,0 1)))",
			revDiff: "MULTIPOLYGON(((4 0,4 1,5 1,5 0,4 0)),((1 0,1 1,2 1,2 2,3 2,3 0,1 0)))",
			symDiff: "MULTIPOLYGON(((0 4,0 5,1 5,1 4,0 4)),((0 1,0 3,2 3,2 2,1 2,1 1,0 1)),((1 1,2 1,2 2,3 2,3 0,1 0,1 1)),((4 0,4 1,5 1,5 0,4 0)))",
		},
		{
			/*

			   Two interlocking rings:

			   +-------------------+
			   |                   |
			   |   +-----------+   |
			   |   |           |   |
			   |   |   +-------+---+-------+
			   |   |   |       |   |       |
			   |   |   |   +---+---+---+   |
			   |   |   |   |   |   |   |   |
			   |   +---+---+---+   |   |   |
			   |       |   |       |   |   |
			   +-------+---+-------+   |   |
			           |   |           |   |
			           |   +-----------+   |
			           |                   |
			           +-------------------+
			*/
			input1:  "POLYGON((0 2,5 2,5 7,0 7,0 2),(1 3,4 3,4 6,1 6,1 3))",
			input2:  "POLYGON((2 0,7 0,7 5,2 5,2 0),(3 1,6 1,6 4,3 4,3 1))",
			union:   "POLYGON((2 2,0 2,0 7,5 7,5 5,7 5,7 0,2 0,2 2),(5 4,5 2,3 2,3 1,6 1,6 4,5 4),(1 3,2 3,2 5,4 5,4 6,1 6,1 3),(3 3,4 3,4 4,3 4,3 3))",
			inter:   "MULTIPOLYGON(((3 2,2 2,2 3,3 3,3 2)),((5 5,5 4,4 4,4 5,5 5)))",
			fwdDiff: "MULTIPOLYGON(((2 2,0 2,0 7,5 7,5 5,4 5,4 6,1 6,1 3,2 3,2 2)),((5 4,5 2,3 2,3 3,4 3,4 4,5 4)))",
			revDiff: "MULTIPOLYGON(((5 5,7 5,7 0,2 0,2 2,3 2,3 1,6 1,6 4,5 4,5 5)),((2 3,2 5,4 5,4 4,3 4,3 3,2 3)))",
			symDiff: "MULTIPOLYGON(((5 5,7 5,7 0,2 0,2 2,3 2,3 1,6 1,6 4,5 4,5 5)),((5 5,4 5,4 6,1 6,1 3,2 3,2 2,0 2,0 7,5 7,5 5)),((2 3,2 5,4 5,4 4,3 4,3 3,2 3)),((4 4,5 4,5 2,3 2,3 3,4 3,4 4)))",
		},
		//{
		//	/*
		//	   +-----+-----+
		//	   | B   | A   |
		//	   |     |     |
		//	   +-----+-----+
		//	   | A   | B   |
		//	   |     |     |
		//	   +-----+-----+
		//	*/
		//	input1: "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((1 1,1 2,2 2,2 1,1 1)))",
		//	input2: "MULTIPOLYGON(((0 1,0 2,1 2,1 1,0 1)),((1 0,1 1,2 1,2 0,1 0)))",
		//	union:  "POLYGON((0 0,2 0,2 2,0 2,0 0))",
		//	// TODO: We don't yet support linear output elements.
		//	//inter:   "MULTILINESTRING((0 1,1 1),(1 1,1 0),(1 1,1 2),(2 1,1 1))",
		//	fwdDiff: "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((1 1,1 2,2 2,2 1,1 1)))",
		//	revDiff: "MULTIPOLYGON(((0 1,0 2,1 2,1 1,0 1)),((1 0,1 1,2 1,2 0,1 0)))",
		//	symDiff: "POLYGON((0 0,0 1,0 2,1 2,2 2,2 1,2 0,1 0,0 0))",
		//},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g1 := geomFromWKT(t, geomCase.input1)
			g2 := geomFromWKT(t, geomCase.input2)
			for _, opCase := range []struct {
				opName string
				op     func(geom.Geometry, geom.Geometry) geom.Geometry
				want   string
			}{
				{"union", geom.Union, geomCase.union},
				{"inter", geom.Intersection, geomCase.inter},
				{"fwd_diff", geom.Difference, geomCase.fwdDiff},
				{"rev_diff", func(a, b geom.Geometry) geom.Geometry { return geom.Difference(b, a) }, geomCase.revDiff},
				{"sym_diff", geom.SymmetricDifference, geomCase.symDiff},
			} {
				t.Run(opCase.opName, func(t *testing.T) {
					if opCase.want == "" {
						// Allows tests to be skipped by just commenting them out.
						t.Skip("Skipping test because it would fail")
					}
					want := geomFromWKT(t, opCase.want)
					got := opCase.op(g1, g2)
					expectGeomEq(t, got, want, geom.IgnoreOrder)
				})
			}
		})
	}
}