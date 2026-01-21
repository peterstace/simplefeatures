// +build gofuzzbeta

package geom_test

import (
	"strings"
	"testing"
	"time"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/extract"
)

func FuzzParseUnmarshalWKT(f *testing.F) {
	corpus, err := extract.StringsFromSource("..")
	if err != nil {
		f.Fatalf("could not build corpus: %v", err)
	}
	for _, str := range corpus {
		if allowInCorpus(str) {
			f.Add(str)
			f.Log(str)
		}
	}

	f.Fuzz(func(t *testing.T, wkt string) {
		done := make(chan struct{})
		go func() {
			geom.UnmarshalWKT(wkt, geom.DisableAllValidations)
			close(done)
		}()
		select {
		case <-done:
			// do nothing
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timed out")
		}
	})
}

func allowInCorpus(s string) bool {
	for _, prefix := range []string{
		"POINT",
		"MULTIPOINT",
		"LINESTRING",
		"MULTILINESTRING",
		"POLYGON",
		"MULTIPOLYGON",
		"GEOMETRYCOLLECTION",
	} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
