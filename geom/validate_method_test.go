package geom_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestValidate(t *testing.T) {
	t.Run("invalid_point", func(t *testing.T) {
		pt, err := geom.XY{X: math.NaN(), Y: math.NaN()}.AsPoint(geom.DisableAllValidations)
		expectNoErr(t, err)
		expectErr(t, pt.Validate())
	})

	for i, tc := range []struct {
		wantValid bool
		wkt       string
	}{
		{true, "POINT(1 2)"},

		{true, "LINESTRING(1 2,3 4)"},
		{false, "LINESTRING(1 2,1 2)"},
		{false, "LINESTRING(1 2)"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.wkt)
			if tc.wantValid {
				g, err := geom.UnmarshalWKT(tc.wkt)
				expectNoErr(t, err)
				expectNoErr(t, g.Validate())
			} else {
				g, err := geom.UnmarshalWKT(tc.wkt, geom.DisableAllValidations)
				expectNoErr(t, err)
				expectErr(t, g.Validate())
			}
		})
	}
}
