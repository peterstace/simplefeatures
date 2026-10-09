package java

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"unsafe"
)

// identityHashSeq is the source of identity hash codes. HotSpot run with
// -XX:hashCode=3 assigns identity hash codes from a global counter in the
// order objects are first hashed; this counter reproduces that so that
// iteration over identity-hashed collections matches the oracle.
var identityHashSeq int32

// SetIdentityHashSeq sets the next identity hash code to be assigned.
func SetIdentityHashSeq(next int32) {
	atomic.StoreInt32(&identityHashSeq, next-1)
}

func nextIdentityHash() int32 {
	return atomic.AddInt32(&identityHashSeq, 1)
}

type identityHashed interface {
	JIdentityHash() *int32
}

var (
	sideHashMu sync.Mutex
	// sideHashes holds identity hash codes for objects without a jhash_
	// field. It keeps those objects alive; it is rarely used.
	sideHashes = map[unsafe.Pointer]int32{}
)

// IdentityHashCode implements System.identityHashCode.
func IdentityHashCode(o any) int32 {
	if o == nil {
		return 0
	}
	if h, ok := o.(identityHashed); ok {
		p := h.JIdentityHash()
		if v := atomic.LoadInt32(p); v != 0 {
			return v
		}
		v := nextIdentityHash()
		if atomic.CompareAndSwapInt32(p, 0, v) {
			return v
		}
		return atomic.LoadInt32(p)
	}
	key := identityKey(o)
	sideHashMu.Lock()
	defer sideHashMu.Unlock()
	v, ok := sideHashes[key]
	if !ok {
		v = nextIdentityHash()
		sideHashes[key] = v
	}
	return v
}

// eface is the layout of a Go interface value.
type eface struct {
	typ, data unsafe.Pointer
}

// identityKey returns the address that identifies a reference value: the
// pointer, or for a slice (Java array) its backing array.
func identityKey(o any) unsafe.Pointer {
	if v := reflect.ValueOf(o); v.Kind() == reflect.Slice {
		return unsafe.Pointer(v.Pointer())
	}
	return (*eface)(unsafe.Pointer(&o)).data
}

type javaEquals interface {
	Equals_Object(any) bool
}

type javaHashCode interface {
	HashCode() int32
}

type javaToString interface {
	ToString() *String
}

type javaClassName interface {
	JClassName() string
}

// Object_Equals_Object implements a virtual call of equals(Object).
func Object_Equals_Object(a, b any) bool {
	switch x := a.(type) {
	case nil:
		panic(NullPointerException_new())
	case javaEquals:
		return x.Equals_Object(b)
	case int32, int64, int16, int8, uint16, bool:
		return a == b
	case float64:
		y, ok := b.(float64)
		return ok && Double_doubleToLongBits(x) == Double_doubleToLongBits(y)
	case float32:
		y, ok := b.(float32)
		return ok && Float_floatToIntBits(x) == Float_floatToIntBits(y)
	}
	return Same(a, b)
}

// Object_HashCode implements a virtual call of hashCode().
func Object_HashCode(a any) int32 {
	switch x := a.(type) {
	case nil:
		panic(NullPointerException_new())
	case javaHashCode:
		return x.HashCode()
	case int32:
		return x
	case int64:
		return int32(x ^ int64(uint64(x)>>32))
	case int16:
		return int32(x)
	case int8:
		return int32(x)
	case uint16:
		return int32(x)
	case bool:
		if x {
			return 1231
		}
		return 1237
	case float64:
		return Double_hashCode_double(x)
	case float32:
		return Float_floatToIntBits(x)
	}
	return IdentityHashCode(a)
}

// Object_ToString implements a virtual call of toString().
func Object_ToString(a any) *String {
	if t, ok := a.(interface{ JClass_Throwable() *Throwable }); ok {
		return throwableToString(className(a), t.JClass_Throwable())
	}
	switch x := a.(type) {
	case nil:
		panic(NullPointerException_new())
	case javaToString:
		return x.ToString()
	case int32:
		return StrOfInt(x)
	case int64:
		return StrOfLong(x)
	case int16:
		return StrOfInt(int32(x))
	case int8:
		return StrOfInt(int32(x))
	case uint16:
		return StrOfChar(x)
	case bool:
		return StrOfBool(x)
	case float64:
		return StrOfDouble(x)
	case float32:
		return StrOfFloat(x)
	}
	return ObjectImpl_ToString(a)
}

// Object_GetClass implements getClass().
func Object_GetClass(a any) *Class {
	if a == nil {
		panic(NullPointerException_new())
	}
	return ClassOf(className(a))
}

// ObjectImpl_Equals_Object is java.lang.Object's equals.
func ObjectImpl_Equals_Object(a, b any) bool {
	return Same(a, b)
}

// ObjectImpl_HashCode is java.lang.Object's hashCode.
func ObjectImpl_HashCode(a any) int32 {
	return IdentityHashCode(a)
}

// ObjectImpl_ToString is java.lang.Object's toString.
func ObjectImpl_ToString(a any) *String {
	return FromGo(className(a) + "@" + strconv.FormatUint(uint64(uint32(IdentityHashCode(a))), 16))
}

// ObjectImpl_GetClass is java.lang.Object's getClass.
func ObjectImpl_GetClass(a any) *Class {
	return Object_GetClass(a)
}

// ObjectImpl_Clone is java.lang.Object's clone: a shallow copy of the whole
// object, made by the JShallowCopy method that every concrete generated class
// has.
func ObjectImpl_Clone(a any) any {
	c, ok := a.(interface{ JShallowCopy() any })
	if !ok {
		panic(Unsupported(fmt.Sprintf("Object.clone of %T", a)))
	}
	return c.JShallowCopy()
}

// Same reports whether a and b are the same Java object (reference equality).
// It compares the dynamic types and data words of the interfaces, so it does
// not panic on uncomparable values such as slices (Java arrays).
func Same(a, b any) bool {
	ea, eb := (*eface)(unsafe.Pointer(&a)), (*eface)(unsafe.Pointer(&b))
	if ea.typ != eb.typ {
		return false
	}
	if ea.data == eb.data {
		return true
	}
	// A slice (Java array) in an interface is a pointer to a copy of its
	// header, so the same array can be held at different addresses.
	if va := reflect.ValueOf(a); va.Kind() == reflect.Slice {
		vb := reflect.ValueOf(b)
		return va.Pointer() == vb.Pointer() && va.Len() == vb.Len()
	}
	return false
}

func className(o any) string {
	switch x := o.(type) {
	case nil:
		return "null"
	case javaClassName:
		return x.JClassName()
	case int32:
		return "java.lang.Integer"
	case int64:
		return "java.lang.Long"
	case int16:
		return "java.lang.Short"
	case int8:
		return "java.lang.Byte"
	case uint16:
		return "java.lang.Character"
	case bool:
		return "java.lang.Boolean"
	case float64:
		return "java.lang.Double"
	case float32:
		return "java.lang.Float"
	}
	return fmt.Sprintf("%T", o)
}

// Class is java.lang.Class, supporting only name queries.
type Class struct {
	name string
}

func (c *Class) GetName() *String {
	return FromGo(c.name)
}

func (c *Class) GetSimpleName() *String {
	n := c.name
	for i := len(n) - 1; i >= 0; i-- {
		if n[i] == '.' || n[i] == '$' {
			return FromGo(n[i+1:])
		}
	}
	return FromGo(n)
}

func (c *Class) Equals_Object(o any) bool {
	d, ok := o.(*Class)
	return ok && d.name == c.name
}

func (c *Class) HashCode() int32 {
	return FromGo(c.name).HashCode()
}

func (c *Class) ToString() *String {
	return FromGo("class " + c.name)
}

func (c *Class) JClassName() string { return "java.lang.Class" }

// plainObject is an instance of java.lang.Object itself.
type plainObject struct {
	jhash_ int32
}

func (o *plainObject) JIdentityHash() *int32 { return &o.jhash_ }
func (o *plainObject) JClassName() string    { return "java.lang.Object" }

// Object_new implements new Object().
func Object_new() any {
	return &plainObject{}
}

var (
	classMu sync.Mutex
	// classes holds the canonical Class object for each class name, since
	// Java compares Class objects by identity.
	classes = map[string]*Class{}
)

// ClassOf returns the Class object for the class with the given binary name.
func ClassOf(name string) *Class {
	classMu.Lock()
	defer classMu.Unlock()
	c, ok := classes[name]
	if !ok {
		c = &Class{name: name}
		classes[name] = c
	}
	return c
}

var (
	monitorsMu sync.Mutex
	// monitors holds the monitors of runtime objects used as locks by
	// synchronized blocks. Generated classes that are locked have their own
	// monitor field instead. Entries keep their objects alive.
	monitors = map[unsafe.Pointer]*sync.Mutex{}
)

// MonitorOf returns the monitor of o, for a synchronized block. Objects that
// carry a monitor return it from JMonitor.
func MonitorOf(o any) *sync.Mutex {
	if o == nil {
		panic(NullPointerException_new())
	}
	if m, ok := o.(interface{ JMonitor() *sync.Mutex }); ok {
		return m.JMonitor()
	}
	key := identityKey(o)
	monitorsMu.Lock()
	defer monitorsMu.Unlock()
	m, ok := monitors[key]
	if !ok {
		m = &sync.Mutex{}
		monitors[key] = m
	}
	return m
}

// Class_forName_String implements Class.forName. Classes are identified by
// name only, so it never throws ClassNotFoundException.
func Class_forName_String(name *String) *Class {
	return ClassOf(name.Go())
}
