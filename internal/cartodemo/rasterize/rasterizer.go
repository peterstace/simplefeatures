package rasterize

import (
	"image"
	"image/draw"

	"github.com/peterstace/simplefeatures/geom"
	"golang.org/x/image/vector"
)

type Rasterizer struct {
	rast *vector.Rasterizer
}

func NewRasterizer(widthPx, heightPx int) *Rasterizer {
	return &Rasterizer{rast: vector.NewRasterizer(widthPx, heightPx)}
}

func (r *Rasterizer) Reset() {
	b := r.rast.Bounds()
	r.rast.Reset(b.Dx(), b.Dy())
}

func (r *Rasterizer) Draw(dst draw.Image, rec image.Rectangle, src image.Image, sp image.Point) {
	r.rast.Draw(dst, rec, src, sp)
}

func (r *Rasterizer) LineString(ls geom.LineString) {
	const strokeWidth = 1.0 // TODO: Make stroke width configurable.

	seq := ls.Coordinates()
	for i := 0; i+1 < seq.Length(); i++ {
		p0 := seq.GetXY(i)
		p1 := seq.GetXY(i + 1)
		if p0 == p1 {
			continue
		}

		// TODO: This naive approach to line drawing is not performant or completely correct.
		mainAxis := p1.Sub(p0)
		sideAxis := rotateCCW90(mainAxis).Scale(0.5 * strokeWidth / mainAxis.Length())

		v0 := p0.Add(sideAxis)
		v1 := p1.Add(sideAxis)
		v2 := p1.Sub(sideAxis)
		v3 := p0.Sub(sideAxis)

		r.rast.MoveTo(float32(v0.X), float32(v0.Y))
		r.rast.LineTo(float32(v1.X), float32(v1.Y))
		r.rast.LineTo(float32(v2.X), float32(v2.Y))
		r.rast.LineTo(float32(v3.X), float32(v3.Y))
		r.rast.ClosePath()
	}
}

func (r *Rasterizer) MultiLineString(mls geom.MultiLineString) {
	for _, ls := range mls.Dump() {
		r.LineString(ls)
	}
}

func (r *Rasterizer) Polygon(p geom.Polygon) {
	// TODO: Support holes.
	ext := p.ExteriorRing()
	seq := ext.Coordinates()
	n := seq.Length()
	if n == 0 {
		return
	}
	r.moveTo(seq.GetXY(0))
	for i := 1; i < n; i++ {
		r.lineTo(seq.GetXY(i))
	}
	r.rast.ClosePath() // Usually not necessary, but just in case.
}

func (r *Rasterizer) MultiPolygon(mp geom.MultiPolygon) {
	for _, p := range mp.Dump() {
		r.Polygon(p)
	}
}

func (r *Rasterizer) moveTo(pt geom.XY) {
	r.rast.MoveTo(float32(pt.X), float32(pt.Y))
}

func (r *Rasterizer) lineTo(pt geom.XY) {
	r.rast.LineTo(float32(pt.X), float32(pt.Y))
}

func rotateCCW90(v geom.XY) geom.XY {
	return geom.XY{-v.Y, v.X}
}
