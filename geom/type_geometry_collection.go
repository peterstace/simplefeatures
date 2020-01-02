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
	geoms []GeometryX
}

// NewGeometryCollection creates a potentially heterogenous collection of
// geometries. There are no constraints on the collection.
func NewGeometryCollection(geoms []GeometryX, opts ...ConstructorOption) GeometryCollection {
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

// GeometryN gives the nth (zero based) GeometryX in the GeometryCollection.
func (c GeometryCollection) GeometryN(n int) GeometryX {
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
		dst = ToGeometry(geom).appendWKT(dst)
		if i != len(c.geoms)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ')')
}

func (c GeometryCollection) Intersection(g GeometryX) (GeometryX, error) {
	return intersection(c, g)
}

func (c GeometryCollection) Intersects(g GeometryX) bool {
	return hasIntersection(c, g)
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

func (c GeometryCollection) Equals(other GeometryX) (bool, error) {
	return equals(c, other)
}

// walk traverses a tree of GeometryCollections, triggering a callback at each
// non-GeometryX collection leaf.
func (c GeometryCollection) walk(fn func(GeometryX)) {
	for _, g := range c.geoms {
		if col, ok := g.(GeometryCollection); ok {
			col.walk(fn)
		} else {
			fn(g)
		}
	}
}

func (c GeometryCollection) flatten() []GeometryX {
	var geoms []GeometryX
	c.walk(func(g GeometryX) {
		geoms = append(geoms, g)
	})
	return geoms
}

func (c GeometryCollection) Envelope() (Envelope, bool) {
	return EnvelopeFromGeoms(c.flatten()...)
}

func (c GeometryCollection) Boundary() GeometryX {
	if c.IsEmpty() {
		return c
	}
	var bounds []GeometryX
	for _, g := range c.geoms {
		bound := g.Boundary()
		if !bound.IsEmpty() {
			bounds = append(bounds, bound)
		}
	}
	return NewGeometryCollection(bounds)
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
		marsh.setErr(ToGeometry(geom).AsBinary(w))
	}
	return marsh.err
}

func (c GeometryCollection) ConvexHull() GeometryX {
	return convexHull(c)
}

func (c GeometryCollection) convexHullPointSet() []XY {
	var points []XY
	n := c.NumGeometries()
	for i := 0; i < n; i++ {
		g := c.GeometryN(i)
		points = append(points, g.convexHullPointSet()...)
	}
	return points
}

func (c GeometryCollection) MarshalJSON() ([]byte, error) {
	buf := []byte(`{"type":"GeometryCollection","geometries":`)
	var geoms = c.geoms
	if geoms == nil {
		geoms = []GeometryX{}
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
func (c GeometryCollection) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (GeometryX, error) {
	transformed := make([]GeometryX, len(c.geoms))
	for i := range c.geoms {
		var err error
		transformed[i], err = c.geoms[i].TransformXY(fn, opts...)
		if err != nil {
			return nil, err
		}
	}
	return NewGeometryCollection(transformed), nil
}

// EqualsExact checks if this GeometryCollection is exactly equal to another GeometryCollection.
func (c GeometryCollection) EqualsExact(other GeometryX, opts ...EqualsExactOption) bool {
	o, ok := other.(GeometryCollection)
	return ok && geometryCollectionExactEqual(c, o, opts)
}

// IsValid checks if this GeometryCollection is valid. However, there is no
// constraints on it, so this function always returns true
func (c GeometryCollection) IsValid() bool {
	all := true
	c.walk(func(g GeometryX) {
		if !g.IsValid() {
			all = false
		}
	})
	return all
}
