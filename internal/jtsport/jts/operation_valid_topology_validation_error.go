package jts

// OperationValid_TopologyValidationError_ERROR is not used.
// Deprecated.
const OperationValid_TopologyValidationError_ERROR = 0

// OperationValid_TopologyValidationError_REPEATED_POINT is no longer used -
// repeated points are considered valid as per the SFS.
// Deprecated.
const OperationValid_TopologyValidationError_REPEATED_POINT = 1

// OperationValid_TopologyValidationError_HOLE_OUTSIDE_SHELL indicates that a
// hole of a polygon lies partially or completely in the exterior of the shell.
const OperationValid_TopologyValidationError_HOLE_OUTSIDE_SHELL = 2

// OperationValid_TopologyValidationError_NESTED_HOLES indicates that a hole
// lies in the interior of another hole in the same polygon.
const OperationValid_TopologyValidationError_NESTED_HOLES = 3

// OperationValid_TopologyValidationError_DISCONNECTED_INTERIOR indicates that
// the interior of a polygon is disjoint (often caused by set of contiguous
// holes splitting the polygon into two parts).
const OperationValid_TopologyValidationError_DISCONNECTED_INTERIOR = 4

// OperationValid_TopologyValidationError_SELF_INTERSECTION indicates that two
// rings of a polygonal geometry intersect.
const OperationValid_TopologyValidationError_SELF_INTERSECTION = 5

// OperationValid_TopologyValidationError_RING_SELF_INTERSECTION indicates that
// a ring self-intersects.
const OperationValid_TopologyValidationError_RING_SELF_INTERSECTION = 6

// OperationValid_TopologyValidationError_NESTED_SHELLS indicates that a polygon
// component of a MultiPolygon lies inside another polygonal component.
const OperationValid_TopologyValidationError_NESTED_SHELLS = 7

// OperationValid_TopologyValidationError_DUPLICATE_RINGS indicates that a
// polygonal geometry contains two rings which are identical.
const OperationValid_TopologyValidationError_DUPLICATE_RINGS = 8

// OperationValid_TopologyValidationError_TOO_FEW_POINTS indicates that either
// a LineString contains a single point or a LinearRing contains 2 or 3 points.
const OperationValid_TopologyValidationError_TOO_FEW_POINTS = 9

// OperationValid_TopologyValidationError_INVALID_COORDINATE indicates that the
// X or Y ordinate of a Coordinate is not a valid numeric value (e.g. NaN).
const OperationValid_TopologyValidationError_INVALID_COORDINATE = 10

// OperationValid_TopologyValidationError_RING_NOT_CLOSED indicates that a ring
// is not correctly closed (the first and the last coordinate are different).
const OperationValid_TopologyValidationError_RING_NOT_CLOSED = 11

// operationValid_TopologyValidationError_errMsg contains messages corresponding
// to error codes.
var operationValid_TopologyValidationError_errMsg = []string{
	"Topology Validation Error",
	"Repeated Point",
	"Hole lies outside shell",
	"Holes are nested",
	"Interior is disconnected",
	"Self-intersection",
	"Ring Self-intersection",
	"Nested shells",
	"Duplicate Rings",
	"Too few distinct points in geometry component",
	"Invalid Coordinate",
	"Ring is not closed",
}

// OperationValid_TopologyValidationError contains information about the nature
// and location of a Geometry validation error.
type OperationValid_TopologyValidationError struct {
	errorType int
	pt        *Geom_Coordinate
}

// OperationValid_NewTopologyValidationError creates a validation error with the
// given type and location.
func OperationValid_NewTopologyValidationError(errorType int, pt *Geom_Coordinate) *OperationValid_TopologyValidationError {
	var ptCopy *Geom_Coordinate
	if pt != nil {
		ptCopy = pt.Copy()
	}
	return &OperationValid_TopologyValidationError{
		errorType: errorType,
		pt:        ptCopy,
	}
}

// OperationValid_NewTopologyValidationErrorWithType creates a validation error
// of the given type with a null location.
func OperationValid_NewTopologyValidationErrorWithType(errorType int) *OperationValid_TopologyValidationError {
	return OperationValid_NewTopologyValidationError(errorType, nil)
}

// GetCoordinate returns the location of this error (on the Geometry containing
// the error).
func (e *OperationValid_TopologyValidationError) GetCoordinate() *Geom_Coordinate {
	return e.pt
}

// GetErrorType gets the type of this error.
func (e *OperationValid_TopologyValidationError) GetErrorType() int {
	return e.errorType
}

// GetMessage gets an error message describing this error. The error message
// does not describe the location of the error.
func (e *OperationValid_TopologyValidationError) GetMessage() string {
	return operationValid_TopologyValidationError_errMsg[e.errorType]
}

// String gets a message describing the type and location of this error.
func (e *OperationValid_TopologyValidationError) String() string {
	locStr := ""
	if e.pt != nil {
		locStr = " at or near point " + e.pt.String()
	}
	return e.GetMessage() + locStr
}
