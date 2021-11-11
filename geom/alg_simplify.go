package geom

func ramerDouglasPeucker(dst []float64, seq Sequence, threshold float64) []float64 {
	if seq.Length() <= 2 {
		return seq.appendAllPoints(dst)
	}

	start := 0
	end := seq.Length() - 1

	for start < end {
		dst = seq.appendPoint(dst, start)
		newEnd := end
		for {
			var maxDist float64
			var maxDistIdx int
			for i := start + 1; i < newEnd; i++ {
				if d := perpendicularDistance(
					seq.GetXY(i),
					seq.GetXY(start),
					seq.GetXY(newEnd),
				); d > maxDist {
					maxDistIdx = i
					maxDist = d
				}
			}
			if maxDist <= threshold {
				break
			}
			newEnd = maxDistIdx
		}
		start = newEnd
	}
	dst = seq.appendPoint(dst, end)
	return dst
}

// perpendicularDistance is the distance from 'p' to the infinite line going
// through 'a' and 'b'. If 'a' and 'b' are the same, then the distance between
// 'a'/'b' and 'p' is returned.
func perpendicularDistance(p, a, b XY) float64 {
	if a == b {
		return p.Sub(a).Length()
	}
	aSubP := a.Sub(p)
	bSubA := b.Sub(a)
	unit := bSubA.Scale(1 / bSubA.Length())
	perpendicular := aSubP.Sub(unit.Scale(aSubP.Dot(unit)))
	return perpendicular.Length()
}
