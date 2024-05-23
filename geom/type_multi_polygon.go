package geom

import (
	"database/sql/driver"
	"fmt"
	"unsafe"

	"github.com/peterstace/simplefeatures/rtree"
)

// MultiPolygon is a planar surface geometry that consists of a collection of
// (possibly empty) Polygons. The zero value is the empty MultiPolygon (i.e.
// the collection of zero Polygons). It is immutable after creation.
type MultiPolygon struct {
	// Invariant: ctype matches the coordinates type of each polygon.
	polys []Polygon
	ctype CoordinatesType
}

// NewMultiPolygon creates a MultiPolygon from its constituent Polygons. The
// coordinates type of the MultiPolygon is the lowest common coordinates type
// of its Polygons.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiPolygon(polys []Polygon) MultiPolygon {
	ctype := DimXY
	if len(polys) > 0 {
		ctype = DimXYZM
		for _, p := range polys {
			ctype &= p.CoordinatesType()
		}
	}
	polys = append([]Polygon(nil), polys...)
	for i := range polys {
		polys[i] = polys[i].ForceCoordinatesType(ctype)
	}
	return MultiPolygon{polys, ctype}
}

// Validate checks if the MultiPolygon is valid.
//
// The Polygons that make up a MultiPolygon are constrained by the following
// rules:
//
//  1. Each child polygon must be valid.
//  2. The interiors of any two child polygons must not intersect.
//  3. The boundaries of any two child polygons may touch only at a finite
//     number of points.
func (m MultiPolygon) Validate() error {
	for i, poly := range m.polys {
		if err := poly.Validate(); err != nil {
			return wrap(err, "validating polygon at index %d", i)
		}
	}
	err := m.checkMultiPolygonConstraints()
	return wrap(err, "validating multipolygon constraints")
}

func (m MultiPolygon) checkMultiPolygonConstraints() error {
	polyBoundaries := make([]indexedLines, len(m.polys))
	polyBoundaryPopulated := make([]bool, len(m.polys))

	// Construct RTree of Polygons.
	boxes := make([]rtree.Box, len(m.polys))
	items := make([]rtree.BulkItem, 0, len(m.polys))
	for i, p := range m.polys {
		if box, ok := p.Envelope().AsBox(); ok {
			boxes[i] = box
			item := rtree.BulkItem{Box: boxes[i], RecordID: i}
			items = append(items, item)
		}
	}
	tree := rtree.BulkLoad(items)

	for i := range m.polys {
		if m.polys[i].IsEmpty() {
			continue
		}
		if err := tree.RangeSearch(boxes[i], func(j int) error {
			// Only consider each pair of polygons once.
			if i <= j {
				return nil
			}

			for _, k := range [...]int{i, j} {
				if !polyBoundaryPopulated[k] {
					polyBoundaries[k] = newIndexedLines(m.polys[k].Boundary().asLines())
					polyBoundaryPopulated[k] = true
				}
			}
			interMP, interMLS := intersectionOfIndexedLines(
				polyBoundaries[i],
				polyBoundaries[j],
			)
			if !interMLS.IsEmpty() {
				return violatePolysMultiTouch.errAtPt(
					arbitraryControlPoint(interMLS.AsGeometry()))
			}

			// Fast case: If both the point and line parts of the intersection
			// are empty, then the only thing we have to worry about is one
			// polygon being nested entirely within the other. But since the
			// boundaries don't intersect in any way, we just have to check a
			// single point.
			if interMP.IsEmpty() {
				// We already know the polygons are NOT empty, so it's safe to
				// directly access index 0.
				for _, dir := range []struct{ inIdx, outIdx int }{{i, j}, {j, i}} {
					in := m.polys[dir.inIdx].ExteriorRing().Coordinates().GetXY(0)
					out := polyBoundaries[dir.outIdx]
					if relatePointToPolygon(in, out) != exterior {
						return violatePolysMultiTouch.errAtXY(in)
					}
				}
				return nil
			}

			// Slow case: The boundaries intersect at a point (or many points).
			// But we still need to ensure that the interiors don't intersect.
			for _, pair := range [...]struct{ pb1, pb2 indexedLines }{
				{polyBoundaries[i], polyBoundaries[j]},
				{polyBoundaries[j], polyBoundaries[i]},
			} {
				if err := validatePolyNotInsidePoly(pair.pb1, pair.pb2); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func validatePolyNotInsidePoly(p1, p2 indexedLines) error {
	// For each point where the boundaries of the two polygons intersect, we
	// take points clockwise and counterclockwise along the boundaries and see
	// if they are inside the opposing polygon. If they are, then the interiors
	// intersect.

	for j := range p2.lines {
		// Find intersection points.
		var pts []XY
		p1.tree.RangeSearch(p2.lines[j].box(), func(i int) error {
			inter := p1.lines[i].intersectLine(p2.lines[j])
			if inter.empty {
				return nil
			}
			if inter.ptA != inter.ptB {
				panic(fmt.Sprintf("already established that boundaries only "+
					"intersect at points, but got: %v", inter))
			}
			pts = append(pts, inter.ptA)
			return nil
		})
		if len(pts) == 0 {
			continue
		}

		// Construct midpoints between intersection points and endpoints.
		pts = append(pts, p2.lines[j].a, p2.lines[j].b)
		pts = sortAndUniquifyXYs(pts)

		// Check if midpoints are inside the other polygon.
		for k := 0; k+1 < len(pts); k++ {
			midpoint := pts[k].Add(pts[k+1]).Scale(0.5)
			if relatePointToPolygon(midpoint, p1) == interior {
				return violatePolysMultiTouch.errAtXY(midpoint)
			}
		}
	}
	return nil
}

// Type returns the GeometryType for a MultiPolygon.
func (m MultiPolygon) Type() GeometryType {
	return TypeMultiPolygon
}

// AsGeometry converts this MultiPolygon into a Geometry.
func (m MultiPolygon) AsGeometry() Geometry {
	return Geometry{TypeMultiPolygon, unsafe.Pointer(&m)}
}

// NumPolygons gives the number of Polygon elements in the MultiPolygon.
func (m MultiPolygon) NumPolygons() int {
	return len(m.polys)
}

// PolygonN gives the nth (zero based) Polygon element.
func (m MultiPolygon) PolygonN(n int) Polygon {
	return m.polys[n]
}

// AsText returns the WKT (Well Known Text) representation of this geometry.
func (m MultiPolygon) AsText() string {
	return string(m.AppendWKT(nil))
}

// AppendWKT appends the WKT (Well Known Text) representation of this geometry
// to the input byte slice.
func (m MultiPolygon) AppendWKT(dst []byte) []byte {
	dst = appendWKTHeader(dst, "MULTIPOLYGON", m.ctype)
	if len(m.polys) == 0 {
		return appendWKTEmpty(dst)
	}
	dst = append(dst, '(')
	for i, poly := range m.polys {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = poly.appendWKTBody(dst)
	}
	return append(dst, ')')
}

// IsSimple returns true if this geometry contains no anomalous geometry
// points, such as self intersection or self tangency. Because MultiPolygons
// are always simple, this method always returns true.
func (m MultiPolygon) IsSimple() bool {
	return true
}

// IsEmpty return true if and only if this MultiPolygon doesn't contain any
// Polygons, or only contains empty Polygons.
func (m MultiPolygon) IsEmpty() bool {
	for _, p := range m.polys {
		if !p.IsEmpty() {
			return false
		}
	}
	return true
}

// Envelope returns the Envelope that most tightly surrounds the geometry.
func (m MultiPolygon) Envelope() Envelope {
	var env Envelope
	for _, poly := range m.polys {
		env = env.ExpandToIncludeEnvelope(poly.Envelope())
	}
	return env
}

// Boundary returns the spatial boundary of this MultiPolygon. This is the
// MultiLineString containing the boundaries of the MultiPolygon's elements.
func (m MultiPolygon) Boundary() MultiLineString {
	var n int
	for _, p := range m.polys {
		n += len(p.rings)
	}
	bounds := make([]LineString, 0, n)
	for _, p := range m.polys {
		for _, r := range p.rings {
			bounds = append(bounds, r.Force2D())
		}
	}
	return NewMultiLineString(bounds)
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (m MultiPolygon) Value() (driver.Value, error) {
	return m.AsBinary(), nil
}

// Scan implements the database/sql.Scanner interface by parsing the src value
// as WKB (Well Known Binary).
//
// If the WKB doesn't represent a MultiPolygon geometry, then an error is returned.
//
// Geometry constraint validation is performed on the resultant geometry (an
// error will be returned if the geometry is invalid). If this validation isn't
// needed or is undesirable, then the WKB should be scanned into a byte slice
// and then UnmarshalWKB called manually (passing in NoValidate{}).
func (m *MultiPolygon) Scan(src interface{}) error {
	return scanAsType(src, m)
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (m MultiPolygon) AsBinary() []byte {
	return m.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (m MultiPolygon) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaler(dst)
	marsh.writeByteOrder()
	marsh.writeGeomType(TypeMultiPolygon, m.ctype)
	n := m.NumPolygons()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		poly := m.PolygonN(i)
		marsh.buf = poly.AppendWKB(marsh.buf)
	}
	return marsh.buf
}

// ConvexHull returns the geometry representing the smallest convex geometry
// that contains this geometry.
func (m MultiPolygon) ConvexHull() Geometry {
	return convexHull(m.AsGeometry())
}

// MarshalJSON implements the encoding/json.Marshaler interface by encoding
// this geometry as a GeoJSON geometry object.
func (m MultiPolygon) MarshalJSON() ([]byte, error) {
	var dst []byte
	dst = append(dst, `{"type":"MultiPolygon","coordinates":`...)
	dst = appendGeoJSONSequenceMatrix(dst, m.Coordinates())
	dst = append(dst, '}')
	return dst, nil
}

// UnmarshalJSON implements the encoding/json.Unmarshaler interface by decoding
// the GeoJSON representation of a MultiPolygon.
func (m *MultiPolygon) UnmarshalJSON(buf []byte) error {
	return unmarshalGeoJSONAsType(buf, m)
}

// Coordinates returns the coordinates of each constituent Polygon of the
// MultiPolygon.
func (m MultiPolygon) Coordinates() [][]Sequence {
	numPolys := m.NumPolygons()
	coords := make([][]Sequence, numPolys)
	for i := 0; i < numPolys; i++ {
		coords[i] = m.PolygonN(i).Coordinates()
	}
	return coords
}

// TransformXY transforms this MultiPolygon into another MultiPolygon according to fn.
func (m MultiPolygon) TransformXY(fn func(XY) XY) MultiPolygon {
	polys := make([]Polygon, m.NumPolygons())
	for i := range polys {
		polys[i] = m.PolygonN(i).TransformXY(fn)
	}
	mp := NewMultiPolygon(polys)
	return mp.ForceCoordinatesType(m.ctype)
}

// Area in the case of a MultiPolygon is the sum of the areas of its polygons.
func (m MultiPolygon) Area(opts ...AreaOption) float64 {
	var area float64
	n := m.NumPolygons()
	for i := 0; i < n; i++ {
		area += m.PolygonN(i).Area(opts...)
	}
	return area
}

// Centroid returns the multi polygon's centroid point. It returns the empty
// Point if the multi polygon is empty.
func (m MultiPolygon) Centroid() Point {
	if m.IsEmpty() {
		return NewEmptyPoint(DimXY)
	}

	areas := make([]float64, m.NumPolygons())
	var totalArea float64
	for i := 0; i < m.NumPolygons(); i++ {
		area := m.PolygonN(i).Area()
		areas[i] = area
		totalArea += area
	}

	var weightedCentroid XY
	for i := 0; i < m.NumPolygons(); i++ {
		centroid, ok := m.PolygonN(i).Centroid().XY()
		if ok {
			weightedCentroid = weightedCentroid.Add(centroid.Scale(areas[i] / totalArea))
		}
	}
	return weightedCentroid.AsPoint()
}

// Reverse in the case of MultiPolygon outputs the component polygons in their original order,
// each individually reversed.
func (m MultiPolygon) Reverse() MultiPolygon {
	polys := make([]Polygon, len(m.polys))
	// Form the reversed slice.
	for i := 0; i < len(m.polys); i++ {
		polys[i] = m.polys[i].Reverse()
	}
	return MultiPolygon{polys, m.ctype}
}

// CoordinatesType returns the CoordinatesType used to represent points making
// up the geometry.
func (m MultiPolygon) CoordinatesType() CoordinatesType {
	return m.ctype
}

// ForceCoordinatesType returns a new MultiPolygon with a different CoordinatesType. If a
// dimension is added, then new values are populated with 0.
func (m MultiPolygon) ForceCoordinatesType(newCType CoordinatesType) MultiPolygon {
	flat := make([]Polygon, len(m.polys))
	for i := range m.polys {
		flat[i] = m.polys[i].ForceCoordinatesType(newCType)
	}
	return MultiPolygon{flat, newCType}
}

// Force2D returns a copy of the MultiPolygon with Z and M values removed.
func (m MultiPolygon) Force2D() MultiPolygon {
	return m.ForceCoordinatesType(DimXY)
}

// PointOnSurface returns a Point on the interior of the MultiPolygon.
func (m MultiPolygon) PointOnSurface() Point {
	var (
		bestWidth float64
		bestPoint Point
	)
	for i := 0; i < m.NumPolygons(); i++ {
		poly := m.PolygonN(i)
		point, bisectorWidth := pointOnAreaSurface(poly)
		if point.IsEmpty() {
			continue
		}
		if bisectorWidth > bestWidth {
			bestWidth = bisectorWidth
			bestPoint = point
		}
	}
	return bestPoint
}

// ForceCW returns the equivalent MultiPolygon that has its exterior rings in a
// clockwise orientation and any inner rings in a counter-clockwise
// orientation.
func (m MultiPolygon) ForceCW() MultiPolygon {
	if m.IsCW() {
		return m
	}
	return m.forceOrientation(true)
}

// ForceCCW returns the equivalent MultiPolygon that has its exterior rings in
// a counter-clockwise orientation and any inner rings in a clockwise
// orientation.
func (m MultiPolygon) ForceCCW() MultiPolygon {
	if m.IsCCW() {
		return m
	}
	return m.forceOrientation(false)
}

func (m MultiPolygon) forceOrientation(forceCW bool) MultiPolygon {
	polys := make([]Polygon, len(m.polys))
	for i, poly := range m.polys {
		polys[i] = poly.forceOrientation(forceCW)
	}
	return MultiPolygon{polys, m.ctype}
}

// IsCW returns true iff all contained polygons are CW.
// An empty multipolygon returns true.
func (m MultiPolygon) IsCW() bool {
	for _, poly := range m.polys {
		if !poly.IsCW() {
			return false
		}
	}
	return true
}

// IsCCW returns true iff all contained polygons are CCW.
// An empty multipolygon returns true.
func (m MultiPolygon) IsCCW() bool {
	for _, poly := range m.polys {
		if !poly.IsCCW() {
			return false
		}
	}
	return true
}

func (m MultiPolygon) controlPoints() int {
	var sum int
	for _, p := range m.polys {
		sum += p.controlPoints()
	}
	return sum
}

// Dump returns the MultiPolygon represented as a Polygon slice.
func (m MultiPolygon) Dump() []Polygon {
	ps := make([]Polygon, len(m.polys))
	copy(ps, m.polys)
	return ps
}

// DumpCoordinates returns the points making up the rings in a MultiPolygon as
// a Sequence.
func (m MultiPolygon) DumpCoordinates() Sequence {
	var n int
	for _, p := range m.polys {
		for _, r := range p.rings {
			n += r.Coordinates().Length()
		}
	}
	ctype := m.CoordinatesType()
	coords := make([]float64, 0, n*ctype.Dimension())

	for _, p := range m.polys {
		for _, r := range p.rings {
			coords = r.Coordinates().appendAllPoints(coords)
		}
	}

	seq := NewSequence(coords, ctype)
	seq.assertNoUnusedCapacity()
	return seq
}

// Summary returns a text summary of the MultiPolygon following a similar format to https://postgis.net/docs/ST_Summary.html.
func (m MultiPolygon) Summary() string {
	numPoints := m.DumpCoordinates().Length()

	var polygonSuffix string
	numPolygons := m.NumPolygons()
	if numPolygons != 1 {
		polygonSuffix = "s"
	}

	var numRings int
	for _, polygon := range m.polys {
		numRings += polygon.NumRings()
	}

	var ringSuffix string
	if numRings != 1 {
		ringSuffix = "s"
	}
	return fmt.Sprintf("%s[%s] with %d polygon%s consisting of %d total ring%s and %d total points",
		m.Type(), m.CoordinatesType(), numPolygons, polygonSuffix, numRings, ringSuffix, numPoints)
}

// String returns the string representation of the MultiPolygon.
func (m MultiPolygon) String() string {
	return m.Summary()
}

// Simplify returns a simplified version of the MultiPolygon by applying
// Simplify to each child Polygon and constructing a new MultiPolygon from the
// result. If the result is invalid, then an error is returned. Geometry
// constraint validation can be skipped by passing in NoValidate{}, potentially
// resulting in an invalid geometry being returned.
func (m MultiPolygon) Simplify(threshold float64, nv ...NoValidate) (MultiPolygon, error) {
	n := m.NumPolygons()
	polys := make([]Polygon, 0, n)
	for i := 0; i < n; i++ {
		poly, err := m.PolygonN(i).Simplify(threshold, nv...)
		if err != nil {
			return MultiPolygon{}, err
		}
		if !poly.IsEmpty() {
			polys = append(polys, poly)
		}
	}
	simpl := NewMultiPolygon(polys)
	if len(nv) == 0 {
		if err := simpl.Validate(); err != nil {
			return MultiPolygon{}, wrapSimplified(err)
		}
	}
	return simpl.ForceCoordinatesType(m.CoordinatesType()), nil
}

// Densify returns a new MultiPolygon with additional linearly interpolated
// control points such that the distance between any two consecutive control
// points is at most the given maxDistance.
//
// Panics if maxDistance is zero or negative.
func (m MultiPolygon) Densify(maxDistance float64) MultiPolygon {
	ps := make([]Polygon, len(m.polys))
	for i, p := range m.polys {
		ps[i] = p.Densify(maxDistance)
	}
	return MultiPolygon{ps, m.ctype}
}

// SnapToGrid returns a MultiPolygon of the geometry with all coordinates
// snapped to a base 10 grid.
//
// The grid spacing is specified by the number of decimal places to round to
// (with negative decimal places being allowed). E.g., a decimalPlaces value of
// 2 would cause all coordinates to be rounded to the nearest 0.01, and a
// decimalPlaces of -1 would cause all coordinates to be rounded to the nearest
// 10.
//
// Returned MultiPolygons may be invalid due to snapping, even if the input
// geometry was valid.
func (m MultiPolygon) SnapToGrid(decimalPlaces int) MultiPolygon {
	return m.TransformXY(snapToGridXY(decimalPlaces))
}
