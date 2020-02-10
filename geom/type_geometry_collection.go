package geom

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"io"
	"unsafe"
)

// GeometryCollection is a collection of geometries.
//
// Its assertions are:
//
// 1. It must contain only valid geometries.
//
// 2. It must contain zero or more geometries.
type GeometryCollection struct {
	geoms []Geometry
}

// NewGeometryCollection creates a potentially heterogenous collection of
// geometries. There are no constraints on the collection.
func NewGeometryCollection(geoms []Geometry, opts ...ConstructorOption) GeometryCollection {
	if len(geoms) == 0 {
		// Store empty geoms as nil to make testing easier.
		geoms = nil
	}
	return GeometryCollection{geoms}
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

func (c GeometryCollection) AsText() string {
	return string(c.AppendWKT(nil))
}

func (c GeometryCollection) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("GEOMETRYCOLLECTION")...)
	if len(c.geoms) == 0 {
		return append(dst, []byte(" EMPTY")...)
	}
	dst = append(dst, '(')
	for i, geom := range c.geoms {
		dst = geom.appendWKT(dst)
		if i != len(c.geoms)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ')')
}

func (c GeometryCollection) Intersection(g Geometry) (Geometry, error) {
	return intersection(c.AsGeometry(), g)
}

func (c GeometryCollection) Intersects(g Geometry) bool {
	return hasIntersection(c.AsGeometry(), g)
}

func (c GeometryCollection) IsEmpty() bool {
	for _, g := range c.geoms {
		if !g.IsEmpty() {
			return false
		}
	}
	return true
}

func (c GeometryCollection) Dimension() int {
	dim := 0
	for _, g := range c.geoms {
		dim = max(dim, g.Dimension())
	}
	return dim
}

func (c GeometryCollection) Equals(other Geometry) (bool, error) {
	return equals(c.AsGeometry(), other)
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

func (c GeometryCollection) Envelope() (Envelope, bool) {
	return EnvelopeFromGeoms(c.flatten()...)
}

func (c GeometryCollection) Boundary() Geometry {
	if c.IsEmpty() {
		return c.AsGeometry()
	}
	var bounds []Geometry
	for _, g := range c.geoms {
		bound := g.Boundary()
		if !bound.IsEmpty() {
			bounds = append(bounds, bound)
		}
	}
	return NewGeometryCollection(bounds).AsGeometry()
}

func (c GeometryCollection) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := c.AsBinary(&buf)
	return buf.Bytes(), err
}

func (c GeometryCollection) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeGeometryCollection)
	n := c.NumGeometries()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		geom := c.GeometryN(i)
		marsh.setErr(geom.AsBinary(w))
	}
	return marsh.err
}

func (c GeometryCollection) ConvexHull() Geometry {
	return convexHull(c.AsGeometry())
}

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
func (c GeometryCollection) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	transformed := make([]Geometry, len(c.geoms))
	for i := range c.geoms {
		var err error
		transformed[i], err = c.geoms[i].TransformXY(fn, opts...)
		if err != nil {
			return Geometry{}, err
		}
	}
	return NewGeometryCollection(transformed).AsGeometry(), nil
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

// Reverse in the case of GeometryCollection reverses each component and
// also returns them in the original order. It also omits empty components.
func (c GeometryCollection) Reverse() GeometryCollection {
	var geoms []Geometry
	var numEmpty int
	for n := 0; n < c.NumGeometries(); n++ {
		if c.GeometryN(n).IsEmpty() {
			numEmpty++
		}
	}
	if numEmpty == c.NumGeometries() {
		return NewGeometryCollection(geoms)
	}
	for n := 0; n < c.NumGeometries(); n++ {
		rev := c.GeometryN(n).Reverse()
		geoms = append(geoms, rev)
	}
	return NewGeometryCollection(geoms)
}

// Area in the case of a GeometryCollection is the sum of the areas of its parts.
func (c GeometryCollection) Area() float64 {
	var sum float64
	n := c.NumGeometries()
	for i := 0; i < n; i++ {
		geom := c.GeometryN(i)
		sum += geom.Area()
	}
	return sum
}

func highestDimensionIgnoreEmpties(geom Geometry) int {
	// We could simply use Dimension() instead of this function
	// except for the fact empties can have different dimensionalities.
	// This function thus exists to treat empties as dimensionality 0.
	if geom.IsEmpty() {
		return 0
	}
	if !geom.IsGeometryCollection() {
		return geom.Dimension()
	}
	c := geom.AsGeometryCollection()
	highestDim := 0
	for _, geom := range c.geoms {
		highestDim = max(highestDim, highestDimensionIgnoreEmpties(geom))
	}
	return highestDim
}

// Centroid of a GeometryCollection is the centroid of its parts' centroids.
// It returns true iff the centroid is well defined.
func (c GeometryCollection) Centroid() (Point, bool) {
	highestDim, numPoints, sumLength, sumArea, sumX, sumY, valid := c.computeCentroid()
	var xy XY
	switch highestDim {
	case 0:
		if numPoints == 0 {
			panic("Invalid centroid, highestDim is 0 and numPoints is 0\n")
		}
		xy = XY{sumX / float64(numPoints), sumY / float64(numPoints)}
	case 1:
		if sumLength == 0 {
			panic("Invalid centroid, highestDim is 1 and sumLength is 0\n")
		}
		xy = XY{sumX / sumLength, sumY / sumLength}
	case 2:
		if sumArea == 0 {
			panic("Invalid centroid, highestDim is 2 and sumArea is 0\n")
		}
		xy = XY{sumX / sumArea, sumY / sumArea}
	default:
		panic("Invalid dimensionality in centroid calculation.")
	}

	return NewPointXY(xy), valid
}

func (c GeometryCollection) computeCentroid() (highestDim, numPoints int, sumLength, sumArea, sumX, sumY float64, valid bool) {
	n := c.NumGeometries()
	if n == 0 {
		panic("Invalid, empty collection\n")
	}

	highestDim = highestDimensionIgnoreEmpties(c.AsGeometry())

	for i := 0; i < n; i++ {
		geom := c.GeometryN(i)
		if highestDimensionIgnoreEmpties(geom) != highestDim {
			continue
		}
		switch {
		case geom.IsPoint():
			pt := geom.AsPoint()
			sumX += pt.XY().X
			sumY += pt.XY().Y
			numPoints++
		case geom.IsMultiPoint():
			mp := geom.AsMultiPoint()
			for m := 0; m < mp.NumPoints(); m++ {
				pt := mp.PointN(m)
				sumX += pt.XY().X
				sumY += pt.XY().Y
				numPoints++
			}
		case geom.IsLine():
			line := geom.AsLine()
			cent := line.Centroid()
			length := line.Length()
			sumX += cent.Coordinates().Scale(length).X
			sumY += cent.Coordinates().Scale(length).Y
			sumLength += length
		case geom.IsLineString():
			ls := geom.AsLineString()
			x, y, length := centroidAndLengthOfLineString(ls)
			sumX += x
			sumY += y
			sumLength += length
		case geom.IsMultiLineString():
			mls := geom.AsMultiLineString()
			for m := 0; m < mls.NumLineStrings(); m++ {
				ls := mls.LineStringN(m)
				x, y, length := centroidAndLengthOfLineString(ls)
				sumX += x
				sumY += y
				sumLength += length
			}
		case geom.IsPolygon():
			poly := geom.AsPolygon()
			x, y, area := centroidAndAreaOfPolygon(poly)
			sumX += x
			sumY += y
			sumArea += area
		case geom.IsMultiPolygon():
			mp := geom.AsMultiPolygon()
			for p := 0; p < mp.NumPolygons(); p++ {
				poly := mp.PolygonN(p)
				x, y, area := centroidAndAreaOfPolygon(poly)
				sumX += x
				sumY += y
				sumArea += area
			}
		case geom.IsGeometryCollection():
			dim, n, length, area, x, y, valid := geom.AsGeometryCollection().computeCentroid()
			if !valid {
				continue
			}
			if dim < highestDim {
				continue // ignore
			}
			sumX += x
			sumY += y
			numPoints += n
			sumLength += length
			sumArea += area
		default:
			panic("unknown geometry type in centroid computation")
		}
	}
	valid = true
	return
}
