package simplefeatures

// Polygon is a planar surface, defined by 1 exiterior boundary and 0 or more
// interior boundaries. Each interior boundary defines a hole in the polygon.
type Polygon struct {
	outer LinearRing
	holes []LinearRing
	empty bool
}

// NewPolygon creates a polygon given its outer and inner rings. No rings may
// cross each other, and can only intersect each with each other at a point.
func NewPolygon(outer LinearRing, holes ...LinearRing) (Polygon, error) {
	// TODO: No rings may cross.
	// TODO: Rings may intersect, but only at a point (and only as a tangent).
	// TODO: check linear ring directions?
	return Polygon{outer: outer, holes: holes}, nil
}

func NewEmptyPolygon() Polygon {
	return Polygon{empty: true}
}

func NewPolygonFromCoords(coords [][]Coordinates) (Polygon, error) {
	if len(coords) == 0 {
		return NewEmptyPolygon(), nil
	}
	outer, err := NewLinearRingFromCoords(coords[0])
	if err != nil {
		return Polygon{}, err
	}
	var holes []LinearRing
	for _, holeCoords := range coords[1:] {
		hole, err := NewLinearRingFromCoords(holeCoords)
		if err != nil {
			return Polygon{}, err
		}
		holes = append(holes, hole)
	}
	return NewPolygon(outer, holes...)
}
