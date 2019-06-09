package simplefeatures

// MultiPolygon is a multi surface whose elements are polygons.
type MultiPolygon struct {
	polys []Polygon
}

func NewMultiPolygon(polys []Polygon) (MultiPolygon, error) {
	// TODO: The interiors of 2 polygons may not intersect.
	// TODO: The boundaries of 2 polygons may touch only at a finite number of points.
	return MultiPolygon{polys}, nil
}

func NewMultiPolygonFromCoords(coords [][][]Coordinates) (MultiPolygon, error) {
	polys := make([]Polygon, len(coords))
	for i, c := range coords {
		poly, err := NewPolygonFromCoords(c)
		if err != nil {
			return MultiPolygon{}, err
		}
		polys[i] = poly
	}
	return NewMultiPolygon(polys)
}
