package jts

import "strconv"

var _ JtstestTestrunner_Result = (*JtstestTestrunner_IntegerResult)(nil)

type JtstestTestrunner_IntegerResult struct {
	value int
}

func JtstestTestrunner_NewIntegerResult(value int) *JtstestTestrunner_IntegerResult {
	return &JtstestTestrunner_IntegerResult{value: value}
}

func (r *JtstestTestrunner_IntegerResult) IsJtstestTestrunner_Result() {}

func (r *JtstestTestrunner_IntegerResult) EqualsResult(other JtstestTestrunner_Result, tolerance float64) bool {
	otherResult, ok := other.(*JtstestTestrunner_IntegerResult)
	if !ok {
		return false
	}
	otherValue := otherResult.value
	diff := r.value - otherValue
	if diff < 0 {
		diff = -diff
	}
	return float64(diff) <= tolerance
}

func (r *JtstestTestrunner_IntegerResult) ToLongString() string {
	return strconv.Itoa(r.value)
}

func (r *JtstestTestrunner_IntegerResult) ToFormattedString() string {
	return strconv.Itoa(r.value)
}

func (r *JtstestTestrunner_IntegerResult) ToShortString() string {
	return strconv.Itoa(r.value)
}
