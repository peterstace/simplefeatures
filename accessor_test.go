package simplefeatures_test

import (
	"reflect"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func TestPointAccessor(t *testing.T) {
	pt := geomFromWKT(t, "POINT(1 2)").(Point)
	want := XY{NewScalarFromFloat64(1), NewScalarFromFloat64(2)}
	got := pt.XY()
	if !want.Equals(got) {
		t.Errorf("got=%v want=%v", got, want)
	}
}

func TestLineAccessor(t *testing.T) {
	line := geomFromWKT(t, "LINESTRING(1 2,3 4)").(Line)
	t.Run("start", func(t *testing.T) {
		got := line.StartPoint()
		want := geomFromWKT(t, "POINT(1 2)")
		if !reflect.DeepEqual(got, want) {
			t.Errorf("want=%v got=%v", want, got)
		}
	})
	t.Run("end", func(t *testing.T) {
		got := line.EndPoint()
		want := geomFromWKT(t, "POINT(3 4)")
		if !reflect.DeepEqual(got, want) {
			t.Errorf("want=%v got=%v", want, got)
		}
	})
	t.Run("num points", func(t *testing.T) {
		if line.NumPoints() != 2 {
			t.Errorf("wanted 2")
		}
	})
	t.Run("point 0", func(t *testing.T) {
		got := line.PointN(0)
		want := geomFromWKT(t, "POINT(1 2)")
		expectDeepEqual(t, got, want)
	})
	t.Run("point 1", func(t *testing.T) {
		got := line.PointN(1)
		want := geomFromWKT(t, "POINT(3 4)")
		expectDeepEqual(t, got, want)
	})
	t.Run("point 2", func(t *testing.T) {
		expectPanics(t, func() {
			line.PointN(2)
		})
	})
	t.Run("point -1", func(t *testing.T) {
		expectPanics(t, func() {
			line.PointN(-1)
		})
	})
}

func TestLineStringAccessor(t *testing.T) {
	ls := geomFromWKT(t, "LINESTRING(1 2,3 4,5 6)").(LineString)
	pt12 := geomFromWKT(t, "POINT(1 2)")
	pt34 := geomFromWKT(t, "POINT(3 4)")
	pt56 := geomFromWKT(t, "POINT(5 6)")

	t.Run("start", func(t *testing.T) {
		expectDeepEqual(t, ls.StartPoint(), pt12)
	})
	t.Run("end", func(t *testing.T) {
		expectDeepEqual(t, ls.EndPoint(), pt56)
	})
	t.Run("num points", func(t *testing.T) {
		expectDeepEqual(t, ls.NumPoints(), 3)
	})
	t.Run("point n", func(t *testing.T) {
		expectPanics(t, func() { ls.PointN(-1) })
		expectDeepEqual(t, ls.PointN(0), pt12)
		expectDeepEqual(t, ls.PointN(1), pt34)
		expectDeepEqual(t, ls.PointN(2), pt56)
		expectPanics(t, func() { ls.PointN(3) })
	})
}

func TestLinearRingAccessor(t *testing.T) {
	ls := geomFromWKT(t, "LINEARRING(0 0,1 0,0 1,0 0)").(LinearRing)
	pt0 := geomFromWKT(t, "POINT(0 0)")
	pt1 := geomFromWKT(t, "POINT(1 0)")
	pt2 := geomFromWKT(t, "POINT(0 1)")
	pt3 := geomFromWKT(t, "POINT(0 0)")

	t.Run("start", func(t *testing.T) {
		expectDeepEqual(t, ls.StartPoint(), pt0)
	})
	t.Run("end", func(t *testing.T) {
		expectDeepEqual(t, ls.EndPoint(), pt3)
	})
	t.Run("num points", func(t *testing.T) {
		expectDeepEqual(t, ls.NumPoints(), 4)
	})
	t.Run("point n", func(t *testing.T) {
		expectPanics(t, func() { ls.PointN(-1) })
		expectDeepEqual(t, ls.PointN(0), pt0)
		expectDeepEqual(t, ls.PointN(1), pt1)
		expectDeepEqual(t, ls.PointN(2), pt2)
		expectDeepEqual(t, ls.PointN(3), pt3)
		expectPanics(t, func() { ls.PointN(4) })
	})
}
