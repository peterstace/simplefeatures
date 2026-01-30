package jts

// Util_AssertionFailedException is thrown when the application is in an inconsistent
// state. Indicates a problem with the code.
type Util_AssertionFailedException struct {
	message string
}

// Util_NewAssertionFailedException creates an Util_AssertionFailedException.
func Util_NewAssertionFailedException() *Util_AssertionFailedException {
	return &Util_AssertionFailedException{}
}

// Util_NewAssertionFailedExceptionWithMessage creates an Util_AssertionFailedException
// with the given detail message.
func Util_NewAssertionFailedExceptionWithMessage(message string) *Util_AssertionFailedException {
	return &Util_AssertionFailedException{message: message}
}

// TRANSLITERATION NOTE: Error() is not in Java. It implements Go's error
// interface, analogous to getMessage()/toString() inherited from RuntimeException.
func (e *Util_AssertionFailedException) Error() string {
	if e.message == "" {
		return "Util_AssertionFailedException"
	}
	return e.message
}
