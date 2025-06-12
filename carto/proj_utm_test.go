package carto_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/carto"
	"github.com/peterstace/simplefeatures/geom"
)

func TestUTMFromCode(t *testing.T) {
	for zone := 1; zone <= 60; zone++ {
		for _, designator := range []string{"N", "S"} {
			code := fmt.Sprintf("%02d%s", zone, designator)
			t.Run(code, func(t *testing.T) {
				proj, err := carto.NewUTMFromCode(code)
				if err != nil {
					t.Fatal(err)
				}
				if proj.Code() != code {
					t.Errorf("got code %s, want %s", proj.Code(), code)
				}
			})
		}
	}
}

func TestUTMFromLocation(t *testing.T) {
	for _, tc := range []struct {
		name     string
		ptWKT    string
		wantCode string
	}{
		{
			"in the southern hemisphere",
			"POINT(151.2020581 -33.8557148)", "56S",
		},
		{
			"zone 01N",
			"POINT(-178.03 51.865)", "01N",
		},
		{
			"zone 02N",
			"POINT(-171.694 63.767)", "02N",
		},
		{
			"zone 59N",
			"POINT(170.249 60.071)", "59N",
		},
		{
			"zone 60N",
			"POINT(177.142 62.652)", "60N",
		},
		{
			"in the east of a zone",
			"POINT(17.9999 10)", "33N",
		},
		{
			"in the west of a zone",
			"POINT(18.0001 10)", "34N",
		},
		{
			"on the boundary between two zones",
			"POINT(18 10)", "34N", // By arbitrary convention, the zone to the east is chosen.
		},
		{
			"at 180 degrees east",
			"POINT(180 10)", "60N",
		},
		{
			"at 180 degrees west",
			"POINT(-180 10)", "01N",
		},
		{
			"at 0 degrees east",
			"POINT(0 10)", "31N",
		},
		{
			"on the equator",
			"POINT(1 0)", "31N", // By arbitrary convention, the zone to the north is chosen.
		},
		{
			"at the northern limit",
			"POINT(1 84)", "31N",
		},
		{
			"at the southern limit",
			"POINT(1 -80)", "31S",
		},

		// Norway exception testing:
		{"Norway exception  1", "POINT(2.9 55.9)", "31N"},
		{"Norway exception  2", "POINT(3.1 55.9)", "31N"},
		{"Norway exception  3", "POINT(2.9 64.1)", "31N"},
		{"Norway exception  4", "POINT(3.1 64.1)", "31N"},
		{"Norway exception  5", "POINT(2.9 56.1)", "31N"},
		{"Norway exception  6", "POINT(3.1 56.1)", "32N"},
		{"Norway exception  7", "POINT(2.9 63.9)", "31N"},
		{"Norway exception  8", "POINT(3.1 63.9)", "32N"},
		{"Norway exception  9", "POINT(11.9 55.9)", "32N"},
		{"Norway exception 10", "POINT(12.1 55.9)", "33N"},
		{"Norway exception 11", "POINT(11.9 56.1)", "32N"},
		{"Norway exception 12", "POINT(12.1 56.1)", "33N"},
		{"Norway exception 13", "POINT(11.9 63.9)", "32N"},
		{"Norway exception 14", "POINT(12.1 63.9)", "33N"},
		{"Norway exception 15", "POINT(11.9 64.1)", "32N"},
		{"Norway exception 16", "POINT(12.1 64.1)", "33N"},

		// Svalbard exception testing:
		{"Svalbard exception  1", "POINT(-0.1 71.9)", "30N"},
		{"Svalbard exception  2", "POINT( 0.1 71.9)", "31N"},
		{"Svalbard exception  3", "POINT(-0.1 72.1)", "30N"},
		{"Svalbard exception  4", "POINT( 0.1 72.1)", "31N"},
		{"Svalbard exception  5", "POINT( 8.9 71.9)", "32N"},
		{"Svalbard exception  6", "POINT( 9.1 71.9)", "32N"},
		{"Svalbard exception  7", "POINT( 8.9 72.1)", "31N"},
		{"Svalbard exception  8", "POINT( 9.1 72.1)", "33N"},
		{"Svalbard exception  9", "POINT(20.9 71.9)", "34N"},
		{"Svalbard exception 10", "POINT(21.1 71.9)", "34N"},
		{"Svalbard exception 11", "POINT(20.9 72.1)", "33N"},
		{"Svalbard exception 12", "POINT(21.1 72.1)", "35N"},
		{"Svalbard exception 13", "POINT(32.9 71.9)", "36N"},
		{"Svalbard exception 14", "POINT(33.1 71.9)", "36N"},
		{"Svalbard exception 15", "POINT(32.9 72.1)", "35N"},
		{"Svalbard exception 16", "POINT(33.1 72.1)", "37N"},
		{"Svalbard exception 17", "POINT(-0.1 83.9)", "30N"},
		{"Svalbard exception 18", "POINT( 0.1 83.9)", "31N"},
		{"Svalbard exception 19", "POINT( 8.9 83.9)", "31N"},
		{"Svalbard exception 20", "POINT( 9.1 83.9)", "33N"},
		{"Svalbard exception 21", "POINT(20.9 83.9)", "33N"},
		{"Svalbard exception 22", "POINT(21.1 83.9)", "35N"},
		{"Svalbard exception 23", "POINT(32.9 83.9)", "35N"},
		{"Svalbard exception 24", "POINT(33.1 83.9)", "37N"},
		{"Svalbard exception 25", "POINT(41.9 83.9)", "37N"},
		{"Svalbard exception 26", "POINT(42.1 83.9)", "38N"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pt, err := geom.UnmarshalWKT(tc.ptWKT)
			if err != nil {
				t.Fatal(err)
			}
			loc, ok := pt.MustAsPoint().XY()
			if !ok {
				t.Fatalf("empty point")
			}
			proj, err := carto.NewUTMFromLocation(loc)
			if err != nil {
				t.Fatalf("got err %v, want nil", err)
			}
			if proj.Code() != tc.wantCode {
				t.Errorf("got code %s, want %s", proj.Code(), tc.wantCode)
			}
		})
	}
}

func TestUTMFromLocationErrorCases(t *testing.T) {
	for _, tc := range []struct {
		ptWKT string
		name  string
	}{
		{"POINT(-180.1 0)", "longitude less than -180"},
		{"POINT(180.1 0)", "longitude greater than 180"},
		{"POINT(0 84.1)", "latitude greater than 84"},
		{"POINT(0 -80.1)", "latitude less than -80"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pt, err := geom.UnmarshalWKT(tc.ptWKT)
			if err != nil {
				t.Fatal(err)
			}
			loc, ok := pt.MustAsPoint().XY()
			if !ok {
				t.Fatalf("expected a point")
			}
			if _, err := carto.NewUTMFromLocation(loc); err == nil {
				t.Errorf("expected an error for %s", tc.ptWKT)
			}
		})
	}
}

func TestUTMForwardReverse(t *testing.T) {
	for i, tc := range []struct {
		code     string
		inputWKT string
		wantWKT  string
	}{
		// echo '151.2020581 -33.8557148' | cs2cs +to EPSG:32756 -d 3
		// 333673.327      6252387.751 0.000
		{"56S", "POINT(151.2020581 -33.8557148)", "POINT(333673.327 6252387.751)"},

		// echo '14.5186965 35.9019739' | cs2cs +to EPSG:32633 -d 3
		// 456567.479      3973182.990 0.000
		{"33N", "POINT(14.5186965 35.9019739)", "POINT(456567.479 3973182.990)"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			pre, err := geom.UnmarshalWKT(tc.inputWKT)
			if err != nil {
				t.Error(err)
			}
			want, err := geom.UnmarshalWKT(tc.wantWKT)
			if err != nil {
				t.Error(err)
			}

			proj, err := carto.NewUTMFromCode(tc.code)
			if err != nil {
				t.Error(err)
			}

			const toleranceInMeters = 1e-3 // 1mm
			got := pre.TransformXY(proj.Forward)
			if !geom.ExactEquals(got, want, geom.ToleranceXY(toleranceInMeters)) {
				t.Errorf("got %v, want %v", got.AsText(), want.AsText())
			}

			const toleranceInDegrees = 1e-9 // ~0.1mm
			roundTrip := got.TransformXY(proj.Reverse)
			if !geom.ExactEquals(roundTrip, pre, geom.ToleranceXY(toleranceInDegrees)) {
				t.Errorf("round trip got %v, want %v", roundTrip.AsText(), pre.AsText())
			}
		})
	}
}
