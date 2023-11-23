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
		input1, input2                                  string
		union, inter, fwdDiff, revDiff, symDiff, relate string
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
			relate:  "212101212",
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
			relate:  "212101212",
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
			relate:  "FF2FF1212",
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
			relate:  "212FF1FF2",
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
			relate:  "212101212",
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
			relate:  "212101212",
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
			input2:  "MULTIPOLYGON(((0 1,1 2,2 1,1 0,0 1)),((2 1,3 0,4 1,3 2,2 1)))",
			union:   "MULTIPOLYGON(((0.5 1.5,0 2,1 3,2 2,1.5 1.5,2 1,1 0,0 1,0.5 1.5)),((2.5 1.5,2 2,3 3,4 2,3.5 1.5,4 1,3 0,2 1,2.5 1.5)))",
			inter:   "MULTIPOLYGON(((1.5 1.5,1 1,0.5 1.5,1 2,1.5 1.5)),((3.5 1.5,3 1,2.5 1.5,3 2,3.5 1.5)))",
			fwdDiff: "MULTIPOLYGON(((0.5 1.5,0 2,1 3,2 2,1.5 1.5,1 2,0.5 1.5)),((2.5 1.5,2 2,3 3,4 2,3.5 1.5,3 2,2.5 1.5)))",
			revDiff: "MULTIPOLYGON(((1 0,0 1,0.5 1.5,1 1,1.5 1.5,2 1,1 0)),((3.5 1.5,4 1,3 0,2 1,2.5 1.5,3 1,3.5 1.5)))",
			symDiff: "MULTIPOLYGON(((1 0,0 1,0.5 1.5,1 1,1.5 1.5,2 1,1 0)),((1.5 1.5,1 2,0.5 1.5,0 2,1 3,2 2,1.5 1.5)),((3.5 1.5,4 1,3 0,2 1,2.5 1.5,3 1,3.5 1.5)),((3.5 1.5,3 2,2.5 1.5,2 2,3 3,4 2,3.5 1.5)))",
			relate:  "212101212",
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
			relate:  "FF2F11212",
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
			relate:  "FF2F11212",
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
			relate:  "FF2F11212",
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
			relate:  "2FFF1FFF2",
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
			relate:  "212111212",
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
			relate:  "FF2F01212",
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
			relate:  "FF2F11212",
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
			input1:  "MULTIPOLYGON(((1 1,1 0,0 0,0 1,1 1)),((1 2,2 2,2 3,1 3,1 2)))",
			input2:  "MULTIPOLYGON(((1 1,1 0,0 0,0 1,1 1)),((2 1,3 1,3 2,2 2,2 1)))",
			union:   "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 2,1 2,1 3,2 3,2 2)),((2 2,3 2,3 1,2 1,2 2)))",
			inter:   "GEOMETRYCOLLECTION(POINT(2 2),POLYGON((0 0,0 1,1 1,1 0,0 0)))",
			fwdDiff: "POLYGON((2 2,1 2,1 3,2 3,2 2))",
			revDiff: "POLYGON((2 2,3 2,3 1,2 1,2 2))",
			symDiff: "MULTIPOLYGON(((2 2,3 2,3 1,2 1,2 2)),((2 2,1 2,1 3,2 3,2 2)))",
			relate:  "2F2F11212",
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
			relate:  "212111212",
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
			relate:  "FF2F11212",
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
			relate:  "FF2F01212",
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
			relate:  "FF1F00102",
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
			relate:  "1010F0102",
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
			relate:  "0F1FF0102",
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
			relate:  "FF1F0F1F2",
		},
		{
			//nolint:dupword
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
			input1:  "LINESTRING(0 2,2 2,2 1,0 1,0 2)",
			input2:  "LINESTRING(1 2,1 0,0 0,0 2,1 2)",
			union:   "MULTILINESTRING((0 2,1 2),(1 2,2 2,2 1,1 1),(1 1,0 1),(0 1,0 2),(1 2,1 1),(1 1,1 0,0 0,0 1))",
			inter:   "GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 2,1 2),LINESTRING(0 1,0 2))",
			fwdDiff: "MULTILINESTRING((1 2,2 2,2 1,1 1),(1 1,0 1))",
			revDiff: "MULTILINESTRING((1 2,1 1),(1 1,1 0,0 0,0 1))",
			symDiff: "MULTILINESTRING((1 2,2 2,2 1,1 1),(1 1,0 1),(1 2,1 1),(1 1,1 0,0 0,0 1))",
			relate:  "1F1FFF1F2",
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
			input1:  "LINESTRING(0 0,2 2,0 2,2 0)",
			input2:  "LINESTRING(2 0,3 1,2 2)",
			union:   "MULTILINESTRING((0 0,1 1),(1 1,2 2),(2 2,0 2,1 1),(1 1,2 0),(2 0,3 1,2 2))",
			inter:   "MULTIPOINT(2 0,2 2)",
			fwdDiff: "MULTILINESTRING((0 0,1 1),(1 1,2 2),(2 2,0 2,1 1),(1 1,2 0))",
			revDiff: "LINESTRING(2 0,3 1,2 2)",
			symDiff: "MULTILINESTRING((0 0,1 1),(1 1,2 2),(2 2,0 2,1 1),(1 1,2 0),(2 0,3 1,2 2))",
			relate:  "F01F001F2",
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
			relate:  "1020F1102",
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
			relate:  "102FF1FF2",
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
			relate:  "212111212",
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
			relate:  "0F0FFF0F2",
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
			union:   "GEOMETRYCOLLECTION(POINT(3 1),POLYGON((0 0,0 2,2 2,2 1,2 0,0 0)))",
			inter:   "POINT(1 1)",
			fwdDiff: "POLYGON((0 0,0 2,2 2,2 1,2 0,0 0))",
			revDiff: "POINT(3 1)",
			symDiff: "GEOMETRYCOLLECTION(POINT(3 1),POLYGON((0 0,0 2,2 2,2 1,2 0,0 0)))",
			relate:  "0F2FF10F2",
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
			relate:  "FF20F1FF2",
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
			relate:  "FF20F1FF2",
		},
		{
			/*
			   +-------+
			   |       |
			   |   +   |
			   |       |
			   +-------+
			*/
			input1:  "LINESTRING(0 0,0 1,1 1,1 0,0 0,0 1)", // overlapping line segment
			input2:  "POINT(0.5 0.5)",
			union:   "GEOMETRYCOLLECTION(LINESTRING(0 0,0 1),LINESTRING(0 1,1 1,1 0,0 0),POINT(0.5 0.5))",
			inter:   "GEOMETRYCOLLECTION EMPTY",
			fwdDiff: "MULTILINESTRING((0 0,0 1),(0 1,1 1,1 0,0 0))",
			revDiff: "POINT(0.5 0.5)",
			symDiff: "GEOMETRYCOLLECTION(LINESTRING(0 0,0 1),LINESTRING(0 1,1 1,1 0,0 0),POINT(0.5 0.5))",
			relate:  "FF1FF00F2",
		},
		{
			/*
			       +
			      /
			     *
			    /
			   +
			*/
			input1:  "LINESTRING(0 0,1 1)",
			input2:  "POINT(0.35355339059327373 0.35355339059327373)",
			union:   "MULTILINESTRING((0 0,0.35355339059327373 0.35355339059327373),(0.35355339059327373 0.35355339059327373,1 1))",
			inter:   "POINT(0.35355339059327373 0.35355339059327373)",
			fwdDiff: "MULTILINESTRING((0 0,0.35355339059327373 0.35355339059327373),(0.35355339059327373 0.35355339059327373,1 1))",
			revDiff: "GEOMETRYCOLLECTION EMPTY",
			symDiff: "MULTILINESTRING((0 0,0.35355339059327373 0.35355339059327373),(0.35355339059327373 0.35355339059327373,1 1))",
			relate:  "0F1FF0FF2",
		},
		{
			// LineString with a Point in the middle of it.
			input1:  "POINT(5 5)",
			input2:  "LINESTRING(1 2,9 8)",
			union:   "MULTILINESTRING((1 2,5 5),(5 5,9 8))",
			inter:   "POINT(5 5)",
			fwdDiff: "GEOMETRYCOLLECTION EMPTY",
			revDiff: "MULTILINESTRING((1 2,5 5),(5 5,9 8))",
			symDiff: "MULTILINESTRING((1 2,5 5),(5 5,9 8))",
			relate:  "0FFFFF102",
		},
		{
			/*
			       *
			   +  /
			    \/
			    /\
			   *  *
			*/

			// Tests a case where intersection between two segments is *not* commutative if done naively.
			input1:  "LINESTRING(0 0,1 2)",
			input2:  "LINESTRING(0 1,1 0)",
			union:   "MULTILINESTRING((0 0,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 2),(0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0))",
			inter:   "POINT(0.3333333333 0.6666666667)",
			fwdDiff: "MULTILINESTRING((0 0,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 2))",
			revDiff: "MULTILINESTRING((0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0))",
			symDiff: "MULTILINESTRING((0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0),(0 0,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 2))",
			relate:  "0F1FF0102",
		},
		{
			// Similar case for when line segment non-commutative operations are
			// done, but this time with a line segment doubling back on itself.
			input1:  "LINESTRING(0 0,1 2,0 0)",
			input2:  "LINESTRING(0 1,1 0)",
			union:   "MULTILINESTRING((0 0,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 2),(0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0))",
			inter:   "POINT(0.3333333333 0.6666666667)",
			fwdDiff: "MULTILINESTRING((0 0,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 2))",
			revDiff: "MULTILINESTRING((0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0))",
			symDiff: "MULTILINESTRING((0 1,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 0),(0 0,0.3333333333 0.6666666667),(0.3333333333 0.6666666667,1 2))",
			relate:  "0F1FFF102",
		},

		// In the following test cases, lines from the first input intersect
		// *almost* exactly with one of the vertices in the second input.
		{
			input1:  "LINESTRING(-1 1,1 -1)",
			input2:  "POLYGON((-1 0,-0.070710678118655 0.070710678118655,0 1,-1 0))",
			union:   "GEOMETRYCOLLECTION(LINESTRING(-1 1,-0.5 0.5),LINESTRING(-0.070710678118655 0.070710678118655,1 -1),POLYGON((-1 0,-0.5 0.5,0 1,-0.070710678118655 0.070710678118655,-1 0)))",
			inter:   "LINESTRING(-0.5 0.5,-0.070710678118655 0.070710678118655)",
			fwdDiff: "MULTILINESTRING((-1 1,-0.5 0.5),(-0.070710678118655 0.070710678118655,1 -1))",
			revDiff: "POLYGON((-1 0,-0.5 0.5,0 1,-0.070710678118655 0.070710678118655,-1 0))",
			symDiff: "GEOMETRYCOLLECTION(LINESTRING(-1 1,-0.5 0.5),LINESTRING(-0.070710678118655 0.070710678118655,1 -1),POLYGON((-1 0,-0.5 0.5,0 1,-0.070710678118655 0.070710678118655,-1 0)))",
			relate:  "101FF0212",
		},
		{
			input1:  "LINESTRING(0 0,1 1)",
			input2:  "LINESTRING(1 0,0.5000000000000001 0.5,0 1)",
			union:   "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1),(1 0,0.5 0.5),(0.5 0.5,0 1))",
			inter:   "POINT(0.5 0.5)",
			fwdDiff: "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1))",
			revDiff: "MULTILINESTRING((1 0,0.5 0.5),(0.5 0.5,0 1))",
			symDiff: "MULTILINESTRING((1 0,0.5 0.5),(0.5 0.5,0 1),(0 0,0.5 0.5),(0.5 0.5,1 1))",
			relate:  "0F1FF0102",
		},
		{
			/*
			     +               +
			     |\              |\
			     | \             | \
			  +--+--+--+  ->  +--+  +--+
			     |   \           |   \
			     |    \          |    \
			     +-----+         +-----+
			*/
			input1:  "GEOMETRYCOLLECTION(POLYGON((1 0,3 2,1 2,1 0)))",
			input2:  "GEOMETRYCOLLECTION(LINESTRING(0 1,3 1))",
			union:   "GEOMETRYCOLLECTION(POLYGON((1 0,2 1,3 2,1 2,1 1,1 0)),LINESTRING(0 1,1 1),LINESTRING(2 1,3 1))",
			inter:   "LINESTRING(1 1,2 1)",
			fwdDiff: "POLYGON((1 0,2 1,3 2,1 2,1 1,1 0))",
			revDiff: "MULTILINESTRING((0 1,1 1),(2 1,3 1))",
			symDiff: "GEOMETRYCOLLECTION(POLYGON((1 0,2 1,3 2,1 2,1 1,1 0)),LINESTRING(0 1,1 1),LINESTRING(2 1,3 1))",
			relate:  "1F20F1102",
		},
		{
			/*
			    Reproduces a bug with set ops between self-intersecting GeometryCollections.
			        +  +
			        |\ |
			        | \|
			     +  |  +
			     |\ |  |\
			     | \|  | \
			     |  +  |  \
			     |  |\ |   \
			     |  | \|    \
			  +--+--+--+-----+--+1B
			     |  |  |\     \
			     |  |  | \  2A \
			     |  +--+--+-----+
			     |     |   \
			     | 1A  |    \
			     +-----+-----+
			           |
			           |2B
			           +
			*/
			input1: `GEOMETRYCOLLECTION(
				POLYGON((1 1,5 5,1 5,1 1)),
				LINESTRING(0 3,6 3))`,
			input2: `GEOMETRYCOLLECTION(
				POLYGON((2 0,6 4,2 4,2 0)),
				LINESTRING(3 0,3 6))`,
			union: `GEOMETRYCOLLECTION(
				POLYGON((2 2,2 0,3 1,5 3,6 4,4 4,5 5,3 5,1 5,1 3,1 1,2 2)),
				LINESTRING(0 3,1 3),
				LINESTRING(5 3,6 3),
				LINESTRING(3 0,3 1),
				LINESTRING(3 5,3 6))`,
			inter: `GEOMETRYCOLLECTION(
				POLYGON((2 2,3 3,4 4,3 4,2 4,2 3,2 2)),
				LINESTRING(3 3,5 3),
				LINESTRING(3 4,3 5))`,
			fwdDiff: `GEOMETRYCOLLECTION(
				POLYGON((1 1,2 2,2 3,2 4,3 4,4 4,5 5,3 5,1 5,1 3,1 1)),
				LINESTRING(0 3,1 3),
				LINESTRING(5 3,6 3))`,
			revDiff: `GEOMETRYCOLLECTION(
				POLYGON((3 1,5 3,6 4,4 4,3 3,2 2,2 0,3 1)),
				LINESTRING(3 0,3 1),
				LINESTRING(3 5,3 6))`,
			symDiff: `GEOMETRYCOLLECTION(
				POLYGON((1 1,2 2,2 3,2 4,3 4,4 4,5 5,3 5,1 5,1 3,1 1)),
				POLYGON((3 1,5 3,6 4,4 4,3 3,2 2,2 0,3 1)),
				LINESTRING(0 3,1 3),
				LINESTRING(5 3,6 3),
				LINESTRING(3 0,3 1),
				LINESTRING(3 5,3 6))`,
			relate: `212101212`,
		},
		{
			/*
			    Reproduces a bug with set ops between self-intersecting GeometryCollections.
			    Similar to the previous case, but none of the crossing points are coincident.
			        +  +
			        |\ |
			        | \|
			     +  |  +
			     |\ |  |\
			     | \|  | \
			     |  +  |  \
			     |  |\ |   \
			     |  | \|    \
			     |  |  +     \
			     |  |  |\     \
			     |  |  | \     \
			  +--+--+--+--+--+--+--+1B
			     |  |  |   \     \
			     |  |  |    \  2A \
			     |  +--+-----+-----+
			     |     |      \
			     | 1A  |       \
			     +-----+--------+
			           |
			           |2B
			           +
			*/
			input1: `GEOMETRYCOLLECTION(
				POLYGON((1 1,6 6,1 6,1 1)),
				LINESTRING(0 4,7 4))`,
			input2: `GEOMETRYCOLLECTION(
				POLYGON((2 0,7 5,2 5,2 0)),
				LINESTRING(3 0,3 7))`,
			union: `GEOMETRYCOLLECTION(
				POLYGON((2 2,2 0,3 1,6 4,7 5,5 5,6 6,3 6,1 6,1 4,1 1,2 2)),
				LINESTRING(0 4,1 4),
				LINESTRING(6 4,7 4),
				LINESTRING(3 0,3 1),
				LINESTRING(3 6,3 7))`,
			inter: `GEOMETRYCOLLECTION(
				POLYGON((2 2,3 3,4 4,5 5,3 5,2 5,2 4,2 2)),
				LINESTRING(4 4,6 4),
				LINESTRING(3 5,3 6))`,
			fwdDiff: `GEOMETRYCOLLECTION(
				POLYGON((5 5,6 6,3 6,1 6,1 4,1 1,2 2,2 4,2 5,3 5,5 5)),
				LINESTRING(0 4,1 4),
				LINESTRING(6 4,7 4))`,
			revDiff: `GEOMETRYCOLLECTION(
				POLYGON((2 0,3 1,6 4,7 5,5 5,4 4,3 3,2 2,2 0)),
				LINESTRING(3 0,3 1),
				LINESTRING(3 6,3 7))`,
			symDiff: `GEOMETRYCOLLECTION(
				POLYGON((3 6,1 6,1 4,1 1,2 2,2 4,2 5,3 5,5 5,6 6,3 6)),
				POLYGON((3 3,2 2,2 0,3 1,6 4,7 5,5 5,4 4,3 3)),
				LINESTRING(0 4,1 4),
				LINESTRING(6 4,7 4),
				LINESTRING(3 0,3 1),
				LINESTRING(3 6,3 7))`,
			relate: `212101212`,
		},
		{
			/*
				+-----+--+      +-----+--+
				| 1A  |2 |      |        |
				|  +--+--+      |        +
				|  |  |  |  ->  |        |
				+--+--+  |      +--+     |
				   |  1B |         |     |
				   +--+--+         +--+--+
			*/
			input1:  "GEOMETRYCOLLECTION(POLYGON((0 0,2 0,2 2,0 2,0 0)),POLYGON((1 1,3 1,3 3,1 3,1 1)))",
			input2:  "POLYGON((2 0,3 0,3 1,2 1,2 0))",
			union:   "POLYGON((2 0,3 0,3 1,3 3,1 3,1 2,0 2,0 0,2 0))",
			inter:   "MULTILINESTRING((2 1,3 1),(2 0,2 1))",
			fwdDiff: "POLYGON((1 2,0 2,0 0,2 0,2 1,3 1,3 3,1 3,1 2))",
			revDiff: "POLYGON((2 0,3 0,3 1,2 1,2 0))",
			symDiff: "POLYGON((0 0,2 0,3 0,3 1,3 3,1 3,1 2,0 2,0 0))",
			relate:  "FF2F11212",
		},
		{
			/*
				      +--------+                  +--------+
				      |        |                  |        |
				      |   1A   |                  |        |
				      |        |                  |        |
				+-----+--+  +--+-----+      +-----+        +-----+
				|     |  |  |  |     |      |                    |
				|     +--+--+--+     |      |        +--+        |
				|  2A    |  |    2B  |  ->  |        |  |        |
				|     +--+--+--+     |      |        +--+        |
				|     |  |  |  |     |      |                    |
				+-----+--+  +--+-----+      +-----+        +-----+
				      |        |                  |        |
				      |   1B   |                  |        |
				      |        |                  |        |
				      +--------+                  +--------+
			*/
			input1: `GEOMETRYCOLLECTION(
				POLYGON((2 0,5 0,5 3,2 3,2 0)),
				POLYGON((2 4,5 4,5 7,2 7,2 4)))`,
			input2: `GEOMETRYCOLLECTION(
				POLYGON((0 2,3 2,3 5,0 5,0 2)),
				POLYGON((4 2,7 2,7 5,4 5,4 2)))`,
			union: `POLYGON(
				(0 2,2 2,2 0,5 0,5 2,7 2,7 5,5 5,5 7,2 7,2 5,0 5,0 2),
				(3 3,3 4,4 4,4 3,3 3))`,
			inter: `MULTIPOLYGON(
				((2 2,3 2,3 3,2 3,2 2)),
				((2 4,3 4,3 5,2 5,2 4)),
				((4 2,5 2,5 3,4 3,4 2)),
				((4 4,5 4,5 5,4 5,4 4)))`,
			fwdDiff: `MULTIPOLYGON(
				((2 0,5 0,5 2,4 2,4 3,3 3,3 2,2 2,2 0)),
				((3 4,4 4,4 5,5 5,5 7,2 7,2 5,3 5,3 4)))`,
			revDiff: `MULTIPOLYGON(
				((0 2,2 2,2 3,3 3,3 4,2 4,2 5,0 5,0 2)),
				((5 2,7 2,7 5,5 5,5 4,4 4,4 3,5 3,5 2)))`,
			symDiff: `MULTIPOLYGON(
				((2 0,5 0,5 2,4 2,4 3,3 3,3 2,2 2,2 0)),
				((2 2,2 3,3 3,3 4,2 4,2 5,0 5,0 2,2 2)),
				((3 4,4 4,4 5,5 5,5 7,2 7,2 5,3 5,3 4)),
				((4 3,5 3,5 2,7 2,7 5,5 5,5 4,4 4,4 3)))`,
			relate: "212101212",
		},

		// Empty cases for relate.
		{input1: "POINT EMPTY", input2: "POINT(0 0)", relate: "FFFFFF0F2"},
		{input1: "POINT EMPTY", input2: "LINESTRING(0 0,1 1)", relate: "FFFFFF102"},
		{input1: "POINT EMPTY", input2: "LINESTRING(0 0,0 1,1 0,0 0)", relate: "FFFFFF1F2"},
		{input1: "POINT EMPTY", input2: "POLYGON((0 0,0 1,1 0,0 0))", relate: "FFFFFF212"},

		// Cases involving geometry collections where polygons from one of the
		// inputs interact with each other.
		{
			input1: `GEOMETRYCOLLECTION(
						POLYGON((0 0,1 0,0 1,0 0)),
						POLYGON((0 0,1 1,0 1,0 0)))`,
			input2:  "LINESTRING(0 0,1 1)",
			union:   "POLYGON((0 0,1 0,0.5 0.5,1 1,0 1,0 0))",
			inter:   "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1))",
			fwdDiff: "POLYGON((0 0,1 0,0.5 0.5,1 1,0 1,0 0))",
			revDiff: "GEOMETRYCOLLECTION EMPTY",
			symDiff: "POLYGON((0 0,1 0,0.5 0.5,1 1,0 1,0 0))",
			relate:  "1F2101FF2",
		},
		{
			input1: `GEOMETRYCOLLECTION(
						POLYGON((0 0,1 0,0 1,0 0)),
						POLYGON((1 1,0 1,1 0,1 1)))`,
			input2:  "POLYGON((0 0,2 0,2 2,0 2,0 0))",
			union:   "POLYGON((0 0,1 0,2 0,2 2,0 2,0 1,0 0))",
			inter:   "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			fwdDiff: "GEOMETRYCOLLECTION EMPTY",
			revDiff: "POLYGON((1 0,2 0,2 2,0 2,0 1,1 1,1 0))",
			symDiff: "POLYGON((1 0,2 0,2 2,0 2,0 1,1 1,1 0))",
			relate:  "2FF11F212",
		},
		{
			input1: `GEOMETRYCOLLECTION(
						POLYGON((0 0,2 0,2 1,0 1,0 0)),
						POLYGON((0 0,1 0,1 2,0 2,0 0)))`,
			input2:  "POLYGON((1 0,2 1,1 2,0 1,1 0))",
			union:   "POLYGON((0 0,1 0,2 0,2 1,1 2,0 2,0 1,0 0))",
			inter:   "POLYGON((1 0,2 1,1 1,1 2,0 1,1 0))",
			fwdDiff: "MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,2 1,1 0)),((0 1,1 2,0 2,0 1)))",
			revDiff: "POLYGON((1 1,2 1,1 2,1 1))",
			symDiff: "MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,2 1,1 0)),((0 1,1 2,0 2,0 1)),((1 1,2 1,1 2,1 1)))",
			relate:  "212101212",
		},

		// Bug reproductions:
		{
			input1:  "LINESTRING(-1 1,1 -1)",
			input2:  "MULTILINESTRING((1 0,0 1),(0 1,1 2),(2 0,3 1),(3 1,2 2))",
			union:   "MULTILINESTRING((-1 1,1 -1),(1 0,0 1),(0 1,1 2),(2 0,3 1),(3 1,2 2))",
			inter:   "GEOMETRYCOLLECTION EMPTY",
			fwdDiff: "LINESTRING(-1 1,1 -1)",
			revDiff: "MULTILINESTRING((1 0,0 1),(0 1,1 2),(2 0,3 1),(3 1,2 2))",
			symDiff: "MULTILINESTRING((1 0,0 1),(0 1,1 2),(2 0,3 1),(3 1,2 2),(-1 1,1 -1))",
			relate:  "FF1FF0102",
		},
		{
			input1:  "LINESTRING(0 1,1 0)",
			input2:  "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))",
			union:   "MULTIPOLYGON(((0 0,0 1,1 1,1 0.5,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))",
			inter:   "LINESTRING(0 1,1 0)",
			fwdDiff: "GEOMETRYCOLLECTION EMPTY",
			revDiff: "MULTIPOLYGON(((0 0,0 1,1 1,1 0.5,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))",
			symDiff: "MULTIPOLYGON(((0 0,0 1,1 1,1 0.5,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))",
			relate:  "1FFF0F212",
		},
		{
			input1:  "POLYGON((1 0,0 1,1 1,1 0))",
			input2:  "POLYGON((2 0,2 1,3 1,3 0,2 0))",
			union:   "MULTIPOLYGON(((1 0,0 1,1 1,1 0)),((2 0,2 1,3 1,3 0,2 0)))",
			inter:   "GEOMETRYCOLLECTION EMPTY",
			fwdDiff: "POLYGON((1 0,0 1,1 1,1 0))",
			revDiff: "POLYGON((2 0,2 1,3 1,3 0,2 0))",
			symDiff: "MULTIPOLYGON(((2 0,2 1,3 1,3 0,2 0)),((1 0,0 1,1 1,1 0)))",
			relate:  "FF2FF1212",
		},
		{
			input1:  "POLYGON((0 0,1 1,1 0,0 0))",
			input2:  "POLYGON((2 2,3 2,3 1,2 1,2 2))",
			union:   "MULTIPOLYGON(((0 0,1 0,1 1,0 0)),((2 1,2 2,3 2,3 1,2 1)))",
			inter:   "GEOMETRYCOLLECTION EMPTY",
			fwdDiff: "POLYGON((0 0,1 1,1 0,0 0))",
			revDiff: "POLYGON((2 1,2 2,3 2,3 1,2 1))",
			symDiff: "MULTIPOLYGON(((2 1,2 2,3 2,3 1,2 1)),((0 0,1 0,1 1,0 0)))",
			relate:  "FF2FF1212",
		},
		{
			input1:  "LINESTRING(0 1,1 0)",
			input2:  "MULTIPOLYGON(((1 1,1 0,0 0,0 1,1 1)),((2 1,2 2,3 2,3 1,2 1)))",
			union:   "MULTIPOLYGON(((1 1,1 0,0 0,0 1,1 1)),((2 1,2 2,3 2,3 1,2 1)))",
			inter:   "LINESTRING(0 1,1 0)",
			fwdDiff: "GEOMETRYCOLLECTION EMPTY",
			revDiff: "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 1,2 2,3 2,3 1,2 1)))",
			symDiff: "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 1,2 2,3 2,3 1,2 1)))",
			relate:  "1FFF0F212",
		},
		{
			input1:  "POINT(5 5)",
			input2:  "LINESTRING(5 3,4 8,1 2,9 8)",
			fwdDiff: "GEOMETRYCOLLECTION EMPTY",
			relate:  "0FFFFF102",
		},
		{
			input1: "LINESTRING(1 1,2 2,3 3,0 0)",
			input2: "LINESTRING(1 2,2 0)",
			inter:  "POINT(1.3333333333 1.3333333333)",
			relate: "0F1FF0102",
		},
		{
			input1: "MULTILINESTRING((0 0,1 1),(0 1,1 0))",
			input2: "LINESTRING(0 1,0.3333333333 0.6666666667,1 0)",
			union:  "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1),(0 1,0.3333333333 0.6666666667,0.5 0.5),(0.5 0.5,1 0))",
			relate: "1F1F00FF2",
		},
		{
			input1: "POLYGON((-1 0,0 0,0 1,-1 0))",
			input2: "POLYGON((1 0,-0.9 -0.2,-1 -0.0000000000000032310891488651735,-0.9 0.2,1 0))",
			union:  "POLYGON((-1 0,-0.9 0.2,-0.80952380952381 0.19047619047619,0 1,0 0.105263157894737,1 0,-0.9 -0.2,-1 0))",
			relate: "212101212",
		},
		{
			input1: "LINESTRING(1 2.1,2.1 1)",
			input2: "POLYGON((0 0,0 10,10 10,10 0,0 0),(1.5 1.5,8.5 1.5,8.5 8.5,1.5 8.5,1.5 1.5))",
			inter:  "MULTILINESTRING((1 2.1,1.5 1.6),(1.6 1.5,2.1 1))",
			relate: "1010FF212",
		},
		{
			input1: "LINESTRING(1 2,2 3)",
			input2: "MULTIPOLYGON(((1 1,1 0,0 0,0 1,1 1)),((1 2,2 2,2 3,1 3,1 2)))",
			union:  "MULTIPOLYGON(((1 1,1 0,0 0,0 1,1 1)),((1 2,2 2,2 3,1 3,1 2)))",
			relate: "1FFF0F212",
		},
		{
			input1: "LINESTRING(0 1,0 0,1 0)",
			input2: "POLYGON((0 0,1 0,1 1,0 1,0 0.5,0 0))",
			union:  "POLYGON((0 0,1 0,1 1,0 1,0 0.5,0 0))",
			relate: "F1FF0F212",
		},
		{
			input1:  "LINESTRING(2 2,3 3,4 4,5 5,0 0)",
			input2:  "LINESTRING(0 0,1 1)",
			fwdDiff: "MULTILINESTRING((2 2,3 3,4 4,5 5),(1 1,2 2))",
			relate:  "101F00FF2",
		},
		{
			input1:  "LINESTRING(0 0,0 0,0 1,1 0,0 0)",
			input2:  "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1),(0 1,0.3333333333 0.6666666667,0.5 0.5),(0.5 0.5,1 0))",
			fwdDiff: "MULTILINESTRING((0 0,0 1),(1 0,0 0))",
			relate:  "101FFF102",
		},
		{
			input1: "LINESTRING(1 0,0.5000000000000001 0.5,0 1)",
			input2: "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0),(0.5 0.5,1 0.5,1 1.5,0.5 1.5,0.5 0.5)))",
			union:  "POLYGON((0 0,1 0,2 0,2 2,0 2,0 1,0 0),(0.5000000000000001 0.5,1 0.5,1 1.5,0.5 1.5,0.5000000000000001 0.5))",
			relate: "10FF0F212",
		},
		{
			input1: "LINESTRING(1 1,3 1,1 1,3 1)",
			input2: "POLYGON((0 0,0 2,2 2,2 0,0 0))",
			relate: "1010F0212",
		},
		{
			input1: "LINESTRING(-1 1,1 -1)",
			input2: "MULTILINESTRING((0 0,0 1),(0 0,1 0))",
			relate: "0F1FF0102",
		},
		{
			input1: "MULTILINESTRING((2 0,2 1),(2 2,2 3))",
			input2: "POLYGON((0 0,0 10,10 10,10 0,0 0),(1.5 1.5,8.5 1.5,8.5 8.5,1.5 8.5,1.5 1.5))",
			union:  "GEOMETRYCOLLECTION(POLYGON((2 0,10 0,10 10,0 10,0 0,2 0),(1.5 1.5,1.5 8.5,8.5 8.5,8.5 1.5,1.5 1.5)),LINESTRING(2 2,2 3))",
		},
		{
			input1: "POINT(0 0)",
			input2: "POINT(0 0)",
			relate: "0FFFFFFF2",
			union:  "POINT(0 0)",
		},
		{
			input1: "GEOMETRYCOLLECTION(POINT(0 0))",
			input2: "GEOMETRYCOLLECTION(LINESTRING(2 0,2 1))",
			union:  "GEOMETRYCOLLECTION(POINT(0 0),LINESTRING(2 0,2 1))",
		},
		{
			input1: "GEOMETRYCOLLECTION(POLYGON((0 0,1 0,0 1,0 0)),POLYGON((0 0,1 1,0 1,0 0)))",
			input2: "POINT(0 0)",
			union:  "POLYGON((0 0,1 0,0.5 0.5,1 1,0 1,0 0))",
		},
		{
			input1: "GEOMETRYCOLLECTION(POLYGON((0 0,0 1,1 0,0 0)),POLYGON((0 1,1 1,1 0,0 1)))",
			input2: "POLYGON EMPTY",
			union:  "POLYGON((0 0,0 1,1 1,1 0,0 0))",
		},
		{
			input1: "LINESTRING(0 0,0 0,0 1,1 0,0 0)",
			input2: "LINESTRING(0.1 0.1,0.5 0.5)",
			inter:  "POINT(0.5 0.5)",
		},

		// Reproduces "no rings to extract" DCEL errors (reported in
		// https://github.com/peterstace/simplefeatures/issues/569).
		{
			input1: "POLYGON((-57.84764391579377 -14.00436771429812, -57.98105430423379 -13.978568346975345, -57.97219 -13.895754, -57.815573 -13.870471, -57.78975494169227 -13.97408746357712, -57.79567678742665 -14.003207561112367, -57.84764391579377 -14.00436771429812))",
			input2: "POLYGON((-57.97219 -13.895754, -57.815573 -13.870471, -57.782572 -14.002915, -57.984142 -14.007415, -57.97219 -13.895754))",
			inter:  "POLYGON((-57.84764391579377 -14.00436771429812, -57.98105430423379 -13.978568346975345, -57.97219 -13.895754, -57.815573 -13.870471, -57.78975494169227 -13.97408746357712, -57.79567678742665 -14.003207561112367, -57.84764391579377 -14.00436771429812))",
		},
		{
			input1: "POLYGON((-91.090505 33.966621, -91.094941 33.966624, -91.09491 33.96729, -91.094691 33.968384, -91.094602 33.968744, -91.094547 33.968945, -91.094484 33.969145, -91.093264 33.972456, -91.093108 33.97274, -91.092382 33.973979, -89.942235 35.721107, -89.941594 35.721928, -89.940438 35.723405, -89.720717 36, -89.711573 36, -89.645271 35.924821, -89.644942 35.924442, -89.644529 35.923925, -89.6429 35.921751, -89.642692 35.921465, -89.642576 35.921135, -89.642146 35.919717, -89.642026 35.91928, -89.641571 35.917498, -89.641166 35.91565, -89.63955 35.907509, -89.639384 35.906472, -89.639338 35.905496, -89.639356 35.904841, -89.639394 35.903992, -89.63944 35.90339, -89.639487 35.902831, -89.639559 35.902218, -89.640275 35.896772, -89.64057 35.894942, -89.640962 35.893237, -89.641113 35.892633, -89.641786 35.890644, -89.642306 35.889248, -89.642587 35.888566, -89.642808 35.888057, -89.643386 35.88681, -89.64378 35.885975, -90.060853 35.140433, -90.585556 34.404858, -90.888428 34.027973, -90.890265 34.026455, -90.890862 34.026091, -90.895918 34.023915, -90.896574 34.023654, -90.896965 34.023521, -91.090505 33.966621))",
			input2: "POLYGON((-90.19553150069916 34.95162878475482, -90.42127335893674 34.993424947208105, -90.30813100280166 35.16529356781885, -90.12850301040231 35.13253239680938, -90.16769780459812 34.99064851784101, -90.19553150069916 34.95162878475482))",
			inter:  "POLYGON((-90.19553150069916 34.95162878475482, -90.42127335893674 34.993424947208105, -90.30813100280166 35.16529356781885, -90.12850301040231 35.13253239680938, -90.16769780459812 34.99064851784101, -90.19553150069916 34.95162878475482))",
		},
		{
			input1: "POLYGON((-91.090505 33.966621, -91.094941 33.966624, -91.09491 33.96729, -91.094691 33.968384, -91.094602 33.968744, -91.094547 33.968945, -91.094484 33.969145, -91.093264 33.972456, -91.093108 33.97274, -91.092382 33.973979, -89.942235 35.721107, -89.941594 35.721928, -89.940438 35.723405, -89.720717 36, -89.711573 36, -89.645271 35.924821, -89.644942 35.924442, -89.644529 35.923925, -89.6429 35.921751, -89.642692 35.921465, -89.642576 35.921135, -89.642146 35.919717, -89.642026 35.91928, -89.641571 35.917498, -89.641166 35.91565, -89.63955 35.907509, -89.639384 35.906472, -89.639338 35.905496, -89.639356 35.904841, -89.639394 35.903992, -89.63944 35.90339, -89.639487 35.902831, -89.639559 35.902218, -89.640275 35.896772, -89.64057 35.894942, -89.640962 35.893237, -89.641113 35.892633, -89.641786 35.890644, -89.642306 35.889248, -89.642587 35.888566, -89.642808 35.888057, -89.643386 35.88681, -89.64378 35.885975, -90.060853 35.140433, -90.585556 34.404858, -90.888428 34.027973, -90.890265 34.026455, -90.890862 34.026091, -90.895918 34.023915, -90.896574 34.023654, -90.896965 34.023521, -91.090505 33.966621))",
			input2: "POLYGON((-90.29716937546225 35.18194480113967, -90.29596586203434 35.18172958540237, -90.34543219833212 34.998800268076835, -90.41002551098103 35.01051096325925, -90.29716937546225 35.18194480113967))",
			inter:  "POLYGON((-90.41002551098103 35.01051096325925,-90.40917779527051 35.01035727335111,-90.34543219833212 34.998800268076835,-90.29596586203434 35.18172958540237,-90.29716937546225 35.18194480113967,-90.41002551098103 35.01051096325925))",
		},
		{
			input1: "POLYGON((-149.845771 -17.472558, -149.888137 -17.477017, -149.929731 -17.480468, -149.934682 -17.50814, -149.920475 -17.541336, -149.895694 -17.571267, -149.861608 -17.600395, -149.832332 -17.611409, -149.791981 -17.611947, -149.774766 -17.577051, -149.753707 -17.535289, -149.744632 -17.494022, -149.765688 -17.465994, -149.805445 -17.46709, -149.845771 -17.472558))",
			input2: "POLYGON((-149.8839047803303 -17.58134141150439, -149.86106842049824 -17.474168045744268, -149.85203718167833 -17.473217512441664, -149.74468306149925 -17.494254193376246, -149.753707 -17.535289, -149.774766 -17.577051, -149.791981 -17.611947, -149.832332 -17.611409, -149.861608 -17.600395, -149.8839047803303 -17.58134141150439))",
			inter:  "POLYGON((-149.8839047803303 -17.58134141150439,-149.861608 -17.600395,-149.832332 -17.611409,-149.791981 -17.611947,-149.774766 -17.577051,-149.753707 -17.535289,-149.74468306149925 -17.494254193376246,-149.84639009954768 -17.47432409190779,-149.85203718167833 -17.473217512441664,-149.86106842049824 -17.474168045744268,-149.8839047803303 -17.58134141150439))",
		},

		// Reproduces a failed DCEL operation (reported in
		// https://github.com/peterstace/simplefeatures/issues/496).
		{
			input1: "POLYGON((-83.5825305152402 32.7316823944815,-83.58376293006216 32.73315376178507,-83.58504085655653 32.734597137036324,-83.58636334101533 32.73601156235818,-83.58772946946186 32.73739605462324,-83.58913829287843 32.738749652823024,-83.59058883640313 32.740071417020225,-83.59208009385833 32.74136043125909,-83.59361103333862 32.74261579844315,-83.59518059080796 32.743836647669376,-83.59678768034422 32.745022130852156,-83.59843118482598 32.746171426099316,-83.60010996734121 32.74728373328835,-83.60182286094272 32.74835828151699,-83.60356867749573 32.74939432363171,-83.6053462070958 32.75039113983657,-83.60715421364506 32.75134803897381,-83.60899144521312 32.75226435508957,-83.61085662379259 32.75313945319649,-83.61274845635847 32.75397272333648,-83.61466562799973 32.7547635884388,-83.61660680960262 32.75551149954699,-83.61857065109865 32.756215934654755,-83.62055578961366 32.75687640673859,-83.62256084632457 32.75749245578336,-83.62458442890414 32.75806365483586,-83.62662513245226 32.75858960587211,-83.62868153809899 32.759069943900975,-83.63075221766115 32.759504334926895,-83.6328357331767 32.759892478947584,-83.6349306369047 32.76023410294001,-83.63703547248946 32.76052897293848,-83.63914877915153 32.76077688192745,-83.64126908772954 32.760977657903275,-83.64339492533685 32.76113116287015,-83.64552481489585 32.76123728881648,-83.64765727560365 32.761295961756694,-83.64979082712301 32.761307141684846,-83.65192398562425 32.76127082060284,-83.65405527030447 32.761187023508135,-83.65618319896379 32.76105580840831,-83.65830629359328 32.76087726729329,-83.66042307921083 32.7606515241745,-83.66253208572374 32.760378736038,-83.66463184629896 32.76005909090909,-83.666720902951 32.75969281274174,-83.66879780351513 32.75928015456466,-83.67086110700251 32.758821403399814,-83.67290937847787 32.75831687924383,-83.67494119511315 32.75776693105163,-83.67695514665314 32.75717194084009,-83.67894983308719 32.756532323629834,-83.68092387070277 32.75584852243937,-83.68287588719608 32.755121013232745,-83.68480452667785 32.75435029997188,-83.6867084511868 32.75353691874934,-83.68858633871054 32.752681435518205,-83.69043688423329 32.75178444323992,-83.69225880474181 32.75084656496929,-83.69405083421947 32.74986845274852,-83.69581173063119 32.74885078644311,-83.69754027103403 32.74779427211197,-83.69923525541884 32.746699642822385,-83.70089550880579 32.74556766051264,-83.70251987926525 32.74439911017125,-83.70410724071183 32.74319480286376,-83.70565649104168 32.74195557654766,-83.70716655737118 32.740682290251605,-83.70863639277735 32.73937582791714,-83.71006497711272 32.738037097583764,-83.71145132142914 32.73666702917708,-83.71279446471812 32.735266572878935,-83.71409347705406 32.73383670052444,-83.71534745831372 32.73237840618403,-83.71655554062085 32.73089270080836,-83.71771688788061 32.729380616419256,-83.7188306961288 32.72784320203518,-83.71989619434675 32.7262815256503,-83.72091264550903 32.72469667027635,-83.72187934768948 32.72308973592171,-83.72279563289698 32.721461838543654,-83.72366086608594 32.71981410620665,-83.72447445225765 32.718147683855385,-83.72523582737989 32.71646372644627,-83.72594446851502 32.71476340406971,-83.72659988462323 32.713047893684596,-83.7272016266906 32.71131838726742,-83.72774927777567 32.70957608482731,-83.72824246383601 32.70782219347475,-83.72868084389114 32.70605793103041,-83.72906411790967 32.70428451962234,-83.7293920229094 32.7025031912739,-83.72966433589676 32.700715179871075,-83.7298808698651 32.69892172546988,-83.7300414788296 32.69712407208444,-83.73014605577549 32.69532346570795,-83.73019452970033 32.693521154312656,-83.73018687161033 32.691718388897606,-83.73012309050767 32.689916417551466,-83.7300032323846 32.68811649010913,-83.7298273852511 32.68631985477568,-83.72959567309675 32.684527756380156,-83.72930826092573 32.68274143695762,-83.72896535074773 32.680962133537285,-83.72856718354245 32.6791910811693,-83.72811403832833 32.67742950582342,-83.72760623312291 32.67567862846354,-83.72704412086836 32.67393966108957,-83.72642809662005 32.672213810695574,-83.72575858936847 32.670502271353506,-83.72503606605545 32.66880622898625,-83.72426102976982 32.66712685869006,-83.7234340194563 32.66546532333759,-83.72255561311698 32.66382277497487,-83.7216264197786 32.66220034969901,-83.72064708740885 32.66059917243128,-83.7196182950583 32.659020351096295,-83.71854075868117 32.65746497883392,-83.71741522729366 32.655934134581365,-83.71624248192614 32.65442887527333,-83.71502333853363 32.65295024597017,-83.71375864112719 32.65149926868198,-83.71244926771118 32.650076949469984,-83.7110961292355 32.64868427122874,-83.70970016179575 32.64732219892485,-83.70826233431666 32.64599167470751,-83.70678364482677 32.6446936215174,-83.70526511731526 32.64342893528683,-83.70370780580646 32.642198493088834,-83.70211278830624 32.64100314591942,-83.70048117076016 32.639843721724354,-83.5825305152402 32.7316823944815))",
			input2: "POLYGON((-83.7004774529266 32.6398466121988,-83.70047002035855 32.63878317849731,-83.70041586825457 32.63698593605796,-83.70032015514218 32.635189934722355,-83.7001829220105 32.633395861340894,-83.70000422287623 32.63160440392813,-83.69978412771954 32.62981624851955,-83.69952272555449 32.62803208010292,-83.69922011737137 32.62625258168692,-83.69887642118621 32.62447843430119,-83.69849177198259 32.62271031792762,-83.69806631976164 32.62094891056905,-83.69760022855255 32.61919488417471,-83.69709368034157 32.617448911858,-83.6965468711139 32.61571166254135,-83.6959600118723 32.61398380014135,-83.69533332962666 32.61226598682834,-83.69466706733576 32.610558879417574,-83.69396148204461 32.60886313204674,-83.69321684575763 32.607179393731265,-83.69243344543858 32.605508308364286,-83.69161158312701 32.60385051506592,-83.69075157582185 32.60220664969666,-83.68985375449185 32.600577340317166,-83.68891846413382 32.59896320998224,-83.68794606575167 32.59736487767214,-83.6869369323401 32.595782953403166,-83.68589145196958 32.59421804311706,-83.68481002662222 32.59267074483933,-83.68369307318127 32.59114165158963,-83.68254101772683 32.58963134730721,-83.68135430531467 32.588140409994146,-83.68013339089586 32.58666941066759,-83.67887874246 32.58521891044939,-83.6775908410352 32.58378946417498,-83.67627018255072 32.58238161992768,-83.67491727108492 32.58099591461497,-83.67353262759636 32.57963287932353,-83.672116782124 32.57829303408062,-83.67067027669756 32.576976891812144,-83.6691936661525 32.57568495659644,-83.66768751766423 32.57441772040466,-83.66615240620806 32.57317566915436,-83.66458891968134 32.57195927688873,-83.66299765913638 32.57076900868697,-83.66137923156268 32.5696053204315,-83.65973425803597 32.56846865624676,-83.65806336847956 32.56735945001265,-83.65636720096583 32.56627812687794,-83.65464640637285 32.56522509872003,-83.65290164081736 32.56420076950008,-83.65113357426955 32.56320553037359,-83.64934288275319 32.562239760214204,-83.64753025020829 32.561303830095774,-83.64569637058659 32.56039809694767,-83.64384194505757 32.55952290582491,-83.64196768247413 32.558678592702115,-83.64007430088597 32.55786547859453,-83.63816252230092 32.557083875495216,-83.63623307966984 32.55633408241693,-83.63428670803907 32.55561638434438,-83.63232415432932 32.55493105625056,-83.63034616674187 32.55427836018635,-83.62835350104498 32.553658545105925,-83.62634691952618 32.553071847041345,-83.62432718982811 32.55251849201296,-83.62229508215452 32.55199869096533,-83.62025137346129 32.551512640931385,-83.61819684489521 32.55106052890324,-83.61613228225971 32.55064252688454,-83.6140584727552 32.55025879587766,-83.61197620917002 32.54990948087789,-83.60988628662082 32.54959471587927,-83.60778950394953 32.54931462087689,-83.6056866623264 32.54906930390417,-83.60357856571565 32.54885885793329,-83.60146601715016 32.548683362970685,-83.59934982455226 32.548542887017454,-83.59723079793977 32.5484374830644,-83.59510974337238 32.548367191122004,-83.59298747366185 32.548332038187716,-83.59086479905883 32.54833203826062,-83.58874252934831 32.548367191340425,-83.58662147571223 32.54843748342834,-83.58450244816842 32.548542887526914,-83.58238625650185 32.54868336362566,-83.58027370793636 32.54885885873379,-83.57816561039428 32.549069304850185,-83.57606276877115 32.54931462196843,-83.57396598609986 32.54959471711633,-83.57187606448198 32.54990948226047,-83.5697938008968 32.55025879740575,-83.5677199913923 32.55064252855815,-83.56565542782548 32.551060530722374,-83.5636008992594 32.551512642896036,-83.56155719173032 32.5519986930755,-83.55952508405673 32.55251849423954,-83.55750535342733 32.55307184944255,-83.55549877190853 32.55365854762355,-83.5535061073758 32.55427836282039,-83.55152811978834 32.55493105905922,-83.54956556514728 32.55561638726946,-83.54761919456425 32.55633408551663,-83.54568975100185 32.55708387876954,-83.54377797346453 32.55786548198527,-83.54188459094505 32.55867859620927,-83.54001032940934 32.55952290944848,-83.53815590388032 32.56039810068766,-83.53632202437504 32.561303833952174,-83.53450939194656 32.56223976424523,-83.53271869949887 32.563205534521025,-83.53095063306748 32.564200773763936,-83.52920586855973 32.565225103100296,-83.52748507315184 32.566278131374624,-83.52578890668585 32.56735945462575,-83.52411801619812 32.56846866097627,-83.52247304278782 32.56960532527743,-83.52085461637827 32.5707690135329,-83.51926335501841 32.57195928185107,-83.51769986965584 32.573175674116705,-83.51616475831608 32.57441772548342,-83.5146586088965 32.575684961791616,-83.51318199858426 32.576976897123735,-83.51173549420557 32.57829303950863,-83.51031964780188 32.579632884867955,-83.50893500442973 32.58099592015939,-83.50758209412808 32.5823816254721,-83.5062614348287 32.583789469835814,-83.50497353456805 32.585218916226644,-83.50371888531728 32.58666941656126,-83.50249797101489 32.588140415887814,-83.50131125871914 32.58963135320088,-83.50015920442885 32.591141657599714,-83.499042250173 32.59267075096583,-83.49796082494206 32.594218049359974,-83.49691534573569 32.595782959762495,-83.49590621255695 32.597364884147886,-83.49493381434942 32.5989632165744,-83.49399852416602 32.600577346909326,-83.49310070301064 32.60220665628882,-83.49224069582189 32.60385052165808,-83.49141883368495 32.60550831507286,-83.49063543354052 32.60717940043984,-83.48989079742816 32.60886313887173,-83.48918521132211 32.610558886242565,-83.48851894920583 32.61226599365333,-83.48789226812434 32.613983807082754,-83.48730540905737 32.61571166948276,-83.48675859895658 32.61744891891582,-83.48625205092023 32.61919489134895,-83.48578595985666 32.620948917743284,-83.48536050781033 32.622710325101856,-83.48497585878133 32.624478441475425,-83.48463216376032 32.626252588861156,-83.48432955575183 32.628032087277155,-83.48406815277187 32.62981625569379,-83.48384805879388 32.631604411102366,-83.48366935983424 32.63339586851513,-83.48353212588765 32.63518994189659,-83.48343641394669 32.63698594334861,-83.48338226201733 32.638783185787965,-83.48336969309977 32.640580979221035,-83.48339871519394 32.642378636564594,-83.48346931829994 32.644175467941444,-83.48358148041709 32.64597078638477,-83.48373515855175 32.64776390178455,-83.4839302976977 32.6495541282217,-83.48416682484796 32.65134078059206,-83.48444465202166 32.653123171928804,-83.48476367520172 32.654900621318724,-83.48512377538256 32.656672445753124,-83.48552481557246 32.658437967112754,-83.48596664579938 32.660196507511195,-83.48644909702827 32.661947394882795,-83.48697198925194 32.663689955299255,-83.48753512252708 32.6654235226321,-83.48813828381364 32.66714743203344,-83.48878124406448 32.66886102133291,-83.4894637593313 32.67056363371527,-83.49018556960047 32.672254617021885,-83.490946399899 32.673933322353754,-83.49174596221539 32.675599105701316,-83.49258395055196 32.67725132806087,-83.49346004587252 32.67888935636589,-83.49437391423973 32.68051256174082,-83.49532520757177 32.68212032101443,-83.49631356294387 32.683712018349674,-83.49733860136595 32.68528704163479,-83.49839993377802 32.68684478888614,-83.49949715412342 32.68838466021558,-83.50062984249209 32.68990606656158,-83.50179756593545 32.69140842188941,-83.50299987939772 32.69289115215512,-83.50423632187423 32.694353686457966,-83.50550642130085 32.6957954636761,-83.50680969176003 32.697215930953156,-83.50814563324795 32.6986145432326,-83.5095137357491 32.699990763490554,-83.51091347620942 32.70134406378353,-83.51234431667373 32.70267392408429,-83.51380571115416 32.70397983332956,-83.51529709864528 32.705261290584225,-83.51681790812992 32.70651780480845,-83.51836755660017 32.70774889299506,-83.51994545115278 32.70895408121729,-83.52155098572959 32.71013290847046,-83.52318354623978 32.711284920618404,-83.52484250578684 32.71240967679629,-83.52652722932518 32.71350674393913,-83.52823706888714 32.71457570108911,-83.52997137045142 32.715616138231475,-83.53172946800598 32.71662765443185,-83.53351068657477 32.7176098626093,-83.53531434419685 32.71856238377374,-83.53713974773538 32.71948485389445,-83.53898619637013 32.720376918021124,-83.5408529819467 32.72123823313601,-83.54273938757953 32.72206846925989,-83.54464468916532 32.72286730636703,-83.54656815573227 32.72363443747023,-83.54850904932366 32.7243695685626,-83.55046662301876 32.72507241762802,-83.55244012558953 32.72574271370983,-83.5544287991722 32.72638019877346,-83.55643188078076 32.72698462986015,-83.55844860044428 32.72755577390641,-83.56047818400086 32.72809340995597,-83.56251985053652 32.72859733399097,-83.5645728170418 32.72906735002622,-83.56663629375518 32.72950327904406,-83.56870948928528 32.72990495206765,-83.57079160595433 32.73027221708768,-83.57288184468754 32.7306049300984,-83.57497940128789 32.73090296609886,-83.57708346992851 32.73116620809166,-83.57919324245424 32.73139455706556,-83.5813079090801 32.73158792303153,-83.58253417348143 32.73167954800889,-83.7004774529266 32.6398466121988))",
			union:  "POLYGON((-83.7004774529266 32.6398466121988,-83.70047002035855 32.63878317849731,-83.70041586825457 32.63698593605796,-83.70032015514218 32.635189934722355,-83.7001829220105 32.633395861340894,-83.70000422287623 32.63160440392813,-83.69978412771954 32.62981624851955,-83.69952272555449 32.62803208010292,-83.69922011737137 32.62625258168692,-83.69887642118621 32.62447843430119,-83.69849177198259 32.62271031792762,-83.69806631976164 32.62094891056905,-83.69760022855255 32.61919488417471,-83.69709368034157 32.617448911858,-83.6965468711139 32.61571166254135,-83.6959600118723 32.61398380014135,-83.69533332962666 32.61226598682834,-83.69466706733576 32.610558879417574,-83.69396148204461 32.60886313204674,-83.69321684575763 32.607179393731265,-83.69243344543858 32.605508308364286,-83.69161158312701 32.60385051506592,-83.69075157582185 32.60220664969666,-83.68985375449185 32.600577340317166,-83.68891846413382 32.59896320998224,-83.68794606575167 32.59736487767214,-83.6869369323401 32.595782953403166,-83.68589145196958 32.59421804311706,-83.68481002662222 32.59267074483933,-83.68369307318127 32.59114165158963,-83.68254101772683 32.58963134730721,-83.68135430531467 32.588140409994146,-83.68013339089586 32.58666941066759,-83.67887874246 32.58521891044939,-83.6775908410352 32.58378946417498,-83.67627018255072 32.58238161992768,-83.67491727108492 32.58099591461497,-83.67353262759636 32.57963287932353,-83.672116782124 32.57829303408062,-83.67067027669756 32.576976891812144,-83.6691936661525 32.57568495659644,-83.66768751766423 32.57441772040466,-83.66615240620806 32.57317566915436,-83.66458891968134 32.57195927688873,-83.66299765913638 32.57076900868697,-83.66137923156268 32.5696053204315,-83.65973425803597 32.56846865624676,-83.65806336847956 32.56735945001265,-83.65636720096583 32.56627812687794,-83.65464640637285 32.56522509872003,-83.65290164081736 32.56420076950008,-83.65113357426955 32.56320553037359,-83.64934288275319 32.562239760214204,-83.64753025020829 32.561303830095774,-83.64569637058659 32.56039809694767,-83.64384194505757 32.55952290582491,-83.64196768247413 32.558678592702115,-83.64007430088597 32.55786547859453,-83.63816252230092 32.557083875495216,-83.63623307966984 32.55633408241693,-83.63428670803907 32.55561638434438,-83.63232415432932 32.55493105625056,-83.63034616674187 32.55427836018635,-83.62835350104498 32.553658545105925,-83.62634691952618 32.553071847041345,-83.62432718982811 32.55251849201296,-83.62229508215452 32.55199869096533,-83.62025137346129 32.551512640931385,-83.61819684489521 32.55106052890324,-83.61613228225971 32.55064252688454,-83.6140584727552 32.55025879587766,-83.61197620917002 32.54990948087789,-83.60988628662082 32.54959471587927,-83.60778950394953 32.54931462087689,-83.6056866623264 32.54906930390417,-83.60357856571565 32.54885885793329,-83.60146601715016 32.548683362970685,-83.59934982455226 32.548542887017454,-83.59723079793977 32.5484374830644,-83.59510974337238 32.548367191122004,-83.59298747366185 32.548332038187716,-83.59086479905883 32.54833203826062,-83.58874252934831 32.548367191340425,-83.58662147571223 32.54843748342834,-83.58450244816842 32.548542887526914,-83.58238625650185 32.54868336362566,-83.58027370793636 32.54885885873379,-83.57816561039428 32.549069304850185,-83.57606276877115 32.54931462196843,-83.57396598609986 32.54959471711633,-83.57187606448198 32.54990948226047,-83.5697938008968 32.55025879740575,-83.5677199913923 32.55064252855815,-83.56565542782548 32.551060530722374,-83.5636008992594 32.551512642896036,-83.56155719173032 32.5519986930755,-83.55952508405673 32.55251849423954,-83.55750535342733 32.55307184944255,-83.55549877190853 32.55365854762355,-83.5535061073758 32.55427836282039,-83.55152811978834 32.55493105905922,-83.54956556514728 32.55561638726946,-83.54761919456425 32.55633408551663,-83.54568975100185 32.55708387876954,-83.54377797346453 32.55786548198527,-83.54188459094505 32.55867859620927,-83.54001032940934 32.55952290944848,-83.53815590388032 32.56039810068766,-83.53632202437504 32.561303833952174,-83.53450939194656 32.56223976424523,-83.53271869949887 32.563205534521025,-83.53095063306748 32.564200773763936,-83.52920586855973 32.565225103100296,-83.52748507315184 32.566278131374624,-83.52578890668585 32.56735945462575,-83.52411801619812 32.56846866097627,-83.52247304278782 32.56960532527743,-83.52085461637827 32.5707690135329,-83.51926335501841 32.57195928185107,-83.51769986965584 32.573175674116705,-83.51616475831608 32.57441772548342,-83.5146586088965 32.575684961791616,-83.51318199858426 32.576976897123735,-83.51173549420557 32.57829303950863,-83.51031964780188 32.579632884867955,-83.50893500442973 32.58099592015939,-83.50758209412808 32.5823816254721,-83.5062614348287 32.583789469835814,-83.50497353456805 32.585218916226644,-83.50371888531728 32.58666941656126,-83.50249797101489 32.588140415887814,-83.50131125871914 32.58963135320088,-83.50015920442885 32.591141657599714,-83.499042250173 32.59267075096583,-83.49796082494206 32.594218049359974,-83.49691534573569 32.595782959762495,-83.49590621255695 32.597364884147886,-83.49493381434942 32.5989632165744,-83.49399852416602 32.600577346909326,-83.49310070301064 32.60220665628882,-83.49224069582189 32.60385052165808,-83.49141883368495 32.60550831507286,-83.49063543354052 32.60717940043984,-83.48989079742816 32.60886313887173,-83.48918521132211 32.610558886242565,-83.48851894920583 32.61226599365333,-83.48789226812434 32.613983807082754,-83.48730540905737 32.61571166948276,-83.48675859895658 32.61744891891582,-83.48625205092023 32.61919489134895,-83.48578595985666 32.620948917743284,-83.48536050781033 32.622710325101856,-83.48497585878133 32.624478441475425,-83.48463216376032 32.626252588861156,-83.48432955575183 32.628032087277155,-83.48406815277187 32.62981625569379,-83.48384805879388 32.631604411102366,-83.48366935983424 32.63339586851513,-83.48353212588765 32.63518994189659,-83.48343641394669 32.63698594334861,-83.48338226201733 32.638783185787965,-83.48336969309977 32.640580979221035,-83.48339871519394 32.642378636564594,-83.48346931829994 32.644175467941444,-83.48358148041709 32.64597078638477,-83.48373515855175 32.64776390178455,-83.4839302976977 32.6495541282217,-83.48416682484796 32.65134078059206,-83.48444465202166 32.653123171928804,-83.48476367520172 32.654900621318724,-83.48512377538256 32.656672445753124,-83.48552481557246 32.658437967112754,-83.48596664579938 32.660196507511195,-83.48644909702827 32.661947394882795,-83.48697198925194 32.663689955299255,-83.48753512252708 32.6654235226321,-83.48813828381364 32.66714743203344,-83.48878124406448 32.66886102133291,-83.4894637593313 32.67056363371527,-83.49018556960047 32.672254617021885,-83.490946399899 32.673933322353754,-83.49174596221539 32.675599105701316,-83.49258395055196 32.67725132806087,-83.49346004587252 32.67888935636589,-83.49437391423973 32.68051256174082,-83.49532520757177 32.68212032101443,-83.49631356294387 32.683712018349674,-83.49733860136595 32.68528704163479,-83.49839993377802 32.68684478888614,-83.49949715412342 32.68838466021558,-83.50062984249209 32.68990606656158,-83.50179756593545 32.69140842188941,-83.50299987939772 32.69289115215512,-83.50423632187423 32.694353686457966,-83.50550642130085 32.6957954636761,-83.50680969176003 32.697215930953156,-83.50814563324795 32.6986145432326,-83.5095137357491 32.699990763490554,-83.51091347620942 32.70134406378353,-83.51234431667373 32.70267392408429,-83.51380571115416 32.70397983332956,-83.51529709864528 32.705261290584225,-83.51681790812992 32.70651780480845,-83.51836755660017 32.70774889299506,-83.51994545115278 32.70895408121729,-83.52155098572959 32.71013290847046,-83.52318354623978 32.711284920618404,-83.52484250578684 32.71240967679629,-83.52652722932518 32.71350674393913,-83.52823706888714 32.71457570108911,-83.52997137045142 32.715616138231475,-83.53172946800598 32.71662765443185,-83.53351068657477 32.7176098626093,-83.53531434419685 32.71856238377374,-83.53713974773538 32.71948485389445,-83.53898619637013 32.720376918021124,-83.5408529819467 32.72123823313601,-83.54273938757953 32.72206846925989,-83.54464468916532 32.72286730636703,-83.54656815573227 32.72363443747023,-83.54850904932366 32.7243695685626,-83.55046662301876 32.72507241762802,-83.55244012558953 32.72574271370983,-83.5544287991722 32.72638019877346,-83.55643188078076 32.72698462986015,-83.55844860044428 32.72755577390641,-83.56047818400086 32.72809340995597,-83.56251985053652 32.72859733399097,-83.5645728170418 32.72906735002622,-83.56663629375518 32.72950327904406,-83.56870948928528 32.72990495206765,-83.57079160595433 32.73027221708768,-83.57288184468754 32.7306049300984,-83.57497940128789 32.73090296609886,-83.57708346992851 32.73116620809166,-83.57919324245424 32.73139455706556,-83.5813079090801 32.73158792303153,-83.5825341712489 32.73167954784208,-83.5825305152402 32.7316823944815,-83.58376293006216 32.73315376178507,-83.58504085655653 32.734597137036324,-83.58636334101533 32.73601156235818,-83.58772946946186 32.73739605462324,-83.58913829287843 32.738749652823024,-83.59058883640313 32.740071417020225,-83.59208009385833 32.74136043125909,-83.59361103333862 32.74261579844315,-83.59518059080796 32.743836647669376,-83.59678768034422 32.745022130852156,-83.59843118482598 32.746171426099316,-83.60010996734121 32.74728373328835,-83.60182286094272 32.74835828151699,-83.60356867749573 32.74939432363171,-83.6053462070958 32.75039113983657,-83.60715421364506 32.75134803897381,-83.60899144521312 32.75226435508957,-83.61085662379259 32.75313945319649,-83.61274845635847 32.75397272333648,-83.61466562799973 32.7547635884388,-83.61660680960262 32.75551149954699,-83.61857065109865 32.756215934654755,-83.62055578961366 32.75687640673859,-83.62256084632457 32.75749245578336,-83.62458442890414 32.75806365483586,-83.62662513245226 32.75858960587211,-83.62868153809899 32.759069943900975,-83.63075221766115 32.759504334926895,-83.6328357331767 32.759892478947584,-83.6349306369047 32.76023410294001,-83.63703547248946 32.76052897293848,-83.63914877915153 32.76077688192745,-83.64126908772954 32.760977657903275,-83.64339492533685 32.76113116287015,-83.64552481489585 32.76123728881648,-83.64765727560365 32.761295961756694,-83.64979082712301 32.761307141684846,-83.65192398562425 32.76127082060284,-83.65405527030447 32.761187023508135,-83.65618319896379 32.76105580840831,-83.65830629359328 32.76087726729329,-83.66042307921083 32.7606515241745,-83.66253208572374 32.760378736038,-83.66463184629896 32.76005909090909,-83.666720902951 32.75969281274174,-83.66879780351513 32.75928015456466,-83.67086110700251 32.758821403399814,-83.67290937847787 32.75831687924383,-83.67494119511315 32.75776693105163,-83.67695514665314 32.75717194084009,-83.67894983308719 32.756532323629834,-83.68092387070277 32.75584852243937,-83.68287588719608 32.755121013232745,-83.68480452667785 32.75435029997188,-83.6867084511868 32.75353691874934,-83.68858633871054 32.752681435518205,-83.69043688423329 32.75178444323992,-83.69225880474181 32.75084656496929,-83.69405083421947 32.74986845274852,-83.69581173063119 32.74885078644311,-83.69754027103403 32.74779427211197,-83.69923525541884 32.746699642822385,-83.70089550880579 32.74556766051264,-83.70251987926525 32.74439911017125,-83.70410724071183 32.74319480286376,-83.70565649104168 32.74195557654766,-83.70716655737118 32.740682290251605,-83.70863639277735 32.73937582791714,-83.71006497711272 32.738037097583764,-83.71145132142914 32.73666702917708,-83.71279446471812 32.735266572878935,-83.71409347705406 32.73383670052444,-83.71534745831372 32.73237840618403,-83.71655554062085 32.73089270080836,-83.71771688788061 32.729380616419256,-83.7188306961288 32.72784320203518,-83.71989619434675 32.7262815256503,-83.72091264550903 32.72469667027635,-83.72187934768948 32.72308973592171,-83.72279563289698 32.721461838543654,-83.72366086608594 32.71981410620665,-83.72447445225765 32.718147683855385,-83.72523582737989 32.71646372644627,-83.72594446851502 32.71476340406971,-83.72659988462323 32.713047893684596,-83.7272016266906 32.71131838726742,-83.72774927777567 32.70957608482731,-83.72824246383601 32.70782219347475,-83.72868084389114 32.70605793103041,-83.72906411790967 32.70428451962234,-83.7293920229094 32.7025031912739,-83.72966433589676 32.700715179871075,-83.7298808698651 32.69892172546988,-83.7300414788296 32.69712407208444,-83.73014605577549 32.69532346570795,-83.73019452970033 32.693521154312656,-83.73018687161033 32.691718388897606,-83.73012309050767 32.689916417551466,-83.7300032323846 32.68811649010913,-83.7298273852511 32.68631985477568,-83.72959567309675 32.684527756380156,-83.72930826092573 32.68274143695762,-83.72896535074773 32.680962133537285,-83.72856718354245 32.6791910811693,-83.72811403832833 32.67742950582342,-83.72760623312291 32.67567862846354,-83.72704412086836 32.67393966108957,-83.72642809662005 32.672213810695574,-83.72575858936847 32.670502271353506,-83.72503606605545 32.66880622898625,-83.72426102976982 32.66712685869006,-83.7234340194563 32.66546532333759,-83.72255561311698 32.66382277497487,-83.7216264197786 32.66220034969901,-83.72064708740885 32.66059917243128,-83.7196182950583 32.659020351096295,-83.71854075868117 32.65746497883392,-83.71741522729366 32.655934134581365,-83.71624248192614 32.65442887527333,-83.71502333853363 32.65295024597017,-83.71375864112719 32.65149926868198,-83.71244926771118 32.650076949469984,-83.7110961292355 32.64868427122874,-83.70970016179575 32.64732219892485,-83.70826233431666 32.64599167470751,-83.70678364482677 32.6446936215174,-83.70526511731526 32.64342893528683,-83.70370780580646 32.642198493088834,-83.70211278830624 32.64100314591942,-83.70048117076016 32.639843721724354,-83.61872797135857 32.7034983516507,-83.7004774529266 32.6398466121988))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g1 := geomFromWKT(t, geomCase.input1)
			g2 := geomFromWKT(t, geomCase.input2)
			t.Logf("input1: %s", geomCase.input1)
			t.Logf("input2: %s", geomCase.input2)
			for _, opCase := range []struct {
				opName string
				op     func(geom.Geometry, geom.Geometry) (geom.Geometry, error)
				want   string
			}{
				{"union", geom.Union, geomCase.union},
				{"inter", geom.Intersection, geomCase.inter},
				{"fwd_diff", geom.Difference, geomCase.fwdDiff},
				{"rev_diff", func(a, b geom.Geometry) (geom.Geometry, error) { return geom.Difference(b, a) }, geomCase.revDiff},
				{"sym_diff", geom.SymmetricDifference, geomCase.symDiff},
			} {
				t.Run(opCase.opName, func(t *testing.T) {
					if opCase.want == "" {
						t.Skip("Skipping test because it's not specified or is commented out")
					}
					want := geomFromWKT(t, opCase.want)
					got, err := opCase.op(g1, g2)
					if err != nil {
						t.Fatalf("could not perform op: %v", err)
					}
					expectGeomEq(t, got, want, geom.IgnoreOrder, geom.ToleranceXY(1e-7))
				})
			}
			t.Run("relate", func(t *testing.T) {
				if geomCase.relate == "" {
					t.Skip("Skipping test because it's not specified or is commented out")
				}
				for _, swap := range []struct {
					description string
					reverse     bool
				}{
					{"fwd", false},
					{"rev", true},
				} {
					t.Run(swap.description, func(t *testing.T) {
						var (
							got string
							err error
						)
						if swap.reverse {
							got, err = geom.Relate(g2, g1)
						} else {
							got, err = geom.Relate(g1, g2)
						}
						if err != nil {
							t.Fatal("could not perform relate op")
						}

						want := geomCase.relate
						if swap.reverse {
							want = ""
							for j := 0; j < 9; j++ {
								k := 3*(j%3) + j/3
								want += geomCase.relate[k : k+1]
							}
						}
						if got != want {
							t.Errorf("\nwant: %v\ngot:  %v\n", want, got)
						}
					})
				}
			})
		})
	}
}

func TestBinaryOpNoCrash(t *testing.T) {
	for i, tc := range []struct {
		inputA, inputB string
	}{
		// Reproduces a node set crash.
		{
			"MULTIPOLYGON(((-73.85559633603238 40.65821829792369,-73.8555908203125 40.6580545462853,-73.85559350252151 40.65822190464714,-73.85559616790695 40.65821836616974,-73.85559633603238 40.65821829792369)),((-73.83276962411851 40.670198784066336,-73.83329428732395 40.66733238233316,-73.83007764816284 40.668112089039745,-73.83276962411851 40.670198784066336),(-73.83250952988594 40.66826467245589,-73.83246950805187 40.66828298244238,-73.83250169456005 40.66826467245589,-73.83250952988594 40.66826467245589),(-73.83128821933425 40.66879546275945,-73.83135303854942 40.668798203056376,-73.83129335939884 40.668798711663115,-73.83128821933425 40.66879546275945)),((-73.82322192192078 40.6723059714534,-73.8232085108757 40.67231004009312,-73.82320448756218 40.67231410873261,-73.82322192192078 40.6723059714534)))",
			"POLYGON((-73.84494431798483 40.65179671514794,-73.84493172168732 40.651798908464365,-73.84487807750702 40.651802469618836,-73.84494431798483 40.65179671514794))",
		},
		{
			"LINESTRING(0 0,1 0,0 1,0 0)",
			"POLYGON((1 0,0.9807852804032305 -0.19509032201612808,0.923879532511287 -0.3826834323650894,0.8314696123025456 -0.5555702330196017,0.7071067811865481 -0.7071067811865469,0.5555702330196031 -0.8314696123025447,0.38268343236509084 -0.9238795325112863,0.19509032201612964 -0.9807852804032302,0.0000000000000016155445744325867 -1,-0.19509032201612647 -0.9807852804032308,-0.38268343236508784 -0.9238795325112875,-0.5555702330196005 -0.8314696123025463,-0.7071067811865459 -0.7071067811865491,-0.8314696123025438 -0.5555702330196043,-0.9238795325112857 -0.38268343236509234,-0.9807852804032299 -0.19509032201613122,-1 -0.0000000000000032310891488651735,-0.9807852804032311 0.19509032201612486,-0.9238795325112882 0.38268343236508634,-0.8314696123025475 0.555570233019599,-0.7071067811865505 0.7071067811865446,-0.5555702330196058 0.8314696123025428,-0.3826834323650936 0.9238795325112852,-0.19509032201613213 0.9807852804032297,-0.000000000000003736410698672604 1,0.1950903220161248 0.9807852804032311,0.38268343236508673 0.9238795325112881,0.5555702330195996 0.8314696123025469,0.7071067811865455 0.7071067811865496,0.8314696123025438 0.5555702330196044,0.9238795325112859 0.38268343236509206,0.98078528040323 0.19509032201613047,1 0))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gA, err := geom.UnmarshalWKT(tc.inputA)
			expectNoErr(t, err)
			gB, err := geom.UnmarshalWKT(tc.inputB)
			expectNoErr(t, err)

			for _, op := range []struct {
				name string
				op   func(_, _ geom.Geometry) (geom.Geometry, error)
			}{
				{"union", geom.Union},
				{"intersection", geom.Intersection},
				{"difference", geom.Difference},
				{"symmetric_difference", geom.SymmetricDifference},
			} {
				t.Run(op.name, func(t *testing.T) {
					if _, err := op.op(gA, gB); err != nil {
						t.Errorf("unexpected error: %v", err)
					}
				})
			}
		})
	}
}

func TestBinaryOpBothInputsEmpty(t *testing.T) {
	for i, wkt := range []string{
		"POINT EMPTY",
		"MULTIPOINT EMPTY",
		"MULTIPOINT(EMPTY)",
		"LINESTRING EMPTY",
		"MULTILINESTRING EMPTY",
		"MULTILINESTRING(EMPTY)",
		"POLYGON EMPTY",
		"MULTIPOLYGON EMPTY",
		"MULTIPOLYGON(EMPTY)",
		"GEOMETRYCOLLECTION EMPTY",
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, wkt)
			for _, opCase := range []struct {
				opName string
				op     func(geom.Geometry, geom.Geometry) (geom.Geometry, error)
			}{
				{"union", geom.Union},
				{"inter", geom.Intersection},
				{"fwd_diff", geom.Difference},
				{"sym_diff", geom.SymmetricDifference},
			} {
				t.Run(opCase.opName, func(t *testing.T) {
					got, err := opCase.op(g, g)
					if err != nil {
						t.Fatalf("could not perform op: %v", err)
					}
					want := geom.Geometry{}
					if opCase.opName == "union" {
						want = got
					}
					expectGeomEq(t, got, want, geom.IgnoreOrder)
				})
			}
			t.Run("relate", func(t *testing.T) {
				got, err := geom.Relate(g, g)
				if err != nil {
					t.Fatal("could not perform relate op")
				}
				if got != "FFFFFFFF2" {
					t.Errorf("got=%v but want=FFFFFFFF2", got)
				}
			})
		})
	}
}

func reverseArgs(fn func(_, _ geom.Geometry) (geom.Geometry, error)) func(_, _ geom.Geometry) (geom.Geometry, error) {
	return func(a, b geom.Geometry) (geom.Geometry, error) {
		return fn(b, a)
	}
}

func TestBinaryOpOneInputEmpty(t *testing.T) {
	for _, opCase := range []struct {
		opName    string
		op        func(geom.Geometry, geom.Geometry) (geom.Geometry, error)
		wantEmpty bool
	}{
		{"fwd_union", geom.Union, false},
		{"rev_union", reverseArgs(geom.Union), false},
		{"fwd_inter", geom.Intersection, true},
		{"rev_inter", reverseArgs(geom.Intersection), true},
		{"fwd_diff", geom.Difference, false},
		{"rev_diff", reverseArgs(geom.Difference), true},
		{"fwd_sym_diff", geom.SymmetricDifference, false},
		{"rev_sym_diff", reverseArgs(geom.SymmetricDifference), false},
	} {
		t.Run(opCase.opName, func(t *testing.T) {
			poly := geomFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")
			empty := geom.Polygon{}.AsGeometry()
			got, err := opCase.op(poly, empty)
			expectNoErr(t, err)
			if opCase.wantEmpty {
				expectTrue(t, got.IsEmpty())
			} else {
				expectGeomEq(t, got, poly, geom.IgnoreOrder)
			}
		})
	}
}

func TestUnaryUnionAndUnionMany(t *testing.T) {
	for i, tc := range []struct {
		inputWKTs []string
		wantWKT   string
	}{
		{
			inputWKTs: nil,
			wantWKT:   `GEOMETRYCOLLECTION EMPTY`,
		},
		{
			inputWKTs: []string{"POINT(1 2)"},
			wantWKT:   "POINT(1 2)",
		},
		{
			inputWKTs: []string{"MULTIPOINT((1 2),(3 4))"},
			wantWKT:   "MULTIPOINT((1 2),(3 4))",
		},
		{
			inputWKTs: []string{"LINESTRING(1 2,3 4)"},
			wantWKT:   "LINESTRING(1 2,3 4)",
		},
		{
			inputWKTs: []string{"MULTILINESTRING((0 1,2 1),(1 0,1 2))"},
			wantWKT:   "MULTILINESTRING((0 1,1 1),(2 1,1 1),(1 0,1 1),(1 2,1 1))",
		},
		{
			inputWKTs: []string{"POLYGON((0 0,0 1,1 0,0 0))"},
			wantWKT:   "POLYGON((0 0,0 1,1 0,0 0))",
		},
		{
			inputWKTs: []string{"MULTIPOLYGON(((1 1,1 0,0 1,1 1)),((1 1,2 1,1 2,1 1)))"},
			wantWKT:   "MULTIPOLYGON(((1 1,1 0,0 1,1 1)),((1 1,2 1,1 2,1 1)))",
		},
		{
			inputWKTs: []string{"GEOMETRYCOLLECTION(POLYGON((0 0,0 1,1 0,0 0)))"},
			wantWKT:   "POLYGON((0 0,0 1,1 0,0 0))",
		},
		{
			inputWKTs: []string{"POINT(2 2)", "POINT(2 2)"},
			wantWKT:   "POINT(2 2)",
		},
		{
			inputWKTs: []string{"MULTIPOINT(1 2,2 2)", "MULTIPOINT(2 2,1 2)"},
			wantWKT:   "MULTIPOINT(1 2,2 2)",
		},
		{
			inputWKTs: []string{"LINESTRING(0 0,0 1,1 1)", "LINESTRING(1 1,0 1,0 0)"},
			wantWKT:   "LINESTRING(0 0,0 1,1 1)",
		},
		{
			inputWKTs: []string{"MULTILINESTRING((0 0,0 1,1 1),(2 2,3 3))", "MULTILINESTRING((1 1,0 1,0 0),(2 2,3 3))"},
			wantWKT:   "MULTILINESTRING((2 2,3 3),(0 0,0 1,1 1))",
		},
		{
			inputWKTs: []string{"POLYGON((0 0,0 1,1 0,0 0))", "POLYGON((0 0,0 1,1 0,0 0))"},
			wantWKT:   "POLYGON((0 0,0 1,1 0,0 0))",
		},
		{
			inputWKTs: []string{
				"MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((1 1,1 2,2 2,2 1,1 1)))",
				"MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((1 1,1 2,2 2,2 1,1 1)))",
			},
			wantWKT: "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((1 1,1 2,2 2,2 1,1 1)))",
		},
		{
			inputWKTs: []string{"POINT(1 2)", "POINT(2 3)", "POINT(3 4)"},
			wantWKT:   "MULTIPOINT(1 2,2 3,3 4)",
		},
		{
			inputWKTs: []string{"MULTIPOINT(1 2,2 3)", "MULTIPOINT(2 3,3 4)", "MULTIPOINT(3 4,4 5)"},
			wantWKT:   "MULTIPOINT(1 2,2 3,3 4,4 5)",
		},
		{
			inputWKTs: []string{"LINESTRING(0 0,0 1,1 1)", "LINESTRING(0 1,1 1,1 0)", "LINESTRING(2 1,2 2,1 2)"},
			wantWKT:   "MULTILINESTRING((0 0,0 1),(0 1,1 1),(1 1,1 0),(2 1,2 2,1 2))",
		},
		{
			inputWKTs: []string{"MULTILINESTRING((0 0,0 1,1 1),(0 1,1 1,1 0))", "LINESTRING(2 1,2 2,1 2)"},
			wantWKT:   "MULTILINESTRING((0 0,0 1),(0 1,1 1),(1 1,1 0),(2 1,2 2,1 2))",
		},
		{
			inputWKTs: []string{
				"POLYGON((0 0,0 1,1 1,1 0,0 0))",
				"POLYGON((1 0,1 1,2 1,2 0,1 0))",
				"POLYGON((1 1,1 2,2 2,2 1,1 1))",
			},
			wantWKT: "POLYGON((0 0,1 0,2 0,2 1,2 2,1 2,1 1,0 1,0 0))",
		},
		{
			inputWKTs: []string{
				"POLYGON((0 0,0 1,2 1,2 0,0 0))",
				"POLYGON((1 0,1 2,2 2,2 0,1 0))",
				"POLYGON((1 0,1 1,2 1,2 0,1 0))",
			},
			wantWKT: "POLYGON((0 0,1 0,2 0,2 1,2 2,1 2,1 1,0 1,0 0))",
		},
		{
			inputWKTs: []string{
				"POLYGON((0 2,1 0,2 2,1 1,0 2))",
				"POLYGON((0 2,1 3,2 2,1 4,0 2))",
				"POLYGON((0 1.5,2 1.5,2 2.5,0 2.5,0 1.5))",
			},
			wantWKT: `POLYGON(
				(1 0,1.75 1.5,2 1.5,2 2,2 2.5,1.75 2.5,1 4,0.25 2.5,0 2.5,0 2,0 1.5,0.25 1.5,1 0),
				(0.5 1.5,1.5 1.5,1 1,0.5 1.5),
				(0.5 2.5,1.5 2.5,1 3,0.5 2.5))`,
		},
		{
			inputWKTs: []string{
				"MULTIPOLYGON(((1 0,2 0,2 1,1 1,1 0)),((3 0,4 0,4 1,3 1,3 0)))",
				"MULTIPOLYGON(((3 0,4 0,4 -1,3 -1,3 0)),((4 0,5 0,5 1,4 1,4 0)))",
				"MULTIPOLYGON(((1 0,1 1,0 1,0 0,1 0)),((5 0,6 0,6 -1,5 -1,5 0)))",
			},
			wantWKT: `MULTIPOLYGON(
				((0 0,1 0,2 0,2 1,1 1,0 1,0 0)),
				((5 0,6 0,6 -1,5 -1,5 0)),
				((4 0,5 0,5 1,4 1,3 1,3 0,3 -1,4 -1,4 0)))`,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var inputs []geom.Geometry
			for _, wkt := range tc.inputWKTs {
				inputs = append(inputs, geomFromWKT(t, wkt))
			}
			t.Run("UnionMany", func(t *testing.T) {
				got, err := geom.UnionMany(inputs)
				expectNoErr(t, err)
				expectGeomEqWKT(t, got, tc.wantWKT, geom.IgnoreOrder)
			})
			t.Run("UnaryUnion", func(t *testing.T) {
				got, err := geom.UnaryUnion(geom.NewGeometryCollection(inputs).AsGeometry())
				expectNoErr(t, err)
				expectGeomEqWKT(t, got, tc.wantWKT, geom.IgnoreOrder)
			})
		})
	}
}

func TestBinaryOpOutputOrdering(t *testing.T) {
	for i, tc := range []struct {
		wkt string
	}{
		{"MULTIPOINT(1 2,2 3)"},
		{"MULTILINESTRING((1 2,2 3),(3 4,4 5))"},
		{"POLYGON((0 0,0 4,4 4,4 0,0 0),(1 1,1 2,2 2,2 1,1 1),(2 2,2 3,3 3,3 2,2 2))"},
		{"MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((1 1,1 2,2 2,2 1,1 1)))"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			in := geomFromWKT(t, tc.wkt)
			got1, err := geom.Union(in, in)
			expectNoErr(t, err)
			got2, err := geom.Union(in, in)
			expectNoErr(t, err)
			// Ensure ordering is stable over multiple executions:
			expectGeomEq(t, got1, got2)
		})
	}
}

func TestNoPanic(t *testing.T) {
	for i, tc := range []struct {
		input1 string
		input2 string
		op     func(_, _ geom.Geometry) (geom.Geometry, error)
	}{
		{
			input1: `POLYGON((
				-83.58253051 32.73168239,
				-83.59843118 32.74617142,
				-83.70048117 32.63984372,
				-83.58253051 32.73168239
			))`,
			input2: `POLYGON((
				-83.70047745 32.63984661,
				-83.68891846 32.59896320,
				-83.58253417 32.73167955,
				-83.70047745 32.63984661
			))`,
			op: geom.Union,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g1 := geomFromWKT(t, tc.input1)
			g2 := geomFromWKT(t, tc.input2)
			// Used to panic before a bug fix was put in place.
			_, _ = tc.op(g1, g2)
		})
	}
}
