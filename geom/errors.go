package geom

import "fmt"

type MixedCoordinatesTypesError struct {
	First  CoordinatesType
	Second CoordinatesType
}

func (e MixedCoordinatesTypesError) Error() string {
	return fmt.Sprintf("mixed coordinate types not "+
		"allowed: %s and %s", e.First, e.Second)
}

type GeoJSONInvalidCoordinatesLengthError struct {
	length int
}

func (e GeoJSONInvalidCoordinatesLengthError) Error() string {
	return fmt.Sprintf(
		"invalid coordinates length in geojson: %d",
		e.length,
	)
}

var GeoJSONNaNOrInfErr = fmt.Errorf("coordinate is NaN or inf")

type ValidationError struct {
	reason string
}

func (e ValidationError) Error() string {
	return e.reason
}
