package java

import "sort"

// SortedKeysString returns the keys of a map[string]V in sorted order. This is
// used when transliterating Java code that iterates over maps, because Go's map
// iteration order is randomized while Java's HashMap iteration order is
// consistent (even though unspecified), and Java's TreeMap iteration order is
// sorted.
func SortedKeysString[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
