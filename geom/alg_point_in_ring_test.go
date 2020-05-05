package geom

import (
	"fmt"
	"strconv"
	"testing"
)

func TestPointInRing(t *testing.T) {
	type subTestCase struct {
		pointWKT string
		side     side
	}
	type testCase struct {
		wkt      string
		subTests []subTestCase
	}
	for i, tc := range []testCase{
		{
			wkt: "POLYGON((1 0,2 0,1 3,1 0))",
			subTests: []subTestCase{
				{"POINT(1.25 1.4166666666666667)", interior},
			},
		},
		{
			wkt: "POLYGON((0 0,2 0,2 2,0 2,0 0))",
			subTests: []subTestCase{
				{"POINT(1 1)", interior},
				{"POINT(0 1)", boundary},
				{"POINT(2 1)", boundary},
				{"POINT(0 2)", boundary},
				{"POINT(2 2)", boundary},
				{"POINT(-1 2)", exterior},
				{"POINT(3 2)", exterior},
				{"POINT(-1 1)", exterior},
				{"POINT(3 1)", exterior},
			},
		},
		{
			wkt: "POLYGON((0 0,4 0,4 2,2 1,0 2,0 0))",
			subTests: []subTestCase{
				{"POINT(0 1)", boundary},
				{"POINT(1 1)", interior},
				{"POINT(3 1)", interior},
				{"POINT(4 1)", boundary},
				{"POINT(-1 1)", exterior},
				{"POINT(5 1)", exterior},
			},
		},
		{
			wkt: "POLYGON((0 0,2 1,4 0,4 2,0 2,0 0))",
			subTests: []subTestCase{
				{"POINT(0 1)", boundary},
				{"POINT(1 1)", interior},
				{"POINT(3 1)", interior},
				{"POINT(4 1)", boundary},
				{"POINT(-1 1)", exterior},
				{"POINT(5 1)", exterior},
			},
		},
		{
			wkt: "POLYGON((0 0,6 0,6 2,4 1,2 1,0 2,0 0))",
			subTests: []subTestCase{
				{"POINT(1 1)", interior},
				{"POINT(2 1)", boundary},
				{"POINT(3 1)", boundary},
				{"POINT(4 1)", boundary},
				{"POINT(5 1)", interior},
			},
		},
		{
			wkt: "POLYGON((0 0,6 0,4 1,2 1,0 2,0 0))",
			subTests: []subTestCase{
				{"POINT(1 1)", interior},
				{"POINT(2 1)", boundary},
				{"POINT(3 1)", boundary},
				{"POINT(4 1)", boundary},
				{"POINT(5 1)", exterior},
			},
		},
		{
			wkt: "POLYGON((0 0,2 1,4 1,6 0,6 2,0 2,0 0))",
			subTests: []subTestCase{
				{"POINT(1 1)", interior},
				{"POINT(2 1)", boundary},
				{"POINT(3 1)", boundary},
				{"POINT(4 1)", boundary},
				{"POINT(5 1)", interior},
			},
		},
		{
			wkt: "POLYGON((0 0,2 1,4 1,6 2,0 2,0 0))",
			subTests: []subTestCase{
				{"POINT(1 1)", interior},
				{"POINT(2 1)", boundary},
				{"POINT(3 1)", boundary},
				{"POINT(4 1)", boundary},
				{"POINT(5 1)", exterior},
			},
		},
		{
			wkt: "POLYGON((0 0,2 0,2 3,1 1,0 3,0 0))",
			subTests: []subTestCase{
				{"POINT(0 1)", boundary},
				{"POINT(0.5 1)", interior},
				{"POINT(1 1)", boundary},
				{"POINT(1.5 1)", interior},
				{"POINT(2 1)", boundary},
				{"POINT(0 1.5)", boundary},
				{"POINT(0.5 1.5)", interior},
				{"POINT(1 1.5)", exterior},
				{"POINT(1.5 1.5)", interior},
				{"POINT(2 1.5)", boundary},
				{"POINT(0 2)", boundary},
				{"POINT(0.5 2)", boundary},
				{"POINT(1 2)", exterior},
				{"POINT(1.5 2)", boundary},
				{"POINT(2 2)", boundary},
				{"POINT(0 2.5)", boundary},
				{"POINT(0.5 2.5)", exterior},
				{"POINT(1 2.5)", exterior},
				{"POINT(1.5 2.5)", exterior},
				{"POINT(2 2.5)", boundary},
				{"POINT(0 3)", boundary},
				{"POINT(0.5 3)", exterior},
				{"POINT(1 3)", exterior},
				{"POINT(1.5 3)", exterior},
				{"POINT(2 3)", boundary},
			},
		},
	} {
		g, err := UnmarshalWKT(tc.wkt)
		if err != nil {
			t.Fatal(err)
		}

		if !g.IsPolygon() {
			t.Fatal("expected a polygon")
		}
		poly := g.AsPolygon()
		ring := poly.ExteriorRing()

		for j, st := range tc.subTests {
			pt, err := UnmarshalWKT(st.pointWKT)
			if err != nil {
				t.Fatal(err)
			}
			xy, ok := pt.AsPoint().XY()
			if !ok {
				panic("point empty not expected in this test")
			}
			t.Run(fmt.Sprintf("non-indexed_%d_%d", i, j), func(t *testing.T) {
				if got := relatePointToRing(xy, ring); got != st.side {
					t.Log(tc.wkt)
					t.Log(st.pointWKT)
					t.Errorf("want=%v got=%v", st.side, got)
				}
			})
			t.Run(fmt.Sprintf("indexed_%d_%d", i, j), func(t *testing.T) {
				il := newIndexedLines(poly.Boundary().asLines())
				if got := relatePointToPolygon(xy, il); got != st.side {
					t.Log(tc.wkt)
					t.Log(st.pointWKT)
					t.Errorf("want=%v got=%v", st.side, got)
				}
			})
		}
	}
}

func TestPointInPolygon(t *testing.T) {
	g, err := UnmarshalWKT(`POLYGON(
		(0 0,1 0,1 1,2 1,2 0,3 0,3 -1,4 -1,4 0,5 0,5 1,6 1,6 -1,7 -1,7 0,8 0,8 -1,9 -1,10 0,11 -1,11 1,12 0,13 1,13 -1,14 -1,14 -2,30 -2,30 3,14 3,14 2,0 2,0 0),
		(15 0,16 0,16 1,17 1,17 0,18 0,18 -1,19 -1,19 0,20 0,20 1,21 1,21 -1,22 -1,22 0,23 0,23 -1,24 -1,25 0,26 -1,26 1,27 0,28 1,28 -1,29 -1,29 2,15 2,15 0)
	)`)
	if err != nil {
		t.Fatal(err)
	}
	poly := g.AsPolygon()
	il := newIndexedLines(poly.Boundary().asLines())

	for i, tt := range []struct {
		x    float64
		side side
	}{
		{-1, exterior},
		{0.0, boundary},
		{0.5, boundary},
		{1.0, boundary},
		{1.5, exterior},
		{2.0, boundary},
		{2.5, boundary},
		{3.0, boundary},
		{3.5, interior},
		{4.0, boundary},
		{4.5, boundary},
		{5.0, boundary},
		{5.5, exterior},
		{6.0, boundary},
		{6.5, interior},
		{7.0, boundary},
		{7.5, boundary},
		{8.0, boundary},
		{8.5, interior},
		{9.0, interior},
		{9.5, interior},
		{10.0, boundary},
		{10.5, interior},
		{11.0, boundary},
		{11.5, exterior},
		{12.0, boundary},
		{12.5, exterior},
		{13.0, boundary},
		{13.5, interior},
		{14.0, interior},
		{14.5, interior},
		{15.0, boundary},
		{15.5, boundary},
		{16.0, boundary},
		{16.5, interior},
		{17.0, boundary},
		{17.5, boundary},
		{18.0, boundary},
		{18.5, exterior},
		{19.0, boundary},
		{19.5, boundary},
		{20.0, boundary},
		{20.5, interior},
		{21.0, boundary},
		{21.5, exterior},
		{22.0, boundary},
		{22.5, boundary},
		{23.0, boundary},
		{23.5, exterior},
		{24.0, exterior},
		{24.5, exterior},
		{25.0, boundary},
		{25.5, exterior},
		{26.0, boundary},
		{26.5, interior},
		{27.0, boundary},
		{27.5, interior},
		{28.0, boundary},
		{28.5, exterior},
		{29.0, boundary},
		{29.5, interior},
		{30.0, boundary},
		{30.5, exterior},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			pt := XY{X: tt.x, Y: 0}
			got := relatePointToPolygon(pt, il)
			if got != tt.side {
				t.Log("point", pt)
				t.Log("want", tt.side)
				t.Log("got", got)
				t.Error("mismatch")
			}
		})
	}
}
