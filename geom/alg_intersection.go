package geom

func intersectionOfIndexedLines(
	lines1 indexedLines, lines2 indexedLines,
) (
	MultiPoint, MultiLineString,
) {
	// TODO: Investigate potential speed up of swapping lines.
	var lss []LineString
	var pts []Point
	seen := make(map[XY]bool)
	for i := range lines1.lines {
		lines2.tree.RangeSearch(lines1.lines[i].envelope().box(), func(j int) error {
			inter := lines1.lines[i].intersectLine(lines2.lines[j])
			if inter.empty {
				return nil
			}
			if inter.ptA == inter.ptB {
				if xy := inter.ptA; !seen[xy] {
					pts = append(pts, xy.AsPoint())
					seen[xy] = true
				}
			} else {
				lss = append(lss, line{inter.ptA, inter.ptB}.asLineString())
			}
			return nil
		})
	}
	return NewMultiPoint(pts), NewMultiLineStringFromLineStrings(lss)
}

func intersectionOfMultiPointAndMultiPoint(mp1, mp2 MultiPoint) MultiPoint {
	inMP1 := make(map[XY]bool)
	for i := 0; i < mp1.NumPoints(); i++ {
		xy, ok := mp1.PointN(i).XY()
		if ok {
			inMP1[xy] = true
		}
	}
	var pts []Point
	for i := 0; i < mp2.NumPoints(); i++ {
		xy, ok := mp2.PointN(i).XY()
		if ok && inMP1[xy] {
			pts = append(pts, xy.AsPoint())
		}
	}
	return NewMultiPoint(pts)
}
