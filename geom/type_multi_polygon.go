package geom

import (
	"database/sql/driver"
	"errors"
	"sort"
	"unsafe"
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

	if skipValidations(opts) {
		return MultiPolygon{polys, ctype}, nil
	}

	type interval struct {
		minX, maxX float64
	}
	intervals := make([]interval, len(polys))
	for i := range intervals {
		env, ok := polys[i].Envelope()
		if ok {
			intervals[i].minX = env.Min().X
			intervals[i].maxX = env.Max().X
		}
	}
	indexes := intSequence(len(polys))
	sort.Slice(indexes, func(i, j int) bool {
		// Empty Polygons with have an interval of (0, 0).
		xi := intervals[indexes[i]].minX
		xj := intervals[indexes[j]].minX
		return xi < xj
	})

	active := intHeap{less: func(i, j int) bool {
		xi := intervals[i].maxX
		xj := intervals[j].maxX
		return xi < xj
	}}

	for _, i := range indexes {
		if polys[i].IsEmpty() {
			continue
		}
		currentX := intervals[i].minX
		for len(active.data) > 0 && intervals[active.data[0]].maxX < currentX {
			active.pop()
		}
		for _, j := range active.data {
			if polys[j].IsEmpty() {
				continue
			}
			bound1 := polys[i].Boundary()
			bound2 := polys[j].Boundary()
			inter := mustIntersection(bound1.AsGeometry(), bound2.AsGeometry())
			if inter.Dimension() > 0 {
				return MultiPolygon{}, errors.New("the boundaries of the polygon elements of multipolygons must only intersect at points")
			}
			if polyInteriorsIntersect(polys[i], polys[j]) {
				return MultiPolygon{}, errors.New("polygon interiors must not intersect")
			}
		}
		active.push(i)
	}

	return MultiPolygon{polys, ctype}, nil
}

func polyInteriorsIntersect(p1, p2 Polygon) bool {
	// Run twice, swapping the order of the polygons each time.
	for order := 0; order < 2; order++ {
		p1, p2 = p2, p1

		// Collect points along the boundary of the first polygon. Do this by
		// first breaking the lines in the ring into multiple segments where
		// they are intersected by the rings from the other polygon. Collect
		// the original points in the boundary, plus the intersection points,
		// then each midpoint between those points. These are enough points
		// that one of the points will be inside the other polygon iff the
		// interior of the polygons intersect.
		allPts := make(map[XY]struct{})
		for _, r1 := range p1.rings {
			seq1 := r1.Coordinates()
			for idx1 := 0; idx1 < seq1.Length(); idx1++ {
				line1, ok := getLine(seq1, idx1)
				if !ok {
					continue
				}
				// Collect boundary control points and intersection points.
				linePts := make(map[XY]struct{})
				linePts[line1.a.XY] = struct{}{}
				linePts[line1.b.XY] = struct{}{}
				for _, r2 := range p2.rings {
					seq2 := r2.Coordinates()
					for idx2 := 0; idx2 < seq2.Length(); idx2++ {
						line2, ok := getLine(seq2, idx2)
						if !ok {
							continue
						}
						inter := intersectLineWithLineNoAlloc(line1, line2)
						if inter.empty {
							continue
						}
						if inter.ptA != inter.ptB {
							continue
						}
						if inter.ptA != line1.StartPoint().XY && inter.ptA != line1.EndPoint().XY {
							linePts[inter.ptA] = struct{}{}
						}
					}
				}
				// Collect midpoints.
				if len(linePts) <= 2 {
					for pt := range linePts {
						allPts[pt] = struct{}{}
					}
				} else {
					linePtsSlice := make([]XY, 0, len(linePts))
					for pt := range linePts {
						linePtsSlice = append(linePtsSlice, pt)
					}
					sort.Slice(linePtsSlice, func(i, j int) bool {
						ptI := linePtsSlice[i]
						ptJ := linePtsSlice[j]
						if ptI.X != ptJ.X {
							return ptI.X < ptJ.X
						}
						return ptI.Y < ptJ.Y
					})
					allPts[linePtsSlice[0]] = struct{}{}
					for i := 0; i+1 < len(linePtsSlice); i++ {
						midpoint := linePtsSlice[i].Midpoint(linePtsSlice[i+1])
						allPts[midpoint] = struct{}{}
						allPts[linePtsSlice[i+1]] = struct{}{}
					}
				}
			}
		}

		// Check to see if any of the points from the boundary from the first
		// polygon are inside the second polygon.
		for pt := range allPts {
			if isPointInteriorToPolygon(pt, p2) {
				return true
			}
		}
	}
	return false
}

// isPointInteriorToPolygon returns true iff the pt is strictly inside the
// polygon (i.e. is not outside the polygon or on its boundary).
func isPointInteriorToPolygon(pt XY, poly Polygon) bool {
	if poly.IsEmpty() || pointRingSide(pt, poly.ExteriorRing()) != interior {
		return false
	}
	for i := 0; i < poly.NumInteriorRings(); i++ {
		hole := poly.InteriorRingN(i)
		if pointRingSide(pt, hole) != exterior {
			return false
		}
	}
	return true
}

// Type return type string for MultiPolygon
func (m MultiPolygon) Type() string {
	return multiPolygonType
}

// AsGeometry converts this MultiPolygon into a Geometry.
func (m MultiPolygon) AsGeometry() Geometry {
	return Geometry{multiPolygonTag, unsafe.Pointer(&m)}
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

// Intersection calculates the of this geometry and another, i.e. the portion
// of the two geometries that are shared. It is not implemented for all
// geometry pairs, and returns an error for those cases.
func (m MultiPolygon) Intersection(g Geometry) (Geometry, error) {
	return intersection(m.AsGeometry(), g)
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
	marsh.writeGeomType(wkbGeomTypeMultiPolygon, m.ctype)
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

// IsValid checks if this MultiPolygon is valid
func (m MultiPolygon) IsValid() bool {
	for _, p := range m.polys {
		if !p.IsValid() {
			return false
		}
	}
	_, err := NewMultiPolygonFromPolygons(m.polys)
	return err == nil
}

// Area in the case of a MultiPolygon is the sum of the areas of its polygons.
func (m MultiPolygon) Area() float64 {
	var area float64
	n := m.NumPolygons()
	for i := 0; i < n; i++ {
		area += m.PolygonN(i).Area()
	}
	return area
}

// SignedArea returns the sum of the signed areas of the constituent polygons.
func (m MultiPolygon) SignedArea() float64 {
	var signedArea float64
	n := m.NumPolygons()
	for i := 0; i < n; i++ {
		signedArea += m.PolygonN(i).SignedArea()
	}
	return signedArea
}

// Centroid returns the multi polygon's centroid point. It returns the empty
// Point if the multi polygon is empty.
func (m MultiPolygon) Centroid() Point {
	var sumArea float64
	var sumXY XY
	n := m.NumPolygons()
	for i := 0; i < n; i++ {
		xy, area := sumCentroidAndAreaOfPolygon(m.PolygonN(i))
		sumXY = sumXY.Add(xy)
		sumArea += area
	}
	if sumArea == 0 {
		return NewEmptyPoint(DimXY)
	}
	return NewPointFromXY(sumXY.Scale(1.0 / sumArea))
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
