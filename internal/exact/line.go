package bentlyottmann

import (
	"fmt"
	"math/big"
)

type XY struct {
	X, Y float64
}

type Line struct {
	A, B XY
}

type Intersection struct {
	Empty bool
	A, B  XY
}

var (
	zero = new(big.Rat)
	one  = new(big.Rat).SetFloat64(1.0)
)

func LineIntersection(lineA, lineB Line) Intersection {
	// Algorithm from https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection

	if lineA.A == lineA.B || lineB.A == lineB.B {
		panic("invalid line")
	}

	x1 := new(big.Rat).SetFloat64(lineA.A.X)
	y1 := new(big.Rat).SetFloat64(lineA.A.Y)
	x2 := new(big.Rat).SetFloat64(lineA.B.X)
	y2 := new(big.Rat).SetFloat64(lineA.B.Y)
	x3 := new(big.Rat).SetFloat64(lineB.A.X)
	y3 := new(big.Rat).SetFloat64(lineB.A.Y)
	x4 := new(big.Rat).SetFloat64(lineB.B.X)
	y4 := new(big.Rat).SetFloat64(lineB.B.Y)

	// d := (x1-x2)*(y3-y4)-(y1-y2)*(x3-x4)
	subX1X2 := sub(x1, x2)
	subY3Y4 := sub(y3, y4)
	subY1Y2 := sub(y1, y2)
	subX3X4 := sub(x3, x4)
	tmp1 := mul(subX1X2, subY3Y4)
	tmp2 := mul(subY1Y2, subX3X4)
	d := sub(tmp1, tmp2)

	if d.Cmp(new(big.Rat)) == 0 {
		if !collinear(x1, y1, x2, y2, x3, y3) {
			return Intersection{Empty: true}
		}

		if xyLess(x2, y2, x1, y1) {
			x1, x2 = x2, x1
			y1, y2 = y2, y1
		}
		if xyLess(x4, y4, x3, y3) {
			x3, x4 = x4, x3
			y3, y4 = y4, y3
		}

		fmt.Println("x1", x1)
		fmt.Println("y1", y1)
		fmt.Println("x2", x2)
		fmt.Println("y2", y2)
		fmt.Println("x3", x3)
		fmt.Println("y3", y3)
		fmt.Println("x4", x4)
		fmt.Println("y4", y4)

		if xyLess(x2, y2, x3, y3) || xyLess(x4, y4, x1, y1) {
			return Intersection{Empty: true}
		}
		xMin, yMin := xyMin(x2, y2, x4, y4)
		xMax, yMax := xyMax(x1, y1, x3, y3)
		fmt.Println("xMin", xMin)
		fmt.Println("yMin", yMin)
		fmt.Println("xMax", xMax)
		fmt.Println("yMax", yMax)
		return Intersection{
			A: ratsToXY(xMax, yMax),
			B: ratsToXY(xMin, yMin),
		}
	}

	// t := [(x1-x3)*(y3-y4)-(y1-y3)*(x3-x4)] / d
	subX1X3 := sub(x1, x3)
	subY1Y3 := sub(y1, y3)
	tmp3 := mul(subX1X3, subY3Y4)
	tmp4 := mul(subY1Y3, subX3X4)
	tmp5 := sub(tmp3, tmp4)
	t := div(tmp5, d)

	// u := [(x2-x1)*(y1-y3)-(y2-y1)*(x1-x3)] / d
	subX2X1 := sub(x2, x1)
	subY2Y1 := sub(y2, y1)
	tmp6 := mul(subX2X1, subY1Y3)
	tmp7 := mul(subY2Y1, subX1X3)
	tmp8 := sub(tmp6, tmp7)
	u := div(tmp8, d)

	if t.Cmp(zero) >= 0 && t.Cmp(one) <= 0 && u.Cmp(zero) >= 0 && u.Cmp(one) <= 0 {
		tmp10 := mul(subX2X1, t)
		tmp11 := add(x1, tmp10)
		tmp12 := mul(subY2Y1, t)
		tmp13 := add(y1, tmp12)
		pt := ratsToXY(tmp11, tmp13)
		return Intersection{A: pt, B: pt}
	}

	return Intersection{Empty: true}
}

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

func xyLess(x1, y1, x2, y2 *big.Rat) bool {
	if xCmp := x1.Cmp(x2); xCmp != 0 {
		return xCmp < 0
	}
	return y1.Cmp(y2) < 0
}

func xyMax(x1, y1, x2, y2 *big.Rat) (*big.Rat, *big.Rat) {
	if xyLess(x1, y1, x2, y2) {
		return x2, y2
	}
	return x1, y1
}

func xyMin(x1, y1, x2, y2 *big.Rat) (*big.Rat, *big.Rat) {
	if xyLess(x2, y2, x1, y1) {
		return x2, y2
	}
	return x1, y1
}

func ratsToXY(x, y *big.Rat) XY {
	xf, _ := x.Float64()
	yf, _ := y.Float64()
	return XY{xf, yf}
}

func collinear(x1, y1, x2, y2, x3, y3 *big.Rat) bool {
	sub2X1X := sub(x2, x1)
	sub2Y1Y := sub(y2, y1)
	sub3X2X := sub(x3, x2)
	sub3Y2Y := sub(y3, y2)
	return cross(sub2X1X, sub2Y1Y, sub3X2X, sub3Y2Y).Cmp(zero) == 0
}

func cross(x1, y1, x2, y2 *big.Rat) *big.Rat {
	tmp1 := mul(x1, y2)
	tmp2 := mul(y1, x2)
	return sub(tmp1, tmp2)
}
