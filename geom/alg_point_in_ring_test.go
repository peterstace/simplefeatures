package simplefeatures

import (
	"fmt"
	"strings"
	"testing"
)

func TestPointInRing(t *testing.T) {
	type subTestCase struct {
		pointWKT string
		inside   bool
	}
	type testCase struct {
		ringWKT  string
		subTests []subTestCase
	}
	for i, tc := range []testCase{
		{
			ringWKT: "LINEARRING(0 0,2 0,2 2,0 2,0 0)",
			subTests: []subTestCase{
				{"POINT(1 1)", true},
				{"POINT(0 1)", true},
				{"POINT(2 1)", true},
				{"POINT(0 2)", true},
				{"POINT(2 2)", true},
				{"POINT(-1 2)", false},
				{"POINT(3 2)", false},
				{"POINT(-1 1)", false},
				{"POINT(3 1)", false},
			},
		},
		{
			ringWKT: "LINEARRING(0 0,4 0,4 2,2 1,0 2,0 0)",
			subTests: []subTestCase{
				{"POINT(0 1)", true},
				{"POINT(1 1)", true},
				{"POINT(3 1)", true},
				{"POINT(4 1)", true},
				{"POINT(-1 1)", false},
				{"POINT(5 1)", false},
			},
		},
		{
			ringWKT: "LINEARRING(0 0,2 1,4 0,4 2,0 2,0 0)",
			subTests: []subTestCase{
				{"POINT(0 1)", true},
				{"POINT(1 1)", true},
				{"POINT(3 1)", true},
				{"POINT(4 1)", true},
				{"POINT(-1 1)", false},
				{"POINT(5 1)", false},
			},
		},
		{
			ringWKT: "LINEARRING(0 0,6 0,6 2,4 1,2 1,0 2,0 0)",
			subTests: []subTestCase{
				{"POINT(1 1)", true},
				{"POINT(2 1)", true},
				{"POINT(3 1)", true},
				{"POINT(4 1)", true},
				{"POINT(5 1)", true},
			},
		},
		{
			ringWKT: "LINEARRING(0 0,6 0,4 1,2 1,0 2,0 0)",
			subTests: []subTestCase{
				{"POINT(1 1)", true},
				{"POINT(2 1)", true},
				{"POINT(3 1)", true},
				{"POINT(4 1)", true},
				{"POINT(5 1)", false},
			},
		},
		{
			ringWKT: "LINEARRING(0 0,2 1,4 1,6 0,6 2,0 2,0 0)",
			subTests: []subTestCase{
				{"POINT(1 1)", true},
				{"POINT(2 1)", true},
				{"POINT(3 1)", true},
				{"POINT(4 1)", true},
				{"POINT(5 1)", true},
			},
		},
		{
			ringWKT: "LINEARRING(0 0,2 1,4 1,6 2,0 2,0 0)",
			subTests: []subTestCase{
				{"POINT(1 1)", true},
				{"POINT(2 1)", true},
				{"POINT(3 1)", true},
				{"POINT(4 1)", true},
				{"POINT(5 1)", false},
			},
		},
		{
			ringWKT: "LINEARRING(0 0,2 0,2 3,1 1,0 3,0 0)",
			subTests: []subTestCase{
				{"POINT(0 1)", true},
				{"POINT(0.5 1)", true},
				{"POINT(1 1)", true},
				{"POINT(1.5 1)", true},
				{"POINT(2 1)", true},
				{"POINT(0 1.5)", true},
				{"POINT(0.5 1.5)", true},
				{"POINT(1 1.5)", false},
				{"POINT(1.5 1.5)", true},
				{"POINT(2 1.5)", true},
				{"POINT(0 2)", true},
				{"POINT(0.5 2)", true},
				{"POINT(1 2)", false},
				{"POINT(1.5 2)", true},
				{"POINT(2 2)", true},
				{"POINT(0 2.5)", true},
				{"POINT(0.5 2.5)", false},
				{"POINT(1 2.5)", false},
				{"POINT(1.5 2.5)", false},
				{"POINT(2 2.5)", true},
				{"POINT(0 3)", true},
				{"POINT(0.5 3)", false},
				{"POINT(1 3)", false},
				{"POINT(1.5 3)", false},
				{"POINT(2 3)", true},
			},
		},
	} {
		ringGeom, err := UnmarshalWKT(strings.NewReader(tc.ringWKT))
		if err != nil {
			t.Fatal(err)
		}
		ring := ringGeom.(LinearRing)
		for j, st := range tc.subTests {
			t.Run(fmt.Sprintf("%d_%d", i, j), func(t *testing.T) {
				pointGeom, err := UnmarshalWKT(strings.NewReader(st.pointWKT))
				if err != nil {
					t.Fatal(err)
				}
				point := pointGeom.(Point)
				got := pointRingSide(point.coords.XY, ring) != exterior
				t.Log(tc.ringWKT)
				t.Log(st.pointWKT)
				if got != st.inside {
					t.Errorf("want=%v got=%v", st.inside, got)
				}
			})
		}
	}
}
