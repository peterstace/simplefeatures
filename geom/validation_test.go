package geom_test

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	. "github.com/peterstace/simplefeatures/geom"
)

func xy(x, y float64) Coordinates {
	return Coordinates{Type: DimXY, XY: XY{x, y}}
}

func TestLineStringValidation(t *testing.T) {
	for i, pts := range [][]float64{
		[]float64{0, 0},
		[]float64{1, 1},
		[]float64{0, 0, 0, 0},
		[]float64{1, 1, 1, 1},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			seq := NewSequence(pts, DimXY)
			_, err := NewLineString(seq)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestPolygonValidation(t *testing.T) {
	for i, wkt := range []string{
		"POLYGON EMPTY",
		"POLYGON((0 0,1 0,1 1,0 1,0 0))",
		"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))",
		`POLYGON(
			(0 0,5 0,5 5,0 5,0 0),
			(1 1,3 1,3 2,1 1),
			(1 1,4 3,3 4,1 1),
			(1 1,2 3,1 3,1 1)
		)`,
		`POLYGON(
			(0 0,5 0,5 5,0 5,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(2 1,3 1,3 2,2 1),
			(1 2,2 3,1 3,1 2),
			(2 2,4 3,3 4,2 2)
		)`,
	} {
		t.Run("valid_"+strconv.Itoa(i), func(t *testing.T) {
			_, err := UnmarshalWKT(wkt)
			if err != nil {
				t.Error(err)
			}
		})
	}
	for i, wkt := range []string{
		// not closed
		"POLYGON((0 0,1 1,0 1))",

		// not simple
		"POLYGON((0 0,1 1,0 1,1 0,0 0))",

		// intersect at a line
		"POLYGON((0 0,3 0,3 3,0 3,0 0),(0 1,1 1,1 2,0 2,0 1))",

		// intersect at two points
		"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 0,3 1,2 2,1 0))",

		// inner ring is outside of the outer ring
		"POLYGON((0 0,3 0,3 3,0 3,0 0),(4 0,7 0,7 3,4 3,4 0))",

		// polygons aren't connected
		`POLYGON(
			(0 0, 4 0, 4 4, 0 4, 0 0),
			(2 0, 3 1, 2 2, 1 1, 2 0),
			(2 2, 3 3, 2 4, 1 3, 2 2)
		)`,
		`POLYGON(
			(0 0, 6 0, 6 5, 0 5, 0 0),
			(2 1, 4 1, 4 2, 2 2, 2 1),
			(2 2, 3 3, 2 4, 1 3, 2 2),
			(4 2, 5 3, 4 4, 3 3, 4 2)
		)`,

		// Nested rings
		`POLYGON(
			(0 0,5 0,5 5,0 5,0 0),
			(1 1,4 1,4 4,1 4,1 1),
			(2 2,3 2,3 3,2 3,2 2)
		)`,
		`POLYGON(
			(0 0,5 0,5 5,0 5,0 0),
			(2 2,3 2,3 3,2 3,2 2),
			(1 1,4 1,4 4,1 4,1 1)
		)`,

		// Contains empty rings.
		`POLYGON(EMPTY)`,
		`POLYGON(EMPTY,(0 0,0 1,1 0,0 0))`,
		`POLYGON((0 0,0 1,1 0,0 0),EMPTY)`,
	} {
		t.Run("invalid_"+strconv.Itoa(i), func(t *testing.T) {
			_, err := UnmarshalWKT(wkt)
			if err == nil {
				t.Log("WKT", wkt)
				t.Error("expected error")
			}
		})
	}
}

func TestMultiPolygonValidation(t *testing.T) {
	for i, wkt := range []string{
		`MULTIPOLYGON EMPTY`,
		`MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)))`,
		`MULTIPOLYGON(
			((0 0,1 0,1 1,0 1,0 0)),
			((2 0,3 0,3 1,2 1,2 0))
		)`,
		`MULTIPOLYGON(
			((0 0,1 0,0 1,0 0)),
			((1 0,2 0,1 1,1 0))
		)`,
		`MULTIPOLYGON(
			((0 0,2 0,2 3,1 1,0 3,0 0)),
			((1 2,2 3,0 3,1 2))
		)`,
		`MULTIPOLYGON(
			((0 0,5 0,5 5,0 5,0 0),(1 1,4 1,4 4,1 4,1 1)),
			((2 2,3 2,3 3,2 3,2 2))
		)`,

		// Child polygons can be empty.
		`MULTIPOLYGON(EMPTY)`,
		`MULTIPOLYGON(((0 0,0 1,1 0,0 0)),EMPTY)`,
		`MULTIPOLYGON(EMPTY,((0 0,0 1,1 0,0 0)))`,

		// Replicates a bug.
		`MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 -1,3 -1,3 0,2 0,2 -1)),((1 1,3 1,3 3,1 3,1 1)))`,
	} {
		t.Run(fmt.Sprintf("valid_%d", i), func(t *testing.T) {
			geomFromWKT(t, wkt)
		})
	}
	for i, wkt := range []string{
		`MULTIPOLYGON(
			((-6 -3,8 4,7 6,-7 -1,-6 -3)),
			((3 -6,5 -5,-2 9,-4 8,3 -6))
		)`,
		`MULTIPOLYGON(
			((0 0,0 1,1 1,1 0,0 0)),
			((1 0,1 1,2 1,2 0,1 0))
		)`,
		`MULTIPOLYGON(
			((0 0,2 0,2 2,0 2,0 0)),
			((1 0,3 0,3 2,1 2,1 0))
		)`,
		`MULTIPOLYGON(
			((1 0,2 0,1 3,1 0)),
			((0 1,3 1,3 2,0 1))
		)`,
		`MULTIPOLYGON(
			((0 0,3 0,3 3,0 3,0 0)),
			((2 1,3 3,1 2,2 1))
		)`,
		`MULTIPOLYGON(
			((2 1,3 3,1 2,2 1)),
			((0 0,3 0,3 3,0 3,0 0))
		)`,
		`MULTIPOLYGON(
			((0 0,0 1,1 0,0 0)),
			((0 0,0 1,1 0,0 0))
		)`,
		`MULTIPOLYGON(
			((0 0,3 0,3 3,0 3,0 0)),
			((1 1,2 1,2 2,1 2,1 1))
		)`,
		`MULTIPOLYGON(
			((1 1,2 1,2 2,1 2,1 1)),
			((0 0,3 0,3 3,0 3,0 0))
		)`,
		`MULTIPOLYGON(
			((0 0,2 0,2 1,0 1,0 0)),
			((0.5 -0.5,1 2,1.5 -0.5,2 2,2 3,0 3,0 2,0.5 -0.5))
		)`,
		`MULTIPOLYGON(
			((0 0,2 0,2 1,0 1,0 0)),
			((0.5 1,1 2,1.5 -0.5,2 2,2 3,0 3,0 2,0.5 1))
		)`,
	} {
		t.Run(fmt.Sprintf("invalid_%d", i), func(t *testing.T) {
			_, err := UnmarshalWKT(wkt)
			if err == nil {
				t.Log(wkt)
				t.Error("expected error")
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
				if _, err := NewPolygonFromRings(rings); err != nil {
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
				if _, err := NewPolygonFromRings(rings); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkPolygonZigZagRingsValidation(b *testing.B) {
	for _, sz := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
			outerRing := NewEnvelope(XY{}, XY{7, float64(sz + 1)}).AsGeometry().AsPolygon().ExteriorRing()
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
				_, err := NewPolygonFromRings([]LineString{outerRing, leftRing, rightRing})
				if err != nil {
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
				polys[i], err = NewPolygonFromRings([]LineString{ring})
				if err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := NewMultiPolygonFromPolygons(polys); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

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
	poly, err := NewPolygonFromRings([]LineString{ring}, geom.DisableAllValidations)
	if err != nil {
		panic(err)
	}
	return poly
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
				if _, err := NewMultiPolygonFromPolygons(polys); err != nil {
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
			p1, err := NewPolygonFromRings([]LineString{ls1})
			if err != nil {
				b.Fatal(err)
			}
			p2, err := NewPolygonFromRings([]LineString{ls2})
			if err != nil {
				b.Fatal(err)
			}
			polys := []Polygon{p1, p2}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := NewMultiPolygonFromPolygons(polys)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
