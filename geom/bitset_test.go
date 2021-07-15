package geom_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestBitSet(t *testing.T) {
	t.Run("one bit at a time", func(t *testing.T) {
		for i := 0; i < 256; i++ {
			t.Run(fmt.Sprintf("bit %d", i), func(t *testing.T) {
				var s geom.BitSet
				expectFalse(t, s.Get(i))
				s.Set(i, true)
				expectTrue(t, s.Get(i))
				s.Set(i, false)
				expectFalse(t, s.Get(i))
			})
		}
	})
	t.Run("many bits at a time", func(t *testing.T) {
		const n = 512
		var want [n]bool
		rnd := rand.New(rand.NewSource(0))
		var s geom.BitSet
		for i := 0; i < n; i++ {
			choice := rnd.Intn(n)
			want[choice] = !want[choice]
			s.Set(choice, want[choice])
			for j := 0; j < n; j++ {
				expectBoolEq(t, s.Get(j), want[j])
			}
		}
	})
}
