package geom_test

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestNullGeometryScan(t *testing.T) {
	wkb := geomFromWKT(t, "POINT(1 2)").AsBinary()

	for _, tc := range []struct {
		description string
		value       interface{}
		wantValid   bool
		wantWKT     string
	}{
		{
			description: "NULL geometry",
			value:       nil,
			wantValid:   false,
		},
		{
			description: "populated geometry with string",
			value:       string(wkb),
			wantValid:   true,
			wantWKT:     "POINT(1 2)",
		},
		{
			description: "populated geometry with []byte",
			value:       wkb,
			wantValid:   true,
			wantWKT:     "POINT(1 2)",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var ng geom.NullGeometry
			scn := sql.Scanner(&ng)
			err := scn.Scan(tc.value)
			expectNoErr(t, err)
			expectBoolEq(t, tc.wantValid, ng.Valid)
			if tc.wantValid {
				expectGeomEq(t, ng.Geometry, geomFromWKT(t, tc.wantWKT))
			}
		})
	}
}

func TestNullGeometryValue(t *testing.T) {
	for _, tc := range []struct {
		description string
		input       geom.NullGeometry
		want        []byte
	}{
		{
			description: "NULL geometry",
			input:       geom.NullGeometry{Valid: false},
			want:        nil,
		},
		{
			description: "point geometry",
			input:       geom.NullGeometry{Valid: true, Geometry: geomFromWKT(t, "POINT(1 2)")},
			want:        geomFromWKT(t, "POINT(1 2)").AsBinary(),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			valuer := driver.Valuer(tc.input)
			got, err := valuer.Value()
			expectNoErr(t, err)
			if got == nil {
				if tc.want != nil {
					t.Fatalf("got nil but didn't want nil")
				}
				return
			}
			gotBytes, ok := got.([]byte)
			if !ok {
				t.Fatalf("didn't get bytes")
			}
			expectBytesEq(t, gotBytes, tc.want)
		})
	}
}
