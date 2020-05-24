package rtree

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkDelete(b *testing.B) {
	for pop := 100; pop <= 10000; pop *= 10 {
		b.Run(fmt.Sprintf("n=%d", pop), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				rnd := rand.New(rand.NewSource(0))
				rt, boxes := testBulkLoad(rnd, pop, 0.9, 0.1)
				b.StartTimer()
				for j, box := range boxes {
					if !rt.Delete(box, j) {
						b.Fatal("could not delete")
					}
				}
			}
		})
	}
}
