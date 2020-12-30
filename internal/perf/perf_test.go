// Package perf contains performance benchmarks that don't make sense to
// include in any particular package (they may test code from multiple
// packages).
package perf

import (
	"fmt"
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/geos"
)

// regularPolygon computes a regular polygon circumscribed by a circle with the
// given center and radius. Sides must be at least 3 or it will panic.
func regularPolygon(center geom.XY, radius float64, sides int) geom.Polygon {
	if sides <= 2 {
		panic(sides)
	}
	coords := make([]float64, 2*(sides+1))
	for i := 0; i < sides; i++ {
		angle := math.Pi/2 + float64(i)/float64(sides)*2*math.Pi
		coords[2*i+0] = center.X + math.Cos(angle)*radius
		coords[2*i+1] = center.Y + math.Sin(angle)*radius
	}
	coords[2*sides+0] = coords[0]
	coords[2*sides+1] = coords[1]
	ring, err := geom.NewLineString(geom.NewSequence(coords, geom.DimXY), geom.DisableAllValidations)
	if err != nil {
		panic(err)
	}
	poly, err := geom.NewPolygonFromRings([]geom.LineString{ring}, geom.DisableAllValidations)
	if err != nil {
		panic(err)
	}
	return poly
}

func BenchmarkSetOperation(b *testing.B) {
	for i := 2; i <= 14; i++ {
		sz := 1 << i
		p1 := regularPolygon(geom.XY{0, 0}, 1.0, sz).AsGeometry()
		p2 := regularPolygon(geom.XY{1, 0}, 1.0, sz).AsGeometry()
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
	setOp func(_, _ geom.Geometry, _ ...geom.ConstructorOption) (geom.Geometry, error),
) func(_, _ geom.Geometry) (geom.Geometry, error) {
	return func(g1, g2 geom.Geometry) (geom.Geometry, error) {
		return setOp(g1, g2)
	}
}
