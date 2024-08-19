package geom

import (
	"database/sql/driver"
	"fmt"
	"unsafe"

	"github.com/peterstace/simplefeatures/rtree"
)

// LineString is a linear geometry defined by linear interpolation between a
// finite set of points. Its zero value is the empty line string. It is
// immutable after creation.
type LineString struct {
	seq Sequence
}

// NewLineString creates a new LineString from a Sequence of points.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewLineString(seq Sequence) LineString {
	return LineString{seq}
}

// Validate checks if the LineString is valid. For it to be valid, the
// following rules must hold.
//
//  1. The XY values must not be NaN or Inf.
//  2. For non-empty LineStrings, there must be at least two distinct XY
//     values.
func (s LineString) Validate() error {
	if s.seq.Length() == 0 {
		return nil
	}
	if err := s.seq.validate(); err != nil {
		return err
	}
	if !hasAtLeast2DistinctPointsInSeq(s.seq) {
		return violateTwoPoints.errAtXY(s.seq.GetXY(0))
	}
	return nil
}

func hasAtLeast2DistinctPointsInSeq(seq Sequence) bool {
	n := seq.Length()
	if n == 0 {
		return false
	}
	first := seq.GetXY(0)
	for i := 1; i < n; i++ {
		if seq.GetXY(i) != first {
			return true
		}
	}
	return false
}

// Type returns the GeometryType for a LineString.
func (s LineString) Type() GeometryType {
	return TypeLineString
}

// AsGeometry converts this LineString into a Geometry.
func (s LineString) AsGeometry() Geometry {
	return Geometry{TypeLineString, unsafe.Pointer(&s)}
}

// StartPoint gives the first point of the LineString. If the LineString is
// empty then it returns the empty Point.
func (s LineString) StartPoint() Point {
	if s.IsEmpty() {
		return NewEmptyPoint(s.CoordinatesType())
	}
	c := s.seq.Get(0)
	return NewPoint(c)
}

// EndPoint gives the last point of the LineString. If the LineString is empty
// then it returns the empty Point.
func (s LineString) EndPoint() Point {
	if s.IsEmpty() {
		return NewEmptyPoint(s.CoordinatesType())
	}
	end := s.seq.Length() - 1
	c := s.seq.Get(end)
	return NewPoint(c)
}

// AsText returns the WKT (Well Known Text) representation of this geometry.
func (s LineString) AsText() string {
	return string(s.AppendWKT(nil))
}

// AppendWKT appends the WKT (Well Known Text) representation of this geometry
// to the input byte slice.
func (s LineString) AppendWKT(dst []byte) []byte {
	dst = appendWKTHeader(dst, "LINESTRING", s.CoordinatesType())
	return s.appendWKTBody(dst)
}

func (s LineString) appendWKTBody(dst []byte) []byte {
	if s.IsEmpty() {
		return appendWKTEmpty(dst)
	}
	return appendWKTSequence(dst, s.seq, false)
}

// IsSimple returns true if this geometry contains no anomalous geometry
// points, such as self intersection or self tangency. LineStrings are simple
// if and only if the curve defined by the LineString doesn't pass through the
// same point twice (with the except of the two endpoints being coincident).
func (s LineString) IsSimple() bool {
	first, last, ok := firstAndLastLines(s.seq)
	if !ok {
		return true
	}

	n := s.seq.Length()
	items := make([]rtree.BulkItem, 0, n-1)
	for i := 0; i < n; i++ {
		ln, ok := getLine(s.seq, i)
		if !ok {
			continue
		}
		items = append(items, rtree.BulkItem{Box: ln.box(), RecordID: i})
	}
	tree := rtree.BulkLoad(items)

	for i := 0; i < n; i++ {
		ln, ok := getLine(s.seq, i)
		if !ok {
			continue
		}

		prev, ok := previousLine(s.seq, i)
		if !ok {
			prev = -1
		}
		next, ok := nextLine(s.seq, i)
		if !ok {
			next = -1
		}

		simple := true // assume simple until proven otherwise
		tree.RangeSearch(ln.box(), func(j int) error {
			// Skip finding the original line (i == j) or cases where we have
			// already checked that pair (i > j).
			if i >= j {
				return nil
			}

			// Extract the other line from the sequence. We previously were
			// able to access line j (otherwise we wouldn't have been able to
			// put it into the tree). So if we can't access it now, something
			// has gone horribly wrong.
			other, ok := getLine(s.seq, j)
			if !ok {
				panic("couldn't get line")
			}

			inter := ln.intersectLine(other)
			if inter.empty {
				return nil
			}
			if inter.ptA != inter.ptB {
				// Two overlapping line segments.
				simple = false
				return rtree.Stop
			}

			// The dimension must be 1. Since the intersection is between two
			// lines, the intersection must be a *single* point in the cases
			// from this point onwards.

			// Adjacent lines will intersect at a point due to construction, so
			// these cases are okay.
			if j == prev || j == next {
				return nil
			}

			// The first and last segment are allowed to intersect at a point,
			// so long as that point is the start of the first segment and the
			// end of the last segment (i.e. the line string is closed).
			if s.IsClosed() && i == first && j == last {
				return nil
			}

			// Any other point intersection (e.g. looping back on
			// itself at a point) is disallowed for simple linestrings.
			simple = false
			return rtree.Stop
		})

		if !simple {
			return false
		}
	}
	return true
}

// IsClosed returns true if and only if this LineString is not empty and its
// start and end points are coincident.
func (s LineString) IsClosed() bool {
	return !s.IsEmpty() && s.seq.GetXY(0) == s.seq.GetXY(s.seq.Length()-1)
}

// IsEmpty returns true if and only if this LineString is the empty LineString.
// The empty LineString is defined by a zero length coordinates sequence.
func (s LineString) IsEmpty() bool {
	return s.seq.Length() == 0
}

// Envelope returns the Envelope that most tightly surrounds the geometry.
func (s LineString) Envelope() Envelope {
	return s.seq.Envelope()
}

// Boundary returns the spatial boundary of this LineString. For closed
// LineStrings (i.e. LineStrings where the start and end points have the same
// XY value), this is the empty MultiPoint. For non-closed LineStrings, this is
// the MultiPoint containing the two endpoints of the LineString.
func (s LineString) Boundary() MultiPoint {
	if s.IsEmpty() || s.IsClosed() {
		return MultiPoint{}
	}
	return NewMultiPoint([]Point{
		s.StartPoint(),
		s.EndPoint(),
	})
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (s LineString) Value() (driver.Value, error) {
	return s.AsBinary(), nil
}

// Scan implements the database/sql.Scanner interface by parsing the src value
// as WKB (Well Known Binary).
//
// If the WKB doesn't represent a LineString geometry, then an error is returned.
//
// Geometry constraint validation is performed on the resultant geometry (an
// error will be returned if the geometry is invalid). If this validation isn't
// needed or is undesirable, then the WKB should be scanned into a byte slice
// and then UnmarshalWKB called manually (passing in NoValidate{}).
func (s *LineString) Scan(src interface{}) error {
	return scanAsType(src, s)
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (s LineString) AsBinary() []byte {
	return s.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (s LineString) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaler(dst)
	marsh.writeByteOrder()
	marsh.writeGeomType(TypeLineString, s.CoordinatesType())
	marsh.writeSequence(s.seq)
	return marsh.buf
}

// ConvexHull returns the geometry representing the smallest convex geometry
// that contains this geometry.
func (s LineString) ConvexHull() Geometry {
	return convexHull(s.AsGeometry())
}

// MarshalJSON implements the encoding/json.Marshaler interface by encoding
// this geometry as a GeoJSON geometry object.
func (s LineString) MarshalJSON() ([]byte, error) {
	var dst []byte
	dst = append(dst, `{"type":"LineString","coordinates":`...)
	dst = appendGeoJSONSequence(dst, s.seq)
	dst = append(dst, '}')
	return dst, nil
}

// UnmarshalJSON implements the encoding/json.Unmarshaler interface by decoding
// the GeoJSON representation of a LineString.
func (s *LineString) UnmarshalJSON(buf []byte) error {
	return unmarshalGeoJSONAsType(buf, s)
}

// Coordinates returns the coordinates of each point along the LineString.
func (s LineString) Coordinates() Sequence {
	return s.seq
}

// TransformXY transforms this LineString into another LineString according to fn.
func (s LineString) TransformXY(fn func(XY) XY) LineString {
	transformed := transformSequence(s.seq, fn)
	return NewLineString(transformed)
}

// IsRing returns true iff this LineString is both simple and closed (i.e. is a
// linear ring).
func (s LineString) IsRing() bool {
	return s.IsClosed() && s.IsSimple()
}

// Length gives the length of the line string.
func (s LineString) Length() float64 {
	var sum float64
	n := s.seq.Length()
	for i := 0; i+1 < n; i++ {
		xyA := s.seq.GetXY(i)
		xyB := s.seq.GetXY(i + 1)
		delta := xyA.Sub(xyB)
		sum += delta.Length()
	}
	return sum
}

// Centroid gives the centroid of the coordinates of the line string.
func (s LineString) Centroid() Point {
	sumXY, sumLength := sumCentroidAndLengthOfLineString(s)
	if sumLength == 0 {
		return NewEmptyPoint(DimXY)
	}
	return sumXY.Scale(1.0 / sumLength).AsPoint()
}

func sumCentroidAndLengthOfLineString(s LineString) (sumXY XY, sumLength float64) {
	seq := s.Coordinates()
	for i := 0; i < seq.Length(); i++ {
		ln, ok := getLine(seq, i)
		if !ok {
			continue
		}
		length := ln.length()
		cent := ln.centroid()
		sumXY = sumXY.Add(cent.Scale(length))
		sumLength += length
	}
	return sumXY, sumLength
}

// AsMultiLineString is a convenience function that converts this LineString
// into a MultiLineString.
func (s LineString) AsMultiLineString() MultiLineString {
	return NewMultiLineString([]LineString{s})
}

// Reverse in the case of LineString outputs the coordinates in reverse order.
func (s LineString) Reverse() LineString {
	return LineString{s.seq.Reverse()}
}

// CoordinatesType returns the CoordinatesType used to represent points making
// up the geometry.
func (s LineString) CoordinatesType() CoordinatesType {
	return s.seq.CoordinatesType()
}

// ForceCoordinatesType returns a new LineString with a different CoordinatesType. If a
// dimension is added, then new values are populated with 0.
func (s LineString) ForceCoordinatesType(newCType CoordinatesType) LineString {
	return LineString{s.seq.ForceCoordinatesType(newCType)}
}

// Force2D returns a copy of the LineString with Z and M values removed.
func (s LineString) Force2D() LineString {
	return s.ForceCoordinatesType(DimXY)
}

func (s LineString) asLines() []line {
	n := s.seq.Length()
	lines := make([]line, 0, maxInt(0, n-1))
	for i := 0; i < n; i++ {
		ln, ok := getLine(s.seq, i)
		if ok {
			lines = append(lines, ln)
		}
	}
	return lines
}

// PointOnSurface returns a Point on the LineString.
func (s LineString) PointOnSurface() Point {
	// Look for the control point on the LineString (other than the first and
	// last point) that is closest to the centroid.
	n := s.seq.Length()
	nearest := newNearestPointAccumulator(s.Centroid())
	for i := 1; i < n-1; i++ {
		candidate := s.seq.GetXY(i).AsPoint()
		nearest.consider(candidate)
	}
	if !nearest.point.IsEmpty() {
		return nearest.point
	}

	// Consider the star/end points if we don't have anything yet.
	nearest.consider(s.StartPoint().Force2D())
	nearest.consider(s.EndPoint().Force2D())
	return nearest.point
}

// Summary returns a text summary of the LineString following a similar format to https://postgis.net/docs/ST_Summary.html.
func (s LineString) Summary() string {
	return fmt.Sprintf("%s[%s] with %d points", s.Type(), s.CoordinatesType(), s.Coordinates().Length())
}

// String returns the string representation of the LineString.
func (s LineString) String() string {
	return s.Summary()
}

// Simplify returns a simplified version of the LineString using the
// Ramer-Douglas-Peucker algorithm. If the Ramer-Douglas-Peucker algorithm were to create
// an invalid LineString (i.e. one having only a single distinct point), then
// the empty LineString is returned.
func (s LineString) Simplify(threshold float64) LineString {
	seq := s.Coordinates()
	floats := ramerDouglasPeucker(nil, seq, threshold)
	seq = NewSequence(floats, seq.CoordinatesType())
	ls := NewLineString(seq)
	if ls.Validate() != nil {
		return LineString{}.ForceCoordinatesType(s.CoordinatesType())
	}
	return ls
}

// InterpolatePoint returns a Point interpolated along the LineString at the
// given fraction. The fraction should be between 0 and 1, and will be clipped
// to that range if outside of it. Z and M coordinates are also interpolated if
// applicable.
func (s LineString) InterpolatePoint(fraction float64) Point {
	if s.IsEmpty() {
		return Point{}.ForceCoordinatesType(s.CoordinatesType())
	}
	interp := newLinearInterpolator(s.Coordinates())
	return interp.interpolate(fraction)
}

// InterpolateEvenlySpacedPoints returns a MultiPoint consisting of n Points
// evenly spaced along the LineString. If n is negative or 0, then an empty
// MultiPoint is returned. If n is 1, then a MultiPoint containing the
// LineString midpoint is returned. If n is 2 or greater, then the returned
// MultiPoint contains the LineString start point, n - 2 evenly spaced
// intermediate Points, and the LineString end point (in that order).
func (s LineString) InterpolateEvenlySpacedPoints(n int) MultiPoint {
	if n < 0 {
		n = 0
	}
	if n == 0 {
		return MultiPoint{}.ForceCoordinatesType(s.CoordinatesType())
	}

	seq := s.Coordinates()
	if seq.Length() == 0 {
		pts := make([]Point, n)
		for i := 0; i < n; i++ {
			pts[i] = Point{}.ForceCoordinatesType(s.CoordinatesType())
		}
		return NewMultiPoint(pts)
	}

	if n == 1 {
		interp := newLinearInterpolator(s.Coordinates())
		return interp.interpolate(0.5).AsMultiPoint()
	}

	interp := newLinearInterpolator(s.Coordinates())
	pts := make([]Point, n)
	for i := 0; i < n; i++ {
		frac := float64(i) / float64(n-1)
		pts[i] = interp.interpolate(frac)
	}
	return NewMultiPoint(pts)
}

// Densify returns a new LineString with additional linearly interpolated
// control points such that the distance between any two consecutive control
// points is at most the given maxDistance.
//
// Panics if maxDistance is zero or negative.
func (s LineString) Densify(minDistance float64) LineString {
	return NewLineString(densify(s.seq, minDistance))
}

// SnapToGrid returns a copy of the LineString with all coordinates snapped to
// a base 10 grid.
//
// The grid spacing is specified by the number of decimal places to round to
// (with negative decimal places being allowed). E.g., a decimalPlaces value of
// 2 would cause all coordinates to be rounded to the nearest 0.01, and a
// decimalPlaces of -1 would cause all coordinates to be rounded to the nearest
// 10.
//
// Returned LineStrings may be invalid due to snapping, even if the input
// geometry was valid.
func (s LineString) SnapToGrid(decimalPlaces int) LineString {
	return s.TransformXY(snapToGridXY(decimalPlaces))
}
