// Package perf contains performance benchmarks that don't make sense to
// include in any particular package (they may test code from multiple
// packages).
package perf

import (
	"fmt"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/geos"
)

func BenchmarkSetOperation(b *testing.B) {
	for i := 2; i <= 14; i++ {
		sz := 1 << i
		p1 := regularPolygon(geom.XY{X: 0, Y: 0}, 1.0, sz).AsGeometry()
		p2 := regularPolygon(geom.XY{X: 1, Y: 0}, 1.0, sz).AsGeometry()
		b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
			for _, op := range []struct {
				name string
				fn   func(geom.Geometry, geom.Geometry) (geom.Geometry, error)
			}{
				{"Go_Intersection", geom.Intersection},
				{"Go_Difference", geom.Difference},
				{"Go_SymmetricDifference", geom.SymmetricDifference},
				{"Go_Union", geom.Union},
				{"GEOS_Intersection", adaptGEOSSetOp(geos.Intersection)},
				{"GEOS_Difference", adaptGEOSSetOp(geos.Difference)},
				{"GEOS_SymmetricDifference", adaptGEOSSetOp(geos.SymmetricDifference)},
				{"GEOS_Union", adaptGEOSSetOp(geos.Union)},
			} {
				b.Run(op.name, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						if _, err := op.fn(p1, p2); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}

func adaptGEOSSetOp(
	setOp func(_, _ geom.Geometry) (geom.Geometry, error),
) func(_, _ geom.Geometry) (geom.Geometry, error) {
	return func(g1, g2 geom.Geometry) (geom.Geometry, error) {
		return setOp(g1, g2)
	}
}
