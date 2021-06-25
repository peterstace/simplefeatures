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
