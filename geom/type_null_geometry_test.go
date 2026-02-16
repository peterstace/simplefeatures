package geom_test

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
)

func TestNullGeometryScan(t *testing.T) {
	wkb := test.FromWKT(t, "POINT(1 2)").AsBinary()

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
			test.NoErr(t, err)
			test.Eq(t, tc.wantValid, ng.Valid)
			if tc.wantValid {
				test.ExactEquals(t, ng.Geometry, test.FromWKT(t, tc.wantWKT))
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
			input:       geom.NullGeometry{Valid: true, Geometry: test.FromWKT(t, "POINT(1 2)")},
			want:        test.FromWKT(t, "POINT(1 2)").AsBinary(),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			valuer := driver.Valuer(tc.input)
			got, err := valuer.Value()
			test.NoErr(t, err)
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
			test.DeepEqual(t, gotBytes, tc.want)
		})
	}
}
