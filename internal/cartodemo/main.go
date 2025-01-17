package main

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
	"log"
	"math"
	"os"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/cartodemo/rasterize"
)

const (
	pixelBudget = 360 * 180 // Adjusts number of pixels in output images.
)

func main() {
	for name, s := range scenarios {
		log.Printf("building %s", name)
		outputPath := fmt.Sprintf("./internal/cartodemo/output/%s.png", name)
		if err := s.build(outputPath); err != nil {
			log.Fatalf("error: %v", err)
		}
	}
}

type worldProjectionFixture struct {
	proj         func(geom.XY) geom.XY // Convert lon/lat to projected coordinates.
	worldMask    geom.Polygon          // Parts of the world (in lon/lat) to include.
	mapMask      geom.Polygon          // Parts of the map (in projected coordinates) to include.
	mapFlipY     bool                  // True iff the map coordinates increase from top to bottom.
	stdParallels [2]float64
	paddingPx    int
}

func (f *worldProjectionFixture) build(outputPath string) error {
	var (
		waterColor = color.RGBA{R: 144, G: 218, B: 238, A: 255}
		landColor  = color.RGBA{R: 188, G: 236, B: 216, A: 255}
		iceColor   = color.RGBA{R: 252, G: 251, B: 250, A: 255}
	)

	var land, lakes, glaciers, iceshelves geom.Geometry
	for path, ptr := range map[string]*geom.Geometry{
		"./internal/cartodemo/assets/ne_50m_land.geojson.gz":                        &land,
		"./internal/cartodemo/assets/ne_50m_lakes.geojson.gz":                       &lakes,
		"./internal/cartodemo/assets/ne_50m_glaciated_areas.geojson.gz":             &glaciers,
		"./internal/cartodemo/assets/ne_50m_antarctic_ice_shelves_polys.geojson.gz": &iceshelves,
	} {
		var err error
		*ptr, err = loadGeom(path)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", path, err)
		}
	}

	f.worldMask = f.worldMask.Densify(1)
	for _, g := range []*geom.Geometry{&land, &lakes, &glaciers, &iceshelves} {
		clipped, err := geom.Intersection(*g, f.worldMask.AsGeometry())
		*g = clipped
		if err != nil {
			return err
		}
	}

	var lines []geom.LineString
	for lon := -180.0; lon < 180; lon += 30 {
		line := geom.NewLineStringXY(lon, -90, lon, +90)
		line = line.Densify(0.1)
		lines = append(lines, line)
	}
	for lat := -60.0; lat <= 60; lat += 30 {
		line := geom.NewLineStringXY(-180, lat, +180, lat)
		line = line.Densify(0.1)
		lines = append(lines, line)
	}

	var clippedLines []geom.LineString
	for i := range lines {
		clipped, err := geom.Intersection(lines[i].AsGeometry(), f.worldMask.AsGeometry())
		if err != nil {
			return err
		}
		clippedLines = append(clippedLines, extractLinearParts(clipped)...)
	}
	lines = clippedLines

	var redLines []geom.LineString
	for i := range f.stdParallels {
		if f.stdParallels[i] == 0 {
			continue
		}
		line := geom.NewLineStringXY(-180, f.stdParallels[i], +180, f.stdParallels[i])
		line = line.Densify(0.1)
		clipped, err := geom.Intersection(line.AsGeometry(), f.worldMask.AsGeometry())
		if err != nil {
			return err
		}
		redLines = append(redLines, extractLinearParts(clipped)...)
	}

	mapMaskEnv := f.mapMask.Envelope()
	mapMaskRatio := mapMaskEnv.Width() / mapMaskEnv.Height()
	fullnessFactor := f.mapMask.Area() / mapMaskEnv.Area()
	pxHigh := int(math.Round(math.Sqrt(pixelBudget / mapMaskRatio / fullnessFactor)))
	pxWide := int(math.Round(float64(pxHigh) * mapMaskRatio))
	mapMaskCenter, ok := mapMaskEnv.Center().XY()
	if !ok {
		return fmt.Errorf("map mask has no center")
	}

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
		imgCoords = imgCoords.Add(geom.XY{
			X: float64(f.paddingPx),
			Y: float64(f.paddingPx),
		})
		return imgCoords
	}
	lonLatToImgCoords := func(lonLat geom.XY) geom.XY {
		mapCoords := f.proj(lonLat)
		return mapCoordsToImgCoords(mapCoords)
	}

	for _, g := range []*geom.Geometry{&land, &lakes, &glaciers, &iceshelves} {
		*g = g.TransformXY(lonLatToImgCoords)
	}
	for i := range lines {
		lines[i] = lines[i].TransformXY(lonLatToImgCoords)
	}
	for i := range redLines {
		redLines[i] = redLines[i].TransformXY(lonLatToImgCoords)
	}

	imgRect := image.Rect(0, 0, pxWide+2*f.paddingPx, pxHigh+2*f.paddingPx)
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

	for _, line := range lines {
		rast.Reset()
		rast.LineString(line)
		rast.Draw(img, img.Bounds(), image.NewUniform(color.Gray{Y: 0xb0}), image.Point{})
	}

	for _, line := range redLines {
		rast.Reset()
		rast.LineString(line)
		rast.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}), image.Point{})
	}

	rast.Reset()
	mapOutline := mapMaskInImgCoords.Boundary()
	rast.MultiLineString(mapOutline)
	rast.Draw(img, img.Bounds(), image.NewUniform(color.Black), image.Point{})

	return writePNGToFile(img, outputPath)
}

func loadGeom(filename string) (geom.Geometry, error) {
	zippedBuf, err := os.ReadFile(filename)
	if err != nil {
		return geom.Geometry{}, err
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(zippedBuf))
	if err != nil {
		return geom.Geometry{}, err
	}

	unzippedBuf, err := io.ReadAll(gzipReader)
	if err != nil {
		return geom.Geometry{}, err
	}

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
	if err != nil {
		return geom.Geometry{}, err
	}
	var gs []geom.Geometry
	for _, rawFeat := range collection.Features {
		g, err := geom.UnmarshalGeoJSON(rawFeat.Geometry, geom.NoValidate{})
		if err != nil {
			return geom.Geometry{}, err
		}
		if err := g.Validate(); err != nil {
			continue
		}
		gs = append(gs, g)
	}

	all, err := geom.UnionMany(gs)
	if err != nil {
		return geom.Geometry{}, err
	}
	return all, nil
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

func writePNGToFile(img image.Image, path string) error {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
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
