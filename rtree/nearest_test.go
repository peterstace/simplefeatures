package rtree

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

func TestNearest(t *testing.T) {
	for pop := 0.0; pop < 1000; pop = (pop + 1) * 1.1 {
		population := int(pop)

		t.Run(fmt.Sprintf("n=%d", population), func(t *testing.T) {
			rnd := rand.New(rand.NewSource(0))
			rt, boxes := testBulkLoad(rnd, population, 0.9, 0.1)
			checkInvariants(t, rt, boxes)
			checkNearest(t, rt, boxes, rnd)
		})
	}
}

func checkNearest(t *testing.T, rt RTree, boxes []Box, rnd *rand.Rand) {
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

func TestNearestEarlyStop(t *testing.T) {
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

	t.Run("stop using sentinal", func(t *testing.T) {
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
