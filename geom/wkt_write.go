package geom

func appendWKTHeader(dst []byte, geomType string, ctype CoordinatesType) []byte {
	dst = append(dst, geomType...)
	dst = append(dst, [4]string{"", " Z ", " M ", " ZM "}[ctype]...)
	return dst
}

func appendWKTCoords(dst []byte, coords Coordinates, parens bool) []byte {
	if parens {
		dst = append(dst, '(')
	}
	dst = appendFloat(dst, coords.X)
	dst = append(dst, ' ')
	dst = appendFloat(dst, coords.Y)
	if coords.Type.Is3D() {
		dst = append(dst, ' ')
		dst = appendFloat(dst, coords.Z)
	}
	if coords.Type.IsMeasured() {
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

func appendWKTSequence(dst []byte, seq Sequence, parens bool) []byte {
	n := seq.Length()
	dst = append(dst, '(')
	for i := 0; i < n; i++ {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = appendWKTCoords(dst, seq.Get(i), parens)
	}
	dst = append(dst, ')')
	return dst
}
