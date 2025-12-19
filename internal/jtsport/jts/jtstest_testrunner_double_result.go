package jts

import (
	"math"
	"strconv"
)

var _ JtstestTestrunner_Result = (*JtstestTestrunner_DoubleResult)(nil)

type JtstestTestrunner_DoubleResult struct {
	value float64
}

func JtstestTestrunner_NewDoubleResult(value float64) *JtstestTestrunner_DoubleResult {
	return &JtstestTestrunner_DoubleResult{value: value}
}

func (r *JtstestTestrunner_DoubleResult) IsJtstestTestrunner_Result() {}

func (r *JtstestTestrunner_DoubleResult) EqualsResult(other JtstestTestrunner_Result, tolerance float64) bool {
	otherResult, ok := other.(*JtstestTestrunner_DoubleResult)
	if !ok {
		return false
	}
	otherValue := otherResult.value
	return math.Abs(r.value-otherValue) <= tolerance
}

func (r *JtstestTestrunner_DoubleResult) ToLongString() string {
	return strconv.FormatFloat(r.value, 'f', -1, 64)
}

func (r *JtstestTestrunner_DoubleResult) ToFormattedString() string {
	return strconv.FormatFloat(r.value, 'f', -1, 64)
}

func (r *JtstestTestrunner_DoubleResult) ToShortString() string {
	return strconv.FormatFloat(r.value, 'f', -1, 64)
}
