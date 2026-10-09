package java

import (
	"fmt"
	"math"
)

// This file implements java.lang.StringBuilder and java.lang.StringBuffer,
// which share their behaviour (java.lang.AbstractStringBuilder in the JDK),
// and the CharSequence interface methods called on receivers of static type
// CharSequence.

// strBuf is the state of an AbstractStringBuilder. capacity tracks the
// length of the JDK's backing array so that capacity() matches Java.
type strBuf struct {
	units    []uint16
	capacity int
}

func newStrBuf(capacity int32) strBuf {
	if capacity < 0 {
		panic(NegativeArraySizeException_new_String(StrOfInt(capacity)))
	}
	return strBuf{capacity: int(capacity)}
}

// newStrBufLen implements the AbstractStringBuilder constructors taking a
// String or CharSequence of length n.
func newStrBufLen(n int) strBuf {
	c := n + 16
	if n >= math.MaxInt32-16 {
		c = math.MaxInt32
	}
	return strBuf{capacity: c}
}

func indexOutOfBounds(index, length int) any {
	return StringIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
		"Index %d out of bounds for length %d", index, length)))
}

func rangeOutOfBounds(from, to, length int) string {
	return fmt.Sprintf("Range [%d, %d) out of bounds for length %d", from, to, length)
}

// checkRangeSIOOBE checks 0 <= from <= to <= length, throwing
// StringIndexOutOfBoundsException.
func checkRangeSIOOBE(from, to, length int) {
	if from < 0 || from > to || to > length {
		panic(StringIndexOutOfBoundsException_new_String(FromGo(rangeOutOfBounds(from, to, length))))
	}
}

// checkOffCountSIOOBE checks that [off, off+count) lies in [0, length), as
// Preconditions.checkFromIndexSize does, throwing
// StringIndexOutOfBoundsException.
func checkOffCountSIOOBE(off, count, length int) {
	if off < 0 || count < 0 || length < 0 || count > length-off {
		panic(StringIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"Range [%d, %d + %d) out of bounds for length %d", off, off, count, length))))
	}
}

// checkRangeIOOBE checks 0 <= from <= to <= length, throwing
// IndexOutOfBoundsException.
func checkRangeIOOBE(from, to, length int) {
	if from < 0 || from > to || to > length {
		panic(IndexOutOfBoundsException_new_String(FromGo(rangeOutOfBounds(from, to, length))))
	}
}

func (b *strBuf) checkIndex(index int32) {
	if index < 0 || int(index) >= len(b.units) {
		panic(indexOutOfBounds(int(index), len(b.units)))
	}
}

func (b *strBuf) checkOffset(offset int32) {
	checkRangeSIOOBE(int(offset), len(b.units), len(b.units))
}

// ensureCapacityInternal grows the capacity as the JDK does: to the larger of
// the minimum and twice the old capacity plus two.
func (b *strBuf) ensureCapacityInternal(min int) {
	if min > b.capacity {
		c := 2*b.capacity + 2
		if c < min {
			c = min
		}
		b.capacity = c
	}
}

func (b *strBuf) appendUnits(u []uint16) {
	b.ensureCapacityInternal(len(b.units) + len(u))
	b.units = append(b.units, u...)
}

func (b *strBuf) appendGo(s string) {
	b.ensureCapacityInternal(len(b.units) + len(s))
	for i := 0; i < len(s); i++ {
		b.units = append(b.units, uint16(s[i]))
	}
}

func (b *strBuf) appendString(s *String) {
	if s == nil {
		b.appendGo("null")
		return
	}
	b.appendUnits(s.units)
}

func (b *strBuf) appendCharSequence(s any) {
	if s == nil {
		b.appendGo("null")
		return
	}
	b.appendUnits(charSequenceUnits(s))
}

func (b *strBuf) appendCharSequenceRange(s any, start, end int32) {
	var u []uint16
	if s == nil {
		u = Lit("null").units
	} else {
		u = charSequenceUnits(s)
	}
	checkRangeIOOBE(int(start), int(end), len(u))
	b.appendUnits(u[start:end])
}

func (b *strBuf) appendChars(cs []uint16, offset, n int32) {
	end := offset + n
	checkRangeIOOBE(int(offset), int(end), len(cs))
	b.appendUnits(cs[offset:end])
}

func (b *strBuf) appendCodePoint(cp int32) {
	b.appendUnits(Character_toChars_int(cp))
}

func (b *strBuf) insertUnits(offset int32, u []uint16) {
	b.ensureCapacityInternal(len(b.units) + len(u))
	b.units = append(b.units, u...)
	copy(b.units[int(offset)+len(u):], b.units[offset:len(b.units)-len(u)])
	copy(b.units[offset:], u)
}

func (b *strBuf) insertString(offset int32, s *String) {
	b.checkOffset(offset)
	if s == nil {
		s = Lit("null")
	}
	b.insertUnits(offset, s.units)
}

func (b *strBuf) insertChars(index int32, cs []uint16, offset, n int32) {
	b.checkOffset(index)
	checkRangeSIOOBE(int(offset), int(offset+n), len(cs))
	b.insertUnits(index, append([]uint16(nil), cs[offset:offset+n]...))
}

func (b *strBuf) insertCharSequence(dst int32, s any) {
	if s == nil {
		s = Lit("null")
	}
	if str, ok := s.(*String); ok {
		b.insertString(dst, str)
		return
	}
	u := charSequenceUnits(s)
	b.insertCharSequenceRange(dst, s, 0, int32(len(u)))
}

func (b *strBuf) insertCharSequenceRange(dst int32, s any, start, end int32) {
	if s == nil {
		s = Lit("null")
	}
	b.checkOffset(dst)
	u := charSequenceUnits(s)
	checkRangeIOOBE(int(start), int(end), len(u))
	b.insertUnits(dst, append([]uint16(nil), u[start:end]...))
}

func (b *strBuf) delete(start, end int32) {
	count := int32(len(b.units))
	if end > count {
		end = count
	}
	checkRangeSIOOBE(int(start), int(end), int(count))
	if end > start {
		b.units = append(b.units[:start], b.units[end:]...)
	}
}

func (b *strBuf) deleteCharAt(index int32) {
	b.checkIndex(index)
	b.units = append(b.units[:index], b.units[index+1:]...)
}

func (b *strBuf) replace(start, end int32, s *String) {
	count := int32(len(b.units))
	if end > count {
		end = count
	}
	checkRangeSIOOBE(int(start), int(end), int(count))
	if s == nil {
		panic(NullPointerException_new())
	}
	newCount := len(b.units) + len(s.units) - int(end-start)
	b.ensureCapacityInternal(newCount)
	rest := append([]uint16(nil), b.units[end:]...)
	b.units = append(append(b.units[:start], s.units...), rest...)
}

func (b *strBuf) reverse() {
	u := b.units
	hasSurr := false
	for i, j := 0, len(u)-1; i < j; i, j = i+1, j-1 {
		if Character_isSurrogate(u[i]) || Character_isSurrogate(u[j]) {
			hasSurr = true
		}
		u[i], u[j] = u[j], u[i]
	}
	if hasSurr {
		for i := 0; i < len(u)-1; i++ {
			if Character_isLowSurrogate(u[i]) && Character_isHighSurrogate(u[i+1]) {
				u[i], u[i+1] = u[i+1], u[i]
				i++
			}
		}
	}
}

func (b *strBuf) setLength(n int32) {
	if n < 0 {
		panic(StringIndexOutOfBoundsException_new_int(n))
	}
	b.ensureCapacityInternal(int(n))
	for len(b.units) < int(n) {
		b.units = append(b.units, 0)
	}
	b.units = b.units[:n]
}

func (b *strBuf) substring(start, end int32) *String {
	checkRangeSIOOBE(int(start), int(end), len(b.units))
	return FromUnits(append([]uint16(nil), b.units[start:end]...))
}

func (b *strBuf) codePointAt(index int32) int32 {
	b.checkIndex(index)
	return codePointAtUnits(b.units, int(index), len(b.units))
}

func (b *strBuf) codePointBefore(index int32) int32 {
	b.checkIndex(index - 1)
	return codePointBeforeUnits(b.units, int(index), 0)
}

func (b *strBuf) codePointCount(begin, end int32) int32 {
	checkRangeIOOBE(int(begin), int(end), len(b.units))
	return codePointCountUnits(b.units, int(begin), int(end))
}

func (b *strBuf) offsetByCodePoints(index, offset int32) int32 {
	if index < 0 || int(index) > len(b.units) {
		panic(IndexOutOfBoundsException_new())
	}
	return offsetByCodePointsUnits(b.units, 0, len(b.units), int(index), int(offset))
}

// offsetByCodePointsUnits implements Character.offsetByCodePoints on
// u[start:start+count].
func offsetByCodePointsUnits(u []uint16, start, count, index, offset int) int32 {
	x := index
	if offset >= 0 {
		limit := start + count
		i := 0
		for ; x < limit && i < offset; i++ {
			x++
			if Character_isHighSurrogate(u[x-1]) && x < limit && Character_isLowSurrogate(u[x]) {
				x++
			}
		}
		if i < offset {
			panic(IndexOutOfBoundsException_new())
		}
	} else {
		i := offset
		for ; x > start && i < 0; i++ {
			x--
			if Character_isLowSurrogate(u[x]) && x > start && Character_isHighSurrogate(u[x-1]) {
				x--
			}
		}
		if i < 0 {
			panic(IndexOutOfBoundsException_new())
		}
	}
	return int32(x)
}

func (b *strBuf) getChars(srcBegin, srcEnd int32, dst []uint16, dstBegin int32) {
	checkRangeSIOOBE(int(srcBegin), int(srcEnd), len(b.units))
	n := srcEnd - srcBegin
	checkRangeIOOBE(int(dstBegin), int(dstBegin+n), len(dst))
	copy(dst[dstBegin:], b.units[srcBegin:srcEnd])
}

func (b *strBuf) ensureCapacity(min int32) {
	if min > 0 {
		b.ensureCapacityInternal(int(min))
	}
}

func (b *strBuf) trimToSize() {
	if len(b.units) < b.capacity {
		b.capacity = len(b.units)
	}
}

func (b *strBuf) toString() *String {
	return FromUnits(append([]uint16(nil), b.units...))
}

// charSequenceUnits returns the UTF-16 units of a CharSequence, which the
// caller must not modify.
func charSequenceUnits(o any) []uint16 {
	switch x := o.(type) {
	case nil:
		panic(NullPointerException_new())
	case *String:
		if x == nil {
			panic(NullPointerException_new())
		}
		return x.units
	case *StringBuilder:
		return x.b.units
	case *StringBuffer:
		return x.b.units
	}
	panic(ClassCastError(o, "java.lang.CharSequence"))
}

// CharSequence_isInstance implements instanceof CharSequence.
func CharSequence_isInstance(o any) bool {
	switch o.(type) {
	case *String, *StringBuilder, *StringBuffer:
		return true
	}
	return false
}

// CharSequence_Length implements CharSequence.length.
func CharSequence_Length(recv any) int32 { return int32(len(charSequenceUnits(recv))) }

// CharSequence_CharAt_int implements CharSequence.charAt.
func CharSequence_CharAt_int(recv any, i int32) uint16 {
	if s, ok := recv.(*String); ok {
		return s.CharAt_int(i)
	}
	u := charSequenceUnits(recv)
	if i < 0 || int(i) >= len(u) {
		panic(indexOutOfBounds(int(i), len(u)))
	}
	return u[i]
}

// CharSequence_SubSequence_int_int implements CharSequence.subSequence.
func CharSequence_SubSequence_int_int(recv any, start, end int32) any {
	if s, ok := recv.(*String); ok {
		return s.Substring_int_int(start, end)
	}
	u := charSequenceUnits(recv)
	checkRangeSIOOBE(int(start), int(end), len(u))
	return FromUnits(append([]uint16(nil), u[start:end]...))
}

// CharSequence_ToString implements CharSequence.toString.
func CharSequence_ToString(recv any) *String {
	if s, ok := recv.(*String); ok {
		return s
	}
	return FromUnits(append([]uint16(nil), charSequenceUnits(recv)...))
}

// ---------------------------------------------------------------------------
// StringBuilder

// StringBuilder is java.lang.StringBuilder.
type StringBuilder struct {
	b strBuf
}

func (this *StringBuilder) JClass_StringBuilder() *StringBuilder { return this }
func (this *StringBuilder) JClassName() string                   { return "java.lang.StringBuilder" }

// StringBuilder_isInstance implements instanceof StringBuilder.
func StringBuilder_isInstance(o any) bool { _, ok := o.(*StringBuilder); return ok }

// StringBuilder_fromAny casts o to StringBuilder.
func StringBuilder_fromAny(o any) *StringBuilder { return AnyTo[*StringBuilder](o) }

// StringBuilder_new implements new StringBuilder().
func StringBuilder_new() *StringBuilder { return &StringBuilder{b: newStrBuf(16)} }

// StringBuilder_new_int implements new StringBuilder(int).
func StringBuilder_new_int(capacity int32) *StringBuilder {
	return &StringBuilder{b: newStrBuf(capacity)}
}

// StringBuilder_new_String implements new StringBuilder(String).
func StringBuilder_new_String(s *String) *StringBuilder {
	if s == nil {
		panic(NullPointerException_new())
	}
	this := &StringBuilder{b: newStrBufLen(len(s.units))}
	this.b.appendString(s)
	return this
}

// StringBuilder_new_CharSequence implements new StringBuilder(CharSequence).
func StringBuilder_new_CharSequence(s any) *StringBuilder {
	u := charSequenceUnits(s)
	this := &StringBuilder{b: newStrBufLen(len(u))}
	this.b.appendUnits(u)
	return this
}

func (this *StringBuilder) Append_Object(o any) *StringBuilder {
	this.b.appendString(StrOfObject(o))
	return this
}

func (this *StringBuilder) Append_String(s *String) *StringBuilder {
	this.b.appendString(s)
	return this
}

func (this *StringBuilder) Append_StringBuffer(s *StringBuffer) *StringBuilder {
	if s == nil {
		this.b.appendGo("null")
	} else {
		this.b.appendUnits(s.b.units)
	}
	return this
}

func (this *StringBuilder) Append_CharSequence(s any) *StringBuilder {
	this.b.appendCharSequence(s)
	return this
}

func (this *StringBuilder) Append_CharSequence_int_int(s any, start, end int32) *StringBuilder {
	this.b.appendCharSequenceRange(s, start, end)
	return this
}

func (this *StringBuilder) Append_charArr(cs []uint16) *StringBuilder {
	if cs == nil {
		panic(NullPointerException_new())
	}
	this.b.appendUnits(cs)
	return this
}

func (this *StringBuilder) Append_charArr_int_int(cs []uint16, offset, n int32) *StringBuilder {
	if cs == nil {
		panic(NullPointerException_new())
	}
	this.b.appendChars(cs, offset, n)
	return this
}

func (this *StringBuilder) Append_boolean(v bool) *StringBuilder {
	this.b.appendString(StrOfBool(v))
	return this
}

func (this *StringBuilder) Append_char(c uint16) *StringBuilder {
	this.b.appendUnits([]uint16{c})
	return this
}

func (this *StringBuilder) Append_int(v int32) *StringBuilder {
	this.b.appendString(StrOfInt(v))
	return this
}

func (this *StringBuilder) Append_long(v int64) *StringBuilder {
	this.b.appendString(StrOfLong(v))
	return this
}

func (this *StringBuilder) Append_float(v float32) *StringBuilder {
	this.b.appendGo(FloatToString(v))
	return this
}

func (this *StringBuilder) Append_double(v float64) *StringBuilder {
	this.b.appendGo(DoubleToString(v))
	return this
}

func (this *StringBuilder) AppendCodePoint_int(cp int32) *StringBuilder {
	this.b.appendCodePoint(cp)
	return this
}

func (this *StringBuilder) Insert_int_Object(offset int32, o any) *StringBuilder {
	this.b.insertString(offset, StrOfObject(o))
	return this
}

func (this *StringBuilder) Insert_int_String(offset int32, s *String) *StringBuilder {
	this.b.insertString(offset, s)
	return this
}

func (this *StringBuilder) Insert_int_charArr(offset int32, cs []uint16) *StringBuilder {
	this.b.checkOffset(offset)
	if cs == nil {
		panic(NullPointerException_new())
	}
	this.b.insertUnits(offset, append([]uint16(nil), cs...))
	return this
}

func (this *StringBuilder) Insert_int_charArr_int_int(index int32, cs []uint16, offset, n int32) *StringBuilder {
	if cs == nil {
		panic(NullPointerException_new())
	}
	this.b.insertChars(index, cs, offset, n)
	return this
}

func (this *StringBuilder) Insert_int_CharSequence(offset int32, s any) *StringBuilder {
	this.b.insertCharSequence(offset, s)
	return this
}

func (this *StringBuilder) Insert_int_CharSequence_int_int(offset int32, s any, start, end int32) *StringBuilder {
	this.b.insertCharSequenceRange(offset, s, start, end)
	return this
}

func (this *StringBuilder) Insert_int_boolean(offset int32, v bool) *StringBuilder {
	this.b.insertString(offset, StrOfBool(v))
	return this
}

func (this *StringBuilder) Insert_int_char(offset int32, c uint16) *StringBuilder {
	this.b.checkOffset(offset)
	this.b.insertUnits(offset, []uint16{c})
	return this
}

func (this *StringBuilder) Insert_int_int(offset int32, v int32) *StringBuilder {
	this.b.insertString(offset, StrOfInt(v))
	return this
}

func (this *StringBuilder) Insert_int_long(offset int32, v int64) *StringBuilder {
	this.b.insertString(offset, StrOfLong(v))
	return this
}

func (this *StringBuilder) Insert_int_float(offset int32, v float32) *StringBuilder {
	this.b.insertString(offset, StrOfFloat(v))
	return this
}

func (this *StringBuilder) Insert_int_double(offset int32, v float64) *StringBuilder {
	this.b.insertString(offset, StrOfDouble(v))
	return this
}

func (this *StringBuilder) Delete_int_int(start, end int32) *StringBuilder {
	this.b.delete(start, end)
	return this
}

func (this *StringBuilder) DeleteCharAt_int(index int32) *StringBuilder {
	this.b.deleteCharAt(index)
	return this
}

func (this *StringBuilder) Replace_int_int_String(start, end int32, s *String) *StringBuilder {
	this.b.replace(start, end, s)
	return this
}

func (this *StringBuilder) Reverse() *StringBuilder {
	this.b.reverse()
	return this
}

func (this *StringBuilder) IndexOf_String(s *String) int32 {
	return unitsIndexOf(this.b.units, s.units, 0)
}

func (this *StringBuilder) IndexOf_String_int(s *String, from int32) int32 {
	return unitsIndexOf(this.b.units, s.units, from)
}

func (this *StringBuilder) LastIndexOf_String(s *String) int32 {
	return unitsLastIndexOf(this.b.units, s.units, int32(len(this.b.units)))
}

func (this *StringBuilder) LastIndexOf_String_int(s *String, from int32) int32 {
	return unitsLastIndexOf(this.b.units, s.units, from)
}

func (this *StringBuilder) CharAt_int(index int32) uint16 {
	this.b.checkIndex(index)
	return this.b.units[index]
}

func (this *StringBuilder) SetCharAt_int_char(index int32, c uint16) {
	this.b.checkIndex(index)
	this.b.units[index] = c
}

func (this *StringBuilder) SetLength_int(n int32)        { this.b.setLength(n) }
func (this *StringBuilder) Length() int32                { return int32(len(this.b.units)) }
func (this *StringBuilder) Capacity() int32              { return int32(this.b.capacity) }
func (this *StringBuilder) EnsureCapacity_int(min int32) { this.b.ensureCapacity(min) }
func (this *StringBuilder) TrimToSize()                  { this.b.trimToSize() }
func (this *StringBuilder) ToString() *String            { return this.b.toString() }
func (this *StringBuilder) Substring_int(start int32) *String {
	return this.b.substring(start, int32(len(this.b.units)))
}

func (this *StringBuilder) Substring_int_int(start, end int32) *String {
	return this.b.substring(start, end)
}

func (this *StringBuilder) SubSequence_int_int(start, end int32) any {
	return this.b.substring(start, end)
}

func (this *StringBuilder) CodePointAt_int(index int32) int32 { return this.b.codePointAt(index) }

func (this *StringBuilder) CodePointBefore_int(index int32) int32 {
	return this.b.codePointBefore(index)
}

func (this *StringBuilder) CodePointCount_int_int(begin, end int32) int32 {
	return this.b.codePointCount(begin, end)
}

func (this *StringBuilder) OffsetByCodePoints_int_int(index, offset int32) int32 {
	return this.b.offsetByCodePoints(index, offset)
}

func (this *StringBuilder) GetChars_int_int_charArr_int(srcBegin, srcEnd int32, dst []uint16, dstBegin int32) {
	this.b.getChars(srcBegin, srcEnd, dst, dstBegin)
}

// ---------------------------------------------------------------------------
// StringBuffer

// StringBuffer is java.lang.StringBuffer.
type StringBuffer struct {
	b strBuf
}

func (this *StringBuffer) JClass_StringBuffer() *StringBuffer { return this }
func (this *StringBuffer) JClassName() string                 { return "java.lang.StringBuffer" }

// StringBuffer_isInstance implements instanceof StringBuffer.
func StringBuffer_isInstance(o any) bool { _, ok := o.(*StringBuffer); return ok }

// StringBuffer_fromAny casts o to StringBuffer.
func StringBuffer_fromAny(o any) *StringBuffer { return AnyTo[*StringBuffer](o) }

// StringBuffer_new implements new StringBuffer().
func StringBuffer_new() *StringBuffer { return &StringBuffer{b: newStrBuf(16)} }

// StringBuffer_new_int implements new StringBuffer(int).
func StringBuffer_new_int(capacity int32) *StringBuffer {
	return &StringBuffer{b: newStrBuf(capacity)}
}

// StringBuffer_new_String implements new StringBuffer(String).
func StringBuffer_new_String(s *String) *StringBuffer {
	if s == nil {
		panic(NullPointerException_new())
	}
	this := &StringBuffer{b: newStrBufLen(len(s.units))}
	this.b.appendString(s)
	return this
}

// StringBuffer_new_CharSequence implements new StringBuffer(CharSequence).
func StringBuffer_new_CharSequence(s any) *StringBuffer {
	u := charSequenceUnits(s)
	this := &StringBuffer{b: newStrBufLen(len(u))}
	this.b.appendUnits(u)
	return this
}

func (this *StringBuffer) Append_Object(o any) *StringBuffer {
	this.b.appendString(StrOfObject(o))
	return this
}

func (this *StringBuffer) Append_String(s *String) *StringBuffer {
	this.b.appendString(s)
	return this
}

func (this *StringBuffer) Append_StringBuffer(s *StringBuffer) *StringBuffer {
	if s == nil {
		this.b.appendGo("null")
	} else {
		this.b.appendUnits(s.b.units)
	}
	return this
}

func (this *StringBuffer) Append_CharSequence(s any) *StringBuffer {
	this.b.appendCharSequence(s)
	return this
}

func (this *StringBuffer) Append_CharSequence_int_int(s any, start, end int32) *StringBuffer {
	this.b.appendCharSequenceRange(s, start, end)
	return this
}

func (this *StringBuffer) Append_charArr(cs []uint16) *StringBuffer {
	if cs == nil {
		panic(NullPointerException_new())
	}
	this.b.appendUnits(cs)
	return this
}

func (this *StringBuffer) Append_charArr_int_int(cs []uint16, offset, n int32) *StringBuffer {
	if cs == nil {
		panic(NullPointerException_new())
	}
	this.b.appendChars(cs, offset, n)
	return this
}

func (this *StringBuffer) Append_boolean(v bool) *StringBuffer {
	this.b.appendString(StrOfBool(v))
	return this
}

func (this *StringBuffer) Append_char(c uint16) *StringBuffer {
	this.b.appendUnits([]uint16{c})
	return this
}

func (this *StringBuffer) Append_int(v int32) *StringBuffer {
	this.b.appendString(StrOfInt(v))
	return this
}

func (this *StringBuffer) Append_long(v int64) *StringBuffer {
	this.b.appendString(StrOfLong(v))
	return this
}

func (this *StringBuffer) Append_float(v float32) *StringBuffer {
	this.b.appendGo(FloatToString(v))
	return this
}

func (this *StringBuffer) Append_double(v float64) *StringBuffer {
	this.b.appendGo(DoubleToString(v))
	return this
}

func (this *StringBuffer) AppendCodePoint_int(cp int32) *StringBuffer {
	this.b.appendCodePoint(cp)
	return this
}

func (this *StringBuffer) Insert_int_Object(offset int32, o any) *StringBuffer {
	this.b.insertString(offset, StrOfObject(o))
	return this
}

func (this *StringBuffer) Insert_int_String(offset int32, s *String) *StringBuffer {
	this.b.insertString(offset, s)
	return this
}

func (this *StringBuffer) Insert_int_charArr(offset int32, cs []uint16) *StringBuffer {
	this.b.checkOffset(offset)
	if cs == nil {
		panic(NullPointerException_new())
	}
	this.b.insertUnits(offset, append([]uint16(nil), cs...))
	return this
}

func (this *StringBuffer) Insert_int_charArr_int_int(index int32, cs []uint16, offset, n int32) *StringBuffer {
	if cs == nil {
		panic(NullPointerException_new())
	}
	this.b.insertChars(index, cs, offset, n)
	return this
}

func (this *StringBuffer) Insert_int_CharSequence(offset int32, s any) *StringBuffer {
	this.b.insertCharSequence(offset, s)
	return this
}

func (this *StringBuffer) Insert_int_CharSequence_int_int(offset int32, s any, start, end int32) *StringBuffer {
	this.b.insertCharSequenceRange(offset, s, start, end)
	return this
}

func (this *StringBuffer) Insert_int_boolean(offset int32, v bool) *StringBuffer {
	this.b.insertString(offset, StrOfBool(v))
	return this
}

func (this *StringBuffer) Insert_int_char(offset int32, c uint16) *StringBuffer {
	this.b.checkOffset(offset)
	this.b.insertUnits(offset, []uint16{c})
	return this
}

func (this *StringBuffer) Insert_int_int(offset int32, v int32) *StringBuffer {
	this.b.insertString(offset, StrOfInt(v))
	return this
}

func (this *StringBuffer) Insert_int_long(offset int32, v int64) *StringBuffer {
	this.b.insertString(offset, StrOfLong(v))
	return this
}

func (this *StringBuffer) Insert_int_float(offset int32, v float32) *StringBuffer {
	this.b.insertString(offset, StrOfFloat(v))
	return this
}

func (this *StringBuffer) Insert_int_double(offset int32, v float64) *StringBuffer {
	this.b.insertString(offset, StrOfDouble(v))
	return this
}

func (this *StringBuffer) Delete_int_int(start, end int32) *StringBuffer {
	this.b.delete(start, end)
	return this
}

func (this *StringBuffer) DeleteCharAt_int(index int32) *StringBuffer {
	this.b.deleteCharAt(index)
	return this
}

func (this *StringBuffer) Replace_int_int_String(start, end int32, s *String) *StringBuffer {
	this.b.replace(start, end, s)
	return this
}

func (this *StringBuffer) Reverse() *StringBuffer {
	this.b.reverse()
	return this
}

func (this *StringBuffer) IndexOf_String(s *String) int32 {
	return unitsIndexOf(this.b.units, s.units, 0)
}

func (this *StringBuffer) IndexOf_String_int(s *String, from int32) int32 {
	return unitsIndexOf(this.b.units, s.units, from)
}

func (this *StringBuffer) LastIndexOf_String(s *String) int32 {
	return unitsLastIndexOf(this.b.units, s.units, int32(len(this.b.units)))
}

func (this *StringBuffer) LastIndexOf_String_int(s *String, from int32) int32 {
	return unitsLastIndexOf(this.b.units, s.units, from)
}

func (this *StringBuffer) CharAt_int(index int32) uint16 {
	this.b.checkIndex(index)
	return this.b.units[index]
}

func (this *StringBuffer) SetCharAt_int_char(index int32, c uint16) {
	this.b.checkIndex(index)
	this.b.units[index] = c
}

func (this *StringBuffer) SetLength_int(n int32)        { this.b.setLength(n) }
func (this *StringBuffer) Length() int32                { return int32(len(this.b.units)) }
func (this *StringBuffer) Capacity() int32              { return int32(this.b.capacity) }
func (this *StringBuffer) EnsureCapacity_int(min int32) { this.b.ensureCapacity(min) }
func (this *StringBuffer) TrimToSize()                  { this.b.trimToSize() }
func (this *StringBuffer) ToString() *String            { return this.b.toString() }
func (this *StringBuffer) Substring_int(start int32) *String {
	return this.b.substring(start, int32(len(this.b.units)))
}

func (this *StringBuffer) Substring_int_int(start, end int32) *String {
	return this.b.substring(start, end)
}

func (this *StringBuffer) SubSequence_int_int(start, end int32) any {
	return this.b.substring(start, end)
}

func (this *StringBuffer) CodePointAt_int(index int32) int32 { return this.b.codePointAt(index) }

func (this *StringBuffer) CodePointBefore_int(index int32) int32 {
	return this.b.codePointBefore(index)
}

func (this *StringBuffer) CodePointCount_int_int(begin, end int32) int32 {
	return this.b.codePointCount(begin, end)
}

func (this *StringBuffer) OffsetByCodePoints_int_int(index, offset int32) int32 {
	return this.b.offsetByCodePoints(index, offset)
}

func (this *StringBuffer) GetChars_int_int_charArr_int(srcBegin, srcEnd int32, dst []uint16, dstBegin int32) {
	this.b.getChars(srcBegin, srcEnd, dst, dstBegin)
}

// unitsIndexOf implements String.indexOf(String, int) on UTF-16 units.
func unitsIndexOf(s, t []uint16, from int32) int32 {
	n, m := int32(len(s)), int32(len(t))
	if from < 0 {
		from = 0
	}
	if from >= n {
		if m == 0 {
			return n
		}
		return -1
	}
	for i := from; i+m <= n; i++ {
		if unitsEqual(s[i:i+m], t) {
			return i
		}
	}
	return -1
}

// unitsLastIndexOf implements String.lastIndexOf(String, int) on UTF-16
// units.
func unitsLastIndexOf(s, t []uint16, from int32) int32 {
	right := int32(len(s) - len(t))
	if from > right {
		from = right
	}
	if from < 0 {
		return -1
	}
	for i := from; i >= 0; i-- {
		if unitsEqual(s[i:int(i)+len(t)], t) {
			return i
		}
	}
	return -1
}
