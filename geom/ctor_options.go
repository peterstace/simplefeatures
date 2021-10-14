package geom

// ConstructorOption allows the behaviour of Geometry constructor functions to be modified.
type ConstructorOption func(o *ctorOptionSet)

// DisableAllValidations causes geometry constructors to skip all validations.
// This allows invalid geometries to be loaded, but also has several
// implications for using the resultant geometries:
//
// 1. If the geometry is invalid, then any geometric calculations resulting
// from the geometry may be invalid.
//
// 2. If the geometry is invalid, then invoking geometric calculations may
// cause a panic or infinite loop.
//
// This option should be used with caution. It is most useful when invalid
// geometries need to be loaded, but no geometric calculations will be
// performed.
func DisableAllValidations(o *ctorOptionSet) {
	o.skipValidations = true
}

// OmitInvalid causes geometry constructors to omit any geometries or
// sub-geometries that are invalid.
//
// The behaviour for each geometry type is:
//
// * Point: if the Point is invalid, then it is replaced with an empty Point.
//
// * MultiPoint: if a child Point is invalid, then it is replace with an empty
// Point within the MultiPoint.
//
// * LineString: if the LineString is invalid (e.g. doesn't contain at least 2
// distinct points), then it is replaced with an empty LineString.
//
// * MultiLineString: if a child LineString is invalid, then it is replaced
// with an empty LineString within the MultiLineString.
//
// * Polygon: if the Polygon is invalid (e.g. self intersecting rings or rings
// that intersect in an invalid way), then it is replaced with an empty
// Polygon.
//
// * MultiPolygon: if a child Polygon is invalid, then it is replaced with an
// empty Polygon within the MultiPolygon. If two child Polygons  interact in an
// invalid way, then the MultiPolygon is replaced with an empty MultiPolygon.
func OmitInvalid(o *ctorOptionSet) {
	o.omitInvalid = true
}

type ctorOptionSet struct {
	skipValidations bool
	omitInvalid     bool
}

func newOptionSet(opts []ConstructorOption) ctorOptionSet {
	// Optimise the case where there are no options. This prevents the `cos`
	// variable escaping to the heap.
	if len(opts) == 0 {
		return ctorOptionSet{}
	}

	var cos ctorOptionSet
	for _, opt := range opts {
		opt(&cos)
	}
	return cos
}
