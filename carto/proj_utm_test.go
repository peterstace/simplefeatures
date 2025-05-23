package carto_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/carto"
	"github.com/peterstace/simplefeatures/geom"
)

func TestUTM(t *testing.T) {
	for i, tc := range []struct {
		zone       int
		northSouth string
		inputWKT   string
		wantWKT    string
	}{
		// echo '151.2020581 -33.8557148' | cs2cs +to EPSG:32756 -d 3
		// 333673.327      6252387.751 0.000
		{56, "S", "POINT(151.2020581 -33.8557148)", "POINT(333673.327 6252387.751)"},

		// echo '14.5186965 35.9019739' | cs2cs +to EPSG:32633 -d 3
		// 456567.479      3973182.990 0.000
		{33, "N", "POINT(14.5186965 35.9019739)", "POINT(456567.479 3973182.990)"},
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

			proj, err := carto.NewUTMFromZoneAndHemisphere(tc.zone, tc.northSouth)
			if err != nil {
				t.Error(err)
			}

			const toleranceInMeters = 1e-3 // 1mm
			got := pre.TransformXY(proj.Forward)
			if !geom.ExactEquals(got, want, geom.ToleranceXY(toleranceInMeters)) {
				t.Errorf("got %v, want %v", got.AsText(), want.AsText())
			}

			const toleranceInDegrees = 1e-9 // ~0.1mm
			roundTrip := got.TransformXY(proj.Recurse)
			if !geom.ExactEquals(roundTrip, pre, geom.ToleranceXY(toleranceInDegrees)) {
				t.Errorf("round trip got %v, want %v", roundTrip.AsText(), pre.AsText())
			}
		})
	}
}
