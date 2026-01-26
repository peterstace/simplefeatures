package jts

import "fmt"

// Util_Assert_IsTrue throws an Util_AssertionFailedException if the given assertion is not true.
func Util_Assert_IsTrue(assertion bool) {
	Util_Assert_IsTrueWithMessage(assertion, "")
}

// Util_Assert_IsTrueWithMessage throws an Util_AssertionFailedException with the given message
// if the given assertion is not true.
func Util_Assert_IsTrueWithMessage(assertion bool, message string) {
	if !assertion {
		if message == "" {
			panic(Util_NewAssertionFailedException())
		}
		panic(Util_NewAssertionFailedExceptionWithMessage(message))
	}
}

// Util_Assert_Equals throws an Util_AssertionFailedException if the given objects are not equal.
func Util_Assert_Equals(expectedValue, actualValue any) {
	Util_Assert_EqualsWithMessage(expectedValue, actualValue, "")
}

// Util_Assert_EqualsWithMessage throws an Util_AssertionFailedException with the given message
// if the given objects are not equal.
func Util_Assert_EqualsWithMessage(expectedValue, actualValue any, message string) {
	// TRANSLITERATION NOTE: Java uses actualValue.equals(expectedValue) which
	// dispatches polymorphically. Go doesn't have this, so we inline special
	// handling for Coordinate (which uses Equals2D in its equals method).
	equal := false
	if expectedCoord, ok := expectedValue.(*Geom_Coordinate); ok {
		if actualCoord, ok := actualValue.(*Geom_Coordinate); ok {
			equal = expectedCoord.Equals2D(actualCoord)
		}
	} else {
		equal = expectedValue == actualValue
	}
	if !equal {
		msg := fmt.Sprintf("Expected %v but encountered %v", expectedValue, actualValue)
		if message != "" {
			msg += ": " + message
		}
		panic(Util_NewAssertionFailedExceptionWithMessage(msg))
	}
}

// Util_Assert_ShouldNeverReachHere always throws an Util_AssertionFailedException.
func Util_Assert_ShouldNeverReachHere() {
	Util_Assert_ShouldNeverReachHereWithMessage("")
}

// Util_Assert_ShouldNeverReachHereWithMessage always throws an Util_AssertionFailedException
// with the given message.
func Util_Assert_ShouldNeverReachHereWithMessage(message string) {
	msg := "Should never reach here"
	if message != "" {
		msg += ": " + message
	}
	panic(Util_NewAssertionFailedExceptionWithMessage(msg))
}
