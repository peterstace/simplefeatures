package geom

import "fmt"

// CoordinatesType controls the dimensionality and type of data used to encode
// a point location.  At minimum, a point location is defined by X and Y
// coordinates. It may optionally include a Z value, representing height. It
// may also optionally include an M value, traditionally representing an
// arbitrary user defined measurement associated with each point location.
type CoordinatesType byte

const (
	// DimXY coordinates only contain X and Y values.
	DimXY CoordinatesType = 0b00

	// DimXYZ coordinates contain X, Y, and Z (height) values.
	DimXYZ CoordinatesType = 0b01

	// DimXYM coordinates contain X, Y, and M (measure) values.
	DimXYM CoordinatesType = 0b10

	// DimXYZM coordinates contain X, Y, Z (height), and M (measure) values.
	DimXYZM CoordinatesType = 0b11
)

// String gives a string representation of a CoordinatesType.
func (t CoordinatesType) String() string {
	if t < 4 {
		return [4]string{"XY", "XYZ", "XYM", "XYZM"}[t]
	}
	return fmt.Sprintf("unknown coordinate type (%d)", t)
}

// Dimension returns the number of float64 coordinates required to encode a
// point location using the CoordinatesType.
func (t CoordinatesType) Dimension() int {
	return [4]int{2, 3, 3, 4}[t]
}

// Is3D returns true if and only if the CoordinatesType includes a Z (3D)
// value.
func (t CoordinatesType) Is3D() bool {
	return (t & DimXYZ) != 0
}

// IsMeasured returns true if and only if the Coordinates type includes an M
// (measure) value.
func (t CoordinatesType) IsMeasured() bool {
	return (t & DimXYM) != 0
}
