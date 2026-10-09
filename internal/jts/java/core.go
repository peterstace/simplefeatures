// Package java implements the JDK classes and Java language semantics that the
// generated JTS translation depends on.
//
// Names follow the translator's rules (see internal/jts/DESIGN.md): a Java
// instance method maps to a Go method named after the method and the erased
// parameter types of its root declaration (List.get(int) is Get_int), a static
// member maps to a function prefixed by its class (Math.abs(double) is
// Math_abs_double when abs is overloaded), and a constructor maps to a
// function suffixed with _new.
package java

import (
	"fmt"
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
)

// UnsupportedError is the panic value of generated code that the translator
// could not translate. Java catch clauses never catch it.
type UnsupportedError struct {
	Reason string
}

func (e UnsupportedError) Error() string {
	return "jts: not translated: " + e.Reason
}

// Unsupported returns the panic value for an untranslated member.
func Unsupported(reason string) any {
	return UnsupportedError{reason}
}

// ClassCastError returns a ClassCastException for a failed cast of o to the
// named class.
func ClassCastError(o any, target string) any {
	return ClassCastException_new_String(Lit(fmt.Sprintf("%s cannot be cast to %s", className(o), target)))
}

// AbstractMethodError returns the panic value for a call to an abstract method
// that has no implementation.
func AbstractMethodError(where string) any {
	return RuntimeException_new_String(Lit("abstract method called: " + where))
}

// Ref converts a pointer to an any, mapping a nil pointer to a nil interface.
// A pointer to a runtime class that a generated class extends (such as a
// Throwable) converts to the whole object, which the runtime class records.
func Ref[T any](p *T) any {
	if p == nil {
		return nil
	}
	if s, ok := any(p).(selfRecorder); ok {
		if self := s.JSelf(); self != nil {
			return self
		}
	}
	return p
}

// selfRecorder is implemented by runtime classes that generated classes
// extend. The generated constructor records the whole object with JSetSelf.
type selfRecorder interface {
	JSelf() any
	JSetSelf(any)
}

// ArrRef converts a slice (a Java array) to an any, mapping nil to nil.
func ArrRef[T any](s []T) any {
	if s == nil {
		return nil
	}
	return s
}

// AnyTo converts o to T as a Java checked cast does: nil converts to the zero
// value, and a value of the wrong type panics with a ClassCastException.
func AnyTo[T any](o any) T {
	if o == nil {
		var zero T
		return zero
	}
	v, ok := o.(T)
	if !ok {
		var zero T
		panic(ClassCastError(o, fmt.Sprintf("%T", zero)))
	}
	return v
}

// ObjArr converts o, a Java array of any element type, to a Java Object[].
// An array whose Go type is not []any is copied.
func ObjArr(o any) []any {
	if o == nil {
		return nil
	}
	if s, ok := o.([]any); ok {
		return s
	}
	v := reflect.ValueOf(o)
	if v.Kind() != reflect.Slice {
		panic(ClassCastError(o, "[Ljava.lang.Object;"))
	}
	out := make([]any, v.Len())
	for i := range out {
		out[i] = v.Index(i).Interface()
	}
	return out
}

// InstanceOf reports whether o is non-nil and has type T.
func InstanceOf[T any](o any) bool {
	_, ok := o.(T)
	return ok
}

// SliceCast reinterprets a slice of pointers to JTS classes as a slice of
// pointers to a related class. All parts of a JTS object share its address, so
// the reinterpretation keeps Java's array aliasing.
func SliceCast[T, S any](s []S) []T {
	if s == nil {
		return nil
	}
	if cap(s) == 0 {
		return []T{}
	}
	return unsafe.Slice((*T)(unsafe.Pointer(&s[:1][0])), cap(s))[:len(s)]
}

// MapSlice converts each element of s, mapping a nil slice to nil.
func MapSlice[S, T any](s []S, f func(S) T) []T {
	if s == nil {
		return nil
	}
	out := make([]T, len(s))
	for i, x := range s {
		out[i] = f(x)
	}
	return out
}

// CloneSlice implements clone() on a Java array.
func CloneSlice[T any](s []T) []T {
	if s == nil {
		panic(NullPointerException_new())
	}
	out := make([]T, len(s))
	copy(out, s)
	return out
}

// NewArray2 implements new T[a][b].
func NewArray2[T any](a, b int32) [][]T {
	out := make([][]T, a)
	for i := range out {
		out[i] = make([]T, b)
	}
	return out
}

// NewArray3 implements new T[a][b][c].
func NewArray3[T any](a, b, c int32) [][][]T {
	out := make([][][]T, a)
	for i := range out {
		out[i] = NewArray2[T](b, c)
	}
	return out
}

// F2I converts a floating-point value to int as Java does: NaN becomes 0 and
// out-of-range values saturate.
func F2I(f float64) int32 {
	switch {
	case f != f:
		return 0
	case f >= math.MaxInt32:
		return math.MaxInt32
	case f <= math.MinInt32:
		return math.MinInt32
	}
	return int32(f)
}

// F2L converts a floating-point value to long as Java does.
func F2L(f float64) int64 {
	switch {
	case f != f:
		return 0
	case f >= math.MaxInt64:
		return math.MaxInt64
	case f <= math.MinInt64:
		return math.MinInt64
	}
	return int64(f)
}

// Number is the set of Go types representing Java numeric primitives.
type Number interface {
	~int8 | ~int16 | ~uint16 | ~int32 | ~int64 | ~float32 | ~float64
}

// Assign implements an assignment used as an expression.
func Assign[T any](p *T, v T) T {
	*p = v
	return v
}

// PreInc implements ++x used as an expression.
func PreInc[T Number](p *T) T {
	*p++
	return *p
}

// PreDec implements --x used as an expression.
func PreDec[T Number](p *T) T {
	*p--
	return *p
}

// PostInc implements x++ used as an expression.
func PostInc[T Number](p *T) T {
	v := *p
	*p++
	return v
}

// PostDec implements x-- used as an expression.
func PostDec[T Number](p *T) T {
	v := *p
	*p--
	return v
}

// BoolAnd implements the non-short-circuit & operator on booleans.
func BoolAnd(a, b bool) bool {
	return a && b
}

// BoolOr implements the non-short-circuit | operator on booleans.
func BoolOr(a, b bool) bool {
	return a || b
}

// StaticInit evaluates the initializer of a Java static field. Go initializes
// package-level variables at program start, whereas Java initializes a class
// on first use, so a failing initializer must not stop the whole program: it
// leaves the zero value and is recorded in StaticInitFailures, which a
// generated test checks is empty.
func StaticInit[T any](f func() T) (v T) {
	defer func() {
		if r := recover(); r != nil {
			staticInitMu.Lock()
			defer staticInitMu.Unlock()
			StaticInitFailures = append(StaticInitFailures, fmt.Sprint(r))
		}
	}()
	return f()
}

var (
	staticInitMu sync.Mutex
	// StaticInitFailures lists the panics of failed static initializers.
	StaticInitFailures []string
)

// VolatileBool is a Java volatile boolean field, read and written atomically.
type VolatileBool struct {
	v int32
}

func (b *VolatileBool) Load() bool { return atomic.LoadInt32(&b.v) != 0 }

func (b *VolatileBool) Store(v bool) {
	var x int32
	if v {
		x = 1
	}
	atomic.StoreInt32(&b.v, x)
}

// VolatilePtr is a Java volatile field of class type T, read and written
// atomically.
type VolatilePtr[T any] struct {
	p unsafe.Pointer
}

func (r *VolatilePtr[T]) Load() *T { return (*T)(atomic.LoadPointer(&r.p)) }

func (r *VolatilePtr[T]) Store(v *T) { atomic.StorePointer(&r.p, unsafe.Pointer(v)) }
