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
