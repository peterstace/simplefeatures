package geom

import (
	"database/sql/driver"
	"errors"
	"io"
	"sort"
)

// MultiPolygon is a multi surface whose elements are polygons.
//
// Its assertions are:
//
// 1. It must be made up of zero or more valid Polygons.
//
// 2. The interiors of any two polygons must not intersect.
//
// 3. The boundaries of any two polygons may touch only at a finite number of points.
type MultiPolygon struct {
	polys []Polygon
}

// NewMultiPolygon creates a MultiPolygon from its constituent Polygons. It
// gives an error if any of the MultiPolygon assertions are not maintained.
func NewMultiPolygon(polys []Polygon, opts ...ConstructorOption) (MultiPolygon, error) {
	if !doExpensiveValidations(opts) {
		return MultiPolygon{polys}, nil
	}
	for i := 0; i < len(polys); i++ {
		for j := i + 1; j < len(polys); j++ {
			bound1 := polys[i].Boundary()
			bound2 := polys[j].Boundary()
			inter := mustIntersection(bound1, bound2)
			if inter.Dimension() > 0 {
				return MultiPolygon{}, errors.New("the boundaries of the polygon elements of multipolygons must only intersect at points")
			}
			if polyInteriorsIntersect(polys[i], polys[j]) {
				return MultiPolygon{}, errors.New("polygon interiors must not intersect")
			}
		}
	}
	return MultiPolygon{polys}, nil
}

// NewMultiPolygonC creates a new MultiPolygon from its constituent Coordinate values.
func NewMultiPolygonC(coords [][][]Coordinates, opts ...ConstructorOption) (MultiPolygon, error) {
	var polys []Polygon
	for _, c := range coords {
		if len(c) == 0 {
			continue
		}
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
		for _, r1 := range p1.rings() {
			for _, line1 := range r1.lines {
				// Collect boundary control points and intersection points.
				linePts := make(map[XY]struct{})
				linePts[line1.a.XY] = struct{}{}
				linePts[line1.b.XY] = struct{}{}
				for _, r2 := range p2.rings() {
					for _, line2 := range r2.lines {
						env, ok := mustIntersection(line1, line2).Envelope()
						if !ok {
							continue
						}
						if !env.Min().Equals(env.Max()) {
							continue
						}
						inter := env.Min()
						if !inter.Equals(line1.a.XY) && !inter.Equals(line1.b.XY) {
							linePts[inter] = struct{}{}
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

func isPointInteriorToPolygon(pt XY, poly Polygon) bool {
	if pointRingSide(pt, poly.outer) != interior {
		return false
	}
	for _, hole := range poly.holes {
		if pointRingSide(pt, hole) != exterior {
			return false
		}
	}
	return true
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
	dst = append(dst, []byte("MULTIPOLYGON")...)
	if len(m.polys) == 0 {
		return append(dst, []byte(" EMPTY")...)
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

func (m MultiPolygon) Intersects(g Geometry) (bool, error) {
	has, _, err := hasIntersection(m, g)
	return has, err
}

func (m MultiPolygon) Intersection(g Geometry) (Geometry, error) {
	return intersection(m, g)
}

func (m MultiPolygon) IsEmpty() bool {
	return len(m.polys) == 0
}

func (m MultiPolygon) Dimension() int {
	return 2
}

func (m MultiPolygon) Equals(other Geometry) (bool, error) {
	return equals(m, other)
}

func (m MultiPolygon) Envelope() (Envelope, bool) {
	if len(m.polys) == 0 {
		return Envelope{}, false
	}
	env := mustEnvelope(m.polys[0])
	for _, poly := range m.polys[1:] {
		env = env.ExpandToIncludeEnvelope(mustEnvelope(poly))
	}
	return env, true
}

func (m MultiPolygon) Boundary() Geometry {
	if m.IsEmpty() {
		return m
	}
	var bounds []LineString
	for _, poly := range m.polys {
		bounds = append(bounds, poly.outer)
		for _, inner := range poly.holes {
			bounds = append(bounds, inner)
		}
	}
	return NewMultiLineString(bounds)
}

func (m MultiPolygon) Value() (driver.Value, error) {
	return wkbAsBytes(m)
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
	return convexHull(m)
}

func (m MultiPolygon) convexHullPointSet() []XY {
	var points []XY
	numPolys := m.NumPolygons()
	for i := 0; i < numPolys; i++ {
		ring := m.PolygonN(i).ExteriorRing()
		numPts := ring.NumPoints()
		for j := 0; j < numPts; j++ {
			points = append(points, ring.PointN(j).XY())
		}
	}
	return points
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
		rings := m.PolygonN(i).rings()
		coords[i] = make([][]Coordinates, len(rings))
		for j, r := range rings {
			n := r.NumPoints()
			coords[i][j] = make([]Coordinates, n)
			for k := 0; k < n; k++ {
				coords[i][j][k] = r.PointN(k).Coordinates()
			}
		}
	}
	return coords
}

// TransformXY transforms this MultiPolygon into another MultiPolygon according to fn.
func (m MultiPolygon) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	coords := m.Coordinates()
	transform3dCoords(coords, fn)
	return NewMultiPolygonC(coords, opts...)
}

// EqualsExact checks if this MultiPolygon is exactly equal to another MultiPolygon.
func (m MultiPolygon) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	o, ok := other.(MultiPolygon)
	return ok && multiPolygonExactEqual(m, o, opts)
}

// IsValid checks if this MultiPolygon is valid
func (m MultiPolygon) IsValid() bool {
	_, err := NewMultiPolygonC(m.Coordinates())
	return err == nil
}

// Area gives the area of the multi polygon.
func (m MultiPolygon) Area() float64 {
	var area float64
	n := m.NumPolygons()
	for i := 0; i < n; i++ {
		area += m.PolygonN(i).Area()
	}
	return area
}

// Centroid returns the multi polygon's centroid point. It returns false if the
// multi polygon is empty (in which case, there is no sensible definition for a
// centroid).
func (m MultiPolygon) Centroid() (Point, bool) {
	if m.IsEmpty() {
		return Point{}, false
	}

	n := m.NumPolygons()
	centroids := make([]XY, n)
	areas := make([]float64, n)
	var totalArea float64
	for i := 0; i < n; i++ {
		centroids[i], areas[i] = centroidAndAreaOfPolygon(m.PolygonN(i))
		totalArea += areas[i]
	}
	var avg XY
	for i := range centroids {
		avg = avg.Add(centroids[i].Scale(areas[i]))
	}
	avg = avg.Scale(1.0 / totalArea)
	return NewPointXY(avg), true
}
