package java

import "math/bits"

// This file holds the static methods of java.util.Collections and the
// collection classes they return.

func Collections_sort_List(l List) { l.Sort_Comparator(nil) }

func Collections_sort_List_Comparator(l List, c Comparator) { l.Sort_Comparator(c) }

func Collections_swap(l List, i, j int32) {
	l.Set_int_Object(i, l.Set_int_Object(j, l.Get_int(i)))
}

func Collections_reverse(l List) {
	size := l.Size()
	if size < 18 || InstanceOf[RandomAccess](l) {
		for i, mid, j := int32(0), size>>1, size-1; i < mid; i, j = i+1, j-1 {
			Collections_swap(l, i, j)
		}
		return
	}
	fwd := l.ListIterator()
	rev := l.ListIterator_int(size)
	for i, mid := int32(0), l.Size()>>1; i < mid; i++ {
		tmp := fwd.Next()
		fwd.Set_Object(rev.Previous())
		rev.Set_Object(tmp)
	}
}

func Collections_fill(l List, o any) {
	size := l.Size()
	if size < 25 || InstanceOf[RandomAccess](l) {
		for i := int32(0); i < size; i++ {
			l.Set_int_Object(i, o)
		}
		return
	}
	it := l.ListIterator()
	for i := int32(0); i < size; i++ {
		it.Next()
		it.Set_Object(o)
	}
}

func Collections_min_Collection(c Collection) any {
	i := c.Iterator()
	candidate := i.Next()
	for i.HasNext() {
		next := i.Next()
		if Comparable_CompareTo_Object(next, candidate) < 0 {
			candidate = next
		}
	}
	return candidate
}

func Collections_min_Collection_Comparator(c Collection, comp Comparator) any {
	if comp == nil {
		return Collections_min_Collection(c)
	}
	i := c.Iterator()
	candidate := i.Next()
	for i.HasNext() {
		next := i.Next()
		if comp.Compare_Object_Object(next, candidate) < 0 {
			candidate = next
		}
	}
	return candidate
}

func Collections_max_Collection(c Collection) any {
	i := c.Iterator()
	candidate := i.Next()
	for i.HasNext() {
		next := i.Next()
		if Comparable_CompareTo_Object(next, candidate) > 0 {
			candidate = next
		}
	}
	return candidate
}

func Collections_max_Collection_Comparator(c Collection, comp Comparator) any {
	if comp == nil {
		return Collections_max_Collection(c)
	}
	i := c.Iterator()
	candidate := i.Next()
	for i.HasNext() {
		next := i.Next()
		if comp.Compare_Object_Object(next, candidate) > 0 {
			candidate = next
		}
	}
	return candidate
}

func Collections_binarySearch_List_Object(l List, key any) int32 {
	return indexedBinarySearch(l, key, nil)
}

func Collections_binarySearch_List_Object_Comparator(l List, key any, c Comparator) int32 {
	return indexedBinarySearch(l, key, c)
}

// indexedBinarySearch is Collections.indexedBinarySearch. Collections also
// has an iterator-based search for large sequential lists, which makes the
// same comparisons.
func indexedBinarySearch(l List, key any, c Comparator) int32 {
	low := int32(0)
	high := l.Size() - 1
	for low <= high {
		mid := int32(uint32(low+high) >> 1)
		cmp := compareWith(c, l.Get_int(mid), key)
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

// Collections_addAll is Collections.addAll(Collection, T...); elements is a
// Java array.
func Collections_addAll(c Collection, elements any) bool {
	result := false
	for _, e := range arrayToAnys(elements, 0, arrayValue(elements).Len()) {
		if c.Add_Object(e) {
			result = true
		}
	}
	return result
}

// Collections_synchronizedList returns l itself: the translation is single
// threaded, so synchronization adds nothing.
func Collections_synchronizedList(l List) List { return l }

// ---------------------------------------------------------------------------
// Comparators

// reverseComparator is Collections.ReverseComparator, whose one instance is
// Collections.reverseOrder().
type reverseComparator struct{ _ byte }

func (*reverseComparator) JClassName() string { return "java.util.Collections$ReverseComparator" }

func (*reverseComparator) Compare_Object_Object(a, b any) int32 {
	return Comparable_CompareTo_Object(b, a)
}

// naturalOrderComparator is Comparators.NaturalOrderComparator, whose one
// instance is Comparator.naturalOrder().
type naturalOrderComparator struct{ _ byte }

func (*naturalOrderComparator) JClassName() string {
	return "java.util.Comparators$NaturalOrderComparator"
}

func (*naturalOrderComparator) Compare_Object_Object(a, b any) int32 {
	return Comparable_CompareTo_Object(a, b)
}

// Collections.ReverseComparator.REVERSE_ORDER and
// Comparators.NaturalOrderComparator.INSTANCE. They have no state to mutate.
var (
	reverseOrderInstance = &reverseComparator{}
	naturalOrderInstance = &naturalOrderComparator{}
)

// reverseComparator2 is Collections.ReverseComparator2, the reverse of a
// given comparator.
type reverseComparator2 struct {
	cmp Comparator
}

func (*reverseComparator2) JClassName() string { return "java.util.Collections$ReverseComparator2" }

func (c *reverseComparator2) Compare_Object_Object(a, b any) int32 {
	return c.cmp.Compare_Object_Object(b, a)
}

func (c *reverseComparator2) Equals_Object(o any) bool {
	if sameRef(o, c) {
		return true
	}
	other, ok := o.(*reverseComparator2)
	return ok && Object_Equals_Object(c.cmp, other.cmp)
}

func (c *reverseComparator2) HashCode() int32 { return Object_HashCode(c.cmp) ^ -0x80000000 }

func Collections_reverseOrder() Comparator { return reverseOrderInstance }

func Collections_reverseOrder_Comparator(cmp Comparator) Comparator {
	if cmp == nil {
		return reverseOrderInstance
	}
	switch c := cmp.(type) {
	case *reverseComparator:
		return naturalOrderInstance
	case *naturalOrderComparator:
		return reverseOrderInstance
	case *reverseComparator2:
		return c.cmp
	}
	return &reverseComparator2{cmp: cmp}
}

// Comparator_naturalOrder is Comparator.naturalOrder.
func Comparator_naturalOrder() Comparator { return naturalOrderInstance }

// Comparator_reverseOrder is Comparator.reverseOrder.
func Comparator_reverseOrder() Comparator { return reverseOrderInstance }

// ---------------------------------------------------------------------------
// Singleton and empty collections

// singletonIterator is Collections.singletonIterator.
type singletonIterator struct {
	e       any
	hasNext bool
}

func (it *singletonIterator) HasNext() bool { return it.hasNext }

func (it *singletonIterator) Next() any {
	if it.hasNext {
		it.hasNext = false
		return it.e
	}
	panic(NoSuchElementException_new())
}

func (it *singletonIterator) Remove() { panic(unsupported()) }

// singletonList is Collections.SingletonList.
type singletonList struct {
	element any
}

func Collections_singletonList(o any) List { return &singletonList{element: o} }

func (this *singletonList) JClassName() string   { return "java.util.Collections$SingletonList" }
func (this *singletonList) JIface_Collection()   {}
func (this *singletonList) JIface_List()         {}
func (this *singletonList) JIface_RandomAccess() {}

func (this *singletonList) Iterator() Iterator {
	return &singletonIterator{e: this.element, hasNext: true}
}
func (this *singletonList) Size() int32   { return 1 }
func (this *singletonList) IsEmpty() bool { return false }

func (this *singletonList) Contains_Object(o any) bool { return objectsEqual(o, this.element) }

func (this *singletonList) Get_int(i int32) any {
	if i != 0 {
		panic(outOfBounds(i, 1))
	}
	return this.element
}

func (this *singletonList) Sort_Comparator(Comparator) {}
func (this *singletonList) HashCode() int32            { return 31 + hashOrZero(this.element) }

func (this *singletonList) Set_int_Object(int32, any) any { panic(unsupported()) }
func (this *singletonList) Add_int_Object(int32, any)     { panic(unsupported()) }
func (this *singletonList) Remove_int(int32) any          { panic(unsupported()) }

func (this *singletonList) Add_Object(e any) bool {
	this.Add_int_Object(1, e)
	return true
}

func (this *singletonList) AddAll_int_Collection(i int32, c Collection) bool {
	return abstractListAddAllAt(this, i, c)
}

func (this *singletonList) ListIterator() ListIterator { return this.ListIterator_int(0) }

func (this *singletonList) ListIterator_int(i int32) ListIterator {
	return abstractListIterator(this, i)
}

func (this *singletonList) SubList_int_int(from, to int32) List {
	return newAbstractSubList(this, nil, from, to)
}

func (this *singletonList) Clear()                         { listRemoveRange(this, 0, this.Size()) }
func (this *singletonList) Remove_Object(o any) bool       { return acRemove(this, o) }
func (this *singletonList) IndexOf_Object(o any) int32     { return alIndexOf(this, o) }
func (this *singletonList) LastIndexOf_Object(o any) int32 { return alLastIndexOf(this, o) }
func (this *singletonList) ToArray() []any                 { return acToArray(this) }
func (this *singletonList) ToArray_ObjectArr(a any) any    { return acToArrayT(this, a) }
func (this *singletonList) Equals_Object(o any) bool       { return alEquals(this, this, o) }
func (this *singletonList) ToString() *String              { return acToString(this, this) }

func (this *singletonList) AddAll_Collection(c Collection) bool      { return acAddAll(this, c) }
func (this *singletonList) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }
func (this *singletonList) RemoveAll_Collection(c Collection) bool   { return acRemoveAll(this, c) }
func (this *singletonList) RetainAll_Collection(c Collection) bool   { return acRetainAll(this, c) }

// abstractListAddAllAt is AbstractList.addAll(int, Collection).
func abstractListAddAllAt(l List, index int32, c Collection) bool {
	if index < 0 || index > l.Size() {
		panic(outOfBounds(index, l.Size()))
	}
	modified := false
	for it := c.Iterator(); it.HasNext(); {
		l.Add_int_Object(index, it.Next())
		index++
		modified = true
	}
	return modified
}

// singletonSet is Collections.SingletonSet.
type singletonSet struct {
	element any
}

func Collections_singleton(o any) Set { return &singletonSet{element: o} }

func (this *singletonSet) JClassName() string { return "java.util.Collections$SingletonSet" }
func (this *singletonSet) JIface_Collection() {}
func (this *singletonSet) JIface_Set()        {}

func (this *singletonSet) Iterator() Iterator {
	return &singletonIterator{e: this.element, hasNext: true}
}
func (this *singletonSet) Size() int32     { return 1 }
func (this *singletonSet) IsEmpty() bool   { return false }
func (this *singletonSet) HashCode() int32 { return hashOrZero(this.element) }

func (this *singletonSet) Contains_Object(o any) bool { return objectsEqual(o, this.element) }

func (this *singletonSet) Add_Object(any) bool                      { panic(unsupported()) }
func (this *singletonSet) Remove_Object(o any) bool                 { return acRemove(this, o) }
func (this *singletonSet) Clear()                                   { acClear(this) }
func (this *singletonSet) AddAll_Collection(c Collection) bool      { return acAddAll(this, c) }
func (this *singletonSet) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }
func (this *singletonSet) RemoveAll_Collection(c Collection) bool   { return asRemoveAll(this, c) }
func (this *singletonSet) RetainAll_Collection(c Collection) bool   { return acRetainAll(this, c) }
func (this *singletonSet) ToArray() []any                           { return acToArray(this) }
func (this *singletonSet) ToArray_ObjectArr(a any) any              { return acToArrayT(this, a) }
func (this *singletonSet) Equals_Object(o any) bool                 { return asEquals(this, this, o) }
func (this *singletonSet) ToString() *String                        { return acToString(this, this) }

// emptyIterator is Collections.emptyIterator() and emptyListIterator().
type emptyIterator struct{ _ byte }

func (*emptyIterator) HasNext() bool        { return false }
func (*emptyIterator) Next() any            { panic(NoSuchElementException_new()) }
func (*emptyIterator) Remove()              { panic(IllegalStateException_new()) }
func (*emptyIterator) HasPrevious() bool    { return false }
func (*emptyIterator) Previous() any        { panic(NoSuchElementException_new()) }
func (*emptyIterator) NextIndex() int32     { return 0 }
func (*emptyIterator) PreviousIndex() int32 { return -1 }
func (*emptyIterator) Set_Object(any)       { panic(IllegalStateException_new()) }
func (*emptyIterator) Add_Object(any)       { panic(unsupported()) }

// emptyArray implements toArray(T[]) on an empty collection.
func emptyArray(a any) any {
	v := arrayValue(a)
	if v.Len() > 0 {
		arraySet(v, 0, nil)
	}
	return a
}

// emptyList is Collections.EmptyList.
type emptyList struct{ _ byte }

func (this *emptyList) JClassName() string   { return "java.util.Collections$EmptyList" }
func (this *emptyList) JIface_Collection()   {}
func (this *emptyList) JIface_List()         {}
func (this *emptyList) JIface_RandomAccess() {}

func (this *emptyList) Iterator() Iterator                       { return &emptyIterator{} }
func (this *emptyList) ListIterator() ListIterator               { return &emptyIterator{} }
func (this *emptyList) Size() int32                              { return 0 }
func (this *emptyList) IsEmpty() bool                            { return true }
func (this *emptyList) Clear()                                   {}
func (this *emptyList) Contains_Object(any) bool                 { return false }
func (this *emptyList) ContainsAll_Collection(c Collection) bool { return c.IsEmpty() }
func (this *emptyList) ToArray() []any                           { return []any{} }
func (this *emptyList) ToArray_ObjectArr(a any) any              { return emptyArray(a) }
func (this *emptyList) Sort_Comparator(Comparator)               {}
func (this *emptyList) HashCode() int32                          { return 1 }

func (this *emptyList) Get_int(i int32) any {
	panic(IndexOutOfBoundsException_new_String(Concat(Lit("Index: "), StrOfInt(i))))
}

func (this *emptyList) Equals_Object(o any) bool {
	l, ok := o.(List)
	return ok && l.IsEmpty()
}

func (this *emptyList) Set_int_Object(int32, any) any { panic(unsupported()) }
func (this *emptyList) Add_int_Object(int32, any)     { panic(unsupported()) }
func (this *emptyList) Remove_int(int32) any          { panic(unsupported()) }

func (this *emptyList) Add_Object(e any) bool {
	this.Add_int_Object(0, e)
	return true
}

func (this *emptyList) AddAll_int_Collection(i int32, c Collection) bool {
	return abstractListAddAllAt(this, i, c)
}

func (this *emptyList) ListIterator_int(i int32) ListIterator {
	return abstractListIterator(this, i)
}

func (this *emptyList) SubList_int_int(from, to int32) List {
	return newAbstractSubList(this, nil, from, to)
}

func (this *emptyList) Remove_Object(o any) bool               { return acRemove(this, o) }
func (this *emptyList) IndexOf_Object(o any) int32             { return alIndexOf(this, o) }
func (this *emptyList) LastIndexOf_Object(o any) int32         { return alLastIndexOf(this, o) }
func (this *emptyList) ToString() *String                      { return acToString(this, this) }
func (this *emptyList) AddAll_Collection(c Collection) bool    { return acAddAll(this, c) }
func (this *emptyList) RemoveAll_Collection(c Collection) bool { return acRemoveAll(this, c) }
func (this *emptyList) RetainAll_Collection(c Collection) bool { return acRetainAll(this, c) }

// emptySet is Collections.EmptySet.
type emptySet struct{ _ byte }

func (this *emptySet) JClassName() string { return "java.util.Collections$EmptySet" }
func (this *emptySet) JIface_Collection() {}
func (this *emptySet) JIface_Set()        {}

func (this *emptySet) Iterator() Iterator                       { return &emptyIterator{} }
func (this *emptySet) Size() int32                              { return 0 }
func (this *emptySet) IsEmpty() bool                            { return true }
func (this *emptySet) Clear()                                   {}
func (this *emptySet) Contains_Object(any) bool                 { return false }
func (this *emptySet) ContainsAll_Collection(c Collection) bool { return c.IsEmpty() }
func (this *emptySet) ToArray() []any                           { return []any{} }
func (this *emptySet) ToArray_ObjectArr(a any) any              { return emptyArray(a) }
func (this *emptySet) HashCode() int32                          { return 0 }

func (this *emptySet) Add_Object(any) bool                    { panic(unsupported()) }
func (this *emptySet) Remove_Object(o any) bool               { return acRemove(this, o) }
func (this *emptySet) AddAll_Collection(c Collection) bool    { return acAddAll(this, c) }
func (this *emptySet) RemoveAll_Collection(c Collection) bool { return asRemoveAll(this, c) }
func (this *emptySet) RetainAll_Collection(c Collection) bool { return acRetainAll(this, c) }
func (this *emptySet) Equals_Object(o any) bool               { return asEquals(this, this, o) }
func (this *emptySet) ToString() *String                      { return acToString(this, this) }

// emptyMap is Collections.EmptyMap.
type emptyMap struct{ _ byte }

func (this *emptyMap) JClassName() string { return "java.util.Collections$EmptyMap" }
func (this *emptyMap) JIface_Map()        {}

func (this *emptyMap) Size() int32                    { return 0 }
func (this *emptyMap) IsEmpty() bool                  { return true }
func (this *emptyMap) Clear()                         {}
func (this *emptyMap) ContainsKey_Object(any) bool    { return false }
func (this *emptyMap) ContainsValue_Object(any) bool  { return false }
func (this *emptyMap) Get_Object(any) any             { return nil }
func (this *emptyMap) KeySet() Set                    { return Collections_EMPTY_SET }
func (this *emptyMap) Values() Collection             { return Collections_EMPTY_SET }
func (this *emptyMap) EntrySet() Set                  { return Collections_EMPTY_SET }
func (this *emptyMap) HashCode() int32                { return 0 }
func (this *emptyMap) Put_Object_Object(any, any) any { panic(unsupported()) }
func (this *emptyMap) Remove_Object(any) any          { return nil }
func (this *emptyMap) PutAll_Map(m Map)               { amPutAll(this, m) }
func (this *emptyMap) ToString() *String              { return Lit("{}") }

func (this *emptyMap) GetOrDefault_Object_Object(_, defaultValue any) any { return defaultValue }

func (this *emptyMap) PutIfAbsent_Object_Object(any, any) any { panic(unsupported()) }

func (this *emptyMap) Equals_Object(o any) bool {
	m, ok := o.(Map)
	return ok && m.IsEmpty()
}

// Collections.EMPTY_LIST, EMPTY_SET and EMPTY_MAP, which are immutable.
var (
	Collections_EMPTY_LIST List = &emptyList{}
	Collections_EMPTY_SET  Set  = &emptySet{}
	Collections_EMPTY_MAP  Map  = &emptyMap{}
)

func Collections_emptyList() List { return Collections_EMPTY_LIST }
func Collections_emptySet() Set   { return Collections_EMPTY_SET }
func Collections_emptyMap() Map   { return Collections_EMPTY_MAP }

// copiesList is Collections.CopiesList, the list that nCopies returns.
type copiesList struct {
	n       int32
	element any
}

func Collections_nCopies(n int32, o any) List {
	if n < 0 {
		panic(IllegalArgumentException_new_String(Concat(Lit("List length = "), StrOfInt(n))))
	}
	return &copiesList{n: n, element: o}
}

func (this *copiesList) JClassName() string   { return "java.util.Collections$CopiesList" }
func (this *copiesList) JIface_Collection()   {}
func (this *copiesList) JIface_List()         {}
func (this *copiesList) JIface_RandomAccess() {}

func (this *copiesList) Size() int32   { return this.n }
func (this *copiesList) IsEmpty() bool { return this.n == 0 }

func (this *copiesList) Contains_Object(o any) bool {
	return this.n != 0 && objectsEqual(o, this.element)
}

func (this *copiesList) IndexOf_Object(o any) int32 {
	if this.Contains_Object(o) {
		return 0
	}
	return -1
}

func (this *copiesList) LastIndexOf_Object(o any) int32 {
	if this.Contains_Object(o) {
		return this.n - 1
	}
	return -1
}

func (this *copiesList) Get_int(i int32) any {
	if i < 0 || i >= this.n {
		panic(outOfBounds(i, this.n))
	}
	return this.element
}

func (this *copiesList) ToArray() []any {
	a := make([]any, this.n)
	for i := range a {
		a[i] = this.element
	}
	return a
}

func (this *copiesList) ToArray_ObjectArr(a any) any { return fillArray(a, this.ToArray()) }

func (this *copiesList) SubList_int_int(from, to int32) List {
	subListRangeCheck(from, to, this.n)
	return &copiesList{n: to - from, element: this.element}
}

func (this *copiesList) HashCode() int32 {
	if this.n == 0 {
		return 1
	}
	pow := int32(31)
	sum := int32(1)
	for i := bits.LeadingZeros32(uint32(this.n)) + 1; i < 32; i++ {
		sum *= pow + 1
		pow *= pow
		if this.n<<uint(i) < 0 {
			pow *= 31
			sum = sum*31 + 1
		}
	}
	return pow + sum*hashOrZero(this.element)
}

func (this *copiesList) Equals_Object(o any) bool {
	if sameRef(o, this) {
		return true
	}
	if other, ok := o.(*copiesList); ok {
		return this.n == other.n && (this.n == 0 || objectsEqual(this.element, other.element))
	}
	l, ok := o.(List)
	if !ok {
		return false
	}
	remaining := this.n
	itr := l.Iterator()
	for itr.HasNext() && remaining > 0 {
		remaining--
		if !objectsEqual(this.element, itr.Next()) {
			return false
		}
	}
	return remaining == 0 && !itr.HasNext()
}

func (this *copiesList) Set_int_Object(int32, any) any { panic(unsupported()) }
func (this *copiesList) Add_int_Object(int32, any)     { panic(unsupported()) }
func (this *copiesList) Remove_int(int32) any          { panic(unsupported()) }

func (this *copiesList) Add_Object(e any) bool {
	this.Add_int_Object(this.n, e)
	return true
}

func (this *copiesList) AddAll_int_Collection(i int32, c Collection) bool {
	return abstractListAddAllAt(this, i, c)
}

func (this *copiesList) Iterator() Iterator         { return newAbstractListItr(this, nil, 0) }
func (this *copiesList) ListIterator() ListIterator { return this.ListIterator_int(0) }

func (this *copiesList) ListIterator_int(i int32) ListIterator {
	return abstractListIterator(this, i)
}

func (this *copiesList) Sort_Comparator(c Comparator)             { listSortDefault(this, c) }
func (this *copiesList) Clear()                                   { listRemoveRange(this, 0, this.n) }
func (this *copiesList) Remove_Object(o any) bool                 { return acRemove(this, o) }
func (this *copiesList) ToString() *String                        { return acToString(this, this) }
func (this *copiesList) AddAll_Collection(c Collection) bool      { return acAddAll(this, c) }
func (this *copiesList) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }
func (this *copiesList) RemoveAll_Collection(c Collection) bool   { return acRemoveAll(this, c) }
func (this *copiesList) RetainAll_Collection(c Collection) bool   { return acRetainAll(this, c) }

// ---------------------------------------------------------------------------
// Unmodifiable views

// unmodifiableCollection is Collections.UnmodifiableCollection.
type unmodifiableCollection struct {
	c Collection
}

func Collections_unmodifiableCollection(c Collection) Collection {
	if u, ok := c.(*unmodifiableCollection); ok {
		return u
	}
	if c == nil {
		panic(NullPointerException_new())
	}
	return &unmodifiableCollection{c: c}
}

func (this *unmodifiableCollection) JClassName() string {
	return "java.util.Collections$UnmodifiableCollection"
}
func (this *unmodifiableCollection) JIface_Collection() {}

func (this *unmodifiableCollection) Size() int32                 { return this.c.Size() }
func (this *unmodifiableCollection) IsEmpty() bool               { return this.c.IsEmpty() }
func (this *unmodifiableCollection) Contains_Object(o any) bool  { return this.c.Contains_Object(o) }
func (this *unmodifiableCollection) ToArray() []any              { return this.c.ToArray() }
func (this *unmodifiableCollection) ToArray_ObjectArr(a any) any { return this.c.ToArray_ObjectArr(a) }
func (this *unmodifiableCollection) ToString() *String           { return Object_ToString(this.c) }

func (this *unmodifiableCollection) Iterator() Iterator {
	return &unmodifiableIterator{i: this.c.Iterator()}
}

func (this *unmodifiableCollection) ContainsAll_Collection(c Collection) bool {
	return this.c.ContainsAll_Collection(c)
}

func (this *unmodifiableCollection) Add_Object(any) bool                  { panic(unsupported()) }
func (this *unmodifiableCollection) Remove_Object(any) bool               { panic(unsupported()) }
func (this *unmodifiableCollection) AddAll_Collection(Collection) bool    { panic(unsupported()) }
func (this *unmodifiableCollection) RemoveAll_Collection(Collection) bool { panic(unsupported()) }
func (this *unmodifiableCollection) RetainAll_Collection(Collection) bool { panic(unsupported()) }
func (this *unmodifiableCollection) Clear()                               { panic(unsupported()) }

type unmodifiableIterator struct {
	i Iterator
}

func (it *unmodifiableIterator) HasNext() bool { return it.i.HasNext() }
func (it *unmodifiableIterator) Next() any     { return it.i.Next() }
func (it *unmodifiableIterator) Remove()       { panic(unsupported()) }

// unmodifiableSet is Collections.UnmodifiableSet.
type unmodifiableSet struct {
	unmodifiableCollection
}

func Collections_unmodifiableSet(s Set) Set {
	if u, ok := s.(*unmodifiableSet); ok {
		return u
	}
	if s == nil {
		panic(NullPointerException_new())
	}
	return &unmodifiableSet{unmodifiableCollection{c: s}}
}

func (this *unmodifiableSet) JClassName() string { return "java.util.Collections$UnmodifiableSet" }
func (this *unmodifiableSet) JIface_Set()        {}

func (this *unmodifiableSet) Equals_Object(o any) bool {
	return sameRef(o, this) || Object_Equals_Object(this.c, o)
}

func (this *unmodifiableSet) HashCode() int32 { return Object_HashCode(this.c) }

// unmodifiableList is Collections.UnmodifiableList and
// UnmodifiableRandomAccessList.
type unmodifiableList struct {
	unmodifiableCollection
	list         List
	randomAccess bool
}

func Collections_unmodifiableList(l List) List {
	if u, ok := l.(*unmodifiableList); ok {
		return u
	}
	if l == nil {
		panic(NullPointerException_new())
	}
	return newUnmodifiableList(l)
}

func newUnmodifiableList(l List) *unmodifiableList {
	return &unmodifiableList{
		unmodifiableCollection: unmodifiableCollection{c: l},
		list:                   l,
		randomAccess:           InstanceOf[RandomAccess](l),
	}
}

func (this *unmodifiableList) JClassName() string {
	if this.randomAccess {
		return "java.util.Collections$UnmodifiableRandomAccessList"
	}
	return "java.util.Collections$UnmodifiableList"
}
func (this *unmodifiableList) JIface_List() {}

func (this *unmodifiableList) Equals_Object(o any) bool {
	return sameRef(o, this) || Object_Equals_Object(this.list, o)
}

func (this *unmodifiableList) HashCode() int32                { return Object_HashCode(this.list) }
func (this *unmodifiableList) Get_int(i int32) any            { return this.list.Get_int(i) }
func (this *unmodifiableList) IndexOf_Object(o any) int32     { return this.list.IndexOf_Object(o) }
func (this *unmodifiableList) LastIndexOf_Object(o any) int32 { return this.list.LastIndexOf_Object(o) }

func (this *unmodifiableList) Set_int_Object(int32, any) any                { panic(unsupported()) }
func (this *unmodifiableList) Add_int_Object(int32, any)                    { panic(unsupported()) }
func (this *unmodifiableList) Remove_int(int32) any                         { panic(unsupported()) }
func (this *unmodifiableList) AddAll_int_Collection(int32, Collection) bool { panic(unsupported()) }
func (this *unmodifiableList) Sort_Comparator(Comparator)                   { panic(unsupported()) }

func (this *unmodifiableList) ListIterator() ListIterator { return this.ListIterator_int(0) }

func (this *unmodifiableList) ListIterator_int(i int32) ListIterator {
	return &unmodifiableListIterator{i: this.list.ListIterator_int(i)}
}

func (this *unmodifiableList) SubList_int_int(from, to int32) List {
	return newUnmodifiableList(this.list.SubList_int_int(from, to))
}

type unmodifiableListIterator struct {
	i ListIterator
}

func (it *unmodifiableListIterator) HasNext() bool        { return it.i.HasNext() }
func (it *unmodifiableListIterator) Next() any            { return it.i.Next() }
func (it *unmodifiableListIterator) HasPrevious() bool    { return it.i.HasPrevious() }
func (it *unmodifiableListIterator) Previous() any        { return it.i.Previous() }
func (it *unmodifiableListIterator) NextIndex() int32     { return it.i.NextIndex() }
func (it *unmodifiableListIterator) PreviousIndex() int32 { return it.i.PreviousIndex() }
func (it *unmodifiableListIterator) Remove()              { panic(unsupported()) }
func (it *unmodifiableListIterator) Set_Object(any)       { panic(unsupported()) }
func (it *unmodifiableListIterator) Add_Object(any)       { panic(unsupported()) }

// unmodifiableMap is Collections.UnmodifiableMap.
type unmodifiableMap struct {
	m        Map
	keySet   Set
	entrySet Set
	values   Collection
}

func Collections_unmodifiableMap(m Map) Map {
	if u, ok := m.(*unmodifiableMap); ok {
		return u
	}
	if m == nil {
		panic(NullPointerException_new())
	}
	return &unmodifiableMap{m: m}
}

func (this *unmodifiableMap) JClassName() string { return "java.util.Collections$UnmodifiableMap" }
func (this *unmodifiableMap) JIface_Map()        {}

func (this *unmodifiableMap) Size() int32                            { return this.m.Size() }
func (this *unmodifiableMap) IsEmpty() bool                          { return this.m.IsEmpty() }
func (this *unmodifiableMap) ContainsKey_Object(k any) bool          { return this.m.ContainsKey_Object(k) }
func (this *unmodifiableMap) ContainsValue_Object(v any) bool        { return this.m.ContainsValue_Object(v) }
func (this *unmodifiableMap) Get_Object(k any) any                   { return this.m.Get_Object(k) }
func (this *unmodifiableMap) Put_Object_Object(any, any) any         { panic(unsupported()) }
func (this *unmodifiableMap) Remove_Object(any) any                  { panic(unsupported()) }
func (this *unmodifiableMap) PutAll_Map(Map)                         { panic(unsupported()) }
func (this *unmodifiableMap) Clear()                                 { panic(unsupported()) }
func (this *unmodifiableMap) PutIfAbsent_Object_Object(any, any) any { panic(unsupported()) }

func (this *unmodifiableMap) GetOrDefault_Object_Object(k, defaultValue any) any {
	return this.m.GetOrDefault_Object_Object(k, defaultValue)
}

func (this *unmodifiableMap) KeySet() Set {
	if this.keySet == nil {
		this.keySet = Collections_unmodifiableSet(this.m.KeySet())
	}
	return this.keySet
}

func (this *unmodifiableMap) EntrySet() Set {
	if this.entrySet == nil {
		this.entrySet = &unmodifiableEntrySet{unmodifiableSet{unmodifiableCollection{c: this.m.EntrySet()}}}
	}
	return this.entrySet
}

func (this *unmodifiableMap) Values() Collection {
	if this.values == nil {
		this.values = Collections_unmodifiableCollection(this.m.Values())
	}
	return this.values
}

func (this *unmodifiableMap) Equals_Object(o any) bool {
	return sameRef(o, this) || Object_Equals_Object(this.m, o)
}

func (this *unmodifiableMap) HashCode() int32   { return Object_HashCode(this.m) }
func (this *unmodifiableMap) ToString() *String { return Object_ToString(this.m) }

// unmodifiableEntrySet is Collections.UnmodifiableMap.UnmodifiableEntrySet,
// which wraps each entry so that setValue is unsupported.
type unmodifiableEntrySet struct {
	unmodifiableSet
}

func (this *unmodifiableEntrySet) JClassName() string {
	return "java.util.Collections$UnmodifiableMap$UnmodifiableEntrySet"
}

func (this *unmodifiableEntrySet) Iterator() Iterator {
	return &unmodifiableEntryIterator{i: this.c.Iterator()}
}

func (this *unmodifiableEntrySet) ToArray() []any {
	a := this.c.ToArray()
	for i, e := range a {
		a[i] = &unmodifiableEntry{e: e.(Map_Entry)}
	}
	return a
}

func (this *unmodifiableEntrySet) ToArray_ObjectArr(a any) any {
	return fillArray(a, this.ToArray())
}

func (this *unmodifiableEntrySet) Contains_Object(o any) bool {
	e, ok := o.(Map_Entry)
	if !ok {
		return false
	}
	return this.c.Contains_Object(&unmodifiableEntry{e: e})
}

func (this *unmodifiableEntrySet) ContainsAll_Collection(c Collection) bool {
	for it := c.Iterator(); it.HasNext(); {
		if !this.Contains_Object(it.Next()) {
			return false
		}
	}
	return true
}

func (this *unmodifiableEntrySet) Equals_Object(o any) bool {
	if sameRef(o, this) {
		return true
	}
	s, ok := o.(Set)
	return ok && s.Size() == this.c.Size() && this.ContainsAll_Collection(s)
}

type unmodifiableEntryIterator struct {
	i Iterator
}

func (it *unmodifiableEntryIterator) HasNext() bool { return it.i.HasNext() }
func (it *unmodifiableEntryIterator) Next() any {
	return &unmodifiableEntry{e: it.i.Next().(Map_Entry)}
}
func (it *unmodifiableEntryIterator) Remove() { panic(unsupported()) }

// unmodifiableEntry is Collections.UnmodifiableMap.UnmodifiableEntrySet.
// UnmodifiableEntry.
type unmodifiableEntry struct {
	e Map_Entry
}

func (this *unmodifiableEntry) JClassName() string {
	return "java.util.Collections$UnmodifiableMap$UnmodifiableEntrySet$UnmodifiableEntry"
}
func (this *unmodifiableEntry) JIface_Map_Entry()       {}
func (this *unmodifiableEntry) GetKey() any             { return this.e.GetKey() }
func (this *unmodifiableEntry) GetValue() any           { return this.e.GetValue() }
func (this *unmodifiableEntry) SetValue_Object(any) any { panic(unsupported()) }
func (this *unmodifiableEntry) HashCode() int32         { return Object_HashCode(this.e) }
func (this *unmodifiableEntry) ToString() *String       { return Object_ToString(this.e) }

func (this *unmodifiableEntry) Equals_Object(o any) bool {
	if sameRef(o, this) {
		return true
	}
	t, ok := o.(Map_Entry)
	return ok && objectsEqual(this.e.GetKey(), t.GetKey()) && objectsEqual(this.e.GetValue(), t.GetValue())
}
