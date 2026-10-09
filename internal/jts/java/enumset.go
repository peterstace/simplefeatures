package java

import "math/bits"

// EnumSet is java.util.EnumSet, implemented as Java's RegularEnumSet: a bit
// vector over the ordinals of an enum class with at most 64 constants.
// Enum constants are generated classes that embed Enum; the set holds each
// as the whole object, and identifies the enum class by its name.
type EnumSet struct {
	elementType string
	elements    uint64
	// universe holds the constants seen so far, indexed by ordinal.
	universe [64]any
}

func (this *EnumSet) JClass_EnumSet() *EnumSet { return this }
func (this *EnumSet) JClassName() string       { return "java.util.RegularEnumSet" }
func (this *EnumSet) JIface_Collection()       {}
func (this *EnumSet) JIface_Set()              {}

func EnumSet_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_EnumSet() *EnumSet })
	return ok
}

func EnumSet_fromAny(o any) *EnumSet {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_EnumSet() *EnumSet }](o).JClass_EnumSet()
}

// EnumSet_noneOf creates an empty set for the enum class c.
func EnumSet_noneOf(c *Class) *EnumSet {
	if c == nil {
		panic(NullPointerException_new())
	}
	return &EnumSet{elementType: c.name}
}

// EnumSet_copyOf_EnumSet is EnumSet.copyOf(EnumSet).
func EnumSet_copyOf_EnumSet(s *EnumSet) *EnumSet {
	return s.Clone().(*EnumSet)
}

// EnumSet_copyOf_Collection is EnumSet.copyOf(Collection).
func EnumSet_copyOf_Collection(c Collection) *EnumSet {
	if s, ok := c.(*EnumSet); ok {
		return s.Clone().(*EnumSet)
	}
	if c.IsEmpty() {
		panic(IllegalArgumentException_new_String(Lit("Collection is empty")))
	}
	it := c.Iterator()
	first := it.Next()
	if first == nil {
		panic(NullPointerException_new())
	}
	result := newEnumSetOf(Enum_fromAny(first))
	result.Add_Object(first)
	for it.HasNext() {
		result.Add_Object(it.Next())
	}
	return result
}

// newEnumSetOf is EnumSet.noneOf(e.getDeclaringClass()).
func newEnumSetOf(e *Enum) *EnumSet {
	if e == nil {
		panic(NullPointerException_new())
	}
	return &EnumSet{elementType: className(Ref(e))}
}

func EnumSet_of_Enum(e *Enum) *EnumSet {
	result := newEnumSetOf(e)
	result.add(e)
	return result
}

func EnumSet_of_Enum_Enum(e1, e2 *Enum) *EnumSet {
	result := newEnumSetOf(e1)
	result.add(e1)
	result.add(e2)
	return result
}

func EnumSet_of_Enum_Enum_Enum(e1, e2, e3 *Enum) *EnumSet {
	result := newEnumSetOf(e1)
	result.add(e1)
	result.add(e2)
	result.add(e3)
	return result
}

func EnumSet_of_Enum_Enum_Enum_Enum(e1, e2, e3, e4 *Enum) *EnumSet {
	result := newEnumSetOf(e1)
	result.add(e1)
	result.add(e2)
	result.add(e3)
	result.add(e4)
	return result
}

func EnumSet_of_Enum_Enum_Enum_Enum_Enum(e1, e2, e3, e4, e5 *Enum) *EnumSet {
	result := newEnumSetOf(e1)
	result.add(e1)
	result.add(e2)
	result.add(e3)
	result.add(e4)
	result.add(e5)
	return result
}

func EnumSet_of_Enum_EnumArr(first *Enum, rest []*Enum) *EnumSet {
	result := newEnumSetOf(first)
	result.add(first)
	for _, e := range rest {
		result.add(e)
	}
	return result
}

// add adds the enum constant e, given by its Enum part.
func (this *EnumSet) add(e *Enum) {
	this.Add_Object(Ref(e))
}

// typeCheck is EnumSet.typeCheck: it returns the Enum part of e, which must
// be a constant of the set's enum class.
func (this *EnumSet) typeCheck(e any) *Enum {
	if e == nil {
		panic(NullPointerException_new())
	}
	if className(e) != this.elementType || !Enum_isInstance(e) {
		panic(ClassCastException_new_String(FromGo("class " + className(e) + " != class " + this.elementType)))
	}
	return Enum_fromAny(e)
}

// member returns the bit for o if it is a constant of the set's enum class,
// or 0.
func (this *EnumSet) member(o any) uint64 {
	if o == nil || !Enum_isInstance(o) || className(o) != this.elementType {
		return 0
	}
	return 1 << uint(Enum_fromAny(o).Ordinal())
}

func (this *EnumSet) Add_Object(o any) bool {
	ord := this.typeCheck(o).Ordinal()
	if ord >= 64 {
		panic(Unsupported("EnumSet of an enum with more than 64 constants"))
	}
	old := this.elements
	this.elements |= 1 << uint(ord)
	this.universe[ord] = o
	return this.elements != old
}

func (this *EnumSet) Remove_Object(o any) bool {
	bit := this.member(o)
	if bit == 0 {
		return false
	}
	old := this.elements
	this.elements &^= bit
	return this.elements != old
}

func (this *EnumSet) Contains_Object(o any) bool { return this.elements&this.member(o) != 0 }

func (this *EnumSet) Size() int32   { return int32(bits.OnesCount64(this.elements)) }
func (this *EnumSet) IsEmpty() bool { return this.elements == 0 }
func (this *EnumSet) Clear()        { this.elements = 0 }

// Iterator returns RegularEnumSet's iterator, which iterates in ordinal order
// over the elements present when it was created.
func (this *EnumSet) Iterator() Iterator {
	return &enumSetItr{set: this, unseen: this.elements}
}

type enumSetItr struct {
	set          *EnumSet
	unseen       uint64
	lastReturned uint64
}

func (it *enumSetItr) HasNext() bool { return it.unseen != 0 }

func (it *enumSetItr) Next() any {
	if it.unseen == 0 {
		panic(NoSuchElementException_new())
	}
	it.lastReturned = it.unseen & -it.unseen
	it.unseen -= it.lastReturned
	return it.set.universe[bits.TrailingZeros64(it.lastReturned)]
}

func (it *enumSetItr) Remove() {
	if it.lastReturned == 0 {
		panic(IllegalStateException_new())
	}
	it.set.elements &^= it.lastReturned
	it.lastReturned = 0
}

func (this *EnumSet) ContainsAll_Collection(c Collection) bool {
	es, ok := c.(*EnumSet)
	if !ok {
		return acContainsAll(this, c)
	}
	if es.elementType != this.elementType {
		return es.IsEmpty()
	}
	return es.elements&^this.elements == 0
}

func (this *EnumSet) AddAll_Collection(c Collection) bool {
	es, ok := c.(*EnumSet)
	if !ok {
		return acAddAll(this, c)
	}
	if es.elementType != this.elementType {
		if es.IsEmpty() {
			return false
		}
		panic(ClassCastException_new_String(FromGo(es.elementType + " != " + this.elementType)))
	}
	old := this.elements
	this.elements |= es.elements
	for i := 0; i < 64; i++ {
		if es.elements&(1<<uint(i)) != 0 {
			this.universe[i] = es.universe[i]
		}
	}
	return this.elements != old
}

func (this *EnumSet) RemoveAll_Collection(c Collection) bool {
	es, ok := c.(*EnumSet)
	if !ok {
		return asRemoveAll(this, c)
	}
	if es.elementType != this.elementType {
		return false
	}
	old := this.elements
	this.elements &^= es.elements
	return this.elements != old
}

func (this *EnumSet) RetainAll_Collection(c Collection) bool {
	es, ok := c.(*EnumSet)
	if !ok {
		return acRetainAll(this, c)
	}
	if es.elementType != this.elementType {
		changed := this.elements != 0
		this.elements = 0
		return changed
	}
	old := this.elements
	this.elements &= es.elements
	return this.elements != old
}

func (this *EnumSet) Equals_Object(o any) bool {
	es, ok := o.(*EnumSet)
	if !ok {
		return asEquals(this, this, o)
	}
	if es.elementType != this.elementType {
		return this.elements == 0 && es.elements == 0
	}
	return es.elements == this.elements
}

func (this *EnumSet) HashCode() int32 { return asHashCode(this) }

func (this *EnumSet) ToArray() []any              { return acToArray(this) }
func (this *EnumSet) ToArray_ObjectArr(a any) any { return acToArrayT(this, a) }
func (this *EnumSet) ToString() *String           { return acToString(this, this) }

// Clone implements EnumSet.clone.
func (this *EnumSet) Clone() any {
	c := *this
	return &c
}
