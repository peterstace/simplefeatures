package jts

// JtstestTestrunner_ResultMatcher is an interface for classes which can
// determine whether two Results match, within a given tolerance.
type JtstestTestrunner_ResultMatcher interface {
	IsJtstestTestrunner_ResultMatcher()
	IsMatch(
		geom *Geom_Geometry,
		opName string,
		args []any,
		actualResult JtstestTestrunner_Result,
		expectedResult JtstestTestrunner_Result,
		tolerance float64,
	) bool
}
