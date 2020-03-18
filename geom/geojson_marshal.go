package geom

func appendGeoJSONCoordinate(dst []byte, coords Coordinates) []byte {
	dst = append(dst, '[')
	dst = appendFloat(dst, coords.X)
	dst = append(dst, ',')
	dst = appendFloat(dst, coords.Y)
	if coords.Type.Is3D() {
		dst = append(dst, ',')
		dst = appendFloat(dst, coords.Z)
	}
	// GeoJSON explicitly prohibits including M values.
	return append(dst, ']')
}

func appendGeoJSONSequence(dst []byte, seq Sequence, empty BitSet) []byte {
	dst = append(dst, '[')
	n := seq.Length()
	var seenFirst bool
	for i := 0; i < n; i++ {
		if empty.Get(i) {
			// GeoJSON doesn't support empty Points within MultiPoints.
			continue
		}
		if seenFirst {
			dst = append(dst, ',')
		}
		dst = appendGeoJSONCoordinate(dst, seq.Get(i))
		seenFirst = true
	}
	dst = append(dst, ']')
	return dst
}

func appendGeoJSONSequences(dst []byte, seqs []Sequence) []byte {
	dst = append(dst, '[')
	for i, seq := range seqs {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = appendGeoJSONSequence(dst, seq, BitSet{})
	}
	dst = append(dst, ']')
	return dst
}

func appendGeoJSONSequenceMatrix(dst []byte, matrix [][]Sequence) []byte {
	dst = append(dst, '[')
	for i, seqs := range matrix {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = appendGeoJSONSequences(dst, seqs)
	}
	dst = append(dst, ']')
	return dst
}
