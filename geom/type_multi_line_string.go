package geom

import (
	"database/sql/driver"
	"fmt"
	"unsafe"

	"github.com/peterstace/simplefeatures/rtree"
)

// MultiLineString is a linear geometry that consists of a collection of
// LineStrings. Its zero value is the empty MultiLineString (i.e. the
// collection of zero LineStrings) of 2D coordinate type. It is immutable after
// creation.
type MultiLineString struct {
	// Invariant: ctype matches the coordinates type of each line.
	lines []LineString
	ctype CoordinatesType
}

// NewMultiLineString creates a MultiLineString from its constituent
// LineStrings. The coordinates type of the MultiLineString is the lowest
// common coordinates type of its LineStrings.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiLineString(lines []LineString) MultiLineString {
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

// Validate checks if the MultiLineString is valid. The only validation rule is
// that each LineString in the collection must be valid.
func (m MultiLineString) Validate() error {
	for i, ls := range m.lines {
		if err := ls.Validate(); err != nil {
			return wrap(err, "validating linestring at index %d", i)
		}
	}
	return nil
}

// Type returns the GeometryType for a MultiLineString.
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
		numItems += maxInt(0, ls.Coordinates().Length()-1)
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
				Box:      ln.box(),
				RecordID: toRecordID(i, j),
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
			tree.RangeSearch(ln.box(), func(recordID int) error {
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
				if !hasIntersectionPointWithMultiPoint(inter.ptA.AsPoint(), boundary) {
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

// Envelope returns the Envelope that most tightly surrounds the geometry.
func (m MultiLineString) Envelope() Envelope {
	var env Envelope
	for _, ls := range m.lines {
		env = env.ExpandToIncludeEnvelope(ls.Envelope())
	}
	return env
}

// Boundary returns the spatial boundary of this MultiLineString. This is
// calculated using the "mod 2 rule". The rule states that a Point is included
// as part of the boundary if and only if it appears on the boundary of an odd
// number of members in the collection.
func (m MultiLineString) Boundary() MultiPoint {
	counts := make(map[XY]int)
	var uniqueEndpoints []Point
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
				uniqueEndpoints = append(uniqueEndpoints, pt)
			}
			counts[xy]++
		}
	}

	var mod2Points []Point
	for _, pt := range uniqueEndpoints {
		xy, ok := pt.XY()
		if !ok {
			// Can't happen, because we already check to make sure pt is not empty.
			panic("MultiLineString Boundary internal error")
		}
		if counts[xy]%2 == 1 {
			mod2Points = append(mod2Points, pt)
		}
	}
	return NewMultiPoint(mod2Points)
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
// Geometry constraint validation is performed on the resultant geometry (an
// error will be returned if the geometry is invalid). If this validation isn't
// needed or is undesirable, then the WKB should be scanned into a byte slice
// and then UnmarshalWKB called manually (passing in NoValidate{}).
func (m *MultiLineString) Scan(src interface{}) error {
	return scanAsType(src, m)
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (m MultiLineString) AsBinary() []byte {
	return m.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (m MultiLineString) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaler(dst)
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

// MarshalJSON implements the encoding/json.Marshaler interface by encoding
// this geometry as a GeoJSON geometry object.
func (m MultiLineString) MarshalJSON() ([]byte, error) {
	var dst []byte
	dst = append(dst, `{"type":"MultiLineString","coordinates":`...)
	dst = appendGeoJSONSequences(dst, m.Coordinates())
	dst = append(dst, '}')
	return dst, nil
}

// UnmarshalJSON implements the encoding/json.Unmarshaler interface by decoding
// the GeoJSON representation of a MultiLineString.
func (m *MultiLineString) UnmarshalJSON(buf []byte) error {
	return unmarshalGeoJSONAsType(buf, m)
}

// Coordinates returns the coordinates of each constituent LineString in the
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
func (m MultiLineString) TransformXY(fn func(XY) XY) MultiLineString {
	n := m.NumLineStrings()
	if n == 0 {
		return MultiLineString{}.ForceCoordinatesType(m.CoordinatesType())
	}
	transformed := make([]LineString, n)
	for i := 0; i < n; i++ {
		seq := transformSequence(m.LineStringN(i).Coordinates(), fn)
		transformed[i] = NewLineString(seq)
	}
	return NewMultiLineString(transformed)
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
	return sumXY.Scale(1.0 / sumLength).AsPoint()
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
		n += maxInt(0, m.LineStringN(i).Coordinates().Length()-1)
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
			candidate := seq.GetXY(j).AsPoint()
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

// DumpCoordinates returns the coordinates (as a Sequence) that constitute the
// MultiLineString.
func (m MultiLineString) DumpCoordinates() Sequence {
	var n int
	for _, ls := range m.lines {
		n += ls.seq.Length() * m.ctype.Dimension()
	}
	coords := make([]float64, 0, n)
	for _, ls := range m.lines {
		coords = ls.Coordinates().appendAllPoints(coords)
	}
	seq := NewSequence(coords, m.ctype)
	seq.assertNoUnusedCapacity()
	return seq
}

// Summary returns a text summary of the MultiLineString following a similar format to https://postgis.net/docs/ST_Summary.html.
func (m MultiLineString) Summary() string {
	numPoints := m.DumpCoordinates().Length()

	var lineStringSuffix string
	numLineStrings := m.NumLineStrings()
	if numLineStrings != 1 {
		lineStringSuffix = "s"
	}
	return fmt.Sprintf("%s[%s] with %d linestring%s consisting of %d total points",
		m.Type(), m.CoordinatesType(), numLineStrings, lineStringSuffix, numPoints)
}

// String returns the string representation of the MultiLineString.
func (m MultiLineString) String() string {
	return m.Summary()
}

// Simplify returns a simplified version of the MultiLineString by using the
// Ramer-Douglas-Peucker algorithm on each of the child LineStrings. If the
// Ramer-Douglas-Peucker were to create an invalid child LineString (i.e. one
// having only a single distinct point), then it is omitted in the output.
// Empty child LineStrings are also omitted from the output.
func (m MultiLineString) Simplify(threshold float64) MultiLineString {
	n := m.NumLineStrings()
	lss := make([]LineString, 0, n)
	for i := 0; i < n; i++ {
		ls := m.LineStringN(i).Simplify(threshold)
		if !ls.IsEmpty() {
			lss = append(lss, ls)
		}
	}
	return NewMultiLineString(lss).ForceCoordinatesType(m.CoordinatesType())
}

// Densify returns a new MultiLineString with additional linearly interpolated
// control points such that the distance between any two consecutive control
// points is at most the given maxDistance.
//
// Panics if maxDistance is zero or negative.
func (m MultiLineString) Densify(maxDistance float64) MultiLineString {
	lss := make([]LineString, len(m.lines))
	for i := range m.lines {
		lss[i] = m.lines[i].Densify(maxDistance)
	}
	return MultiLineString{lss, m.ctype}
}

// SnapToGrid returns a copy of the MultiLineString with all coordinates
// snapped to a base 10 grid.
//
// The grid spacing is specified by the number of decimal places to round to
// (with negative decimal places being allowed). E.g., a decimalPlaces value of
// 2 would cause all coordinates to be rounded to the nearest 0.01, and a
// decimalPlaces of -1 would cause all coordinates to be rounded to the nearest
// 10.
//
// Returned MultiLineStrings may be invalid due to snapping, even if the input
// geometry was valid.
func (m MultiLineString) SnapToGrid(decimalPlaces int) MultiLineString {
	return m.TransformXY(snapToGridXY(decimalPlaces))
}
