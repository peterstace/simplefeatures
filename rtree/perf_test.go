package rtree

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkDelete(b *testing.B) {
	for pop := 100; pop <= 10000; pop *= 10 {
		b.Run(fmt.Sprintf("n=%d", pop), func(b *testing.B) {
			var boxes []Box
			rts := make([]RTree, b.N)
			for i := 0; i < b.N; i++ {
				rnd := rand.New(rand.NewSource(0))
				rts[i], boxes = testBulkLoad(rnd, pop, 0.9, 0.1)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j, box := range boxes {
					if !rts[i].Delete(box, j) {
						b.Fatal("could not delete")
					}
				}
			}
		})
	}
}
