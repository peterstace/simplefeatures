package simplefeatures_test

import (
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func geomFromWKT(t *testing.T, wkt string) Geometry {
	geom, err := UnmarshalWKT(strings.NewReader(wkt))
	if err != nil {
		t.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
	}
	return geom
}
