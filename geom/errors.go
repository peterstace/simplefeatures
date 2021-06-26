package geom

import "fmt"

func wrap(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %v", append(args, err)...)
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
	// gtype is the type of geometry that was attempted to be created.
	gtype GeometryType

	// reason should describe the invalid state (as opposed to describing the
	// validation rule). E.g. "non-closed ring" rather than "rings must be
	// closed".
	reason string
}

func (e validationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.gtype, e.reason)
}

// childValidationError is an error is used in 2 cases:
//
// 1. To indicate that multi-geometry (MultiPoint, MultiLineString,
// MultiPolygon, or GeometryCollection) could not be created because one of its
// children did not pass all validation checks.
//
// 2. To indicate that Polygon could not be created because of its rings is not
// a valid LineString.
type childValidationError struct {
	gtype    GeometryType // type of multi-geometry or Polygon that was attempted to be created
	childIdx int          // 0-based index of the child
	childErr error        // error returned when attempting to validate the child
}

func (e childValidationError) Error() string {
	var subject string
	if e.gtype == TypePolygon {
		subject = ringName(e.childIdx)
	} else {
		subject = fmt.Sprintf("child with index %d", e.childIdx)
	}
	return fmt.Sprintf("invalid %s, %s is invalid: %s", e.gtype, subject, e.childErr.Error())
}
