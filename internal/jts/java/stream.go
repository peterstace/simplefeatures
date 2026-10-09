package java

// Predicate is java.util.function.Predicate.
type Predicate interface {
	JIface_Predicate()
	Test_Object(any) bool
}

// PredicateFunc adapts a function literal (a translated lambda or method
// reference) to Predicate.
type PredicateFunc func(a any) bool

func (f PredicateFunc) JIface_Predicate()      {}
func (f PredicateFunc) Test_Object(a any) bool { return f(a) }

// Stream is java.util.stream.Stream, supporting the sequential operations
// JTS uses. A stream holds the elements of its source in encounter order.
type Stream interface {
	JIface_Stream()
	Filter_Predicate(Predicate) Stream
	Map_Function(Function) Stream
	Collect_Collector(Collector) any
	ForEach_Consumer(Consumer)
}

// Collector is java.util.stream.Collector.
type Collector interface {
	JIface_Collector()
	collect(elems []any) any
}

type sliceStream struct {
	elems []any
}

// Collection_default_Stream is the default method Collection.stream.
func Collection_default_Stream(c any) Stream {
	coll := AnyTo[Collection](c)
	var elems []any
	for it := coll.Iterator(); it.HasNext(); {
		elems = append(elems, it.Next())
	}
	return &sliceStream{elems: elems}
}

func (s *sliceStream) JIface_Stream() {}

func (s *sliceStream) Filter_Predicate(p Predicate) Stream {
	var out []any
	for _, e := range s.elems {
		if p.Test_Object(e) {
			out = append(out, e)
		}
	}
	return &sliceStream{elems: out}
}

func (s *sliceStream) Map_Function(f Function) Stream {
	out := make([]any, len(s.elems))
	for i, e := range s.elems {
		out[i] = f.Apply_Object(e)
	}
	return &sliceStream{elems: out}
}

func (s *sliceStream) Collect_Collector(c Collector) any {
	return c.collect(s.elems)
}

func (s *sliceStream) ForEach_Consumer(c Consumer) {
	for _, e := range s.elems {
		c.Accept_Object(e)
	}
}

type collectorFunc func(elems []any) any

func (f collectorFunc) JIface_Collector()       {}
func (f collectorFunc) collect(elems []any) any { return f(elems) }

// Collectors_toSet is Collectors.toSet, which collects into a HashSet in
// encounter order.
func Collectors_toSet() Collector {
	return collectorFunc(func(elems []any) any {
		s := HashSet_new()
		for _, e := range elems {
			s.Add_Object(e)
		}
		return s
	})
}

// Collectors_toList is Collectors.toList, which collects into an ArrayList.
func Collectors_toList() Collector {
	return collectorFunc(func(elems []any) any {
		l := ArrayList_new()
		for _, e := range elems {
			l.Add_Object(e)
		}
		return l
	})
}
