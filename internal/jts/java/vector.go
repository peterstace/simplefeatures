package java

// Vector is java.util.Vector. The length of Java's backing array is tracked
// in capacity: Capacity returns it, and it appears in the messages of the
// exceptions that negative indexes cause.
type Vector struct {
	jself             any
	elems             []any
	capacity          int32
	capacityIncrement int32
	modCount          int32
}

func (this *Vector) JSelf() any        { return this.jself }
func (this *Vector) JSetSelf(self any) { this.jself = self }

func (this *Vector) JClass_Vector() *Vector { return this }
func (this *Vector) JClassName() string     { return "java.util.Vector" }
func (this *Vector) JIface_Collection()     {}
func (this *Vector) JIface_List()           {}
func (this *Vector) JIface_RandomAccess()   {}

func Vector_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_Vector() *Vector })
	return ok
}

func Vector_fromAny(o any) *Vector {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_Vector() *Vector }](o).JClass_Vector()
}

func (this *Vector) self() any {
	if this.jself != nil {
		return this.jself
	}
	return this
}

func Vector_new() *Vector {
	this := &Vector{}
	this.Vector_init()
	return this
}

func Vector_new_int(initialCapacity int32) *Vector {
	this := &Vector{}
	this.Vector_init_int(initialCapacity)
	return this
}

func Vector_new_int_int(initialCapacity, capacityIncrement int32) *Vector {
	this := &Vector{}
	this.Vector_init_int_int(initialCapacity, capacityIncrement)
	return this
}

func Vector_new_Collection(c Collection) *Vector {
	this := &Vector{}
	this.Vector_init_Collection(c)
	return this
}

func (this *Vector) Vector_init() { this.Vector_init_int_int(10, 0) }

func (this *Vector) Vector_init_int(initialCapacity int32) {
	this.Vector_init_int_int(initialCapacity, 0)
}

func (this *Vector) Vector_init_int_int(initialCapacity, capacityIncrement int32) {
	if initialCapacity < 0 {
		panic(IllegalArgumentException_new_String(Concat(Lit("Illegal Capacity: "), StrOfInt(initialCapacity))))
	}
	this.capacity = initialCapacity
	this.capacityIncrement = capacityIncrement
}

func (this *Vector) Vector_init_Collection(c Collection) {
	this.elems = c.ToArray()
	this.capacity = int32(len(this.elems))
}

// arrayIndexOutOfBounds is the exception the JVM throws for an access to
// element i of Vector's backing array.
func (this *Vector) arrayIndexOutOfBounds(i int32) any {
	return ArrayIndexOutOfBoundsException_new_String(FromGo(
		"Index " + StrOfInt(i).Go() + " out of bounds for length " + StrOfInt(this.capacity).Go()))
}

// elementData reads element i of the backing array, which must be below the
// element count if it is non-negative.
func (this *Vector) elementData(i int32) any {
	if i < 0 {
		panic(this.arrayIndexOutOfBounds(i))
	}
	return this.elems[i]
}

// grow is Vector.grow(int), which sets the capacity as
// ArraysSupport.newLength does.
func (this *Vector) grow(minCapacity int32) {
	old := this.capacity
	pref := old
	if this.capacityIncrement > 0 {
		pref = this.capacityIncrement
	}
	growth := minCapacity - old
	if pref > growth {
		growth = pref
	}
	this.capacity = old + growth
}

func (this *Vector) CopyInto_ObjectArr(a any) {
	v := arrayValue(a)
	if v.Len() < len(this.elems) {
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(
			"arraycopy: last destination index " + StrOfInt(int32(len(this.elems))).Go() +
				" out of bounds for object array[" + StrOfInt(int32(v.Len())).Go() + "]")))
	}
	for i, e := range this.elems {
		arraySet(v, i, e)
	}
}

func (this *Vector) TrimToSize() {
	this.modCount++
	if int32(len(this.elems)) < this.capacity {
		this.capacity = int32(len(this.elems))
	}
}

func (this *Vector) EnsureCapacity_int(minCapacity int32) {
	if minCapacity > 0 {
		this.modCount++
		if minCapacity > this.capacity {
			this.grow(minCapacity)
		}
	}
}

func (this *Vector) SetSize_int(newSize int32) {
	this.modCount++
	if newSize > this.capacity {
		this.grow(newSize)
	}
	if newSize < 0 {
		panic(this.arrayIndexOutOfBounds(newSize))
	}
	n := int32(len(this.elems))
	for i := newSize; i < n; i++ {
		this.elems[i] = nil
	}
	if newSize <= n {
		this.elems = this.elems[:newSize]
	} else {
		this.elems = append(this.elems, make([]any, newSize-n)...)
	}
}

func (this *Vector) Capacity() int32 { return this.capacity }
func (this *Vector) Size() int32     { return int32(len(this.elems)) }
func (this *Vector) IsEmpty() bool   { return len(this.elems) == 0 }

func (this *Vector) Contains_Object(o any) bool { return this.IndexOf_Object_int(o, 0) >= 0 }

func (this *Vector) IndexOf_Object(o any) int32 { return this.IndexOf_Object_int(o, 0) }

func (this *Vector) IndexOf_Object_int(o any, index int32) int32 {
	n := int32(len(this.elems))
	if index < 0 && index < n {
		panic(this.arrayIndexOutOfBounds(index))
	}
	if o == nil {
		for i := index; i < n; i++ {
			if this.elems[i] == nil {
				return i
			}
		}
	} else {
		for i := index; i < n; i++ {
			if Object_Equals_Object(o, this.elems[i]) {
				return i
			}
		}
	}
	return -1
}

func (this *Vector) LastIndexOf_Object(o any) int32 {
	return this.LastIndexOf_Object_int(o, int32(len(this.elems))-1)
}

func (this *Vector) LastIndexOf_Object_int(o any, index int32) int32 {
	n := int32(len(this.elems))
	if index >= n {
		panic(IndexOutOfBoundsException_new_String(FromGo(StrOfInt(index).Go() + " >= " + StrOfInt(n).Go())))
	}
	if o == nil {
		for i := index; i >= 0; i-- {
			if this.elems[i] == nil {
				return i
			}
		}
	} else {
		for i := index; i >= 0; i-- {
			if Object_Equals_Object(o, this.elems[i]) {
				return i
			}
		}
	}
	return -1
}

func (this *Vector) ElementAt_int(index int32) any {
	n := int32(len(this.elems))
	if index >= n {
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(StrOfInt(index).Go() + " >= " + StrOfInt(n).Go())))
	}
	return this.elementData(index)
}

func (this *Vector) FirstElement() any {
	if len(this.elems) == 0 {
		panic(NoSuchElementException_new())
	}
	return this.elems[0]
}

func (this *Vector) LastElement() any {
	if len(this.elems) == 0 {
		panic(NoSuchElementException_new())
	}
	return this.elems[len(this.elems)-1]
}

func (this *Vector) SetElementAt_Object_int(o any, index int32) {
	n := int32(len(this.elems))
	if index >= n {
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(StrOfInt(index).Go() + " >= " + StrOfInt(n).Go())))
	}
	if index < 0 {
		panic(this.arrayIndexOutOfBounds(index))
	}
	this.elems[index] = o
}

func (this *Vector) RemoveElementAt_int(index int32) {
	n := int32(len(this.elems))
	if index >= n {
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(StrOfInt(index).Go() + " >= " + StrOfInt(n).Go())))
	}
	if index < 0 {
		panic(ArrayIndexOutOfBoundsException_new_int(index))
	}
	this.modCount++
	this.deleteAt(index)
}

func (this *Vector) deleteAt(index int32) {
	n := len(this.elems) - 1
	copy(this.elems[index:], this.elems[index+1:])
	this.elems[n] = nil
	this.elems = this.elems[:n]
}

func (this *Vector) InsertElementAt_Object_int(o any, index int32) {
	n := int32(len(this.elems))
	if index > n {
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(StrOfInt(index).Go() + " > " + StrOfInt(n).Go())))
	}
	this.modCount++
	if n == this.capacity {
		this.grow(n + 1)
	}
	if index < 0 {
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(
			"arraycopy: source index " + StrOfInt(index).Go() + " out of bounds for object array[" +
				StrOfInt(this.capacity).Go() + "]")))
	}
	this.elems = append(this.elems, nil)
	copy(this.elems[index+1:], this.elems[index:])
	this.elems[index] = o
}

func (this *Vector) AddElement_Object(o any) {
	this.modCount++
	this.appendElement(o)
}

func (this *Vector) appendElement(o any) {
	n := int32(len(this.elems))
	if n == this.capacity {
		this.grow(n + 1)
	}
	this.elems = append(this.elems, o)
}

func (this *Vector) RemoveElement_Object(o any) bool {
	this.modCount++
	i := this.IndexOf_Object(o)
	if i >= 0 {
		this.RemoveElementAt_int(i)
		return true
	}
	return false
}

func (this *Vector) RemoveAllElements() {
	for i := range this.elems {
		this.elems[i] = nil
	}
	this.elems = this.elems[:0]
	this.modCount++
}

// Clone implements Vector.clone, a shallow copy of the whole object.
func (this *Vector) Clone() any {
	var c any
	var cl *Vector
	if this.jself != nil && !sameRef(this.jself, this) {
		c = ObjectImpl_Clone(this.jself)
		cl = c.(interface{ JClass_Vector() *Vector }).JClass_Vector()
		cl.jself = c
	} else {
		cl = &Vector{capacityIncrement: this.capacityIncrement}
		c = cl
	}
	cl.elems = append(make([]any, 0, len(this.elems)), this.elems...)
	cl.capacity = int32(len(this.elems))
	cl.modCount = 0
	return c
}

func (this *Vector) ToArray() []any {
	return append(make([]any, 0, len(this.elems)), this.elems...)
}

func (this *Vector) ToArray_ObjectArr(a any) any { return fillArray(a, this.elems) }

func (this *Vector) Get_int(index int32) any {
	if index >= int32(len(this.elems)) {
		panic(ArrayIndexOutOfBoundsException_new_int(index))
	}
	return this.elementData(index)
}

func (this *Vector) Set_int_Object(index int32, o any) any {
	if index >= int32(len(this.elems)) {
		panic(ArrayIndexOutOfBoundsException_new_int(index))
	}
	old := this.elementData(index)
	this.elems[index] = o
	return old
}

func (this *Vector) Add_Object(o any) bool {
	this.modCount++
	this.appendElement(o)
	return true
}

func (this *Vector) Remove_Object(o any) bool { return this.RemoveElement_Object(o) }

func (this *Vector) Add_int_Object(index int32, o any) { this.InsertElementAt_Object_int(o, index) }

func (this *Vector) Remove_int(index int32) any {
	this.modCount++
	if index >= int32(len(this.elems)) {
		panic(ArrayIndexOutOfBoundsException_new_int(index))
	}
	old := this.elementData(index)
	this.deleteAt(index)
	return old
}

func (this *Vector) Clear() { this.RemoveAllElements() }

func (this *Vector) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }

func (this *Vector) AddAll_Collection(c Collection) bool {
	a := c.ToArray()
	this.modCount++
	if len(a) == 0 {
		return false
	}
	n := int32(len(this.elems))
	if int32(len(a)) > this.capacity-n {
		this.grow(n + int32(len(a)))
	}
	this.elems = append(this.elems, a...)
	return true
}

func (this *Vector) AddAll_int_Collection(index int32, c Collection) bool {
	n := int32(len(this.elems))
	if index < 0 || index > n {
		panic(ArrayIndexOutOfBoundsException_new_int(index))
	}
	a := c.ToArray()
	this.modCount++
	if len(a) == 0 {
		return false
	}
	if int32(len(a)) > this.capacity-n {
		this.grow(n + int32(len(a)))
	}
	this.elems = append(this.elems, a...)
	copy(this.elems[int(index)+len(a):], this.elems[index:n])
	copy(this.elems[index:], a)
	return true
}

func (this *Vector) RemoveAll_Collection(c Collection) bool {
	if c == nil {
		panic(NullPointerException_new())
	}
	return this.bulkRemove(func(e any) bool { return c.Contains_Object(e) })
}

func (this *Vector) RetainAll_Collection(c Collection) bool {
	if c == nil {
		panic(NullPointerException_new())
	}
	return this.bulkRemove(func(e any) bool { return !c.Contains_Object(e) })
}

// bulkRemove is Vector.bulkRemove: it tests every element once, in order,
// then removes those that matched.
func (this *Vector) bulkRemove(filter func(any) bool) bool {
	expectedModCount := this.modCount
	es := this.elems
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
	this.modCount++
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
	this.elems = es[:w]
	return true
}

func (this *Vector) Equals_Object(o any) bool { return alEquals(this, this.self(), o) }
func (this *Vector) HashCode() int32          { return alHashCode(this) }
func (this *Vector) ToString() *String        { return acToString(this, this.self()) }

func (this *Vector) SubList_int_int(from, to int32) List {
	return newAbstractSubList(this, &this.modCount, from, to)
}

// removeRange is Vector.removeRange.
func (this *Vector) removeRange(from, to int32) {
	this.modCount++
	n := len(this.elems)
	copy(this.elems[from:], this.elems[to:])
	newSize := n - int(to-from)
	for i := newSize; i < n; i++ {
		this.elems[i] = nil
	}
	this.elems = this.elems[:newSize]
}

func (this *Vector) Sort_Comparator(c Comparator) {
	expectedModCount := this.modCount
	timSort(this.elems, 0, len(this.elems), c)
	if this.modCount != expectedModCount {
		panic(ConcurrentModificationException_new())
	}
	this.modCount++
}

func (this *Vector) Iterator() Iterator {
	return &vectorItr{vec: this, lastRet: -1, expectedModCount: this.modCount}
}

func (this *Vector) ListIterator() ListIterator { return this.ListIterator_int(0) }

func (this *Vector) ListIterator_int(index int32) ListIterator {
	if index < 0 || index > int32(len(this.elems)) {
		panic(IndexOutOfBoundsException_new_String(Concat(Lit("Index: "), StrOfInt(index))))
	}
	return &vectorItr{vec: this, cursor: index, lastRet: -1, expectedModCount: this.modCount}
}

// vectorItr is Vector.Itr and Vector.ListItr.
type vectorItr struct {
	vec              *Vector
	cursor           int32
	lastRet          int32
	expectedModCount int32
}

func (it *vectorItr) checkForComodification() {
	if it.vec.modCount != it.expectedModCount {
		panic(ConcurrentModificationException_new())
	}
}

func (it *vectorItr) HasNext() bool { return it.cursor != int32(len(it.vec.elems)) }

func (it *vectorItr) Next() any {
	it.checkForComodification()
	i := it.cursor
	if i >= int32(len(it.vec.elems)) {
		panic(NoSuchElementException_new())
	}
	it.cursor = i + 1
	it.lastRet = i
	return it.vec.elementData(i)
}

func (it *vectorItr) Remove() {
	if it.lastRet == -1 {
		panic(IllegalStateException_new())
	}
	it.checkForComodification()
	it.vec.Remove_int(it.lastRet)
	it.expectedModCount = it.vec.modCount
	it.cursor = it.lastRet
	it.lastRet = -1
}

func (it *vectorItr) HasPrevious() bool    { return it.cursor != 0 }
func (it *vectorItr) NextIndex() int32     { return it.cursor }
func (it *vectorItr) PreviousIndex() int32 { return it.cursor - 1 }

func (it *vectorItr) Previous() any {
	it.checkForComodification()
	i := it.cursor - 1
	if i < 0 {
		panic(NoSuchElementException_new())
	}
	it.cursor = i
	it.lastRet = i
	return it.vec.elementData(i)
}

func (it *vectorItr) Set_Object(o any) {
	if it.lastRet == -1 {
		panic(IllegalStateException_new())
	}
	it.checkForComodification()
	it.vec.Set_int_Object(it.lastRet, o)
}

func (it *vectorItr) Add_Object(o any) {
	i := it.cursor
	it.checkForComodification()
	it.vec.Add_int_Object(i, o)
	it.expectedModCount = it.vec.modCount
	it.cursor = i + 1
	it.lastRet = -1
}

// Stack is java.util.Stack.
type Stack struct {
	Vector
}

func (this *Stack) JClass_Stack() *Stack { return this }
func (this *Stack) JClassName() string   { return "java.util.Stack" }
func (this *Stack) JShallowCopy() any    { c := new(Stack); *c = *this; return c }

func Stack_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_Stack() *Stack })
	return ok
}

func Stack_fromAny(o any) *Stack {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_Stack() *Stack }](o).JClass_Stack()
}

func Stack_new() *Stack {
	this := &Stack{}
	this.jself = this
	this.Stack_init()
	return this
}

func (this *Stack) Stack_init() { this.Vector.Vector_init() }

func (this *Stack) Push_Object(item any) any {
	this.AddElement_Object(item)
	return item
}

func (this *Stack) Pop() any {
	n := this.Size()
	obj := this.Peek()
	this.RemoveElementAt_int(n - 1)
	return obj
}

func (this *Stack) Peek() any {
	n := this.Size()
	if n == 0 {
		panic(EmptyStackException_new())
	}
	return this.ElementAt_int(n - 1)
}

func (this *Stack) Empty() bool { return this.Size() == 0 }

func (this *Stack) Search_Object(o any) int32 {
	i := this.LastIndexOf_Object(o)
	if i >= 0 {
		return this.Size() - i
	}
	return -1
}
