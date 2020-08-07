package geom

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"sort"
	"unsafe"

	"github.com/peterstace/simplefeatures/rtree"
)

// MultiPolygon is a planar surface geometry that consists of a collection of
// Polygons. The zero value is the empty MultiPolygon (i.e. the collection of
// zero Polygons). It is immutable after creation.
//
// For a MultiPolygon to be valid, the following assertions must hold:
//
// 1. It must be made up of zero or more valid Polygons (any of which may be empty).
//
// 2. The interiors of any two polygons must not intersect.
//
// 3. The boundaries of any two polygons may touch only at a finite number of points.
type MultiPolygon struct {
	polys []Polygon
	ctype CoordinatesType
}

// NewMultiPolygonFromPolygons creates a MultiPolygon from its constituent
// Polygons. It gives an error if any of the MultiPolygon assertions are not
// maintained. The coordinates type of the MultiPolygon is the lowest common
// coordinates type its Polygons.
func NewMultiPolygonFromPolygons(polys []Polygon, opts ...ConstructorOption) (MultiPolygon, error) {
	if len(polys) == 0 {
		return MultiPolygon{}, nil
	}

	ctype := DimXYZM
	for _, p := range polys {
		ctype &= p.CoordinatesType()
	}
	polys = append([]Polygon(nil), polys...)
	for i := range polys {
		polys[i] = polys[i].ForceCoordinatesType(ctype)
	}

	ctorOpts := newOptionSet(opts)
	if err := validateMultiPolygon(polys, ctorOpts); err != nil {
		if ctorOpts.omitInvalid {
			return MultiPolygon{}, nil
		}
		return MultiPolygon{}, err
	}
	return MultiPolygon{polys, ctype}, nil
}

func validateMultiPolygon(polys []Polygon, opts ctorOptionSet) error {
	if opts.skipValidations {
		return nil
	}

	polyBoundaries := make([]indexedLines, len(polys))
	polyBoundaryPopulated := make([]bool, len(polys))

	var tree rtree.RTree
	for i := range polys {
		env, ok := polys[i].Envelope()
		if !ok {
			continue
		}
		box := env.box()

		err := tree.RangeSearch(box, func(j int) error {
			for _, k := range [...]int{i, j} {
				if !polyBoundaryPopulated[k] {
					polyBoundaries[k] = newIndexedLines(polys[k].Boundary().asLines())
					polyBoundaryPopulated[k] = true
				}
			}

			interMP, interMLS := intersectionOfIndexedLines(
				polyBoundaries[i],
				polyBoundaries[j],
			)
			if !interMLS.IsEmpty() {
				return errors.New("the boundaries of the polygon elements" +
					" of multipolygons must only intersect at points")
			}

			// Fast case: If both the point and line parts of the intersection
			// are empty, then the only thing we have to worry about is one
			// polygon being nested entirely within the other. But since the
			// boundaries don't intersect in any way, we just have to check a
			// single point.
			if interMP.IsEmpty() {
				// We already know the polygons are NOT empty, so it's safe to
				// directly access index 0.
				ptI := polys[i].ExteriorRing().Coordinates().GetXY(0)
				ptJ := polys[j].ExteriorRing().Coordinates().GetXY(0)
				if relatePointToPolygon(ptI, polyBoundaries[j]) != exterior ||
					relatePointToPolygon(ptJ, polyBoundaries[i]) != exterior {
					return errors.New("polygons must not be nested")
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
		})
		if err != nil {
			return err
		}

		tree.Insert(box, i)
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
		p1.tree.RangeSearch(p2.lines[j].envelope().box(), func(i int) error {
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
				return fmt.Errorf("polygon interiors intersect at %s",
					NewPointFromXY(midpoint).AsText())
			}
		}
	}
	return nil
}

func sortAndUniquifyXYs(xys []XY) []XY {
	if len(xys) == 0 {
		return xys
	}
	sort.Slice(xys, func(i, j int) bool {
		ptI := xys[i]
		ptJ := xys[j]
		if ptI.X != ptJ.X {
			return ptI.X < ptJ.X
		}
		return ptI.Y < ptJ.Y
	})
	n := 1
	for i := 1; i < len(xys); i++ {
		if xys[i] != xys[i-1] {
			xys[n] = xys[i]
			n++
		}
	}
	return xys[:n]
}

// Type returns the GeometryType for a MultiPolygon
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

// Intersects return true if and only if this geometry intersects with the
// other, i.e. they shared at least one common point.
func (m MultiPolygon) Intersects(g Geometry) bool {
	return hasIntersection(m.AsGeometry(), g)
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

// Envelope returns the Envelope that most tightly surrounds the geometry. If
// the geometry is empty, then false is returned.
func (m MultiPolygon) Envelope() (Envelope, bool) {
	var env Envelope
	var has bool
	for _, poly := range m.polys {
		e, ok := poly.Envelope()
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
	return NewMultiLineStringFromLineStrings(bounds)
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (m MultiPolygon) Value() (driver.Value, error) {
	return m.AsBinary(), nil
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (m MultiPolygon) AsBinary() []byte {
	return m.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (m MultiPolygon) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaller(dst)
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

// MarshalJSON implements the encoding/json.Marshaller interface by encoding
// this geometry as a GeoJSON geometry object.
func (m MultiPolygon) MarshalJSON() ([]byte, error) {
	var dst []byte
	dst = append(dst, `{"type":"MultiPolygon","coordinates":`...)
	dst = appendGeoJSONSequenceMatrix(dst, m.Coordinates())
	dst = append(dst, '}')
	return dst, nil
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
func (m MultiPolygon) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (MultiPolygon, error) {
	polys := make([]Polygon, m.NumPolygons())
	for i := range polys {
		transformed, err := m.PolygonN(i).TransformXY(fn, opts...)
		if err != nil {
			return MultiPolygon{}, err
		}
		polys[i] = transformed
	}
	mp, err := NewMultiPolygonFromPolygons(polys, opts...)
	return mp.ForceCoordinatesType(m.ctype), err
}

// EqualsExact checks if this MultiPolygon is exactly equal to another MultiPolygon.
func (m MultiPolygon) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	return other.IsMultiPolygon() &&
		multiPolygonExactEqual(m, other.AsMultiPolygon(), opts)
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
	return NewPointFromXY(weightedCentroid)
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
	return m.forceOrientation(true)
}

// ForceCCW returns the equivalent MultiPolygon that has its exterior rings in
// a counter-clockwise orientation and any inner rings in a clockwise
// orientation.
func (m MultiPolygon) ForceCCW() MultiPolygon {
	return m.forceOrientation(false)
}

func (m MultiPolygon) forceOrientation(forceCW bool) MultiPolygon {
	polys := make([]Polygon, len(m.polys))
	for i, poly := range m.polys {
		polys[i] = poly.forceOrientation(forceCW)
	}
	return MultiPolygon{polys, m.ctype}
}
