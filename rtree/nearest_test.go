package rtree

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
)

func TestNearest(t *testing.T) {
	for _, population := range testPopulations(66, 1000, 1.1) {
		t.Run(fmt.Sprintf("n=%d", population), func(t *testing.T) {
			rnd := rand.New(rand.NewSource(0))
			rt, boxes := testBulkLoad(rnd, population, 0.9, 0.1)
			checkInvariants(t, rt, boxes)
			checkPrioritySearch(t, rt, boxes, rnd)
			checkNearest(t, rt, boxes, rnd)
		})
	}
}

func checkNearest(t *testing.T, rt *RTree, boxes []Box, rnd *rand.Rand) {
	for i := 0; i < 10; i++ {
		originBB := randomBox(rnd, 0.9, 0.1)
		got, ok := rt.Nearest(originBB)

		if ok && len(boxes) == 0 {
			t.Fatal("found nearest but no boxes")
		}
		if !ok && len(boxes) != 0 {
			t.Fatal("could not find nearest but have some boxes")
		}
		if !ok {
			continue
		}

		bestDist := math.Inf(+1)
		for j := range boxes {
			bestDist = math.Min(bestDist, squaredEuclideanDistance(originBB, boxes[j]))
		}
		if bestDist != squaredEuclideanDistance(originBB, boxes[got]) {
			t.Errorf("mismatched distance")
		}
	}
}

func checkPrioritySearch(t *testing.T, rt *RTree, boxes []Box, rnd *rand.Rand) {
	for i := 0; i < 10; i++ {
		var got []int
		originBB := randomBox(rnd, 0.9, 0.1)
		t.Logf("origin: %v", originBB)
		rt.PrioritySearch(originBB, func(recordID int) error {
			got = append(got, recordID)
			return nil
		})
		t.Logf("got: %v", got)

		if len(got) != len(boxes) {
			t.Fatal("didn't get all of the boxes")
		}
		if !sort.SliceIsSorted(got, func(i, j int) bool {
			di := squaredEuclideanDistance(originBB, boxes[got[i]])
			dj := squaredEuclideanDistance(originBB, boxes[got[j]])
			return di < dj
		}) {
			t.Fatal("records not in sorted order")
		}
	}
}

func TestPrioritySearchEarlyStop(t *testing.T) {
	rnd := rand.New(rand.NewSource(0))
	boxes := make([]Box, 100)
	for i := range boxes {
		boxes[i] = randomBox(rnd, 0.9, 0.1)
	}

	inserts := make([]BulkItem, len(boxes))
	for i := range inserts {
		inserts[i].Box = boxes[i]
		inserts[i].RecordID = i
	}
	rt := BulkLoad(inserts)
	origin := randomBox(rnd, 0.9, 0.1)

	t.Run("stop using sentinel", func(t *testing.T) {
		var count int
		err := rt.PrioritySearch(origin, func(int) error {
			count++
			if count >= 3 {
				return Stop
			}
			return nil
		})
		if err != nil {
			t.Fatal("got an error but didn't expect to")
		}
		if count != 3 {
			t.Fatalf("didn't stop after 3: %v", count)
		}
	})

	t.Run("stop with user error", func(t *testing.T) {
		var count int
		userErr := errors.New("user error")
		err := rt.PrioritySearch(origin, func(int) error {
			count++
			if count >= 3 {
				return userErr
			}
			return nil
		})
		if err != userErr {
			t.Fatalf("expected to get userErr but got: %v", userErr)
		}
		if count != 3 {
			t.Fatalf("didn't stop after 3: %v", count)
		}
	})
}
