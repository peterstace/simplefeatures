package wrapgeos_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/check"
	"github.com/peterstace/simplefeatures/internal/wrapgeos"
)

func TestVersion(t *testing.T) {
	x, y, z := wrapgeos.Version()
	t.Logf("version: v%d.%d.%d", x, y, z)
	check.Eq(t, x, 3)
	check.GE(t, y, 0)
	check.GE(t, z, 0)
}

func TestUnion(t *testing.T) {
	g1 := check.GeomFromWKT(t, "POLYGON((0 0,0 2,2 2,2 0,0 0))")
	g2 := check.GeomFromWKT(t, "POLYGON((1 1,1 3,3 3,3 1,1 1))")

	got, err := wrapgeos.Union(g1, g2)
	check.NoError(t, err)

	want := check.GeomFromWKT(t, "POLYGON((0 0,0 2,1 2,1 3,3 3,3 1,2 1,2 0,0 0))")
	check.GeomEq(t, got, want, geom.IgnoreOrder)
}
