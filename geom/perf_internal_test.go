package geom

import (
	"strconv"
	"testing"
)

var dummyEnv Envelope

func BenchmarkLineEnvelope(b *testing.B) {
	for i, ln := range []line{
		{XY{0, 0}, XY{1, 1}},
		{XY{1, 1}, XY{0, 0}},
		{XY{0, 1}, XY{1, 0}},
		{XY{1, 0}, XY{0, 1}},
	} {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				dummyEnv = ln.envelope()
			}
		})
	}
}
