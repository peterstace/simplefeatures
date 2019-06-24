package simplefeatures

import (
	"crypto/sha256"
	"fmt"
)

type XY struct {
	X, Y Scalar
}

func (w XY) Equals(o XY) bool {
	return w.X.Equals(o.X) && w.Y.Equals(o.Y)
}

func (w XY) Sub(o XY) XY {
	return XY{
		w.X.Sub(o.X),
		w.Y.Sub(o.Y),
	}
}

func (w XY) Cross(o XY) Scalar {
	return w.X.Mul(o.Y).Sub(w.Y.Mul(o.X))
}

type xyHash [sha256.Size]byte

func (w XY) hash() xyHash {
	h := sha256.New()
	fmt.Fprintf(h, "%s,%s", w.X.val, w.Y.val)
	var sum xyHash
	h.Sum(sum[:0])
	return sum
}
