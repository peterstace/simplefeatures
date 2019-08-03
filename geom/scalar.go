package geom

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
)

// Scalar represents a rational number.
type Scalar struct {
	val *big.Rat
}

var (
	zero = Scalar{val: big.NewRat(0, 1)}
	one  = Scalar{val: big.NewRat(1, 1)}
	half = Scalar{val: big.NewRat(1, 2)}
)

// NewScalarS parses a string and returns the corresponding scalar. The string
// can be a decimal representation, or a fractional representation (the same as
// big.Rat.SetString).
func NewScalarS(s string) (Scalar, error) {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return Scalar{}, errors.New("invalid scalar")
	}
	return Scalar{r}, nil
}

// MustNewScalarS parses a string by calling NewScalarS. It panics if the
// string cannot be parsed.
func MustNewScalarS(s string) Scalar {
	scal, err := NewScalarS(s)
	if err != nil {
		panic("MustNewScalarS: " + err.Error())
	}
	return scal
}

// NewScalarF creates a scalar that exactly equals the provided float64. It
// gives an error if f is NaN or infinite.
func NewScalarF(f float64) (Scalar, error) {
	z := new(big.Rat).SetFloat64(f)
	if z == nil {
		return Scalar{}, fmt.Errorf("f must be finite: %v", f)
	}
	return Scalar{z}, nil
}

// MustNewScalarF creates a scalar that exactly equals the provided float64.
// It panics if f is NaN or infinite.
func MustNewScalarF(f float64) Scalar {
	scal, err := NewScalarF(f)
	if err != nil {
		panic("MustNewScalarF: " + err.Error())
	}
	return scal
}

func (s Scalar) String() string {
	return s.val.RatString()
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

func (s Scalar) Equals(o Scalar) bool {
	return s.val.Cmp(o.val) == 0
}

func (s Scalar) Min(o Scalar) Scalar {
	if s.val.Cmp(o.val) < 0 {
		return s
	}
	return o
}

func (s Scalar) Max(o Scalar) Scalar {
	if s.val.Cmp(o.val) > 0 {
		return s
	}
	return o
}

func (s Scalar) Add(o Scalar) Scalar {
	return Scalar{new(big.Rat).Add(s.val, o.val)}
}

func (s Scalar) Sub(o Scalar) Scalar {
	return Scalar{new(big.Rat).Sub(s.val, o.val)}
}

func (s Scalar) Mul(o Scalar) Scalar {
	return Scalar{new(big.Rat).Mul(s.val, o.val)}
}

func (s Scalar) Div(o Scalar) Scalar {
	z := new(big.Rat)
	z.Inv(o.val)
	z.Mul(z, s.val)
	return Scalar{z}
}

func (s Scalar) GT(o Scalar) bool {
	return s.val.Cmp(o.val) > 0
}

func (s Scalar) LT(o Scalar) bool {
	return s.val.Cmp(o.val) < 0
}

func (s Scalar) GTE(o Scalar) bool {
	return s.val.Cmp(o.val) >= 0
}

func (s Scalar) LTE(o Scalar) bool {
	return s.val.Cmp(o.val) <= 0
}

func (s Scalar) appendAsFloat(dst []byte) []byte {
	return strconv.AppendFloat(dst, s.AsFloat(), 'f', -1, 64)
}
