package jts

import "math"

// Noding_Octant provides methods for computing and working with octants of the
// Cartesian plane. Octants are numbered as follows:
//
//	 \2|1/
//	3 \|/ 0
//	---+--
//	4 /|\ 7
//	 /5|6\
//
// If line segments lie along a coordinate axis, the octant is the lower of the
// two possible values.

// Noding_Octant_OctantFromDxDy returns the octant of a directed line segment
// (specified as x and y displacements, which cannot both be 0).
func Noding_Octant_OctantFromDxDy(dx, dy float64) int {
	if dx == 0.0 && dy == 0.0 {
		panic("Cannot compute the octant for point (0, 0)")
	}

	adx := math.Abs(dx)
	ady := math.Abs(dy)

	if dx >= 0 {
		if dy >= 0 {
			if adx >= ady {
				return 0
			}
			return 1
		}
		// dy < 0
		if adx >= ady {
			return 7
		}
		return 6
	}
	// dx < 0
	if dy >= 0 {
		if adx >= ady {
			return 3
		}
		return 2
	}
	// dy < 0
	if adx >= ady {
		return 4
	}
	return 5
}

// Noding_Octant_Octant returns the octant of a directed line segment from p0 to
// p1.
func Noding_Octant_Octant(p0, p1 *Geom_Coordinate) int {
	dx := p1.GetX() - p0.GetX()
	dy := p1.GetY() - p0.GetY()
	if dx == 0.0 && dy == 0.0 {
		panic("Cannot compute the octant for two identical points")
	}
	return Noding_Octant_OctantFromDxDy(dx, dy)
}
