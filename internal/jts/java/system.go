package java

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// ---------------------------------------------------------------------------
// java.lang.System

// System_currentTimeMillis implements System.currentTimeMillis.
func System_currentTimeMillis() int64 {
	return time.Now().UnixMilli()
}

// processStart is the origin of System.nanoTime. It is set once at package
// initialization and only read afterwards.
var processStart = time.Now()

// System_nanoTime implements System.nanoTime: a monotonic clock with an
// arbitrary origin.
func System_nanoTime() int64 {
	return int64(time.Since(processStart))
}

// System_identityHashCode implements System.identityHashCode.
func System_identityHashCode(o any) int32 {
	return IdentityHashCode(o)
}

// System_lineSeparator implements System.lineSeparator for a Unix host.
func System_lineSeparator() *String {
	return Lit("\n")
}

// System_getProperty_String implements System.getProperty(String). Only the
// separator properties and user.dir are defined; every other property is
// unset (null), as it is in a JVM started without -D options.
func System_getProperty_String(key *String) *String {
	checkPropertyKey(key)
	switch key.Go() {
	case "line.separator":
		return Lit("\n")
	case "file.separator":
		return Lit("/")
	case "path.separator":
		return Lit(":")
	case "user.dir":
		if wd, err := os.Getwd(); err == nil {
			return FromGo(wd)
		}
	}
	return nil
}

// System_getProperty_String_String implements System.getProperty(String,
// String).
func System_getProperty_String_String(key, def *String) *String {
	if v := System_getProperty_String(key); v != nil {
		return v
	}
	return def
}

func checkPropertyKey(key *String) {
	if key == nil {
		panic(NullPointerException_new_String(Lit("key can't be null")))
	}
	if key.IsEmpty() {
		panic(IllegalArgumentException_new_String(Lit("key can't be empty")))
	}
}

// System_gc implements System.gc.
func System_gc() {
	runtime.GC()
}

// System_exit implements System.exit, which the port does not support: it
// would end the Go process.
func System_exit(status int32) {
	panic(Unsupported("System.exit(" + strconv.Itoa(int(status)) + ")"))
}

// System_setOut implements System.setOut.
func System_setOut(out *PrintStream) {
	System_out = out
}

// System_setErr implements System.setErr.
func System_setErr(err *PrintStream) {
	System_err = err
}

// System_arraycopy implements System.arraycopy. The arrays arrive as Go
// slices of any element type. Copies within one array behave as if the
// source range were first copied to a temporary array, as in Java.
func System_arraycopy(src any, srcPos int32, dst any, dstPos int32, length int32) {
	switch s := src.(type) {
	case []int32:
		if d, ok := dst.([]int32); ok {
			arraycopySame(s, srcPos, d, dstPos, length)
			return
		}
	case []float64:
		if d, ok := dst.([]float64); ok {
			arraycopySame(s, srcPos, d, dstPos, length)
			return
		}
	case []int8:
		if d, ok := dst.([]int8); ok {
			arraycopySame(s, srcPos, d, dstPos, length)
			return
		}
	case []any:
		if d, ok := dst.([]any); ok {
			arraycopySame(s, srcPos, d, dstPos, length)
			return
		}
	}
	if src == nil || dst == nil {
		panic(NullPointerException_new())
	}
	sv := reflect.ValueOf(src)
	dv := reflect.ValueOf(dst)
	if sv.Kind() != reflect.Slice {
		panic(arrayStoreError("arraycopy: source type " + className(src) + " is not an array"))
	}
	if dv.Kind() != reflect.Slice {
		panic(arrayStoreError("arraycopy: destination type " + className(dst) + " is not an array"))
	}
	st, dt := sv.Type().Elem(), dv.Type().Elem()
	sPrim, dPrim := isPrimitiveElem(st), isPrimitiveElem(dt)
	if sPrim || dPrim {
		if st != dt {
			panic(arrayStoreError("arraycopy: type mismatch: can not copy " +
				arrayTypeName(sv.Type()) + " into " + arrayTypeName(dv.Type())))
		}
	}
	checkArraycopyBounds(sv, srcPos, dv, dstPos, length)
	if length == 0 {
		return
	}
	s := sv.Slice(int(srcPos), int(srcPos+length))
	d := dv.Slice(int(dstPos), int(dstPos+length))
	if st == dt {
		reflect.Copy(d, s)
		return
	}
	// Object arrays of different component types: Java checks each element
	// against the destination's component type and stops at the first that
	// does not fit, leaving the earlier ones copied.
	vals := make([]reflect.Value, length)
	for i := range vals {
		vals[i] = s.Index(i)
	}
	for i, v := range vals {
		e, ok := storableElem(v, dt)
		if !ok {
			panic(arrayStoreError("arraycopy: element type mismatch: can not cast one of the elements of " +
				javaTypeName(sv.Type()) + " to the type of the destination array, " + javaTypeName(dt)))
		}
		d.Index(i).Set(e)
	}
}

// arraycopySame is System.arraycopy between two arrays of the same Go type.
func arraycopySame[T any](s []T, srcPos int32, d []T, dstPos int32, length int32) {
	if s == nil || d == nil {
		panic(NullPointerException_new())
	}
	if srcPos < 0 || dstPos < 0 || length < 0 ||
		int64(srcPos)+int64(length) > int64(len(s)) || int64(dstPos)+int64(length) > int64(len(d)) {
		checkArraycopyBounds(reflect.ValueOf(s), srcPos, reflect.ValueOf(d), dstPos, length)
	}
	copy(d[dstPos:dstPos+length], s[srcPos:srcPos+length])
}

func checkArraycopyBounds(sv reflect.Value, srcPos int32, dv reflect.Value, dstPos int32, length int32) {
	switch {
	case srcPos < 0:
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"arraycopy: source index %d out of bounds for %s", srcPos, arrayDesc(sv)))))
	case dstPos < 0:
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"arraycopy: destination index %d out of bounds for %s", dstPos, arrayDesc(dv)))))
	case length < 0:
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"arraycopy: length %d is negative", length))))
	case int64(srcPos)+int64(length) > int64(sv.Len()):
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"arraycopy: last source index %d out of bounds for %s", int64(srcPos)+int64(length), arrayDesc(sv)))))
	case int64(dstPos)+int64(length) > int64(dv.Len()):
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"arraycopy: last destination index %d out of bounds for %s", int64(dstPos)+int64(length), arrayDesc(dv)))))
	}
}

// arrayDesc describes an array in HotSpot's arraycopy messages: "int[5]" for
// a primitive array and "object array[5]" otherwise.
func arrayDesc(v reflect.Value) string {
	et := v.Type().Elem()
	if isPrimitiveElem(et) {
		return primitiveName(et) + "[" + strconv.Itoa(v.Len()) + "]"
	}
	return "object array[" + strconv.Itoa(v.Len()) + "]"
}

// arrayTypeName names an array type as HotSpot's arraycopy type-mismatch
// messages do: "int[]" for a primitive array and "object array[]" otherwise.
func arrayTypeName(t reflect.Type) string {
	if isPrimitiveElem(t.Elem()) {
		return primitiveName(t.Elem()) + "[]"
	}
	return "object array[]"
}

func isPrimitiveElem(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Uint16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func primitiveName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int8:
		return "byte"
	case reflect.Int16:
		return "short"
	case reflect.Uint16:
		return "char"
	case reflect.Int32:
		return "int"
	case reflect.Int64:
		return "long"
	case reflect.Float32:
		return "float"
	case reflect.Float64:
		return "double"
	}
	return t.String()
}

// javaTypeName approximates the Java name of the type of a Go slice or
// slice element, for exception messages.
func javaTypeName(t reflect.Type) string {
	switch {
	case t.Kind() == reflect.Slice:
		if isPrimitiveElem(t.Elem()) {
			return "[" + primitiveName(t.Elem()) + "[]"
		}
		return javaTypeName(t.Elem()) + "[]"
	case t.Kind() == reflect.Interface && t.NumMethod() == 0:
		return "java.lang.Object"
	case t == reflect.TypeOf((*String)(nil)):
		return "java.lang.String"
	}
	if z, ok := reflect.Zero(t).Interface().(javaClassName); ok && t.Kind() == reflect.Ptr {
		return z.JClassName()
	}
	return t.String()
}

// storableElem converts v, an element of an object array, for storing in an
// object array with element type dt. It reports false where Java would throw
// ArrayStoreException.
func storableElem(v reflect.Value, dt reflect.Type) (reflect.Value, bool) {
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return reflect.Zero(dt), true
		}
		v = v.Elem()
	}
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Slice || v.Kind() == reflect.Func) && v.IsNil() {
		return reflect.Zero(dt), true
	}
	if v.Type().AssignableTo(dt) {
		return v, true
	}
	if dt.Kind() == reflect.Ptr {
		if m := v.MethodByName("JClass_" + dt.Elem().Name()); m.IsValid() && m.Type().NumIn() == 0 {
			if r := m.Call(nil)[0]; r.Type() == dt {
				return r, true
			}
		}
	}
	return reflect.Value{}, false
}

// ---------------------------------------------------------------------------
// java.lang.Runtime

// Runtime is java.lang.Runtime. Memory figures come from the Go runtime.
type Runtime struct {
	jhash int32
}

func (this *Runtime) JClass_Runtime() *Runtime { return this }
func (this *Runtime) JClassName() string       { return "java.lang.Runtime" }
func (this *Runtime) JIdentityHash() *int32    { return &this.jhash }

func Runtime_isInstance(o any) bool { _, ok := o.(*Runtime); return ok }

func Runtime_fromAny(o any) *Runtime { return AnyTo[*Runtime](o) }

// theRuntime is Runtime.getRuntime()'s singleton. It has no mutable state.
var theRuntime = &Runtime{}

// Runtime_getRuntime implements Runtime.getRuntime.
func Runtime_getRuntime() *Runtime { return theRuntime }

func (this *Runtime) AvailableProcessors() int32 { return int32(runtime.NumCPU()) }

func (this *Runtime) FreeMemory() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.HeapIdle)
}

func (this *Runtime) TotalMemory() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.HeapSys)
}

func (this *Runtime) MaxMemory() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Sys)
}

func (this *Runtime) Gc() { runtime.GC() }

func (this *Runtime) Exit_int(status int32) { System_exit(status) }

// ---------------------------------------------------------------------------
// Functional interfaces

// Runnable is java.lang.Runnable.
type Runnable interface {
	JIface_Runnable()
	Run()
}

// Consumer is java.util.function.Consumer.
type Consumer interface {
	JIface_Consumer()
	Accept_Object(any)
}

// Function is java.util.function.Function.
type Function interface {
	JIface_Function()
	Apply_Object(any) any
}

// Supplier is java.util.function.Supplier.
type Supplier interface {
	JIface_Supplier()
	Get() any
}

func Runnable_isInstance(o any) bool { return InstanceOf[Runnable](o) }
func Consumer_isInstance(o any) bool { return InstanceOf[Consumer](o) }
func Function_isInstance(o any) bool { return InstanceOf[Function](o) }
func Supplier_isInstance(o any) bool { return InstanceOf[Supplier](o) }

// RunnableFunc adapts a function literal (a translated lambda or method
// reference) to Runnable.
type RunnableFunc func()

func (f RunnableFunc) JIface_Runnable() {}
func (f RunnableFunc) Run()             { f() }

// ConsumerFunc adapts a function literal to Consumer.
type ConsumerFunc func(a any)

func (f ConsumerFunc) JIface_Consumer()    {}
func (f ConsumerFunc) Accept_Object(a any) { f(a) }

// FunctionFunc adapts a function literal to Function.
type FunctionFunc func(a any) any

func (f FunctionFunc) JIface_Function()       {}
func (f FunctionFunc) Apply_Object(a any) any { return f(a) }

// SupplierFunc adapts a function literal to Supplier.
type SupplierFunc func() any

func (f SupplierFunc) JIface_Supplier() {}
func (f SupplierFunc) Get() any         { return f() }

// ---------------------------------------------------------------------------
// java.lang.Thread

// Thread is java.lang.Thread. A started thread runs in its own goroutine.
type Thread struct {
	jself   any
	jhash   int32
	target  Runnable
	name    *String
	id      int64
	mu      sync.Mutex
	started bool
	done    chan struct{}
}

func (this *Thread) JSelf() any             { return this.jself }
func (this *Thread) JSetSelf(self any)      { this.jself = self }
func (this *Thread) JClass_Thread() *Thread { return this }
func (this *Thread) JClassName() string     { return "java.lang.Thread" }
func (this *Thread) JIdentityHash() *int32  { return &this.jhash }
func (this *Thread) JIface_Runnable()       {}

func Thread_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_Thread() *Thread })
	return ok
}

func Thread_fromAny(o any) *Thread {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_Thread() *Thread }](o).JClass_Thread()
}

var (
	threadNumMu sync.Mutex
	// threadNum numbers threads created without a name, as Thread's
	// threadInitNumber does.
	threadNum int32
)

func nextThreadName() *String {
	threadNumMu.Lock()
	defer threadNumMu.Unlock()
	n := threadNum
	threadNum++
	return FromGo("Thread-" + strconv.Itoa(int(n)))
}

// lastThreadID numbers threads for Thread.toString; the main thread is 1.
var lastThreadID int64 = 1

func newThread() *Thread {
	this := &Thread{}
	this.jself = this
	return this
}

func Thread_new() *Thread {
	this := newThread()
	this.Thread_init()
	return this
}

func Thread_new_Runnable(target Runnable) *Thread {
	this := newThread()
	this.Thread_init_Runnable(target)
	return this
}

func Thread_new_String(name *String) *Thread {
	this := newThread()
	this.Thread_init_String(name)
	return this
}

func Thread_new_Runnable_String(target Runnable, name *String) *Thread {
	this := newThread()
	this.Thread_init_Runnable_String(target, name)
	return this
}

func (this *Thread) Thread_init() {
	this.Thread_init_Runnable_String(nil, nextThreadName())
}

func (this *Thread) Thread_init_Runnable(target Runnable) {
	this.Thread_init_Runnable_String(target, nextThreadName())
}

func (this *Thread) Thread_init_String(name *String) {
	this.Thread_init_Runnable_String(nil, name)
}

func (this *Thread) Thread_init_Runnable_String(target Runnable, name *String) {
	if name == nil {
		panic(NullPointerException_new_String(Lit("'name' is null")))
	}
	this.target = target
	this.name = name
	this.done = make(chan struct{})
	this.id = atomic.AddInt64(&lastThreadID, 1)
}

// Run implements Thread.run: it runs the target, if any.
func (this *Thread) Run() {
	if this.target != nil {
		this.target.Run()
	}
}

// Start implements Thread.start: run (as overridden by the object's class)
// executes in a new goroutine. An exception that escapes run is reported on
// System.err and ends only that thread, as in Java.
func (this *Thread) Start() {
	this.mu.Lock()
	if this.started {
		this.mu.Unlock()
		panic(IllegalThreadStateException_new())
	}
	this.started = true
	this.mu.Unlock()
	run := this.Run
	if r, ok := this.jself.(Runnable); ok {
		run = r.Run
	}
	go func() {
		defer close(this.done)
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(UnsupportedError); ok {
					panic(r)
				}
				ex := Throwable_ofPanic(r)
				System_err.Print_String(Concat(Lit("Exception in thread \""), this.name, Lit("\" ")))
				Throwable_fromAny(ex).PrintStackTrace_PrintStream(System_err)
			}
		}()
		run()
	}()
}

// Join implements Thread.join: it waits for a started thread to finish.
func (this *Thread) Join() {
	this.Join_long(0)
}

// Join_long implements Thread.join(long); a timeout of 0 waits forever.
func (this *Thread) Join_long(millis int64) {
	if millis < 0 {
		panic(IllegalArgumentException_new_String(Lit("timeout value is negative")))
	}
	this.mu.Lock()
	started := this.started
	this.mu.Unlock()
	if !started {
		return
	}
	if millis == 0 {
		<-this.done
		return
	}
	select {
	case <-this.done:
	case <-time.After(time.Duration(millis) * time.Millisecond):
	}
}

// Join_long_int implements Thread.join(long, int).
func (this *Thread) Join_long_int(millis int64, nanos int32) {
	if millis < 0 {
		panic(IllegalArgumentException_new_String(Lit("timeout value is negative")))
	}
	if nanos < 0 || nanos > 999999 {
		panic(IllegalArgumentException_new_String(Lit("nanosecond timeout value out of range")))
	}
	if nanos > 0 && millis < 1<<63-1 {
		millis++
	}
	this.Join_long(millis)
}

func (this *Thread) GetName() *String { return this.name }

func (this *Thread) SetName_String(name *String) {
	if name == nil {
		panic(NullPointerException_new_String(Lit("name cannot be null")))
	}
	this.name = name
}

// ToString implements Thread.toString in JDK 21's format. The thread IDs
// differ from the JVM's, which also count the JVM's own threads.
func (this *Thread) ToString() *String {
	group := "main"
	select {
	case <-this.done:
		group = ""
	default:
	}
	return Concat(Lit("Thread[#"), StrOfLong(this.id), Lit(","), this.name, FromGo(",5,"+group+"]"))
}

// Thread_sleep_long implements Thread.sleep(long).
func Thread_sleep_long(millis int64) {
	if millis < 0 {
		panic(IllegalArgumentException_new_String(Lit("timeout value is negative")))
	}
	time.Sleep(time.Duration(millis) * time.Millisecond)
}

// Thread_sleep_long_int implements Thread.sleep(long, int).
func Thread_sleep_long_int(millis int64, nanos int32) {
	if millis < 0 {
		panic(IllegalArgumentException_new_String(Lit("timeout value is negative")))
	}
	if nanos < 0 || nanos > 999999 {
		panic(IllegalArgumentException_new_String(Lit("nanosecond timeout value out of range")))
	}
	time.Sleep(time.Duration(millis)*time.Millisecond + time.Duration(nanos))
}

// ---------------------------------------------------------------------------
// java.lang.ref

// SoftReference is java.lang.ref.SoftReference. The referent is held
// strongly: the JVM clears soft references only under memory pressure, and
// code using them must already cope with either outcome.
type SoftReference struct {
	jhash    int32
	referent any
}

func (this *SoftReference) JClass_SoftReference() *SoftReference { return this }
func (this *SoftReference) JClassName() string                   { return "java.lang.ref.SoftReference" }
func (this *SoftReference) JIdentityHash() *int32                { return &this.jhash }

func SoftReference_isInstance(o any) bool { _, ok := o.(*SoftReference); return ok }

func SoftReference_fromAny(o any) *SoftReference { return AnyTo[*SoftReference](o) }

func SoftReference_new_Object(referent any) *SoftReference {
	return &SoftReference{referent: referent}
}

func (this *SoftReference) Get() any { return this.referent }
func (this *SoftReference) Clear()   { this.referent = nil }

// WeakReference is java.lang.ref.WeakReference, holding its referent
// strongly like SoftReference.
type WeakReference struct {
	jhash    int32
	referent any
}

func (this *WeakReference) JClass_WeakReference() *WeakReference { return this }
func (this *WeakReference) JClassName() string                   { return "java.lang.ref.WeakReference" }
func (this *WeakReference) JIdentityHash() *int32                { return &this.jhash }

func WeakReference_isInstance(o any) bool { _, ok := o.(*WeakReference); return ok }

func WeakReference_fromAny(o any) *WeakReference { return AnyTo[*WeakReference](o) }

func WeakReference_new_Object(referent any) *WeakReference {
	return &WeakReference{referent: referent}
}

func (this *WeakReference) Get() any { return this.referent }
func (this *WeakReference) Clear()   { this.referent = nil }

// ---------------------------------------------------------------------------
// java.util.Date

// Date is java.util.Date, supporting the millisecond-instant API.
type Date struct {
	millis int64
}

func (this *Date) JClass_Date() *Date { return this }
func (this *Date) JClassName() string { return "java.util.Date" }

func Date_isInstance(o any) bool { _, ok := o.(*Date); return ok }

func Date_fromAny(o any) *Date { return AnyTo[*Date](o) }

func Date_new() *Date { return &Date{millis: System_currentTimeMillis()} }

func Date_new_long(millis int64) *Date { return &Date{millis: millis} }

func (this *Date) GetTime() int64 { return this.millis }

func (this *Date) SetTime_long(millis int64) { this.millis = millis }

func (this *Date) Before_Date(when *Date) bool { return this.millis < when.millis }

func (this *Date) After_Date(when *Date) bool { return this.millis > when.millis }

func (this *Date) CompareTo_Object(o any) int32 {
	return Long_compare(this.millis, AnyTo[*Date](o).millis)
}

func (this *Date) Equals_Object(o any) bool {
	d, ok := o.(*Date)
	return ok && d != nil && d.millis == this.millis
}

func (this *Date) HashCode() int32 {
	return int32(this.millis) ^ int32(this.millis>>32)
}

func (this *Date) Clone() any { return &Date{millis: this.millis} }

// ToString implements Date.toString ("EEE MMM dd HH:mm:ss zzz yyyy") in the
// local time zone.
func (this *Date) ToString() *String {
	t := time.UnixMilli(this.millis)
	return FromGo(t.Format("Mon Jan 02 15:04:05 MST ") + strconv.Itoa(t.Year()))
}

// arrayStoreError returns the exception System.arraycopy throws when the
// arrays' element types are incompatible.
func arrayStoreError(msg string) any {
	return ArrayStoreException_new_String(FromGo(msg))
}
