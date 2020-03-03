package geom

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"io"
	"sort"
	"unsafe"
)

// MultiPolygon is a multi surface whose elements are polygons.
//
// Its assertions are:
//
// 1. It must be made up of zero or more valid Polygons (any of which may be empty).
//
// 2. The interiors of any two polygons must not intersect.
//
// 3. The boundaries of any two polygons may touch only at a finite number of points.
type MultiPolygon struct {
	polys []Polygon
}

// NewEmptyMultiPolygon returns the empty MultiPolygon. It is equivalent to
// calling NewMultiPolygon with a zero length polygon slice.
func NewEmptyMultiPolygon() MultiPolygon {
	return MultiPolygon{}
}

// NewMultiPolygon creates a MultiPolygon from its constituent Polygons. It
// gives an error if any of the MultiPolygon assertions are not maintained.
func NewMultiPolygon(polys []Polygon, opts ...ConstructorOption) (MultiPolygon, error) {
	if skipValidations(opts) {
		return MultiPolygon{polys}, nil
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
	indexes := seq(len(polys))
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

	return MultiPolygon{polys}, nil
}

// NewMultiPolygonC creates a new MultiPolygon from its constituent Coordinate values.
func NewMultiPolygonC(coords [][][]Coordinates, opts ...ConstructorOption) (MultiPolygon, error) {
	var polys []Polygon
	for _, c := range coords {
		poly, err := NewPolygonC(c, opts...)
		if err != nil {
			return MultiPolygon{}, err
		}
		polys = append(polys, poly)
	}
	return NewMultiPolygon(polys, opts...)
}

// NewMultiPolygonXY creates a new MultiPolygon from its constituent XY values.
func NewMultiPolygonXY(pts [][][]XY, opts ...ConstructorOption) (MultiPolygon, error) {
	return NewMultiPolygonC(threeDimXYToCoords(pts), opts...)
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
			for ln1 := 0; ln1 < r1.NumLines(); ln1++ {
				line1 := r1.LineN(ln1)
				// Collect boundary control points and intersection points.
				linePts := make(map[XY]struct{})
				linePts[line1.a.XY] = struct{}{}
				linePts[line1.b.XY] = struct{}{}
				for _, r2 := range p2.rings {
					for ln2 := 0; ln2 < r2.NumLines(); ln2++ {
						line2 := r2.LineN(ln2)
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

func (m MultiPolygon) AsText() string {
	return string(m.AppendWKT(nil))
}

func (m MultiPolygon) AppendWKT(dst []byte) []byte {
	dst = append(dst, "MULTIPOLYGON"...)
	if len(m.polys) == 0 {
		return append(dst, " EMPTY"...)
	}
	dst = append(dst, '(')
	for i, poly := range m.polys {
		dst = poly.appendWKTBody(dst)
		if i != len(m.polys)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ')')
}

// IsSimple returns true. All MultiPolygons are simple by definition.
func (m MultiPolygon) IsSimple() bool {
	return true
}

func (m MultiPolygon) Intersects(g Geometry) bool {
	return hasIntersection(m.AsGeometry(), g)
}

func (m MultiPolygon) Intersection(g Geometry) (Geometry, error) {
	return intersection(m.AsGeometry(), g)
}

func (m MultiPolygon) IsEmpty() bool {
	for _, p := range m.polys {
		if !p.IsEmpty() {
			return false
		}
	}
	return true
}

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

func (m MultiPolygon) Boundary() MultiLineString {
	var n int
	for _, p := range m.polys {
		n += len(p.rings)
	}
	bounds := make([]LineString, n)
	var i int
	for _, p := range m.polys {
		copy(bounds[i:], p.rings)
		i += len(p.rings)
	}
	return NewMultiLineString(bounds)
}

func (m MultiPolygon) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := m.AsBinary(&buf)
	return buf.Bytes(), err
}

func (m MultiPolygon) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeMultiPolygon)
	n := m.NumPolygons()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		poly := m.PolygonN(i)
		marsh.setErr(poly.AsBinary(w))
	}
	return marsh.err
}

func (m MultiPolygon) ConvexHull() Geometry {
	return convexHull(m.AsGeometry())
}

func (m MultiPolygon) MarshalJSON() ([]byte, error) {
	return marshalGeoJSON("MultiPolygon", m.Coordinates())
}

// Coordinates returns the coordinates of each constituent Polygon of the
// MultiPolygon.
func (m MultiPolygon) Coordinates() [][][]Coordinates {
	numPolys := m.NumPolygons()
	coords := make([][][]Coordinates, numPolys)
	for i := 0; i < numPolys; i++ {
		coords[i] = m.PolygonN(i).Coordinates()
	}
	return coords
}

// TransformXY transforms this MultiPolygon into another MultiPolygon according to fn.
func (m MultiPolygon) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (MultiPolygon, error) {
	coords := m.Coordinates()
	transform3dCoords(coords, fn)
	mp, err := NewMultiPolygonC(coords, opts...)
	return mp, err
}

// EqualsExact checks if this MultiPolygon is exactly equal to another MultiPolygon.
func (m MultiPolygon) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	return other.IsMultiPolygon() &&
		multiPolygonExactEqual(m, other.AsMultiPolygon(), opts)
}

// IsValid checks if this MultiPolygon is valid
func (m MultiPolygon) IsValid() bool {
	_, err := NewMultiPolygonC(m.Coordinates())
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
		return NewEmptyPoint()
	}
	return NewPointXY(sumXY.Scale(1.0 / sumArea))
}

// Reverse in the case of MultiPolygon outputs the component polygons in their original order,
// each individually reversed.
func (m MultiPolygon) Reverse() MultiPolygon {
	polys := make([]Polygon, len(m.polys))
	// Form the reversed slice.
	for i := 0; i < len(m.polys); i++ {
		polys[i] = m.polys[i].Reverse()
	}
	m2, err := NewMultiPolygon(polys)
	if err != nil {
		panic("Reverse of an existing MultiPolygon should not fail")
	}
	return m2
}
