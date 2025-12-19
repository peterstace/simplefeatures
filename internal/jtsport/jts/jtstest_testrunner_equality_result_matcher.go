package jts

var _ JtstestTestrunner_ResultMatcher = (*JtstestTestrunner_EqualityResultMatcher)(nil)

// JtstestTestrunner_EqualityResultMatcher compares results for equality,
// up to the given tolerance.
type JtstestTestrunner_EqualityResultMatcher struct{}

func JtstestTestrunner_NewEqualityResultMatcher() *JtstestTestrunner_EqualityResultMatcher {
	return &JtstestTestrunner_EqualityResultMatcher{}
}

func (m *JtstestTestrunner_EqualityResultMatcher) IsJtstestTestrunner_ResultMatcher() {}

func (m *JtstestTestrunner_EqualityResultMatcher) IsMatch(
	geom *Geom_Geometry,
	opName string,
	args []any,
	actualResult JtstestTestrunner_Result,
	expectedResult JtstestTestrunner_Result,
	tolerance float64,
) bool {
	return actualResult.EqualsResult(expectedResult, tolerance)
}
