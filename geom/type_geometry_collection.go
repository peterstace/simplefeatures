package geom

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"io"
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

// Type return type string for GeometryCollection
func (c GeometryCollection) Type() string {
	return geometryCollectionType
}

// AsGeometry converts this GeometryCollection into a Geometry.
func (c GeometryCollection) AsGeometry() Geometry {
	return Geometry{geometryCollectionTag, unsafe.Pointer(&c)}
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

// Intersection calculates the intersection between this geometry and another
// (i.e. the overlap between the two geometries). It is not implemented for all
// geometry pairs, and returns an error for those cases.
func (c GeometryCollection) Intersection(g Geometry) (Geometry, error) {
	return intersection(c.AsGeometry(), g)
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
	var buf bytes.Buffer
	err := c.AsBinary(&buf)
	return buf.Bytes(), err
}

// AsBinary writes the WKB (Well Known Binary) representation of the geometry
// to the writer.
func (c GeometryCollection) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeGeometryCollection, c.ctype)
	n := c.NumGeometries()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		g := c.GeometryN(i)
		marsh.setErr(g.AsBinary(w))
	}
	return marsh.err
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

// IsValid checks if this GeometryCollection is valid. However, there is no
// constraints on it, so this function always returns true
func (c GeometryCollection) IsValid() bool {
	all := true
	c.walk(func(g Geometry) {
		if !g.IsValid() {
			all = false
		}
	})
	return all
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
	result := c.sumCentroidCalc()
	var xy XY
	switch result.highestDim {
	case 0:
		if result.numPoints == 0 {
			return NewEmptyPoint(DimXY) // Invalid centroid, highestDim is 0 and numPoints is 0
		}
		xy = result.sumXY.Scale(1.0 / float64(result.numPoints))
	case 1:
		if result.sumLength == 0 {
			return NewEmptyPoint(DimXY) // Invalid centroid, highestDim is 1 and sumLength is 0
		}
		xy = result.sumXY.Scale(1.0 / result.sumLength)
	case 2:
		if result.sumArea == 0 {
			return NewEmptyPoint(DimXY) // Invalid centroid, highestDim is 2 and sumArea is 0
		}
		xy = result.sumXY.Scale(1.0 / result.sumArea)
	default:
		panic("Invalid dimensionality in centroid calculation.")
	}

	return NewPointFromXY(xy)
}

type centroidCalc struct {
	highestDim int
	numPoints  int
	sumLength  float64
	sumArea    float64
	sumXY      XY
}

func (c GeometryCollection) sumCentroidCalc() centroidCalc {
	var result centroidCalc

	n := c.NumGeometries()
	if n == 0 {
		return result
	}

	result.highestDim = highestDimensionIgnoreEmpties(c.AsGeometry())

	for i := 0; i < n; i++ {
		g := c.GeometryN(i)
		if highestDimensionIgnoreEmpties(g) != result.highestDim {
			continue
		}
		switch {
		case g.IsPoint():
			xy, ok := g.AsPoint().XY()
			if ok {
				result.sumXY = result.sumXY.Add(xy)
				result.numPoints++
			}
		case g.IsMultiPoint():
			mp := g.AsMultiPoint()
			for m := 0; m < mp.NumPoints(); m++ {
				xy, ok := mp.PointN(i).XY()
				if ok {
					result.sumXY = result.sumXY.Add(xy)
					result.numPoints++
				}
			}
		case g.IsLine():
			line := g.AsLine()
			cent, ok := line.Centroid().XY()
			if ok {
				length := line.Length()
				result.sumXY = result.sumXY.Add(cent.Scale(length))
				result.sumLength += length
			}
		case g.IsLineString():
			ls := g.AsLineString()
			xy, length := sumCentroidAndLengthOfLineString(ls)
			result.sumXY = result.sumXY.Add(xy)
			result.sumLength += length
		case g.IsMultiLineString():
			mls := g.AsMultiLineString()
			for m := 0; m < mls.NumLineStrings(); m++ {
				ls := mls.LineStringN(m)
				xy, length := sumCentroidAndLengthOfLineString(ls)
				result.sumXY = result.sumXY.Add(xy)
				result.sumLength += length
			}
		case g.IsPolygon():
			poly := g.AsPolygon()
			xy, area := sumCentroidAndAreaOfPolygon(poly)
			result.sumXY = result.sumXY.Add(xy)
			result.sumArea += area
		case g.IsMultiPolygon():
			mp := g.AsMultiPolygon()
			for p := 0; p < mp.NumPolygons(); p++ {
				poly := mp.PolygonN(p)
				xy, area := sumCentroidAndAreaOfPolygon(poly)
				result.sumXY = result.sumXY.Add(xy)
				result.sumArea += area
			}
		case g.IsGeometryCollection():
			child := g.AsGeometryCollection().sumCentroidCalc()
			if child.highestDim != result.highestDim {
				continue // ignore
			}
			result.sumXY = result.sumXY.Add(child.sumXY)
			result.numPoints += child.numPoints
			result.sumLength += child.sumLength
			result.sumArea += child.sumArea
		default:
			panic("unknown geometry type in centroid computation")
		}
	}

	return result
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
