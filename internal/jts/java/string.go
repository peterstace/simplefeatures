package java

import (
	"strconv"
	"sync"
	"unicode/utf16"
)

// String is java.lang.String. A nil *String is Java's null.
//
// The value is held as UTF-16 code units, as Java holds it, so that indexing,
// length, hashing, and ordering match Java for all strings.
type String struct {
	units []uint16
}

var (
	litMu sync.Mutex
	// litTable interns string literals, as the JVM does, so that identical
	// literals are the same object. It is only written during package
	// initialization of generated code.
	litTable = map[string]*String{}
)

// Lit returns the interned String for a Java string literal.
func Lit(s string) *String {
	litMu.Lock()
	defer litMu.Unlock()
	if v, ok := litTable[s]; ok {
		return v
	}
	v := FromGo(s)
	litTable[s] = v
	return v
}

// FromGo creates a new String from a Go string.
func FromGo(s string) *String {
	return &String{units: utf16.Encode([]rune(s))}
}

// FromUnits creates a new String from UTF-16 code units, which it takes
// ownership of.
func FromUnits(u []uint16) *String {
	return &String{units: u}
}

// Go returns the Go string for s, or "null" if s is nil.
func (s *String) Go() string {
	if s == nil {
		return "null"
	}
	return string(utf16.Decode(s.units))
}

func (s *String) JClassName() string { return "java.lang.String" }

func (s *String) Length() int32 {
	return int32(len(s.units))
}

func (s *String) IsEmpty() bool {
	return len(s.units) == 0
}

func (s *String) CharAt_int(i int32) uint16 {
	if i < 0 || int(i) >= len(s.units) {
		panic(indexOutOfBounds(int(i), len(s.units)))
	}
	return s.units[i]
}

// HashCode is not cached, unlike Java's, because literals are shared between
// goroutines and caching would race.
func (s *String) HashCode() int32 {
	var h int32
	for _, c := range s.units {
		h = 31*h + int32(c)
	}
	return h
}

func (s *String) Equals_Object(o any) bool {
	t, ok := o.(*String)
	if !ok || t == nil {
		return false
	}
	return unitsEqual(s.units, t.units)
}

func unitsEqual(a, b []uint16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// EqualsIgnoreCase_String compares code points after folding each to upper
// and then lower case, as String.regionMatches(true, ...) does.
func (s *String) EqualsIgnoreCase_String(o *String) bool {
	if o == nil || len(s.units) != len(o.units) {
		return false
	}
	return compareToCI(s.units, 0, len(s.units), o.units, 0, len(o.units)) == 0
}

func (s *String) CompareTo_Object(o any) int32 {
	t := o.(*String)
	n := len(s.units)
	if len(t.units) < n {
		n = len(t.units)
	}
	for i := 0; i < n; i++ {
		if s.units[i] != t.units[i] {
			return int32(s.units[i]) - int32(t.units[i])
		}
	}
	return int32(len(s.units) - len(t.units))
}

func (s *String) ToString() *String {
	return s
}

func (s *String) IndexOf_int(ch int32) int32 {
	return s.IndexOf_int_int(ch, 0)
}

func (s *String) IndexOf_String(t *String) int32 {
	return s.IndexOf_String_int(t, 0)
}

func (s *String) IndexOf_String_int(t *String, from int32) int32 {
	return unitsIndexOf(s.units, t.units, from)
}

func (s *String) LastIndexOf_int(ch int32) int32 {
	return s.LastIndexOf_int_int(ch, int32(len(s.units)-1))
}

func (s *String) Contains_CharSequence(t any) bool {
	return unitsIndexOf(s.units, charSequenceUnits(t), 0) >= 0
}

func (s *String) StartsWith_String(p *String) bool {
	return len(p.units) <= len(s.units) && unitsEqual(s.units[:len(p.units)], p.units)
}

func (s *String) EndsWith_String(p *String) bool {
	n, m := len(s.units), len(p.units)
	return m <= n && unitsEqual(s.units[n-m:], p.units)
}

func (s *String) Substring_int(begin int32) *String {
	return s.Substring_int_int(begin, int32(len(s.units)))
}

func (s *String) Substring_int_int(begin, end int32) *String {
	checkRangeSIOOBE(int(begin), int(end), len(s.units))
	if begin == 0 && end == int32(len(s.units)) {
		return s
	}
	return FromUnits(append([]uint16(nil), s.units[begin:end]...))
}

func (s *String) Trim() *String {
	b, e := 0, len(s.units)
	for b < e && s.units[b] <= ' ' {
		b++
	}
	for e > b && s.units[e-1] <= ' ' {
		e--
	}
	return s.Substring_int_int(int32(b), int32(e))
}

// ToUpperCase converts with the default locale, as String.toUpperCase does.
func (s *String) ToUpperCase() *String {
	return s.ToUpperCase_Locale(Locale_getDefault())
}

// ToLowerCase converts with the default locale, as String.toLowerCase does.
func (s *String) ToLowerCase() *String {
	return s.ToLowerCase_Locale(Locale_getDefault())
}

func (s *String) ToCharArray() []uint16 {
	out := make([]uint16, len(s.units))
	copy(out, s.units)
	return out
}

func String_new_charArr(cs []uint16) *String {
	return FromUnits(append([]uint16(nil), cs...))
}

func String_valueOf_double(d float64) *String {
	return StrOfDouble(d)
}

func String_valueOf_int(i int32) *String {
	return StrOfInt(i)
}

func String_valueOf_Object(o any) *String {
	return StrOfObject(o)
}

func String_valueOf_char(c uint16) *String {
	return StrOfChar(c)
}

// Concat implements Java's string concatenation operator. The operands have
// already been converted with the StrOf functions.
func Concat(parts ...*String) *String {
	n := 0
	for _, p := range parts {
		if p == nil {
			n += 4
		} else {
			n += len(p.units)
		}
	}
	u := make([]uint16, 0, n)
	for _, p := range parts {
		if p == nil {
			u = append(u, 'n', 'u', 'l', 'l')
		} else {
			u = append(u, p.units...)
		}
	}
	return FromUnits(u)
}

// StrOfChar implements string conversion of a char.
func StrOfChar(c uint16) *String {
	return FromUnits([]uint16{c})
}

// StrOfInt implements string conversion of an int.
func StrOfInt(i int32) *String {
	return FromGo(strconv.FormatInt(int64(i), 10))
}

// StrOfLong implements string conversion of a long.
func StrOfLong(i int64) *String {
	return FromGo(strconv.FormatInt(i, 10))
}

// StrOfBool implements string conversion of a boolean.
func StrOfBool(b bool) *String {
	if b {
		return Lit("true")
	}
	return Lit("false")
}

// StrOfDouble implements string conversion of a double (Double.toString).
func StrOfDouble(d float64) *String {
	return FromGo(DoubleToString(d))
}

// StrOfFloat implements string conversion of a float (Float.toString).
func StrOfFloat(f float32) *String {
	return FromGo(FloatToString(f))
}

// StrOfObject implements string conversion of a reference (String.valueOf).
func StrOfObject(o any) *String {
	if o == nil {
		return Lit("null")
	}
	return Object_ToString(o)
}
