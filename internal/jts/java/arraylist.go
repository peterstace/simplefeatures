package java

import "sync"

// ArrayList is java.util.ArrayList.
type ArrayList struct {
	jself any
	// jmon is the list's monitor, locked by synchronized blocks.
	jmon     sync.Mutex
	elems    []any
	modCount int32
	// capacity is the length of Java's backing array, and defaultEmpty
	// reports whether that array is the shared empty array of
	// new ArrayList(), which grows to 10 elements at once. They decide
	// whether ensureCapacity counts as a modification.
	capacity     int32
	defaultEmpty bool
}

// JMonitor returns the list's monitor.
func (this *ArrayList) JMonitor() *sync.Mutex { return &this.jmon }

func (this *ArrayList) JSelf() any        { return this.jself }
func (this *ArrayList) JSetSelf(self any) { this.jself = self }

func (this *ArrayList) JClass_ArrayList() *ArrayList { return this }
func (this *ArrayList) JClassName() string           { return "java.util.ArrayList" }
func (this *ArrayList) JIface_Collection()           {}
func (this *ArrayList) JIface_List()                 {}
func (this *ArrayList) JIface_RandomAccess()         {}

func ArrayList_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_ArrayList() *ArrayList })
	return ok
}

func ArrayList_fromAny(o any) *ArrayList {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_ArrayList() *ArrayList }](o).JClass_ArrayList()
}

// self returns the whole object, which is a subclass instance when a
// generated class extends ArrayList.
func (this *ArrayList) self() any {
	if this.jself != nil {
		return this.jself
	}
	return this
}

func ArrayList_new() *ArrayList {
	this := &ArrayList{}
	this.ArrayList_init()
	return this
}

func ArrayList_new_int(capacity int32) *ArrayList {
	this := &ArrayList{}
	this.ArrayList_init_int(capacity)
	return this
}

func ArrayList_new_Collection(c Collection) *ArrayList {
	this := &ArrayList{}
	this.ArrayList_init_Collection(c)
	return this
}

func (this *ArrayList) ArrayList_init() { this.defaultEmpty = true }

func (this *ArrayList) ArrayList_init_int(capacity int32) {
	if capacity < 0 {
		panic(IllegalArgumentException_new_String(Concat(Lit("Illegal Capacity: "), StrOfInt(capacity))))
	}
	this.elems = make([]any, 0, capacity)
	this.capacity = capacity
}

func (this *ArrayList) ArrayList_init_Collection(c Collection) {
	this.elems = c.ToArray()
	this.capacity = int32(len(this.elems))
}

// grow is ArrayList.grow(int), which sets the capacity as
// ArraysSupport.newLength does.
func (this *ArrayList) grow(minCapacity int32) {
	old := this.capacity
	if old > 0 || !this.defaultEmpty {
		growth := minCapacity - old
		if pref := old >> 1; pref > growth {
			growth = pref
		}
		this.capacity = old + growth
		return
	}
	this.capacity = minCapacity
	if this.capacity < 10 {
		this.capacity = 10
	}
	this.defaultEmpty = false
}

func (this *ArrayList) Size() int32   { return int32(len(this.elems)) }
func (this *ArrayList) IsEmpty() bool { return len(this.elems) == 0 }

func (this *ArrayList) Get_int(i int32) any {
	checkIndex(i, int32(len(this.elems)))
	return this.elems[i]
}

func (this *ArrayList) Set_int_Object(i int32, o any) any {
	checkIndex(i, int32(len(this.elems)))
	old := this.elems[i]
	this.elems[i] = o
	return old
}

func (this *ArrayList) Add_Object(o any) bool {
	this.modCount++
	if int32(len(this.elems)) == this.capacity {
		this.grow(this.capacity + 1)
	}
	this.elems = append(this.elems, o)
	return true
}

func (this *ArrayList) rangeCheckForAdd(i int32) {
	if i > int32(len(this.elems)) || i < 0 {
		panic(outOfBounds(i, int32(len(this.elems))))
	}
}

func (this *ArrayList) Add_int_Object(i int32, o any) {
	this.rangeCheckForAdd(i)
	this.modCount++
	if int32(len(this.elems)) == this.capacity {
		this.grow(this.capacity + 1)
	}
	this.elems = append(this.elems, nil)
	copy(this.elems[i+1:], this.elems[i:])
	this.elems[i] = o
}

func (this *ArrayList) Remove_int(i int32) any {
	checkIndex(i, int32(len(this.elems)))
	old := this.elems[i]
	this.fastRemove(int(i))
	return old
}

func (this *ArrayList) fastRemove(i int) {
	this.modCount++
	n := len(this.elems) - 1
	copy(this.elems[i:], this.elems[i+1:])
	this.elems[n] = nil
	this.elems = this.elems[:n]
}

func (this *ArrayList) Remove_Object(o any) bool {
	i := this.indexOfRange(o, 0, len(this.elems))
	if i < 0 {
		return false
	}
	this.fastRemove(i)
	return true
}

func (this *ArrayList) IndexOf_Object(o any) int32 {
	return int32(this.indexOfRange(o, 0, len(this.elems)))
}

func (this *ArrayList) indexOfRange(o any, from, to int) int {
	es := this.elems
	if o == nil {
		for i := from; i < to; i++ {
			if es[i] == nil {
				return i
			}
		}
	} else {
		for i := from; i < to; i++ {
			if Object_Equals_Object(o, es[i]) {
				return i
			}
		}
	}
	return -1
}

func (this *ArrayList) LastIndexOf_Object(o any) int32 {
	return int32(this.lastIndexOfRange(o, 0, len(this.elems)))
}

func (this *ArrayList) lastIndexOfRange(o any, from, to int) int {
	es := this.elems
	if o == nil {
		for i := to - 1; i >= from; i-- {
			if es[i] == nil {
				return i
			}
		}
	} else {
		for i := to - 1; i >= from; i-- {
			if Object_Equals_Object(o, es[i]) {
				return i
			}
		}
	}
	return -1
}

func (this *ArrayList) Contains_Object(o any) bool { return this.IndexOf_Object(o) >= 0 }

func (this *ArrayList) AddAll_Collection(c Collection) bool {
	a := c.ToArray()
	this.modCount++
	if len(a) == 0 {
		return false
	}
	this.growForAll(len(a))
	this.elems = append(this.elems, a...)
	return true
}

func (this *ArrayList) growForAll(numNew int) {
	if s := int32(len(this.elems)); int32(numNew) > this.capacity-s {
		this.grow(s + int32(numNew))
	}
}

func (this *ArrayList) AddAll_int_Collection(i int32, c Collection) bool {
	this.rangeCheckForAdd(i)
	a := c.ToArray()
	this.modCount++
	if len(a) == 0 {
		return false
	}
	this.growForAll(len(a))
	n := len(this.elems)
	this.elems = append(this.elems, a...)
	copy(this.elems[int(i)+len(a):], this.elems[i:n])
	copy(this.elems[i:], a)
	return true
}

func (this *ArrayList) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }

func (this *ArrayList) RemoveAll_Collection(c Collection) bool {
	return this.batchRemove(c, false, 0, len(this.elems))
}

func (this *ArrayList) RetainAll_Collection(c Collection) bool {
	return this.batchRemove(c, true, 0, len(this.elems))
}

// batchRemove is ArrayList.batchRemove: it removes the elements of
// [from, end) for which c.contains differs from complement.
func (this *ArrayList) batchRemove(c Collection, complement bool, from, end int) bool {
	if c == nil {
		panic(NullPointerException_new())
	}
	es := this.elems
	r := from
	for ; ; r++ {
		if r == end {
			return false
		}
		if c.Contains_Object(es[r]) != complement {
			break
		}
	}
	w := r
	r++
	defer func() {
		if p := recover(); p != nil {
			copy(es[w:], es[r:end])
			w += end - r
			this.modCount += int32(end - w)
			this.shiftTailOverGap(w, end)
			panic(p)
		}
		this.modCount += int32(end - w)
		this.shiftTailOverGap(w, end)
	}()
	for ; r < end; r++ {
		if e := es[r]; c.Contains_Object(e) == complement {
			es[w] = e
			w++
		}
	}
	return true
}

// shiftTailOverGap removes elements [lo, hi).
func (this *ArrayList) shiftTailOverGap(lo, hi int) {
	n := len(this.elems)
	copy(this.elems[lo:], this.elems[hi:])
	newSize := n - (hi - lo)
	for i := newSize; i < n; i++ {
		this.elems[i] = nil
	}
	this.elems = this.elems[:newSize]
}

// removeRange is ArrayList.removeRange.
func (this *ArrayList) removeRange(from, to int32) {
	if from > to {
		panic(IndexOutOfBoundsException_new_String(FromGo(
			"From Index: " + StrOfInt(from).Go() + " > To Index: " + StrOfInt(to).Go())))
	}
	this.modCount++
	this.shiftTailOverGap(int(from), int(to))
}

func (this *ArrayList) Clear() {
	this.modCount++
	for i := range this.elems {
		this.elems[i] = nil
	}
	this.elems = this.elems[:0]
}

func (this *ArrayList) ToArray() []any {
	return append(make([]any, 0, len(this.elems)), this.elems...)
}

// ToArray_ObjectArr implements toArray(T[]): a is a Java array of any element
// type, and the result has a's element type.
func (this *ArrayList) ToArray_ObjectArr(a any) any {
	return fillArray(a, this.elems)
}

func (this *ArrayList) SubList_int_int(from, to int32) List {
	subListRangeCheck(from, to, int32(len(this.elems)))
	return &arrayListSubList{root: this, offset: from, size: to - from, modCount: this.modCount}
}

func (this *ArrayList) Sort_Comparator(c Comparator) {
	expectedModCount := this.modCount
	timSort(this.elems, 0, len(this.elems), c)
	if this.modCount != expectedModCount {
		panic(ConcurrentModificationException_new())
	}
	this.modCount++
}

func (this *ArrayList) Iterator() Iterator {
	return &arrayListItr{list: this, lastRet: -1, expectedModCount: this.modCount}
}

func (this *ArrayList) ListIterator() ListIterator { return this.ListIterator_int(0) }

func (this *ArrayList) ListIterator_int(i int32) ListIterator {
	this.rangeCheckForAdd(i)
	return &arrayListItr{list: this, cursor: i, lastRet: -1, expectedModCount: this.modCount}
}

func (this *ArrayList) EnsureCapacity_int(minCapacity int32) {
	if minCapacity > this.capacity && !(this.defaultEmpty && minCapacity <= 10) {
		this.modCount++
		this.grow(minCapacity)
	}
}

func (this *ArrayList) TrimToSize() {
	this.modCount++
	if n := int32(len(this.elems)); n < this.capacity {
		this.capacity = n
		this.defaultEmpty = false
	}
}

func (this *ArrayList) Equals_Object(o any) bool {
	if sameRef(o, this.self()) {
		return true
	}
	l, ok := o.(List)
	if !ok {
		return false
	}
	expectedModCount := this.modCount
	var equal bool
	if other, ok := o.(*ArrayList); ok && other.jself == nil {
		equal = this.equalsArrayList(other)
	} else {
		equal = this.equalsRange(l, 0, len(this.elems))
	}
	this.checkForComodification(expectedModCount)
	return equal
}

func (this *ArrayList) equalsArrayList(other *ArrayList) bool {
	otherModCount := other.modCount
	s := len(this.elems)
	equal := s == len(other.elems)
	if equal {
		for i := 0; i < s; i++ {
			if !javaObjectsEquals(this.elems[i], other.elems[i]) {
				equal = false
				break
			}
		}
	}
	other.checkForComodification(otherModCount)
	return equal
}

func (this *ArrayList) equalsRange(other List, from, to int) bool {
	oit := other.Iterator()
	for ; from < to; from++ {
		if !oit.HasNext() || !javaObjectsEquals(this.elems[from], oit.Next()) {
			return false
		}
	}
	return !oit.HasNext()
}

func (this *ArrayList) checkForComodification(expectedModCount int32) {
	if this.modCount != expectedModCount {
		panic(ConcurrentModificationException_new())
	}
}

func (this *ArrayList) HashCode() int32 {
	expectedModCount := this.modCount
	h := this.hashCodeRange(0, len(this.elems))
	this.checkForComodification(expectedModCount)
	return h
}

func (this *ArrayList) hashCodeRange(from, to int) int32 {
	h := int32(1)
	for i := from; i < to; i++ {
		h = 31*h + hashOrZero(this.elems[i])
	}
	return h
}

func (this *ArrayList) ToString() *String { return acToString(this, this.self()) }

// Clone implements ArrayList.clone. When a generated class extends ArrayList
// the clone is a shallow copy of the whole object, as Object.clone makes.
func (this *ArrayList) Clone() any {
	var c any
	var cl *ArrayList
	if this.jself != nil && !sameRef(this.jself, this) {
		c = ObjectImpl_Clone(this.jself)
		cl = c.(interface{ JClass_ArrayList() *ArrayList }).JClass_ArrayList()
		cl.jself = c
		cl.jmon = sync.Mutex{}
	} else {
		cl = &ArrayList{}
		c = cl
	}
	cl.elems = append(make([]any, 0, len(this.elems)), this.elems...)
	cl.modCount = 0
	cl.capacity = int32(len(this.elems))
	cl.defaultEmpty = false
	return c
}

// arrayListItr is ArrayList.Itr and ArrayList.ListItr.
type arrayListItr struct {
	list             *ArrayList
	cursor           int32
	lastRet          int32
	expectedModCount int32
}

func (it *arrayListItr) checkForComodification() {
	if it.list.modCount != it.expectedModCount {
		panic(ConcurrentModificationException_new())
	}
}

func (it *arrayListItr) HasNext() bool { return it.cursor != int32(len(it.list.elems)) }

func (it *arrayListItr) Next() any {
	it.checkForComodification()
	i := it.cursor
	if i >= int32(len(it.list.elems)) {
		panic(NoSuchElementException_new())
	}
	it.cursor = i + 1
	it.lastRet = i
	return it.list.elems[i]
}

func (it *arrayListItr) Remove() {
	if it.lastRet < 0 {
		panic(IllegalStateException_new())
	}
	it.checkForComodification()
	if it.lastRet >= int32(len(it.list.elems)) {
		panic(ConcurrentModificationException_new())
	}
	it.list.Remove_int(it.lastRet)
	it.cursor = it.lastRet
	it.lastRet = -1
	it.expectedModCount = it.list.modCount
}

func (it *arrayListItr) HasPrevious() bool    { return it.cursor != 0 }
func (it *arrayListItr) NextIndex() int32     { return it.cursor }
func (it *arrayListItr) PreviousIndex() int32 { return it.cursor - 1 }

func (it *arrayListItr) Previous() any {
	it.checkForComodification()
	i := it.cursor - 1
	if i < 0 {
		panic(NoSuchElementException_new())
	}
	if i >= int32(len(it.list.elems)) {
		panic(ConcurrentModificationException_new())
	}
	it.cursor = i
	it.lastRet = i
	return it.list.elems[i]
}

func (it *arrayListItr) Set_Object(o any) {
	if it.lastRet < 0 {
		panic(IllegalStateException_new())
	}
	it.checkForComodification()
	if it.lastRet >= int32(len(it.list.elems)) {
		panic(ConcurrentModificationException_new())
	}
	it.list.Set_int_Object(it.lastRet, o)
}

func (it *arrayListItr) Add_Object(o any) {
	it.checkForComodification()
	i := it.cursor
	if i > int32(len(it.list.elems)) {
		panic(ConcurrentModificationException_new())
	}
	it.list.Add_int_Object(i, o)
	it.cursor = i + 1
	it.lastRet = -1
	it.expectedModCount = it.list.modCount
}

// arrayListSubList is ArrayList.SubList, a view of a range of an ArrayList.
type arrayListSubList struct {
	root     *ArrayList
	parent   *arrayListSubList
	offset   int32
	size     int32
	modCount int32
}

func (this *arrayListSubList) JClassName() string   { return "java.util.ArrayList$SubList" }
func (this *arrayListSubList) JIface_Collection()   {}
func (this *arrayListSubList) JIface_List()         {}
func (this *arrayListSubList) JIface_RandomAccess() {}

func (this *arrayListSubList) checkForComodification() {
	if this.root.modCount != this.modCount {
		panic(ConcurrentModificationException_new())
	}
}

func (this *arrayListSubList) rangeCheckForAdd(i int32) {
	if i < 0 || i > this.size {
		panic(outOfBounds(i, this.size))
	}
}

func (this *arrayListSubList) updateSizeAndModCount(sizeChange int32) {
	for s := this; s != nil; s = s.parent {
		s.size += sizeChange
		s.modCount = this.root.modCount
	}
}

func (this *arrayListSubList) Set_int_Object(i int32, o any) any {
	checkIndex(i, this.size)
	this.checkForComodification()
	old := this.root.elems[this.offset+i]
	this.root.elems[this.offset+i] = o
	return old
}

func (this *arrayListSubList) Get_int(i int32) any {
	checkIndex(i, this.size)
	this.checkForComodification()
	return this.root.elems[this.offset+i]
}

func (this *arrayListSubList) Size() int32 {
	this.checkForComodification()
	return this.size
}

func (this *arrayListSubList) IsEmpty() bool { return this.Size() == 0 }

func (this *arrayListSubList) Add_Object(o any) bool {
	this.Add_int_Object(this.Size(), o)
	return true
}

func (this *arrayListSubList) Add_int_Object(i int32, o any) {
	this.rangeCheckForAdd(i)
	this.checkForComodification()
	this.root.Add_int_Object(this.offset+i, o)
	this.updateSizeAndModCount(1)
}

func (this *arrayListSubList) Remove_int(i int32) any {
	checkIndex(i, this.size)
	this.checkForComodification()
	result := this.root.Remove_int(this.offset + i)
	this.updateSizeAndModCount(-1)
	return result
}

func (this *arrayListSubList) removeRange(from, to int32) {
	this.checkForComodification()
	this.root.removeRange(this.offset+from, this.offset+to)
	this.updateSizeAndModCount(from - to)
}

func (this *arrayListSubList) AddAll_Collection(c Collection) bool {
	return this.AddAll_int_Collection(this.size, c)
}

func (this *arrayListSubList) AddAll_int_Collection(i int32, c Collection) bool {
	this.rangeCheckForAdd(i)
	cSize := c.Size()
	if cSize == 0 {
		return false
	}
	this.checkForComodification()
	this.root.AddAll_int_Collection(this.offset+i, c)
	this.updateSizeAndModCount(cSize)
	return true
}

func (this *arrayListSubList) RemoveAll_Collection(c Collection) bool {
	return this.batchRemove(c, false)
}

func (this *arrayListSubList) RetainAll_Collection(c Collection) bool {
	return this.batchRemove(c, true)
}

func (this *arrayListSubList) batchRemove(c Collection, complement bool) bool {
	this.checkForComodification()
	oldSize := int32(len(this.root.elems))
	modified := this.root.batchRemove(c, complement, int(this.offset), int(this.offset+this.size))
	if modified {
		this.updateSizeAndModCount(int32(len(this.root.elems)) - oldSize)
	}
	return modified
}

func (this *arrayListSubList) ToArray() []any {
	this.checkForComodification()
	return append([]any(nil), this.root.elems[this.offset:this.offset+this.size]...)
}

func (this *arrayListSubList) ToArray_ObjectArr(a any) any {
	this.checkForComodification()
	return fillArray(a, this.root.elems[this.offset:this.offset+this.size])
}

func (this *arrayListSubList) Equals_Object(o any) bool {
	if sameRef(o, this) {
		return true
	}
	l, ok := o.(List)
	if !ok {
		return false
	}
	equal := this.root.equalsRange(l, int(this.offset), int(this.offset+this.size))
	this.checkForComodification()
	return equal
}

func (this *arrayListSubList) HashCode() int32 {
	h := this.root.hashCodeRange(int(this.offset), int(this.offset+this.size))
	this.checkForComodification()
	return h
}

func (this *arrayListSubList) IndexOf_Object(o any) int32 {
	i := this.root.indexOfRange(o, int(this.offset), int(this.offset+this.size))
	this.checkForComodification()
	if i >= 0 {
		return int32(i) - this.offset
	}
	return -1
}

func (this *arrayListSubList) LastIndexOf_Object(o any) int32 {
	i := this.root.lastIndexOfRange(o, int(this.offset), int(this.offset+this.size))
	this.checkForComodification()
	if i >= 0 {
		return int32(i) - this.offset
	}
	return -1
}

func (this *arrayListSubList) Contains_Object(o any) bool { return this.IndexOf_Object(o) >= 0 }

func (this *arrayListSubList) ContainsAll_Collection(c Collection) bool {
	return acContainsAll(this, c)
}

func (this *arrayListSubList) Remove_Object(o any) bool { return acRemove(this, o) }

func (this *arrayListSubList) Clear() { this.removeRange(0, this.Size()) }

func (this *arrayListSubList) Sort_Comparator(c Comparator) { listSortDefault(this, c) }

func (this *arrayListSubList) ToString() *String { return acToString(this, this) }

func (this *arrayListSubList) Iterator() Iterator { return this.ListIterator_int(0) }

func (this *arrayListSubList) ListIterator() ListIterator { return this.ListIterator_int(0) }

func (this *arrayListSubList) ListIterator_int(i int32) ListIterator {
	this.checkForComodification()
	this.rangeCheckForAdd(i)
	return &arrayListSubListItr{sub: this, cursor: i, lastRet: -1, expectedModCount: this.modCount}
}

func (this *arrayListSubList) SubList_int_int(from, to int32) List {
	subListRangeCheck(from, to, this.size)
	return &arrayListSubList{
		root:     this.root,
		parent:   this,
		offset:   this.offset + from,
		size:     to - from,
		modCount: this.modCount,
	}
}

type arrayListSubListItr struct {
	sub              *arrayListSubList
	cursor           int32
	lastRet          int32
	expectedModCount int32
}

func (it *arrayListSubListItr) checkForComodification() {
	if it.sub.root.modCount != it.expectedModCount {
		panic(ConcurrentModificationException_new())
	}
}

func (it *arrayListSubListItr) HasNext() bool { return it.cursor != it.sub.size }

func (it *arrayListSubListItr) Next() any {
	it.checkForComodification()
	i := it.cursor
	if i >= it.sub.size {
		panic(NoSuchElementException_new())
	}
	if it.sub.offset+i >= int32(len(it.sub.root.elems)) {
		panic(ConcurrentModificationException_new())
	}
	it.cursor = i + 1
	it.lastRet = i
	return it.sub.root.elems[it.sub.offset+i]
}

func (it *arrayListSubListItr) HasPrevious() bool    { return it.cursor != 0 }
func (it *arrayListSubListItr) NextIndex() int32     { return it.cursor }
func (it *arrayListSubListItr) PreviousIndex() int32 { return it.cursor - 1 }

func (it *arrayListSubListItr) Previous() any {
	it.checkForComodification()
	i := it.cursor - 1
	if i < 0 {
		panic(NoSuchElementException_new())
	}
	if it.sub.offset+i >= int32(len(it.sub.root.elems)) {
		panic(ConcurrentModificationException_new())
	}
	it.cursor = i
	it.lastRet = i
	return it.sub.root.elems[it.sub.offset+i]
}

func (it *arrayListSubListItr) Remove() {
	if it.lastRet < 0 {
		panic(IllegalStateException_new())
	}
	it.checkForComodification()
	if it.lastRet >= it.sub.size {
		panic(ConcurrentModificationException_new())
	}
	it.sub.Remove_int(it.lastRet)
	it.cursor = it.lastRet
	it.lastRet = -1
	it.expectedModCount = it.sub.modCount
}

func (it *arrayListSubListItr) Set_Object(o any) {
	if it.lastRet < 0 {
		panic(IllegalStateException_new())
	}
	it.checkForComodification()
	if it.sub.offset+it.lastRet >= int32(len(it.sub.root.elems)) {
		panic(ConcurrentModificationException_new())
	}
	it.sub.root.Set_int_Object(it.sub.offset+it.lastRet, o)
}

func (it *arrayListSubListItr) Add_Object(o any) {
	it.checkForComodification()
	i := it.cursor
	if i > it.sub.size {
		panic(ConcurrentModificationException_new())
	}
	it.sub.Add_int_Object(i, o)
	it.cursor = i + 1
	it.lastRet = -1
	it.expectedModCount = it.sub.modCount
}
