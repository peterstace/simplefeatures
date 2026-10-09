package java

// This file holds the parts of java.util.AbstractList that list classes
// inherit without overriding: its iterators and its subList view.

// catchIndexOutOfBounds runs f and reports whether it threw an
// IndexOutOfBoundsException, which AbstractList's iterators translate.
func catchIndexOutOfBounds(f func()) (caught bool) {
	defer func() {
		if r := recover(); r != nil {
			ex := Throwable_ofPanic(r)
			if IndexOutOfBoundsException_isInstance(ex) {
				caught = true
				return
			}
			panic(ex)
		}
	}()
	f()
	return false
}

func readModCount(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

// abstractListItr is AbstractList.Itr and AbstractList.ListItr, which access
// the list through get, set, add and remove. modCount points at the list's
// modification count, or is nil for a list that is never structurally
// modified.
type abstractListItr struct {
	list             List
	modCount         *int32
	cursor           int32
	lastRet          int32
	expectedModCount int32
}

func newAbstractListItr(l List, modCount *int32, index int32) *abstractListItr {
	return &abstractListItr{list: l, modCount: modCount, cursor: index, lastRet: -1, expectedModCount: readModCount(modCount)}
}

// abstractListIterator is AbstractList.listIterator(int).
func abstractListIterator(l List, index int32) ListIterator {
	if index < 0 || index > l.Size() {
		panic(outOfBounds(index, l.Size()))
	}
	return newAbstractListItr(l, nil, index)
}

func (it *abstractListItr) checkForComodification() {
	if readModCount(it.modCount) != it.expectedModCount {
		panic(ConcurrentModificationException_new())
	}
}

func (it *abstractListItr) HasNext() bool { return it.cursor != it.list.Size() }

func (it *abstractListItr) Next() any {
	it.checkForComodification()
	var next any
	i := it.cursor
	if catchIndexOutOfBounds(func() { next = it.list.Get_int(i) }) {
		it.checkForComodification()
		panic(NoSuchElementException_new())
	}
	it.lastRet = i
	it.cursor = i + 1
	return next
}

func (it *abstractListItr) Remove() {
	if it.lastRet < 0 {
		panic(IllegalStateException_new())
	}
	it.checkForComodification()
	if catchIndexOutOfBounds(func() { it.list.Remove_int(it.lastRet) }) {
		panic(ConcurrentModificationException_new())
	}
	if it.lastRet < it.cursor {
		it.cursor--
	}
	it.lastRet = -1
	it.expectedModCount = readModCount(it.modCount)
}

func (it *abstractListItr) HasPrevious() bool    { return it.cursor != 0 }
func (it *abstractListItr) NextIndex() int32     { return it.cursor }
func (it *abstractListItr) PreviousIndex() int32 { return it.cursor - 1 }

func (it *abstractListItr) Previous() any {
	it.checkForComodification()
	var prev any
	i := it.cursor - 1
	if catchIndexOutOfBounds(func() { prev = it.list.Get_int(i) }) {
		it.checkForComodification()
		panic(NoSuchElementException_new())
	}
	it.lastRet = i
	it.cursor = i
	return prev
}

func (it *abstractListItr) Set_Object(e any) {
	if it.lastRet < 0 {
		panic(IllegalStateException_new())
	}
	it.checkForComodification()
	if catchIndexOutOfBounds(func() { it.list.Set_int_Object(it.lastRet, e) }) {
		panic(ConcurrentModificationException_new())
	}
	it.expectedModCount = readModCount(it.modCount)
}

func (it *abstractListItr) Add_Object(e any) {
	it.checkForComodification()
	i := it.cursor
	if catchIndexOutOfBounds(func() { it.list.Add_int_Object(i, e) }) {
		panic(ConcurrentModificationException_new())
	}
	it.lastRet = -1
	it.cursor = i + 1
	it.expectedModCount = readModCount(it.modCount)
}

// listRemoveRange is a call of the protected method removeRange on l: the
// list's own version if it has one, else AbstractList's.
func listRemoveRange(l List, from, to int32) {
	if r, ok := l.(interface{ removeRange(from, to int32) }); ok {
		r.removeRange(from, to)
		return
	}
	it := l.ListIterator_int(from)
	for i, n := int32(0), to-from; i < n; i++ {
		it.Next()
		it.Remove()
	}
}

// abstractSubList is AbstractList.SubList, the view that AbstractList.subList
// returns.
type abstractSubList struct {
	root         List
	rootModCount *int32
	parent       *abstractSubList
	offset       int32
	size         int32
	modCount     int32
}

// newAbstractSubList is AbstractList.subList on root, whose modification
// count rootModCount points at (nil if it has none).
func newAbstractSubList(root List, rootModCount *int32, from, to int32) List {
	subListRangeCheck(from, to, root.Size())
	return &abstractSubList{
		root:         root,
		rootModCount: rootModCount,
		offset:       from,
		size:         to - from,
		modCount:     readModCount(rootModCount),
	}
}

func (this *abstractSubList) JClassName() string { return "java.util.AbstractList$SubList" }
func (this *abstractSubList) JIface_Collection() {}
func (this *abstractSubList) JIface_List()       {}

func (this *abstractSubList) checkForComodification() {
	if readModCount(this.rootModCount) != this.modCount {
		panic(ConcurrentModificationException_new())
	}
}

func (this *abstractSubList) rangeCheckForAdd(i int32) {
	if i < 0 || i > this.size {
		panic(outOfBounds(i, this.size))
	}
}

func (this *abstractSubList) updateSizeAndModCount(sizeChange int32) {
	for s := this; s != nil; s = s.parent {
		s.size += sizeChange
		s.modCount = readModCount(this.rootModCount)
	}
}

func (this *abstractSubList) Set_int_Object(i int32, e any) any {
	checkIndex(i, this.size)
	this.checkForComodification()
	return this.root.Set_int_Object(this.offset+i, e)
}

func (this *abstractSubList) Get_int(i int32) any {
	checkIndex(i, this.size)
	this.checkForComodification()
	return this.root.Get_int(this.offset + i)
}

func (this *abstractSubList) Size() int32 {
	this.checkForComodification()
	return this.size
}

func (this *abstractSubList) IsEmpty() bool { return this.Size() == 0 }

func (this *abstractSubList) Add_Object(e any) bool {
	this.Add_int_Object(this.Size(), e)
	return true
}

func (this *abstractSubList) Add_int_Object(i int32, e any) {
	this.rangeCheckForAdd(i)
	this.checkForComodification()
	this.root.Add_int_Object(this.offset+i, e)
	this.updateSizeAndModCount(1)
}

func (this *abstractSubList) Remove_int(i int32) any {
	checkIndex(i, this.size)
	this.checkForComodification()
	result := this.root.Remove_int(this.offset + i)
	this.updateSizeAndModCount(-1)
	return result
}

func (this *abstractSubList) removeRange(from, to int32) {
	this.checkForComodification()
	listRemoveRange(this.root, this.offset+from, this.offset+to)
	this.updateSizeAndModCount(from - to)
}

func (this *abstractSubList) AddAll_Collection(c Collection) bool {
	return this.AddAll_int_Collection(this.size, c)
}

func (this *abstractSubList) AddAll_int_Collection(i int32, c Collection) bool {
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

func (this *abstractSubList) Iterator() Iterator         { return this.ListIterator_int(0) }
func (this *abstractSubList) ListIterator() ListIterator { return this.ListIterator_int(0) }

func (this *abstractSubList) ListIterator_int(i int32) ListIterator {
	this.checkForComodification()
	this.rangeCheckForAdd(i)
	return &abstractSubListItr{sub: this, i: this.root.ListIterator_int(this.offset + i)}
}

func (this *abstractSubList) SubList_int_int(from, to int32) List {
	subListRangeCheck(from, to, this.size)
	return &abstractSubList{
		root:         this.root,
		rootModCount: this.rootModCount,
		parent:       this,
		offset:       this.offset + from,
		size:         to - from,
		modCount:     readModCount(this.rootModCount),
	}
}

func (this *abstractSubList) Clear()                         { this.removeRange(0, this.Size()) }
func (this *abstractSubList) Contains_Object(o any) bool     { return acContains(this, o) }
func (this *abstractSubList) Remove_Object(o any) bool       { return acRemove(this, o) }
func (this *abstractSubList) IndexOf_Object(o any) int32     { return alIndexOf(this, o) }
func (this *abstractSubList) LastIndexOf_Object(o any) int32 { return alLastIndexOf(this, o) }
func (this *abstractSubList) ToArray() []any                 { return acToArray(this) }
func (this *abstractSubList) ToArray_ObjectArr(a any) any    { return acToArrayT(this, a) }
func (this *abstractSubList) Sort_Comparator(c Comparator)   { listSortDefault(this, c) }
func (this *abstractSubList) Equals_Object(o any) bool       { return alEquals(this, this, o) }
func (this *abstractSubList) HashCode() int32                { return alHashCode(this) }
func (this *abstractSubList) ToString() *String              { return acToString(this, this) }

func (this *abstractSubList) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }
func (this *abstractSubList) RemoveAll_Collection(c Collection) bool   { return acRemoveAll(this, c) }
func (this *abstractSubList) RetainAll_Collection(c Collection) bool   { return acRetainAll(this, c) }

// abstractSubListItr is the list iterator of AbstractList.SubList, which
// wraps a list iterator of the root list.
type abstractSubListItr struct {
	sub *abstractSubList
	i   ListIterator
}

func (it *abstractSubListItr) HasNext() bool { return it.NextIndex() < it.sub.size }

func (it *abstractSubListItr) Next() any {
	if it.HasNext() {
		return it.i.Next()
	}
	panic(NoSuchElementException_new())
}

func (it *abstractSubListItr) HasPrevious() bool { return it.PreviousIndex() >= 0 }

func (it *abstractSubListItr) Previous() any {
	if it.HasPrevious() {
		return it.i.Previous()
	}
	panic(NoSuchElementException_new())
}

func (it *abstractSubListItr) NextIndex() int32     { return it.i.NextIndex() - it.sub.offset }
func (it *abstractSubListItr) PreviousIndex() int32 { return it.i.PreviousIndex() - it.sub.offset }

func (it *abstractSubListItr) Remove() {
	it.i.Remove()
	it.sub.updateSizeAndModCount(-1)
}

func (it *abstractSubListItr) Set_Object(e any) { it.i.Set_Object(e) }

func (it *abstractSubListItr) Add_Object(e any) {
	it.i.Add_Object(e)
	it.sub.updateSizeAndModCount(1)
}
