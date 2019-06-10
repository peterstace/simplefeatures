package simplefeatures

import (
	"errors"
	"math"
)

type XY struct {
	X, Y float64
}

type Coordinates struct {
	XY
	// TODO: Put optional Z and M here.
}

// TODO: should the Validate function be on the XY instead of the Coordinates?

// Validate checks if the coordinates could represent a valid point.
// Coordinates can represent valid points if their X and Y values are not -Inf
// or +Inf, and are not NaN.
func (c Coordinates) Validate() error {
	if math.IsNaN(c.X) || math.IsNaN(c.Y) {
		return errors.New("coordinate is NaN")
	}
	if math.IsInf(c.X, 0) || math.IsInf(c.Y, 0) {
		return errors.New("coordinate is Inf")
	}
	return nil
}

type OptionalCoordinates struct {
	Empty bool
	Value Coordinates
}
