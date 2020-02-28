package geom

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"io"
	"math"
	"sort"
	"unsafe"
)

// Polygon is a planar surface. Its zero value is the empty Polygon. When not
// empty, it is defined by one outer ring and zero or more interior rings. The
// outer ring defines the exterior boundary of the Polygon, and each inner ring
// defines a hole in the polygon.
//
// Its assertions are:
//
// 1. The rings (outer and inner) must be valid linear rings (i.e. be simple
// and closed LineStrings).
//
// 2. Each pair of rings must only intersect at a single point.
//
// 3. The interior of the polygon is connected.
//
// 4. The holes must be fully inside the outer ring.
//
type Polygon struct {
	rings []LineString
}

// NewEmptyPolygon returns an empty Polygon. It is equivalent to calling
// NewPolygon with a zero length rings argument.
func NewEmptyPolygon() Polygon {
	return Polygon{}
}

// NewPolygon creates a polygon given its rings. The outer ring is first, and
// any inner rings follow. No rings may cross each other, and can only
// intersect each with each other at a point. If no rings are provided, then
// the returned Polygon is the empty Polygon.
func NewPolygon(rings []LineString, opts ...ConstructorOption) (Polygon, error) {
	if err := validatePolygon(rings, opts...); err != nil {
		return Polygon{}, err
	}
	tmp := make([]LineString, len(rings))
	copy(tmp, rings)
	return Polygon{rings}, nil
}

// NewPolygonC creates a new polygon from its Coordinates. The first dimension
// of the Coordinates slice is the ring number, and the second dimension of the
// Coordinates slice is the position within the ring.
func NewPolygonC(coords [][]Coordinates, opts ...ConstructorOption) (Polygon, error) {
	rings := make([]LineString, len(coords))
	for i := range rings {
		var err error
		rings[i], err = NewLineStringC(coords[i], opts...)
		if err != nil {
			return Polygon{}, err
		}
	}
	if err := validatePolygon(rings, opts...); err != nil {
		return Polygon{}, err
	}
	return Polygon{rings}, nil
}

// NewPolygonXY creates a new polygon from its XYs. The first dimension of the
// XYs slice is the ring number, and the second dimension of the XYs slice is
// the position within the ring.
func NewPolygonXY(pts [][]XY, opts ...ConstructorOption) (Polygon, error) {
	rings := make([]LineString, len(pts))
	for i := range rings {
		var err error
		rings[i], err = NewLineStringXY(pts[i], opts...)
		if err != nil {
			return Polygon{}, err
		}
	}
	if err := validatePolygon(rings, opts...); err != nil {
		return Polygon{}, err
	}
	return Polygon{rings}, nil
}

func validatePolygon(rings []LineString, opts ...ConstructorOption) error {
	if len(rings) == 0 || skipValidations(opts) {
		return nil
	}

	// Overview:
	//
	// 1. Create slice of all rings, ordered by min X coordinate.
	// 2. Loop over each ring.
	//    a. Remove any ring from the heap that has max X coordinate less than
	//       the min X of the current ring.
	//    b. Check to see if the current ring intersects with any in the heap.
	//    c. Insert the current ring into the heap.

	orderedRings := seq(len(rings))
	type interval struct {
		minX, maxX float64
	}
	intervals := make([]interval, len(orderedRings))
	for i := range intervals {
		env, ok := rings[i].Envelope()
		if !ok {
			return errors.New("polygon rings must not be empty")
		}
		intervals[i].minX = env.Min().X
		intervals[i].maxX = env.Max().X
	}

	for _, r := range rings {
		if !r.IsClosed() {
			return errors.New("polygon rings must be closed")
		}
		if !r.IsSimple() {
			return errors.New("polygon rings must be simple")
		}
	}

	sort.Slice(orderedRings, func(i, j int) bool {
		return intervals[i].minX < intervals[j].minX
	})

	nextInterVert := len(orderedRings)
	interVerts := make(map[XY]int)
	graph := newGraph()

	h := intHeap{less: func(i, j int) bool {
		return intervals[i].maxX < intervals[j].maxX
	}}

	for _, current := range orderedRings {
		currentRing := rings[current]
		currentX := intervals[current].minX
		for len(h.data) > 0 && intervals[h.data[0]].maxX < currentX {
			h.pop()
		}
		for _, other := range h.data {
			otherRing := rings[other]
			if current > 0 && other > 0 {
				// Check is skipped if the outer ring is involved.
				startXY, ok := currentRing.StartPoint().XY()
				if !ok {
					panic("already checked that all rings are non-empty")
				}
				nestedFwd := pointRingSide(startXY, otherRing) == interior
				nestedRev := pointRingSide(startXY, currentRing) == interior
				if nestedFwd || nestedRev {
					return errors.New("polygon must not have nested rings")
				}
			}

			intersects, ext := hasIntersectionMultiLineStringWithMultiLineString(
				currentRing.AsMultiLineString(),
				otherRing.AsMultiLineString(),
				true,
			)
			if !intersects {
				continue
			}
			if ext.multiplePoints {
				return errors.New("polygon rings must not intersect at multiple points")
			}

			interVert, ok := interVerts[ext.singlePoint]
			if !ok {
				interVert = nextInterVert
				nextInterVert++
				interVerts[ext.singlePoint] = interVert
			}
			graph.addEdge(interVert, current)
			graph.addEdge(interVert, other)
		}
		h.push(current)
	}

	// All inner rings must be inside the outer ring.
	for _, hole := range rings[1:] {
		for i := 0; i < hole.NumPoints(); i++ {
			pt := hole.PointN(i)
			if pointRingSide(pt.XY, rings[0]) == exterior {
				return errors.New("hole must be inside outer ring")
			}
		}
	}

	// Connectedness check: a graph is created where the intersections and
	// rings are modelled as vertices. Edges are added to the graph between an
	// intersection vertex and a ring vertex if the ring participates in that
	// intersection. The interior of the polygon is connected iff the graph
	// does not contain a cycle.
	if graph.hasCycle() {
		return errors.New("polygon interiors must be connected")
	}
	return nil
}

// AsGeometry converts this Polygon into a Geometry.
func (p Polygon) AsGeometry() Geometry {
	return Geometry{polygonTag, unsafe.Pointer(&p)}
}

// ExteriorRing gives the exterior ring of the polygon boundary. If the polygon
// is empty, then it returns the empty LineString.
func (p Polygon) ExteriorRing() LineString {
	if p.IsEmpty() {
		return NewEmptyLineString()
	}
	return p.rings[0]
}

// NumInteriorRings gives the number of interior rings in the polygon boundary.
func (p Polygon) NumInteriorRings() int {
	return max(0, len(p.rings)-1)
}

// InteriorRingN gives the nth (zero indexed) interior ring in the polygon
// boundary. It will panic if n is out of bounds with respect to the number of
// interior rings.
func (p Polygon) InteriorRingN(n int) LineString {
	// Outer ring is at the 0th position.
	if n == -1 {
		panic("n out of range")
	}
	return p.rings[n+1]
}

func (p Polygon) AsText() string {
	return string(p.AppendWKT(nil))
}

func (p Polygon) AppendWKT(dst []byte) []byte {
	dst = append(dst, "POLYGON"...)
	if p.IsEmpty() {
		dst = append(dst, ' ')
	}
	return p.appendWKTBody(dst)
}

func (p Polygon) appendWKTBody(dst []byte) []byte {
	if p.IsEmpty() {
		return append(dst, "EMPTY"...)
	}

	dst = append(dst, '(')
	for i, r := range p.rings {
		dst = r.appendWKTBody(dst)
		if i+1 < len(p.rings) {
			dst = append(dst, ',')
		}
	}
	return append(dst, ')')
}

// IsSimple always returns true. All Polygons are simple.
func (p Polygon) IsSimple() bool {
	return true
}

func (p Polygon) Intersection(g Geometry) (Geometry, error) {
	return intersection(p.AsGeometry(), g)
}

func (p Polygon) Intersects(g Geometry) bool {
	return hasIntersection(p.AsGeometry(), g)
}

func (p Polygon) IsEmpty() bool {
	return len(p.rings) == 0
}

func (p Polygon) Equals(other Geometry) (bool, error) {
	return equals(p.AsGeometry(), other)
}

func (p Polygon) Envelope() (Envelope, bool) {
	return p.ExteriorRing().Envelope()
}

func (p Polygon) Boundary() MultiLineString {
	return NewMultiLineString(p.rings)
}

func (p Polygon) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := p.AsBinary(&buf)
	return buf.Bytes(), err
}

func (p Polygon) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypePolygon)
	marsh.writeCount(len(p.rings))
	for _, ring := range p.rings {
		numPts := ring.NumPoints()
		marsh.writeCount(numPts)
		for i := 0; i < numPts; i++ {
			pt := ring.PointN(i)
			marsh.writeFloat64(pt.X)
			marsh.writeFloat64(pt.Y)
		}
	}
	return marsh.err
}

// ConvexHull returns the convex hull of the Polygon, which is always another
// Polygon.
func (p Polygon) ConvexHull() Geometry {
	return convexHull(p.AsGeometry())
}

func (p Polygon) MarshalJSON() ([]byte, error) {
	return marshalGeoJSON("Polygon", p.Coordinates())
}

// Coordinates returns the coordinates of the rings making up the Polygon
// (external ring first, then internal rings after).
func (p Polygon) Coordinates() [][]Coordinates {
	coords := make([][]Coordinates, len(p.rings))
	for i, r := range p.rings {
		coords[i] = r.Coordinates()
	}
	return coords
}

// TransformXY transforms this Polygon into another Polygon according to fn.
func (p Polygon) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	coords := p.Coordinates()
	transform2dCoords(coords, fn)
	poly, err := NewPolygonC(coords, opts...)
	return poly.AsGeometry(), err
}

// EqualsExact checks if this Polygon is exactly equal to another Polygon.
func (p Polygon) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	return other.IsPolygon() &&
		polygonExactEqual(p, other.AsPolygon(), opts)
}

// IsValid checks if this Polygon is valid
func (p Polygon) IsValid() bool {
	_, err := NewPolygonC(p.Coordinates())
	return err == nil
}

// Area of a Polygon is the outer ring's area minus the areas of all inner rings.
func (p Polygon) Area() float64 {
	area := math.Abs(signedAreaOfLinearRing(p.ExteriorRing()))
	n := p.NumInteriorRings()
	for i := 0; i < n; i++ {
		area -= math.Abs(signedAreaOfLinearRing(p.InteriorRingN(i)))
	}
	return area
}

// SignedArea gives the positive area of the polygon when the outer rings are
// wound CCW and any inner rings are wound CW, and the negative area of the
// polygon when the outer rings are wound CW and any inner rings are wound CCW.
// If the windings of the inner and outer rings are the same, then the area
// will be inconsistent.
func (p Polygon) SignedArea() float64 {
	signedArea := signedAreaOfLinearRing(p.ExteriorRing())
	n := p.NumInteriorRings()
	for i := 0; i < n; i++ {
		signedArea += signedAreaOfLinearRing(p.InteriorRingN(i))
	}
	return signedArea
}

func signedAreaOfLinearRing(lr LineString) float64 {
	// This is the "Shoelace Formula".
	var sum float64
	n := lr.NumPoints()
	for i := 0; i < n; i++ {
		pt0 := lr.PointN(i)
		pt1 := lr.PointN((i + 1) % n)
		sum += (pt1.X + pt0.X) * (pt1.Y - pt0.Y)
	}
	return sum / 2
}

// Centroid returns the polygon's centroid point. If returns an empty Point if
// the Polygon is empty.
func (p Polygon) Centroid() Point {
	sumXY, sumArea := sumCentroidAndAreaOfPolygon(p)
	if sumArea == 0 {
		return NewEmptyPoint()
	}
	return NewPointXY(sumXY.Scale(1.0 / sumArea))
}

func sumCentroidAndAreaOfPolygon(p Polygon) (sumXY XY, sumArea float64) {
	n := p.NumInteriorRings()
	xy, area := sumCentroidAndAreaOfLinearRing(p.ExteriorRing())
	if area > 0 {
		sumXY = sumXY.Add(xy)
		sumArea += area // Define exterior ring as having positive area.
	} else {
		sumXY = sumXY.Sub(xy)
		sumArea -= area // Define exterior ring as having positive area.
	}
	for i := 0; i < n; i++ {
		r := p.InteriorRingN(i)
		xy, area = sumCentroidAndAreaOfLinearRing(r)
		if area > 0 {
			sumXY = sumXY.Sub(xy)
			sumArea -= area // Holes have negative area.
		} else {
			sumXY = sumXY.Add(xy)
			sumArea += area // Holes have negative area.
		}
	}
	return sumXY, sumArea
}

func centroidAndAreaOfLinearRing(lr LineString) (XY, float64) {
	xy, area := sumCentroidAndAreaOfLinearRing(lr)
	if area == 0 {
		return XY{}, 0
	}
	return XY{xy.X / area, xy.Y / area}, math.Abs(area)
}

func sumCentroidAndAreaOfLinearRing(lr LineString) (XY, float64) {
	n := lr.NumPoints()
	var x, y float64
	var area float64
	for i := 0; i < n; i++ {
		pt0 := lr.PointN(i)
		pt1 := lr.PointN((i + 1) % n)
		x += (pt0.X + pt1.X) * (pt0.X*pt1.Y - pt1.X*pt0.Y)
		y += (pt0.Y + pt1.Y) * (pt0.X*pt1.Y - pt1.X*pt0.Y)
		area += pt0.X*pt1.Y - pt1.X*pt0.Y
	}
	area /= 2
	return XY{x / 6, y / 6}, area
}

// AsMultiPolygon is a helper that converts this Polygon into a MultiPolygon.
func (p Polygon) AsMultiPolygon() MultiPolygon {
	var polys []Polygon
	if !p.IsEmpty() {
		polys = []Polygon{p}
	}
	mp, err := NewMultiPolygon(polys)
	if err != nil {
		// Cannot occur due to construction. A valid polygon will always be a
		// valid multipolygon.
		panic(err)
	}
	return mp
}

// Reverse in the case of Polygon outputs the coordinates of each ring in reverse order,
// but note the order of the inner rings is unchanged.
func (p Polygon) Reverse() Polygon {
	reversed := make([]LineString, len(p.rings))
	for i := range reversed {
		reversed[i] = p.rings[i].Reverse()
	}
	p2, err := NewPolygon(reversed)
	if err != nil {
		panic("Reverse of an existing Polygon should not fail")
	}
	return p2
}
