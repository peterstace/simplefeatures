package rtree

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
)

func BenchmarkDelete(b *testing.B) {
	// We test from 100 to 10000 instead of the regular 10 to 10000 because this
	// benchmark has a long startup time (which `go test` doesn't consider when
	// limiting the total runtime of the entire benchmark).
	for pop := 100; pop <= 10000; pop *= 10 {
		b.Run(fmt.Sprintf("n=%d", pop), func(b *testing.B) {

			var boxes []Box
			trees := make([]*RTree, b.N)
			for i := 0; i < b.N; i++ {
				rnd := rand.New(rand.NewSource(0))
				trees[i], boxes = testBulkLoad(rnd, pop, 0.9, 0.1)
			}
			runtime.GC()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				for j, box := range boxes {
					if !trees[i].Delete(box, j) {
						b.Fatal("could not delete")
					}
				}
			}
		})
	}
}

func BenchmarkBulk(b *testing.B) {
	for _, pop := range [...]int{10, 100, 1000, 10_000, 100_000} {
		rnd := rand.New(rand.NewSource(0))
		boxes := make([]Box, pop)
		for i := range boxes {
			boxes[i] = randomBox(rnd, 0.9, 0.1)
		}
		inserts := make([]BulkItem, len(boxes))
		for i := range inserts {
			inserts[i].Box = boxes[i]
			inserts[i].RecordID = i
		}
		b.Run(fmt.Sprintf("n=%d", pop), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				BulkLoad(inserts)
			}
		})
	}
}

func BenchmarkInsert(b *testing.B) {
	for _, pop := range [...]int{10, 100, 1000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("n=%d", pop), func(b *testing.B) {
			rnd := rand.New(rand.NewSource(0))
			boxes := make([]Box, pop)
			for i := range boxes {
				boxes[i] = randomBox(rnd, 0.9, 0.1)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var tree RTree
				for j, b := range boxes {
					tree.Insert(b, j)
				}
			}
		})
	}
}

func BenchmarkRangeSearch(b *testing.B) {
	for _, pop := range [...]int{10, 100, 1000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("n=%d", pop), func(b *testing.B) {
			rnd := rand.New(rand.NewSource(0))
			tree, _ := testBulkLoad(rnd, pop, 0.9, 0.1)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tree.RangeSearch(Box{0.5, 0.5, 0.5, 0.5}, func(int) error { return nil })
			}
		})
	}
}
