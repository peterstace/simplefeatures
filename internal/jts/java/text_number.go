package java

import (
	"math"
	"math/bits"
	"strconv"
)

// This file implements the boxed primitive classes (Integer, Long, Short,
// Byte, Double, Float, Boolean, Character) and Number. Boxed values are Go
// primitives held in an any (Integer is an int32, Double a float64, and so
// on), so instance methods called on them are functions taking the receiver
// as their first argument.

// unboxRecv returns the receiver of an instance method of a boxed class,
// throwing NullPointerException for null and ClassCastException for a value
// of another class.
func unboxRecv[T any](recv any, class string) T {
	if recv == nil {
		panic(NullPointerException_new())
	}
	v, ok := recv.(T)
	if !ok {
		panic(ClassCastError(recv, class))
	}
	return v
}

// ---------------------------------------------------------------------------
// Number

// Number_isInstance implements instanceof Number.
func Number_isInstance(o any) bool {
	switch o.(type) {
	case int32, int64, int16, int8, float64, float32, *BigDecimal:
		return true
	}
	return false
}

// Number_DoubleValue implements Number.doubleValue on a boxed receiver.
func Number_DoubleValue(recv any) float64 {
	switch x := recv.(type) {
	case nil:
		panic(NullPointerException_new())
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case int16:
		return float64(x)
	case int8:
		return float64(x)
	case float64:
		return x
	case float32:
		return float64(x)
	case *BigDecimal:
		return x.DoubleValue()
	}
	panic(ClassCastError(recv, "java.lang.Number"))
}

// Number_FloatValue implements Number.floatValue on a boxed receiver.
func Number_FloatValue(recv any) float32 {
	switch x := recv.(type) {
	case int32:
		return float32(x)
	case int64:
		return float32(x)
	case int16:
		return float32(x)
	case int8:
		return float32(x)
	case float32:
		return x
	case *BigDecimal:
		return x.FloatValue()
	}
	return float32(Number_DoubleValue(recv))
}

// Number_LongValue implements Number.longValue on a boxed receiver.
func Number_LongValue(recv any) int64 {
	switch x := recv.(type) {
	case int32:
		return int64(x)
	case int64:
		return x
	case int16:
		return int64(x)
	case int8:
		return int64(x)
	case float32:
		return F2L(float64(x))
	case *BigDecimal:
		return x.LongValue()
	}
	return F2L(Number_DoubleValue(recv))
}

// Number_IntValue implements Number.intValue on a boxed receiver.
func Number_IntValue(recv any) int32 {
	switch x := recv.(type) {
	case int32:
		return x
	case int64:
		return int32(x)
	case int16:
		return int32(x)
	case int8:
		return int32(x)
	case float32:
		return F2I(float64(x))
	case *BigDecimal:
		return x.IntValue()
	}
	return F2I(Number_DoubleValue(recv))
}

// Number_ShortValue implements Number.shortValue on a boxed receiver.
func Number_ShortValue(recv any) int16 {
	switch x := recv.(type) {
	case int64:
		return int16(x)
	case *BigDecimal:
		return int16(x.IntValue())
	}
	return int16(Number_IntValue(recv))
}

// Number_ByteValue implements Number.byteValue on a boxed receiver.
func Number_ByteValue(recv any) int8 {
	switch x := recv.(type) {
	case int64:
		return int8(x)
	case *BigDecimal:
		return int8(x.IntValue())
	}
	return int8(Number_IntValue(recv))
}

// ---------------------------------------------------------------------------
// Parsing integers

// numberFormatForInput returns NumberFormatException.forInputString.
func numberFormatForInput(s *String, radix int32) any {
	msg := Concat(Lit("For input string: \""), s, Lit("\""))
	if radix != 10 {
		msg = Concat(msg, Lit(" under radix "), StrOfInt(radix))
	}
	return NumberFormatException_new_String(msg)
}

func checkRadix(s *String, radix int32) {
	if s == nil {
		panic(NumberFormatException_new_String(Lit("Cannot parse null string")))
	}
	if radix < 2 {
		panic(NumberFormatException_new_String(Concat(Lit("radix "), StrOfInt(radix), Lit(" less than Character.MIN_RADIX"))))
	}
	if radix > 36 {
		panic(NumberFormatException_new_String(Concat(Lit("radix "), StrOfInt(radix), Lit(" greater than Character.MAX_RADIX"))))
	}
}

// parseSigned implements Long.parseLong(String, int), and
// Integer.parseInt(String, int) for the int range.
func parseSigned(s *String, radix int32, min, max int64) int64 {
	checkRadix(s, radix)
	u := s.units
	if len(u) == 0 {
		panic(numberFormatForInput(s, radix))
	}
	neg := false
	i := 0
	limit := -max
	if u[0] < '0' {
		switch u[0] {
		case '-':
			neg = true
			limit = min
		case '+':
		default:
			panic(numberFormatForInput(s, radix))
		}
		if len(u) == 1 {
			panic(numberFormatForInput(s, radix))
		}
		i++
	}
	multmin := limit / int64(radix)
	var result int64
	for ; i < len(u); i++ {
		digit := int64(Character_digit_char_int(u[i], radix))
		if digit < 0 || result < multmin {
			panic(numberFormatForInput(s, radix))
		}
		result *= int64(radix)
		if result < limit+digit {
			panic(numberFormatForInput(s, radix))
		}
		result -= digit
	}
	if neg {
		return result
	}
	return -result
}

// parseUnsignedInt implements Integer.parseUnsignedInt(String, int).
func parseUnsignedInt(s *String, radix int32) int32 {
	if s == nil {
		panic(NumberFormatException_new_String(Lit("Cannot parse null string")))
	}
	n := len(s.units)
	if n == 0 {
		panic(numberFormatForInput(s, radix))
	}
	if s.units[0] == '-' {
		panic(NumberFormatException_new_String(Concat(Lit("Illegal leading minus sign on unsigned string "), s, Lit("."))))
	}
	if n <= 5 || radix == 10 && n <= 9 {
		return Integer_parseInt_String_int(s, radix)
	}
	ell := Long_parseLong_String_int(s, radix)
	if uint64(ell)>>32 != 0 {
		panic(NumberFormatException_new_String(Concat(Lit("String value "), s, Lit(" exceeds range of unsigned int."))))
	}
	return int32(ell)
}

// parseUnsignedLong implements Long.parseUnsignedLong(String, int).
func parseUnsignedLong(s *String, radix int32) int64 {
	if s == nil {
		panic(NumberFormatException_new_String(Lit("Cannot parse null string")))
	}
	n := len(s.units)
	if n == 0 {
		panic(numberFormatForInput(s, radix))
	}
	if s.units[0] == '-' {
		panic(NumberFormatException_new_String(Concat(Lit("Illegal leading minus sign on unsigned string "), s, Lit("."))))
	}
	if n <= 12 || radix == 10 && n <= 18 {
		return Long_parseLong_String_int(s, radix)
	}
	first := parseLongRange(s.units, 0, n-1, radix)
	second := Character_digit_char_int(s.units[n-1], radix)
	if second < 0 {
		panic(NumberFormatException_new_String(Concat(Lit("Bad digit at end of "), s)))
	}
	result := first*int64(radix) + int64(second)
	guard := radix * int32(uint64(first)>>57)
	if guard >= 128 || result >= 0 && guard >= 128-36 {
		panic(NumberFormatException_new_String(Concat(Lit("String value "), s, Lit(" exceeds range of unsigned long."))))
	}
	return result
}

// parseLongRange implements Long.parseLong(CharSequence, int, int, int) on
// u[begin:end], with its "Error at index" messages.
func parseLongRange(u []uint16, begin, end int, radix int32) int64 {
	fail := func(i int) {
		panic(NumberFormatException_new_String(Concat(Lit("Error at index "), StrOfInt(int32(i-begin)),
			Lit(" in: \""), FromUnits(append([]uint16(nil), u[begin:end]...)), Lit("\""))))
	}
	if radix < 2 {
		panic(NumberFormatException_new_String(Concat(Lit("radix "), StrOfInt(radix), Lit(" less than Character.MIN_RADIX"))))
	}
	if radix > 36 {
		panic(NumberFormatException_new_String(Concat(Lit("radix "), StrOfInt(radix), Lit(" greater than Character.MAX_RADIX"))))
	}
	if begin >= end {
		panic(NumberFormatException_new_String(Lit("")))
	}
	neg := false
	i := begin
	limit := int64(-math.MaxInt64)
	if u[i] < '0' {
		switch u[i] {
		case '-':
			neg = true
			limit = math.MinInt64
		case '+':
		default:
			fail(i)
		}
		i++
	}
	if i >= end {
		fail(i)
	}
	multmin := limit / int64(radix)
	var result int64
	for ; i < end; i++ {
		digit := int64(Character_digit_char_int(u[i], radix))
		if digit < 0 || result < multmin {
			fail(i)
		}
		result *= int64(radix)
		if result < limit+digit {
			fail(i)
		}
		result -= digit
	}
	if neg {
		return result
	}
	return -result
}

// decode implements Integer.decode and Long.decode.
func decode(nm *String, min, max int64) int64 {
	u := nm.units
	if len(u) == 0 {
		panic(NumberFormatException_new_String(Lit("Zero length string")))
	}
	radix := int32(10)
	index := 0
	neg := false
	switch u[0] {
	case '-':
		neg = true
		index++
	case '+':
		index++
	}
	rest := FromUnits(u[index:])
	switch {
	case rest.StartsWith_String(Lit("0x")) || rest.StartsWith_String(Lit("0X")):
		index += 2
		radix = 16
	case rest.StartsWith_String(Lit("#")):
		index++
		radix = 16
	case rest.StartsWith_String(Lit("0")) && len(u) > 1+index:
		index++
		radix = 8
	}
	if index < len(u) && (u[index] == '-' || u[index] == '+') {
		panic(NumberFormatException_new_String(Lit("Sign character in wrong position")))
	}
	digits := FromUnits(u[index:])
	var result int64
	func() {
		defer func() {
			if r := recover(); r != nil {
				if !NumberFormatException_isInstance(r) {
					panic(r)
				}
				constant := digits
				if neg {
					constant = Concat(Lit("-"), digits)
				}
				result = parseSigned(constant, radix, min, max)
			}
		}()
		result = parseSigned(digits, radix, min, max)
		if neg {
			result = -result
		}
	}()
	return result
}

// ---------------------------------------------------------------------------
// Integer

// Integer_isInstance implements instanceof Integer.
func Integer_isInstance(o any) bool { _, ok := o.(int32); return ok }

// Integer_new_int implements new Integer(int).
func Integer_new_int(i int32) any { return i }

// Integer_new_String implements new Integer(String).
func Integer_new_String(s *String) any { return Integer_parseInt_String_int(s, 10) }

// Integer_valueOf_int implements Integer.valueOf(int).
func Integer_valueOf_int(i int32) any { return i }

// Integer_valueOf_String implements Integer.valueOf(String).
func Integer_valueOf_String(s *String) any { return Integer_parseInt_String_int(s, 10) }

// Integer_valueOf_String_int implements Integer.valueOf(String, int).
func Integer_valueOf_String_int(s *String, radix int32) any {
	return Integer_parseInt_String_int(s, radix)
}

// Integer_parseInt_String implements Integer.parseInt(String).
func Integer_parseInt_String(s *String) int32 { return Integer_parseInt_String_int(s, 10) }

// Integer_parseInt_String_int implements Integer.parseInt(String, int).
func Integer_parseInt_String_int(s *String, radix int32) int32 {
	return int32(parseSigned(s, radix, math.MinInt32, math.MaxInt32))
}

// Integer_parseUnsignedInt_String implements Integer.parseUnsignedInt(String).
func Integer_parseUnsignedInt_String(s *String) int32 {
	return Integer_parseUnsignedInt_String_int(s, 10)
}

// Integer_parseUnsignedInt_String_int implements
// Integer.parseUnsignedInt(String, int).
func Integer_parseUnsignedInt_String_int(s *String, radix int32) int32 {
	return parseUnsignedInt(s, radix)
}

// Integer_decode implements Integer.decode.
func Integer_decode(s *String) any { return int32(decode(s, math.MinInt32, math.MaxInt32)) }

// Integer_toString_int implements Integer.toString(int).
func Integer_toString_int(i int32) *String { return StrOfInt(i) }

// Integer_toString_int_int implements Integer.toString(int, int).
func Integer_toString_int_int(i, radix int32) *String {
	if radix < 2 || radix > 36 {
		radix = 10
	}
	return FromGo(strconv.FormatInt(int64(i), int(radix)))
}

// Integer_toUnsignedString_int implements Integer.toUnsignedString(int).
func Integer_toUnsignedString_int(i int32) *String {
	return FromGo(strconv.FormatUint(uint64(uint32(i)), 10))
}

// Integer_toUnsignedString_int_int implements Integer.toUnsignedString(int, int).
func Integer_toUnsignedString_int_int(i, radix int32) *String {
	return Long_toUnsignedString_long_int(int64(uint32(i)), radix)
}

// Integer_toBinaryString implements Integer.toBinaryString.
func Integer_toBinaryString(i int32) *String { return FromGo(strconv.FormatUint(uint64(uint32(i)), 2)) }

// Integer_toHexString implements Integer.toHexString.
func Integer_toHexString(i int32) *String { return FromGo(strconv.FormatUint(uint64(uint32(i)), 16)) }

// Integer_toOctalString implements Integer.toOctalString.
func Integer_toOctalString(i int32) *String { return FromGo(strconv.FormatUint(uint64(uint32(i)), 8)) }

// Integer_compare implements Integer.compare.
func Integer_compare(a, b int32) int32 {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// Integer_compareUnsigned implements Integer.compareUnsigned.
func Integer_compareUnsigned(a, b int32) int32 {
	return Integer_compare(a+math.MinInt32, b+math.MinInt32)
}

// Integer_hashCode implements Integer.hashCode(int).
func Integer_hashCode(i int32) int32 { return i }

// Integer_max implements Integer.max.
func Integer_max(a, b int32) int32 {
	if a >= b {
		return a
	}
	return b
}

// Integer_min implements Integer.min.
func Integer_min(a, b int32) int32 {
	if a <= b {
		return a
	}
	return b
}

// Integer_sum implements Integer.sum.
func Integer_sum(a, b int32) int32 { return a + b }

// Integer_signum implements Integer.signum.
func Integer_signum(i int32) int32 { return i>>31 | int32(uint32(-i)>>31) }

// Integer_bitCount implements Integer.bitCount.
func Integer_bitCount(i int32) int32 { return int32(bits.OnesCount32(uint32(i))) }

// Integer_reverse implements Integer.reverse.
func Integer_reverse(i int32) int32 { return int32(bits.Reverse32(uint32(i))) }

// Integer_reverseBytes implements Integer.reverseBytes.
func Integer_reverseBytes(i int32) int32 { return int32(bits.ReverseBytes32(uint32(i))) }

// Integer_rotateLeft implements Integer.rotateLeft.
func Integer_rotateLeft(i, distance int32) int32 {
	return int32(bits.RotateLeft32(uint32(i), int(distance&31)))
}

// Integer_rotateRight implements Integer.rotateRight.
func Integer_rotateRight(i, distance int32) int32 {
	return int32(bits.RotateLeft32(uint32(i), -int(distance&31)))
}

// Integer_highestOneBit implements Integer.highestOneBit.
func Integer_highestOneBit(i int32) int32 {
	if i == 0 {
		return 0
	}
	return int32(uint32(1) << (31 - bits.LeadingZeros32(uint32(i))))
}

// Integer_lowestOneBit implements Integer.lowestOneBit.
func Integer_lowestOneBit(i int32) int32 { return i & -i }

// Integer_numberOfLeadingZeros implements Integer.numberOfLeadingZeros.
func Integer_numberOfLeadingZeros(i int32) int32 { return int32(bits.LeadingZeros32(uint32(i))) }

// Integer_numberOfTrailingZeros implements Integer.numberOfTrailingZeros.
func Integer_numberOfTrailingZeros(i int32) int32 { return int32(bits.TrailingZeros32(uint32(i))) }

// Integer_toUnsignedLong implements Integer.toUnsignedLong.
func Integer_toUnsignedLong(i int32) int64 { return int64(uint32(i)) }

// Integer_divideUnsigned implements Integer.divideUnsigned.
func Integer_divideUnsigned(a, b int32) int32 { return int32(uint32(a) / uint32(b)) }

// Integer_remainderUnsigned implements Integer.remainderUnsigned.
func Integer_remainderUnsigned(a, b int32) int32 { return int32(uint32(a) % uint32(b)) }

// Integer_IntValue implements Integer.intValue on a boxed receiver.
func Integer_IntValue(recv any) int32 { return unboxRecv[int32](recv, "java.lang.Integer") }

// Integer_LongValue implements Integer.longValue on a boxed receiver.
func Integer_LongValue(recv any) int64 { return int64(Integer_IntValue(recv)) }

// Integer_DoubleValue implements Integer.doubleValue on a boxed receiver.
func Integer_DoubleValue(recv any) float64 { return float64(Integer_IntValue(recv)) }

// Integer_FloatValue implements Integer.floatValue on a boxed receiver.
func Integer_FloatValue(recv any) float32 { return float32(Integer_IntValue(recv)) }

// Integer_ShortValue implements Integer.shortValue on a boxed receiver.
func Integer_ShortValue(recv any) int16 { return int16(Integer_IntValue(recv)) }

// Integer_ByteValue implements Integer.byteValue on a boxed receiver.
func Integer_ByteValue(recv any) int8 { return int8(Integer_IntValue(recv)) }

// Integer_CompareTo_Object implements Integer.compareTo on a boxed receiver.
func Integer_CompareTo_Object(recv, o any) int32 {
	return Integer_compare(Integer_IntValue(recv), unboxRecv[int32](o, "java.lang.Integer"))
}

// Integer_Equals_Object implements Integer.equals on a boxed receiver.
func Integer_Equals_Object(recv, o any) bool {
	v, ok := o.(int32)
	return ok && Integer_IntValue(recv) == v
}

// Integer_HashCode implements Integer.hashCode on a boxed receiver.
func Integer_HashCode(recv any) int32 { return Integer_IntValue(recv) }

// Integer_ToString implements Integer.toString on a boxed receiver.
func Integer_ToString(recv any) *String { return StrOfInt(Integer_IntValue(recv)) }

// ---------------------------------------------------------------------------
// Long

// Long_isInstance implements instanceof Long.
func Long_isInstance(o any) bool { _, ok := o.(int64); return ok }

// Long_new_long implements new Long(long).
func Long_new_long(l int64) any { return l }

// Long_new_String implements new Long(String).
func Long_new_String(s *String) any { return Long_parseLong_String_int(s, 10) }

// Long_valueOf_long implements Long.valueOf(long).
func Long_valueOf_long(l int64) any { return l }

// Long_valueOf_String implements Long.valueOf(String).
func Long_valueOf_String(s *String) any { return Long_parseLong_String_int(s, 10) }

// Long_valueOf_String_int implements Long.valueOf(String, int).
func Long_valueOf_String_int(s *String, radix int32) any { return Long_parseLong_String_int(s, radix) }

// Long_parseLong_String implements Long.parseLong(String).
func Long_parseLong_String(s *String) int64 { return Long_parseLong_String_int(s, 10) }

// Long_parseLong_String_int implements Long.parseLong(String, int).
func Long_parseLong_String_int(s *String, radix int32) int64 {
	return parseSigned(s, radix, math.MinInt64, math.MaxInt64)
}

// Long_parseUnsignedLong_String implements Long.parseUnsignedLong(String).
func Long_parseUnsignedLong_String(s *String) int64 { return Long_parseUnsignedLong_String_int(s, 10) }

// Long_parseUnsignedLong_String_int implements
// Long.parseUnsignedLong(String, int).
func Long_parseUnsignedLong_String_int(s *String, radix int32) int64 {
	return parseUnsignedLong(s, radix)
}

// Long_decode implements Long.decode.
func Long_decode(s *String) any { return decode(s, math.MinInt64, math.MaxInt64) }

// Long_toString_long implements Long.toString(long).
func Long_toString_long(l int64) *String { return StrOfLong(l) }

// Long_toString_long_int implements Long.toString(long, int).
func Long_toString_long_int(l int64, radix int32) *String {
	if radix < 2 || radix > 36 {
		radix = 10
	}
	return FromGo(strconv.FormatInt(l, int(radix)))
}

// Long_toUnsignedString_long implements Long.toUnsignedString(long).
func Long_toUnsignedString_long(l int64) *String { return FromGo(strconv.FormatUint(uint64(l), 10)) }

// Long_toUnsignedString_long_int implements Long.toUnsignedString(long, int).
func Long_toUnsignedString_long_int(l int64, radix int32) *String {
	if radix < 2 || radix > 36 {
		radix = 10
	}
	return FromGo(strconv.FormatUint(uint64(l), int(radix)))
}

// Long_toBinaryString implements Long.toBinaryString.
func Long_toBinaryString(l int64) *String { return FromGo(strconv.FormatUint(uint64(l), 2)) }

// Long_toHexString implements Long.toHexString.
func Long_toHexString(l int64) *String { return FromGo(strconv.FormatUint(uint64(l), 16)) }

// Long_toOctalString implements Long.toOctalString.
func Long_toOctalString(l int64) *String { return FromGo(strconv.FormatUint(uint64(l), 8)) }

// Long_compare implements Long.compare.
func Long_compare(a, b int64) int32 {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// Long_compareUnsigned implements Long.compareUnsigned.
func Long_compareUnsigned(a, b int64) int32 {
	return Long_compare(a+math.MinInt64, b+math.MinInt64)
}

// Long_hashCode implements Long.hashCode(long).
func Long_hashCode(l int64) int32 { return int32(l ^ int64(uint64(l)>>32)) }

// Long_max implements Long.max.
func Long_max(a, b int64) int64 {
	if a >= b {
		return a
	}
	return b
}

// Long_min implements Long.min.
func Long_min(a, b int64) int64 {
	if a <= b {
		return a
	}
	return b
}

// Long_sum implements Long.sum.
func Long_sum(a, b int64) int64 { return a + b }

// Long_signum implements Long.signum.
func Long_signum(l int64) int32 { return int32(l>>63 | int64(uint64(-l)>>63)) }

// Long_bitCount implements Long.bitCount.
func Long_bitCount(l int64) int32 { return int32(bits.OnesCount64(uint64(l))) }

// Long_reverse implements Long.reverse.
func Long_reverse(l int64) int64 { return int64(bits.Reverse64(uint64(l))) }

// Long_reverseBytes implements Long.reverseBytes.
func Long_reverseBytes(l int64) int64 { return int64(bits.ReverseBytes64(uint64(l))) }

// Long_rotateLeft implements Long.rotateLeft.
func Long_rotateLeft(l int64, distance int32) int64 {
	return int64(bits.RotateLeft64(uint64(l), int(distance&63)))
}

// Long_rotateRight implements Long.rotateRight.
func Long_rotateRight(l int64, distance int32) int64 {
	return int64(bits.RotateLeft64(uint64(l), -int(distance&63)))
}

// Long_highestOneBit implements Long.highestOneBit.
func Long_highestOneBit(l int64) int64 {
	if l == 0 {
		return 0
	}
	return int64(uint64(1) << (63 - bits.LeadingZeros64(uint64(l))))
}

// Long_lowestOneBit implements Long.lowestOneBit.
func Long_lowestOneBit(l int64) int64 { return l & -l }

// Long_numberOfLeadingZeros implements Long.numberOfLeadingZeros.
func Long_numberOfLeadingZeros(l int64) int32 { return int32(bits.LeadingZeros64(uint64(l))) }

// Long_numberOfTrailingZeros implements Long.numberOfTrailingZeros.
func Long_numberOfTrailingZeros(l int64) int32 { return int32(bits.TrailingZeros64(uint64(l))) }

// Long_divideUnsigned implements Long.divideUnsigned.
func Long_divideUnsigned(a, b int64) int64 { return int64(uint64(a) / uint64(b)) }

// Long_remainderUnsigned implements Long.remainderUnsigned.
func Long_remainderUnsigned(a, b int64) int64 { return int64(uint64(a) % uint64(b)) }

// Long_LongValue implements Long.longValue on a boxed receiver.
func Long_LongValue(recv any) int64 { return unboxRecv[int64](recv, "java.lang.Long") }

// Long_IntValue implements Long.intValue on a boxed receiver.
func Long_IntValue(recv any) int32 { return int32(Long_LongValue(recv)) }

// Long_DoubleValue implements Long.doubleValue on a boxed receiver.
func Long_DoubleValue(recv any) float64 { return float64(Long_LongValue(recv)) }

// Long_FloatValue implements Long.floatValue on a boxed receiver.
func Long_FloatValue(recv any) float32 { return float32(Long_LongValue(recv)) }

// Long_ShortValue implements Long.shortValue on a boxed receiver.
func Long_ShortValue(recv any) int16 { return int16(Long_LongValue(recv)) }

// Long_ByteValue implements Long.byteValue on a boxed receiver.
func Long_ByteValue(recv any) int8 { return int8(Long_LongValue(recv)) }

// Long_CompareTo_Object implements Long.compareTo on a boxed receiver.
func Long_CompareTo_Object(recv, o any) int32 {
	return Long_compare(Long_LongValue(recv), unboxRecv[int64](o, "java.lang.Long"))
}

// Long_Equals_Object implements Long.equals on a boxed receiver.
func Long_Equals_Object(recv, o any) bool {
	v, ok := o.(int64)
	return ok && Long_LongValue(recv) == v
}

// Long_HashCode implements Long.hashCode on a boxed receiver.
func Long_HashCode(recv any) int32 { return Long_hashCode(Long_LongValue(recv)) }

// Long_ToString implements Long.toString on a boxed receiver.
func Long_ToString(recv any) *String { return StrOfLong(Long_LongValue(recv)) }

// ---------------------------------------------------------------------------
// Short and Byte

func parseNarrow(s *String, radix int32, min, max int32) int32 {
	i := Integer_parseInt_String_int(s, radix)
	if i < min || i > max {
		panic(NumberFormatException_new_String(Concat(Lit("Value out of range. Value:\""), s, Lit("\" Radix:"), StrOfInt(radix))))
	}
	return i
}

func decodeNarrow(s *String, min, max int32) int32 {
	i := int32(decode(s, math.MinInt32, math.MaxInt32))
	if i < min || i > max {
		panic(NumberFormatException_new_String(Concat(Lit("Value "), StrOfInt(i), Lit(" out of range from input "), s)))
	}
	return i
}

// Short_isInstance implements instanceof Short.
func Short_isInstance(o any) bool { _, ok := o.(int16); return ok }

// Short_new_short implements new Short(short).
func Short_new_short(v int16) any { return v }

// Short_new_String implements new Short(String).
func Short_new_String(s *String) any { return Short_parseShort_String_int(s, 10) }

// Short_valueOf_short implements Short.valueOf(short).
func Short_valueOf_short(v int16) any { return v }

// Short_valueOf_String implements Short.valueOf(String).
func Short_valueOf_String(s *String) any { return Short_parseShort_String_int(s, 10) }

// Short_valueOf_String_int implements Short.valueOf(String, int).
func Short_valueOf_String_int(s *String, radix int32) any {
	return Short_parseShort_String_int(s, radix)
}

// Short_parseShort_String implements Short.parseShort(String).
func Short_parseShort_String(s *String) int16 { return Short_parseShort_String_int(s, 10) }

// Short_parseShort_String_int implements Short.parseShort(String, int).
func Short_parseShort_String_int(s *String, radix int32) int16 {
	return int16(parseNarrow(s, radix, math.MinInt16, math.MaxInt16))
}

// Short_decode implements Short.decode.
func Short_decode(s *String) any { return int16(decodeNarrow(s, math.MinInt16, math.MaxInt16)) }

// Short_toString implements Short.toString(short).
func Short_toString(v int16) *String { return StrOfInt(int32(v)) }

// Short_compare implements Short.compare.
func Short_compare(a, b int16) int32 { return int32(a) - int32(b) }

// Short_hashCode implements Short.hashCode(short).
func Short_hashCode(v int16) int32 { return int32(v) }

// Short_reverseBytes implements Short.reverseBytes.
func Short_reverseBytes(v int16) int16 { return int16(bits.ReverseBytes16(uint16(v))) }

// Short_toUnsignedInt implements Short.toUnsignedInt.
func Short_toUnsignedInt(v int16) int32 { return int32(uint16(v)) }

// Short_toUnsignedLong implements Short.toUnsignedLong.
func Short_toUnsignedLong(v int16) int64 { return int64(uint16(v)) }

// Short_ShortValue implements Short.shortValue on a boxed receiver.
func Short_ShortValue(recv any) int16 { return unboxRecv[int16](recv, "java.lang.Short") }

// Short_IntValue implements Short.intValue on a boxed receiver.
func Short_IntValue(recv any) int32 { return int32(Short_ShortValue(recv)) }

// Short_LongValue implements Short.longValue on a boxed receiver.
func Short_LongValue(recv any) int64 { return int64(Short_ShortValue(recv)) }

// Short_DoubleValue implements Short.doubleValue on a boxed receiver.
func Short_DoubleValue(recv any) float64 { return float64(Short_ShortValue(recv)) }

// Short_FloatValue implements Short.floatValue on a boxed receiver.
func Short_FloatValue(recv any) float32 { return float32(Short_ShortValue(recv)) }

// Short_ByteValue implements Short.byteValue on a boxed receiver.
func Short_ByteValue(recv any) int8 { return int8(Short_ShortValue(recv)) }

// Short_CompareTo_Object implements Short.compareTo on a boxed receiver.
func Short_CompareTo_Object(recv, o any) int32 {
	return Short_compare(Short_ShortValue(recv), unboxRecv[int16](o, "java.lang.Short"))
}

// Short_Equals_Object implements Short.equals on a boxed receiver.
func Short_Equals_Object(recv, o any) bool {
	v, ok := o.(int16)
	return ok && Short_ShortValue(recv) == v
}

// Short_HashCode implements Short.hashCode on a boxed receiver.
func Short_HashCode(recv any) int32 { return int32(Short_ShortValue(recv)) }

// Short_ToString implements Short.toString on a boxed receiver.
func Short_ToString(recv any) *String { return StrOfInt(int32(Short_ShortValue(recv))) }

// Byte_isInstance implements instanceof Byte.
func Byte_isInstance(o any) bool { _, ok := o.(int8); return ok }

// Byte_new_byte implements new Byte(byte).
func Byte_new_byte(v int8) any { return v }

// Byte_new_String implements new Byte(String).
func Byte_new_String(s *String) any { return Byte_parseByte_String_int(s, 10) }

// Byte_valueOf_byte implements Byte.valueOf(byte).
func Byte_valueOf_byte(v int8) any { return v }

// Byte_valueOf_String implements Byte.valueOf(String).
func Byte_valueOf_String(s *String) any { return Byte_parseByte_String_int(s, 10) }

// Byte_valueOf_String_int implements Byte.valueOf(String, int).
func Byte_valueOf_String_int(s *String, radix int32) any { return Byte_parseByte_String_int(s, radix) }

// Byte_parseByte_String implements Byte.parseByte(String).
func Byte_parseByte_String(s *String) int8 { return Byte_parseByte_String_int(s, 10) }

// Byte_parseByte_String_int implements Byte.parseByte(String, int).
func Byte_parseByte_String_int(s *String, radix int32) int8 {
	return int8(parseNarrow(s, radix, math.MinInt8, math.MaxInt8))
}

// Byte_decode implements Byte.decode.
func Byte_decode(s *String) any { return int8(decodeNarrow(s, math.MinInt8, math.MaxInt8)) }

// Byte_toString implements Byte.toString(byte).
func Byte_toString(v int8) *String { return StrOfInt(int32(v)) }

// Byte_compare implements Byte.compare.
func Byte_compare(a, b int8) int32 { return int32(a) - int32(b) }

// Byte_hashCode implements Byte.hashCode(byte).
func Byte_hashCode(v int8) int32 { return int32(v) }

// Byte_toUnsignedInt implements Byte.toUnsignedInt.
func Byte_toUnsignedInt(v int8) int32 { return int32(uint8(v)) }

// Byte_toUnsignedLong implements Byte.toUnsignedLong.
func Byte_toUnsignedLong(v int8) int64 { return int64(uint8(v)) }

// Byte_ByteValue implements Byte.byteValue on a boxed receiver.
func Byte_ByteValue(recv any) int8 { return unboxRecv[int8](recv, "java.lang.Byte") }

// Byte_IntValue implements Byte.intValue on a boxed receiver.
func Byte_IntValue(recv any) int32 { return int32(Byte_ByteValue(recv)) }

// Byte_LongValue implements Byte.longValue on a boxed receiver.
func Byte_LongValue(recv any) int64 { return int64(Byte_ByteValue(recv)) }

// Byte_DoubleValue implements Byte.doubleValue on a boxed receiver.
func Byte_DoubleValue(recv any) float64 { return float64(Byte_ByteValue(recv)) }

// Byte_FloatValue implements Byte.floatValue on a boxed receiver.
func Byte_FloatValue(recv any) float32 { return float32(Byte_ByteValue(recv)) }

// Byte_ShortValue implements Byte.shortValue on a boxed receiver.
func Byte_ShortValue(recv any) int16 { return int16(Byte_ByteValue(recv)) }

// Byte_CompareTo_Object implements Byte.compareTo on a boxed receiver.
func Byte_CompareTo_Object(recv, o any) int32 {
	return Byte_compare(Byte_ByteValue(recv), unboxRecv[int8](o, "java.lang.Byte"))
}

// Byte_Equals_Object implements Byte.equals on a boxed receiver.
func Byte_Equals_Object(recv, o any) bool {
	v, ok := o.(int8)
	return ok && Byte_ByteValue(recv) == v
}

// Byte_HashCode implements Byte.hashCode on a boxed receiver.
func Byte_HashCode(recv any) int32 { return int32(Byte_ByteValue(recv)) }

// Byte_ToString implements Byte.toString on a boxed receiver.
func Byte_ToString(recv any) *String { return StrOfInt(int32(Byte_ByteValue(recv))) }

// ---------------------------------------------------------------------------
// Double

// Double_isInstance implements instanceof Double.
func Double_isInstance(o any) bool { _, ok := o.(float64); return ok }

// Double_new_double implements new Double(double).
func Double_new_double(d float64) any { return d }

// Double_new_String implements new Double(String).
func Double_new_String(s *String) any { return parseDouble(s) }

// Double_valueOf_double implements Double.valueOf(double).
func Double_valueOf_double(d float64) any { return d }

// Double_valueOf_String implements Double.valueOf(String).
func Double_valueOf_String(s *String) any { return parseDouble(s) }

// Double_parseDouble implements Double.parseDouble.
func Double_parseDouble(s *String) float64 { return parseDouble(s) }

// Double_toString implements Double.toString(double).
func Double_toString(d float64) *String { return StrOfDouble(d) }

// Double_toString_double is Double_toString under the name with an explicit
// parameter suffix.
func Double_toString_double(d float64) *String { return StrOfDouble(d) }

// Double_toHexString implements Double.toHexString.
func Double_toHexString(d float64) *String { return FromGo(doubleToHexString(d)) }

func doubleToHexString(d float64) string {
	if math.IsInf(d, 0) || d != d {
		return DoubleToString(d)
	}
	b := math.Float64bits(d)
	s := ""
	if b>>63 != 0 {
		s = "-"
	}
	s += "0x"
	b &^= 1 << 63
	if b == 0 {
		return s + "0.0p0"
	}
	subnormal := b < 1<<52
	if subnormal {
		s += "0."
	} else {
		s += "1."
	}
	signif := strconv.FormatUint(b&(1<<52-1)|1<<60, 16)[3:16]
	n := len(signif)
	for n > 1 && signif[n-1] == '0' {
		n--
	}
	s += signif[:n] + "p"
	if subnormal {
		return s + "-1022"
	}
	return s + strconv.Itoa(int(b>>52)-1023)
}

// Double_compare implements Double.compare: -0.0 is less than 0.0, and NaN
// is equal to itself and greater than every other value.
func Double_compare(a, b float64) int32 {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	ab, bb := Double_doubleToLongBits(a), Double_doubleToLongBits(b)
	switch {
	case ab == bb:
		return 0
	case ab < bb:
		return -1
	}
	return 1
}

// Double_isNaN implements Double.isNaN(double).
func Double_isNaN(d float64) bool { return d != d }

// Double_isInfinite implements Double.isInfinite(double).
func Double_isInfinite(d float64) bool { return math.IsInf(d, 0) }

// Double_isFinite implements Double.isFinite.
func Double_isFinite(d float64) bool { return !math.IsInf(d, 0) && d == d }

// Double_doubleToLongBits implements Double.doubleToLongBits, which maps
// every NaN to the canonical NaN.
func Double_doubleToLongBits(d float64) int64 {
	if d != d {
		return 0x7ff8000000000000
	}
	return int64(math.Float64bits(d))
}

// Double_doubleToRawLongBits implements Double.doubleToRawLongBits.
func Double_doubleToRawLongBits(d float64) int64 { return int64(math.Float64bits(d)) }

// Double_longBitsToDouble implements Double.longBitsToDouble.
func Double_longBitsToDouble(b int64) float64 { return math.Float64frombits(uint64(b)) }

// Double_hashCode implements Double.hashCode(double).
func Double_hashCode(d float64) int32 {
	b := Double_doubleToLongBits(d)
	return int32(b ^ int64(uint64(b)>>32))
}

// Double_hashCode_double is Double_hashCode under the name with an explicit
// parameter suffix.
func Double_hashCode_double(d float64) int32 { return Double_hashCode(d) }

// Double_max implements Double.max, which is Math.max.
func Double_max(a, b float64) float64 {
	if a != a {
		return a
	}
	if a == 0 && b == 0 && math.Float64bits(a) == 1<<63 {
		return b
	}
	if a >= b {
		return a
	}
	return b
}

// Double_min implements Double.min, which is Math.min.
func Double_min(a, b float64) float64 {
	if a != a {
		return a
	}
	if a == 0 && b == 0 && math.Float64bits(b) == 1<<63 {
		return b
	}
	if a <= b {
		return a
	}
	return b
}

// Double_sum implements Double.sum.
func Double_sum(a, b float64) float64 { return a + b }

// Double_DoubleValue implements Double.doubleValue on a boxed receiver.
func Double_DoubleValue(recv any) float64 { return unboxRecv[float64](recv, "java.lang.Double") }

// Double_FloatValue implements Double.floatValue on a boxed receiver.
func Double_FloatValue(recv any) float32 { return float32(Double_DoubleValue(recv)) }

// Double_IntValue implements Double.intValue on a boxed receiver.
func Double_IntValue(recv any) int32 { return F2I(Double_DoubleValue(recv)) }

// Double_LongValue implements Double.longValue on a boxed receiver.
func Double_LongValue(recv any) int64 { return F2L(Double_DoubleValue(recv)) }

// Double_ShortValue implements Double.shortValue on a boxed receiver.
func Double_ShortValue(recv any) int16 { return int16(F2I(Double_DoubleValue(recv))) }

// Double_ByteValue implements Double.byteValue on a boxed receiver.
func Double_ByteValue(recv any) int8 { return int8(F2I(Double_DoubleValue(recv))) }

// Double_IsNaN implements Double.isNaN() on a boxed receiver.
func Double_IsNaN(recv any) bool { d := Double_DoubleValue(recv); return d != d }

// Double_IsInfinite implements Double.isInfinite() on a boxed receiver.
func Double_IsInfinite(recv any) bool { return math.IsInf(Double_DoubleValue(recv), 0) }

// Double_CompareTo_Object implements Double.compareTo on a boxed receiver.
func Double_CompareTo_Object(recv, o any) int32 {
	return Double_compare(Double_DoubleValue(recv), unboxRecv[float64](o, "java.lang.Double"))
}

// Double_Equals_Object implements Double.equals on a boxed receiver, which
// compares doubleToLongBits.
func Double_Equals_Object(recv, o any) bool {
	v, ok := o.(float64)
	return ok && Double_doubleToLongBits(Double_DoubleValue(recv)) == Double_doubleToLongBits(v)
}

// Double_HashCode implements Double.hashCode on a boxed receiver.
func Double_HashCode(recv any) int32 { return Double_hashCode(Double_DoubleValue(recv)) }

// Double_ToString implements Double.toString on a boxed receiver.
func Double_ToString(recv any) *String { return StrOfDouble(Double_DoubleValue(recv)) }

// ---------------------------------------------------------------------------
// Float

// Float_isInstance implements instanceof Float.
func Float_isInstance(o any) bool { _, ok := o.(float32); return ok }

// Float_new_float implements new Float(float).
func Float_new_float(f float32) any { return f }

// Float_new_double implements new Float(double).
func Float_new_double(d float64) any { return float32(d) }

// Float_new_String implements new Float(String).
func Float_new_String(s *String) any { return parseFloat(s) }

// Float_valueOf_float implements Float.valueOf(float).
func Float_valueOf_float(f float32) any { return f }

// Float_valueOf_String implements Float.valueOf(String).
func Float_valueOf_String(s *String) any { return parseFloat(s) }

// Float_parseFloat implements Float.parseFloat.
func Float_parseFloat(s *String) float32 { return parseFloat(s) }

// Float_toString implements Float.toString(float).
func Float_toString(f float32) *String { return StrOfFloat(f) }

// Float_toString_float is Float_toString under the name with an explicit
// parameter suffix.
func Float_toString_float(f float32) *String { return StrOfFloat(f) }

// Float_toHexString implements Float.toHexString.
func Float_toHexString(f float32) *String {
	a := math.Abs(float64(f))
	if a < 0x1p-126 && f != 0 {
		s := doubleToHexString(float64(f) * 0x1p-896)
		return FromGo(s[:len(s)-len("p-1022")] + "p-126")
	}
	return FromGo(doubleToHexString(float64(f)))
}

// Float_compare implements Float.compare.
func Float_compare(a, b float32) int32 {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	ab, bb := Float_floatToIntBits(a), Float_floatToIntBits(b)
	switch {
	case ab == bb:
		return 0
	case ab < bb:
		return -1
	}
	return 1
}

// Float_isNaN implements Float.isNaN(float).
func Float_isNaN(f float32) bool { return f != f }

// Float_isInfinite implements Float.isInfinite(float).
func Float_isInfinite(f float32) bool { return math.IsInf(float64(f), 0) }

// Float_isFinite implements Float.isFinite.
func Float_isFinite(f float32) bool { return !math.IsInf(float64(f), 0) && f == f }

// Float_floatToIntBits implements Float.floatToIntBits, which maps every NaN
// to the canonical NaN.
func Float_floatToIntBits(f float32) int32 {
	if f != f {
		return 0x7fc00000
	}
	return int32(math.Float32bits(f))
}

// Float_floatToRawIntBits implements Float.floatToRawIntBits.
func Float_floatToRawIntBits(f float32) int32 { return int32(math.Float32bits(f)) }

// Float_intBitsToFloat implements Float.intBitsToFloat.
func Float_intBitsToFloat(b int32) float32 { return math.Float32frombits(uint32(b)) }

// Float_hashCode implements Float.hashCode(float).
func Float_hashCode(f float32) int32 { return Float_floatToIntBits(f) }

// Float_max implements Float.max, which is Math.max.
func Float_max(a, b float32) float32 {
	if a != a {
		return a
	}
	if a == 0 && b == 0 && math.Float32bits(a) == 1<<31 {
		return b
	}
	if a >= b {
		return a
	}
	return b
}

// Float_min implements Float.min, which is Math.min.
func Float_min(a, b float32) float32 {
	if a != a {
		return a
	}
	if a == 0 && b == 0 && math.Float32bits(b) == 1<<31 {
		return b
	}
	if a <= b {
		return a
	}
	return b
}

// Float_sum implements Float.sum.
func Float_sum(a, b float32) float32 { return a + b }

// Float_FloatValue implements Float.floatValue on a boxed receiver.
func Float_FloatValue(recv any) float32 { return unboxRecv[float32](recv, "java.lang.Float") }

// Float_DoubleValue implements Float.doubleValue on a boxed receiver.
func Float_DoubleValue(recv any) float64 { return float64(Float_FloatValue(recv)) }

// Float_IntValue implements Float.intValue on a boxed receiver.
func Float_IntValue(recv any) int32 { return F2I(float64(Float_FloatValue(recv))) }

// Float_LongValue implements Float.longValue on a boxed receiver.
func Float_LongValue(recv any) int64 { return F2L(float64(Float_FloatValue(recv))) }

// Float_ShortValue implements Float.shortValue on a boxed receiver.
func Float_ShortValue(recv any) int16 { return int16(Float_IntValue(recv)) }

// Float_ByteValue implements Float.byteValue on a boxed receiver.
func Float_ByteValue(recv any) int8 { return int8(Float_IntValue(recv)) }

// Float_IsNaN implements Float.isNaN() on a boxed receiver.
func Float_IsNaN(recv any) bool { f := Float_FloatValue(recv); return f != f }

// Float_IsInfinite implements Float.isInfinite() on a boxed receiver.
func Float_IsInfinite(recv any) bool { return math.IsInf(float64(Float_FloatValue(recv)), 0) }

// Float_CompareTo_Object implements Float.compareTo on a boxed receiver.
func Float_CompareTo_Object(recv, o any) int32 {
	return Float_compare(Float_FloatValue(recv), unboxRecv[float32](o, "java.lang.Float"))
}

// Float_Equals_Object implements Float.equals on a boxed receiver, which
// compares floatToIntBits.
func Float_Equals_Object(recv, o any) bool {
	v, ok := o.(float32)
	return ok && Float_floatToIntBits(Float_FloatValue(recv)) == Float_floatToIntBits(v)
}

// Float_HashCode implements Float.hashCode on a boxed receiver.
func Float_HashCode(recv any) int32 { return Float_floatToIntBits(Float_FloatValue(recv)) }

// Float_ToString implements Float.toString on a boxed receiver.
func Float_ToString(recv any) *String { return StrOfFloat(Float_FloatValue(recv)) }

// ---------------------------------------------------------------------------
// Boolean

// Boolean_TRUE is Boolean.TRUE.
var Boolean_TRUE any = true

// Boolean_FALSE is Boolean.FALSE.
var Boolean_FALSE any = false

// Boolean_isInstance implements instanceof Boolean.
func Boolean_isInstance(o any) bool { _, ok := o.(bool); return ok }

// Boolean_new_boolean implements new Boolean(boolean).
func Boolean_new_boolean(b bool) any { return b }

// Boolean_new_String implements new Boolean(String).
func Boolean_new_String(s *String) any { return Boolean_parseBoolean(s) }

// Boolean_valueOf_boolean implements Boolean.valueOf(boolean).
func Boolean_valueOf_boolean(b bool) any { return b }

// Boolean_valueOf_String implements Boolean.valueOf(String).
func Boolean_valueOf_String(s *String) any { return Boolean_parseBoolean(s) }

// Boolean_parseBoolean implements Boolean.parseBoolean: true exactly when s
// equals "true" ignoring case.
func Boolean_parseBoolean(s *String) bool {
	return s != nil && s.EqualsIgnoreCase_String(Lit("true"))
}

// Boolean_toString implements Boolean.toString(boolean).
func Boolean_toString(b bool) *String { return StrOfBool(b) }

// Boolean_compare implements Boolean.compare.
func Boolean_compare(a, b bool) int32 {
	switch {
	case a == b:
		return 0
	case a:
		return 1
	}
	return -1
}

// Boolean_hashCode implements Boolean.hashCode(boolean).
func Boolean_hashCode(b bool) int32 {
	if b {
		return 1231
	}
	return 1237
}

// Boolean_logicalAnd implements Boolean.logicalAnd.
func Boolean_logicalAnd(a, b bool) bool { return a && b }

// Boolean_logicalOr implements Boolean.logicalOr.
func Boolean_logicalOr(a, b bool) bool { return a || b }

// Boolean_logicalXor implements Boolean.logicalXor.
func Boolean_logicalXor(a, b bool) bool { return a != b }

// Boolean_BooleanValue implements Boolean.booleanValue on a boxed receiver.
func Boolean_BooleanValue(recv any) bool { return unboxRecv[bool](recv, "java.lang.Boolean") }

// Boolean_CompareTo_Object implements Boolean.compareTo on a boxed receiver.
func Boolean_CompareTo_Object(recv, o any) int32 {
	return Boolean_compare(Boolean_BooleanValue(recv), unboxRecv[bool](o, "java.lang.Boolean"))
}

// Boolean_Equals_Object implements Boolean.equals on a boxed receiver.
func Boolean_Equals_Object(recv, o any) bool {
	v, ok := o.(bool)
	return ok && Boolean_BooleanValue(recv) == v
}

// Boolean_HashCode implements Boolean.hashCode on a boxed receiver.
func Boolean_HashCode(recv any) int32 { return Boolean_hashCode(Boolean_BooleanValue(recv)) }

// Boolean_ToString implements Boolean.toString on a boxed receiver.
func Boolean_ToString(recv any) *String { return StrOfBool(Boolean_BooleanValue(recv)) }

// systemProperty implements the lookup of Integer.getInteger,
// Long.getLong and Boolean.getBoolean: System.getProperty, with null for a
// name it rejects.
func systemProperty(name *String) (v *String) {
	defer func() {
		if r := recover(); r != nil {
			if !NullPointerException_isInstance(r) && !IllegalArgumentException_isInstance(r) {
				panic(r)
			}
			v = nil
		}
	}()
	return System_getProperty_String(name)
}

// decodeProperty decodes the system property name with Integer.decode or
// Long.decode, returning ok false when it is unset or malformed.
func decodeProperty(name *String, min, max int64) (v int64, ok bool) {
	s := systemProperty(name)
	if s == nil {
		return 0, false
	}
	defer func() {
		if r := recover(); r != nil {
			if !NumberFormatException_isInstance(r) {
				panic(r)
			}
			ok = false
		}
	}()
	return decode(s, min, max), true
}

// Boolean_getBoolean implements Boolean.getBoolean.
func Boolean_getBoolean(name *String) bool { return Boolean_parseBoolean(systemProperty(name)) }

// Integer_getInteger_String implements Integer.getInteger(String).
func Integer_getInteger_String(name *String) any { return Integer_getInteger_String_Integer(name, nil) }

// Integer_getInteger_String_int implements Integer.getInteger(String, int).
func Integer_getInteger_String_int(name *String, val int32) any {
	if v := Integer_getInteger_String_Integer(name, nil); v != nil {
		return v
	}
	return val
}

// Integer_getInteger_String_Integer implements Integer.getInteger(String,
// Integer).
func Integer_getInteger_String_Integer(name *String, val any) any {
	if v, ok := decodeProperty(name, math.MinInt32, math.MaxInt32); ok {
		return int32(v)
	}
	return val
}

// Long_getLong_String implements Long.getLong(String).
func Long_getLong_String(name *String) any { return Long_getLong_String_Long(name, nil) }

// Long_getLong_String_long implements Long.getLong(String, long).
func Long_getLong_String_long(name *String, val int64) any {
	if v := Long_getLong_String_Long(name, nil); v != nil {
		return v
	}
	return val
}

// Long_getLong_String_Long implements Long.getLong(String, Long).
func Long_getLong_String_Long(name *String, val any) any {
	if v, ok := decodeProperty(name, math.MinInt64, math.MaxInt64); ok {
		return v
	}
	return val
}
