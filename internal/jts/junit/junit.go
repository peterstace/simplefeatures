// Package junit implements the parts of JUnit that translated JTS tests use.
package junit

import (
	"fmt"
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jts/java"
)

// Run runs a translated test, reporting a thrown AssertionFailedError or any
// other panic as a test failure.
func Run(t *testing.T, test func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if u, ok := r.(java.UnsupportedError); ok {
				t.Skip(u.Error())
			}
			if java.Throwable_isInstance(r) {
				t.Fatalf("%s", java.Object_ToString(r).Go())
			}
			t.Fatalf("panic: %v", r)
		}
	}()
	test()
}

// AssertionFailedError is junit.framework.AssertionFailedError.
type AssertionFailedError struct {
	java.Error
}

func (this *AssertionFailedError) JClass_AssertionFailedError() *AssertionFailedError { return this }
func (this *AssertionFailedError) JClassName() string {
	return "junit.framework.AssertionFailedError"
}

func fail(format string, args ...any) {
	e := &AssertionFailedError{}
	e.Error_init_String(java.FromGo(fmt.Sprintf(format, args...)))
	panic(e)
}

// TestCase is junit.framework.TestCase.
type TestCase struct {
	jself any
	fName *java.String
}

func (this *TestCase) JSelf() any        { return this.jself }
func (this *TestCase) JSetSelf(self any) { this.jself = self }

func (this *TestCase) JClass_TestCase() *TestCase { return this }
func (this *TestCase) JClassName() string         { return "junit.framework.TestCase" }

func (this *TestCase) TestCase_init()                      {}
func (this *TestCase) TestCase_init_String(n *java.String) { this.fName = n }
func (this *TestCase) SetUp()                              {}
func (this *TestCase) TearDown()                           {}
func (this *TestCase) GetName() *java.String               { return this.fName }

func msgPrefix(msg *java.String) string {
	if msg == nil {
		return ""
	}
	return msg.Go() + " "
}

func TestCase_fail_String(msg *java.String) { fail("%s", msgPrefix(msg)) }

func TestCase_fail() { fail("") }

func TestCase_assertTrue_boolean(c bool) { TestCase_assertTrue_String_boolean(nil, c) }

func TestCase_assertTrue_String_boolean(msg *java.String, c bool) {
	if !c {
		fail("%sexpected true", msgPrefix(msg))
	}
}

func TestCase_assertFalse_boolean(c bool) { TestCase_assertFalse_String_boolean(nil, c) }

func TestCase_assertFalse_String_boolean(msg *java.String, c bool) {
	if c {
		fail("%sexpected false", msgPrefix(msg))
	}
}

func TestCase_assertEquals_Object_Object(a, b any) {
	TestCase_assertEquals_String_Object_Object(nil, a, b)
}

func TestCase_assertEquals_String_Object_Object(msg *java.String, a, b any) {
	if a == nil && b == nil {
		return
	}
	if a != nil && java.Object_Equals_Object(a, b) {
		return
	}
	fail("%sexpected:<%s> but was:<%s>", msgPrefix(msg), str(a), str(b))
}

func TestCase_assertEquals_String_String(a, b *java.String) {
	TestCase_assertEquals_String_String_String(nil, a, b)
}

func TestCase_assertEquals_String_String_String(msg, a, b *java.String) {
	TestCase_assertEquals_String_Object_Object(msg, java.Ref(a), java.Ref(b))
}

func TestCase_assertEquals_int_int(a, b int32) { TestCase_assertEquals_String_int_int(nil, a, b) }

func TestCase_assertEquals_String_int_int(msg *java.String, a, b int32) {
	if a != b {
		fail("%sexpected:<%d> but was:<%d>", msgPrefix(msg), a, b)
	}
}

func TestCase_assertEquals_long_long(a, b int64) { TestCase_assertEquals_String_long_long(nil, a, b) }

func TestCase_assertEquals_String_long_long(msg *java.String, a, b int64) {
	if a != b {
		fail("%sexpected:<%d> but was:<%d>", msgPrefix(msg), a, b)
	}
}

func TestCase_assertEquals_boolean_boolean(a, b bool) {
	TestCase_assertEquals_String_boolean_boolean(nil, a, b)
}

func TestCase_assertEquals_String_boolean_boolean(msg *java.String, a, b bool) {
	if a != b {
		fail("%sexpected:<%v> but was:<%v>", msgPrefix(msg), a, b)
	}
}

func TestCase_assertEquals_char_char(a, b uint16) {
	if a != b {
		fail("expected:<%c> but was:<%c>", rune(a), rune(b))
	}
}

func TestCase_assertEquals_double_double_double(a, b, delta float64) {
	TestCase_assertEquals_String_double_double_double(nil, a, b, delta)
}

// TestCase_assertEquals_String_double_double_double follows JUnit 3: equal
// values (including infinities) pass, and otherwise the difference must be
// within delta.
func TestCase_assertEquals_String_double_double_double(msg *java.String, a, b, delta float64) {
	if java.Double_compare(a, b) == 0 {
		return
	}
	if !(math.Abs(a-b) <= delta) {
		fail("%sexpected:<%s> but was:<%s>", msgPrefix(msg), java.DoubleToString(a), java.DoubleToString(b))
	}
}

func TestCase_assertNull_Object(o any) { TestCase_assertNull_String_Object(nil, o) }

func TestCase_assertNull_String_Object(msg *java.String, o any) {
	if o != nil {
		fail("%sexpected null, but was:<%s>", msgPrefix(msg), str(o))
	}
}

func TestCase_assertNotNull_Object(o any) { TestCase_assertNotNull_String_Object(nil, o) }

func TestCase_assertNotNull_String_Object(msg *java.String, o any) {
	if o == nil {
		fail("%sexpected not null", msgPrefix(msg))
	}
}

func TestCase_assertSame_Object_Object(a, b any) {
	if !java.Same(a, b) {
		fail("expected same:<%s> was not:<%s>", str(a), str(b))
	}
}

func TestCase_assertNotSame_Object_Object(a, b any) {
	if java.Same(a, b) {
		fail("expected not same")
	}
}

func str(o any) string {
	if o == nil {
		return "null"
	}
	return java.Object_ToString(o).Go()
}

// J4_Assert_assertTrue_String_boolean is org.junit.Assert.assertTrue(String, boolean).
func J4_Assert_assertTrue_String_boolean(msg *java.String, c bool) {
	TestCase_assertTrue_String_boolean(msg, c)
}

// J4_Assert_assertEquals_Object_Object is org.junit.Assert.assertEquals(Object, Object).
func J4_Assert_assertEquals_Object_Object(a, b any) {
	TestCase_assertEquals_Object_Object(a, b)
}

// J4_Assert_assertEquals_String_double_double_double is
// org.junit.Assert.assertEquals(String message, double expected, double
// actual, double delta).
func J4_Assert_assertEquals_String_double_double_double(msg *java.String, a, b, delta float64) {
	TestCase_assertEquals_String_double_double_double(msg, a, b, delta)
}
