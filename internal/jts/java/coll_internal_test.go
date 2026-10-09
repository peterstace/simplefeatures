package java

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unsafe"
)

// This file replays the vectors in testdata/coll_*.txt, which
// internal/jts/tools/jdkvectors/CollVectors.java generates from the JDK, and
// requires every result, every intermediate state and every sequence of
// comparisons to match. See the Java program for the format.

// ---------------------------------------------------------------------------
// Test element types, mirroring the Java program's.

// collCmpLog records the comparisons made through collItem.compareTo and the
// test comparators.
type collCmpLog struct {
	count int32
	hash  int32
}

func (l *collCmpLog) add(a, b int32) {
	l.count++
	l.hash = l.hash*1000003 + (a*1009 + b)
}

// collItem is the Java program's Item: compared by key only, equal when key
// and id are.
type collItem struct {
	log *collCmpLog
	key int32
	id  int32
}

func (it *collItem) CompareTo_Object(o any) int32 {
	if o == nil {
		panic(NullPointerException_new())
	}
	other, ok := o.(*collItem)
	if !ok {
		panic(ClassCastException_new())
	}
	it.log.add(it.key, other.key)
	return Integer_compare(it.key, other.key)
}

func (it *collItem) Equals_Object(o any) bool {
	other, ok := o.(*collItem)
	return ok && other.key == it.key && other.id == it.id
}

func (it *collItem) HashCode() int32    { return it.key*31 + it.id }
func (it *collItem) ToString() *String  { return FromGo(fmt.Sprintf("k%d.%d", it.key, it.id)) }
func (it *collItem) JClassName() string { return "CollVectors$Item" }

// collOrd is the Java program's enum Ord.
type collOrd struct {
	Enum
}

func (o *collOrd) JClassName() string { return "CollVectors$Ord" }

func newCollOrds() []*collOrd {
	var ords []*collOrd
	for i, name := range []string{"X", "Y", "Z", "M", "A", "B"} {
		o := &collOrd{}
		o.JSetSelf(o)
		ords = append(ords, EnumConst(o, name, int32(i)))
	}
	return ords
}

func collKeyOf(o any) int32 {
	switch x := o.(type) {
	case nil:
		panic(NullPointerException_new())
	case int32:
		return x
	case *collItem:
		return x.key
	case *String:
		return x.HashCode()
	case *collOrd:
		return x.Ordinal()
	}
	panic(ClassCastException_new())
}

type collModCmp struct{ log *collCmpLog }

func (c *collModCmp) Compare_Object_Object(a, b any) int32 {
	x, y := collKeyOf(a), collKeyOf(b)
	c.log.add(x, y)
	fm := func(v int32) int32 { return ((v % 3) + 3) % 3 }
	return Integer_compare(fm(x), fm(y))
}

type collChaosCmp struct{ log *collCmpLog }

func (c *collChaosCmp) Compare_Object_Object(a, b any) int32 {
	x, y := collKeyOf(a), collKeyOf(b)
	c.log.add(x, y)
	h := x*int32(-1640531535) + y*int32(-2048144789)
	h ^= int32(uint32(h) >> 15)
	h *= 0x2C1B3C6D
	h ^= int32(uint32(h) >> 12)
	return (h & 3) - 1
}

type collLogNatCmp struct{ log *collCmpLog }

func (c *collLogNatCmp) Compare_Object_Object(a, b any) int32 {
	c.log.add(collKeyOf(a), collKeyOf(b))
	return Comparable_CompareTo_Object(a, b)
}

// ---------------------------------------------------------------------------
// The replayer

type collTester struct {
	t     *testing.T
	file  string
	line  int
	log   *collCmpLog
	mod   *collModCmp
	ords  []*collOrd
	state any
	held  []any
	// lastState is the rendered state after the previous operation, which
	// the vectors abbreviate to "=" when it is unchanged.
	lastState string
}

func newCollTester(t *testing.T, file string) *collTester {
	t.Helper()
	log := &collCmpLog{}
	return &collTester{t: t, file: file, log: log, mod: &collModCmp{log: log}, ords: newCollOrds()}
}

func (ct *collTester) cmp(name string) Comparator {
	switch name {
	case "nat":
		return nil
	case "rev":
		return Collections_reverseOrder()
	case "mod":
		return ct.mod
	case "chaos":
		return &collChaosCmp{log: ct.log}
	case "lognat":
		return &collLogNatCmp{log: ct.log}
	case "revmod":
		return Collections_reverseOrder_Comparator(ct.mod)
	}
	ct.t.Fatalf("%s:%d: unknown comparator %q", ct.file, ct.line, name)
	return nil
}

// r renders a value as the Java program's r does.
func (ct *collTester) r(o any) string {
	switch x := o.(type) {
	case nil:
		return "null"
	case int32:
		return "i" + strconv.Itoa(int(x))
	case *String:
		return "s" + x.Go()
	case *collItem:
		return fmt.Sprintf("k%d.%d", x.key, x.id)
	case bool:
		return strconv.FormatBool(x)
	case *collOrd:
		return "e" + x.Name().Go()
	case Map_Entry:
		return "<" + ct.r(x.GetKey()) + "=" + ct.r(x.GetValue()) + ">"
	case Map:
		var parts []string
		for it := x.EntrySet().Iterator(); it.HasNext(); {
			parts = append(parts, ct.r(it.Next()))
		}
		return "{" + strings.Join(parts, ",") + "}"
	case Collection:
		var parts []string
		for it := x.Iterator(); it.HasNext(); {
			parts = append(parts, ct.r(it.Next()))
		}
		return "[" + strings.Join(parts, ",") + "]"
	}
	v := reflect.ValueOf(o)
	if v.Kind() == reflect.Slice {
		var parts []string
		for i := 0; i < v.Len(); i++ {
			parts = append(parts, ct.r(arrayGet(v, i)))
		}
		return "A[" + strings.Join(parts, ",") + "]"
	}
	ct.t.Fatalf("%s:%d: cannot render %T", ct.file, ct.line, o)
	return ""
}

func (ct *collTester) parse(s string) any {
	if s == "null" {
		return nil
	}
	switch s[0] {
	case 'i':
		v, err := strconv.ParseInt(s[1:], 10, 32)
		if err != nil {
			ct.t.Fatalf("%s:%d: bad int %q", ct.file, ct.line, s)
		}
		return int32(v)
	case 's':
		return FromGo(s[1:])
	case 'e':
		for _, o := range ct.ords {
			if o.Name().Go() == s[1:] {
				return o
			}
		}
	case 'k':
		var k, id int32
		if _, err := fmt.Sscanf(s[1:], "%d.%d", &k, &id); err == nil {
			return &collItem{log: ct.log, key: k, id: id}
		}
	}
	ct.t.Fatalf("%s:%d: bad value %q", ct.file, ct.line, s)
	return nil
}

// listLit splits "<tag>[a,b]" into the tag and the parsed elements.
func (ct *collTester) listLit(s string) (string, []any) {
	open := strings.IndexByte(s, '[')
	if open < 0 || s[len(s)-1] != ']' {
		ct.t.Fatalf("%s:%d: bad literal %q", ct.file, ct.line, s)
	}
	tag, body := s[:open], s[open+1:len(s)-1]
	var elems []any
	if body != "" {
		for _, e := range strings.Split(body, ",") {
			elems = append(elems, ct.parse(e))
		}
	}
	return tag, elems
}

// quietly runs f without recording its comparisons.
func (ct *collTester) quietly(f func()) {
	saved := *ct.log
	f()
	*ct.log = saved
}

func (ct *collTester) coll(s string) Collection {
	var c Collection
	ct.quietly(func() {
		tag, elems := ct.listLit(s)
		switch tag {
		case "AL":
			l := ArrayList_new()
			for _, e := range elems {
				l.Add_Object(e)
			}
			c = l
		case "LL":
			l := LinkedList_new()
			for _, e := range elems {
				l.Add_Object(e)
			}
			c = l
		case "TS":
			ts := TreeSet_new()
			for _, e := range elems {
				ts.Add_Object(e)
			}
			c = ts
		case "TSmod":
			ts := TreeSet_new_Comparator(ct.mod)
			for _, e := range elems {
				ts.Add_Object(e)
			}
			c = ts
		case "EM":
			es := EnumSet_noneOf(&Class{name: "CollVectors$Ord"})
			for _, e := range elems {
				if e != nil {
					es.Add_Object(e)
				}
			}
			c = es
		default:
			ct.t.Fatalf("%s:%d: bad collection literal %q", ct.file, ct.line, s)
		}
	})
	return c
}

func (ct *collTester) objArray(s string) []any {
	tag, elems := ct.listLit(s)
	if tag != "A" {
		ct.t.Fatalf("%s:%d: bad array literal %q", ct.file, ct.line, s)
	}
	if elems == nil {
		elems = []any{}
	}
	return elems
}

func (ct *collTester) mapLit(s string) Map {
	var m Map
	ct.quietly(func() {
		open := strings.IndexByte(s, '[')
		tag, body := s[:open], s[open+1:len(s)-1]
		var tm *TreeMap
		if tag == "TMmod" {
			tm = TreeMap_new_Comparator(ct.mod)
		} else {
			tm = TreeMap_new()
		}
		if body != "" {
			for _, kv := range strings.Split(body, ",") {
				p := strings.SplitN(kv, "=", 2)
				tm.Put_Object_Object(ct.parse(p[0]), ct.parse(p[1]))
			}
		}
		m = tm
		if tag == "UM" {
			m = Collections_unmodifiableMap(tm)
		}
	})
	return m
}

func collSimpleName(ex any) string {
	n := className(ex)
	if i := strings.LastIndexAny(n, ".$"); i >= 0 {
		n = n[i+1:]
	}
	return n
}

// try runs f, rendering a Java exception it throws as "!<SimpleName>".
func (ct *collTester) try(f func() string) (res string) {
	defer func() {
		if r := recover(); r != nil {
			if !Throwable_isInstance(r) {
				if _, ok := r.(interface{ RuntimeError() }); !ok {
					ct.t.Fatalf("%s:%d: panic %v", ct.file, ct.line, r)
				}
			}
			res = "!" + collSimpleName(Throwable_ofPanic(r))
		}
	}()
	return f()
}

func atoi32(s string) int32 {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		panic(err)
	}
	return int32(v)
}

func bstr(b bool) string { return strconv.FormatBool(b) }

func istr(i int32) string { return strconv.Itoa(int(i)) }

func (ct *collTester) toStr(o any) string {
	return strings.ReplaceAll(Object_ToString(o).Go(), " ", "_")
}

// iterProg runs an iterator program; owner is the collection or map that
// external modifications ('x' commands) apply to.
func (ct *collTester) iterProg(it Iterator, prog string, owner any) string {
	var out []string
	var last any
	for _, c := range strings.Split(prog, "/") {
		out = append(out, ct.try(func() string {
			rest := c[1:]
			switch c[0] {
			case 'h':
				return bstr(it.HasNext())
			case 'n':
				last = it.Next()
				return ct.r(last)
			case 'r':
				it.Remove()
				return "ok"
			case 'p':
				last = it.(ListIterator).Previous()
				return ct.r(last)
			case 'H':
				return bstr(it.(ListIterator).HasPrevious())
			case 'i':
				return istr(it.(ListIterator).NextIndex())
			case 'I':
				return istr(it.(ListIterator).PreviousIndex())
			case 'a':
				it.(ListIterator).Add_Object(ct.parse(rest))
				return "ok"
			case 's':
				it.(ListIterator).Set_Object(ct.parse(rest))
				return "ok"
			case 'v':
				return ct.r(last.(Map_Entry).SetValue_Object(ct.parse(rest)))
			case 'x':
				return ct.external(owner, rest)
			}
			ct.t.Fatalf("%s:%d: bad iterator command %q", ct.file, ct.line, c)
			return ""
		}))
	}
	return strings.Join(out, "/")
}

func (ct *collTester) external(owner any, c string) string {
	if m, ok := owner.(Map); ok {
		if c[0] == 'p' {
			kv := strings.SplitN(c[1:], "=", 2)
			return ct.r(m.Put_Object_Object(ct.parse(kv[0]), ct.parse(kv[1])))
		}
		return ct.r(m.Remove_Object(ct.parse(c[1:])))
	}
	coll := owner.(Collection)
	switch c[0] {
	case 'a':
		return bstr(coll.Add_Object(ct.parse(c[1:])))
	case 'r':
		return bstr(coll.Remove_Object(ct.parse(c[1:])))
	case 'c':
		coll.Clear()
		return "ok"
	}
	ct.t.Fatalf("%s:%d: bad external command %q", ct.file, ct.line, c)
	return ""
}

// collectionOp runs an operation of the Java program's collectionOp.
func (ct *collTester) collectionOp(c Collection, op string, args []string) (string, bool) {
	switch op {
	case "add":
		return bstr(c.Add_Object(ct.parse(args[0]))), true
	case "remove":
		return bstr(c.Remove_Object(ct.parse(args[0]))), true
	case "contains":
		return bstr(c.Contains_Object(ct.parse(args[0]))), true
	case "size":
		return istr(c.Size()), true
	case "isEmpty":
		return bstr(c.IsEmpty()), true
	case "addAll":
		return bstr(c.AddAll_Collection(ct.coll(args[0]))), true
	case "removeAll":
		return bstr(c.RemoveAll_Collection(ct.coll(args[0]))), true
	case "retainAll":
		return bstr(c.RetainAll_Collection(ct.coll(args[0]))), true
	case "containsAll":
		return bstr(c.ContainsAll_Collection(ct.coll(args[0]))), true
	case "toArray":
		return ct.r(c.ToArray()), true
	case "toArrayT":
		n := atoi32(args[0])
		arr := make([]any, n)
		for i := range arr {
			arr[i] = int32(-1)
		}
		res := c.ToArray_ObjectArr(arr)
		rv := reflect.ValueOf(res)
		prefix := "new:"
		if rv.Pointer() == reflect.ValueOf(arr).Pointer() && rv.Len() == len(arr) {
			prefix = "same:"
		}
		return prefix + ct.r(res), true
	case "iter":
		return ct.iterProg(c.Iterator(), args[0], c), true
	case "toString":
		return ct.toStr(c), true
	case "hashCode":
		return istr(Object_HashCode(c)), true
	case "clear":
		c.Clear()
		return "ok", true
	case "equals":
		return bstr(Object_Equals_Object(c, ct.coll(args[0]))), true
	case "cme":
		e := ct.parse(args[0])
		it := c.Iterator()
		out := []string{
			ct.try(func() string { return ct.r(it.Next()) }),
			ct.try(func() string { return bstr(c.Add_Object(e)) }),
		}
		var hasNext string
		next := ct.try(func() string {
			hasNext = bstr(it.HasNext())
			return ct.r(it.Next())
		})
		if hasNext != "" {
			out = append(out, hasNext)
		}
		return strings.Join(append(out, next), "/"), true
	}
	return "", false
}

func (ct *collTester) listOp(l List, op string, args []string) (string, bool) {
	switch op {
	case "addAt":
		l.Add_int_Object(atoi32(args[0]), ct.parse(args[1]))
		return "ok", true
	case "get":
		return ct.r(l.Get_int(atoi32(args[0]))), true
	case "set":
		return ct.r(l.Set_int_Object(atoi32(args[0]), ct.parse(args[1]))), true
	case "removeAt":
		return ct.r(l.Remove_int(atoi32(args[0]))), true
	case "indexOf":
		return istr(l.IndexOf_Object(ct.parse(args[0]))), true
	case "lastIndexOf":
		return istr(l.LastIndexOf_Object(ct.parse(args[0]))), true
	case "addAllAt":
		return bstr(l.AddAll_int_Collection(atoi32(args[0]), ct.coll(args[1]))), true
	case "sort":
		l.Sort_Comparator(ct.cmp(args[0]))
		return "ok", true
	case "lit":
		i := atoi32(args[0])
		var it ListIterator
		if i < 0 {
			it = l.ListIterator()
		} else {
			it = l.ListIterator_int(i)
		}
		return ct.iterProg(it, args[1], l), true
	case "sub":
		return ct.subProg(l, atoi32(args[0]), atoi32(args[1]), args[2]), true
	case "ensureCME":
		al := l.(*ArrayList)
		n := atoi32(args[0])
		it := al.Iterator()
		if n < 0 {
			al.TrimToSize()
		} else {
			al.EnsureCapacity_int(n)
		}
		return ct.try(func() string {
			it.Next()
			return "ok"
		}), true
	case "clone":
		c, ok := l.(interface{ Clone() any })
		if !ok {
			return "noclone", true
		}
		res := c.Clone()
		return collSimpleName(res) + ":" + ct.r(res), true
	}
	return "", false
}

func (ct *collTester) subProg(l List, f, t int32, prog string) string {
	s := l.SubList_int_int(f, t)
	var out []string
	for _, c := range strings.Split(prog, "/") {
		out = append(out, ct.try(func() string {
			rest := c[1:]
			switch c[0] {
			case 'a':
				return bstr(s.Add_Object(ct.parse(rest)))
			case 'A':
				p := strings.SplitN(rest, ",", 2)
				s.Add_int_Object(atoi32(p[0]), ct.parse(p[1]))
				return "ok"
			case 'r':
				return ct.r(s.Remove_int(atoi32(rest)))
			case 'R':
				return bstr(s.Remove_Object(ct.parse(rest)))
			case 'g':
				return ct.r(s.Get_int(atoi32(rest)))
			case 's':
				p := strings.SplitN(rest, ",", 2)
				return ct.r(s.Set_int_Object(atoi32(p[0]), ct.parse(p[1])))
			case 'c':
				s.Clear()
				return "ok"
			case 'z':
				return istr(s.Size())
			case 't':
				return ct.r(s)
			case 'o':
				s.Sort_Comparator(nil)
				return "ok"
			case 'i':
				return istr(s.IndexOf_Object(ct.parse(rest)))
			case 'l':
				return istr(s.LastIndexOf_Object(ct.parse(rest)))
			case 'h':
				return istr(Object_HashCode(s))
			case 'n':
				p := strings.SplitN(rest, ",", 2)
				s = s.SubList_int_int(atoi32(p[0]), atoi32(p[1]))
				return "ok"
			case 'x':
				if rest[0] == 'a' {
					return bstr(l.Add_Object(ct.parse(rest[1:])))
				}
				return ct.r(l.Remove_int(atoi32(rest[1:])))
			case 'U':
				return bstr(s.AddAll_Collection(ct.coll(rest)))
			case 'V':
				p := strings.SplitN(rest, ";", 2)
				return bstr(s.AddAll_int_Collection(atoi32(p[0]), ct.coll(p[1])))
			case 'D':
				return bstr(s.RemoveAll_Collection(ct.coll(rest)))
			case 'K':
				return bstr(s.RetainAll_Collection(ct.coll(rest)))
			case 'E':
				return bstr(Object_Equals_Object(s, ct.coll(rest)))
			case 'C':
				return bstr(s.Contains_Object(ct.parse(rest)))
			case 'Y':
				return bstr(s.ContainsAll_Collection(ct.coll(rest)))
			case 'T':
				return ct.toStr(s)
			case 'L':
				p := strings.SplitN(rest, ":", 2)
				it := s.ListIterator_int(atoi32(p[0]))
				return strings.ReplaceAll(ct.iterProg(it, strings.ReplaceAll(p[1], "~", "/"), s), "/", "~")
			case 'Q':
				arr := make([]any, atoi32(rest))
				for i := range arr {
					arr[i] = int32(-1)
				}
				res := s.ToArray_ObjectArr(arr)
				prefix := "new:"
				if rv := reflect.ValueOf(res); rv.Pointer() == reflect.ValueOf(arr).Pointer() && rv.Len() == len(arr) {
					prefix = "same:"
				}
				return prefix + ct.r(res)
			case 'w':
				var w []string
				it := s.Iterator()
				for k := 0; it.HasNext(); k++ {
					w = append(w, ct.r(it.Next()))
					if k%2 == 1 {
						it.Remove()
					}
				}
				return strings.Join(w, ",")
			}
			ct.t.Fatalf("%s:%d: bad sublist command %q", ct.file, ct.line, c)
			return ""
		}))
	}
	return strings.Join(out, "/")
}

func (ct *collTester) dequeOp(d Deque, op string, args []string) (string, bool) {
	switch op {
	case "addFirst":
		d.AddFirst_Object(ct.parse(args[0]))
		return "ok", true
	case "addLast":
		d.AddLast_Object(ct.parse(args[0]))
		return "ok", true
	case "offerFirst":
		return bstr(d.OfferFirst_Object(ct.parse(args[0]))), true
	case "offerLast":
		return bstr(d.OfferLast_Object(ct.parse(args[0]))), true
	case "removeFirst":
		return ct.r(d.RemoveFirst()), true
	case "removeLast":
		return ct.r(d.RemoveLast()), true
	case "pollFirst":
		return ct.r(d.PollFirst()), true
	case "pollLast":
		return ct.r(d.PollLast()), true
	case "getFirst":
		return ct.r(d.GetFirst()), true
	case "getLast":
		return ct.r(d.GetLast()), true
	case "peekFirst":
		return ct.r(d.PeekFirst()), true
	case "peekLast":
		return ct.r(d.PeekLast()), true
	case "push":
		d.Push_Object(ct.parse(args[0]))
		return "ok", true
	case "pop":
		return ct.r(d.Pop()), true
	case "peek":
		return ct.r(d.Peek()), true
	case "poll":
		return ct.r(d.Poll()), true
	case "element":
		return ct.r(d.Element()), true
	case "offer":
		return bstr(d.Offer_Object(ct.parse(args[0]))), true
	case "removeHead":
		return ct.r(d.Remove()), true
	case "removeFirstOcc":
		return bstr(d.RemoveFirstOccurrence_Object(ct.parse(args[0]))), true
	case "removeLastOcc":
		return bstr(d.RemoveLastOccurrence_Object(ct.parse(args[0]))), true
	case "desc":
		return ct.iterProg(d.DescendingIterator(), args[0], d), true
	case "clone":
		return "ArrayDeque:" + ct.r(d.(*ArrayDeque).Clone()), true
	}
	return "", false
}

func (ct *collTester) vectorOp(v *Vector, op string, args []string) (string, bool) {
	switch op {
	case "addElement":
		v.AddElement_Object(ct.parse(args[0]))
		return "ok", true
	case "elementAt":
		return ct.r(v.ElementAt_int(atoi32(args[0]))), true
	case "firstElement":
		return ct.r(v.FirstElement()), true
	case "lastElement":
		return ct.r(v.LastElement()), true
	case "insertElementAt":
		v.InsertElementAt_Object_int(ct.parse(args[0]), atoi32(args[1]))
		return "ok", true
	case "removeElementAt":
		v.RemoveElementAt_int(atoi32(args[0]))
		return "ok", true
	case "setElementAt":
		v.SetElementAt_Object_int(ct.parse(args[0]), atoi32(args[1]))
		return "ok", true
	case "removeElement":
		return bstr(v.RemoveElement_Object(ct.parse(args[0]))), true
	case "setSize":
		v.SetSize_int(atoi32(args[0]))
		return "ok", true
	case "indexOfFrom":
		return istr(v.IndexOf_Object_int(ct.parse(args[0]), atoi32(args[1]))), true
	case "lastIndexOfFrom":
		return istr(v.LastIndexOf_Object_int(ct.parse(args[0]), atoi32(args[1]))), true
	case "capacity":
		return istr(v.Capacity()), true
	case "ensureCapacity":
		v.EnsureCapacity_int(atoi32(args[0]))
		return istr(v.Capacity()), true
	case "trimToSize":
		v.TrimToSize()
		return istr(v.Capacity()), true
	case "removeAllElements":
		v.RemoveAllElements()
		return "ok", true
	case "copyInto":
		arr := make([]any, atoi32(args[0]))
		v.CopyInto_ObjectArr(arr)
		return ct.r(arr), true
	}
	st, ok := ct.state.(*Stack)
	if !ok {
		return "", false
	}
	switch op {
	case "stackPush":
		return ct.r(st.Push_Object(ct.parse(args[0]))), true
	case "stackPop":
		return ct.r(st.Pop()), true
	case "stackPeek":
		return ct.r(st.Peek()), true
	case "empty":
		return bstr(st.Empty()), true
	case "search":
		return istr(st.Search_Object(ct.parse(args[0]))), true
	}
	return "", false
}

func (ct *collTester) pqOp(q *PriorityQueue, op string, args []string) (string, bool) {
	switch op {
	case "offer":
		return bstr(q.Offer_Object(ct.parse(args[0]))), true
	case "poll":
		return ct.r(q.Poll()), true
	case "peek":
		return ct.r(q.Peek()), true
	case "element":
		return ct.r(q.Element()), true
	case "removeHead":
		return ct.r(q.Remove()), true
	case "drain":
		return ct.iterProg(q.Iterator(), args[0], q), true
	}
	return "", false
}

func (ct *collTester) navSetOp(s NavigableSet, op string, args []string) (string, bool) {
	switch op {
	case "first":
		return ct.r(s.First()), true
	case "last":
		return ct.r(s.Last()), true
	case "lower":
		return ct.r(s.Lower_Object(ct.parse(args[0]))), true
	case "floor":
		return ct.r(s.Floor_Object(ct.parse(args[0]))), true
	case "ceiling":
		return ct.r(s.Ceiling_Object(ct.parse(args[0]))), true
	case "higher":
		return ct.r(s.Higher_Object(ct.parse(args[0]))), true
	case "pollFirst":
		return ct.r(s.PollFirst()), true
	case "pollLast":
		return ct.r(s.PollLast()), true
	case "desc":
		return ct.iterProg(s.DescendingIterator(), args[0], s), true
	case "clone":
		return "TreeSet:" + ct.r(s.(*TreeSet).Clone()), true
	}
	return "", false
}

func (ct *collTester) mapOp(m Map, op string, args []string) (string, bool) {
	p := func(i int) any { return ct.parse(args[i]) }
	switch op {
	case "put":
		return ct.r(m.Put_Object_Object(p(0), p(1))), true
	case "get":
		return ct.r(m.Get_Object(p(0))), true
	case "mremove":
		return ct.r(m.Remove_Object(p(0))), true
	case "containsKey":
		return bstr(m.ContainsKey_Object(p(0))), true
	case "containsValue":
		return bstr(m.ContainsValue_Object(p(0))), true
	case "putIfAbsent":
		return ct.r(m.PutIfAbsent_Object_Object(p(0), p(1))), true
	case "getOrDefault":
		return ct.r(m.GetOrDefault_Object_Object(p(0), p(1))), true
	case "msize":
		return istr(m.Size()), true
	case "mclear":
		m.Clear()
		return "ok", true
	case "putAll":
		m.PutAll_Map(ct.mapLit(args[0]))
		return "ok", true
	case "entries":
		return ct.iterProg(m.EntrySet().Iterator(), args[0], m), true
	case "mequals":
		return bstr(Object_Equals_Object(m, ct.mapLit(args[0]))), true
	case "mhash":
		return istr(Object_HashCode(m)), true
	case "mtostring":
		return ct.toStr(m), true
	case "mcme":
		k, v := p(0), p(1)
		it := m.EntrySet().Iterator()
		out := []string{
			ct.try(func() string { return ct.r(it.Next()) }),
			ct.try(func() string { return ct.r(m.Put_Object_Object(k, v)) }),
		}
		var hasNext string
		next := ct.try(func() string {
			hasNext = bstr(it.HasNext())
			return ct.r(it.Next())
		})
		if hasNext != "" {
			out = append(out, hasNext)
		}
		return strings.Join(append(out, next), "/"), true
	case "esops":
		return ct.esOps(m, p(0), p(1), ct.mapLit(args[2])), true
	case "misEmpty":
		return bstr(m.IsEmpty()), true
	}
	tm, ok := m.(*TreeMap)
	if !ok {
		return "", false
	}
	switch op {
	case "firstKey":
		return ct.r(tm.FirstKey()), true
	case "lastKey":
		return ct.r(tm.LastKey()), true
	case "firstEntry":
		return ct.r(tm.FirstEntry()), true
	case "lastEntry":
		return ct.r(tm.LastEntry()), true
	case "pollFirstEntry":
		return ct.r(tm.PollFirstEntry()), true
	case "pollLastEntry":
		return ct.r(tm.PollLastEntry()), true
	case "lowerKey":
		return ct.r(tm.LowerKey_Object(p(0))), true
	case "floorKey":
		return ct.r(tm.FloorKey_Object(p(0))), true
	case "ceilingKey":
		return ct.r(tm.CeilingKey_Object(p(0))), true
	case "higherKey":
		return ct.r(tm.HigherKey_Object(p(0))), true
	case "lowerEntry":
		return ct.r(tm.LowerEntry_Object(p(0))), true
	case "floorEntry":
		return ct.r(tm.FloorEntry_Object(p(0))), true
	case "ceilingEntry":
		return ct.r(tm.CeilingEntry_Object(p(0))), true
	case "higherEntry":
		return ct.r(tm.HigherEntry_Object(p(0))), true
	case "replace3":
		return bstr(tm.Replace_Object_Object_Object(p(0), p(1), p(2))), true
	case "replace":
		return ct.r(tm.Replace_Object_Object(p(0), p(1))), true
	case "remove2":
		return bstr(tm.Remove_Object_Object(p(0), p(1))), true
	case "keys":
		return ct.iterProg(tm.KeySet().Iterator(), args[0], tm), true
	case "values":
		return ct.iterProg(tm.Values().Iterator(), args[0], tm), true
	case "desckeys":
		return ct.iterProg(tm.NavigableKeySet().DescendingIterator(), args[0], tm), true
	case "escontains":
		return bstr(tm.EntrySet().Contains_Object(&treeImmutableEntry{key: p(0), value: p(1)})), true
	case "esremove":
		return bstr(tm.EntrySet().Remove_Object(&treeImmutableEntry{key: p(0), value: p(1)})), true
	case "hold":
		ct.held = tm.EntrySet().ToArray()
		return istr(int32(len(ct.held))), true
	case "held":
		var parts []string
		for _, e := range ct.held {
			parts = append(parts, ct.r(e))
		}
		return strings.Join(parts, ","), true
	case "mclone":
		return "TreeMap:" + ct.r(tm.Clone()), true
	case "exportedEntry":
		var e Map_Entry
		if len(args) == 0 {
			e = tm.FirstEntry()
		} else {
			e = tm.CeilingEntry_Object(p(0))
		}
		if e == nil {
			return "null", true
		}
		res := ct.toStr(e) + ";" + istr(Object_HashCode(e)) + ";" +
			bstr(Object_Equals_Object(e, &treeImmutableEntry{key: e.GetKey(), value: e.GetValue()}))
		if ex := ct.try(func() string { e.SetValue_Object(int32(1)); return "" }); ex != "" {
			res += ";" + ex
		}
		return res, true
	}
	return "", false
}

// esOps mirrors the Java program's esOps.
func (ct *collTester) esOps(m Map, k, val any, other Map) string {
	es := m.EntrySet()
	out := []string{ct.r(es.ToArray()), ct.r(es.ToArray_ObjectArr(make([]any, 1)))}
	out = append(out, ct.try(func() string {
		return bstr(es.Contains_Object(&treeImmutableEntry{key: k, value: val}))
	}))
	out = append(out, ct.try(func() string { return bstr(es.ContainsAll_Collection(other.EntrySet())) }))
	out = append(out, bstr(Object_Equals_Object(es, other.EntrySet())))
	out = append(out, istr(Object_HashCode(es)))
	out = append(out, ct.toStr(es))
	if it := es.Iterator(); it.HasNext() {
		e := it.Next().(Map_Entry)
		out = append(out, ct.toStr(e), istr(Object_HashCode(e)),
			bstr(Object_Equals_Object(e, &treeImmutableEntry{key: e.GetKey(), value: e.GetValue()})),
			bstr(Object_Equals_Object(e, &treeImmutableEntry{key: e.GetKey(), value: val})))
	}
	ks := m.KeySet()
	out = append(out, ct.r(ks)+";"+ct.toStr(ks)+";"+istr(Object_HashCode(ks)))
	out = append(out, ct.try(func() string { return bstr(ks.Contains_Object(k)) }))
	out = append(out, ct.r(m.Values())+";"+bstr(m.Values().Contains_Object(val))+";"+bstr(m.ContainsValue_Object(val)))
	return strings.Join(out, "/")
}

// newScenario creates the collection that a "new" line describes. It
// returns the rendered result for lines that give one.
func (ct *collTester) newScenario(fields []string) string {
	kind := fields[0]
	ctor := ""
	if len(fields) > 2 {
		ctor = fields[2]
	}
	switch kind {
	case "ArrayList", "LinkedList", "Vector", "Stack":
		var l List
		switch kind {
		case "ArrayList":
			switch ctor {
			case "collection":
				l = ArrayList_new_Collection(ct.coll(fields[3]))
			case "cap":
				l = ArrayList_new_int(atoi32(fields[3]))
			case "capfail":
				return ct.try(func() string {
					ArrayList_new_int(atoi32(fields[3]))
					return "ok"
				})
			default:
				l = ArrayList_new()
			}
		case "LinkedList":
			if ctor == "collection" {
				l = LinkedList_new_Collection(ct.coll(fields[3]))
			} else {
				l = LinkedList_new()
			}
		case "Vector":
			switch ctor {
			case "collection":
				l = Vector_new_Collection(ct.coll(fields[3]))
			case "capinc":
				l = Vector_new_int_int(atoi32(fields[3]), atoi32(fields[4]))
			case "cap":
				l = Vector_new_int(atoi32(fields[3]))
			case "capfail":
				return ct.try(func() string {
					Vector_new_int(atoi32(fields[3]))
					return "ok"
				})
			default:
				l = Vector_new()
			}
		default:
			l = Stack_new()
		}
		ct.state = l
	case "ArrayDeque":
		switch ctor {
		case "cap":
			ct.state = ArrayDeque_new_int(atoi32(fields[3]))
		case "collection":
			ct.state = ArrayDeque_new_Collection(ct.coll(fields[3]))
		default:
			ct.state = ArrayDeque_new()
		}
	case "PriorityQueue":
		if ctor == "fail" {
			return ""
		}
		cn := fields[2]
		ctor = fields[3]
		return ct.try(func() string {
			switch ctor {
			case "new":
				if cn == "nat" {
					ct.state = PriorityQueue_new()
				} else {
					ct.state = PriorityQueue_new_Comparator(ct.cmp(cn))
				}
			case "capcmp":
				ct.state = PriorityQueue_new_int_Comparator(atoi32(fields[4]), ct.cmp(cn))
			case "collection":
				ct.state = PriorityQueue_new_Collection(ct.coll(fields[4]))
			case "sortedset":
				ct.state = PriorityQueue_new_SortedSet(ct.coll(fields[4]).(SortedSet))
			case "pq":
				var src *PriorityQueue
				ct.quietly(func() {
					if cn == "nat" {
						src = PriorityQueue_new()
					} else {
						src = PriorityQueue_new_Comparator(ct.cmp(cn))
					}
					src.AddAll_Collection(ct.coll(fields[4]))
				})
				ct.state = PriorityQueue_new_PriorityQueue(src)
			}
			return ct.r(ct.state)
		})
	case "TreeSet":
		if ctor == "fail" {
			return ""
		}
		cn := fields[2]
		ctor = fields[3]
		return ct.try(func() string {
			switch ctor {
			case "new":
				if cn == "nat" {
					ct.state = TreeSet_new()
				} else {
					ct.state = TreeSet_new_Comparator(ct.cmp(cn))
				}
			case "collection":
				ct.state = TreeSet_new_Collection(ct.coll(fields[4]))
			case "sortedset":
				ct.state = TreeSet_new_SortedSet(ct.coll(fields[4]).(SortedSet))
			case "sortedcollection":
				ct.state = TreeSet_new_Collection(ct.coll(fields[4]))
			}
			return ct.r(ct.state)
		})
	case "TreeMap":
		if ctor == "fail" {
			return ""
		}
		cn := fields[2]
		ctor = fields[3]
		ct.held = nil
		return ct.try(func() string {
			switch ctor {
			case "new":
				if cn == "nat" {
					ct.state = TreeMap_new()
				} else {
					ct.state = TreeMap_new_Comparator(ct.cmp(cn))
				}
			case "map":
				ct.state = TreeMap_new_Map(ct.mapLit(fields[4]))
			case "sortedmap":
				ct.state = TreeMap_new_SortedMap(ct.mapLit(fields[4]).(SortedMap))
			case "puts":
				var m *TreeMap
				if cn == "nat" {
					m = TreeMap_new()
				} else {
					m = TreeMap_new_Comparator(ct.cmp(cn))
				}
				for _, kv := range fields[4:] {
					p := strings.SplitN(kv, "=", 2)
					m.Put_Object_Object(ct.parse(p[0]), ct.parse(p[1]))
				}
				ct.state = m
			}
			return ct.r(ct.state)
		})
	case "EnumSet":
		_, elems := ct.listLit(fields[3])
		parts := make([]*Enum, len(elems))
		for i, e := range elems {
			parts[i] = &e.(*collOrd).Enum
		}
		var s *EnumSet
		switch len(parts) {
		case 1:
			s = EnumSet_of_Enum(parts[0])
		case 2:
			s = EnumSet_of_Enum_Enum(parts[0], parts[1])
		case 3:
			s = EnumSet_of_Enum_Enum_Enum(parts[0], parts[1], parts[2])
		case 4:
			s = EnumSet_of_Enum_Enum_Enum_Enum(parts[0], parts[1], parts[2], parts[3])
		case 5:
			s = EnumSet_of_Enum_Enum_Enum_Enum_Enum(parts[0], parts[1], parts[2], parts[3], parts[4])
		default:
			s = EnumSet_of_Enum_EnumArr(parts[0], parts[1:])
		}
		ct.state = s
		return ct.r(s)
	case "AsList":
		arr := ct.objArray(fields[2])
		if fields[1] == "strarr" {
			strs := make([]*String, len(arr))
			for i, e := range arr {
				if e != nil {
					strs[i] = e.(*String)
				}
			}
			ct.held = []any{strs}
			ct.state = Arrays_asList(strs)
		} else {
			ct.held = []any{arr}
			ct.state = Arrays_asList(arr)
		}
	case "singletonList":
		ct.state = Collections_singletonList(ct.parse(fields[2]))
	case "emptyList":
		ct.state = Collections_emptyList()
	case "nCopies":
		ct.state = Collections_nCopies(atoi32(fields[2]), ct.parse(fields[3]))
	case "unmodifiableList", "unmodifiableLL":
		ct.state = Collections_unmodifiableList(ct.coll(fields[2]).(List))
	case "singleton":
		ct.state = Collections_singleton(ct.parse(fields[2]))
	case "emptySet":
		ct.state = Collections_emptySet()
	case "unmodifiableSet":
		ct.state = Collections_unmodifiableSet(ct.coll(fields[2]).(Set))
	case "unmodifiableCollection":
		ct.state = Collections_unmodifiableCollection(ct.coll(fields[2]))
	case "unmodifiableMap":
		ct.state = Collections_unmodifiableMap(ct.mapLit(fields[2]))
	case "emptyMap":
		ct.state = Collections_emptyMap()
	default:
		ct.t.Fatalf("%s:%d: unknown scenario %q", ct.file, ct.line, kind)
	}
	return ""
}

// op runs one operation line's operation on the current scenario.
func (ct *collTester) op(name string, args []string) string {
	if name == "ks" {
		ks := ct.state.(*TreeMap).NavigableKeySet()
		name, args = args[0], args[1:]
		return ct.try(func() string {
			if res, ok := ct.collectionOp(ks, name, args); ok {
				return res
			}
			if res, ok := ct.navSetOp(ks, name, args); ok {
				return res
			}
			ct.t.Fatalf("%s:%d: unknown key set operation %q", ct.file, ct.line, name)
			return ""
		})
	}
	if name == "vs" {
		vs := ct.state.(*TreeMap).Values()
		name, args = args[0], args[1:]
		return ct.try(func() string {
			if res, ok := ct.collectionOp(vs, name, args); ok {
				return res
			}
			ct.t.Fatalf("%s:%d: unknown values operation %q", ct.file, ct.line, name)
			return ""
		})
	}
	if static(name) {
		ct.state = nil
		return ct.try(func() string { return ct.staticOp(name, args) })
	}
	return ct.try(func() string {
		if name == "array" {
			return ct.r(ct.held[0])
		}
		if m, ok := ct.state.(Map); ok {
			if res, ok := ct.mapOp(m, name, args); ok {
				return res
			}
			ct.t.Fatalf("%s:%d: unknown map operation %q", ct.file, ct.line, name)
		}
		c := ct.state.(Collection)
		if res, ok := ct.collectionOp(c, name, args); ok {
			return res
		}
		if l, ok := c.(List); ok {
			if res, ok := ct.listOp(l, name, args); ok {
				return res
			}
		}
		if d, ok := c.(Deque); ok {
			if res, ok := ct.dequeOp(d, name, args); ok {
				return res
			}
		}
		if v, ok := c.(interface{ JClass_Vector() *Vector }); ok {
			if res, ok := ct.vectorOp(v.JClass_Vector(), name, args); ok {
				return res
			}
		}
		if q, ok := c.(*PriorityQueue); ok {
			if res, ok := ct.pqOp(q, name, args); ok {
				return res
			}
		}
		if s, ok := c.(NavigableSet); ok {
			if res, ok := ct.navSetOp(s, name, args); ok {
				return res
			}
		}
		if s, ok := c.(*EnumSet); ok {
			switch name {
			case "clone":
				return collSimpleName(s.Clone()) + ":" + ct.r(s.Clone())
			case "containsInt":
				return bstr(s.Contains_Object(int32(5)))
			case "copyOfSet":
				return ct.r(EnumSet_copyOf_EnumSet(s))
			case "copyOf":
				return ct.r(EnumSet_copyOf_Collection(ct.coll(args[0])))
			}
		}
		ct.t.Fatalf("%s:%d: unknown operation %q on %T", ct.file, ct.line, name, c)
		return ""
	})
}

func static(name string) bool {
	switch name {
	case "sortarr", "sortrange", "sortlist", "sortgen", "sortd", "sortdr", "sortf", "sortfr",
		"sorti", "sortir", "sortl", "sortlr", "sorts", "sortsr", "sortc", "sortcr", "sortb", "sortbr",
		"bsi", "bsir", "bsd", "bsf", "bso", "bsl", "copyofi", "copyofd", "copyoff", "copyofo",
		"copyofrangei", "copyofrangeo", "eqo", "eqd", "hashd", "hashi", "hasho", "tostro", "tostrd",
		"filldr", "fillo", "deep", "deepself", "reverse", "reverseLL", "min", "max", "swap",
		"addAllArr", "revcmp", "fill", "minnat", "cmpboxed", "setEqualsCCE", "setEqualsNPE", "mapEqualsCCE", "primj", "prims", "primc", "primb", "primz", "primf":
		return true
	}
	return false
}

func TestCollVectors(t *testing.T) {
	for _, name := range []string{"lists", "deques", "pq", "treeset", "treemap", "enumset", "static"} {
		t.Run(name, func(t *testing.T) {
			runCollVectors(t, "testdata/coll_"+name+".txt")
		})
	}
}

func runCollVectors(t *testing.T, path string) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ct := newCollTester(t, path)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	failures := 0
	ops := 0
	for sc.Scan() {
		ct.line++
		line := sc.Text()
		lhs, rhs, hasResult := strings.Cut(line, " => ")
		fields := strings.Fields(lhs)
		*ct.log = collCmpLog{}
		var got, want string
		if fields[0] == "new" {
			ct.state = nil
			ct.lastState = ""
			res := ct.newScenario(fields[1:])
			if !hasResult {
				continue
			}
			if strings.Contains(rhs, " | ") {
				got = res + " | " + fmt.Sprintf("%d:%d", ct.log.count, ct.log.hash)
			} else {
				got = res
			}
			want = rhs
		} else {
			res := ct.op(fields[0], fields[1:])
			st := ct.r(ct.state)
			if st == ct.lastState {
				st = "="
			} else {
				ct.lastState = st
			}
			got = res + " | " + st + " | " + fmt.Sprintf("%d:%d", ct.log.count, ct.log.hash)
			want = rhs
		}
		ops++
		if got != want {
			failures++
			t.Errorf("%s:%d: %s\n got: %s\nwant: %s", path, ct.line, lhs, got, want)
			if failures > 10 {
				t.Fatal("too many failures")
			}
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d vectors", ops)
}

// ---------------------------------------------------------------------------
// Arrays, Collections and TimSort

func parseHexBits(s string) uint64 {
	v, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func primLit(s string) []string {
	open := strings.IndexByte(s, '[')
	body := s[open+1 : len(s)-1]
	if body == "" {
		return nil
	}
	return strings.Split(body, ",")
}

func doublesLit(s string) []float64 {
	p := primLit(s)
	out := make([]float64, len(p))
	for i, x := range p {
		out[i] = math.Float64frombits(parseHexBits(x))
	}
	return out
}

func floatsLit(s string) []float32 {
	p := primLit(s)
	out := make([]float32, len(p))
	for i, x := range p {
		out[i] = math.Float32frombits(uint32(parseHexBits(x)))
	}
	return out
}

func intsLit[T int8 | int16 | uint16 | int32 | int64](s string) []T {
	p := primLit(s)
	out := make([]T, len(p))
	for i, x := range p {
		v, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			panic(err)
		}
		out[i] = T(v)
	}
	return out
}

func renderPrims[T any](tag string, a []T, f func(T) string) string {
	parts := make([]string, len(a))
	for i, x := range a {
		parts[i] = f(x)
	}
	return tag + "[" + strings.Join(parts, ",") + "]"
}

// collNonNegative returns a, or 0 if a is negative.
func collNonNegative(a int32) int32 {
	if a > 0 {
		return a
	}
	return 0
}

func zLit(a []bool) string {
	return renderPrims("Z", a, func(b bool) string {
		if b {
			return "1"
		}
		return "0"
	})
}

func dLit(a []float64) string {
	return renderPrims("D", a, func(d float64) string { return fmt.Sprintf("%016x", math.Float64bits(d)) })
}

func fLit(a []float32) string {
	return renderPrims("F", a, func(f float32) string { return fmt.Sprintf("%08x", math.Float32bits(f)) })
}

func intLit[T int8 | int16 | uint16 | int32 | int64](tag string, a []T) string {
	return renderPrims(tag, a, func(x T) string { return strconv.FormatInt(int64(x), 10) })
}

// boxed parses the Java program's boxLit.
func (ct *collTester) boxed(s string) any {
	n := func() int64 {
		v, err := strconv.ParseInt(s[1:], 10, 64)
		if err != nil {
			panic(err)
		}
		return v
	}
	switch s[0] {
	case 'L':
		return n()
	case 'D':
		return math.Float64frombits(parseHexBits(s[1:]))
	case 'F':
		return math.Float32frombits(uint32(parseHexBits(s[1:])))
	case 'S':
		return int16(n())
	case 'B':
		return int8(n())
	case 'C':
		return uint16(n())
	case 'Z':
		return s[1] == '1'
	}
	return ct.parse(s)
}

// collLCG mirrors the Java program's generator for large sort inputs.
type collLCG struct{ x uint64 }

func (g *collLCG) next(domain int32) int32 {
	g.x = g.x*6364136223846793005 + 1442695040888963407
	return int32((g.x >> 33) % uint64(domain))
}

func (ct *collTester) staticOp(name string, args []string) string {
	rangeArgs := func() (int32, int32, string) { return atoi32(args[0]), atoi32(args[1]), args[2] }
	switch name {
	case "sortarr":
		a := ct.objArray(args[1])
		res := ct.try(func() string {
			Arrays_sort_ObjectArr_Comparator(a, ct.cmp(args[0]))
			return ""
		})
		if res != "" {
			return res + ":" + ct.r(a)
		}
		return ct.r(a)
	case "sortrange":
		a := ct.objArray(args[3])
		res := ct.try(func() string {
			Arrays_sort_ObjectArr_int_int_Comparator(a, atoi32(args[1]), atoi32(args[2]), ct.cmp(args[0]))
			return ""
		})
		if res != "" {
			return res + ":" + ct.r(a)
		}
		return ct.r(a)
	case "sortlist":
		l := ct.coll(args[1]).(List)
		res := ct.try(func() string {
			Collections_sort_List_Comparator(l, ct.cmp(args[0]))
			return ""
		})
		if res != "" {
			return res + ":" + ct.r(l)
		}
		return ct.r(l)
	case "sortgen":
		seed, _ := strconv.ParseInt(args[0], 10, 64)
		n, domain, typ := atoi32(args[1]), atoi32(args[2]), atoi32(args[3])
		g := &collLCG{x: uint64(seed)}
		a := make([]any, n)
		for i := range a {
			x := g.next(domain)
			if typ == 0 {
				a[i] = x
			} else {
				a[i] = &collItem{log: ct.log, key: x, id: g.next(1000)}
			}
		}
		keyCmp := ComparatorFunc(func(x, y any) int32 { return Integer_compare(collKeyOf(x), collKeyOf(y)) })
		revCmp := ComparatorFunc(func(x, y any) int32 { return Integer_compare(collKeyOf(y), collKeyOf(x)) })
		if domain%3 == 0 {
			Arrays_sort_ObjectArr_int_int_Comparator(a, 0, n/2, keyCmp)
		}
		if domain%5 == 0 {
			Arrays_sort_ObjectArr_int_int_Comparator(a, n/3, n, revCmp)
		}
		*ct.log = collCmpLog{}
		ex := ct.try(func() string {
			Arrays_sort_ObjectArr_Comparator(a, ct.cmp(args[4]))
			return ""
		})
		if ex != "" {
			ex += ":"
		}
		var h int32
		for _, o := range a {
			h = h*31 + FromGo(ct.r(o)).HashCode()
		}
		return fmt.Sprintf("%s%d#%d", ex, len(a), h)
	case "sortd", "sortdr":
		var a []float64
		if name == "sortdr" {
			f, t, s := rangeArgs()
			a = doublesLit(s)
			Arrays_sort_doubleArr_int_int(a, f, t)
		} else {
			a = doublesLit(args[0])
			Arrays_sort_doubleArr(a)
		}
		return dLit(a)
	case "sortf", "sortfr":
		var a []float32
		if name == "sortfr" {
			f, t, s := rangeArgs()
			a = floatsLit(s)
			Arrays_sort_floatArr_int_int(a, f, t)
		} else {
			a = floatsLit(args[0])
			Arrays_sort_floatArr(a)
		}
		return fLit(a)
	case "sorti", "sortir":
		var a []int32
		if name == "sortir" {
			f, t, s := rangeArgs()
			a = intsLit[int32](s)
			Arrays_sort_intArr_int_int(a, f, t)
		} else {
			a = intsLit[int32](args[0])
			Arrays_sort_intArr(a)
		}
		return intLit("I", a)
	case "sortl", "sortlr":
		var a []int64
		if name == "sortlr" {
			f, t, s := rangeArgs()
			a = intsLit[int64](s)
			Arrays_sort_longArr_int_int(a, f, t)
		} else {
			a = intsLit[int64](args[0])
			Arrays_sort_longArr(a)
		}
		return intLit("J", a)
	case "sorts", "sortsr":
		var a []int16
		if name == "sortsr" {
			f, t, s := rangeArgs()
			a = intsLit[int16](s)
			Arrays_sort_shortArr_int_int(a, f, t)
		} else {
			a = intsLit[int16](args[0])
			Arrays_sort_shortArr(a)
		}
		return intLit("S", a)
	case "sortc", "sortcr":
		var a []uint16
		if name == "sortcr" {
			f, t, s := rangeArgs()
			a = intsLit[uint16](s)
			Arrays_sort_charArr_int_int(a, f, t)
		} else {
			a = intsLit[uint16](args[0])
			Arrays_sort_charArr(a)
		}
		return intLit("C", a)
	case "sortb", "sortbr":
		var a []int8
		if name == "sortbr" {
			f, t, s := rangeArgs()
			a = intsLit[int8](s)
			Arrays_sort_byteArr_int_int(a, f, t)
		} else {
			a = intsLit[int8](args[0])
			Arrays_sort_byteArr(a)
		}
		return intLit("B", a)
	case "bsi":
		return istr(Arrays_binarySearch_intArr_int(intsLit[int32](args[1]), atoi32(args[0])))
	case "bsir":
		return istr(Arrays_binarySearch_intArr_int_int_int(intsLit[int32](args[3]), atoi32(args[0]), atoi32(args[1]), atoi32(args[2])))
	case "bsd":
		return istr(Arrays_binarySearch_doubleArr_double(doublesLit(args[1]), math.Float64frombits(parseHexBits(args[0]))))
	case "bsf":
		return istr(Arrays_binarySearch_floatArr_float(floatsLit(args[1]), math.Float32frombits(uint32(parseHexBits(args[0])))))
	case "bso":
		return istr(Arrays_binarySearch_ObjectArr_Object_Comparator(ct.objArray(args[2]), ct.parse(args[1]), ct.cmp(args[0])))
	case "bsl":
		return istr(Collections_binarySearch_List_Object_Comparator(ct.coll(args[2]).(List), ct.parse(args[1]), ct.cmp(args[0])))
	case "copyofi":
		return intLit("I", Arrays_copyOf_intArr_int(intsLit[int32](args[1]), atoi32(args[0])))
	case "copyofd":
		return dLit(Arrays_copyOf_doubleArr_int(doublesLit(args[1]), atoi32(args[0])))
	case "copyoff":
		return fLit(Arrays_copyOf_floatArr_int(floatsLit(args[1]), atoi32(args[0])))
	case "copyofo":
		return ct.r(Arrays_copyOf_ObjectArr_int(ct.objArray(args[1]), atoi32(args[0])))
	case "copyofrangei":
		return intLit("I", Arrays_copyOfRange_intArr_int_int(intsLit[int32](args[2]), atoi32(args[0]), atoi32(args[1])))
	case "copyofrangeo":
		return ct.r(Arrays_copyOfRange_ObjectArr_int_int(ct.objArray(args[2]), atoi32(args[0]), atoi32(args[1])))
	case "eqo":
		return bstr(Arrays_equals_ObjectArr_ObjectArr(ct.objArray(args[0]), ct.objArray(args[1])))
	case "eqd":
		return bstr(Arrays_equals_doubleArr_doubleArr(doublesLit(args[0]), doublesLit(args[1])))
	case "hashd":
		return istr(Arrays_hashCode_doubleArr(doublesLit(args[0])))
	case "hashi":
		return istr(Arrays_hashCode_intArr(intsLit[int32](args[0])))
	case "hasho":
		return istr(Arrays_hashCode_ObjectArr(ct.objArray(args[0])))
	case "tostro":
		return strings.ReplaceAll(Arrays_toString_ObjectArr(ct.objArray(args[0])).Go(), " ", "_")
	case "tostrd":
		return strings.ReplaceAll(Arrays_toString_doubleArr(doublesLit(args[0])).Go(), " ", "_")
	case "filldr":
		a := doublesLit(args[3])
		Arrays_fill_doubleArr_int_int_double(a, atoi32(args[1]), atoi32(args[2]), math.Float64frombits(parseHexBits(args[0])))
		return dLit(a)
	case "fillo":
		a := ct.objArray(args[1])
		Arrays_fill_ObjectArr_Object(a, ct.parse(args[0]))
		return ct.r(a)
	case "cmpboxed":
		return istr(Comparable_CompareTo_Object(ct.boxed(args[0]), ct.boxed(args[1])))
	case "setEqualsCCE":
		a, b := TreeSet_new(), TreeSet_new()
		a.Add_Object(int32(1))
		a.Add_Object(int32(2))
		b.Add_Object(FromGo("a"))
		b.Add_Object(FromGo("b"))
		return bstr(a.Equals_Object(b))
	case "setEqualsNPE":
		a := TreeSet_new()
		a.Add_Object(int32(1))
		return bstr(a.Equals_Object(Collections_singleton(nil)))
	case "mapEqualsCCE":
		a, b := TreeMap_new(), TreeMap_new()
		a.Put_Object_Object(int32(1), int32(1))
		b.Put_Object_Object(FromGo("a"), int32(1))
		return bstr(a.Equals_Object(b))
	case "primj":
		j := intsLit[int64](args[0])
		newLen := atoi32(args[1])
		key, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			panic(err)
		}
		sorted := append([]int64(nil), j...)
		Arrays_sort_longArr(sorted)
		return strings.ReplaceAll(Arrays_toString_longArr(j).Go(), " ", "_") + ";" + istr(Arrays_hashCode_longArr(j)) + ";" +
			bstr(Arrays_equals_longArr_longArr(j, append([]int64{}, j...))) + ";" + istr(Arrays_binarySearch_longArr_long(sorted, key)) +
			";" + intLit("J", Arrays_copyOf_longArr_int(j, collNonNegative(newLen)))
	case "prims":
		a := intsLit[int16](args[0])
		newLen := atoi32(args[1])
		return strings.ReplaceAll(Arrays_toString_shortArr(a).Go(), " ", "_") + ";" + istr(Arrays_hashCode_shortArr(a)) + ";" +
			intLit("S", Arrays_copyOf_shortArr_int(a, collNonNegative(newLen)))
	case "primc":
		a := intsLit[uint16](args[0])
		newLen := atoi32(args[1])
		sorted := append([]uint16(nil), a...)
		Arrays_sort_charArr(sorted)
		key := uint16('a')
		if len(a) > 0 {
			key = a[0]
		}
		return istr(Arrays_hashCode_charArr(a)) + ";" + intLit("C", Arrays_copyOf_charArr_int(a, collNonNegative(newLen))) + ";" +
			istr(Arrays_binarySearch_charArr_char(sorted, key))
	case "primb":
		a := intsLit[int8](args[0])
		newLen := atoi32(args[1])
		return strings.ReplaceAll(Arrays_toString_byteArr(a).Go(), " ", "_") + ";" + istr(Arrays_hashCode_byteArr(a)) + ";" +
			intLit("B", Arrays_copyOf_byteArr_int(a, collNonNegative(newLen)))
	case "primz":
		p := primLit(args[0])
		z := make([]bool, len(p))
		for i, x := range p {
			z[i] = x == "1"
		}
		newLen := atoi32(args[1])
		filled := append([]bool{}, z...)
		Arrays_fill_booleanArr_boolean(filled, true)
		return strings.ReplaceAll(Arrays_toString_booleanArr(z).Go(), " ", "_") + ";" + istr(Arrays_hashCode_booleanArr(z)) + ";" +
			zLit(Arrays_copyOf_booleanArr_int(z, collNonNegative(newLen))) + ";" + zLit(filled) + ";" + bstr(Arrays_equals_booleanArr_booleanArr(z, filled))
	case "primf":
		f, g := floatsLit(args[0]), floatsLit(args[1])
		return strings.ReplaceAll(Arrays_toString_floatArr(f).Go(), " ", "_") + ";" + istr(Arrays_hashCode_floatArr(f)) + ";" +
			bstr(Arrays_equals_floatArr_floatArr(f, g))
	case "deep":
		a := []any{int32(1), []any{FromGo("a"), nil, []float64{1.5, math.Copysign(0, -1)}}, []int32{3, 4}, nil, []any{}}
		return strings.ReplaceAll(Arrays_deepToString(a).Go(), " ", "_")
	case "deepself":
		a := make([]any, 2)
		a[0] = a
		a[1] = FromGo("x")
		return strings.ReplaceAll(Arrays_deepToString(a).Go(), " ", "_")
	case "reverse", "reverseLL":
		l := ct.coll(args[0]).(List)
		Collections_reverse(l)
		return ct.r(l)
	case "min":
		return ct.r(Collections_min_Collection_Comparator(ct.coll(args[1]), ct.cmp(args[0])))
	case "max":
		return ct.r(Collections_max_Collection_Comparator(ct.coll(args[1]), ct.cmp(args[0])))
	case "minnat":
		c := ct.coll(args[0])
		return ct.r(Collections_min_Collection(c)) + "," + ct.r(Collections_max_Collection(c))
	case "swap":
		l := ct.coll(args[2]).(List)
		Collections_swap(l, atoi32(args[0]), atoi32(args[1]))
		return ct.r(l)
	case "addAllArr":
		l := ct.coll(args[0]).(List)
		b := Collections_addAll(l, ct.objArray(args[1]))
		return bstr(b) + ":" + ct.r(l)
	case "revcmp":
		c := Collections_reverseOrder_Comparator(ct.cmp(args[0]))
		c2 := Collections_reverseOrder_Comparator(c)
		x, y := ct.parse(args[1]), ct.parse(args[2])
		return istr(c.Compare_Object_Object(x, y)) + "," + istr(c2.Compare_Object_Object(x, y))
	case "fill":
		l := ct.coll(args[1]).(List)
		Collections_fill(l, ct.parse(args[0]))
		return ct.r(l)
	}
	ct.t.Fatalf("%s:%d: unknown static operation %q", ct.file, ct.line, name)
	return ""
}

// ---------------------------------------------------------------------------
// Conventions shared with the generated code, which the JDK vectors cannot
// exercise.

// CollTestBase and CollTestSub model a generated class hierarchy: arrays of
// the base class hold part pointers, JToAny returns the whole object, and
// JClass_ methods return the part for an array's element class.
type CollTestBase struct {
	jkind_ int32
	v      int32
}

type CollTestSub struct {
	CollTestBase
}

func (b *CollTestBase) JClass_CollTestBase() *CollTestBase { return b }
func (b *CollTestBase) JClassName() string                 { return "CollTestBase" }

func (b *CollTestBase) JToAny() any {
	if b.jkind_ == 1 {
		return (*CollTestSub)(unsafe.Pointer(b))
	}
	return b
}

func (b *CollTestBase) CompareTo_Object(o any) int32 {
	return Integer_compare(b.v, o.(interface{ JClass_CollTestBase() *CollTestBase }).JClass_CollTestBase().v)
}

func (s *CollTestSub) JClassName() string { return "CollTestSub" }

func newCollTestSub(v int32) *CollTestSub {
	return &CollTestSub{CollTestBase{jkind_: 1, v: v}}
}

func TestCollArraysOfPartPointers(t *testing.T) {
	s1, s2 := newCollTestSub(3), newCollTestSub(1)
	b := &CollTestBase{v: 2}
	arr := []*CollTestBase{&s1.CollTestBase, b, &s2.CollTestBase, nil}

	l := Arrays_asList(arr)
	if got, ok := l.Get_int(0).(*CollTestSub); !ok || got != s1 {
		t.Fatalf("Get_int(0) = %#v, want the whole subclass object", l.Get_int(0))
	}
	if l.Get_int(3) != nil {
		t.Fatalf("Get_int(3) = %#v, want nil", l.Get_int(3))
	}
	s3 := newCollTestSub(0)
	l.Set_int_Object(3, s3)
	if arr[3] != &s3.CollTestBase {
		t.Fatalf("Set_int_Object did not write the part pointer through")
	}

	Arrays_sort_ObjectArr(arr)
	var got []int32
	for _, e := range arr {
		got = append(got, e.v)
	}
	if fmt.Sprint(got) != "[0 1 2 3]" || arr[0] != &s3.CollTestBase || arr[2] != b {
		t.Fatalf("sorted array = %v", got)
	}

	al := ArrayList_new_Collection(l)
	out := al.ToArray_ObjectArr([]*CollTestBase{}).([]*CollTestBase)
	for i := range arr {
		if out[i] != arr[i] {
			t.Fatalf("ToArray_ObjectArr element %d = %p, want %p", i, out[i], arr[i])
		}
	}
}

// CollTestList models a generated class that extends ArrayList.
type CollTestList struct {
	ArrayList
	extra int32
}

func (l *CollTestList) JClassName() string { return "CollTestList" }

// JShallowCopy copies every field, as the generated JShallowCopy's *c = *l
// does. It copies through reflect because vet rejects copying the
// ArrayList's monitor, which ArrayList.Clone resets in the copy.
func (l *CollTestList) JShallowCopy() any {
	c := new(CollTestList)
	reflect.ValueOf(c).Elem().Set(reflect.ValueOf(l).Elem())
	return c
}

func TestCollArrayListSubclassClone(t *testing.T) {
	l := &CollTestList{extra: 7}
	l.JSetSelf(l)
	l.ArrayList_init()
	l.Add_Object(int32(1))
	c, ok := l.ArrayList.Clone().(*CollTestList)
	if !ok {
		t.Fatalf("Clone returned %T", l.ArrayList.Clone())
	}
	if c == l || c.extra != 7 || c.JSelf() != any(c) {
		t.Fatalf("clone is not a separate copy of the whole object")
	}
	c.Add_Object(int32(2))
	if l.Size() != 1 || c.Size() != 2 {
		t.Fatalf("sizes after adding to the clone: %d, %d", l.Size(), c.Size())
	}
	if got := Object_ToString(c).Go(); got != "[1, 2]" {
		t.Fatalf("toString = %q", got)
	}

	s := Stack_new()
	s.Push_Object(int32(1))
	sc, ok := s.Clone().(*Stack)
	if !ok || sc == s || sc.JSelf() != any(sc) || Object_ToString(sc).Go() != "[1]" {
		t.Fatalf("Stack clone = %#v", s.Clone())
	}
	if !Vector_isInstance(s) || Vector_fromAny(Ref(&s.Vector)) != &s.Vector || Ref(&s.Vector) != any(s) {
		t.Fatalf("Stack does not convert to its Vector part and back")
	}
}
