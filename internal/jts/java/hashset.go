package java

// HashSet is java.util.HashSet: the keys of a HashMap, so that its iteration
// order is the map's.
//
// A LinkedHashSet embeds a HashSet whose map is a LinkedHashMap.
type HashSet struct {
	jself any
	m     *HashMap
}

func (this *HashSet) JSelf() any        { return this.jself }
func (this *HashSet) JSetSelf(self any) { this.jself = self }

func (this *HashSet) JClass_HashSet() *HashSet { return this }
func (this *HashSet) JClassName() string       { return "java.util.HashSet" }
func (this *HashSet) JIface_Collection()       {}
func (this *HashSet) JIface_Set()              {}

func HashSet_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_HashSet() *HashSet })
	return ok
}

func HashSet_fromAny(o any) *HashSet {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_HashSet() *HashSet }](o).JClass_HashSet()
}

// self returns the whole object, which is a LinkedHashSet when this is
// embedded in one.
func (this *HashSet) self() any {
	if this.jself != nil {
		return this.jself
	}
	return this
}

func HashSet_new() *HashSet {
	return &HashSet{m: HashMap_new()}
}

func HashSet_new_int(initialCapacity int32) *HashSet {
	return &HashSet{m: HashMap_new_int(initialCapacity)}
}

func HashSet_new_int_float(initialCapacity int32, loadFactor float32) *HashSet {
	return &HashSet{m: HashMap_new_int_float(initialCapacity, loadFactor)}
}

// HashSet_new_Collection sizes the map for c (and at least 12 elements),
// then adds c's elements in its iteration order.
func HashSet_new_Collection(c Collection) *HashSet {
	this := &HashSet{m: HashMap_new_int(hashMapCapacityFor(hashSetMinSize(c)))}
	this.AddAll_Collection(c)
	return this
}

func hashSetMinSize(c Collection) int32 {
	if n := c.Size(); n > 12 {
		return n
	}
	return 12
}

func (this *HashSet) Iterator() Iterator { return this.m.newIterator(iterKeys) }
func (this *HashSet) Size() int32        { return this.m.size }
func (this *HashSet) IsEmpty() bool      { return this.m.size == 0 }
func (this *HashSet) Clear()             { this.m.Clear() }

func (this *HashSet) Contains_Object(o any) bool {
	return this.m.ContainsKey_Object(o)
}

// Add_Object maps e to a non-null placeholder value, standing in for Java's
// HashSet.PRESENT.
func (this *HashSet) Add_Object(e any) bool {
	return this.m.Put_Object_Object(e, true) == nil
}

func (this *HashSet) Remove_Object(o any) bool {
	return this.m.Remove_Object(o) != nil
}

func (this *HashSet) ToArray() []any { return this.m.keysArray() }

func (this *HashSet) ToArray_ObjectArr(a any) any {
	return fillArray(a, this.m.keysArray())
}

func (this *HashSet) AddAll_Collection(c Collection) bool      { return acAddAll(this, c) }
func (this *HashSet) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }
func (this *HashSet) RemoveAll_Collection(c Collection) bool   { return asRemoveAll(this, c) }
func (this *HashSet) RetainAll_Collection(c Collection) bool   { return acRetainAll(this, c) }
func (this *HashSet) Equals_Object(o any) bool                 { return asEquals(this, this.self(), o) }
func (this *HashSet) HashCode() int32                          { return asHashCode(this) }
func (this *HashSet) ToString() *String                        { return acToString(this, this.self()) }

// Clone is HashSet.clone: a set of the same class with a clone of the map.
func (this *HashSet) Clone() any {
	m := HashMap_fromAny(this.m.Clone())
	if this.m.linked {
		s := &LinkedHashSet{}
		s.jself = s
		s.m = m
		return s
	}
	return &HashSet{m: m}
}

// LinkedHashSet is java.util.LinkedHashSet, which iterates in insertion
// order.
type LinkedHashSet struct {
	HashSet
}

func (this *LinkedHashSet) JClass_LinkedHashSet() *LinkedHashSet { return this }
func (this *LinkedHashSet) JClassName() string                   { return "java.util.LinkedHashSet" }

func LinkedHashSet_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_LinkedHashSet() *LinkedHashSet })
	return ok
}

func LinkedHashSet_fromAny(o any) *LinkedHashSet {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_LinkedHashSet() *LinkedHashSet }](o).JClass_LinkedHashSet()
}

func LinkedHashSet_new() *LinkedHashSet {
	return LinkedHashSet_new_int_float(16, hashMapDefaultLoadFactor)
}

func LinkedHashSet_new_int(initialCapacity int32) *LinkedHashSet {
	return LinkedHashSet_new_int_float(initialCapacity, hashMapDefaultLoadFactor)
}

func LinkedHashSet_new_int_float(initialCapacity int32, loadFactor float32) *LinkedHashSet {
	this := &LinkedHashSet{}
	this.jself = this
	this.m = &LinkedHashMap_new_int_float(initialCapacity, loadFactor).HashMap
	return this
}

// LinkedHashSet_new_Collection sizes the map for c (and at least 12
// elements), then adds c's elements in its iteration order.
func LinkedHashSet_new_Collection(c Collection) *LinkedHashSet {
	this := LinkedHashSet_new_int(hashMapCapacityFor(hashSetMinSize(c)))
	this.AddAll_Collection(c)
	return this
}
