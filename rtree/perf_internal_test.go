package rtree

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkBulk(b *testing.B) {
	for _, pop := range [...]int{10, 100, 1000, 10_000, 100_000} {
		rnd := rand.New(rand.NewSource(0))
		boxes := make([]Box, pop)
		for i := range boxes {
			boxes[i] = randomBox(rnd, 0.9, 0.1)
		}
		inserts := make([]BulkItem[int], len(boxes))
		for i := range inserts {
			inserts[i].Box = boxes[i]
			inserts[i].Record = i
		}
		b.Run(fmt.Sprintf("n=%d", pop), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				BulkLoad(inserts)
			}
		})
	}
}

func BenchmarkRangeSearch(b *testing.B) {
	for _, pop := range [...]int{10, 100, 1000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("n=%d", pop), func(b *testing.B) {
			rnd := rand.New(rand.NewSource(0))
			tree, _ := testBulkLoad(rnd, pop)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tree.RangeSearch(Box{0.5, 0.5, 0.5, 0.5}, func(int) error { return nil })
			}
		})
	}
}
