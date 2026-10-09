package java

import (
	"reflect"
	"sync"
	"unsafe"
)

// Iterator is java.util.Iterator.
type Iterator interface {
	HasNext() bool
	Next() any
	Remove()
}

// Iterable is java.lang.Iterable.
type Iterable interface {
	Iterator() Iterator
}

// Collection is java.util.Collection.
type Collection interface {
	JIface_Collection()
	Iterator() Iterator
	Size() int32
	IsEmpty() bool
	Contains_Object(any) bool
	Add_Object(any) bool
	Remove_Object(any) bool
	AddAll_Collection(Collection) bool
	ContainsAll_Collection(Collection) bool
	RemoveAll_Collection(Collection) bool
	RetainAll_Collection(Collection) bool
	Clear()
	ToArray() []any
	ToArray_ObjectArr(any) any
}

// List is java.util.List.
type List interface {
	Collection
	JIface_List()
	Get_int(int32) any
	Set_int_Object(int32, any) any
	Add_int_Object(int32, any)
	Remove_int(int32) any
	IndexOf_Object(any) int32
	LastIndexOf_Object(any) int32
	SubList_int_int(int32, int32) List
	Sort_Comparator(Comparator)
	AddAll_int_Collection(int32, Collection) bool
	ListIterator() ListIterator
	ListIterator_int(int32) ListIterator
}

// ListIterator is java.util.ListIterator.
type ListIterator interface {
	Iterator
	HasPrevious() bool
	Previous() any
	NextIndex() int32
	PreviousIndex() int32
	Set_Object(any)
	Add_Object(any)
}

// Queue is java.util.Queue.
type Queue interface {
	Collection
	JIface_Queue()
	Offer_Object(any) bool
	Poll() any
	Peek() any
	Remove() any
	Element() any
}

// Deque is java.util.Deque.
type Deque interface {
	Queue
	JIface_Deque()
	AddFirst_Object(any)
	AddLast_Object(any)
	OfferFirst_Object(any) bool
	OfferLast_Object(any) bool
	RemoveFirst() any
	RemoveLast() any
	PollFirst() any
	PollLast() any
	GetFirst() any
	GetLast() any
	PeekFirst() any
	PeekLast() any
	Push_Object(any)
	Pop() any
	DescendingIterator() Iterator
	RemoveFirstOccurrence_Object(any) bool
	RemoveLastOccurrence_Object(any) bool
}

// Map is java.util.Map.
type Map interface {
	JIface_Map()
	Size() int32
	IsEmpty() bool
	ContainsKey_Object(any) bool
	ContainsValue_Object(any) bool
	Get_Object(any) any
	Put_Object_Object(any, any) any
	Remove_Object(any) any
	PutAll_Map(Map)
	Clear()
	KeySet() Set
	Values() Collection
	EntrySet() Set
	GetOrDefault_Object_Object(any, any) any
	PutIfAbsent_Object_Object(any, any) any
}

// Map_Entry is java.util.Map.Entry.
type Map_Entry interface {
	JIface_Map_Entry()
	GetKey() any
	GetValue() any
	SetValue_Object(any) any
}

// Set is java.util.Set.
type Set interface {
	Collection
	JIface_Set()
}

// SortedSet is java.util.SortedSet, without the subset views.
type SortedSet interface {
	Set
	JIface_SortedSet()
	Comparator() Comparator
	First() any
	Last() any
}

// NavigableSet is java.util.NavigableSet, without the subset views.
type NavigableSet interface {
	SortedSet
	JIface_NavigableSet()
	Lower_Object(any) any
	Floor_Object(any) any
	Ceiling_Object(any) any
	Higher_Object(any) any
	PollFirst() any
	PollLast() any
	DescendingIterator() Iterator
}

// SortedMap is java.util.SortedMap, without the submap views.
type SortedMap interface {
	Map
	JIface_SortedMap()
	Comparator() Comparator
	FirstKey() any
	LastKey() any
}

// NavigableMap is java.util.NavigableMap, without the submap views.
type NavigableMap interface {
	SortedMap
	JIface_NavigableMap()
	LowerEntry_Object(any) Map_Entry
	LowerKey_Object(any) any
	FloorEntry_Object(any) Map_Entry
	FloorKey_Object(any) any
	CeilingEntry_Object(any) Map_Entry
	CeilingKey_Object(any) any
	HigherEntry_Object(any) Map_Entry
	HigherKey_Object(any) any
	FirstEntry() Map_Entry
	LastEntry() Map_Entry
	PollFirstEntry() Map_Entry
	PollLastEntry() Map_Entry
	NavigableKeySet() NavigableSet
}

// RandomAccess is java.util.RandomAccess.
type RandomAccess interface {
	JIface_RandomAccess()
}

// Comparator is java.util.Comparator.
type Comparator interface {
	Compare_Object_Object(a, b any) int32
}

func Collection_isInstance(o any) bool { return InstanceOf[Collection](o) }

func List_isInstance(o any) bool { return InstanceOf[List](o) }

func Set_isInstance(o any) bool { return InstanceOf[Set](o) }

func Map_isInstance(o any) bool { return InstanceOf[Map](o) }

func Queue_isInstance(o any) bool { return InstanceOf[Queue](o) }

func Deque_isInstance(o any) bool { return InstanceOf[Deque](o) }

func Map_Entry_isInstance(o any) bool { return InstanceOf[Map_Entry](o) }

func Comparator_isInstance(o any) bool { return InstanceOf[Comparator](o) }

// Comparable_CompareTo_Object implements a virtual call of compareTo on a
// Comparable.
func Comparable_CompareTo_Object(a, b any) int32 {
	switch x := a.(type) {
	case nil:
		panic(NullPointerException_new())
	case interface{ CompareTo_Object(any) int32 }:
		return x.CompareTo_Object(b)
	case int32:
		return Integer_compare(x, boxedArg[int32](b, "java.lang.Integer"))
	case int64:
		return Long_compare(x, boxedArg[int64](b, "java.lang.Long"))
	case float64:
		return Double_compare(x, boxedArg[float64](b, "java.lang.Double"))
	case float32:
		return floatCompare(x, boxedArg[float32](b, "java.lang.Float"))
	case int16:
		return int32(x) - int32(boxedArg[int16](b, "java.lang.Short"))
	case int8:
		return int32(x) - int32(boxedArg[int8](b, "java.lang.Byte"))
	case uint16:
		return int32(x) - int32(boxedArg[uint16](b, "java.lang.Character"))
	case bool:
		y := boxedArg[bool](b, "java.lang.Boolean")
		switch {
		case x == y:
			return 0
		case x:
			return 1
		}
		return -1
	}
	panic(ClassCastError(a, "java.lang.Comparable"))
}

// boxedArg converts the argument of a boxed type's compareTo, throwing the
// exceptions that Java's would for null and for another type.
func boxedArg[T any](b any, class string) T {
	if b == nil {
		panic(NullPointerException_new())
	}
	v, ok := b.(T)
	if !ok {
		panic(ClassCastError(b, class))
	}
	return v
}

// floatCompare is Float.compare.
func floatCompare(a, b float32) int32 {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	ab, bb := Float_floatToIntBits(a), Float_floatToIntBits(b)
	switch {
	case ab == bb:
		return 0
	case ab < bb:
		return -1
	}
	return 1
}

func Comparable_isInstance(o any) bool {
	switch o.(type) {
	case interface{ CompareTo_Object(any) int32 }, int32, int64, float64, float32, int16, int8, uint16, bool:
		return true
	}
	return false
}

// requireComparable implements the cast to Comparable that Java's natural
// ordering code performs before comparing.
func requireComparable(o any) {
	if o == nil {
		panic(NullPointerException_new())
	}
	if !Comparable_isInstance(o) {
		panic(ClassCastError(o, "java.lang.Comparable"))
	}
}

// compareWith compares a and b with c, or by natural ordering if c is nil.
func compareWith(c Comparator, a, b any) int32 {
	if c == nil {
		return Comparable_CompareTo_Object(a, b)
	}
	return c.Compare_Object_Object(a, b)
}

// sameRef implements Java's == on references: a and b have the same dynamic
// type and the same data word, so pointers compare by address and boxed
// primitives are the same only if they are the same box.
func sameRef(a, b any) bool {
	type eface struct {
		typ, data unsafe.Pointer
	}
	ea := (*eface)(unsafe.Pointer(&a))
	eb := (*eface)(unsafe.Pointer(&b))
	return ea.typ == eb.typ && ea.data == eb.data
}

// objectsEqual is (o == null ? e == null : o.equals(e)), the test that most
// collection methods apply to an argument o and an element e.
func objectsEqual(o, e any) bool {
	if o == nil {
		return e == nil
	}
	return Object_Equals_Object(o, e)
}

// javaObjectsEquals is java.util.Objects.equals, which tests identity before
// calling equals.
func javaObjectsEquals(a, b any) bool {
	return sameRef(a, b) || a != nil && Object_Equals_Object(a, b)
}

// hashOrZero is (o == null ? 0 : o.hashCode()).
func hashOrZero(o any) int32 {
	if o == nil {
		return 0
	}
	return Object_HashCode(o)
}

// catchClassCastOrNull runs f, returning false if it throws a
// ClassCastException or NullPointerException, as AbstractSet.equals and
// AbstractMap.equals do.
func catchClassCastOrNull(f func() bool) (result bool) {
	defer func() {
		if r := recover(); r != nil {
			ex := Throwable_ofPanic(r)
			if ClassCastException_isInstance(ex) || NullPointerException_isInstance(ex) {
				result = false
				return
			}
			panic(ex)
		}
	}()
	return f()
}

// ---------------------------------------------------------------------------
// Exceptions thrown by the collection classes

// checkIndex implements Objects.checkIndex.
func checkIndex(index, length int32) {
	if index < 0 || index >= length {
		panic(IndexOutOfBoundsException_new_String(FromGo(
			"Index " + StrOfInt(index).Go() + " out of bounds for length " + StrOfInt(length).Go())))
	}
}

// outOfBounds returns an IndexOutOfBoundsException with the message format
// "Index: i, Size: n" used by the older collection classes.
func outOfBounds(index, size int32) any {
	return IndexOutOfBoundsException_new_String(FromGo(
		"Index: " + StrOfInt(index).Go() + ", Size: " + StrOfInt(size).Go()))
}

// subListRangeCheck is AbstractList.subListRangeCheck.
func subListRangeCheck(from, to, size int32) {
	if from < 0 {
		panic(IndexOutOfBoundsException_new_String(FromGo("fromIndex = " + StrOfInt(from).Go())))
	}
	if to > size {
		panic(IndexOutOfBoundsException_new_String(FromGo("toIndex = " + StrOfInt(to).Go())))
	}
	if from > to {
		panic(IllegalArgumentException_new_String(FromGo(
			"fromIndex(" + StrOfInt(from).Go() + ") > toIndex(" + StrOfInt(to).Go() + ")")))
	}
}

func unsupported() any {
	return UnsupportedOperationException_new()
}

// iteratorRemoveUnsupported is the exception thrown by Iterator.remove's
// default implementation.
func iteratorRemoveUnsupported() any {
	return UnsupportedOperationException_new_String(Lit("remove"))
}

// ---------------------------------------------------------------------------
// Java arrays held in an any

// fillArray implements Collection.toArray(T[]) for a collection holding elems.
func fillArray(a any, elems []any) any {
	if s, ok := a.([]any); ok {
		if len(s) < len(elems) {
			return append([]any(nil), elems...)
		}
		copy(s, elems)
		if len(s) > len(elems) {
			s[len(elems)] = nil
		}
		return s
	}
	av := reflect.ValueOf(a)
	if av.Kind() != reflect.Slice {
		panic(NullPointerException_new())
	}
	out := av
	if av.Len() < len(elems) {
		out = reflect.MakeSlice(av.Type(), len(elems), len(elems))
	}
	et := av.Type().Elem()
	if et.Kind() == reflect.Ptr {
		fillPointerArray(out, et, elems)
	} else {
		for i, e := range elems {
			if e == nil {
				out.Index(i).Set(reflect.Zero(et))
			} else {
				out.Index(i).Set(arrayElem(e, et))
			}
		}
	}
	if out.Len() > len(elems) {
		out.Index(len(elems)).Set(reflect.Zero(et))
	}
	return out.Interface()
}

// fillPointerArray stores elems in the Java array out, whose element type et
// is a pointer to a class. A pointer to any part of an object (the part for a
// superclass) has the object's address, so storing an element only needs a
// check that its class has the part et points to, which is the same for every
// element of the same type.
func fillPointerArray(out reflect.Value, et reflect.Type, elems []any) {
	base := unsafe.Pointer(out.Pointer())
	var checked reflect.Type
	for i, e := range elems {
		var p unsafe.Pointer
		if e != nil {
			if t := reflect.TypeOf(e); t != checked {
				if !hasPart(t, et) {
					panic(ArrayStoreException_new_String(FromGo(className(e))))
				}
				checked = t
			}
			p = (*eface)(unsafe.Pointer(&elems[i])).data
		}
		*(*unsafe.Pointer)(unsafe.Add(base, uintptr(i)*unsafe.Sizeof(p))) = p
	}
}

// partCache memoizes hasPart. Its entries are facts about pairs of Go types,
// which never change, so sharing it between goroutines is safe.
var partCache sync.Map

// hasPart reports whether values of pointer type t can be stored in an array
// with element type et (a pointer to a class): t is et, or t's class extends
// et's class.
func hasPart(t, et reflect.Type) bool {
	if t == et {
		return true
	}
	key := [2]reflect.Type{t, et}
	if v, ok := partCache.Load(key); ok {
		return v.(bool)
	}
	ok := false
	if t.Kind() == reflect.Ptr {
		m, found := t.MethodByName("JClass_" + et.Elem().Name())
		ok = found && m.Type.NumOut() == 1 && m.Type.Out(0) == et
	}
	partCache.Store(key, ok)
	return ok
}

// arrayElem converts a collection element to the element type of a Java
// array. Generated classes are stored in arrays by the pointer to the part of
// the object for the array's element class, which JClass_ methods return.
func arrayElem(e any, et reflect.Type) reflect.Value {
	v := reflect.ValueOf(e)
	if v.Type().AssignableTo(et) {
		return v
	}
	if et.Kind() == reflect.Ptr {
		name := "JClass_" + et.Elem().Name()
		if m := v.MethodByName(name); m.IsValid() {
			return m.Call(nil)[0]
		}
	}
	panic(ArrayStoreException_new_String(FromGo(className(e))))
}

// arrayValue returns the reflect.Value of the Java array a, throwing a
// NullPointerException if a is null.
func arrayValue(a any) reflect.Value {
	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Slice || v.IsNil() {
		panic(NullPointerException_new())
	}
	return v
}

// arrayGet reads element i of a Java array as an any. A null element becomes
// a nil any, and an element that is a part pointer becomes the whole object:
// through JSelf for a runtime class that generated classes extend, and
// through JToAny for a generated class.
func arrayGet(v reflect.Value, i int) any {
	e := v.Index(i)
	switch e.Kind() {
	case reflect.Interface, reflect.Ptr, reflect.Slice, reflect.Map, reflect.Func:
		if e.IsNil() {
			return nil
		}
	}
	switch x := e.Interface().(type) {
	case selfRecorder:
		if self := x.JSelf(); self != nil {
			return self
		}
		return x
	case interface{ JToAny() any }:
		return x.JToAny()
	default:
		return x
	}
}

func arraySet(v reflect.Value, i int, o any) {
	if o == nil {
		v.Index(i).Set(reflect.Zero(v.Type().Elem()))
		return
	}
	v.Index(i).Set(arrayElem(o, v.Type().Elem()))
}

// arrayToAnys copies elements [from, to) of the Java array a into a []any.
func arrayToAnys(a any, from, to int) []any {
	if s, ok := a.([]any); ok {
		return append(make([]any, 0, to-from), s[from:to]...)
	}
	v := arrayValue(a)
	out := make([]any, to-from)
	for i := range out {
		out[i] = arrayGet(v, from+i)
	}
	return out
}

// ---------------------------------------------------------------------------
// AbstractCollection, AbstractSet, AbstractList and AbstractMap algorithms.
// Each takes the receiver as an interface so that the calls it makes on it
// dispatch as Java's do.

// acToString is AbstractCollection.toString; self is the whole receiver.
func acToString(c Collection, self any) *String {
	it := c.Iterator()
	if !it.HasNext() {
		return Lit("[]")
	}
	parts := []*String{Lit("[")}
	for {
		e := it.Next()
		if e != nil && sameRef(e, self) {
			parts = append(parts, Lit("(this Collection)"))
		} else {
			parts = append(parts, StrOfObject(e))
		}
		if !it.HasNext() {
			parts = append(parts, Lit("]"))
			return Concat(parts...)
		}
		parts = append(parts, Lit(", "))
	}
}

// acContains is AbstractCollection.contains.
func acContains(c Collection, o any) bool {
	it := c.Iterator()
	if o == nil {
		for it.HasNext() {
			if it.Next() == nil {
				return true
			}
		}
		return false
	}
	for it.HasNext() {
		if Object_Equals_Object(o, it.Next()) {
			return true
		}
	}
	return false
}

// acRemove is AbstractCollection.remove.
func acRemove(c Collection, o any) bool {
	it := c.Iterator()
	if o == nil {
		for it.HasNext() {
			if it.Next() == nil {
				it.Remove()
				return true
			}
		}
		return false
	}
	for it.HasNext() {
		if Object_Equals_Object(o, it.Next()) {
			it.Remove()
			return true
		}
	}
	return false
}

// acToArray is AbstractCollection.toArray.
func acToArray(c Collection) []any {
	out := make([]any, 0, c.Size())
	for it := c.Iterator(); it.HasNext(); {
		out = append(out, it.Next())
	}
	return out
}

// acToArrayT is AbstractCollection.toArray(T[]).
func acToArrayT(c Collection, a any) any {
	return fillArray(a, acToArray(c))
}

// acContainsAll is AbstractCollection.containsAll.
func acContainsAll(c Collection, other Collection) bool {
	for it := other.Iterator(); it.HasNext(); {
		if !c.Contains_Object(it.Next()) {
			return false
		}
	}
	return true
}

// acAddAll is AbstractCollection.addAll.
func acAddAll(c Collection, other Collection) bool {
	modified := false
	for it := other.Iterator(); it.HasNext(); {
		if c.Add_Object(it.Next()) {
			modified = true
		}
	}
	return modified
}

// acRemoveAll is AbstractCollection.removeAll.
func acRemoveAll(c Collection, other Collection) bool {
	if other == nil {
		panic(NullPointerException_new())
	}
	modified := false
	for it := c.Iterator(); it.HasNext(); {
		if other.Contains_Object(it.Next()) {
			it.Remove()
			modified = true
		}
	}
	return modified
}

// acRetainAll is AbstractCollection.retainAll.
func acRetainAll(c Collection, other Collection) bool {
	if other == nil {
		panic(NullPointerException_new())
	}
	modified := false
	for it := c.Iterator(); it.HasNext(); {
		if !other.Contains_Object(it.Next()) {
			it.Remove()
			modified = true
		}
	}
	return modified
}

// acClear is AbstractCollection.clear.
func acClear(c Collection) {
	for it := c.Iterator(); it.HasNext(); {
		it.Next()
		it.Remove()
	}
}

// asRemoveAll is AbstractSet.removeAll.
func asRemoveAll(s Set, c Collection) bool {
	if c == nil {
		panic(NullPointerException_new())
	}
	modified := false
	if s.Size() > c.Size() {
		for it := c.Iterator(); it.HasNext(); {
			if s.Remove_Object(it.Next()) {
				modified = true
			}
		}
	} else {
		for it := s.Iterator(); it.HasNext(); {
			if c.Contains_Object(it.Next()) {
				it.Remove()
				modified = true
			}
		}
	}
	return modified
}

// asEquals is AbstractSet.equals; self is the whole receiver.
func asEquals(s Set, self any, o any) bool {
	if sameRef(o, self) {
		return true
	}
	c, ok := o.(Set)
	if !ok {
		return false
	}
	if c.Size() != s.Size() {
		return false
	}
	return catchClassCastOrNull(func() bool { return s.ContainsAll_Collection(c) })
}

// asHashCode is AbstractSet.hashCode.
func asHashCode(s Collection) int32 {
	var h int32
	for it := s.Iterator(); it.HasNext(); {
		h += hashOrZero(it.Next())
	}
	return h
}

// alEquals is AbstractList.equals; self is the whole receiver.
func alEquals(l List, self any, o any) bool {
	if sameRef(o, self) {
		return true
	}
	ol, ok := o.(List)
	if !ok {
		return false
	}
	e1 := l.ListIterator()
	e2 := ol.ListIterator()
	for e1.HasNext() && e2.HasNext() {
		o1 := e1.Next()
		o2 := e2.Next()
		if !objectsEqual(o1, o2) {
			return false
		}
	}
	return !(e1.HasNext() || e2.HasNext())
}

// alHashCode is AbstractList.hashCode.
func alHashCode(l Collection) int32 {
	h := int32(1)
	for it := l.Iterator(); it.HasNext(); {
		h = 31*h + hashOrZero(it.Next())
	}
	return h
}

// alIndexOf is AbstractList.indexOf.
func alIndexOf(l List, o any) int32 {
	it := l.ListIterator()
	if o == nil {
		for it.HasNext() {
			if it.Next() == nil {
				return it.PreviousIndex()
			}
		}
	} else {
		for it.HasNext() {
			if Object_Equals_Object(o, it.Next()) {
				return it.PreviousIndex()
			}
		}
	}
	return -1
}

// alLastIndexOf is AbstractList.lastIndexOf.
func alLastIndexOf(l List, o any) int32 {
	it := l.ListIterator_int(l.Size())
	if o == nil {
		for it.HasPrevious() {
			if it.Previous() == nil {
				return it.NextIndex()
			}
		}
	} else {
		for it.HasPrevious() {
			if Object_Equals_Object(o, it.Previous()) {
				return it.NextIndex()
			}
		}
	}
	return -1
}

// amEquals is AbstractMap.equals; self is the whole receiver.
func amEquals(m Map, self any, o any) bool {
	if sameRef(o, self) {
		return true
	}
	om, ok := o.(Map)
	if !ok {
		return false
	}
	if om.Size() != m.Size() {
		return false
	}
	return catchClassCastOrNull(func() bool {
		for it := m.EntrySet().Iterator(); it.HasNext(); {
			e := it.Next().(Map_Entry)
			key := e.GetKey()
			value := e.GetValue()
			if value == nil {
				if !(om.Get_Object(key) == nil && om.ContainsKey_Object(key)) {
					return false
				}
			} else if !Object_Equals_Object(value, om.Get_Object(key)) {
				return false
			}
		}
		return true
	})
}

// amHashCode is AbstractMap.hashCode.
func amHashCode(m Map) int32 {
	var h int32
	for it := m.EntrySet().Iterator(); it.HasNext(); {
		h += Object_HashCode(it.Next())
	}
	return h
}

// amToString is AbstractMap.toString; self is the whole receiver.
func amToString(m Map, self any) *String {
	it := m.EntrySet().Iterator()
	if !it.HasNext() {
		return Lit("{}")
	}
	parts := []*String{Lit("{")}
	for {
		e := it.Next().(Map_Entry)
		key := e.GetKey()
		value := e.GetValue()
		if key != nil && sameRef(key, self) {
			parts = append(parts, Lit("(this Map)"))
		} else {
			parts = append(parts, StrOfObject(key))
		}
		parts = append(parts, Lit("="))
		if value != nil && sameRef(value, self) {
			parts = append(parts, Lit("(this Map)"))
		} else {
			parts = append(parts, StrOfObject(value))
		}
		if !it.HasNext() {
			parts = append(parts, Lit("}"))
			return Concat(parts...)
		}
		parts = append(parts, Lit(", "))
	}
}

// amPutAll is AbstractMap.putAll.
func amPutAll(m Map, other Map) {
	for it := other.EntrySet().Iterator(); it.HasNext(); {
		e := it.Next().(Map_Entry)
		m.Put_Object_Object(e.GetKey(), e.GetValue())
	}
}

// mapGetOrDefault is Map.getOrDefault's default implementation.
func mapGetOrDefault(m Map, key, defaultValue any) any {
	if v := m.Get_Object(key); v != nil || m.ContainsKey_Object(key) {
		return v
	}
	return defaultValue
}

// entryEquals is Map.Entry equality as AbstractMap.SimpleEntry and
// TreeMap.Entry define it.
func entryEquals(key, value any, o any) bool {
	e, ok := o.(Map_Entry)
	return ok && objectsEqual(key, e.GetKey()) && objectsEqual(value, e.GetValue())
}

// entryHashCode is Map.Entry's hashCode.
func entryHashCode(key, value any) int32 {
	return hashOrZero(key) ^ hashOrZero(value)
}

// entryToString is Map.Entry's toString.
func entryToString(key, value any) *String {
	return Concat(StrOfObject(key), Lit("="), StrOfObject(value))
}

// Iterator_default_Remove is the default method Iterator.remove.
func Iterator_default_Remove(_ any) {
	panic(UnsupportedOperationException_new_String(Lit("remove")))
}
