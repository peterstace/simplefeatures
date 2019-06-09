package simplefeatures

// MultiPoint is a 0-dimensional geometric collection of points. The points are
// not connected or ordered.
type MultiPoint struct {
	pts []Point
}

func NewMultiPoint(pts []Point) (MultiPoint, error) {
	// TODO: error checking
	return MultiPoint{pts}, nil
}

func NewMultiPointFromCoords(coords []OptionalCoordinates) (MultiPoint, error) {
	pts := make([]Point, len(coords))
	for i, c := range coords {
		pt, err := NewPointFromOptionalCoords(c)
		if err != nil {
			return MultiPoint{}, err
		}
		pts[i] = pt
	}
	return MultiPoint{pts}, nil
}
