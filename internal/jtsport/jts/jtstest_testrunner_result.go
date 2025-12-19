package jts

// JtstestTestrunner_Result is the interface for test results.
type JtstestTestrunner_Result interface {
	IsJtstestTestrunner_Result()
	EqualsResult(other JtstestTestrunner_Result, tolerance float64) bool
	ToLongString() string
	ToFormattedString() string
	ToShortString() string
}
