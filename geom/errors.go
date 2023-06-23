package geom

import "fmt"

func wrap(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// wrapTransformed wraps errors to indicate that they occurred as the result of
// pointwise-transforming a geometry.
func wrapTransformed(err error) error {
	return wrap(err, "transformed geometry")
}

// wrapSimplified wraps errors to indicate that they occurred as a result of no
// longer being valid after simplification.
func wrapSimplified(err error) error {
	return wrap(err, "simplified geometry")
}

// validationError is an error used to indicate that a geometry could not be
// created because it didn't pass all validation checks.
// TODO: remove me
type validationError struct {
	// reason should begin with the name of the invalid geometry being created,
	// and describe the invalid state (as opposed to describing the validation
	// rule). E.g. "polygon has non-closed ring" rather than "polygon rings
	// must be closed".
	reason string
}

func (e validationError) Error() string {
	return fmt.Sprintf("failed geometry constraint: %s", e.reason)
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
	defyNoInf              ruleViolation = "Inf not allowed"
	defyNoNaN              ruleViolation = "NaN not allowed"
	defyAtLeastTwoPoints   ruleViolation = "non-empty LineString contains only one distinct XY value"
	defyRingNotEmpty       ruleViolation = "polygon ring empty"
	defyRingClosed         ruleViolation = "polygon ring not closed"
	defyRingSimple         ruleViolation = "polygon ring not simple"
	defyRingNotNested      ruleViolation = "polygon has nested rings"
	defyInteriorInExterior ruleViolation = "polygon interior ring outside of exterior ring"
	defyInteriorConnected  ruleViolation = "polygon has disconnected interior"
)

func (v ruleViolation) errAt(location XY) error {
	return validationError2{
		ruleViolation: v,
		hasLocation:   true,
		location:      location,
	}
}

func (v ruleViolation) err() error {
	return validationError2{ruleViolation: v}
}

// TODO: rename to validationError once the old validationError is removed.
type validationError2 struct {
	ruleViolation ruleViolation
	hasLocation   bool
	location      XY
}

func (e validationError2) Error() string {
	if e.hasLocation {
		return fmt.Sprintf("%s (at or near %v)", e.ruleViolation, e.location)
	}
	return string(e.ruleViolation)
}
