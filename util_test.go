package simplefeatures_test

import (
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

// eq first unmarshals each WKT, then sees if they are equals via the Equals
// method. If the method panics (usually because it's not implemented for that
// combination of types), then the WKTs are compared textually.
func eq(t *testing.T, wkt1, wkt2 string) bool {
	g1, err := UnmarshalWKT(strings.NewReader(wkt1))
	if err != nil {
		t.Fatal(err)
	}
	g2, err := UnmarshalWKT(strings.NewReader(wkt2))
	if err != nil {
		t.Fatal(err)
	}

	var panicked bool
	var equal bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		equal = g1.Equals(g2)
	}()

	if !panicked {
		return equal
	}
	return wkt1 == wkt2
}
