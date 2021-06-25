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
// created because it would not pass all validation checks.
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
