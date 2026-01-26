package jts

import "math"

// Functions for computing length.

// Algorithm_Length_OfLine computes the length of a linestring specified by a
// sequence of points.
func Algorithm_Length_OfLine(pts Geom_CoordinateSequence) float64 {
	// Optimized for processing CoordinateSequences.
	n := pts.Size()
	if n <= 1 {
		return 0.0
	}

	len := 0.0

	p := pts.CreateCoordinate()
	pts.GetCoordinateInto(0, p)
	x0 := p.GetX()
	y0 := p.GetY()

	for i := 1; i < n; i++ {
		pts.GetCoordinateInto(i, p)
		x1 := p.GetX()
		y1 := p.GetY()
		dx := x1 - x0
		dy := y1 - y0

		len += math.Hypot(dx, dy)

		x0 = x1
		y0 = y1
	}
	return len
}
