package geom

// The following types are exported for testing purposes only.
//
// Because this file ends in `_test.go`, it is not included in non-test builds.
// But because it's in the `geom` package, it's able to access unexported types
// from other files in the `geom` package.

type MismatchedGeometryCollectionDimsError = mismatchedGeometryCollectionDimsError

type UnmarshalGeoJSONSourceDestinationMismatchError = unmarshalGeoJSONSourceDestinationMismatchError

type (
	ValidationError = validationError
	RuleViolation   = ruleViolation
)

const (
	ViolateInf                = violateInf
	ViolateNaN                = violateNaN
	ViolateTwoPoints          = violateTwoPoints
	ViolateRingEmpty          = violateRingEmpty
	ViolateRingClosed         = violateRingClosed
	ViolateRingSimple         = violateRingSimple
	ViolateRingNested         = violateRingNested
	ViolateInteriorInExterior = violateInteriorInExterior
	ViolateInteriorConnected  = violateInteriorConnected
	ViolateRingsMultiTouch    = violateRingsMultiTouch
	ViolatePolysMultiTouch    = violatePolysMultiTouch
)
