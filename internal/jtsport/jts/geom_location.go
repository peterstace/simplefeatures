package jts

import "fmt"

// Geom_Location_Interior is the location value for the interior of a geometry. Also,
// DE-9IM row index of the interior of the first geometry and column index
// of the interior of the second geometry.
const Geom_Location_Interior = 0

// Geom_Location_Boundary is the location value for the boundary of a geometry. Also,
// DE-9IM row index of the boundary of the first geometry and column index
// of the boundary of the second geometry.
const Geom_Location_Boundary = 1

// Geom_Location_Exterior is the location value for the exterior of a geometry. Also,
// DE-9IM row index of the exterior of the first geometry and column index
// of the exterior of the second geometry.
const Geom_Location_Exterior = 2

// Geom_Location_None is used for uninitialized location values.
const Geom_Location_None = -1

// Geom_Location_ToLocationSymbol converts the location value to a location symbol, for
// example, Exterior => 'e'.
//
// locationValue is either Exterior, Boundary, Interior or None.
//
// Returns either 'e', 'b', 'i' or '-'.
func Geom_Location_ToLocationSymbol(locationValue int) byte {
	switch locationValue {
	case Geom_Location_Exterior:
		return 'e'
	case Geom_Location_Boundary:
		return 'b'
	case Geom_Location_Interior:
		return 'i'
	case Geom_Location_None:
		return '-'
	default:
		panic(fmt.Sprintf("unknown location value: %d", locationValue))
	}
}
