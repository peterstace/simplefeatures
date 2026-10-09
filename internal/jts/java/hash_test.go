package java

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

// The vectors in testdata/hash.txt come from
// internal/jts/tools/jdkvectors/HashVectors.java, which documents the
// format. Each sequence is replayed twice: once with every key's identity
// hash code pinned to the value it had in Java, and once with identity hash
// codes assigned lazily from the counter, starting where Java's was, which
// also checks that the Go code computes identity hash codes in the same
// order as Java.

// hvCKey is HashVectors.CKey (or CKeySub, if sub is set).
type hvCKey struct {
	idx, id, hash, rank int32
	sub                 bool
	jhash               int32
}

func (k *hvCKey) HashCode() int32 { return k.hash }

func (k *hvCKey) Equals_Object(o any) bool {
	x, ok := o.(*hvCKey)
	return ok && x.sub == k.sub && x.id == k.id
}

func (k *hvCKey) CompareTo_Object(o any) int32 {
	return Integer_compare(k.rank, o.(*hvCKey).rank)
}

// JComparableSelf is inherited by CKeySub, as a Go method of a generated
// class is inherited by its subclasses.
func (k *hvCKey) JComparableSelf() string { return "HashVectors$CKey" }

func (k *hvCKey) JClassName() string {
	if k.sub {
		return "HashVectors$CKeySub"
	}
	return "HashVectors$CKey"
}

func (k *hvCKey) JIdentityHash() *int32 { return &k.jhash }

func (k *hvCKey) ToString() *String {
	if k.sub {
		return FromGo(fmt.Sprintf("t%d", k.idx))
	}
	return FromGo(fmt.Sprintf("c%d", k.idx))
}

// hvNKey is HashVectors.NKey.
type hvNKey struct {
	idx, id, hash int32
	jhash         int32
}

func (k *hvNKey) HashCode() int32 { return k.hash }

func (k *hvNKey) Equals_Object(o any) bool {
	x, ok := o.(*hvNKey)
	return ok && x.id == k.id
}

func (k *hvNKey) JClassName() string    { return "HashVectors$NKey" }
func (k *hvNKey) JIdentityHash() *int32 { return &k.jhash }
func (k *hvNKey) ToString() *String     { return FromGo(fmt.Sprintf("n%d", k.idx)) }

// hvOKey is HashVectors.OKey, which has Object's hashCode and equals.
type hvOKey struct {
	idx   int32
	jhash int32
}

func (k *hvOKey) JClassName() string    { return "HashVectors$OKey" }
func (k *hvOKey) JIdentityHash() *int32 { return &k.jhash }
func (k *hvOKey) ToString() *String     { return FromGo(fmt.Sprintf("o%d", k.idx)) }

// hvXKey is HashVectors.XKey, which is not equal even to itself.
type hvXKey struct {
	idx, hash int32
	jhash     int32
}

func (k *hvXKey) HashCode() int32          { return k.hash }
func (k *hvXKey) Equals_Object(o any) bool { return false }
func (k *hvXKey) JClassName() string       { return "HashVectors$XKey" }
func (k *hvXKey) JIdentityHash() *int32    { return &k.jhash }
func (k *hvXKey) ToString() *String        { return FromGo(fmt.Sprintf("x%d", k.idx)) }

// hvEntry is AbstractMap.SimpleEntry.
type hvEntry struct {
	key, value any
}

func (e *hvEntry) JIface_Map_Entry()         {}
func (e *hvEntry) JClassName() string        { return "java.util.AbstractMap$SimpleEntry" }
func (e *hvEntry) GetKey() any               { return e.key }
func (e *hvEntry) GetValue() any             { return e.value }
func (e *hvEntry) SetValue_Object(v any) any { old := e.value; e.value = v; return old }
func (e *hvEntry) Equals_Object(o any) bool  { return entryEquals(e.key, e.value, o) }
func (e *hvEntry) HashCode() int32           { return entryHashCode(e.key, e.value) }
func (e *hvEntry) ToString() *String         { return entryToString(e.key, e.value) }

type hvSeq struct {
	line   int
	header []string
	h0     int32
	keys   [][]string
	steps  []hvStep
	ih     [][2]int32 // pool index, identity hash
	ctorEx string
}

type hvStep struct {
	line int
	op   []string
	want string
}

func readHashVectors(t *testing.T) []*hvSeq {
	f, err := os.Open("testdata/hash.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var seqs []*hvSeq
	var cur *hvSeq
	sc := bufio.NewScanner(f)
	sc.Buffer(nil, 1<<20)
	for ln := 1; sc.Scan(); ln++ {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "seq "):
			cur = &hvSeq{line: ln, header: strings.Fields(line)[2:]}
			seqs = append(seqs, cur)
		case strings.HasPrefix(line, "= "):
			cur.ctorEx = line[2:]
		case strings.HasPrefix(line, "h0 "):
			cur.h0 = hvAtoi(t, line[3:])
		case strings.HasPrefix(line, "k "):
			cur.keys = append(cur.keys, strings.Split(line[2:], " "))
		case strings.HasPrefix(line, "ih"):
			for _, f := range strings.Fields(line)[1:] {
				i := strings.IndexByte(f, ':')
				cur.ih = append(cur.ih, [2]int32{hvAtoi(t, f[:i]), hvAtoi(t, f[i+1:])})
			}
		case line == "end":
			cur = nil
		default:
			i := strings.Index(line, " = ")
			if i < 0 {
				t.Fatalf("line %d: bad line %q", ln, line)
			}
			cur.steps = append(cur.steps, hvStep{line: ln, op: strings.Fields(line[:i]), want: line[i+3:]})
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return seqs
}

func hvAtoi(t *testing.T, s string) int32 {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		t.Fatal(err)
	}
	return int32(v)
}

func TestHashVectors(t *testing.T) {
	seqs := readHashVectors(t)
	if len(seqs) < 100 {
		t.Fatalf("only %d sequences", len(seqs))
	}
	steps, lazy := 0, 0
	for _, s := range seqs {
		if s.ctorEx != "" {
			hvCtorError(t, s)
			continue
		}
		steps += len(s.steps)
		if !hvRun(t, s, false) {
			continue
		}
		if len(s.ih) > 0 {
			lazy++
			hvRun(t, s, true)
		}
	}
	t.Logf("%d sequences, %d steps, %d sequences replayed with lazy identity hashes", len(seqs), steps, lazy)
}

func hvCtorError(t *testing.T, s *hvSeq) {
	t.Helper()
	got := hvCatch(func() string {
		r := &hvRunner{}
		r.construct(s.header)
		if r.m != nil {
			return className(r.m)
		}
		return className(r.s)
	})
	if got != s.ctorEx {
		t.Errorf("line %d: %v: got %s, want %s", s.line, s.header, got, s.ctorEx)
	}
}

type hvRunner struct {
	t    *testing.T
	pool []any
	m    Map
	s    Set
}

// hvRun replays s, reporting whether it matched.
func hvRun(t *testing.T, s *hvSeq, lazy bool) bool {
	t.Helper()
	r := &hvRunner{t: t}
	if lazy {
		SetIdentityHashSeq(s.h0 + 1)
	}
	for i, k := range s.keys {
		r.pool = append(r.pool, hvKey(t, int32(i), k))
	}
	if !lazy {
		for _, ih := range s.ih {
			*r.pool[ih[0]].(identityHashed).JIdentityHash() = ih[1]
		}
	}
	r.construct(s.header)
	mode := "pinned"
	if lazy {
		mode = "lazy"
	}
	for _, st := range s.steps {
		res := hvCatch(func() string { return r.do(st.op) })
		got := fmt.Sprintf("%s %d %s %s", res, r.size(), hvDigest(r.order()), hvDigest(r.state()))
		if got != st.want {
			t.Errorf("line %d (%s identity hashes): %s: got %s, want %s\norder: %s\nstate: %s",
				st.line, mode, strings.Join(st.op, " "), got, st.want, r.order(), r.state())
			return false
		}
	}
	if lazy {
		for _, ih := range s.ih {
			if got := IdentityHashCode(r.pool[ih[0]]); got != ih[1] {
				t.Errorf("line %d: identity hash of key %d: got %d, want %d", s.line, ih[0], got, ih[1])
				return false
			}
		}
	}
	return true
}

func hvKey(t *testing.T, idx int32, f []string) any {
	switch f[0] {
	case "z":
		return nil
	case "i":
		return hvAtoi(t, f[1])
	case "l":
		v, err := strconv.ParseInt(f[1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return v
	case "d":
		v, err := strconv.ParseUint(f[1], 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		return math.Float64frombits(v)
	case "b":
		return f[1] == "1"
	case "s":
		if len(f) < 2 {
			return FromGo("")
		}
		return FromGo(f[1])
	case "c", "t":
		return &hvCKey{idx: idx, id: hvAtoi(t, f[1]), hash: hvAtoi(t, f[2]), rank: hvAtoi(t, f[3]), sub: f[0] == "t"}
	case "n":
		return &hvNKey{idx: idx, id: hvAtoi(t, f[1]), hash: hvAtoi(t, f[2])}
	case "o":
		return &hvOKey{idx: idx}
	case "x":
		return &hvXKey{idx: idx, hash: hvAtoi(t, f[1])}
	}
	t.Fatalf("bad key %v", f)
	return nil
}

// hvCatch returns f's result, or the token for the Java exception it throws.
func hvCatch(f func() string) (res string) {
	defer func() {
		if r := recover(); r != nil {
			ex := Throwable_ofPanic(r)
			res = "EX:" + className(ex)
			if msg := Throwable_fromAny(ex).GetMessage(); msg != nil {
				res += ":" + strings.ReplaceAll(msg.Go(), " ", "_")
			}
		}
	}()
	return f()
}

func hvFloat(s string) float32 {
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		panic(err)
	}
	return math.Float32frombits(uint32(v))
}

func hvInt(s string) int32 {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		panic(err)
	}
	return int32(v)
}

func (r *hvRunner) key(s string) any { return r.pool[hvInt(s)] }

func hvValue(s string) any {
	if s == "null" {
		return nil
	}
	return hvInt(s)
}

func (r *hvRunner) construct(h []string) {
	container, ctor, args := h[0], h[1], h[2:]
	linked := strings.HasPrefix(container, "L")
	if container == "HS" || container == "LHS" {
		switch {
		case ctor == "new" && linked:
			r.s = LinkedHashSet_new()
		case ctor == "new":
			r.s = HashSet_new()
		case ctor == "new_int" && linked:
			r.s = LinkedHashSet_new_int(hvInt(args[0]))
		case ctor == "new_int":
			r.s = HashSet_new_int(hvInt(args[0]))
		case ctor == "new_int_float" && linked:
			r.s = LinkedHashSet_new_int_float(hvInt(args[0]), hvFloat(args[1]))
		case ctor == "new_int_float":
			r.s = HashSet_new_int_float(hvInt(args[0]), hvFloat(args[1]))
		case ctor == "new_coll" && linked:
			r.s = LinkedHashSet_new_Collection(r.keyList(args))
		case ctor == "new_coll":
			r.s = HashSet_new_Collection(r.keyList(args))
		default:
			panic("bad constructor " + ctor)
		}
		return
	}
	switch {
	case ctor == "new" && linked:
		r.m = LinkedHashMap_new()
	case ctor == "new":
		r.m = HashMap_new()
	case ctor == "new_int" && linked:
		r.m = LinkedHashMap_new_int(hvInt(args[0]))
	case ctor == "new_int":
		r.m = HashMap_new_int(hvInt(args[0]))
	case ctor == "new_int_float" && linked:
		r.m = LinkedHashMap_new_int_float(hvInt(args[0]), hvFloat(args[1]))
	case ctor == "new_int_float":
		r.m = HashMap_new_int_float(hvInt(args[0]), hvFloat(args[1]))
	case ctor == "new_int_float_boolean":
		r.m = LinkedHashMap_new_int_float_boolean(hvInt(args[0]), hvFloat(args[1]), args[2] == "true")
	case ctor == "new_map" && linked:
		r.m = LinkedHashMap_new_Map(r.sourceMap(args))
	case ctor == "new_map":
		r.m = HashMap_new_Map(r.sourceMap(args))
	default:
		panic("bad constructor " + ctor)
	}
}

// keyList is the ArrayList of keys described by "n k...".
func (r *hvRunner) keyList(args []string) *ArrayList {
	l := ArrayList_new()
	for _, a := range args[1 : 1+hvInt(args[0])] {
		l.Add_Object(r.key(a))
	}
	return l
}

func hvValueList(args []string) *ArrayList {
	l := ArrayList_new()
	for _, a := range args[1 : 1+hvInt(args[0])] {
		l.Add_Object(hvValue(a))
	}
	return l
}

func (r *hvRunner) entryList(args []string) *ArrayList {
	l := ArrayList_new()
	n := int(hvInt(args[0]))
	for i := 0; i < n; i++ {
		l.Add_Object(&hvEntry{r.key(args[1+2*i]), hvValue(args[2+2*i])})
	}
	return l
}

// sourceMap is the map described by "H|L n k v...".
func (r *hvRunner) sourceMap(args []string) Map {
	var m Map = HashMap_new()
	if args[0] == "L" {
		m = LinkedHashMap_new()
	}
	n := int(hvInt(args[1]))
	for i := 0; i < n; i++ {
		m.Put_Object_Object(r.key(args[2+2*i]), hvValue(args[3+2*i]))
	}
	return m
}

func hvV(v any) string {
	if v == nil {
		return "null"
	}
	return strconv.Itoa(int(v.(int32)))
}

func hvB(b bool) string { return strconv.FormatBool(b) }

func (r *hvRunner) do(op []string) string {
	if r.m != nil {
		return r.mapOp(op[0], op[1:])
	}
	return r.setOp(op[0], op[1:])
}

func (r *hvRunner) mapOp(name string, a []string) string {
	m := r.m
	hm := HashMap_fromAny(m)
	switch name {
	case "init":
		return "ok"
	case "put":
		return hvV(m.Put_Object_Object(r.key(a[0]), hvValue(a[1])))
	case "putIfAbsent":
		return hvV(m.PutIfAbsent_Object_Object(r.key(a[0]), hvValue(a[1])))
	case "get":
		return hvV(m.Get_Object(r.key(a[0])))
	case "getOrDefault":
		return hvV(m.GetOrDefault_Object_Object(r.key(a[0]), hvValue(a[1])))
	case "containsKey":
		return hvB(m.ContainsKey_Object(r.key(a[0])))
	case "containsValue":
		return hvB(m.ContainsValue_Object(hvValue(a[0])))
	case "remove":
		return hvV(m.Remove_Object(r.key(a[0])))
	case "remove2":
		return hvB(hm.Remove_Object_Object(r.key(a[0]), hvValue(a[1])))
	case "replace2":
		return hvV(hm.Replace_Object_Object(r.key(a[0]), hvValue(a[1])))
	case "replace3":
		return hvB(hm.Replace_Object_Object_Object(r.key(a[0]), hvValue(a[1]), hvValue(a[2])))
	case "clear":
		m.Clear()
		return "ok"
	case "putAll":
		m.PutAll_Map(r.sourceMap(a))
		return "ok"
	case "iterRemove":
		var it Iterator
		switch a[0] {
		case "K":
			it = m.KeySet().Iterator()
		case "V":
			it = m.Values().Iterator()
		default:
			it = m.EntrySet().Iterator()
		}
		return hvIterRemove(it, hvInt(a[1]), hvInt(a[2]))
	case "ksRemove":
		return hvB(m.KeySet().Remove_Object(r.key(a[0])))
	case "vsRemove":
		return hvB(m.Values().Remove_Object(hvValue(a[0])))
	case "esRemove":
		return hvB(m.EntrySet().Remove_Object(&hvEntry{r.key(a[0]), hvValue(a[1])}))
	case "esContains":
		return hvB(m.EntrySet().Contains_Object(&hvEntry{r.key(a[0]), hvValue(a[1])}))
	case "ksContains":
		return hvB(m.KeySet().Contains_Object(r.key(a[0])))
	case "vsContains":
		return hvB(m.Values().Contains_Object(hvValue(a[0])))
	case "ksRemoveAll":
		return hvB(m.KeySet().RemoveAll_Collection(r.keyList(a)))
	case "ksRetainAll":
		return hvB(m.KeySet().RetainAll_Collection(r.keyList(a)))
	case "vsRemoveAll":
		return hvB(m.Values().RemoveAll_Collection(hvValueList(a)))
	case "vsRetainAll":
		return hvB(m.Values().RetainAll_Collection(hvValueList(a)))
	case "esRemoveAll":
		return hvB(m.EntrySet().RemoveAll_Collection(r.entryList(a)))
	case "esSetValue":
		pos := hvInt(a[0])
		it := m.EntrySet().Iterator()
		for i := int32(0); i < pos && it.HasNext(); i++ {
			it.Next()
		}
		if !it.HasNext() {
			return "none"
		}
		return hvV(it.Next().(Map_Entry).SetValue_Object(hvValue(a[1])))
	case "clone":
		r.m = hm.Clone().(Map)
		return className(r.m)
	case "copy":
		if a[0] == "L" {
			r.m = LinkedHashMap_new_Map(m)
		} else {
			r.m = HashMap_new_Map(m)
		}
		return className(r.m)
	case "equalsCopy":
		c := HashMap_new_Map(m)
		x := Object_Equals_Object(m, c)
		y := Object_Equals_Object(c, m)
		ks := Object_Equals_Object(m.KeySet(), c.KeySet())
		es := Object_Equals_Object(m.EntrySet(), c.EntrySet())
		return hvB(x) + "/" + hvB(y) + "/" + hvB(ks) + "/" + hvB(es)
	case "equalsMod":
		c := LinkedHashMap_new_Map(m)
		c.Put_Object_Object(r.key(a[0]), hvValue(a[1]))
		return hvB(Object_Equals_Object(m, c)) + "/" + hvB(Object_Equals_Object(c, m))
	case "hashCode":
		return fmt.Sprintf("%d/%d/%d", Object_HashCode(m), Object_HashCode(m.KeySet()), Object_HashCode(m.EntrySet()))
	case "toString":
		return hvDigest(Object_ToString(m).Go()) + "/" + hvDigest(Object_ToString(m.KeySet()).Go()) + "/" +
			hvDigest(Object_ToString(m.Values()).Go()) + "/" + hvDigest(Object_ToString(m.EntrySet()).Go())
	case "toArray":
		var sb strings.Builder
		for _, o := range m.KeySet().ToArray() {
			sb.WriteString(hvTok(o) + ",")
		}
		sb.WriteString("|")
		for _, o := range m.Values().ToArray_ObjectArr(make([]any, hvInt(a[0]))).([]any) {
			if o == nil {
				sb.WriteString("_,")
			} else {
				sb.WriteString(hvV(o) + ",")
			}
		}
		sb.WriteString("|")
		for _, o := range m.EntrySet().ToArray() {
			e := o.(Map_Entry)
			sb.WriteString(hvTok(e.GetKey()) + "=" + hvV(e.GetValue()) + ",")
		}
		return hvDigest(sb.String())
	case "iterNSEE":
		it := m.Values().Iterator()
		for it.HasNext() {
			it.Next()
		}
		it.Next()
		return "none"
	case "iterISE":
		it := m.KeySet().Iterator()
		if a[0] == "true" && it.HasNext() {
			it.Next()
			it.Remove()
		}
		it.Remove()
		return "none"
	case "iterCME":
		it := m.EntrySet().Iterator()
		if it.HasNext() {
			it.Next()
		}
		put := hvV(m.Put_Object_Object(r.key(a[0]), hvValue(a[1])))
		if !it.HasNext() {
			return put + "/end"
		}
		return put + "/" + hvCatch(func() string { it.Next(); return "ok" })
	case "addView":
		m.KeySet().Add_Object(r.key(a[0]))
		return "none"
	case "ksContainsAll":
		return hvB(m.KeySet().ContainsAll_Collection(r.keyList(a)))
	case "mutate":
		return r.mutate(a)
	}
	panic("bad map op " + name)
}

// mutate changes the hash code of a key, which may be in the collection.
func (r *hvRunner) mutate(a []string) string {
	h := hvInt(a[1])
	switch k := r.key(a[0]).(type) {
	case *hvCKey:
		k.hash = h
	case *hvNKey:
		k.hash = h
	case *hvXKey:
		k.hash = h
	default:
		panic("mutate of an immutable key")
	}
	return "ok"
}

func hvIterRemove(it Iterator, mod, rem int32) string {
	removed := 0
	for i := int32(0); it.HasNext(); i++ {
		it.Next()
		if i%mod == rem {
			it.Remove()
			removed++
		}
	}
	return strconv.Itoa(removed)
}

func (r *hvRunner) setOp(name string, a []string) string {
	s := r.s
	switch name {
	case "init":
		return "ok"
	case "add":
		return hvB(s.Add_Object(r.key(a[0])))
	case "remove":
		return hvB(s.Remove_Object(r.key(a[0])))
	case "contains":
		return hvB(s.Contains_Object(r.key(a[0])))
	case "clear":
		s.Clear()
		return "ok"
	case "addAll":
		return hvB(s.AddAll_Collection(r.keyList(a)))
	case "removeAll":
		return hvB(s.RemoveAll_Collection(r.keyList(a)))
	case "retainAll":
		return hvB(s.RetainAll_Collection(r.keyList(a)))
	case "containsAll":
		return hvB(s.ContainsAll_Collection(r.keyList(a)))
	case "iterRemove":
		return hvIterRemove(s.Iterator(), hvInt(a[0]), hvInt(a[1]))
	case "clone":
		r.s = HashSet_fromAny(s).Clone().(Set)
		return className(r.s)
	case "copy":
		if a[0] == "L" {
			r.s = LinkedHashSet_new_Collection(s)
		} else {
			r.s = HashSet_new_Collection(s)
		}
		return className(r.s)
	case "equalsCopy":
		c := HashSet_new_Collection(s)
		return hvB(Object_Equals_Object(s, c)) + "/" + hvB(Object_Equals_Object(c, s))
	case "equalsMod":
		c := LinkedHashSet_new_Collection(s)
		c.Add_Object(r.key(a[0]))
		return hvB(Object_Equals_Object(s, c)) + "/" + hvB(Object_Equals_Object(c, s))
	case "hashCode":
		return strconv.Itoa(int(Object_HashCode(s)))
	case "toString":
		return hvDigest(Object_ToString(s).Go())
	case "toArray":
		var sb strings.Builder
		for _, o := range s.ToArray() {
			sb.WriteString(hvTok(o) + ",")
		}
		sb.WriteString("|")
		for _, o := range s.ToArray_ObjectArr(make([]any, hvInt(a[0]))).([]any) {
			if o == nil {
				sb.WriteString("_,")
			} else {
				sb.WriteString(hvTok(o) + ",")
			}
		}
		return hvDigest(sb.String())
	case "iterNSEE":
		it := s.Iterator()
		for it.HasNext() {
			it.Next()
		}
		it.Next()
		return "none"
	case "iterCME":
		it := s.Iterator()
		if it.HasNext() {
			it.Next()
		}
		add := hvB(s.Add_Object(r.key(a[0])))
		if !it.HasNext() {
			return add + "/end"
		}
		return add + "/" + hvCatch(func() string { it.Next(); return "ok" })
	case "mutate":
		return r.mutate(a)
	}
	panic("bad set op " + name)
}

func (r *hvRunner) size() int32 {
	if r.m != nil {
		return r.m.Size()
	}
	return r.s.Size()
}

func hvTok(o any) string {
	switch x := o.(type) {
	case nil:
		return "null"
	case int32:
		return "i" + strconv.Itoa(int(x))
	case int64:
		return "l" + strconv.FormatInt(x, 10)
	case float64:
		return "d" + strconv.FormatUint(math.Float64bits(x), 16)
	case bool:
		if x {
			return "b1"
		}
		return "b0"
	case *String:
		return "s" + x.Go()
	case *hvCKey, *hvNKey, *hvOKey, *hvXKey:
		return Object_ToString(x).Go()
	}
	panic(fmt.Sprintf("bad key %T", o))
}

// order is HashVectors.Seq.order.
func (r *hvRunner) order() string {
	var sb strings.Builder
	if r.m != nil {
		for it := r.m.EntrySet().Iterator(); it.HasNext(); {
			e := it.Next().(Map_Entry)
			sb.WriteString(hvTok(e.GetKey()) + "=" + hvV(e.GetValue()) + ",")
		}
		return sb.String()
	}
	for it := r.s.Iterator(); it.HasNext(); {
		sb.WriteString(hvTok(it.Next()) + ",")
	}
	return sb.String()
}

// state is HashVectors.mapState for the collection's map.
func (r *hvRunner) state() string {
	var m *HashMap
	if r.m != nil {
		m = HashMap_fromAny(r.m)
	} else {
		m = HashSet_fromAny(r.s).m
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "c%dt%ds%d;", len(m.table), m.threshold, m.size)
	for i, first := range m.table {
		if first == nil {
			continue
		}
		fmt.Fprintf(&sb, "%d:", i)
		tree := first.kind == nodeTree
		prevOK := true
		var prev *treeNode
		for e := first; e != nil; e = e.next {
			sb.WriteString(hvTok(e.key) + ",")
			if tree {
				if asTree(e).prev != prev {
					prevOK = false
				}
				prev = asTree(e)
			}
		}
		if tree {
			sb.WriteString("T")
			hvTree(&sb, asTree(first).root())
			if !prevOK {
				sb.WriteString("P!")
			}
		}
		sb.WriteString(";")
	}
	if m.linked {
		var before *hashNode
		ok := true
		for e := m.head; e != nil; e = e.links().after {
			if e.links().before != before {
				ok = false
			}
			before = e
		}
		if m.tail != before {
			ok = false
		}
		if !ok {
			sb.WriteString("L!")
		}
	}
	return sb.String()
}

func hvTree(sb *strings.Builder, n *treeNode) {
	if n == nil {
		sb.WriteString(".")
		return
	}
	sb.WriteString("(" + hvTok(n.key))
	if n.red {
		sb.WriteString("r")
	} else {
		sb.WriteString("b")
	}
	hvTree(sb, n.left)
	hvTree(sb, n.right)
	sb.WriteString(")")
}

// hvDigest is HashVectors.digest, the 32-bit FNV-1a hash of s.
func hvDigest(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprintf("%08x", h.Sum32())
}
