package java

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

// BigDecimal is java.math.BigDecimal: the value intVal * 10^-scale. intVal
// is never modified after construction.
type BigDecimal struct {
	intVal *big.Int
	scale  int32
}

func (this *BigDecimal) JClass_BigDecimal() *BigDecimal { return this }
func (this *BigDecimal) JClassName() string             { return "java.math.BigDecimal" }

// BigDecimal_isInstance implements instanceof BigDecimal.
func BigDecimal_isInstance(o any) bool { _, ok := o.(*BigDecimal); return ok }

// BigDecimal_fromAny casts o to BigDecimal.
func BigDecimal_fromAny(o any) *BigDecimal { return AnyTo[*BigDecimal](o) }

func newBigDecimal(v *big.Int, scale int32) *BigDecimal {
	return &BigDecimal{intVal: v, scale: scale}
}

// BigDecimal_ZERO is BigDecimal.ZERO.
var BigDecimal_ZERO = newBigDecimal(big.NewInt(0), 0)

// BigDecimal_ONE is BigDecimal.ONE.
var BigDecimal_ONE = newBigDecimal(big.NewInt(1), 0)

// BigDecimal_TEN is BigDecimal.TEN.
var BigDecimal_TEN = newBigDecimal(big.NewInt(10), 0)

// BigDecimal_new_double implements new BigDecimal(double): the exact value
// of the double, with the smallest scale that represents it.
func BigDecimal_new_double(d float64) *BigDecimal {
	if math.IsInf(d, 0) || d != d {
		panic(NumberFormatException_new_String(Lit("Infinite or NaN")))
	}
	b := math.Float64bits(d)
	exponent := int((b >> 52) & 0x7ff)
	var significand uint64
	if exponent == 0 {
		significand = (b & (1<<52 - 1)) << 1
	} else {
		significand = b&(1<<52-1) | 1<<52
	}
	exponent -= 1075
	if significand == 0 {
		return newBigDecimal(big.NewInt(0), 0)
	}
	for significand&1 == 0 {
		significand >>= 1
		exponent++
	}
	v := new(big.Int).SetUint64(significand)
	if b>>63 != 0 {
		v.Neg(v)
	}
	switch {
	case exponent == 0:
		return newBigDecimal(v, 0)
	case exponent < 0:
		p := new(big.Int).Exp(big.NewInt(5), big.NewInt(int64(-exponent)), nil)
		return newBigDecimal(v.Mul(v, p), int32(-exponent))
	}
	return newBigDecimal(v.Lsh(v, uint(exponent)), 0)
}

// BigDecimal_new_int implements new BigDecimal(int).
func BigDecimal_new_int(i int32) *BigDecimal { return newBigDecimal(big.NewInt(int64(i)), 0) }

// BigDecimal_new_long implements new BigDecimal(long).
func BigDecimal_new_long(l int64) *BigDecimal { return newBigDecimal(big.NewInt(l), 0) }

// BigDecimal_new_String implements new BigDecimal(String).
func BigDecimal_new_String(s *String) *BigDecimal {
	if s == nil {
		panic(NullPointerException_new())
	}
	return parseBigDecimal(s.units)
}

// BigDecimal_new_charArr implements new BigDecimal(char[]).
func BigDecimal_new_charArr(cs []uint16) *BigDecimal {
	if cs == nil {
		panic(NullPointerException_new())
	}
	return parseBigDecimal(cs)
}

// BigDecimal_valueOf_double implements BigDecimal.valueOf(double), which is
// new BigDecimal(Double.toString(d)).
func BigDecimal_valueOf_double(d float64) *BigDecimal {
	return BigDecimal_new_String(StrOfDouble(d))
}

// BigDecimal_valueOf_long implements BigDecimal.valueOf(long).
func BigDecimal_valueOf_long(l int64) *BigDecimal { return newBigDecimal(big.NewInt(l), 0) }

// BigDecimal_valueOf_long_int implements BigDecimal.valueOf(long, int).
func BigDecimal_valueOf_long_int(l int64, scale int32) *BigDecimal {
	return newBigDecimal(big.NewInt(l), scale)
}

// parseBigDecimal implements the BigDecimal(char[], int, int, MathContext)
// constructor with unlimited precision, including its error messages.
func parseBigDecimal(in []uint16) *BigDecimal {
	if len(in) == 0 {
		panic(NumberFormatException_new())
	}
	offset, n := 0, len(in)
	isneg := false
	if in[0] == '-' {
		isneg = true
		offset++
		n--
	} else if in[0] == '+' {
		offset++
		n--
	}
	isCompact := n <= 18
	dot := false
	prec := 0
	var scl int64
	var digits []byte
	nonZeroSeen := false
	for ; n > 0; offset, n = offset+1, n-1 {
		c := in[offset]
		d := int32(-1)
		if c >= '0' && c <= '9' {
			d = int32(c - '0')
		} else if Character_isDigit_char(c) {
			d = Character_digit_char_int(c, 10)
		}
		if d >= 0 {
			if d == 0 {
				if prec == 0 {
					prec = 1
				} else if nonZeroSeen {
					prec++
					digits = append(digits, '0')
				}
			} else {
				if prec != 1 || nonZeroSeen {
					prec++
				}
				nonZeroSeen = true
				digits = append(digits, byte('0'+d))
			}
			if dot {
				scl++
			}
			continue
		}
		if c == '.' {
			if dot {
				panic(NumberFormatException_new_String(Lit("Character array contains more than one decimal point.")))
			}
			dot = true
			continue
		}
		if c == 'e' || c == 'E' {
			scl -= parseBigDecimalExp(in, offset, n)
			break
		}
		if isCompact {
			panic(NumberFormatException_new_String(Concat(Lit("Character "), StrOfChar(c),
				Lit(" is neither a decimal digit number, decimal point, nor \"e\" notation exponential mark."))))
		}
		panic(NumberFormatException_new_String(Lit("Character array is missing \"e\" notation exponential mark.")))
	}
	if prec == 0 {
		panic(NumberFormatException_new_String(Lit("No digits found.")))
	}
	if int64(int32(scl)) != scl {
		panic(NumberFormatException_new_String(Lit("Exponent overflow.")))
	}
	v := new(big.Int)
	if len(digits) > 0 {
		v.SetString(string(digits), 10)
	}
	if isneg {
		v.Neg(v)
	}
	return newBigDecimal(v, int32(scl))
}

func parseBigDecimalExp(in []uint16, offset, n int) int64 {
	at := func(i int) uint16 {
		if i >= len(in) {
			panic(NumberFormatException_new())
		}
		return in[i]
	}
	offset++
	c := at(offset)
	n--
	negexp := c == '-'
	if negexp || c == '+' {
		offset++
		c = at(offset)
		n--
	}
	if n <= 0 {
		panic(NumberFormatException_new_String(Lit("No exponent digits.")))
	}
	for n > 10 && (c == '0' || Character_digit_char_int(c, 10) == 0) {
		offset++
		c = at(offset)
		n--
	}
	if n > 10 {
		panic(NumberFormatException_new_String(Lit("Too many nonzero exponent digits.")))
	}
	var exp int64
	for ; ; n-- {
		var v int64
		if c >= '0' && c <= '9' {
			v = int64(c - '0')
		} else {
			v = int64(Character_digit_char_int(c, 10))
			if v < 0 {
				panic(NumberFormatException_new_String(Lit("Not a digit.")))
			}
		}
		exp = exp*10 + v
		if n == 1 {
			break
		}
		offset++
		c = at(offset)
	}
	if negexp {
		exp = -exp
	}
	return exp
}

// bigPow10 returns 10^n.
func bigPow10(n int64) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(n), nil)
}

// checkRoundingMode validates a legacy BigDecimal.ROUND_ constant.
func checkRoundingMode(rm int32) {
	if rm < rmUp || rm > rmUnnecessary {
		panic(IllegalArgumentException_new_String(Lit("Invalid rounding mode")))
	}
}

// divideAndRound returns n/d rounded to an integer with the given rounding
// mode (a RoundingMode ordinal).
func divideAndRound(n, d *big.Int, mode int32) *big.Int {
	if d.Sign() == 0 {
		if n.IsInt64() && n.Int64() != math.MinInt64 {
			panic(ArithmeticException_new_String(Lit("/ by zero")))
		}
		panic(ArithmeticException_new_String(Lit("BigInteger divide by zero")))
	}
	q, r := new(big.Int).QuoRem(n, d, new(big.Int))
	if r.Sign() == 0 {
		return q
	}
	sign := n.Sign() * d.Sign()
	twoR := new(big.Int).Abs(r)
	twoR.Lsh(twoR, 1)
	cmpHalf := twoR.Cmp(new(big.Int).Abs(d))
	var inc bool
	switch mode {
	case rmUp:
		inc = true
	case rmDown:
		inc = false
	case rmCeiling:
		inc = sign > 0
	case rmFloor:
		inc = sign < 0
	case rmHalfUp:
		inc = cmpHalf >= 0
	case rmHalfDown:
		inc = cmpHalf > 0
	case rmHalfEven:
		inc = cmpHalf > 0 || cmpHalf == 0 && q.Bit(0) == 1
	default:
		panic(ArithmeticException_new_String(Lit("Rounding necessary")))
	}
	if inc {
		q.Add(q, big.NewInt(int64(sign)))
	}
	return q
}

// precisionOf returns the number of decimal digits of |v| (1 for zero).
func precisionOf(v *big.Int) int32 {
	if v.Sign() == 0 {
		return 1
	}
	return int32(len(new(big.Int).Abs(v).String()))
}

// roundToPrecision implements BigDecimal's doRound for a MathContext with
// precision prec and the given rounding mode.
func (this *BigDecimal) roundToPrecision(prec int32, mode int32) *BigDecimal {
	if prec <= 0 {
		return this
	}
	v, scale := this.intVal, this.scale
	for {
		drop := precisionOf(v) - prec
		if drop <= 0 {
			break
		}
		v = divideAndRound(v, bigPow10(int64(drop)), mode)
		scale -= drop
	}
	if v == this.intVal {
		return this
	}
	return newBigDecimal(v, scale)
}

func (this *BigDecimal) Signum() int32 { return int32(this.intVal.Sign()) }
func (this *BigDecimal) Scale() int32  { return this.scale }

// Precision implements BigDecimal.precision.
func (this *BigDecimal) Precision() int32 { return precisionOf(this.intVal) }

// Negate implements BigDecimal.negate.
func (this *BigDecimal) Negate() *BigDecimal {
	return newBigDecimal(new(big.Int).Neg(this.intVal), this.scale)
}

// Abs implements BigDecimal.abs.
func (this *BigDecimal) Abs() *BigDecimal {
	if this.intVal.Sign() < 0 {
		return this.Negate()
	}
	return this
}

// Plus implements BigDecimal.plus.
func (this *BigDecimal) Plus() *BigDecimal { return this }

// aligned returns the unscaled values of a and b at their common (larger)
// scale.
func aligned(a, b *BigDecimal) (*big.Int, *big.Int, int32) {
	switch {
	case a.scale == b.scale:
		return a.intVal, b.intVal, a.scale
	case a.scale < b.scale:
		return new(big.Int).Mul(a.intVal, bigPow10(int64(b.scale-a.scale))), b.intVal, b.scale
	}
	return a.intVal, new(big.Int).Mul(b.intVal, bigPow10(int64(a.scale-b.scale))), a.scale
}

// Add_BigDecimal implements BigDecimal.add.
func (this *BigDecimal) Add_BigDecimal(o *BigDecimal) *BigDecimal {
	a, b, s := aligned(this, o)
	return newBigDecimal(new(big.Int).Add(a, b), s)
}

// Subtract_BigDecimal implements BigDecimal.subtract.
func (this *BigDecimal) Subtract_BigDecimal(o *BigDecimal) *BigDecimal {
	a, b, s := aligned(this, o)
	return newBigDecimal(new(big.Int).Sub(a, b), s)
}

// Multiply_BigDecimal implements BigDecimal.multiply.
func (this *BigDecimal) Multiply_BigDecimal(o *BigDecimal) *BigDecimal {
	return newBigDecimal(new(big.Int).Mul(this.intVal, o.intVal), this.scale+o.scale)
}

// Divide_BigDecimal_int implements BigDecimal.divide(BigDecimal, int): the
// quotient at this BigDecimal's scale, rounded with a legacy rounding mode.
func (this *BigDecimal) Divide_BigDecimal_int(o *BigDecimal, roundingMode int32) *BigDecimal {
	return this.Divide_BigDecimal_int_int(o, this.scale, roundingMode)
}

// Divide_BigDecimal_RoundingMode implements BigDecimal.divide(BigDecimal,
// RoundingMode).
func (this *BigDecimal) Divide_BigDecimal_RoundingMode(o *BigDecimal, rm *RoundingMode) *BigDecimal {
	return this.Divide_BigDecimal_int_int(o, this.scale, rm.ordinal)
}

// Divide_BigDecimal_int_RoundingMode implements BigDecimal.divide(BigDecimal,
// int, RoundingMode).
func (this *BigDecimal) Divide_BigDecimal_int_RoundingMode(o *BigDecimal, scale int32, rm *RoundingMode) *BigDecimal {
	return this.Divide_BigDecimal_int_int(o, scale, rm.ordinal)
}

// Divide_BigDecimal_int_int implements BigDecimal.divide(BigDecimal, int,
// int): the quotient at the given scale.
func (this *BigDecimal) Divide_BigDecimal_int_int(o *BigDecimal, scale int32, roundingMode int32) *BigDecimal {
	checkRoundingMode(roundingMode)
	dividend, divisor := this.intVal, o.intVal
	if int64(scale)+int64(o.scale) > int64(this.scale) {
		dividend = new(big.Int).Mul(dividend, bigPow10(int64(scale)+int64(o.scale)-int64(this.scale)))
	} else {
		divisor = new(big.Int).Mul(divisor, bigPow10(int64(this.scale)-int64(scale)-int64(o.scale)))
	}
	return newBigDecimal(divideAndRound(dividend, divisor, roundingMode), scale)
}

// Divide_BigDecimal implements BigDecimal.divide(BigDecimal): the exact
// quotient, at the scale this.scale() - divisor.scale() if that represents
// it, and otherwise at the smallest scale that does.
func (this *BigDecimal) Divide_BigDecimal(o *BigDecimal) *BigDecimal {
	if o.intVal.Sign() == 0 {
		if this.intVal.Sign() == 0 {
			panic(ArithmeticException_new_String(Lit("Division undefined")))
		}
		panic(ArithmeticException_new_String(Lit("Division by zero")))
	}
	preferred := this.scale - o.scale
	if this.intVal.Sign() == 0 {
		return newBigDecimal(big.NewInt(0), preferred)
	}
	// this/o = (a/b) * 10^-(preferred); reduce a/b and require b = 2^i 5^j.
	a := new(big.Int).Set(this.intVal)
	b := new(big.Int).Set(o.intVal)
	g := new(big.Int).GCD(nil, nil, new(big.Int).Abs(a), new(big.Int).Abs(b))
	a.Quo(a, g)
	b.Quo(b, g)
	if b.Sign() < 0 {
		a.Neg(a)
		b.Neg(b)
	}
	twos, fives := 0, 0
	rest := new(big.Int).Set(b)
	two, five, r := big.NewInt(2), big.NewInt(5), new(big.Int)
	for {
		q, m := new(big.Int).QuoRem(rest, two, r)
		if m.Sign() != 0 {
			break
		}
		rest = q
		twos++
	}
	for {
		q, m := new(big.Int).QuoRem(rest, five, r)
		if m.Sign() != 0 {
			break
		}
		rest = q
		fives++
	}
	if rest.Cmp(big.NewInt(1)) != 0 {
		panic(ArithmeticException_new_String(Lit("Non-terminating decimal expansion; no exact representable decimal result.")))
	}
	k := twos
	if fives > k {
		k = fives
	}
	// a/b = a * 10^k / b / 10^k, and 10^k / b is an integer.
	q := new(big.Int).Mul(a, new(big.Int).Quo(bigPow10(int64(k)), b))
	res := newBigDecimal(q, preferred+int32(k))
	// Remove trailing zeros down to the preferred scale.
	ten := big.NewInt(10)
	for res.scale > preferred {
		qq, m := new(big.Int).QuoRem(res.intVal, ten, r)
		if m.Sign() != 0 {
			break
		}
		res = newBigDecimal(qq, res.scale-1)
	}
	return res
}

// SetScale_int implements BigDecimal.setScale(int), which requires no
// rounding.
func (this *BigDecimal) SetScale_int(newScale int32) *BigDecimal {
	return this.SetScale_int_int(newScale, rmUnnecessary)
}

// SetScale_int_RoundingMode implements BigDecimal.setScale(int, RoundingMode).
func (this *BigDecimal) SetScale_int_RoundingMode(newScale int32, rm *RoundingMode) *BigDecimal {
	return this.SetScale_int_int(newScale, rm.ordinal)
}

// SetScale_int_int implements BigDecimal.setScale(int, int).
func (this *BigDecimal) SetScale_int_int(newScale int32, roundingMode int32) *BigDecimal {
	checkRoundingMode(roundingMode)
	switch {
	case newScale == this.scale:
		return this
	case this.intVal.Sign() == 0:
		return newBigDecimal(big.NewInt(0), newScale)
	case newScale > this.scale:
		return newBigDecimal(new(big.Int).Mul(this.intVal, bigPow10(int64(newScale)-int64(this.scale))), newScale)
	}
	return newBigDecimal(divideAndRound(this.intVal, bigPow10(int64(this.scale)-int64(newScale)), roundingMode), newScale)
}

// Round_MathContext implements BigDecimal.round(MathContext).
func (this *BigDecimal) Round_MathContext(mc *MathContext) *BigDecimal {
	return this.roundToPrecision(mc.precision, mc.roundingMode.ordinal)
}

// MovePointLeft_int implements BigDecimal.movePointLeft.
func (this *BigDecimal) MovePointLeft_int(n int32) *BigDecimal {
	if n == 0 {
		return this
	}
	r := newBigDecimal(this.intVal, this.scale+n)
	if r.scale < 0 {
		return r.SetScale_int_int(0, rmUnnecessary)
	}
	return r
}

// MovePointRight_int implements BigDecimal.movePointRight.
func (this *BigDecimal) MovePointRight_int(n int32) *BigDecimal {
	if n == 0 {
		return this
	}
	r := newBigDecimal(this.intVal, this.scale-n)
	if r.scale < 0 {
		return r.SetScale_int_int(0, rmUnnecessary)
	}
	return r
}

// ScaleByPowerOfTen_int implements BigDecimal.scaleByPowerOfTen.
func (this *BigDecimal) ScaleByPowerOfTen_int(n int32) *BigDecimal {
	return newBigDecimal(this.intVal, this.scale-n)
}

// StripTrailingZeros implements BigDecimal.stripTrailingZeros.
func (this *BigDecimal) StripTrailingZeros() *BigDecimal {
	if this.intVal.Sign() == 0 {
		return BigDecimal_ZERO
	}
	v, s := this.intVal, this.scale
	ten, r := big.NewInt(10), new(big.Int)
	for {
		q, m := new(big.Int).QuoRem(v, ten, r)
		if m.Sign() != 0 {
			break
		}
		v = q
		s--
	}
	return newBigDecimal(v, s)
}

// Pow_int implements BigDecimal.pow(int).
func (this *BigDecimal) Pow_int(n int32) *BigDecimal {
	if n < 0 || n > 999999999 {
		panic(ArithmeticException_new_String(Lit("Invalid operation")))
	}
	return newBigDecimal(new(big.Int).Exp(this.intVal, big.NewInt(int64(n)), nil), this.scale*n)
}

// Max_BigDecimal implements BigDecimal.max.
func (this *BigDecimal) Max_BigDecimal(o *BigDecimal) *BigDecimal {
	if this.compare(o) >= 0 {
		return this
	}
	return o
}

// Min_BigDecimal implements BigDecimal.min.
func (this *BigDecimal) Min_BigDecimal(o *BigDecimal) *BigDecimal {
	if this.compare(o) <= 0 {
		return this
	}
	return o
}

// Ulp implements BigDecimal.ulp.
func (this *BigDecimal) Ulp() *BigDecimal { return newBigDecimal(big.NewInt(1), this.scale) }

func (this *BigDecimal) compare(o *BigDecimal) int32 {
	a, b, _ := aligned(this, o)
	return int32(a.Cmp(b))
}

// CompareTo_Object implements BigDecimal.compareTo, which ignores the scale.
func (this *BigDecimal) CompareTo_Object(o any) int32 {
	return this.compare(unboxRecv[*BigDecimal](o, "java.math.BigDecimal"))
}

// Equals_Object implements BigDecimal.equals, which requires equal scales.
func (this *BigDecimal) Equals_Object(o any) bool {
	b, ok := o.(*BigDecimal)
	return ok && b != nil && b.scale == this.scale && b.intVal.Cmp(this.intVal) == 0
}

// HashCode implements BigDecimal.hashCode.
func (this *BigDecimal) HashCode() int32 {
	var h int32
	for _, w := range bigMagnitudeWords(this.intVal) {
		h = 31*h + int32(w)
	}
	return 31*h*int32(this.intVal.Sign()) + this.scale
}

// bigMagnitudeWords returns |v| as big-endian 32-bit words, as
// BigInteger.mag holds it.
func bigMagnitudeWords(v *big.Int) []uint32 {
	b := new(big.Int).Abs(v).Bytes()
	for len(b)%4 != 0 {
		b = append([]byte{0}, b...)
	}
	out := make([]uint32, len(b)/4)
	for i := range out {
		out[i] = uint32(b[4*i])<<24 | uint32(b[4*i+1])<<16 | uint32(b[4*i+2])<<8 | uint32(b[4*i+3])
	}
	return out
}

// DoubleValue implements BigDecimal.doubleValue, which rounds correctly.
func (this *BigDecimal) DoubleValue() float64 {
	v, _ := strconv.ParseFloat(this.intVal.String()+"e"+strconv.Itoa(-int(this.scale)), 64)
	return v
}

// FloatValue implements BigDecimal.floatValue, which rounds correctly.
func (this *BigDecimal) FloatValue() float32 {
	v, _ := strconv.ParseFloat(this.intVal.String()+"e"+strconv.Itoa(-int(this.scale)), 32)
	return float32(v)
}

// toBigInteger truncates toward zero.
func (this *BigDecimal) toBigInteger() *big.Int {
	if this.scale <= 0 {
		return new(big.Int).Mul(this.intVal, bigPow10(-int64(this.scale)))
	}
	return new(big.Int).Quo(this.intVal, bigPow10(int64(this.scale)))
}

// LongValue implements BigDecimal.longValue: the low 64 bits of the integer
// part.
func (this *BigDecimal) LongValue() int64 {
	i := this.toBigInteger()
	lo := new(big.Int).And(i, new(big.Int).SetUint64(math.MaxUint64))
	return int64(lo.Uint64())
}

// IntValue implements BigDecimal.intValue: the low 32 bits of the integer
// part.
func (this *BigDecimal) IntValue() int32 { return int32(this.LongValue()) }

// ToString implements BigDecimal.toString, which uses scientific notation
// when the scale is negative or the adjusted exponent is below -6.
func (this *BigDecimal) ToString() *String { return FromGo(this.layout(true)) }

// ToEngineeringString implements BigDecimal.toEngineeringString.
func (this *BigDecimal) ToEngineeringString() *String { return FromGo(this.layout(false)) }

func (this *BigDecimal) layout(sci bool) string {
	if this.scale == 0 {
		return this.intVal.String()
	}
	coeff := new(big.Int).Abs(this.intVal).String()
	var buf strings.Builder
	if this.intVal.Sign() < 0 {
		buf.WriteByte('-')
	}
	coeffLen := len(coeff)
	adjusted := -int64(this.scale) + int64(coeffLen-1)
	if this.scale >= 0 && adjusted >= -6 {
		pad := int(this.scale) - coeffLen
		if pad >= 0 {
			buf.WriteString("0.")
			buf.WriteString(strings.Repeat("0", pad))
			buf.WriteString(coeff)
		} else {
			buf.WriteString(coeff[:-pad])
			buf.WriteByte('.')
			buf.WriteString(coeff[-pad:])
		}
		return buf.String()
	}
	if sci {
		buf.WriteByte(coeff[0])
		if coeffLen > 1 {
			buf.WriteByte('.')
			buf.WriteString(coeff[1:])
		}
	} else {
		sig := int(adjusted % 3)
		if sig < 0 {
			sig += 3
		}
		adjusted -= int64(sig)
		sig++
		switch {
		case this.intVal.Sign() == 0:
			switch sig {
			case 1:
				buf.WriteByte('0')
			case 2:
				buf.WriteString("0.00")
				adjusted += 3
			case 3:
				buf.WriteString("0.0")
				adjusted += 3
			}
		case sig >= coeffLen:
			buf.WriteString(coeff)
			buf.WriteString(strings.Repeat("0", sig-coeffLen))
		default:
			buf.WriteString(coeff[:sig])
			buf.WriteByte('.')
			buf.WriteString(coeff[sig:])
		}
	}
	if adjusted != 0 {
		buf.WriteByte('E')
		if adjusted > 0 {
			buf.WriteByte('+')
		}
		buf.WriteString(strconv.FormatInt(adjusted, 10))
	}
	return buf.String()
}

// ToPlainString implements BigDecimal.toPlainString.
func (this *BigDecimal) ToPlainString() *String {
	if this.scale == 0 {
		return FromGo(this.intVal.String())
	}
	if this.scale < 0 {
		if this.intVal.Sign() == 0 {
			return Lit("0")
		}
		return FromGo(this.intVal.String() + strings.Repeat("0", int(-this.scale)))
	}
	s := new(big.Int).Abs(this.intVal).String()
	sign := ""
	if this.intVal.Sign() < 0 {
		sign = "-"
	}
	insertionPoint := len(s) - int(this.scale)
	switch {
	case insertionPoint == 0:
		return FromGo(sign + "0." + s)
	case insertionPoint > 0:
		return FromGo(sign + s[:insertionPoint] + "." + s[insertionPoint:])
	}
	return FromGo(sign + "0." + strings.Repeat("0", -insertionPoint) + s)
}

// ---------------------------------------------------------------------------
// MathContext

// MathContext is java.math.MathContext.
type MathContext struct {
	precision    int32
	roundingMode *RoundingMode
}

func (this *MathContext) JClass_MathContext() *MathContext { return this }
func (this *MathContext) JClassName() string               { return "java.math.MathContext" }

// MathContext_isInstance implements instanceof MathContext.
func MathContext_isInstance(o any) bool { _, ok := o.(*MathContext); return ok }

// MathContext_fromAny casts o to MathContext.
func MathContext_fromAny(o any) *MathContext { return AnyTo[*MathContext](o) }

// MathContext_new_int implements new MathContext(int), which rounds HALF_UP.
func MathContext_new_int(precision int32) *MathContext {
	return MathContext_new_int_RoundingMode(precision, RoundingMode_HALF_UP)
}

// MathContext_new_int_RoundingMode implements new MathContext(int,
// RoundingMode).
func MathContext_new_int_RoundingMode(precision int32, rm *RoundingMode) *MathContext {
	if precision < 0 {
		panic(IllegalArgumentException_new_String(Lit("Digits < 0")))
	}
	if rm == nil {
		panic(NullPointerException_new_String(Lit("null RoundingMode")))
	}
	return &MathContext{precision: precision, roundingMode: rm}
}

// MathContext_UNLIMITED is MathContext.UNLIMITED.
var MathContext_UNLIMITED = &MathContext{precision: 0, roundingMode: RoundingMode_HALF_UP}

// MathContext_DECIMAL32 is MathContext.DECIMAL32.
var MathContext_DECIMAL32 = &MathContext{precision: 7, roundingMode: RoundingMode_HALF_EVEN}

// MathContext_DECIMAL64 is MathContext.DECIMAL64.
var MathContext_DECIMAL64 = &MathContext{precision: 16, roundingMode: RoundingMode_HALF_EVEN}

// MathContext_DECIMAL128 is MathContext.DECIMAL128.
var MathContext_DECIMAL128 = &MathContext{precision: 34, roundingMode: RoundingMode_HALF_EVEN}

func (this *MathContext) GetPrecision() int32            { return this.precision }
func (this *MathContext) GetRoundingMode() *RoundingMode { return this.roundingMode }

// ToString implements MathContext.toString.
func (this *MathContext) ToString() *String {
	return Concat(Lit("precision="), StrOfInt(this.precision), Lit(" roundingMode="), this.roundingMode.name)
}

// Equals_Object implements MathContext.equals.
func (this *MathContext) Equals_Object(o any) bool {
	m, ok := o.(*MathContext)
	return ok && m != nil && m.precision == this.precision && m.roundingMode == this.roundingMode
}

// HashCode implements MathContext.hashCode.
func (this *MathContext) HashCode() int32 {
	return this.roundingMode.HashCode() + this.precision<<2
}
