package simplefeatures

// Polygon is a planar surface, defined by 1 exiterior boundary and 0 or more
// interior boundaries. Each interior boundary defines a hole in the polygon.
type Polygon struct {
	outer LinearRing
	holes []LinearRing
}

// NewPolygon creates a polygon given its outer and inner rings. No rings may cross each other, and can only intersect each with each other at a point.
func NewPolygon(outer LinearRing, holes ...LinearRing) (Polygon, error) {
	// TODO: No rings may cross.
	// TODO: Rings may intersect, but only at a point (and only as a tangent).
	return Polygon{outer, holes}, nil
}
