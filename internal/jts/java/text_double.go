package java

import (
	"math"
	"math/big"
	"math/bits"
	"strconv"
)

// This file converts between binary floating-point values and decimal
// strings as the JDK does:
//
//   - Double.toString and Float.toString use Raffaello Giulietti's Schubfach
//     algorithm (JDK 19 and later), which selects the shortest decimal that
//     rounds to the value, preferring the one closest to the value. The port
//     follows jdk.internal.math.DoubleToDecimal and FloatToDecimal.
//   - Double.parseDouble and Float.parseFloat accept the grammar of
//     jdk.internal.math.FloatingDecimal.readJavaFormatString and round
//     correctly, as FloatingDecimal does.
//   - DecimalFormat uses FloatingDecimal's binary to decimal conversion
//     (BinaryToASCIIBuffer.dtoa), which does not always select the shortest
//     decimal. fdConvert ports it.

// DoubleToString formats d as Java's Double.toString does.
func DoubleToString(d float64) string {
	b := math.Float64bits(d)
	switch {
	case d != d:
		return "NaN"
	case math.IsInf(d, 1):
		return "Infinity"
	case math.IsInf(d, -1):
		return "-Infinity"
	case b == 0:
		return "0.0"
	case b == 1<<63:
		return "-0.0"
	}
	f, e := doubleToDecimal(b)
	return javaDecimalString(b>>63 != 0, f, e)
}

// FloatToString formats f as Java's Float.toString does.
func FloatToString(f float32) string {
	b := math.Float32bits(f)
	switch {
	case f != f:
		return "NaN"
	case math.IsInf(float64(f), 1):
		return "Infinity"
	case math.IsInf(float64(f), -1):
		return "-Infinity"
	case b == 0:
		return "0.0"
	case b == 1<<31:
		return "-0.0"
	}
	c, e := floatToDecimal(b)
	return javaDecimalString(b>>31 != 0, int64(c), e)
}

// javaDecimalString renders the decimal f*10^e (f > 0) in the format of
// Double.toString: plain notation with at least one fractional digit when
// 10^-3 <= |v| < 10^7, otherwise computerized scientific notation.
func javaDecimalString(neg bool, f int64, e int) string {
	digits := strconv.FormatInt(f, 10)
	// v = 0.DIGITS * 10^exp.
	exp := e + len(digits)
	n := len(digits)
	for n > 1 && digits[n-1] == '0' {
		n--
	}
	digits = digits[:n]
	buf := make([]byte, 0, 32)
	if neg {
		buf = append(buf, '-')
	}
	switch {
	case 0 < exp && exp <= 7:
		if len(digits) <= exp {
			buf = append(buf, digits...)
			for i := len(digits); i < exp; i++ {
				buf = append(buf, '0')
			}
			buf = append(buf, '.', '0')
		} else {
			buf = append(buf, digits[:exp]...)
			buf = append(buf, '.')
			buf = append(buf, digits[exp:]...)
		}
	case -3 < exp && exp <= 0:
		buf = append(buf, '0', '.')
		for i := exp; i < 0; i++ {
			buf = append(buf, '0')
		}
		buf = append(buf, digits...)
	default:
		buf = append(buf, digits[0], '.')
		if len(digits) > 1 {
			buf = append(buf, digits[1:]...)
		} else {
			buf = append(buf, '0')
		}
		buf = append(buf, 'E')
		buf = strconv.AppendInt(buf, int64(exp-1), 10)
	}
	return string(buf)
}

// ---------------------------------------------------------------------------
// Schubfach

const (
	sfKMin = -324
	sfKMax = 292
)

// sfG holds, for each k in [sfKMin, sfKMax], the 126-bit value
// g = floor(10^-k / 2^r) + 1, where r is chosen so that 2^125 <= 10^-k / 2^r
// < 2^126, split into its upper and lower 63 bits. It is computed once at
// package initialization and only read afterwards.
var sfG = computeSchubfachTable()

func computeSchubfachTable() [][2]int64 {
	out := make([][2]int64, sfKMax-sfKMin+1)
	ten := big.NewInt(10)
	mask63 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 63), big.NewInt(1))
	for k := sfKMin; k <= sfKMax; k++ {
		r := flog2pow10(-k) - 125
		var beta big.Int
		if k <= 0 {
			beta.Exp(ten, big.NewInt(int64(-k)), nil)
			if r <= 0 {
				beta.Lsh(&beta, uint(-r))
			} else {
				beta.Rsh(&beta, uint(r))
			}
		} else {
			// r < 0 here.
			var den big.Int
			den.Exp(ten, big.NewInt(int64(k)), nil)
			beta.Lsh(big.NewInt(1), uint(-r))
			beta.Quo(&beta, &den)
		}
		beta.Add(&beta, big.NewInt(1))
		var lo big.Int
		lo.And(&beta, mask63)
		beta.Rsh(&beta, 63)
		out[k-sfKMin] = [2]int64{beta.Int64(), lo.Int64()}
	}
	return out
}

// flog10pow2 returns floor(log10(2^e)).
func flog10pow2(e int) int {
	return int(int64(e) * 661_971_961_083 >> 41)
}

// flog10threeQuartersPow2 returns floor(log10(3/4 * 2^e)).
func flog10threeQuartersPow2(e int) int {
	return int((int64(e)*661_971_961_083 + -274_743_187_321) >> 41)
}

// flog2pow10 returns floor(log2(10^e)).
func flog2pow10(e int) int {
	return int(int64(e) * 913_124_641_741 >> 38)
}

// multiplyHigh is Java's Math.multiplyHigh: the high 64 bits of the signed
// 128-bit product.
func multiplyHigh(a, b int64) int64 {
	hi, _ := bits.Mul64(uint64(a), uint64(b))
	h := int64(hi)
	if a < 0 {
		h -= b
	}
	if b < 0 {
		h -= a
	}
	return h
}

// doubleToDecimal returns f and e such that f*10^e is the decimal that
// Double.toString selects for the finite non-zero double with bits b
// (ignoring the sign).
func doubleToDecimal(b uint64) (int64, int) {
	const (
		p     = 53
		qMin  = -1074
		cTiny = 3
		cMin  = int64(1) << (p - 1)
	)
	t := int64(b & (1<<(p-1) - 1))
	bq := int(b>>(p-1)) & (1<<11 - 1)
	if bq != 0 {
		mq := -qMin + 1 - bq
		c := cMin | t
		if 0 < mq && mq < p {
			f := c >> uint(mq)
			if f<<uint(mq) == c {
				return f, 0
			}
		}
		return schubfachDouble(-mq, c, 0)
	}
	if t < cTiny {
		return schubfachDouble(qMin, 10*t, -1)
	}
	return schubfachDouble(qMin, t, 0)
}

func schubfachDouble(q int, c int64, dk int) (int64, int) {
	const (
		qMin   = -1074
		cMin   = int64(1) << 52
		mask63 = math.MaxInt64
	)
	out := c & 1
	cb := c << 2
	cbr := cb + 2
	var cbl int64
	var k int
	if c != cMin || q == qMin {
		cbl = cb - 2
		k = flog10pow2(q)
	} else {
		cbl = cb - 1
		k = flog10threeQuartersPow2(q)
	}
	h := uint(q + flog2pow10(-k) + 2)
	g := sfG[k-sfKMin]
	g1, g0 := g[0], g[1]
	rop := func(cp int64) int64 {
		x1 := multiplyHigh(g0, cp)
		y0 := g1 * cp
		y1 := multiplyHigh(g1, cp)
		z := int64(uint64(y0)>>1) + x1
		vbp := y1 + int64(uint64(z)>>63)
		return vbp | int64(uint64((z&mask63)+mask63)>>63)
	}
	vb := rop(cb << h)
	vbl := rop(cbl << h)
	vbr := rop(cbr << h)
	s := vb >> 2
	if s >= 100 {
		sp10 := 10 * multiplyHigh(s, 115_292_150_460_684_698<<4)
		tp10 := sp10 + 10
		upin := vbl+out <= sp10<<2
		wpin := (tp10<<2)+out <= vbr
		if upin != wpin {
			if upin {
				return sp10, k
			}
			return tp10, k
		}
	}
	t := s + 1
	uin := vbl+out <= s<<2
	win := (t<<2)+out <= vbr
	if uin != win {
		if uin {
			return s, k + dk
		}
		return t, k + dk
	}
	cmp := vb - (s+t)<<1
	if cmp < 0 || cmp == 0 && s&1 == 0 {
		return s, k + dk
	}
	return t, k + dk
}

// floatToDecimal is doubleToDecimal for the float with bits b.
func floatToDecimal(b uint32) (int32, int) {
	const (
		p     = 24
		qMin  = -149
		cTiny = 8
		cMin  = int32(1) << (p - 1)
	)
	t := int32(b & (1<<(p-1) - 1))
	bq := int(b>>(p-1)) & (1<<8 - 1)
	if bq != 0 {
		mq := -qMin + 1 - bq
		c := cMin | t
		if 0 < mq && mq < p {
			f := c >> uint(mq)
			if f<<uint(mq) == c {
				return f, 0
			}
		}
		return schubfachFloat(-mq, c, 0)
	}
	if t < cTiny {
		return schubfachFloat(qMin, 10*t, -1)
	}
	return schubfachFloat(qMin, t, 0)
}

func schubfachFloat(q int, c int32, dk int) (int32, int) {
	const (
		qMin   = -149
		cMin   = int32(1) << 23
		mask32 = int64(1)<<32 - 1
	)
	out := c & 1
	cb := int64(c << 2)
	cbr := cb + 2
	var cbl int64
	var k int
	if c != cMin || q == qMin {
		cbl = cb - 2
		k = flog10pow2(q)
	} else {
		cbl = cb - 1
		k = flog10threeQuartersPow2(q)
	}
	h := uint(q + flog2pow10(-k) + 33)
	g := sfG[k-sfKMin][0] + 1
	rop := func(cp int64) int32 {
		x1 := multiplyHigh(g, cp)
		vbp := int64(uint64(x1) >> 31)
		return int32(vbp | int64(uint64((x1&mask32)+mask32)>>32))
	}
	vb := rop(cb << h)
	vbl := rop(cbl << h)
	vbr := rop(cbr << h)
	s := vb >> 2
	if s >= 100 {
		sp10 := 10 * int32(uint64(int64(s)*1_717_986_919)>>34)
		tp10 := sp10 + 10
		upin := vbl+out <= sp10<<2
		wpin := (tp10<<2)+out <= vbr
		if upin != wpin {
			if upin {
				return sp10, k
			}
			return tp10, k
		}
	}
	t := s + 1
	uin := vbl+out <= s<<2
	win := (t<<2)+out <= vbr
	if uin != win {
		if uin {
			return s, k + dk
		}
		return t, k + dk
	}
	cmp := vb - (s+t)<<1
	if cmp < 0 || cmp == 0 && s&1 == 0 {
		return s, k + dk
	}
	return t, k + dk
}

// ---------------------------------------------------------------------------
// Parsing

// javaFloatLiteral is the result of scanning a string with Java's
// floating-point grammar.
type javaFloatLiteral struct {
	special   float64 // NaN or an infinity, when isSpecial
	isSpecial bool
	neg       bool
	zero      bool
	hex       bool
	// For a decimal literal, the value is 0.digits * 10^decExp.
	digits string
	decExp int
	// For a hexadecimal literal: the significand's digits without leading
	// zeros; the number of them before the point, or, when there are none,
	// minus the number of zeros between the point and the first digit; and
	// the binary exponent, with hexExpHuge set when it does not fit in an
	// int.
	hexSig     string
	hexIntLen  int
	hexExp     int64
	hexExpHuge bool
	hexExpNeg  bool
}

// parseDouble implements Double.parseDouble.
func parseDouble(s *String) float64 {
	lit := scanJavaFloat(s)
	switch {
	case lit.isSpecial:
		return lit.special
	case lit.zero:
		if lit.neg {
			return math.Copysign(0, -1)
		}
		return 0
	case lit.hex:
		v, _ := hexToFloat(lit)
		return v
	}
	v, _ := strconv.ParseFloat("0."+lit.digits+"e"+strconv.Itoa(lit.decExp), 64)
	if lit.neg {
		v = -v
	}
	return v
}

// parseFloat implements Float.parseFloat. The decimal value is rounded to a
// float directly, not through a double.
func parseFloat(s *String) float32 {
	lit := scanJavaFloat(s)
	switch {
	case lit.isSpecial && lit.special != lit.special:
		return math.Float32frombits(0x7fc00000)
	case lit.isSpecial:
		return float32(lit.special)
	case lit.zero:
		if lit.neg {
			return float32(math.Copysign(0, -1))
		}
		return 0
	case lit.hex:
		_, f := hexToFloat(lit)
		return f
	}
	v, _ := strconv.ParseFloat("0."+lit.digits+"e"+strconv.Itoa(lit.decExp), 32)
	f := float32(v)
	if lit.neg {
		f = -f
	}
	return f
}

// scanJavaFloat implements FloatingDecimal.readJavaFormatString, throwing
// NumberFormatException with the JDK's messages.
func scanJavaFloat(str *String) javaFloatLiteral {
	if str == nil {
		panic(NullPointerException_new())
	}
	in := str.Trim().units
	fail := func() {
		panic(NumberFormatException_new_String(Concat(Lit("For input string: \""), FromUnits(in), Lit("\""))))
	}
	n := len(in)
	if n == 0 {
		panic(NumberFormatException_new_String(Lit("empty String")))
	}
	var lit javaFloatLiteral
	signSeen := false
	i := 0
	// at returns in[j], or fails as Java's String.charAt failure does.
	at := func(j int) uint16 {
		if j >= n {
			fail()
		}
		return in[j]
	}
	switch in[0] {
	case '-':
		lit.neg = true
		i++
		signSeen = true
	case '+':
		i++
		signSeen = true
	}
	c := at(i)
	switch c {
	case 'N':
		if n-i == 3 && in[i+1] == 'a' && in[i+2] == 'N' {
			lit.isSpecial = true
			lit.special = math.Float64frombits(0x7ff8000000000000)
			return lit
		}
		fail()
	case 'I':
		if n-i == 8 && FromUnits(in[i:]).Go() == "Infinity" {
			lit.isSpecial = true
			lit.special = math.Inf(1)
			if lit.neg {
				lit.special = math.Inf(-1)
			}
			return lit
		}
		fail()
	case '0':
		if n > i+1 && (in[i+1] == 'x' || in[i+1] == 'X') {
			return scanJavaHexFloat(in)
		}
	}
	multiplePoints := func() {
		panic(NumberFormatException_new_String(Lit("multiple points")))
	}
	digits := make([]byte, 0, n)
	decSeen := false
	decPt := 0
	nLeadZero := 0
	nTrailZero := 0
skipLeadingZerosLoop:
	for ; i < n; i++ {
		c = in[i]
		switch {
		case c == '0':
			nLeadZero++
		case c == '.':
			if decSeen {
				multiplePoints()
			}
			decPt = i
			if signSeen {
				decPt--
			}
			decSeen = true
		default:
			break skipLeadingZerosLoop
		}
	}
digitLoop:
	for ; i < n; i++ {
		c = in[i]
		switch {
		case c >= '1' && c <= '9':
			digits = append(digits, byte(c))
			nTrailZero = 0
		case c == '0':
			digits = append(digits, '0')
			nTrailZero++
		case c == '.':
			if decSeen {
				multiplePoints()
			}
			decPt = i
			if signSeen {
				decPt--
			}
			decSeen = true
		default:
			break digitLoop
		}
	}
	nDigits := len(digits) - nTrailZero
	isZero := nDigits == 0
	if isZero && nLeadZero == 0 {
		fail()
	}
	var decExp int
	if decSeen {
		decExp = decPt - nLeadZero
	} else {
		decExp = nDigits + nTrailZero
	}
	if i < n {
		if c = in[i]; c == 'e' || c == 'E' {
			expSign := 1
			expVal := 0
			const reallyBig = math.MaxInt32 / 10
			expOverflow := false
			i++
			switch at(i) {
			case '-':
				expSign = -1
				i++
			case '+':
				i++
			}
			expAt := i
			for i < n {
				if expVal >= reallyBig {
					expOverflow = true
				}
				c = in[i]
				i++
				if c >= '0' && c <= '9' {
					expVal = int(int32(expVal*10 + int(c-'0')))
				} else {
					i--
					break
				}
			}
			expLimit := 324 + nDigits + nTrailZero
			if expOverflow || expVal > expLimit {
				if !expOverflow && expSign == 1 && decExp < 0 && expVal+decExp < expLimit {
					decExp += expVal
				} else {
					decExp = expSign * expLimit
				}
			} else {
				decExp += expSign * expVal
			}
			if i == expAt {
				fail()
			}
		}
	}
	if i < n && (i != n-1 || in[i] != 'f' && in[i] != 'F' && in[i] != 'd' && in[i] != 'D') {
		fail()
	}
	if isZero {
		lit.zero = true
		return lit
	}
	lit.digits = string(digits[:nDigits])
	lit.decExp = decExp
	return lit
}

// scanJavaHexFloat matches in against Java's hexadecimal floating-point
// grammar: [+-]?0[xX](H+\.?|H*\.H+)[pP][+-]?D+[fFdD]?.
func scanJavaHexFloat(in []uint16) javaFloatLiteral {
	fail := func() {
		panic(NumberFormatException_new_String(Concat(Lit("For input string: \""), FromUnits(in), Lit("\""))))
	}
	isHex := func(c uint16) bool {
		return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
	}
	var lit javaFloatLiteral
	lit.hex = true
	i, n := 0, len(in)
	if in[0] == '-' || in[0] == '+' {
		lit.neg = in[0] == '-'
		i++
	}
	i += 2 // "0x"
	intStart := i
	for i < n && isHex(in[i]) {
		i++
	}
	intEnd := i
	fracStart, fracEnd := i, i
	hasPoint := false
	if i < n && in[i] == '.' {
		hasPoint = true
		i++
		fracStart = i
		for i < n && isHex(in[i]) {
			i++
		}
		fracEnd = i
	}
	if intEnd == intStart && (!hasPoint || fracEnd == fracStart) {
		fail()
	}
	if i >= n || in[i] != 'p' && in[i] != 'P' {
		fail()
	}
	i++
	if i < n && (in[i] == '+' || in[i] == '-') {
		lit.hexExpNeg = in[i] == '-'
		i++
	}
	expStart := i
	for i < n && in[i] >= '0' && in[i] <= '9' {
		i++
	}
	expEnd := i
	if expEnd == expStart {
		fail()
	}
	if i < n && (i != n-1 || in[i] != 'f' && in[i] != 'F' && in[i] != 'd' && in[i] != 'D') {
		fail()
	}
	intDigits := stripLeadingZeros(FromUnits(in[intStart:intEnd]).Go())
	fracDigits := FromUnits(in[fracStart:fracEnd]).Go()
	sig := stripLeadingZeros(intDigits + fracDigits)
	if sig == "" {
		lit.zero = true
		return lit
	}
	lit.hexSig = sig
	if len(intDigits) >= 1 {
		lit.hexIntLen = len(intDigits)
	} else {
		lit.hexIntLen = -(len(fracDigits) - len(sig))
	}
	e, err := strconv.ParseInt(FromUnits(in[expStart:expEnd]).Go(), 10, 32)
	if err != nil {
		lit.hexExpHuge = true
	}
	lit.hexExp = e
	if lit.hexExpNeg {
		lit.hexExp = -e
	}
	return lit
}

func stripLeadingZeros(s string) string {
	i := 0
	for i < len(s) && s[i] == '0' {
		i++
	}
	return s[i:]
}

// hexToFloat implements the rounding of FloatingDecimal.parseHexString for a
// non-zero significand, returning the double and float results.
func hexToFloat(lit javaFloatLiteral) (float64, float32) {
	if lit.hexExpHuge {
		switch {
		case !lit.hexExpNeg && lit.neg:
			return math.Inf(-1), float32(math.Inf(-1))
		case !lit.hexExpNeg:
			return math.Inf(1), float32(math.Inf(1))
		case lit.neg:
			return math.Copysign(0, -1), float32(math.Copysign(0, -1))
		}
		return 0, 0
	}
	sig := lit.hexSig
	var exponentAdjust int64
	if lit.hexIntLen >= 1 {
		exponentAdjust = 4 * int64(lit.hexIntLen-1)
	} else {
		exponentAdjust = -4 * (int64(-lit.hexIntLen) + 1)
	}
	exponent := lit.hexExp + exponentAdjust
	hexDigit := func(i int) int64 {
		c := sig[i]
		switch {
		case c <= '9':
			return int64(c - '0')
		case c <= 'F':
			return int64(c-'A') + 10
		}
		return int64(c-'a') + 10
	}
	round, sticky := false, false
	var nextShift int
	var significand int64
	lead := hexDigit(0)
	switch {
	case lead == 1:
		significand |= lead << 52
		nextShift = 52 - 4
	case lead <= 3:
		significand |= lead << 51
		nextShift = 52 - 5
		exponent++
	case lead <= 7:
		significand |= lead << 50
		nextShift = 52 - 6
		exponent += 2
	default:
		significand |= lead << 49
		nextShift = 52 - 7
		exponent += 3
	}
	i := 1
	for ; i < len(sig) && nextShift >= 0; i++ {
		significand |= hexDigit(i) << uint(nextShift)
		nextShift -= 4
	}
	if i < len(sig) {
		d := hexDigit(i)
		switch nextShift {
		case -1:
			significand |= (d & 0xE) >> 1
			round = d&0x1 != 0
		case -2:
			significand |= (d & 0xC) >> 2
			round = d&0x2 != 0
			sticky = d&0x1 != 0
		case -3:
			significand |= (d & 0x8) >> 3
			round = d&0x4 != 0
			sticky = d&0x3 != 0
		case -4:
			round = d&0x8 != 0
			sticky = d&0x7 != 0
		}
		i++
		for i < len(sig) && !sticky {
			sticky = hexDigit(i) != 0
			i++
		}
	}

	// Float result.
	var floatBits int32
	if lit.neg {
		floatBits = math.MinInt32
	}
	if exponent >= -126 {
		if exponent > 127 {
			floatBits |= 0x7f800000
		} else {
			const threshShift = 53 - 24 - 1
			floatSticky := significand&(1<<threshShift-1) != 0 || round || sticky
			iValue := int32(significand >> threshShift)
			if iValue&3 != 1 || floatSticky {
				iValue++
			}
			floatBits |= (int32(exponent)+(127-1))<<23 + iValue>>1
		}
	} else if exponent >= -149-1 {
		threshShift := uint((53 - 2 + -149) - exponent)
		floatSticky := significand&(1<<threshShift-1) != 0 || round || sticky
		iValue := int32(significand >> threshShift)
		if iValue&3 != 1 || floatSticky {
			iValue++
		}
		floatBits |= iValue >> 1
	}
	fValue := math.Float32frombits(uint32(floatBits))

	// Double result.
	if exponent > 1023 {
		if lit.neg {
			return math.Inf(-1), fValue
		}
		return math.Inf(1), fValue
	}
	if exponent >= -1022 {
		significand = (exponent+1023)<<52&0x7ff0000000000000 | significand&(1<<52-1)
	} else {
		if exponent < -1074-1 {
			if lit.neg {
				return math.Copysign(0, -1), fValue
			}
			return 0, fValue
		}
		sticky = sticky || round
		bitsDiscarded := 53 - (int(exponent) - -1074 + 1)
		round = significand&(1<<uint(bitsDiscarded-1)) != 0
		if bitsDiscarded > 1 {
			mask := ^(^int64(0) << uint(bitsDiscarded-1))
			sticky = sticky || significand&mask != 0
		}
		significand >>= uint(bitsDiscarded)
		significand = (int64(-1022-1)+1023)<<52&0x7ff0000000000000 | significand&(1<<52-1)
	}
	leastZero := significand&1 == 0
	if leastZero && round && sticky || !leastZero && round {
		significand++
	}
	if lit.neg {
		significand |= math.MinInt64
	}
	return math.Float64frombits(uint64(significand)), fValue
}

// ---------------------------------------------------------------------------
// FloatingDecimal.BinaryToASCIIBuffer

// fdDecimal is the result of FloatingDecimal's binary to decimal conversion
// of a finite non-zero double: the value is 0.digits * 10^decExponent.
type fdDecimal struct {
	neg         bool
	digits      []byte
	decExponent int
	// roundedUp reports that the last digit was rounded up
	// (digitsRoundedUp), and exact that the digits are the exact value
	// (decimalDigitsExact).
	roundedUp bool
	exact     bool
}

// fdConvert implements FloatingDecimal.getBinaryToASCIIConverter for a
// finite double. Zero yields the single digit '0'.
func fdConvert(d float64) fdDecimal {
	const expShift = 52
	const fractHOB = int64(1) << expShift
	dBits := int64(math.Float64bits(d))
	r := fdDecimal{neg: dBits < 0}
	fractBits := dBits & (1<<52 - 1)
	binExp := int(dBits>>expShift) & 0x7ff
	var nSignificantBits int
	if binExp == 0 {
		if fractBits == 0 {
			r.digits = []byte{'0'}
			return r
		}
		leadingZeros := bits.LeadingZeros64(uint64(fractBits))
		shift := leadingZeros - (63 - expShift)
		fractBits <<= uint(shift)
		binExp = 1 - shift
		nSignificantBits = 64 - leadingZeros
	} else {
		fractBits |= fractHOB
		nSignificantBits = expShift + 1
	}
	binExp -= 1023
	r.dtoa(binExp, fractBits, nSignificantBits)
	return r
}

// fdInsignificantDigitsNumber[p] is the number of decimal digits of 2^p
// minus one, for small p. It is a read-only lookup table.
var fdInsignificantDigitsNumber = [...]int{
	0, 0, 0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3, 3,
	4, 4, 4, 5, 5, 5, 6, 6, 6, 6, 7, 7, 7,
	8, 8, 8, 9, 9, 9, 9, 10, 10, 10, 11, 11, 11,
	12, 12, 12, 12, 13, 13, 13, 14, 14, 14,
	15, 15, 15, 15, 16, 16, 16, 17, 17, 17,
	18, 18, 18, 19,
}

// fdN5Bits[i] is approximately ceil(log2(5^i)), as FloatingDecimal's
// N_5_BITS has it. It is a read-only lookup table.
var fdN5Bits = [...]int{
	0, 3, 5, 7, 10, 12, 14, 17, 19, 21, 24, 26, 28, 31, 33, 35, 38, 40, 42,
	45, 47, 49, 52, 54, 56, 59, 61,
}

// fdPow5 returns 5^i as an int64 (i <= 27).
func fdPow5(i int) int64 {
	p := int64(1)
	for ; i > 0; i-- {
		p *= 5
	}
	return p
}

func (r *fdDecimal) developLongDigits(decExponent int, lvalue int64, insignificantDigits int) {
	if insignificantDigits != 0 {
		pow10 := fdPow5(insignificantDigits) << uint(insignificantDigits)
		residue := lvalue % pow10
		lvalue /= pow10
		decExponent += insignificantDigits
		if residue >= pow10>>1 {
			lvalue++
		}
	}
	var buf [20]byte
	digitno := len(buf) - 1
	c := int(lvalue % 10)
	lvalue /= 10
	for c == 0 {
		decExponent++
		c = int(lvalue % 10)
		lvalue /= 10
	}
	for lvalue != 0 {
		buf[digitno] = byte(c + '0')
		digitno--
		decExponent++
		c = int(lvalue % 10)
		lvalue /= 10
	}
	buf[digitno] = byte(c + '0')
	r.decExponent = decExponent + 1
	r.digits = append([]byte(nil), buf[digitno:]...)
}

func (r *fdDecimal) dtoa(binExp int, fractBits int64, nSignificantBits int) {
	const expShift = 52
	tailZeros := bits.TrailingZeros64(uint64(fractBits))
	nFractBits := expShift + 1 - tailZeros
	r.roundedUp = false
	r.exact = false
	nTinyBits := nFractBits - binExp - 1
	if nTinyBits < 0 {
		nTinyBits = 0
	}
	if binExp <= 62 && binExp >= -21 {
		if nTinyBits < 27 && nFractBits+fdN5Bits[nTinyBits] < 64 && nTinyBits == 0 {
			insignificant := 0
			if binExp > nSignificantBits {
				p2 := binExp - nSignificantBits - 1
				if p2 > 1 && p2 < len(fdInsignificantDigitsNumber) {
					insignificant = fdInsignificantDigitsNumber[p2]
				}
			}
			if binExp >= expShift {
				fractBits <<= uint(binExp - expShift)
			} else {
				fractBits = int64(uint64(fractBits) >> uint(expShift-binExp))
			}
			r.developLongDigits(0, fractBits, insignificant)
			return
		}
	}
	decExp := fdEstimateDecExp(fractBits, binExp)
	B5 := 0
	if -decExp > 0 {
		B5 = -decExp
	}
	B2 := B5 + nTinyBits + binExp
	S5 := 0
	if decExp > 0 {
		S5 = decExp
	}
	S2 := S5 + nTinyBits
	M5 := B5
	M2 := B2 - nSignificantBits
	fractBits = int64(uint64(fractBits) >> uint(tailZeros))
	B2 -= nFractBits - 1
	common2 := B2
	if S2 < common2 {
		common2 = S2
	}
	B2 -= common2
	S2 -= common2
	M2 -= common2
	if nFractBits == 1 {
		M2--
	}
	if M2 < 0 {
		B2 -= M2
		S2 -= M2
		M2 = 0
	}
	n5bits := func(i int) int {
		if i < len(fdN5Bits) {
			return fdN5Bits[i]
		}
		return i * 3
	}
	Bbits := nFractBits + B2 + n5bits(B5)
	tenSbits := S2 + 1 + n5bits(S5+1)
	digits := make([]byte, 0, 20)
	var low, high bool
	var lowDigitDifference int64
	if Bbits < 64 && tenSbits < 64 {
		if Bbits < 32 && tenSbits < 32 {
			b := int32(fractBits) * int32(fdPow5(B5)) << uint(B2)
			s := int32(fdPow5(S5)) << uint(S2)
			m := int32(fdPow5(M5)) << uint(M2)
			tens := s * 10
			q := b / s
			b = 10 * (b % s)
			m *= 10
			low = b < m
			high = b+m > tens
			if q == 0 && !high {
				decExp--
			} else {
				digits = append(digits, byte('0'+q))
			}
			if decExp < -3 || decExp >= 8 {
				high, low = false, false
			}
			for !low && !high {
				q = b / s
				b = 10 * (b % s)
				m *= 10
				if m > 0 {
					low = b < m
					high = b+m > tens
				} else {
					low = true
					high = true
				}
				digits = append(digits, byte('0'+q))
			}
			lowDigitDifference = int64(b<<1 - tens)
			r.exact = b == 0
		} else {
			b := fractBits * fdPow5(B5) << uint(B2)
			s := fdPow5(S5) << uint(S2)
			m := fdPow5(M5) << uint(M2)
			tens := s * 10
			q := b / s
			b = 10 * (b % s)
			m *= 10
			low = b < m
			high = b+m > tens
			if q == 0 && !high {
				decExp--
			} else {
				digits = append(digits, byte('0'+q))
			}
			if decExp < -3 || decExp >= 8 {
				high, low = false, false
			}
			for !low && !high {
				q = b / s
				b = 10 * (b % s)
				m *= 10
				if m > 0 {
					low = b < m
					high = b+m > tens
				} else {
					low = true
					high = true
				}
				digits = append(digits, byte('0'+q))
			}
			lowDigitDifference = b<<1 - tens
			r.exact = b == 0
		}
	} else {
		// The JDK uses FDBigInteger here; the arithmetic is exact, so
		// math/big gives the same results.
		pow52 := func(p5, p2 int) *big.Int {
			v := new(big.Int).Exp(big.NewInt(5), big.NewInt(int64(p5)), nil)
			return v.Lsh(v, uint(p2))
		}
		S := pow52(S5, S2)
		B := pow52(B5, B2)
		B.Mul(B, big.NewInt(fractBits))
		M := pow52(M5+1, M2+1)
		tenS := pow52(S5+1, S2+1)
		var q, rem, sum big.Int
		ten := big.NewInt(10)
		quoRem := func() int {
			q.QuoRem(B, S, &rem)
			B.Mul(&rem, ten)
			return int(q.Int64())
		}
		qd := quoRem()
		low = B.Cmp(M) < 0
		high = tenS.Cmp(sum.Add(B, M)) <= 0
		if qd == 0 && !high {
			decExp--
		} else {
			digits = append(digits, byte('0'+qd))
		}
		if decExp < -3 || decExp >= 8 {
			high, low = false, false
		}
		for !low && !high {
			qd = quoRem()
			M.Mul(M, ten)
			low = B.Cmp(M) < 0
			high = tenS.Cmp(sum.Add(B, M)) <= 0
			digits = append(digits, byte('0'+qd))
		}
		if high && low {
			B.Lsh(B, 1)
			lowDigitDifference = int64(B.Cmp(tenS))
		}
		r.exact = B.Sign() == 0
	}
	r.decExponent = decExp + 1
	r.digits = digits
	if high {
		if low {
			if lowDigitDifference == 0 {
				if digits[len(digits)-1]&1 != 0 {
					r.roundup()
				}
			} else if lowDigitDifference > 0 {
				r.roundup()
			}
		} else {
			r.roundup()
		}
	}
}

func (r *fdDecimal) roundup() {
	i := len(r.digits) - 1
	q := r.digits[i]
	if q == '9' {
		for q == '9' && i > 0 {
			r.digits[i] = '0'
			i--
			q = r.digits[i]
		}
		if q == '9' {
			r.decExponent++
			r.digits[0] = '1'
			return
		}
	}
	r.digits[i] = q + 1
	r.roundedUp = true
}

// fdEstimateDecExp is FloatingDecimal's estimate of floor(log10(d)).
func fdEstimateDecExp(fractBits int64, binExp int) int {
	const expShift = 52
	d2 := math.Float64frombits(uint64(1023<<expShift | fractBits&(1<<52-1)))
	d := float64((d2-1.5)*0.289529654) + 0.176091259 + float64(float64(binExp)*0.301029995663981)
	dBits := int64(math.Float64bits(d))
	exponent := int((dBits&0x7ff0000000000000)>>expShift) - 1023
	isNegative := dBits < 0
	switch {
	case exponent >= 0 && exponent < 52:
		mask := int64(1<<52-1) >> uint(exponent)
		r := int((dBits&(1<<52-1) | 1<<52) >> uint(expShift-exponent))
		if isNegative {
			if mask&dBits == 0 {
				return -r
			}
			return -r - 1
		}
		return r
	case exponent < 0:
		if dBits&^math.MinInt64 == 0 {
			return 0
		}
		if isNegative {
			return -1
		}
		return 0
	}
	return int(d)
}

// javaFormatString is BinaryToASCIIBuffer.toJavaFormatString: the converted
// digits in the format of Double.toString before JDK 19.
func (r *fdDecimal) javaFormatString() string {
	buf := make([]byte, 0, 26)
	if r.neg {
		buf = append(buf, '-')
	}
	nDigits := len(r.digits)
	decExponent := r.decExponent
	switch {
	case decExponent > 0 && decExponent < 8:
		charLength := nDigits
		if decExponent < charLength {
			charLength = decExponent
		}
		buf = append(buf, r.digits[:charLength]...)
		if charLength < decExponent {
			for i := charLength; i < decExponent; i++ {
				buf = append(buf, '0')
			}
			buf = append(buf, '.', '0')
		} else {
			buf = append(buf, '.')
			if charLength < nDigits {
				buf = append(buf, r.digits[charLength:]...)
			} else {
				buf = append(buf, '0')
			}
		}
	case decExponent <= 0 && decExponent > -3:
		buf = append(buf, '0', '.')
		for i := decExponent; i < 0; i++ {
			buf = append(buf, '0')
		}
		buf = append(buf, r.digits...)
	default:
		buf = append(buf, r.digits[0], '.')
		if nDigits > 1 {
			buf = append(buf, r.digits[1:]...)
		} else {
			buf = append(buf, '0')
		}
		buf = append(buf, 'E')
		var e int
		if decExponent <= 0 {
			buf = append(buf, '-')
			e = -decExponent + 1
		} else {
			e = decExponent - 1
		}
		buf = strconv.AppendInt(buf, int64(e), 10)
	}
	return string(buf)
}
