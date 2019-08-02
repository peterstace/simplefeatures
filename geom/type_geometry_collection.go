package geom

import (
	"database/sql/driver"
	"encoding/json"
	"io"
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
func NewGeometryCollection(geoms []Geometry) GeometryCollection {
	if len(geoms) == 0 {
		// Store empty geoms as nil to make testing easier.
		geoms = nil
	}
	return GeometryCollection{geoms}
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
		dst = geom.AppendWKT(dst)
		if i != len(c.geoms)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ')')
}

func (c GeometryCollection) Intersection(g Geometry) Geometry {
	return intersection(c, g)
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

func (c GeometryCollection) Equals(other Geometry) bool {
	return equals(c, other)
}

// walk traverses a tree of GeometryCollections, triggering a callback at each
// non-Geometry collection leaf.
func (c GeometryCollection) walk(fn func(Geometry)) {
	for _, g := range c.geoms {
		if col, ok := g.(GeometryCollection); ok {
			col.walk(fn)
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
		return c
	}
	var bounds []Geometry
	for _, g := range c.geoms {
		bound := g.Boundary()
		if !bound.IsEmpty() {
			bounds = append(bounds, bound)
		}
	}
	return NewGeometryCollection(bounds)
}

func (c GeometryCollection) Value() (driver.Value, error) {
	return c.AsText(), nil
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
