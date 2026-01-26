package jts

import (
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Standard ordinate index values for Geom_CoordinateXYM.
const Geom_CoordinateXYM_X = 0
const Geom_CoordinateXYM_Y = 1
const Geom_CoordinateXYM_Z = -1 // Geom_CoordinateXYM does not support Z values.
const Geom_CoordinateXYM_M = 2  // Standard ordinate index value for M in XYM sequences.

// Geom_CoordinateXYM is a Geom_Coordinate subclass supporting XYM ordinates.
//
// This data object is suitable for use with coordinate sequences with
// dimension = 3 and measures = 1.
//
// The Geom_Coordinate.Z field is visible, but intended to be ignored.
type Geom_CoordinateXYM struct {
	*Geom_Coordinate
	M      float64
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (c *Geom_CoordinateXYM) GetChild() java.Polymorphic {
	return c.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (c *Geom_CoordinateXYM) GetParent() java.Polymorphic {
	return c.Geom_Coordinate
}

// Geom_NewCoordinateXYM3D constructs a Geom_CoordinateXYM at (0,0) with M=0.
func Geom_NewCoordinateXYM3D() *Geom_CoordinateXYM {
	coord := &Geom_Coordinate{
		X: 0,
		Y: 0,
		Z: math.NaN(),
	}
	c := &Geom_CoordinateXYM{
		Geom_Coordinate: coord,
		M:               0.0,
	}
	coord.child = c
	return c
}

// Geom_NewCoordinateXYM3DWithXYM constructs a Geom_CoordinateXYM instance with the given
// ordinates and measure.
func Geom_NewCoordinateXYM3DWithXYM(x, y, m float64) *Geom_CoordinateXYM {
	coord := &Geom_Coordinate{
		X: x,
		Y: y,
		Z: Geom_Coordinate_NullOrdinate,
	}
	c := &Geom_CoordinateXYM{
		Geom_Coordinate: coord,
		M:               m,
	}
	coord.child = c
	return c
}

// Geom_NewCoordinateXYM3DFromCoordinate constructs a Geom_CoordinateXYM instance with
// the x and y ordinates of the given Geom_Coordinate.
func Geom_NewCoordinateXYM3DFromCoordinate(other *Geom_Coordinate) *Geom_CoordinateXYM {
	coord := &Geom_Coordinate{
		X: other.X,
		Y: other.Y,
		Z: math.NaN(),
	}
	c := &Geom_CoordinateXYM{
		Geom_Coordinate: coord,
		M:               other.GetM(),
	}
	coord.child = c
	return c
}

// Geom_NewCoordinateXYM3DFromCoordinateXYM constructs a Geom_CoordinateXYM instance with
// the x and y ordinates of the given Geom_CoordinateXYM.
func Geom_NewCoordinateXYM3DFromCoordinateXYM(other *Geom_CoordinateXYM) *Geom_CoordinateXYM {
	coord := &Geom_Coordinate{
		X: other.X,
		Y: other.Y,
		Z: math.NaN(),
	}
	c := &Geom_CoordinateXYM{
		Geom_Coordinate: coord,
		M:               other.M,
	}
	coord.child = c
	return c
}

func (c *Geom_CoordinateXYM) Copy_BODY() *Geom_Coordinate {
	return Geom_NewCoordinateXYM3DFromCoordinateXYM(c).Geom_Coordinate
}

func (c *Geom_CoordinateXYM) Create_BODY() *Geom_Coordinate {
	return Geom_NewCoordinateXYM3D().Geom_Coordinate
}

func (c *Geom_CoordinateXYM) GetM_BODY() float64 {
	return c.M
}

func (c *Geom_CoordinateXYM) SetM_BODY(m float64) {
	c.M = m
}

func (c *Geom_CoordinateXYM) GetZ_BODY() float64 {
	return Geom_Coordinate_NullOrdinate
}

func (c *Geom_CoordinateXYM) SetZ_BODY(z float64) {
	panic("Geom_CoordinateXYM dimension 2 does not support z-ordinate")
}

func (c *Geom_CoordinateXYM) SetCoordinate_BODY(other *Geom_Coordinate) {
	c.X = other.X
	c.Y = other.Y
	c.Z = other.GetZ()
	c.M = other.GetM()
}

func (c *Geom_CoordinateXYM) GetOrdinate_BODY(ordinateIndex int) float64 {
	switch ordinateIndex {
	case Geom_CoordinateXYM_X:
		return c.X
	case Geom_CoordinateXYM_Y:
		return c.Y
	case Geom_CoordinateXYM_M:
		return c.M
	}
	panic(fmt.Sprintf("Invalid ordinate index: %d", ordinateIndex))
}

func (c *Geom_CoordinateXYM) SetOrdinate_BODY(ordinateIndex int, value float64) {
	switch ordinateIndex {
	case Geom_CoordinateXYM_X:
		c.X = value
	case Geom_CoordinateXYM_Y:
		c.Y = value
	case Geom_CoordinateXYM_M:
		c.M = value
	default:
		panic(fmt.Sprintf("Invalid ordinate index: %d", ordinateIndex))
	}
}

func (c *Geom_CoordinateXYM) String_BODY() string {
	return fmt.Sprintf("(%v, %v m=%v)", c.X, c.Y, c.GetM())
}
