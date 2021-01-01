package redblack_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/peterstace/simplefeatures/internal/redblack"
)

func TestIterator(t *testing.T) {
	for pop := 0; pop <= 100; pop++ {
		t.Run(fmt.Sprintf("pop=%d", pop), func(t *testing.T) {
			var tr redblack.Tree
			cmp := func(k1, k2 int) int {
				return k1 - k2
			}
			var want []int
			for i := 0; i < pop; i++ {
				tr.Insert(i, cmp)
				want = append(want, i)
			}

			var got []int
			iter := tr.Begin()
			for iter.Next() {
				got = append(got, iter.Key())
			}

			if !reflect.DeepEqual(want, got) {
				t.Logf("want: %v", want)
				t.Logf("got:  %v", got)
				t.Fatal("mismatch")
			}
		})
	}
}
