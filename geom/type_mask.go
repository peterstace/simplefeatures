package geom

type BitSet struct {
	masks []uint64
}

func (b *BitSet) Get(i int) bool {
	idx := i / 64
	if idx >= len(b.masks) {
		return false
	}
	return (b.masks[idx] & (1 << (i % 64))) != 0
}

func (b *BitSet) Set(i int) {
	idx := i / 64
	if idx >= len(b.masks) {
		b.masks = append(
			b.masks,
			make([]uint64, idx-len(b.masks)+1)...,
		)
	}
	b.masks[idx] |= (1 << i % 64)
}

func (b *BitSet) Clear(i int) {
	idx := i / 64
	if idx < len(b.masks) {
		b.masks[i] &= ^(1 << (i % 64))
	}
}

func (b *BitSet) Clone() BitSet {
	return BitSet{append([]uint64(nil), b.masks...)}
}
