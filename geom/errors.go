package geom

import "fmt"

type mixedCoordinatesTypeError struct {
	first  CoordinatesType
	second CoordinatesType
}

func (e mixedCoordinatesTypeError) Error() string {
	return fmt.Sprintf("mixed coordinate types not "+
		"allowed: %s and %s", e.first, e.second)
}

type geojsonInvalidCoordinatesLengthError struct {
	length int
}

func (e geojsonInvalidCoordinatesLengthError) Error() string {
	return fmt.Sprintf(
		"invalid coordinates length in geojson: %d",
		e.length,
	)
}
