package jts

import "math"

// Functions for computing area.

// Algorithm_Area_OfRing computes the area for a ring.
func Algorithm_Area_OfRing(ring []*Geom_Coordinate) float64 {
	return math.Abs(Algorithm_Area_OfRingSigned(ring))
}

// Algorithm_Area_OfRingSeq computes the area for a ring.
func Algorithm_Area_OfRingSeq(ring Geom_CoordinateSequence) float64 {
	return math.Abs(Algorithm_Area_OfRingSignedSeq(ring))
}

// Algorithm_Area_OfRingSigned computes the signed area for a ring. The signed area is
// positive if the ring is oriented CW, negative if the ring is oriented CCW,
// and zero if the ring is degenerate or flat.
func Algorithm_Area_OfRingSigned(ring []*Geom_Coordinate) float64 {
	if len(ring) < 3 {
		return 0.0
	}
	sum := 0.0
	x0 := ring[0].GetX()
	for i := 1; i < len(ring)-1; i++ {
		x := ring[i].GetX() - x0
		y1 := ring[i+1].GetY()
		y2 := ring[i-1].GetY()
		sum += x * (y2 - y1)
	}
	return sum / 2.0
}

// Algorithm_Area_OfRingSignedSeq computes the signed area for a ring. The signed area is:
//   - positive if the ring is oriented CW
//   - negative if the ring is oriented CCW
//   - zero if the ring is degenerate or flat
func Algorithm_Area_OfRingSignedSeq(ring Geom_CoordinateSequence) float64 {
	n := ring.Size()
	if n < 3 {
		return 0.0
	}
	p0 := ring.CreateCoordinate()
	p1 := ring.CreateCoordinate()
	p2 := ring.CreateCoordinate()
	ring.GetCoordinateInto(0, p1)
	ring.GetCoordinateInto(1, p2)
	x0 := p1.GetX()
	p2.SetX(p2.GetX() - x0)
	sum := 0.0
	for i := 1; i < n-1; i++ {
		p0.SetY(p1.GetY())
		p1.SetX(p2.GetX())
		p1.SetY(p2.GetY())
		ring.GetCoordinateInto(i+1, p2)
		p2.SetX(p2.GetX() - x0)
		sum += p1.GetX() * (p0.GetY() - p2.GetY())
	}
	return sum / 2.0
}
