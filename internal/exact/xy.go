package exact

import "math/big"

type XY64 struct {
	X, Y float64
}

func (f XY64) ToRat() XYRat {
	return XYRat{
		new(big.Rat).SetFloat64(f.X),
		new(big.Rat).SetFloat64(f.Y),
	}
}

type XYRat struct {
	X, Y *big.Rat
}

func (r XYRat) Less(o XYRat) bool {
	if xCmp := r.X.Cmp(o.X); xCmp != 0 {
		return xCmp < 0
	}
	return r.Y.Cmp(o.Y) < 0
}

func (r XYRat) Sub(o XYRat) XYRat {
	return XYRat{
		sub(r.X, o.X),
		sub(r.Y, o.Y),
	}
}

func (r XYRat) Add(o XYRat) XYRat {
	return XYRat{
		add(r.X, o.X),
		add(r.Y, o.Y),
	}
}

func (r XYRat) Neg() XYRat {
	return XYRat{
		neg(r.X),
		neg(r.Y),
	}
}

func (r XYRat) Max(o XYRat) XYRat {
	if r.Less(o) {
		return o
	}
	return r
}

func (r XYRat) Min(o XYRat) XYRat {
	if r.Less(o) {
		return r
	}
	return o
}

func (r XYRat) ToXY64() XY64 {
	xf, _ := r.X.Float64()
	yf, _ := r.Y.Float64()
	return XY64{xf, yf}

}

func (r XYRat) Cross(o XYRat) *big.Rat {
	tmp1 := mul(r.X, o.Y)
	tmp2 := mul(r.Y, o.X)
	return sub(tmp1, tmp2)
}

func (r XYRat) Scale(s *big.Rat) XYRat {
	return XYRat{
		mul(r.X, s),
		mul(r.Y, s),
	}
}
