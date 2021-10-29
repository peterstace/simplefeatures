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
