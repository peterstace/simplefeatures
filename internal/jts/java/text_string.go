package java

import (
	"unicode/utf16"
)

// This file implements java.lang.String members beyond those in string.go:
// locale-sensitive case conversion, code point access, comparison and
// search variants, replacement, regular expressions (see text_regex.go), and
// the remaining static factories.

// ToUpperCase_Locale implements String.toUpperCase(Locale).
func (s *String) ToUpperCase_Locale(l *Locale) *String {
	if l == nil {
		panic(NullPointerException_new())
	}
	u := javaToUpperCase(s.units, l)
	if unitsEqual(u, s.units) {
		return s
	}
	return FromUnits(u)
}

// ToLowerCase_Locale implements String.toLowerCase(Locale).
func (s *String) ToLowerCase_Locale(l *Locale) *String {
	if l == nil {
		panic(NullPointerException_new())
	}
	u := javaToLowerCase(s.units, l)
	if unitsEqual(u, s.units) {
		return s
	}
	return FromUnits(u)
}

// appendCodePoint appends cp as UTF-16.
func appendCodePoint(out []uint16, cp int32) []uint16 {
	if cp < 0x10000 {
		return append(out, uint16(cp))
	}
	return append(out, Character_highSurrogate(cp), Character_lowSurrogate(cp))
}

// javaToUpperCase implements String.toUpperCase(Locale) on UTF-16 units,
// including the full (multi-character) mappings and the Turkish, Azeri and
// Lithuanian rules of java.lang.ConditionalSpecialCasing.
func javaToUpperCase(u []uint16, l *Locale) []uint16 {
	lang := l.language
	localeDependent := lang == "tr" || lang == "az" || lang == "lt"
	out := make([]uint16, 0, len(u))
	for i := 0; i < len(u); {
		cp := codePointAtUnits(u, i, len(u))
		n := int(Character_charCount(cp))
		if localeDependent {
			switch {
			case cp == 'i' && (lang == "tr" || lang == "az"):
				out = append(out, 0x130)
				i += n
				continue
			case cp == 0x307 && lang == "lt" && isAfterSoftDotted(u, i):
				i += n
				continue
			}
		}
		if special, ok := charSpecialUpper[cp]; ok {
			out = append(out, utf16.Encode([]rune(special))...)
		} else {
			out = appendCodePoint(out, Character_toUpperCase_int(cp))
		}
		i += n
	}
	return out
}

// javaToLowerCase implements String.toLowerCase(Locale) on UTF-16 units.
func javaToLowerCase(u []uint16, l *Locale) []uint16 {
	lang := l.language
	trAz := lang == "tr" || lang == "az"
	lt := lang == "lt"
	out := make([]uint16, 0, len(u))
	for i := 0; i < len(u); {
		cp := codePointAtUnits(u, i, len(u))
		n := int(Character_charCount(cp))
		i0 := i
		i += n
		switch {
		case trAz && cp == 0x130:
			out = append(out, 'i')
			continue
		case trAz && cp == 0x307 && isAfterI(u, i0):
			continue
		case trAz && cp == 'I' && !isBeforeDot(u, i0):
			out = append(out, 0x131)
			continue
		case lt && (cp == 'I' || cp == 'J' || cp == 0x12e) && isMoreAbove(u, i0):
			out = appendCodePoint(out, Character_toLowerCase_int(cp))
			out = append(out, 0x307)
			continue
		case lt && cp == 0xcc:
			out = append(out, 'i', 0x307, 0x300)
			continue
		case lt && cp == 0xcd:
			out = append(out, 'i', 0x307, 0x301)
			continue
		case lt && cp == 0x128:
			out = append(out, 'i', 0x307, 0x303)
			continue
		case cp == 0x130:
			out = append(out, 'i', 0x307)
			continue
		case cp == 0x3a3 && isFinalCased(u, i0):
			out = append(out, 0x3c2)
			continue
		}
		out = appendCodePoint(out, Character_toLowerCase_int(cp))
	}
	return out
}

// combiningClassKind returns 0 for canonical combining class 0, 230 for
// class 230 (Above), and 1 for any other class.
func combiningClassKind(cp int32) int {
	switch {
	case inRanges(charCombiningAboveRanges[:], cp):
		return 230
	case inRanges(charCombiningOtherRanges[:], cp):
		return 1
	}
	return 0
}

// isAfterI implements ConditionalSpecialCasing.isAfterI.
func isAfterI(u []uint16, index int) bool {
	for i := index; i > 0; {
		ch := codePointBeforeUnits(u, i, 0)
		if ch == 'I' {
			return true
		}
		if k := combiningClassKind(ch); k == 0 || k == 230 {
			return false
		}
		i -= int(Character_charCount(ch))
	}
	return false
}

// isBeforeDot implements ConditionalSpecialCasing.isBeforeDot.
func isBeforeDot(u []uint16, index int) bool {
	for i := index + int(Character_charCount(codePointAtUnits(u, index, len(u)))); i < len(u); {
		ch := codePointAtUnits(u, i, len(u))
		if ch == 0x307 {
			return true
		}
		if k := combiningClassKind(ch); k == 0 || k == 230 {
			return false
		}
		i += int(Character_charCount(ch))
	}
	return false
}

// isAfterSoftDotted implements ConditionalSpecialCasing.isAfterSoftDotted.
func isAfterSoftDotted(u []uint16, index int) bool {
	for i := index; i > 0; {
		ch := codePointBeforeUnits(u, i, 0)
		switch ch {
		case 0x69, 0x6a, 0x12f, 0x268, 0x456, 0x458, 0x1e2d, 0x1ecb:
			return true
		}
		if k := combiningClassKind(ch); k == 0 || k == 230 {
			return false
		}
		i -= int(Character_charCount(ch))
	}
	return false
}

// isMoreAbove implements ConditionalSpecialCasing.isMoreAbove.
func isMoreAbove(u []uint16, index int) bool {
	for i := index + int(Character_charCount(codePointAtUnits(u, index, len(u)))); i < len(u); {
		ch := codePointAtUnits(u, i, len(u))
		switch combiningClassKind(ch) {
		case 230:
			return true
		case 0:
			return false
		}
		i += int(Character_charCount(ch))
	}
	return false
}

// isCased implements ConditionalSpecialCasing.isCased.
func isCased(ch int32) bool {
	switch charType(ch) {
	case charLowercaseLetter, charUppercaseLetter, charTitlecaseLetter:
		return true
	}
	return ch >= 0x2b0 && ch <= 0x2b8 || ch >= 0x2c0 && ch <= 0x2c1 || ch >= 0x2e0 && ch <= 0x2e4 ||
		ch == 0x345 || ch == 0x37a || ch >= 0x1d2c && ch <= 0x1d61 || ch >= 0x2160 && ch <= 0x217f ||
		ch >= 0x24b6 && ch <= 0x24e9
}

// isFinalCased implements ConditionalSpecialCasing.isFinalCased: the sigma
// at index follows a cased letter and is not followed by one, within its
// word.
//
// TODO: The JDK finds word boundaries with BreakIterator.getWordInstance;
// this approximates a word as a run of letters, marks and digits.
func isFinalCased(u []uint16, index int) bool {
	isWordChar := func(cp int32) bool {
		t := charType(cp)
		return charIsLetterType(t) || t == charNonSpacingMark || t == charEnclosingMark ||
			t == charCombiningSpacingMark || t == charDecimalDigitNumber
	}
	for i := index; i > 0; {
		ch := codePointBeforeUnits(u, i, 0)
		if !isWordChar(ch) {
			return false
		}
		if isCased(ch) {
			for j := index + int(Character_charCount(codePointAtUnits(u, index, len(u)))); j < len(u); {
				c := codePointAtUnits(u, j, len(u))
				if !isWordChar(c) {
					break
				}
				if isCased(c) {
					return false
				}
				j += int(Character_charCount(c))
			}
			return true
		}
		i -= int(Character_charCount(ch))
	}
	return false
}

// CodePointAt_int implements String.codePointAt.
func (s *String) CodePointAt_int(index int32) int32 {
	if index < 0 || int(index) >= len(s.units) {
		panic(indexOutOfBounds(int(index), len(s.units)))
	}
	return codePointAtUnits(s.units, int(index), len(s.units))
}

// CodePointBefore_int implements String.codePointBefore.
func (s *String) CodePointBefore_int(index int32) int32 {
	i := index - 1
	if i < 0 || int(i) >= len(s.units) {
		panic(indexOutOfBounds(int(i), len(s.units)))
	}
	return codePointBeforeUnits(s.units, int(index), 0)
}

// CodePointCount_int_int implements String.codePointCount.
func (s *String) CodePointCount_int_int(begin, end int32) int32 {
	checkRangeIOOBE(int(begin), int(end), len(s.units))
	return codePointCountUnits(s.units, int(begin), int(end))
}

// OffsetByCodePoints_int_int implements String.offsetByCodePoints.
func (s *String) OffsetByCodePoints_int_int(index, offset int32) int32 {
	if index < 0 || int(index) > len(s.units) {
		panic(IndexOutOfBoundsException_new())
	}
	return offsetByCodePointsUnits(s.units, 0, len(s.units), int(index), int(offset))
}

// GetChars_int_int_charArr_int implements String.getChars.
func (s *String) GetChars_int_int_charArr_int(srcBegin, srcEnd int32, dst []uint16, dstBegin int32) {
	checkRangeSIOOBE(int(srcBegin), int(srcEnd), len(s.units))
	checkOffCountSIOOBE(int(dstBegin), int(srcEnd-srcBegin), len(dst))
	copy(dst[dstBegin:], s.units[srcBegin:srcEnd])
}

// ContentEquals_StringBuffer implements String.contentEquals(StringBuffer).
func (s *String) ContentEquals_StringBuffer(sb *StringBuffer) bool {
	return unitsEqual(s.units, charSequenceUnits(sb))
}

// ContentEquals_CharSequence implements String.contentEquals(CharSequence).
func (s *String) ContentEquals_CharSequence(cs any) bool {
	return unitsEqual(s.units, charSequenceUnits(cs))
}

// CompareToIgnoreCase_String implements String.compareToIgnoreCase.
func (s *String) CompareToIgnoreCase_String(t *String) int32 {
	return compareToCI(s.units, 0, len(s.units), t.units, 0, len(t.units))
}

// compareToCI implements StringUTF16.compareToCIImpl: compare code points
// after folding each to upper and then lower case.
func compareToCI(a []uint16, aoff, alen int, b []uint16, boff, blen int) int32 {
	alast, blast := aoff+alen, boff+blen
	for k1, k2 := aoff, boff; k1 < alast && k2 < blast; k1, k2 = k1+1, k2+1 {
		cp1, cp2 := int32(a[k1]), int32(b[k2])
		if cp1 == cp2 || compareCodePointCI(cp1, cp2) == 0 {
			continue
		}
		cp1 = codePointIncluding(a, cp1, k1, aoff, alast)
		if cp1 < 0 {
			k1++
			cp1 = -cp1
		}
		cp2 = codePointIncluding(b, cp2, k2, boff, blast)
		if cp2 < 0 {
			k2++
			cp2 = -cp2
		}
		if diff := compareCodePointCI(cp1, cp2); diff != 0 {
			return diff
		}
	}
	return int32(alen - blen)
}

func compareCodePointCI(cp1, cp2 int32) int32 {
	cp1 = Character_toUpperCase_int(cp1)
	cp2 = Character_toUpperCase_int(cp2)
	if cp1 != cp2 {
		cp1 = Character_toLowerCase_int(cp1)
		cp2 = Character_toLowerCase_int(cp2)
		if cp1 != cp2 {
			return cp1 - cp2
		}
	}
	return 0
}

// codePointIncluding returns the code point that the surrogate cp at
// u[index] is part of, negated when it continues at index+1.
func codePointIncluding(u []uint16, cp int32, index, start, end int) int32 {
	c := uint16(cp)
	if !Character_isSurrogate(c) {
		return cp
	}
	if Character_isLowSurrogate(c) {
		if index > start && Character_isHighSurrogate(u[index-1]) {
			return Character_toCodePoint(u[index-1], c)
		}
	} else if index+1 < end && Character_isLowSurrogate(u[index+1]) {
		return -Character_toCodePoint(c, u[index+1])
	}
	return cp
}

// RegionMatches_int_String_int_int implements String.regionMatches(int,
// String, int, int).
func (s *String) RegionMatches_int_String_int_int(toffset int32, other *String, ooffset, n int32) bool {
	if ooffset < 0 || toffset < 0 || int64(toffset) > int64(len(s.units))-int64(n) ||
		int64(ooffset) > int64(len(other.units))-int64(n) {
		return false
	}
	if n <= 0 {
		return true
	}
	return unitsEqual(s.units[toffset:toffset+n], other.units[ooffset:ooffset+n])
}

// RegionMatches_boolean_int_String_int_int implements
// String.regionMatches(boolean, int, String, int, int).
func (s *String) RegionMatches_boolean_int_String_int_int(ignoreCase bool, toffset int32, other *String, ooffset, n int32) bool {
	if !ignoreCase {
		return s.RegionMatches_int_String_int_int(toffset, other, ooffset, n)
	}
	if ooffset < 0 || toffset < 0 || int64(toffset) > int64(len(s.units))-int64(n) ||
		int64(ooffset) > int64(len(other.units))-int64(n) {
		return false
	}
	if n <= 0 {
		return true
	}
	return compareToCI(s.units, int(toffset), int(n), other.units, int(ooffset), int(n)) == 0
}

// StartsWith_String_int implements String.startsWith(String, int).
func (s *String) StartsWith_String_int(p *String, offset int32) bool {
	if offset < 0 || int(offset) > len(s.units)-len(p.units) {
		return false
	}
	return unitsEqual(s.units[offset:int(offset)+len(p.units)], p.units)
}

// IndexOf_int_int implements String.indexOf(int, int).
func (s *String) IndexOf_int_int(ch, from int32) int32 {
	if from < 0 {
		from = 0
	}
	if ch >= 0x10000 && ch <= 0x10ffff {
		pair := []uint16{Character_highSurrogate(ch), Character_lowSurrogate(ch)}
		if int(from) >= len(s.units) {
			return -1
		}
		return unitsIndexOf(s.units, pair, from)
	}
	for i := int(from); i < len(s.units); i++ {
		if int32(s.units[i]) == ch {
			return int32(i)
		}
	}
	return -1
}

// LastIndexOf_int_int implements String.lastIndexOf(int, int).
func (s *String) LastIndexOf_int_int(ch, from int32) int32 {
	if int(from) >= len(s.units) {
		from = int32(len(s.units) - 1)
	}
	if ch >= 0x10000 && ch <= 0x10ffff {
		pair := []uint16{Character_highSurrogate(ch), Character_lowSurrogate(ch)}
		return unitsLastIndexOf(s.units, pair, from)
	}
	for i := from; i >= 0; i-- {
		if int32(s.units[i]) == ch {
			return i
		}
	}
	return -1
}

// LastIndexOf_String implements String.lastIndexOf(String).
func (s *String) LastIndexOf_String(t *String) int32 {
	return unitsLastIndexOf(s.units, t.units, int32(len(s.units)))
}

// LastIndexOf_String_int implements String.lastIndexOf(String, int).
func (s *String) LastIndexOf_String_int(t *String, from int32) int32 {
	return unitsLastIndexOf(s.units, t.units, from)
}

// SubSequence_int_int implements String.subSequence.
func (s *String) SubSequence_int_int(begin, end int32) any { return s.Substring_int_int(begin, end) }

// Concat_String implements String.concat.
func (s *String) Concat_String(t *String) *String {
	if len(t.units) == 0 {
		return s
	}
	return Concat(s, t)
}

// Replace_char_char implements String.replace(char, char).
func (s *String) Replace_char_char(oldChar, newChar uint16) *String {
	if oldChar == newChar {
		return s
	}
	i := s.IndexOf_int(int32(oldChar))
	if i < 0 {
		return s
	}
	u := append([]uint16(nil), s.units...)
	for j := i; j < int32(len(u)); j++ {
		if u[j] == oldChar {
			u[j] = newChar
		}
	}
	return FromUnits(u)
}

// Replace_CharSequence_CharSequence implements String.replace(CharSequence,
// CharSequence): replace each literal occurrence, scanning left to right. An
// empty target matches before every char and at the end.
func (s *String) Replace_CharSequence_CharSequence(target, replacement any) *String {
	t := charSequenceUnits(target)
	r := charSequenceUnits(replacement)
	if len(t) == 0 {
		out := make([]uint16, 0, len(s.units)+(len(s.units)+1)*len(r))
		out = append(out, r...)
		for _, c := range s.units {
			out = append(out, c)
			out = append(out, r...)
		}
		return FromUnits(out)
	}
	j := unitsIndexOf(s.units, t, 0)
	if j < 0 {
		return s
	}
	var out []uint16
	i := int32(0)
	for j >= 0 {
		out = append(out, s.units[i:j]...)
		out = append(out, r...)
		i = j + int32(len(t))
		j = unitsIndexOf(s.units, t, i)
	}
	out = append(out, s.units[i:]...)
	return FromUnits(out)
}

// Intern implements String.intern.
func (s *String) Intern() *String { return Lit(s.Go()) }

// Matches_String implements String.matches.
func (s *String) Matches_String(regex *String) bool {
	return compileRegex(regex).matches(s.units)
}

// ReplaceAll_String_String implements String.replaceAll.
func (s *String) ReplaceAll_String_String(regex, replacement *String) *String {
	return compileRegex(regex).replace(s, replacement, true)
}

// ReplaceFirst_String_String implements String.replaceFirst.
func (s *String) ReplaceFirst_String_String(regex, replacement *String) *String {
	return compileRegex(regex).replace(s, replacement, false)
}

// Split_String implements String.split(String).
func (s *String) Split_String(regex *String) []*String { return s.Split_String_int(regex, 0) }

// Split_String_int implements String.split(String, int).
func (s *String) Split_String_int(regex *String, limit int32) []*String {
	return compileRegex(regex).split(s, limit)
}

// String_new implements new String().
func String_new() *String { return FromUnits([]uint16{}) }

// String_new_String implements new String(String).
func String_new_String(s *String) *String {
	if s == nil {
		panic(NullPointerException_new())
	}
	return FromUnits(s.units)
}

// String_new_StringBuilder implements new String(StringBuilder).
func String_new_StringBuilder(sb *StringBuilder) *String { return sb.ToString() }

// String_new_StringBuffer implements new String(StringBuffer).
func String_new_StringBuffer(sb *StringBuffer) *String { return sb.ToString() }

// String_new_charArr_int_int implements new String(char[], int, int).
func String_new_charArr_int_int(cs []uint16, offset, count int32) *String {
	checkOffCountSIOOBE(int(offset), int(count), len(cs))
	return FromUnits(append([]uint16(nil), cs[offset:offset+count]...))
}

// String_new_intArr_int_int implements new String(int[], int, int).
func String_new_intArr_int_int(cps []int32, offset, count int32) *String {
	checkOffCountSIOOBE(int(offset), int(count), len(cps))
	var u []uint16
	for _, cp := range cps[offset : offset+count] {
		if !Character_isValidCodePoint(cp) {
			panic(IllegalArgumentException_new_String(StrOfInt(cp)))
		}
		u = appendCodePoint(u, cp)
	}
	return FromUnits(u)
}

// String_valueOf_boolean implements String.valueOf(boolean).
func String_valueOf_boolean(b bool) *String { return StrOfBool(b) }

// String_valueOf_long implements String.valueOf(long).
func String_valueOf_long(l int64) *String { return StrOfLong(l) }

// String_valueOf_float implements String.valueOf(float).
func String_valueOf_float(f float32) *String { return StrOfFloat(f) }

// String_valueOf_charArr implements String.valueOf(char[]).
func String_valueOf_charArr(cs []uint16) *String { return String_new_charArr(cs) }

// String_valueOf_charArr_int_int implements String.valueOf(char[], int, int).
func String_valueOf_charArr_int_int(cs []uint16, offset, count int32) *String {
	return String_new_charArr_int_int(cs, offset, count)
}

// String_copyValueOf_charArr implements String.copyValueOf(char[]).
func String_copyValueOf_charArr(cs []uint16) *String { return String_new_charArr(cs) }

// String_copyValueOf_charArr_int_int implements String.copyValueOf(char[],
// int, int).
func String_copyValueOf_charArr_int_int(cs []uint16, offset, count int32) *String {
	return String_new_charArr_int_int(cs, offset, count)
}

// String_join_CharSequence_CharSequenceArr implements String.join with an
// array of elements.
func String_join_CharSequence_CharSequenceArr(delimiter any, elements []any) *String {
	d := charSequenceUnits(delimiter)
	var u []uint16
	for i, e := range elements {
		if i > 0 {
			u = append(u, d...)
		}
		if e == nil {
			u = append(u, 'n', 'u', 'l', 'l')
		} else {
			u = append(u, charSequenceUnits(e)...)
		}
	}
	return FromUnits(u)
}

// String_join_CharSequence_Iterable implements String.join with an
// Iterable of elements.
func String_join_CharSequence_Iterable(delimiter any, elements Iterable) *String {
	var es []any
	for it := elements.Iterator(); it.HasNext(); {
		es = append(es, it.Next())
	}
	return String_join_CharSequence_CharSequenceArr(delimiter, es)
}

type caseInsensitiveComparator struct{}

func (caseInsensitiveComparator) Compare_Object_Object(a, b any) int32 {
	return AnyTo[*String](a).CompareToIgnoreCase_String(AnyTo[*String](b))
}

// String_CASE_INSENSITIVE_ORDER is String.CASE_INSENSITIVE_ORDER.
var String_CASE_INSENSITIVE_ORDER Comparator = caseInsensitiveComparator{}
