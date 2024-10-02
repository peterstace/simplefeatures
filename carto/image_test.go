package carto_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestImage(t *testing.T) {
	land := loadLandGeom(t)
	land = land.TransformXY(func(in geom.XY) geom.XY {
		in.X += 180      // -180..180 -> 0..360
		in.Y = 90 - in.Y // 90..-90 -> 0..180
		return in
	})
	t.Log(land.Envelope())

	landBoundary := land.Boundary()
	mls, ok := landBoundary.AsMultiLineString()
	if !ok {
		t.Fatalf("expected MultiLineString, got %v", landBoundary.Type())
	}

	img := image.NewRGBA(image.Rect(0, 0, 360, 180))
	drawLines(t, img, mls)

	err := os.WriteFile("testdata/land.png", imageToPNG(t, img), 0644)
	expectNoErr(t, err)
}

func imageToPNG(t *testing.T, img *image.RGBA) []byte {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	expectNoErr(t, err)
	return buf.Bytes()
}

func drawLines(t *testing.T, img *image.RGBA, mls geom.MultiLineString) {
	for _, ls := range mls.Dump() {
		fmt.Println("DEBUG carto/image_test.go:48 ls.Envelope()", ls.Envelope()) // XXX
	}
}

func loadLandGeom(t *testing.T) geom.Geometry {
	zippedBuf, err := os.ReadFile("testdata/ne_110m_land.geojson.gz")
	expectNoErr(t, err)

	gzipReader, err := gzip.NewReader(bytes.NewReader(zippedBuf))
	expectNoErr(t, err)

	unzippedBuf, err := io.ReadAll(gzipReader)
	expectNoErr(t, err)

	// TODO: There is currently no way to disable a GeoJSON GeometryCollection
	// without validation directly. See
	// https://github.com/peterstace/simplefeatures/issues/638. For now, we
	// unmarshal the GeoJSON FeatureCollection "manually" to avoid validation.
	var collection struct {
		Features []struct {
			Geometry json.RawMessage `json:"geometry"`
		} `json:"features"`
	}
	err = json.Unmarshal(unzippedBuf, &collection)
	expectNoErr(t, err)
	var gs []geom.Geometry
	for _, rawFeat := range collection.Features {
		g, err := geom.UnmarshalGeoJSON(rawFeat.Geometry, geom.NoValidate{})
		expectNoErr(t, err)
		if err := g.Validate(); err != nil {
			continue
		}
		gs = append(gs, g)
	}

	all, err := geom.UnionMany(gs)
	expectNoErr(t, err)

	return all
}
