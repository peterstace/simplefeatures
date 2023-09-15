package perf

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

var dummyBool bool

func BenchmarkLineStringIsSimpleCircle(b *testing.B) {
	for _, sz := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
			circ := regularPolygon(geom.XY{}, 1.0, sz)
			ring := circ.ExteriorRing()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dummyBool = ring.IsSimple()
			}
		})
	}
}

func BenchmarkLineStringIsSimpleZigZag(b *testing.B) {
	for _, sz := range []int{10, 100, 1000, 10000} {
		b.Run(strconv.Itoa(sz), func(b *testing.B) {
			floats := make([]float64, 2*sz)
			for i := 0; i < sz; i++ {
				floats[2*i+0] = float64(i%2) * 0.01
				floats[2*i+1] = float64(i) * 0.01
			}
			seq := geom.NewSequence(floats, geom.DimXY)
			ls := geom.NewLineString(seq)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if !ls.IsSimple() {
					b.Fatal("not simple")
				}
			}
		})
	}
}
