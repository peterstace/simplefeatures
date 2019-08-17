package geom

import (
	"crypto/sha256"
	"encoding/binary"
)

type XY struct {
	X, Y float64
}

// TODO: no must...
// MustNewXYF creates an XY from x and y float64 values. If either value is not
// finite, then it panics.
func MustNewXYF(x, y float64) XY {
	return XY{x, y}
}

func (w XY) Equals(o XY) bool {
	// TODO: epsilon check?
	return w.X == o.X && w.Y == o.Y
}

func (w XY) Sub(o XY) XY {
	return XY{
		w.X - o.X,
		w.Y - o.Y,
	}
}

func (w XY) Add(o XY) XY {
	return XY{
		w.X + o.X,
		w.Y + o.Y,
	}
}

func (w XY) Scale(s float64) XY {
	return XY{
		w.X * s,
		w.Y * s,
	}
}

func (w XY) Cross(o XY) float64 {
	return (w.X * o.Y) - (w.Y * o.X)
}

func (w XY) Midpoint(o XY) XY {
	return w.Add(o).Scale(0.5)
}

func (w XY) Dot(o XY) float64 {
	return w.X*o.X + w.Y*o.Y
}

// Less gives an ordering on XYs. If two XYs have different X values, then the
// one with the lower X value is ordered before the one with the higher X
// value. If the X values are then same, then the Y values are used (the lower
// Y value comes first).
func (w XY) Less(o XY) bool {
	if w.X != o.X {
		return w.X < o.X
	}
	return w.Y < o.Y
}

type xyHash [sha256.Size]byte

// TODO: don't need a hash anymore, since XY is now fixed size.
func (w XY) hash() xyHash {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, w.X)
	binary.Write(h, binary.LittleEndian, w.Y)
	var sum xyHash
	h.Sum(sum[:0])
	return sum
}

type xySet map[xyHash]XY

func newXYSet() xySet {
	return make(map[xyHash]XY)
}

func (s xySet) add(xy XY) {
	s[xy.hash()] = xy
}

func (s xySet) contains(xy XY) bool {
	_, ok := s[xy.hash()]
	return ok
}

type xyxyHash struct {
	a, b xyHash
}

func hashXYXY(a, b XY) xyxyHash {
	return xyxyHash{
		a.hash(),
		b.hash(),
	}
}
