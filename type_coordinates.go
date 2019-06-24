package simplefeatures

type Coordinates struct {
	XY
	// TODO: Put optional Z and M here.
}

type OptionalCoordinates struct {
	Empty bool
	Value Coordinates
}
