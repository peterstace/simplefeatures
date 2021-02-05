package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/de9im"
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

		// Empty cases for relate.
		{input1: "POINT EMPTY", input2: "POINT(0 0)", relate: "FFFFFF0F2"},
		{input1: "POINT EMPTY", input2: "LINESTRING(0 0,1 1)", relate: "FFFFFF102"},
		{input1: "POINT EMPTY", input2: "LINESTRING(0 0,0 1,1 0,0 0)", relate: "FFFFFF1F2"},
		{input1: "POINT EMPTY", input2: "POLYGON((0 0,0 1,1 0,0 0))", relate: "FFFFFF212"},

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
							got de9im.Matrix
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
						if got.StringCode() != want {
							t.Errorf("\nwant: %v\ngot:  %v\n", want, got.StringCode())
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

func TestBinaryOpEmptyInputs(t *testing.T) {
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
				if got.StringCode() != "FFFFFFFF2" {
					t.Errorf("got=%v but want=FFFFFFFF2", got.StringCode())
				}
			})

		})
	}
}

func TestBinaryOpGeometryCollection(t *testing.T) {
	for i, wkt := range []string{
		"GEOMETRYCOLLECTION(POINT(0 0))",
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, wkt)
			_, err := geom.Union(g, g)
			if err == nil {
				t.Error("expected to fail, but not nil err")
			}
		})
	}
}
