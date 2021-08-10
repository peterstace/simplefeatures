package geom

import (
	"strconv"
	"strings"
)

// OptionalCoordinates represent a point location that may be empty.
type OptionalCoordinates struct {
	// Type indicates the coordinates type, and therefore whether or not Z and
	// M are populated. Type must be populated even when the coordinates are
	// empty (i.e. empty coordinates have a well defined type).
	Type CoordinatesType

	// Empty indicates if the coordinates are empty or not.
	Empty bool

	// XY represents the XY position of the point location. It's ignored for
	// empty coordinates.
	XY

	// Z represents the height of the location. It's ignored for empty or
	// non-3D coordinate types.
	Z float64

	// M represents a user defined measure associated with the location. It's
	// ignored for empty or non-measure coordinate types.
	M float64
}

// String gives a string representation of the optional coordinates.
func (c OptionalCoordinates) String() string {
	var sb strings.Builder
	sb.WriteString("OptionalCoordinates[")
	sb.WriteString(c.Type.String())
	sb.WriteString("] ")
	if c.Empty {
		sb.WriteString("EMPTY")
		return sb.String()
	}
	sb.WriteString(strconv.FormatFloat(c.X, 'f', -1, 64))
	sb.WriteRune(' ')
	sb.WriteString(strconv.FormatFloat(c.Y, 'f', -1, 64))
	if c.Type.Is3D() {
		sb.WriteRune(' ')
		sb.WriteString(strconv.FormatFloat(c.Z, 'f', -1, 64))
	}
	if c.Type.IsMeasured() {
		sb.WriteRune(' ')
		sb.WriteString(strconv.FormatFloat(c.M, 'f', -1, 64))
	}
	return sb.String()
}
