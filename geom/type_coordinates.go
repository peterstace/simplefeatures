package geom

type Coordinates struct {
	XY
	Z float64
	M float64
}

func oneDimXYToCoords(pts []XY) []Coordinates {
	coords := make([]Coordinates, len(pts))
	for i, pt := range pts {
		coords[i] = Coordinates{XY: pt}
	}
	return coords
}

func twoDimXYToCoords(xys [][]XY) [][]Coordinates {
	coords := make([][]Coordinates, len(xys))
	for i, pts := range xys {
		coords[i] = oneDimXYToCoords(pts)
	}
	return coords
}

func threeDimXYToCoords(xys [][][]XY) [][][]Coordinates {
	coords := make([][][]Coordinates, len(xys))
	for i, pts := range xys {
		coords[i] = twoDimXYToCoords(pts)
	}
	return coords
}

// TODO: Add Z and M values
// TODO: Want to have these methods removed.
func (c Coordinates) MarshalJSON() ([]byte, error) {
	buf := []byte{'['}
	buf = appendFloat(buf, c.XY.X)
	buf = append(buf, ',')
	buf = appendFloat(buf, c.XY.Y)
	buf = append(buf, ']')
	return buf, nil
}
