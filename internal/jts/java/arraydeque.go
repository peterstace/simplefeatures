package java

// ArrayDeque is java.util.ArrayDeque: a circular buffer whose capacity grows
// as Java's does, so that iterators observe concurrent modifications as
// Java's do.
type ArrayDeque struct {
	elements []any
	head     int
	tail     int
}

func (this *ArrayDeque) JClass_ArrayDeque() *ArrayDeque { return this }
func (this *ArrayDeque) JClassName() string             { return "java.util.ArrayDeque" }
func (this *ArrayDeque) JIface_Collection()             {}
func (this *ArrayDeque) JIface_Queue()                  {}
func (this *ArrayDeque) JIface_Deque()                  {}

func ArrayDeque_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_ArrayDeque() *ArrayDeque })
	return ok
}

func ArrayDeque_fromAny(o any) *ArrayDeque {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_ArrayDeque() *ArrayDeque }](o).JClass_ArrayDeque()
}

func ArrayDeque_new() *ArrayDeque {
	return &ArrayDeque{elements: make([]any, 16+1)}
}

func ArrayDeque_new_int(numElements int32) *ArrayDeque {
	n := int(numElements) + 1
	if numElements < 1 {
		n = 1
	} else if numElements == 0x7fffffff {
		n = 0x7fffffff
	}
	return &ArrayDeque{elements: make([]any, n)}
}

func ArrayDeque_new_Collection(c Collection) *ArrayDeque {
	this := ArrayDeque_new_int(c.Size())
	this.copyElements(c)
	return this
}

const arrayDequeMaxArraySize = 0x7fffffff - 8

func (this *ArrayDeque) grow(needed int) {
	oldCapacity := len(this.elements)
	var newCapacity int
	jump := oldCapacity >> 1
	if oldCapacity < 64 {
		jump = oldCapacity + 2
	}
	if jump < needed || oldCapacity+jump-arrayDequeMaxArraySize > 0 {
		newCapacity = this.newCapacity(needed, jump)
	} else {
		newCapacity = oldCapacity + jump
	}
	es := make([]any, newCapacity)
	copy(es, this.elements)
	this.elements = es
	if this.tail < this.head || (this.tail == this.head && es[this.head] != nil) {
		newSpace := newCapacity - oldCapacity
		copy(es[this.head+newSpace:], es[this.head:oldCapacity])
		to := this.head + newSpace
		for i := this.head; i < to; i++ {
			es[i] = nil
		}
		this.head = to
	}
}

func (this *ArrayDeque) newCapacity(needed, jump int) int {
	oldCapacity := len(this.elements)
	minCapacity := oldCapacity + needed
	if minCapacity-arrayDequeMaxArraySize > 0 {
		if minCapacity < 0 {
			panic(IllegalStateException_new_String(Lit("Sorry, deque too big")))
		}
		return 0x7fffffff
	}
	if needed > jump {
		return minCapacity
	}
	if oldCapacity+jump-arrayDequeMaxArraySize < 0 {
		return oldCapacity + jump
	}
	return arrayDequeMaxArraySize
}

func dequeInc(i, modulus int) int {
	i++
	if i >= modulus {
		i = 0
	}
	return i
}

func dequeDec(i, modulus int) int {
	i--
	if i < 0 {
		i = modulus - 1
	}
	return i
}

func dequeSub(i, j, modulus int) int {
	i -= j
	if i < 0 {
		i += modulus
	}
	return i
}

func (this *ArrayDeque) AddFirst_Object(e any) {
	if e == nil {
		panic(NullPointerException_new())
	}
	es := this.elements
	this.head = dequeDec(this.head, len(es))
	es[this.head] = e
	if this.head == this.tail {
		this.grow(1)
	}
}

func (this *ArrayDeque) AddLast_Object(e any) {
	if e == nil {
		panic(NullPointerException_new())
	}
	es := this.elements
	es[this.tail] = e
	this.tail = dequeInc(this.tail, len(es))
	if this.head == this.tail {
		this.grow(1)
	}
}

func (this *ArrayDeque) AddAll_Collection(c Collection) bool {
	s := this.size()
	if needed := s + int(c.Size()) + 1 - len(this.elements); needed > 0 {
		this.grow(needed)
	}
	this.copyElements(c)
	return this.size() > s
}

func (this *ArrayDeque) copyElements(c Collection) {
	for it := c.Iterator(); it.HasNext(); {
		this.AddLast_Object(it.Next())
	}
}

func (this *ArrayDeque) OfferFirst_Object(e any) bool {
	this.AddFirst_Object(e)
	return true
}

func (this *ArrayDeque) OfferLast_Object(e any) bool {
	this.AddLast_Object(e)
	return true
}

func (this *ArrayDeque) RemoveFirst() any {
	e := this.PollFirst()
	if e == nil {
		panic(NoSuchElementException_new())
	}
	return e
}

func (this *ArrayDeque) RemoveLast() any {
	e := this.PollLast()
	if e == nil {
		panic(NoSuchElementException_new())
	}
	return e
}

func (this *ArrayDeque) PollFirst() any {
	es := this.elements
	h := this.head
	e := es[h]
	if e != nil {
		es[h] = nil
		this.head = dequeInc(h, len(es))
	}
	return e
}

func (this *ArrayDeque) PollLast() any {
	es := this.elements
	t := dequeDec(this.tail, len(es))
	e := es[t]
	if e != nil {
		this.tail = t
		es[t] = nil
	}
	return e
}

func (this *ArrayDeque) GetFirst() any {
	e := this.elements[this.head]
	if e == nil {
		panic(NoSuchElementException_new())
	}
	return e
}

func (this *ArrayDeque) GetLast() any {
	es := this.elements
	e := es[dequeDec(this.tail, len(es))]
	if e == nil {
		panic(NoSuchElementException_new())
	}
	return e
}

func (this *ArrayDeque) PeekFirst() any { return this.elements[this.head] }

func (this *ArrayDeque) PeekLast() any {
	es := this.elements
	return es[dequeDec(this.tail, len(es))]
}

func (this *ArrayDeque) RemoveFirstOccurrence_Object(o any) bool {
	if o != nil {
		es := this.elements
		end := this.tail
		i := this.head
		to := end
		if i > end {
			to = len(es)
		}
		for {
			for ; i < to; i++ {
				if Object_Equals_Object(o, es[i]) {
					this.delete(i)
					return true
				}
			}
			if to == end {
				break
			}
			i, to = 0, end
		}
	}
	return false
}

func (this *ArrayDeque) RemoveLastOccurrence_Object(o any) bool {
	if o != nil {
		es := this.elements
		end := this.head
		i := this.tail
		to := end
		if i < end {
			to = 0
		}
		for {
			for i--; i > to-1; i-- {
				if Object_Equals_Object(o, es[i]) {
					this.delete(i)
					return true
				}
			}
			if to == end {
				break
			}
			i, to = len(es), end
		}
	}
	return false
}

func (this *ArrayDeque) Add_Object(e any) bool {
	this.AddLast_Object(e)
	return true
}

func (this *ArrayDeque) Offer_Object(e any) bool { return this.OfferLast_Object(e) }
func (this *ArrayDeque) Remove() any             { return this.RemoveFirst() }
func (this *ArrayDeque) Poll() any               { return this.PollFirst() }
func (this *ArrayDeque) Element() any            { return this.GetFirst() }
func (this *ArrayDeque) Peek() any               { return this.PeekFirst() }
func (this *ArrayDeque) Push_Object(e any)       { this.AddFirst_Object(e) }
func (this *ArrayDeque) Pop() any                { return this.RemoveFirst() }

// delete removes element i, moving whichever side is shorter. It reports
// whether the elements after i moved back one place.
func (this *ArrayDeque) delete(i int) bool {
	es := this.elements
	capacity := len(es)
	h := this.head
	t := this.tail
	front := dequeSub(i, h, capacity)
	back := dequeSub(t, i, capacity) - 1
	if front < back {
		if h <= i {
			copy(es[h+1:h+1+front], es[h:h+front])
		} else {
			copy(es[1:1+i], es[0:i])
			es[0] = es[capacity-1]
			n := front - (i + 1)
			copy(es[h+1:h+1+n], es[h:h+n])
		}
		es[h] = nil
		this.head = dequeInc(h, capacity)
		return false
	}
	this.tail = dequeDec(t, capacity)
	if i <= this.tail {
		copy(es[i:i+back], es[i+1:i+1+back])
	} else {
		n := capacity - (i + 1)
		copy(es[i:i+n], es[i+1:i+1+n])
		es[capacity-1] = es[0]
		copy(es[0:t-1], es[1:t])
	}
	es[this.tail] = nil
	return true
}

func (this *ArrayDeque) size() int {
	return dequeSub(this.tail, this.head, len(this.elements))
}

func (this *ArrayDeque) Size() int32   { return int32(this.size()) }
func (this *ArrayDeque) IsEmpty() bool { return this.head == this.tail }

func (this *ArrayDeque) Iterator() Iterator {
	return &arrayDequeItr{deque: this, cursor: this.head, remaining: this.size(), lastRet: -1}
}

func (this *ArrayDeque) DescendingIterator() Iterator {
	return &arrayDequeItr{
		deque:      this,
		cursor:     dequeDec(this.tail, len(this.elements)),
		remaining:  this.size(),
		lastRet:    -1,
		descending: true,
	}
}

// arrayDequeItr is ArrayDeque.DeqIterator and ArrayDeque.DescendingIterator.
type arrayDequeItr struct {
	deque      *ArrayDeque
	cursor     int
	remaining  int
	lastRet    int
	descending bool
}

func (it *arrayDequeItr) HasNext() bool { return it.remaining > 0 }

func (it *arrayDequeItr) Next() any {
	if it.remaining <= 0 {
		panic(NoSuchElementException_new())
	}
	es := it.deque.elements
	e := es[it.cursor]
	if e == nil {
		panic(ConcurrentModificationException_new())
	}
	it.lastRet = it.cursor
	if it.descending {
		it.cursor = dequeDec(it.cursor, len(es))
	} else {
		it.cursor = dequeInc(it.cursor, len(es))
	}
	it.remaining--
	return e
}

func (it *arrayDequeItr) Remove() {
	if it.lastRet < 0 {
		panic(IllegalStateException_new())
	}
	leftShifted := it.deque.delete(it.lastRet)
	if it.descending {
		if !leftShifted {
			it.cursor = dequeInc(it.cursor, len(it.deque.elements))
		}
	} else if leftShifted {
		it.cursor = dequeDec(it.cursor, len(it.deque.elements))
	}
	it.lastRet = -1
}

func (this *ArrayDeque) Contains_Object(o any) bool {
	if o != nil {
		es := this.elements
		end := this.tail
		i := this.head
		to := end
		if i > end {
			to = len(es)
		}
		for {
			for ; i < to; i++ {
				if Object_Equals_Object(o, es[i]) {
					return true
				}
			}
			if to == end {
				break
			}
			i, to = 0, end
		}
	}
	return false
}

func (this *ArrayDeque) Remove_Object(o any) bool { return this.RemoveFirstOccurrence_Object(o) }

func (this *ArrayDeque) Clear() {
	es := this.elements
	for i := this.head; i != this.tail; i = dequeInc(i, len(es)) {
		es[i] = nil
	}
	this.head = 0
	this.tail = 0
}

func (this *ArrayDeque) ToArray() []any {
	out := make([]any, 0, this.size())
	es := this.elements
	for i := this.head; i != this.tail; i = dequeInc(i, len(es)) {
		out = append(out, es[i])
	}
	return out
}

func (this *ArrayDeque) ToArray_ObjectArr(a any) any { return fillArray(a, this.ToArray()) }

func (this *ArrayDeque) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }

func (this *ArrayDeque) RemoveAll_Collection(c Collection) bool {
	if c == nil {
		panic(NullPointerException_new())
	}
	return this.bulkRemove(func(e any) bool { return c.Contains_Object(e) })
}

func (this *ArrayDeque) RetainAll_Collection(c Collection) bool {
	if c == nil {
		panic(NullPointerException_new())
	}
	return this.bulkRemove(func(e any) bool { return !c.Contains_Object(e) })
}

// bulkRemove is ArrayDeque.bulkRemove: it tests every element once, in
// order, then closes the gaps by moving the survivors that follow the first
// removed element towards it.
func (this *ArrayDeque) bulkRemove(filter func(any) bool) bool {
	es := this.elements
	capacity := len(es)
	n := this.size()
	beg := -1
	k := 0
	for ; k < n; k++ {
		if filter(es[(this.head+k)%capacity]) {
			beg = k
			break
		}
	}
	if beg < 0 {
		return false
	}
	dead := make([]bool, n-beg)
	dead[0] = true
	for k = beg + 1; k < n; k++ {
		if filter(es[(this.head+k)%capacity]) {
			dead[k-beg] = true
		}
	}
	w := beg
	for k = beg + 1; k < n; k++ {
		if !dead[k-beg] {
			es[(this.head+w)%capacity] = es[(this.head+k)%capacity]
			w++
		}
	}
	newTail := (this.head + w) % capacity
	for i := newTail; i != this.tail; i = dequeInc(i, capacity) {
		es[i] = nil
	}
	this.tail = newTail
	return true
}

// Clone implements ArrayDeque.clone.
func (this *ArrayDeque) Clone() any {
	es := make([]any, len(this.elements))
	copy(es, this.elements)
	return &ArrayDeque{elements: es, head: this.head, tail: this.tail}
}

func (this *ArrayDeque) ToString() *String { return acToString(this, this) }
