package proj_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
	"github.com/peterstace/simplefeatures/proj"
)

func TestBadSourceCRS(t *testing.T) {
	_, err := proj.NewTransformation("EPSG:999999", "EPSG:4326")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBadTargetCRS(t *testing.T) {
	_, err := proj.NewTransformation("EPSG:4326", "EPSG:999999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGeneric(t *testing.T) {
	for i, tc := range []struct {
		sourceCRS   string
		targetCRS   string
		input       string
		want        string
		toleranceXY float64
		fails       func() bool
	}{
		{
			// echo "55 12" | cs2cs +init=epsg:4326 +to "+proj=utm +zone=32 +datum=WGS84"
			sourceCRS:   "EPSG:4326",
			targetCRS:   "+proj=utm +zone=32 +datum=WGS84",
			input:       "POINT(55 12)",
			want:        "POINT(6080642.11 1886936.97)",
			toleranceXY: 1e-2, // cs2cs rounds to 2 decimal places
		},
		{
			// echo "55 12" | cs2cs +init=epsg:4326 +to "+proj=utm +zone=32 +datum=WGS84"
			// echo "55.1 12.1" | cs2cs +init=epsg:4326 +to "+proj=utm +zone=32 +datum=WGS84"
			sourceCRS:   "EPSG:4326",
			targetCRS:   "+proj=utm +zone=32 +datum=WGS84",
			input:       "LINESTRING(55 12, 55.1 12.1)",
			want:        "LINESTRING(6080642.11 1886936.97, 6092335.13 1905469.04)",
			toleranceXY: 1e-2, // cs2cs rounds to 2 decimal places
		},
		{
			// Z value always unmodified (for this transformation).
			sourceCRS:   "EPSG:4326",
			targetCRS:   "EPSG:32756", // UTM zone 56S
			input:       "POINT Z(151.2093 -33.8688 42)",
			want:        "POINT Z(334368.633648097 6250948.345385009 42)",
			toleranceXY: 0, // X and Y want values are exact.
		},
		{
			// M value always passed through untouched.
			sourceCRS:   "EPSG:4326",
			targetCRS:   "EPSG:32756", // UTM zone 56S
			input:       "POINT M(151.2093 -33.8688 42)",
			want:        "POINT M(334368.633648097 6250948.345385009 42)",
			toleranceXY: 0, // X and Y want values are exact.
		},
		{
			// Z and M values.
			sourceCRS:   "EPSG:4326",
			targetCRS:   "EPSG:32756", // UTM zone 56S
			input:       "POINT ZM(151.2093 -33.8688 100 42)",
			want:        "POINT ZM(334368.633648097 6250948.345385009 100 42)",
			toleranceXY: 0, // X and Y want values are exact.
		},
		{
			// See https://www.spatial.nsw.gov.au/__data/assets/pdf_file/0013/230431/3D_Data_and_Transformations_Information_Sheet.pdf
			sourceCRS:   "EPSG:4939", // GDA94 (3D) (defined with ellipsoid height)
			targetCRS:   "EPSG:7843", // GDA2020 (3D) (defined with ellipsoid height)
			input:       "POINT Z(151.209300000000 -33.86880000000000 10.123000000000000)",
			want:        "POINT Z(151.209305429808 -33.86878731685864 10.028012086637318)",
			toleranceXY: 1e-10, // Slightly different numbers on each version of PROJ.
			fails: func() bool {
				// PROJ 9.2.1-r0 that comes with Alpine Linux 3.18 doesn't have
				// the necessary grid files for this transformation.
				major, minor, _ := proj.Version()
				return major == 9 && minor == 2
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			pj, err := proj.NewTransformation(tc.sourceCRS, tc.targetCRS)
			test.NoErr(t, err)
			defer pj.Release()

			input, err := geom.UnmarshalWKT(tc.input)
			test.NoErr(t, err)
			want, err := geom.UnmarshalWKT(tc.want)
			test.NoErr(t, err)

			transformed, err := input.Transform(pj.Forward)
			if tc.fails != nil && tc.fails() {
				test.Err(t, err)
				return
			}
			test.NoErr(t, err)
			if !geom.ExactEquals(transformed, want, geom.ToleranceXY(tc.toleranceXY)) {
				t.Fatalf("got %v, want %v", transformed.AsText(), want.AsText())
			}

			roundTrip, err := transformed.Transform(pj.Inverse)
			test.NoErr(t, err)

			const (
				// Round trip tolerances are limited by float64 rounding error.
				// We basically want these to be as small as possible without
				// false positive test failures.
				roundTripXYTolerance = 1e-11
				roundTripZTolerance  = 1e-6
			)
			if !geom.ExactEquals(
				roundTrip, input,
				geom.ToleranceXY(roundTripXYTolerance),
				geom.ToleranceZ(roundTripZTolerance),
			) {
				t.Fatalf("round trip failed: got %v, want %v", roundTrip.AsText(), input.AsText())
			}
		})
	}
}
