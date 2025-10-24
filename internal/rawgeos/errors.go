package rawgeos

import "fmt"

// #include "geos_c.h"
import "C"

func wrap(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

var currentGEOSVersion = fmt.Sprintf(
	"%d.%d.%d",
	C.GEOS_VERSION_MAJOR,
	C.GEOS_VERSION_MINOR,
	C.GEOS_VERSION_PATCH,
)

type UnsupportedGEOSVersionError struct {
	requiredGEOSVersion string
	operation           string
}

func (e UnsupportedGEOSVersionError) Error() string {
	return fmt.Sprintf("%s is unsupported in GEOS %s, requires at least GEOS %s",
		e.operation, currentGEOSVersion, e.requiredGEOSVersion)
}
