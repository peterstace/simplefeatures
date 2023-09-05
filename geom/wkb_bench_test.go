package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func BenchmarkWKBParse(b *testing.B) {
	for i, tt := range validWKBCases {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_, err := geom.UnmarshalWKBWithoutValidation(hexStringToBytes(b, tt.wkb))
				if err != nil {
					panic(err)
				}
			}
		})
	}
}
