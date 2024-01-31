package geom

import "math"

// densify returns a copy of the sequence with additional sets of coordinates
// inserted such that the distance between adjacent sets of coordinates is at
// most maxDist.
func densify(seq Sequence, maxDist float64) Sequence {
	if maxDist <= 0 {
		panic("maxDist must be positive")
	}

	if seq.Length() == 0 {
		return seq
	}

	var dense []float64
	n := seq.Length()
	for i := 0; i+1 < n; i++ {
		c0 := seq.Get(i + 0)
		c1 := seq.Get(i + 1)

		// Copy start of segment:
		dense = c0.appendFloat64s(dense)

		// Copy any additional inter-segment coordinates:
		dist := c0.XY.distanceTo(c1.XY)
		subsections := int(math.Ceil(dist / maxDist))
		for j := 1; j < subsections; j++ {
			cj := interpolateCoords(c0, c1, float64(j)/float64(subsections))
			dense = cj.appendFloat64s(dense)
		}
	}

	// Copy end of last segment:
	dense = seq.Get(n - 1).appendFloat64s(dense)

	return NewSequence(dense, seq.CoordinatesType())
}
