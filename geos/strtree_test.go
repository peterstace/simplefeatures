package geos_test

import (
	"math"
	"math/rand"
	"sort"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/geos"
)

func TestSTRTreeNodeCapacityValidation(t *testing.T) {
	for _, tc := range []struct {
		capacity int
		wantErr  bool
	}{
		{1, true},
		{2, false},
		{3, false},
		{63, false},
		{64, false},
		{65, true},
	} {
		t.Run(strconv.Itoa(tc.capacity), func(t *testing.T) {
			tr, err := geos.NewSTRTree(tc.capacity, nil)
			if err == nil {
				defer tr.Close()
			}
			if tc.wantErr {
				expectErr(t, err)
			} else {
				expectNoErr(t, err)
			}
		})
	}
}

func TestSTRTreeEmptyEnvelope(t *testing.T) {
	var emptyEnv geom.Envelope
	tree, err := geos.NewSTRTree(4, []geom.Envelope{emptyEnv})
	expectNoErr(t, err)
	var got []int
	tree.Iterate(func(ridx int) {
		got = append(got, ridx)
	})
	expectIntEq(t, 0, len(got))
}

func TestSTRTreeIterate(t *testing.T) {
	for i := 0; i < 64; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var boxes []geom.Envelope
			for j := 0; j < i; j++ {
				boxes = append(boxes, randomEnv(t))
			}

			tree, err := geos.NewSTRTree(4, boxes)
			expectNoErr(t, err)
			defer tree.Close()

			var found []int
			tree.Iterate(func(ridx int) {
				found = append(found, ridx)
			})
			sort.Ints(found)

			expectIntEq(t, i, len(found))
			for j := 0; j < i; j++ {
				expectIntEq(t, j, found[j])
			}
		})
	}
}

func randomEnv(t *testing.T) geom.Envelope {
	t.Helper()
	bbox, err := geom.NewEnvelope([]geom.XY{
		{X: randomFloat2DP(), Y: randomFloat2DP()},
		{X: randomFloat2DP(), Y: randomFloat2DP()},
	})
	expectNoErr(t, err)
	return bbox
}

func randomFloat2DP() float64 {
	return math.Trunc(rand.Float64()*100) / 100
}
