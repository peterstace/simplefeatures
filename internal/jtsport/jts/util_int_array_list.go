package jts

// Util_IntArrayList is an extendable array of primitive int values.
type Util_IntArrayList struct {
	data []int
	size int
}

// Util_NewIntArrayList constructs an empty list.
func Util_NewIntArrayList() *Util_IntArrayList {
	return Util_NewIntArrayListWithCapacity(10)
}

// Util_NewIntArrayListWithCapacity constructs an empty list with the specified
// initial capacity.
func Util_NewIntArrayListWithCapacity(initialCapacity int) *Util_IntArrayList {
	return &Util_IntArrayList{
		data: make([]int, initialCapacity),
		size: 0,
	}
}

// Size returns the number of values in this list.
func (l *Util_IntArrayList) Size() int {
	return l.size
}

// EnsureCapacity increases the capacity of this list instance, if necessary,
// to ensure that it can hold at least the number of elements specified by the
// capacity argument.
func (l *Util_IntArrayList) EnsureCapacity(capacity int) {
	if capacity <= len(l.data) {
		return
	}
	newLength := capacity
	if len(l.data)*2 > newLength {
		newLength = len(l.data) * 2
	}
	newData := make([]int, newLength)
	copy(newData, l.data[:l.size])
	l.data = newData
}

// Add adds a value to the end of this list.
func (l *Util_IntArrayList) Add(value int) {
	l.EnsureCapacity(l.size + 1)
	l.data[l.size] = value
	l.size++
}

// AddAll adds all values in an array to the end of this list.
func (l *Util_IntArrayList) AddAll(values []int) {
	if values == nil {
		return
	}
	if len(values) == 0 {
		return
	}
	l.EnsureCapacity(l.size + len(values))
	copy(l.data[l.size:], values)
	l.size += len(values)
}

// ToArray returns an int array containing a copy of the values in this list.
func (l *Util_IntArrayList) ToArray() []int {
	array := make([]int, l.size)
	copy(array, l.data[:l.size])
	return array
}
