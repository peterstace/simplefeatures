package geom_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	. "github.com/peterstace/simplefeatures/geom"
)

// regularPolygon computes a regular polygon circumscribed by a circle with the
// given center and radius. Sides must be at least 3 or it will panic.
func regularPolygon(center XY, radius float64, sides int) Polygon {
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
	ring, err := NewLineString(NewSequence(coords, DimXY), geom.DisableAllValidations)
	if err != nil {
		panic(err)
	}
	poly, err := NewPolygon([]LineString{ring}, geom.DisableAllValidations)
	if err != nil {
		panic(err)
	}
	return poly
}

func BenchmarkMarshalWKB(b *testing.B) {
	b.Run("polygon", func(b *testing.B) {
		for _, sz := range []int{10, 100, 1000, 10000} {
			b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
				poly := regularPolygon(XY{}, 1.0, sz)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					poly.AsBinary()
				}
			})
		}
	})
}

func BenchmarkUnmarshalWKB(b *testing.B) {
	b.Run("polygon", func(b *testing.B) {
		for _, sz := range []int{10, 100, 1000, 10000} {
			b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
				wkb := regularPolygon(XY{}, 1.0, sz).AsBinary()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := UnmarshalWKB(wkb, DisableAllValidations)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	})
}

func BenchmarkIntersectsLineStringWithLineString(b *testing.B) {
	for _, sz := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
			var floats1, floats2 []float64
			for i := 0; i < sz; i++ {
				x := float64(i) / float64(sz)
				floats1 = append(floats1, x, 1)
				floats2 = append(floats2, x, 2)
			}
			seq1 := geom.NewSequence(floats1, geom.DimXY)
			seq2 := geom.NewSequence(floats2, geom.DimXY)
			ls1, err := geom.NewLineString(seq1)
			if err != nil {
				b.Fatal(err)
			}
			ls2, err := geom.NewLineString(seq2)
			if err != nil {
				b.Fatal(err)
			}
			ls1g := ls1.AsGeometry()
			ls2g := ls2.AsGeometry()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if Intersects(ls1g, ls2g) {
					b.Fatal("should not intersect")
				}
			}
		})
	}
}

func BenchmarkIntersectsMultiPointWithMultiPoint(b *testing.B) {
	for _, sz := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", 2*sz), func(b *testing.B) {
			rnd := rand.New(rand.NewSource(0))
			var pointsA, pointsB []Point
			for i := 0; i < sz; i++ {
				ptA, err := XY{X: rnd.Float64(), Y: rnd.Float64()}.AsPoint()
				expectNoErr(b, err)
				pointsA = append(pointsA, ptA)
				ptB, err := XY{X: rnd.Float64(), Y: rnd.Float64()}.AsPoint()
				expectNoErr(b, err)
				pointsB = append(pointsB, ptB)
			}
			mpA := NewMultiPoint(pointsA).AsGeometry()
			mpB := NewMultiPoint(pointsB).AsGeometry()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if Intersects(mpA, mpB) {
					b.Fatal("shouldn't have intersected")
				}
			}
		})
	}
}

func BenchmarkPolygonSingleRingValidation(b *testing.B) {
	for _, sz := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
			floats := make([]float64, 2*(sz+1))
			for i := 0; i < sz; i++ {
				angle := float64(i) / float64(sz) * 2 * math.Pi
				floats[2*i+0] = math.Cos(angle)
				floats[2*i+1] = math.Sin(angle)
			}
			floats[2*sz+0] = floats[0]
			floats[2*sz+1] = floats[1]
			ring, err := NewLineString(NewSequence(floats, DimXY))
			if err != nil {
				b.Fatal(err)
			}
			rings := []LineString{ring}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := NewPolygon(rings); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkPolygonMultipleRingsValidation(b *testing.B) {
	for _, sz := range []int{2, 6, 20, 64} {
		b.Run(fmt.Sprintf("n=%d", sz*sz), func(b *testing.B) {
			rnd := rand.New(rand.NewSource(0))
			rings := make([]LineString, sz*sz+1)
			var err error
			rings[0], err = NewLineString(NewSequence([]float64{0, 0, 0, 1, 1, 1, 1, 0, 0, 0}, DimXY))
			if err != nil {
				b.Fatal(err)
			}
			for i := 0; i < sz*sz; i++ {
				center := XY{
					X: (0.5 + float64(i/sz)) / float64(sz),
					Y: (0.5 + float64(i%sz)) / float64(sz),
				}
				dx := rnd.Float64() * 0.5 / float64(sz)
				dy := rnd.Float64() * 0.5 / float64(sz)
				rings[1+i], err = NewLineString(NewSequence([]float64{
					center.X - dx, center.Y - dy,
					center.X + dx, center.Y - dy,
					center.X + dx, center.Y + dy,
					center.X - dx, center.Y + dy,
					center.X - dx, center.Y - dy,
				}, DimXY))
				if err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := NewPolygon(rings); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkPolygonZigZagRingsValidation(b *testing.B) {
	for _, sz := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
			outerRingEnv, err := NewEnvelope(XY{}, XY{7, float64(sz + 1)})
			expectNoErr(b, err)
			outerRing := outerRingEnv.AsGeometry().AsPolygon().ExteriorRing()
			var leftFloats, rightFloats []float64
			for i := 0; i < sz; i++ {
				leftFloats = append(leftFloats, float64(2+(i%2)*2), float64(1+i))
				rightFloats = append(rightFloats, float64(3+(i%2)*2), float64(1+i))
			}
			leftFloats = append(leftFloats,
				1, float64(sz),
				1, 1,
				2, 1,
			)
			rightFloats = append(rightFloats,
				6, float64(sz),
				6, 1,
				3, 1,
			)
			leftRing, err := NewLineString(NewSequence(leftFloats, DimXY))
			if err != nil {
				b.Fatal(err)
			}
			rightRing, err := NewLineString(NewSequence(rightFloats, DimXY))
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := NewPolygon([]LineString{outerRing, leftRing, rightRing})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkPolygonAnnulusValidation(b *testing.B) {
	for _, sz := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
			outer := regularPolygon(XY{}, 1.0, sz/2).ExteriorRing()
			inner := regularPolygon(XY{}, 0.5, sz/2).ExteriorRing()
			rings := []LineString{outer, inner}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := NewPolygon(rings); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkMultipolygonValidation(b *testing.B) {
	for _, sz := range []int{1, 2, 4, 8, 16, 32} {
		b.Run(fmt.Sprintf("n=%d", sz*sz), func(b *testing.B) {
			rnd := rand.New(rand.NewSource(0))
			polys := make([]Polygon, sz*sz)
			for i := 0; i < sz*sz; i++ {
				cx := (0.5 + float64(i/sz)) / float64(sz)
				cy := (0.5 + float64(i%sz)) / float64(sz)
				dx := rnd.Float64() * 0.5 / float64(sz)
				dy := rnd.Float64() * 0.5 / float64(sz)
				ring, err := NewLineString(NewSequence([]float64{
					cx - dx, cy - dy,
					cx + dx, cy - dy,
					cx + dx, cy + dy,
					cx - dx, cy + dy,
					cx - dx, cy - dy,
				}, DimXY))
				if err != nil {
					b.Fatal(err)
				}
				polys[i], err = NewPolygon([]LineString{ring})
				if err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := NewMultiPolygon(polys); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkMultiPolygonTwoCircles(b *testing.B) {
	for _, sz := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
			const eps = 0.1
			polys := []Polygon{
				regularPolygon(XY{X: -eps, Y: -eps}, 1.0, sz),
				regularPolygon(XY{X: math.Sqrt2, Y: math.Sqrt2}, 1.0, sz),
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := NewMultiPolygon(polys); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkMultiPolygonMultipleTouchingPoints(b *testing.B) {
	for _, sz := range []int{1, 10, 100, 1000} {
		b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
			fs1 := []float64{0, 0}
			fs2 := []float64{4, 0}
			for i := 0; i < 2*sz+1; i++ {
				fs1 = append(fs1, float64(1+i%2), float64(i))
				fs2 = append(fs2, float64(3-i%2), float64(i))
			}
			fs1 = append(fs1, 0, float64(2*sz), 0, 0)
			fs2 = append(fs2, 4, float64(2*sz), 4, 0)

			ls1, err := NewLineString(NewSequence(fs1, DimXY))
			if err != nil {
				b.Fatal(err)
			}
			ls2, err := NewLineString(NewSequence(fs2, DimXY))
			if err != nil {
				b.Fatal(err)
			}
			p1, err := NewPolygon([]LineString{ls1})
			if err != nil {
				b.Fatal(err)
			}
			p2, err := NewPolygon([]LineString{ls2})
			if err != nil {
				b.Fatal(err)
			}
			polys := []Polygon{p1, p2}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := NewMultiPolygon(polys)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkWKTParsing(b *testing.B) {
	for _, tc := range []struct {
		desc string
		wkt  string
	}{
		{
			"point",
			"POINT(-3.14159265359 3.14159265359)",
		},
	} {
		b.Run(tc.desc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := UnmarshalWKT(tc.wkt); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDistancePolygonToPolygonOrdering(b *testing.B) {
	for _, sz := range []int{100, 1000} {
		for _, swap := range []bool{false, true} {
			b.Run(fmt.Sprintf("n=%d_swap=%t", sz, swap), func(b *testing.B) {
				p1 := regularPolygon(geom.XY{0, 0}, 1.0, sz/10).AsGeometry()
				p2 := regularPolygon(geom.XY{3, 0}, 1.0, sz).AsGeometry()
				if swap {
					p1, p2 = p2, p1
				}
				for i := 0; i < b.N; i++ {
					Distance(p1, p2)
				}
			})
		}
	}
}

func BenchmarkIntersectionPolygonWithPolygonOrdering(b *testing.B) {
	for _, sz := range []int{100, 1000} {
		for _, swap := range []bool{false, true} {
			b.Run(fmt.Sprintf("n=%d_swap=%t", sz, swap), func(b *testing.B) {
				p1 := regularPolygon(geom.XY{0, 0}, 1.0, sz/10).AsGeometry()
				p2 := regularPolygon(geom.XY{1, 0}, 1.0, sz).AsGeometry()
				if swap {
					p1, p2 = p2, p1
				}
				for i := 0; i < b.N; i++ {
					Distance(p1, p2)
				}
			})
		}
	}
}

func BenchmarkMultiLineStringIsSimpleManyLineStrings(b *testing.B) {
	for _, sz := range []int{100, 1000} {
		b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
			var lss []LineString
			for i := 0; i < sz; i++ {
				seq := NewSequence([]float64{
					float64(2*i + 0),
					float64(2*i + 0),
					float64(2*i + 1),
					float64(2*i + 1),
				}, DimXY)
				ls, err := NewLineString(seq)
				if err != nil {
					b.Fatal(err)
				}
				lss = append(lss, ls)
			}
			mls := NewMultiLineString(lss)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				mls.IsSimple()
			}
		})
	}
}
