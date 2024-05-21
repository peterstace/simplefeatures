package geom

import "fmt"

func wrap(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// wrapSimplified wraps errors to indicate that they occurred as a result of no
// longer being valid after simplification.
func wrapSimplified(err error) error {
	return wrap(err, "simplified geometry")
}

// wkbSyntaxError is an error used to indicate that a serialised WKB geometry
// cannot be unmarshalled because some aspect of it's syntax is invalid.
type wkbSyntaxError struct {
	// reason should describe the invalid syntax (as opposed to describing the
	// syntax rule that was broken).
	reason string
}

func (e wkbSyntaxError) Error() string {
	return "invalid WKB syntax: " + e.reason
}

// wktSyntaxError is an error used to indicate that a serialised WKT geometry
// cannot be unmarshalled because some aspect of it's syntax is invalid.
type wktSyntaxError struct {
	// reason should describe the invalid syntax (as opposed to describing the
	// syntax rule that was broken).
	reason string
}

func (e wktSyntaxError) Error() string {
	return "invalid WKT syntax: " + e.reason
}

// geojsonSyntaxError is an error used to indicate that a serialised GeoJSON geometry
// cannot be unmarshalled because some aspect of it's syntax is invalid.
type geojsonSyntaxError struct {
	// reason should describe the invalid syntax (as opposed to describing the
	// syntax rule that was broken).
	reason string
}

func (e geojsonSyntaxError) Error() string {
	return "invalid GeoJSON syntax: " + e.reason
}

func wrapWithGeoJSONSyntaxError(err error) error {
	if err == nil {
		return nil
	}
	return geojsonSyntaxError{err.Error()}
}

type unmarshalGeoJSONSourceDestinationMismatchError struct {
	SourceType      GeometryType
	DestinationType GeometryType
}

func (e unmarshalGeoJSONSourceDestinationMismatchError) Error() string {
	return fmt.Sprintf(
		"cannot unmarshal GeoJSON of type %s into %s",
		e.SourceType, e.DestinationType,
	)
}

type mismatchedGeometryCollectionDimsError struct {
	CT1, CT2 CoordinatesType
}

func (e mismatchedGeometryCollectionDimsError) Error() string {
	return fmt.Sprintf("mixed dimensions in geometry collection: %s and %s", e.CT1, e.CT2)
}

type ruleViolation string

const (
	violateInf                ruleViolation = "Inf not allowed"
	violateNaN                ruleViolation = "NaN not allowed"
	violateTwoPoints          ruleViolation = "non-empty LineString contains only one distinct XY value"
	violateRingEmpty          ruleViolation = "polygon ring empty"
	violateRingClosed         ruleViolation = "polygon ring not closed"
	violateRingSimple         ruleViolation = "polygon ring not simple"
	violateRingNested         ruleViolation = "polygon has nested rings"
	violateInteriorInExterior ruleViolation = "polygon interior ring outside of exterior ring"
	violateInteriorConnected  ruleViolation = "polygon has disconnected interior"
	violateRingsMultiTouch    ruleViolation = "polygon rings intersect at multiple points"
	violatePolysMultiTouch    ruleViolation = "multipolygon child polygons touch at multiple points"
)

func (v ruleViolation) errAtXY(location XY) error {
	return validationError{
		RuleViolation: v,
		HasLocation:   true,
		Location:      location,
	}
}

func (v ruleViolation) errAtPt(location Point) error {
	xy, ok := location.XY()
	return validationError{
		RuleViolation: v,
		HasLocation:   ok,
		Location:      xy,
	}
}

func (v ruleViolation) err() error {
	return validationError{RuleViolation: v}
}

type validationError struct {
	RuleViolation ruleViolation
	HasLocation   bool
	Location      XY
}

func (e validationError) Error() string {
	if e.HasLocation {
		return fmt.Sprintf("%s (at or near %v)", e.RuleViolation, e.Location)
	}
	return string(e.RuleViolation)
}
