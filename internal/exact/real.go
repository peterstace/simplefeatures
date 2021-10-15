package exact

import "math/big"

var (
	one = new(big.Rat).SetFloat64(1.0)
)

func sub(a, b *big.Rat) *big.Rat {
	tmp := new(big.Rat).Set(a)
	tmp.Sub(tmp, b)
	return tmp
}

func mul(a, b *big.Rat) *big.Rat {
	tmp := new(big.Rat).Set(a)
	tmp.Mul(tmp, b)
	return tmp
}

func div(a, b *big.Rat) *big.Rat {
	tmp := new(big.Rat).Set(b)
	tmp.Inv(tmp)
	tmp.Mul(tmp, a)
	return tmp
}

func add(a, b *big.Rat) *big.Rat {
	tmp := new(big.Rat).Set(a)
	tmp.Add(tmp, b)
	return tmp
}

func neg(r *big.Rat) *big.Rat {
	tmp := new(big.Rat).Set(r)
	tmp.Neg(tmp)
	return tmp
}

func inUnitInterval(r *big.Rat) bool {
	return r.Sign() >= 0 && r.Cmp(one) <= 0
}
