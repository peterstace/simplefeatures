package java

import (
	"cmp"
	"slices"
)

// SortedKeys returns the keys of a map in sorted order. This is used when
// transliterating Java code that iterates over maps, because Go's map iteration
// order is randomized while Java's HashMap iteration order is consistent (even
// though unspecified), and Java's TreeMap iteration order is sorted.
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
