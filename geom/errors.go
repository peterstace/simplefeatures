package geom

import "fmt"

type geojsonInvalidCoordinatesLengthError struct {
	length int
}

func (e geojsonInvalidCoordinatesLengthError) Error() string {
	return fmt.Sprintf(
		"invalid coordinates length in geojson: %d",
		e.length,
	)
}
