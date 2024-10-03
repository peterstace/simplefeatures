package rasterize_test

import (
	"image"
	"os"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/rasterize"
)

func TestDrawLine(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	g, err := geom.UnmarshalWKT("LINESTRING(4 4, 12 8, 4 12)")
	expectNoErr(t, err)
	rasterize.LineString(img, g.MustAsLineString())
	err = os.WriteFile("testdata/line.png", imageToPNG(t, img), 0644)
	expectNoErr(t, err)
}
