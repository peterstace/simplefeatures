package oracle

import (
	"fmt"
	"math"
	"testing"

	sf "github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/jts"
)

// wavy returns the ring of a star-shaped polygon with n vertices centred on
// (cx, cy).
func wavy(cx, cy float64, n int) [][2]float64 {
	pts := make([][2]float64, 0, n+1)
	for i := 0; i < n; i++ {
		a := 2 * math.Pi * float64(i) / float64(n)
		r := 1 + 0.2*math.Sin(7*a)
		pts = append(pts, [2]float64{cx + r*math.Cos(a), cy + r*math.Sin(a)})
	}
	return append(pts, pts[0])
}

func portPolygon(ring [][2]float64) *jts.Geom_Geometry {
	f := jts.Geom_GeometryFactory_new()
	cs := make([]*jts.Geom_Coordinate, len(ring))
	for i, p := range ring {
		cs[i] = jts.Geom_Coordinate_new_double_double(p[0], p[1])
	}
	return geom(f.CreatePolygon_CoordinateArr(cs))
}

func sfPolygon(ring [][2]float64) sf.Geometry {
	flat := make([]float64, 0, 2*len(ring))
	for _, p := range ring {
		flat = append(flat, p[0], p[1])
	}
	ls := sf.NewLineString(sf.NewSequence(flat, sf.DimXY))
	return sf.NewPolygon([]sf.LineString{ls}).AsGeometry()
}

// portFromSF converts a simplefeatures polygon to the port, allocating each
// coordinate separately, as JTS's WKBReader would create them.
func portFromSF(g sf.Geometry) *jts.Geom_Geometry {
	f := jts.Geom_GeometryFactory_new()
	seq := g.MustAsPolygon().ExteriorRing().Coordinates()
	cs := make([]*jts.Geom_Coordinate, seq.Length())
	for i := range cs {
		xy := seq.GetXY(i)
		cs[i] = geomCoord(jts.Geom_CoordinateXY_new_double_double(xy.X, xy.Y))
	}
	return geom(f.CreatePolygon_CoordinateArr(cs))
}

func geomCoord(c *jts.Geom_CoordinateXY) *jts.Geom_Coordinate {
	return &c.Geom_Coordinate
}

// BenchmarkOverlay compares this port with the round 1 port used by the geom
// package. The geom functions include converting to and from JTS geometries,
// the port benchmarks do not.
func BenchmarkOverlay(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		ra, rb := wavy(0, 0, n), wavy(0.5, 0.3, n)
		pa, pb := portPolygon(ra), portPolygon(rb)
		sa, sb := sfPolygon(ra), sfPolygon(rb)
		b.Run(fmt.Sprintf("port/union/n=%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				jts.OperationOverlayng_OverlayNGRobust_overlay(pa, pb, 2)
			}
		})
		b.Run(fmt.Sprintf("round1/union/n=%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sf.Union(sa, sb); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("port+convert/union/n=%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				jts.OperationOverlayng_OverlayNGRobust_overlay(portFromSF(sa), portFromSF(sb), 2)
			}
		})
		b.Run(fmt.Sprintf("port+convert/relate/n=%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				jts.OperationRelateng_RelateNG_relate_Geometry_Geometry(portFromSF(sa), portFromSF(sb))
			}
		})
		b.Run(fmt.Sprintf("port/relate/n=%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				jts.OperationRelateng_RelateNG_relate_Geometry_Geometry(pa, pb)
			}
		})
		b.Run(fmt.Sprintf("round1/relate/n=%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sf.Relate(sa, sb); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
