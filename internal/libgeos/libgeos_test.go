package libgeos_test

import (
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/libgeos"
)

func TestAsText(t *testing.T) {
	g, err := geom.UnmarshalWKT(strings.NewReader("POINT(1 2)"))
	if err != nil {
		t.Fatal("could not unmarshal WKT")
	}

	h := libgeos.NewHandle()
	defer h.Close()
	wkt, err := h.AsText(g)
	if err != nil {
		t.Fatal("could not convert to text")
	}
	if want := "POINT (1.0000000000000000 2.0000000000000000)"; wkt != want {
		t.Errorf("want: %v got: %v", want, wkt)
	}
}
