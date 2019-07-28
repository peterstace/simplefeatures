package simplefeatures

import (
	"database/sql/driver"
	"io"
)

// MultiLineString is a multicurve whose elements are LineStrings.
//
// Its assertions are:
//
// 1. It must be made of up zero or more valid LineStrings.
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

// NumLineStrings gives the number of LineString elements in the
// MultiLineString.
func (m MultiLineString) NumLineStrings() int {
	return len(m.lines)
}

// LineStringN gives the nth (zero indexed) LineString element.
func (m MultiLineString) LineStringN(n int) LineString {
	return m.lines[n]
}

func (m MultiLineString) AsText() string {
	return string(m.AppendWKT(nil))
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

// IsSimple returns true iff the following conditions hold:
//
// 1. Each element (a LineString) is simple.
//
// 2. The intersection between any two elements occurs at points that are on
// the boundaries of both elements.
func (m MultiLineString) IsSimple() bool {
	for _, ls := range m.lines {
		if !ls.IsSimple() {
			return false
		}
	}
	for i := 0; i < len(m.lines); i++ {
		for j := i + 1; j < len(m.lines); j++ {
			inter := m.lines[i].Intersection(m.lines[j])
			bound := m.lines[i].Boundary().Intersection(m.lines[j].Boundary())
			if !inter.Equals(inter.Intersection(bound)) {
				return false
			}
		}
	}
	return true
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

func (m MultiLineString) Envelope() (Envelope, bool) {
	if len(m.lines) == 0 {
		return Envelope{}, false
	}
	env := mustEnvelope(m.lines[0])
	for _, line := range m.lines[1:] {
		e := mustEnvelope(line)
		env = env.Union(e)
	}
	return env, true
}
func (m MultiLineString) Boundary() Geometry {
	if m.IsEmpty() {
		// Postgis behaviour (but any other empty set would be ok).
		return NewMultiLineString(nil)
	}

	counts := make(map[xyHash]int)
	var uniqueEndpoints []Point
	for _, ls := range m.lines {
		if ls.IsClosed() {
			continue
		}
		for _, pt := range [2]Point{
			NewPointFromCoords(ls.lines[0].a),
			NewPointFromCoords(ls.lines[len(ls.lines)-1].b),
		} {
			hash := pt.coords.XY.hash()
			_, seen := counts[hash]
			if !seen {
				uniqueEndpoints = append(uniqueEndpoints, pt)
			}
			counts[hash]++
		}
	}

	var bound []Point
	for _, pt := range uniqueEndpoints {
		if counts[pt.coords.XY.hash()]%2 == 1 {
			bound = append(bound, pt)
		}
	}
	return NewMultiPoint(bound)
}

func (m MultiLineString) Value() (driver.Value, error) {
	return m.AsText(), nil
}

func (m MultiLineString) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeMultiLineString)
	n := m.NumLineStrings()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		ls := m.LineStringN(i)
		marsh.setErr(ls.AsBinary(w))
	}
	return marsh.err
}

func (m MultiLineString) MarshalJSON() ([]byte, error) {
	numLines := m.NumLineStrings()
	coords := make([][]Coordinates, numLines)
	for i := 0; i < numLines; i++ {
		numPts := m.LineStringN(i).NumPoints()
		coords[i] = make([]Coordinates, numPts)
		for j := 0; j < numPts; j++ {
			coords[i][j] = m.LineStringN(i).PointN(j).Coordinates()
		}
	}
	return marshalGeoJSON("MultiLineString", coords)
}
