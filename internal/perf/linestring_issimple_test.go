package perf

import (
	"fmt"
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
