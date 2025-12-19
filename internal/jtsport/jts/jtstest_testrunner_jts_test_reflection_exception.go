package jts

import (
	"fmt"
	"reflect"
)

// JtstestTestrunner_JTSTestReflectionException indicates a problem during
// reflection.
type JtstestTestrunner_JTSTestReflectionException struct {
	message string
}

func JtstestTestrunner_NewJTSTestReflectionExceptionWithMessage(message string) *JtstestTestrunner_JTSTestReflectionException {
	return &JtstestTestrunner_JTSTestReflectionException{message: message}
}

func JtstestTestrunner_NewJTSTestReflectionException(opName string, args []any) *JtstestTestrunner_JTSTestReflectionException {
	return &JtstestTestrunner_JTSTestReflectionException{
		message: jtstestTestrunner_JTSTestReflectionException_createMessage(opName, args),
	}
}

func jtstestTestrunner_JTSTestReflectionException_createMessage(opName string, args []any) string {
	msg := "Could not find Geometry method: " + opName + "("
	for j := 0; j < len(args); j++ {
		if j > 0 {
			msg += ", "
		}
		msg += reflect.TypeOf(args[j]).String()
	}
	msg += ")"
	return msg
}

func (e *JtstestTestrunner_JTSTestReflectionException) Error() string {
	return fmt.Sprintf("JTSTestReflectionException: %s", e.message)
}
