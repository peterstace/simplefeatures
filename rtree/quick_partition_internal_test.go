package rtree

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

func TestQuickPartition(t *testing.T) {
	testCases := [][]int{
		{1},
		{1, 1},
		{1, 1, 1},
		{1, 1, 1, 1},
		{1, 1, 1, 1, 1},
		{1, 2},
		{2, 1},
		{1, 1, 2},
		{1, 2, 1},
		{2, 1, 1},
	}
	for i := 1; i <= 100; i++ {
		allNums := make([]int, i)
		for j := range allNums {
			allNums[j] = j
		}
		rand.New(rand.NewSource(0)).Shuffle(i, func(a, b int) {
			allNums[a], allNums[b] = allNums[b], allNums[a]
		})
		testCases = append(testCases, allNums)
	}
	for i := 1; i <= 20; i++ {
		allNums := make([]int, i*5)
		for j := range allNums {
			allNums[j] = j / 5 // results in duplicates
		}
		rand.New(rand.NewSource(0)).Shuffle(i*5, func(a, b int) {
			allNums[a], allNums[b] = allNums[b], allNums[a]
		})
		testCases = append(testCases, allNums)
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for k := range tc {
				t.Run(fmt.Sprintf("k=%d", k), func(t *testing.T) {
					items := make([]BulkItem, 0, len(tc))
					for _, num := range tc {
						f := float64(num)
						items = append(items, BulkItem{
							Box{f, f, f, f},
							len(items),
						})
					}

					quickPartition(items, k, false)
					kth := items[k]
					for j, item := range items {
						switch {
						case j < k:
							if item.Box.MaxX > kth.Box.MaxX {
								t.Errorf("item at index %d not partitioned", j)
							}
						case j > k:
							if item.Box.MaxX < kth.Box.MaxX {
								t.Errorf("item at index %d not partitioned", j)
							}
						default:
							if item.Box.MaxX != kth.Box.MaxX {
								t.Errorf("item at index %d not partitioned", j)
							}
						}
					}
				})
			}
		})
	}
}
