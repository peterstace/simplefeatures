package java

import "math/bits"

// TreeMap is java.util.TreeMap, a red-black tree transliterated from JDK 21.
// The tree's shape decides which comparisons are made and which entry
// objects survive a removal, so it follows Java's exactly, including the
// balanced trees that buildFromSorted makes when copying sorted maps.
type TreeMap struct {
	comparator      Comparator
	root            *treeMapEntry
	size            int32
	modCount        int32
	entrySet        *treeMapEntrySet
	navigableKeySet *treeMapKeySet
	values          *treeMapValues
}

const (
	treeRed   = false
	treeBlack = true
)

// treeMapEntry is TreeMap.Entry, a node of the tree. Iterating a TreeMap's
// entry set returns these nodes, so setValue writes through.
type treeMapEntry struct {
	key    any
	value  any
	left   *treeMapEntry
	right  *treeMapEntry
	parent *treeMapEntry
	color  bool
}

func (this *TreeMap) JClass_TreeMap() *TreeMap { return this }
func (this *TreeMap) JClassName() string       { return "java.util.TreeMap" }
func (this *TreeMap) JIface_Map()              {}
func (this *TreeMap) JIface_SortedMap()        {}
func (this *TreeMap) JIface_NavigableMap()     {}

func TreeMap_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_TreeMap() *TreeMap })
	return ok
}

func TreeMap_fromAny(o any) *TreeMap {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_TreeMap() *TreeMap }](o).JClass_TreeMap()
}

func TreeMap_new() *TreeMap { return &TreeMap{} }

func TreeMap_new_Comparator(c Comparator) *TreeMap { return &TreeMap{comparator: c} }

func TreeMap_new_Map(m Map) *TreeMap {
	this := &TreeMap{}
	this.PutAll_Map(m)
	return this
}

func TreeMap_new_SortedMap(m SortedMap) *TreeMap {
	this := &TreeMap{comparator: m.Comparator()}
	this.buildFromSorted(m.Size(), m.EntrySet().Iterator(), nil)
	return this
}

func (this *TreeMap) Size() int32   { return this.size }
func (this *TreeMap) IsEmpty() bool { return this.size == 0 }

func (this *TreeMap) ContainsKey_Object(key any) bool { return this.getEntry(key) != nil }

func (this *TreeMap) ContainsValue_Object(value any) bool {
	for e := this.getFirstEntry(); e != nil; e = treeSuccessor(e) {
		if objectsEqual(value, e.value) {
			return true
		}
	}
	return false
}

func (this *TreeMap) Get_Object(key any) any {
	p := this.getEntry(key)
	if p == nil {
		return nil
	}
	return p.value
}

func (this *TreeMap) Comparator() Comparator { return this.comparator }

func (this *TreeMap) FirstKey() any { return treeKey(this.getFirstEntry()) }
func (this *TreeMap) LastKey() any  { return treeKey(this.getLastEntry()) }

func (this *TreeMap) PutAll_Map(m Map) {
	mapSize := m.Size()
	if this.size == 0 && mapSize != 0 {
		if sm, ok := m.(SortedMap); ok && javaObjectsEquals(this.comparator, sm.Comparator()) {
			this.modCount++
			this.buildFromSorted(mapSize, m.EntrySet().Iterator(), nil)
			return
		}
	}
	amPutAll(this, m)
}

// getEntry returns the entry for key, or nil.
func (this *TreeMap) getEntry(key any) *treeMapEntry {
	if this.comparator != nil {
		return this.getEntryUsingComparator(key)
	}
	requireComparable(key)
	p := this.root
	for p != nil {
		cmp := Comparable_CompareTo_Object(key, p.key)
		switch {
		case cmp < 0:
			p = p.left
		case cmp > 0:
			p = p.right
		default:
			return p
		}
	}
	return nil
}

func (this *TreeMap) getEntryUsingComparator(key any) *treeMapEntry {
	cpr := this.comparator
	p := this.root
	for p != nil {
		cmp := cpr.Compare_Object_Object(key, p.key)
		switch {
		case cmp < 0:
			p = p.left
		case cmp > 0:
			p = p.right
		default:
			return p
		}
	}
	return nil
}

func (this *TreeMap) getCeilingEntry(key any) *treeMapEntry {
	p := this.root
	for p != nil {
		cmp := this.compare(key, p.key)
		switch {
		case cmp < 0:
			if p.left == nil {
				return p
			}
			p = p.left
		case cmp > 0:
			if p.right == nil {
				parent := p.parent
				ch := p
				for parent != nil && ch == parent.right {
					ch = parent
					parent = parent.parent
				}
				return parent
			}
			p = p.right
		default:
			return p
		}
	}
	return nil
}

func (this *TreeMap) getFloorEntry(key any) *treeMapEntry {
	p := this.root
	for p != nil {
		cmp := this.compare(key, p.key)
		switch {
		case cmp > 0:
			if p.right == nil {
				return p
			}
			p = p.right
		case cmp < 0:
			if p.left == nil {
				parent := p.parent
				ch := p
				for parent != nil && ch == parent.left {
					ch = parent
					parent = parent.parent
				}
				return parent
			}
			p = p.left
		default:
			return p
		}
	}
	return nil
}

func (this *TreeMap) getHigherEntry(key any) *treeMapEntry {
	p := this.root
	for p != nil {
		cmp := this.compare(key, p.key)
		if cmp < 0 {
			if p.left == nil {
				return p
			}
			p = p.left
		} else {
			if p.right == nil {
				parent := p.parent
				ch := p
				for parent != nil && ch == parent.right {
					ch = parent
					parent = parent.parent
				}
				return parent
			}
			p = p.right
		}
	}
	return nil
}

func (this *TreeMap) getLowerEntry(key any) *treeMapEntry {
	p := this.root
	for p != nil {
		cmp := this.compare(key, p.key)
		if cmp > 0 {
			if p.right == nil {
				return p
			}
			p = p.right
		} else {
			if p.left == nil {
				parent := p.parent
				ch := p
				for parent != nil && ch == parent.left {
					ch = parent
					parent = parent.parent
				}
				return parent
			}
			p = p.left
		}
	}
	return nil
}

func (this *TreeMap) Put_Object_Object(key, value any) any {
	return this.put(key, value, true)
}

func (this *TreeMap) PutIfAbsent_Object_Object(key, value any) any {
	return this.put(key, value, false)
}

func (this *TreeMap) put(key, value any, replaceOld bool) any {
	t := this.root
	if t == nil {
		this.addEntryToEmptyMap(key, value)
		return nil
	}
	var cmp int32
	var parent *treeMapEntry
	cpr := this.comparator
	if cpr == nil {
		requireComparable(key)
	}
	for t != nil {
		parent = t
		if cpr != nil {
			cmp = cpr.Compare_Object_Object(key, t.key)
		} else {
			cmp = Comparable_CompareTo_Object(key, t.key)
		}
		switch {
		case cmp < 0:
			t = t.left
		case cmp > 0:
			t = t.right
		default:
			oldValue := t.value
			if replaceOld || oldValue == nil {
				t.value = value
			}
			return oldValue
		}
	}
	this.addEntry(key, value, parent, cmp < 0)
	return nil
}

func (this *TreeMap) addEntry(key, value any, parent *treeMapEntry, addToLeft bool) {
	e := &treeMapEntry{key: key, value: value, parent: parent, color: treeBlack}
	if addToLeft {
		parent.left = e
	} else {
		parent.right = e
	}
	this.fixAfterInsertion(e)
	this.size++
	this.modCount++
}

func (this *TreeMap) addEntryToEmptyMap(key, value any) {
	this.compare(key, key)
	this.root = &treeMapEntry{key: key, value: value, color: treeBlack}
	this.size = 1
	this.modCount++
}

func (this *TreeMap) Remove_Object(key any) any {
	p := this.getEntry(key)
	if p == nil {
		return nil
	}
	oldValue := p.value
	this.deleteEntry(p)
	return oldValue
}

// Remove_Object_Object is Map.remove(Object, Object).
func (this *TreeMap) Remove_Object_Object(key, value any) bool {
	curValue := this.Get_Object(key)
	if !javaObjectsEquals(curValue, value) || curValue == nil && !this.ContainsKey_Object(key) {
		return false
	}
	this.Remove_Object(key)
	return true
}

func (this *TreeMap) Replace_Object_Object_Object(key, oldValue, newValue any) bool {
	p := this.getEntry(key)
	if p != nil && javaObjectsEquals(oldValue, p.value) {
		p.value = newValue
		return true
	}
	return false
}

func (this *TreeMap) Replace_Object_Object(key, value any) any {
	p := this.getEntry(key)
	if p != nil {
		oldValue := p.value
		p.value = value
		return oldValue
	}
	return nil
}

func (this *TreeMap) GetOrDefault_Object_Object(key, defaultValue any) any {
	return mapGetOrDefault(this, key, defaultValue)
}

func (this *TreeMap) Clear() {
	this.modCount++
	this.size = 0
	this.root = nil
}

// Clone implements TreeMap.clone, whose copy is a balanced tree rather than
// one of the original's shape.
func (this *TreeMap) Clone() any {
	c := &TreeMap{comparator: this.comparator}
	c.buildFromSorted(this.size, this.EntrySet().Iterator(), nil)
	return c
}

func (this *TreeMap) FirstEntry() Map_Entry { return treeExportEntry(this.getFirstEntry()) }
func (this *TreeMap) LastEntry() Map_Entry  { return treeExportEntry(this.getLastEntry()) }

func (this *TreeMap) PollFirstEntry() Map_Entry {
	p := this.getFirstEntry()
	result := treeExportEntry(p)
	if p != nil {
		this.deleteEntry(p)
	}
	return result
}

func (this *TreeMap) PollLastEntry() Map_Entry {
	p := this.getLastEntry()
	result := treeExportEntry(p)
	if p != nil {
		this.deleteEntry(p)
	}
	return result
}

func (this *TreeMap) LowerEntry_Object(key any) Map_Entry {
	return treeExportEntry(this.getLowerEntry(key))
}

func (this *TreeMap) LowerKey_Object(key any) any { return treeKeyOrNull(this.getLowerEntry(key)) }

func (this *TreeMap) FloorEntry_Object(key any) Map_Entry {
	return treeExportEntry(this.getFloorEntry(key))
}

func (this *TreeMap) FloorKey_Object(key any) any { return treeKeyOrNull(this.getFloorEntry(key)) }

func (this *TreeMap) CeilingEntry_Object(key any) Map_Entry {
	return treeExportEntry(this.getCeilingEntry(key))
}

func (this *TreeMap) CeilingKey_Object(key any) any {
	return treeKeyOrNull(this.getCeilingEntry(key))
}

func (this *TreeMap) HigherEntry_Object(key any) Map_Entry {
	return treeExportEntry(this.getHigherEntry(key))
}

func (this *TreeMap) HigherKey_Object(key any) any { return treeKeyOrNull(this.getHigherEntry(key)) }

func (this *TreeMap) KeySet() Set { return this.NavigableKeySet() }

func (this *TreeMap) NavigableKeySet() NavigableSet {
	if this.navigableKeySet == nil {
		this.navigableKeySet = &treeMapKeySet{m: this}
	}
	return this.navigableKeySet
}

func (this *TreeMap) Values() Collection {
	if this.values == nil {
		this.values = &treeMapValues{m: this}
	}
	return this.values
}

func (this *TreeMap) EntrySet() Set {
	if this.entrySet == nil {
		this.entrySet = &treeMapEntrySet{m: this}
	}
	return this.entrySet
}

func (this *TreeMap) Equals_Object(o any) bool { return amEquals(this, this, o) }
func (this *TreeMap) HashCode() int32          { return amHashCode(this) }
func (this *TreeMap) ToString() *String        { return amToString(this, this) }

func (this *TreeMap) keyIterator() Iterator {
	return &treeMapItr{m: this, next: this.getFirstEntry(), expectedModCount: this.modCount, kind: treeItrKeys}
}

func (this *TreeMap) descendingKeyIterator() Iterator {
	return &treeMapItr{m: this, next: this.getLastEntry(), expectedModCount: this.modCount, kind: treeItrKeys, descending: true}
}

// compare is TreeMap.compare.
func (this *TreeMap) compare(k1, k2 any) int32 {
	return compareWith(this.comparator, k1, k2)
}

func treeExportEntry(e *treeMapEntry) Map_Entry {
	if e == nil {
		return nil
	}
	return &treeImmutableEntry{key: e.key, value: e.value}
}

func treeKeyOrNull(e *treeMapEntry) any {
	if e == nil {
		return nil
	}
	return e.key
}

func treeKey(e *treeMapEntry) any {
	if e == nil {
		panic(NoSuchElementException_new())
	}
	return e.key
}

func (this *TreeMap) getFirstEntry() *treeMapEntry {
	p := this.root
	if p != nil {
		for p.left != nil {
			p = p.left
		}
	}
	return p
}

func (this *TreeMap) getLastEntry() *treeMapEntry {
	p := this.root
	if p != nil {
		for p.right != nil {
			p = p.right
		}
	}
	return p
}

func treeSuccessor(t *treeMapEntry) *treeMapEntry {
	if t == nil {
		return nil
	}
	if t.right != nil {
		p := t.right
		for p.left != nil {
			p = p.left
		}
		return p
	}
	p := t.parent
	ch := t
	for p != nil && ch == p.right {
		ch = p
		p = p.parent
	}
	return p
}

func treePredecessor(t *treeMapEntry) *treeMapEntry {
	if t == nil {
		return nil
	}
	if t.left != nil {
		p := t.left
		for p.right != nil {
			p = p.right
		}
		return p
	}
	p := t.parent
	ch := t
	for p != nil && ch == p.left {
		ch = p
		p = p.parent
	}
	return p
}

func treeColorOf(p *treeMapEntry) bool {
	if p == nil {
		return treeBlack
	}
	return p.color
}

func treeParentOf(p *treeMapEntry) *treeMapEntry {
	if p == nil {
		return nil
	}
	return p.parent
}

func treeSetColor(p *treeMapEntry, c bool) {
	if p != nil {
		p.color = c
	}
}

func treeLeftOf(p *treeMapEntry) *treeMapEntry {
	if p == nil {
		return nil
	}
	return p.left
}

func treeRightOf(p *treeMapEntry) *treeMapEntry {
	if p == nil {
		return nil
	}
	return p.right
}

func (this *TreeMap) rotateLeft(p *treeMapEntry) {
	if p != nil {
		r := p.right
		p.right = r.left
		if r.left != nil {
			r.left.parent = p
		}
		r.parent = p.parent
		switch {
		case p.parent == nil:
			this.root = r
		case p.parent.left == p:
			p.parent.left = r
		default:
			p.parent.right = r
		}
		r.left = p
		p.parent = r
	}
}

func (this *TreeMap) rotateRight(p *treeMapEntry) {
	if p != nil {
		l := p.left
		p.left = l.right
		if l.right != nil {
			l.right.parent = p
		}
		l.parent = p.parent
		switch {
		case p.parent == nil:
			this.root = l
		case p.parent.right == p:
			p.parent.right = l
		default:
			p.parent.left = l
		}
		l.right = p
		p.parent = l
	}
}

func (this *TreeMap) fixAfterInsertion(x *treeMapEntry) {
	x.color = treeRed
	for x != nil && x != this.root && x.parent.color == treeRed {
		if treeParentOf(x) == treeLeftOf(treeParentOf(treeParentOf(x))) {
			y := treeRightOf(treeParentOf(treeParentOf(x)))
			if treeColorOf(y) == treeRed {
				treeSetColor(treeParentOf(x), treeBlack)
				treeSetColor(y, treeBlack)
				treeSetColor(treeParentOf(treeParentOf(x)), treeRed)
				x = treeParentOf(treeParentOf(x))
			} else {
				if x == treeRightOf(treeParentOf(x)) {
					x = treeParentOf(x)
					this.rotateLeft(x)
				}
				treeSetColor(treeParentOf(x), treeBlack)
				treeSetColor(treeParentOf(treeParentOf(x)), treeRed)
				this.rotateRight(treeParentOf(treeParentOf(x)))
			}
		} else {
			y := treeLeftOf(treeParentOf(treeParentOf(x)))
			if treeColorOf(y) == treeRed {
				treeSetColor(treeParentOf(x), treeBlack)
				treeSetColor(y, treeBlack)
				treeSetColor(treeParentOf(treeParentOf(x)), treeRed)
				x = treeParentOf(treeParentOf(x))
			} else {
				if x == treeLeftOf(treeParentOf(x)) {
					x = treeParentOf(x)
					this.rotateRight(x)
				}
				treeSetColor(treeParentOf(x), treeBlack)
				treeSetColor(treeParentOf(treeParentOf(x)), treeRed)
				this.rotateLeft(treeParentOf(treeParentOf(x)))
			}
		}
	}
	this.root.color = treeBlack
}

// deleteEntry removes p. An entry with two children is removed by moving
// its successor's key and value into it and unlinking the successor.
func (this *TreeMap) deleteEntry(p *treeMapEntry) {
	this.modCount++
	this.size--

	if p.left != nil && p.right != nil {
		s := treeSuccessor(p)
		p.key = s.key
		p.value = s.value
		p = s
	}

	replacement := p.left
	if replacement == nil {
		replacement = p.right
	}

	switch {
	case replacement != nil:
		replacement.parent = p.parent
		switch {
		case p.parent == nil:
			this.root = replacement
		case p == p.parent.left:
			p.parent.left = replacement
		default:
			p.parent.right = replacement
		}
		p.left = nil
		p.right = nil
		p.parent = nil
		if p.color == treeBlack {
			this.fixAfterDeletion(replacement)
		}
	case p.parent == nil:
		this.root = nil
	default:
		if p.color == treeBlack {
			this.fixAfterDeletion(p)
		}
		if p.parent != nil {
			if p == p.parent.left {
				p.parent.left = nil
			} else if p == p.parent.right {
				p.parent.right = nil
			}
			p.parent = nil
		}
	}
}

func (this *TreeMap) fixAfterDeletion(x *treeMapEntry) {
	for x != this.root && treeColorOf(x) == treeBlack {
		if x == treeLeftOf(treeParentOf(x)) {
			sib := treeRightOf(treeParentOf(x))
			if treeColorOf(sib) == treeRed {
				treeSetColor(sib, treeBlack)
				treeSetColor(treeParentOf(x), treeRed)
				this.rotateLeft(treeParentOf(x))
				sib = treeRightOf(treeParentOf(x))
			}
			if treeColorOf(treeLeftOf(sib)) == treeBlack && treeColorOf(treeRightOf(sib)) == treeBlack {
				treeSetColor(sib, treeRed)
				x = treeParentOf(x)
			} else {
				if treeColorOf(treeRightOf(sib)) == treeBlack {
					treeSetColor(treeLeftOf(sib), treeBlack)
					treeSetColor(sib, treeRed)
					this.rotateRight(sib)
					sib = treeRightOf(treeParentOf(x))
				}
				treeSetColor(sib, treeColorOf(treeParentOf(x)))
				treeSetColor(treeParentOf(x), treeBlack)
				treeSetColor(treeRightOf(sib), treeBlack)
				this.rotateLeft(treeParentOf(x))
				x = this.root
			}
		} else {
			sib := treeLeftOf(treeParentOf(x))
			if treeColorOf(sib) == treeRed {
				treeSetColor(sib, treeBlack)
				treeSetColor(treeParentOf(x), treeRed)
				this.rotateRight(treeParentOf(x))
				sib = treeLeftOf(treeParentOf(x))
			}
			if treeColorOf(treeRightOf(sib)) == treeBlack && treeColorOf(treeLeftOf(sib)) == treeBlack {
				treeSetColor(sib, treeRed)
				x = treeParentOf(x)
			} else {
				if treeColorOf(treeLeftOf(sib)) == treeBlack {
					treeSetColor(treeRightOf(sib), treeBlack)
					treeSetColor(sib, treeRed)
					this.rotateLeft(sib)
					sib = treeLeftOf(treeParentOf(x))
				}
				treeSetColor(sib, treeColorOf(treeParentOf(x)))
				treeSetColor(treeParentOf(x), treeBlack)
				treeSetColor(treeLeftOf(sib), treeBlack)
				this.rotateRight(treeParentOf(x))
				x = this.root
			}
		}
	}
	treeSetColor(x, treeBlack)
}

// buildFromSorted replaces the tree with a balanced one holding the size
// entries of it, in order. If keysOnly is non-nil the iterator yields keys,
// each mapped to *keysOnly; otherwise it yields Map.Entry values.
func (this *TreeMap) buildFromSorted(size int32, it Iterator, keysOnly *any) {
	this.size = size
	this.root = this.buildFromSortedRange(0, 0, size-1, treeComputeRedLevel(size), it, keysOnly)
}

func (this *TreeMap) buildFromSortedRange(level, lo, hi, redLevel int32, it Iterator, keysOnly *any) *treeMapEntry {
	if hi < lo {
		return nil
	}
	mid := int32(uint32(lo+hi) >> 1)
	var left *treeMapEntry
	if lo < mid {
		left = this.buildFromSortedRange(level+1, lo, mid-1, redLevel, it, keysOnly)
	}
	var key, value any
	if keysOnly == nil {
		entry := it.Next().(Map_Entry)
		key = entry.GetKey()
		value = entry.GetValue()
	} else {
		key = it.Next()
		value = *keysOnly
	}
	middle := &treeMapEntry{key: key, value: value, color: treeBlack}
	if level == redLevel {
		middle.color = treeRed
	}
	if left != nil {
		middle.left = left
		left.parent = middle
	}
	if mid < hi {
		right := this.buildFromSortedRange(level+1, mid+1, hi, redLevel, it, keysOnly)
		middle.right = right
		right.parent = middle
	}
	return middle
}

// treeComputeRedLevel is TreeMap.computeRedLevel.
func treeComputeRedLevel(size int32) int32 {
	return 31 - int32(bits.LeadingZeros32(uint32(size+1)))
}

func (e *treeMapEntry) JClassName() string { return "java.util.TreeMap$Entry" }
func (e *treeMapEntry) JIface_Map_Entry()  {}
func (e *treeMapEntry) GetKey() any        { return e.key }
func (e *treeMapEntry) GetValue() any      { return e.value }

func (e *treeMapEntry) SetValue_Object(value any) any {
	old := e.value
	e.value = value
	return old
}

func (e *treeMapEntry) Equals_Object(o any) bool { return entryEquals(e.key, e.value, o) }
func (e *treeMapEntry) HashCode() int32          { return entryHashCode(e.key, e.value) }
func (e *treeMapEntry) ToString() *String        { return entryToString(e.key, e.value) }

// treeImmutableEntry is AbstractMap.SimpleImmutableEntry, the snapshot of an
// entry that TreeMap's navigation methods return.
type treeImmutableEntry struct {
	key   any
	value any
}

func (e *treeImmutableEntry) JClassName() string {
	return "java.util.AbstractMap$SimpleImmutableEntry"
}
func (e *treeImmutableEntry) JIface_Map_Entry() {}
func (e *treeImmutableEntry) GetKey() any       { return e.key }
func (e *treeImmutableEntry) GetValue() any     { return e.value }

func (e *treeImmutableEntry) SetValue_Object(any) any { panic(unsupported()) }

func (e *treeImmutableEntry) Equals_Object(o any) bool { return entryEquals(e.key, e.value, o) }
func (e *treeImmutableEntry) HashCode() int32          { return entryHashCode(e.key, e.value) }
func (e *treeImmutableEntry) ToString() *String        { return entryToString(e.key, e.value) }

type treeItrKind int

const (
	treeItrEntries treeItrKind = iota
	treeItrKeys
	treeItrValues
)

// treeMapItr is TreeMap.PrivateEntryIterator and its subclasses.
type treeMapItr struct {
	m                *TreeMap
	next             *treeMapEntry
	lastReturned     *treeMapEntry
	expectedModCount int32
	kind             treeItrKind
	descending       bool
}

func (it *treeMapItr) HasNext() bool { return it.next != nil }

func (it *treeMapItr) Next() any {
	e := it.next
	if e == nil {
		panic(NoSuchElementException_new())
	}
	if it.m.modCount != it.expectedModCount {
		panic(ConcurrentModificationException_new())
	}
	if it.descending {
		it.next = treePredecessor(e)
	} else {
		it.next = treeSuccessor(e)
	}
	it.lastReturned = e
	switch it.kind {
	case treeItrKeys:
		return e.key
	case treeItrValues:
		return e.value
	}
	return e
}

func (it *treeMapItr) Remove() {
	if it.lastReturned == nil {
		panic(IllegalStateException_new())
	}
	if it.m.modCount != it.expectedModCount {
		panic(ConcurrentModificationException_new())
	}
	if !it.descending && it.lastReturned.left != nil && it.lastReturned.right != nil {
		it.next = it.lastReturned
	}
	it.m.deleteEntry(it.lastReturned)
	it.expectedModCount = it.m.modCount
	it.lastReturned = nil
}

// treeMapKeySet is TreeMap.KeySet, the navigable key set view.
type treeMapKeySet struct {
	m *TreeMap
}

func (this *treeMapKeySet) JClassName() string   { return "java.util.TreeMap$KeySet" }
func (this *treeMapKeySet) JIface_Collection()   {}
func (this *treeMapKeySet) JIface_Set()          {}
func (this *treeMapKeySet) JIface_SortedSet()    {}
func (this *treeMapKeySet) JIface_NavigableSet() {}

func (this *treeMapKeySet) Iterator() Iterator           { return this.m.keyIterator() }
func (this *treeMapKeySet) DescendingIterator() Iterator { return this.m.descendingKeyIterator() }
func (this *treeMapKeySet) Size() int32                  { return this.m.Size() }
func (this *treeMapKeySet) IsEmpty() bool                { return this.m.IsEmpty() }
func (this *treeMapKeySet) Contains_Object(o any) bool   { return this.m.ContainsKey_Object(o) }
func (this *treeMapKeySet) Clear()                       { this.m.Clear() }
func (this *treeMapKeySet) Lower_Object(e any) any       { return this.m.LowerKey_Object(e) }
func (this *treeMapKeySet) Floor_Object(e any) any       { return this.m.FloorKey_Object(e) }
func (this *treeMapKeySet) Ceiling_Object(e any) any     { return this.m.CeilingKey_Object(e) }
func (this *treeMapKeySet) Higher_Object(e any) any      { return this.m.HigherKey_Object(e) }
func (this *treeMapKeySet) First() any                   { return this.m.FirstKey() }
func (this *treeMapKeySet) Last() any                    { return this.m.LastKey() }
func (this *treeMapKeySet) Comparator() Comparator       { return this.m.Comparator() }

func (this *treeMapKeySet) PollFirst() any {
	e := this.m.PollFirstEntry()
	if e == nil {
		return nil
	}
	return e.GetKey()
}

func (this *treeMapKeySet) PollLast() any {
	e := this.m.PollLastEntry()
	if e == nil {
		return nil
	}
	return e.GetKey()
}

func (this *treeMapKeySet) Remove_Object(o any) bool {
	oldSize := this.Size()
	this.m.Remove_Object(o)
	return this.Size() != oldSize
}

func (this *treeMapKeySet) Add_Object(any) bool                      { panic(unsupported()) }
func (this *treeMapKeySet) AddAll_Collection(c Collection) bool      { return acAddAll(this, c) }
func (this *treeMapKeySet) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }
func (this *treeMapKeySet) RemoveAll_Collection(c Collection) bool   { return asRemoveAll(this, c) }
func (this *treeMapKeySet) RetainAll_Collection(c Collection) bool   { return acRetainAll(this, c) }
func (this *treeMapKeySet) ToArray() []any                           { return acToArray(this) }
func (this *treeMapKeySet) ToArray_ObjectArr(a any) any              { return acToArrayT(this, a) }
func (this *treeMapKeySet) Equals_Object(o any) bool                 { return asEquals(this, this, o) }
func (this *treeMapKeySet) HashCode() int32                          { return asHashCode(this) }
func (this *treeMapKeySet) ToString() *String                        { return acToString(this, this) }

// treeMapValues is TreeMap.Values, the values view.
type treeMapValues struct {
	m *TreeMap
}

func (this *treeMapValues) JClassName() string { return "java.util.TreeMap$Values" }
func (this *treeMapValues) JIface_Collection() {}

func (this *treeMapValues) Iterator() Iterator {
	return &treeMapItr{m: this.m, next: this.m.getFirstEntry(), expectedModCount: this.m.modCount, kind: treeItrValues}
}

func (this *treeMapValues) Size() int32                { return this.m.Size() }
func (this *treeMapValues) IsEmpty() bool              { return this.Size() == 0 }
func (this *treeMapValues) Contains_Object(o any) bool { return this.m.ContainsValue_Object(o) }
func (this *treeMapValues) Clear()                     { this.m.Clear() }

func (this *treeMapValues) Remove_Object(o any) bool {
	for e := this.m.getFirstEntry(); e != nil; e = treeSuccessor(e) {
		if objectsEqual(e.value, o) {
			this.m.deleteEntry(e)
			return true
		}
	}
	return false
}

func (this *treeMapValues) Add_Object(any) bool                      { panic(unsupported()) }
func (this *treeMapValues) AddAll_Collection(c Collection) bool      { return acAddAll(this, c) }
func (this *treeMapValues) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }
func (this *treeMapValues) RemoveAll_Collection(c Collection) bool   { return acRemoveAll(this, c) }
func (this *treeMapValues) RetainAll_Collection(c Collection) bool   { return acRetainAll(this, c) }
func (this *treeMapValues) ToArray() []any                           { return acToArray(this) }
func (this *treeMapValues) ToArray_ObjectArr(a any) any              { return acToArrayT(this, a) }
func (this *treeMapValues) ToString() *String                        { return acToString(this, this) }

// treeMapEntrySet is TreeMap.EntrySet, the entry set view.
type treeMapEntrySet struct {
	m *TreeMap
}

func (this *treeMapEntrySet) JClassName() string { return "java.util.TreeMap$EntrySet" }
func (this *treeMapEntrySet) JIface_Collection() {}
func (this *treeMapEntrySet) JIface_Set()        {}

func (this *treeMapEntrySet) Iterator() Iterator {
	return &treeMapItr{m: this.m, next: this.m.getFirstEntry(), expectedModCount: this.m.modCount, kind: treeItrEntries}
}

func (this *treeMapEntrySet) Contains_Object(o any) bool {
	entry, ok := o.(Map_Entry)
	if !ok {
		return false
	}
	value := entry.GetValue()
	p := this.m.getEntry(entry.GetKey())
	return p != nil && objectsEqual(p.value, value)
}

func (this *treeMapEntrySet) Remove_Object(o any) bool {
	entry, ok := o.(Map_Entry)
	if !ok {
		return false
	}
	value := entry.GetValue()
	p := this.m.getEntry(entry.GetKey())
	if p != nil && objectsEqual(p.value, value) {
		this.m.deleteEntry(p)
		return true
	}
	return false
}

func (this *treeMapEntrySet) Size() int32   { return this.m.Size() }
func (this *treeMapEntrySet) IsEmpty() bool { return this.Size() == 0 }
func (this *treeMapEntrySet) Clear()        { this.m.Clear() }

func (this *treeMapEntrySet) Add_Object(any) bool                      { panic(unsupported()) }
func (this *treeMapEntrySet) AddAll_Collection(c Collection) bool      { return acAddAll(this, c) }
func (this *treeMapEntrySet) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }
func (this *treeMapEntrySet) RemoveAll_Collection(c Collection) bool   { return asRemoveAll(this, c) }
func (this *treeMapEntrySet) RetainAll_Collection(c Collection) bool   { return acRetainAll(this, c) }
func (this *treeMapEntrySet) ToArray() []any                           { return acToArray(this) }
func (this *treeMapEntrySet) ToArray_ObjectArr(a any) any              { return acToArrayT(this, a) }
func (this *treeMapEntrySet) Equals_Object(o any) bool                 { return asEquals(this, this, o) }
func (this *treeMapEntrySet) HashCode() int32                          { return asHashCode(this) }
func (this *treeMapEntrySet) ToString() *String                        { return acToString(this, this) }

// TreeSet is java.util.TreeSet, a set of the keys of a TreeMap.
type TreeSet struct {
	m *TreeMap
}

// treeSetPresent is the value TreeSet maps every key to, like Java's
// TreeSet.PRESENT. It is never mutated.
var treeSetPresent any = true

func (this *TreeSet) JClass_TreeSet() *TreeSet { return this }
func (this *TreeSet) JClassName() string       { return "java.util.TreeSet" }
func (this *TreeSet) JIface_Collection()       {}
func (this *TreeSet) JIface_Set()              {}
func (this *TreeSet) JIface_SortedSet()        {}
func (this *TreeSet) JIface_NavigableSet()     {}

func TreeSet_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_TreeSet() *TreeSet })
	return ok
}

func TreeSet_fromAny(o any) *TreeSet {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_TreeSet() *TreeSet }](o).JClass_TreeSet()
}

func TreeSet_new() *TreeSet { return &TreeSet{m: TreeMap_new()} }

func TreeSet_new_Comparator(c Comparator) *TreeSet {
	return &TreeSet{m: TreeMap_new_Comparator(c)}
}

func TreeSet_new_Collection(c Collection) *TreeSet {
	this := TreeSet_new()
	this.AddAll_Collection(c)
	return this
}

func TreeSet_new_SortedSet(s SortedSet) *TreeSet {
	this := TreeSet_new_Comparator(s.Comparator())
	this.AddAll_Collection(s)
	return this
}

func (this *TreeSet) Iterator() Iterator           { return this.m.keyIterator() }
func (this *TreeSet) DescendingIterator() Iterator { return this.m.descendingKeyIterator() }
func (this *TreeSet) Size() int32                  { return this.m.Size() }
func (this *TreeSet) IsEmpty() bool                { return this.m.IsEmpty() }
func (this *TreeSet) Contains_Object(o any) bool   { return this.m.ContainsKey_Object(o) }

func (this *TreeSet) Add_Object(e any) bool {
	return this.m.Put_Object_Object(e, treeSetPresent) == nil
}

func (this *TreeSet) Remove_Object(o any) bool { return this.m.Remove_Object(o) != nil }

func (this *TreeSet) Clear() { this.m.Clear() }

func (this *TreeSet) AddAll_Collection(c Collection) bool {
	if this.m.size == 0 && c.Size() > 0 {
		if ss, ok := c.(SortedSet); ok && javaObjectsEquals(this.m.comparator, ss.Comparator()) {
			this.m.modCount++
			this.m.buildFromSorted(ss.Size(), ss.Iterator(), &treeSetPresent)
			return true
		}
	}
	return acAddAll(this, c)
}

func (this *TreeSet) First() any             { return this.m.FirstKey() }
func (this *TreeSet) Last() any              { return this.m.LastKey() }
func (this *TreeSet) Lower_Object(e any) any { return this.m.LowerKey_Object(e) }
func (this *TreeSet) Floor_Object(e any) any { return this.m.FloorKey_Object(e) }

func (this *TreeSet) Ceiling_Object(e any) any { return this.m.CeilingKey_Object(e) }
func (this *TreeSet) Higher_Object(e any) any  { return this.m.HigherKey_Object(e) }
func (this *TreeSet) Comparator() Comparator   { return this.m.Comparator() }

func (this *TreeSet) PollFirst() any {
	e := this.m.PollFirstEntry()
	if e == nil {
		return nil
	}
	return e.GetKey()
}

func (this *TreeSet) PollLast() any {
	e := this.m.PollLastEntry()
	if e == nil {
		return nil
	}
	return e.GetKey()
}

// Clone implements TreeSet.clone, which copies the map with
// new TreeMap(SortedMap).
func (this *TreeSet) Clone() any { return &TreeSet{m: TreeMap_new_SortedMap(this.m)} }

func (this *TreeSet) ContainsAll_Collection(c Collection) bool { return acContainsAll(this, c) }
func (this *TreeSet) RemoveAll_Collection(c Collection) bool   { return asRemoveAll(this, c) }
func (this *TreeSet) RetainAll_Collection(c Collection) bool   { return acRetainAll(this, c) }
func (this *TreeSet) ToArray() []any                           { return acToArray(this) }
func (this *TreeSet) ToArray_ObjectArr(a any) any              { return acToArrayT(this, a) }
func (this *TreeSet) Equals_Object(o any) bool                 { return asEquals(this, this, o) }
func (this *TreeSet) HashCode() int32                          { return asHashCode(this) }
func (this *TreeSet) ToString() *String                        { return acToString(this, this) }
