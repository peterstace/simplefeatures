// Package cartodemo_test contains demo code for the
// github.com/peterstace/simplefeatures/carto package.
//
// It's provided as a set of executable tests.
package cartodemo_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"testing"

	"github.com/peterstace/simplefeatures/carto"
	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/cartodemo/rasterize"
)

const (
	pxWide      = 512
	pxPadding   = 2
	earthRadius = 6371000
	earthCircum = 2 * pi * earthRadius
	pi          = math.Pi
)

var fullWorldMask = rectangle(
	xy(-180, +90),
	xy(+180, -90),
)

func xy(x, y float64) geom.XY {
	return geom.XY{X: x, Y: y}
}

func TestDrawMapEquirectangularMarinus(t *testing.T) {
	p := carto.NewEquirectangular(earthRadius)
	p.SetStandardParallels(36)
	cos36 := math.Cos(36 * pi / 180)
	f := &worldProjectionFixture{
		proj:      p.Forward,
		worldMask: fullWorldMask,
		mapMask: rectangle(
			xy(-0.5*earthCircum*cos36, +0.25*earthCircum),
			xy(+0.5*earthCircum*cos36, -0.25*earthCircum),
		),
	}
	f.build(t, "testdata/marinus.png")
}

func TestDrawMapWebMercator(t *testing.T) {
	f := &worldProjectionFixture{
		proj:      carto.NewWebMercator(0).Forward,
		worldMask: fullWorldMask,
		mapMask:   rectangle(xy(0, 0), xy(1, 1)),
		mapFlipY:  true,
	}
	f.build(t, "testdata/web_mercator.png")
}

func TestDrawMapLambertCylindricalEqualArea(t *testing.T) {
	f := &worldProjectionFixture{
		proj:      carto.NewLambertCylindricalEqualArea(earthRadius).Forward,
		worldMask: fullWorldMask,
		mapMask: rectangle(
			xy(-0.5*earthCircum, +0.25*earthCircum*2/pi),
			xy(+0.5*earthCircum, -0.25*earthCircum*2/pi),
		),
	}
	f.build(t, "testdata/lambert_cylindrical_equal_area.png")
}

func TestDrawMapSinusoidal(t *testing.T) {
	var lhs, rhs []float64
	for lat := -90.0; lat <= 90; lat++ {
		lhs = append(lhs,
			-0.5*earthCircum*math.Cos(lat*pi/180),
			lat/90*0.25*earthCircum,
		)
		rhs = append(rhs,
			+0.5*earthCircum*math.Cos(lat*pi/180),
			-lat/90*0.25*earthCircum,
		)
	}
	all := append(lhs, rhs...) //nolint:gocritic
	mapMask := geom.NewPolygonXY(all)

	f := &worldProjectionFixture{
		proj:      carto.NewSinusoidal(earthRadius).Forward,
		worldMask: fullWorldMask,
		mapMask:   mapMask,
	}
	f.build(t, "testdata/sinusoidal.png")
}

func TestDrawMapOrthographicNorthAmerica(t *testing.T) {
	const centralMeridian = -105
	var worldMaskCoords []float64
	latAtLon := func(lon float64) float64 {
		return 45 * math.Cos((285+lon)*pi/180)
	}
	for lon := -180.0; lon <= 180.0; lon++ {
		lat := latAtLon(lon)
		worldMaskCoords = append(worldMaskCoords, lon, lat)
	}
	worldMaskCoords = append(
		worldMaskCoords,
		180, latAtLon(180),
		180, 90,
		-180, 90,
		-180, latAtLon(-180),
	)
	worldMask := geom.NewPolygonXY(worldMaskCoords)

	proj := carto.NewOrthographic(earthRadius)
	proj.SetCenter(geom.XY{X: centralMeridian, Y: 45})

	f := &worldProjectionFixture{
		proj:      proj.Forward,
		worldMask: worldMask,
		mapMask:   circle(xy(0, 0), earthRadius),
	}
	f.build(t, "testdata/orthographic_north_america.png")
}

func TestDrawMapAzimuthalEquidistantSydney(t *testing.T) {
	p := carto.NewAzimuthalEquidistant(earthRadius)
	p.SetCenter(geom.XY{X: 151, Y: -34})
	f := &worldProjectionFixture{
		proj:      p.Forward,
		worldMask: fullWorldMask,
		mapMask:   circle(xy(0, 0), earthCircum/2),
	}
	f.build(t, "testdata/azimuthal_equidistant_sydney.png")
}

const (
	stdParallel1 = 30
	stdParallel2 = 60
)

func TestDrawEquidistantConic(t *testing.T) {
	p := carto.NewEquidistantConic(earthRadius)
	p.SetStandardParallels(stdParallel1, stdParallel2)

	const eps = 0.1
	mapMask := rectangle(
		xy(-180+eps, -90+eps),
		xy(+180-eps, +90-eps),
	)
	mapMask = mapMask.Densify(0.1)
	mapMask = mapMask.TransformXY(p.Forward)

	f := &worldProjectionFixture{
		proj:      p.Forward,
		worldMask: fullWorldMask,
		mapMask:   mapMask,
	}
	f.build(t, "testdata/equidistant_conic.png")
}

func TestDrawLambertConformalConic(t *testing.T) {
	p := carto.NewLambertConformalConic(earthRadius)
	p.SetStandardParallels(stdParallel1, stdParallel2)

	const (
		minLat = -45
		maxLat = 90
	)
	worldMask := rectangle(
		xy(-180, minLat),
		xy(+180, maxLat),
	)
	mapMask := worldMask.Densify(0.1).TransformXY(p.Forward)

	f := &worldProjectionFixture{
		proj:      p.Forward,
		worldMask: worldMask,
		mapMask:   mapMask,
	}
	f.build(t, "testdata/lambert_conformal_conic.png")
}

func TestDrawAlbersEqualAreaConic(t *testing.T) {
	p := carto.NewAlbersEqualAreaConic(earthRadius)
	p.SetStandardParallels(stdParallel1, stdParallel2)

	worldMask := rectangle(
		xy(-180, -90),
		xy(+180, +90),
	)
	mapMask := worldMask.Densify(0.1).TransformXY(p.Forward)

	f := &worldProjectionFixture{
		proj:      p.Forward,
		worldMask: worldMask,
		mapMask:   mapMask,
	}
	f.build(t, "testdata/albers_equal_area_conic.png")
}

type worldProjectionFixture struct {
	proj      func(geom.XY) geom.XY // Convert lon/lat to projected coordinates.
	worldMask geom.Polygon          // Parts of the world (in lon/lat) to include.
	mapMask   geom.Polygon          // Parts of the map (in projected coordinates) to include.
	mapFlipY  bool                  // True iff the map coordinates increase from top to bottom.
}

func (f *worldProjectionFixture) build(t *testing.T, outputPath string) {
	t.Helper()
	var (
		waterColor = color.RGBA{R: 144, G: 218, B: 238, A: 255}
		landColor  = color.RGBA{R: 188, G: 236, B: 216, A: 255}
		iceColor   = color.RGBA{R: 252, G: 251, B: 250, A: 255}

		land       = loadGeom(t, "testdata/ne_50m_land.geojson.gz")
		lakes      = loadGeom(t, "testdata/ne_50m_lakes.geojson.gz")
		glaciers   = loadGeom(t, "testdata/ne_50m_glaciated_areas.geojson.gz")
		iceshelves = loadGeom(t, "testdata/ne_50m_antarctic_ice_shelves_polys.geojson.gz")
	)

	f.worldMask = f.worldMask.Densify(1)
	for _, g := range []*geom.Geometry{&land, &lakes, &glaciers, &iceshelves} {
		clipped, err := geom.Intersection(*g, f.worldMask.AsGeometry())
		*g = clipped
		expectNoErr(t, err)
	}

	var graticules []geom.LineString
	for lon := -180.0; lon < 180; lon += 30 {
		grat := geom.NewLineStringXY(lon, -90, lon, +90)
		grat = grat.Densify(0.1)
		graticules = append(graticules, grat)
	}
	for lat := -60.0; lat <= 60; lat += 30 {
		grat := geom.NewLineStringXY(-180, lat, +180, lat)
		grat = grat.Densify(0.1)
		graticules = append(graticules, grat)
	}

	var clippedGraticules []geom.LineString
	for i := range graticules {
		clipped, err := geom.Intersection(graticules[i].AsGeometry(), f.worldMask.AsGeometry())
		clippedGraticules = append(clippedGraticules, extractLinearParts(clipped)...)
		expectNoErr(t, err)
	}
	graticules = clippedGraticules

	mapMaskEnv := f.mapMask.Envelope()
	mapMaskRatio := mapMaskEnv.Width() / mapMaskEnv.Height()
	pxHigh := int(pxWide / mapMaskRatio)
	mapMaskCenter, ok := mapMaskEnv.Center().XY()
	expectTrue(t, ok)

	mapUnitsPerPixel := f.mapMask.Envelope().Width() / float64(pxWide)

	imgDims := geom.XY{X: float64(pxWide), Y: float64(pxHigh)}
	mapCoordsToImgCoords := func(mapCoords geom.XY) geom.XY {
		imgCoords := mapCoords.
			Sub(mapMaskCenter).
			Scale(1 / mapUnitsPerPixel).
			Add(imgDims.Scale(0.5))
		if !f.mapFlipY {
			imgCoords.Y = imgDims.Y - imgCoords.Y
		}
		imgCoords = imgCoords.Add(geom.XY{X: pxPadding, Y: pxPadding})
		return imgCoords
	}
	lonLatToImgCoords := func(lonLat geom.XY) geom.XY {
		mapCoords := f.proj(lonLat)
		return mapCoordsToImgCoords(mapCoords)
	}

	for _, g := range []*geom.Geometry{&land, &lakes, &glaciers, &iceshelves} {
		*g = g.TransformXY(lonLatToImgCoords)
	}
	for i := range graticules {
		graticules[i] = graticules[i].TransformXY(lonLatToImgCoords)
	}

	imgRect := image.Rect(0, 0, pxWide+2*pxPadding, pxHigh+2*pxPadding)
	rast := rasterize.NewRasterizer(imgRect.Dx(), imgRect.Dy())

	mapMaskImage := image.NewAlpha(imgRect)
	rast.Reset()
	mapMaskInImgCoords := f.mapMask.TransformXY(mapCoordsToImgCoords)
	rast.Polygon(mapMaskInImgCoords)
	rast.Draw(mapMaskImage, mapMaskImage.Bounds(), image.NewUniform(color.Opaque), image.Point{})

	img := image.NewRGBA(imgRect)
	draw.DrawMask(img, img.Bounds(), image.NewUniform(waterColor), image.Point{}, mapMaskImage, image.Point{}, draw.Src)

	rasterisePolygons := func(g geom.Geometry) {
		for _, p := range extractPolygonalParts(g) {
			rast.Polygon(p)
		}
	}

	rast.Reset()
	rasterisePolygons(land)
	rast.Draw(img, img.Bounds(), image.NewUniform(landColor), image.Point{})

	rast.Reset()
	rasterisePolygons(lakes)
	rast.Draw(img, img.Bounds(), image.NewUniform(waterColor), image.Point{})

	rast.Reset()
	rasterisePolygons(glaciers)
	rasterisePolygons(iceshelves)
	rast.Draw(img, img.Bounds(), image.NewUniform(iceColor), image.Point{})

	for _, line := range graticules {
		rast.Reset()
		rast.LineString(line)
		rast.Draw(img, img.Bounds(), image.NewUniform(color.Gray{Y: 0xb0}), image.Point{})
	}

	rast.Reset()
	mapOutline := mapMaskInImgCoords.Boundary()
	rast.MultiLineString(mapOutline)
	rast.Draw(img, img.Bounds(), image.NewUniform(color.Black), image.Point{})

	err := os.WriteFile(outputPath, imageToPNG(t, img), 0o600)
	expectNoErr(t, err)
}

func imageToPNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	expectNoErr(t, err)
	return buf.Bytes()
}

func loadGeom(t *testing.T, filename string) geom.Geometry {
	t.Helper()
	zippedBuf, err := os.ReadFile(filename)
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

func extractPolygonalParts(g geom.Geometry) []geom.Polygon {
	switch gt := g.Type(); gt {
	case geom.TypeGeometryCollection:
		var ps []geom.Polygon
		for _, g := range g.Dump() {
			ps = append(ps, extractPolygonalParts(g)...)
		}
		return ps
	case geom.TypeMultiPolygon:
		return g.MustAsMultiPolygon().Dump()
	case geom.TypePolygon:
		return []geom.Polygon{g.MustAsPolygon()}
	default:
		return nil
	}
}

func extractLinearParts(g geom.Geometry) []geom.LineString {
	switch gt := g.Type(); gt {
	case geom.TypeGeometryCollection:
		var lss []geom.LineString
		for _, g := range g.Dump() {
			lss = append(lss, extractLinearParts(g)...)
		}
		return lss
	case geom.TypeMultiLineString:
		return g.MustAsMultiLineString().Dump()
	case geom.TypeLineString:
		return []geom.LineString{g.MustAsLineString()}
	default:
		return nil
	}
}

func circle(c geom.XY, r float64) geom.Polygon {
	var coords []float64
	for i := 0.0; i <= 360; i++ {
		coords = append(coords,
			c.X+r*math.Cos(i*pi/180),
			c.Y+r*math.Sin(i*pi/180),
		)
	}
	return geom.NewPolygonXY(coords)
}

func rectangle(tl, br geom.XY) geom.Polygon {
	return geom.NewEnvelope(tl, br).AsGeometry().MustAsPolygon()
}

func expectNoErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}

func expectTrue(tb testing.TB, b bool) {
	tb.Helper()
	if !b {
		tb.Fatalf("expected true, got false")
	}
}
