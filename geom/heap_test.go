package geom

import (
	"math/rand"
	"testing"
)

func TestHeap(t *testing.T) {
	for _, tt := range []struct {
		name string
		less func(i, j int) bool
	}{
		{
			name: "forward",
			less: func(i, j int) bool { return i < j },
		},
		{
			name: "backwards",
			less: func(i, j int) bool { return i > j },
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			heap := intHeap{less: tt.less}
			list := []int{}

			seed := int64(1577816310618750611)
			rnd := rand.New(rand.NewSource(seed))
			t.Logf("seed %v", seed)

			minListIndex := func() int {
				minI := -1
				for i := range list {
					if minI == -1 || tt.less(list[i], list[minI]) {
						minI = i
					}
				}
				return minI
			}
			check := func() {
				if len(heap.data) != len(list) {
					t.Fatal("lengths differ")
				}
				if len(heap.data) > 0 {
					minI := minListIndex()
					if tt.less(heap.data[0], list[minI]) || tt.less(list[minI], heap.data[0]) {
						t.Log("heap min:", heap.data[0])
						t.Log("list min:", list[minI])
						t.Fatal("min from heap and list don't match")
					}
				}
			}
			push := func() {
				x := rnd.Int()
				t.Log("PUSH", x)
				heap.push(x)
				list = append(list, x)
			}
			pop := func() {
				heap.pop()
				t.Log("POP")
				minI := minListIndex()
				list[minI] = list[len(list)-1]
				list = list[:len(list)-1]
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

			if len(heap.data) != 0 {
				t.Fatalf("not empty: %d", len(heap.data))
			}
		})
	}
}
