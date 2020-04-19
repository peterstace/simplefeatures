package geom

import (
	"database/sql/driver"
	"errors"
	"math"
	"unsafe"

	"github.com/peterstace/simplefeatures/rtree"
)

// Polygon is a planar surface geometry. Its zero value is the empty Polygon.
// It is immutable after creation. When not empty, it is defined by one outer
// ring and zero or more interior rings. The outer ring defines the exterior
// boundary of the Polygon, and each inner ring defines a hole in the polygon.
//
// For a Polygon to be valid, the following assertions must hold:
//
// 1. The rings (outer and inner) must be valid linear rings. This means that
//    they must be non-empty, simple, and closed.
//
// 2. Each pair of rings must only intersect at a single point.
//
// 3. The interior of the polygon must be connected.
//
// 4. The holes must be fully inside the outer ring.
//
type Polygon struct {
	rings []LineString
	ctype CoordinatesType
}

// NewPolygonFromRings creates a polygon given its rings. The outer ring is
// first, and any inner rings follow. If no rings are provided, then the
// returned Polygon is the empty Polygon. The coordinate type of the polygon is
// the lowest common coordinate type of its rings.
func NewPolygonFromRings(rings []LineString, opts ...ConstructorOption) (Polygon, error) {
	if len(rings) == 0 {
		return Polygon{}, nil
	}

	ctype := DimXYZM
	for _, r := range rings {
		ctype &= r.CoordinatesType()
	}
	rings = append([]LineString(nil), rings...)
	for i := range rings {
		rings[i] = rings[i].ForceCoordinatesType(ctype)
	}

	if err := validatePolygon(rings, opts...); err != nil {
		return Polygon{}, err
	}
	return Polygon{rings, ctype}, nil
}

func validatePolygon(rings []LineString, opts ...ConstructorOption) error {
	if len(rings) == 0 || skipValidations(opts) {
		return nil
	}

	for _, r := range rings {
		if !r.IsClosed() {
			return errors.New("polygon rings must be closed")
		}
		if !r.IsSimple() {
			return errors.New("polygon rings must be simple")
		}
	}

	// Data structures used to track connectedness.
	nextInterVert := len(rings)
	interVerts := make(map[XY]int)
	graph := newGraph()

	// Check each pair of rings (skipping any pairs that could not possibly intersect).
	var tree rtree.RTree
	for i, currentRing := range rings {
		env, ok := currentRing.Envelope()
		if !ok {
			return errors.New("polygon rings must not be empty")
		}
		box := toBox(env)
		if err := tree.Search(box, func(j int) error {
			otherRing := rings[j]
			if i > 0 && j > 0 { // Check is skipped if the outer ring is involved.
				// It's ok to access the first coord (index 0), since we've
				// already checked to ensure that no ring is empty.
				startCurrent := currentRing.Coordinates().GetXY(0)
				startOther := otherRing.Coordinates().GetXY(0)
				nestedFwd := relatePointToRing(startCurrent, otherRing) == interior
				nestedRev := relatePointToRing(startOther, currentRing) == interior
				if nestedFwd || nestedRev {
					return errors.New("polygon must not have nested rings")
				}
			}

			intersects, ext := hasIntersectionLineStringWithLineString(
				currentRing, otherRing, true,
			)
			if !intersects {
				return nil
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
			graph.addEdge(interVert, i)
			graph.addEdge(interVert, j)
			return nil
		}); err != nil {
			return err
		}

		tree.Insert(box, i)
	}

	// All inner rings must be inside the outer ring.
	for _, hole := range rings[1:] {
		holeSeq := hole.Coordinates()
		holeLen := holeSeq.Length()
		for i := 0; i < holeLen; i++ {
			xy := holeSeq.GetXY(i)
			if relatePointToRing(xy, rings[0]) == exterior {
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

// Type returns the GeometryType for a Polygon
func (p Polygon) Type() GeometryType {
	return TypePolygon
}

// AsGeometry converts this Polygon into a Geometry.
func (p Polygon) AsGeometry() Geometry {
	return Geometry{TypePolygon, unsafe.Pointer(&p)}
}

// ExteriorRing gives the exterior ring of the polygon boundary. If the polygon
// is empty, then it returns the empty LineString.
func (p Polygon) ExteriorRing() LineString {
	if p.IsEmpty() {
		return LineString{}.ForceCoordinatesType(p.ctype)
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

// AsText returns the WKT (Well Known Text) representation of this geometry.
func (p Polygon) AsText() string {
	return string(p.AppendWKT(nil))
}

// AppendWKT appends the WKT (Well Known Text) representation of this geometry
// to the input byte slice.
func (p Polygon) AppendWKT(dst []byte) []byte {
	dst = appendWKTHeader(dst, "POLYGON", p.ctype)
	return p.appendWKTBody(dst)
}

func (p Polygon) appendWKTBody(dst []byte) []byte {
	if p.IsEmpty() {
		return appendWKTEmpty(dst)
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

// IsSimple returns true if this geometry contains no anomalous geometry
// points, such as self intersection or self tangency. Because Polygons are
// always simple, this method always returns true.
func (p Polygon) IsSimple() bool {
	return true
}

// Intersects return true if and only if this geometry intersects with the
// other, i.e. they shared at least one common point.
func (p Polygon) Intersects(g Geometry) bool {
	return hasIntersection(p.AsGeometry(), g)
}

// IsEmpty returns true if and only if this Polygon is the empty Polygon. The
// empty Polygon doesn't have any rings and doesn't enclose any area.
func (p Polygon) IsEmpty() bool {
	// Rings are not allowed to be empty, so we don't have to check IsEmpty on
	// each ring.
	return len(p.rings) == 0
}

// Envelope returns the Envelope that most tightly surrounds the geometry. If
// the geometry is empty, then false is returned.
func (p Polygon) Envelope() (Envelope, bool) {
	return p.ExteriorRing().Envelope()
}

// Boundary returns the spatial boundary of this Polygon. For non-empty
// Polygons, this is the MultiLineString collection containing all of the
// rings.
func (p Polygon) Boundary() MultiLineString {
	return NewMultiLineStringFromLineStrings(p.rings).Force2D()
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (p Polygon) Value() (driver.Value, error) {
	return p.AsBinary(), nil
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (p Polygon) AsBinary() []byte {
	return p.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (p Polygon) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaller(dst)
	marsh.writeByteOrder()
	marsh.writeGeomType(TypePolygon, p.ctype)
	marsh.writeCount(len(p.rings))
	for _, ring := range p.rings {
		seq := ring.Coordinates()
		marsh.writeSequence(seq)
	}
	return marsh.buf
}

// ConvexHull returns the geometry representing the smallest convex geometry
// that contains this geometry.
func (p Polygon) ConvexHull() Geometry {
	return convexHull(p.AsGeometry())
}

// MarshalJSON implements the encoding/json.Marshaller interface by encoding
// this geometry as a GeoJSON geometry object.
func (p Polygon) MarshalJSON() ([]byte, error) {
	var dst []byte
	dst = append(dst, `{"type":"Polygon","coordinates":`...)
	dst = appendGeoJSONSequences(dst, p.Coordinates())
	dst = append(dst, '}')
	return dst, nil
}

// Coordinates returns the coordinates of the rings making up the Polygon
// (external ring first, then internal rings after).
func (p Polygon) Coordinates() []Sequence {
	coords := make([]Sequence, len(p.rings))
	for i, r := range p.rings {
		coords[i] = r.Coordinates()
	}
	return coords
}

// TransformXY transforms this Polygon into another Polygon according to fn.
func (p Polygon) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Polygon, error) {
	n := len(p.rings)
	transformed := make([]LineString, n)
	for i, r := range p.rings {
		var err error
		transformed[i], err = NewLineString(
			transformSequence(r.Coordinates(), fn),
			opts...,
		)
		if err != nil {
			return Polygon{}, err
		}
	}
	poly, err := NewPolygonFromRings(transformed, opts...)
	return poly.ForceCoordinatesType(p.ctype), err
}

// EqualsExact checks if this Polygon is exactly equal to another Polygon.
func (p Polygon) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	return other.IsPolygon() &&
		polygonExactEqual(p, other.AsPolygon(), opts)
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
	seq := lr.Coordinates()
	n := seq.Length()
	for i := 0; i < n; i++ {
		pt0 := seq.GetXY(i)
		pt1 := seq.GetXY((i + 1) % n)
		sum += (pt1.X + pt0.X) * (pt1.Y - pt0.Y)
	}
	return sum / 2
}

// Centroid returns the polygon's centroid point. If returns an empty Point if
// the Polygon is empty.
func (p Polygon) Centroid() Point {
	sumXY, sumArea := sumCentroidAndAreaOfPolygon(p)
	if sumArea == 0 {
		return NewEmptyPoint(DimXY)
	}
	return NewPointFromXY(sumXY.Scale(1.0 / sumArea))
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
	seq := lr.Coordinates()
	n := seq.Length()
	var x, y float64
	var area float64
	for i := 0; i < n; i++ {
		pt0 := seq.GetXY(i)
		pt1 := seq.GetXY((i + 1) % n)
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
	mp, err := NewMultiPolygonFromPolygons(polys)
	if err != nil {
		// Cannot occur due to construction. A valid polygon will always be a
		// valid multipolygon.
		panic(err)
	}
	return mp.ForceCoordinatesType(p.ctype)
}

// Reverse in the case of Polygon outputs the coordinates of each ring in reverse order,
// but note the order of the inner rings is unchanged.
func (p Polygon) Reverse() Polygon {
	reversed := make([]LineString, len(p.rings))
	for i := range reversed {
		reversed[i] = p.rings[i].Reverse()
	}
	return Polygon{reversed, p.ctype}
}

// CoordinatesType returns the CoordinatesType used to represent points making
// up the geometry.
func (p Polygon) CoordinatesType() CoordinatesType {
	return p.ctype
}

// ForceCoordinatesType returns a new Polygon with a different CoordinatesType. If a dimension
// is added, then new values are populated with 0.
func (p Polygon) ForceCoordinatesType(newCType CoordinatesType) Polygon {
	flatRings := make([]LineString, len(p.rings))
	for i := range p.rings {
		flatRings[i] = p.rings[i].ForceCoordinatesType(newCType)
	}
	return Polygon{flatRings, newCType}
}

// Force2D returns a copy of the Polygon with Z and M values removed.
func (p Polygon) Force2D() Polygon {
	return p.ForceCoordinatesType(DimXY)
}
