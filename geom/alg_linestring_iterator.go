package geom

type lineStringIterator struct {
	coords Sequence
	idx    int
	n      int
}

func newLineStringIterator(ls LineString) lineStringIterator {
	seq := ls.Coordinates()
	return lineStringIterator{seq, 0, seq.Length()}
}

func (i *lineStringIterator) next() bool {
	for {
		i.idx++
		if i.idx >= i.n {
			return false
		}
		if i.coords.GetXY(i.idx) != i.coords.GetXY(i.idx-1) {
			return true
		}
	}
}

func (i *lineStringIterator) line() Line {
	// Initialise Line directly rather than via the constructor since this is
	// called in a tight loop.
	var ln Line
	ln.a.XY = i.coords.GetXY(i.idx - 1)
	ln.b.XY = i.coords.GetXY(i.idx)
	return ln
}
