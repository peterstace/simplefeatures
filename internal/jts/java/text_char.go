package java

import (
	"fmt"
	"sort"
)

// This file implements java.lang.Character. Character properties come from
// the tables in text_chardata.go, which are generated from the JDK, so they
// follow the JDK's Unicode version rather than the Go toolchain's.

// General categories, as returned by Character.getType.
const (
	charUnassigned           = 0
	charUppercaseLetter      = 1
	charLowercaseLetter      = 2
	charTitlecaseLetter      = 3
	charModifierLetter       = 4
	charOtherLetter          = 5
	charNonSpacingMark       = 6
	charEnclosingMark        = 7
	charCombiningSpacingMark = 8
	charDecimalDigitNumber   = 9
	charLetterNumber         = 10
	charOtherNumber          = 11
	charSpaceSeparator       = 12
	charLineSeparator        = 13
	charParagraphSeparator   = 14
	charControl              = 15
	charFormat               = 16
	charPrivateUse           = 18
	charSurrogate            = 19
	charDashPunctuation      = 20
	charStartPunctuation     = 21
	charEndPunctuation       = 22
	charConnectorPunctuation = 23
	charOtherPunctuation     = 24
	charMathSymbol           = 25
	charCurrencySymbol       = 26
	charModifierSymbol       = 27
	charOtherSymbol          = 28
)

// charType implements Character.getType(int).
func charType(cp int32) int32 {
	if cp < 0 || cp > 0x10ffff {
		return charUnassigned
	}
	u := uint32(cp) << 5
	i := sort.Search(len(charTypeRuns), func(i int) bool { return charTypeRuns[i]&^31 > u })
	return int32(charTypeRuns[i-1] & 31)
}

// charMap applies case mapping runs (see TextCharDataGen) to cp.
func charMap(runs [][4]int32, cp int32) int32 {
	i := sort.Search(len(runs), func(i int) bool { return runs[i][0] > cp })
	if i == 0 {
		return cp
	}
	r := runs[i-1]
	if cp <= r[1] && (cp-r[0])%r[3] == 0 {
		return cp + r[2]
	}
	return cp
}

// inRanges reports whether cp is in one of the sorted inclusive ranges.
func inRanges(ranges [][2]int32, cp int32) bool {
	i := sort.Search(len(ranges), func(i int) bool { return ranges[i][0] > cp })
	return i > 0 && cp <= ranges[i-1][1]
}

func charIsLetterType(t int32) bool {
	return t >= charUppercaseLetter && t <= charOtherLetter
}

// Character_isInstance implements instanceof Character.
func Character_isInstance(o any) bool { _, ok := o.(uint16); return ok }

// Character_new implements new Character(char).
func Character_new(c uint16) any { return c }

// Character_valueOf implements Character.valueOf(char).
func Character_valueOf(c uint16) any { return c }

// Character_toString implements Character.toString(char).
func Character_toString(c uint16) *String { return StrOfChar(c) }

// Character_compare implements Character.compare.
func Character_compare(a, b uint16) int32 { return int32(a) - int32(b) }

// Character_hashCode implements Character.hashCode(char).
func Character_hashCode(c uint16) int32 { return int32(c) }

// Character_reverseBytes implements Character.reverseBytes.
func Character_reverseBytes(c uint16) uint16 { return c<<8 | c>>8 }

// Character_getType_char implements Character.getType(char).
func Character_getType_char(c uint16) int32 { return charType(int32(c)) }

// Character_getType_int implements Character.getType(int).
func Character_getType_int(cp int32) int32 { return charType(cp) }

// Character_isLetter_char implements Character.isLetter(char).
func Character_isLetter_char(c uint16) bool { return Character_isLetter_int(int32(c)) }

// Character_isLetter_int implements Character.isLetter(int).
func Character_isLetter_int(cp int32) bool { return charIsLetterType(charType(cp)) }

// Character_isDigit_char implements Character.isDigit(char).
func Character_isDigit_char(c uint16) bool { return Character_isDigit_int(int32(c)) }

// Character_isDigit_int implements Character.isDigit(int).
func Character_isDigit_int(cp int32) bool { return charType(cp) == charDecimalDigitNumber }

// Character_isLetterOrDigit_char implements Character.isLetterOrDigit(char).
func Character_isLetterOrDigit_char(c uint16) bool { return Character_isLetterOrDigit_int(int32(c)) }

// Character_isLetterOrDigit_int implements Character.isLetterOrDigit(int).
func Character_isLetterOrDigit_int(cp int32) bool {
	t := charType(cp)
	return charIsLetterType(t) || t == charDecimalDigitNumber
}

// Character_isAlphabetic implements Character.isAlphabetic.
func Character_isAlphabetic(cp int32) bool {
	t := charType(cp)
	return charIsLetterType(t) || t == charLetterNumber || inRanges(charOtherAlphabeticRanges[:], cp)
}

// Character_isIdeographic implements Character.isIdeographic.
func Character_isIdeographic(cp int32) bool { return inRanges(charIdeographicRanges[:], cp) }

// Character_isLowerCase_char implements Character.isLowerCase(char).
func Character_isLowerCase_char(c uint16) bool { return Character_isLowerCase_int(int32(c)) }

// Character_isLowerCase_int implements Character.isLowerCase(int).
func Character_isLowerCase_int(cp int32) bool {
	return charType(cp) == charLowercaseLetter || inRanges(charOtherLowerRanges[:], cp)
}

// Character_isUpperCase_char implements Character.isUpperCase(char).
func Character_isUpperCase_char(c uint16) bool { return Character_isUpperCase_int(int32(c)) }

// Character_isUpperCase_int implements Character.isUpperCase(int).
func Character_isUpperCase_int(cp int32) bool {
	return charType(cp) == charUppercaseLetter || inRanges(charOtherUpperRanges[:], cp)
}

// Character_isTitleCase_char implements Character.isTitleCase(char).
func Character_isTitleCase_char(c uint16) bool { return Character_isTitleCase_int(int32(c)) }

// Character_isTitleCase_int implements Character.isTitleCase(int).
func Character_isTitleCase_int(cp int32) bool { return charType(cp) == charTitlecaseLetter }

// Character_isDefined_char implements Character.isDefined(char).
func Character_isDefined_char(c uint16) bool { return Character_isDefined_int(int32(c)) }

// Character_isDefined_int implements Character.isDefined(int).
func Character_isDefined_int(cp int32) bool { return charType(cp) != charUnassigned }

// Character_isSpaceChar_char implements Character.isSpaceChar(char).
func Character_isSpaceChar_char(c uint16) bool { return Character_isSpaceChar_int(int32(c)) }

// Character_isSpaceChar_int implements Character.isSpaceChar(int).
func Character_isSpaceChar_int(cp int32) bool {
	t := charType(cp)
	return t == charSpaceSeparator || t == charLineSeparator || t == charParagraphSeparator
}

// Character_isWhitespace_char implements Character.isWhitespace(char).
func Character_isWhitespace_char(c uint16) bool { return Character_isWhitespace_int(int32(c)) }

// Character_isWhitespace_int implements Character.isWhitespace(int): space
// separators other than the non-breaking ones, and the ASCII controls
// \t \n \u000B \f \r \u001C-\u001F.
func Character_isWhitespace_int(cp int32) bool {
	switch {
	case cp >= 0x09 && cp <= 0x0d, cp >= 0x1c && cp <= 0x1f:
		return true
	case cp == 0xa0, cp == 0x2007, cp == 0x202f:
		return false
	}
	return Character_isSpaceChar_int(cp)
}

// Character_isSpace implements the deprecated Character.isSpace.
func Character_isSpace(c uint16) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\f' || c == '\r'
}

// Character_isISOControl_char implements Character.isISOControl(char).
func Character_isISOControl_char(c uint16) bool { return Character_isISOControl_int(int32(c)) }

// Character_isISOControl_int implements Character.isISOControl(int).
func Character_isISOControl_int(cp int32) bool {
	return cp <= 0x9f && (cp >= 0x7f || cp>>5 == 0)
}

// Character_isIdentifierIgnorable_char implements
// Character.isIdentifierIgnorable(char).
func Character_isIdentifierIgnorable_char(c uint16) bool {
	return Character_isIdentifierIgnorable_int(int32(c))
}

// Character_isIdentifierIgnorable_int implements
// Character.isIdentifierIgnorable(int).
func Character_isIdentifierIgnorable_int(cp int32) bool {
	return cp >= 0 && cp <= 8 || cp >= 0x0e && cp <= 0x1b || cp >= 0x7f && cp <= 0x9f || charType(cp) == charFormat
}

// Character_isJavaIdentifierStart_char implements
// Character.isJavaIdentifierStart(char).
func Character_isJavaIdentifierStart_char(c uint16) bool {
	return Character_isJavaIdentifierStart_int(int32(c))
}

// Character_isJavaIdentifierStart_int implements
// Character.isJavaIdentifierStart(int).
func Character_isJavaIdentifierStart_int(cp int32) bool {
	t := charType(cp)
	return charIsLetterType(t) || t == charLetterNumber || t == charCurrencySymbol || t == charConnectorPunctuation
}

// Character_isJavaIdentifierPart_char implements
// Character.isJavaIdentifierPart(char).
func Character_isJavaIdentifierPart_char(c uint16) bool {
	return Character_isJavaIdentifierPart_int(int32(c))
}

// Character_isJavaIdentifierPart_int implements
// Character.isJavaIdentifierPart(int).
func Character_isJavaIdentifierPart_int(cp int32) bool {
	switch charType(cp) {
	case charUppercaseLetter, charLowercaseLetter, charTitlecaseLetter, charModifierLetter, charOtherLetter,
		charCurrencySymbol, charConnectorPunctuation, charDecimalDigitNumber, charLetterNumber,
		charCombiningSpacingMark, charNonSpacingMark:
		return true
	}
	return Character_isIdentifierIgnorable_int(cp)
}

// Character_isJavaLetter implements the deprecated Character.isJavaLetter.
func Character_isJavaLetter(c uint16) bool { return Character_isJavaIdentifierStart_char(c) }

// Character_isJavaLetterOrDigit implements the deprecated
// Character.isJavaLetterOrDigit.
func Character_isJavaLetterOrDigit(c uint16) bool { return Character_isJavaIdentifierPart_char(c) }

// Character_toUpperCase_char implements Character.toUpperCase(char).
func Character_toUpperCase_char(c uint16) uint16 {
	if c >= 'a' && c <= 'z' {
		return c - 32
	}
	if c < 0x80 {
		return c
	}
	return uint16(Character_toUpperCase_int(int32(c)))
}

// Character_toUpperCase_int implements Character.toUpperCase(int).
func Character_toUpperCase_int(cp int32) int32 { return charMap(charUpperRuns[:], cp) }

// Character_toLowerCase_char implements Character.toLowerCase(char).
func Character_toLowerCase_char(c uint16) uint16 {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	if c < 0x80 {
		return c
	}
	return uint16(Character_toLowerCase_int(int32(c)))
}

// Character_toLowerCase_int implements Character.toLowerCase(int).
func Character_toLowerCase_int(cp int32) int32 { return charMap(charLowerRuns[:], cp) }

// Character_toTitleCase_char implements Character.toTitleCase(char).
func Character_toTitleCase_char(c uint16) uint16 { return uint16(Character_toTitleCase_int(int32(c))) }

// Character_toTitleCase_int implements Character.toTitleCase(int).
func Character_toTitleCase_int(cp int32) int32 {
	runs := charTitleRuns[:]
	i := sort.Search(len(runs), func(i int) bool { return runs[i][0] > cp })
	if i > 0 && cp <= runs[i-1][1] {
		return cp + runs[i-1][2]
	}
	return Character_toUpperCase_int(cp)
}

// Character_digit_char_int implements Character.digit(char, int).
func Character_digit_char_int(c uint16, radix int32) int32 {
	return Character_digit_int_int(int32(c), radix)
}

// Character_digit_int_int implements Character.digit(int, int).
func Character_digit_int_int(cp, radix int32) int32 {
	if radix < 2 || radix > 36 {
		return -1
	}
	var v int32
	switch {
	case cp >= '0' && cp <= '9':
		v = cp - '0'
	case cp >= 'a' && cp <= 'z':
		v = cp - 'a' + 10
	case cp >= 'A' && cp <= 'Z':
		v = cp - 'A' + 10
	case cp < 0x80:
		return -1
	default:
		runs := charDigitRuns[:]
		i := sort.Search(len(runs), func(i int) bool { return runs[i][0] > cp })
		if i == 0 || cp > runs[i-1][1] {
			return -1
		}
		v = runs[i-1][2] + cp - runs[i-1][0]
	}
	if v < radix {
		return v
	}
	return -1
}

// Character_forDigit implements Character.forDigit.
func Character_forDigit(digit, radix int32) uint16 {
	if digit >= radix || digit < 0 || radix < 2 || radix > 36 {
		return 0
	}
	if digit < 10 {
		return uint16('0' + digit)
	}
	return uint16('a' - 10 + digit)
}

// Character_isHighSurrogate implements Character.isHighSurrogate.
func Character_isHighSurrogate(c uint16) bool { return c >= 0xd800 && c <= 0xdbff }

// Character_isLowSurrogate implements Character.isLowSurrogate.
func Character_isLowSurrogate(c uint16) bool { return c >= 0xdc00 && c <= 0xdfff }

// Character_isSurrogate implements Character.isSurrogate.
func Character_isSurrogate(c uint16) bool { return c >= 0xd800 && c <= 0xdfff }

// Character_isSurrogatePair implements Character.isSurrogatePair.
func Character_isSurrogatePair(high, low uint16) bool {
	return Character_isHighSurrogate(high) && Character_isLowSurrogate(low)
}

// Character_toCodePoint implements Character.toCodePoint.
func Character_toCodePoint(high, low uint16) int32 {
	return int32(high)<<10 + int32(low) + (0x10000 - 0xd800<<10 - 0xdc00)
}

// Character_charCount implements Character.charCount.
func Character_charCount(cp int32) int32 {
	if cp >= 0x10000 {
		return 2
	}
	return 1
}

// Character_isBmpCodePoint implements Character.isBmpCodePoint.
func Character_isBmpCodePoint(cp int32) bool { return uint32(cp)>>16 == 0 }

// Character_isValidCodePoint implements Character.isValidCodePoint.
func Character_isValidCodePoint(cp int32) bool { return uint32(cp) <= 0x10ffff }

// Character_isSupplementaryCodePoint implements
// Character.isSupplementaryCodePoint.
func Character_isSupplementaryCodePoint(cp int32) bool { return cp >= 0x10000 && cp <= 0x10ffff }

// Character_highSurrogate implements Character.highSurrogate.
func Character_highSurrogate(cp int32) uint16 {
	return uint16(uint32(cp)>>10 + (0xd800 - 0x10000>>10))
}

// Character_lowSurrogate implements Character.lowSurrogate.
func Character_lowSurrogate(cp int32) uint16 { return uint16(cp&0x3ff + 0xdc00) }

// Character_toChars_int implements Character.toChars(int).
func Character_toChars_int(cp int32) []uint16 {
	if Character_isBmpCodePoint(cp) {
		return []uint16{uint16(cp)}
	}
	if !Character_isValidCodePoint(cp) {
		panic(IllegalArgumentException_new_String(FromGo(fmt.Sprintf("Not a valid Unicode code point: 0x%X", uint32(cp)))))
	}
	return []uint16{Character_highSurrogate(cp), Character_lowSurrogate(cp)}
}

// Character_toChars_int_charArr_int implements Character.toChars(int,
// char[], int).
func Character_toChars_int_charArr_int(cp int32, dst []uint16, index int32) int32 {
	cs := Character_toChars_int(cp)
	if index < 0 || int(index)+len(cs) > len(dst) {
		panic(ArrayIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"Index %d out of bounds for length %d", int(index)+len(cs)-1, len(dst)))))
	}
	copy(dst[index:], cs)
	return int32(len(cs))
}

// codePointAtUnits implements Character.codePointAt on u with limit.
func codePointAtUnits(u []uint16, index, limit int) int32 {
	c1 := u[index]
	if Character_isHighSurrogate(c1) && index+1 < limit {
		if c2 := u[index+1]; Character_isLowSurrogate(c2) {
			return Character_toCodePoint(c1, c2)
		}
	}
	return int32(c1)
}

// codePointBeforeUnits implements Character.codePointBefore on u with start.
func codePointBeforeUnits(u []uint16, index, start int) int32 {
	index--
	c2 := u[index]
	if Character_isLowSurrogate(c2) && index > start {
		if c1 := u[index-1]; Character_isHighSurrogate(c1) {
			return Character_toCodePoint(c1, c2)
		}
	}
	return int32(c2)
}

// arrayIndexOutOfBounds returns the ArrayIndexOutOfBoundsException the JVM
// throws for an array access.
func arrayIndexOutOfBounds(index, length int) any {
	return ArrayIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
		"Index %d out of bounds for length %d", index, length)))
}

// Character_codePointAt_charArr_int implements Character.codePointAt(char[],
// int).
func Character_codePointAt_charArr_int(a []uint16, index int32) int32 {
	if index < 0 || int(index) >= len(a) {
		panic(arrayIndexOutOfBounds(int(index), len(a)))
	}
	return codePointAtUnits(a, int(index), len(a))
}

// Character_codePointAt_CharSequence_int implements
// Character.codePointAt(CharSequence, int).
func Character_codePointAt_CharSequence_int(seq any, index int32) int32 {
	u := charSequenceUnits(seq)
	if index < 0 || int(index) >= len(u) {
		panic(StringIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"index %d, length %d", index, len(u)))))
	}
	return codePointAtUnits(u, int(index), len(u))
}

// Character_codePointBefore_charArr_int implements
// Character.codePointBefore(char[], int).
func Character_codePointBefore_charArr_int(a []uint16, index int32) int32 {
	if index-1 < 0 || int(index-1) >= len(a) {
		panic(arrayIndexOutOfBounds(int(index-1), len(a)))
	}
	return codePointBeforeUnits(a, int(index), 0)
}

// Character_codePointBefore_CharSequence_int implements
// Character.codePointBefore(CharSequence, int).
func Character_codePointBefore_CharSequence_int(seq any, index int32) int32 {
	u := charSequenceUnits(seq)
	if index < 1 || int(index) > len(u) {
		panic(StringIndexOutOfBoundsException_new_String(FromGo(fmt.Sprintf(
			"index %d, length %d", index-1, len(u)))))
	}
	return codePointBeforeUnits(u, int(index), 0)
}

// codePointCountUnits counts the code points in u[begin:end].
func codePointCountUnits(u []uint16, begin, end int) int32 {
	n := int32(end - begin)
	for i := begin; i < end; {
		if Character_isHighSurrogate(u[i]) && i+1 < end && Character_isLowSurrogate(u[i+1]) {
			n--
			i += 2
		} else {
			i++
		}
	}
	return n
}

// Character_codePointCount_CharSequence_int_int implements
// Character.codePointCount(CharSequence, int, int).
func Character_codePointCount_CharSequence_int_int(seq any, begin, end int32) int32 {
	u := charSequenceUnits(seq)
	if begin < 0 || end > int32(len(u)) || begin > end {
		panic(IndexOutOfBoundsException_new())
	}
	return codePointCountUnits(u, int(begin), int(end))
}

// Character_codePointCount_charArr_int_int implements
// Character.codePointCount(char[], int, int).
func Character_codePointCount_charArr_int_int(a []uint16, offset, count int32) int32 {
	if count < 0 || offset < 0 || offset > int32(len(a))-count {
		panic(IndexOutOfBoundsException_new())
	}
	return codePointCountUnits(a, int(offset), int(offset+count))
}

// Character_CharValue implements Character.charValue on a boxed receiver.
func Character_CharValue(recv any) uint16 { return unboxRecv[uint16](recv, "java.lang.Character") }

// Character_CompareTo_Object implements Character.compareTo on a boxed
// receiver.
func Character_CompareTo_Object(recv, o any) int32 {
	return Character_compare(Character_CharValue(recv), unboxRecv[uint16](o, "java.lang.Character"))
}

// Character_Equals_Object implements Character.equals on a boxed receiver.
func Character_Equals_Object(recv, o any) bool {
	v, ok := o.(uint16)
	return ok && Character_CharValue(recv) == v
}

// Character_HashCode implements Character.hashCode on a boxed receiver.
func Character_HashCode(recv any) int32 { return int32(Character_CharValue(recv)) }

// Character_ToString implements Character.toString on a boxed receiver.
func Character_ToString(recv any) *String { return StrOfChar(Character_CharValue(recv)) }
