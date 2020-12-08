package geom

import "fmt"

// SyntaxError indicates an error in the structural representation of a
// serialised geometry.
type SyntaxError struct {
	reason string
}

// Error gives the error text of the syntax error.
func (e SyntaxError) Error() string {
	return e.reason
}

//type TopologyError struct {
//}

type geojsonInvalidCoordinatesLengthError struct {
	length int
}

func (e geojsonInvalidCoordinatesLengthError) Error() string {
	return fmt.Sprintf(
		"invalid coordinates length in geojson: %d",
		e.length,
	)
}
