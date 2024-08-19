package geom

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"unsafe"
)

// GeometryCollection is a non-homogeneous collection of geometries. Its zero
// value is the empty GeometryCollection (i.e. a collection of zero
// geometries).
type GeometryCollection struct {
	// Invariant: ctype matches the coordinates type of each geometry.
	geoms []Geometry
	ctype CoordinatesType
}

// NewGeometryCollection creates a collection of geometries. The coordinates
// type of the GeometryCollection is the lowest common coordinates type of its
// child geometries.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewGeometryCollection(geoms []Geometry) GeometryCollection {
	if len(geoms) == 0 {
		return GeometryCollection{}
	}

	ctype := DimXYZM
	for _, g := range geoms {
		ctype &= g.CoordinatesType()
	}
	geoms = append([]Geometry(nil), geoms...)
	for i := range geoms {
		geoms[i] = geoms[i].ForceCoordinatesType(ctype)
	}
	return GeometryCollection{geoms, ctype}
}

// Validate checks if the GeometryCollection is valid. The only validation rule
// is that each geometry in the collection must be valid.
func (c GeometryCollection) Validate() error {
	for i, g := range c.geoms {
		if err := g.Validate(); err != nil {
			return wrap(err, "validating geometry at index %d", i)
		}
	}
	return nil
}

// Type returns the GeometryType for a GeometryCollection.
func (c GeometryCollection) Type() GeometryType {
	return TypeGeometryCollection
}

// AsGeometry converts this GeometryCollection into a Geometry.
func (c GeometryCollection) AsGeometry() Geometry {
	return Geometry{TypeGeometryCollection, unsafe.Pointer(&c)}
}

// NumGeometries gives the number of Geometry elements in the GeometryCollection.
func (c GeometryCollection) NumGeometries() int {
	return len(c.geoms)
}

// NumTotalGeometries gives the total number of Geometry elements in the GeometryCollection.
// If there are GeometryCollection-type child geometries, this will recursively count its children.
func (c GeometryCollection) NumTotalGeometries() int {
	var n int
	for _, geom := range c.geoms {
		if geom.IsGeometryCollection() {
			n += geom.MustAsGeometryCollection().NumTotalGeometries()
		}
	}
	return n + c.NumGeometries()
}

// GeometryN gives the nth (zero based) Geometry in the GeometryCollection.
func (c GeometryCollection) GeometryN(n int) Geometry {
	return c.geoms[n]
}

// AsText returns the WKT (Well Known Text) representation of this geometry.
func (c GeometryCollection) AsText() string {
	return string(c.AppendWKT(nil))
}

// AppendWKT appends the WKT (Well Known Text) representation of this geometry
// to the input byte slice.
func (c GeometryCollection) AppendWKT(dst []byte) []byte {
	dst = appendWKTHeader(dst, "GEOMETRYCOLLECTION", c.ctype)
	if len(c.geoms) == 0 {
		return appendWKTEmpty(dst)
	}
	dst = append(dst, '(')
	for i, g := range c.geoms {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = g.AppendWKT(dst)
	}
	return append(dst, ')')
}

// IsEmpty return true if and only if this GeometryCollection doesn't contain
// any elements, or only contains empty elements.
func (c GeometryCollection) IsEmpty() bool {
	for _, g := range c.geoms {
		if !g.IsEmpty() {
			return false
		}
	}
	return true
}

// Dimension returns the maximum dimension over the collection, or 0 if the
// collection is the empty collection. Points and MultiPoints have dimension 0,
// LineStrings and MultiLineStrings have dimension 1, and Polygons and
// MultiPolygons have dimension 2.
func (c GeometryCollection) Dimension() int {
	dim := 0
	for _, g := range c.geoms {
		dim = maxInt(dim, g.Dimension())
	}
	return dim
}

// walk traverses a tree of GeometryCollections, triggering a callback at each
// non-Geometry collection leaf.
func (c GeometryCollection) walk(fn func(Geometry)) {
	for _, g := range c.geoms {
		if g.IsGeometryCollection() {
			g.MustAsGeometryCollection().walk(fn)
		} else {
			fn(g)
		}
	}
}

// Envelope returns the Envelope that most tightly surrounds the geometry.
func (c GeometryCollection) Envelope() Envelope {
	var env Envelope
	for _, g := range c.geoms {
		env = env.ExpandToIncludeEnvelope(g.Envelope())
	}
	return env
}

// Boundary returns the spatial boundary of this GeometryCollection. This is
// the GeometryCollection containing the boundaries of each child geometry.
func (c GeometryCollection) Boundary() GeometryCollection {
	if c.IsEmpty() {
		return c
	}
	var bounds []Geometry
	for _, g := range c.geoms {
		bound := g.Boundary().Force2D()
		if !bound.IsEmpty() {
			bounds = append(bounds, bound)
		}
	}
	return GeometryCollection{bounds, DimXY}
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (c GeometryCollection) Value() (driver.Value, error) {
	return c.AsBinary(), nil
}

// Scan implements the database/sql.Scanner interface by parsing the src value
// as WKB (Well Known Binary).
//
// If the WKB doesn't represent a GeometryCollection geometry, then an error is
// returned.
//
// Geometry constraint validation is performed on the resultant geometry (an
// error will be returned if the geometry is invalid). If this validation isn't
// needed or is undesirable, then the WKB should be scanned into a byte slice
// and then UnmarshalWKB called manually (passing in NoValidate{}).
func (c *GeometryCollection) Scan(src interface{}) error {
	return scanAsType(src, c)
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (c GeometryCollection) AsBinary() []byte {
	return c.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (c GeometryCollection) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaler(dst)
	marsh.writeByteOrder()
	marsh.writeGeomType(TypeGeometryCollection, c.ctype)
	n := c.NumGeometries()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		g := c.GeometryN(i)
		marsh.buf = g.AppendWKB(marsh.buf)
	}
	return marsh.buf
}

// ConvexHull returns the geometry representing the smallest convex geometry
// that contains this geometry.
func (c GeometryCollection) ConvexHull() Geometry {
	return convexHull(c.AsGeometry())
}

// MarshalJSON implements the encoding/json.Marshaler interface by encoding
// this geometry as a GeoJSON geometry object.
func (c GeometryCollection) MarshalJSON() ([]byte, error) {
	buf := []byte(`{"type":"GeometryCollection","geometries":`)
	geoms := c.geoms
	if geoms == nil {
		geoms = []Geometry{}
	}
	geomsJSON, err := json.Marshal(geoms)
	if err != nil {
		return nil, err
	}
	buf = append(buf, geomsJSON...)
	buf = append(buf, '}')
	return buf, nil
}

// UnmarshalJSON implements the encoding/json.Unmarshaler interface by decoding
// the GeoJSON representation of a GeometryCollection.
func (c *GeometryCollection) UnmarshalJSON(buf []byte) error {
	return unmarshalGeoJSONAsType(buf, c)
}

// TransformXY transforms this GeometryCollection into another GeometryCollection according to fn.
func (c GeometryCollection) TransformXY(fn func(XY) XY) GeometryCollection {
	transformed := make([]Geometry, len(c.geoms))
	for i := range c.geoms {
		transformed[i] = c.geoms[i].TransformXY(fn)
	}
	return GeometryCollection{transformed, c.ctype}
}

// Reverse in the case of GeometryCollection reverses each component and also
// returns them in the original order.
func (c GeometryCollection) Reverse() GeometryCollection {
	if c.IsEmpty() {
		return c
	}
	var geoms []Geometry
	for n := 0; n < c.NumGeometries(); n++ {
		rev := c.GeometryN(n).Reverse()
		geoms = append(geoms, rev)
	}
	return GeometryCollection{geoms, c.ctype}
}

// Length of a GeometryCollection is the sum of the lengths of its parts.
func (c GeometryCollection) Length() float64 {
	var sum float64
	n := c.NumGeometries()
	for i := 0; i < n; i++ {
		geom := c.GeometryN(i)
		sum += geom.Length()
	}
	return sum
}

// Area in the case of a GeometryCollection is the sum of the areas of its parts.
func (c GeometryCollection) Area(opts ...AreaOption) float64 {
	var sum float64
	n := c.NumGeometries()
	for i := 0; i < n; i++ {
		g := c.GeometryN(i)
		sum += g.Area(opts...)
	}
	return sum
}

func highestDimensionIgnoreEmpties(g Geometry) int {
	// We could simply use Dimension() instead of this function
	// except for the fact empties can have different dimensionalities.
	// This function thus exists to treat empties as dimensionality 0.
	if g.IsEmpty() {
		return 0
	}
	if !g.IsGeometryCollection() {
		return g.Dimension()
	}
	c := g.MustAsGeometryCollection()
	highestDim := 0
	for _, g2 := range c.geoms {
		highestDim = maxInt(highestDim, highestDimensionIgnoreEmpties(g2))
	}
	return highestDim
}

// Centroid of a GeometryCollection is the centroid of its parts' centroids.
func (c GeometryCollection) Centroid() Point {
	if c.IsEmpty() {
		return NewEmptyPoint(DimXY)
	}
	switch highestDimensionIgnoreEmpties(c.AsGeometry()) {
	case 0:
		return c.pointCentroid()
	case 1:
		return c.linearCentroid()
	case 2:
		return c.arealCentroid()
	default:
		panic("Invalid dimensionality in centroid calculation.")
	}
}

func (c GeometryCollection) pointCentroid() Point {
	var (
		numPoints int
		sumPoints XY
	)
	c.walk(func(g Geometry) {
		switch {
		case g.IsPoint():
			xy, ok := g.MustAsPoint().XY()
			if ok {
				numPoints++
				sumPoints = sumPoints.Add(xy)
			}
		case g.IsMultiPoint():
			mp := g.MustAsMultiPoint()
			for i := 0; i < mp.NumPoints(); i++ {
				xy, ok := mp.PointN(i).XY()
				if ok {
					numPoints++
					sumPoints = sumPoints.Add(xy)
				}
			}
		}
	})
	return sumPoints.Scale(1 / float64(numPoints)).AsPoint()
}

func (c GeometryCollection) linearCentroid() Point {
	var (
		lengthSum        float64
		weightedCentroid XY
	)
	c.walk(func(g Geometry) {
		switch {
		case g.IsLineString():
			ls := g.MustAsLineString()
			centroid, ok := ls.Centroid().XY()
			if ok {
				length := ls.Length()
				lengthSum += length
				weightedCentroid = weightedCentroid.Add(centroid.Scale(length))
			}
		case g.IsMultiLineString():
			mls := g.MustAsMultiLineString()
			for i := 0; i < mls.NumLineStrings(); i++ {
				ls := mls.LineStringN(i)
				centroid, ok := ls.Centroid().XY()
				if ok {
					length := ls.Length()
					lengthSum += length
					weightedCentroid = weightedCentroid.Add(centroid.Scale(length))
				}
			}
		}
	})
	return weightedCentroid.Scale(1 / lengthSum).AsPoint()
}

func (c GeometryCollection) arealCentroid() Point {
	var (
		areas   []float64
		areaSum float64
	)
	c.walk(func(g Geometry) {
		area := g.Area()
		areas = append(areas, area)
		areaSum += area
	})
	var weightedCentroid XY
	c.walk(func(g Geometry) {
		area := areas[0]
		areas = areas[1:]
		centroid, ok := g.Centroid().XY()
		if ok {
			weightedCentroid = weightedCentroid.Add(
				centroid.Scale(area / areaSum))
		}
	})
	return weightedCentroid.AsPoint()
}

// CoordinatesType returns the CoordinatesType used to represent points making
// up the geometry.
func (c GeometryCollection) CoordinatesType() CoordinatesType {
	return c.ctype
}

// ForceCoordinatesType returns a new GeometryCollection with a different CoordinatesType. If
// a dimension is added, then new values are populated with 0.
func (c GeometryCollection) ForceCoordinatesType(newCType CoordinatesType) GeometryCollection {
	gs := make([]Geometry, len(c.geoms))
	for i := range c.geoms {
		gs[i] = c.geoms[i].ForceCoordinatesType(newCType)
	}
	return GeometryCollection{gs, newCType}
}

// Force2D returns a copy of the GeometryCollection with Z and M values removed.
func (c GeometryCollection) Force2D() GeometryCollection {
	return c.ForceCoordinatesType(DimXY)
}

// PointOnSurface returns a Point that's on one of the geometries in the
// collection.
func (c GeometryCollection) PointOnSurface() Point {
	// For collection with mixed dimension, we only consider members who have
	// the highest dimension.
	var maxDim int
	c.walk(func(g Geometry) {
		if !g.IsEmpty() {
			maxDim = maxInt(maxDim, g.Dimension())
		}
	})

	// Find the point-on-surface of a member that is closest to the overall
	// centroid.
	nearest := newNearestPointAccumulator(c.Centroid())
	c.walk(func(g Geometry) {
		if g.Dimension() == maxDim {
			nearest.consider(g.PointOnSurface())
		}
	})

	return nearest.point
}

// ForceCW returns the equivalent GeometryCollection that has its constituent
// Polygons and MultiPolygons reoriented in a clockwise direction (i.e.
// exterior rings clockwise and interior rings counter-clockwise). Geometries
// other that Polygons and MultiPolygons are unchanged.
func (c GeometryCollection) ForceCW() GeometryCollection {
	if c.IsCW() {
		return c
	}
	return c.forceOrientation(true)
}

// ForceCCW returns the equivalent GeometryCollection that has its constituent
// Polygons and MultiPolygons reoriented in a counter-clockwise direction (i.e.
// exterior rings counter-clockwise and interior rings clockwise). Geometries
// other that Polygons and MultiPolygons are unchanged.
func (c GeometryCollection) ForceCCW() GeometryCollection {
	if c.IsCCW() {
		return c
	}
	return c.forceOrientation(false)
}

func (c GeometryCollection) forceOrientation(forceCW bool) GeometryCollection {
	geoms := make([]Geometry, len(c.geoms))
	for i, g := range c.geoms {
		geoms[i] = g.forceOrientation(forceCW)
	}
	return GeometryCollection{geoms, c.ctype}
}

// IsCW returns true iff all contained geometries are CW.
// An empty geometry collection returns true.
func (c GeometryCollection) IsCW() bool {
	for _, g := range c.geoms {
		if !g.IsCW() {
			return false
		}
	}
	return true
}

// IsCCW returns true iff all contained geometries are CCW.
// An empty geometry collection returns true.
func (c GeometryCollection) IsCCW() bool {
	for _, g := range c.geoms {
		if !g.IsCCW() {
			return false
		}
	}
	return true
}

// Dump breaks this GeometryCollection into its constituent non-multi types
// (Points, LineStrings, and Polygons).
//
// The returned slice will only ever contain Points, LineStrings, and Polygons.
func (c GeometryCollection) Dump() []Geometry {
	var gs []Geometry
	for _, g := range c.geoms {
		gs = g.appendDump(gs)
	}
	return gs
}

// DumpCoordinates returns a Sequence holding all control points in the
// GeometryCollection.
func (c GeometryCollection) DumpCoordinates() Sequence {
	var coords []float64
	for _, g := range c.geoms {
		coords = g.DumpCoordinates().appendAllPoints(coords)
	}
	return NewSequence(coords, c.ctype)
}

// Summary returns a text summary of the GeometryCollection following a similar format to https://postgis.net/docs/ST_Summary.html.
func (c GeometryCollection) Summary() string {
	var pointSuffix string
	numPoints := c.DumpCoordinates().Length()
	if numPoints != 1 {
		pointSuffix = "s"
	}

	geometrySuffix := "y"
	numGeometries := c.NumTotalGeometries()
	if numGeometries != 1 {
		geometrySuffix = "ies"
	}

	return fmt.Sprintf("%s[%s] with %d child geometr%s consisting of %d total point%s",
		c.Type(), c.CoordinatesType(), numGeometries, geometrySuffix, numPoints, pointSuffix)
}

// String returns the string representation of the GeometryCollection.
func (c GeometryCollection) String() string {
	return c.Summary()
}

// Simplify returns a simplified version of the GeometryCollection by applying
// Simplify to each child geometry. If simplification causes a child geometry
// to become invalid, then an error is returned. NoValidate{} can be passed in
// to disable geometry constraint validation, potentially resulting in an
// invalid geometry being returned.
func (c GeometryCollection) Simplify(threshold float64, nv ...NoValidate) (GeometryCollection, error) {
	n := c.NumGeometries()
	geoms := make([]Geometry, n)
	for i := 0; i < n; i++ {
		var err error
		geoms[i], err = c.GeometryN(i).Simplify(threshold, nv...)
		if err != nil {
			return GeometryCollection{}, wrapSimplified(err)
		}
	}
	return NewGeometryCollection(geoms).ForceCoordinatesType(c.CoordinatesType()), nil
}

// Densify returns a new GeometryCollection with additional linearly
// interpolated control points such that the distance between any two
// consecutive control points is at most the given maxDistance.
//
// Panics if maxDistance is zero or negative.
func (c GeometryCollection) Densify(maxDistance float64) GeometryCollection {
	gs := make([]Geometry, len(c.geoms))
	for i, g := range c.geoms {
		gs[i] = g.Densify(maxDistance)
	}
	return GeometryCollection{gs, c.ctype}
}

// SnapToGrid returns a copy of the GeometryCollection with all coordinates
// snapped to a base 10 grid.
//
// The grid spacing is specified by the number of decimal places to round to
// (with negative decimal places being allowed). E.g., a decimalPlaces value of
// 2 would cause all coordinates to be rounded to the nearest 0.01, and a
// decimalPlaces of -1 would cause all coordinates to be rounded to the nearest
// 10.
//
// Returned GeometryCollections may be invalid due to snapping, even if the
// input geometry was valid.
func (c GeometryCollection) SnapToGrid(decimalPlaces int) GeometryCollection {
	return c.TransformXY(snapToGridXY(decimalPlaces))
}
