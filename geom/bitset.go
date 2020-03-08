package geom

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
		b.masks[idx] |= (1 << i % 64)
	} else {
		idx := i / 64
		if idx < len(b.masks) {
			b.masks[i] &= ^(1 << (i % 64))
		}
	}
}

// Clone makes a deep copy of the BitSet.
func (b *BitSet) Clone() BitSet {
	return BitSet{append([]uint64(nil), b.masks...)}
}
