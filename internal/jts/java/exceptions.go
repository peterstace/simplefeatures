// Code below the Throwable type is boilerplate for the JDK exception classes.

package java

// Throwable is java.lang.Throwable.
type Throwable struct {
	jself   any
	message *String
	cause   *Throwable
}

func (this *Throwable) JSelf() any        { return this.jself }
func (this *Throwable) JSetSelf(self any) { this.jself = self }

func (this *Throwable) JClass_Throwable() *Throwable { return this }
func (this *Throwable) JClassName() string           { return "java.lang.Throwable" }

func Throwable_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_Throwable() *Throwable })
	return ok
}

func Throwable_fromAny(o any) *Throwable {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_Throwable() *Throwable }](o).JClass_Throwable()
}

func Throwable_new() *Throwable {
	t := &Throwable{}
	t.jself = t
	return t
}

func Throwable_new_String(msg *String) *Throwable {
	t := Throwable_new()
	t.message = msg
	return t
}

func Throwable_new_String_Throwable(msg *String, cause *Throwable) *Throwable {
	t := Throwable_new()
	t.Throwable_init_String_Throwable(msg, cause)
	return t
}

func Throwable_new_Throwable(cause *Throwable) *Throwable {
	t := Throwable_new()
	t.Throwable_init_Throwable(cause)
	return t
}

func (this *Throwable) Throwable_init() {}

func (this *Throwable) Throwable_init_String(msg *String) { this.message = msg }

func (this *Throwable) Throwable_init_String_Throwable(msg *String, cause *Throwable) {
	this.message = msg
	this.cause = cause
}

func (this *Throwable) Throwable_init_Throwable(cause *Throwable) {
	this.cause = cause
	if cause != nil {
		this.message = Object_ToString(Ref(cause))
	}
}

func (this *Throwable) GetMessage() *String { return this.message }

func (this *Throwable) GetLocalizedMessage() *String { return this.message }

func (this *Throwable) GetCause() *Throwable { return this.cause }

// PrintStackTrace prints the exception and its causes to System.err. There
// are no stack frames to print.
func (this *Throwable) PrintStackTrace() {
	this.PrintStackTrace_PrintStream(System_err)
}

// ToString is only reached through a *Throwable, which does not know the
// concrete class; Object_ToString handles throwables with their class name.
func (this *Throwable) ToString() *String {
	return throwableToString("java.lang.Throwable", this)
}

func throwableToString(class string, t *Throwable) *String {
	if t.message == nil {
		return FromGo(class)
	}
	return Concat(FromGo(class+": "), t.message)
}

// Exception is java.lang.Exception.
type Exception struct {
	Throwable
}

func (this *Exception) JClass_Exception() *Exception { return this }
func (this *Exception) JClassName() string           { return "java.lang.Exception" }

func Exception_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_Exception() *Exception })
	return ok
}

func Exception_fromAny(o any) *Exception {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_Exception() *Exception }](o).JClass_Exception()
}

func Exception_new() *Exception {
	this := &Exception{}
	this.jself = this
	this.Exception_init()
	return this
}

func (this *Exception) Exception_init() {
	this.Throwable.Throwable_init()
}

func Exception_new_String(a0 *String) *Exception {
	this := &Exception{}
	this.jself = this
	this.Exception_init_String(a0)
	return this
}

func (this *Exception) Exception_init_String(a0 *String) {
	this.Throwable.Throwable_init_String(a0)
}

func Exception_new_String_Throwable(a0 *String, a1 *Throwable) *Exception {
	this := &Exception{}
	this.jself = this
	this.Exception_init_String_Throwable(a0, a1)
	return this
}

func (this *Exception) Exception_init_String_Throwable(a0 *String, a1 *Throwable) {
	this.Throwable.Throwable_init_String_Throwable(a0, a1)
}

func Exception_new_Throwable(a0 *Throwable) *Exception {
	this := &Exception{}
	this.jself = this
	this.Exception_init_Throwable(a0)
	return this
}

func (this *Exception) Exception_init_Throwable(a0 *Throwable) {
	this.Throwable.Throwable_init_Throwable(a0)
}

// RuntimeException is java.lang.RuntimeException.
type RuntimeException struct {
	Exception
}

func (this *RuntimeException) JClass_RuntimeException() *RuntimeException { return this }
func (this *RuntimeException) JClassName() string                         { return "java.lang.RuntimeException" }

func RuntimeException_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_RuntimeException() *RuntimeException })
	return ok
}

func RuntimeException_fromAny(o any) *RuntimeException {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_RuntimeException() *RuntimeException }](o).JClass_RuntimeException()
}

func RuntimeException_new() *RuntimeException {
	this := &RuntimeException{}
	this.jself = this
	this.RuntimeException_init()
	return this
}

func (this *RuntimeException) RuntimeException_init() {
	this.Exception.Exception_init()
}

func RuntimeException_new_String(a0 *String) *RuntimeException {
	this := &RuntimeException{}
	this.jself = this
	this.RuntimeException_init_String(a0)
	return this
}

func (this *RuntimeException) RuntimeException_init_String(a0 *String) {
	this.Exception.Exception_init_String(a0)
}

func RuntimeException_new_String_Throwable(a0 *String, a1 *Throwable) *RuntimeException {
	this := &RuntimeException{}
	this.jself = this
	this.RuntimeException_init_String_Throwable(a0, a1)
	return this
}

func (this *RuntimeException) RuntimeException_init_String_Throwable(a0 *String, a1 *Throwable) {
	this.Exception.Exception_init_String_Throwable(a0, a1)
}

func RuntimeException_new_Throwable(a0 *Throwable) *RuntimeException {
	this := &RuntimeException{}
	this.jself = this
	this.RuntimeException_init_Throwable(a0)
	return this
}

func (this *RuntimeException) RuntimeException_init_Throwable(a0 *Throwable) {
	this.Exception.Exception_init_Throwable(a0)
}

// Error is java.lang.Error.
type Error struct {
	Throwable
}

func (this *Error) JClass_Error() *Error { return this }
func (this *Error) JClassName() string   { return "java.lang.Error" }

func Error_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_Error() *Error })
	return ok
}

func Error_fromAny(o any) *Error {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_Error() *Error }](o).JClass_Error()
}

func Error_new() *Error {
	this := &Error{}
	this.jself = this
	this.Error_init()
	return this
}

func (this *Error) Error_init() {
	this.Throwable.Throwable_init()
}

func Error_new_String(a0 *String) *Error {
	this := &Error{}
	this.jself = this
	this.Error_init_String(a0)
	return this
}

func (this *Error) Error_init_String(a0 *String) {
	this.Throwable.Throwable_init_String(a0)
}

func Error_new_String_Throwable(a0 *String, a1 *Throwable) *Error {
	this := &Error{}
	this.jself = this
	this.Error_init_String_Throwable(a0, a1)
	return this
}

func (this *Error) Error_init_String_Throwable(a0 *String, a1 *Throwable) {
	this.Throwable.Throwable_init_String_Throwable(a0, a1)
}

func Error_new_Throwable(a0 *Throwable) *Error {
	this := &Error{}
	this.jself = this
	this.Error_init_Throwable(a0)
	return this
}

func (this *Error) Error_init_Throwable(a0 *Throwable) {
	this.Throwable.Throwable_init_Throwable(a0)
}

// IllegalArgumentException is java.lang.IllegalArgumentException.
type IllegalArgumentException struct {
	RuntimeException
}

func (this *IllegalArgumentException) JClass_IllegalArgumentException() *IllegalArgumentException {
	return this
}

func (this *IllegalArgumentException) JClassName() string {
	return "java.lang.IllegalArgumentException"
}

func IllegalArgumentException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_IllegalArgumentException() *IllegalArgumentException
	})
	return ok
}

func IllegalArgumentException_fromAny(o any) *IllegalArgumentException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_IllegalArgumentException() *IllegalArgumentException
	}](o).JClass_IllegalArgumentException()
}

func IllegalArgumentException_new() *IllegalArgumentException {
	this := &IllegalArgumentException{}
	this.jself = this
	this.IllegalArgumentException_init()
	return this
}

func (this *IllegalArgumentException) IllegalArgumentException_init() {
	this.RuntimeException.RuntimeException_init()
}

func IllegalArgumentException_new_String(a0 *String) *IllegalArgumentException {
	this := &IllegalArgumentException{}
	this.jself = this
	this.IllegalArgumentException_init_String(a0)
	return this
}

func (this *IllegalArgumentException) IllegalArgumentException_init_String(a0 *String) {
	this.RuntimeException.RuntimeException_init_String(a0)
}

func IllegalArgumentException_new_String_Throwable(a0 *String, a1 *Throwable) *IllegalArgumentException {
	this := &IllegalArgumentException{}
	this.jself = this
	this.IllegalArgumentException_init_String_Throwable(a0, a1)
	return this
}

func (this *IllegalArgumentException) IllegalArgumentException_init_String_Throwable(a0 *String, a1 *Throwable) {
	this.RuntimeException.RuntimeException_init_String_Throwable(a0, a1)
}

func IllegalArgumentException_new_Throwable(a0 *Throwable) *IllegalArgumentException {
	this := &IllegalArgumentException{}
	this.jself = this
	this.IllegalArgumentException_init_Throwable(a0)
	return this
}

func (this *IllegalArgumentException) IllegalArgumentException_init_Throwable(a0 *Throwable) {
	this.RuntimeException.RuntimeException_init_Throwable(a0)
}

// IllegalStateException is java.lang.IllegalStateException.
type IllegalStateException struct {
	RuntimeException
}

func (this *IllegalStateException) JClass_IllegalStateException() *IllegalStateException { return this }
func (this *IllegalStateException) JClassName() string                                   { return "java.lang.IllegalStateException" }

func IllegalStateException_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_IllegalStateException() *IllegalStateException })
	return ok
}

func IllegalStateException_fromAny(o any) *IllegalStateException {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_IllegalStateException() *IllegalStateException }](o).JClass_IllegalStateException()
}

func IllegalStateException_new() *IllegalStateException {
	this := &IllegalStateException{}
	this.jself = this
	this.IllegalStateException_init()
	return this
}

func (this *IllegalStateException) IllegalStateException_init() {
	this.RuntimeException.RuntimeException_init()
}

func IllegalStateException_new_String(a0 *String) *IllegalStateException {
	this := &IllegalStateException{}
	this.jself = this
	this.IllegalStateException_init_String(a0)
	return this
}

func (this *IllegalStateException) IllegalStateException_init_String(a0 *String) {
	this.RuntimeException.RuntimeException_init_String(a0)
}

func IllegalStateException_new_String_Throwable(a0 *String, a1 *Throwable) *IllegalStateException {
	this := &IllegalStateException{}
	this.jself = this
	this.IllegalStateException_init_String_Throwable(a0, a1)
	return this
}

func (this *IllegalStateException) IllegalStateException_init_String_Throwable(a0 *String, a1 *Throwable) {
	this.RuntimeException.RuntimeException_init_String_Throwable(a0, a1)
}

func IllegalStateException_new_Throwable(a0 *Throwable) *IllegalStateException {
	this := &IllegalStateException{}
	this.jself = this
	this.IllegalStateException_init_Throwable(a0)
	return this
}

func (this *IllegalStateException) IllegalStateException_init_Throwable(a0 *Throwable) {
	this.RuntimeException.RuntimeException_init_Throwable(a0)
}

// UnsupportedOperationException is java.lang.UnsupportedOperationException.
type UnsupportedOperationException struct {
	RuntimeException
}

func (this *UnsupportedOperationException) JClass_UnsupportedOperationException() *UnsupportedOperationException {
	return this
}

func (this *UnsupportedOperationException) JClassName() string {
	return "java.lang.UnsupportedOperationException"
}

func UnsupportedOperationException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_UnsupportedOperationException() *UnsupportedOperationException
	})
	return ok
}

func UnsupportedOperationException_fromAny(o any) *UnsupportedOperationException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_UnsupportedOperationException() *UnsupportedOperationException
	}](o).JClass_UnsupportedOperationException()
}

func UnsupportedOperationException_new() *UnsupportedOperationException {
	this := &UnsupportedOperationException{}
	this.jself = this
	this.UnsupportedOperationException_init()
	return this
}

func (this *UnsupportedOperationException) UnsupportedOperationException_init() {
	this.RuntimeException.RuntimeException_init()
}

func UnsupportedOperationException_new_String(a0 *String) *UnsupportedOperationException {
	this := &UnsupportedOperationException{}
	this.jself = this
	this.UnsupportedOperationException_init_String(a0)
	return this
}

func (this *UnsupportedOperationException) UnsupportedOperationException_init_String(a0 *String) {
	this.RuntimeException.RuntimeException_init_String(a0)
}

func UnsupportedOperationException_new_String_Throwable(a0 *String, a1 *Throwable) *UnsupportedOperationException {
	this := &UnsupportedOperationException{}
	this.jself = this
	this.UnsupportedOperationException_init_String_Throwable(a0, a1)
	return this
}

func (this *UnsupportedOperationException) UnsupportedOperationException_init_String_Throwable(a0 *String, a1 *Throwable) {
	this.RuntimeException.RuntimeException_init_String_Throwable(a0, a1)
}

func UnsupportedOperationException_new_Throwable(a0 *Throwable) *UnsupportedOperationException {
	this := &UnsupportedOperationException{}
	this.jself = this
	this.UnsupportedOperationException_init_Throwable(a0)
	return this
}

func (this *UnsupportedOperationException) UnsupportedOperationException_init_Throwable(a0 *Throwable) {
	this.RuntimeException.RuntimeException_init_Throwable(a0)
}

// NullPointerException is java.lang.NullPointerException.
type NullPointerException struct {
	RuntimeException
}

func (this *NullPointerException) JClass_NullPointerException() *NullPointerException { return this }
func (this *NullPointerException) JClassName() string                                 { return "java.lang.NullPointerException" }

func NullPointerException_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_NullPointerException() *NullPointerException })
	return ok
}

func NullPointerException_fromAny(o any) *NullPointerException {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_NullPointerException() *NullPointerException }](o).JClass_NullPointerException()
}

func NullPointerException_new() *NullPointerException {
	this := &NullPointerException{}
	this.jself = this
	this.NullPointerException_init()
	return this
}

func (this *NullPointerException) NullPointerException_init() {
	this.RuntimeException.RuntimeException_init()
}

func NullPointerException_new_String(a0 *String) *NullPointerException {
	this := &NullPointerException{}
	this.jself = this
	this.NullPointerException_init_String(a0)
	return this
}

func (this *NullPointerException) NullPointerException_init_String(a0 *String) {
	this.RuntimeException.RuntimeException_init_String(a0)
}

// ClassCastException is java.lang.ClassCastException.
type ClassCastException struct {
	RuntimeException
}

func (this *ClassCastException) JClass_ClassCastException() *ClassCastException { return this }
func (this *ClassCastException) JClassName() string                             { return "java.lang.ClassCastException" }

func ClassCastException_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_ClassCastException() *ClassCastException })
	return ok
}

func ClassCastException_fromAny(o any) *ClassCastException {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_ClassCastException() *ClassCastException }](o).JClass_ClassCastException()
}

func ClassCastException_new() *ClassCastException {
	this := &ClassCastException{}
	this.jself = this
	this.ClassCastException_init()
	return this
}

func (this *ClassCastException) ClassCastException_init() {
	this.RuntimeException.RuntimeException_init()
}

func ClassCastException_new_String(a0 *String) *ClassCastException {
	this := &ClassCastException{}
	this.jself = this
	this.ClassCastException_init_String(a0)
	return this
}

func (this *ClassCastException) ClassCastException_init_String(a0 *String) {
	this.RuntimeException.RuntimeException_init_String(a0)
}

// IndexOutOfBoundsException is java.lang.IndexOutOfBoundsException.
type IndexOutOfBoundsException struct {
	RuntimeException
}

func (this *IndexOutOfBoundsException) JClass_IndexOutOfBoundsException() *IndexOutOfBoundsException {
	return this
}

func (this *IndexOutOfBoundsException) JClassName() string {
	return "java.lang.IndexOutOfBoundsException"
}

func IndexOutOfBoundsException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_IndexOutOfBoundsException() *IndexOutOfBoundsException
	})
	return ok
}

func IndexOutOfBoundsException_fromAny(o any) *IndexOutOfBoundsException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_IndexOutOfBoundsException() *IndexOutOfBoundsException
	}](o).JClass_IndexOutOfBoundsException()
}

func IndexOutOfBoundsException_new() *IndexOutOfBoundsException {
	this := &IndexOutOfBoundsException{}
	this.jself = this
	this.IndexOutOfBoundsException_init()
	return this
}

func (this *IndexOutOfBoundsException) IndexOutOfBoundsException_init() {
	this.RuntimeException.RuntimeException_init()
}

func IndexOutOfBoundsException_new_String(a0 *String) *IndexOutOfBoundsException {
	this := &IndexOutOfBoundsException{}
	this.jself = this
	this.IndexOutOfBoundsException_init_String(a0)
	return this
}

func (this *IndexOutOfBoundsException) IndexOutOfBoundsException_init_String(a0 *String) {
	this.RuntimeException.RuntimeException_init_String(a0)
}

// ArrayIndexOutOfBoundsException is java.lang.ArrayIndexOutOfBoundsException.
type ArrayIndexOutOfBoundsException struct {
	IndexOutOfBoundsException
}

func (this *ArrayIndexOutOfBoundsException) JClass_ArrayIndexOutOfBoundsException() *ArrayIndexOutOfBoundsException {
	return this
}

func (this *ArrayIndexOutOfBoundsException) JClassName() string {
	return "java.lang.ArrayIndexOutOfBoundsException"
}

func ArrayIndexOutOfBoundsException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_ArrayIndexOutOfBoundsException() *ArrayIndexOutOfBoundsException
	})
	return ok
}

func ArrayIndexOutOfBoundsException_fromAny(o any) *ArrayIndexOutOfBoundsException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_ArrayIndexOutOfBoundsException() *ArrayIndexOutOfBoundsException
	}](o).JClass_ArrayIndexOutOfBoundsException()
}

func ArrayIndexOutOfBoundsException_new() *ArrayIndexOutOfBoundsException {
	this := &ArrayIndexOutOfBoundsException{}
	this.jself = this
	this.ArrayIndexOutOfBoundsException_init()
	return this
}

func (this *ArrayIndexOutOfBoundsException) ArrayIndexOutOfBoundsException_init() {
	this.IndexOutOfBoundsException.IndexOutOfBoundsException_init()
}

func ArrayIndexOutOfBoundsException_new_int(a0 int32) *ArrayIndexOutOfBoundsException {
	this := &ArrayIndexOutOfBoundsException{}
	this.jself = this
	this.ArrayIndexOutOfBoundsException_init_int(a0)
	return this
}

func (this *ArrayIndexOutOfBoundsException) ArrayIndexOutOfBoundsException_init_int(a0 int32) {
	this.IndexOutOfBoundsException.IndexOutOfBoundsException_init_String(Concat(Lit("Array index out of range: "), StrOfInt(a0)))
}

func ArrayIndexOutOfBoundsException_new_String(a0 *String) *ArrayIndexOutOfBoundsException {
	this := &ArrayIndexOutOfBoundsException{}
	this.jself = this
	this.ArrayIndexOutOfBoundsException_init_String(a0)
	return this
}

func (this *ArrayIndexOutOfBoundsException) ArrayIndexOutOfBoundsException_init_String(a0 *String) {
	this.IndexOutOfBoundsException.IndexOutOfBoundsException_init_String(a0)
}

// StringIndexOutOfBoundsException is java.lang.StringIndexOutOfBoundsException.
type StringIndexOutOfBoundsException struct {
	IndexOutOfBoundsException
}

func (this *StringIndexOutOfBoundsException) JClass_StringIndexOutOfBoundsException() *StringIndexOutOfBoundsException {
	return this
}

func (this *StringIndexOutOfBoundsException) JClassName() string {
	return "java.lang.StringIndexOutOfBoundsException"
}

func StringIndexOutOfBoundsException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_StringIndexOutOfBoundsException() *StringIndexOutOfBoundsException
	})
	return ok
}

func StringIndexOutOfBoundsException_fromAny(o any) *StringIndexOutOfBoundsException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_StringIndexOutOfBoundsException() *StringIndexOutOfBoundsException
	}](o).JClass_StringIndexOutOfBoundsException()
}

func StringIndexOutOfBoundsException_new() *StringIndexOutOfBoundsException {
	this := &StringIndexOutOfBoundsException{}
	this.jself = this
	this.StringIndexOutOfBoundsException_init()
	return this
}

func (this *StringIndexOutOfBoundsException) StringIndexOutOfBoundsException_init() {
	this.IndexOutOfBoundsException.IndexOutOfBoundsException_init()
}

func StringIndexOutOfBoundsException_new_String(a0 *String) *StringIndexOutOfBoundsException {
	this := &StringIndexOutOfBoundsException{}
	this.jself = this
	this.StringIndexOutOfBoundsException_init_String(a0)
	return this
}

func (this *StringIndexOutOfBoundsException) StringIndexOutOfBoundsException_init_String(a0 *String) {
	this.IndexOutOfBoundsException.IndexOutOfBoundsException_init_String(a0)
}

func StringIndexOutOfBoundsException_new_int(a0 int32) *StringIndexOutOfBoundsException {
	this := &StringIndexOutOfBoundsException{}
	this.jself = this
	this.StringIndexOutOfBoundsException_init_int(a0)
	return this
}

func (this *StringIndexOutOfBoundsException) StringIndexOutOfBoundsException_init_int(a0 int32) {
	this.IndexOutOfBoundsException.IndexOutOfBoundsException_init_String(Concat(Lit("String index out of range: "), StrOfInt(a0)))
}

// ArithmeticException is java.lang.ArithmeticException.
type ArithmeticException struct {
	RuntimeException
}

func (this *ArithmeticException) JClass_ArithmeticException() *ArithmeticException { return this }
func (this *ArithmeticException) JClassName() string                               { return "java.lang.ArithmeticException" }

func ArithmeticException_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_ArithmeticException() *ArithmeticException })
	return ok
}

func ArithmeticException_fromAny(o any) *ArithmeticException {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_ArithmeticException() *ArithmeticException }](o).JClass_ArithmeticException()
}

func ArithmeticException_new() *ArithmeticException {
	this := &ArithmeticException{}
	this.jself = this
	this.ArithmeticException_init()
	return this
}

func (this *ArithmeticException) ArithmeticException_init() {
	this.RuntimeException.RuntimeException_init()
}

func ArithmeticException_new_String(a0 *String) *ArithmeticException {
	this := &ArithmeticException{}
	this.jself = this
	this.ArithmeticException_init_String(a0)
	return this
}

func (this *ArithmeticException) ArithmeticException_init_String(a0 *String) {
	this.RuntimeException.RuntimeException_init_String(a0)
}

// NegativeArraySizeException is java.lang.NegativeArraySizeException.
type NegativeArraySizeException struct {
	RuntimeException
}

func (this *NegativeArraySizeException) JClass_NegativeArraySizeException() *NegativeArraySizeException {
	return this
}

func (this *NegativeArraySizeException) JClassName() string {
	return "java.lang.NegativeArraySizeException"
}

func NegativeArraySizeException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_NegativeArraySizeException() *NegativeArraySizeException
	})
	return ok
}

func NegativeArraySizeException_fromAny(o any) *NegativeArraySizeException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_NegativeArraySizeException() *NegativeArraySizeException
	}](o).JClass_NegativeArraySizeException()
}

func NegativeArraySizeException_new() *NegativeArraySizeException {
	this := &NegativeArraySizeException{}
	this.jself = this
	this.NegativeArraySizeException_init()
	return this
}

func (this *NegativeArraySizeException) NegativeArraySizeException_init() {
	this.RuntimeException.RuntimeException_init()
}

func NegativeArraySizeException_new_String(a0 *String) *NegativeArraySizeException {
	this := &NegativeArraySizeException{}
	this.jself = this
	this.NegativeArraySizeException_init_String(a0)
	return this
}

func (this *NegativeArraySizeException) NegativeArraySizeException_init_String(a0 *String) {
	this.RuntimeException.RuntimeException_init_String(a0)
}

// NumberFormatException is java.lang.NumberFormatException.
type NumberFormatException struct {
	IllegalArgumentException
}

func (this *NumberFormatException) JClass_NumberFormatException() *NumberFormatException { return this }
func (this *NumberFormatException) JClassName() string                                   { return "java.lang.NumberFormatException" }

func NumberFormatException_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_NumberFormatException() *NumberFormatException })
	return ok
}

func NumberFormatException_fromAny(o any) *NumberFormatException {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_NumberFormatException() *NumberFormatException }](o).JClass_NumberFormatException()
}

func NumberFormatException_new() *NumberFormatException {
	this := &NumberFormatException{}
	this.jself = this
	this.NumberFormatException_init()
	return this
}

func (this *NumberFormatException) NumberFormatException_init() {
	this.IllegalArgumentException.IllegalArgumentException_init()
}

func NumberFormatException_new_String(a0 *String) *NumberFormatException {
	this := &NumberFormatException{}
	this.jself = this
	this.NumberFormatException_init_String(a0)
	return this
}

func (this *NumberFormatException) NumberFormatException_init_String(a0 *String) {
	this.IllegalArgumentException.IllegalArgumentException_init_String(a0)
}

// NoSuchElementException is java.util.NoSuchElementException.
type NoSuchElementException struct {
	RuntimeException
}

func (this *NoSuchElementException) JClass_NoSuchElementException() *NoSuchElementException {
	return this
}
func (this *NoSuchElementException) JClassName() string { return "java.util.NoSuchElementException" }

func NoSuchElementException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_NoSuchElementException() *NoSuchElementException
	})
	return ok
}

func NoSuchElementException_fromAny(o any) *NoSuchElementException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_NoSuchElementException() *NoSuchElementException
	}](o).JClass_NoSuchElementException()
}

func NoSuchElementException_new() *NoSuchElementException {
	this := &NoSuchElementException{}
	this.jself = this
	this.NoSuchElementException_init()
	return this
}

func (this *NoSuchElementException) NoSuchElementException_init() {
	this.RuntimeException.RuntimeException_init()
}

func NoSuchElementException_new_String(a0 *String) *NoSuchElementException {
	this := &NoSuchElementException{}
	this.jself = this
	this.NoSuchElementException_init_String(a0)
	return this
}

func (this *NoSuchElementException) NoSuchElementException_init_String(a0 *String) {
	this.RuntimeException.RuntimeException_init_String(a0)
}

// ConcurrentModificationException is java.util.ConcurrentModificationException.
type ConcurrentModificationException struct {
	RuntimeException
}

func (this *ConcurrentModificationException) JClass_ConcurrentModificationException() *ConcurrentModificationException {
	return this
}

func (this *ConcurrentModificationException) JClassName() string {
	return "java.util.ConcurrentModificationException"
}

func ConcurrentModificationException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_ConcurrentModificationException() *ConcurrentModificationException
	})
	return ok
}

func ConcurrentModificationException_fromAny(o any) *ConcurrentModificationException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_ConcurrentModificationException() *ConcurrentModificationException
	}](o).JClass_ConcurrentModificationException()
}

func ConcurrentModificationException_new() *ConcurrentModificationException {
	this := &ConcurrentModificationException{}
	this.jself = this
	this.ConcurrentModificationException_init()
	return this
}

func (this *ConcurrentModificationException) ConcurrentModificationException_init() {
	this.RuntimeException.RuntimeException_init()
}

func ConcurrentModificationException_new_String(a0 *String) *ConcurrentModificationException {
	this := &ConcurrentModificationException{}
	this.jself = this
	this.ConcurrentModificationException_init_String(a0)
	return this
}

func (this *ConcurrentModificationException) ConcurrentModificationException_init_String(a0 *String) {
	this.RuntimeException.RuntimeException_init_String(a0)
}

func ConcurrentModificationException_new_Throwable(a0 *Throwable) *ConcurrentModificationException {
	this := &ConcurrentModificationException{}
	this.jself = this
	this.ConcurrentModificationException_init_Throwable(a0)
	return this
}

func (this *ConcurrentModificationException) ConcurrentModificationException_init_Throwable(a0 *Throwable) {
	this.RuntimeException.RuntimeException_init_Throwable(a0)
}

func ConcurrentModificationException_new_String_Throwable(a0 *String, a1 *Throwable) *ConcurrentModificationException {
	this := &ConcurrentModificationException{}
	this.jself = this
	this.ConcurrentModificationException_init_String_Throwable(a0, a1)
	return this
}

func (this *ConcurrentModificationException) ConcurrentModificationException_init_String_Throwable(a0 *String, a1 *Throwable) {
	this.RuntimeException.RuntimeException_init_String_Throwable(a0, a1)
}

// EmptyStackException is java.util.EmptyStackException.
type EmptyStackException struct {
	RuntimeException
}

func (this *EmptyStackException) JClass_EmptyStackException() *EmptyStackException { return this }
func (this *EmptyStackException) JClassName() string                               { return "java.util.EmptyStackException" }

func EmptyStackException_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_EmptyStackException() *EmptyStackException })
	return ok
}

func EmptyStackException_fromAny(o any) *EmptyStackException {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_EmptyStackException() *EmptyStackException }](o).JClass_EmptyStackException()
}

func EmptyStackException_new() *EmptyStackException {
	this := &EmptyStackException{}
	this.jself = this
	this.EmptyStackException_init()
	return this
}

func (this *EmptyStackException) EmptyStackException_init() {
	this.RuntimeException.RuntimeException_init()
}

// IOException is java.io.IOException.
type IOException struct {
	Exception
}

func (this *IOException) JClass_IOException() *IOException { return this }
func (this *IOException) JClassName() string               { return "java.io.IOException" }

func IOException_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_IOException() *IOException })
	return ok
}

func IOException_fromAny(o any) *IOException {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_IOException() *IOException }](o).JClass_IOException()
}

func IOException_new() *IOException {
	this := &IOException{}
	this.jself = this
	this.IOException_init()
	return this
}

func (this *IOException) IOException_init() {
	this.Exception.Exception_init()
}

func IOException_new_String(a0 *String) *IOException {
	this := &IOException{}
	this.jself = this
	this.IOException_init_String(a0)
	return this
}

func (this *IOException) IOException_init_String(a0 *String) {
	this.Exception.Exception_init_String(a0)
}

func IOException_new_String_Throwable(a0 *String, a1 *Throwable) *IOException {
	this := &IOException{}
	this.jself = this
	this.IOException_init_String_Throwable(a0, a1)
	return this
}

func (this *IOException) IOException_init_String_Throwable(a0 *String, a1 *Throwable) {
	this.Exception.Exception_init_String_Throwable(a0, a1)
}

func IOException_new_Throwable(a0 *Throwable) *IOException {
	this := &IOException{}
	this.jself = this
	this.IOException_init_Throwable(a0)
	return this
}

func (this *IOException) IOException_init_Throwable(a0 *Throwable) {
	this.Exception.Exception_init_Throwable(a0)
}

// PrintStackTrace_PrintStream implements Throwable.printStackTrace(PrintStream).
// Stack frames are not available, so it prints the throwable and its causes
// without "at" lines.
func (this *Throwable) PrintStackTrace_PrintStream(s *PrintStream) {
	s.Println_Object(Ref(this))
	seen := map[*Throwable]bool{this: true}
	for c := this.cause; c != nil && !seen[c]; c = c.cause {
		seen[c] = true
		s.Println_String(Concat(Lit("Caused by: "), StrOfObject(Ref(c))))
	}
}

// CloneNotSupportedException is java.lang.CloneNotSupportedException.
type CloneNotSupportedException struct {
	Exception
}

func (this *CloneNotSupportedException) JClass_CloneNotSupportedException() *CloneNotSupportedException {
	return this
}

func (this *CloneNotSupportedException) JClassName() string {
	return "java.lang.CloneNotSupportedException"
}

func CloneNotSupportedException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_CloneNotSupportedException() *CloneNotSupportedException
	})
	return ok
}

func CloneNotSupportedException_fromAny(o any) *CloneNotSupportedException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_CloneNotSupportedException() *CloneNotSupportedException
	}](o).JClass_CloneNotSupportedException()
}

func CloneNotSupportedException_new() *CloneNotSupportedException {
	this := &CloneNotSupportedException{}
	this.jself = this
	this.CloneNotSupportedException_init()
	return this
}

func (this *CloneNotSupportedException) CloneNotSupportedException_init() {
	this.Exception.Exception_init()
}

func CloneNotSupportedException_new_String(a0 *String) *CloneNotSupportedException {
	this := &CloneNotSupportedException{}
	this.jself = this
	this.CloneNotSupportedException_init_String(a0)
	return this
}

func (this *CloneNotSupportedException) CloneNotSupportedException_init_String(a0 *String) {
	this.Exception.Exception_init_String(a0)
}

// IllegalThreadStateException is java.lang.IllegalThreadStateException.
type IllegalThreadStateException struct {
	IllegalArgumentException
}

func (this *IllegalThreadStateException) JClass_IllegalThreadStateException() *IllegalThreadStateException {
	return this
}

func (this *IllegalThreadStateException) JClassName() string {
	return "java.lang.IllegalThreadStateException"
}

func IllegalThreadStateException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_IllegalThreadStateException() *IllegalThreadStateException
	})
	return ok
}

func IllegalThreadStateException_fromAny(o any) *IllegalThreadStateException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_IllegalThreadStateException() *IllegalThreadStateException
	}](o).JClass_IllegalThreadStateException()
}

func IllegalThreadStateException_new() *IllegalThreadStateException {
	this := &IllegalThreadStateException{}
	this.jself = this
	this.IllegalThreadStateException_init()
	return this
}

func (this *IllegalThreadStateException) IllegalThreadStateException_init() {
	this.IllegalArgumentException.IllegalArgumentException_init()
}

func IllegalThreadStateException_new_String(a0 *String) *IllegalThreadStateException {
	this := &IllegalThreadStateException{}
	this.jself = this
	this.IllegalThreadStateException_init_String(a0)
	return this
}

func (this *IllegalThreadStateException) IllegalThreadStateException_init_String(a0 *String) {
	this.IllegalArgumentException.IllegalArgumentException_init_String(a0)
}

// FileNotFoundException is java.io.FileNotFoundException.
type FileNotFoundException struct {
	IOException
}

func (this *FileNotFoundException) JClass_FileNotFoundException() *FileNotFoundException { return this }
func (this *FileNotFoundException) JClassName() string                                   { return "java.io.FileNotFoundException" }

func FileNotFoundException_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_FileNotFoundException() *FileNotFoundException })
	return ok
}

func FileNotFoundException_fromAny(o any) *FileNotFoundException {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_FileNotFoundException() *FileNotFoundException }](o).JClass_FileNotFoundException()
}

func FileNotFoundException_new() *FileNotFoundException {
	this := &FileNotFoundException{}
	this.jself = this
	this.FileNotFoundException_init()
	return this
}

func (this *FileNotFoundException) FileNotFoundException_init() {
	this.IOException.IOException_init()
}

func FileNotFoundException_new_String(a0 *String) *FileNotFoundException {
	this := &FileNotFoundException{}
	this.jself = this
	this.FileNotFoundException_init_String(a0)
	return this
}

func (this *FileNotFoundException) FileNotFoundException_init_String(a0 *String) {
	this.IOException.IOException_init_String(a0)
}

// ArrayStoreException is java.lang.ArrayStoreException.
type ArrayStoreException struct {
	RuntimeException
}

func (this *ArrayStoreException) JClass_ArrayStoreException() *ArrayStoreException { return this }
func (this *ArrayStoreException) JClassName() string                               { return "java.lang.ArrayStoreException" }

func ArrayStoreException_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_ArrayStoreException() *ArrayStoreException })
	return ok
}

func ArrayStoreException_fromAny(o any) *ArrayStoreException {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_ArrayStoreException() *ArrayStoreException }](o).JClass_ArrayStoreException()
}

func ArrayStoreException_new() *ArrayStoreException {
	this := &ArrayStoreException{}
	this.jself = this
	this.ArrayStoreException_init()
	return this
}

func (this *ArrayStoreException) ArrayStoreException_init() {
	this.RuntimeException.RuntimeException_init()
}

func ArrayStoreException_new_String(a0 *String) *ArrayStoreException {
	this := &ArrayStoreException{}
	this.jself = this
	this.ArrayStoreException_init_String(a0)
	return this
}

func (this *ArrayStoreException) ArrayStoreException_init_String(a0 *String) {
	this.RuntimeException.RuntimeException_init_String(a0)
}

// IllegalFormatException is java.util.IllegalFormatException.
type IllegalFormatException struct {
	IllegalArgumentException
}

func (this *IllegalFormatException) JClass_IllegalFormatException() *IllegalFormatException {
	return this
}
func (this *IllegalFormatException) JClassName() string { return "java.util.IllegalFormatException" }

func IllegalFormatException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_IllegalFormatException() *IllegalFormatException
	})
	return ok
}

func IllegalFormatException_fromAny(o any) *IllegalFormatException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_IllegalFormatException() *IllegalFormatException
	}](o).JClass_IllegalFormatException()
}

func (this *IllegalFormatException) IllegalFormatException_init() {
	this.IllegalArgumentException.IllegalArgumentException_init()
}

// UnknownFormatConversionException is java.util.UnknownFormatConversionException. Its message is computed as the JDK's
// getMessage computes it.
type UnknownFormatConversionException struct {
	IllegalFormatException
	s *String
}

func (this *UnknownFormatConversionException) JClass_UnknownFormatConversionException() *UnknownFormatConversionException {
	return this
}

func (this *UnknownFormatConversionException) JClassName() string {
	return "java.util.UnknownFormatConversionException"
}

func UnknownFormatConversionException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_UnknownFormatConversionException() *UnknownFormatConversionException
	})
	return ok
}

func UnknownFormatConversionException_fromAny(o any) *UnknownFormatConversionException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_UnknownFormatConversionException() *UnknownFormatConversionException
	}](o).JClass_UnknownFormatConversionException()
}

func UnknownFormatConversionException_new(s *String) *UnknownFormatConversionException {
	this := &UnknownFormatConversionException{}
	this.jself = this
	this.UnknownFormatConversionException_init(s)
	return this
}

func (this *UnknownFormatConversionException) UnknownFormatConversionException_init(s *String) {
	this.IllegalFormatException.IllegalFormatException_init()
	if s == nil {
		panic(NullPointerException_new())
	}
	this.s = s
	this.message = Concat(Lit("Conversion = '"), s, Lit("'"))
}

func (this *UnknownFormatConversionException) GetConversion() *String { return this.s }

// MissingFormatArgumentException is java.util.MissingFormatArgumentException. Its message is computed as the JDK's
// getMessage computes it.
type MissingFormatArgumentException struct {
	IllegalFormatException
	s *String
}

func (this *MissingFormatArgumentException) JClass_MissingFormatArgumentException() *MissingFormatArgumentException {
	return this
}

func (this *MissingFormatArgumentException) JClassName() string {
	return "java.util.MissingFormatArgumentException"
}

func MissingFormatArgumentException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_MissingFormatArgumentException() *MissingFormatArgumentException
	})
	return ok
}

func MissingFormatArgumentException_fromAny(o any) *MissingFormatArgumentException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_MissingFormatArgumentException() *MissingFormatArgumentException
	}](o).JClass_MissingFormatArgumentException()
}

func MissingFormatArgumentException_new(s *String) *MissingFormatArgumentException {
	this := &MissingFormatArgumentException{}
	this.jself = this
	this.MissingFormatArgumentException_init(s)
	return this
}

func (this *MissingFormatArgumentException) MissingFormatArgumentException_init(s *String) {
	this.IllegalFormatException.IllegalFormatException_init()
	if s == nil {
		panic(NullPointerException_new())
	}
	this.s = s
	this.message = Concat(Lit("Format specifier '"), s, Lit("'"))
}

func (this *MissingFormatArgumentException) GetFormatSpecifier() *String { return this.s }

// IllegalFormatConversionException is java.util.IllegalFormatConversionException. Its message is computed as the JDK's
// getMessage computes it.
type IllegalFormatConversionException struct {
	IllegalFormatException
	c   uint16
	arg *Class
}

func (this *IllegalFormatConversionException) JClass_IllegalFormatConversionException() *IllegalFormatConversionException {
	return this
}

func (this *IllegalFormatConversionException) JClassName() string {
	return "java.util.IllegalFormatConversionException"
}

func IllegalFormatConversionException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_IllegalFormatConversionException() *IllegalFormatConversionException
	})
	return ok
}

func IllegalFormatConversionException_fromAny(o any) *IllegalFormatConversionException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_IllegalFormatConversionException() *IllegalFormatConversionException
	}](o).JClass_IllegalFormatConversionException()
}

func IllegalFormatConversionException_new(c uint16, arg *Class) *IllegalFormatConversionException {
	this := &IllegalFormatConversionException{}
	this.jself = this
	this.IllegalFormatConversionException_init(c, arg)
	return this
}

func (this *IllegalFormatConversionException) IllegalFormatConversionException_init(c uint16, arg *Class) {
	this.IllegalFormatException.IllegalFormatException_init()
	if arg == nil {
		panic(NullPointerException_new())
	}
	this.c = c
	this.arg = arg
	this.message = Concat(StrOfChar(c), Lit(" != "), arg.GetName())
}

func (this *IllegalFormatConversionException) GetConversion() uint16 { return this.c }

func (this *IllegalFormatConversionException) GetArgumentClass() *Class { return this.arg }

// FormatFlagsConversionMismatchException is java.util.FormatFlagsConversionMismatchException. Its message is computed as the JDK's
// getMessage computes it.
type FormatFlagsConversionMismatchException struct {
	IllegalFormatException
	f *String
	c uint16
}

func (this *FormatFlagsConversionMismatchException) JClass_FormatFlagsConversionMismatchException() *FormatFlagsConversionMismatchException {
	return this
}

func (this *FormatFlagsConversionMismatchException) JClassName() string {
	return "java.util.FormatFlagsConversionMismatchException"
}

func FormatFlagsConversionMismatchException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_FormatFlagsConversionMismatchException() *FormatFlagsConversionMismatchException
	})
	return ok
}

func FormatFlagsConversionMismatchException_fromAny(o any) *FormatFlagsConversionMismatchException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_FormatFlagsConversionMismatchException() *FormatFlagsConversionMismatchException
	}](o).JClass_FormatFlagsConversionMismatchException()
}

func FormatFlagsConversionMismatchException_new(f *String, c uint16) *FormatFlagsConversionMismatchException {
	this := &FormatFlagsConversionMismatchException{}
	this.jself = this
	this.FormatFlagsConversionMismatchException_init(f, c)
	return this
}

func (this *FormatFlagsConversionMismatchException) FormatFlagsConversionMismatchException_init(f *String, c uint16) {
	this.IllegalFormatException.IllegalFormatException_init()
	if f == nil {
		panic(NullPointerException_new())
	}
	this.f = f
	this.c = c
	this.message = Concat(Lit("Conversion = "), StrOfChar(c), Lit(", Flags = "), f)
}

func (this *FormatFlagsConversionMismatchException) GetFlags() *String { return this.f }

func (this *FormatFlagsConversionMismatchException) GetConversion() uint16 { return this.c }

// IllegalFormatPrecisionException is java.util.IllegalFormatPrecisionException. Its message is computed as the JDK's
// getMessage computes it.
type IllegalFormatPrecisionException struct {
	IllegalFormatException
	p int32
}

func (this *IllegalFormatPrecisionException) JClass_IllegalFormatPrecisionException() *IllegalFormatPrecisionException {
	return this
}

func (this *IllegalFormatPrecisionException) JClassName() string {
	return "java.util.IllegalFormatPrecisionException"
}

func IllegalFormatPrecisionException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_IllegalFormatPrecisionException() *IllegalFormatPrecisionException
	})
	return ok
}

func IllegalFormatPrecisionException_fromAny(o any) *IllegalFormatPrecisionException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_IllegalFormatPrecisionException() *IllegalFormatPrecisionException
	}](o).JClass_IllegalFormatPrecisionException()
}

func IllegalFormatPrecisionException_new(p int32) *IllegalFormatPrecisionException {
	this := &IllegalFormatPrecisionException{}
	this.jself = this
	this.IllegalFormatPrecisionException_init(p)
	return this
}

func (this *IllegalFormatPrecisionException) IllegalFormatPrecisionException_init(p int32) {
	this.IllegalFormatException.IllegalFormatException_init()
	this.p = p
	this.message = StrOfInt(p)
}

func (this *IllegalFormatPrecisionException) GetPrecision() int32 { return this.p }

// IllegalFormatWidthException is java.util.IllegalFormatWidthException. Its message is computed as the JDK's
// getMessage computes it.
type IllegalFormatWidthException struct {
	IllegalFormatException
	w int32
}

func (this *IllegalFormatWidthException) JClass_IllegalFormatWidthException() *IllegalFormatWidthException {
	return this
}

func (this *IllegalFormatWidthException) JClassName() string {
	return "java.util.IllegalFormatWidthException"
}

func IllegalFormatWidthException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_IllegalFormatWidthException() *IllegalFormatWidthException
	})
	return ok
}

func IllegalFormatWidthException_fromAny(o any) *IllegalFormatWidthException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_IllegalFormatWidthException() *IllegalFormatWidthException
	}](o).JClass_IllegalFormatWidthException()
}

func IllegalFormatWidthException_new(w int32) *IllegalFormatWidthException {
	this := &IllegalFormatWidthException{}
	this.jself = this
	this.IllegalFormatWidthException_init(w)
	return this
}

func (this *IllegalFormatWidthException) IllegalFormatWidthException_init(w int32) {
	this.IllegalFormatException.IllegalFormatException_init()
	this.w = w
	this.message = StrOfInt(w)
}

func (this *IllegalFormatWidthException) GetWidth() int32 { return this.w }

// MissingFormatWidthException is java.util.MissingFormatWidthException. Its message is computed as the JDK's
// getMessage computes it.
type MissingFormatWidthException struct {
	IllegalFormatException
	s *String
}

func (this *MissingFormatWidthException) JClass_MissingFormatWidthException() *MissingFormatWidthException {
	return this
}

func (this *MissingFormatWidthException) JClassName() string {
	return "java.util.MissingFormatWidthException"
}

func MissingFormatWidthException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_MissingFormatWidthException() *MissingFormatWidthException
	})
	return ok
}

func MissingFormatWidthException_fromAny(o any) *MissingFormatWidthException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_MissingFormatWidthException() *MissingFormatWidthException
	}](o).JClass_MissingFormatWidthException()
}

func MissingFormatWidthException_new(s *String) *MissingFormatWidthException {
	this := &MissingFormatWidthException{}
	this.jself = this
	this.MissingFormatWidthException_init(s)
	return this
}

func (this *MissingFormatWidthException) MissingFormatWidthException_init(s *String) {
	this.IllegalFormatException.IllegalFormatException_init()
	if s == nil {
		panic(NullPointerException_new())
	}
	this.s = s
	this.message = s
}

func (this *MissingFormatWidthException) GetFormatSpecifier() *String { return this.s }

// DuplicateFormatFlagsException is java.util.DuplicateFormatFlagsException. Its message is computed as the JDK's
// getMessage computes it.
type DuplicateFormatFlagsException struct {
	IllegalFormatException
	f *String
}

func (this *DuplicateFormatFlagsException) JClass_DuplicateFormatFlagsException() *DuplicateFormatFlagsException {
	return this
}

func (this *DuplicateFormatFlagsException) JClassName() string {
	return "java.util.DuplicateFormatFlagsException"
}

func DuplicateFormatFlagsException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_DuplicateFormatFlagsException() *DuplicateFormatFlagsException
	})
	return ok
}

func DuplicateFormatFlagsException_fromAny(o any) *DuplicateFormatFlagsException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_DuplicateFormatFlagsException() *DuplicateFormatFlagsException
	}](o).JClass_DuplicateFormatFlagsException()
}

func DuplicateFormatFlagsException_new(f *String) *DuplicateFormatFlagsException {
	this := &DuplicateFormatFlagsException{}
	this.jself = this
	this.DuplicateFormatFlagsException_init(f)
	return this
}

func (this *DuplicateFormatFlagsException) DuplicateFormatFlagsException_init(f *String) {
	this.IllegalFormatException.IllegalFormatException_init()
	if f == nil {
		panic(NullPointerException_new())
	}
	this.f = f
	this.message = Concat(Lit("Flags = '"), f, Lit("'"))
}

func (this *DuplicateFormatFlagsException) GetFlags() *String { return this.f }

// IllegalFormatFlagsException is java.util.IllegalFormatFlagsException. Its message is computed as the JDK's
// getMessage computes it.
type IllegalFormatFlagsException struct {
	IllegalFormatException
	f *String
}

func (this *IllegalFormatFlagsException) JClass_IllegalFormatFlagsException() *IllegalFormatFlagsException {
	return this
}

func (this *IllegalFormatFlagsException) JClassName() string {
	return "java.util.IllegalFormatFlagsException"
}

func IllegalFormatFlagsException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_IllegalFormatFlagsException() *IllegalFormatFlagsException
	})
	return ok
}

func IllegalFormatFlagsException_fromAny(o any) *IllegalFormatFlagsException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_IllegalFormatFlagsException() *IllegalFormatFlagsException
	}](o).JClass_IllegalFormatFlagsException()
}

func IllegalFormatFlagsException_new(f *String) *IllegalFormatFlagsException {
	this := &IllegalFormatFlagsException{}
	this.jself = this
	this.IllegalFormatFlagsException_init(f)
	return this
}

func (this *IllegalFormatFlagsException) IllegalFormatFlagsException_init(f *String) {
	this.IllegalFormatException.IllegalFormatException_init()
	if f == nil {
		panic(NullPointerException_new())
	}
	this.f = f
	this.message = Concat(Lit("Flags = '"), f, Lit("'"))
}

func (this *IllegalFormatFlagsException) GetFlags() *String { return this.f }

// UnknownFormatFlagsException is java.util.UnknownFormatFlagsException. Its message is computed as the JDK's
// getMessage computes it.
type UnknownFormatFlagsException struct {
	IllegalFormatException
	f *String
}

func (this *UnknownFormatFlagsException) JClass_UnknownFormatFlagsException() *UnknownFormatFlagsException {
	return this
}

func (this *UnknownFormatFlagsException) JClassName() string {
	return "java.util.UnknownFormatFlagsException"
}

func UnknownFormatFlagsException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_UnknownFormatFlagsException() *UnknownFormatFlagsException
	})
	return ok
}

func UnknownFormatFlagsException_fromAny(o any) *UnknownFormatFlagsException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_UnknownFormatFlagsException() *UnknownFormatFlagsException
	}](o).JClass_UnknownFormatFlagsException()
}

func UnknownFormatFlagsException_new(f *String) *UnknownFormatFlagsException {
	this := &UnknownFormatFlagsException{}
	this.jself = this
	this.UnknownFormatFlagsException_init(f)
	return this
}

func (this *UnknownFormatFlagsException) UnknownFormatFlagsException_init(f *String) {
	this.IllegalFormatException.IllegalFormatException_init()
	if f == nil {
		panic(NullPointerException_new())
	}
	this.f = f
	this.message = Concat(Lit("Flags = "), f)
}

func (this *UnknownFormatFlagsException) GetFlags() *String { return this.f }

// IllegalFormatCodePointException is java.util.IllegalFormatCodePointException. Its message is computed as the JDK's
// getMessage computes it.
type IllegalFormatCodePointException struct {
	IllegalFormatException
	c int32
}

func (this *IllegalFormatCodePointException) JClass_IllegalFormatCodePointException() *IllegalFormatCodePointException {
	return this
}

func (this *IllegalFormatCodePointException) JClassName() string {
	return "java.util.IllegalFormatCodePointException"
}

func IllegalFormatCodePointException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_IllegalFormatCodePointException() *IllegalFormatCodePointException
	})
	return ok
}

func IllegalFormatCodePointException_fromAny(o any) *IllegalFormatCodePointException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_IllegalFormatCodePointException() *IllegalFormatCodePointException
	}](o).JClass_IllegalFormatCodePointException()
}

func IllegalFormatCodePointException_new(c int32) *IllegalFormatCodePointException {
	this := &IllegalFormatCodePointException{}
	this.jself = this
	this.IllegalFormatCodePointException_init(c)
	return this
}

func (this *IllegalFormatCodePointException) IllegalFormatCodePointException_init(c int32) {
	this.IllegalFormatException.IllegalFormatException_init()
	this.c = c
	this.message = Concat(Lit("Code point = 0x"), Integer_toHexString(c))
}

func (this *IllegalFormatCodePointException) GetCodePoint() int32 { return this.c }

// IllegalFormatArgumentIndexException is the JDK's package-private
// java.util.IllegalFormatArgumentIndexException.
type IllegalFormatArgumentIndexException struct {
	IllegalFormatException
}

func (this *IllegalFormatArgumentIndexException) JClass_IllegalFormatArgumentIndexException() *IllegalFormatArgumentIndexException {
	return this
}

func (this *IllegalFormatArgumentIndexException) JClassName() string {
	return "java.util.IllegalFormatArgumentIndexException"
}

func newIllegalFormatArgumentIndexException(index int32) *IllegalFormatArgumentIndexException {
	this := &IllegalFormatArgumentIndexException{}
	this.jself = this
	this.IllegalFormatException.IllegalFormatException_init()
	if index == -1<<31 {
		this.message = Lit("Format argument index: (not representable as int)")
	} else {
		this.message = Concat(Lit("Illegal format argument index = "), StrOfInt(index))
	}
	return this
}

// PatternSyntaxException is java.util.regex.PatternSyntaxException. Its
// message is computed as the JDK's getMessage computes it.
type PatternSyntaxException struct {
	IllegalArgumentException
	desc    *String
	pattern *String
	index   int32
}

func (this *PatternSyntaxException) JClass_PatternSyntaxException() *PatternSyntaxException {
	return this
}

func (this *PatternSyntaxException) JClassName() string {
	return "java.util.regex.PatternSyntaxException"
}

func PatternSyntaxException_isInstance(o any) bool {
	_, ok := o.(interface {
		JClass_PatternSyntaxException() *PatternSyntaxException
	})
	return ok
}

func PatternSyntaxException_fromAny(o any) *PatternSyntaxException {
	if o == nil {
		return nil
	}
	return AnyTo[interface {
		JClass_PatternSyntaxException() *PatternSyntaxException
	}](o).JClass_PatternSyntaxException()
}

func PatternSyntaxException_new(desc, regex *String, index int32) *PatternSyntaxException {
	this := &PatternSyntaxException{}
	this.jself = this
	this.PatternSyntaxException_init(desc, regex, index)
	return this
}

func (this *PatternSyntaxException) PatternSyntaxException_init(desc, regex *String, index int32) {
	this.IllegalArgumentException.IllegalArgumentException_init()
	this.desc = desc
	this.pattern = regex
	this.index = index
	parts := []*String{desc}
	if index >= 0 {
		parts = append(parts, Lit(" near index "), StrOfInt(index))
	}
	parts = append(parts, Lit("\n"), regex)
	if index >= 0 && regex != nil && index < regex.Length() {
		caret := make([]uint16, 0, index+2)
		caret = append(caret, '\n')
		for i := int32(0); i < index; i++ {
			if regex.units[i] == '\t' {
				caret = append(caret, '\t')
			} else {
				caret = append(caret, ' ')
			}
		}
		parts = append(parts, FromUnits(append(caret, '^')))
	}
	this.message = Concat(parts...)
}

func (this *PatternSyntaxException) GetDescription() *String { return this.desc }
func (this *PatternSyntaxException) GetIndex() int32         { return this.index }
func (this *PatternSyntaxException) GetPattern() *String     { return this.pattern }
