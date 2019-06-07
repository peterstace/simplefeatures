package simplefeatures_test

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func TestUnmarshalWKTValid(t *testing.T) {
	must := func(g Geometry, err error) Geometry {
		if err != nil {
			t.Fatalf("could not create geometry: %v", err)
		}
		return g
	}
	for _, tt := range []struct {
		name string
		wkt  string
		want Geometry
	}{
		{
			name: "basic point (wikipedia)",
			wkt:  "POINT (30 10)",
			want: NewPoint(30, 10),
		},
		{
			name: "empty point",
			wkt:  "POINT EMPTY",
			want: NewEmptyPoint(),
		},
		{
			name: "basic line string (wikipedia)",
			wkt:  "LINESTRING (30 10, 10 30, 40 40)",
			want: must(NewLineString([]Point{
				NewPoint(30, 10),
				NewPoint(10, 30),
				NewPoint(40, 40),
			})),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalWKT(strings.NewReader(tt.wkt))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("want=%#v got=%#v", got, tt.want)
			}
		})
	}
}
