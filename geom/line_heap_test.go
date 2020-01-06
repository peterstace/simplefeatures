package geom

import (
	"math"
	"math/rand"
	"testing"
)

func TestLineHeap(t *testing.T) {
	var heap lineHeap

	seed := int64(1577816310618750611)
	rnd := rand.New(rand.NewSource(seed))
	t.Logf("seed %v", seed)

	check := func() {
		for i := range heap {
			childA := 2*i + 1
			childB := 2*i + 2
			le := func(i, j int) bool {
				return heap[i].EndPoint().XY().X <= heap[j].EndPoint().XY().X
			}
			if childA < len(heap) {
				if !le(i, childA) {
					t.Fatal("heap invariant doesn't hold")
				}
			}
			if childB < len(heap) {
				if !le(i, childB) {
					t.Fatal("heap invariant doesn't hold")
				}
			}
		}
	}
	push := func() {
		ln, err := NewLineXY(
			XY{
				X: math.Round(10 + 90*rnd.Float64()),
				Y: math.Round(10 + 90*rnd.Float64()),
			},
			XY{
				X: math.Round(10 + 90*rnd.Float64()),
				Y: math.Round(10 + 90*rnd.Float64()),
			},
		)
		if err != nil {
			t.Fatalf("could not make line: %v", err)
		}
		heap.push(ln)
	}
	pop := func() {
		heap.pop()
	}

	const n = 100

	for i := 0; i < n; i++ {
		push()
		check()
		push()
		check()
		pop()
		check()
	}

	for i := 0; i < n; i++ {
		pop()
		check()
	}

	if len(heap) != 0 {
		t.Fatalf("not empty: %d", len(heap))
	}
}
