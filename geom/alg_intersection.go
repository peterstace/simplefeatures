package geom

func intersectionOfLineStringAndLineString(ls1, ls2 LineString) (MultiPoint, MultiLineString) {
	// TODO: This is quadratic in the length of the input. Should use an
	// acceleration structure.
	var points []XY
	var lines []line
	seq1 := ls1.Coordinates()
	for i := 0; i < seq1.Length(); i++ {
		ln1, ok := getLine(seq1, i)
		if !ok {
			continue
		}
		seq2 := ls2.Coordinates()
		for j := 0; j < seq2.Length(); j++ {
			ln2, ok := getLine(seq2, j)
			if !ok {
				continue
			}
			inter := ln1.intersectLine(ln2)
			if inter.empty {
				continue
			}
			if inter.ptA == inter.ptB {
				points = append(points, inter.ptA)
			} else {
				lines = append(lines, line{inter.ptA, inter.ptB})
			}
		}
	}

	mpFloats := make([]float64, 0, 2*len(points))
	for _, xy := range points {
		mpFloats = append(mpFloats, xy.X, xy.Y)
	}

	lss := make([]LineString, 0, len(lines))
	for _, ln := range lines {
		lss = append(lss, ln.asLineString())
	}

	return NewMultiPoint(NewSequence(mpFloats, DimXY)), NewMultiLineStringFromLineStrings(lss)
}

func intersectionOfMultiLineStringAndMultiLineString(
	mls1, mls2 MultiLineString,
) (
	MultiPoint, MultiLineString,
) {
	// TODO: This has horrible time complexity. Should use acceleration structure.
	var pts []Point
	var lss []LineString
	for i := 0; i < mls1.NumLineStrings(); i++ {
		ls1 := mls1.LineStringN(i)
		for j := 0; j < mls2.NumLineStrings(); j++ {
			ls2 := mls2.LineStringN(j)
			interMP, interMLS := intersectionOfLineStringAndLineString(ls1, ls2)
			for k := 0; k < interMP.NumPoints(); k++ {
				pts = append(pts, interMP.PointN(k))
			}
			for k := 0; k < interMLS.NumLineStrings(); k++ {
				lss = append(lss, interMLS.LineStringN(k))
			}
		}
	}
	return NewMultiPointFromPoints(pts), NewMultiLineStringFromLineStrings(lss)
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
