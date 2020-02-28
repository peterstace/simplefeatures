package geom

type Coordinates struct {
	XY
	// TODO: Put optional Z and M here.
}

func oneDimXYToCoords(pts []XY) []Coordinates {
	coords := make([]Coordinates, len(pts))
	for i, pt := range pts {
		coords[i] = Coordinates{pt}
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

type OptionalCoordinates struct {
	Empty bool
	Value Coordinates
}

func (c Coordinates) MarshalJSON() ([]byte, error) {
	buf := []byte{'['}
	buf = appendFloat(buf, c.XY.X)
	buf = append(buf, ',')
	buf = appendFloat(buf, c.XY.Y)
	buf = append(buf, ']')
	return buf, nil
}

func (c Coordinates) Equals(other Coordinates) bool {
	return c.XY == other.XY
}

func (oc OptionalCoordinates) MarshalJSON() ([]byte, error) {
	buf := []byte{'['}
	if !oc.Empty {
		buf = appendFloat(buf, oc.Value.X)
		buf = append(buf, ',')
		buf = appendFloat(buf, oc.Value.Y)
	}
	buf = append(buf, ']')
	return buf, nil
}
