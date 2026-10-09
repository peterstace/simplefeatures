package java

// LinkedList is java.util.LinkedList, a doubly linked list.
type LinkedList struct {
	size     int32
	first    *linkedListNode
	last     *linkedListNode
	modCount int32
}

type linkedListNode struct {
	item any
	next *linkedListNode
	prev *linkedListNode
}

func (this *LinkedList) JClass_LinkedList() *LinkedList { return this }
func (this *LinkedList) JClassName() string             { return "java.util.LinkedList" }
func (this *LinkedList) JIface_Collection()             {}
func (this *LinkedList) JIface_List()                   {}
func (this *LinkedList) JIface_Queue()                  {}
func (this *LinkedList) JIface_Deque()                  {}

func LinkedList_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_LinkedList() *LinkedList })
	return ok
}

func LinkedList_fromAny(o any) *LinkedList {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_LinkedList() *LinkedList }](o).JClass_LinkedList()
}

func LinkedList_new() *LinkedList { return &LinkedList{} }

func LinkedList_new_Collection(c Collection) *LinkedList {
	this := &LinkedList{}
	this.AddAll_Collection(c)
	return this
}

func (this *LinkedList) linkFirst(e any) {
	f := this.first
	n := &linkedListNode{item: e, next: f}
	this.first = n
	if f == nil {
		this.last = n
	} else {
		f.prev = n
	}
	this.size++
	this.modCount++
}

func (this *LinkedList) linkLast(e any) {
	l := this.last
	n := &linkedListNode{item: e, prev: l}
	this.last = n
	if l == nil {
		this.first = n
	} else {
		l.next = n
	}
	this.size++
	this.modCount++
}

func (this *LinkedList) linkBefore(e any, succ *linkedListNode) {
	pred := succ.prev
	n := &linkedListNode{item: e, next: succ, prev: pred}
	succ.prev = n
	if pred == nil {
		this.first = n
	} else {
		pred.next = n
	}
	this.size++
	this.modCount++
}

func (this *LinkedList) unlinkFirst(f *linkedListNode) any {
	e := f.item
	next := f.next
	f.item = nil
	f.next = nil
	this.first = next
	if next == nil {
		this.last = nil
	} else {
		next.prev = nil
	}
	this.size--
	this.modCount++
	return e
}

func (this *LinkedList) unlinkLast(l *linkedListNode) any {
	e := l.item
	prev := l.prev
	l.item = nil
	l.prev = nil
	this.last = prev
	if prev == nil {
		this.first = nil
	} else {
		prev.next = nil
	}
	this.size--
	this.modCount++
	return e
}

func (this *LinkedList) unlink(x *linkedListNode) any {
	e := x.item
	next := x.next
	prev := x.prev
	if prev == nil {
		this.first = next
	} else {
		prev.next = next
		x.prev = nil
	}
	if next == nil {
		this.last = prev
	} else {
		next.prev = prev
		x.next = nil
	}
	x.item = nil
	this.size--
	this.modCount++
	return e
}

func (this *LinkedList) GetFirst() any {
	if this.first == nil {
		panic(NoSuchElementException_new())
	}
	return this.first.item
}

func (this *LinkedList) GetLast() any {
	if this.last == nil {
		panic(NoSuchElementException_new())
	}
	return this.last.item
}

func (this *LinkedList) RemoveFirst() any {
	if this.first == nil {
		panic(NoSuchElementException_new())
	}
	return this.unlinkFirst(this.first)
}

func (this *LinkedList) RemoveLast() any {
	if this.last == nil {
		panic(NoSuchElementException_new())
	}
	return this.unlinkLast(this.last)
}

func (this *LinkedList) AddFirst_Object(e any) { this.linkFirst(e) }
func (this *LinkedList) AddLast_Object(e any)  { this.linkLast(e) }

func (this *LinkedList) Contains_Object(o any) bool { return this.IndexOf_Object(o) >= 0 }

func (this *LinkedList) Size() int32   { return this.size }
func (this *LinkedList) IsEmpty() bool { return this.size == 0 }

func (this *LinkedList) Add_Object(e any) bool {
	this.linkLast(e)
	return true
}

func (this *LinkedList) Remove_Object(o any) bool {
	for x := this.first; x != nil; x = x.next {
		if objectsEqual(o, x.item) {
			this.unlink(x)
			return true
		}
	}
	return false
}

func (this *LinkedList) AddAll_Collection(c Collection) bool {
	return this.AddAll_int_Collection(this.size, c)
}

func (this *LinkedList) AddAll_int_Collection(index int32, c Collection) bool {
	this.checkPositionIndex(index)
	a := c.ToArray()
	if len(a) == 0 {
		return false
	}
	var pred, succ *linkedListNode
	if index == this.size {
		pred = this.last
	} else {
		succ = this.node(index)
		pred = succ.prev
	}
	for _, e := range a {
		n := &linkedListNode{item: e, prev: pred}
		if pred == nil {
			this.first = n
		} else {
			pred.next = n
		}
		pred = n
	}
	if succ == nil {
		this.last = pred
	} else {
		pred.next = succ
		succ.prev = pred
	}
	this.size += int32(len(a))
	this.modCount++
	return true
}

func (this *LinkedList) Clear() {
	for x := this.first; x != nil; {
		next := x.next
		x.item = nil
		x.next = nil
		x.prev = nil
		x = next
	}
	this.first = nil
	this.last = nil
	this.size = 0
	this.modCount++
}

func (this *LinkedList) Get_int(index int32) any {
	this.checkElementIndex(index)
	return this.node(index).item
}

func (this *LinkedList) Set_int_Object(index int32, e any) any {
	this.checkElementIndex(index)
	x := this.node(index)
	old := x.item
	x.item = e
	return old
}

func (this *LinkedList) Add_int_Object(index int32, e any) {
	this.checkPositionIndex(index)
	if index == this.size {
		this.linkLast(e)
	} else {
		this.linkBefore(e, this.node(index))
	}
}

func (this *LinkedList) Remove_int(index int32) any {
	this.checkElementIndex(index)
	return this.unlink(this.node(index))
}

func (this *LinkedList) checkElementIndex(index int32) {
	if !(index >= 0 && index < this.size) {
		panic(outOfBounds(index, this.size))
	}
}

func (this *LinkedList) checkPositionIndex(index int32) {
	if !(index >= 0 && index <= this.size) {
		panic(outOfBounds(index, this.size))
	}
}

// node returns the node at index.
func (this *LinkedList) node(index int32) *linkedListNode {
	if index < this.size>>1 {
		x := this.first
		for i := int32(0); i < index; i++ {
			x = x.next
		}
		return x
	}
	x := this.last
	for i := this.size - 1; i > index; i-- {
		x = x.prev
	}
	return x
}

func (this *LinkedList) IndexOf_Object(o any) int32 {
	index := int32(0)
	for x := this.first; x != nil; x = x.next {
		if objectsEqual(o, x.item) {
			return index
		}
		index++
	}
	return -1
}

func (this *LinkedList) LastIndexOf_Object(o any) int32 {
	index := this.size
	for x := this.last; x != nil; x = x.prev {
		index--
		if objectsEqual(o, x.item) {
			return index
		}
	}
	return -1
}

func (this *LinkedList) Peek() any {
	if this.first == nil {
		return nil
	}
	return this.first.item
}

func (this *LinkedList) Element() any { return this.GetFirst() }

func (this *LinkedList) Poll() any {
	if this.first == nil {
		return nil
	}
	return this.unlinkFirst(this.first)
}

func (this *LinkedList) Remove() any { return this.RemoveFirst() }

func (this *LinkedList) Offer_Object(e any) bool { return this.Add_Object(e) }

func (this *LinkedList) OfferFirst_Object(e any) bool {
	this.AddFirst_Object(e)
	return true
}

func (this *LinkedList) OfferLast_Object(e any) bool {
	this.AddLast_Object(e)
	return true
}

func (this *LinkedList) PeekFirst() any { return this.Peek() }

func (this *LinkedList) PeekLast() any {
	if this.last == nil {
		return nil
	}
	return this.last.item
}

func (this *LinkedList) PollFirst() any { return this.Poll() }

func (this *LinkedList) PollLast() any {
	if this.last == nil {
		return nil
	}
	return this.unlinkLast(this.last)
}

func (this *LinkedList) Push_Object(e any) { this.AddFirst_Object(e) }

func (this *LinkedList) Pop() any { return this.RemoveFirst() }

func (this *LinkedList) RemoveFirstOccurrence_Object(o any) bool { return this.Remove_Object(o) }

func (this *LinkedList) RemoveLastOccurrence_Object(o any) bool {
	for x := this.last; x != nil; x = x.prev {
		if objectsEqual(o, x.item) {
			this.unlink(x)
			return true
		}
	}
	return false
}

func (this *LinkedList) Iterator() Iterator { return this.ListIterator_int(0) }

func (this *LinkedList) ListIterator() ListIterator { return this.ListIterator_int(0) }

func (this *LinkedList) ListIterator_int(index int32) ListIterator {
	this.checkPositionIndex(index)
	return this.listItr(index)
}

func (this *LinkedList) listItr(index int32) *linkedListItr {
	it := &linkedListItr{list: this, nextIndex: index, expectedModCount: this.modCount}
	if index != this.size {
		it.next = this.node(index)
	}
	return it
}

func (this *LinkedList) DescendingIterator() Iterator {
	return &linkedListDescendingItr{itr: this.listItr(this.size)}
}

func (this *LinkedList) ToArray() []any {
	out := make([]any, 0, this.size)
	for x := this.first; x != nil; x = x.next {
		out = append(out, x.item)
	}
	return out
}

func (this *LinkedList) ToArray_ObjectArr(a any) any { return fillArray(a, this.ToArray()) }

// Clone implements LinkedList.clone.
func (this *LinkedList) Clone() any {
	c := &LinkedList{}
	for x := this.first; x != nil; x = x.next {
		c.Add_Object(x.item)
	}
	return c
}

func (this *LinkedList) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }
func (this *LinkedList) RemoveAll_Collection(c Collection) bool   { return acRemoveAll(this, c) }
func (this *LinkedList) RetainAll_Collection(c Collection) bool   { return acRetainAll(this, c) }

func (this *LinkedList) Sort_Comparator(c Comparator) { listSortDefault(this, c) }

func (this *LinkedList) SubList_int_int(from, to int32) List {
	return newAbstractSubList(this, &this.modCount, from, to)
}

func (this *LinkedList) Equals_Object(o any) bool { return alEquals(this, this, o) }
func (this *LinkedList) HashCode() int32          { return alHashCode(this) }
func (this *LinkedList) ToString() *String        { return acToString(this, this) }

// linkedListItr is LinkedList.ListItr.
type linkedListItr struct {
	list             *LinkedList
	lastReturned     *linkedListNode
	next             *linkedListNode
	nextIndex        int32
	expectedModCount int32
}

func (it *linkedListItr) checkForComodification() {
	if it.list.modCount != it.expectedModCount {
		panic(ConcurrentModificationException_new())
	}
}

func (it *linkedListItr) HasNext() bool { return it.nextIndex < it.list.size }

func (it *linkedListItr) Next() any {
	it.checkForComodification()
	if !it.HasNext() {
		panic(NoSuchElementException_new())
	}
	it.lastReturned = it.next
	it.next = it.next.next
	it.nextIndex++
	return it.lastReturned.item
}

func (it *linkedListItr) HasPrevious() bool { return it.nextIndex > 0 }

func (it *linkedListItr) Previous() any {
	it.checkForComodification()
	if !it.HasPrevious() {
		panic(NoSuchElementException_new())
	}
	if it.next == nil {
		it.next = it.list.last
	} else {
		it.next = it.next.prev
	}
	it.lastReturned = it.next
	it.nextIndex--
	return it.lastReturned.item
}

func (it *linkedListItr) NextIndex() int32     { return it.nextIndex }
func (it *linkedListItr) PreviousIndex() int32 { return it.nextIndex - 1 }

func (it *linkedListItr) Remove() {
	it.checkForComodification()
	if it.lastReturned == nil {
		panic(IllegalStateException_new())
	}
	lastNext := it.lastReturned.next
	it.list.unlink(it.lastReturned)
	if it.next == it.lastReturned {
		it.next = lastNext
	} else {
		it.nextIndex--
	}
	it.lastReturned = nil
	it.expectedModCount++
}

func (it *linkedListItr) Set_Object(e any) {
	if it.lastReturned == nil {
		panic(IllegalStateException_new())
	}
	it.checkForComodification()
	it.lastReturned.item = e
}

func (it *linkedListItr) Add_Object(e any) {
	it.checkForComodification()
	it.lastReturned = nil
	if it.next == nil {
		it.list.linkLast(e)
	} else {
		it.list.linkBefore(e, it.next)
	}
	it.nextIndex++
	it.expectedModCount++
}

type linkedListDescendingItr struct {
	itr *linkedListItr
}

func (it *linkedListDescendingItr) HasNext() bool { return it.itr.HasPrevious() }
func (it *linkedListDescendingItr) Next() any     { return it.itr.Previous() }
func (it *linkedListDescendingItr) Remove()       { it.itr.Remove() }
