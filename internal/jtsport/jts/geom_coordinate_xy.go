package jts

import (
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Standard ordinate index values for Geom_CoordinateXY.
const Geom_CoordinateXY_X = 0
const Geom_CoordinateXY_Y = 1
const Geom_CoordinateXY_Z = -1 // Geom_CoordinateXY does not support Z values.
const Geom_CoordinateXY_M = -1 // Geom_CoordinateXY does not support M measures.

// Geom_CoordinateXY is a Geom_Coordinate subclass supporting XY ordinates.
//
// This data object is suitable for use with coordinate sequences with
// dimension = 2.
//
// The Geom_Coordinate.Z field is visible, but intended to be ignored.
type Geom_CoordinateXY struct {
	*Geom_Coordinate
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (c *Geom_CoordinateXY) GetChild() java.Polymorphic {
	return c.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (c *Geom_CoordinateXY) GetParent() java.Polymorphic {
	return c.Geom_Coordinate
}

// Geom_NewCoordinateXY2D constructs a Geom_CoordinateXY at (0,0).
func Geom_NewCoordinateXY2D() *Geom_CoordinateXY {
	coord := &Geom_Coordinate{
		X: 0,
		Y: 0,
		Z: math.NaN(),
	}
	c := &Geom_CoordinateXY{Geom_Coordinate: coord}
	coord.child = c
	return c
}

// Geom_NewCoordinateXY2DWithXY constructs a Geom_CoordinateXY instance with the given
// ordinates.
func Geom_NewCoordinateXY2DWithXY(x, y float64) *Geom_CoordinateXY {
	coord := &Geom_Coordinate{
		X: x,
		Y: y,
		Z: Geom_Coordinate_NullOrdinate,
	}
	c := &Geom_CoordinateXY{Geom_Coordinate: coord}
	coord.child = c
	return c
}

// Geom_NewCoordinateXY2DFromCoordinate constructs a Geom_CoordinateXY instance with the
// x and y ordinates of the given Geom_Coordinate.
func Geom_NewCoordinateXY2DFromCoordinate(other *Geom_Coordinate) *Geom_CoordinateXY {
	coord := &Geom_Coordinate{
		X: other.X,
		Y: other.Y,
		Z: math.NaN(),
	}
	c := &Geom_CoordinateXY{Geom_Coordinate: coord}
	coord.child = c
	return c
}

// Geom_NewCoordinateXY2DFromCoordinateXY constructs a Geom_CoordinateXY instance with
// the x and y ordinates of the given Geom_CoordinateXY.
func Geom_NewCoordinateXY2DFromCoordinateXY(other *Geom_CoordinateXY) *Geom_CoordinateXY {
	coord := &Geom_Coordinate{
		X: other.X,
		Y: other.Y,
		Z: math.NaN(),
	}
	c := &Geom_CoordinateXY{Geom_Coordinate: coord}
	coord.child = c
	return c
}

func (c *Geom_CoordinateXY) Copy_BODY() *Geom_Coordinate {
	return Geom_NewCoordinateXY2DFromCoordinateXY(c).Geom_Coordinate
}

func (c *Geom_CoordinateXY) Create_BODY() *Geom_Coordinate {
	return Geom_NewCoordinateXY2D().Geom_Coordinate
}

func (c *Geom_CoordinateXY) GetZ_BODY() float64 {
	return Geom_Coordinate_NullOrdinate
}

func (c *Geom_CoordinateXY) SetZ_BODY(z float64) {
	panic("Geom_CoordinateXY dimension 2 does not support z-ordinate")
}

func (c *Geom_CoordinateXY) SetCoordinate_BODY(other *Geom_Coordinate) {
	c.X = other.X
	c.Y = other.Y
	c.Z = other.GetZ()
}

func (c *Geom_CoordinateXY) GetOrdinate_BODY(ordinateIndex int) float64 {
	switch ordinateIndex {
	case Geom_CoordinateXY_X:
		return c.X
	case Geom_CoordinateXY_Y:
		return c.Y
	}
	return math.NaN()
}

func (c *Geom_CoordinateXY) SetOrdinate_BODY(ordinateIndex int, value float64) {
	switch ordinateIndex {
	case Geom_CoordinateXY_X:
		c.X = value
	case Geom_CoordinateXY_Y:
		c.Y = value
	default:
		panic(fmt.Sprintf("Invalid ordinate index: %d", ordinateIndex))
	}
}

func (c *Geom_CoordinateXY) String_BODY() string {
	return fmt.Sprintf("(%v, %v)", c.X, c.Y)
}
