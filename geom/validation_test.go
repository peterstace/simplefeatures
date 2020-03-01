package geom_test

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
	. "github.com/peterstace/simplefeatures/internal/geomtest"
)

func xy(x, y float64) Coordinates {
	return Coordinates{XY: XY{x, y}}
}

func TestLineValidation(t *testing.T) {
	for i, pts := range [][2]Coordinates{
		{xy(0, 0), xy(0, 0)},
		{xy(-1, -1), xy(-1, -1)},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := NewLineC(pts[0], pts[1])
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestLineStringValidation(t *testing.T) {
	for i, pts := range [][]Coordinates{
		{xy(0, 0)},
		{xy(1, 1)},
		{xy(0, 0), xy(0, 0)},
		{xy(1, 1), xy(1, 1)},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := NewLineStringC(pts)
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
			_, err := UnmarshalWKT(strings.NewReader(wkt))
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
			_, err := UnmarshalWKT(strings.NewReader(wkt))
			if err == nil {
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
	} {
		t.Run(fmt.Sprintf("valid_%d", i), func(t *testing.T) {
			GeomFromWKT(t, wkt)
		})
	}
	for i, wkt := range []string{
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
	} {
		t.Run(fmt.Sprintf("invalid_%d", i), func(t *testing.T) {
			_, err := UnmarshalWKT(strings.NewReader(wkt))
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
			coords := [][]Coordinates{{}}
			coords[0] = make([]Coordinates, sz+1)
			for i := 0; i < sz; i++ {
				angle := float64(i) / float64(sz) * 2 * math.Pi
				coords[0][i].X = math.Cos(angle)
				coords[0][i].Y = math.Sin(angle)
			}
			coords[0][sz] = coords[0][0]

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := NewPolygonC(coords); err != nil {
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
			coords := make([][]XY, sz*sz+1)
			coords[0] = []XY{XY{0, 0}, XY{0, 1}, XY{1, 1}, XY{1, 0}, XY{0, 0}}
			for i := 0; i < sz*sz; i++ {
				center := XY{
					X: (0.5 + float64(i/sz)) / float64(sz),
					Y: (0.5 + float64(i%sz)) / float64(sz),
				}
				dx := rnd.Float64() * 0.5 / float64(sz)
				dy := rnd.Float64() * 0.5 / float64(sz)
				coords[1+i] = []XY{
					center.Add(XY{-dx, -dy}),
					center.Add(XY{dx, -dy}),
					center.Add(XY{dx, dy}),
					center.Add(XY{-dx, dy}),
					center.Add(XY{-dx, -dy}),
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := NewPolygonXY(coords); err != nil {
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
			coords := make([][][]XY, sz*sz)
			for i := 0; i < sz*sz; i++ {
				center := XY{
					X: (0.5 + float64(i/sz)) / float64(sz),
					Y: (0.5 + float64(i%sz)) / float64(sz),
				}
				dx := rnd.Float64() * 0.5 / float64(sz)
				dy := rnd.Float64() * 0.5 / float64(sz)
				coords[i] = [][]XY{{
					center.Add(XY{-dx, -dy}),
					center.Add(XY{dx, -dy}),
					center.Add(XY{dx, dy}),
					center.Add(XY{-dx, dy}),
					center.Add(XY{-dx, -dy}),
				}}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := NewMultiPolygonXY(coords); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
