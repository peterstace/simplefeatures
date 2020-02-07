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

func TestIsSimpleValid(t *testing.T) {
	g, err := geom.UnmarshalWKT(strings.NewReader("LINESTRING(0 0,0 1,1 0,0 0)"))
	if err != nil {
		t.Fatal("could not unmarshal WKT")
	}

	h := libgeos.NewHandle()
	defer h.Close()
	got, err := h.IsSimple(g)
	if err != nil {
		t.Fatal("could not get IsSimple")
	}
	if !got {
		t.Errorf("expected it to be simple")
	}
}

func TestIsNotSimpleInvalid(t *testing.T) {
	g, err := geom.UnmarshalWKT(strings.NewReader("LINESTRING(0 0,1 1,1 0,0 1)"))
	if err != nil {
		t.Fatal("could not unmarshal WKT")
	}

	h := libgeos.NewHandle()
	defer h.Close()
	got, err := h.IsSimple(g)
	if err != nil {
		t.Fatal("could not get IsSimple")
	}
	if got {
		t.Errorf("expected it not to be simple")
	}
}
