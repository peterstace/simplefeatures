package geom

func appendWKTHeader(dst []byte, geomType string, ctype CoordinatesType) []byte {
	dst = append(dst, geomType...)
	dst = append(dst, [4]string{"", " Z ", " M ", " ZM "}[ctype]...)
	return dst
}

func appendWKTCoords(dst []byte, coords Coordinates, ctype CoordinatesType, parens bool) []byte {
	if parens {
		dst = append(dst, '(')
	}
	dst = appendFloat(dst, coords.X)
	dst = append(dst, ' ')
	dst = appendFloat(dst, coords.Y)
	if ctype.Is3D() {
		dst = append(dst, ' ')
		dst = appendFloat(dst, coords.Z)
	}
	if ctype.IsMeasured() {
		dst = append(dst, ' ')
		dst = appendFloat(dst, coords.M)
	}
	if parens {
		dst = append(dst, ')')
	}
	return dst
}

func appendWKTEmpty(dst []byte) []byte {
	if len(dst) > 0 {
		switch dst[len(dst)-1] {
		case '(', ',', ' ':
		default:
			dst = append(dst, ' ')
		}
	}
	return append(dst, "EMPTY"...)
}

// TODO: Might need to pass in an empty BitSet for MultiPoints
func appendWKTSequence(dst []byte, seq Sequence, parens bool, empty BitSet) []byte {
	ctype := seq.CoordinatesType()
	n := seq.Length()
	dst = append(dst, '(')
	for i := 0; i < n; i++ {
		if i > 0 {
			dst = append(dst, ',')
		}
		if empty.Get(i) {
			dst = appendWKTEmpty(dst)
		} else {
			c := seq.Get(i)
			dst = appendWKTCoords(dst, c, ctype, parens)
		}
	}
	dst = append(dst, ')')
	return dst
}
