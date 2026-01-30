package jts

// JtstestGeomop_GeometryOperation is an interface for classes which execute
// operations on Geometries.
type JtstestGeomop_GeometryOperation interface {
	IsJtstestGeomop_GeometryOperation()

	// GetReturnType gets the class of the return type of the given operation.
	// Returns "boolean", "int", "double", "geometry", or "" if unknown.
	GetReturnType(opName string) string

	// Invoke invokes an operation on a Geometry.
	Invoke(opName string, geometry *Geom_Geometry, args []any) (JtstestTestrunner_Result, error)
}
