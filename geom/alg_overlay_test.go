package geom_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
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
			inter:   "POLYGON EMPTY",
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
			revDiff: "POLYGON EMPTY",
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
			fwdDiff: "POLYGON EMPTY",
			revDiff: "POLYGON EMPTY",
			symDiff: "POLYGON EMPTY",
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
			revDiff: "LINESTRING EMPTY",
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
			union:   "GEOMETRYCOLLECTION(POINT(3 1),POLYGON((0 0,0 2,2 2,2 0,0 0)))",
			inter:   "POINT(1 1)",
			fwdDiff: "POLYGON((0 0,0 2,2 2,2 0,0 0))",
			revDiff: "POINT(3 1)",
			symDiff: "GEOMETRYCOLLECTION(POINT(3 1),POLYGON((0 0,0 2,2 2,2 0,0 0)))",
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
			revDiff: "POINT EMPTY",
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
			union:   "POLYGON((0 0,0 1,1 0,0 0))",
			inter:   "POINT(0.5 0.5)",
			fwdDiff: "POLYGON((0 0,0 1,1 0,0 0))",
			revDiff: "POINT EMPTY",
			symDiff: "POLYGON((0 0,0 1,1 0,0 0))",
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
			inter:   "POINT EMPTY",
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
			union:   "LINESTRING(0 0,1 1)",
			inter:   "POINT(0.35355339059327373 0.35355339059327373)",
			fwdDiff: "LINESTRING(0 0,1 1)",
			revDiff: "POINT EMPTY",
			symDiff: "LINESTRING(0 0,1 1)",
			relate:  "0F1FF0FF2",
		},
		{
			// LineString with a Point in the middle of it.
			input1:  "POINT(5 5)",
			input2:  "LINESTRING(1 2,9 8)",
			union:   "LINESTRING(1 2,9 8)",
			inter:   "POINT(5 5)",
			fwdDiff: "POINT EMPTY",
			revDiff: "LINESTRING(1 2,9 8)",
			symDiff: "LINESTRING(1 2,9 8)",
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
			union:   "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1),(1 0,0.5000000000000001 0.5,0.5 0.5),(0.5 0.5,0 1))",
			inter:   "POINT(0.5 0.5)",
			fwdDiff: "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1))",
			revDiff: "MULTILINESTRING((1 0,0.5000000000000001 0.5,0.5 0.5),(0.5 0.5,0 1))",
			symDiff: "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1),(1 0,0.5000000000000001 0.5,0.5 0.5),(0.5 0.5,0 1))",
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
				LINESTRING(3 3,5 3),
				LINESTRING(3 4,3 5),
				POLYGON((3 3,2 2,2 4,3 4,4 4,3 3)))`,
			fwdDiff: `GEOMETRYCOLLECTION(
				LINESTRING(0 3,1 3),
				LINESTRING(5 3,6 3),
				POLYGON((2 2,1 1,1 3,1 5,3 5,5 5,4 4,3 4,2 4,2 2)))`,
			revDiff: `GEOMETRYCOLLECTION(
				LINESTRING(3 0,3 1),
				LINESTRING(3 5,3 6),
				POLYGON((2 0,2 2,3 3,4 4,6 4,5 3,3 1,2 0)))`,
			symDiff: `GEOMETRYCOLLECTION(
				LINESTRING(0 3,1 3),
				LINESTRING(5 3,6 3),
				LINESTRING(3 0,3 1),
				LINESTRING(3 5,3 6),
				POLYGON((2 0,2 2,3 3,4 4,6 4,5 3,3 1,2 0)),
				POLYGON((1 1,1 3,1 5,3 5,5 5,4 4,3 4,2 4,2 2,1 1)))`,
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
				LINESTRING(4 4,6 4),
				LINESTRING(3 5,3 6),
				POLYGON((4 4,2 2,2 5,3 5,5 5,4 4)))`,
			fwdDiff: `GEOMETRYCOLLECTION(
				LINESTRING(0 4,1 4),
				LINESTRING(6 4,7 4),
				POLYGON((2 2,1 1,1 4,1 6,3 6,6 6,5 5,3 5,2 5,2 2)))`,
			revDiff: `GEOMETRYCOLLECTION(
				LINESTRING(3 0,3 1),
				LINESTRING(3 6,3 7),
				POLYGON((3 1,2 0,2 2,4 4,5 5,7 5,6 4,3 1)))`,
			symDiff: `GEOMETRYCOLLECTION(
				LINESTRING(0 4,1 4),
				LINESTRING(6 4,7 4),
				LINESTRING(3 0,3 1),
				LINESTRING(3 6,3 7),
				POLYGON((2 0,2 2,4 4,5 5,7 5,6 4,3 1,2 0)),
				POLYGON((1 1,1 4,1 6,3 6,6 6,5 5,3 5,2 5,2 2,1 1)))`,
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
			revDiff: "LINESTRING EMPTY",
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
			fwdDiff: "POLYGON EMPTY",
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
			inter:   "LINESTRING EMPTY",
			fwdDiff: "LINESTRING(-1 1,1 -1)",
			revDiff: "MULTILINESTRING((1 0,0 1),(0 1,1 2),(2 0,3 1),(3 1,2 2))",
			symDiff: "MULTILINESTRING((1 0,0 1),(0 1,1 2),(2 0,3 1),(3 1,2 2),(-1 1,1 -1))",
			relate:  "FF1FF0102",
		},
		{
			input1:  "LINESTRING(0 1,1 0)",
			input2:  "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))",
			union:   "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))",
			inter:   "LINESTRING(0 1,1 0)",
			fwdDiff: "LINESTRING EMPTY",
			revDiff: "MULTIPOLYGON(((0 1,1 1,1 0,0 0,0 1)),((2 1,3 1,3 0,2 0,2 1)))",
			symDiff: "MULTIPOLYGON(((1 1,1 0,0 0,0 1,1 1)),((3 1,3 0,2 0,2 1,3 1)))",
			relate:  "1FFF0F212",
		},
		{
			input1:  "POLYGON((1 0,0 1,1 1,1 0))",
			input2:  "POLYGON((2 0,2 1,3 1,3 0,2 0))",
			union:   "MULTIPOLYGON(((1 0,0 1,1 1,1 0)),((2 0,2 1,3 1,3 0,2 0)))",
			inter:   "POLYGON EMPTY",
			fwdDiff: "POLYGON((1 0,0 1,1 1,1 0))",
			revDiff: "POLYGON((2 0,2 1,3 1,3 0,2 0))",
			symDiff: "MULTIPOLYGON(((2 0,2 1,3 1,3 0,2 0)),((1 0,0 1,1 1,1 0)))",
			relate:  "FF2FF1212",
		},
		{
			input1:  "POLYGON((0 0,1 1,1 0,0 0))",
			input2:  "POLYGON((2 2,3 2,3 1,2 1,2 2))",
			union:   "MULTIPOLYGON(((0 0,1 0,1 1,0 0)),((2 1,2 2,3 2,3 1,2 1)))",
			inter:   "POLYGON EMPTY",
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
			fwdDiff: "LINESTRING EMPTY",
			revDiff: "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 1,2 2,3 2,3 1,2 1)))",
			symDiff: "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 1,2 2,3 2,3 1,2 1)))",
			relate:  "1FFF0F212",
		},
		{
			input1:  "POINT(5 5)",
			input2:  "LINESTRING(5 3,4 8,1 2,9 8)",
			fwdDiff: "POINT EMPTY",
			relate:  "0FFFFF102",
		},
		{
			input1: "LINESTRING(1 1,2 2,3 3,0 0)",
			input2: "LINESTRING(1 2,2 0)",
			inter:  "POINT(1.3333333333 1.3333333333)",
			relate: "0F1FF0102",
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
			fwdDiff: "MULTILINESTRING((2 2,3 3),(3 3,4 4),(4 4,5 5),(2 2,1 1))",
			relate:  "101F00FF2",
		},
		{
			input1:  "LINESTRING(0 0,0 0,0 1,1 0,0 0)",
			input2:  "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1),(0 1,0.3333333333 0.6666666667,0.5 0.5),(0.5 0.5,1 0))",
			fwdDiff: "MULTILINESTRING((0 0,0 1),(0 1,0.5 0.5),(1 0,0 0))",
			relate:  "101FFF102",
		},
		{
			input1: "LINESTRING(1 0,0.5000000000000001 0.5,0 1)",
			input2: "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0),(0.5 0.5,1 0.5,1 1.5,0.5 1.5,0.5 0.5)))",
			union:  "GEOMETRYCOLLECTION(POLYGON((0 0,0 1,0 2,2 2,2 0,1 0,0 0),(0.5000000000000001 0.5,1 0.5,1 1.5,0.5 1.5,0.5 0.5000000000000001,0.5 0.5,0.5000000000000001 0.5)),LINESTRING(0.5000000000000001 0.5,0.5 0.5000000000000001))",
			relate: "101F0F212",
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

		// NESTED GEOMETRYCOLLECTION TESTS
		{
			// GC containing a GC with a polygon.
			input1:  "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POLYGON((0 0,2 0,2 2,0 2,0 0))))",
			input2:  "POLYGON((1 1,3 1,3 3,1 3,1 1))",
			union:   "POLYGON((0 0,0 2,1 2,1 3,3 3,3 1,2 1,2 0,0 0))",
			inter:   "POLYGON((1 1,1 2,2 2,2 1,1 1))",
			fwdDiff: "POLYGON((0 0,0 2,1 2,1 1,2 1,2 0,0 0))",
			revDiff: "POLYGON((1 2,1 3,3 3,3 1,2 1,2 2,1 2))",
			symDiff: "MULTIPOLYGON(((0 0,0 2,1 2,1 1,2 1,2 0,0 0)),((1 2,1 3,3 3,3 1,2 1,2 2,1 2)))",
		},
		{
			// Deeply nested GC.
			input1:  "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POLYGON((0 0,1 0,1 1,0 1,0 0)))))",
			input2:  "POLYGON((0.5 0,1.5 0,1.5 1,0.5 1,0.5 0))",
			union:   "POLYGON((0 0,0 1,0.5 1,1 1,1.5 1,1.5 0,1 0,0.5 0,0 0))",
			inter:   "POLYGON((0.5 0,0.5 1,1 1,1 0,0.5 0))",
			fwdDiff: "POLYGON((0 0,0 1,0.5 1,0.5 0,0 0))",
			revDiff: "POLYGON((1 0,1 1,1.5 1,1.5 0,1 0))",
			symDiff: "MULTIPOLYGON(((0 0,0 1,0.5 1,0.5 0,0 0)),((1 0,1 1,1.5 1,1.5 0,1 0)))",
		},
		{
			// Both inputs are nested GCs.
			input1:  "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POLYGON((0 0,2 0,2 2,0 2,0 0))))",
			input2:  "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POLYGON((1 1,3 1,3 3,1 3,1 1))))",
			union:   "POLYGON((0 0,0 2,1 2,1 3,3 3,3 1,2 1,2 0,0 0))",
			inter:   "POLYGON((1 1,1 2,2 2,2 1,1 1))",
			fwdDiff: "POLYGON((0 0,0 2,1 2,1 1,2 1,2 0,0 0))",
			revDiff: "POLYGON((1 2,1 3,3 3,3 1,2 1,2 2,1 2))",
			symDiff: "MULTIPOLYGON(((0 0,0 2,1 2,1 1,2 1,2 0,0 0)),((1 2,1 3,3 3,3 1,2 1,2 2,1 2)))",
		},

		// EMPTY GEOMETRYCOLLECTION TESTS
		{
			// Empty GC × Polygon.
			input1:  "GEOMETRYCOLLECTION EMPTY",
			input2:  "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			union:   "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			inter:   "GEOMETRYCOLLECTION EMPTY",
			fwdDiff: "GEOMETRYCOLLECTION EMPTY",
			revDiff: "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			symDiff: "POLYGON((0 0,0 1,1 1,1 0,0 0))",
		},
		{
			// Polygon × Empty GC.
			input1:  "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			input2:  "GEOMETRYCOLLECTION EMPTY",
			union:   "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			inter:   "GEOMETRYCOLLECTION EMPTY",
			fwdDiff: "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			revDiff: "GEOMETRYCOLLECTION EMPTY",
			symDiff: "POLYGON((0 0,0 1,1 1,1 0,0 0))",
		},
		{
			// GC containing empty polygon.
			input1:  "GEOMETRYCOLLECTION(POLYGON EMPTY,POLYGON((0 0,1 0,1 1,0 1,0 0)))",
			input2:  "POLYGON((0.5 0,1.5 0,1.5 1,0.5 1,0.5 0))",
			union:   "POLYGON((0 0,0 1,0.5 1,1 1,1.5 1,1.5 0,1 0,0.5 0,0 0))",
			inter:   "POLYGON((0.5 0,0.5 1,1 1,1 0,0.5 0))",
			fwdDiff: "POLYGON((0 0,0 1,0.5 1,0.5 0,0 0))",
			revDiff: "POLYGON((1 0,1 1,1.5 1,1.5 0,1 0))",
			symDiff: "MULTIPOLYGON(((0 0,0 1,0.5 1,0.5 0,0 0)),((1 0,1 1,1.5 1,1.5 0,1 0)))",
		},

		// GC WITH MULTI* TYPES TESTS
		{
			// GC containing MultiPolygon.
			input1:  "GEOMETRYCOLLECTION(MULTIPOLYGON(((0 0,1 0,1 1,0 1,0 0)),((2 0,3 0,3 1,2 1,2 0))))",
			input2:  "POLYGON((0.5 0,2.5 0,2.5 1,0.5 1,0.5 0))",
			union:   "POLYGON((0 0,0 1,0.5 1,1 1,2 1,2.5 1,3 1,3 0,2.5 0,2 0,1 0,0.5 0,0 0))",
			inter:   "MULTIPOLYGON(((0.5 0,0.5 1,1 1,1 0,0.5 0)),((2 0,2 1,2.5 1,2.5 0,2 0)))",
			fwdDiff: "MULTIPOLYGON(((0 0,0 1,0.5 1,0.5 0,0 0)),((2.5 0,2.5 1,3 1,3 0,2.5 0)))",
			revDiff: "POLYGON((1 0,1 1,2 1,2 0,1 0))",
			symDiff: "MULTIPOLYGON(((0 0,0 1,0.5 1,0.5 0,0 0)),((1 0,1 1,2 1,2 0,1 0)),((2.5 0,2.5 1,3 1,3 0,2.5 0)))",
		},

		// MIXED-DIMENSION NESTED GC TESTS
		{
			// Nested GC with mixed dimensions (polygon in nested GC, line at top level).
			input1:  "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POLYGON((0 0,2 0,2 2,0 2,0 0))),LINESTRING(1 0,1 3))",
			input2:  "POLYGON((1 1,3 1,3 3,1 3,1 1))",
			union:   "POLYGON((1 2,1 3,3 3,3 1,2 1,2 0,1 0,0 0,0 2,1 2))",
			inter:   "GEOMETRYCOLLECTION(LINESTRING(1 2,1 3),POLYGON((1 2,2 2,2 1,1 1,1 2)))",
			fwdDiff: "POLYGON((1 0,0 0,0 2,1 2,1 1,2 1,2 0,1 0))",
			revDiff: "POLYGON((1 2,1 3,3 3,3 1,2 1,2 2,1 2))",
			symDiff: "MULTIPOLYGON(((0 0,0 2,1 2,1 1,2 1,2 0,1 0,0 0)),((3 3,3 1,2 1,2 2,1 2,1 3,3 3)))",
		},

		// GC COMPOSITION TEST
		{
			// GC COMPOSITION TEST
			input1: `
				GEOMETRYCOLLECTION(
					POLYGON((0 0,0 1,3 1,3 0,0 0)),
					LINESTRING(0.5 1.5,2.5 1.5),
					MULTIPOINT(0.5 2.5,1.5 2.5,2.5 2.5)
				)`,
			input2: `
				GEOMETRYCOLLECTION(
					POLYGON((0 0,1 0,1 3,0 3,0 0)),
					LINESTRING(1.5 0,1.5 3),
					MULTIPOINT(2.5 0.5,2.5 1.5,2.5 2.5)
				)`,
			union: `
				GEOMETRYCOLLECTION(
					POINT(2.5 2.5),
					LINESTRING(1 1.5,1.5 1.5),
					LINESTRING(1.5 1.5,2.5 1.5),
					LINESTRING(1.5 1,1.5 1.5),
					LINESTRING(1.5 1.5,1.5 3),
					POLYGON((0 1,0 3,1 3,1 1.5,1 1,1.5 1,3 1,3 0,1.5 0,1 0,0 0,0 1))
				)`,
			inter: `
				GEOMETRYCOLLECTION(
					POINT(0.5 2.5),
					POINT(1.5 1.5),
					POINT(1.5 2.5),
					POINT(2.5 0.5),
					POINT(2.5 1.5),
					POINT(2.5 2.5),
					LINESTRING(0.5 1.5,1 1.5),
					LINESTRING(1.5 0,1.5 1),
					POLYGON((0 1,1 1,1 0,0 0,0 1))
				)`,
			fwdDiff: `
				GEOMETRYCOLLECTION(
					LINESTRING(1 1.5,1.5 1.5),
					LINESTRING(1.5 1.5,2.5 1.5),
					POLYGON((1 0,1 1,1.5 1,3 1,3 0,1.5 0,1 0))
				)`,
			revDiff: `
				GEOMETRYCOLLECTION(
					LINESTRING(1.5 1,1.5 1.5),
					LINESTRING(1.5 1.5,1.5 3),
					POLYGON((0 3,1 3,1 1.5,1 1,0 1,0 3))
				)`,
			symDiff: `
				GEOMETRYCOLLECTION(
					LINESTRING(1 1.5,1.5 1.5),
					LINESTRING(1.5 1.5,2.5 1.5),
					LINESTRING(1.5 1,1.5 1.5),
					LINESTRING(1.5 1.5,1.5 3),
					POLYGON((1 1,1.5 1,3 1,3 0,1.5 0,1 0,1 1)),
					POLYGON((1 3,1 1.5,1 1,0 1,0 3,1 3))
				)`,
			relate: "212111212",
		},

		// Reproduces "no rings to extract" DCEL errors (reported in
		// https://github.com/peterstace/simplefeatures/issues/569).
		{
			input1: "POLYGON((-57.84764391579377 -14.00436771429812, -57.98105430423379 -13.978568346975345, -57.97219 -13.895754, -57.815573 -13.870471, -57.78975494169227 -13.97408746357712, -57.79567678742665 -14.003207561112367, -57.84764391579377 -14.00436771429812))",
			input2: "POLYGON((-57.97219 -13.895754, -57.815573 -13.870471, -57.782572 -14.002915, -57.984142 -14.007415, -57.97219 -13.895754))",
			inter:  "POLYGON((-57.84764391579377 -14.00436771429812,-57.98105430423379 -13.978568346975345,-57.97219 -13.895754,-57.815573 -13.870471,-57.78975494169227 -13.974087463577124,-57.79567678742665 -14.003207561112367,-57.82788570034102 -14.003926617063723,-57.84764391579377 -14.00436771429812))",
		},
		{
			input1: "POLYGON((-91.090505 33.966621, -91.094941 33.966624, -91.09491 33.96729, -91.094691 33.968384, -91.094602 33.968744, -91.094547 33.968945, -91.094484 33.969145, -91.093264 33.972456, -91.093108 33.97274, -91.092382 33.973979, -89.942235 35.721107, -89.941594 35.721928, -89.940438 35.723405, -89.720717 36, -89.711573 36, -89.645271 35.924821, -89.644942 35.924442, -89.644529 35.923925, -89.6429 35.921751, -89.642692 35.921465, -89.642576 35.921135, -89.642146 35.919717, -89.642026 35.91928, -89.641571 35.917498, -89.641166 35.91565, -89.63955 35.907509, -89.639384 35.906472, -89.639338 35.905496, -89.639356 35.904841, -89.639394 35.903992, -89.63944 35.90339, -89.639487 35.902831, -89.639559 35.902218, -89.640275 35.896772, -89.64057 35.894942, -89.640962 35.893237, -89.641113 35.892633, -89.641786 35.890644, -89.642306 35.889248, -89.642587 35.888566, -89.642808 35.888057, -89.643386 35.88681, -89.64378 35.885975, -90.060853 35.140433, -90.585556 34.404858, -90.888428 34.027973, -90.890265 34.026455, -90.890862 34.026091, -90.895918 34.023915, -90.896574 34.023654, -90.896965 34.023521, -91.090505 33.966621))",
			input2: "POLYGON((-90.19553150069916 34.95162878475482, -90.42127335893674 34.993424947208105, -90.30813100280166 35.16529356781885, -90.12850301040231 35.13253239680938, -90.16769780459812 34.99064851784101, -90.19553150069916 34.95162878475482))",
			inter:  "POLYGON((-90.38057125559546 35.055253378188205,-90.30813100280166 35.16529356781885,-90.12850301040231 35.13253239680938,-90.16769780459812 34.99064851784102,-90.19553150069916 34.95162878475482,-90.42127335893674 34.993424947208105,-90.38057125559546 35.055253378188205))",
		},
		{
			input1: "POLYGON((-91.090505 33.966621, -91.094941 33.966624, -91.09491 33.96729, -91.094691 33.968384, -91.094602 33.968744, -91.094547 33.968945, -91.094484 33.969145, -91.093264 33.972456, -91.093108 33.97274, -91.092382 33.973979, -89.942235 35.721107, -89.941594 35.721928, -89.940438 35.723405, -89.720717 36, -89.711573 36, -89.645271 35.924821, -89.644942 35.924442, -89.644529 35.923925, -89.6429 35.921751, -89.642692 35.921465, -89.642576 35.921135, -89.642146 35.919717, -89.642026 35.91928, -89.641571 35.917498, -89.641166 35.91565, -89.63955 35.907509, -89.639384 35.906472, -89.639338 35.905496, -89.639356 35.904841, -89.639394 35.903992, -89.63944 35.90339, -89.639487 35.902831, -89.639559 35.902218, -89.640275 35.896772, -89.64057 35.894942, -89.640962 35.893237, -89.641113 35.892633, -89.641786 35.890644, -89.642306 35.889248, -89.642587 35.888566, -89.642808 35.888057, -89.643386 35.88681, -89.64378 35.885975, -90.060853 35.140433, -90.585556 34.404858, -90.888428 34.027973, -90.890265 34.026455, -90.890862 34.026091, -90.895918 34.023915, -90.896574 34.023654, -90.896965 34.023521, -91.090505 33.966621))",
			input2: "POLYGON((-90.29716937546225 35.18194480113967, -90.29596586203434 35.18172958540237, -90.34543219833212 34.998800268076835, -90.41002551098103 35.01051096325925, -90.29716937546225 35.18194480113967))",
			inter:  "POLYGON((-90.41002551098103 35.01051096325925,-90.39289363505974 35.036535097665215,-90.29716937546225 35.18194480113967,-90.29596586203434 35.18172958540237,-90.34543219833212 34.998800268076835,-90.41002551098103 35.01051096325925))",
		},
		{
			input1: "POLYGON((-149.845771 -17.472558, -149.888137 -17.477017, -149.929731 -17.480468, -149.934682 -17.50814, -149.920475 -17.541336, -149.895694 -17.571267, -149.861608 -17.600395, -149.832332 -17.611409, -149.791981 -17.611947, -149.774766 -17.577051, -149.753707 -17.535289, -149.744632 -17.494022, -149.765688 -17.465994, -149.805445 -17.46709, -149.845771 -17.472558))",
			input2: "POLYGON((-149.8839047803303 -17.58134141150439, -149.86106842049824 -17.474168045744268, -149.85203718167833 -17.473217512441664, -149.74468306149925 -17.494254193376246, -149.753707 -17.535289, -149.774766 -17.577051, -149.791981 -17.611947, -149.832332 -17.611409, -149.861608 -17.600395, -149.8839047803303 -17.58134141150439))",
			inter:  "POLYGON((-149.85203718167833 -17.473217512441664,-149.74468306149925 -17.494254193376246,-149.753707 -17.535289,-149.774766 -17.577051,-149.791981 -17.611947,-149.832332 -17.611409,-149.861608 -17.600395,-149.8839047803303 -17.581341411504383,-149.86106842049824 -17.474168045744268,-149.85668348649625 -17.473706533665833,-149.85203718167833 -17.473217512441664))",
		},

		// Reproduces a failed DCEL operation (reported in
		// https://github.com/peterstace/simplefeatures/issues/496).
		{
			input1: "POLYGON((-83.5825305152402 32.7316823944815,-83.58376293006216 32.73315376178507,-83.58504085655653 32.734597137036324,-83.58636334101533 32.73601156235818,-83.58772946946186 32.73739605462324,-83.58913829287843 32.738749652823024,-83.59058883640313 32.740071417020225,-83.59208009385833 32.74136043125909,-83.59361103333862 32.74261579844315,-83.59518059080796 32.743836647669376,-83.59678768034422 32.745022130852156,-83.59843118482598 32.746171426099316,-83.60010996734121 32.74728373328835,-83.60182286094272 32.74835828151699,-83.60356867749573 32.74939432363171,-83.6053462070958 32.75039113983657,-83.60715421364506 32.75134803897381,-83.60899144521312 32.75226435508957,-83.61085662379259 32.75313945319649,-83.61274845635847 32.75397272333648,-83.61466562799973 32.7547635884388,-83.61660680960262 32.75551149954699,-83.61857065109865 32.756215934654755,-83.62055578961366 32.75687640673859,-83.62256084632457 32.75749245578336,-83.62458442890414 32.75806365483586,-83.62662513245226 32.75858960587211,-83.62868153809899 32.759069943900975,-83.63075221766115 32.759504334926895,-83.6328357331767 32.759892478947584,-83.6349306369047 32.76023410294001,-83.63703547248946 32.76052897293848,-83.63914877915153 32.76077688192745,-83.64126908772954 32.760977657903275,-83.64339492533685 32.76113116287015,-83.64552481489585 32.76123728881648,-83.64765727560365 32.761295961756694,-83.64979082712301 32.761307141684846,-83.65192398562425 32.76127082060284,-83.65405527030447 32.761187023508135,-83.65618319896379 32.76105580840831,-83.65830629359328 32.76087726729329,-83.66042307921083 32.7606515241745,-83.66253208572374 32.760378736038,-83.66463184629896 32.76005909090909,-83.666720902951 32.75969281274174,-83.66879780351513 32.75928015456466,-83.67086110700251 32.758821403399814,-83.67290937847787 32.75831687924383,-83.67494119511315 32.75776693105163,-83.67695514665314 32.75717194084009,-83.67894983308719 32.756532323629834,-83.68092387070277 32.75584852243937,-83.68287588719608 32.755121013232745,-83.68480452667785 32.75435029997188,-83.6867084511868 32.75353691874934,-83.68858633871054 32.752681435518205,-83.69043688423329 32.75178444323992,-83.69225880474181 32.75084656496929,-83.69405083421947 32.74986845274852,-83.69581173063119 32.74885078644311,-83.69754027103403 32.74779427211197,-83.69923525541884 32.746699642822385,-83.70089550880579 32.74556766051264,-83.70251987926525 32.74439911017125,-83.70410724071183 32.74319480286376,-83.70565649104168 32.74195557654766,-83.70716655737118 32.740682290251605,-83.70863639277735 32.73937582791714,-83.71006497711272 32.738037097583764,-83.71145132142914 32.73666702917708,-83.71279446471812 32.735266572878935,-83.71409347705406 32.73383670052444,-83.71534745831372 32.73237840618403,-83.71655554062085 32.73089270080836,-83.71771688788061 32.729380616419256,-83.7188306961288 32.72784320203518,-83.71989619434675 32.7262815256503,-83.72091264550903 32.72469667027635,-83.72187934768948 32.72308973592171,-83.72279563289698 32.721461838543654,-83.72366086608594 32.71981410620665,-83.72447445225765 32.718147683855385,-83.72523582737989 32.71646372644627,-83.72594446851502 32.71476340406971,-83.72659988462323 32.713047893684596,-83.7272016266906 32.71131838726742,-83.72774927777567 32.70957608482731,-83.72824246383601 32.70782219347475,-83.72868084389114 32.70605793103041,-83.72906411790967 32.70428451962234,-83.7293920229094 32.7025031912739,-83.72966433589676 32.700715179871075,-83.7298808698651 32.69892172546988,-83.7300414788296 32.69712407208444,-83.73014605577549 32.69532346570795,-83.73019452970033 32.693521154312656,-83.73018687161033 32.691718388897606,-83.73012309050767 32.689916417551466,-83.7300032323846 32.68811649010913,-83.7298273852511 32.68631985477568,-83.72959567309675 32.684527756380156,-83.72930826092573 32.68274143695762,-83.72896535074773 32.680962133537285,-83.72856718354245 32.6791910811693,-83.72811403832833 32.67742950582342,-83.72760623312291 32.67567862846354,-83.72704412086836 32.67393966108957,-83.72642809662005 32.672213810695574,-83.72575858936847 32.670502271353506,-83.72503606605545 32.66880622898625,-83.72426102976982 32.66712685869006,-83.7234340194563 32.66546532333759,-83.72255561311698 32.66382277497487,-83.7216264197786 32.66220034969901,-83.72064708740885 32.66059917243128,-83.7196182950583 32.659020351096295,-83.71854075868117 32.65746497883392,-83.71741522729366 32.655934134581365,-83.71624248192614 32.65442887527333,-83.71502333853363 32.65295024597017,-83.71375864112719 32.65149926868198,-83.71244926771118 32.650076949469984,-83.7110961292355 32.64868427122874,-83.70970016179575 32.64732219892485,-83.70826233431666 32.64599167470751,-83.70678364482677 32.6446936215174,-83.70526511731526 32.64342893528683,-83.70370780580646 32.642198493088834,-83.70211278830624 32.64100314591942,-83.70048117076016 32.639843721724354,-83.5825305152402 32.7316823944815))",
			input2: "POLYGON((-83.7004774529266 32.6398466121988,-83.70047002035855 32.63878317849731,-83.70041586825457 32.63698593605796,-83.70032015514218 32.635189934722355,-83.7001829220105 32.633395861340894,-83.70000422287623 32.63160440392813,-83.69978412771954 32.62981624851955,-83.69952272555449 32.62803208010292,-83.69922011737137 32.62625258168692,-83.69887642118621 32.62447843430119,-83.69849177198259 32.62271031792762,-83.69806631976164 32.62094891056905,-83.69760022855255 32.61919488417471,-83.69709368034157 32.617448911858,-83.6965468711139 32.61571166254135,-83.6959600118723 32.61398380014135,-83.69533332962666 32.61226598682834,-83.69466706733576 32.610558879417574,-83.69396148204461 32.60886313204674,-83.69321684575763 32.607179393731265,-83.69243344543858 32.605508308364286,-83.69161158312701 32.60385051506592,-83.69075157582185 32.60220664969666,-83.68985375449185 32.600577340317166,-83.68891846413382 32.59896320998224,-83.68794606575167 32.59736487767214,-83.6869369323401 32.595782953403166,-83.68589145196958 32.59421804311706,-83.68481002662222 32.59267074483933,-83.68369307318127 32.59114165158963,-83.68254101772683 32.58963134730721,-83.68135430531467 32.588140409994146,-83.68013339089586 32.58666941066759,-83.67887874246 32.58521891044939,-83.6775908410352 32.58378946417498,-83.67627018255072 32.58238161992768,-83.67491727108492 32.58099591461497,-83.67353262759636 32.57963287932353,-83.672116782124 32.57829303408062,-83.67067027669756 32.576976891812144,-83.6691936661525 32.57568495659644,-83.66768751766423 32.57441772040466,-83.66615240620806 32.57317566915436,-83.66458891968134 32.57195927688873,-83.66299765913638 32.57076900868697,-83.66137923156268 32.5696053204315,-83.65973425803597 32.56846865624676,-83.65806336847956 32.56735945001265,-83.65636720096583 32.56627812687794,-83.65464640637285 32.56522509872003,-83.65290164081736 32.56420076950008,-83.65113357426955 32.56320553037359,-83.64934288275319 32.562239760214204,-83.64753025020829 32.561303830095774,-83.64569637058659 32.56039809694767,-83.64384194505757 32.55952290582491,-83.64196768247413 32.558678592702115,-83.64007430088597 32.55786547859453,-83.63816252230092 32.557083875495216,-83.63623307966984 32.55633408241693,-83.63428670803907 32.55561638434438,-83.63232415432932 32.55493105625056,-83.63034616674187 32.55427836018635,-83.62835350104498 32.553658545105925,-83.62634691952618 32.553071847041345,-83.62432718982811 32.55251849201296,-83.62229508215452 32.55199869096533,-83.62025137346129 32.551512640931385,-83.61819684489521 32.55106052890324,-83.61613228225971 32.55064252688454,-83.6140584727552 32.55025879587766,-83.61197620917002 32.54990948087789,-83.60988628662082 32.54959471587927,-83.60778950394953 32.54931462087689,-83.6056866623264 32.54906930390417,-83.60357856571565 32.54885885793329,-83.60146601715016 32.548683362970685,-83.59934982455226 32.548542887017454,-83.59723079793977 32.5484374830644,-83.59510974337238 32.548367191122004,-83.59298747366185 32.548332038187716,-83.59086479905883 32.54833203826062,-83.58874252934831 32.548367191340425,-83.58662147571223 32.54843748342834,-83.58450244816842 32.548542887526914,-83.58238625650185 32.54868336362566,-83.58027370793636 32.54885885873379,-83.57816561039428 32.549069304850185,-83.57606276877115 32.54931462196843,-83.57396598609986 32.54959471711633,-83.57187606448198 32.54990948226047,-83.5697938008968 32.55025879740575,-83.5677199913923 32.55064252855815,-83.56565542782548 32.551060530722374,-83.5636008992594 32.551512642896036,-83.56155719173032 32.5519986930755,-83.55952508405673 32.55251849423954,-83.55750535342733 32.55307184944255,-83.55549877190853 32.55365854762355,-83.5535061073758 32.55427836282039,-83.55152811978834 32.55493105905922,-83.54956556514728 32.55561638726946,-83.54761919456425 32.55633408551663,-83.54568975100185 32.55708387876954,-83.54377797346453 32.55786548198527,-83.54188459094505 32.55867859620927,-83.54001032940934 32.55952290944848,-83.53815590388032 32.56039810068766,-83.53632202437504 32.561303833952174,-83.53450939194656 32.56223976424523,-83.53271869949887 32.563205534521025,-83.53095063306748 32.564200773763936,-83.52920586855973 32.565225103100296,-83.52748507315184 32.566278131374624,-83.52578890668585 32.56735945462575,-83.52411801619812 32.56846866097627,-83.52247304278782 32.56960532527743,-83.52085461637827 32.5707690135329,-83.51926335501841 32.57195928185107,-83.51769986965584 32.573175674116705,-83.51616475831608 32.57441772548342,-83.5146586088965 32.575684961791616,-83.51318199858426 32.576976897123735,-83.51173549420557 32.57829303950863,-83.51031964780188 32.579632884867955,-83.50893500442973 32.58099592015939,-83.50758209412808 32.5823816254721,-83.5062614348287 32.583789469835814,-83.50497353456805 32.585218916226644,-83.50371888531728 32.58666941656126,-83.50249797101489 32.588140415887814,-83.50131125871914 32.58963135320088,-83.50015920442885 32.591141657599714,-83.499042250173 32.59267075096583,-83.49796082494206 32.594218049359974,-83.49691534573569 32.595782959762495,-83.49590621255695 32.597364884147886,-83.49493381434942 32.5989632165744,-83.49399852416602 32.600577346909326,-83.49310070301064 32.60220665628882,-83.49224069582189 32.60385052165808,-83.49141883368495 32.60550831507286,-83.49063543354052 32.60717940043984,-83.48989079742816 32.60886313887173,-83.48918521132211 32.610558886242565,-83.48851894920583 32.61226599365333,-83.48789226812434 32.613983807082754,-83.48730540905737 32.61571166948276,-83.48675859895658 32.61744891891582,-83.48625205092023 32.61919489134895,-83.48578595985666 32.620948917743284,-83.48536050781033 32.622710325101856,-83.48497585878133 32.624478441475425,-83.48463216376032 32.626252588861156,-83.48432955575183 32.628032087277155,-83.48406815277187 32.62981625569379,-83.48384805879388 32.631604411102366,-83.48366935983424 32.63339586851513,-83.48353212588765 32.63518994189659,-83.48343641394669 32.63698594334861,-83.48338226201733 32.638783185787965,-83.48336969309977 32.640580979221035,-83.48339871519394 32.642378636564594,-83.48346931829994 32.644175467941444,-83.48358148041709 32.64597078638477,-83.48373515855175 32.64776390178455,-83.4839302976977 32.6495541282217,-83.48416682484796 32.65134078059206,-83.48444465202166 32.653123171928804,-83.48476367520172 32.654900621318724,-83.48512377538256 32.656672445753124,-83.48552481557246 32.658437967112754,-83.48596664579938 32.660196507511195,-83.48644909702827 32.661947394882795,-83.48697198925194 32.663689955299255,-83.48753512252708 32.6654235226321,-83.48813828381364 32.66714743203344,-83.48878124406448 32.66886102133291,-83.4894637593313 32.67056363371527,-83.49018556960047 32.672254617021885,-83.490946399899 32.673933322353754,-83.49174596221539 32.675599105701316,-83.49258395055196 32.67725132806087,-83.49346004587252 32.67888935636589,-83.49437391423973 32.68051256174082,-83.49532520757177 32.68212032101443,-83.49631356294387 32.683712018349674,-83.49733860136595 32.68528704163479,-83.49839993377802 32.68684478888614,-83.49949715412342 32.68838466021558,-83.50062984249209 32.68990606656158,-83.50179756593545 32.69140842188941,-83.50299987939772 32.69289115215512,-83.50423632187423 32.694353686457966,-83.50550642130085 32.6957954636761,-83.50680969176003 32.697215930953156,-83.50814563324795 32.6986145432326,-83.5095137357491 32.699990763490554,-83.51091347620942 32.70134406378353,-83.51234431667373 32.70267392408429,-83.51380571115416 32.70397983332956,-83.51529709864528 32.705261290584225,-83.51681790812992 32.70651780480845,-83.51836755660017 32.70774889299506,-83.51994545115278 32.70895408121729,-83.52155098572959 32.71013290847046,-83.52318354623978 32.711284920618404,-83.52484250578684 32.71240967679629,-83.52652722932518 32.71350674393913,-83.52823706888714 32.71457570108911,-83.52997137045142 32.715616138231475,-83.53172946800598 32.71662765443185,-83.53351068657477 32.7176098626093,-83.53531434419685 32.71856238377374,-83.53713974773538 32.71948485389445,-83.53898619637013 32.720376918021124,-83.5408529819467 32.72123823313601,-83.54273938757953 32.72206846925989,-83.54464468916532 32.72286730636703,-83.54656815573227 32.72363443747023,-83.54850904932366 32.7243695685626,-83.55046662301876 32.72507241762802,-83.55244012558953 32.72574271370983,-83.5544287991722 32.72638019877346,-83.55643188078076 32.72698462986015,-83.55844860044428 32.72755577390641,-83.56047818400086 32.72809340995597,-83.56251985053652 32.72859733399097,-83.5645728170418 32.72906735002622,-83.56663629375518 32.72950327904406,-83.56870948928528 32.72990495206765,-83.57079160595433 32.73027221708768,-83.57288184468754 32.7306049300984,-83.57497940128789 32.73090296609886,-83.57708346992851 32.73116620809166,-83.57919324245424 32.73139455706556,-83.5813079090801 32.73158792303153,-83.58253417348143 32.73167954800889,-83.7004774529266 32.6398466121988))",
			union:  "POLYGON((-83.7004774529266 32.6398466121988,-83.70047002035855 32.63878317849731,-83.70041586825457 32.63698593605796,-83.70032015514218 32.635189934722355,-83.7001829220105 32.633395861340894,-83.70000422287623 32.63160440392813,-83.69978412771954 32.62981624851955,-83.69952272555449 32.62803208010292,-83.69922011737137 32.62625258168692,-83.69887642118621 32.62447843430119,-83.69849177198259 32.62271031792762,-83.69806631976164 32.62094891056905,-83.69760022855255 32.61919488417471,-83.69709368034157 32.617448911858,-83.6965468711139 32.61571166254135,-83.6959600118723 32.61398380014135,-83.69533332962666 32.61226598682834,-83.69466706733576 32.610558879417574,-83.69396148204461 32.60886313204674,-83.69321684575763 32.607179393731265,-83.69243344543858 32.605508308364286,-83.69161158312701 32.60385051506592,-83.69075157582185 32.60220664969666,-83.68985375449185 32.600577340317166,-83.68891846413382 32.59896320998224,-83.68794606575167 32.59736487767214,-83.6869369323401 32.595782953403166,-83.68589145196958 32.59421804311706,-83.68481002662222 32.59267074483933,-83.68369307318127 32.59114165158963,-83.68254101772683 32.58963134730721,-83.68135430531467 32.588140409994146,-83.68013339089586 32.58666941066759,-83.67887874246 32.58521891044939,-83.6775908410352 32.58378946417498,-83.67627018255072 32.58238161992768,-83.67491727108492 32.58099591461497,-83.67353262759636 32.57963287932353,-83.672116782124 32.57829303408062,-83.67067027669756 32.576976891812144,-83.6691936661525 32.57568495659644,-83.66768751766423 32.57441772040466,-83.66615240620806 32.57317566915436,-83.66458891968134 32.57195927688873,-83.66299765913638 32.57076900868697,-83.66137923156268 32.5696053204315,-83.65973425803597 32.56846865624676,-83.65806336847956 32.56735945001265,-83.65636720096583 32.56627812687794,-83.65464640637285 32.56522509872003,-83.65290164081736 32.56420076950008,-83.65113357426955 32.56320553037359,-83.64934288275319 32.562239760214204,-83.64753025020829 32.561303830095774,-83.64569637058659 32.56039809694767,-83.64384194505757 32.55952290582491,-83.64196768247413 32.558678592702115,-83.64007430088597 32.55786547859453,-83.63816252230092 32.557083875495216,-83.63623307966984 32.55633408241693,-83.63428670803907 32.55561638434438,-83.63232415432932 32.55493105625056,-83.63034616674187 32.55427836018635,-83.62835350104498 32.553658545105925,-83.62634691952618 32.553071847041345,-83.62432718982811 32.55251849201296,-83.62229508215452 32.55199869096533,-83.62025137346129 32.551512640931385,-83.61819684489521 32.55106052890324,-83.61613228225971 32.55064252688454,-83.6140584727552 32.55025879587766,-83.61197620917002 32.54990948087789,-83.60988628662082 32.54959471587927,-83.60778950394953 32.54931462087689,-83.6056866623264 32.54906930390417,-83.60357856571565 32.54885885793329,-83.60146601715016 32.548683362970685,-83.59934982455226 32.548542887017454,-83.59723079793977 32.5484374830644,-83.59510974337238 32.548367191122004,-83.59298747366185 32.548332038187716,-83.59086479905883 32.54833203826062,-83.58874252934831 32.548367191340425,-83.58662147571223 32.54843748342834,-83.58450244816842 32.548542887526914,-83.58238625650185 32.54868336362566,-83.58027370793636 32.54885885873379,-83.57816561039428 32.549069304850185,-83.57606276877115 32.54931462196843,-83.57396598609986 32.54959471711633,-83.57187606448198 32.54990948226047,-83.5697938008968 32.55025879740575,-83.5677199913923 32.55064252855815,-83.56565542782548 32.551060530722374,-83.5636008992594 32.551512642896036,-83.56155719173032 32.5519986930755,-83.55952508405673 32.55251849423954,-83.55750535342733 32.55307184944255,-83.55549877190853 32.55365854762355,-83.5535061073758 32.55427836282039,-83.55152811978834 32.55493105905922,-83.54956556514728 32.55561638726946,-83.54761919456425 32.55633408551663,-83.54568975100185 32.55708387876954,-83.54377797346453 32.55786548198527,-83.54188459094505 32.55867859620927,-83.54001032940934 32.55952290944848,-83.53815590388032 32.56039810068766,-83.53632202437504 32.561303833952174,-83.53450939194656 32.56223976424523,-83.53271869949887 32.563205534521025,-83.53095063306748 32.564200773763936,-83.52920586855973 32.565225103100296,-83.52748507315184 32.566278131374624,-83.52578890668585 32.56735945462575,-83.52411801619812 32.56846866097627,-83.52247304278782 32.56960532527743,-83.52085461637827 32.5707690135329,-83.51926335501841 32.57195928185107,-83.51769986965584 32.573175674116705,-83.51616475831608 32.57441772548342,-83.5146586088965 32.575684961791616,-83.51318199858426 32.576976897123735,-83.51173549420557 32.57829303950863,-83.51031964780188 32.579632884867955,-83.50893500442973 32.58099592015939,-83.50758209412808 32.5823816254721,-83.5062614348287 32.583789469835814,-83.50497353456805 32.585218916226644,-83.50371888531728 32.58666941656126,-83.50249797101489 32.588140415887814,-83.50131125871914 32.58963135320088,-83.50015920442885 32.591141657599714,-83.499042250173 32.59267075096583,-83.49796082494206 32.594218049359974,-83.49691534573569 32.595782959762495,-83.49590621255695 32.597364884147886,-83.49493381434942 32.5989632165744,-83.49399852416602 32.600577346909326,-83.49310070301064 32.60220665628882,-83.49224069582189 32.60385052165808,-83.49141883368495 32.60550831507286,-83.49063543354052 32.60717940043984,-83.48989079742816 32.60886313887173,-83.48918521132211 32.610558886242565,-83.48851894920583 32.61226599365333,-83.48789226812434 32.613983807082754,-83.48730540905737 32.61571166948276,-83.48675859895658 32.61744891891582,-83.48625205092023 32.61919489134895,-83.48578595985666 32.620948917743284,-83.48536050781033 32.622710325101856,-83.48497585878133 32.624478441475425,-83.48463216376032 32.626252588861156,-83.48432955575183 32.628032087277155,-83.48406815277187 32.62981625569379,-83.48384805879388 32.631604411102366,-83.48366935983424 32.63339586851513,-83.48353212588765 32.63518994189659,-83.48343641394669 32.63698594334861,-83.48338226201733 32.638783185787965,-83.48336969309977 32.640580979221035,-83.48339871519394 32.642378636564594,-83.48346931829994 32.644175467941444,-83.48358148041709 32.64597078638477,-83.48373515855175 32.64776390178455,-83.4839302976977 32.6495541282217,-83.48416682484796 32.65134078059206,-83.48444465202166 32.653123171928804,-83.48476367520172 32.654900621318724,-83.48512377538256 32.656672445753124,-83.48552481557246 32.658437967112754,-83.48596664579938 32.660196507511195,-83.48644909702827 32.661947394882795,-83.48697198925194 32.663689955299255,-83.48753512252708 32.6654235226321,-83.48813828381364 32.66714743203344,-83.48878124406448 32.66886102133291,-83.4894637593313 32.67056363371527,-83.49018556960047 32.672254617021885,-83.490946399899 32.673933322353754,-83.49174596221539 32.675599105701316,-83.49258395055196 32.67725132806087,-83.49346004587252 32.67888935636589,-83.49437391423973 32.68051256174082,-83.49532520757177 32.68212032101443,-83.49631356294387 32.683712018349674,-83.49733860136595 32.68528704163479,-83.49839993377802 32.68684478888614,-83.49949715412342 32.68838466021558,-83.50062984249209 32.68990606656158,-83.50179756593545 32.69140842188941,-83.50299987939772 32.69289115215512,-83.50423632187423 32.694353686457966,-83.50550642130085 32.6957954636761,-83.50680969176003 32.697215930953156,-83.50814563324795 32.6986145432326,-83.5095137357491 32.699990763490554,-83.51091347620942 32.70134406378353,-83.51234431667373 32.70267392408429,-83.51380571115416 32.70397983332956,-83.51529709864528 32.705261290584225,-83.51681790812992 32.70651780480845,-83.51836755660017 32.70774889299506,-83.51994545115278 32.70895408121729,-83.52155098572959 32.71013290847046,-83.52318354623978 32.711284920618404,-83.52484250578684 32.71240967679629,-83.52652722932518 32.71350674393913,-83.52823706888714 32.71457570108911,-83.52997137045142 32.715616138231475,-83.53172946800598 32.71662765443185,-83.53351068657477 32.7176098626093,-83.53531434419685 32.71856238377374,-83.53713974773538 32.71948485389445,-83.53898619637013 32.720376918021124,-83.5408529819467 32.72123823313601,-83.54273938757953 32.72206846925989,-83.54464468916532 32.72286730636703,-83.54656815573227 32.72363443747023,-83.54850904932366 32.7243695685626,-83.55046662301876 32.72507241762802,-83.55244012558953 32.72574271370983,-83.5544287991722 32.72638019877346,-83.55643188078076 32.72698462986015,-83.55844860044428 32.72755577390641,-83.56047818400086 32.72809340995597,-83.56251985053652 32.72859733399097,-83.5645728170418 32.72906735002622,-83.56663629375518 32.72950327904406,-83.56870948928528 32.72990495206765,-83.57079160595433 32.73027221708768,-83.57288184468754 32.7306049300984,-83.57497940128789 32.73090296609886,-83.57708346992851 32.73116620809166,-83.57919324245424 32.73139455706556,-83.5813079090801 32.73158792303153,-83.5825341712489 32.73167954784208,-83.5825305152402 32.7316823944815,-83.58376293006216 32.73315376178507,-83.58504085655653 32.734597137036324,-83.58636334101533 32.73601156235818,-83.58772946946186 32.73739605462324,-83.58913829287843 32.738749652823024,-83.59058883640313 32.740071417020225,-83.59208009385833 32.74136043125909,-83.59361103333862 32.74261579844315,-83.59518059080796 32.743836647669376,-83.59678768034422 32.745022130852156,-83.59843118482598 32.746171426099316,-83.60010996734121 32.74728373328835,-83.60182286094272 32.74835828151699,-83.60356867749573 32.74939432363171,-83.6053462070958 32.75039113983657,-83.60715421364506 32.75134803897381,-83.60899144521312 32.75226435508957,-83.61085662379259 32.75313945319649,-83.61274845635847 32.75397272333648,-83.61466562799973 32.7547635884388,-83.61660680960262 32.75551149954699,-83.61857065109865 32.756215934654755,-83.62055578961366 32.75687640673859,-83.62256084632457 32.75749245578336,-83.62458442890414 32.75806365483586,-83.62662513245226 32.75858960587211,-83.62868153809899 32.759069943900975,-83.63075221766115 32.759504334926895,-83.6328357331767 32.759892478947584,-83.6349306369047 32.76023410294001,-83.63703547248946 32.76052897293848,-83.63914877915153 32.76077688192745,-83.64126908772954 32.760977657903275,-83.64339492533685 32.76113116287015,-83.64552481489585 32.76123728881648,-83.64765727560365 32.761295961756694,-83.64979082712301 32.761307141684846,-83.65192398562425 32.76127082060284,-83.65405527030447 32.761187023508135,-83.65618319896379 32.76105580840831,-83.65830629359328 32.76087726729329,-83.66042307921083 32.7606515241745,-83.66253208572374 32.760378736038,-83.66463184629896 32.76005909090909,-83.666720902951 32.75969281274174,-83.66879780351513 32.75928015456466,-83.67086110700251 32.758821403399814,-83.67290937847787 32.75831687924383,-83.67494119511315 32.75776693105163,-83.67695514665314 32.75717194084009,-83.67894983308719 32.756532323629834,-83.68092387070277 32.75584852243937,-83.68287588719608 32.755121013232745,-83.68480452667785 32.75435029997188,-83.6867084511868 32.75353691874934,-83.68858633871054 32.752681435518205,-83.69043688423329 32.75178444323992,-83.69225880474181 32.75084656496929,-83.69405083421947 32.74986845274852,-83.69581173063119 32.74885078644311,-83.69754027103403 32.74779427211197,-83.69923525541884 32.746699642822385,-83.70089550880579 32.74556766051264,-83.70251987926525 32.74439911017125,-83.70410724071183 32.74319480286376,-83.70565649104168 32.74195557654766,-83.70716655737118 32.740682290251605,-83.70863639277735 32.73937582791714,-83.71006497711272 32.738037097583764,-83.71145132142914 32.73666702917708,-83.71279446471812 32.735266572878935,-83.71409347705406 32.73383670052444,-83.71534745831372 32.73237840618403,-83.71655554062085 32.73089270080836,-83.71771688788061 32.729380616419256,-83.7188306961288 32.72784320203518,-83.71989619434675 32.7262815256503,-83.72091264550903 32.72469667027635,-83.72187934768948 32.72308973592171,-83.72279563289698 32.721461838543654,-83.72366086608594 32.71981410620665,-83.72447445225765 32.718147683855385,-83.72523582737989 32.71646372644627,-83.72594446851502 32.71476340406971,-83.72659988462323 32.713047893684596,-83.7272016266906 32.71131838726742,-83.72774927777567 32.70957608482731,-83.72824246383601 32.70782219347475,-83.72868084389114 32.70605793103041,-83.72906411790967 32.70428451962234,-83.7293920229094 32.7025031912739,-83.72966433589676 32.700715179871075,-83.7298808698651 32.69892172546988,-83.7300414788296 32.69712407208444,-83.73014605577549 32.69532346570795,-83.73019452970033 32.693521154312656,-83.73018687161033 32.691718388897606,-83.73012309050767 32.689916417551466,-83.7300032323846 32.68811649010913,-83.7298273852511 32.68631985477568,-83.72959567309675 32.684527756380156,-83.72930826092573 32.68274143695762,-83.72896535074773 32.680962133537285,-83.72856718354245 32.6791910811693,-83.72811403832833 32.67742950582342,-83.72760623312291 32.67567862846354,-83.72704412086836 32.67393966108957,-83.72642809662005 32.672213810695574,-83.72575858936847 32.670502271353506,-83.72503606605545 32.66880622898625,-83.72426102976982 32.66712685869006,-83.7234340194563 32.66546532333759,-83.72255561311698 32.66382277497487,-83.7216264197786 32.66220034969901,-83.72064708740885 32.66059917243128,-83.7196182950583 32.659020351096295,-83.71854075868117 32.65746497883392,-83.71741522729366 32.655934134581365,-83.71624248192614 32.65442887527333,-83.71502333853363 32.65295024597017,-83.71375864112719 32.65149926868198,-83.71244926771118 32.650076949469984,-83.7110961292355 32.64868427122874,-83.70970016179575 32.64732219892485,-83.70826233431666 32.64599167470751,-83.70678364482677 32.6446936215174,-83.70526511731526 32.64342893528683,-83.70370780580646 32.642198493088834,-83.70211278830624 32.64100314591942,-83.70048117076016 32.639843721724354,-83.61872797135857 32.7034983516507,-83.7004774529266 32.6398466121988))",
		},

		// Reproduces https://github.com/peterstace/simplefeatures/issues/667.
		{
			input1:  "POLYGON((-87.62444917095299 41.39729331748811,-87.62446704222914 41.397301659153754,-87.62449506935019 41.39731775684101,-87.62452090105552 41.397335817239,-87.62454429589731 41.39735567153923,-87.6245650352042 41.3973771341657,-87.6245829251253 41.397400004509386,-87.62459779844208 41.397424068803204,-87.62460951613143 41.397449102120135,-87.62461796866529 41.39747487047549,-87.6246230770344 41.39750113301402,-87.62462479348707 41.39752764426096,-87.6246231019755 41.39755415641672,-87.62461801830612 41.397580421672856,-87.62460958999183 41.39760619452835,-87.62456285356899 41.397724210992784,-87.62455132416264 41.39774894508767,-87.62453671452448 41.3977727384771,-87.62451915784027 41.397795374248304,-87.62449881416285 41.39781664604174,-87.62447586895344 41.39783635993238,-87.62445053139074 41.39785433619775,-87.62442303246401 41.397870410956294,-87.62439362286723 41.39788443766158,-87.62436257071374 41.39789628843825,-87.62433015909178 41.39790585524787,-87.62429668348378 41.3979130508739,-87.62426244907233 41.39791780971686,-87.62422776795812 41.397920088392326,-87.62419295631425 41.397919866126564,-87.62415833150395 41.39791714494585,-87.62412420918706 41.397911949658024,-87.62412431963479 41.397911970638134,-87.6240142654138 41.3978910789883,-87.62397972826206 41.397883133976144,-87.62397604150179 41.39788197468104,-87.62397154652997 41.39788163144765,-87.62393679249253 41.397876372263475,-87.6238779544157 41.39786520300083,-87.62384341726417 41.3978572579968,-87.6238100933212 41.39784677934415,-87.62377830950128 41.397833869840746,-87.62377439835728 41.39783188189482,-87.62374847092677 41.39782212145373,-87.62371745215871 41.397807336601254,-87.62368856511434 41.397790277980455,-87.62366210264413 41.397771118528894,-87.62364238065591 41.39775363973907,-87.62363973273321 41.3977514470298,-87.62363850541901 41.39775020527393,-87.62363833301796 41.39775005248186,-87.62363806435583 41.39774975902078,-87.6236183787311 41.3977298417487,-87.62359993880196 41.397706764993615,-87.62358459039267 41.39768243883536,-87.62357248119949 41.39765709736778,-87.62356372774683 41.3976309844548,-87.62355841426607 41.39760435138386,-87.62355659188512 41.3975774544476,-87.62355827813658 41.39755055247751,-87.62356345678923 41.39752390435329,-87.62357207800432 41.39749776651146,-87.6236187070669 41.39738117370699,-87.62363074852786 41.397355686766915,-87.62364606262244 41.39733121587506,-87.62366450066513 41.39730799861518,-87.62368588364099 41.39728626039952,-87.62371000394393 41.3972662122803,-87.62373662739223 41.397248048900735,-87.62376549550245 41.39723194660534,-87.62379632799865 41.3972180617278,-87.62382882553379 41.39720652907334,-87.62386267259592 41.397197460609846,-87.62389754057125 41.397190944380924,-87.62393309093454 41.39718704365108,-87.62396897853569 41.39718579629156,-87.62400485495054 41.39718721441263,-87.62400561478837 41.397187301481495,-87.62400775375312 41.39718708148513,-87.6240424599759 41.39718602384076,-87.6240771423665 41.39718745779836,-87.62411148616091 41.39719137034386,-87.62414517966795 41.39719772596847,-87.62423643421413 41.3972184638187,-87.6242932704773 41.39722944573794,-87.62432844349291 41.39723768847929,-87.62436233243534 41.39724856156757,-87.62439459138113 41.397261954015285,-87.62442489104455 41.39727772911831,-87.62444917095299 41.39729331748811))",
			input2:  "POLYGON((-87.6242938761274 41.39744938185984,-87.62429387983731 41.3974493827643,-87.62430377142961 41.397452198873204,-87.6243132556424 41.39745572107071,-87.62432224382728 41.39745991643503,-87.6243306519721 41.39746474575231,-87.62433840148628 41.397470163883234,-87.6243454199354 41.397476120184834,-87.62435164171818 41.39748255898387,-87.62435700867968 41.39748942009721,-87.62436147065489 41.39749663939436,-87.62436498593763 41.397504149396866,-87.6243675216704 41.39751187990905,-87.6243690541514 41.397519758674136,-87.62436956905626 41.39752771204957,-87.62436906157178 41.39753566569541,-87.62436753644107 41.397543545269144,-87.6243650079191 41.39755127712062,-87.62431827109069 41.397669293489656,-87.6243148122446 41.39767671371113,-87.62431042933302 41.397683851719385,-87.62430516231241 41.39769064244089,-87.62429905919917 41.397697023968156,-87.62429217563205 41.39770293812402,-87.62428457436476 41.3977083309922,-87.6242763246941 41.397713153408624,-87.62426750182793 41.39771736140979,-87.62425818619981 41.39772091663344,-87.6242484627356 41.39772378666837,-87.62423842007912 41.39772594534997,-87.62422814978432 41.39772737299857,-87.62421774548025 41.39772805659901,-87.62420730201784 41.39772798991927,-87.62419691460487 41.3977271735672,-87.62418667793823 41.3977256149851,-87.62418671102395 41.3977256212699,-87.62407665709884 41.39770472967701,-87.62406629597992 41.397702346179805,-87.62405629881992 41.397699202590374,-87.62405211939647 41.3976975050517,-87.62403011344777 41.397692436957406,-87.62402019045922 41.39769240856049,-87.62400961007666 41.39769160064994,-87.6239991838944 41.39769002289897,-87.62394034597573 41.397678853666775,-87.62392998485687 41.397676470172,-87.6239199876966 41.39767332658449,-87.62391045256837 41.39766945374326,-87.62390147301304 41.3976648896414,-87.62389709425379 41.39766215257271,-87.62389409652785 41.397661423260395,-87.6238838004308 41.39765847960019,-87.62387394964163 41.3976547712504,-87.62386464402535 41.397650335805466,-87.62385597792006 41.3976452182309,-87.62384803918056 41.397639470407476,-87.62384090828776 41.3976331506053,-87.623839286541 41.39763137916821,-87.62383675849487 41.397629285732634,-87.62383035228544 41.397622804159624,-87.62382482029217 41.39761588114358,-87.62382021574966 41.397608583305356,-87.62381658296758 41.39760098087273,-87.62381395690413 41.39759314700461,-87.62381236282981 41.39758515708701,-87.62381181608414 41.397577088007544,-87.62381232192811 41.39756901741562,-87.6238138754936 41.39756102297521,-87.62381646183007 41.39755318161735,-87.62386309049305 41.397436588717746,-87.62386670290655 41.39742894262843,-87.62387129711452 41.39742160135193,-87.62387682851207 41.39741463616366,-87.6238832433953 41.39740811468775,-87.62389047948277 41.39740210024026,-87.62389846652009 41.397396651214606,-87.6239071269621 41.39739182051462,-87.62391637672572 41.39738765504086,-87.6239261260062 41.39738419523529,-87.62393628014925 41.3973814746886,-87.62394674056974 41.397379519814216,-87.62395740570904 41.39737834959169,-87.62396817202092 41.397377975382526,-87.62397893497695 41.397378400819846,-87.62398959008122 41.3973796217732,-87.62400003388497 41.39738162638845,-87.62400459928844 41.397382685573255,-87.6240097418674 41.39738137496058,-87.62401988537127 41.39737957809739,-87.62403020992313 41.39737851619985,-87.6240406218205 41.39737819890541,-87.62405102656818 41.39737862909373,-87.62406132973588 41.39737980286054,-87.62407143781516 41.39738170955312,-87.62416781499175 41.397403611535694,-87.62422983409283 41.39741559488651,-87.62424038602445 41.3974180677156,-87.62425055272993 41.397421329650655,-87.62426023043129 41.39742534739508,-87.62426932034207 41.39743007993722,-87.62427772967571 41.39743547896907,-87.6242853725927 41.397441489379304,-87.62429217107669 41.397448049815836,-87.62429313249845 41.397449200566754,-87.6242938761274 41.39744938185984))",
			fwdDiff: "POLYGON((-87.62446704222914 41.397301659153754,-87.62449506935019 41.39731775684101,-87.62452090105552 41.397335817239,-87.62454429589731 41.39735567153923,-87.6245650352042 41.3973771341657,-87.6245829251253 41.397400004509386,-87.62459779844208 41.397424068803204,-87.62460951613143 41.397449102120135,-87.62461796866529 41.39747487047549,-87.6246230770344 41.39750113301402,-87.62462479348707 41.39752764426096,-87.6246231019755 41.39755415641672,-87.62461801830612 41.397580421672856,-87.62460958999183 41.39760619452835,-87.62456285356899 41.397724210992784,-87.62455132416264 41.39774894508767,-87.62453671452448 41.3977727384771,-87.62451915784027 41.397795374248304,-87.62449881416285 41.39781664604174,-87.62447586895344 41.39783635993238,-87.62445053139074 41.39785433619775,-87.62442303246401 41.397870410956294,-87.62439362286723 41.39788443766158,-87.62436257071374 41.39789628843825,-87.62433015909178 41.39790585524787,-87.62429668348378 41.3979130508739,-87.62426244907233 41.39791780971686,-87.62422776795812 41.397920088392326,-87.62419295631425 41.397919866126564,-87.62415833150395 41.39791714494585,-87.62412420918706 41.397911949658024,-87.62412431963479 41.397911970638134,-87.6240142654138 41.3978910789883,-87.62397972826206 41.397883133976144,-87.62397604150179 41.39788197468104,-87.62397154652997 41.39788163144765,-87.62393679249253 41.397876372263475,-87.6238779544157 41.39786520300083,-87.62384341726417 41.3978572579968,-87.6238100933212 41.39784677934415,-87.62377830950128 41.397833869840746,-87.62377439835728 41.39783188189482,-87.62374847092677 41.39782212145373,-87.62371745215871 41.397807336601254,-87.62368856511434 41.397790277980455,-87.62366210264413 41.397771118528894,-87.62364238065591 41.39775363973907,-87.62363973273321 41.3977514470298,-87.62363850541901 41.39775020527393,-87.62363833301796 41.39775005248186,-87.62363806435583 41.39774975902078,-87.6236183787311 41.3977298417487,-87.62359993880196 41.397706764993615,-87.62358459039267 41.39768243883536,-87.62357248119949 41.39765709736778,-87.62356372774683 41.3976309844548,-87.62355841426607 41.39760435138386,-87.62355659188512 41.3975774544476,-87.62355827813658 41.39755055247751,-87.62356345678923 41.39752390435329,-87.62357207800432 41.39749776651146,-87.6236187070669 41.39738117370699,-87.62363074852786 41.397355686766915,-87.62364606262244 41.39733121587506,-87.62366450066513 41.39730799861518,-87.62368588364099 41.39728626039952,-87.62371000394393 41.3972662122803,-87.62373662739223 41.397248048900735,-87.62376549550245 41.39723194660534,-87.62379632799865 41.3972180617278,-87.62382882553379 41.39720652907334,-87.62386267259592 41.397197460609846,-87.62389754057125 41.397190944380924,-87.62393309093454 41.39718704365108,-87.62396897853569 41.39718579629156,-87.62400485495054 41.39718721441263,-87.62400561478837 41.397187301481495,-87.62400775375312 41.39718708148513,-87.6240424599759 41.39718602384076,-87.6240771423665 41.39718745779836,-87.62411148616091 41.39719137034386,-87.62414517966795 41.39719772596847,-87.62423643421413 41.3972184638187,-87.6242932704773 41.39722944573794,-87.62432844349291 41.39723768847929,-87.62436233243534 41.39724856156757,-87.62439459138113 41.397261954015285,-87.62442489104455 41.39727772911831,-87.62444917095299 41.39729331748811,-87.62446704222914 41.397301659153754),(-87.62429313249845 41.397449200566754,-87.62429217107669 41.397448049815836,-87.6242853725927 41.397441489379304,-87.62427772967571 41.39743547896907,-87.62426932034207 41.39743007993722,-87.62426023043129 41.39742534739508,-87.62425055272993 41.397421329650655,-87.62424038602445 41.3974180677156,-87.62422983409283 41.39741559488651,-87.62416781499175 41.397403611535694,-87.62407143781516 41.39738170955312,-87.62406132973588 41.39737980286054,-87.62405102656818 41.39737862909373,-87.6240406218205 41.39737819890541,-87.62403020992313 41.39737851619985,-87.62401988537127 41.39737957809739,-87.6240097418674 41.39738137496058,-87.62400459928844 41.397382685573255,-87.62400003388497 41.39738162638845,-87.62398959008122 41.3973796217732,-87.62397893497695 41.397378400819846,-87.62396817202092 41.397377975382526,-87.62395740570904 41.39737834959169,-87.62394674056974 41.397379519814216,-87.62393628014925 41.3973814746886,-87.6239261260062 41.39738419523529,-87.62391637672572 41.39738765504086,-87.6239071269621 41.39739182051462,-87.62389846652009 41.397396651214606,-87.62389047948277 41.39740210024026,-87.6238832433953 41.39740811468775,-87.62387682851207 41.39741463616366,-87.62387129711452 41.39742160135193,-87.62386670290655 41.39742894262843,-87.62386309049305 41.397436588717746,-87.62381646183007 41.39755318161735,-87.6238138754936 41.39756102297521,-87.62381232192811 41.39756901741562,-87.62381181608414 41.397577088007544,-87.62381236282981 41.39758515708701,-87.62381395690413 41.39759314700461,-87.62381658296758 41.39760098087273,-87.62382021574966 41.397608583305356,-87.62382482029217 41.39761588114358,-87.62383035228544 41.397622804159624,-87.62383675849487 41.397629285732634,-87.623839286541 41.39763137916821,-87.62384090828776 41.3976331506053,-87.62384803918056 41.397639470407476,-87.62385597792006 41.3976452182309,-87.62386464402535 41.397650335805466,-87.62387394964163 41.3976547712504,-87.6238838004308 41.39765847960019,-87.62389409652785 41.397661423260395,-87.62389709425379 41.39766215257271,-87.62390147301304 41.3976648896414,-87.62391045256837 41.39766945374326,-87.6239199876966 41.39767332658449,-87.62392998485687 41.397676470172,-87.62394034597573 41.397678853666775,-87.6239991838944 41.39769002289897,-87.62400961007666 41.39769160064994,-87.62402019045922 41.39769240856049,-87.62403011344777 41.397692436957406,-87.62405211939647 41.3976975050517,-87.62405629881992 41.397699202590374,-87.62406629597992 41.397702346179805,-87.62407665709884 41.39770472967701,-87.62418671102395 41.3977256212699,-87.62418667793823 41.3977256149851,-87.62419691460487 41.3977271735672,-87.62420730201784 41.39772798991927,-87.62421774548025 41.39772805659901,-87.62422814978432 41.39772737299857,-87.62423842007912 41.39772594534997,-87.6242484627356 41.39772378666837,-87.62425818619981 41.39772091663344,-87.62426750182793 41.39771736140979,-87.6242763246941 41.397713153408624,-87.62428457436476 41.3977083309922,-87.62429217563205 41.39770293812402,-87.62429905919917 41.397697023968156,-87.62430516231241 41.39769064244089,-87.62431042933302 41.397683851719385,-87.6243148122446 41.39767671371113,-87.62431827109069 41.397669293489656,-87.6243650079191 41.39755127712062,-87.62436753644107 41.397543545269144,-87.62436906157178 41.39753566569541,-87.62436956905626 41.39752771204957,-87.6243690541514 41.397519758674136,-87.6243675216704 41.39751187990905,-87.62436498593763 41.397504149396866,-87.62436147065489 41.39749663939436,-87.62435700867968 41.39748942009721,-87.62435164171818 41.39748255898387,-87.6243454199354 41.397476120184834,-87.62433840148628 41.397470163883234,-87.6243306519721 41.39746474575231,-87.62432224382728 41.39745991643503,-87.6243132556424 41.39745572107071,-87.62430377142961 41.397452198873204,-87.62429387983731 41.3974493827643,-87.6242938761274 41.39744938185984,-87.62429313249845 41.397449200566754))",
		},

		// Reproduces https://github.com/peterstace/simplefeatures/issues/573.
		{
			input1: "MULTIPOLYGON(((-84.68833861614341 33.586881506063484,-84.69297198085386 33.58116806479595,-84.69314358695425 33.58002533114709,-84.69417322355658 33.57959680212774,-84.71356471290028 33.55516713116744,-84.71373631900066 33.55373828152835,-84.71476595560297 33.553595395264594,-84.71631041050647 33.5517378523297,-84.86389165683904 33.36492286730951,-84.02061927953866 33.34788067611323,-84.01787358193248 33.3506019220889,-84.01787358193248 33.351461245252004,-84.01684394533015 33.3516044649545,-84.01547109652708 33.352893431673934,-83.96776460061955 33.400285501397065,-83.9672497823184 33.401430609399185,-83.96587693351529 33.402146294237056,-83.79856098563825 33.5681685767283,-83.79856098563825 33.569740047434045,-83.80422398695102 33.56988290698898,-84.68833861614341 33.586881506063484),(-84.2308367525124 33.4876996684837,-84.22860587320739 33.48784266364406,-84.22740463050468 33.48669869575038,-84.22912069150854 33.48555471274636,-84.2310083586128 33.48598370814359,-84.23169478301435 33.48727068158603,-84.2308367525124 33.4876996684837),(-84.14554852062021 33.495135103738065,-84.1443472779175 33.494849137265035,-84.14572012672059 33.49399123217892,-84.1469213694233 33.49413421695025,-84.1469213694233 33.495135103738065,-84.14554852062021 33.495135103738065),(-84.14846582432678 33.49913853518449,-84.1443472779175 33.4984236502898,-84.14383245961633 33.49742280151938,-84.14486209621866 33.49556405167663,-84.14812261212599 33.49542106926658,-84.14983867312988 33.496135978955564,-84.14846582432678 33.49913853518449),(-84.43590604247454 33.501426127175165,-84.43521961807299 33.5011401814824,-84.43539122417339 33.50042531311787,-84.43401837537029 33.4998534141756,-84.43453319367146 33.49856662774103,-84.43693567907687 33.499281511454996,-84.4377937095788 33.500282338736525,-84.43727889127763 33.5011401814824,-84.43590604247454 33.501426127175165),(-84.35713884239703 33.51643694937166,-84.35696723629664 33.516579897065206,-84.35679563019626 33.51629400144192,-84.35696723629664 33.516151053275976,-84.35713884239703 33.51643694937166)))",
			input2: "POLYGON((-84.91779342802035 33.944641723933344,-84.8848339500958 33.84748446314561,-84.8848343784715 33.84747361663,-84.88437073489857 33.84611900705978,-84.88229142057187 33.83998964770643,-84.88222714480544 33.83985616244309,-84.88125822370353 33.837025302976684,-84.88126237687771 33.83699770204221,-84.85536477724138 33.761309422514195,-84.85534194530501 33.76130671052121,-84.85449468317503 33.758831297387545,-84.85450724082959 33.75874254545525,-84.81368850762794 33.639440725461384,-84.81364280754319 33.63943194420541,-84.813631809119 33.6393995771171,-84.81245474992689 33.63920365933259,-84.8092293182866 33.63858389372383,-84.80916438043222 33.638655987723595,-84.80841510657781 33.63853127345728,-84.8083717818748 33.63844108760532,-84.79629514127976 33.63651549182358,-84.78394716531548 33.6344387969546,-84.7839497741841 33.6344018854751,-84.78350448493246 33.63432738751653,-84.78350322593263 33.634299607151,-84.75514064693321 33.62957592907422,-84.74704761565542 33.62822806594716,-84.7156639365883 33.6229775012537,-84.7156482619808 33.623012293828445,-84.71102277801262 33.62225009809144,-84.71102219943961 33.62225000275312,-84.71102210848343 33.622249987765194,-84.701763091656 33.62072426998016,-84.7017665676303 33.6206924426037,-84.69137613551851 33.61887890940932,-84.68574407498753 33.59073059778281,-84.6885541839693 33.587218663264146,-84.6885542340275 33.587218600703885,-84.68855451712274 33.58721824690553,-84.68855455244871 33.58721820275688,-84.69431943910551 33.580013533978004,-84.69424890091736 33.579543341019956,-84.69738559738585 33.57558565245055,-84.69863491361976 33.57402573125029,-84.70667667404871 33.56388161091509,-84.71550795902019 33.55280561973021,-84.77318683091957 33.48046612680534,-84.779117011864 33.4799318239731,-84.799023055071 33.4593348340223,-84.85210310244086 33.38149118368702,-84.8616600590725 33.36950507206812,-84.8616600589832 33.3695050720681,-84.86044709861349 33.369473606293106,-84.86389165683013 33.36492286730951,-84.86389165676792 33.36492286730951,-84.85608672289517 33.36476404808989,-84.8568535426636 33.3347906275748,-84.7156235981867 33.3312091337486,-84.7087594453567 33.3299197599267,-84.7053273689417 33.330063024627,-84.7048125574795 33.3292034328914,-84.6114600789917 33.3130128714351,-84.5967021504072 33.3128695787046,-84.5965305465864 33.3104335662518,-84.5670146894175 33.3052747269735,-84.5440197774371 33.302981811551,-84.40155431850484 33.300961131709805,-84.44207668304443 33.26879753476913,-84.37753200531006 33.12479312032584,-84.37375545501709 33.12234910903114,-84.24037456512451 33.119689372392806,-84.16698932647705 33.12119896253218,-84.03944492340088 33.223072176711646,-83.95395755767822 33.221707944680595,-83.76710414886475 33.369208714244394,-83.7661600112915 33.40102923261309,-83.87138859428424 33.493634468317246,-83.81152280628423 33.55489126855177,-83.80353783116196 33.554698369639816,-83.80353783116196 33.56317098221176,-83.80261577856116 33.564085642231284,-83.8021656961152 33.56413252691967,-83.8021656961152 33.56453211612927,-83.79907591409109 33.567597125743774,-83.79905058771362 33.567744715646896,-83.79718785744615 33.569597025871445,-83.79718769453356 33.569597187642735,-83.79718769476483 33.569597187642735,-83.79719087268145 33.56959725026669,-83.79718760387247 33.569600479883945,-83.79838932427909 33.56962378359338,-83.79838932427909 33.569740047434045,-83.79990208203515 33.569766842039066,-83.8021656961152 33.569827630003665,-83.8021656961152 33.57585231701821,-83.80422389868532 33.57599523146829,-83.8033663142811 33.588999455846874,-83.80559603373209 33.58928524096701,-83.80559603373209 33.60085874271436,-83.80490996620871 33.601430233425845,-83.8033663142811 33.623715416663686,-83.80782575318308 33.62385825181623,-83.80834030382562 33.61228783729619,-83.80872638072529 33.61225210854396,-83.80877375697825 33.61374537703404,-83.80873745211973 33.61978823044128,-83.80748271942139 33.63399894237652,-83.80748271942139 33.634427395972715,-83.80864866822309 33.63456608921711,-83.8080883026123 33.7278375514204,-83.7866306304931 33.7321204874728,-83.7862873077392 33.7326915294623,-83.7861156463623 33.7322632483265,-83.7843990325927 33.7325487693212,-83.7323856353759 33.7426841490739,-83.732213973999 33.7442543120441,-83.7310123443603 33.74468253332,-83.7227725982666 33.7445397931323,-83.6676692962646 33.7552446479486,-83.6669826507568 33.7568145809432,-83.6649227142334 33.756957300699,-83.6647510528564 33.7561009786006,-83.6633777618408 33.7561009786006,-83.6475849151611 33.7592407844763,-83.6482715606689 33.8513856127035,-83.8275754226885 33.85513355791846,-83.8275501944802 33.85536729056149,-83.83129565682832 33.8557561295505,-83.78336906433105 33.88659122617693,-83.76688957214355 33.88587870592943,-83.76603126525879 33.98699697606636,-83.76931058875978 33.987022872220784,-83.7490430439492 34.00267818454505,-83.64711284637451 33.915478908148366,-83.64737033843994 33.9032981198015,-83.63123416900635 33.88299293807057,-83.55510234832764 33.8499956195856,-83.54068279266357 33.81491754086023,-83.5376787185669 33.809854272019145,-83.44995975494385 33.77461703628178,-83.44695568084717 33.77426030996451,-83.44704151153564 33.77347550683796,-83.431077003479 33.77725676486296,-83.42747211456299 33.779254343232445,-83.42764377593994 33.780253114947854,-83.42498302459717 33.78089517632915,-83.42506885528564 33.78167991148108,-83.42343807220459 33.78146589351633,-83.4171724319458 33.784961453304234,-83.38687419891357 33.7968024933337,-83.38550090789795 33.79808636311057,-83.38352680206299 33.79815768864465,-83.38318347930908 33.7991562398809,-83.3812952041626 33.799013590417516,-83.36799144744873 33.80422014179627,-83.30533504486084 33.836949966753664,-83.2505750656128 33.87893132165219,-83.23984622955322 33.89004686495183,-83.22971820831299 33.902514500728934,-83.22980403900146 33.90386802005763,-83.22825908660889 33.90429544274963,-83.22774410247803 33.90564893380339,-83.22611331939697 33.906931168663235,-83.20448398590088 33.93356879543034,-83.18628787994385 33.96446939658514,-83.1791639328003 33.985466855532835,-83.17221164703369 34.062294647776454,-83.17229747772217 34.06329010689234,-83.19590091705322 34.085827015146165,-83.2084321975708 34.095422931402005,-83.34893703460693 34.14700897818791,-83.3549451828003 34.15375688963505,-83.40016880669104 34.2758303445348,-83.3948564529419 34.28002078078785,-83.41331005096436 34.286970795241245,-83.65500926971436 34.287254457092466,-83.66779804229736 34.31306367686605,-83.68427753448486 34.342266740026,-83.74110848004483 34.29797689604221,-83.74019610452991 34.35426286075068,-83.74021213701118 34.35426506565391,-83.74019610013654 34.35526011593728,-83.74234899568803 34.35804422266522,-83.74277009895376 34.35865228381865,-83.75673750028577 34.376651317852854,-83.75924341504098 34.37989194609645,-83.75926101895718 34.37990323976539,-83.76078779192514 34.38187070944423,-83.76078779192514 34.38187350268598,-83.7611859990874 34.38238385745909,-83.78000721828539 34.406637743837614,-83.78025195195042 34.40681937989462,-83.78187253060538 34.40889636426392,-83.78189481419624 34.409043370162216,-83.78767406169894 34.41633178824798,-83.78772858868186 34.416401671737376,-83.78812134410857 34.41641153210681,-83.89028598851462 34.41907023133361,-83.8438367843628 34.4557218032159,-83.84907245635986 34.45579257605265,-84.04588222503662 34.455155618362454,-84.12012577056885 34.4543771079208,-84.24123287200928 34.360478065560066,-84.25857067108154 34.34977850540455,-84.26389217376709 34.34744000908005,-84.29745197296143 34.336455290073765,-84.42319393157959 34.34850297004076,-84.4279146194458 34.350345403781446,-84.45383548736572 34.3710345708007,-84.4707441329956 34.37974799906729,-84.58592891693115 34.376347744669054,-84.65811252593994 34.32100337957915,-84.6710729598999 34.30292528635627,-84.6710729598999 34.30164903857466,-84.67210292816162 34.301436328725515,-84.67321872711182 34.29994734469774,-84.7406816482544 34.20544930169491,-84.66401968871665 34.14012064778763,-84.67060928893602 34.13279527622302,-84.67078088856428 34.13180147295538,-84.67163888670557 34.13165950010613,-84.69119534965195 34.10966154253311,-84.74514070447981 34.110667743702116,-84.74671109181762 34.10990574958218,-84.79917532247262 34.110810185410706,-84.79917532247262 34.110806892155296,-84.79934698383909 34.11080987193478,-84.79934698383909 34.08436540087182,-84.80160005800013 34.083272148766966,-84.80341416254085 34.076002429774626,-84.81043107648884 34.04788332668517,-84.81790291084036 34.017941206551576,-84.80353483436086 33.999542448512706,-84.80339109812107 33.999358389894894,-84.80314453879552 33.99904266319485,-84.82699818663541 33.977723470613775,-84.82768462254448 33.97701273875627,-84.82768462254448 33.97516480812015,-84.82905749436262 33.97573340643993,-84.8311168020898 33.97402760007304,-84.84524080534064 33.961194578554384,-84.8510205396477 33.95594313153772,-84.86171282535385 33.94622815695766,-84.86345885103923 33.944641723933344,-84.91779342802035 33.944641723933344))",
			union:  "POLYGON((-84.86389165683904 33.36492286730951,-84.86389165662615 33.364922867579,-84.86044709861349 33.369473606293106,-84.8616600589832 33.3695050720681,-84.8616600590725 33.36950507206812,-84.85210310244086 33.38149118368702,-84.799023055071 33.4593348340223,-84.779117011864 33.4799318239731,-84.77318683091957 33.48046612680534,-84.71550795902019 33.55280561973021,-84.70667667404871 33.56388161091509,-84.69863491361976 33.57402573125029,-84.69738559738585 33.57558565245055,-84.69424890091736 33.579543341019956,-84.69431943910551 33.580013533978004,-84.68855455244871 33.58721820275688,-84.68855451712274 33.58721824690553,-84.6885542340275 33.587218600703885,-84.6885541839693 33.587218663264146,-84.68574407498753 33.59073059778281,-84.69137613551851 33.61887890940932,-84.7017665676303 33.6206924426037,-84.701763091656 33.62072426998016,-84.71102210848343 33.622249987765194,-84.71102219943961 33.62225000275312,-84.71102277801262 33.62225009809144,-84.7156482619808 33.623012293828445,-84.7156639365883 33.6229775012537,-84.74704761565542 33.62822806594716,-84.75514064693321 33.62957592907422,-84.78350322593263 33.634299607151,-84.78350448493246 33.63432738751653,-84.7839497741841 33.6344018854751,-84.78394716531548 33.6344387969546,-84.79629514127976 33.63651549182358,-84.8083717818748 33.63844108760532,-84.80841510657781 33.63853127345728,-84.80916438043222 33.638655987723595,-84.8092293182866 33.63858389372383,-84.81245474992689 33.63920365933259,-84.813631809119 33.6393995771171,-84.81364280754319 33.63943194420541,-84.81368850762794 33.639440725461384,-84.85450724082959 33.75874254545525,-84.85449468317503 33.758831297387545,-84.85534194530501 33.76130671052121,-84.85536477724138 33.761309422514195,-84.88126237687771 33.83699770204221,-84.88125822370353 33.837025302976684,-84.88222714480544 33.83985616244309,-84.88229142057187 33.83998964770643,-84.88437073489857 33.84611900705978,-84.8848343784715 33.84747361663,-84.8848339500958 33.84748446314561,-84.91779342802035 33.944641723933344,-84.86345885103923 33.944641723933344,-84.86171282535385 33.94622815695766,-84.8510205396477 33.95594313153772,-84.84524080534064 33.961194578554384,-84.8311168020898 33.97402760007304,-84.82905749436262 33.97573340643993,-84.82768462254448 33.97516480812015,-84.82768462254448 33.97701273875627,-84.82699818663541 33.977723470613775,-84.80314453879552 33.99904266319485,-84.80339109812107 33.999358389894894,-84.80353483436086 33.999542448512706,-84.81790291084036 34.017941206551576,-84.81043107648884 34.04788332668517,-84.80341416254085 34.076002429774626,-84.80160005800013 34.083272148766966,-84.79934698383909 34.08436540087182,-84.79934698383909 34.11080987193478,-84.79917532247262 34.110806892155296,-84.79917532247262 34.110810185410706,-84.74671109181762 34.10990574958218,-84.74514070447981 34.110667743702116,-84.69119534965195 34.10966154253311,-84.67163888670557 34.13165950010613,-84.67078088856428 34.13180147295538,-84.67060928893602 34.13279527622302,-84.66401968871665 34.14012064778763,-84.7406816482544 34.20544930169491,-84.67321872711182 34.29994734469774,-84.67210292816162 34.301436328725515,-84.6710729598999 34.30164903857466,-84.6710729598999 34.30292528635627,-84.65811252593994 34.32100337957915,-84.58592891693115 34.376347744669054,-84.4707441329956 34.37974799906729,-84.45383548736572 34.3710345708007,-84.4279146194458 34.350345403781446,-84.42319393157959 34.34850297004076,-84.29745197296143 34.336455290073765,-84.26389217376709 34.34744000908005,-84.25857067108154 34.34977850540455,-84.24123287200928 34.360478065560066,-84.12012577056885 34.4543771079208,-84.04588222503662 34.455155618362454,-83.84907245635986 34.45579257605265,-83.8438367843628 34.4557218032159,-83.89028598851462 34.41907023133361,-83.78812134410857 34.41641153210681,-83.78772858868186 34.416401671737376,-83.78767406169894 34.41633178824798,-83.78189481419624 34.409043370162216,-83.78187253060538 34.40889636426392,-83.78025195195042 34.40681937989462,-83.78000721828539 34.406637743837614,-83.7611859990874 34.38238385745909,-83.76078779192514 34.38187350268598,-83.76078779192514 34.38187070944423,-83.75926101895718 34.37990323976539,-83.75924341504098 34.37989194609645,-83.75673750028577 34.376651317852854,-83.74277009895376 34.35865228381865,-83.74234899568803 34.35804422266522,-83.74019610013654 34.35526011593728,-83.74021213701118 34.35426506565391,-83.74019610452991 34.35426286075068,-83.74110848004483 34.29797689604221,-83.68427753448486 34.342266740026,-83.66779804229736 34.31306367686605,-83.65500926971436 34.287254457092466,-83.41331005096436 34.286970795241245,-83.3948564529419 34.28002078078785,-83.40016880669104 34.2758303445348,-83.3549451828003 34.15375688963505,-83.34893703460693 34.14700897818791,-83.2084321975708 34.095422931402005,-83.19590091705322 34.085827015146165,-83.17229747772217 34.06329010689234,-83.17221164703369 34.062294647776454,-83.1791639328003 33.985466855532835,-83.18628787994385 33.96446939658514,-83.20448398590088 33.93356879543034,-83.22611331939697 33.906931168663235,-83.22774410247803 33.90564893380339,-83.22825908660889 33.90429544274963,-83.22980403900146 33.90386802005763,-83.22971820831299 33.902514500728934,-83.23984622955322 33.89004686495183,-83.2505750656128 33.87893132165219,-83.30533504486084 33.836949966753664,-83.36799144744873 33.80422014179627,-83.3812952041626 33.799013590417516,-83.38318347930908 33.7991562398809,-83.38352680206299 33.79815768864465,-83.38550090789795 33.79808636311057,-83.38687419891357 33.7968024933337,-83.4171724319458 33.784961453304234,-83.42343807220459 33.78146589351633,-83.42506885528564 33.78167991148108,-83.42498302459717 33.78089517632915,-83.42764377593994 33.780253114947854,-83.42747211456299 33.779254343232445,-83.431077003479 33.77725676486296,-83.44704151153564 33.77347550683796,-83.44695568084717 33.77426030996451,-83.44995975494385 33.77461703628178,-83.5376787185669 33.809854272019145,-83.54068279266357 33.81491754086023,-83.55510234832764 33.8499956195856,-83.63123416900635 33.88299293807057,-83.64737033843994 33.9032981198015,-83.64711284637451 33.915478908148366,-83.7490430439492 34.00267818454505,-83.76931058875978 33.987022872220784,-83.76603126525879 33.98699697606636,-83.76688957214355 33.88587870592943,-83.78336906433105 33.88659122617693,-83.83129565682832 33.8557561295505,-83.8275501944802 33.85536729056149,-83.8275754226885 33.85513355791846,-83.6482715606689 33.8513856127035,-83.6475849151611 33.7592407844763,-83.6633777618408 33.7561009786006,-83.6647510528564 33.7561009786006,-83.6649227142334 33.756957300699,-83.6669826507568 33.7568145809432,-83.6676692962646 33.7552446479486,-83.7227725982666 33.7445397931323,-83.7310123443603 33.74468253332,-83.732213973999 33.7442543120441,-83.7323856353759 33.7426841490739,-83.7843990325927 33.7325487693212,-83.7861156463623 33.7322632483265,-83.7862873077392 33.7326915294623,-83.7866306304931 33.7321204874728,-83.8080883026123 33.7278375514204,-83.80864866822309 33.63456608921711,-83.80748271942139 33.634427395972715,-83.80748271942139 33.63399894237652,-83.80873745211973 33.61978823044128,-83.80877375697825 33.61374537703404,-83.80872638072529 33.61225210854396,-83.80834030382562 33.61228783729619,-83.80782575318308 33.62385825181623,-83.8033663142811 33.623715416663686,-83.80490996620871 33.601430233425845,-83.80559603373209 33.60085874271436,-83.80559603373209 33.58928524096701,-83.8033663142811 33.588999455846874,-83.80422389868532 33.57599523146829,-83.8021656961152 33.57585231701821,-83.8021656961152 33.569830982841964,-83.79896561366643 33.56975025491535,-83.79838932427909 33.569740047434045,-83.79838932427909 33.56962378359338,-83.79718760387247 33.569600479883945,-83.79719087268145 33.56959725026669,-83.79718769476483 33.569597187642735,-83.79718769453356 33.569597187642735,-83.79718785744615 33.569597025871445,-83.79856098563825 33.56823157890671,-83.79856098563825 33.5681685767283,-83.79906340100439 33.56767004597314,-83.79907591409109 33.567597125743774,-83.8021656961152 33.56453211612927,-83.8021656961152 33.56413252691967,-83.80261577856116 33.564085642231284,-83.80353783116196 33.56317098221176,-83.80353783116196 33.554698369639816,-83.81152280628423 33.55489126855177,-83.87138859428424 33.493634468317246,-83.7661600112915 33.40102923261309,-83.76710414886475 33.369208714244394,-83.95395755767822 33.221707944680595,-84.03944492340088 33.223072176711646,-84.16698932647705 33.12119896253218,-84.24037456512451 33.119689372392806,-84.37375545501709 33.12234910903114,-84.37753200531006 33.12479312032584,-84.44207668304443 33.26879753476913,-84.40155431850484 33.300961131709805,-84.5440197774371 33.302981811551,-84.5670146894175 33.3052747269735,-84.5965305465864 33.3104335662518,-84.5967021504072 33.3128695787046,-84.6114600789917 33.3130128714351,-84.7048125574795 33.3292034328914,-84.7053273689417 33.330063024627,-84.7087594453567 33.3299197599267,-84.7156235981867 33.3312091337486,-84.8568535426636 33.3347906275748,-84.85608672289517 33.36476404808989,-84.8638916464249 33.36492286709905,-84.86389165683904 33.36492286730951))",
		},

		// Reproduces https://github.com/peterstace/simplefeatures/issues/677.
		{
			input1:  "POLYGON((-80.19201390445232 25.76988776280538,-80.19201993942261 25.769890782125756,-80.19203066825867 25.76989198985386,-80.1920534670353 25.76989138598981,-80.19206285476685 25.76988715894131,-80.1920735836029 25.769885951213148,-80.1920809596777 25.769872666202453,-80.1920809596777 25.769864212103958,-80.19207760691643 25.76985273868362,-80.19207961857319 25.7698460961766,-80.19207626581192 25.769830999568352,-80.19208297133446 25.7698195261448,-80.192089676857 25.76981771455153,-80.19208498299122 25.76981288363592,-80.1920910179615 25.769804429533195,-80.19209168851376 25.76979295610708,-80.19207894802094 25.769785105867502,-80.19207626581192 25.76978087881521,-80.19207425415516 25.769754308768835,-80.19206956028938 25.769744646932292,-80.19206888973713 25.76972109620247,-80.19207492470741 25.769682448840825,-80.19206821918488 25.769679429515154,-80.19206620752811 25.769676410189426,-80.1920635253191 25.769651651715503,-80.19206754863262 25.769616023658585,-80.19207492470741 25.769599115424455,-80.19206821918488 25.769577376262735,-80.19206084311008 25.769562883486046,-80.19205883145332 25.76954778684177,-80.19206017255783 25.769536917256723,-80.19206553697586 25.769528463134314,-80.19206553697586 25.76949947756721,-80.19206888973713 25.769496458236908,-80.19207626581192 25.769494646638677,-80.19207827746868 25.76949102344214,-80.19207693636417 25.76947773838723,-80.19207894802094 25.76946264173215,-80.19207090139389 25.76945720693583,-80.19206821918488 25.76944150641177,-80.1920635253191 25.76943727934727,-80.19206821918488 25.76942580588563,-80.19206888973713 25.76941252082345,-80.19206084311008 25.769408293757895,-80.19205816090107 25.76939983962632,-80.19205279648304 25.769396216426884,-80.1920286566019 25.769395008693728,-80.19201926887035 25.769398028026618,-80.19200183451176 25.769399235759764,-80.19199848175049 25.769401651225998,-80.19197769463062 25.769398028026618,-80.19195556640625 25.76939742416005,-80.19195556640625 25.76939561256031,-80.19194282591343 25.76939198936077,-80.19193679094315 25.769384742961304,-80.19193813204765 25.769373873361324,-80.19194349646568 25.769366023094033,-80.19194349646568 25.769361192160066,-80.19194148480892 25.76935756895945,-80.19193209707737 25.769354549625533,-80.19192807376385 25.769350926424714,-80.1919274032116 25.769339452954757,-80.19192136824131 25.769335225886593,-80.19192203879356 25.76932496014902,-80.19192673265934 25.76931952534642,-80.19192203879356 25.769315298277547,-80.1919186860323 25.76929537066513,-80.19192136824131 25.769286916525516,-80.19192606210709 25.76928148172117,-80.1919274032116 25.76927242371336,-80.19193410873413 25.769258534766717,-80.19195556640625 25.769244645818425,-80.19195556640625 25.769243438083727,-80.1919562369585 25.769242834216378,-80.19197568297386 25.769232568470795,-80.19198305904865 25.769225925929057,-80.19198641180992 25.76921324471201,-80.19199848175049 25.76920116736116,-80.19200049340725 25.769184259067927,-80.19200585782528 25.76917338944956,-80.19200317561626 25.769155877284554,-80.19201658666134 25.76914198832427,-80.19202262163162 25.76913051483411,-80.19203133881092 25.769091263412033,-80.19203133881092 25.769087036335034,-80.1920273154974 25.769083413126044,-80.19202530384064 25.76907737444419,-80.19203133881092 25.76907012802552,-80.19203335046768 25.769053823581945,-80.19203670322895 25.769050200371957,-80.19204340875149 25.76902725337277,-80.19204340875149 25.769012156660384,-80.1920460909605 25.769003702500598,-80.19204206764698 25.768991625128436,-80.19203402101994 25.768986794179234,-80.19200786948204 25.768981963229805,-80.19199579954147 25.76897713228021,-80.19199311733246 25.76897713228021,-80.19199311733246 25.768976528411503,-80.1919723302126 25.768971093592942,-80.19195556640625 25.76896988585544,-80.1919475197792 25.768971093592942,-80.19194014370441 25.768966262642895,-80.19191399216652 25.768960827823847,-80.19190192222595 25.768962035561444,-80.19188582897186 25.768955393004575,-80.19186101853848 25.768948750447326,-80.19185096025467 25.76894331562748,-80.19184827804565 25.76893546533173,-80.19184358417988 25.768932445987073,-80.19182346761227 25.768927615035437,-80.19180469214916 25.768925803428548,-80.19180066883564 25.76892278408362,-80.19178256392479 25.76892157634565,-80.19177317619324 25.76891734926261,-80.19175842404366 25.768916141524553,-80.19174568355083 25.76891131057227,-80.19173093140125 25.768910102834184,-80.19170343875885 25.76890225253622,-80.19168801605701 25.76890225253622,-80.19166387617588 25.768898629321612,-80.1916591823101 25.768894402237752,-80.19165180623531 25.76889319449947,-80.19164443016052 25.76889319449947,-80.19164174795151 25.768902856405326,-80.19163303077221 25.768910102834184,-80.19160620868206 25.76892338795263,-80.19159279763699 25.768934861462807,-80.1915592700243 25.768945127234122,-80.19154720008373 25.76894573110299,-80.19154116511345 25.768948750447326,-80.19153781235218 25.768953581398083,-80.19153110682964 25.76895478913573,-80.19152708351612 25.76895841234864,-80.19151635468006 25.768962035561444,-80.19151166081429 25.768962035561444,-80.19150897860527 25.76895720461103,-80.19151233136654 25.768953581398083,-80.19151367247105 25.7689475427096,-80.19149824976921 25.76893848467631,-80.19149489700794 25.76892338795263,-80.19149288535118 25.768928822773365,-80.19148081541061 25.768939088545235,-80.19147276878357 25.768940900151918,-80.1914680749178 25.76893606920065,-80.19146740436554 25.768927011166472,-80.19146338105202 25.768933653724932,-80.19145533442497 25.76893003051129,-80.19145533442497 25.76892338795263,-80.19145801663399 25.76891855700062,-80.19144997000694 25.768913726048442,-80.19145600497723 25.76891131057227,-80.19145734608173 25.768908895096065,-80.19144997000694 25.76890406414349,-80.19145600497723 25.768901648667125,-80.19145734608173 25.76889923319072,-80.1914506405592 25.768898629321612,-80.19144527614117 25.768894402237752,-80.1914419233799 25.768901044798035,-80.19143521785736 25.768905271881664,-80.19142918288708 25.768895609976003,-80.19143052399158 25.768888363546253,-80.1914244890213 25.768885948069574,-80.19141644239426 25.768885948069574,-80.1914244890213 25.768873266816176,-80.19143521785736 25.768873870685436,-80.19143991172314 25.768880513246867,-80.19144460558891 25.768882324854445,-80.1914519816637 25.76887990937765,-80.19145600497723 25.768874474554664,-80.19146136939526 25.768877493900796,-80.19147880375385 25.768878097770028,-80.19148282706738 25.76886964360069,-80.19148752093315 25.768866624254365,-80.19148349761963 25.768861189430748,-80.191460698843 25.768859377822825,-80.19144661724567 25.768861189430748,-80.1914419233799 25.768859377822825,-80.19143454730511 25.768859981692145,-80.19142851233482 25.768850923652174,-80.19141912460327 25.768846092697213,-80.19139431416988 25.768846696566612,-80.19139230251312 25.76885152752152,-80.19139565527439 25.76885454686823,-80.19139565527439 25.768858773953518,-80.1913882791996 25.768866020385065,-80.19140034914017 25.768873870685436,-80.19140303134918 25.76888353259284,-80.19141040742397 25.76888353259284,-80.19141241908073 25.768886551938746,-80.19139900803566 25.768886551938746,-80.19139565527439 25.76888896741541,-80.19139431416988 25.768894402237752,-80.19139029085636 25.768895006106877,-80.19138425588608 25.76889017515374,-80.19137419760227 25.768886551938746,-80.19137218594551 25.76888353259284,-80.19137218594551 25.768874474554664,-80.19138492643833 25.768866020385065,-80.19137687981129 25.768856962345552,-80.19137755036354 25.768846092697213,-80.19137553870678 25.768844281089084,-80.19137151539326 25.768842469480898,-80.19136145710945 25.768842469480898,-80.19134938716888 25.768836430786763,-80.1913420110941 25.768837034656183,-80.19133396446705 25.76883401530902,-80.19132725894451 25.768834619178453,-80.19132256507874 25.76883884626445,-80.1913172006607 25.768821937919597,-80.19131250679493 25.768826768875517,-80.19131183624268 25.768838242395034,-80.1913071423769 25.768834619178453,-80.1912970840931 25.768833411439587,-80.1912984251976 25.76882616500604,-80.1913071423769 25.768824957267064,-80.19130848348141 25.768822541789092,-80.19130177795887 25.76881831470252,-80.19130177795887 25.768803217963562,-80.19129976630211 25.768802010224345,-80.1912897080183 25.76880563344192,-80.1912883669138 25.768801406354747,-80.19129574298859 25.768792952179947,-80.19130110740662 25.76879114057098,-80.19130177795887 25.768788121222656,-80.19129373133183 25.768783894134856,-80.19129104912281 25.76877543995881,-80.19128434360027 25.768777251568014,-80.19128434360027 25.76877181674033,-80.1912883669138 25.76876879739149,-80.19128903746605 25.768764570303027,-80.19128032028675 25.768759135474742,-80.19127763807774 25.768762758693626,-80.19127897918224 25.768770609000796,-80.19127428531647 25.768776043828556,-80.19127160310745 25.76878510187425,-80.19126892089844 25.76878691348329,-80.19126825034618 25.768794763788886,-80.19127026200294 25.76880382183315,-80.1912796497345 25.768808652789804,-80.19128635525703 25.768809256659363,-80.19129239022732 25.768817106963485,-80.1912796497345 25.768822541789092,-80.19128501415253 25.768826768875517,-80.19128702580929 25.768835826917336,-80.19128300249577 25.768842469480898,-80.19128501415253 25.76884548882784,-80.19129440188408 25.768847904305343,-80.19129373133183 25.768852131390865,-80.19129037857056 25.76885515073756,-80.19128903746605 25.76884850817473,-80.19128434360027 25.768847904305343,-80.19127562642097 25.768856358476235,-80.19125819206238 25.76885756621489,-80.19125483930111 25.768849715913447,-80.19125282764435 25.768847904305343,-80.19124612212181 25.76884850817473,-80.1912434399128 25.768842469480898,-80.19122667610645 25.768839450133882,-80.19122198224068 25.76884126174209,-80.1912159472704 25.76884065787269,-80.19120991230011 25.768834619178453,-80.19120454788208 25.76883522304791,-80.19120052456856 25.768831599831234,-80.19119314849377 25.768834619178453,-80.19118912518024 25.768822541789092,-80.19118309020996 25.768820126311077,-80.19117705523968 25.76881227600718,-80.19117571413517 25.76880382183315,-80.1911623030901 25.7687989908763,-80.19115760922432 25.76879355604959,-80.19114889204502 25.7687923483103,-80.1911448687315 25.76878872509232,-80.19112542271614 25.768782082525746,-80.19110731780529 25.768782082525746,-80.19107446074486 25.768774232219332,-80.19105769693851 25.768773024479827,-80.1910362392664 25.768768193521726,-80.19103221595287 25.768764570303027,-80.19101612269878 25.768763362563416,-80.19100606441498 25.768759135474742,-80.19099332392216 25.768758531604924,-80.19098460674286 25.768754304516076,-80.19097320735455 25.768753700646233,-80.190963819623 25.768749473557207,-80.19095040857792 25.768747661947575,-80.19094370305538 25.76874283098844,-80.19093364477158 25.768740415508795,-80.19093833863735 25.768731357459693,-80.19093699753284 25.7687235071499,-80.19093297421932 25.768725922629887,-80.1909202337265 25.768724111019917,-80.19091688096523 25.768734376809466,-80.1909101754427 25.76873739615918,-80.19090078771114 25.768736792289225,-80.1908940821886 25.76873377293952,-80.19088134169579 25.768733169069574,-80.19087128341198 25.76872833810985,-80.1908565312624 25.76872894197982,-80.19085384905338 25.768741623248605,-80.19084580242634 25.768748869687336,-80.19082769751549 25.768747661947575,-80.19082367420197 25.768744642598136,-80.19080691039562 25.768744038728226,-80.1908028870821 25.768741623248605,-80.19079752266407 25.76874283098844,-80.1907928287983 25.7687470580777,-80.1907854527235 25.768747661947575,-80.19077941775322 25.768744642598136,-80.19076466560364 25.768743434858333,-80.19075863063335 25.76873981163888,-80.19074320793152 25.76873860389902,-80.19073650240898 25.768734980679405,-80.1907230913639 25.76873377293952,-80.19071504473686 25.76872894197982,-80.19071370363235 25.7687235071499,-80.19071772694588 25.768716260709642,-80.19071504473686 25.76870056008757,-80.1907217502594 25.768696936866807,-80.1907230913639 25.768681236242173,-80.19072711467743 25.768675801410073,-80.19072644412518 25.76865828917167,-80.1907130330801 25.768655269819938,-80.19070632755756 25.76865828917167,-80.19069962203503 25.768656477560647,-80.19069626927376 25.768658893041994,-80.19070163369179 25.768662516263916,-80.19070029258728 25.76866553561546,-80.19069157540798 25.7686685549669,-80.19068956375122 25.768667347226344,-80.19068486988544 25.768672178188552,-80.19067279994488 25.768671574318272,-80.19066207110882 25.768674593669584,-80.19065603613853 25.768671574318272,-80.19065134227276 25.768663120134246,-80.19065201282501 25.76865828917167,-80.19065871834755 25.768653458208867,-80.19065737724304 25.76864802337552,-80.19065134227276 25.76864440015313,-80.19064530730247 25.768643796282728,-80.1906406134367 25.768640173060234,-80.19063457846642 25.76864440015313,-80.19062787294388 25.76866493174517,-80.19061915576458 25.76867036657774,-80.19061110913754 25.768671574318272,-80.19059836864471 25.76866976270747,-80.19058965146542 25.768663120134246,-80.19058227539062 25.768661912393604,-80.19057422876358 25.768661308523296,-80.19057154655457 25.76865828917167,-80.19055344164371 25.768657081430984,-80.19054874777794 25.768652854338516,-80.19053131341934 25.768651042727406,-80.19052863121033 25.76864802337552,-80.19052863121033 25.768654062079246,-80.19053533673286 25.76865828917167,-80.19053600728512 25.768662516263916,-80.19054003059864 25.768667347226344,-80.19053734838963 25.76867398979934,-80.1905407011509 25.768679424631497,-80.19054740667343 25.768681840112382,-80.19055545330048 25.768690898165264,-80.19055478274822 25.76869270977576,-80.19053466618061 25.768696332996647,-80.19052997231483 25.768695125256375,-80.19052527844906 25.768689086554733,-80.19051723182201 25.76868787881439,-80.19051320850849 25.768680632371932,-80.19050918519497 25.76868365172302,-80.19048973917961 25.768684255593215,-80.19047565758228 25.768680632371932,-80.19046761095524 25.768680632371932,-80.19046559929848 25.768675801410073,-80.19046492874622 25.76866070465298,-80.1904595643282 25.76866553561546,-80.19045487046242 25.768661912393604,-80.19044950604439 25.76866493174517,-80.19044615328312 25.76866493174517,-80.1904434710741 25.768661912393604,-80.19043877720833 25.768664327874845,-80.19043140113354 25.768664327874845,-80.19042737782001 25.768671574318272,-80.19042938947678 25.768676405280328,-80.19042871892452 25.768682443982588,-80.19043140113354 25.768686067203816,-80.19043073058128 25.7686921059056,-80.19044280052185 25.768696332996647,-80.19043609499931 25.76870237169792,-80.19042067229748 25.76871082587917,-80.19041866064072 25.76871686457968,-80.19041396677494 25.76872048779985,-80.19041329622269 25.768725922629887,-80.1904072612524 25.768730149719747,-80.19040659070015 25.768744642598136,-80.19040122628212 25.768753096776358,-80.19039988517761 25.76876940126128,-80.19039586186409 25.768774836089072,-80.19039586186409 25.76878691348329,-80.19039116799831 25.7687923483103,-80.1903898268938 25.768805029572317,-80.19038647413254 25.76881227600718,-80.19038513302803 25.76882858048394,-80.19038110971451 25.76883522304791,-80.19037976861 25.768846696566612,-80.19037574529648 25.76885515073756,-80.19037373363972 25.768874474554664,-80.1903710514307 25.768878701639235,-80.19036903977394 25.76889077902289,-80.19036501646042 25.768896213845142,-80.19036300480366 25.768908895096065,-80.19035898149014 25.76891855700062,-80.1903522759676 25.768952977529256,-80.19034758210182 25.768955996873387,-80.1903335005045 25.768955996873387,-80.19032947719097 25.768960827823847,-80.19032746553421 25.76897713228021,-80.19032143056393 25.769000079289064,-80.19031874835491 25.769003702500598,-80.19029594957829 25.76900430636917,-80.19028656184673 25.76900189089485,-80.19028186798096 25.769004910237754,-80.19027650356293 25.769015176003013,-80.19027315080166 25.7690459732935,-80.19027717411518 25.769053823581945,-80.19028522074223 25.769059862265017,-80.19029460847378 25.769061070001584,-80.19030064344406 25.769064089342987,-80.19030131399632 25.76907737444419,-80.19029729068279 25.769097302093186,-80.19029729068279 25.769117229738857,-80.19030667841434 25.76912206068275,-80.1903086900711 25.769127495494423,-80.19030533730984 25.76913594964537,-80.19030064344406 25.769163123697968,-80.19029662013054 25.769170370110963,-80.190289914608 25.76920901763935,-80.1902798563242 25.769238003277394,-80.19027784466743 25.76925249609375,-80.19027382135391 25.76925793089942,-80.19026979804039 25.76927121597893,-80.19026845693588 25.76928570879125,-80.1902637630701 25.76929537066513,-80.19026108086109 25.769310467341512,-80.19025169312954 25.76933401815282,-80.19025102257729 25.769348510957446,-80.19024699926376 25.76935817282624,-80.190244987607 25.769373269494615,-80.19024096429348 25.76937810042809,-80.19023559987545 25.769417955621794,-80.19024163484573 25.769424598152774,-80.19025504589081 25.769427617484915,-80.19027382135391 25.769437883213623,-80.19028455018997 25.769440298679076,-80.1902885735035 25.769445733476143,-80.1902885735035 25.769460830133404,-80.19028320908546 25.76948196545032,-80.19027851521969 25.769486192513256,-80.19027650356293 25.76949162730825,-80.19026510417461 25.769501893031403,-80.19026443362236 25.769508535557687,-80.19025772809982 25.769521216743208,-80.1902624219656 25.76953510565911,-80.19026175141335 25.769543559781024,-80.19025705754757 25.76955080617079,-80.19025705754757 25.769553221633938,-80.19026108086109 25.769556240962803,-80.19026108086109 25.769562279620317,-80.19026577472687 25.769567714411806,-80.1902711391449 25.769569526008922,-80.1902885735035 25.76957012987461,-80.190289914608 25.769572545337365,-80.19028723239899 25.769576772397073,-80.19031472504139 25.769579187859698,-80.19032076001167 25.769577376262735,-80.19032545387745 25.769571337606003,-80.19033081829548 25.7695707337403,-80.190334841609 25.769576772397073,-80.19035160541534 25.769576772397073,-80.1903522759676 25.769580999456632,-80.19035696983337 25.769584018784787,-80.19036769866943 25.769586434247262,-80.19037373363972 25.76958220718792,-80.19038446247578 25.769585226516046,-80.19039183855057 25.769594888365543,-80.19038580358028 25.769605757945303,-80.19037440419197 25.769610588869313,-80.19036501646042 25.769611796600298,-80.1903623342514 25.7696142120622,-80.19035696983337 25.769610588869313,-80.19034422934055 25.769607569541837,-80.19034020602703 25.769609985003832,-80.19034020602703 25.76961481592767,-80.19034557044506 25.769617231389518,-80.1903522759676 25.769616627524066,-80.19035562872887 25.76962025071675,-80.19036300480366 25.7696208545822,-80.19036702811718 25.76962508164017,-80.19037440419197 25.769627497101833,-80.19038647413254 25.769643197601265,-80.19039519131184 25.76964742465846,-80.19040390849113 25.769655878772394,-80.19041329622269 25.76967278699846,-80.19041329622269 25.76967761791974,-80.19041664898396 25.769678221784886,-80.19041866064072 25.76967399472879,-80.19042603671551 25.769668559942172,-80.19045419991016 25.769660709694364,-80.19045755267143 25.76965285944607,-80.19045352935791 25.769648028523775,-80.19044749438763 25.769645613062554,-80.19044481217861 25.76963897054394,-80.1904521882534 25.76962930869801,-80.19046224653721 25.769623873909346,-80.1904609054327 25.76960877727283,-80.19046559929848 25.76960394634874,-80.1904883980751 25.769599719289996,-80.19049040973186 25.769594888365543,-80.19050247967243 25.7695954922311,-80.19051253795624 25.769600927021095,-80.1905232667923 25.76960153088662,-80.19052930176258 25.76960394634874,-80.1905319839716 25.769605757945303,-80.19053131341934 25.769610588869313,-80.19055679440498 25.769623270043926,-80.19056752324104 25.76962508164017,-80.19057624042034 25.7696208545822,-80.19058227539062 25.7696208545822,-80.19058227539062 25.76962206231306,-80.19058294594288 25.769621458447634,-80.19058831036091 25.76962025071675,-80.19059501588345 25.769622666178485,-80.1906057447195 25.769623270043926,-80.1906144618988 25.76961964685131,-80.19062049686909 25.76961964685131,-80.19062519073486 25.769616023658585,-80.19062787294388 25.769604550214254,-80.19063457846642 25.769596096096684,-80.19063591957092 25.76959005744089,-80.1906406134367 25.769587038112892,-80.19064128398895 25.769566506680395,-80.19065134227276 25.769568922143225,-80.19065402448177 25.769564695083233,-80.1906593888998 25.76956348735177,-80.19067414104939 25.769571941471686,-80.19068822264671 25.769575564665757,-80.1906955987215 25.769571337606003,-80.19069761037827 25.769540540451843,-80.1907029747963 25.76953148246381,-80.19070900976658 25.76952785926842,-80.19070498645306 25.769520009011305,-80.19070498645306 25.76951397035164,-80.19070833921432 25.769509139423697,-80.19070968031883 25.769490419576044,-80.1907116919756 25.76948860797772,-80.19071571528912 25.769492231174333,-80.19071906805038 25.769492231174333,-80.1907217502594 25.769490419576044,-80.1907217502594 25.769494042772607,-80.19073247909546 25.769504912361572,-80.19072040915489 25.76951759354748,-80.19071772694588 25.769529670866113,-80.19071772694588 25.769539332720157,-80.1907230913639 25.76954235204937,-80.1907217502594 25.76956107188883,-80.19071571528912 25.769568318277503,-80.19071035087109 25.769568318277503,-80.19070833921432 25.7695707337403,-80.19070900976658 25.76957616853142,-80.19071370363235 25.769580395591,-80.19071370363235 25.769586434247262,-80.19071102142334 25.769590661306477,-80.19071102142334 25.769596096096684,-80.19070364534855 25.769599719289996,-80.19070096313953 25.76960938113834,-80.19070565700531 25.769616023658585,-80.19070699810982 25.76962508164017,-80.1907029747963 25.769635347351805,-80.19070096313953 25.769651651715503,-80.19069761037827 25.769657690368163,-80.19069626927376 25.769672183133267,-80.19069224596024 25.769677014054572,-80.19068621098995 25.769696337737866,-80.1906855404377 25.76971143436325,-80.19068822264671 25.769713849823127,-80.19071638584137 25.769722303932305,-80.19073449075222 25.76972411552706,-80.19074521958828 25.76972773871649,-80.19081093370914 25.76974041987856,-80.1908927410841 25.76976155514571,-80.19095242023468 25.769771820845516,-80.19096851348877 25.76977725562743,-80.19099600613117 25.769782086544467,-80.1910288631916 25.769791144513377,-80.19108720123768 25.769801410210622,-80.19109457731247 25.769804429533195,-80.19116967916489 25.769818922280383,-80.19119516015053 25.769825564788924,-80.19120253622532 25.769829187975244,-80.19121997058392 25.769831603432735,-80.19122667610645 25.769835830483217,-80.19123740494251 25.769836434347553,-80.19124813377857 25.769823753195716,-80.19125282764435 25.769822545466912,-80.19125819206238 25.76981771455153,-80.19126087427139 25.769798390888,-80.19126422703266 25.76979295610708,-80.19126623868942 25.769779067221325,-80.19127026200294 25.76977363243952,-80.19127160310745 25.7697591396868,-80.19127562642097 25.769752497174544,-80.191280990839 25.769718680742738,-80.19128568470478 25.769708415038313,-80.19128702580929 25.769698753198053,-80.19129104912281 25.769693318412568,-80.19129239022732 25.769671579268092,-80.19129574298859 25.769667956076958,-80.1912984251976 25.769651651715503,-80.19130311906338 25.769643801466607,-80.19130311906338 25.76963474348645,-80.1913071423769 25.76962508164017,-80.19131049513817 25.769594888365543,-80.19132323563099 25.769562883486046,-80.19132390618324 25.769551410036577,-80.19132860004902 25.769545371378506,-80.19132927060127 25.769536313390844,-80.19133999943733 25.769507327825657,-80.19134670495987 25.769470491993033,-80.19135072827339 25.769465661063315,-80.1913520693779 25.76945781080211,-80.19135676324368 25.769454791470725,-80.19134804606438 25.769448148941436,-80.19133128225803 25.76944633734249,-80.19132524728775 25.76944271414447,-80.19130580127239 25.76944150641177,-80.19129641354084 25.76943727934727,-80.19128367304802 25.769436675480875,-80.19127629697323 25.769433052282587,-80.19125953316689 25.769431844549782,-80.19125819206238 25.769425202019207,-80.19125483930111 25.76942701361849,-80.19123673439026 25.769426409752075,-80.19123271107674 25.769422786553452,-80.19119381904602 25.769413124689944,-80.19114822149277 25.769408293757895,-80.19113816320896 25.769404670558714,-80.19112341105938 25.769403462825643,-80.1911187171936 25.769400443492874,-80.191101282835 25.76939863189318,-80.19107043743134 25.76938957389431,-80.19105903804302 25.769388970027705,-80.19104965031147 25.769385346827956,-80.19101142883301 25.76937870429478,-80.19100606441498 25.769375684961375,-80.19099064171314 25.769373873361324,-80.19094035029411 25.76936058829331,-80.19092760980129 25.76935938055977,-80.19091621041298 25.769349114824287,-80.19091419875622 25.769349718691103,-80.19091352820396 25.769354549625533,-80.19090749323368 25.76935515349232,-80.1908940821886 25.769350926424714,-80.19088000059128 25.769349718691103,-80.19087195396423 25.769346095490132,-80.1908565312624 25.76934488775647,-80.19085183739662 25.769341868422206,-80.19083440303802 25.7693406606885,-80.19082836806774 25.769337037487258,-80.1908129453659 25.76933462201971,-80.19080959260464 25.769331602685202,-80.19079014658928 25.769330394951375,-80.19078746438026 25.769328583350628,-80.19078075885773 25.769330394951375,-80.19077003002167 25.769318921479428,-80.19075930118561 25.76931710987849,-80.19075326621532 25.76931710987849,-80.19074723124504 25.769320733080345,-80.19074723124504 25.769325564015958,-80.19075192511082 25.76932918721755,-80.19075393676758 25.769339452954757,-80.19075326621532 25.769356361225906,-80.19075594842434 25.76936723082751,-80.19075125455856 25.769373269494615,-80.19075058400631 25.76937810042809,-80.19074521958828 25.76938232749473,-80.19075125455856 25.769387158427833,-80.1907492429018 25.76939198936077,-80.19074991345406 25.769398028026618,-80.1907405257225 25.769402255092544,-80.1907405257225 25.769408293757895,-80.19073247909546 25.769407689891352,-80.19072510302067 25.76939198936077,-80.19071504473686 25.769384742961304,-80.19070498645306 25.769372665627916,-80.1907043159008 25.76936723082751,-80.19070632755756 25.769364211493812,-80.19070766866207 25.769351530291537,-80.19071236252785 25.76934488775647,-80.19071236252785 25.76934005682164,-80.19070498645306 25.7693358297535,-80.19070364534855 25.76932435628208,-80.19070833921432 25.769317713745487,-80.1907217502594 25.769315298277547,-80.19071035087109 25.769311071208524,-80.19070833921432 25.769306844139354,-80.1907130330801 25.769298993867622,-80.19071571528912 25.769288728126913,-80.1907217502594 25.769282689455483,-80.19072644412518 25.76928752039264,-80.1907318085432 25.76928752039264,-80.19073247909546 25.76928148172117,-80.19073784351349 25.76927725465094,-80.19072711467743 25.76926215797034,-80.19072644412518 25.76925189222644,-80.19073382019997 25.769248269022487,-80.19073650240898 25.7692066021692,-80.19074723124504 25.769175804920394,-80.19074723124504 25.769164331433505,-80.19075125455856 25.769151650209853,-80.19075192511082 25.769139572852755,-80.19075527787209 25.769132930305815,-80.1907566189766 25.76912024907883,-80.19076131284237 25.769112398794768,-80.19076265394688 25.76910152916982,-80.1907666772604 25.769093075016425,-80.1907680183649 25.76907918604877,-80.19077338278294 25.769069524157285,-80.19077271223068 25.76905865452842,-80.19077807664871 25.769045369425115,-80.19077874720097 25.769036915267716,-80.1907854527235 25.769022422425014,-80.1907928287983 25.769020610819574,-80.19079148769379 25.76900611797488,-80.19079752266407 25.769003098632023,-80.1908129453659 25.769000079289064,-80.19081830978394 25.768996456077446,-80.19080825150013 25.76898800191654,-80.19080825150013 25.768981359361124,-80.19081093370914 25.76897713228021,-80.19082300364971 25.76897713228021,-80.19082568585873 25.76897290519914,-80.19082367420197 25.768966866511658,-80.19082568585873 25.768962035561444,-80.19081898033619 25.768960223955055,-80.19081696867943 25.76895720461103,-80.19082501530647 25.768952977529256,-80.1908303797245 25.76894391949638,-80.1908303797245 25.7689378808074,-80.19083574414253 25.768934257593862,-80.1908377557993 25.76892942664232,-80.19083708524704 25.768923991821605,-80.1908303797245 25.768919160869647,-80.19083373248577 25.768915537655538,-80.19083239138126 25.768908895096065,-80.19083641469479 25.76890406414349,-80.1908303797245 25.76890044092892,-80.19082903862 25.76889319449947,-80.19082501530647 25.768888363546253,-80.19082836806774 25.76885515073756,-80.19083641469479 25.768840054003284,-80.19084110856056 25.768837034656183,-80.1908390969038 25.768831599831234,-80.19084580242634 25.768833411439587,-80.19084647297859 25.768822541789092,-80.19085049629211 25.768820126311077,-80.19085519015789 25.768820126311077,-80.19085586071014 25.76882858048394,-80.19085854291916 25.768830995961792,-80.19087061285973 25.76882858048394,-80.19088335335255 25.768830392092337,-80.19088804721832 25.76884126174209,-80.19088804721832 25.76884730043598,-80.19088938832283 25.76884307335029,-80.1908927410841 25.76884126174209,-80.19089207053185 25.768836430786763,-80.1908940821886 25.76883522304791,-80.1908940821886 25.76884126174209,-80.19089810550213 25.76884367721967,-80.19089810550213 25.768846092697213,-80.19090548157692 25.768847904305343,-80.190918892622 25.768861189430748,-80.19094236195087 25.768858773953518,-80.19094571471214 25.768861189430748,-80.19094705581665 25.768868435862156,-80.19095979630947 25.768868435862156,-80.19095912575722 25.768874474554664,-80.19095242023468 25.768875682293142,-80.1909463852644 25.768872662946944,-80.19093833863735 25.76889077902289,-80.1909289509058 25.768895006106877,-80.19093431532383 25.76890406414349,-80.19093431532383 25.768912518310373,-80.19093632698059 25.76891493378651,-80.19093230366707 25.76891493378651,-80.19092828035355 25.76891191444134,-80.19092626869678 25.76889923319072,-80.19091822206974 25.768901044798035,-80.19090816378593 25.768898025452483,-80.1908940821886 25.76891191444134,-80.19088335335255 25.76891734926261,-80.190873965621 25.76891855700062,-80.19086323678493 25.76892338795263,-80.19086390733719 25.768928822773365,-80.19087195396423 25.7689378808074,-80.19087128341198 25.76894391949638,-80.190873965621 25.7689475427096,-80.190873965621 25.768952977529256,-80.19086457788944 25.768962639430224,-80.19087262451649 25.76896324329901,-80.19088067114353 25.76897713228021,-80.19088335335255 25.76897713228021,-80.19088536500931 25.768995248340204,-80.1908840239048 25.769011552791856,-80.1908940821886 25.769019403082574,-80.19089877605438 25.769015176003013,-80.19090816378593 25.769013968265977,-80.19091285765171 25.7690109489233,-80.1909101754427 25.76900189089485,-80.19091084599495 25.76899585220884,-80.19091822206974 25.768991625128436,-80.19091956317425 25.76898800191654,-80.19091621041298 25.768981359361124,-80.19091956317425 25.768978943886335,-80.19093096256256 25.76897713228021,-80.19092626869678 25.76897713228021,-80.19092559814453 25.76897290519914,-80.19093163311481 25.76897290519914,-80.19093163311481 25.768966866511658,-80.19093699753284 25.768963847167775,-80.19094839692116 25.768967470380417,-80.19094437360764 25.768971093592942,-80.19094571471214 25.76897713228021,-80.19093632698059 25.76897713228021,-80.19093297421932 25.768989209653856,-80.19093297421932 25.769001287026263,-80.19094102084637 25.769006721843436,-80.1909477263689 25.769008533449114,-80.19095778465271 25.769004910237754,-80.19096449017525 25.768996456077446,-80.19096650183201 25.768989813522506,-80.1909738779068 25.76898619031055,-80.19097924232483 25.768991625128436,-80.19098058342934 25.769003098632023,-80.19098594784737 25.769017591477063,-80.19099399447441 25.769020610819574,-80.19099198281765 25.769022422425014,-80.19098460674286 25.769022422425014,-80.19098192453384 25.7690254417674,-80.19098460674286 25.769030272715035,-80.19098058342934 25.769035103662475,-80.19098326563835 25.769045369425115,-80.19097991287708 25.769047181030206,-80.19096583127975 25.769046577161856,-80.190963819623 25.76904838876692,-80.19096650183201 25.769054427450257,-80.1909651607275 25.769059258396734,-80.19095979630947 25.769064693211252,-80.19095107913017 25.769068920289044,-80.19094973802567 25.769074958971355,-80.19095242023468 25.769081601521517,-80.19095845520496 25.769087036335034,-80.19096449017525 25.76908764020318,-80.1909738779068 25.769083413126044,-80.19097253680229 25.76907797831239,-80.19097588956356 25.769072543498453,-80.19097320735455 25.76907012802552,-80.1909738779068 25.769067108684293,-80.19099064171314 25.769065297079507,-80.19099466502666 25.769059258396734,-80.1910013705492 25.76905865452842,-80.19101478159428 25.769068316420807,-80.19101075828075 25.769074958971355,-80.19100271165371 25.769078582180583,-80.19100069999695 25.76908280925788,-80.19100271165371 25.769087036335034,-80.19100204110146 25.769093678884538,-80.19099466502666 25.769096094356996,-80.19099466502666 25.769102736905985,-80.1909913122654 25.769106963982427,-80.19099935889244 25.769116625870865,-80.19099600613117 25.769121456814783,-80.19099868834019 25.76912508002261,-80.19100673496723 25.769123268418713,-80.19101209938526 25.76912508002261,-80.19101545214653 25.76912206068275,-80.19101612269878 25.769116625870865,-80.19102483987808 25.76910817171852,-80.19102685153484 25.769102736905985,-80.19103355705738 25.76910575624631,-80.19104026257992 25.769100321433672,-80.19106239080429 25.769096698225088,-80.19107043743134 25.769091867280157,-80.19107580184937 25.769091263412033,-80.19107982516289 25.769086432466885,-80.19108720123768 25.76908582859871,-80.19108921289444 25.769084016994235,-80.19108653068542 25.76907918604877,-80.19108720123768 25.76907435510313,-80.19109725952148 25.769068920289044,-80.19109927117825 25.769059258396734,-80.19109390676022 25.76905503131859,-80.19109390676022 25.76904416168839,-80.19109658896923 25.769039934609715,-80.19108854234219 25.769030272715035,-80.19108653068542 25.76901879921408,-80.19109658896923 25.769010345054763,-80.19109189510345 25.769007929580567,-80.19108586013317 25.76900974118622,-80.19108049571514 25.769007325712018,-80.19107915461063 25.769000683157675,-80.19107580184937 25.768996456077446,-80.1910650730133 25.76899585220884,-80.19106842577457 25.768981963229805,-80.19108049571514 25.76897713228021,-80.19107848405838 25.76897713228021,-80.19108720123768 25.768967470380417,-80.19109591841698 25.768962639430224,-80.19110396504402 25.76896324329901,-80.19111067056656 25.768957808479843,-80.19111067056656 25.768952977529256,-80.19111402332783 25.768951165922733,-80.19114352762699 25.76894573110299,-80.19115224480629 25.76894573110299,-80.19115559756756 25.768948146578477,-80.19116565585136 25.768949958185026,-80.19117973744869 25.768948750447326,-80.19117973744869 25.76894693884074,-80.19117504358292 25.768944523365253,-80.19117437303066 25.768932445987073,-80.19117638468742 25.768928822773365,-80.19116029143333 25.768910102834184,-80.19116632640362 25.768899837059823,-80.19116565585136 25.768894402237752,-80.19117772579193 25.768881720985263,-80.19118443131447 25.768878701639235,-80.19118644297123 25.768872059077705,-80.19118510186672 25.768860585561455,-80.1911897957325 25.768855754606907,-80.1911998540163 25.76885756621489,-80.19120387732983 25.768855754606907,-80.19121527671814 25.768855754606907,-80.1912159472704 25.768861793300047,-80.19121930003166 25.768865416515794,-80.19121795892715 25.768873870685436,-80.19120454788208 25.768885344200402,-80.19119918346405 25.76887990937765,-80.1911985129118 25.768885948069574,-80.1911998540163 25.768889571284586,-80.19120454788208 25.768892590630333,-80.19121930003166 25.76889138289203,-80.1912260055542 25.768894402237752,-80.19123405218124 25.768903460274412,-80.19123405218124 25.76891070670323,-80.19123069941998 25.768919160869647,-80.1912260055542 25.76891795313161,-80.19122064113617 25.768923991821605,-80.19121527671814 25.768923991821605,-80.19120924174786 25.768933653724932,-80.19120320677757 25.7689378808074,-80.19120320677757 25.76894391949638,-80.19119784235954 25.768952373660404,-80.19120320677757 25.768962035561444,-80.1911985129118 25.768970489724175,-80.19120119512081 25.76897713228021,-80.19120052456856 25.76897713228021,-80.19120924174786 25.768981359361124,-80.19120253622532 25.76899404060297,-80.19120186567307 25.76900189089485,-80.19122332334518 25.769006721843436,-80.19123874604702 25.76900551410632,-80.19124880433083 25.769011552791856,-80.19124880433083 25.769015176003013,-80.1912534981966 25.76902000695106,-80.19124612212181 25.769029064978135,-80.19124813377857 25.769040538478098,-80.19124142825603 25.769044765556757,-80.1912447810173 25.769057446791827,-80.1912434399128 25.769065297079507,-80.19124679267406 25.769071335762003,-80.19125953316689 25.769080997653333,-80.19126355648041 25.769093678884538,-80.19125886261463 25.769097905961306,-80.19125953316689 25.769107567850476,-80.19125483930111 25.769112398794768,-80.19125148653984 25.769121456814783,-80.19125081598759 25.76913655351329,-80.1912534981966 25.76914078058853,-80.1912521570921 25.769146215399317,-80.19126757979393 25.76915406568113,-80.19127428531647 25.76915346181332,-80.1912796497345 25.76915044247421,-80.19128769636154 25.76914319606002,-80.19129104912281 25.76913655351329,-80.19129574298859 25.76913655351329,-80.1912984251976 25.769132326437898,-80.19131049513817 25.769127495494423,-80.19131384789944 25.769123268418713,-80.19130915403366 25.769116625870865,-80.19129976630211 25.769112398794768,-80.19129574298859 25.76909247114828,-80.19129171967506 25.769087036335034,-80.19129574298859 25.769083413126044,-80.1912970840931 25.769074958971355,-80.19130177795887 25.769069524157285,-80.19129976630211 25.769064089342987,-80.19130513072014 25.769059258396734,-80.19130446016788 25.76904416168839,-80.19130177795887 25.769039330741325,-80.19130580127239 25.769030272715035,-80.19131451845169 25.7690254417674,-80.1913158595562 25.76902000695106,-80.19133396446705 25.76901879921408,-80.19133999943733 25.769020610819574,-80.19134670495987 25.769016383740052,-80.19134603440762 25.76900913731766,-80.19134871661663 25.76900551410632,-80.1913607865572 25.769006721843436,-80.19136480987072 25.7690109489233,-80.19136279821396 25.769015779871523,-80.19136480987072 25.769016987608563,-80.19137620925903 25.769016383740052,-80.19137889146805 25.76901276052891,-80.19138224422932 25.769012156660384,-80.19138693809509 25.769019403082574,-80.19138559699059 25.769028461109688,-80.19138358533382 25.769030272715035,-80.19138358533382 25.769046577161856,-80.19139498472214 25.76906771255254,-80.19139230251312 25.769073751234913,-80.19139632582664 25.76907435510313,-80.1913983374834 25.769073147366694,-80.19139900803566 25.769064089342987,-80.19140169024467 25.769059258396734,-80.19141107797623 25.76905503131859,-80.19141308963299 25.769041746214885,-80.191415771842 25.769039934609715,-80.19141308963299 25.769035103662475,-80.19141845405102 25.76903389592562,-80.19142046570778 25.769029064978135,-80.19143119454384 25.76902483789892,-80.19143521785736 25.769014572134495,-80.1914419233799 25.769007929580567,-80.19144661724567 25.769007325712018,-80.19145734608173 25.769015176003013,-80.19145734608173 25.769020610819574,-80.19145131111145 25.76902483789892,-80.19145265221596 25.769030876583475,-80.19144997000694 25.76903389592562,-80.19144997000694 25.769035707530882,-80.19145667552948 25.769040538478098,-80.19145399332047 25.769041746214885,-80.1914519816637 25.769048992635263,-80.19144862890244 25.769051408108627,-80.19144125282764 25.769066504816042,-80.19143722951412 25.769069524157285,-80.19143588840961 25.76907918604877,-80.19143052399158 25.76908764020318,-80.19142985343933 25.769093075016425,-80.19142650067806 25.769094886620756,-80.19142247736454 25.76910152916982,-80.19142650067806 25.76910636011436,-80.19142046570778 25.769120852946813,-80.19142113626003 25.76913474190957,-80.1914244890213 25.769142592192143,-80.19141845405102 25.769149234738535,-80.19141308963299 25.769138365116987,-80.1914070546627 25.769138365116987,-80.19139632582664 25.769131722569984,-80.19138626754284 25.7691287032303,-80.19138358533382 25.769132930305815,-80.19137419760227 25.76913594964537,-80.191370844841 25.769143799927885,-80.19136548042297 25.769143799927885,-80.19136145710945 25.769147423135028,-80.19135475158691 25.76914681926716,-80.19135005772114 25.76915285794551,-80.19134268164635 25.769151046342035,-80.19133932888508 25.76915285794551,-80.19133664667606 25.76915708502017,-80.19133999943733 25.769160104359106,-80.19133932888508 25.76916493530126,-80.19133396446705 25.76916191596244,-80.19132658839226 25.769164331433505,-80.19132658839226 25.76916674690452,-80.19133128225803 25.769169162375487,-80.19133128225803 25.7691715778464,-80.19132390618324 25.769178824258855,-80.1913245767355 25.769181239729576,-80.19133061170578 25.769184862935575,-80.1913245767355 25.76919029774435,-80.1913332939148 25.769198751890837,-80.19133061170578 25.769212640844483,-80.19132792949677 25.769215660182034,-80.19132524728775 25.769213848579515,-80.1913245767355 25.769203582831416,-80.19132122397423 25.769202978963865,-80.1913172006607 25.769208413771807,-80.1913172006607 25.769218075652,-80.19131250679493 25.76922290659179,-80.19131518900394 25.769229549133694,-80.19131116569042 25.769232568470795,-80.19132122397423 25.76923498394041,-80.19132927060127 25.769234380073016,-80.19133262336254 25.76923619167522,-80.19135542213917 25.769235587807806,-80.19136413931847 25.76921988725445,-80.19136413931847 25.769212640844483,-80.19136011600494 25.76920901763935,-80.19135944545269 25.76920479056654,-80.1913607865572 25.769191505479604,-80.19137352705002 25.769181843597266,-80.19137620925903 25.769181843597266,-80.1913795620203 25.769175804920394,-80.19138626754284 25.76917097397869,-80.19140973687172 25.76917097397869,-80.19142046570778 25.769175804920394,-80.19141979515553 25.76919090161199,-80.19141644239426 25.769202375096306,-80.19141912460327 25.76921928338696,-80.19142851233482 25.76922109498939,-80.19144594669342 25.76922049112192,-80.1914506405592 25.7692241143267,-80.19146136939526 25.769225925929057,-80.19146673381329 25.76923075686853,-80.19148081541061 25.769232568470795,-80.19148752093315 25.769236795542607,-80.19149623811245 25.769237399409988,-80.19150696694851 25.769256723164858,-80.19150696694851 25.769267592775584,-80.19150093197823 25.769289935861156,-80.19150227308273 25.7692965783993,-80.1914955675602 25.769300805468838,-80.19149892032146 25.76930563640526,-80.19149422645569 25.769310467341512,-80.19149154424667 25.769309259607464,-80.19148752093315 25.769305032538227,-80.19149355590343 25.769300805468838,-80.1914881914854 25.76929778613348,-80.19148550927639 25.769298993867622,-80.19148416817188 25.769303220937076,-80.19147746264935 25.769302617070025,-80.19147410988808 25.76930563640526,-80.19147276878357 25.769310467341512,-80.19146740436554 25.769314694410536,-80.19146203994751 25.769341868422206,-80.19145667552948 25.769347907090626,-80.19146271049976 25.76935273802514,-80.19146673381329 25.769359984426533,-80.19147478044033 25.76935817282624,-80.19147880375385 25.76936058829331,-80.19147612154484 25.769384742961304,-80.19147545099258 25.769388970027705,-80.19147209823132 25.76939259322736,-80.19147075712681 25.769407086024838,-80.19146539270878 25.76941554015589,-80.19146472215652 25.769428221351344,-80.191460698843 25.769436071614507,-80.19145868718624 25.76944935667407,-80.19145533442497 25.769454791470725,-80.19145332276821 25.769470491993033,-80.19144997000694 25.7694741151902,-80.19144460558891 25.76948860797772,-80.1914432644844 25.769507931691688,-80.19143924117088 25.76951336648566,-80.19143722951412 25.769528463134314,-80.19143387675285 25.76953208632969,-80.19143253564835 25.769549598439188,-80.19142650067806 25.769560468023094,-80.19142784178257 25.769571337606003,-80.19142381846905 25.769575564665757,-80.19142381846905 25.769588849709685,-80.19142918288708 25.7695954922311,-80.19142784178257 25.76959851155891,-80.19142381846905 25.76960032315556,-80.1914244890213 25.769602134752155,-80.19142985343933 25.769600927021095,-80.19143924117088 25.769604550214254,-80.1914506405592 25.76961481592767,-80.19145600497723 25.76961481592767,-80.19145667552948 25.769622666178485,-80.19145131111145 25.769626289371008,-80.19143521785736 25.769629912563417,-80.19142985343933 25.769634139621065,-80.19142583012581 25.76962930869801,-80.19141778349876 25.7696305164288,-80.19141510128975 25.769622666178485,-80.19141308963299 25.769622666178485,-80.19141241908073 25.769626289371008,-80.19140303134918 25.769633535755695,-80.19140504300594 25.769643197601265,-80.19140169024467 25.76965104785023,-80.19140034914017 25.769663125155276,-80.1913969963789 25.769666748346562,-80.19138090312481 25.769667956076958,-80.1913795620203 25.769675806324276,-80.19137553870678 25.769681241110565,-80.19137419760227 25.76969392227763,-80.19137017428875 25.76969935706309,-80.19137017428875 25.769710226633276,-80.19136749207973 25.769713245958144,-80.19136145710945 25.769713849823127,-80.19135810434818 25.769716265282955,-80.19135676324368 25.76972653098668,-80.19135273993015 25.76973377736525,-80.19135273993015 25.76974041987856,-80.19135475158691 25.76974223147304,-80.19137687981129 25.76974947785065,-80.19137419760227 25.769765178333973,-80.19139297306538 25.769773028574864,-80.19138962030411 25.769781482679853,-80.19138894975185 25.769794163836185,-80.19138425588608 25.769802617939664,-80.19138291478157 25.769822545466912,-80.19138760864735 25.769824960924527,-80.19139900803566 25.769825564788924,-80.19140839576721 25.769829791839626,-80.19142583012581 25.769830999568352,-80.19143186509609 25.769835226618856,-80.19144125282764 25.769835226618856,-80.19144594669342 25.769829791839626,-80.19146136939526 25.769827376382104,-80.19145600497723 25.769818922280383,-80.19145600497723 25.769811675907007,-80.19146002829075 25.76980503339769,-80.19146136939526 25.769791748377934,-80.19146606326103 25.76978631359671,-80.19147075712681 25.769774840168843,-80.19147008657455 25.7697591396868,-80.19147478044033 25.76974947785065,-80.1914694160223 25.769739816013733,-80.19147410988808 25.769736192824702,-80.19148215651512 25.769736192824702,-80.1914881914854 25.769739816013733,-80.19148215651512 25.76974525079711,-80.19148416817188 25.769754912633587,-80.19149422645569 25.76976095128097,-80.19149892032146 25.769767593792754,-80.191505625844 25.76976940538682,-80.19149892032146 25.769770613116187,-80.19149288535118 25.769774236304166,-80.19149355590343 25.769783898138304,-80.19148886203766 25.769787521325885,-80.1914881914854 25.769795371565298,-80.19149489700794 25.769792352242497,-80.19149757921696 25.769798390888,-80.19149288535118 25.76980020248159,-80.19148014485836 25.769812279771468,-80.19148148596287 25.76981771455153,-80.19148550927639 25.769820733873658,-80.19148886203766 25.76982798024648,-80.19149959087372 25.76982798024648,-80.19150026142597 25.769824960924527,-80.19151233136654 25.769818318415957,-80.19151568412781 25.76981288363592,-80.19152037799358 25.76981529909374,-80.19153982400894 25.769814695229297,-80.19154317677021 25.76981288363592,-80.19154250621796 25.769808052720148,-80.19154720008373 25.76980322180417,-80.19155390560627 25.76980322180417,-80.19155859947205 25.769807448855666,-80.19156999886036 25.769807448855666,-80.19157268106937 25.769804429533195,-80.19157268106937 25.769800806346108,-80.19156865775585 25.76979295610708,-80.19156195223331 25.769784502002903,-80.19153647124767 25.769764574469274,-80.19153580069542 25.769754308768835,-80.19154250621796 25.769753101039303,-80.19155256450176 25.769756724227843,-80.1915592700243 25.769762762875132,-80.19159078598022 25.76975974355153,-80.19160754978657 25.769750081715447,-80.19160822033882 25.769736192824702,-80.19161090254784 25.769730154176038,-80.19162364304066 25.769722303932305,-80.19162431359291 25.769739212148913,-80.19162029027939 25.769747062391495,-80.1916216313839 25.76975974355153,-80.19161693751812 25.76976397060457,-80.1916142553091 25.769776651762772,-80.19160889089108 25.769783294273683,-80.19160956144333 25.76979355997163,-80.19162900745869 25.76980865658464,-80.19163101911545 25.769827376382104,-80.19163772463799 25.769835226618856,-80.1916491240263 25.769836434347553,-80.19165381789207 25.769841869126484,-80.19166052341461 25.76984428458371,-80.1916665583849 25.76984368071941,-80.19167594611645 25.76983824594057,-80.19168063998222 25.769841265262176,-80.19168265163898 25.769857569597576,-80.19167728722095 25.769871458474125,-80.19168198108673 25.769875685523186,-80.19168466329575 25.769884139620864,-80.19168399274349 25.769895009174128,-80.19168734550476 25.769896820766224,-80.19169203937054 25.769904067134426,-80.19169270992279 25.769915540549796,-80.19169874489307 25.769917352141597,-80.19171953201294 25.769917956005525,-80.19172959029675 25.769926410100187,-80.19174300134182 25.76992459850852,-80.19175171852112 25.769928825555695,-80.19179530441761 25.769928825555695,-80.19180536270142 25.76993365646655,-80.19180871546268 25.769939091241035,-80.19181407988071 25.769939091241035,-80.19182614982128 25.76992701396406,-80.19184961915016 25.769926410100187,-80.19186437129974 25.769923390780722,-80.19186839461327 25.76992701396406,-80.19187174737453 25.76992701396406,-80.19188046455383 25.76992459850852,-80.19188649952412 25.769917352141597,-80.19189186394215 25.769917352141597,-80.19189588725567 25.769921579189006,-80.1919025927782 25.769920975325107,-80.19190795719624 25.76991855986944,-80.19191198050976 25.76991252123005,-80.19193477928638 25.7699107096382,-80.1919388025999 25.76990708645436,-80.191954895854 25.769902859406415,-80.19195556640625 25.769902255542416,-80.19195556640625 25.76989984008637,-80.19196964800358 25.76989621690219,-80.19198440015316 25.769895613038162,-80.19201390445232 25.76988776280538),(-80.1919736713171 25.76986602369655,-80.19196294248104 25.769881120300347,-80.19195556640625 25.769882931892653,-80.19195556640625 25.76986723142491,-80.19196763634682 25.76986662756073,-80.19197098910809 25.769864212103958,-80.1919736713171 25.76986602369655),(-80.1909738779068 25.76897713228021,-80.19097186625004 25.76897713228021,-80.19096985459328 25.768971697461673,-80.1909738779068 25.768966866511658,-80.19097991287708 25.76896565877411,-80.19097991287708 25.76897713228021,-80.19097454845905 25.76897713228021,-80.1909738779068 25.76897713228021),(-80.19102953374386 25.76909247114828,-80.1910275220871 25.76908824407132,-80.1910188049078 25.76908824407132,-80.19101344048977 25.76909247114828,-80.1910188049078 25.769099113697482,-80.19102819263935 25.76910152916982,-80.19102953374386 25.76909247114828),(-80.19074186682701 25.76946445333087,-80.19074186682701 25.769460226267157,-80.19073717296124 25.76945720693583,-80.19072040915489 25.769460226267157,-80.19072040915489 25.769470491993033,-80.19072711467743 25.76947653065491,-80.19073247909546 25.7694741151902,-80.19073583185673 25.769468680394404,-80.19074186682701 25.76946445333087),(-80.19068822264671 25.769399235759764,-80.19069023430347 25.76939983962632,-80.19069157540798 25.769391385494142,-80.19069023430347 25.769388970027705,-80.19068822264671 25.769399235759764),(-80.19169136881828 25.76981288363592,-80.19169002771378 25.769802617939664,-80.19168734550476 25.769798994752524,-80.19168265163898 25.769802014075157,-80.1916766166687 25.769821941602483,-80.19167862832546 25.769824357060113,-80.19168935716152 25.769822545466912,-80.19169136881828 25.76981288363592),(-80.19167058169842 25.76977967108597,-80.19167594611645 25.769783294273683,-80.19167996942997 25.76978087881521,-80.19167728722095 25.769775444033492,-80.19167058169842 25.76977967108597),(-80.19153580069542 25.769613608196718,-80.19153110682964 25.76962930869801,-80.19152842462063 25.769623873909346,-80.19152507185936 25.76962447777476,-80.19152574241161 25.769640178274614,-80.1915331184864 25.76964621692786,-80.19153714179993 25.76965889809864,-80.19154787063599 25.769661917424834,-80.1915492117405 25.769657690368163,-80.19155390560627 25.769654067176596,-80.19155658781528 25.769661917424834,-80.1915592700243 25.769663729020515,-80.19157133996487 25.76965346331133,-80.19156999886036 25.769645009197223,-80.19155994057655 25.76963957440928,-80.19156329333782 25.769626289371008,-80.19156195223331 25.7696142120622,-80.19156396389008 25.76960877727283,-80.1915592700243 25.769600927021095,-80.1915492117405 25.7696027386177,-80.1915418356657 25.769611796600298,-80.19153580069542 25.769613608196718),(-80.19154720008373 25.769526651536573,-80.19154317677021 25.76953027473203,-80.19153714179993 25.769529670866113,-80.19153378903866 25.76953510565911,-80.1915331184864 25.7695429559152,-80.19153580069542 25.769545371378506,-80.1915418356657 25.7695429559152,-80.19154720008373 25.769545371378506,-80.19155323505402 25.76954235204937,-80.19155792891979 25.769544767512695,-80.19156329333782 25.76954235204937,-80.1915693283081 25.769543559781024,-80.19157133996487 25.769539332720157,-80.19156999886036 25.76953087859792,-80.1915606111288 25.76952785926842,-80.19155256450176 25.76953027473203,-80.19154720008373 25.769526651536573),(-80.19156195223331 25.769569526008922,-80.1915693283081 25.7695671105461,-80.1915592700243 25.76956348735177,-80.19155658781528 25.769565902814666,-80.19155859947205 25.76957496080007,-80.19156195223331 25.769569526008922),(-80.19155859947205 25.769576772397073,-80.19155457615852 25.76958341491917,-80.19154787063599 25.769581603322273,-80.19154116511345 25.769585226516046,-80.19154720008373 25.7695954922311,-80.19155390560627 25.769597303827794,-80.19155457615852 25.769590661306477,-80.19155725836754 25.7695894535753,-80.1915606111288 25.769582811053542,-80.1915606111288 25.769578583994054,-80.19155859947205 25.769576772397073),(-80.19151970744133 25.76971083049825,-80.19151635468006 25.769707207308333,-80.19150294363499 25.769712038228214,-80.19150026142597 25.769718680742738,-80.19149959087372 25.76973136190578,-80.1915042847395 25.76973498509498,-80.19151300191879 25.7697331735004,-80.19151769578457 25.769741023743403,-80.19152641296387 25.769744646932292,-80.19153244793415 25.769741627608223,-80.1915404945612 25.769747666256308,-80.19154720008373 25.769747062391495,-80.19155323505402 25.769741023743403,-80.19155994057655 25.76974041987856,-80.19154854118824 25.76972653098668,-80.19153714179993 25.769728342581374,-80.19153378903866 25.76973136190578,-80.1915317773819 25.76973800441922,-80.19151769578457 25.769734381230137,-80.19151568412781 25.76972955031116,-80.19151836633682 25.76972411552706,-80.19151970744133 25.76971083049825),(-80.19147410988808 25.769540540451843,-80.19146539270878 25.769540540451843,-80.19146271049976 25.769538124988436,-80.19146002829075 25.769541748183542,-80.19146203994751 25.76954899457339,-80.19147276878357 25.76954899457339,-80.1914781332016 25.7695429559152,-80.19147410988808 25.769540540451843),(-80.1914244890213 25.76915708502017,-80.1914244890213 25.769154669548957,-80.19141979515553 25.769151650209853,-80.19140772521496 25.769151046342035,-80.1914057135582 25.76916674690452,-80.19140772521496 25.769169162375487,-80.19141979515553 25.769166143036774,-80.19142046570778 25.769160104359106,-80.1914244890213 25.76915708502017),(-80.19144125282764 25.768866624254365,-80.19144058227539 25.768864812646513,-80.19143588840961 25.76886360490793,-80.19143052399158 25.768864812646513,-80.19143052399158 25.768867228123625,-80.19144125282764 25.768866624254365),(-80.19141845405102 25.768898629321612,-80.1914231479168 25.768895609976003,-80.19141979515553 25.768892590630333,-80.19141376018524 25.76889319449947,-80.19141308963299 25.768895006106877,-80.19141845405102 25.768898629321612),(-80.19130982458591 25.76884730043598,-80.19130311906338 25.76884307335029,-80.19129775464535 25.76884307335029,-80.19129574298859 25.768846696566612,-80.1912970840931 25.768855754606907,-80.19130446016788 25.768861189430748,-80.19131183624268 25.768863001038643,-80.19131183624268 25.76887628616236,-80.19131653010845 25.768880513246867,-80.1913172006607 25.76888474033123,-80.19131988286972 25.76887990937765,-80.19133262336254 25.768875682293142,-80.1913332939148 25.768866020385065,-80.19132323563099 25.768861189430748,-80.19132256507874 25.768855754606907,-80.19131183624268 25.768850923652174,-80.19130982458591 25.76884730043598),(-80.1912085711956 25.768962639430224,-80.19121661782265 25.76896324329901,-80.19121259450912 25.768957808479843,-80.1912085711956 25.768957808479843,-80.1912085711956 25.768962639430224),(-80.1909738779068 25.768820126311077,-80.19096046686172 25.76881952244157,-80.1909463852644 25.768820126311077,-80.19094303250313 25.768822541789092,-80.19094303250313 25.768824957267064,-80.19094705581665 25.76882858048394,-80.19095309078693 25.768827372744987,-80.19095309078693 25.768844281089084,-80.19095778465271 25.768853339129567,-80.19096180796623 25.768856962345552,-80.19097454845905 25.768860585561455,-80.1909752190113 25.768866020385065,-80.19097790122032 25.768868435862156,-80.19098460674286 25.768867831992893,-80.19098728895187 25.768862397169357,-80.19099399447441 25.768858773953518,-80.19099533557892 25.768855754606907,-80.19097991287708 25.76884126174209,-80.19097723066807 25.768833411439587,-80.19097991287708 25.768827976614457,-80.19097857177258 25.768822541789092,-80.1909738779068 25.768820126311077),(-80.19095107913017 25.76888353259284,-80.19095376133919 25.768885344200402,-80.19095711410046 25.768879305508438,-80.19095309078693 25.76887628616236,-80.19095107913017 25.76888353259284),(-80.190604403615 25.768684255593215,-80.19060708582401 25.768680632371932,-80.1906144618988 25.768677009150565,-80.19060306251049 25.768676405280328,-80.19060172140598 25.768682443982588,-80.190604403615 25.768684255593215))",
			input2:  "MULTIPOLYGON(((-80.191965 25.769143,-80.191865 25.769121,-80.191876 25.769082,-80.19185 25.769076,-80.191844 25.769101,-80.191834 25.769098,-80.191816 25.769165,-80.191644 25.769127,-80.191643 25.769132,-80.191622 25.769128,-80.191641 25.769056,-80.191616 25.76905,-80.19168 25.768812,-80.191886 25.768858,-80.191879 25.768882,-80.192082 25.768926,-80.192027 25.76913,-80.191975 25.769118,-80.191953 25.769201,-80.191909 25.769191,-80.191917 25.769161,-80.191958 25.76917,-80.191965 25.769143)))",
			fwdDiff: "POLYGON((-80.19202262163162 25.76913051483411,-80.19201658666134 25.76914198832427,-80.19200317561626 25.769155877284554,-80.19200585782528 25.76917338944956,-80.19200049340725 25.769184259067927,-80.19199848175049 25.76920116736116,-80.19198641180992 25.76921324471201,-80.19198305904865 25.769225925929057,-80.19197568297386 25.769232568470795,-80.1919562369585 25.769242834216378,-80.19195556640625 25.769243438083727,-80.19195556640625 25.769244645818425,-80.19193410873413 25.769258534766717,-80.1919274032116 25.76927242371336,-80.19192606210709 25.76928148172117,-80.19192136824131 25.769286916525516,-80.1919186860323 25.76929537066513,-80.19192203879356 25.769315298277547,-80.19192673265934 25.76931952534642,-80.19192203879356 25.76932496014902,-80.19192136824131 25.769335225886593,-80.1919274032116 25.769339452954757,-80.19192807376385 25.769350926424714,-80.19193209707737 25.769354549625533,-80.19194148480892 25.76935756895945,-80.19194349646568 25.769361192160066,-80.19194349646568 25.769366023094033,-80.19193813204765 25.769373873361324,-80.19193679094315 25.769384742961304,-80.19194282591343 25.76939198936077,-80.19195556640625 25.76939561256031,-80.19195556640625 25.76939742416005,-80.19197769463062 25.769398028026618,-80.19199848175049 25.769401651225998,-80.19200183451176 25.769399235759764,-80.19201926887035 25.769398028026618,-80.1920286566019 25.769395008693728,-80.19205279648304 25.769396216426884,-80.19205816090107 25.76939983962632,-80.19206084311008 25.769408293757895,-80.19206888973713 25.76941252082345,-80.19206821918488 25.76942580588563,-80.1920635253191 25.76943727934727,-80.19206821918488 25.76944150641177,-80.19207090139389 25.76945720693583,-80.19207894802094 25.76946264173215,-80.19207693636417 25.76947773838723,-80.19207827746868 25.76949102344214,-80.19207626581192 25.769494646638677,-80.19206888973713 25.769496458236908,-80.19206553697586 25.76949947756721,-80.19206553697586 25.769528463134314,-80.19206017255783 25.769536917256723,-80.19205883145332 25.76954778684177,-80.19206084311008 25.769562883486046,-80.19206821918488 25.769577376262735,-80.19207492470741 25.769599115424455,-80.19206754863262 25.769616023658585,-80.1920635253191 25.769651651715503,-80.19206620752811 25.769676410189426,-80.19206821918488 25.769679429515154,-80.19207492470741 25.769682448840825,-80.19206888973713 25.76972109620247,-80.19206956028938 25.769744646932292,-80.19207425415516 25.769754308768835,-80.19207626581192 25.76978087881521,-80.19207894802094 25.769785105867502,-80.19209168851376 25.76979295610708,-80.1920910179615 25.769804429533195,-80.19208498299122 25.76981288363592,-80.192089676857 25.76981771455153,-80.19208297133446 25.7698195261448,-80.19207626581192 25.769830999568352,-80.19207961857319 25.7698460961766,-80.19207760691643 25.76985273868362,-80.1920809596777 25.769864212103958,-80.1920809596777 25.769872666202453,-80.1920735836029 25.769885951213148,-80.19206285476685 25.76988715894131,-80.1920534670353 25.76989138598981,-80.19203066825867 25.76989198985386,-80.19201993942261 25.769890782125756,-80.19201390445232 25.76988776280538,-80.19198440015316 25.769895613038162,-80.19196964800358 25.76989621690219,-80.19195556640625 25.76989984008637,-80.19195556640625 25.769902255542416,-80.191954895854 25.769902859406415,-80.1919388025999 25.76990708645436,-80.19193477928638 25.7699107096382,-80.19191198050976 25.76991252123005,-80.19190795719624 25.76991855986944,-80.1919025927782 25.769920975325107,-80.19189588725567 25.769921579189006,-80.19189186394215 25.769917352141597,-80.19188649952412 25.769917352141597,-80.19188046455383 25.76992459850852,-80.19187174737453 25.76992701396406,-80.19186839461327 25.76992701396406,-80.19186437129974 25.769923390780722,-80.19184961915016 25.769926410100187,-80.19182614982128 25.76992701396406,-80.19181407988071 25.769939091241035,-80.19180871546268 25.769939091241035,-80.19180536270142 25.76993365646655,-80.19179530441761 25.769928825555695,-80.19175171852112 25.769928825555695,-80.19174300134182 25.76992459850852,-80.19172959029675 25.769926410100187,-80.19171953201294 25.769917956005525,-80.19169874489307 25.769917352141597,-80.19169270992279 25.769915540549796,-80.19169203937054 25.769904067134426,-80.19168734550476 25.769896820766224,-80.19168399274349 25.769895009174128,-80.19168466329575 25.769884139620864,-80.19168198108673 25.769875685523186,-80.19167728722095 25.769871458474125,-80.19168265163898 25.769857569597576,-80.19168063998222 25.769841265262176,-80.19167594611645 25.76983824594057,-80.1916665583849 25.76984368071941,-80.19166052341461 25.76984428458371,-80.19165381789207 25.769841869126484,-80.1916491240263 25.769836434347553,-80.19163772463799 25.769835226618856,-80.19163101911545 25.769827376382104,-80.19162900745869 25.76980865658464,-80.19160956144333 25.76979355997163,-80.19160889089108 25.769783294273683,-80.1916142553091 25.769776651762772,-80.19161693751812 25.76976397060457,-80.1916216313839 25.76975974355153,-80.19162029027939 25.769747062391495,-80.19162431359291 25.769739212148913,-80.19162364304066 25.769722303932305,-80.19161090254784 25.769730154176038,-80.19160822033882 25.769736192824702,-80.19160754978657 25.769750081715447,-80.19159078598022 25.76975974355153,-80.1915592700243 25.769762762875132,-80.19155256450176 25.769756724227843,-80.19154250621796 25.769753101039303,-80.19153580069542 25.769754308768835,-80.19153647124767 25.769764574469274,-80.19156195223331 25.769784502002903,-80.19156865775585 25.76979295610708,-80.19157268106937 25.769800806346108,-80.19157268106937 25.769804429533195,-80.19156999886036 25.769807448855666,-80.19155859947205 25.769807448855666,-80.19155390560627 25.76980322180417,-80.19154720008373 25.76980322180417,-80.19154250621796 25.769808052720148,-80.19154317677021 25.76981288363592,-80.19153982400894 25.769814695229297,-80.19152037799358 25.76981529909374,-80.19151568412781 25.76981288363592,-80.19151233136654 25.769818318415957,-80.19150026142597 25.769824960924527,-80.19149959087372 25.76982798024648,-80.19148886203766 25.76982798024648,-80.19148550927639 25.769820733873658,-80.19148148596287 25.76981771455153,-80.19148014485836 25.769812279771468,-80.19149288535118 25.76980020248159,-80.19149757921696 25.769798390888,-80.19149489700794 25.769792352242497,-80.1914881914854 25.769795371565298,-80.19148886203766 25.769787521325885,-80.19149355590343 25.769783898138304,-80.19149288535118 25.769774236304166,-80.19149892032146 25.769770613116187,-80.191505625844 25.76976940538682,-80.19149892032146 25.769767593792754,-80.19149422645569 25.76976095128097,-80.19148416817188 25.769754912633587,-80.19148215651512 25.76974525079711,-80.1914881914854 25.769739816013733,-80.19148215651512 25.769736192824702,-80.19147410988808 25.769736192824702,-80.1914694160223 25.769739816013733,-80.19147478044033 25.76974947785065,-80.19147008657455 25.7697591396868,-80.19147075712681 25.769774840168843,-80.19146606326103 25.76978631359671,-80.19146136939526 25.769791748377934,-80.19146002829075 25.76980503339769,-80.19145600497723 25.769811675907007,-80.19145600497723 25.769818922280383,-80.19146136939526 25.769827376382104,-80.19144594669342 25.769829791839626,-80.19144125282764 25.769835226618856,-80.19143186509609 25.769835226618856,-80.19142583012581 25.769830999568352,-80.19140839576721 25.769829791839626,-80.19139900803566 25.769825564788924,-80.19138760864735 25.769824960924527,-80.19138291478157 25.769822545466912,-80.19138425588608 25.769802617939664,-80.19138894975185 25.769794163836185,-80.19138962030411 25.769781482679853,-80.19139297306538 25.769773028574864,-80.19137419760227 25.769765178333973,-80.19137687981129 25.76974947785065,-80.19135475158691 25.76974223147304,-80.19135273993015 25.76974041987856,-80.19135273993015 25.76973377736525,-80.19135676324368 25.76972653098668,-80.19135810434818 25.769716265282955,-80.19136145710945 25.769713849823127,-80.19136749207973 25.769713245958144,-80.19137017428875 25.769710226633276,-80.19137017428875 25.76969935706309,-80.19137419760227 25.76969392227763,-80.19137553870678 25.769681241110565,-80.1913795620203 25.769675806324276,-80.19138090312481 25.769667956076958,-80.1913969963789 25.769666748346562,-80.19140034914017 25.769663125155276,-80.19140169024467 25.76965104785023,-80.19140504300594 25.769643197601265,-80.19140303134918 25.769633535755695,-80.19141241908073 25.769626289371008,-80.19141308963299 25.769622666178485,-80.19141510128975 25.769622666178485,-80.19141778349876 25.7696305164288,-80.19142583012581 25.76962930869801,-80.19142985343933 25.769634139621065,-80.19143521785736 25.769629912563417,-80.19145131111145 25.769626289371008,-80.19145667552948 25.769622666178485,-80.19145600497723 25.76961481592767,-80.1914506405592 25.76961481592767,-80.19143924117088 25.769604550214254,-80.19142985343933 25.769600927021095,-80.1914244890213 25.769602134752155,-80.19142381846905 25.76960032315556,-80.19142784178257 25.76959851155891,-80.19142918288708 25.7695954922311,-80.19142381846905 25.769588849709685,-80.19142381846905 25.769575564665757,-80.19142784178257 25.769571337606003,-80.19142650067806 25.769560468023094,-80.19143253564835 25.769549598439188,-80.19143387675285 25.76953208632969,-80.19143722951412 25.769528463134314,-80.19143924117088 25.76951336648566,-80.1914432644844 25.769507931691688,-80.19144460558891 25.76948860797772,-80.19144997000694 25.7694741151902,-80.19145332276821 25.769470491993033,-80.19145533442497 25.769454791470725,-80.19145868718624 25.76944935667407,-80.191460698843 25.769436071614507,-80.19146472215652 25.769428221351344,-80.19146539270878 25.76941554015589,-80.19147075712681 25.769407086024838,-80.19147209823132 25.76939259322736,-80.19147545099258 25.769388970027705,-80.19147612154484 25.769384742961304,-80.19147880375385 25.76936058829331,-80.19147478044033 25.76935817282624,-80.19146673381329 25.769359984426533,-80.19146271049976 25.76935273802514,-80.19145667552948 25.769347907090626,-80.19146203994751 25.769341868422206,-80.19146740436554 25.769314694410536,-80.19147276878357 25.769310467341512,-80.19147410988808 25.76930563640526,-80.19147746264935 25.769302617070025,-80.19148416817188 25.769303220937076,-80.19148550927639 25.769298993867622,-80.1914881914854 25.76929778613348,-80.19149355590343 25.769300805468838,-80.19148752093315 25.769305032538227,-80.19149154424667 25.769309259607464,-80.19149422645569 25.769310467341512,-80.19149892032146 25.76930563640526,-80.1914955675602 25.769300805468838,-80.19150227308273 25.7692965783993,-80.19150093197823 25.769289935861156,-80.19150696694851 25.769267592775584,-80.19150696694851 25.769256723164858,-80.19149623811245 25.769237399409988,-80.19148752093315 25.769236795542607,-80.19148081541061 25.769232568470795,-80.19146673381329 25.76923075686853,-80.19146136939526 25.769225925929057,-80.1914506405592 25.7692241143267,-80.19144594669342 25.76922049112192,-80.19142851233482 25.76922109498939,-80.19141912460327 25.76921928338696,-80.19141644239426 25.769202375096306,-80.19141979515553 25.76919090161199,-80.19142046570778 25.769175804920394,-80.19140973687172 25.76917097397869,-80.19138626754284 25.76917097397869,-80.1913795620203 25.769175804920394,-80.19137620925903 25.769181843597266,-80.19137352705002 25.769181843597266,-80.1913607865572 25.769191505479604,-80.19135944545269 25.76920479056654,-80.19136011600494 25.76920901763935,-80.19136413931847 25.769212640844483,-80.19136413931847 25.76921988725445,-80.19135542213917 25.769235587807806,-80.19133262336254 25.76923619167522,-80.19132927060127 25.769234380073016,-80.19132122397423 25.76923498394041,-80.19131116569042 25.769232568470795,-80.19131518900394 25.769229549133694,-80.19131250679493 25.76922290659179,-80.1913172006607 25.769218075652,-80.1913172006607 25.769208413771807,-80.19132122397423 25.769202978963865,-80.1913245767355 25.769203582831416,-80.19132524728775 25.769213848579515,-80.19132792949677 25.769215660182034,-80.19133061170578 25.769212640844483,-80.1913332939148 25.769198751890837,-80.1913245767355 25.76919029774435,-80.19133061170578 25.769184862935575,-80.1913245767355 25.769181239729576,-80.19132390618324 25.769178824258855,-80.19133128225803 25.7691715778464,-80.19133128225803 25.769169162375487,-80.19132658839226 25.76916674690452,-80.19132658839226 25.769164331433505,-80.19133396446705 25.76916191596244,-80.19133932888508 25.76916493530126,-80.19133999943733 25.769160104359106,-80.19133664667606 25.76915708502017,-80.19133932888508 25.76915285794551,-80.19134268164635 25.769151046342035,-80.19135005772114 25.76915285794551,-80.19135475158691 25.76914681926716,-80.19136145710945 25.769147423135028,-80.19136548042297 25.769143799927885,-80.191370844841 25.769143799927885,-80.19137419760227 25.76913594964537,-80.19138358533382 25.769132930305815,-80.19138626754284 25.7691287032303,-80.19139632582664 25.769131722569984,-80.1914070546627 25.769138365116987,-80.19141308963299 25.769138365116987,-80.19141845405102 25.769149234738535,-80.1914244890213 25.769142592192143,-80.19142113626003 25.76913474190957,-80.19142046570778 25.769120852946813,-80.19142650067806 25.76910636011436,-80.19142247736454 25.76910152916982,-80.19142650067806 25.769094886620756,-80.19142985343933 25.769093075016425,-80.19143052399158 25.76908764020318,-80.19143588840961 25.76907918604877,-80.19143722951412 25.769069524157285,-80.19144125282764 25.769066504816042,-80.19144862890244 25.769051408108627,-80.1914519816637 25.769048992635263,-80.19145399332047 25.769041746214885,-80.19145667552948 25.769040538478098,-80.19144997000694 25.769035707530882,-80.19144997000694 25.76903389592562,-80.19145265221596 25.769030876583475,-80.19145131111145 25.76902483789892,-80.19145734608173 25.769020610819574,-80.19145734608173 25.769015176003013,-80.19144661724567 25.769007325712018,-80.1914419233799 25.769007929580567,-80.19143521785736 25.769014572134495,-80.19143119454384 25.76902483789892,-80.19142046570778 25.769029064978135,-80.19141845405102 25.76903389592562,-80.19141308963299 25.769035103662475,-80.191415771842 25.769039934609715,-80.19141308963299 25.769041746214885,-80.19141107797623 25.76905503131859,-80.19140169024467 25.769059258396734,-80.19139900803566 25.769064089342987,-80.1913983374834 25.769073147366694,-80.19139632582664 25.76907435510313,-80.19139230251312 25.769073751234913,-80.19139498472214 25.76906771255254,-80.19138358533382 25.769046577161856,-80.19138358533382 25.769030272715035,-80.19138559699059 25.769028461109688,-80.19138693809509 25.769019403082574,-80.19138224422932 25.769012156660384,-80.19137889146805 25.76901276052891,-80.19137620925903 25.769016383740052,-80.19136480987072 25.769016987608563,-80.19136279821396 25.769015779871523,-80.19136480987072 25.7690109489233,-80.1913607865572 25.769006721843436,-80.19134871661663 25.76900551410632,-80.19134603440762 25.76900913731766,-80.19134670495987 25.769016383740052,-80.19133999943733 25.769020610819574,-80.19133396446705 25.76901879921408,-80.1913158595562 25.76902000695106,-80.19131451845169 25.7690254417674,-80.19130580127239 25.769030272715035,-80.19130177795887 25.769039330741325,-80.19130446016788 25.76904416168839,-80.19130513072014 25.769059258396734,-80.19129976630211 25.769064089342987,-80.19130177795887 25.769069524157285,-80.1912970840931 25.769074958971355,-80.19129574298859 25.769083413126044,-80.19129171967506 25.769087036335034,-80.19129574298859 25.76909247114828,-80.19129976630211 25.769112398794768,-80.19130915403366 25.769116625870865,-80.19131384789944 25.769123268418713,-80.19131049513817 25.769127495494423,-80.1912984251976 25.769132326437898,-80.19129574298859 25.76913655351329,-80.19129104912281 25.76913655351329,-80.19128769636154 25.76914319606002,-80.1912796497345 25.76915044247421,-80.19127428531647 25.76915346181332,-80.19126757979393 25.76915406568113,-80.1912521570921 25.769146215399317,-80.1912534981966 25.76914078058853,-80.19125081598759 25.76913655351329,-80.19125148653984 25.769121456814783,-80.19125483930111 25.769112398794768,-80.19125953316689 25.769107567850476,-80.19125886261463 25.769097905961306,-80.19126355648041 25.769093678884538,-80.19125953316689 25.769080997653333,-80.19124679267406 25.769071335762003,-80.1912434399128 25.769065297079507,-80.1912447810173 25.769057446791827,-80.19124142825603 25.769044765556757,-80.19124813377857 25.769040538478098,-80.19124612212181 25.769029064978135,-80.1912534981966 25.76902000695106,-80.19124880433083 25.769015176003013,-80.19124880433083 25.769011552791856,-80.19123874604702 25.76900551410632,-80.19122332334518 25.769006721843436,-80.19120186567307 25.76900189089485,-80.19120253622532 25.76899404060297,-80.19120924174786 25.768981359361124,-80.19120052456856 25.76897713228021,-80.19120119512081 25.76897713228021,-80.1911985129118 25.768970489724175,-80.19120320677757 25.768962035561444,-80.19119784235954 25.768952373660404,-80.19120320677757 25.76894391949638,-80.19120320677757 25.7689378808074,-80.19120924174786 25.768933653724932,-80.19121527671814 25.768923991821605,-80.19122064113617 25.768923991821605,-80.1912260055542 25.76891795313161,-80.19123069941998 25.768919160869647,-80.19123405218124 25.76891070670323,-80.19123405218124 25.768903460274412,-80.1912260055542 25.768894402237752,-80.19121930003166 25.76889138289203,-80.19120454788208 25.768892590630333,-80.1911998540163 25.768889571284586,-80.1911985129118 25.768885948069574,-80.19119918346405 25.76887990937765,-80.19120454788208 25.768885344200402,-80.19121795892715 25.768873870685436,-80.19121930003166 25.768865416515794,-80.1912159472704 25.768861793300047,-80.19121527671814 25.768855754606907,-80.19120387732983 25.768855754606907,-80.1911998540163 25.76885756621489,-80.1911897957325 25.768855754606907,-80.19118510186672 25.768860585561455,-80.19118644297123 25.768872059077705,-80.19118443131447 25.768878701639235,-80.19117772579193 25.768881720985263,-80.19116565585136 25.768894402237752,-80.19116632640362 25.768899837059823,-80.19116029143333 25.768910102834184,-80.19117638468742 25.768928822773365,-80.19117437303066 25.768932445987073,-80.19117504358292 25.768944523365253,-80.19117973744869 25.76894693884074,-80.19117973744869 25.768948750447326,-80.19116565585136 25.768949958185026,-80.19115559756756 25.768948146578477,-80.19115224480629 25.76894573110299,-80.19114352762699 25.76894573110299,-80.19111402332783 25.768951165922733,-80.19111067056656 25.768952977529256,-80.19111067056656 25.768957808479843,-80.19110396504402 25.76896324329901,-80.19109591841698 25.768962639430224,-80.19108720123768 25.768967470380417,-80.19107848405838 25.76897713228021,-80.19108049571514 25.76897713228021,-80.19106842577457 25.768981963229805,-80.1910650730133 25.76899585220884,-80.19107580184937 25.768996456077446,-80.19107915461063 25.769000683157675,-80.19108049571514 25.769007325712018,-80.19108586013317 25.76900974118622,-80.19109189510345 25.769007929580567,-80.19109658896923 25.769010345054763,-80.19108653068542 25.76901879921408,-80.19108854234219 25.769030272715035,-80.19109658896923 25.769039934609715,-80.19109390676022 25.76904416168839,-80.19109390676022 25.76905503131859,-80.19109927117825 25.769059258396734,-80.19109725952148 25.769068920289044,-80.19108720123768 25.76907435510313,-80.19108653068542 25.76907918604877,-80.19108921289444 25.769084016994235,-80.19108720123768 25.76908582859871,-80.19107982516289 25.769086432466885,-80.19107580184937 25.769091263412033,-80.19107043743134 25.769091867280157,-80.19106239080429 25.769096698225088,-80.19104026257992 25.769100321433672,-80.19103355705738 25.76910575624631,-80.19102685153484 25.769102736905985,-80.19102483987808 25.76910817171852,-80.19101612269878 25.769116625870865,-80.19101545214653 25.76912206068275,-80.19101209938526 25.76912508002261,-80.19100673496723 25.769123268418713,-80.19099868834019 25.76912508002261,-80.19099600613117 25.769121456814783,-80.19099935889244 25.769116625870865,-80.1909913122654 25.769106963982427,-80.19099466502666 25.769102736905985,-80.19099466502666 25.769096094356996,-80.19100204110146 25.769093678884538,-80.19100271165371 25.769087036335034,-80.19100069999695 25.76908280925788,-80.19100271165371 25.769078582180583,-80.19101075828075 25.769074958971355,-80.19101478159428 25.769068316420807,-80.1910013705492 25.76905865452842,-80.19099466502666 25.769059258396734,-80.19099064171314 25.769065297079507,-80.1909738779068 25.769067108684293,-80.19097320735455 25.76907012802552,-80.19097588956356 25.769072543498453,-80.19097253680229 25.76907797831239,-80.1909738779068 25.769083413126044,-80.19096449017525 25.76908764020318,-80.19095845520496 25.769087036335034,-80.19095242023468 25.769081601521517,-80.19094973802567 25.769074958971355,-80.19095107913017 25.769068920289044,-80.19095979630947 25.769064693211252,-80.1909651607275 25.769059258396734,-80.19096650183201 25.769054427450257,-80.190963819623 25.76904838876692,-80.19096583127975 25.769046577161856,-80.19097991287708 25.769047181030206,-80.19098326563835 25.769045369425115,-80.19098058342934 25.769035103662475,-80.19098460674286 25.769030272715035,-80.19098192453384 25.7690254417674,-80.19098460674286 25.769022422425014,-80.19099198281765 25.769022422425014,-80.19099399447441 25.769020610819574,-80.19098594784737 25.769017591477063,-80.19098058342934 25.769003098632023,-80.19097924232483 25.768991625128436,-80.1909738779068 25.76898619031055,-80.19096650183201 25.768989813522506,-80.19096449017525 25.768996456077446,-80.19095778465271 25.769004910237754,-80.1909477263689 25.769008533449114,-80.19094102084637 25.769006721843436,-80.19093297421932 25.769001287026263,-80.19093297421932 25.768989209653856,-80.19093632698059 25.76897713228021,-80.19094571471214 25.76897713228021,-80.19094437360764 25.768971093592942,-80.19094839692116 25.768967470380417,-80.19093699753284 25.768963847167775,-80.19093163311481 25.768966866511658,-80.19093163311481 25.76897290519914,-80.19092559814453 25.76897290519914,-80.19092626869678 25.76897713228021,-80.19093096256256 25.76897713228021,-80.19091956317425 25.768978943886335,-80.19091621041298 25.768981359361124,-80.19091956317425 25.76898800191654,-80.19091822206974 25.768991625128436,-80.19091084599495 25.76899585220884,-80.1909101754427 25.76900189089485,-80.19091285765171 25.7690109489233,-80.19090816378593 25.769013968265977,-80.19089877605438 25.769015176003013,-80.1908940821886 25.769019403082574,-80.1908840239048 25.769011552791856,-80.19088536500931 25.768995248340204,-80.19088335335255 25.76897713228021,-80.19088067114353 25.76897713228021,-80.19087262451649 25.76896324329901,-80.19086457788944 25.768962639430224,-80.190873965621 25.768952977529256,-80.190873965621 25.7689475427096,-80.19087128341198 25.76894391949638,-80.19087195396423 25.7689378808074,-80.19086390733719 25.768928822773365,-80.19086323678493 25.76892338795263,-80.190873965621 25.76891855700062,-80.19088335335255 25.76891734926261,-80.1908940821886 25.76891191444134,-80.19090816378593 25.768898025452483,-80.19091822206974 25.768901044798035,-80.19092626869678 25.76889923319072,-80.19092828035355 25.76891191444134,-80.19093230366707 25.76891493378651,-80.19093632698059 25.76891493378651,-80.19093431532383 25.768912518310373,-80.19093431532383 25.76890406414349,-80.1909289509058 25.768895006106877,-80.19093833863735 25.76889077902289,-80.1909463852644 25.768872662946944,-80.19095242023468 25.768875682293142,-80.19095912575722 25.768874474554664,-80.19095979630947 25.768868435862156,-80.19094705581665 25.768868435862156,-80.19094571471214 25.768861189430748,-80.19094236195087 25.768858773953518,-80.190918892622 25.768861189430748,-80.19090548157692 25.768847904305343,-80.19089810550213 25.768846092697213,-80.19089810550213 25.76884367721967,-80.1908940821886 25.76884126174209,-80.1908940821886 25.76883522304791,-80.19089207053185 25.768836430786763,-80.1908927410841 25.76884126174209,-80.19088938832283 25.76884307335029,-80.19088804721832 25.76884730043598,-80.19088804721832 25.76884126174209,-80.19088335335255 25.768830392092337,-80.19087061285973 25.76882858048394,-80.19085854291916 25.768830995961792,-80.19085586071014 25.76882858048394,-80.19085519015789 25.768820126311077,-80.19085049629211 25.768820126311077,-80.19084647297859 25.768822541789092,-80.19084580242634 25.768833411439587,-80.1908390969038 25.768831599831234,-80.19084110856056 25.768837034656183,-80.19083641469479 25.768840054003284,-80.19082836806774 25.76885515073756,-80.19082501530647 25.768888363546253,-80.19082903862 25.76889319449947,-80.1908303797245 25.76890044092892,-80.19083641469479 25.76890406414349,-80.19083239138126 25.768908895096065,-80.19083373248577 25.768915537655538,-80.1908303797245 25.768919160869647,-80.19083708524704 25.768923991821605,-80.1908377557993 25.76892942664232,-80.19083574414253 25.768934257593862,-80.1908303797245 25.7689378808074,-80.1908303797245 25.76894391949638,-80.19082501530647 25.768952977529256,-80.19081696867943 25.76895720461103,-80.19081898033619 25.768960223955055,-80.19082568585873 25.768962035561444,-80.19082367420197 25.768966866511658,-80.19082568585873 25.76897290519914,-80.19082300364971 25.76897713228021,-80.19081093370914 25.76897713228021,-80.19080825150013 25.768981359361124,-80.19080825150013 25.76898800191654,-80.19081830978394 25.768996456077446,-80.1908129453659 25.769000079289064,-80.19079752266407 25.769003098632023,-80.19079148769379 25.76900611797488,-80.1907928287983 25.769020610819574,-80.1907854527235 25.769022422425014,-80.19077874720097 25.769036915267716,-80.19077807664871 25.769045369425115,-80.19077271223068 25.76905865452842,-80.19077338278294 25.769069524157285,-80.1907680183649 25.76907918604877,-80.1907666772604 25.769093075016425,-80.19076265394688 25.76910152916982,-80.19076131284237 25.769112398794768,-80.1907566189766 25.76912024907883,-80.19075527787209 25.769132930305815,-80.19075192511082 25.769139572852755,-80.19075125455856 25.769151650209853,-80.19074723124504 25.769164331433505,-80.19074723124504 25.769175804920394,-80.19073650240898 25.7692066021692,-80.19073382019997 25.769248269022487,-80.19072644412518 25.76925189222644,-80.19072711467743 25.76926215797034,-80.19073784351349 25.76927725465094,-80.19073247909546 25.76928148172117,-80.1907318085432 25.76928752039264,-80.19072644412518 25.76928752039264,-80.1907217502594 25.769282689455483,-80.19071571528912 25.769288728126913,-80.1907130330801 25.769298993867622,-80.19070833921432 25.769306844139354,-80.19071035087109 25.769311071208524,-80.1907217502594 25.769315298277547,-80.19070833921432 25.769317713745487,-80.19070364534855 25.76932435628208,-80.19070498645306 25.7693358297535,-80.19071236252785 25.76934005682164,-80.19071236252785 25.76934488775647,-80.19070766866207 25.769351530291537,-80.19070632755756 25.769364211493812,-80.1907043159008 25.76936723082751,-80.19070498645306 25.769372665627916,-80.19071504473686 25.769384742961304,-80.19072510302067 25.76939198936077,-80.19073247909546 25.769407689891352,-80.1907405257225 25.769408293757895,-80.1907405257225 25.769402255092544,-80.19074991345406 25.769398028026618,-80.1907492429018 25.76939198936077,-80.19075125455856 25.769387158427833,-80.19074521958828 25.76938232749473,-80.19075058400631 25.76937810042809,-80.19075125455856 25.769373269494615,-80.19075594842434 25.76936723082751,-80.19075326621532 25.769356361225906,-80.19075393676758 25.769339452954757,-80.19075192511082 25.76932918721755,-80.19074723124504 25.769325564015958,-80.19074723124504 25.769320733080345,-80.19075326621532 25.76931710987849,-80.19075930118561 25.76931710987849,-80.19077003002167 25.769318921479428,-80.19078075885773 25.769330394951375,-80.19078746438026 25.769328583350628,-80.19079014658928 25.769330394951375,-80.19080959260464 25.769331602685202,-80.1908129453659 25.76933462201971,-80.19082836806774 25.769337037487258,-80.19083440303802 25.7693406606885,-80.19085183739662 25.769341868422206,-80.1908565312624 25.76934488775647,-80.19087195396423 25.769346095490132,-80.19088000059128 25.769349718691103,-80.1908940821886 25.769350926424714,-80.19090749323368 25.76935515349232,-80.19091352820396 25.769354549625533,-80.19091419875622 25.769349718691103,-80.19091621041298 25.769349114824287,-80.19092760980129 25.76935938055977,-80.19094035029411 25.76936058829331,-80.19099064171314 25.769373873361324,-80.19100606441498 25.769375684961375,-80.19101142883301 25.76937870429478,-80.19104965031147 25.769385346827956,-80.19105903804302 25.769388970027705,-80.19107043743134 25.76938957389431,-80.191101282835 25.76939863189318,-80.1911187171936 25.769400443492874,-80.19112341105938 25.769403462825643,-80.19113816320896 25.769404670558714,-80.19114822149277 25.769408293757895,-80.19119381904602 25.769413124689944,-80.19123271107674 25.769422786553452,-80.19123673439026 25.769426409752075,-80.19125483930111 25.76942701361849,-80.19125819206238 25.769425202019207,-80.19125953316689 25.769431844549782,-80.19127629697323 25.769433052282587,-80.19128367304802 25.769436675480875,-80.19129641354084 25.76943727934727,-80.19130580127239 25.76944150641177,-80.19132524728775 25.76944271414447,-80.19133128225803 25.76944633734249,-80.19134804606438 25.769448148941436,-80.19135676324368 25.769454791470725,-80.1913520693779 25.76945781080211,-80.19135072827339 25.769465661063315,-80.19134670495987 25.769470491993033,-80.19133999943733 25.769507327825657,-80.19132927060127 25.769536313390844,-80.19132860004902 25.769545371378506,-80.19132390618324 25.769551410036577,-80.19132323563099 25.769562883486046,-80.19131049513817 25.769594888365543,-80.1913071423769 25.76962508164017,-80.19130311906338 25.76963474348645,-80.19130311906338 25.769643801466607,-80.1912984251976 25.769651651715503,-80.19129574298859 25.769667956076958,-80.19129239022732 25.769671579268092,-80.19129104912281 25.769693318412568,-80.19128702580929 25.769698753198053,-80.19128568470478 25.769708415038313,-80.191280990839 25.769718680742738,-80.19127562642097 25.769752497174544,-80.19127160310745 25.7697591396868,-80.19127026200294 25.76977363243952,-80.19126623868942 25.769779067221325,-80.19126422703266 25.76979295610708,-80.19126087427139 25.769798390888,-80.19125819206238 25.76981771455153,-80.19125282764435 25.769822545466912,-80.19124813377857 25.769823753195716,-80.19123740494251 25.769836434347553,-80.19122667610645 25.769835830483217,-80.19121997058392 25.769831603432735,-80.19120253622532 25.769829187975244,-80.19119516015053 25.769825564788924,-80.19116967916489 25.769818922280383,-80.19109457731247 25.769804429533195,-80.19108720123768 25.769801410210622,-80.1910288631916 25.769791144513377,-80.19099600613117 25.769782086544467,-80.19096851348877 25.76977725562743,-80.19095242023468 25.769771820845516,-80.1908927410841 25.76976155514571,-80.19081093370914 25.76974041987856,-80.19074521958828 25.76972773871649,-80.19073449075222 25.76972411552706,-80.19071638584137 25.769722303932305,-80.19068822264671 25.769713849823127,-80.1906855404377 25.76971143436325,-80.19068621098995 25.769696337737866,-80.19069224596024 25.769677014054572,-80.19069626927376 25.769672183133267,-80.19069761037827 25.769657690368163,-80.19070096313953 25.769651651715503,-80.1907029747963 25.769635347351805,-80.19070699810982 25.76962508164017,-80.19070565700531 25.769616023658585,-80.19070096313953 25.76960938113834,-80.19070364534855 25.769599719289996,-80.19071102142334 25.769596096096684,-80.19071102142334 25.769590661306477,-80.19071370363235 25.769586434247262,-80.19071370363235 25.769580395591,-80.19070900976658 25.76957616853142,-80.19070833921432 25.7695707337403,-80.19071035087109 25.769568318277503,-80.19071571528912 25.769568318277503,-80.1907217502594 25.76956107188883,-80.1907230913639 25.76954235204937,-80.19071772694588 25.769539332720157,-80.19071772694588 25.769529670866113,-80.19072040915489 25.76951759354748,-80.19073247909546 25.769504912361572,-80.1907217502594 25.769494042772607,-80.1907217502594 25.769490419576044,-80.19071906805038 25.769492231174333,-80.19071571528912 25.769492231174333,-80.1907116919756 25.76948860797772,-80.19070968031883 25.769490419576044,-80.19070833921432 25.769509139423697,-80.19070498645306 25.76951397035164,-80.19070498645306 25.769520009011305,-80.19070900976658 25.76952785926842,-80.1907029747963 25.76953148246381,-80.19069761037827 25.769540540451843,-80.1906955987215 25.769571337606003,-80.19068822264671 25.769575564665757,-80.19067414104939 25.769571941471686,-80.1906593888998 25.76956348735177,-80.19065402448177 25.769564695083233,-80.19065134227276 25.769568922143225,-80.19064128398895 25.769566506680395,-80.1906406134367 25.769587038112892,-80.19063591957092 25.76959005744089,-80.19063457846642 25.769596096096684,-80.19062787294388 25.769604550214254,-80.19062519073486 25.769616023658585,-80.19062049686909 25.76961964685131,-80.1906144618988 25.76961964685131,-80.1906057447195 25.769623270043926,-80.19059501588345 25.769622666178485,-80.19058831036091 25.76962025071675,-80.19058294594288 25.769621458447634,-80.19058227539062 25.76962206231306,-80.19058227539062 25.7696208545822,-80.19057624042034 25.7696208545822,-80.19056752324104 25.76962508164017,-80.19055679440498 25.769623270043926,-80.19053131341934 25.769610588869313,-80.1905319839716 25.769605757945303,-80.19052930176258 25.76960394634874,-80.1905232667923 25.76960153088662,-80.19051253795624 25.769600927021095,-80.19050247967243 25.7695954922311,-80.19049040973186 25.769594888365543,-80.1904883980751 25.769599719289996,-80.19046559929848 25.76960394634874,-80.1904609054327 25.76960877727283,-80.19046224653721 25.769623873909346,-80.1904521882534 25.76962930869801,-80.19044481217861 25.76963897054394,-80.19044749438763 25.769645613062554,-80.19045352935791 25.769648028523775,-80.19045755267143 25.76965285944607,-80.19045419991016 25.769660709694364,-80.19042603671551 25.769668559942172,-80.19041866064072 25.76967399472879,-80.19041664898396 25.769678221784886,-80.19041329622269 25.76967761791974,-80.19041329622269 25.76967278699846,-80.19040390849113 25.769655878772394,-80.19039519131184 25.76964742465846,-80.19038647413254 25.769643197601265,-80.19037440419197 25.769627497101833,-80.19036702811718 25.76962508164017,-80.19036300480366 25.7696208545822,-80.19035562872887 25.76962025071675,-80.1903522759676 25.769616627524066,-80.19034557044506 25.769617231389518,-80.19034020602703 25.76961481592767,-80.19034020602703 25.769609985003832,-80.19034422934055 25.769607569541837,-80.19035696983337 25.769610588869313,-80.1903623342514 25.7696142120622,-80.19036501646042 25.769611796600298,-80.19037440419197 25.769610588869313,-80.19038580358028 25.769605757945303,-80.19039183855057 25.769594888365543,-80.19038446247578 25.769585226516046,-80.19037373363972 25.76958220718792,-80.19036769866943 25.769586434247262,-80.19035696983337 25.769584018784787,-80.1903522759676 25.769580999456632,-80.19035160541534 25.769576772397073,-80.190334841609 25.769576772397073,-80.19033081829548 25.7695707337403,-80.19032545387745 25.769571337606003,-80.19032076001167 25.769577376262735,-80.19031472504139 25.769579187859698,-80.19028723239899 25.769576772397073,-80.190289914608 25.769572545337365,-80.1902885735035 25.76957012987461,-80.1902711391449 25.769569526008922,-80.19026577472687 25.769567714411806,-80.19026108086109 25.769562279620317,-80.19026108086109 25.769556240962803,-80.19025705754757 25.769553221633938,-80.19025705754757 25.76955080617079,-80.19026175141335 25.769543559781024,-80.1902624219656 25.76953510565911,-80.19025772809982 25.769521216743208,-80.19026443362236 25.769508535557687,-80.19026510417461 25.769501893031403,-80.19027650356293 25.76949162730825,-80.19027851521969 25.769486192513256,-80.19028320908546 25.76948196545032,-80.1902885735035 25.769460830133404,-80.1902885735035 25.769445733476143,-80.19028455018997 25.769440298679076,-80.19027382135391 25.769437883213623,-80.19025504589081 25.769427617484915,-80.19024163484573 25.769424598152774,-80.19023559987545 25.769417955621794,-80.19024096429348 25.76937810042809,-80.190244987607 25.769373269494615,-80.19024699926376 25.76935817282624,-80.19025102257729 25.769348510957446,-80.19025169312954 25.76933401815282,-80.19026108086109 25.769310467341512,-80.1902637630701 25.76929537066513,-80.19026845693588 25.76928570879125,-80.19026979804039 25.76927121597893,-80.19027382135391 25.76925793089942,-80.19027784466743 25.76925249609375,-80.1902798563242 25.769238003277394,-80.190289914608 25.76920901763935,-80.19029662013054 25.769170370110963,-80.19030064344406 25.769163123697968,-80.19030533730984 25.76913594964537,-80.1903086900711 25.769127495494423,-80.19030667841434 25.76912206068275,-80.19029729068279 25.769117229738857,-80.19029729068279 25.769097302093186,-80.19030131399632 25.76907737444419,-80.19030064344406 25.769064089342987,-80.19029460847378 25.769061070001584,-80.19028522074223 25.769059862265017,-80.19027717411518 25.769053823581945,-80.19027315080166 25.7690459732935,-80.19027650356293 25.769015176003013,-80.19028186798096 25.769004910237754,-80.19028656184673 25.76900189089485,-80.19029594957829 25.76900430636917,-80.19031874835491 25.769003702500598,-80.19032143056393 25.769000079289064,-80.19032746553421 25.76897713228021,-80.19032947719097 25.768960827823847,-80.1903335005045 25.768955996873387,-80.19034758210182 25.768955996873387,-80.1903522759676 25.768952977529256,-80.19035898149014 25.76891855700062,-80.19036300480366 25.768908895096065,-80.19036501646042 25.768896213845142,-80.19036903977394 25.76889077902289,-80.1903710514307 25.768878701639235,-80.19037373363972 25.768874474554664,-80.19037574529648 25.76885515073756,-80.19037976861 25.768846696566612,-80.19038110971451 25.76883522304791,-80.19038513302803 25.76882858048394,-80.19038647413254 25.76881227600718,-80.1903898268938 25.768805029572317,-80.19039116799831 25.7687923483103,-80.19039586186409 25.76878691348329,-80.19039586186409 25.768774836089072,-80.19039988517761 25.76876940126128,-80.19040122628212 25.768753096776358,-80.19040659070015 25.768744642598136,-80.1904072612524 25.768730149719747,-80.19041329622269 25.768725922629887,-80.19041396677494 25.76872048779985,-80.19041866064072 25.76871686457968,-80.19042067229748 25.76871082587917,-80.19043609499931 25.76870237169792,-80.19044280052185 25.768696332996647,-80.19043073058128 25.7686921059056,-80.19043140113354 25.768686067203816,-80.19042871892452 25.768682443982588,-80.19042938947678 25.768676405280328,-80.19042737782001 25.768671574318272,-80.19043140113354 25.768664327874845,-80.19043877720833 25.768664327874845,-80.1904434710741 25.768661912393604,-80.19044615328312 25.76866493174517,-80.19044950604439 25.76866493174517,-80.19045487046242 25.768661912393604,-80.1904595643282 25.76866553561546,-80.19046492874622 25.76866070465298,-80.19046559929848 25.768675801410073,-80.19046761095524 25.768680632371932,-80.19047565758228 25.768680632371932,-80.19048973917961 25.768684255593215,-80.19050918519497 25.76868365172302,-80.19051320850849 25.768680632371932,-80.19051723182201 25.76868787881439,-80.19052527844906 25.768689086554733,-80.19052997231483 25.768695125256375,-80.19053466618061 25.768696332996647,-80.19055478274822 25.76869270977576,-80.19055545330048 25.768690898165264,-80.19054740667343 25.768681840112382,-80.1905407011509 25.768679424631497,-80.19053734838963 25.76867398979934,-80.19054003059864 25.768667347226344,-80.19053600728512 25.768662516263916,-80.19053533673286 25.76865828917167,-80.19052863121033 25.768654062079246,-80.19052863121033 25.76864802337552,-80.19053131341934 25.768651042727406,-80.19054874777794 25.768652854338516,-80.19055344164371 25.768657081430984,-80.19057154655457 25.76865828917167,-80.19057422876358 25.768661308523296,-80.19058227539062 25.768661912393604,-80.19058965146542 25.768663120134246,-80.19059836864471 25.76866976270747,-80.19061110913754 25.768671574318272,-80.19061915576458 25.76867036657774,-80.19062787294388 25.76866493174517,-80.19063457846642 25.76864440015313,-80.1906406134367 25.768640173060234,-80.19064530730247 25.768643796282728,-80.19065134227276 25.76864440015313,-80.19065737724304 25.76864802337552,-80.19065871834755 25.768653458208867,-80.19065201282501 25.76865828917167,-80.19065134227276 25.768663120134246,-80.19065603613853 25.768671574318272,-80.19066207110882 25.768674593669584,-80.19067279994488 25.768671574318272,-80.19068486988544 25.768672178188552,-80.19068956375122 25.768667347226344,-80.19069157540798 25.7686685549669,-80.19070029258728 25.76866553561546,-80.19070163369179 25.768662516263916,-80.19069626927376 25.768658893041994,-80.19069962203503 25.768656477560647,-80.19070632755756 25.76865828917167,-80.1907130330801 25.768655269819938,-80.19072644412518 25.76865828917167,-80.19072711467743 25.768675801410073,-80.1907230913639 25.768681236242173,-80.1907217502594 25.768696936866807,-80.19071504473686 25.76870056008757,-80.19071772694588 25.768716260709642,-80.19071370363235 25.7687235071499,-80.19071504473686 25.76872894197982,-80.1907230913639 25.76873377293952,-80.19073650240898 25.768734980679405,-80.19074320793152 25.76873860389902,-80.19075863063335 25.76873981163888,-80.19076466560364 25.768743434858333,-80.19077941775322 25.768744642598136,-80.1907854527235 25.768747661947575,-80.1907928287983 25.7687470580777,-80.19079752266407 25.76874283098844,-80.1908028870821 25.768741623248605,-80.19080691039562 25.768744038728226,-80.19082367420197 25.768744642598136,-80.19082769751549 25.768747661947575,-80.19084580242634 25.768748869687336,-80.19085384905338 25.768741623248605,-80.1908565312624 25.76872894197982,-80.19087128341198 25.76872833810985,-80.19088134169579 25.768733169069574,-80.1908940821886 25.76873377293952,-80.19090078771114 25.768736792289225,-80.1909101754427 25.76873739615918,-80.19091688096523 25.768734376809466,-80.1909202337265 25.768724111019917,-80.19093297421932 25.768725922629887,-80.19093699753284 25.7687235071499,-80.19093833863735 25.768731357459693,-80.19093364477158 25.768740415508795,-80.19094370305538 25.76874283098844,-80.19095040857792 25.768747661947575,-80.190963819623 25.768749473557207,-80.19097320735455 25.768753700646233,-80.19098460674286 25.768754304516076,-80.19099332392216 25.768758531604924,-80.19100606441498 25.768759135474742,-80.19101612269878 25.768763362563416,-80.19103221595287 25.768764570303027,-80.1910362392664 25.768768193521726,-80.19105769693851 25.768773024479827,-80.19107446074486 25.768774232219332,-80.19110731780529 25.768782082525746,-80.19112542271614 25.768782082525746,-80.1911448687315 25.76878872509232,-80.19114889204502 25.7687923483103,-80.19115760922432 25.76879355604959,-80.1911623030901 25.7687989908763,-80.19117571413517 25.76880382183315,-80.19117705523968 25.76881227600718,-80.19118309020996 25.768820126311077,-80.19118912518024 25.768822541789092,-80.19119314849377 25.768834619178453,-80.19120052456856 25.768831599831234,-80.19120454788208 25.76883522304791,-80.19120991230011 25.768834619178453,-80.1912159472704 25.76884065787269,-80.19122198224068 25.76884126174209,-80.19122667610645 25.768839450133882,-80.1912434399128 25.768842469480898,-80.19124612212181 25.76884850817473,-80.19125282764435 25.768847904305343,-80.19125483930111 25.768849715913447,-80.19125819206238 25.76885756621489,-80.19127562642097 25.768856358476235,-80.19128434360027 25.768847904305343,-80.19128903746605 25.76884850817473,-80.19129037857056 25.76885515073756,-80.19129373133183 25.768852131390865,-80.19129440188408 25.768847904305343,-80.19128501415253 25.76884548882784,-80.19128300249577 25.768842469480898,-80.19128702580929 25.768835826917336,-80.19128501415253 25.768826768875517,-80.1912796497345 25.768822541789092,-80.19129239022732 25.768817106963485,-80.19128635525703 25.768809256659363,-80.1912796497345 25.768808652789804,-80.19127026200294 25.76880382183315,-80.19126825034618 25.768794763788886,-80.19126892089844 25.76878691348329,-80.19127160310745 25.76878510187425,-80.19127428531647 25.768776043828556,-80.19127897918224 25.768770609000796,-80.19127763807774 25.768762758693626,-80.19128032028675 25.768759135474742,-80.19128903746605 25.768764570303027,-80.1912883669138 25.76876879739149,-80.19128434360027 25.76877181674033,-80.19128434360027 25.768777251568014,-80.19129104912281 25.76877543995881,-80.19129373133183 25.768783894134856,-80.19130177795887 25.768788121222656,-80.19130110740662 25.76879114057098,-80.19129574298859 25.768792952179947,-80.1912883669138 25.768801406354747,-80.1912897080183 25.76880563344192,-80.19129976630211 25.768802010224345,-80.19130177795887 25.768803217963562,-80.19130177795887 25.76881831470252,-80.19130848348141 25.768822541789092,-80.1913071423769 25.768824957267064,-80.1912984251976 25.76882616500604,-80.1912970840931 25.768833411439587,-80.1913071423769 25.768834619178453,-80.19131183624268 25.768838242395034,-80.19131250679493 25.768826768875517,-80.1913172006607 25.768821937919597,-80.19132256507874 25.76883884626445,-80.19132725894451 25.768834619178453,-80.19133396446705 25.76883401530902,-80.1913420110941 25.768837034656183,-80.19134938716888 25.768836430786763,-80.19136145710945 25.768842469480898,-80.19137151539326 25.768842469480898,-80.19137553870678 25.768844281089084,-80.19137755036354 25.768846092697213,-80.19137687981129 25.768856962345552,-80.19138492643833 25.768866020385065,-80.19137218594551 25.768874474554664,-80.19137218594551 25.76888353259284,-80.19137419760227 25.768886551938746,-80.19138425588608 25.76889017515374,-80.19139029085636 25.768895006106877,-80.19139431416988 25.768894402237752,-80.19139565527439 25.76888896741541,-80.19139900803566 25.768886551938746,-80.19141241908073 25.768886551938746,-80.19141040742397 25.76888353259284,-80.19140303134918 25.76888353259284,-80.19140034914017 25.768873870685436,-80.1913882791996 25.768866020385065,-80.19139565527439 25.768858773953518,-80.19139565527439 25.76885454686823,-80.19139230251312 25.76885152752152,-80.19139431416988 25.768846696566612,-80.19141912460327 25.768846092697213,-80.19142851233482 25.768850923652174,-80.19143454730511 25.768859981692145,-80.1914419233799 25.768859377822825,-80.19144661724567 25.768861189430748,-80.191460698843 25.768859377822825,-80.19148349761963 25.768861189430748,-80.19148752093315 25.768866624254365,-80.19148282706738 25.76886964360069,-80.19147880375385 25.768878097770028,-80.19146136939526 25.768877493900796,-80.19145600497723 25.768874474554664,-80.1914519816637 25.76887990937765,-80.19144460558891 25.768882324854445,-80.19143991172314 25.768880513246867,-80.19143521785736 25.768873870685436,-80.1914244890213 25.768873266816176,-80.19141644239426 25.768885948069574,-80.1914244890213 25.768885948069574,-80.19143052399158 25.768888363546253,-80.19142918288708 25.768895609976003,-80.19143521785736 25.768905271881664,-80.1914419233799 25.768901044798035,-80.19144527614117 25.768894402237752,-80.1914506405592 25.768898629321612,-80.19145734608173 25.76889923319072,-80.19145600497723 25.768901648667125,-80.19144997000694 25.76890406414349,-80.19145734608173 25.768908895096065,-80.19145600497723 25.76891131057227,-80.19144997000694 25.768913726048442,-80.19145801663399 25.76891855700062,-80.19145533442497 25.76892338795263,-80.19145533442497 25.76893003051129,-80.19146338105202 25.768933653724932,-80.19146740436554 25.768927011166472,-80.1914680749178 25.76893606920065,-80.19147276878357 25.768940900151918,-80.19148081541061 25.768939088545235,-80.19149288535118 25.768928822773365,-80.19149489700794 25.76892338795263,-80.19149824976921 25.76893848467631,-80.19151367247105 25.7689475427096,-80.19151233136654 25.768953581398083,-80.19150897860527 25.76895720461103,-80.19151166081429 25.768962035561444,-80.19151635468006 25.768962035561444,-80.19152708351612 25.76895841234864,-80.19153110682964 25.76895478913573,-80.19153781235218 25.768953581398083,-80.19154116511345 25.768948750447326,-80.19154720008373 25.76894573110299,-80.1915592700243 25.768945127234122,-80.19159279763699 25.768934861462807,-80.19160620868206 25.76892338795263,-80.19163303077221 25.768910102834184,-80.19164174795151 25.768902856405326,-80.19164443016052 25.76889319449947,-80.19165180623531 25.76889319449947,-80.19165789796502 25.768894191942604,-80.191616 25.76905,-80.191641 25.769056,-80.191622 25.769128,-80.191643 25.769132,-80.191644 25.769127,-80.191816 25.769165,-80.191834 25.769098,-80.191844 25.769101,-80.19185 25.769076,-80.191876 25.769082,-80.191865 25.769121,-80.191965 25.769143,-80.191958 25.76917,-80.191917 25.769161,-80.191909 25.769191,-80.191953 25.769201,-80.191975 25.769118,-80.19202294384887 25.769129063965124,-80.19202262163162 25.76913051483411),(-80.19197098910809 25.769864212103958,-80.19196763634682 25.76986662756073,-80.19195556640625 25.76986723142491,-80.19195556640625 25.769882931892653,-80.19196294248104 25.769881120300347,-80.1919736713171 25.76986602369655,-80.19197098910809 25.769864212103958),(-80.19097454845905 25.76897713228021,-80.19097991287708 25.76897713228021,-80.19097991287708 25.76896565877411,-80.1909738779068 25.768966866511658,-80.19096985459328 25.768971697461673,-80.19097186625004 25.76897713228021,-80.1909738779068 25.76897713228021,-80.19097454845905 25.76897713228021),(-80.1910275220871 25.76908824407132,-80.1910188049078 25.76908824407132,-80.19101344048977 25.76909247114828,-80.1910188049078 25.769099113697482,-80.19102819263935 25.76910152916982,-80.19102953374386 25.76909247114828,-80.1910275220871 25.76908824407132),(-80.19074186682701 25.769460226267157,-80.19073717296124 25.76945720693583,-80.19072040915489 25.769460226267157,-80.19072040915489 25.769470491993033,-80.19072711467743 25.76947653065491,-80.19073247909546 25.7694741151902,-80.19073583185673 25.769468680394404,-80.19074186682701 25.76946445333087,-80.19074186682701 25.769460226267157),(-80.19069023430347 25.76939983962632,-80.19069157540798 25.769391385494142,-80.19069023430347 25.769388970027705,-80.19068822264671 25.769399235759764,-80.19069023430347 25.76939983962632),(-80.19169002771378 25.769802617939664,-80.19168734550476 25.769798994752524,-80.19168265163898 25.769802014075157,-80.1916766166687 25.769821941602483,-80.19167862832546 25.769824357060113,-80.19168935716152 25.769822545466912,-80.19169136881828 25.76981288363592,-80.19169002771378 25.769802617939664),(-80.19167594611645 25.769783294273683,-80.19167996942997 25.76978087881521,-80.19167728722095 25.769775444033492,-80.19167058169842 25.76977967108597,-80.19167594611645 25.769783294273683),(-80.19153110682964 25.76962930869801,-80.19152842462063 25.769623873909346,-80.19152507185936 25.76962447777476,-80.19152574241161 25.769640178274614,-80.1915331184864 25.76964621692786,-80.19153714179993 25.76965889809864,-80.19154787063599 25.769661917424834,-80.1915492117405 25.769657690368163,-80.19155390560627 25.769654067176596,-80.19155658781528 25.769661917424834,-80.1915592700243 25.769663729020515,-80.19157133996487 25.76965346331133,-80.19156999886036 25.769645009197223,-80.19155994057655 25.76963957440928,-80.19156329333782 25.769626289371008,-80.19156195223331 25.7696142120622,-80.19156396389008 25.76960877727283,-80.1915592700243 25.769600927021095,-80.1915492117405 25.7696027386177,-80.1915418356657 25.769611796600298,-80.19153580069542 25.769613608196718,-80.19153110682964 25.76962930869801),(-80.19154317677021 25.76953027473203,-80.19153714179993 25.769529670866113,-80.19153378903866 25.76953510565911,-80.1915331184864 25.7695429559152,-80.19153580069542 25.769545371378506,-80.1915418356657 25.7695429559152,-80.19154720008373 25.769545371378506,-80.19155323505402 25.76954235204937,-80.19155792891979 25.769544767512695,-80.19156329333782 25.76954235204937,-80.1915693283081 25.769543559781024,-80.19157133996487 25.769539332720157,-80.19156999886036 25.76953087859792,-80.1915606111288 25.76952785926842,-80.19155256450176 25.76953027473203,-80.19154720008373 25.769526651536573,-80.19154317677021 25.76953027473203),(-80.1915693283081 25.7695671105461,-80.1915592700243 25.76956348735177,-80.19155658781528 25.769565902814666,-80.19155859947205 25.76957496080007,-80.19156195223331 25.769569526008922,-80.1915693283081 25.7695671105461),(-80.19155457615852 25.76958341491917,-80.19154787063599 25.769581603322273,-80.19154116511345 25.769585226516046,-80.19154720008373 25.7695954922311,-80.19155390560627 25.769597303827794,-80.19155457615852 25.769590661306477,-80.19155725836754 25.7695894535753,-80.1915606111288 25.769582811053542,-80.1915606111288 25.769578583994054,-80.19155859947205 25.769576772397073,-80.19155457615852 25.76958341491917),(-80.19151635468006 25.769707207308333,-80.19150294363499 25.769712038228214,-80.19150026142597 25.769718680742738,-80.19149959087372 25.76973136190578,-80.1915042847395 25.76973498509498,-80.19151300191879 25.7697331735004,-80.19151769578457 25.769741023743403,-80.19152641296387 25.769744646932292,-80.19153244793415 25.769741627608223,-80.1915404945612 25.769747666256308,-80.19154720008373 25.769747062391495,-80.19155323505402 25.769741023743403,-80.19155994057655 25.76974041987856,-80.19154854118824 25.76972653098668,-80.19153714179993 25.769728342581374,-80.19153378903866 25.76973136190578,-80.1915317773819 25.76973800441922,-80.19151769578457 25.769734381230137,-80.19151568412781 25.76972955031116,-80.19151836633682 25.76972411552706,-80.19151970744133 25.76971083049825,-80.19151635468006 25.769707207308333),(-80.19146539270878 25.769540540451843,-80.19146271049976 25.769538124988436,-80.19146002829075 25.769541748183542,-80.19146203994751 25.76954899457339,-80.19147276878357 25.76954899457339,-80.1914781332016 25.7695429559152,-80.19147410988808 25.769540540451843,-80.19146539270878 25.769540540451843),(-80.1914244890213 25.769154669548957,-80.19141979515553 25.769151650209853,-80.19140772521496 25.769151046342035,-80.1914057135582 25.76916674690452,-80.19140772521496 25.769169162375487,-80.19141979515553 25.769166143036774,-80.19142046570778 25.769160104359106,-80.1914244890213 25.76915708502017,-80.1914244890213 25.769154669548957),(-80.19144058227539 25.768864812646513,-80.19143588840961 25.76886360490793,-80.19143052399158 25.768864812646513,-80.19143052399158 25.768867228123625,-80.19144125282764 25.768866624254365,-80.19144058227539 25.768864812646513),(-80.1914231479168 25.768895609976003,-80.19141979515553 25.768892590630333,-80.19141376018524 25.76889319449947,-80.19141308963299 25.768895006106877,-80.19141845405102 25.768898629321612,-80.1914231479168 25.768895609976003),(-80.19130311906338 25.76884307335029,-80.19129775464535 25.76884307335029,-80.19129574298859 25.768846696566612,-80.1912970840931 25.768855754606907,-80.19130446016788 25.768861189430748,-80.19131183624268 25.768863001038643,-80.19131183624268 25.76887628616236,-80.19131653010845 25.768880513246867,-80.1913172006607 25.76888474033123,-80.19131988286972 25.76887990937765,-80.19133262336254 25.768875682293142,-80.1913332939148 25.768866020385065,-80.19132323563099 25.768861189430748,-80.19132256507874 25.768855754606907,-80.19131183624268 25.768850923652174,-80.19130982458591 25.76884730043598,-80.19130311906338 25.76884307335029),(-80.19121661782265 25.76896324329901,-80.19121259450912 25.768957808479843,-80.1912085711956 25.768957808479843,-80.1912085711956 25.768962639430224,-80.19121661782265 25.76896324329901),(-80.19096046686172 25.76881952244157,-80.1909463852644 25.768820126311077,-80.19094303250313 25.768822541789092,-80.19094303250313 25.768824957267064,-80.19094705581665 25.76882858048394,-80.19095309078693 25.768827372744987,-80.19095309078693 25.768844281089084,-80.19095778465271 25.768853339129567,-80.19096180796623 25.768856962345552,-80.19097454845905 25.768860585561455,-80.1909752190113 25.768866020385065,-80.19097790122032 25.768868435862156,-80.19098460674286 25.768867831992893,-80.19098728895187 25.768862397169357,-80.19099399447441 25.768858773953518,-80.19099533557892 25.768855754606907,-80.19097991287708 25.76884126174209,-80.19097723066807 25.768833411439587,-80.19097991287708 25.768827976614457,-80.19097857177258 25.768822541789092,-80.1909738779068 25.768820126311077,-80.19096046686172 25.76881952244157),(-80.19095376133919 25.768885344200402,-80.19095711410046 25.768879305508438,-80.19095309078693 25.76887628616236,-80.19095107913017 25.76888353259284,-80.19095376133919 25.768885344200402),(-80.19060708582401 25.768680632371932,-80.1906144618988 25.768677009150565,-80.19060306251049 25.768676405280328,-80.19060172140598 25.768682443982588,-80.190604403615 25.768684255593215,-80.19060708582401 25.768680632371932))",
		},

		// Reproduces https://github.com/peterstace/simplefeatures/issues/630.
		{
			input1: "POLYGON((112.409282 34.638641,112.409268 34.638646,112.40926 34.638642,112.409252 34.638642,112.408256 34.63827,112.408249 34.638271,112.408247 34.638254,112.408246 34.638259,112.408248 34.638249,112.408352 34.638021,112.408362 34.638014,112.408362 34.638011,112.408375 34.638008,112.408376 34.63801,112.409396 34.638386,112.409402 34.638393,112.409411 34.638395,112.409404 34.638407,112.409288 34.638638,112.409282 34.638641))",
			input2: "POLYGON((112.3681640625 34.6728515625,112.3681640625 34.62890625,112.412109375 34.62890625,112.412109375 34.6728515625,112.3681640625 34.6728515625))",
			inter:  "POLYGON((112.409288 34.638638,112.409404 34.638407,112.409411 34.638395,112.409402 34.638393,112.409396 34.638386,112.408376 34.63801,112.408375 34.638008,112.408362 34.638011,112.408362 34.638014,112.408352 34.638021,112.408248 34.638249,112.408246 34.638259,112.408247 34.638254,112.408249 34.638271,112.408256 34.63827,112.409252 34.638642,112.40926 34.638642,112.409268 34.638646,112.409282 34.638641,112.409288 34.638638))",
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

// dimToEmpty maps a dimension to the canonical "empty" geometry for that dimension.
var dimToEmpty = map[int]string{
	-1: "GEOMETRYCOLLECTION EMPTY",
	0:  "POINT EMPTY",
	1:  "LINESTRING EMPTY",
	2:  "POLYGON EMPTY",
}

func TestOverlayAndRelateBothInputsEmpty(t *testing.T) {
	inputs := []struct {
		wkt string
		dim int
	}{
		{"POINT EMPTY", 0},
		{"LINESTRING EMPTY", 1},
		{"POLYGON EMPTY", 2},
		{"MULTIPOINT EMPTY", 0},
		{"MULTILINESTRING EMPTY", 1},
		{"MULTIPOLYGON EMPTY", 2},
		{"GEOMETRYCOLLECTION EMPTY", -1},
	}

	for i, inputA := range inputs {
		for j, inputB := range inputs {
			gA := geomFromWKT(t, inputA.wkt)
			gB := geomFromWKT(t, inputB.wkt)
			t.Run(fmt.Sprintf("%d_%d", i, j), func(t *testing.T) {
				t.Run("union", func(t *testing.T) {
					got, err := geom.Union(gA, gB)
					test.NoErr(t, err)
					test.ExactEqualsWKT(t, got, dimToEmpty[maxInt(inputA.dim, inputB.dim)])
				})
				t.Run("intersection", func(t *testing.T) {
					got, err := geom.Intersection(gA, gB)
					test.NoErr(t, err)
					test.ExactEqualsWKT(t, got, dimToEmpty[minInt(inputA.dim, inputB.dim)])
				})
				t.Run("difference", func(t *testing.T) {
					got, err := geom.Difference(gA, gB)
					test.NoErr(t, err)
					test.ExactEqualsWKT(t, got, dimToEmpty[inputA.dim])
				})
				t.Run("symmetric_difference", func(t *testing.T) {
					got, err := geom.SymmetricDifference(gA, gB)
					test.NoErr(t, err)
					test.ExactEqualsWKT(t, got, dimToEmpty[maxInt(inputA.dim, inputB.dim)])
				})
				t.Run("relate", func(t *testing.T) {
					got, err := geom.Relate(gA, gB)
					test.NoErr(t, err)
					test.Eq(t, got, "FFFFFFFF2")
				})
			})
		}
	}
}

func TestOverlayAndRelateOnlyOneInputEmpty(t *testing.T) {
	emptyInputs := []struct {
		wkt string
		dim int
	}{
		{"POINT EMPTY", 0},
		{"LINESTRING EMPTY", 1},
		{"POLYGON EMPTY", 2},
		{"MULTIPOINT EMPTY", 0},
		{"MULTILINESTRING EMPTY", 1},
		{"MULTIPOLYGON EMPTY", 2},
		{"GEOMETRYCOLLECTION EMPTY", -1},
	}

	nonEmptyInputs := []struct {
		wkt string
		dim int
	}{
		{"POINT(0 0)", 0},
		{"LINESTRING(10 10,11 11)", 1},
		{"POLYGON((20 20,20 21,21 20,20 20))", 2},
		{"MULTIPOINT((30 30),(31 31))", 0},
		{"MULTILINESTRING((40 40,41 41),(42 42,43 43))", 1},
		{"MULTIPOLYGON(((50 50,50 51,51 50,50 50)),((52 52,52 53,53 52,52 52)))", 2},
		{"GEOMETRYCOLLECTION(POINT(60 60),LINESTRING(70 70,71 71),POLYGON((80 80,80 81,81 80,80 80)))", 2},
	}

	fwdRelate := map[int]string{
		0: "FFFFFF0F2",
		1: "FFFFFF102",
		2: "FFFFFF212",
	}
	revRelate := map[int]string{
		0: "FF0FFFFF2",
		1: "FF1FF0FF2",
		2: "FF2FF1FF2",
	}

	for i, emptyInput := range emptyInputs {
		emptyGeom := geomFromWKT(t, emptyInput.wkt)
		for j, nonEmptyInput := range nonEmptyInputs {
			nonEmptyGeom := geomFromWKT(t, nonEmptyInput.wkt)
			t.Run(fmt.Sprintf("%d_%d", i, j), func(t *testing.T) {
				t.Run("union", func(t *testing.T) {
					t.Run("fwd", func(t *testing.T) {
						got, err := geom.Union(emptyGeom, nonEmptyGeom)
						test.NoErr(t, err)
						test.ExactEquals(t, got, nonEmptyGeom, geom.IgnoreOrder)
					})
					t.Run("rev", func(t *testing.T) {
						got, err := geom.Union(nonEmptyGeom, emptyGeom)
						test.NoErr(t, err)
						test.ExactEquals(t, got, nonEmptyGeom, geom.IgnoreOrder)
					})
				})
				t.Run("intersection", func(t *testing.T) {
					t.Run("fwd", func(t *testing.T) {
						got, err := geom.Intersection(emptyGeom, nonEmptyGeom)
						test.NoErr(t, err)
						test.ExactEqualsWKT(t, got, dimToEmpty[minInt(emptyInput.dim, nonEmptyInput.dim)])
					})
					t.Run("rev", func(t *testing.T) {
						got, err := geom.Intersection(nonEmptyGeom, emptyGeom)
						test.NoErr(t, err)
						test.ExactEqualsWKT(t, got, dimToEmpty[minInt(emptyInput.dim, nonEmptyInput.dim)])
					})
				})
				t.Run("difference", func(t *testing.T) {
					t.Run("fwd", func(t *testing.T) {
						got, err := geom.Difference(emptyGeom, nonEmptyGeom)
						test.NoErr(t, err)
						test.ExactEqualsWKT(t, got, dimToEmpty[emptyInput.dim])
					})
					t.Run("rev", func(t *testing.T) {
						got, err := geom.Difference(nonEmptyGeom, emptyGeom)
						test.NoErr(t, err)
						test.ExactEquals(t, got, nonEmptyGeom, geom.IgnoreOrder)
					})
				})
				t.Run("symmetric_difference", func(t *testing.T) {
					t.Run("fwd", func(t *testing.T) {
						got, err := geom.SymmetricDifference(emptyGeom, nonEmptyGeom)
						test.NoErr(t, err)
						test.ExactEquals(t, got, nonEmptyGeom, geom.IgnoreOrder)
					})
					t.Run("rev", func(t *testing.T) {
						got, err := geom.SymmetricDifference(nonEmptyGeom, emptyGeom)
						test.NoErr(t, err)
						test.ExactEquals(t, got, nonEmptyGeom, geom.IgnoreOrder)
					})
				})
				t.Run("relate", func(t *testing.T) {
					t.Run("fwd", func(t *testing.T) {
						got, err := geom.Relate(emptyGeom, nonEmptyGeom)
						test.NoErr(t, err)
						test.Eq(t, got, fwdRelate[nonEmptyInput.dim])
					})
					t.Run("rev", func(t *testing.T) {
						got, err := geom.Relate(nonEmptyGeom, emptyGeom)
						test.NoErr(t, err)
						test.Eq(t, got, revRelate[nonEmptyInput.dim])
					})
				})
			})
		}
	}
}

func TestIntersectionEnvelopesDisjoint(t *testing.T) {
	inputs := []struct {
		wkt string
		dim int
	}{
		{"POINT(0 0)", 0},
		{"LINESTRING(10 10,11 11)", 1},
		{"POLYGON((20 20,20 21,21 20,20 20))", 2},
		{"MULTIPOINT((30 30),(31 31))", 0},
		{"MULTILINESTRING((40 40,41 41),(42 42,43 43))", 1},
		{"MULTIPOLYGON(((50 50,50 51,51 50,50 50)),((52 52,52 53,53 52,52 52)))", 2},
		{"GEOMETRYCOLLECTION(POINT(60 60),LINESTRING(70 70,71 71),POLYGON((80 80,80 81,81 80,80 80)))", 2},
	}

	for i, inputA := range inputs {
		gA := geomFromWKT(t, inputA.wkt)
		for j, inputB := range inputs {
			gB := geomFromWKT(t, inputB.wkt).TransformXY(func(xy geom.XY) geom.XY {
				return xy.Add(geom.XY{X: 100, Y: 100})
			})
			t.Run(fmt.Sprintf("%d_%d", i, j), func(t *testing.T) {
				got, err := geom.Intersection(gA, gB)
				test.NoErr(t, err)
				want := dimToEmpty[minInt(inputA.dim, inputB.dim)]
				test.ExactEqualsWKT(t, got, want)
			})
		}
	}
}

func maxInt(a, b int) int {
	// Once on Go 1.21, we can use max builtin instead.
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	// Once on Go 1.21, we can use min builtin instead.
	if a < b {
		return a
	}
	return b
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
			wantWKT:   "MULTILINESTRING((0 0,0 1),(0 1,1 1))",
		},
		{
			inputWKTs: []string{"MULTILINESTRING((0 0,0 1,1 1),(2 2,3 3))", "MULTILINESTRING((1 1,0 1,0 0),(2 2,3 3))"},
			wantWKT:   "MULTILINESTRING((0 0,0 1),(0 1,1 1),(2 2,3 3))",
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

		// Reproduces https://github.com/peterstace/simplefeatures/issues/668.
		{
			inputWKTs: []string{
				"MULTIPOLYGON(((-87.62421257793903 41.39764795906342,-87.62420654296875 41.397646450042785,-87.62420654296875 41.39764695304968,-87.62416161596775 41.397637395918245,-87.62415960431099 41.39764393500832,-87.62410126626492 41.39763085682753,-87.62410339298516 41.39762487432261,-87.62410339336095 41.397624874409054,-87.62421344649789 41.39764576585199,-87.62421257793903 41.39764795906342)),((-87.62404158711433 41.39745681617114,-87.62420319020748 41.39749403877844,-87.6242026502297 41.397495364425644,-87.62420264993521 41.39749536435664,-87.62413840985234 41.397482951866046,-87.62403983775401 41.39746055108124,-87.62404158711433 41.39745681617114)),((-87.62391954660416 41.39758105911507,-87.62392119420248 41.39757693935211,-87.62392119442447 41.39757693941165,-87.6239291859734 41.397578755513464,-87.62392771482985 41.39758304634879,-87.62391954660416 41.39758105911507)),((-87.62402482330799 41.39761325158001,-87.62396648526192 41.39760067640029,-87.62396708180846 41.397598998302605,-87.62396708218431 41.39759899838907,-87.62402591965431 41.39761016753592,-87.62402482330799 41.39761325158001)))",
				"MULTIPOLYGON(((-87.62422129511833 41.397514159097845,-87.6242246478796 41.39751063804239,-87.62422859770999 41.39751063804239,-87.62422934174538 41.39751063804239,-87.62423202395439 41.39751315308203,-87.62424007058144 41.397514159097845,-87.6242434233427 41.397517177145225,-87.6242434233427 41.39752069820032,-87.62423946264127 41.397522678919835,-87.62422593614133 41.39751938122142,-87.62422129511833 41.397514159097845)),((-87.62420654296875 41.39748599064894,-87.62420654296875 41.397485487640814,-87.62420251965523 41.39748498463269,-87.62419891701943 41.39748048049657,-87.624207178155 41.39748207671596,-87.62420893681185 41.39748248885456,-87.62421063126413 41.3974830325109,-87.6242122442154 41.397483702135574,-87.62421375920124 41.39748449089327,-87.62421516075717 41.39748539073262,-87.6242164345766 41.39748639246838,-87.62421756765681 41.39748748587517,-87.62421854843173 41.397488659791875,-87.62421936688986 41.39748990223556,-87.62422001467672 41.397491200523774,-87.62422048517993 41.39749254140405,-87.62422077359668 41.39749391118913,-87.62422087698293 41.39749529589673,-87.62422082211445 41.39749621512772,-87.62420654296875 41.39748599064894)),((-87.62420654296875 41.39765701318648,-87.62420654296875 41.39765600717286,-87.62421056628227 41.397653995145596,-87.62421324849129 41.39765449815243,-87.62421701337621 41.3976592051252,-87.62421590008874 41.39765935988056,-87.62421416603986 41.39765947381383,-87.62421242546463 41.39765946270053,-87.62421069423094 41.397659326641985,-87.62420898812152 41.39765906687855,-87.62420899357804 41.397659067915036,-87.62420604474384 41.39765850813634,-87.62420654296875 41.39765701318648)),((-87.6240348815918 41.397450780070656,-87.62402415275574 41.39746084023782,-87.62401610612869 41.39746084023782,-87.62401387572902 41.39745665745553,-87.6240215021472 41.39745842680161,-87.62402264165466 41.397455993937264,-87.62402329666267 41.39745478277123,-87.62402410175937 41.39745362403454,-87.62402504963794 41.397452528243555,-87.62402613169566 41.3974515053434,-87.62402733811206 41.39745056461764,-87.62402865793804 41.39744971460401,-87.62403007919521 41.39744896301703,-87.62403158898465 41.39744831667789,-87.62403317360392 41.397447781452556,-87.62403481867145 41.39744736219862,-87.62403650925704 41.39744706272111,-87.62403823001745 41.397446885738006,-87.62403996533548 41.39744683285552,-87.62404169946193 41.39744690455365,-87.62404341665828 41.39744710018164,-87.62404510133979 41.397447417964045,-87.62405168073352 41.39744891314965,-87.6240348815918 41.397450780070656)),((-87.62401409447193 41.39761777864412,-87.62403421103954 41.39761727563702,-87.62404292821884 41.39761274857286,-87.6240523159504 41.39761224556572,-87.62406505644321 41.3976167726299,-87.62407913804054 41.39761777864412,-87.62409187853336 41.39762934780651,-87.62410193681717 41.3976333718625,-87.62410461902618 41.397637395918245,-87.62410445659496 41.397639223606355,-87.62409893975861 41.397638176342504,-87.62409721290705 41.397637779093365,-87.62409554671505 41.39763725516227,-87.62409395752817 41.39763660968908,-87.62409246093647 41.39763584900594,-87.62409107162166 41.39763498057522,-87.62409054781364 41.397634580966205,-87.62403539291145 41.397621878478574,-87.62403515702614 41.397621994791656,-87.62403358694466 41.39762261406486,-87.62403194441826 41.3976231156037,-87.62403024497316 41.39762349466726,-87.62402850467363 41.39762374767239,-87.62402673997009 41.39762387222754,-87.6240249675436 41.39762386715533,-87.62402320414834 41.39762373250369,-87.62402146645303 41.39762346954546,-87.6240114235828 41.39762156310229,-87.62401409447193 41.39761777864412)),((-87.62397654354572 41.39745228909582,-87.6239725202322 41.39745731917949,-87.6239624619484 41.397458325196176,-87.62395575642586 41.397455307146075,-87.62395441532135 41.39745128307905,-87.62395534881179 41.39745034941373,-87.62395626638761 41.39744972341221,-87.62395770979515 41.39744891829488,-87.62395925142329 41.397448224048624,-87.62396087630455 41.39744764741381,-87.62396256866319 41.39744719398892,-87.62396431206825 41.397446868176175,-87.62396608959328 41.39744667313887,-87.62396788398048 41.397446610770594,-87.62396964164054 41.39744668024725,-87.62397654354572 41.39745228909582)),((-87.62397654354572 41.39760369444365,-87.623982578516 41.39760369444365,-87.62398861348629 41.39760822150841,-87.62398861348629 41.39761224556572,-87.62398229974659 41.39761603450963,-87.62397074639594 41.3976138413312,-87.62397654354572 41.39760369444365)),((-87.62391149997711 41.397583071144595,-87.62391552329063 41.39758105911507,-87.62391820549965 41.397581562122454,-87.623921403287 41.3975844406643,-87.62392155826092 41.39758458016668,-87.62392289936543 41.39759212527668,-87.62391987732425 41.39759542265905,-87.62391876331321 41.39759500328741,-87.62391721237802 41.39759426404723,-87.6239157680276 41.39759341111883,-87.62391444490447 41.39759245314897,-87.62391325642203 41.39759139984933,-87.62391221462886 41.39759026189796,-87.62391133008636 41.39758905083112,-87.62391086664078 41.39758823022903,-87.62390965476884 41.39758755564437,-87.62390935879748 41.39758735429706,-87.62391149997711 41.397583071144595)))",
			},
			wantWKT: "MULTIPOLYGON(((-87.62395575642586 41.397455307146075,-87.62395441532135 41.39745128307905,-87.62395534881179 41.39745034941373,-87.62395626638761 41.39744972341221,-87.62395770979515 41.39744891829488,-87.62395925142329 41.397448224048624,-87.62396087630455 41.39744764741381,-87.62396256866319 41.39744719398892,-87.62396431206825 41.397446868176175,-87.62396608959328 41.39744667313887,-87.62396788398048 41.397446610770594,-87.62396964164054 41.39744668024725,-87.62397654354572 41.39745228909582,-87.6239725202322 41.39745731917949,-87.6239624619484 41.397458325196176,-87.62395575642586 41.397455307146075)),((-87.62401387572902 41.39745665745553,-87.6240215021472 41.39745842680161,-87.62402264165466 41.397455993937264,-87.62402329666267 41.39745478277123,-87.62402410175937 41.39745362403454,-87.62402504963794 41.397452528243555,-87.62402613169566 41.3974515053434,-87.62402733811206 41.39745056461764,-87.62402865793804 41.39744971460401,-87.62403007919521 41.39744896301703,-87.62403158898465 41.39744831667789,-87.62403317360392 41.397447781452556,-87.62403481867145 41.39744736219862,-87.62403650925704 41.39744706272111,-87.62403823001745 41.397446885738006,-87.62403996533548 41.39744683285552,-87.62404169946193 41.39744690455365,-87.62404341665828 41.39744710018164,-87.62404510133979 41.397447417964045,-87.62405168073352 41.39744891314965,-87.6240348815918 41.397450780070656,-87.62402415275574 41.39746084023782,-87.62401610612869 41.39746084023782,-87.62401387572902 41.39745665745553)),((-87.62420264993521 41.39749536435664,-87.62413840985234 41.397482951866046,-87.62403983775401 41.39746055108124,-87.62404158711433 41.39745681617114,-87.62420319020748 41.39749403877844,-87.6242026502297 41.397495364425644,-87.62420264993521 41.39749536435664)),((-87.62419891701943 41.39748048049657,-87.624207178155 41.39748207671596,-87.62420893681185 41.39748248885456,-87.62421063126413 41.3974830325109,-87.6242122442154 41.397483702135574,-87.62421375920124 41.39748449089327,-87.62421516075717 41.39748539073262,-87.6242164345766 41.39748639246838,-87.62421756765681 41.39748748587517,-87.62421854843173 41.397488659791875,-87.62421936688986 41.39748990223556,-87.62422001467672 41.397491200523774,-87.62422048517993 41.39749254140405,-87.62422077359668 41.39749391118913,-87.62422087698293 41.39749529589673,-87.62422082211445 41.39749621512772,-87.62420654296875 41.39748599064894,-87.62420654296875 41.397485487640814,-87.62420251965523 41.39748498463269,-87.62419891701943 41.39748048049657)),((-87.62423202395439 41.39751315308203,-87.62424007058144 41.397514159097845,-87.6242434233427 41.397517177145225,-87.6242434233427 41.39752069820032,-87.62423946264127 41.397522678919835,-87.62422593614133 41.39751938122142,-87.62422129511833 41.397514159097845,-87.6242246478796 41.39751063804239,-87.62422859770999 41.39751063804239,-87.62422934174538 41.39751063804239,-87.62423202395439 41.39751315308203)),((-87.62392771482985 41.39758304634879,-87.62391954660416 41.39758105911507,-87.62392119420248 41.39757693935211,-87.62392119442447 41.39757693941165,-87.6239291859734 41.397578755513464,-87.62392771482985 41.39758304634879)),((-87.62392155826092 41.39758458016668,-87.62392289936543 41.39759212527668,-87.62391987732425 41.39759542265905,-87.62391876331321 41.39759500328741,-87.62391721237802 41.39759426404723,-87.6239157680276 41.39759341111883,-87.62391444490447 41.39759245314897,-87.62391325642203 41.39759139984933,-87.62391221462886 41.39759026189796,-87.62391133008636 41.39758905083112,-87.62391086664078 41.39758823022903,-87.62390965476884 41.39758755564437,-87.62390935879748 41.39758735429706,-87.62391149997711 41.397583071144595,-87.62391552329063 41.39758105911507,-87.62391820549965 41.397581562122454,-87.623921403287 41.3975844406643,-87.62392155826092 41.39758458016668)),((-87.62402591965431 41.39761016753592,-87.62402482330799 41.39761325158001,-87.62398342211355 41.39760432725882,-87.62398861348629 41.39760822150841,-87.62398861348629 41.39761224556572,-87.62398229974659 41.39761603450963,-87.62397074639594 41.3976138413312,-87.62397654354572 41.39760369444365,-87.62398048639405 41.39760369444365,-87.62396648526192 41.39760067640029,-87.62396708180846 41.397598998302605,-87.62396708218431 41.39759899838907,-87.62402591965431 41.39761016753592)),((-87.62406505644321 41.3976167726299,-87.62407913804054 41.39761777864412,-87.62409187853336 41.39762934780651,-87.62410193681717 41.3976333718625,-87.62410461902618 41.397637395918245,-87.62410445659496 41.397639223606355,-87.62409893975861 41.397638176342504,-87.62409721290705 41.397637779093365,-87.62409554671505 41.39763725516227,-87.62409395752817 41.39763660968908,-87.62409246093647 41.39763584900594,-87.62409107162166 41.39763498057522,-87.62409054781364 41.397634580966205,-87.62403539291145 41.397621878478574,-87.62403515702614 41.397621994791656,-87.62403358694466 41.39762261406486,-87.62403194441826 41.3976231156037,-87.62403024497316 41.39762349466726,-87.62402850467363 41.39762374767239,-87.62402673997009 41.39762387222754,-87.6240249675436 41.39762386715533,-87.62402320414834 41.39762373250369,-87.62402146645303 41.39762346954546,-87.6240114235828 41.39762156310229,-87.62401409447193 41.39761777864412,-87.62403421103954 41.39761727563702,-87.62404292821884 41.39761274857286,-87.6240523159504 41.39761224556572,-87.62406505644321 41.3976167726299)),((-87.62415960431099 41.39764393500832,-87.62410126626492 41.39763085682753,-87.62410339298516 41.39762487432261,-87.62410339336095 41.397624874409054,-87.62421344649789 41.39764576585199,-87.62421257793903 41.39764795906342,-87.62420654296875 41.397646450042785,-87.62420654296875 41.39764695304968,-87.62416161596775 41.397637395918245,-87.62415960431099 41.39764393500832)),((-87.62421701337621 41.3976592051252,-87.62421590008874 41.39765935988056,-87.62421416603986 41.39765947381383,-87.62421242546463 41.39765946270053,-87.62421069423094 41.397659326641985,-87.62420898812152 41.39765906687855,-87.62420899357804 41.397659067915036,-87.62420604474384 41.39765850813634,-87.62420654296875 41.39765701318648,-87.62420654296875 41.39765600717286,-87.62421056628227 41.397653995145596,-87.62421324849129 41.39765449815243,-87.62421701337621 41.3976592051252)))",
		},

		// Reproduces https://github.com/peterstace/simplefeatures/issues/657.
		{
			inputWKTs: []string{
				"MULTIPOLYGON(((-110.957357 32.2328185,-110.957357 32.232822999999996,-110.95735599999999 32.232898,-110.9574775 32.232898,-110.95760999999999 32.232898,-110.957731 32.232898,-110.9577355 32.232898,-110.957866 32.2328985,-110.95786799999999 32.232751,-110.9578705 32.2325175,-110.957872 32.232405,-110.95787349999999 32.232271,-110.957741 32.23227,-110.957628 32.232268999999995,-110.957616 32.232268999999995,-110.95747399999999 32.232268,-110.957473 32.232279,-110.9574715 32.2322905,-110.95746899999999 32.2323015,-110.957466 32.232312,-110.95746199999999 32.232323,-110.95745749999999 32.232333,-110.95745199999999 32.2323435,-110.95744549999999 32.232352999999996,-110.9574385 32.2323625,-110.957431 32.232372,-110.95742249999999 32.2323805,-110.957414 32.2323885,-110.9574045 32.2323965,-110.95739449999999 32.2324035,-110.95738349999999 32.2324105,-110.95737249999999 32.2324165,-110.95736099999999 32.232422,-110.95736 32.2325155,-110.95736 32.2325265,-110.9573575 32.232748,-110.9573575 32.2327485,-110.9573575 32.2327495,-110.957357 32.2327945,-110.957357 32.2328185),(-110.9576385 32.232718,-110.9577245 32.2327185,-110.957724 32.232742,-110.95769449999999 32.232742,-110.95769449999999 32.2327505,-110.9576675 32.2327505,-110.95763799999999 32.232749999999996,-110.957504 32.232749,-110.95750299999999 32.2328245,-110.957471 32.232824,-110.957472 32.232749,-110.9574715 32.2327255,-110.957538 32.232726,-110.9575395 32.2325875,-110.957467 32.232586999999995,-110.95746749999999 32.232561,-110.957601 32.232562,-110.957599 32.2327175,-110.9576385 32.232718),(-110.9577025 32.2323885,-110.957737 32.232389,-110.957737 32.2324045,-110.9577365 32.232474499999995,-110.9577015 32.232474499999995,-110.9577015 32.2324585,-110.9577025 32.232389,-110.9577025 32.2323885)))",
				"MULTIPOLYGON(((-110.95736 32.2325265,-110.9573575 32.232748,-110.9573575 32.2327485,-110.9573575 32.2327495,-110.957357 32.2327945,-110.957357 32.2328185,-110.957357 32.232822999999996,-110.95735599999999 32.232898,-110.9574775 32.232898,-110.95760999999999 32.232898,-110.957731 32.232898,-110.9577355 32.232898,-110.957866 32.2328985,-110.95786799999999 32.232751,-110.95775549999999 32.232749999999996,-110.95775549999999 32.232787,-110.95773249999999 32.232787,-110.957735 32.232562,-110.95770499999999 32.232562,-110.95770499999999 32.232547499999995,-110.95767649999999 32.232547,-110.95767699999999 32.2325375,-110.957509 32.232537,-110.957484 32.232537,-110.9574835 32.232527499999996,-110.957422 32.232527,-110.9573855 32.2325265,-110.95736 32.2325265),(-110.9576385 32.232718,-110.9577245 32.2327185,-110.957724 32.232742,-110.95769449999999 32.232742,-110.95769449999999 32.2327505,-110.9576675 32.2327505,-110.95763799999999 32.232749999999996,-110.9576385 32.232718)))",
				"MULTIPOLYGON(((-110.95736 32.2325265,-110.9573575 32.232748,-110.9573575 32.2327485,-110.9573575 32.2327495,-110.957357 32.2327945,-110.957357 32.2328185,-110.957357 32.232822999999996,-110.95735599999999 32.232898,-110.9574775 32.232898,-110.95760999999999 32.232898,-110.957731 32.232898,-110.9577355 32.232898,-110.957866 32.2328985,-110.95786799999999 32.232751,-110.9578705 32.2325175,-110.957872 32.232405,-110.95787349999999 32.232271,-110.957741 32.23227,-110.957628 32.232268999999995,-110.957616 32.232268999999995,-110.95747399999999 32.232268,-110.957473 32.232279,-110.9574715 32.2322905,-110.95746899999999 32.2323015,-110.957466 32.232312,-110.95746199999999 32.232323,-110.95745749999999 32.232333,-110.95745199999999 32.2323435,-110.95744549999999 32.232352999999996,-110.9574385 32.2323625,-110.957431 32.232372,-110.95742249999999 32.2323805,-110.957414 32.2323885,-110.9574045 32.2323965,-110.95739449999999 32.2324035,-110.95738349999999 32.2324105,-110.95737249999999 32.2324165,-110.95736099999999 32.232422,-110.95736 32.2325155,-110.95736 32.2325265)))",
			},
			wantWKT: "POLYGON((-110.957357 32.2328185,-110.957357 32.2327945,-110.9573575 32.2327495,-110.9573575 32.2327485,-110.9573575 32.232748,-110.95736 32.2325265,-110.95736 32.2325155,-110.95736099999999 32.232422,-110.95737249999999 32.2324165,-110.95738349999999 32.2324105,-110.95739449999999 32.2324035,-110.9574045 32.2323965,-110.957414 32.2323885,-110.95742249999999 32.2323805,-110.957431 32.232372,-110.9574385 32.2323625,-110.95744549999999 32.232352999999996,-110.95745199999999 32.2323435,-110.95745749999999 32.232333,-110.95746199999999 32.232323,-110.957466 32.232312,-110.95746899999999 32.2323015,-110.9574715 32.2322905,-110.957473 32.232279,-110.95747399999999 32.232268,-110.957616 32.232268999999995,-110.957628 32.232268999999995,-110.957741 32.23227,-110.95787349999999 32.232271,-110.957872 32.232405,-110.9578705 32.2325175,-110.95786799999999 32.232751,-110.957866 32.2328985,-110.9577355 32.232898,-110.957731 32.232898,-110.95760999999999 32.232898,-110.9574775 32.232898,-110.95735599999999 32.232898,-110.957357 32.232822999999996,-110.957357 32.2328185))",
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
