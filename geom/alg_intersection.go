package geom

import "github.com/peterstace/simplefeatures/rtree"

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
		lines2.tree.Search(lines1.lines[i].envelope().box(), func(j int) error {
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

func intersectionOfLines(lines1, lines2 []line) (MultiPoint, MultiLineString) {
	// TODO: Should we swap lines1 and lines2 depending on length?

	bulk := make([]rtree.BulkItem, len(lines1))
	for i, ln := range lines1 {
		bulk[i] = rtree.BulkItem{
			Box:      ln.envelope().box(),
			RecordID: i,
		}
	}
	tree := rtree.BulkLoad(bulk)

	var lines []LineString
	var ptFloats []float64
	seen := make(map[XY]bool)

	for j := range lines2 {
		tree.Search(lines2[j].envelope().box(), func(i int) error {
			inter := lines2[j].intersectLine(lines1[i])
			if inter.empty {
				return nil
			}
			if inter.ptA == inter.ptB {
				if pt := inter.ptA; !seen[pt] {
					ptFloats = append(ptFloats, pt.X, pt.Y)
					seen[pt] = true
				}
			} else {
				lines = append(lines, line{inter.ptA, inter.ptB}.asLineString())
			}
			return nil
		})
	}
	return NewMultiPoint(NewSequence(ptFloats, DimXY)),
		NewMultiLineStringFromLineStrings(lines)
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
