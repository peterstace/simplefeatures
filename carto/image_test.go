package carto_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/peterstace/simplefeatures/carto"
	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/rasterize"
)

func TestWorldProjections(t *testing.T) {
	const (
		earthRadius = 6371000
		earthCircum = 2 * math.Pi * earthRadius
	)
	for i, tc := range []struct {
		name     string
		proj     func(geom.XY) geom.XY
		pxWide   int
		pxHigh   int
		tlXY     geom.XY
		brXY     geom.XY
		mask     geom.Polygon
		filename string
	}{
		{
			name:     "plate carree",
			proj:     func(pt geom.XY) geom.XY { return pt },
			pxWide:   720,
			pxHigh:   360,
			tlXY:     geom.XY{X: -180, Y: +90},
			brXY:     geom.XY{X: +180, Y: -90},
			filename: "plate_carree.png",
		},
		{
			name:     "web mercator",
			proj:     carto.NewWebMercator(0).To,
			pxWide:   512,
			pxHigh:   512,
			tlXY:     geom.XY{},
			brXY:     geom.XY{X: 1, Y: 1},
			filename: "web_mercator.png",
		},
		{
			name:     "lambert cylindrical equal area",
			proj:     carto.NewLambertCylindricalEqualArea(1, 0).To,
			pxWide:   902, // h*pi
			pxHigh:   287, // h
			tlXY:     geom.XY{X: -math.Pi, Y: +1.0},
			brXY:     geom.XY{X: +math.Pi, Y: -1.0},
			filename: "lambert_cylindrical_equal_area.png",
		},
		{
			name:     "sinusoidal",
			proj:     carto.NewSinusoidal(earthRadius, 0).To,
			pxWide:   720,
			pxHigh:   360,
			tlXY:     geom.XY{X: -0.5 * earthCircum, Y: +0.25 * earthCircum},
			brXY:     geom.XY{X: +0.5 * earthCircum, Y: -0.25 * earthCircum},
			filename: "sinusoidal.png",
			mask: func() geom.Polygon {
				// TODO: Consider a programmatic "mask" image instead.
				var lhs, rhs []float64
				for i := 0; i <= 360; i++ {
					lhs = append(lhs, 360+math.Sin(float64(i)*math.Pi/360)*360, float64(i))
					rhs = append(rhs, 360-math.Sin(float64(i)*math.Pi/360)*360, 360-float64(i))
				}
				all := append(lhs, rhs...)
				seq := geom.NewSequence(all, geom.DimXY)
				ring := geom.NewLineString(seq)
				poly := geom.NewPolygon([]geom.LineString{ring})
				return poly
			}(),
		},
		{
			name:     "equirectangular",
			proj:     (&carto.Equirectangular{Radius: earthRadius}).To,
			pxWide:   720,
			pxHigh:   360,
			tlXY:     geom.XY{X: -0.5 * earthCircum, Y: +0.25 * earthCircum},
			brXY:     geom.XY{X: +0.5 * earthCircum, Y: -0.25 * earthCircum},
			filename: "equirectangular.png",
		},
		{
			name:     "marinus",
			proj:     (&carto.Equirectangular{Radius: earthRadius, StandardParallels: 36}).To,
			pxWide:   int(math.Round(720 * math.Cos(36*math.Pi/180))),
			pxHigh:   360,
			tlXY:     geom.XY{X: -0.5 * earthCircum * math.Cos(36*math.Pi/180), Y: +0.25 * earthCircum},
			brXY:     geom.XY{X: +0.5 * earthCircum * math.Cos(36*math.Pi/180), Y: -0.25 * earthCircum},
			filename: "marinus.png",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join("./testdata", fmt.Sprintf("%d_%s", i, tc.filename))
			worldMask := geom.NewSingleRingPolygonXY(-180, 90, 180, 90, 180, -90, -180, -90, -180, 90)
			writeProjectedWorld(t, path, tc.proj, tc.pxWide, tc.pxHigh, tc.tlXY, tc.brXY, tc.mask, worldMask)
		})
	}

	t.Run("orthographic from south pole", func(t *testing.T) {
		worldMask := geom.NewSingleRingPolygonXY(-180, 0, 180, 0, 180, -90, -180, -90, -180, 0)
		path := filepath.Join("./testdata", "orthographic_south_pole.png")
		writeProjectedWorld(
			t,
			path,
			carto.NewOrthographic(earthRadius, geom.XY{X: 135, Y: -90}).To,
			512, 512,
			geom.XY{X: -earthRadius, Y: +earthRadius},
			geom.XY{X: +earthRadius, Y: -earthRadius},
			circle(geom.XY{X: 256, Y: 256}, 256),
			worldMask,
		)
	})

	t.Run("orthographic from north pole", func(t *testing.T) {
		worldMask := geom.NewSingleRingPolygonXY(-180, 0, 180, 0, 180, 90, -180, 90, -180, 0)
		path := filepath.Join("./testdata", "orthographic_north_pole.png")
		writeProjectedWorld(
			t,
			path,
			carto.NewOrthographic(earthRadius, geom.XY{X: 15, Y: 90}).To,
			512, 512,
			geom.XY{X: -earthRadius, Y: +earthRadius},
			geom.XY{X: +earthRadius, Y: -earthRadius},
			circle(geom.XY{X: 256, Y: 256}, 256),
			worldMask,
		)
	})

	t.Run("orthographic from north america", func(t *testing.T) {
		const centralMeridian = -105
		var worldMaskCoords []float64
		latAtLon := func(lon float64) float64 {
			return 45 * math.Cos((285+lon)*math.Pi/180)
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

		path := filepath.Join("./testdata", "orthographic_north_america.png")
		writeProjectedWorld(
			t,
			path,
			carto.NewOrthographic(earthRadius, geom.XY{X: centralMeridian, Y: 45}).To,
			512, 512,
			geom.XY{X: -earthRadius, Y: +earthRadius},
			geom.XY{X: +earthRadius, Y: -earthRadius},
			circle(geom.XY{X: 256, Y: 256}, 256),
			worldMask,
		)
	})
}

func writeProjectedWorld(
	t *testing.T,
	outputFilename string,
	projection func(geom.XY) geom.XY,
	pxWide int,
	pxHigh int,
	tlXY geom.XY,
	brXY geom.XY,
	mapMask geom.Polygon,
	worldMask geom.Polygon,
) {
	var (
		waterColor = color.RGBA{R: 144, G: 218, B: 238, A: 255}
		landColor  = color.RGBA{R: 188, G: 236, B: 216, A: 255}
		iceColor   = color.RGBA{R: 252, G: 251, B: 250, A: 255}
	)

	land := loadGeom(t, "testdata/ne_50m_land.geojson.gz")
	lakes := loadGeom(t, "testdata/ne_50m_lakes.geojson.gz")
	glaciers := loadGeom(t, "testdata/ne_50m_glaciated_areas.geojson.gz")
	iceshelves := loadGeom(t, "testdata/ne_50m_antarctic_ice_shelves_polys.geojson.gz")

	worldMask = worldMask.Densify(1)
	for _, g := range []*geom.Geometry{&land, &lakes, &glaciers, &iceshelves} {
		clipped, err := geom.Intersection(*g, worldMask.AsGeometry())
		*g = clipped
		expectNoErr(t, err)
	}

	var lines []geom.LineString
	for lon := -180; lon <= 180; lon += 30 {
		line := geom.NewLineString(geom.NewSequence([]float64{float64(lon), -90, float64(lon), +90}, geom.DimXY))
		line = line.Densify(1)
		lines = append(lines, line)
	}
	for lat := -90; lat <= 90; lat += 30 {
		line := geom.NewLineString(geom.NewSequence([]float64{-180, float64(lat), +180, float64(lat)}, geom.DimXY))
		line = line.Densify(1)
		lines = append(lines, line)
	}

	var clippedLines []geom.LineString
	for i := range lines {
		clipped, err := geom.Intersection(lines[i].AsGeometry(), worldMask.AsGeometry())
		clippedLines = append(clippedLines, extractLinearParts(clipped)...)
		expectNoErr(t, err)
	}
	lines = clippedLines

	for _, g := range []*geom.Geometry{&land, &lakes, &glaciers, &iceshelves} {
		*g = g.TransformXY(func(latlon geom.XY) geom.XY {
			// Project from lat/lon to map coordinates:
			xy := projection(latlon)

			// Project from map coordinates to image coordinates:
			return geom.XY{
				X: linearRemap(tlXY.X, 0, brXY.X, float64(pxWide))(xy.X),
				Y: linearRemap(tlXY.Y, 0, brXY.Y, float64(pxHigh))(xy.Y),
			}
		})
	}

	// TODO: Remove duplication.
	for i := range lines {
		lines[i] = lines[i].TransformXY(func(latlon geom.XY) geom.XY {
			// Project from lat/lon to map coordinates:
			xy := projection(latlon)

			// Project from map coordinates to image coordinates:
			return geom.XY{
				X: linearRemap(tlXY.X, 0, brXY.X, float64(pxWide))(xy.X),
				Y: linearRemap(tlXY.Y, 0, brXY.Y, float64(pxHigh))(xy.Y),
			}
		})
	}

	rast := rasterize.NewRasterizer(pxWide, pxHigh)

	mapMaskImage := image.NewAlpha(image.Rect(0, 0, pxWide, pxHigh))
	var mapOutline geom.MultiLineString
	if mapMask.IsEmpty() {
		draw.Draw(mapMaskImage, mapMaskImage.Bounds(), image.NewUniform(color.Opaque), image.Point{}, draw.Src)
	} else {
		mapMaskImage = image.NewAlpha(image.Rect(0, 0, pxWide, pxHigh))
		rast.Reset()
		rast.Polygon(mapMask)
		rast.Draw(mapMaskImage, mapMaskImage.Bounds(), image.NewUniform(color.Opaque), image.Point{})
		mapOutline = mapMask.Boundary()
	}

	img := image.NewRGBA(image.Rect(0, 0, pxWide, pxHigh))
	draw.DrawMask(img, img.Bounds(), image.NewUniform(waterColor), image.Point{}, mapMaskImage, image.Point{}, draw.Src)

	rast.Reset()
	rast.MultiLineString(mapOutline)
	rast.Draw(img, img.Bounds(), image.NewUniform(color.Black), image.Point{})

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

	rast.Reset()
	for _, line := range lines {
		rast.LineString(line)
	}
	rast.Draw(img, img.Bounds(), image.NewUniform(color.Gray{Y: 0x80}), image.Point{})

	err := os.WriteFile(outputFilename, imageToPNG(t, img), 0644)
	expectNoErr(t, err)
}

func imageToPNG(t *testing.T, img image.Image) []byte {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	expectNoErr(t, err)
	return buf.Bytes()
}

func loadGeom(t *testing.T, filename string) geom.Geometry {
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

func linearRemap(fromA, toA, fromB, toB float64) func(float64) float64 {
	fromDelta := fromB - fromA
	toDelta := toB - toA
	return func(f float64) float64 {
		t := (f - fromA) / fromDelta
		return toA + t*toDelta
	}
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

func circle(center geom.XY, radius float64) geom.Polygon {
	var coords []float64
	for i := 0.0; i <= 360; i++ {
		coords = append(coords,
			center.X+radius*math.Cos(i*math.Pi/180),
			center.Y+radius*math.Sin(i*math.Pi/180),
		)
	}
	return geom.NewPolygonXY(coords)
}
