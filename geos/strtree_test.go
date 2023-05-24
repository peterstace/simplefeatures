package geos_test

import (
	"strconv"
	"testing"

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
