package jts

import (
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geom_CoordinateXYZM is a Geom_Coordinate subclass supporting XYZM ordinates.
//
// This data object is suitable for use with coordinate sequences with
// dimension = 4 and measures = 1.
type Geom_CoordinateXYZM struct {
	*Geom_Coordinate
	M      float64
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (c *Geom_CoordinateXYZM) GetChild() java.Polymorphic {
	return c.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (c *Geom_CoordinateXYZM) GetParent() java.Polymorphic {
	return c.Geom_Coordinate
}

// Geom_NewCoordinateXYZM4D constructs a Geom_CoordinateXYZM at (0,0,NaN) with M=0.
func Geom_NewCoordinateXYZM4D() *Geom_CoordinateXYZM {
	coord := &Geom_Coordinate{
		X: 0,
		Y: 0,
		Z: math.NaN(),
	}
	c := &Geom_CoordinateXYZM{
		Geom_Coordinate: coord,
		M:               0.0,
	}
	coord.child = c
	return c
}

// Geom_NewCoordinateXYZM4DWithXYZM constructs a Geom_CoordinateXYZM instance with the
// given ordinates and measure.
func Geom_NewCoordinateXYZM4DWithXYZM(x, y, z, m float64) *Geom_CoordinateXYZM {
	coord := &Geom_Coordinate{
		X: x,
		Y: y,
		Z: z,
	}
	c := &Geom_CoordinateXYZM{
		Geom_Coordinate: coord,
		M:               m,
	}
	coord.child = c
	return c
}

// Geom_NewCoordinateXYZM4DFromCoordinate constructs a Geom_CoordinateXYZM instance with
// the ordinates of the given Geom_Coordinate.
func Geom_NewCoordinateXYZM4DFromCoordinate(other *Geom_Coordinate) *Geom_CoordinateXYZM {
	coord := &Geom_Coordinate{
		X: other.X,
		Y: other.Y,
		Z: other.GetZ(),
	}
	c := &Geom_CoordinateXYZM{
		Geom_Coordinate: coord,
		M:               other.GetM(),
	}
	coord.child = c
	return c
}

// Geom_NewCoordinateXYZM4DFromCoordinateXYZM constructs a Geom_CoordinateXYZM instance
// with the ordinates of the given Geom_CoordinateXYZM.
func Geom_NewCoordinateXYZM4DFromCoordinateXYZM(other *Geom_CoordinateXYZM) *Geom_CoordinateXYZM {
	coord := &Geom_Coordinate{
		X: other.X,
		Y: other.Y,
		Z: other.Z,
	}
	c := &Geom_CoordinateXYZM{
		Geom_Coordinate: coord,
		M:               other.M,
	}
	coord.child = c
	return c
}

func (c *Geom_CoordinateXYZM) Copy_BODY() *Geom_Coordinate {
	return Geom_NewCoordinateXYZM4DFromCoordinateXYZM(c).Geom_Coordinate
}

func (c *Geom_CoordinateXYZM) Create_BODY() *Geom_Coordinate {
	return Geom_NewCoordinateXYZM4D().Geom_Coordinate
}

func (c *Geom_CoordinateXYZM) GetM_BODY() float64 {
	return c.M
}

func (c *Geom_CoordinateXYZM) SetM_BODY(m float64) {
	c.M = m
}

func (c *Geom_CoordinateXYZM) SetCoordinate_BODY(other *Geom_Coordinate) {
	c.X = other.X
	c.Y = other.Y
	c.Z = other.GetZ()
	c.M = other.GetM()
}

func (c *Geom_CoordinateXYZM) GetOrdinate_BODY(ordinateIndex int) float64 {
	switch ordinateIndex {
	case Geom_Coordinate_X:
		return c.X
	case Geom_Coordinate_Y:
		return c.Y
	case Geom_Coordinate_Z:
		return c.GetZ()
	case Geom_Coordinate_M:
		return c.GetM()
	}
	panic(fmt.Sprintf("Invalid ordinate index: %d", ordinateIndex))
}

func (c *Geom_CoordinateXYZM) SetOrdinate_BODY(ordinateIndex int, value float64) {
	switch ordinateIndex {
	case Geom_Coordinate_X:
		c.X = value
	case Geom_Coordinate_Y:
		c.Y = value
	case Geom_Coordinate_Z:
		c.Z = value
	case Geom_Coordinate_M:
		c.M = value
	default:
		panic(fmt.Sprintf("Invalid ordinate index: %d", ordinateIndex))
	}
}

func (c *Geom_CoordinateXYZM) String_BODY() string {
	return fmt.Sprintf("(%v, %v, %v m=%v)", c.X, c.Y, c.GetZ(), c.GetM())
}
