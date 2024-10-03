package carto_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"golang.org/x/image/vector"
)

func TestImage(t *testing.T) {
	const (
		pxWide = 360 * 2
		pxHigh = 180 * 2
	)
	land := loadLandGeom(t)
	land = land.TransformXY(func(in geom.XY) geom.XY {
		// TODO: Consider adding a "Linear" transformation to do this.
		in.X += 180                   // -180..180 -> 0..360
		in.X *= float64(pxWide) / 360 // 0..360 -> 0..pxWide
		in.Y = 90 - in.Y              // 90..-90 -> 0..180
		in.Y *= float64(pxHigh) / 180 // 0..180 -> 0..pxHigh
		return in
	})
	t.Log(land.Envelope())

	landBoundary := land.Boundary()
	mls, ok := landBoundary.AsMultiLineString()
	if !ok {
		t.Fatalf("expected MultiLineString, got %v", landBoundary.Type())
	}

	img := image.NewRGBA(image.Rect(0, 0, pxWide, pxHigh))
	drawMultiLineString(t, img, mls)

	err := os.WriteFile("testdata/land.png", imageToPNG(t, img), 0644)
	expectNoErr(t, err)
}

func imageToPNG(t *testing.T, img *image.RGBA) []byte {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	expectNoErr(t, err)
	return buf.Bytes()
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

func TestDrawLine(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	g, err := geom.UnmarshalWKT("LINESTRING(4 4, 12 8, 4 12)")
	expectNoErr(t, err)
	drawLineString(t, img, g.MustAsLineString())
	err = os.WriteFile("testdata/line.png", imageToPNG(t, img), 0644)
	expectNoErr(t, err)
}

func drawMultiLineString(t *testing.T, img *image.RGBA, mls geom.MultiLineString) {
	for _, ls := range mls.Dump() {
		drawLineString(t, img, ls)
	}
}

func drawLineString(t *testing.T, img *image.RGBA, ls geom.LineString) {
	const strokeWidth = 1                     // TODO: Make stroke width configurable.
	blackImg := image.NewUniform(color.Black) // TODO: Make color configurable.

	rast := vector.NewRasterizer(img.Rect.Dx(), img.Rect.Dy())
	seq := ls.Coordinates()

	for i := 0; i+1 < seq.Length(); i++ {
		p0 := seq.GetXY(i)
		p1 := seq.GetXY(i + 1)
		if p0 == p1 {
			continue
		}

		// TODO: This is a pretty basic/stupid way to draw a line. Consider
		// something that is both more accurate and faster. We should be able
		// to do one pass per line string, rather than one pass per line
		// segment.
		mainAxis := p1.Sub(p0)
		sideAxis := rotateCCW90(mainAxis).Scale(0.5 * strokeWidth / mainAxis.Length())

		v0 := p0.Add(sideAxis)
		v1 := p1.Add(sideAxis)
		v2 := p1.Sub(sideAxis)
		v3 := p0.Sub(sideAxis)

		rast.MoveTo(float32(v0.X), float32(v0.Y))
		rast.LineTo(float32(v1.X), float32(v1.Y))
		rast.LineTo(float32(v2.X), float32(v2.Y))
		rast.LineTo(float32(v3.X), float32(v3.Y))
		rast.LineTo(float32(v0.X), float32(v0.Y))

		rast.Draw(img, img.Bounds(), blackImg, image.Point{})
		rast.Reset(img.Rect.Dx(), img.Rect.Dy())
	}
}

// TODO: This is duplicated from geom/xy.go. Could be a better solution?
func rotateCCW90(v geom.XY) geom.XY {
	return geom.XY{-v.Y, v.X}
}
