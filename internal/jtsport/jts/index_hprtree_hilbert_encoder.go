package jts

import "math"

// IndexHprtree_HilbertEncoder encodes envelopes as Hilbert codes for spatial
// indexing.
type IndexHprtree_HilbertEncoder struct {
	level   int
	minx    float64
	miny    float64
	strideX float64
	strideY float64
}

// IndexHprtree_NewHilbertEncoder creates a new HilbertEncoder for the given
// level and extent.
func IndexHprtree_NewHilbertEncoder(level int, extent *Geom_Envelope) *IndexHprtree_HilbertEncoder {
	hside := int(math.Pow(2, float64(level))) - 1

	minx := extent.GetMinX()
	strideX := extent.GetWidth() / float64(hside)

	miny := extent.GetMinY()
	strideY := extent.GetHeight() / float64(hside)

	return &IndexHprtree_HilbertEncoder{
		level:   level,
		minx:    minx,
		miny:    miny,
		strideX: strideX,
		strideY: strideY,
	}
}

// Encode encodes the given envelope as a Hilbert code.
func (e *IndexHprtree_HilbertEncoder) Encode(env *Geom_Envelope) int {
	midx := env.GetWidth()/2 + env.GetMinX()
	x := int((midx - e.minx) / e.strideX)

	midy := env.GetHeight()/2 + env.GetMinY()
	y := int((midy - e.miny) / e.strideY)

	return ShapeFractal_HilbertCode_Encode(e.level, x, y)
}
