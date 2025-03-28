package projdemo_test

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/cartodemo/rasterize"
	"github.com/peterstace/simplefeatures/proj"
)

func TestDemo(t *testing.T) {
	glWGS84 := loadGreenlandWGS84(t)
	t.Log(glWGS84.Centroid().AsText())

	tf, err := proj.NewTransformation("EPSG:4326", "EPSG:5938")
	if err != nil {
		t.Fatal(err)
	}

	//glPolarStereo, err := glWGS84.MustAsMultiPolygon().PolygonN(0).ExteriorRing().Transform(tf.Forward)
	glPolarStereo, err := glWGS84.Transform(tf.Forward)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(glPolarStereo.Centroid().AsText())

	img := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	drawGeom(t, img, glPolarStereo)
	//drawGeom(t, img, glWGS84)
	saveImage(t, img, "/tmp/gl.png")
}

func saveImage(t *testing.T, img image.Image, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func drawGeom(t *testing.T, img *image.RGBA, g geom.Geometry) {
	t.Helper()

	gEnv := g.Envelope()
	gEnvCenter, ok := gEnv.Center().XY()
	if !ok {
		t.Fatal("empty envelope")
	}

	// TODO: This is useful enough that it could be its own function that's exposed.
	// There is something very similar in the carto demo.

	pxWide, pxHigh := img.Bounds().Dx(), img.Bounds().Dy()
	imgDims := geom.XY{float64(pxWide), float64(pxHigh)}

	mapUnitsPerPixel := math.Max(
		gEnv.Width()/float64(img.Bounds().Dx()),
		gEnv.Height()/float64(img.Bounds().Dy()),
	)
	mapCoordsToImageCoords := func(mapCoords geom.XY) geom.XY {
		imgCoords := mapCoords.
			Sub(gEnvCenter).
			Scale(1 / mapUnitsPerPixel).
			Add(imgDims.Scale(0.5))
		imgCoords.Y = imgDims.Y - imgCoords.Y
		return imgCoords
	}

	g = g.TransformXY(mapCoordsToImageCoords)
	fmt.Println("DEBUG internal/projdemo/projdemo_test.go:74 g.Envelope().AsGeometry().AsText()", g.Envelope()) // XXX

	r := rasterize.NewRasterizer(img.Bounds().Dx(), img.Bounds().Dy())

	r.Reset()
	r.MultiPolygon(g.MustAsMultiPolygon())
	r.Draw(img, img.Bounds(), image.NewUniform(color.Black), image.Point{})
}

func loadGreenlandWGS84(t *testing.T) geom.Geometry {
	t.Helper()
	// https://simplemaps.com/gis/country/gl#all
	buf, err := os.ReadFile("/home/petsta/downloads/gl.json")
	if err != nil {
		t.Fatal(err)
	}

	var fc geom.GeoJSONFeatureCollection
	if err := json.Unmarshal(buf, &fc); err != nil {
		t.Fatal(err)
	}
	if len(fc) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(fc))
	}
	return fc[0].Geometry
}
