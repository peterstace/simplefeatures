package geos

import (
	"C"
	"errors"
	"fmt"
)

var (
	errGeometryCollectionNotSupported = errors.New("GeometryCollection not supported")
	errBadMalloc                      = fmt.Errorf("malloc failed")
	errGEOSContextNotCreated          = fmt.Errorf("could not create GEOS context")
)

func errGEOSInternalFailure(msg string) error {
	return fmt.Errorf("GEOS internal error: %v", msg)
}

func errGEOSIllegalReturnValue(c C.char) error {
	return fmt.Errorf("illegal result from GEOS: %v", c)
}

func errInvalidMaskLength(mask string) error {
	return fmt.Errorf("mask has invalid length %d (must be 9)", len(mask))
}
