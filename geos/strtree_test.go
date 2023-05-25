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

func TestSTRTreeIterate(t *testing.T) {
	for i := 0; i < 64; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var entries []geos.STRTreeEntry
			for j := 0; j < i; j++ {
				entries = append(entries, geos.STRTreeEntry{
					BBox: randomEnv(t),
					Item: int(-j),
				})
			}

			tree, err := geos.NewSTRTree(4, entries)
			expectNoErr(t, err)
			defer tree.Close()

			var found []geos.STRTreeEntry
			tree.Iterate(func(entry geos.STRTreeEntry) {
				found = append(found, entry)
			})
			sort.Slice(found, func(i, j int) bool {
				return found[i].Item.(int) < found[j].Item.(int)
			})
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Item.(int) < entries[j].Item.(int)
			})
			expectDeepEq(t, entries, found)
		})
	}
}
