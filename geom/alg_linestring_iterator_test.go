package geom

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestLineStringIterator(t *testing.T) {
	convertWKTToLineString := func(t *testing.T, wkt string) LineString {
		g, err := UnmarshalWKT(strings.NewReader(wkt))
		if err != nil {
			t.Fatal(err)
		}
		if g.IsLineString() {
			return g.AsLineString()
		}
		if g.IsLine() {
			return g.AsLine().AsLineString()
		}
		t.Fatal("not a line or linestring")
		return LineString{}
	}
	convertWKTToLine := func(t *testing.T, wkt string) Line {
		g, err := UnmarshalWKT(strings.NewReader(wkt))
		if err != nil {
			t.Fatal(err)
		}
		if g.IsLine() {
			return g.AsLine()
		}
		t.Fatal("not a line")
		return Line{}
	}
	for i, tt := range []struct {
		wkt   string
		lines []string
	}{
		{
			"LINESTRING EMPTY",
			nil,
		},
		{
			"LINESTRING(0 1,2 3)",
			[]string{
				"LINESTRING(0 1,2 3)",
			},
		},
		{
			"LINESTRING(0 1,2 3,4 5)",
			[]string{
				"LINESTRING(0 1,2 3)",
				"LINESTRING(2 3,4 5)",
			},
		},
		{
			"LINESTRING(0 1,2 3,4 5,6 7)",
			[]string{
				"LINESTRING(0 1,2 3)",
				"LINESTRING(2 3,4 5)",
				"LINESTRING(4 5,6 7)",
			},
		},
		{
			"LINESTRING(0 1,0 1,2 3,4 5)",
			[]string{
				"LINESTRING(0 1,2 3)",
				"LINESTRING(2 3,4 5)",
			},
		},
		{
			"LINESTRING(0 1,2 3,2 3,4 5)",
			[]string{
				"LINESTRING(0 1,2 3)",
				"LINESTRING(2 3,4 5)",
			},
		},
		{
			"LINESTRING(0 1,2 3,4 5,4 5)",
			[]string{
				"LINESTRING(0 1,2 3)",
				"LINESTRING(2 3,4 5)",
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ls := convertWKTToLineString(t, tt.wkt)
			var got []Line
			iter := newLineStringIterator(ls)
			for iter.next() {
				got = append(got, iter.line())
			}

			var want []Line
			for _, wkt := range tt.lines {
				want = append(want, convertWKTToLine(t, wkt))
			}

			if !reflect.DeepEqual(got, want) {
				t.Logf("got:  %v", got)
				t.Logf("want: %v", want)
				t.Error("mismatch")
			}
		})
	}
}
