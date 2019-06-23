package simplefeatures

import (
	"errors"
	"math/big"
)

// Scalar represents a rational number.
type Scalar struct {
	val *big.Rat
}

var (
	zero = Scalar{val: big.NewRat(0, 1)}
	one  = Scalar{val: big.NewRat(1, 1)}
)

// NewScalar parses a string and returns the corresponding scalar. The string
// can be a decimal representation, or a fractional representation (the same as
// big.Rat.SetString).
func NewScalar(s string) (Scalar, error) {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return Scalar{}, errors.New("invalid scalar")
	}
	return Scalar{r}, nil
}

// AsFloat converts the scalar into a float64. If the scalar is too large to be
// represented as a float64, then the retuned value will be infinity or
// negative infinity.
func (s Scalar) AsFloat() float64 {
	f, _ := s.val.Float64()
	return f
}

// AsRat copies the scalar into r. If r is nil, then a new Rat is allocated.
func (s Scalar) AsRat(r *big.Rat) *big.Rat {
	if r == nil {
		r = new(big.Rat)
	}
	return r.Set(s.val)
}

func seq(a, b Scalar) bool {
	return a.val.Cmp(b.val) == 0
}

func smin(a, b Scalar) Scalar {
	if a.val.Cmp(b.val) < 0 {
		return a
	}
	return b
}

func smax(a, b Scalar) Scalar {
	if a.val.Cmp(b.val) > 0 {
		return a
	}
	return b
}

func sadd(a, b Scalar) Scalar {
	return Scalar{new(big.Rat).Add(a.val, b.val)}
}

func ssub(a, b Scalar) Scalar {
	return Scalar{new(big.Rat).Sub(a.val, b.val)}
}

func smul(a, b Scalar) Scalar {
	return Scalar{new(big.Rat).Mul(a.val, b.val)}
}

func sdiv(a, b Scalar) Scalar {
	z := new(big.Rat)
	z.Inv(b.val)
	return Scalar{z.Mul(a.val, z)}
}

func sgt(a, b Scalar) bool {
	return a.val.Cmp(b.val) > 0
}

func slt(a, b Scalar) bool {
	return a.val.Cmp(b.val) < 0
}

func sge(a, b Scalar) bool {
	return a.val.Cmp(b.val) >= 0
}

func sle(a, b Scalar) bool {
	return a.val.Cmp(b.val) <= 0
}
