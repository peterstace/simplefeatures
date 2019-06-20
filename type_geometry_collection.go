package simplefeatures

// GeometryCollection is a collection of 1 or more geometries.
type GeometryCollection struct {
	geoms []Geometry
}

// NewGeometryCollection creates a potentially heterogenous collection of
// geometries. There are no constraints on the collection.
func NewGeometryCollection(geoms []Geometry) GeometryCollection {
	return GeometryCollection{geoms}
}

func (c GeometryCollection) AsText() []byte {
	return c.AppendWKT(nil)
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

func (c GeometryCollection) IsSimple() bool {
	panic("not implemented")
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

func (c GeometryCollection) FiniteNumberOfPoints() (int, bool) {
	pts := make(map[XY]struct{})
	finite := true
	c.walk(func(g Geometry) {
		_, fin := g.FiniteNumberOfPoints()
		if !fin {
			finite = false
		}
		switch g := g.(type) {
		case Point:
			pts[g.coords.XY] = struct{}{}
		case MultiPoint:
			for _, pt := range g.pts {
				pts[pt.coords.XY] = struct{}{}
			}
		}
	})
	return len(pts), finite
}

func (c GeometryCollection) walk(fn func(Geometry)) {
	for _, g := range c.geoms {
		if col, ok := g.(GeometryCollection); ok {
			col.walk(fn)
		} else {
			fn(g)
		}
	}
}
