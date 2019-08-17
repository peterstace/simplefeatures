package geom

type Coordinates struct {
	XY
	// TODO: Put optional Z and M here.
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
	return c.XY.Equals(other.XY)
}
