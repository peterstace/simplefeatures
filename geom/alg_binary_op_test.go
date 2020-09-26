package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

// Results for the following tests can be found using the following style of
// SQL query:
//
// WITH const AS (
//   SELECT
//     ST_GeomFromText('POLYGON((0 0,1 2,2 0,0 0))') AS input1,
//     ST_GeomFromText('POLYGON((0 1,2 1,1 3,0 1))') AS input2
// )
// SELECT
//   ST_AsText(input1) AS input1,
//   ST_AsText(input2) AS input2,
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
			inter:   "GEOMETRYCOLLECTION EMPTY",
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
			revDiff: "GEOMETRYCOLLECTION EMPTY",
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
		{
			/*

			      /\      /\
			     /  \    /  \
			    / A  \  / A  \
			   /      \/      \
			   \  /\  /\  /\  /
			    \/AB\/  \/AB\/
			    /\  /\  /\  /\
			   /  \/  \/  \/  \
			   \      /\      /
			    \ B  /  \ B  /
			     \  /    \  /
			      \/      \/

			*/
			input1:  "MULTIPOLYGON(((0 2,1 1,2 2,1 3,0 2)),((2 2,3 1,4 2,3 3,2 2)))",
			input2:  "MULTIPOLYGON(((1 0,0 1,1 2,2 1,1 0)),((3 0,4 1,3 2,2 1,3 0)))",
			union:   "MULTIPOLYGON(((0.5 1.5,0 2,1 3,2 2,1.5 1.5,2 1,1 0,0 1,0.5 1.5)),((2.5 1.5,2 2,3 3,4 2,3.5 1.5,4 1,3 0,2 1,2.5 1.5)))",
			inter:   "MULTIPOLYGON(((1.5 1.5,1 1,0.5 1.5,1 2,1.5 1.5)),((3.5 1.5,3 1,2.5 1.5,3 2,3.5 1.5)))",
			fwdDiff: "MULTIPOLYGON(((0.5 1.5,0 2,1 3,2 2,1.5 1.5,1 2,0.5 1.5)),((2.5 1.5,2 2,3 3,4 2,3.5 1.5,3 2,2.5 1.5)))",
			revDiff: "MULTIPOLYGON(((1 0,0 1,0.5 1.5,1 1,1.5 1.5,2 1,1 0)),((3.5 1.5,4 1,3 0,2 1,2.5 1.5,3 1,3.5 1.5)))",
			symDiff: "MULTIPOLYGON(((1 0,0 1,0.5 1.5,1 1,1.5 1.5,2 1,1 0)),((1.5 1.5,1 2,0.5 1.5,0 2,1 3,2 2,1.5 1.5)),((3.5 1.5,4 1,3 0,2 1,2.5 1.5,3 1,3.5 1.5)),((3.5 1.5,3 2,2.5 1.5,2 2,3 3,4 2,3.5 1.5)))",
		},

		{
			/*
			   +-----+-----+
			   | B   | A   |
			   |     |     |
			   +-----+-----+
			   | A   | B   |
			   |     |     |
			   +-----+-----+
			*/
			input1:  "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((1 1,1 2,2 2,2 1,1 1)))",
			input2:  "MULTIPOLYGON(((0 1,0 2,1 2,1 1,0 1)),((1 0,1 1,2 1,2 0,1 0)))",
			union:   "POLYGON((0 0,0 1,0 2,1 2,2 2,2 1,2 0,1 0,0 0))",
			inter:   "MULTILINESTRING((0 1,1 1),(1 1,1 0),(1 1,1 2),(2 1,1 1))",
			fwdDiff: "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((1 1,1 2,2 2,2 1,1 1)))",
			revDiff: "MULTIPOLYGON(((0 1,0 2,1 2,1 1,0 1)),((1 0,1 1,2 1,2 0,1 0)))",
			symDiff: "POLYGON((0 0,0 1,0 2,1 2,2 2,2 1,2 0,1 0,0 0))",
		},
		{
			/*
			   +-----+-----+
			   | A   | B   |
			   |     |     |
			   +-----+-----+
			*/
			input1:  "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			input2:  "POLYGON((1 0,1 1,2 1,2 0,1 0))",
			union:   "POLYGON((0 0,0 1,1 1,2 1,2 0,1 0,0 0))",
			inter:   "LINESTRING(1 1,1 0)",
			fwdDiff: "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			revDiff: "POLYGON((1 0,1 1,2 1,2 0,1 0))",
			symDiff: "POLYGON((1 1,2 1,2 0,1 0,0 0,0 1,1 1))",
		},
		{
			/*
			   +-------+
			   | A     |
			   |       +-------+
			   |       | B     |
			   +-------+       |
			           |       |
			           +-------+
			*/
			input1:  "POLYGON((0 0.5,0 1.5,1 1.5,1 0.5,0 0.5))",
			input2:  "POLYGON((1 0,1 1,2 1,2 0,1 0))",
			union:   "POLYGON((0 0.5,0 1.5,1 1.5,1 1,2 1,2 0,1 0,1 0.5,0 0.5))",
			inter:   "LINESTRING(1 1,1 0.5)",
			fwdDiff: "POLYGON((0 0.5,0 1.5,1 1.5,1 1,1 0.5,0 0.5))",
			revDiff: "POLYGON((1 0,1 0.5,1 1,2 1,2 0,1 0))",
			symDiff: "POLYGON((1 0,1 0.5,0 0.5,0 1.5,1 1.5,1 1,2 1,2 0,1 0))",
		},
		{
			/*
			   +-----+
			   | A&B |
			   |     |
			   +-----+
			*/
			input1:  "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			input2:  "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			union:   "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			inter:   "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			fwdDiff: "GEOMETRYCOLLECTION EMPTY",
			revDiff: "GEOMETRYCOLLECTION EMPTY",
			symDiff: "GEOMETRYCOLLECTION EMPTY",
		},
		{
			/*
			   *-------*
			   |\ A&B /|
			   | \   / |
			   |  \ /  |
			   *   *   *
			   | A | B |
			   |   |   |
			   *---*---*
			*/
			input1:  "POLYGON((0 0,0 2,2 2,1 1,1 0,0 0))",
			input2:  "POLYGON((1 0,1 1,0 2,2 2,2 0,1 0))",
			union:   "POLYGON((0 0,0 2,2 2,2 0,1 0,0 0))",
			inter:   "GEOMETRYCOLLECTION(LINESTRING(1 1,1 0),POLYGON((0 2,2 2,1 1,0 2)))",
			fwdDiff: "POLYGON((0 0,0 2,1 1,1 0,0 0))",
			revDiff: "POLYGON((1 0,1 1,2 2,2 0,1 0))",
			symDiff: "POLYGON((0 2,1 1,2 2,2 0,1 0,0 0,0 2))",
		},
		{
			/*
			   +---+
			   | A |
			   +---+---+
			       | B |
			       +---+
			*/
			input1:  "POLYGON((0 1,1 1,1 2,0 2,0 1))",
			input2:  "POLYGON((1 0,2 0,2 1,1 1,1 0))",
			union:   "MULTIPOLYGON(((1 1,0 1,0 2,1 2,1 1)),((1 1,2 1,2 0,1 0,1 1)))",
			inter:   "POINT(1 1)",
			fwdDiff: "POLYGON((1 1,0 1,0 2,1 2,1 1))",
			revDiff: "POLYGON((1 1,2 1,2 0,1 0,1 1))",
			symDiff: "MULTIPOLYGON(((1 1,2 1,2 0,1 0,1 1)),((1 1,0 1,0 2,1 2,1 1)))",
		},
		{
			/*
			   +-----+-----+
			   |    / \    |
			   |   +-+-+   |
			   | A   |   B |
			   +-----+-----+
			*/
			input1:  "POLYGON((0 0,2 0,2 1,1 1,2 2,0 2,0 0))",
			input2:  "POLYGON((2 0,4 0,4 2,2 2,3 1,2 1,2 0))",
			union:   "POLYGON((2 0,0 0,0 2,2 2,4 2,4 0,2 0),(2 2,1 1,2 1,3 1,2 2))",
			inter:   "GEOMETRYCOLLECTION(POINT(2 2),LINESTRING(2 0,2 1))",
			fwdDiff: "POLYGON((2 0,0 0,0 2,2 2,1 1,2 1,2 0))",
			revDiff: "POLYGON((2 2,4 2,4 0,2 0,2 1,3 1,2 2))",
			symDiff: "POLYGON((2 2,4 2,4 0,2 0,0 0,0 2,2 2),(2 2,1 1,2 1,3 1,2 2))",
		},
		{
			/*
			        +---+
			        | A |
			        +---+---+
			            | B |
			   +---+    +---+
			   |A&B|
			   +---+
			*/
			input1:  "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((1 2,2 2,2 3,1 3,1 2)))",
			input2:  "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 1,3 1,3 2,2 2,2 1)))",
			union:   "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 2,1 2,1 3,2 3,2 2)),((2 2,3 2,3 1,2 1,2 2)))",
			inter:   "GEOMETRYCOLLECTION(POINT(2 2),POLYGON((0 0,0 1,1 1,1 0,0 0)))",
			fwdDiff: "POLYGON((2 2,1 2,1 3,2 3,2 2))",
			revDiff: "POLYGON((2 2,3 2,3 1,2 1,2 2))",
			symDiff: "MULTIPOLYGON(((2 2,3 2,3 1,2 1,2 2)),((2 2,1 2,1 3,2 3,2 2)))",
		},
		{
			/*
			       +-------+
			       |       |
			   +---+---+   |
			   |   |   |   |
			   |   +---+   |
			   | A |       |
			   |   +---+   |
			   |   |   |   |
			   +---+---+   |
			   |A&B|     B |
			   +---+-------+
			*/
			input1:  "POLYGON((0 0,1 0,1 4,0 4,0 0))",
			input2:  "POLYGON((0 0,3 0,3 5,1 5,1 4,2 4,2 3,1 3,1 2,2 2,2 1,0 1,0 0))",
			union:   "POLYGON((1 0,0 0,0 1,0 4,1 4,1 5,3 5,3 0,1 0),(1 4,1 3,2 3,2 4,1 4),(1 2,1 1,2 1,2 2,1 2))",
			inter:   "GEOMETRYCOLLECTION(POINT(1 4),LINESTRING(1 2,1 3),POLYGON((1 0,0 0,0 1,1 1,1 0)))",
			fwdDiff: "POLYGON((1 2,1 1,0 1,0 4,1 4,1 3,1 2))",
			revDiff: "POLYGON((1 4,1 5,3 5,3 0,1 0,1 1,2 1,2 2,1 2,1 3,2 3,2 4,1 4))",
			symDiff: "POLYGON((1 4,1 5,3 5,3 0,1 0,1 1,0 1,0 4,1 4),(1 1,2 1,2 2,1 2,1 1),(1 4,1 3,2 3,2 4,1 4))",
		},
		{
			/*
			   +-------+-------+
			   | A     |     B |
			   |   +---+---+   |
			   |   |       |   |
			   |   +---+---+   |
			   |       |       |
			   +-------+-------+
			*/

			input1:  "POLYGON((0 0,2 0,2 1,1 1,1 2,2 2,2 3,0 3,0 0))",
			input2:  "POLYGON((2 0,4 0,4 3,2 3,2 2,3 2,3 1,2 1,2 0))",
			union:   "POLYGON((2 0,0 0,0 3,2 3,4 3,4 0,2 0),(2 2,1 2,1 1,2 1,3 1,3 2,2 2))",
			inter:   "MULTILINESTRING((2 0,2 1),(2 2,2 3))",
			fwdDiff: "POLYGON((2 0,0 0,0 3,2 3,2 2,1 2,1 1,2 1,2 0))",
			revDiff: "POLYGON((2 3,4 3,4 0,2 0,2 1,3 1,3 2,2 2,2 3))",
			symDiff: "POLYGON((2 3,4 3,4 0,2 0,0 0,0 3,2 3),(2 1,3 1,3 2,2 2,1 2,1 1,2 1))",
		},
		{
			/*
			   *-------------+
			   |\`.        B |
			   | \ `.        |
			   |  \  `.      |
			   |   \   `*    |
			   |    *    \   |
			   |     `.   \  |
			   |       `.  \ |
			   | A       `. \|
			   +-----------`-*
			*/

			input1:  "POLYGON((0 0,3 0,1 1,0 3,0 0))",
			input2:  "POLYGON((3 0,3 3,0 3,2 2,3 0))",
			union:   "MULTIPOLYGON(((3 0,0 0,0 3,1 1,3 0)),((0 3,3 3,3 0,2 2,0 3)))",
			inter:   "MULTIPOINT(0 3,3 0)",
			fwdDiff: "POLYGON((3 0,0 0,0 3,1 1,3 0))",
			revDiff: "POLYGON((0 3,3 3,3 0,2 2,0 3))",
			symDiff: "MULTIPOLYGON(((0 3,3 3,3 0,2 2,0 3)),((3 0,0 0,0 3,1 1,3 0)))",
		},
		{
			/*
			   +
			   |A
			   |   B
			   +----+
			*/
			input1:  "LINESTRING(0 0,0 1)",
			input2:  "LINESTRING(0 0,1 0)",
			union:   "MULTILINESTRING((0 0,0 1),(0 0,1 0))",
			inter:   "POINT(0 0)",
			fwdDiff: "LINESTRING(0 0,0 1)",
			revDiff: "LINESTRING(0 0,1 0)",
			symDiff: "MULTILINESTRING((0 0,1 0),(0 0,0 1))",
		},
		{
			/*
			   +       +
			   |       |
			   A       B
			   |       |
			   +--A&B--+
			*/
			input1:  "LINESTRING(0 1,0 0,1 0)",
			input2:  "LINESTRING(0 0,1 0,1 1)",
			union:   "MULTILINESTRING((0 1,0 0),(0 0,1 0),(1 0,1 1))",
			inter:   "LINESTRING(0 0,1 0)",
			fwdDiff: "LINESTRING(0 1,0 0)",
			revDiff: "LINESTRING(1 0,1 1)",
			symDiff: "MULTILINESTRING((1 0,1 1),(0 1,0 0))",
		},
		{
			/*
			   \      /
			    \    /
			     B  A
			      \/
			      /\
			     A  B
			    /    \
			   /      \
			*/
			input1:  "LINESTRING(0 0,1 1)",
			input2:  "LINESTRING(0 1,1 0)",
			union:   "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1),(0 1,0.5 0.5),(0.5 0.5,1 0))",
			inter:   "POINT(0.5 0.5)",
			fwdDiff: "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1))",
			revDiff: "MULTILINESTRING((0 1,0.5 0.5),(0.5 0.5,1 0))",
			symDiff: "MULTILINESTRING((0 1,0.5 0.5),(0.5 0.5,1 0),(0 0,0.5 0.5),(0.5 0.5,1 1))",
		},
		{
			//    +---A---+
			//    |       |
			//    B       B
			//    |       |
			//    +---A---+
			//
			input1:  "MULTILINESTRING((0 0,1 0),(0 1,1 1))",
			input2:  "MULTILINESTRING((0 0,0 1),(1 0,1 1))",
			union:   "MULTILINESTRING((0 0,1 0),(0 1,1 1),(0 0,0 1),(1 0,1 1))",
			inter:   "MULTIPOINT(0 0,0 1,1 0,1 1)",
			fwdDiff: "MULTILINESTRING((0 0,1 0),(0 1,1 1))",
			revDiff: "MULTILINESTRING((0 0,0 1),(1 0,1 1))",
			symDiff: "MULTILINESTRING((0 0,0 1),(1 0,1 1),(0 0,1 0),(0 1,1 1))",
		},
		{
			/*
			   +--A&B--+---A---+
			   |       |       |
			  A&B      B       A
			   |       |       |
			   +---A---+---A---+
			   |       |
			   B       B
			   |       |
			   +---B---+
			*/
			input1: "LINESTRING(0 2,2 2,2 1,0 1,0 2)",
			input2: "LINESTRING(1 2,1 0,0 0,0 2,1 2)",
			// TODO: I needed to make some structural modifications to the linear elements manually compared to PostGIS output.
			union:   "MULTILINESTRING((0 2,1 2),(1 2,2 2),(2 2,2 1),(2 1,1 1),(1 1,0 1),(0 1,0 2),(1 2,1 1),(1 1,1 0),(1 0,0 0),(0 0,0 1))",
			inter:   "GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 2,1 2),LINESTRING(0 1,0 2))",
			fwdDiff: "MULTILINESTRING((1 2,2 2),(2 2,2 1),(2 1,1 1),(1 1,0 1))",
			revDiff: "MULTILINESTRING((1 2,1 1),(1 1,1 0),(1 0,0 0),(0 0,0 1))",
			symDiff: "MULTILINESTRING((1 2,1 1),(1 1,1 0),(1 0,0 0),(0 0,0 1),(1 2,2 2),(2 2,2 1),(2 1,1 1),(1 1,0 1))",
		},
		{
			/*
			  +---------+
			   `,     ,` `,
			     `, ,`     `,
			      ,`,       ,`
			    ,`   `,   ,`
			  +`       `+`

			*/
			input1: "LINESTRING(0 0,2 2,0 2,2 0)",
			input2: "LINESTRING(2 0,3 1,2 2)",
			// TODO: I needed to make some structural modifications to the linear elements manually compared to PostGIS output.
			union:   "MULTILINESTRING((0 0,1 1),(1 1,2 2),(2 2,0 2),(0 2,1 1),(1 1,2 0),(2 0,3 1),(3 1,2 2))",
			inter:   "MULTIPOINT(2 0,2 2)",
			fwdDiff: "MULTILINESTRING((0 0,1 1),(1 1,2 2),(2 2,0 2),(0 2,1 1),(1 1,2 0))",
			revDiff: "MULTILINESTRING((2 0,3 1),(3 1,2 2))",
			symDiff: "MULTILINESTRING((2 0,3 1),(3 1,2 2),(0 0,1 1),(1 1,2 2),(2 2,0 2),(0 2,1 1),(1 1,2 0))",
		},
		{
			/*
			       +
			       |
			   +---+---+
			   |   |   |
			   |   +   |
			   |       |
			   +-------+
			*/
			input1:  "POLYGON((0 0,0 2,2 2,2 0,0 0))",
			input2:  "LINESTRING(1 1,1 3)",
			union:   "GEOMETRYCOLLECTION(LINESTRING(1 2,1 3),POLYGON((0 0,0 2,1 2,2 2,2 0,0 0)))",
			inter:   "LINESTRING(1 1,1 2)",
			fwdDiff: "POLYGON((0 0,0 2,1 2,2 2,2 0,0 0))",
			revDiff: "LINESTRING(1 2,1 3)",
			symDiff: "GEOMETRYCOLLECTION(LINESTRING(1 2,1 3),POLYGON((0 0,0 2,1 2,2 2,2 0,0 0)))",
		},
		{
			/*
			   +--------+
			   |     ,  |
			   |   ,`   |
			   |  `     |
			   +--------+
			*/
			input1:  "POLYGON((0 0,0 3,3 3,3 0,0 0))",
			input2:  "LINESTRING(1 1,2 2)",
			union:   "POLYGON((0 0,0 3,3 3,3 0,0 0))",
			inter:   "LINESTRING(1 1,2 2)",
			fwdDiff: "POLYGON((0 0,0 3,3 3,3 0,0 0))",
			revDiff: "GEOMETRYCOLLECTION EMPTY",
			symDiff: "POLYGON((0 0,0 3,3 3,3 0,0 0))",
		},
		{
			/*
			   +---+---+---+
			   |   A   |A&B|
			   +---+---+---+
			   |A&B|   B   |
			   +---+---+---+
			   |   A   |A&B|
			   +---+---+---+
			*/
			input1:  "POLYGON((0 0,3 0,3 1,1 1,1 2,3 2,3 3,0 3,0 0))",
			input2:  "POLYGON((0 1,0 2,2 2,2 3,3 3,3 0,2 0,2 1,0 1))",
			union:   "POLYGON((2 0,0 0,0 1,0 2,0 3,2 3,3 3,3 2,3 1,3 0,2 0))",
			inter:   "GEOMETRYCOLLECTION(LINESTRING(2 1,1 1),LINESTRING(1 2,2 2),POLYGON((3 0,2 0,2 1,3 1,3 0)),POLYGON((1 2,1 1,0 1,0 2,1 2)),POLYGON((3 2,2 2,2 3,3 3,3 2)))",
			fwdDiff: "MULTIPOLYGON(((2 0,0 0,0 1,1 1,2 1,2 0)),((2 2,1 2,0 2,0 3,2 3,2 2)))",
			revDiff: "POLYGON((1 2,2 2,3 2,3 1,2 1,1 1,1 2))",
			symDiff: "POLYGON((1 2,0 2,0 3,2 3,2 2,3 2,3 1,2 1,2 0,0 0,0 1,1 1,1 2))",
		},
		{
			/*
			   +   +   +
			   A  A&B  B
			*/
			input1:  "MULTIPOINT(0 0,1 1)",
			input2:  "MULTIPOINT(1 1,2 2)",
			union:   "MULTIPOINT(0 0,1 1,2 2)",
			inter:   "POINT(1 1)",
			fwdDiff: "POINT(0 0)",
			revDiff: "POINT(2 2)",
			symDiff: "MULTIPOINT(0 0,2 2)",
		},
		{
			/*
			   +-------+
			   |       |
			   |   +   |   +
			   |       |
			   +-------+
			*/
			input1:  "POLYGON((0 0,0 2,2 2,2 0,0 0))",
			input2:  "MULTIPOINT(1 1,3 1)",
			union:   "GEOMETRYCOLLECTION(POINT(3 1),POLYGON((0 0,0 2,2 2,2 0,0 0)))",
			inter:   "POINT(1 1)",
			fwdDiff: "POLYGON((0 0,0 2,2 2,2 0,0 0))",
			revDiff: "POINT(3 1)",
			symDiff: "GEOMETRYCOLLECTION(POINT(3 1),POLYGON((0 0,0 2,2 2,2 0,0 0)))",
		},
		{
			/*
			   +
			   |\
			   | \
			   |  \
			   |   \
			   |    \
			   O-----+
			*/
			input1:  "POLYGON((0 0,0 1,1 0,0 0))",
			input2:  "POINT(0 0)",
			union:   "POLYGON((0 0,0 1,1 0,0 0))",
			inter:   "POINT(0 0)",
			fwdDiff: "POLYGON((0 0,0 1,1 0,0 0))",
			revDiff: "GEOMETRYCOLLECTION EMPTY",
			symDiff: "POLYGON((0 0,0 1,1 0,0 0))",
		},
		{
			/*
			   +
			   |\
			   | \
			   |  O
			   |   \
			   |    \
			   +-----+
			*/
			input1:  "POLYGON((0 0,0 1,1 0,0 0))",
			input2:  "POINT(0.5 0.5)",
			union:   "POLYGON((0 0,0 1,0.5 0.5,1 0,0 0))",
			inter:   "POINT(0.5 0.5)",
			fwdDiff: "POLYGON((0 0,0 1,0.5 0.5,1 0,0 0))",
			revDiff: "GEOMETRYCOLLECTION EMPTY",
			symDiff: "POLYGON((0 0,0 1,0.5 0.5,1 0,0 0))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g1 := geomFromWKT(t, geomCase.input1)
			g2 := geomFromWKT(t, geomCase.input2)
			t.Logf("input1: %s", geomCase.input1)
			t.Logf("input2: %s", geomCase.input2)
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
