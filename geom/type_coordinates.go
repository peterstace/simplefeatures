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
	buf = c.XY.X.appendAsFloat(buf)
	buf = append(buf, ',')
	buf = c.XY.Y.appendAsFloat(buf)
	buf = append(buf, ']')
	return buf, nil
}
