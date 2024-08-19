package geom

import (
	"database/sql/driver"
	"fmt"
	"math"
	"unsafe"

	"github.com/peterstace/simplefeatures/rtree"
)

// Polygon is a planar surface geometry. Its zero value is the empty Polygon.
// It is immutable after creation. When not empty, it is defined by one outer
// ring and zero or more interior rings. The outer ring defines the exterior
// boundary of the Polygon, and each inner ring defines a hole in the polygon.
type Polygon struct {
	rings []LineString
	ctype CoordinatesType
}

// NewPolygon creates a polygon given its rings. The outer ring is first, and
// any inner rings follow. If no rings are provided, then the returned Polygon
// is the empty Polygon. The coordinate type of the polygon is the lowest
// common coordinate type of its rings.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewPolygon(rings []LineString) Polygon {
	ctype := DimXY
	if len(rings) > 0 {
		ctype = DimXYZM
		for _, r := range rings {
			ctype &= r.CoordinatesType()
		}
	}
	rings = append([]LineString(nil), rings...)
	for i := range rings {
		rings[i] = rings[i].ForceCoordinatesType(ctype)
	}
	return Polygon{rings, ctype}
}

// Validate checks if the Polygon is valid. For non-empty Polygons to be valid,
// the following validation rules must hold:
//
//  1. The rings (outer and inner) must be valid linear rings. This means that
//     they must be valid LineStrings that are non-empty, simple, and closed.
//  2. Each pair of rings must at most intersect at a single point.
//  3. The interior of the polygon must be connected.
//  4. Any interior rings must be fully inside the exterior ring.
func (p Polygon) Validate() error {
	if len(p.rings) == 0 {
		return nil
	}

	for i, r := range p.rings {
		if err := validateRing(r); err != nil {
			return wrap(err, "validating ring at index %d", i)
		}
	}

	// Data structures used to track connectedness.
	nextInterVert := len(p.rings)
	interVerts := make(map[XY]int)
	graph := newGraph()

	// Construct RTree of rings.
	boxes := make([]rtree.Box, len(p.rings))
	items := make([]rtree.BulkItem, len(p.rings))
	for i, r := range p.rings {
		box, ok := r.Envelope().AsBox()
		if !ok {
			// Cannot occur, because we have already checked to ensure rings
			// are closed. Closed rings by definition are non-empty.
			panic("unexpected empty ring")
		}
		boxes[i] = box
		items[i] = rtree.BulkItem{Box: boxes[i], RecordID: i}
	}
	tree := rtree.BulkLoad(items)

	// Check each pair of rings (skipping any pairs that could not possibly intersect).
	for i := range p.rings {
		if err := tree.RangeSearch(boxes[i], func(j int) error {
			// Only compare each pair once.
			if i <= j {
				return nil
			}
			if i > 0 && j > 0 { // Check is skipped if the outer ring is involved.
				// It's ok to access the first coord (index 0), since we've
				// already checked to ensure that no ring is empty.
				iStart := p.rings[i].Coordinates().GetXY(0)
				jStart := p.rings[j].Coordinates().GetXY(0)
				nestedFwd := relatePointToRing(iStart, p.rings[j]) == interior
				nestedRev := relatePointToRing(jStart, p.rings[i]) == interior
				if nestedFwd || nestedRev {
					return violateRingNested.errAtXY(iStart)
				}
			}

			intersects, ext := hasIntersectionLineStringWithLineString(p.rings[i], p.rings[j], true)
			if !intersects {
				return nil
			}
			if ext.multiplePoints {
				return violateRingsMultiTouch.errAtXY(ext.singlePoint)
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
	}

	// All inner rings must be inside the outer ring. Because we have already
	// made sure that the rings don't intersect at multiple points, we can
	// check arbitrary points in the inner rings until we find one that's
	// unambiguously inside or outside the outer ring.
	for _, hole := range p.rings[1:] {
		holeSeq := hole.Coordinates()
		for i := 0; i < holeSeq.Length(); i++ {
			xy := holeSeq.GetXY(i)
			relate := relatePointToRing(xy, p.rings[0])
			if relate == exterior {
				return violateInteriorInExterior.errAtXY(xy)
			}
			if relate == interior {
				break
			}
			// Continue to next iteration if relate == boundary.
		}
	}

	// Connectedness check: a graph is created where the intersections and
	// rings are modelled as vertices. Edges are added to the graph between an
	// intersection vertex and a ring vertex if the ring participates in that
	// intersection. The interior of the polygon is connected iff the graph
	// does not contain a cycle.
	if graph.hasCycle() {
		return violateInteriorConnected.err()
	}
	return nil
}

func validateRing(r LineString) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if r.IsEmpty() {
		return violateRingEmpty.err()
	}
	if !r.IsClosed() {
		return violateRingClosed.errAtXY(r.Coordinates().GetXY(0))
	}
	if !r.IsSimple() {
		return violateRingSimple.errAtXY(r.Coordinates().GetXY(0))
	}
	return nil
}

// Type returns the GeometryType for a Polygon.
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
	return maxInt(0, len(p.rings)-1)
}

// NumRings gives the total number of rings: ExternalRing + NumInteriorRings().
func (p Polygon) NumRings() int {
	if p.IsEmpty() {
		return 0
	}
	return 1 + p.NumInteriorRings()
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

// IsEmpty returns true if and only if this Polygon is the empty Polygon. The
// empty Polygon doesn't have any rings and doesn't enclose any area.
func (p Polygon) IsEmpty() bool {
	// Rings are not allowed to be empty, so we don't have to check IsEmpty on
	// each ring.
	return len(p.rings) == 0
}

// Envelope returns the Envelope that most tightly surrounds the geometry.
func (p Polygon) Envelope() Envelope {
	return p.ExteriorRing().Envelope()
}

// Boundary returns the spatial boundary of this Polygon. For non-empty
// Polygons, this is the MultiLineString collection containing all of the
// rings.
func (p Polygon) Boundary() MultiLineString {
	return NewMultiLineString(p.rings).Force2D()
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (p Polygon) Value() (driver.Value, error) {
	return p.AsBinary(), nil
}

// Scan implements the database/sql.Scanner interface by parsing the src value
// as WKB (Well Known Binary).
//
// If the WKB doesn't represent a Polygon geometry, then an error is returned.
//
// Geometry constraint validation is performed on the resultant geometry (an
// error will be returned if the geometry is invalid). If this validation isn't
// needed or is undesirable, then the WKB should be scanned into a byte slice
// and then UnmarshalWKB called manually (passing in NoValidate{}).
func (p *Polygon) Scan(src interface{}) error {
	return scanAsType(src, p)
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (p Polygon) AsBinary() []byte {
	return p.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (p Polygon) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaler(dst)
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

// MarshalJSON implements the encoding/json.Marshaler interface by encoding
// this geometry as a GeoJSON geometry object.
func (p Polygon) MarshalJSON() ([]byte, error) {
	var dst []byte
	dst = append(dst, `{"type":"Polygon","coordinates":`...)
	dst = appendGeoJSONSequences(dst, p.Coordinates())
	dst = append(dst, '}')
	return dst, nil
}

// UnmarshalJSON implements the encoding/json.Unmarshaler interface by decoding
// the GeoJSON representation of a Polygon.
func (p *Polygon) UnmarshalJSON(buf []byte) error {
	return unmarshalGeoJSONAsType(buf, p)
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
func (p Polygon) TransformXY(fn func(XY) XY) Polygon {
	n := len(p.rings)
	transformed := make([]LineString, n)
	for i, r := range p.rings {
		seq := transformSequence(r.Coordinates(), fn)
		transformed[i] = NewLineString(seq)
	}
	poly := NewPolygon(transformed)
	return poly.ForceCoordinatesType(p.ctype)
}

// AreaOption allows the behaviour of area calculations to be modified.
type AreaOption func(o *areaOptionSet)

type areaOptionSet struct {
	signed    bool
	transform func(XY) XY
}

func newAreaOptionSet(opts []AreaOption) areaOptionSet {
	var os areaOptionSet
	for _, opt := range opts {
		opt(&os)
	}
	return os
}

// WithTransform alters the behaviour of area calculations by first
// transforming the geometry with the provided transform function.
func WithTransform(tr func(XY) XY) AreaOption {
	return func(o *areaOptionSet) {
		o.transform = tr
	}
}

// SignedArea alters the behaviour of area calculations. It causes them to give
// a positive areas when the outer rings are wound CCW and any inner rings are
// wound CW, and a negative area when the outer rings are wound CW and any
// inner rings are wound CCW.  If the windings of the inner and outer rings are
// the same, then the area will be inconsistent.
func SignedArea(o *areaOptionSet) {
	o.signed = true
}

// Area of a Polygon is the area enclosed by the polygon's boundary.
func (p Polygon) Area(opts ...AreaOption) float64 {
	os := newAreaOptionSet(opts)
	totalArea := signedAreaOfLinearRing(p.ExteriorRing(), os.transform)
	if !os.signed {
		totalArea = math.Abs(totalArea)
	}
	n := p.NumInteriorRings()
	for i := 0; i < n; i++ {
		area := signedAreaOfLinearRing(p.InteriorRingN(i), os.transform)
		if os.signed {
			totalArea += area
		} else {
			totalArea -= math.Abs(area)
		}
	}
	return totalArea
}

func signedAreaOfLinearRing(lr LineString, transform func(XY) XY) float64 {
	// This is the "Shoelace Formula".
	var sum float64
	seq := lr.Coordinates()
	n := seq.Length()
	if n == 0 {
		return 0
	}

	nthPt := func(i int) XY {
		pt := seq.GetXY(i)
		if transform != nil {
			pt = transform(pt)
		}
		return pt
	}

	pt1 := nthPt(0)
	for i := 0; i < n-1; i++ {
		pt0 := pt1
		pt1 = nthPt(i + 1)
		sum += (pt1.X + pt0.X) * (pt1.Y - pt0.Y)
	}
	return sum / 2
}

// Centroid returns the polygon's centroid point. If returns an empty Point if
// the Polygon is empty.
func (p Polygon) Centroid() Point {
	if p.IsEmpty() {
		return NewEmptyPoint(DimXY)
	}

	// The basis of this approach is taken from:
	// https://stackoverflow.com/questions/2792443/finding-the-centroid-of-a-polygon
	// The original sources that the SO answer links to are gone (servers no
	// longer up), so it's hard to trace it through to the original sources.
	// GEOS and JTS seem to use a very similar calculation method.

	areas := make([]float64, 1+p.NumInteriorRings())
	areas[0] = math.Abs(signedAreaOfLinearRing(p.ExteriorRing(), nil))
	sumAreas := areas[0]
	for i := 0; i < p.NumInteriorRings(); i++ {
		areas[i+1] = -math.Abs(signedAreaOfLinearRing(p.InteriorRingN(i), nil))
		sumAreas += areas[i+1]
	}

	centroid := weightedCentroid(p.ExteriorRing(), areas[0], sumAreas)
	for i := 0; i < p.NumInteriorRings(); i++ {
		centroid = centroid.Add(
			weightedCentroid(p.InteriorRingN(i), areas[i+1], sumAreas))
	}
	return centroid.AsPoint()
}

func weightedCentroid(ring LineString, ringArea, totalArea float64) XY {
	centroid := centroidOfRing(ring)
	return centroid.Scale(ringArea / totalArea)
}

func centroidOfRing(ring LineString) XY {
	var areaSum2 float64 // double the area
	var cent6 XY         // sextuple the centroid (also scaled by area)

	seq := ring.Coordinates()
	n := seq.Length()

	base := seq.GetXY(0)
	for i := 1; i+1 < n; i++ {
		cent3 := centroid3(base, seq.GetXY(i), seq.GetXY(i+1))
		area2 := triangleArea2(base, seq.GetXY(i), seq.GetXY(i+1))
		cent6 = cent6.Add(cent3.Scale(area2))
		areaSum2 += area2
	}
	return cent6.Scale(1.0 / 3.0 / areaSum2)
}

// centroid3 returns triple the centroid of 3 points.
func centroid3(pt1, pt2, pt3 XY) XY {
	return pt1.Add(pt2).Add(pt3)
}

// triangleArea2 returns double the signed area of the triangle defined by 3 points.
func triangleArea2(pt1, pt2, pt3 XY) float64 {
	return (pt2.X-pt1.X)*(pt3.Y-pt1.Y) - (pt3.X-pt1.X)*(pt2.Y-pt1.Y)
}

// AsMultiPolygon is a helper that converts this Polygon into a MultiPolygon.
func (p Polygon) AsMultiPolygon() MultiPolygon {
	var polys []Polygon
	if !p.IsEmpty() {
		polys = []Polygon{p}
	}
	mp := NewMultiPolygon(polys)
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

// PointOnSurface returns a Point that lies inside the Polygon.
func (p Polygon) PointOnSurface() Point {
	pt, _ := pointOnAreaSurface(p)
	return pt
}

// ForceCW returns the equivalent Polygon that has its exterior ring in a
// clockwise orientation and any inner rings in a counter-clockwise
// orientation.
func (p Polygon) ForceCW() Polygon {
	if p.IsCW() {
		return p
	}
	return p.forceOrientation(true)
}

// ForceCCW returns the equivalent Polygon that has its exterior ring in a
// counter-clockwise orientation and any inner rings in a clockwise
// orientation.
func (p Polygon) ForceCCW() Polygon {
	if p.IsCCW() {
		return p
	}
	return p.forceOrientation(false)
}

func (p Polygon) forceOrientation(forceCW bool) Polygon {
	orientedRings := make([]LineString, len(p.rings))
	for i, ring := range p.rings {
		alreadyCW := signedAreaOfLinearRing(ring, nil) < 0
		if (i == 0) == (alreadyCW == forceCW) {
			orientedRings[i] = ring
		} else {
			orientedRings[i] = ring.Reverse()
		}
	}
	return Polygon{orientedRings, p.ctype}
}

// IsCW returns true iff the outer ring is CW and all inner rings are CCW.
// Any linear ring with a negative signed area is assumed to be CW.
// Any linear ring with a positive signed area is assumed to be CCW.
// Any linear ring of zero area is assumed to be neither CW nor CCW.
// An empty polygon returns true.
func (p Polygon) IsCW() bool {
	for i, ring := range p.rings {
		isCW := signedAreaOfLinearRing(ring, nil) < 0
		if (i == 0) != isCW {
			return false
		}
	}
	return true
}

// IsCCW returns true iff the outer ring is CCW and all inner rings are CW.
// Any linear ring with a negative signed area is assumed to be CW.
// Any linear ring with a positive signed area is assumed to be CCW.
// Any linear ring of zero area is assumed to be neither CW nor CCW.
// An empty polygon returns true.
func (p Polygon) IsCCW() bool {
	for i, ring := range p.rings {
		isCCW := signedAreaOfLinearRing(ring, nil) > 0
		if (i == 0) != isCCW {
			return false
		}
	}
	return true
}

func (p Polygon) controlPoints() int {
	var sum int
	for _, r := range p.rings {
		sum += r.Coordinates().Length()
	}
	return sum
}

// DumpCoordinates returns the points making up the rings in a Polygon as a
// Sequence.
func (p Polygon) DumpCoordinates() Sequence {
	var n int
	for _, r := range p.rings {
		n += r.Coordinates().Length()
	}
	ctype := p.CoordinatesType()
	coords := make([]float64, 0, n*ctype.Dimension())
	for _, r := range p.rings {
		coords = r.Coordinates().appendAllPoints(coords)
	}
	seq := NewSequence(coords, ctype)
	seq.assertNoUnusedCapacity()
	return seq
}

// DumpRings returns a copy of the Polygon's rings as a slice of LineStrings.
// If the Polygon is empty, then the slice will have length zero. Otherwise,
// the slice will consist of the exterior ring, followed by any interior rings.
func (p Polygon) DumpRings() []LineString {
	tmp := make([]LineString, len(p.rings))
	copy(tmp, p.rings)
	return tmp
}

// Summary returns a text summary of the Polygon following a similar format to https://postgis.net/docs/ST_Summary.html.
func (p Polygon) Summary() string {
	numPoints := p.DumpCoordinates().Length()

	var ringSuffix string
	numRings := p.NumRings()
	if numRings != 1 {
		ringSuffix = "s"
	}
	return fmt.Sprintf("%s[%s] with %d ring%s consisting of %d total points",
		p.Type(), p.CoordinatesType(), numRings, ringSuffix, numPoints)
}

// String returns the string representation of the Polygon.
func (p Polygon) String() string {
	return p.Summary()
}

// Simplify returns a simplified version of the Polygon by applying the
// Ramer-Douglas-Peucker algorithm to each constituent ring. If the exterior
// ring collapses to a point or single linear element, the empty Polygon is
// returned. If any interior ring collapses to a point or a single linear
// element, then it is omitted from the final output.
//
// The output Polygon will be invalid if any rings in the input become
// non-rings (e.g. via self intersection) in the output, or if any two rings
// were to interact in ways prohibited by Polygon validation rules (such as
// intersecting at more than one point). In these cases, an error is returned.
// These validations can be skipped by passing in NoValidate{}, potentially
// resulting in a non-valid Polygon being returned.
func (p Polygon) Simplify(threshold float64, nv ...NoValidate) (Polygon, error) {
	exterior := p.ExteriorRing().Simplify(threshold)

	// If we don't have at least 4 coordinates, then we can't form a ring, and
	// the polygon has collapsed either to a point or a single linear element.
	// Both cases are represented by an empty Polygon.
	hasCollapsed := func(ring LineString) bool {
		return ring.Coordinates().Length() < 4
	}
	if hasCollapsed(exterior) {
		return Polygon{}.ForceCoordinatesType(p.CoordinatesType()), nil
	}

	n := p.NumInteriorRings()
	rings := make([]LineString, 0, n+1)
	rings = append(rings, exterior)
	for i := 0; i < n; i++ {
		interior := p.InteriorRingN(i).Simplify(threshold)
		if !hasCollapsed(interior) {
			rings = append(rings, interior)
		}
	}
	simpl := NewPolygon(rings)
	if len(nv) == 0 {
		if err := simpl.Validate(); err != nil {
			return Polygon{}, wrapSimplified(err)
		}
	}
	return simpl, nil
}

// Densify returns a new Polygon with additional linearly interpolated control
// points such that the distance between any two consecutive control points is
// at most the given maxDistance.
//
// Panics if maxDistance is zero or negative.
func (p Polygon) Densify(maxDistance float64) Polygon {
	rings := make([]LineString, len(p.rings))
	for i, r := range p.rings {
		rings[i] = r.Densify(maxDistance)
	}
	return Polygon{rings, p.ctype}
}

// SnapToGrid returns a copy of the Polygon with all coordinates snapped to a
// base 10 grid.
//
// The grid spacing is specified by the number of decimal places to round to
// (with negative decimal places being allowed). E.g., a decimalPlaces value of
// 2 would cause all coordinates to be rounded to the nearest 0.01, and a
// decimalPlaces of -1 would cause all coordinates to be rounded to the nearest
// 10.
//
// Returned Polygons may be invalid due to snapping, even if the input geometry
// was valid.
func (p Polygon) SnapToGrid(decimalPlaces int) Polygon {
	return p.TransformXY(snapToGridXY(decimalPlaces))
}
