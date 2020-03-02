package geom

import (
	"bytes"
	"database/sql/driver"
	"io"
	"unsafe"
)

// MultiLineString is a multicurve whose elements are LineStrings.
//
// Its assertions are:
//
// 1. It must be made of up zero or more valid LineStrings.
type MultiLineString struct {
	lines []LineString
	ctype CoordinatesType
}

// NewMultiLineString creates a MultiLineString from its constintuent
// LineStrings.
func NewMultiLineString(lines []LineString, opts ...ConstructorOption) (MultiLineString, error) {
	var agg coordinateTypeAggregator
	for _, ls := range lines {
		agg.add(ls.CoordinatesType())
	}
	if agg.err != nil {
		return MultiLineString{}, agg.err
	}
	return MultiLineString{lines, agg.ctype}, nil
}

func NewEmptyMultiLineString(ctype CoordinatesType) MultiLineString {
	return MultiLineString{nil, ctype}
}

// Type return type string for MultiLineString
func (m MultiLineString) Type() string {
	return multiLineStringType
}

// AsGeometry converts this MultiLineString into a Geometry.
func (m MultiLineString) AsGeometry() Geometry {
	return Geometry{multiLineStringTag, unsafe.Pointer(&m)}
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
	dst = appendWKTHeader(dst, "MULTILINESTRING", m.ctype)
	if len(m.lines) == 0 {
		return appendWKTEmpty(dst)
	}
	dst = append(dst, '(')
	for i, line := range m.lines {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = line.appendWKTBody(dst)
	}
	return append(dst, ')')
}

// IsSimple returns true iff the following conditions hold:
//
// 1. Each element (a LineString) is simple.
//
// 2. The intersection between any two distinct elements occurs at points that
// are on the boundaries of both elements.
func (m MultiLineString) IsSimple() bool {
	for _, ls := range m.lines {
		if !ls.IsSimple() {
			return false
		}
	}
	for i := 0; i < len(m.lines); i++ {
		for j := i + 1; j < len(m.lines); j++ {
			// Ignore any intersections if the lines are *exactly* the same
			// (ignoring order). This is to match PostGIS and libgeos
			// behaviour. The OGC spec is ambiguous around this case, so it's
			// just easier to follow other implementations for better
			// interoperability.
			if m.lines[i].EqualsExact(m.lines[j].AsGeometry(), IgnoreOrder) {
				continue
			}

			inter := mustIntersection(
				m.lines[i].AsGeometry(),
				m.lines[j].AsGeometry(),
			)
			if inter.IsEmpty() {
				continue
			}
			bound := mustIntersection(
				m.lines[i].Boundary().AsGeometry(),
				m.lines[j].Boundary().AsGeometry(),
			)
			if !inter.EqualsExact(mustIntersection(inter, bound), IgnoreOrder) {
				return false
			}
		}
	}
	return true
}

func (m MultiLineString) Intersection(g Geometry) (Geometry, error) {
	return intersection(m.AsGeometry(), g)
}

func (m MultiLineString) Intersects(g Geometry) bool {
	return hasIntersection(m.AsGeometry(), g)
}

func (m MultiLineString) IsEmpty() bool {
	for _, ls := range m.lines {
		if !ls.IsEmpty() {
			return false
		}
	}
	return true
}

func (m MultiLineString) Envelope() (Envelope, bool) {
	var env Envelope
	var has bool
	for _, ls := range m.lines {
		e, ok := ls.Envelope()
		if !ok {
			continue
		}
		if has {
			env = env.ExpandToIncludeEnvelope(e)
		} else {
			env = e
			has = true
		}
	}
	return env, has
}

func (m MultiLineString) Boundary() MultiPoint {
	counts := make(map[XY]int)
	var uniqueEndpoints []XY
	for _, ls := range m.lines {
		if ls.IsClosed() {
			continue
		}
		for _, pt := range [2]Point{
			ls.StartPoint(),
			ls.EndPoint(),
		} {
			xy, ok := pt.XY()
			if !ok {
				continue
			}
			if counts[xy] == 0 {
				uniqueEndpoints = append(uniqueEndpoints, xy)
			}
			counts[xy]++
		}
	}

	var bound []XY
	for _, xy := range uniqueEndpoints {
		if counts[xy]%2 == 1 {
			bound = append(bound, xy)
		}
	}
	return NewMultiPointXY(bound)
}

func (m MultiLineString) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := m.AsBinary(&buf)
	return buf.Bytes(), err
}

func (m MultiLineString) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeMultiLineString, m.ctype)
	n := m.NumLineStrings()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		ls := m.LineStringN(i)
		marsh.setErr(ls.AsBinary(w))
	}
	return marsh.err
}

func (m MultiLineString) ConvexHull() Geometry {
	return convexHull(m.AsGeometry())
}

func (m MultiLineString) MarshalJSON() ([]byte, error) {
	var dst []byte
	dst = append(dst, `{"type":"MultiLineString","coordinates":`...)
	dst = appendGeoJSONSequences(dst, m.Coordinates())
	dst = append(dst, '}')
	return dst, nil
}

// Coordinates returns the coordinates of each constintuent LineString in the
// MultiLineString.
func (m MultiLineString) Coordinates() []Sequence {
	n := m.NumLineStrings()
	seqs := make([]Sequence, n)
	for i := 0; i < n; i++ {
		seqs[i] = m.LineStringN(i).Coordinates()
	}
	return seqs
}

// TransformXY transforms this MultiLineString into another MultiLineString according to fn.
func (m MultiLineString) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	n := m.NumLineStrings()
	transformed := make([]LineString, n)
	for i := 0; i < n; i++ {
		var err error
		transformed[i], err = NewLineStringFromSequence(
			transformSequence(m.LineStringN(i).Coordinates(), fn),
			opts...,
		)
		if err != nil {
			return Geometry{}, err
		}
	}
	mls, err := NewMultiLineString(transformed, opts...)
	return mls.AsGeometry(), err
}

// EqualsExact checks if this MultiLineString is exactly equal to another MultiLineString.
func (m MultiLineString) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	return other.IsMultiLineString() &&
		multiLineStringExactEqual(m, other.AsMultiLineString(), opts)
}

// IsValid checks if this MultiLineString is valid
func (m MultiLineString) IsValid() bool {
	for _, ls := range m.lines {
		if !ls.IsValid() {
			return false
		}
	}
	return true
}

// Length gives the sum of the lengths of the constituent members of the multi
// line string.
func (m MultiLineString) Length() float64 {
	var sum float64
	for _, ln := range m.lines {
		sum += ln.Length()
	}
	return sum
}

// Centroid gives the centroid of the coordinates of the multi line string.
func (m MultiLineString) Centroid() Point {
	var sumXY XY
	var sumLength float64
	for i := 0; i < m.NumLineStrings(); i++ {
		ls := m.LineStringN(i)
		xy, length := sumCentroidAndLengthOfLineString(ls)
		sumXY = sumXY.Add(xy)
		sumLength += length
	}
	if sumLength == 0 {
		return NewEmptyPoint(XYOnly)
	}
	return NewPointXY(sumXY.Scale(1.0 / sumLength))
}

// Reverse in the case of MultiLineString outputs each component line string in their
// original order, each individually reversed.
func (m MultiLineString) Reverse() MultiLineString {
	linestrings := make([]LineString, len(m.lines))
	// Form the reversed slice.
	for i := 0; i < len(m.lines); i++ {
		linestrings[i] = m.lines[i].Reverse()
	}
	return MultiLineString{linestrings, m.ctype}
}

func (m MultiLineString) CoordinatesType() CoordinatesType {
	return m.ctype
}
