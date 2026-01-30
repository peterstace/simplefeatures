package jts

// JtstestTestrunner_TestParseException represents an error during test parsing.
type JtstestTestrunner_TestParseException struct {
	message string
}

func JtstestTestrunner_NewTestParseException(message string) *JtstestTestrunner_TestParseException {
	return &JtstestTestrunner_TestParseException{message: message}
}

func (e *JtstestTestrunner_TestParseException) Error() string {
	return e.message
}
