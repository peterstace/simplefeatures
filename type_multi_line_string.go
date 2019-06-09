package simplefeatures

// MultiLineString is a multicurve whose elements are LineStrings.
type MultiLineString struct {
	lines []LineString
}

func NewMultiLineString(lines []LineString) (MultiLineString, error) {
	// TODO: validation
	return MultiLineString{lines}, nil
}

func NewMultiLineStringFromCoords(coords [][]Coordinates) (MultiLineString, error) {
	lines := make([]LineString, len(coords))
	for i, c := range coords {
		line, err := NewLineStringFromCoords(c)
		if err != nil {
			return MultiLineString{}, err
		}
		lines[i] = line
	}
	return MultiLineString{lines}, nil
}

func (m MultiLineString) AsText() []byte {
	return m.AppendWKT(nil)
}

func (m MultiLineString) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("MULTILINESTRING")...)
	if len(m.lines) == 0 {
		return append(dst, []byte(" EMPTY")...)
	}
	dst = append(dst, '(')
	for i, line := range m.lines {
		dst = line.appendWKTBody(dst)
		if i != len(m.lines)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ')')
}
