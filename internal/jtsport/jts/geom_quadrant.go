package jts

import "fmt"

// Quadrant constants reference and number quadrants as follows:
//
//	1 - NW | 0 - NE
//	-------+-------
//	2 - SW | 3 - SE
const Geom_Quadrant_NE = 0
const Geom_Quadrant_NW = 1
const Geom_Quadrant_SW = 2
const Geom_Quadrant_SE = 3

// Geom_Quadrant_QuadrantFromDeltas returns the quadrant of a directed line segment
// (specified as x and y displacements, which cannot both be 0).
//
// Panics if the displacements are both 0.
func Geom_Quadrant_QuadrantFromDeltas(dx, dy float64) int {
	if dx == 0.0 && dy == 0.0 {
		panic(fmt.Sprintf("cannot compute the quadrant for point ( %v, %v )", dx, dy))
	}
	if dx >= 0.0 {
		if dy >= 0.0 {
			return Geom_Quadrant_NE
		}
		return Geom_Quadrant_SE
	}
	if dy >= 0.0 {
		return Geom_Quadrant_NW
	}
	return Geom_Quadrant_SW
}

// Geom_Quadrant_QuadrantFromCoords returns the quadrant of a directed line segment from p0
// to p1.
//
// Panics if the points are equal.
func Geom_Quadrant_QuadrantFromCoords(p0, p1 *Geom_Coordinate) int {
	if p1.X == p0.X && p1.Y == p0.Y {
		panic(fmt.Sprintf("cannot compute the quadrant for two identical points %v", p0))
	}
	if p1.X >= p0.X {
		if p1.Y >= p0.Y {
			return Geom_Quadrant_NE
		}
		return Geom_Quadrant_SE
	}
	if p1.Y >= p0.Y {
		return Geom_Quadrant_NW
	}
	return Geom_Quadrant_SW
}

// Geom_Quadrant_IsOpposite returns true if the quadrants are 1 and 3, or 2 and 4.
func Geom_Quadrant_IsOpposite(quad1, quad2 int) bool {
	if quad1 == quad2 {
		return false
	}
	diff := (quad1 - quad2 + 4) % 4
	// If quadrants are not adjacent, they are opposite.
	return diff == 2
}

// Geom_Quadrant_CommonHalfPlane returns the right-hand quadrant of the halfplane defined by
// the two quadrants, or -1 if the quadrants are opposite, or the quadrant if
// they are identical.
func Geom_Quadrant_CommonHalfPlane(quad1, quad2 int) int {
	// If quadrants are the same they do not determine a unique common
	// halfplane. Simply return one of the two possibilities.
	if quad1 == quad2 {
		return quad1
	}
	diff := (quad1 - quad2 + 4) % 4
	// If quadrants are not adjacent, they do not share a common halfplane.
	if diff == 2 {
		return -1
	}
	min := quad1
	if quad2 < quad1 {
		min = quad2
	}
	max := quad1
	if quad2 > quad1 {
		max = quad2
	}
	// For this one case, the righthand plane is NOT the minimum index.
	if min == 0 && max == 3 {
		return 3
	}
	// In general, the halfplane index is the minimum of the two adjacent
	// quadrants.
	return min
}

// Geom_Quadrant_IsInHalfPlane returns whether the given quadrant lies within the given
// halfplane (specified by its right-hand quadrant).
func Geom_Quadrant_IsInHalfPlane(quad, halfPlane int) bool {
	if halfPlane == Geom_Quadrant_SE {
		return quad == Geom_Quadrant_SE || quad == Geom_Quadrant_SW
	}
	return quad == halfPlane || quad == halfPlane+1
}

// Geom_Quadrant_IsNorthern returns true if the given quadrant is 0 or 1.
func Geom_Quadrant_IsNorthern(quad int) bool {
	return quad == Geom_Quadrant_NE || quad == Geom_Quadrant_NW
}
