package rasterize

import (
	"image"
	"image/color"

	"github.com/peterstace/simplefeatures/geom"
	"golang.org/x/image/vector"
)

func MultiLineString(img *image.RGBA, mls geom.MultiLineString) {
	for _, ls := range mls.Dump() {
		LineString(img, ls)
	}
}

func LineString(img *image.RGBA, ls geom.LineString) {
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
