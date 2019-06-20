package simplefeatures

// MultiLineString is a multicurve whose elements are LineStrings.
type MultiLineString struct {
	lines []LineString
}

func NewMultiLineString(lines []LineString) MultiLineString {
	return MultiLineString{lines}
}

func NewMultiLineStringFromCoords(coords [][]Coordinates) (MultiLineString, error) {
	var lines []LineString
	for _, c := range coords {
		if len(c) == 0 {
			continue
		}
		line, err := NewLineString(c)
		if err != nil {
			return MultiLineString{}, err
		}
		lines = append(lines, line)
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

func (m MultiLineString) IsSimple() bool {
	panic("not implemented")
}

func (m MultiLineString) Intersection(g Geometry) Geometry {
	return intersection(m, g)
}

func (m MultiLineString) IsEmpty() bool {
	return len(m.lines) == 0
}

func (m MultiLineString) Dimension() int {
	if m.IsEmpty() {
		return 0
	}
	return 1
}

func (m MultiLineString) Equals(other Geometry) bool {
	return equals(m, other)
}

func (m MultiLineString) FiniteNumberOfPoints() (int, bool) {
	return 0, m.IsEmpty()
}
