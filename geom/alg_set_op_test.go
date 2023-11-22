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

		// Reported in https://github.com/peterstace/simplefeatures/issues/573
		{
			input1: "MULTIPOLYGON(((144.14852129205005 -36.60216120134051,144.1498935327475 -36.88032819802394,144.15040812300904 -36.887324318786426,144.1526380141424 -36.887324318786426,144.15280954422957 -36.88801017847503,144.15503943536294 -36.88787300703022,144.15538249553728 -36.88938187936745,144.15778391675784 -36.88910754116068,144.1582985070194 -36.89102788790482,144.16138604858864 -36.89089072188227,144.16121451850145 -36.89418263839662,144.16344440963482 -36.894319798503766,144.16361593972198 -36.895279912352436,144.16653195120406 -36.89541707048774,144.2114728340455 -36.89637717053346,144.2114728340455 -36.895142753970624,144.2126735446558 -36.89500559534236,144.2126735446558 -36.89637717053346,144.2152464959635 -36.89651432669692,144.31902219870804 -36.89884594376332,144.32193821019013 -36.89884594376332,144.32210974027728 -36.89829733439166,144.32228127036447 -36.89884594376332,144.32519728184656 -36.89898309549001,144.50976365565342 -36.90309753267125,144.53223409707417 -36.87429181406224,144.53309174751004 -36.87223384684748,144.53309174751004 -36.86962700878087,144.53480704838185 -36.86948980431315,144.53497857846904 -36.87045023041224,144.53566469881775 -36.86990141697698,144.59655787976703 -36.79165520168818,144.58455077366435 -36.61193608998413,144.5713429569514 -36.61166075796511,144.57117142686425 -36.61262441573086,144.57031377642832 -36.612762080142936,144.56756929503345 -36.61234908616945,144.5670547047719 -36.61152309158697,144.55213158718715 -36.61124775809338,144.55196005709996 -36.61179842409749,144.54904404561788 -36.61166075796511,144.54870098544353 -36.61207375562499,144.54750027483325 -36.61179842409749,144.54767180492044 -36.61111009097795,144.54578497396145 -36.61097242361676,144.52863196524333 -36.61097242361676,144.51542414853037 -36.61028408312449,144.50667611408414 -36.61055942005868,144.5063330539098 -36.61138542496305,144.50358857251487 -36.61124775809338,144.50307398225334 -36.61179842409749,144.5018732716431 -36.61152309158697,144.5018732716431 -36.61083475600983,144.4963843088533 -36.60987107588005,144.39432390698053 -36.60753065971684,144.39243707602154 -36.60753065971684,144.39243707602154 -36.608632040875996,144.39140789549847 -36.60876971241501,144.39140789549847 -36.60794367949463,144.38763423358048 -36.60753065971684,144.3864335229702 -36.60849436909124,144.38626199288302 -36.607806006481134,144.3821452707907 -36.60725531196955,144.35795952849816 -36.60670461352599,144.35298515596992 -36.60711763772727,144.35281362588273 -36.60876971241501,144.35041220466218 -36.608632040875996,144.350240674575 -36.609320396113546,144.34938302413912 -36.60904505475579,144.34955455422627 -36.60725531196955,144.34372253126213 -36.60642926282975,144.30958804391307 -36.605603204843085,144.3068435625182 -36.605603204843085,144.306672032431 -36.60629158711301,144.3058143819951 -36.60629158711301,144.30564285190792 -36.605603204843085,144.2987816484207 -36.60546552765187,144.2986101183335 -36.60670461352599,144.2975809378104 -36.606842288505504,144.29740940772322 -36.60532785021491,144.2648186911588 -36.60463945934398,144.26464716107165 -36.605603204843085,144.26361798054856 -36.60574088178856,144.263103390287 -36.60463945934398,144.24286283999965 -36.604088742223894,144.2416621293894 -36.604088742223894,144.2414905993022 -36.60505249460376,144.23908917808166 -36.60491481642958,144.23874611790728 -36.60725531196955,144.2378884674714 -36.60697996323927,144.23840305773294 -36.604088742223894,144.23685928694832 -36.603951062329536,144.1768237564349 -36.602711932221936,144.1768237564349 -36.604501780432564,144.17630916617338 -36.60463945934398,144.17545151573748 -36.60367570180357,144.17270703434258 -36.603400340294655,144.17270703434258 -36.60257424987018,144.15332413449113 -36.60216120134051,144.15315260440394 -36.603813382189415,144.15229495396804 -36.603951062329536,144.1500650628347 -36.60367570180357,144.15040812300904 -36.60216120134051,144.14852129205005 -36.60216120134051),(144.1498935327475 -36.603951062329536,144.1500650628347 -36.60670461352599,144.1492074123988 -36.60711763772727,144.1490358823116 -36.60697996323927,144.1498935327475 -36.603951062329536)))",
			input2: "MULTIPOLYGON(((144.14852129205005 -36.60216120134051,144.1498935327475 -36.88032819802394,144.15040812300904 -36.887324318786426,144.1526380141424 -36.887324318786426,144.15280954422957 -36.88801017847503,144.15503943536294 -36.88787300703022,144.15538249553728 -36.88938187936745,144.15778391675784 -36.88910754116068,144.1582985070194 -36.89102788790482,144.16138604858864 -36.89089072188227,144.16121451850145 -36.89418263839662,144.16344440963482 -36.894319798503766,144.16361593972198 -36.895279912352436,144.16653195120406 -36.89541707048774,144.2114728340455 -36.89637717053346,144.2114728340455 -36.895142753970624,144.2126735446558 -36.89500559534236,144.2126735446558 -36.89637717053346,144.2152464959635 -36.89651432669692,144.31902219870804 -36.89884594376332,144.32193821019013 -36.89884594376332,144.32210974027728 -36.89829733439166,144.32228127036447 -36.89884594376332,144.32519728184656 -36.89898309549001,144.50976365565342 -36.90309753267125,144.53223409707417 -36.87429181406224,144.53309174751004 -36.87223384684748,144.53309174751004 -36.86962700878087,144.53480704838185 -36.86948980431315,144.53497857846904 -36.87045023041224,144.53566469881775 -36.86990141697698,144.59655787976703 -36.79165520168818,144.58455077366435 -36.61193608998413,144.5713429569514 -36.61166075796511,144.57117142686425 -36.61262441573086,144.57031377642832 -36.612762080142936,144.56756929503345 -36.61234908616945,144.5670547047719 -36.61152309158697,144.55213158718715 -36.61124775809338,144.55196005709996 -36.61179842409749,144.54904404561788 -36.61166075796511,144.54870098544353 -36.61207375562499,144.54750027483325 -36.61179842409749,144.54767180492044 -36.61111009097795,144.54578497396145 -36.61097242361676,144.52863196524333 -36.61097242361676,144.51542414853037 -36.61028408312449,144.50667611408414 -36.61055942005868,144.5063330539098 -36.61138542496305,144.50358857251487 -36.61124775809338,144.50307398225334 -36.61179842409749,144.5018732716431 -36.61152309158697,144.5018732716431 -36.61083475600983,144.4963843088533 -36.60987107588005,144.39432390698053 -36.60753065971684,144.39243707602154 -36.60753065971684,144.39243707602154 -36.608632040875996,144.39140789549847 -36.60876971241501,144.39140789549847 -36.60794367949463,144.38763423358048 -36.60753065971684,144.3864335229702 -36.60849436909124,144.38626199288302 -36.607806006481134,144.3821452707907 -36.60725531196955,144.35795952849816 -36.60670461352599,144.35298515596992 -36.60711763772727,144.35281362588273 -36.60876971241501,144.35041220466218 -36.608632040875996,144.350240674575 -36.609320396113546,144.34938302413912 -36.60904505475579,144.34955455422627 -36.60725531196955,144.34372253126213 -36.60642926282975,144.30958804391307 -36.605603204843085,144.3068435625182 -36.605603204843085,144.306672032431 -36.60629158711301,144.3058143819951 -36.60629158711301,144.30564285190792 -36.605603204843085,144.2987816484207 -36.60546552765187,144.2986101183335 -36.60670461352599,144.2975809378104 -36.606842288505504,144.29740940772322 -36.60532785021491,144.2648186911588 -36.60463945934398,144.26464716107165 -36.605603204843085,144.26361798054856 -36.60574088178856,144.263103390287 -36.60463945934398,144.24286283999965 -36.604088742223894,144.2416621293894 -36.604088742223894,144.2414905993022 -36.60505249460376,144.23908917808166 -36.60491481642958,144.23874611790728 -36.60725531196955,144.2378884674714 -36.60697996323927,144.23840305773294 -36.604088742223894,144.23685928694832 -36.603951062329536,144.1768237564349 -36.602711932221936,144.1768237564349 -36.604501780432564,144.17630916617338 -36.60463945934398,144.17545151573748 -36.60367570180357,144.17270703434258 -36.603400340294655,144.17270703434258 -36.60257424987018,144.15332413449113 -36.60216120134051,144.15315260440394 -36.603813382189415,144.15229495396804 -36.603951062329536,144.1500650628347 -36.60367570180357,144.15040812300904 -36.60216120134051,144.14852129205005 -36.60216120134051),(144.1498935327475 -36.603951062329536,144.1500650628347 -36.60670461352599,144.1492074123988 -36.60711763772727,144.1490358823116 -36.60697996323927,144.1498935327475 -36.603951062329536)))",
			union:  "POLYGON((144.14852129205005 -36.60216120134051,144.1498935327475 -36.88032819802394,144.15040812300904 -36.887324318786426,144.1526380141424 -36.887324318786426,144.15280954422957 -36.88801017847503,144.15503943536294 -36.88787300703022,144.15538249553728 -36.88938187936745,144.15778391675784 -36.88910754116068,144.1582985070194 -36.89102788790482,144.16138604858864 -36.89089072188227,144.16121451850145 -36.89418263839662,144.16344440963482 -36.894319798503766,144.16361593972198 -36.895279912352436,144.16653195120406 -36.89541707048774,144.2114728340455 -36.89637717053346,144.2114728340455 -36.895142753970624,144.2126735446558 -36.89500559534236,144.2126735446558 -36.89637717053346,144.2152464959635 -36.89651432669692,144.31902219870804 -36.89884594376332,144.32193821019013 -36.89884594376332,144.32210974027728 -36.89829733439166,144.32228127036447 -36.89884594376332,144.32519728184656 -36.89898309549001,144.50976365565342 -36.90309753267125,144.53223409707417 -36.87429181406224,144.53309174751004 -36.87223384684748,144.53309174751004 -36.86962700878087,144.53480704838185 -36.86948980431315,144.53497857846904 -36.87045023041224,144.53566469881775 -36.86990141697698,144.59655787976703 -36.79165520168818,144.58455077366435 -36.61193608998413,144.5713429569514 -36.61166075796511,144.57117142686425 -36.61262441573086,144.57031377642832 -36.612762080142936,144.56756929503345 -36.61234908616945,144.5670547047719 -36.61152309158697,144.55213158718715 -36.61124775809338,144.55196005709996 -36.61179842409749,144.54904404561788 -36.61166075796511,144.54870098544353 -36.61207375562499,144.54750027483325 -36.61179842409749,144.54767180492044 -36.61111009097795,144.54578497396145 -36.61097242361676,144.52863196524333 -36.61097242361676,144.51542414853037 -36.61028408312449,144.50667611408414 -36.61055942005868,144.5063330539098 -36.61138542496305,144.50358857251487 -36.61124775809338,144.50307398225334 -36.61179842409749,144.5018732716431 -36.61152309158697,144.5018732716431 -36.61083475600983,144.4963843088533 -36.60987107588005,144.39432390698053 -36.60753065971684,144.39243707602154 -36.60753065971684,144.39243707602154 -36.608632040875996,144.39140789549847 -36.60876971241501,144.39140789549847 -36.60794367949463,144.38763423358048 -36.60753065971684,144.3864335229702 -36.60849436909124,144.38626199288302 -36.607806006481134,144.3821452707907 -36.60725531196955,144.35795952849816 -36.60670461352599,144.35298515596992 -36.60711763772727,144.35281362588273 -36.60876971241501,144.35041220466218 -36.608632040875996,144.350240674575 -36.609320396113546,144.34938302413912 -36.60904505475579,144.34955455422627 -36.60725531196955,144.34372253126213 -36.60642926282975,144.30958804391307 -36.605603204843085,144.3068435625182 -36.605603204843085,144.306672032431 -36.60629158711301,144.3058143819951 -36.60629158711301,144.30564285190792 -36.605603204843085,144.2987816484207 -36.60546552765187,144.2986101183335 -36.60670461352599,144.2975809378104 -36.606842288505504,144.29740940772322 -36.60532785021491,144.2648186911588 -36.60463945934398,144.26464716107165 -36.605603204843085,144.26361798054856 -36.60574088178856,144.263103390287 -36.60463945934398,144.24286283999965 -36.604088742223894,144.2416621293894 -36.604088742223894,144.2414905993022 -36.60505249460376,144.23908917808166 -36.60491481642958,144.23874611790728 -36.60725531196955,144.2378884674714 -36.60697996323927,144.23840305773294 -36.604088742223894,144.23685928694832 -36.603951062329536,144.1768237564349 -36.602711932221936,144.1768237564349 -36.604501780432564,144.17630916617338 -36.60463945934398,144.17545151573748 -36.60367570180357,144.17270703434258 -36.603400340294655,144.17270703434258 -36.60257424987018,144.15332413449113 -36.60216120134051,144.15315260440394 -36.603813382189415,144.15229495396804 -36.603951062329536,144.1500650628347 -36.60367570180357,144.15040812300904 -36.60216120134051,144.14852129205005 -36.60216120134051),(144.1498935327475 -36.603951062329536,144.1500650628347 -36.60670461352599,144.1492074123988 -36.60711763772727,144.1490358823116 -36.60697996323927,144.1498935327475 -36.603951062329536))",
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

func TestErrInsteadOfPanic(t *testing.T) {
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
			_, err := tc.op(g1, g2)
			expectErr(t, err)
			t.Log(err)
		})
	}
}
