package geom

func intersectionOfIndexedLines(
	lines1 indexedLines, lines2 indexedLines,
) (
	MultiPoint, MultiLineString,
) {
	// TODO: Investigate potential speed up of swapping lines.
	var lss []LineString
	var ptFloats []float64
	seen := make(map[XY]bool)
	for i := range lines1.lines {
		lines2.tree.RangeSearch(lines1.lines[i].envelope().box(), func(j int) error {
			inter := lines1.lines[i].intersectLine(lines2.lines[j])
			if inter.empty {
				return nil
			}
			if inter.ptA == inter.ptB {
				if pt := inter.ptA; !seen[pt] {
					ptFloats = append(ptFloats, pt.X, pt.Y)
					seen[pt] = true
				}
			} else {
				lss = append(lss, line{inter.ptA, inter.ptB}.asLineString())
			}
			return nil
		})
	}
	return NewMultiPoint(NewSequence(ptFloats, DimXY)),
		NewMultiLineStringFromLineStrings(lss)
}

func intersectionOfMultiPointAndMultiPoint(mp1, mp2 MultiPoint) MultiPoint {
	inMP1 := make(map[XY]bool)
	for i := 0; i < mp1.NumPoints(); i++ {
		xy, ok := mp1.PointN(i).XY()
		if ok {
			inMP1[xy] = true
		}
	}
	var floats []float64
	for i := 0; i < mp2.NumPoints(); i++ {
		xy, ok := mp2.PointN(i).XY()
		if ok && inMP1[xy] {
			floats = append(floats, xy.X, xy.Y)
		}
	}
	return NewMultiPoint(NewSequence(floats, DimXY))
}
