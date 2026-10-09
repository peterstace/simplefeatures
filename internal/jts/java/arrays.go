package java

import (
	"math"
	"reflect"
	"sort"
)

// This file holds the static methods of java.util.Arrays. Java arrays of
// object type arrive as an any holding a slice of any element type; they are
// read and written in place through reflection.

// ---------------------------------------------------------------------------
// asList

// arraysArrayList is Arrays.ArrayList, the fixed-size list view of an array
// that Arrays.asList returns. Writes go through to the array.
type arraysArrayList struct {
	a    any
	anys []any
	v    reflect.Value
}

func Arrays_asList(a any) List {
	l := &arraysArrayList{a: a}
	if s, ok := a.([]any); ok && s != nil {
		l.anys = s
	} else {
		l.v = arrayValue(a)
	}
	return l
}

func (this *arraysArrayList) JClassName() string   { return "java.util.Arrays$ArrayList" }
func (this *arraysArrayList) JIface_Collection()   {}
func (this *arraysArrayList) JIface_List()         {}
func (this *arraysArrayList) JIface_RandomAccess() {}

func (this *arraysArrayList) length() int32 {
	if this.anys != nil {
		return int32(len(this.anys))
	}
	return int32(this.v.Len())
}

func (this *arraysArrayList) at(i int32) any {
	n := this.length()
	if i < 0 || i >= n {
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(
			"Index " + StrOfInt(i).Go() + " out of bounds for length " + StrOfInt(n).Go())))
	}
	if this.anys != nil {
		return this.anys[i]
	}
	return arrayGet(this.v, int(i))
}

func (this *arraysArrayList) Size() int32   { return this.length() }
func (this *arraysArrayList) IsEmpty() bool { return this.length() == 0 }

func (this *arraysArrayList) ToArray() []any { return arrayToAnys(this.a, 0, int(this.length())) }

func (this *arraysArrayList) ToArray_ObjectArr(a any) any { return fillArray(a, this.ToArray()) }

func (this *arraysArrayList) Get_int(i int32) any { return this.at(i) }

func (this *arraysArrayList) Set_int_Object(i int32, e any) any {
	old := this.at(i)
	if this.anys != nil {
		this.anys[i] = e
	} else {
		arraySet(this.v, int(i), e)
	}
	return old
}

func (this *arraysArrayList) IndexOf_Object(o any) int32 {
	n := this.length()
	for i := int32(0); i < n; i++ {
		if objectsEqual(o, this.at(i)) {
			return i
		}
	}
	return -1
}

func (this *arraysArrayList) Contains_Object(o any) bool { return this.IndexOf_Object(o) >= 0 }

func (this *arraysArrayList) Sort_Comparator(c Comparator) {
	sortObjectArray(this.a, 0, int(this.length()), c)
}

func (this *arraysArrayList) Iterator() Iterator { return &arrayItr{list: this} }

// arrayItr is Arrays.ArrayItr.
type arrayItr struct {
	list   *arraysArrayList
	cursor int32
}

func (it *arrayItr) HasNext() bool { return it.cursor < it.list.length() }

func (it *arrayItr) Next() any {
	i := it.cursor
	if i >= it.list.length() {
		panic(NoSuchElementException_new())
	}
	it.cursor = i + 1
	return it.list.at(i)
}

func (it *arrayItr) Remove() { panic(iteratorRemoveUnsupported()) }

func (this *arraysArrayList) Add_Object(e any) bool {
	this.Add_int_Object(this.length(), e)
	return true
}

func (this *arraysArrayList) Add_int_Object(int32, any) { panic(unsupported()) }
func (this *arraysArrayList) Remove_int(int32) any      { panic(unsupported()) }

func (this *arraysArrayList) AddAll_int_Collection(i int32, c Collection) bool {
	return abstractListAddAllAt(this, i, c)
}

func (this *arraysArrayList) ListIterator() ListIterator { return this.ListIterator_int(0) }

func (this *arraysArrayList) ListIterator_int(i int32) ListIterator {
	return abstractListIterator(this, nil, i)
}

func (this *arraysArrayList) SubList_int_int(from, to int32) List {
	return newAbstractSubList(this, nil, from, to)
}

func (this *arraysArrayList) LastIndexOf_Object(o any) int32 { return alLastIndexOf(this, o) }
func (this *arraysArrayList) Clear()                         { listRemoveRange(this, 0, this.length()) }
func (this *arraysArrayList) Remove_Object(o any) bool       { return acRemove(this, o) }
func (this *arraysArrayList) Equals_Object(o any) bool       { return alEquals(this, this, o) }
func (this *arraysArrayList) HashCode() int32                { return alHashCode(this) }
func (this *arraysArrayList) ToString() *String              { return acToString(this, this) }

func (this *arraysArrayList) AddAll_Collection(c Collection) bool      { return acAddAll(this, c) }
func (this *arraysArrayList) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }
func (this *arraysArrayList) RemoveAll_Collection(c Collection) bool   { return acRemoveAll(this, c) }
func (this *arraysArrayList) RetainAll_Collection(c Collection) bool   { return acRetainAll(this, c) }

// ---------------------------------------------------------------------------
// sort

// arraysRangeCheck is Arrays.rangeCheck.
func arraysRangeCheck(length, from, to int32) {
	if from > to {
		panic(IllegalArgumentException_new_String(FromGo(
			"fromIndex(" + StrOfInt(from).Go() + ") > toIndex(" + StrOfInt(to).Go() + ")")))
	}
	if from < 0 {
		panic(ArrayIndexOutOfBoundsException_new_int(from))
	}
	if to > length {
		panic(ArrayIndexOutOfBoundsException_new_int(to))
	}
}

func Arrays_sort_ObjectArr(a any) {
	sortObjectArray(a, 0, arrayValue(a).Len(), nil)
}

func Arrays_sort_ObjectArr_int_int(a any, from, to int32) {
	arraysRangeCheck(int32(arrayValue(a).Len()), from, to)
	sortObjectArray(a, int(from), int(to), nil)
}

func Arrays_sort_ObjectArr_Comparator(a any, c Comparator) {
	sortObjectArray(a, 0, arrayValue(a).Len(), c)
}

func Arrays_sort_ObjectArr_int_int_Comparator(a any, from, to int32, c Comparator) {
	arraysRangeCheck(int32(arrayValue(a).Len()), from, to)
	sortObjectArray(a, int(from), int(to), c)
}

// sortObjectArray sorts elements [from, to) of the Java array a with TimSort.
func sortObjectArray(a any, from, to int, c Comparator) {
	if s, ok := a.([]any); ok {
		timSort(s, from, to, c)
		return
	}
	v := arrayValue(a)
	n := to - from
	if n < 2 {
		return
	}
	elems := make([]any, n)
	perm := make([]int, n)
	for i := range elems {
		elems[i] = arrayGet(v, from+i)
		perm[i] = i
	}
	cmp := Comparable_CompareTo_Object
	if c != nil {
		cmp = c.Compare_Object_Object
	}
	timSortFunc(perm, 0, n, func(x, y int) int32 { return cmp(elems[x], elems[y]) })
	orig := reflect.MakeSlice(v.Type(), n, n)
	reflect.Copy(orig, v.Slice(from, to))
	for k, j := range perm {
		v.Index(from + k).Set(orig.Index(j))
	}
}

// primitive is the set of Go types of Java's primitive numeric types.
type primitive interface {
	~int8 | ~int16 | ~uint16 | ~int32 | ~int64 | ~float32 | ~float64
}

type primitiveSlice[T primitive] []T

func (s primitiveSlice[T]) Len() int           { return len(s) }
func (s primitiveSlice[T]) Less(i, j int) bool { return s[i] < s[j] }
func (s primitiveSlice[T]) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// sortIntegral sorts an array of an integral type. Equal elements are
// indistinguishable, so any correct sort gives Java's result.
func sortIntegral[T primitive](a []T, from, to int32) {
	arraysRangeCheck(int32(len(a)), from, to)
	sort.Sort(primitiveSlice[T](a[from:to]))
}

// sortFloating is DualPivotQuicksort.sort for float and double arrays:
// negative zeros are made positive and NaNs moved to the end (both in Java's
// order), the rest are sorted, and the negative zeros restored.
func sortFloating[T float32 | float64](a []T, low, high int32) {
	arraysRangeCheck(int32(len(a)), low, high)
	numNegativeZero := 0
	for k := high; k > low; {
		k--
		ak := a[k]
		if ak == 0 && math.Signbit(float64(ak)) {
			numNegativeZero++
			a[k] = 0
		} else if ak != ak {
			high--
			a[k] = a[high]
			a[high] = ak
		}
	}
	sort.Sort(primitiveSlice[T](a[low:high]))
	numNegativeZero++
	if numNegativeZero == 1 {
		return
	}
	for low <= high {
		middle := int32(uint32(low+high) >> 1)
		if a[middle] < 0 {
			low = middle + 1
		} else {
			high = middle - 1
		}
	}
	negZero := T(math.Copysign(0, -1))
	for numNegativeZero--; numNegativeZero > 0; numNegativeZero-- {
		high++
		a[high] = negZero
	}
}

func Arrays_sort_intArr(a []int32)                     { sortIntegral(a, 0, int32(len(a))) }
func Arrays_sort_intArr_int_int(a []int32, f, t int32) { sortIntegral(a, f, t) }
func Arrays_sort_longArr(a []int64)                    { sortIntegral(a, 0, int32(len(a))) }

func Arrays_sort_longArr_int_int(a []int64, f, t int32)  { sortIntegral(a, f, t) }
func Arrays_sort_shortArr(a []int16)                     { sortIntegral(a, 0, int32(len(a))) }
func Arrays_sort_shortArr_int_int(a []int16, f, t int32) { sortIntegral(a, f, t) }
func Arrays_sort_charArr(a []uint16)                     { sortIntegral(a, 0, int32(len(a))) }
func Arrays_sort_charArr_int_int(a []uint16, f, t int32) { sortIntegral(a, f, t) }
func Arrays_sort_byteArr(a []int8)                       { sortIntegral(a, 0, int32(len(a))) }
func Arrays_sort_byteArr_int_int(a []int8, f, t int32)   { sortIntegral(a, f, t) }

func Arrays_sort_floatArr(a []float32)                      { sortFloating(a, 0, int32(len(a))) }
func Arrays_sort_floatArr_int_int(a []float32, f, t int32)  { sortFloating(a, f, t) }
func Arrays_sort_doubleArr(a []float64)                     { sortFloating(a, 0, int32(len(a))) }
func Arrays_sort_doubleArr_int_int(a []float64, f, t int32) { sortFloating(a, f, t) }

// ---------------------------------------------------------------------------
// binarySearch

func binarySearchIntegral[T primitive](a []T, from, to int32, key T) int32 {
	low := from
	high := to - 1
	for low <= high {
		mid := int32(uint32(low+high) >> 1)
		midVal := a[mid]
		switch {
		case midVal < key:
			low = mid + 1
		case midVal > key:
			high = mid - 1
		default:
			return mid
		}
	}
	return -(low + 1)
}

func binarySearchFloating[T float32 | float64](a []T, from, to int32, key T, bitsOf func(T) int64) int32 {
	low := from
	high := to - 1
	for low <= high {
		mid := int32(uint32(low+high) >> 1)
		midVal := a[mid]
		switch {
		case midVal < key:
			low = mid + 1
		case midVal > key:
			high = mid - 1
		default:
			midBits := bitsOf(midVal)
			keyBits := bitsOf(key)
			switch {
			case midBits == keyBits:
				return mid
			case midBits < keyBits:
				low = mid + 1
			default:
				high = mid - 1
			}
		}
	}
	return -(low + 1)
}

func doubleBits(d float64) int64 { return Double_doubleToLongBits(d) }
func floatBits(f float32) int64  { return int64(Float_floatToIntBits(f)) }

func Arrays_binarySearch_intArr_int(a []int32, key int32) int32 {
	return binarySearchIntegral(a, 0, int32(len(a)), key)
}

func Arrays_binarySearch_intArr_int_int_int(a []int32, from, to, key int32) int32 {
	arraysRangeCheck(int32(len(a)), from, to)
	return binarySearchIntegral(a, from, to, key)
}

func Arrays_binarySearch_longArr_long(a []int64, key int64) int32 {
	return binarySearchIntegral(a, 0, int32(len(a)), key)
}

func Arrays_binarySearch_longArr_int_int_long(a []int64, from, to int32, key int64) int32 {
	arraysRangeCheck(int32(len(a)), from, to)
	return binarySearchIntegral(a, from, to, key)
}

func Arrays_binarySearch_shortArr_short(a []int16, key int16) int32 {
	return binarySearchIntegral(a, 0, int32(len(a)), key)
}

func Arrays_binarySearch_shortArr_int_int_short(a []int16, from, to int32, key int16) int32 {
	arraysRangeCheck(int32(len(a)), from, to)
	return binarySearchIntegral(a, from, to, key)
}

func Arrays_binarySearch_charArr_char(a []uint16, key uint16) int32 {
	return binarySearchIntegral(a, 0, int32(len(a)), key)
}

func Arrays_binarySearch_charArr_int_int_char(a []uint16, from, to int32, key uint16) int32 {
	arraysRangeCheck(int32(len(a)), from, to)
	return binarySearchIntegral(a, from, to, key)
}

func Arrays_binarySearch_byteArr_byte(a []int8, key int8) int32 {
	return binarySearchIntegral(a, 0, int32(len(a)), key)
}

func Arrays_binarySearch_byteArr_int_int_byte(a []int8, from, to int32, key int8) int32 {
	arraysRangeCheck(int32(len(a)), from, to)
	return binarySearchIntegral(a, from, to, key)
}

func Arrays_binarySearch_doubleArr_double(a []float64, key float64) int32 {
	return binarySearchFloating(a, 0, int32(len(a)), key, doubleBits)
}

func Arrays_binarySearch_doubleArr_int_int_double(a []float64, from, to int32, key float64) int32 {
	arraysRangeCheck(int32(len(a)), from, to)
	return binarySearchFloating(a, from, to, key, doubleBits)
}

func Arrays_binarySearch_floatArr_float(a []float32, key float32) int32 {
	return binarySearchFloating(a, 0, int32(len(a)), key, floatBits)
}

func Arrays_binarySearch_floatArr_int_int_float(a []float32, from, to int32, key float32) int32 {
	arraysRangeCheck(int32(len(a)), from, to)
	return binarySearchFloating(a, from, to, key, floatBits)
}

func Arrays_binarySearch_ObjectArr_Object(a any, key any) int32 {
	return binarySearchObjects(a, 0, int32(arrayValue(a).Len()), key, nil)
}

func Arrays_binarySearch_ObjectArr_int_int_Object(a any, from, to int32, key any) int32 {
	arraysRangeCheck(int32(arrayValue(a).Len()), from, to)
	return binarySearchObjects(a, from, to, key, nil)
}

func Arrays_binarySearch_ObjectArr_Object_Comparator(a any, key any, c Comparator) int32 {
	return binarySearchObjects(a, 0, int32(arrayValue(a).Len()), key, c)
}

func Arrays_binarySearch_ObjectArr_int_int_Object_Comparator(a any, from, to int32, key any, c Comparator) int32 {
	arraysRangeCheck(int32(arrayValue(a).Len()), from, to)
	return binarySearchObjects(a, from, to, key, c)
}

func binarySearchObjects(a any, from, to int32, key any, c Comparator) int32 {
	v := arrayValue(a)
	low := from
	high := to - 1
	for low <= high {
		mid := int32(uint32(low+high) >> 1)
		midVal := arrayGet(v, int(mid))
		var cmp int32
		if c == nil {
			requireComparable(midVal)
			cmp = Comparable_CompareTo_Object(midVal, key)
		} else {
			cmp = c.Compare_Object_Object(midVal, key)
		}
		switch {
		case cmp < 0:
			low = mid + 1
		case cmp > 0:
			high = mid - 1
		default:
			return mid
		}
	}
	return -(low + 1)
}

// ---------------------------------------------------------------------------
// copyOf, copyOfRange

func negativeArraySize(n int32) any {
	return NegativeArraySizeException_new_String(StrOfInt(n))
}

func copyOf[T any](a []T, newLength int32) []T {
	if a == nil {
		panic(NullPointerException_new())
	}
	if newLength < 0 {
		panic(negativeArraySize(newLength))
	}
	out := make([]T, newLength)
	copy(out, a)
	return out
}

func copyOfRange[T any](a []T, from, to int32) []T {
	if a == nil {
		panic(NullPointerException_new())
	}
	newLength := to - from
	if newLength < 0 {
		panic(IllegalArgumentException_new_String(FromGo(StrOfInt(from).Go() + " > " + StrOfInt(to).Go())))
	}
	if from < 0 || from > int32(len(a)) {
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(
			"arraycopy: source index " + StrOfInt(from).Go() + " out of bounds for length " + StrOfInt(int32(len(a))).Go())))
	}
	out := make([]T, newLength)
	copy(out, a[from:])
	return out
}

func Arrays_copyOf_intArr_int(a []int32, n int32) []int32       { return copyOf(a, n) }
func Arrays_copyOf_longArr_int(a []int64, n int32) []int64      { return copyOf(a, n) }
func Arrays_copyOf_shortArr_int(a []int16, n int32) []int16     { return copyOf(a, n) }
func Arrays_copyOf_charArr_int(a []uint16, n int32) []uint16    { return copyOf(a, n) }
func Arrays_copyOf_byteArr_int(a []int8, n int32) []int8        { return copyOf(a, n) }
func Arrays_copyOf_booleanArr_int(a []bool, n int32) []bool     { return copyOf(a, n) }
func Arrays_copyOf_floatArr_int(a []float32, n int32) []float32 { return copyOf(a, n) }

func Arrays_copyOf_doubleArr_int(a []float64, n int32) []float64 { return copyOf(a, n) }

func Arrays_copyOfRange_intArr_int_int(a []int32, f, t int32) []int32    { return copyOfRange(a, f, t) }
func Arrays_copyOfRange_longArr_int_int(a []int64, f, t int32) []int64   { return copyOfRange(a, f, t) }
func Arrays_copyOfRange_shortArr_int_int(a []int16, f, t int32) []int16  { return copyOfRange(a, f, t) }
func Arrays_copyOfRange_charArr_int_int(a []uint16, f, t int32) []uint16 { return copyOfRange(a, f, t) }
func Arrays_copyOfRange_byteArr_int_int(a []int8, f, t int32) []int8     { return copyOfRange(a, f, t) }
func Arrays_copyOfRange_booleanArr_int_int(a []bool, f, t int32) []bool  { return copyOfRange(a, f, t) }

func Arrays_copyOfRange_floatArr_int_int(a []float32, f, t int32) []float32 {
	return copyOfRange(a, f, t)
}

func Arrays_copyOfRange_doubleArr_int_int(a []float64, f, t int32) []float64 {
	return copyOfRange(a, f, t)
}

// Arrays_copyOf_ObjectArr_int is Arrays.copyOf(T[], int); the copy has the
// element type of a.
func Arrays_copyOf_ObjectArr_int(a any, newLength int32) any {
	v := arrayValue(a)
	if newLength < 0 {
		panic(negativeArraySize(newLength))
	}
	out := reflect.MakeSlice(v.Type(), int(newLength), int(newLength))
	reflect.Copy(out, v)
	return out.Interface()
}

// Arrays_copyOfRange_ObjectArr_int_int is Arrays.copyOfRange(T[], int, int).
func Arrays_copyOfRange_ObjectArr_int_int(a any, from, to int32) any {
	v := arrayValue(a)
	newLength := to - from
	if newLength < 0 {
		panic(IllegalArgumentException_new_String(FromGo(StrOfInt(from).Go() + " > " + StrOfInt(to).Go())))
	}
	if from < 0 || from > int32(v.Len()) {
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(
			"arraycopy: source index " + StrOfInt(from).Go() + " out of bounds for length " + StrOfInt(int32(v.Len())).Go())))
	}
	out := reflect.MakeSlice(v.Type(), int(newLength), int(newLength))
	reflect.Copy(out, v.Slice(int(from), v.Len()))
	return out.Interface()
}

// ---------------------------------------------------------------------------
// fill

func fill[T any](a []T, from, to int32, val T) {
	arraysRangeCheck(int32(len(a)), from, to)
	for i := from; i < to; i++ {
		a[i] = val
	}
}

func Arrays_fill_intArr_int(a []int32, val int32)             { fill(a, 0, int32(len(a)), val) }
func Arrays_fill_intArr_int_int_int(a []int32, f, t, v int32) { fill(a, f, t, v) }
func Arrays_fill_longArr_long(a []int64, val int64)           { fill(a, 0, int32(len(a)), val) }

func Arrays_fill_longArr_int_int_long(a []int64, f, t int32, v int64)     { fill(a, f, t, v) }
func Arrays_fill_shortArr_short(a []int16, val int16)                     { fill(a, 0, int32(len(a)), val) }
func Arrays_fill_shortArr_int_int_short(a []int16, f, t int32, v int16)   { fill(a, f, t, v) }
func Arrays_fill_charArr_char(a []uint16, val uint16)                     { fill(a, 0, int32(len(a)), val) }
func Arrays_fill_charArr_int_int_char(a []uint16, f, t int32, v uint16)   { fill(a, f, t, v) }
func Arrays_fill_byteArr_byte(a []int8, val int8)                         { fill(a, 0, int32(len(a)), val) }
func Arrays_fill_byteArr_int_int_byte(a []int8, f, t int32, v int8)       { fill(a, f, t, v) }
func Arrays_fill_booleanArr_boolean(a []bool, val bool)                   { fill(a, 0, int32(len(a)), val) }
func Arrays_fill_booleanArr_int_int_boolean(a []bool, f, t int32, v bool) { fill(a, f, t, v) }
func Arrays_fill_floatArr_float(a []float32, val float32)                 { fill(a, 0, int32(len(a)), val) }

func Arrays_fill_floatArr_int_int_float(a []float32, f, t int32, v float32) { fill(a, f, t, v) }
func Arrays_fill_doubleArr_double(a []float64, val float64)                 { fill(a, 0, int32(len(a)), val) }

func Arrays_fill_doubleArr_int_int_double(a []float64, f, t int32, v float64) { fill(a, f, t, v) }

func Arrays_fill_ObjectArr_Object(a any, val any) {
	Arrays_fill_ObjectArr_int_int_Object(a, 0, int32(arrayValue(a).Len()), val)
}

func Arrays_fill_ObjectArr_int_int_Object(a any, from, to int32, val any) {
	v := arrayValue(a)
	arraysRangeCheck(int32(v.Len()), from, to)
	for i := from; i < to; i++ {
		arraySet(v, int(i), val)
	}
}

// ---------------------------------------------------------------------------
// equals, hashCode, toString

func equalsArr[T comparable](a, b []T) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func Arrays_equals_intArr_intArr(a, b []int32) bool     { return equalsArr(a, b) }
func Arrays_equals_longArr_longArr(a, b []int64) bool   { return equalsArr(a, b) }
func Arrays_equals_shortArr_shortArr(a, b []int16) bool { return equalsArr(a, b) }
func Arrays_equals_charArr_charArr(a, b []uint16) bool  { return equalsArr(a, b) }
func Arrays_equals_byteArr_byteArr(a, b []int8) bool    { return equalsArr(a, b) }

func Arrays_equals_booleanArr_booleanArr(a, b []bool) bool { return equalsArr(a, b) }

func Arrays_equals_doubleArr_doubleArr(a, b []float64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if Double_doubleToLongBits(a[i]) != Double_doubleToLongBits(b[i]) {
			return false
		}
	}
	return true
}

func Arrays_equals_floatArr_floatArr(a, b []float32) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if Float_floatToIntBits(a[i]) != Float_floatToIntBits(b[i]) {
			return false
		}
	}
	return true
}

func Arrays_equals_ObjectArr_ObjectArr(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	va, vb := arrayValue(a), arrayValue(b)
	if va.Len() != vb.Len() {
		return false
	}
	for i := 0; i < va.Len(); i++ {
		if !javaObjectsEquals(arrayGet(va, i), arrayGet(vb, i)) {
			return false
		}
	}
	return true
}

func hashArr[T any](a []T, h func(T) int32) int32 {
	if a == nil {
		return 0
	}
	result := int32(1)
	for _, e := range a {
		result = 31*result + h(e)
	}
	return result
}

func Arrays_hashCode_intArr(a []int32) int32 {
	return hashArr(a, func(e int32) int32 { return e })
}

func Arrays_hashCode_longArr(a []int64) int32 {
	return hashArr(a, func(e int64) int32 { return int32(e ^ int64(uint64(e)>>32)) })
}

func Arrays_hashCode_shortArr(a []int16) int32 {
	return hashArr(a, func(e int16) int32 { return int32(e) })
}

func Arrays_hashCode_charArr(a []uint16) int32 {
	return hashArr(a, func(e uint16) int32 { return int32(e) })
}

func Arrays_hashCode_byteArr(a []int8) int32 {
	return hashArr(a, func(e int8) int32 { return int32(e) })
}

func Arrays_hashCode_booleanArr(a []bool) int32 {
	return hashArr(a, func(e bool) int32 {
		if e {
			return 1231
		}
		return 1237
	})
}

func Arrays_hashCode_floatArr(a []float32) int32 { return hashArr(a, Float_floatToIntBits) }

func Arrays_hashCode_doubleArr(a []float64) int32 { return hashArr(a, Double_hashCode_double) }

func Arrays_hashCode_ObjectArr(a any) int32 {
	if a == nil {
		return 0
	}
	v := arrayValue(a)
	result := int32(1)
	for i := 0; i < v.Len(); i++ {
		result = 31*result + hashOrZero(arrayGet(v, i))
	}
	return result
}

func toStringArr[T any](a []T, str func(T) *String) *String {
	if a == nil {
		return Lit("null")
	}
	if len(a) == 0 {
		return Lit("[]")
	}
	parts := make([]*String, 0, 2*len(a)+1)
	parts = append(parts, Lit("["))
	for i, e := range a {
		if i > 0 {
			parts = append(parts, Lit(", "))
		}
		parts = append(parts, str(e))
	}
	parts = append(parts, Lit("]"))
	return Concat(parts...)
}

func Arrays_toString_intArr(a []int32) *String  { return toStringArr(a, StrOfInt) }
func Arrays_toString_longArr(a []int64) *String { return toStringArr(a, StrOfLong) }

func Arrays_toString_shortArr(a []int16) *String {
	return toStringArr(a, func(e int16) *String { return StrOfInt(int32(e)) })
}

func Arrays_toString_charArr(a []uint16) *String { return toStringArr(a, StrOfChar) }

func Arrays_toString_byteArr(a []int8) *String {
	return toStringArr(a, func(e int8) *String { return StrOfInt(int32(e)) })
}

func Arrays_toString_booleanArr(a []bool) *String  { return toStringArr(a, StrOfBool) }
func Arrays_toString_floatArr(a []float32) *String { return toStringArr(a, StrOfFloat) }
func Arrays_toString_doubleArr(a []float64) *String {
	return toStringArr(a, StrOfDouble)
}

func Arrays_toString_ObjectArr(a any) *String {
	if a == nil {
		return Lit("null")
	}
	return toStringArr(arrayToAnys(a, 0, arrayValue(a).Len()), StrOfObject)
}

// Arrays_deepToString is Arrays.deepToString: nested arrays are formatted
// recursively, and an array that contains itself is shown as "[...]".
func Arrays_deepToString(a any) *String {
	if a == nil {
		return Lit("null")
	}
	var parts []*String
	deepToString(arrayValue(a), &parts, nil)
	return Concat(parts...)
}

func deepToString(v reflect.Value, parts *[]*String, dejaVu []reflect.Value) {
	n := v.Len()
	if n == 0 {
		*parts = append(*parts, Lit("[]"))
		return
	}
	dejaVu = append(dejaVu, v)
	*parts = append(*parts, Lit("["))
	for i := 0; i < n; i++ {
		if i > 0 {
			*parts = append(*parts, Lit(", "))
		}
		ev := v.Index(i)
		if ev.Kind() == reflect.Interface && !ev.IsNil() {
			ev = ev.Elem()
		}
		if ev.Kind() != reflect.Slice {
			*parts = append(*parts, StrOfObject(arrayGet(v, i)))
			continue
		}
		if ev.IsNil() {
			*parts = append(*parts, Lit("null"))
			continue
		}
		switch e := ev.Interface().(type) {
		case []int8:
			*parts = append(*parts, Arrays_toString_byteArr(e))
		case []int16:
			*parts = append(*parts, Arrays_toString_shortArr(e))
		case []int32:
			*parts = append(*parts, Arrays_toString_intArr(e))
		case []int64:
			*parts = append(*parts, Arrays_toString_longArr(e))
		case []uint16:
			*parts = append(*parts, Arrays_toString_charArr(e))
		case []float32:
			*parts = append(*parts, Arrays_toString_floatArr(e))
		case []float64:
			*parts = append(*parts, Arrays_toString_doubleArr(e))
		case []bool:
			*parts = append(*parts, Arrays_toString_booleanArr(e))
		default:
			seen := false
			for _, d := range dejaVu {
				if d.Pointer() == ev.Pointer() && d.Len() == ev.Len() && d.Type() == ev.Type() {
					seen = true
					break
				}
			}
			if seen {
				*parts = append(*parts, Lit("[...]"))
			} else {
				deepToString(ev, parts, dejaVu)
			}
		}
	}
	*parts = append(*parts, Lit("]"))
}
