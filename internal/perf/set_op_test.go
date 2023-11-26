// Package perf_test contains performance benchmarks that don't make sense to
// include in any particular package (they may test code from multiple
// packages).
package perf_test

import (
	"fmt"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/geos"
)

func BenchmarkSetOperation(b *testing.B) {
	for sz := range []int{10, 100, 1000} {
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
				{"GEOS_Intersection", geos.Intersection},
				{"GEOS_Difference", geos.Difference},
				{"GEOS_SymmetricDifference", geos.SymmetricDifference},
				{"GEOS_Union", geos.Union},
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
