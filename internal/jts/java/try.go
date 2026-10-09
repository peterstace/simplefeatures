package java

import (
	"runtime"
	"strings"
)

// Try implements a Java try statement. body, catch, and finally return a flow
// code: 0 for normal completion, otherwise the code of a return, break, or
// continue that leaves the statement. catch reports whether it handled the
// exception. A non-zero flow from finally overrides the outcome of the body
// and catch, including a pending exception, as in Java.
func Try(body func() int, catch func(ex any) (int, bool), finally func() int) (flow int) {
	if finally != nil {
		defer func() {
			r := recover()
			if f := finally(); f != 0 {
				flow = f
				return
			}
			if r != nil {
				panic(r)
			}
		}()
	}
	if catch == nil {
		return body()
	}
	ex, f := runCatching(body)
	if ex == nil {
		return f
	}
	if f, ok := catch(ex); ok {
		return f
	}
	panic(ex)
}

// runCatching runs body, returning the Java exception it throws, if any.
func runCatching(body func() int) (ex any, flow int) {
	defer func() {
		if r := recover(); r != nil {
			ex = Throwable_ofPanic(r)
		}
	}()
	return nil, body()
}

// Throwable_ofPanic converts a recovered panic value to a Java exception. A
// Go runtime error becomes the exception the JVM would throw in its place.
// Panics that are not Java exceptions (such as untranslated code) re-panic.
func Throwable_ofPanic(r any) any {
	if Throwable_isInstance(r) {
		return r
	}
	if err, ok := r.(runtime.Error); ok {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "nil pointer dereference"):
			return NullPointerException_new()
		case strings.Contains(msg, "interface conversion: ") && strings.Contains(msg, " is nil, not "):
			return NullPointerException_new()
		case strings.Contains(msg, "interface conversion"):
			return ClassCastException_new_String(FromGo(msg))
		case strings.Contains(msg, "index out of range"), strings.Contains(msg, "slice bounds out of range"):
			return ArrayIndexOutOfBoundsException_new_String(FromGo(msg))
		case strings.Contains(msg, "integer divide by zero"):
			return ArithmeticException_new_String(Lit("/ by zero"))
		case strings.Contains(msg, "makeslice: len out of range"):
			return NegativeArraySizeException_new()
		}
	}
	panic(r)
}
