package java

import (
	"math"
	"math/bits"
	"unsafe"
)

// The constants of java.util.HashMap.
const (
	hashMapDefaultInitialCapacity = 1 << 4
	hashMapMaximumCapacity        = 1 << 30
	hashMapDefaultLoadFactor      = float32(0.75)
	hashMapTreeifyThreshold       = 8
	hashMapUntreeifyThreshold     = 6
	hashMapMinTreeifyCapacity     = 64
)

// HashMap is java.util.HashMap, ported from JDK 21 so that iteration order,
// including the order within treeified bins, matches Java.
//
// A LinkedHashMap embeds a HashMap with linked set, and the HashMap methods
// perform LinkedHashMap's overrides of them.
type HashMap struct {
	jself      any
	table      []*hashNode
	entrySet   *hashEntrySet
	keySet     *hashKeySet
	values     *hashValues
	size       int32
	modCount   int32
	threshold  int32
	loadFactor float32

	// LinkedHashMap state.
	linked      bool
	accessOrder bool
	head        *hashNode
	tail        *hashNode
}

func (this *HashMap) JSelf() any        { return this.jself }
func (this *HashMap) JSetSelf(self any) { this.jself = self }

func (this *HashMap) JClass_HashMap() *HashMap { return this }
func (this *HashMap) JClassName() string       { return "java.util.HashMap" }
func (this *HashMap) JIface_Map()              {}

func HashMap_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_HashMap() *HashMap })
	return ok
}

func HashMap_fromAny(o any) *HashMap {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_HashMap() *HashMap }](o).JClass_HashMap()
}

// self returns the whole object, which is a LinkedHashMap when this is
// embedded in one.
func (this *HashMap) self() any {
	if this.jself != nil {
		return this.jself
	}
	return this
}

func HashMap_new() *HashMap {
	return &HashMap{loadFactor: hashMapDefaultLoadFactor}
}

func HashMap_new_int(initialCapacity int32) *HashMap {
	return HashMap_new_int_float(initialCapacity, hashMapDefaultLoadFactor)
}

func HashMap_new_int_float(initialCapacity int32, loadFactor float32) *HashMap {
	this := &HashMap{}
	this.init(initialCapacity, loadFactor)
	return this
}

func HashMap_new_Map(m Map) *HashMap {
	this := HashMap_new()
	this.putMapEntries(m)
	return this
}

// init is the HashMap(int, float) constructor. The initial capacity is held
// in threshold until the table is allocated.
func (this *HashMap) init(initialCapacity int32, loadFactor float32) {
	if initialCapacity < 0 {
		panic(IllegalArgumentException_new_String(Concat(Lit("Illegal initial capacity: "), StrOfInt(initialCapacity))))
	}
	if initialCapacity > hashMapMaximumCapacity {
		initialCapacity = hashMapMaximumCapacity
	}
	if loadFactor <= 0 || loadFactor != loadFactor {
		panic(IllegalArgumentException_new_String(Concat(Lit("Illegal load factor: "), StrOfFloat(loadFactor))))
	}
	this.loadFactor = loadFactor
	this.threshold = hashMapTableSizeFor(initialCapacity)
}

// hashMapHash is HashMap.hash: the key's hashCode with its high bits spread
// into the low bits.
func hashMapHash(key any) int32 {
	if key == nil {
		return 0
	}
	h := Object_HashCode(key)
	return h ^ int32(uint32(h)>>16)
}

// hashMapTableSizeFor is HashMap.tableSizeFor: the smallest power of two that
// is at least c, clamped to [1, MAXIMUM_CAPACITY].
func hashMapTableSizeFor(c int32) int32 {
	n := int32(uint32(math.MaxUint32) >> (uint(bits.LeadingZeros32(uint32(c-1))) & 31))
	if n < 0 {
		return 1
	}
	if n >= hashMapMaximumCapacity {
		return hashMapMaximumCapacity
	}
	return n + 1
}

// hashMapCapacityFor is HashMap.calculateHashMapCapacity.
func hashMapCapacityFor(numMappings int32) int32 {
	return int32(math.Ceil(float64(numMappings) / float64(hashMapDefaultLoadFactor)))
}

// keyMatches reports whether the key k of a node with the right hash is key,
// testing identity before equals as HashMap does.
func keyMatches(k, key any) bool {
	return sameRef(k, key) || (key != nil && Object_Equals_Object(key, k))
}

// putMapEntries puts m's mappings, first sizing the table for them.
func (this *HashMap) putMapEntries(m Map) {
	s := m.Size()
	if s <= 0 {
		return
	}
	if this.table == nil {
		dt := math.Ceil(float64(s) / float64(this.loadFactor))
		t := int32(hashMapMaximumCapacity)
		if dt < hashMapMaximumCapacity {
			t = int32(dt)
		}
		if t > this.threshold {
			this.threshold = hashMapTableSizeFor(t)
		}
	} else {
		for s > this.threshold && len(this.table) < hashMapMaximumCapacity {
			this.resize()
		}
	}
	for it := m.EntrySet().Iterator(); it.HasNext(); {
		e := it.Next().(Map_Entry)
		key := e.GetKey()
		value := e.GetValue()
		this.putVal(hashMapHash(key), key, value, false)
	}
}

func (this *HashMap) Size() int32   { return this.size }
func (this *HashMap) IsEmpty() bool { return this.size == 0 }

func (this *HashMap) Get_Object(key any) any {
	e := this.getNode(key)
	if e == nil {
		return nil
	}
	if this.accessOrder {
		this.afterNodeAccess(e)
	}
	return e.value
}

func (this *HashMap) GetOrDefault_Object_Object(key, defaultValue any) any {
	e := this.getNode(key)
	if e == nil {
		return defaultValue
	}
	if this.accessOrder {
		this.afterNodeAccess(e)
	}
	return e.value
}

// getNode finds the node for key. The key's hashCode is only computed when
// the table has been allocated, as in Java.
func (this *HashMap) getNode(key any) *hashNode {
	tab := this.table
	n := len(tab)
	if n == 0 {
		return nil
	}
	hash := hashMapHash(key)
	first := tab[int32(n-1)&hash]
	if first == nil {
		return nil
	}
	if first.hash == hash && keyMatches(first.key, key) {
		return first
	}
	e := first.next
	if e == nil {
		return nil
	}
	if first.kind == nodeTree {
		return asTree(first).getTreeNode(hash, key).node()
	}
	for ; e != nil; e = e.next {
		if e.hash == hash && keyMatches(e.key, key) {
			return e
		}
	}
	return nil
}

func (this *HashMap) ContainsKey_Object(key any) bool {
	return this.getNode(key) != nil
}

func (this *HashMap) Put_Object_Object(key, value any) any {
	return this.putVal(hashMapHash(key), key, value, false)
}

func (this *HashMap) PutIfAbsent_Object_Object(key, value any) any {
	return this.putVal(hashMapHash(key), key, value, true)
}

// putVal is HashMap.putVal. A new node goes at the tail of its bin. Java's
// afterNodeInsertion hook is omitted: it only acts for LinkedHashMap
// subclasses that override removeEldestEntry.
func (this *HashMap) putVal(hash int32, key, value any, onlyIfAbsent bool) any {
	tab := this.table
	n := len(tab)
	if n == 0 {
		tab = this.resize()
		n = len(tab)
	}
	i := int32(n-1) & hash
	p := tab[i]
	if p == nil {
		tab[i] = this.newNode(hash, key, value, nil)
	} else {
		var e *hashNode
		if p.hash == hash && keyMatches(p.key, key) {
			e = p
		} else if p.kind == nodeTree {
			e = asTree(p).putTreeVal(this, tab, hash, key, value).node()
		} else {
			for binCount := 0; ; binCount++ {
				e = p.next
				if e == nil {
					p.next = this.newNode(hash, key, value, nil)
					if binCount >= hashMapTreeifyThreshold-1 {
						this.treeifyBin(tab, hash)
					}
					break
				}
				if e.hash == hash && keyMatches(e.key, key) {
					break
				}
				p = e
			}
		}
		if e != nil {
			oldValue := e.value
			if !onlyIfAbsent || oldValue == nil {
				e.value = value
			}
			this.afterNodeAccess(e)
			return oldValue
		}
	}
	this.modCount++
	this.size++
	if this.size > this.threshold {
		this.resize()
	}
	return nil
}

// resize allocates or doubles the table. Each bin is split into a low and a
// high list that keep the relative order of the nodes.
func (this *HashMap) resize() []*hashNode {
	oldTab := this.table
	oldCap := int32(len(oldTab))
	oldThr := this.threshold
	var newCap, newThr int32
	if oldCap > 0 {
		if oldCap >= hashMapMaximumCapacity {
			this.threshold = math.MaxInt32
			return oldTab
		}
		newCap = oldCap << 1
		if newCap < hashMapMaximumCapacity && oldCap >= hashMapDefaultInitialCapacity {
			newThr = oldThr << 1
		}
	} else if oldThr > 0 {
		newCap = oldThr
	} else {
		newCap = hashMapDefaultInitialCapacity
		newThr = int32(hashMapDefaultLoadFactor * hashMapDefaultInitialCapacity)
	}
	if newThr == 0 {
		ft := float32(float32(newCap) * this.loadFactor)
		if newCap < hashMapMaximumCapacity && ft < hashMapMaximumCapacity {
			newThr = F2I(float64(ft))
		} else {
			newThr = math.MaxInt32
		}
	}
	this.threshold = newThr
	newTab := make([]*hashNode, newCap)
	this.table = newTab
	for j := int32(0); j < oldCap; j++ {
		e := oldTab[j]
		if e == nil {
			continue
		}
		oldTab[j] = nil
		if e.next == nil {
			newTab[e.hash&(newCap-1)] = e
		} else if e.kind == nodeTree {
			asTree(e).split(this, newTab, j, oldCap)
		} else {
			var loHead, loTail, hiHead, hiTail *hashNode
			for e != nil {
				next := e.next
				if e.hash&oldCap == 0 {
					if loTail == nil {
						loHead = e
					} else {
						loTail.next = e
					}
					loTail = e
				} else {
					if hiTail == nil {
						hiHead = e
					} else {
						hiTail.next = e
					}
					hiTail = e
				}
				e = next
			}
			if loTail != nil {
				loTail.next = nil
				newTab[j] = loHead
			}
			if hiTail != nil {
				hiTail.next = nil
				newTab[j+oldCap] = hiHead
			}
		}
	}
	return newTab
}

// treeifyBin converts the bin for hash to a tree, unless the table is too
// small, in which case it resizes instead.
func (this *HashMap) treeifyBin(tab []*hashNode, hash int32) {
	n := len(tab)
	if n < hashMapMinTreeifyCapacity {
		this.resize()
		return
	}
	index := int32(n-1) & hash
	e := tab[index]
	if e == nil {
		return
	}
	var hd, tl *treeNode
	for ; e != nil; e = e.next {
		p := this.replacementTreeNode(e, nil)
		if tl == nil {
			hd = p
		} else {
			p.prev = tl
			tl.next = p.node()
		}
		tl = p
	}
	tab[index] = hd.node()
	if hd != nil {
		hd.treeify(tab)
	}
}

func (this *HashMap) PutAll_Map(m Map) {
	this.putMapEntries(m)
}

func (this *HashMap) Remove_Object(key any) any {
	e := this.removeNode(hashMapHash(key), key, nil, false, true)
	if e == nil {
		return nil
	}
	return e.value
}

// Remove_Object_Object is Map.remove(key, value).
func (this *HashMap) Remove_Object_Object(key, value any) bool {
	return this.removeNode(hashMapHash(key), key, value, true, true) != nil
}

// removeNode is HashMap.removeNode. When movable is false, removal from a
// tree bin does not move the tree's root to the front of the bin.
func (this *HashMap) removeNode(hash int32, key, value any, matchValue, movable bool) *hashNode {
	tab := this.table
	n := len(tab)
	if n == 0 {
		return nil
	}
	index := int32(n-1) & hash
	p := tab[index]
	if p == nil {
		return nil
	}
	var node *hashNode
	if p.hash == hash && keyMatches(p.key, key) {
		node = p
	} else if e := p.next; e != nil {
		if p.kind == nodeTree {
			node = asTree(p).getTreeNode(hash, key).node()
		} else {
			for ; e != nil; e = e.next {
				if e.hash == hash && keyMatches(e.key, key) {
					node = e
					break
				}
				p = e
			}
		}
	}
	if node == nil {
		return nil
	}
	if matchValue {
		v := node.value
		if !sameRef(v, value) && (value == nil || !Object_Equals_Object(value, v)) {
			return nil
		}
	}
	if node.kind == nodeTree {
		asTree(node).removeTreeNode(this, tab, movable)
	} else if node == p {
		tab[index] = node.next
	} else {
		p.next = node.next
	}
	this.modCount++
	this.size--
	this.afterNodeRemoval(node)
	return node
}

// Clear empties the map but keeps the table, so the capacity is unchanged.
func (this *HashMap) Clear() {
	this.modCount++
	tab := this.table
	if tab != nil && this.size > 0 {
		this.size = 0
		for i := range tab {
			tab[i] = nil
		}
	}
	this.head = nil
	this.tail = nil
}

func (this *HashMap) ContainsValue_Object(value any) bool {
	found := false
	this.eachNode(func(e *hashNode) bool {
		v := e.value
		found = sameRef(v, value) || (value != nil && Object_Equals_Object(value, v))
		return !found
	})
	return found
}

// Replace_Object_Object_Object is Map.replace(key, oldValue, newValue).
func (this *HashMap) Replace_Object_Object_Object(key, oldValue, newValue any) bool {
	e := this.getNode(key)
	if e == nil {
		return false
	}
	v := e.value
	if !sameRef(v, oldValue) && (v == nil || !Object_Equals_Object(v, oldValue)) {
		return false
	}
	e.value = newValue
	this.afterNodeAccess(e)
	return true
}

// Replace_Object_Object is Map.replace(key, value).
func (this *HashMap) Replace_Object_Object(key, value any) any {
	e := this.getNode(key)
	if e == nil {
		return nil
	}
	oldValue := e.value
	e.value = value
	this.afterNodeAccess(e)
	return oldValue
}

func (this *HashMap) KeySet() Set {
	if this.keySet == nil {
		this.keySet = &hashKeySet{m: this}
	}
	return this.keySet
}

func (this *HashMap) Values() Collection {
	if this.values == nil {
		this.values = &hashValues{m: this}
	}
	return this.values
}

func (this *HashMap) EntrySet() Set {
	if this.entrySet == nil {
		this.entrySet = &hashEntrySet{m: this}
	}
	return this.entrySet
}

// Clone is HashMap.clone: a map of the same class whose entries are put in
// this map's iteration order into a table sized for them.
func (this *HashMap) Clone() any {
	var result *HashMap
	var out any
	if this.linked {
		l := &LinkedHashMap{}
		l.jself = l
		result = &l.HashMap
		out = l
	} else {
		result = &HashMap{}
		out = result
	}
	result.loadFactor = this.loadFactor
	result.linked = this.linked
	result.accessOrder = this.accessOrder
	result.putMapEntries(this)
	return out
}

func (this *HashMap) Equals_Object(o any) bool { return amEquals(this, this.self(), o) }
func (this *HashMap) HashCode() int32          { return amHashCode(this) }
func (this *HashMap) ToString() *String        { return amToString(this, this.self()) }

// eachNode calls f on each node in iteration order until f returns false.
func (this *HashMap) eachNode(f func(e *hashNode) bool) {
	if this.linked {
		for e := this.head; e != nil; e = e.links().after {
			if !f(e) {
				return
			}
		}
		return
	}
	if this.size == 0 {
		return
	}
	for _, e := range this.table {
		for ; e != nil; e = e.next {
			if !f(e) {
				return
			}
		}
	}
}

func (this *HashMap) keysArray() []any {
	out := make([]any, 0, this.size)
	this.eachNode(func(e *hashNode) bool {
		out = append(out, e.key)
		return true
	})
	return out
}

func (this *HashMap) valuesArray() []any {
	out := make([]any, 0, this.size)
	this.eachNode(func(e *hashNode) bool {
		out = append(out, e.value)
		return true
	})
	return out
}

// ---------------------------------------------------------------------------
// LinkedHashMap

// LinkedHashMap is java.util.LinkedHashMap. Its entries are also kept in a
// doubly linked list, in insertion order or, if accessOrder is set, in order
// of last access.
type LinkedHashMap struct {
	HashMap
}

func (this *LinkedHashMap) JClass_LinkedHashMap() *LinkedHashMap { return this }
func (this *LinkedHashMap) JClassName() string                   { return "java.util.LinkedHashMap" }

func LinkedHashMap_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_LinkedHashMap() *LinkedHashMap })
	return ok
}

func LinkedHashMap_fromAny(o any) *LinkedHashMap {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_LinkedHashMap() *LinkedHashMap }](o).JClass_LinkedHashMap()
}

func newLinkedHashMap() *LinkedHashMap {
	this := &LinkedHashMap{}
	this.jself = this
	this.linked = true
	this.loadFactor = hashMapDefaultLoadFactor
	return this
}

func LinkedHashMap_new() *LinkedHashMap {
	return newLinkedHashMap()
}

func LinkedHashMap_new_int(initialCapacity int32) *LinkedHashMap {
	return LinkedHashMap_new_int_float_boolean(initialCapacity, hashMapDefaultLoadFactor, false)
}

func LinkedHashMap_new_int_float(initialCapacity int32, loadFactor float32) *LinkedHashMap {
	return LinkedHashMap_new_int_float_boolean(initialCapacity, loadFactor, false)
}

func LinkedHashMap_new_Map(m Map) *LinkedHashMap {
	this := newLinkedHashMap()
	this.putMapEntries(m)
	return this
}

func LinkedHashMap_new_int_float_boolean(initialCapacity int32, loadFactor float32, accessOrder bool) *LinkedHashMap {
	this := newLinkedHashMap()
	this.init(initialCapacity, loadFactor)
	this.accessOrder = accessOrder
	return this
}

// newNode creates a node for a new mapping, linking it at the end of a
// LinkedHashMap's list.
func (this *HashMap) newNode(hash int32, key, value any, next *hashNode) *hashNode {
	if !this.linked {
		return &hashNode{hash: hash, key: key, value: value, next: next}
	}
	p := &linkedNode{hashNode: hashNode{hash: hash, kind: nodeLinked, key: key, value: value, next: next}}
	this.linkNodeAtEnd(&p.hashNode)
	return &p.hashNode
}

// replacementNode converts a tree node to a plain node.
func (this *HashMap) replacementNode(p, next *hashNode) *hashNode {
	if !this.linked {
		return &hashNode{hash: p.hash, key: p.key, value: p.value, next: next}
	}
	t := &linkedNode{hashNode: hashNode{hash: p.hash, kind: nodeLinked, key: p.key, value: p.value, next: next}}
	this.transferLinks(p, &t.hashNode)
	return &t.hashNode
}

func (this *HashMap) newTreeNode(hash int32, key, value any, next *hashNode) *treeNode {
	p := &treeNode{}
	p.hashNode = hashNode{hash: hash, kind: nodeTree, key: key, value: value, next: next}
	if this.linked {
		this.linkNodeAtEnd(p.node())
	}
	return p
}

// replacementTreeNode converts a plain node to a tree node.
func (this *HashMap) replacementTreeNode(p, next *hashNode) *treeNode {
	t := &treeNode{}
	t.hashNode = hashNode{hash: p.hash, kind: nodeTree, key: p.key, value: p.value, next: next}
	if this.linked {
		this.transferLinks(p, t.node())
	}
	return t
}

func (this *HashMap) linkNodeAtEnd(p *hashNode) {
	last := this.tail
	this.tail = p
	if last == nil {
		this.head = p
	} else {
		p.links().before = last
		last.links().after = p
	}
}

// transferLinks puts dst in src's place in the linked list.
func (this *HashMap) transferLinks(src, dst *hashNode) {
	b := src.links().before
	a := src.links().after
	dst.links().before = b
	dst.links().after = a
	if b == nil {
		this.head = dst
	} else {
		b.links().after = dst
	}
	if a == nil {
		this.tail = dst
	} else {
		a.links().before = dst
	}
}

func (this *HashMap) afterNodeRemoval(e *hashNode) {
	if !this.linked {
		return
	}
	p := e.links()
	b := p.before
	a := p.after
	p.before = nil
	p.after = nil
	if b == nil {
		this.head = a
	} else {
		b.links().after = a
	}
	if a == nil {
		this.tail = b
	} else {
		a.links().before = b
	}
}

// afterNodeAccess moves e to the end of an access-ordered LinkedHashMap's
// list.
func (this *HashMap) afterNodeAccess(e *hashNode) {
	if !this.accessOrder {
		return
	}
	last := this.tail
	if last == e {
		return
	}
	p := e.links()
	b := p.before
	a := p.after
	p.after = nil
	if b == nil {
		this.head = a
	} else {
		b.links().after = a
	}
	if a != nil {
		a.links().before = b
	} else {
		last = b
	}
	if last == nil {
		this.head = e
	} else {
		p.before = last
		last.links().after = e
	}
	this.tail = e
	this.modCount++
}

// ---------------------------------------------------------------------------
// Nodes

// The kinds of hashNode, which are distinct classes in Java.
const (
	nodePlain  uint8 = iota // HashMap.Node
	nodeLinked              // LinkedHashMap.Entry
	nodeTree                // HashMap.TreeNode
)

// hashNode is HashMap.Node, the Map.Entry for a mapping. A node of kind
// nodeLinked is the first field of a linkedNode, and one of kind nodeTree the
// first field of a treeNode.
type hashNode struct {
	hash  int32
	kind  uint8
	key   any
	value any
	next  *hashNode
}

// linkedNode is LinkedHashMap.Entry, a node of a LinkedHashMap's list.
type linkedNode struct {
	hashNode
	before *hashNode
	after  *hashNode
}

// links returns n's LinkedHashMap.Entry links. All nodes of a LinkedHashMap
// have them.
func (n *hashNode) links() *linkedNode {
	if n.kind == nodePlain {
		panic("java: HashMap node is not a LinkedHashMap entry")
	}
	return (*linkedNode)(unsafe.Pointer(n))
}

func (e *hashNode) JIface_Map_Entry() {}

func (e *hashNode) JClassName() string {
	switch e.kind {
	case nodeLinked:
		return "java.util.LinkedHashMap$Entry"
	case nodeTree:
		return "java.util.HashMap$TreeNode"
	}
	return "java.util.HashMap$Node"
}

func (e *hashNode) GetKey() any   { return e.key }
func (e *hashNode) GetValue() any { return e.value }

func (e *hashNode) SetValue_Object(value any) any {
	old := e.value
	e.value = value
	return old
}

func (e *hashNode) ToString() *String {
	return Concat(StrOfObject(e.key), Lit("="), StrOfObject(e.value))
}

func (e *hashNode) HashCode() int32 {
	return hashOrZero(e.key) ^ hashOrZero(e.value)
}

func (e *hashNode) Equals_Object(o any) bool {
	if sameRef(o, e) {
		return true
	}
	x, ok := o.(Map_Entry)
	return ok && javaObjectsEquals(e.key, x.GetKey()) && javaObjectsEquals(e.value, x.GetValue())
}

// treeNode is HashMap.TreeNode, a node of a bin that holds a red-black tree.
// The bin's next links (with prev as their reverse) still list all its nodes,
// and give its iteration order.
type treeNode struct {
	linkedNode
	parent *treeNode
	left   *treeNode
	right  *treeNode
	prev   *treeNode
	red    bool
}

// asTree is a Java cast of a node to TreeNode.
func asTree(n *hashNode) *treeNode {
	if n != nil && n.kind != nodeTree {
		panic("java: HashMap node is not a tree node")
	}
	return (*treeNode)(unsafe.Pointer(n))
}

// node returns t as a hashNode, mapping nil to nil.
func (t *treeNode) node() *hashNode {
	return (*hashNode)(unsafe.Pointer(t))
}

// comparableClassFor is HashMap.comparableClassFor: the name of x's class if
// it is Comparable to itself, otherwise "". That is the case for String and
// the boxed primitives, and for a class that directly implements
// Comparable<itself>; the translator gives such classes a JComparableSelf
// method that returns the class name. A subclass inherits the method but is
// not such a class, which comparing the name with the object's class name
// detects.
func comparableClassFor(x any) string {
	switch v := x.(type) {
	case *String:
		return "java.lang.String"
	case int32, int64, int16, int8, uint16, bool, float64, float32:
		return className(x)
	case interface{ JComparableSelf() string }:
		if c := v.JComparableSelf(); c == className(x) {
			return c
		}
	}
	return ""
}

// compareComparables is HashMap.compareComparables: k.compareTo(x) if x has
// class kc, otherwise 0.
func compareComparables(kc string, k, x any) int32 {
	if x == nil || className(x) != kc {
		return 0
	}
	return Comparable_CompareTo_Object(k, x)
}

// treeTieBreakOrder is TreeNode.tieBreakOrder, which orders keys with equal
// hashes that are not comparable by class name, then identity hash code.
func treeTieBreakOrder(a, b any) int32 {
	var d int32
	if a != nil && b != nil {
		d = FromGo(className(a)).CompareTo_Object(FromGo(className(b)))
	}
	if a == nil || b == nil || d == 0 {
		ha := IdentityHashCode(a)
		hb := IdentityHashCode(b)
		if ha <= hb {
			d = -1
		} else {
			d = 1
		}
	}
	return d
}

func (t *treeNode) root() *treeNode {
	r := t
	for r.parent != nil {
		r = r.parent
	}
	return r
}

// treeMoveRootToFront makes root the first node of its bin.
func treeMoveRootToFront(tab []*hashNode, root *treeNode) {
	n := len(tab)
	if root == nil || n == 0 {
		return
	}
	index := int32(n-1) & root.hash
	first := asTree(tab[index])
	if root == first {
		return
	}
	tab[index] = root.node()
	rp := root.prev
	rn := root.next
	if rn != nil {
		asTree(rn).prev = rp
	}
	if rp != nil {
		rp.next = rn
	}
	if first != nil {
		first.prev = root
	}
	root.next = first.node()
	root.prev = nil
}

// find finds the node for key k with hash h in the subtree rooted at t. kc is
// k's comparable class if known.
func (t *treeNode) find(h int32, k any, kc string) *treeNode {
	p := t
	for p != nil {
		pl := p.left
		pr := p.right
		if ph := p.hash; ph > h {
			p = pl
		} else if ph < h {
			p = pr
		} else if pk := p.key; sameRef(pk, k) || (k != nil && Object_Equals_Object(k, pk)) {
			return p
		} else if pl == nil {
			p = pr
		} else if pr == nil {
			p = pl
		} else {
			if kc == "" {
				kc = comparableClassFor(k)
			}
			if kc != "" {
				if dir := compareComparables(kc, k, pk); dir != 0 {
					if dir < 0 {
						p = pl
					} else {
						p = pr
					}
					continue
				}
			}
			if q := pr.find(h, k, kc); q != nil {
				return q
			}
			p = pl
		}
	}
	return nil
}

func (t *treeNode) getTreeNode(h int32, k any) *treeNode {
	r := t
	if t.parent != nil {
		r = t.root()
	}
	return r.find(h, k, "")
}

// treeify builds a tree from the nodes linked from t.
func (t *treeNode) treeify(tab []*hashNode) {
	var root *treeNode
	for x := t; x != nil; {
		next := asTree(x.next)
		x.left = nil
		x.right = nil
		if root == nil {
			x.parent = nil
			x.red = false
			root = x
		} else {
			k := x.key
			h := x.hash
			kc := ""
			for p := root; ; {
				var dir int32
				pk := p.key
				if ph := p.hash; ph > h {
					dir = -1
				} else if ph < h {
					dir = 1
				} else {
					if kc == "" {
						kc = comparableClassFor(k)
					}
					if kc != "" {
						dir = compareComparables(kc, k, pk)
					}
					if dir == 0 {
						dir = treeTieBreakOrder(k, pk)
					}
				}
				xp := p
				if dir <= 0 {
					p = p.left
				} else {
					p = p.right
				}
				if p == nil {
					x.parent = xp
					if dir <= 0 {
						xp.left = x
					} else {
						xp.right = x
					}
					root = treeBalanceInsertion(root, x)
					break
				}
			}
		}
		x = next
	}
	treeMoveRootToFront(tab, root)
}

// untreeify returns a list of plain nodes replacing those linked from t.
func (t *treeNode) untreeify(m *HashMap) *hashNode {
	var hd, tl *hashNode
	for q := t.node(); q != nil; q = q.next {
		p := m.replacementNode(q, nil)
		if tl == nil {
			hd = p
		} else {
			tl.next = p
		}
		tl = p
	}
	return hd
}

// putTreeVal is TreeNode.putTreeVal. It returns the existing node for k, or
// nil after adding a node. A new node is linked into the bin's list directly
// after its tree parent.
func (t *treeNode) putTreeVal(m *HashMap, tab []*hashNode, h int32, k, v any) *treeNode {
	kc := ""
	searched := false
	root := t
	if t.parent != nil {
		root = t.root()
	}
	for p := root; ; {
		var dir int32
		if ph := p.hash; ph > h {
			dir = -1
		} else if ph < h {
			dir = 1
		} else if pk := p.key; sameRef(pk, k) || (k != nil && Object_Equals_Object(k, pk)) {
			return p
		} else {
			if kc == "" {
				kc = comparableClassFor(k)
			}
			if kc != "" {
				dir = compareComparables(kc, k, pk)
			}
			if dir == 0 {
				if !searched {
					searched = true
					if ch := p.left; ch != nil {
						if q := ch.find(h, k, kc); q != nil {
							return q
						}
					}
					if ch := p.right; ch != nil {
						if q := ch.find(h, k, kc); q != nil {
							return q
						}
					}
				}
				dir = treeTieBreakOrder(k, pk)
			}
		}
		xp := p
		if dir <= 0 {
			p = p.left
		} else {
			p = p.right
		}
		if p == nil {
			xpn := xp.next
			x := m.newTreeNode(h, k, v, xpn)
			if dir <= 0 {
				xp.left = x
			} else {
				xp.right = x
			}
			xp.next = x.node()
			x.parent = xp
			x.prev = xp
			if xpn != nil {
				asTree(xpn).prev = x
			}
			treeMoveRootToFront(tab, treeBalanceInsertion(root, x))
			return nil
		}
	}
}

// removeTreeNode is TreeNode.removeTreeNode, which removes t from its bin and
// converts the bin back to plain nodes when the tree has become too small.
func (t *treeNode) removeTreeNode(m *HashMap, tab []*hashNode, movable bool) {
	n := len(tab)
	if n == 0 {
		return
	}
	index := int32(n-1) & t.hash
	first := asTree(tab[index])
	root := first
	succ := asTree(t.next)
	pred := t.prev
	if pred == nil {
		first = succ
		tab[index] = first.node()
	} else {
		pred.next = succ.node()
	}
	if succ != nil {
		succ.prev = pred
	}
	if first == nil {
		return
	}
	if root.parent != nil {
		root = root.root()
	}
	if root == nil || (movable && (root.right == nil || root.left == nil || root.left.left == nil)) {
		tab[index] = first.untreeify(m)
		return
	}
	p := t
	pl := t.left
	pr := t.right
	var replacement *treeNode
	if pl != nil && pr != nil {
		s := pr
		for s.left != nil {
			s = s.left
		}
		s.red, p.red = p.red, s.red
		sr := s.right
		pp := p.parent
		if s == pr {
			p.parent = s
			s.right = p
		} else {
			sp := s.parent
			p.parent = sp
			if sp != nil {
				if s == sp.left {
					sp.left = p
				} else {
					sp.right = p
				}
			}
			s.right = pr
			if pr != nil {
				pr.parent = s
			}
		}
		p.left = nil
		p.right = sr
		if sr != nil {
			sr.parent = p
		}
		s.left = pl
		if pl != nil {
			pl.parent = s
		}
		s.parent = pp
		if pp == nil {
			root = s
		} else if p == pp.left {
			pp.left = s
		} else {
			pp.right = s
		}
		if sr != nil {
			replacement = sr
		} else {
			replacement = p
		}
	} else if pl != nil {
		replacement = pl
	} else if pr != nil {
		replacement = pr
	} else {
		replacement = p
	}
	if replacement != p {
		pp := p.parent
		replacement.parent = pp
		if pp == nil {
			root = replacement
			root.red = false
		} else if p == pp.left {
			pp.left = replacement
		} else {
			pp.right = replacement
		}
		p.left = nil
		p.right = nil
		p.parent = nil
	}
	r := root
	if !p.red {
		r = treeBalanceDeletion(root, replacement)
	}
	if replacement == p {
		pp := p.parent
		p.parent = nil
		if pp != nil {
			if p == pp.left {
				pp.left = nil
			} else if p == pp.right {
				pp.right = nil
			}
		}
	}
	if movable {
		treeMoveRootToFront(tab, r)
	}
}

// split divides the tree bin at index of the old table between index and
// index+bit of the new table tab, keeping the order of the nodes, and
// converts a part to plain nodes if it is small.
func (t *treeNode) split(m *HashMap, tab []*hashNode, index, bit int32) {
	var loHead, loTail, hiHead, hiTail *treeNode
	lc, hc := 0, 0
	for e := t; e != nil; {
		next := asTree(e.next)
		e.next = nil
		if e.hash&bit == 0 {
			e.prev = loTail
			if loTail == nil {
				loHead = e
			} else {
				loTail.next = e.node()
			}
			loTail = e
			lc++
		} else {
			e.prev = hiTail
			if hiTail == nil {
				hiHead = e
			} else {
				hiTail.next = e.node()
			}
			hiTail = e
			hc++
		}
		e = next
	}
	if loHead != nil {
		if lc <= hashMapUntreeifyThreshold {
			tab[index] = loHead.untreeify(m)
		} else {
			tab[index] = loHead.node()
			if hiHead != nil {
				loHead.treeify(tab)
			}
		}
	}
	if hiHead != nil {
		if hc <= hashMapUntreeifyThreshold {
			tab[index+bit] = hiHead.untreeify(m)
		} else {
			tab[index+bit] = hiHead.node()
			if loHead != nil {
				hiHead.treeify(tab)
			}
		}
	}
}

// The red-black tree methods below are TreeNode's, which are adapted from
// CLR.

func treeRotateLeft(root, p *treeNode) *treeNode {
	if p == nil {
		return root
	}
	r := p.right
	if r == nil {
		return root
	}
	rl := r.left
	p.right = rl
	if rl != nil {
		rl.parent = p
	}
	pp := p.parent
	r.parent = pp
	if pp == nil {
		root = r
		root.red = false
	} else if pp.left == p {
		pp.left = r
	} else {
		pp.right = r
	}
	r.left = p
	p.parent = r
	return root
}

func treeRotateRight(root, p *treeNode) *treeNode {
	if p == nil {
		return root
	}
	l := p.left
	if l == nil {
		return root
	}
	lr := l.right
	p.left = lr
	if lr != nil {
		lr.parent = p
	}
	pp := p.parent
	l.parent = pp
	if pp == nil {
		root = l
		root.red = false
	} else if pp.right == p {
		pp.right = l
	} else {
		pp.left = l
	}
	l.right = p
	p.parent = l
	return root
}

func treeBalanceInsertion(root, x *treeNode) *treeNode {
	x.red = true
	for {
		xp := x.parent
		if xp == nil {
			x.red = false
			return x
		}
		if !xp.red {
			return root
		}
		xpp := xp.parent
		if xpp == nil {
			return root
		}
		if xppl := xpp.left; xp == xppl {
			if xppr := xpp.right; xppr != nil && xppr.red {
				xppr.red = false
				xp.red = false
				xpp.red = true
				x = xpp
			} else {
				if x == xp.right {
					x = xp
					root = treeRotateLeft(root, x)
					xp = x.parent
					xpp = nil
					if xp != nil {
						xpp = xp.parent
					}
				}
				if xp != nil {
					xp.red = false
					if xpp != nil {
						xpp.red = true
						root = treeRotateRight(root, xpp)
					}
				}
			}
		} else {
			if xppl != nil && xppl.red {
				xppl.red = false
				xp.red = false
				xpp.red = true
				x = xpp
			} else {
				if x == xp.left {
					x = xp
					root = treeRotateRight(root, x)
					xp = x.parent
					xpp = nil
					if xp != nil {
						xpp = xp.parent
					}
				}
				if xp != nil {
					xp.red = false
					if xpp != nil {
						xpp.red = true
						root = treeRotateLeft(root, xpp)
					}
				}
			}
		}
	}
}

func treeBalanceDeletion(root, x *treeNode) *treeNode {
	for {
		if x == nil || x == root {
			return root
		}
		xp := x.parent
		if xp == nil {
			x.red = false
			return x
		}
		if x.red {
			x.red = false
			return root
		}
		if xpl := xp.left; xpl == x {
			xpr := xp.right
			if xpr != nil && xpr.red {
				xpr.red = false
				xp.red = true
				root = treeRotateLeft(root, xp)
				xp = x.parent
				xpr = nil
				if xp != nil {
					xpr = xp.right
				}
			}
			if xpr == nil {
				x = xp
			} else {
				sl := xpr.left
				sr := xpr.right
				if (sr == nil || !sr.red) && (sl == nil || !sl.red) {
					xpr.red = true
					x = xp
				} else {
					if sr == nil || !sr.red {
						if sl != nil {
							sl.red = false
						}
						xpr.red = true
						root = treeRotateRight(root, xpr)
						xp = x.parent
						xpr = nil
						if xp != nil {
							xpr = xp.right
						}
					}
					if xpr != nil {
						xpr.red = xp != nil && xp.red
						if sr = xpr.right; sr != nil {
							sr.red = false
						}
					}
					if xp != nil {
						xp.red = false
						root = treeRotateLeft(root, xp)
					}
					x = root
				}
			}
		} else {
			if xpl != nil && xpl.red {
				xpl.red = false
				xp.red = true
				root = treeRotateRight(root, xp)
				xp = x.parent
				xpl = nil
				if xp != nil {
					xpl = xp.left
				}
			}
			if xpl == nil {
				x = xp
			} else {
				sl := xpl.left
				sr := xpl.right
				if (sl == nil || !sl.red) && (sr == nil || !sr.red) {
					xpl.red = true
					x = xp
				} else {
					if sl == nil || !sl.red {
						if sr != nil {
							sr.red = false
						}
						xpl.red = true
						root = treeRotateLeft(root, xpl)
						xp = x.parent
						xpl = nil
						if xp != nil {
							xpl = xp.left
						}
					}
					if xpl != nil {
						xpl.red = xp != nil && xp.red
						if sl = xpl.left; sl != nil {
							sl.red = false
						}
					}
					if xp != nil {
						xp.red = false
						root = treeRotateRight(root, xp)
					}
					x = root
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Iterators and views

// The kinds of hashIterator.
const (
	iterKeys uint8 = iota
	iterValues
	iterEntries
)

// hashIterator is HashMap.HashIterator and LinkedHashMap.LinkedHashIterator.
type hashIterator struct {
	m                *HashMap
	next             *hashNode
	current          *hashNode
	expectedModCount int32
	index            int32
	kind             uint8
}

func (this *HashMap) newIterator(kind uint8) *hashIterator {
	it := &hashIterator{m: this, kind: kind, expectedModCount: this.modCount}
	if this.linked {
		it.next = this.head
		return it
	}
	if t := this.table; t != nil && this.size > 0 {
		it.advance(t)
	}
	return it
}

// advance sets next to the first node of the next non-empty bin.
func (it *hashIterator) advance(t []*hashNode) {
	for it.index < int32(len(t)) {
		it.next = t[it.index]
		it.index++
		if it.next != nil {
			return
		}
	}
}

func (it *hashIterator) HasNext() bool { return it.next != nil }

func (it *hashIterator) nextNode() *hashNode {
	e := it.next
	if it.m.modCount != it.expectedModCount {
		panic(ConcurrentModificationException_new())
	}
	if e == nil {
		panic(NoSuchElementException_new())
	}
	it.current = e
	if it.m.linked {
		it.next = e.links().after
		return e
	}
	it.next = e.next
	if it.next == nil {
		if t := it.m.table; t != nil {
			it.advance(t)
		}
	}
	return e
}

func (it *hashIterator) Next() any {
	e := it.nextNode()
	switch it.kind {
	case iterKeys:
		return e.key
	case iterValues:
		return e.value
	}
	return e
}

// Remove removes the current mapping without moving a tree bin's root, as
// Java's iterators do.
func (it *hashIterator) Remove() {
	p := it.current
	if p == nil {
		panic(IllegalStateException_new())
	}
	if it.m.modCount != it.expectedModCount {
		panic(ConcurrentModificationException_new())
	}
	it.current = nil
	it.m.removeNode(p.hash, p.key, nil, false, false)
	it.expectedModCount = it.m.modCount
}

// hashKeySet is HashMap.KeySet and LinkedHashMap.LinkedKeySet.
type hashKeySet struct {
	m *HashMap
}

func (s *hashKeySet) JIface_Collection() {}
func (s *hashKeySet) JIface_Set()        {}

func (s *hashKeySet) JClassName() string {
	if s.m.linked {
		return "java.util.LinkedHashMap$LinkedKeySet"
	}
	return "java.util.HashMap$KeySet"
}

func (s *hashKeySet) Iterator() Iterator         { return s.m.newIterator(iterKeys) }
func (s *hashKeySet) Size() int32                { return s.m.size }
func (s *hashKeySet) IsEmpty() bool              { return s.m.size == 0 }
func (s *hashKeySet) Contains_Object(o any) bool { return s.m.ContainsKey_Object(o) }
func (s *hashKeySet) Add_Object(any) bool        { panic(UnsupportedOperationException_new()) }
func (s *hashKeySet) Clear()                     { s.m.Clear() }
func (s *hashKeySet) ToArray() []any             { return s.m.keysArray() }
func (s *hashKeySet) ToArray_ObjectArr(a any) any {
	return fillArray(a, s.m.keysArray())
}

func (s *hashKeySet) Remove_Object(key any) bool {
	return s.m.removeNode(hashMapHash(key), key, nil, false, true) != nil
}

func (s *hashKeySet) AddAll_Collection(c Collection) bool      { return acAddAll(s, c) }
func (s *hashKeySet) ContainsAll_Collection(c Collection) bool { return acContainsAll(s, c) }
func (s *hashKeySet) RemoveAll_Collection(c Collection) bool   { return asRemoveAll(s, c) }
func (s *hashKeySet) RetainAll_Collection(c Collection) bool   { return acRetainAll(s, c) }
func (s *hashKeySet) Equals_Object(o any) bool                 { return asEquals(s, s, o) }
func (s *hashKeySet) HashCode() int32                          { return asHashCode(s) }
func (s *hashKeySet) ToString() *String                        { return acToString(s, s) }

// hashValues is HashMap.Values and LinkedHashMap.LinkedValues. Like any
// AbstractCollection that is not a Set or List, it has identity equality.
type hashValues struct {
	m     *HashMap
	jhash int32
}

func (s *hashValues) JIface_Collection() {}

func (s *hashValues) JClassName() string {
	if s.m.linked {
		return "java.util.LinkedHashMap$LinkedValues"
	}
	return "java.util.HashMap$Values"
}

func (s *hashValues) JIdentityHash() *int32 { return &s.jhash }

func (s *hashValues) Iterator() Iterator         { return s.m.newIterator(iterValues) }
func (s *hashValues) Size() int32                { return s.m.size }
func (s *hashValues) IsEmpty() bool              { return s.m.size == 0 }
func (s *hashValues) Contains_Object(o any) bool { return s.m.ContainsValue_Object(o) }
func (s *hashValues) Add_Object(any) bool        { panic(UnsupportedOperationException_new()) }
func (s *hashValues) Remove_Object(o any) bool   { return acRemove(s, o) }
func (s *hashValues) Clear()                     { s.m.Clear() }
func (s *hashValues) ToArray() []any             { return s.m.valuesArray() }
func (s *hashValues) ToArray_ObjectArr(a any) any {
	return fillArray(a, s.m.valuesArray())
}

func (s *hashValues) AddAll_Collection(c Collection) bool      { return acAddAll(s, c) }
func (s *hashValues) ContainsAll_Collection(c Collection) bool { return acContainsAll(s, c) }
func (s *hashValues) RemoveAll_Collection(c Collection) bool   { return acRemoveAll(s, c) }
func (s *hashValues) RetainAll_Collection(c Collection) bool   { return acRetainAll(s, c) }
func (s *hashValues) Equals_Object(o any) bool                 { return sameRef(o, s) }
func (s *hashValues) HashCode() int32                          { return IdentityHashCode(s) }
func (s *hashValues) ToString() *String                        { return acToString(s, s) }

// hashEntrySet is HashMap.EntrySet and LinkedHashMap.LinkedEntrySet.
type hashEntrySet struct {
	m *HashMap
}

func (s *hashEntrySet) JIface_Collection() {}
func (s *hashEntrySet) JIface_Set()        {}

func (s *hashEntrySet) JClassName() string {
	if s.m.linked {
		return "java.util.LinkedHashMap$LinkedEntrySet"
	}
	return "java.util.HashMap$EntrySet"
}

func (s *hashEntrySet) Iterator() Iterator  { return s.m.newIterator(iterEntries) }
func (s *hashEntrySet) Size() int32         { return s.m.size }
func (s *hashEntrySet) IsEmpty() bool       { return s.m.size == 0 }
func (s *hashEntrySet) Add_Object(any) bool { panic(UnsupportedOperationException_new()) }
func (s *hashEntrySet) Clear()              { s.m.Clear() }

func (s *hashEntrySet) Contains_Object(o any) bool {
	e, ok := o.(Map_Entry)
	if !ok {
		return false
	}
	candidate := s.m.getNode(e.GetKey())
	return candidate != nil && candidate.Equals_Object(e)
}

func (s *hashEntrySet) Remove_Object(o any) bool {
	e, ok := o.(Map_Entry)
	if !ok {
		return false
	}
	key := e.GetKey()
	value := e.GetValue()
	return s.m.removeNode(hashMapHash(key), key, value, true, true) != nil
}

func (s *hashEntrySet) ToArray() []any {
	out := make([]any, 0, s.m.size)
	s.m.eachNode(func(e *hashNode) bool {
		out = append(out, e)
		return true
	})
	return out
}

func (s *hashEntrySet) ToArray_ObjectArr(a any) any {
	return fillArray(a, s.ToArray())
}

func (s *hashEntrySet) AddAll_Collection(c Collection) bool      { return acAddAll(s, c) }
func (s *hashEntrySet) ContainsAll_Collection(c Collection) bool { return acContainsAll(s, c) }
func (s *hashEntrySet) RemoveAll_Collection(c Collection) bool   { return asRemoveAll(s, c) }
func (s *hashEntrySet) RetainAll_Collection(c Collection) bool   { return acRetainAll(s, c) }
func (s *hashEntrySet) Equals_Object(o any) bool                 { return asEquals(s, s, o) }
func (s *hashEntrySet) HashCode() int32                          { return asHashCode(s) }
func (s *hashEntrySet) ToString() *String                        { return acToString(s, s) }
