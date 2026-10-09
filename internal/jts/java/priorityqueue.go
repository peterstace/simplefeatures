package java

// PriorityQueue is java.util.PriorityQueue, a binary heap in an array. The
// sift operations are Java's, which decide the order of elements that
// compare equal and the iteration order.
type PriorityQueue struct {
	queue      []any
	comparator Comparator
	modCount   int32
}

func (this *PriorityQueue) JClass_PriorityQueue() *PriorityQueue { return this }
func (this *PriorityQueue) JClassName() string                   { return "java.util.PriorityQueue" }
func (this *PriorityQueue) JIface_Collection()                   {}
func (this *PriorityQueue) JIface_Queue()                        {}

func PriorityQueue_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_PriorityQueue() *PriorityQueue })
	return ok
}

func PriorityQueue_fromAny(o any) *PriorityQueue {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_PriorityQueue() *PriorityQueue }](o).JClass_PriorityQueue()
}

func PriorityQueue_new() *PriorityQueue { return &PriorityQueue{} }

func PriorityQueue_new_int(initialCapacity int32) *PriorityQueue {
	return PriorityQueue_new_int_Comparator(initialCapacity, nil)
}

func PriorityQueue_new_Comparator(c Comparator) *PriorityQueue {
	return &PriorityQueue{comparator: c}
}

func PriorityQueue_new_int_Comparator(initialCapacity int32, c Comparator) *PriorityQueue {
	if initialCapacity < 1 {
		panic(IllegalArgumentException_new())
	}
	return &PriorityQueue{queue: make([]any, 0, initialCapacity), comparator: c}
}

func PriorityQueue_new_Collection(c Collection) *PriorityQueue {
	if ss, ok := c.(SortedSet); ok {
		this := &PriorityQueue{comparator: ss.Comparator()}
		this.initElementsFromCollection(ss)
		return this
	}
	if pq, ok := c.(interface{ JClass_PriorityQueue() *PriorityQueue }); ok {
		this := &PriorityQueue{comparator: pq.JClass_PriorityQueue().comparator}
		this.initFromPriorityQueue(c)
		return this
	}
	this := &PriorityQueue{}
	this.initFromCollection(c)
	return this
}

func PriorityQueue_new_PriorityQueue(c *PriorityQueue) *PriorityQueue {
	this := &PriorityQueue{comparator: c.comparator}
	this.initFromPriorityQueue(c)
	return this
}

func PriorityQueue_new_SortedSet(c SortedSet) *PriorityQueue {
	this := &PriorityQueue{comparator: c.Comparator()}
	this.initElementsFromCollection(c)
	return this
}

func (this *PriorityQueue) initFromPriorityQueue(c Collection) {
	if _, ok := c.(*PriorityQueue); ok {
		this.queue = c.ToArray()
		return
	}
	this.initFromCollection(c)
}

func (this *PriorityQueue) initElementsFromCollection(c Collection) {
	es := c.ToArray()
	if len(es) == 1 || this.comparator != nil {
		for _, e := range es {
			if e == nil {
				panic(NullPointerException_new())
			}
		}
	}
	this.queue = es
}

func (this *PriorityQueue) initFromCollection(c Collection) {
	this.initElementsFromCollection(c)
	this.heapify()
}

func (this *PriorityQueue) Add_Object(e any) bool { return this.Offer_Object(e) }

func (this *PriorityQueue) Offer_Object(e any) bool {
	if e == nil {
		panic(NullPointerException_new())
	}
	this.modCount++
	i := len(this.queue)
	this.queue = append(this.queue, nil)
	this.siftUp(i, e)
	return true
}

func (this *PriorityQueue) Peek() any {
	if len(this.queue) == 0 {
		return nil
	}
	return this.queue[0]
}

func (this *PriorityQueue) indexOf(o any) int {
	if o != nil {
		for i, e := range this.queue {
			if Object_Equals_Object(o, e) {
				return i
			}
		}
	}
	return -1
}

func (this *PriorityQueue) Remove_Object(o any) bool {
	i := this.indexOf(o)
	if i == -1 {
		return false
	}
	this.removeAt(i)
	return true
}

// removeEq removes the element identical to o.
func (this *PriorityQueue) removeEq(o any) {
	for i, e := range this.queue {
		if sameRef(o, e) {
			this.removeAt(i)
			break
		}
	}
}

func (this *PriorityQueue) Contains_Object(o any) bool { return this.indexOf(o) >= 0 }

func (this *PriorityQueue) ToArray() []any {
	return append(make([]any, 0, len(this.queue)), this.queue...)
}

func (this *PriorityQueue) ToArray_ObjectArr(a any) any { return fillArray(a, this.queue) }

func (this *PriorityQueue) Iterator() Iterator {
	return &priorityQueueItr{pq: this, lastRet: -1, expectedModCount: this.modCount}
}

// priorityQueueItr is PriorityQueue.Itr. It iterates in array order; when
// removing an element moves a later element before the cursor, the moved
// element is returned at the end instead.
type priorityQueueItr struct {
	pq               *PriorityQueue
	cursor           int
	lastRet          int
	forgetMeNot      *ArrayDeque
	lastRetElt       any
	expectedModCount int32
}

func (it *priorityQueueItr) HasNext() bool {
	return it.cursor < len(it.pq.queue) || it.forgetMeNot != nil && !it.forgetMeNot.IsEmpty()
}

func (it *priorityQueueItr) Next() any {
	if it.expectedModCount != it.pq.modCount {
		panic(ConcurrentModificationException_new())
	}
	if it.cursor < len(it.pq.queue) {
		it.lastRet = it.cursor
		it.cursor++
		return it.pq.queue[it.lastRet]
	}
	if it.forgetMeNot != nil {
		it.lastRet = -1
		it.lastRetElt = it.forgetMeNot.Poll()
		if it.lastRetElt != nil {
			return it.lastRetElt
		}
	}
	panic(NoSuchElementException_new())
}

func (it *priorityQueueItr) Remove() {
	if it.expectedModCount != it.pq.modCount {
		panic(ConcurrentModificationException_new())
	}
	switch {
	case it.lastRet != -1:
		moved := it.pq.removeAt(it.lastRet)
		it.lastRet = -1
		if moved == nil {
			it.cursor--
		} else {
			if it.forgetMeNot == nil {
				it.forgetMeNot = ArrayDeque_new()
			}
			it.forgetMeNot.Add_Object(moved)
		}
	case it.lastRetElt != nil:
		it.pq.removeEq(it.lastRetElt)
		it.lastRetElt = nil
	default:
		panic(IllegalStateException_new())
	}
	it.expectedModCount = it.pq.modCount
}

func (this *PriorityQueue) Size() int32   { return int32(len(this.queue)) }
func (this *PriorityQueue) IsEmpty() bool { return len(this.queue) == 0 }

func (this *PriorityQueue) Clear() {
	this.modCount++
	for i := range this.queue {
		this.queue[i] = nil
	}
	this.queue = this.queue[:0]
}

func (this *PriorityQueue) Poll() any {
	es := this.queue
	if len(es) == 0 {
		return nil
	}
	result := es[0]
	this.modCount++
	n := len(es) - 1
	x := es[n]
	es[n] = nil
	this.queue = es[:n]
	if n > 0 {
		if this.comparator == nil {
			siftDownComparable(0, x, this.queue, n)
		} else {
			siftDownUsingComparator(0, x, this.queue, n, this.comparator)
		}
	}
	return result
}

// removeAt removes the element at index i. It returns the element that moved
// from the end of the array to before i, or nil if none did.
func (this *PriorityQueue) removeAt(i int) any {
	es := this.queue
	this.modCount++
	s := len(es) - 1
	if s == i {
		es[i] = nil
		this.queue = es[:s]
		return nil
	}
	moved := es[s]
	es[s] = nil
	this.queue = es[:s]
	this.siftDown(i, moved)
	if sameRef(this.queue[i], moved) {
		this.siftUp(i, moved)
		if !sameRef(this.queue[i], moved) {
			return moved
		}
	}
	return nil
}

func (this *PriorityQueue) siftUp(k int, x any) {
	if this.comparator != nil {
		siftUpUsingComparator(k, x, this.queue, this.comparator)
	} else {
		siftUpComparable(k, x, this.queue)
	}
}

func siftUpComparable(k int, x any, es []any) {
	requireComparable(x)
	for k > 0 {
		parent := (k - 1) >> 1
		e := es[parent]
		if Comparable_CompareTo_Object(x, e) >= 0 {
			break
		}
		es[k] = e
		k = parent
	}
	es[k] = x
}

func siftUpUsingComparator(k int, x any, es []any, c Comparator) {
	for k > 0 {
		parent := (k - 1) >> 1
		e := es[parent]
		if c.Compare_Object_Object(x, e) >= 0 {
			break
		}
		es[k] = e
		k = parent
	}
	es[k] = x
}

func (this *PriorityQueue) siftDown(k int, x any) {
	if this.comparator != nil {
		siftDownUsingComparator(k, x, this.queue, len(this.queue), this.comparator)
	} else {
		siftDownComparable(k, x, this.queue, len(this.queue))
	}
}

func siftDownComparable(k int, x any, es []any, n int) {
	requireComparable(x)
	half := n >> 1
	for k < half {
		child := (k << 1) + 1
		c := es[child]
		right := child + 1
		if right < n && Comparable_CompareTo_Object(c, es[right]) > 0 {
			child = right
			c = es[child]
		}
		if Comparable_CompareTo_Object(x, c) <= 0 {
			break
		}
		es[k] = c
		k = child
	}
	es[k] = x
}

func siftDownUsingComparator(k int, x any, es []any, n int, cmp Comparator) {
	half := n >> 1
	for k < half {
		child := (k << 1) + 1
		c := es[child]
		right := child + 1
		if right < n && cmp.Compare_Object_Object(c, es[right]) > 0 {
			child = right
			c = es[child]
		}
		if cmp.Compare_Object_Object(x, c) <= 0 {
			break
		}
		es[k] = c
		k = child
	}
	es[k] = x
}

func (this *PriorityQueue) heapify() {
	es := this.queue
	n := len(es)
	i := (n >> 1) - 1
	if this.comparator == nil {
		for ; i >= 0; i-- {
			siftDownComparable(i, es[i], es, n)
		}
	} else {
		for ; i >= 0; i-- {
			siftDownUsingComparator(i, es[i], es, n, this.comparator)
		}
	}
}

func (this *PriorityQueue) Comparator() Comparator { return this.comparator }

func (this *PriorityQueue) RemoveAll_Collection(c Collection) bool {
	if c == nil {
		panic(NullPointerException_new())
	}
	return this.bulkRemove(func(e any) bool { return c.Contains_Object(e) })
}

func (this *PriorityQueue) RetainAll_Collection(c Collection) bool {
	if c == nil {
		panic(NullPointerException_new())
	}
	return this.bulkRemove(func(e any) bool { return !c.Contains_Object(e) })
}

// bulkRemove is PriorityQueue.bulkRemove: it tests every element once, in
// array order, removes those that matched, and heapifies the survivors.
func (this *PriorityQueue) bulkRemove(filter func(any) bool) bool {
	this.modCount++
	expectedModCount := this.modCount
	es := this.queue
	end := len(es)
	i := 0
	for i < end && !filter(es[i]) {
		i++
	}
	if i >= end {
		if this.modCount != expectedModCount {
			panic(ConcurrentModificationException_new())
		}
		return false
	}
	beg := i
	dead := make([]bool, end-beg)
	dead[0] = true
	for i = beg + 1; i < end; i++ {
		if filter(es[i]) {
			dead[i-beg] = true
		}
	}
	if this.modCount != expectedModCount {
		panic(ConcurrentModificationException_new())
	}
	w := beg
	for i = beg; i < end; i++ {
		if !dead[i-beg] {
			es[w] = es[i]
			w++
		}
	}
	for i = w; i < end; i++ {
		es[i] = nil
	}
	this.queue = es[:w]
	this.heapify()
	return true
}

// Remove is AbstractQueue.remove.
func (this *PriorityQueue) Remove() any {
	x := this.Poll()
	if x == nil {
		panic(NoSuchElementException_new())
	}
	return x
}

// Element is AbstractQueue.element.
func (this *PriorityQueue) Element() any {
	x := this.Peek()
	if x == nil {
		panic(NoSuchElementException_new())
	}
	return x
}

// AddAll_Collection is AbstractQueue.addAll.
func (this *PriorityQueue) AddAll_Collection(c Collection) bool {
	if c == nil {
		panic(NullPointerException_new())
	}
	if sameRef(c, this) {
		panic(IllegalArgumentException_new())
	}
	return acAddAll(this, c)
}

func (this *PriorityQueue) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }

func (this *PriorityQueue) ToString() *String { return acToString(this, this) }
