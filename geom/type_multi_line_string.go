package geom

import (
	"database/sql/driver"
	"unsafe"

	"github.com/peterstace/simplefeatures/rtree"
)

// MultiLineString is a linear geometry that consists of a collection of
// LineStrings. Its zero value is the empty MultiLineString (i.e. the
// collection of zero LineStrings) of 2D coordinate type. It is immutable after
// creation.
type MultiLineString struct {
	lines []LineString
	ctype CoordinatesType
}

// NewMultiLineStringFromLineStrings creates a MultiLineString from its
// constituent LineStrings. The coordinates type of the MultiLineString is the
// lowest common coordinates type of its LineStrings.
func NewMultiLineStringFromLineStrings(lines []LineString, opts ...ConstructorOption) MultiLineString {
	if len(lines) == 0 {
		return MultiLineString{}
	}

	ctype := DimXYZM
	for _, ls := range lines {
		ctype &= ls.CoordinatesType()
	}

	lines = append([]LineString(nil), lines...)
	for i := range lines {
		lines[i] = lines[i].ForceCoordinatesType(ctype)
	}

	return MultiLineString{lines, ctype}
}

// Type returns the GeometryType for a MultiLineString
func (m MultiLineString) Type() GeometryType {
	return TypeMultiLineString
}

// AsGeometry converts this MultiLineString into a Geometry.
func (m MultiLineString) AsGeometry() Geometry {
	return Geometry{TypeMultiLineString, unsafe.Pointer(&m)}
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

// AsText returns the WKT (Well Known Text) representation of this geometry.
func (m MultiLineString) AsText() string {
	return string(m.AppendWKT(nil))
}

// AppendWKT appends the WKT (Well Known Text) representation of this geometry
// to the input byte slice.
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

// IsSimple returns true if this geometry contains no anomalous geometry
// points, such as self intersection or self tangency. A MultiLineString is
// simple if and only if the following conditions hold:
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

	// Map between record ID in the rtree and a particular line segment:
	toRecordID := func(lineStringIdx, seqIdx int) int {
		return int(uint64(lineStringIdx)<<32 | uint64(seqIdx))
	}
	fromRecordID := func(recordID int) (lineStringIdx, seqIdx int) {
		lineStringIdx = int(uint64(recordID) >> 32)
		seqIdx = int((uint64(recordID) << 32) >> 32)
		return
	}

	// Create an RTree containing all line segments.
	var numItems int
	for _, ls := range m.lines {
		numItems += max(0, ls.Coordinates().Length()-1)
	}
	items := make([]rtree.BulkItem, 0, numItems)
	for i, ls := range m.lines {
		seq := ls.Coordinates()
		seqLen := seq.Length()
		for j := 0; j < seqLen; j++ {
			ln, ok := getLine(seq, j)
			if !ok {
				continue
			}
			items = append(items, rtree.BulkItem{
				ln.envelope().box(),
				toRecordID(i, j),
			})
		}
	}
	tree := rtree.BulkLoad(items)

	for i, ls := range m.lines {
		seq := ls.Coordinates()
		seqLen := seq.Length()
		for j := 0; j < seqLen; j++ {
			ln, ok := getLine(seq, j)
			if !ok {
				continue
			}
			isSimple := true // assume simple until proven otherwise
			tree.RangeSearch(ln.envelope().box(), func(recordID int) error {
				// Ignore the intersection if it's for the same LineString that we're currently looking up.
				lineStringIdx, seqIdx := fromRecordID(recordID)
				if lineStringIdx == i {
					return nil
				}

				otherLS := m.lines[lineStringIdx]
				other, ok := getLine(otherLS.Coordinates(), seqIdx)
				if !ok {
					// Shouldn't even happen, since we were able to insert this
					// entry into the RTree.
					panic("could not getLine")
				}

				inter := ln.intersectLine(other)
				if inter.empty {
					return nil
				}

				// The MLS is NOT simple if the intersection is NOT on the
				// boundary of each LineString.
				if inter.ptA != inter.ptB {
					// Intersection is a line segment, so CANNOT be only on the
					// boundary.
					isSimple = false
					return rtree.Stop
				}
				boundary := intersectionOfMultiPointAndMultiPoint(ls.Boundary(), otherLS.Boundary())
				if !hasIntersectionPointWithMultiPoint(NewPointFromXY(inter.ptA), boundary) {
					isSimple = false
					return rtree.Stop
				}
				return nil
			})
			if !isSimple {
				return false
			}
		}
	}
	return true
}

// IsEmpty return true if and only if this MultiLineString doesn't contain any
// LineStrings, or only contains empty LineStrings.
func (m MultiLineString) IsEmpty() bool {
	for _, ls := range m.lines {
		if !ls.IsEmpty() {
			return false
		}
	}
	return true
}

// Envelope returns the Envelope that most tightly surrounds the geometry. If
// the geometry is empty, then false is returned.
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

// Boundary returns the spatial boundary of this MultiLineString. This is
// calculated using the "mod 2 rule". The rule states that a Point is included
// as part of the boundary if and only if it appears on the boundry of an odd
// number of members in the collection.
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

	var floats []float64
	for _, xy := range uniqueEndpoints {
		if counts[xy]%2 == 1 {
			floats = append(floats, xy.X, xy.Y)
		}
	}
	seq := NewSequence(floats, DimXY)
	return NewMultiPoint(seq)
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (m MultiLineString) Value() (driver.Value, error) {
	return m.AsBinary(), nil
}

// Scan implements the database/sql.Scanner interface by parsing the src value
// as WKB (Well Known Binary).
//
// If the WKB doesn't represent a MultiLineString geometry, then an error is returned.
//
// It constructs the resultant geometry with no ConstructionOptions. If
// ConstructionOptions are needed, then the value should be scanned into a byte
// slice and then UnmarshalWKB called manually (passing in the
// ConstructionOptions as desired).
func (m *MultiLineString) Scan(src interface{}) error {
	return scanAsType(src, m, TypeMultiLineString)
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (m MultiLineString) AsBinary() []byte {
	return m.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (m MultiLineString) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaller(dst)
	marsh.writeByteOrder()
	marsh.writeGeomType(TypeMultiLineString, m.ctype)
	n := m.NumLineStrings()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		ls := m.LineStringN(i)
		marsh.buf = ls.AppendWKB(marsh.buf)
	}
	return marsh.buf
}

// ConvexHull returns the geometry representing the smallest convex geometry
// that contains this geometry.
func (m MultiLineString) ConvexHull() Geometry {
	return convexHull(m.AsGeometry())
}

// MarshalJSON implements the encoding/json.Marshaller interface by encoding
// this geometry as a GeoJSON geometry object.
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
func (m MultiLineString) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (MultiLineString, error) {
	n := m.NumLineStrings()
	transformed := make([]LineString, n)
	for i := 0; i < n; i++ {
		var err error
		transformed[i], err = NewLineString(
			transformSequence(m.LineStringN(i).Coordinates(), fn),
			opts...,
		)
		if err != nil {
			return MultiLineString{}, wrapTransformed(err)
		}
	}
	return NewMultiLineStringFromLineStrings(transformed, opts...), nil
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
		return NewEmptyPoint(DimXY)
	}
	return NewPointFromXY(sumXY.Scale(1.0 / sumLength))
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

// CoordinatesType returns the CoordinatesType used to represent points making
// up the geometry.
func (m MultiLineString) CoordinatesType() CoordinatesType {
	return m.ctype
}

// ForceCoordinatesType returns a new MultiLineString with a different CoordinatesType. If a
// dimension is added, then new values are populated with 0.
func (m MultiLineString) ForceCoordinatesType(newCType CoordinatesType) MultiLineString {
	flat := make([]LineString, len(m.lines))
	for i := range m.lines {
		flat[i] = m.lines[i].ForceCoordinatesType(newCType)
	}
	return MultiLineString{flat, newCType}
}

// Force2D returns a copy of the MultiLineString with Z and M values removed.
func (m MultiLineString) Force2D() MultiLineString {
	return m.ForceCoordinatesType(DimXY)
}

func (m MultiLineString) asLines() []line {
	var n int
	numLineStrings := m.NumLineStrings()
	for i := 0; i < numLineStrings; i++ {
		n += max(0, m.LineStringN(i).Coordinates().Length()-1)
	}

	lines := make([]line, 0, n)
	for i := 0; i < numLineStrings; i++ {
		seq := m.LineStringN(i).Coordinates()
		length := seq.Length()
		for j := 0; j < length; j++ {
			ln, ok := getLine(seq, j)
			if ok {
				lines = append(lines, ln)
			}
		}
	}
	return lines
}

// PointOnSurface returns a Point on one of the LineStrings in the collection.
func (m MultiLineString) PointOnSurface() Point {
	// Find the nearest control point on the LineString, ignoring the start/end points.
	nearest := newNearestPointAccumulator(m.Centroid())
	for i := 0; i < m.NumLineStrings(); i++ {
		seq := m.LineStringN(i).Coordinates()
		n := seq.Length()
		for j := 1; j < n-1; j++ {
			candidate := NewPointFromXY(seq.GetXY(j))
			nearest.consider(candidate)
		}
	}
	if !nearest.point.IsEmpty() {
		return nearest.point
	}

	// If we still don't have a point, then consider the start/end points.
	for i := 0; i < m.NumLineStrings(); i++ {
		ls := m.LineStringN(i)
		nearest.consider(ls.StartPoint().Force2D())
		nearest.consider(ls.EndPoint().Force2D())
	}
	return nearest.point
}

func (m MultiLineString) controlPoints() int {
	var sum int
	for _, ls := range m.lines {
		sum += ls.Coordinates().Length()
	}
	return sum
}

// Dump returns the MultiLineString represented as a LineString slice.
func (m MultiLineString) Dump() []LineString {
	lss := make([]LineString, len(m.lines))
	copy(lss, m.lines)
	return lss
}
