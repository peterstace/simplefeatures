package geom

import (
	"database/sql/driver"
	"encoding/json"
	"unsafe"
)

// GeometryCollection is a non-homogeneous collection of geometries. Its zero
// value is the empty GeometryCollection (i.e. a collection of zero
// geometries).
type GeometryCollection struct {
	geoms []Geometry
	ctype CoordinatesType
}

// NewGeometryCollection creates a collection of geometries. The coordinates
// type of the GeometryCollection is the lowest common coordinates type of its
// child geometries.
func NewGeometryCollection(geoms []Geometry, opts ...ConstructorOption) GeometryCollection {
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

// Type returns the GeometryType for a GeometryCollection
func (c GeometryCollection) Type() GeometryType {
	return TypeGeometryCollection
}

// AsGeometry converts this GeometryCollection into a Geometry.
func (c GeometryCollection) AsGeometry() Geometry {
	return Geometry{TypeGeometryCollection, unsafe.Pointer(&c)}
}

// NumGeometries gives the number of Geomety elements is the GeometryCollection.
func (c GeometryCollection) NumGeometries() int {
	return len(c.geoms)
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

// Intersects return true if and only if this geometry intersects with the
// other, i.e. they shared at least one common point.
func (c GeometryCollection) Intersects(g Geometry) bool {
	return hasIntersection(c.AsGeometry(), g)
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
		dim = max(dim, g.Dimension())
	}
	return dim
}

// walk traverses a tree of GeometryCollections, triggering a callback at each
// non-Geometry collection leaf.
func (c GeometryCollection) walk(fn func(Geometry)) {
	for _, g := range c.geoms {
		if g.IsGeometryCollection() {
			g.AsGeometryCollection().walk(fn)
		} else {
			fn(g)
		}
	}
}

func (c GeometryCollection) flatten() []Geometry {
	var geoms []Geometry
	c.walk(func(g Geometry) {
		geoms = append(geoms, g)
	})
	return geoms
}

// Envelope returns the Envelope that most tightly surrounds the geometry. If
// the geometry is empty, then false is returned.
func (c GeometryCollection) Envelope() (Envelope, bool) {
	return EnvelopeFromGeoms(c.flatten()...)
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

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (c GeometryCollection) AsBinary() []byte {
	return c.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (c GeometryCollection) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaller(dst)
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

// MarshalJSON implements the encoding/json.Marshaller interface by encoding
// this geometry as a GeoJSON geometry object.
func (c GeometryCollection) MarshalJSON() ([]byte, error) {
	buf := []byte(`{"type":"GeometryCollection","geometries":`)
	var geoms = c.geoms
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

// TransformXY transforms this GeometryCollection into another GeometryCollection according to fn.
func (c GeometryCollection) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (GeometryCollection, error) {
	transformed := make([]Geometry, len(c.geoms))
	for i := range c.geoms {
		var err error
		transformed[i], err = c.geoms[i].TransformXY(fn, opts...)
		if err != nil {
			return GeometryCollection{}, err
		}
	}
	return GeometryCollection{transformed, c.ctype}, nil
}

// EqualsExact checks if this GeometryCollection is exactly equal to another GeometryCollection.
func (c GeometryCollection) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	return other.IsGeometryCollection() &&
		geometryCollectionExactEqual(c, other.AsGeometryCollection(), opts)
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
func (c GeometryCollection) Area() float64 {
	var sum float64
	n := c.NumGeometries()
	for i := 0; i < n; i++ {
		g := c.GeometryN(i)
		sum += g.Area()
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
	c := g.AsGeometryCollection()
	highestDim := 0
	for _, g2 := range c.geoms {
		highestDim = max(highestDim, highestDimensionIgnoreEmpties(g2))
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
			xy, ok := g.AsPoint().XY()
			if ok {
				numPoints++
				sumPoints = sumPoints.Add(xy)
			}
		case g.IsMultiPoint():
			mp := g.AsMultiPoint()
			for i := 0; i < mp.NumPoints(); i++ {
				xy, ok := mp.PointN(i).XY()
				if ok {
					numPoints++
					sumPoints = sumPoints.Add(xy)
				}
			}
		}
	})
	return NewPointFromXY(sumPoints.Scale(1 / float64(numPoints)))
}

func (c GeometryCollection) linearCentroid() Point {
	var (
		lengthSum        float64
		weightedCentroid XY
	)
	c.walk(func(g Geometry) {
		switch {
		case g.IsLineString():
			ls := g.AsLineString()
			centroid, ok := ls.Centroid().XY()
			if ok {
				length := ls.Length()
				lengthSum += length
				weightedCentroid = weightedCentroid.Add(centroid.Scale(length))
			}
		case g.IsMultiLineString():
			mls := g.AsMultiLineString()
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
	return NewPointFromXY(weightedCentroid.Scale(1 / lengthSum))
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
	return NewPointFromXY(weightedCentroid)
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
			maxDim = max(maxDim, g.Dimension())
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
