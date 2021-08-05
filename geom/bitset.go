package geom

import "math/bits"

// BitSet is a set data structure that holds a mapping from non-negative
// integers to boolean values (bits). The zero value is the BitSet with all
// bits set to false.
type BitSet struct {
	masks []uint64
}

// Get gets the bit as position i. It panics if i is negative. Get returns
// false for bits that haven't been explicitly set.
func (b *BitSet) Get(i int) bool {
	idx := i / 64
	if idx >= len(b.masks) {
		return false
	}
	return (b.masks[idx] & (1 << (i % 64))) != 0
}

// Set sets the bit in position i to a new value.
func (b *BitSet) Set(i int, newVal bool) {
	if newVal {
		idx := i / 64
		if idx >= len(b.masks) {
			b.masks = append(
				b.masks,
				make([]uint64, idx-len(b.masks)+1)...,
			)
		}
		b.masks[idx] |= (1 << (i % 64))
	} else {
		idx := i / 64
		if idx < len(b.masks) {
			b.masks[idx] &= ^(1 << (i % 64))
		}
	}
}

// CountTrue counts the number of elements in the set that are true.
func (b *BitSet) CountTrue() int {
	var count int
	for _, mask := range b.masks {
		count += bits.OnesCount64(mask)
	}
	return count
}

// Clone makes a deep copy of the BitSet.
func (b *BitSet) Clone() BitSet {
	return BitSet{append([]uint64(nil), b.masks...)}
}
