package gatig

import "errors"

type Geometry interface {
	// TODO - not yet clear what this interface should be
}

// GeometryCollection is a collection of 1 or more geometries.
type GeometryCollection struct {
	geoms []Geometry
}

// NewGeometryCollection creates a collection of 1 or more Geometries. There
// are no constraints on the geometries in the collection.
func NewGeometryCollection(geoms []Geometry) (GeometryCollection, error) {
	if len(geoms) == 0 {
		return GeometryCollection{}, errors.New("geometry collection must contain at least 1 geometry")
	}
	return GeometryCollection{geoms}, nil
}
