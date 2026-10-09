package java

// ComparatorFunc adapts a function literal (a translated lambda or method
// reference) to Comparator.
type ComparatorFunc func(a, b any) int32

func (f ComparatorFunc) Compare_Object_Object(a, b any) int32 { return f(a, b) }
