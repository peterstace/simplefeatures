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
	return fmt.Sprintf("invalid WKB syntax: %s", e.reason)
}

// wktSyntaxError is an error used to indicate that a serialised WKT geometry
// cannot be unmarshalled because some aspect of it's syntax is invalid.
type wktSyntaxError struct {
	// reason should describe the invalid syntax (as opposed to describing the
	// syntax rule that was broken).
	reason string
}

func (e wktSyntaxError) Error() string {
	return fmt.Sprintf("invalid WKT syntax: %s", e.reason)
}

// geojsonSyntaxError is an error used to indicate that a serialised GeoJSON geometry
// cannot be unmarshalled because some aspect of it's syntax is invalid.
type geojsonSyntaxError struct {
	// reason should describe the invalid syntax (as opposed to describing the
	// syntax rule that was broken).
	reason string
}

func (e geojsonSyntaxError) Error() string {
	return fmt.Sprintf("invalid GeoJSON syntax: %s", e.reason)
}

func wrapWithGeoJSONSyntaxError(err error) error {
	if err == nil {
		return nil
	}
	return geojsonSyntaxError{err.Error()}
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
		ruleViolation: v,
		hasLocation:   true,
		location:      location,
	}
}

func (v ruleViolation) errAtPt(location Point) error {
	xy, ok := location.XY()
	return validationError{
		ruleViolation: v,
		hasLocation:   ok,
		location:      xy,
	}
}

func (v ruleViolation) err() error {
	return validationError{ruleViolation: v}
}

type validationError struct {
	ruleViolation ruleViolation
	hasLocation   bool
	location      XY
}

func (e validationError) Error() string {
	if e.hasLocation {
		return fmt.Sprintf("%s (at or near %v)", e.ruleViolation, e.location)
	}
	return string(e.ruleViolation)
}
