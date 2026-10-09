package java

import (
	"bufio"
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"
)

// The vectors in testdata/text_*.txt come from the JVM; see
// internal/jts/tools/jdkvectors/TextVectors.java for their format and how
// to regenerate them. TEXT_VECTORS_DIR overrides the directory, for
// checking a larger local set of vectors.

func readTextVectors(t *testing.T, name string) [][]string {
	t.Helper()
	dir := os.Getenv("TEXT_VECTORS_DIR")
	if dir == "" {
		dir = "testdata"
	}
	f, err := os.Open(dir + "/" + name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out [][]string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		out = append(out, strings.Split(sc.Text(), "\t"))
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// unesc reverses TextVectors.esc.
func unesc(s string) *String {
	var u []uint16
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' {
			u = append(u, uint16(c))
			continue
		}
		if s[i+1] == '\\' {
			u = append(u, '\\')
			i++
			continue
		}
		v, err := strconv.ParseUint(s[i+2:i+6], 16, 16)
		if err != nil {
			panic(err)
		}
		u = append(u, uint16(v))
		i += 5
	}
	return FromUnits(u)
}

// esc implements TextVectors.esc.
func esc(s *String) string {
	var sb strings.Builder
	for _, c := range s.units {
		switch {
		case c == '\\':
			sb.WriteString(`\\`)
		case c >= 0x20 && c < 0x7f:
			sb.WriteByte(byte(c))
		default:
			fmt.Fprintf(&sb, `\u%04x`, c)
		}
	}
	return sb.String()
}

func parseBits64(t *testing.T, s string) float64 {
	t.Helper()
	b, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		t.Fatal(err)
	}
	return math.Float64frombits(b)
}

func parseBits32(t *testing.T, s string) float32 {
	t.Helper()
	b, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		t.Fatal(err)
	}
	return math.Float32frombits(uint32(b))
}

func dbits(d float64) string { return fmt.Sprintf("%016x", math.Float64bits(d)) }
func fbits(f float32) string { return fmt.Sprintf("%08x", math.Float32bits(f)) }

// attempt runs f, rendering a thrown Java exception as TextVectors.exc does.
func attempt(f func() string) (out string) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if _, ok := r.(UnsupportedError); ok {
			out = fmt.Sprint("unsupported: ", r)
			return
		}
		ex := Throwable_ofPanic(r)
		msg := Throwable_fromAny(ex).GetMessage()
		m := "null"
		if msg != nil {
			m = esc(msg)
		}
		out = "!" + className(ex) + ":" + m
	}()
	return f()
}

// splitmix64 implements TextVectors.Rng.
type splitmix64 struct{ state uint64 }

func (r *splitmix64) next() uint64 {
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ z>>30) * 0xbf58476d1ce4e5b9
	z = (z ^ z>>27) * 0x94d049bb133111eb
	return z ^ z>>31
}

func (r *splitmix64) below(n int) int { return int(r.next() % uint64(n)) }

const (
	fnvOffset = 0xcbf29ce484222325
	fnvPrime  = 0x100000001b3
)

func fnvString(h uint64, s *String) uint64 {
	for _, b := range []byte(s.Go()) {
		h ^= uint64(b)
		h *= fnvPrime
	}
	h ^= '\n'
	h *= fnvPrime
	return h
}

// testPow10 holds the exact powers of ten that TextVectors.POW10 holds. It is
// a read-only lookup table.
var testPow10 = [...]float64{
	1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11, 1e12, 1e13, 1e14,
	1e15, 1e16, 1e17, 1e18, 1e19, 1e20, 1e21, 1e22,
}

func genDouble(r *splitmix64) float64 {
	switch r.below(8) {
	case 0, 1:
		return math.Float64frombits(r.next())
	case 2:
		k := int64(r.next() >> uint(63-r.below(40)))
		return float64(k) / testPow10[r.below(23)]
	case 3:
		k := int64(r.next() >> uint(63-r.below(30)))
		return float64(float64(k) * testPow10[r.below(23)])
	case 4:
		return math.Float64frombits(r.next() >> 12)
	case 5:
		e := int64(r.below(2046) + 1)
		bits := e<<52 + int64(r.below(5)-2)
		return math.Float64frombits(uint64(bits & 0x7fffffffffffffff))
	case 6:
		k := int64(r.next() >> uint(63-r.below(50)))
		v := float64(k) / testPow10[r.below(16)]
		if r.next()&1 == 0 {
			return v
		}
		return -v
	}
	k := int64(r.below(2000000) - 1000000)
	return float64(k) / float64(int32(1)<<r.below(12))
}

func genFloat(r *splitmix64) float32 {
	switch r.below(4) {
	case 0, 1:
		return math.Float32frombits(uint32(r.next()))
	case 2:
		return float32(genDouble(r))
	}
	k := int32(r.below(2000000) - 1000000)
	return float32(k) / float32(int32(1)<<r.below(20))
}

func genDecimalString(r *splitmix64) *String {
	var sb strings.Builder
	if r.below(4) == 0 {
		if r.below(2) == 0 {
			sb.WriteByte('-')
		} else {
			sb.WriteByte('+')
		}
	}
	mode := r.below(6)
	var nd int
	switch mode {
	case 0:
		nd = 1 + r.below(60)
	case 1:
		nd = 17 + r.below(30)
	default:
		nd = 1 + r.below(20)
	}
	point := r.below(nd+2) - 1
	for i := 0; i < nd; i++ {
		if i == point {
			sb.WriteByte('.')
		}
		sb.WriteByte(byte('0' + r.below(10)))
	}
	if mode == 2 {
		for i := 0; i < 20; i++ {
			if i == 0 {
				sb.WriteByte('5')
			} else {
				sb.WriteByte('0')
			}
		}
		sb.WriteByte(byte('0' + r.below(10)))
	}
	if r.below(3) > 0 {
		if r.below(2) == 0 {
			sb.WriteByte('e')
		} else {
			sb.WriteByte('E')
		}
		var e int
		if r.below(5) == 0 {
			e = r.below(700) - 350
		} else {
			e = r.below(80) - 40
		}
		sb.WriteString(strconv.Itoa(e))
	}
	if r.below(10) == 0 {
		sb.WriteByte("fFdD"[r.below(4)])
	}
	return FromGo(sb.String())
}

func genHexString(r *splitmix64) *String {
	const hexDigits = "0123456789abcdefABCDEF"
	var sb strings.Builder
	if r.below(3) == 0 {
		sb.WriteByte('-')
	}
	if r.below(2) == 0 {
		sb.WriteString("0x")
	} else {
		sb.WriteString("0X")
	}
	n := 16
	if r.below(4) == 0 {
		n = 40
	}
	nd := 1 + r.below(n)
	point := r.below(nd+2) - 1
	for i := 0; i < nd; i++ {
		if i == point {
			sb.WriteByte('.')
		}
		sb.WriteByte(hexDigits[r.below(len(hexDigits))])
	}
	if r.below(2) == 0 {
		sb.WriteByte('p')
	} else {
		sb.WriteByte('P')
	}
	var e int
	if r.below(3) == 0 {
		e = r.below(2400) - 1200
	} else {
		e = r.below(300) - 150
	}
	sb.WriteString(strconv.Itoa(e))
	return FromGo(sb.String())
}

func genFloatFormat(r *splitmix64) *String {
	prec := r.below(22) - 1
	conv := [...]string{"f", "e", "g", "a"}[r.below(4)]
	if prec >= 0 {
		return FromGo("%." + strconv.Itoa(prec) + conv)
	}
	return FromGo("%" + conv)
}

func genBigDecimal(r *splitmix64) *BigDecimal {
	var sb strings.Builder
	if r.below(2) == 0 {
		sb.WriteByte('-')
	}
	n := 20
	if r.below(4) == 0 {
		n = 60
	}
	nd := 1 + r.below(n)
	for i := 0; i < nd; i++ {
		sb.WriteByte(byte('0' + r.below(10)))
	}
	var scale int
	if r.below(5) == 0 {
		scale = r.below(800) - 400
	} else {
		scale = r.below(60) - 30
	}
	v, _ := new(big.Int).SetString(sb.String(), 10)
	return newBigDecimal(v, int32(scale))
}

func TestTextToString(t *testing.T) {
	fails := 0
	for _, v := range readTextVectors(t, "text_tostring.txt") {
		var got, want string
		switch v[0] {
		case "d":
			got, want = esc(StrOfDouble(parseBits64(t, v[1]))), v[2]
		case "f":
			got, want = esc(StrOfFloat(parseBits32(t, v[1]))), v[2]
		case "dd", "fd":
			batch, _ := strconv.Atoi(v[1])
			seed := uint64(0x1000000)
			if v[0] == "fd" {
				seed = 0x2000000
			}
			r := &splitmix64{state: seed + uint64(batch)}
			h := uint64(fnvOffset)
			for i := 0; i < 1000; i++ {
				if v[0] == "dd" {
					h = fnvString(h, StrOfDouble(genDouble(r)))
				} else {
					h = fnvString(h, StrOfFloat(genFloat(r)))
				}
			}
			got, want = strconv.FormatUint(h, 16), v[2]
		default:
			t.Fatalf("unknown record %q", v[0])
		}
		if got != want {
			fails++
			if fails <= 20 {
				t.Errorf("%v: got %q", v, got)
			}
		}
	}
}

func TestTextParse(t *testing.T) {
	fails := 0
	for _, v := range readTextVectors(t, "text_parse.txt") {
		var got, want string
		switch v[0] {
		case "pd":
			s := unesc(v[1])
			got, want = attempt(func() string { return dbits(Double_parseDouble(s)) }), v[2]
		case "pf":
			s := unesc(v[1])
			got, want = attempt(func() string { return fbits(Float_parseFloat(s)) }), v[2]
		case "pdd":
			batch, _ := strconv.Atoi(v[1])
			r := &splitmix64{state: 0x3000000 + uint64(batch)}
			h := uint64(fnvOffset)
			for i := 0; i < 1000; i++ {
				s := genDecimalString(r)
				h = fnvString(h, FromGo(attempt(func() string { return dbits(Double_parseDouble(s)) })))
				h = fnvString(h, FromGo(attempt(func() string { return fbits(Float_parseFloat(s)) })))
			}
			got, want = strconv.FormatUint(h, 16), v[2]
		case "phd":
			batch, _ := strconv.Atoi(v[1])
			r := &splitmix64{state: 0x6000000 + uint64(batch)}
			h := uint64(fnvOffset)
			for i := 0; i < 1000; i++ {
				s := genHexString(r)
				h = fnvString(h, FromGo(attempt(func() string { return dbits(Double_parseDouble(s)) })))
				h = fnvString(h, FromGo(attempt(func() string { return fbits(Float_parseFloat(s)) })))
			}
			got, want = strconv.FormatUint(h, 16), v[2]
		default:
			t.Fatalf("unknown record %q", v[0])
		}
		if got != want {
			fails++
			if fails <= 20 {
				t.Errorf("%v: got %q", v, got)
			}
		}
	}
}

// dfSetup implements TextVectors.dfSetup.
func dfSetup(t *testing.T, setup string) *NumberFormat {
	kind, arg, _ := strings.Cut(setup, ":")
	switch kind {
	case "ordinate":
		n, _ := strconv.Atoi(arg)
		f := DecimalFormat_fromAny(Ref(NumberFormat_getInstance_Locale(Locale_US)))
		f.ApplyPattern_String(Lit("0"))
		f.SetMaximumFractionDigits_int(int32(n))
		return &f.NumberFormat
	case "us", "us-obj":
		return NumberFormat_getInstance_Locale(Locale_US)
	case "root":
		return NumberFormat_getInstance_Locale(Locale_ROOT)
	case "de":
		return NumberFormat_getInstance_Locale(Locale_GERMANY)
	case "tr":
		return NumberFormat_getInstance_Locale(Locale_forLanguageTag(Lit("tr")))
	case "fr":
		return NumberFormat_getInstance_Locale(Locale_FRANCE)
	case "int":
		return NumberFormat_getIntegerInstance_Locale(Locale_US)
	case "pct":
		return NumberFormat_getPercentInstance_Locale(Locale_US)
	case "pat":
		return &DecimalFormat_new_String_DecimalFormatSymbols(FromGo(arg), DecimalFormatSymbols_getInstance_Locale(Locale_US)).NumberFormat
	case "mode":
		f := NumberFormat_getInstance_Locale(Locale_US)
		f.SetRoundingMode_RoundingMode(RoundingMode_valueOf_String(FromGo(arg)))
		f.SetMaximumFractionDigits_int(2)
		return f
	case "maxint", "minint", "minfrac":
		n, _ := strconv.Atoi(arg)
		f := NumberFormat_getInstance_Locale(Locale_US)
		switch kind {
		case "maxint":
			f.SetMaximumIntegerDigits_int(int32(n))
		case "minint":
			f.SetMinimumIntegerDigits_int(int32(n))
		default:
			f.SetMinimumFractionDigits_int(int32(n))
		}
		return f
	}
	t.Fatalf("unknown setup %q", setup)
	return nil
}

func dfFormat(setup string, f *NumberFormat, d float64) *String {
	if setup == "us-obj" {
		return f.Format_Object(d)
	}
	return f.Format_double(d)
}

func TestTextDecimalFormat(t *testing.T) {
	fails := 0
	formats := map[string]*NumberFormat{}
	format := func(setup string) *NumberFormat {
		if f, ok := formats[setup]; ok {
			return f
		}
		f := dfSetup(t, setup)
		formats[setup] = f
		return f
	}
	for _, v := range readTextVectors(t, "text_decimalformat.txt") {
		var got, want string
		switch v[0] {
		case "df":
			setup := unesc(v[1]).Go()
			d := parseBits64(t, v[2])
			got = attempt(func() string { return esc(dfFormat(setup, format(setup), d)) })
			want = v[3]
		case "dfl":
			setup := unesc(v[1]).Go()
			l, _ := strconv.ParseInt(v[2], 10, 64)
			got = attempt(func() string { return esc(format(setup).Format_long(l)) })
			want = v[3]
		case "dfbad":
			p := unesc(v[1])
			got = attempt(func() string {
				DecimalFormat_new_String_DecimalFormatSymbols(p, DecimalFormatSymbols_getInstance_Locale(Locale_US))
				return "ok"
			})
			want = v[2]
		case "dfd":
			setup := v[1]
			batch, _ := strconv.Atoi(v[2])
			r := &splitmix64{state: 0x4000000 + uint64(batch)}
			f := dfSetup(t, setup)
			h := uint64(fnvOffset)
			for i := 0; i < 1000; i++ {
				h = fnvString(h, dfFormat(setup, f, genDouble(r)))
			}
			got, want = strconv.FormatUint(h, 16), v[3]
		default:
			t.Fatalf("unknown record %q", v[0])
		}
		if got != want {
			fails++
			if fails <= 30 {
				t.Errorf("%q: got %q", v, got)
			}
		}
	}
}

// decodeArg reverses TextVectors.encodeArg.
func decodeArg(t *testing.T, s string) any {
	kind, val, _ := strings.Cut(s, ":")
	switch kind {
	case "n":
		return nil
	case "i":
		v, _ := strconv.ParseInt(val, 10, 32)
		return int32(v)
	case "l":
		v, _ := strconv.ParseInt(val, 10, 64)
		return v
	case "sh":
		v, _ := strconv.ParseInt(val, 10, 16)
		return int16(v)
	case "by":
		v, _ := strconv.ParseInt(val, 10, 8)
		return int8(v)
	case "d":
		return parseBits64(t, val)
	case "f":
		return parseBits32(t, val)
	case "s":
		return unesc(val)
	case "c":
		v, _ := strconv.ParseUint(val, 16, 16)
		return uint16(v)
	case "b":
		return val == "true"
	case "bd":
		return BigDecimal_new_String(FromGo(val))
	}
	t.Fatalf("unknown arg %q", s)
	return nil
}

func TestTextFormat(t *testing.T) {
	fails := 0
	for _, v := range readTextVectors(t, "text_format.txt") {
		if v[0] == "fmtd" {
			batch, _ := strconv.Atoi(v[1])
			r := &splitmix64{state: 0x5000000 + uint64(batch)}
			h := uint64(fnvOffset)
			for i := 0; i < 1000; i++ {
				f := genFloatFormat(r)
				d := genDouble(r)
				h = fnvString(h, Format(Locale_ROOT, f, []any{d}))
				x := genFloat(r)
				h = fnvString(h, Format(Locale_ROOT, f, []any{x}))
			}
			if got := strconv.FormatUint(h, 16); got != v[2] {
				t.Errorf("%q: got %q", v, got)
			}
			continue
		}
		if v[0] != "fmt" {
			t.Fatalf("unknown record %q", v[0])
		}
		var l *Locale
		if v[1] != "-" {
			l = Locale_forLanguageTag(FromGo(v[1]))
		}
		format := unesc(v[2])
		var args []any
		for _, a := range v[4:] {
			args = append(args, decodeArg(t, a))
		}
		got := attempt(func() string {
			if l == nil {
				return esc(String_format_String_ObjectArr(format, args))
			}
			return esc(String_format_Locale_String_ObjectArr(l, format, args))
		})
		if got != v[3] {
			fails++
			if fails <= 40 {
				t.Errorf("%q: got %q", v, got)
			}
		}
	}
}

func boolDigit(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// charProps implements TextVectors.charProps.
func charProps(cp int32) string {
	h := func(v int32) string { return strconv.FormatUint(uint64(uint32(v)), 16) }
	s := fmt.Sprintf("%d %s %s %s %d %d ", Character_getType_int(cp), h(Character_toUpperCase_int(cp)),
		h(Character_toLowerCase_int(cp)), h(Character_toTitleCase_int(cp)), Character_digit_int_int(cp, 36),
		Character_digit_int_int(cp, 10))
	s += boolDigit(Character_isLetter_int(cp)) + boolDigit(Character_isDigit_int(cp)) +
		boolDigit(Character_isLetterOrDigit_int(cp)) + boolDigit(Character_isWhitespace_int(cp)) +
		boolDigit(Character_isSpaceChar_int(cp)) + boolDigit(Character_isUpperCase_int(cp)) +
		boolDigit(Character_isLowerCase_int(cp)) + boolDigit(Character_isTitleCase_int(cp)) +
		boolDigit(Character_isAlphabetic(cp)) + boolDigit(Character_isIdeographic(cp)) +
		boolDigit(Character_isDefined_int(cp)) + boolDigit(Character_isISOControl_int(cp)) +
		boolDigit(Character_isJavaIdentifierStart_int(cp)) + boolDigit(Character_isJavaIdentifierPart_int(cp)) +
		boolDigit(Character_isIdentifierIgnorable_int(cp))
	if cp <= 0xffff {
		s += " " + h(int32(Character_toUpperCase_char(uint16(cp)))) + " " + h(int32(Character_toLowerCase_char(uint16(cp))))
	}
	return s
}

func renderSplit(a []*String) string {
	parts := []string{strconv.Itoa(len(a))}
	for _, s := range a {
		parts = append(parts, esc(s))
	}
	return strings.Join(parts, "\t")
}

func splitUnits(s *String, sep uint16) []*String {
	var out []*String
	start := 0
	for i, c := range s.units {
		if c == sep {
			out = append(out, FromUnits(s.units[start:i]))
			start = i + 1
		}
	}
	return append(out, FromUnits(s.units[start:]))
}

// runBuilderScript implements TextVectors.runScript.
func runBuilderScript(script *String) *String {
	sb := StringBuilder_new()
	var out []uint16
	atoi := func(s *String) int32 {
		v, _ := strconv.ParseInt(s.Go(), 10, 32)
		return int32(v)
	}
	for _, op := range splitUnits(script, ';') {
		a := splitUnits(op, ':')
		prefixes := map[string]string{
			"charAt": " charAt=", "codePointAt": " cp=", "indexOf": " indexOf=",
			"lastIndexOf": " lastIndexOf=", "indexOffrom": " indexOf=", "lastIndexOffrom": " lastIndexOf=",
			"substring": " sub=", "length": " len=", "capacity": " cap=",
		}
		prefix := prefixes[a[0].Go()]
		out = append(out, FromGo(prefix).units...)
		res := attempt(func() string {
			switch a[0].Go() {
			case "new":
				sb = StringBuilder_new_String(a[1])
			case "newcap":
				sb = StringBuilder_new_int(atoi(a[1]))
			case "append":
				sb.Append_String(a[1])
			case "appendnull":
				sb.Append_String(nil)
			case "appendnullsb":
				sb.Append_StringBuffer(nil)
			case "appendnullcs":
				sb.Append_CharSequence(nil)
			case "appendd":
				sb.Append_double(Double_parseDouble(a[1]))
			case "appendf":
				sb.Append_float(Float_parseFloat(a[1]))
			case "appendi":
				sb.Append_int(atoi(a[1]))
			case "appendl":
				v, _ := strconv.ParseInt(a[1].Go(), 10, 64)
				sb.Append_long(v)
			case "appendb":
				sb.Append_boolean(true)
			case "appendcp":
				sb.AppendCodePoint_int(atoi(a[1]))
			case "appendcsr":
				sb.Append_CharSequence_int_int(a[1], atoi(a[2]), atoi(a[3]))
			case "insertcsr":
				sb.Insert_int_CharSequence_int_int(atoi(a[1]), a[2], atoi(a[3]), atoi(a[4]))
			case "insert":
				sb.Insert_int_String(atoi(a[1]), a[2])
			case "reverse":
				sb.Reverse()
			case "deleteCharAt":
				sb.DeleteCharAt_int(atoi(a[1]))
			case "setCharAt":
				sb.SetCharAt_int_char(atoi(a[1]), a[2].units[0])
			case "delete":
				sb.Delete_int_int(atoi(a[1]), atoi(a[2]))
			case "replace":
				sb.Replace_int_int_String(atoi(a[1]), atoi(a[2]), a[3])
			case "setLength":
				sb.SetLength_int(atoi(a[1]))
			case "trimToSize":
				sb.TrimToSize()
			case "ensureCapacity":
				sb.EnsureCapacity_int(atoi(a[1]))
			case "length":
				return fmt.Sprint(sb.Length())
			case "capacity":
				return fmt.Sprint(sb.Capacity())
			case "charAt":
				return fmt.Sprint(sb.CharAt_int(atoi(a[1])))
			case "codePointAt":
				return fmt.Sprint(sb.CodePointAt_int(atoi(a[1])))
			case "indexOf":
				return fmt.Sprint(sb.IndexOf_String(a[1]))
			case "lastIndexOf":
				return fmt.Sprint(sb.LastIndexOf_String(a[1]))
			case "indexOffrom":
				return fmt.Sprint(sb.IndexOf_String_int(a[1], atoi(a[2])))
			case "lastIndexOffrom":
				return fmt.Sprint(sb.LastIndexOf_String_int(a[1], atoi(a[2])))
			case "substring":
				if len(a) == 2 {
					return esc(sb.Substring_int(atoi(a[1])))
				}
				return esc(sb.Substring_int_int(atoi(a[1]), atoi(a[2])))
			default:
				panic("unknown op " + op.Go())
			}
			return ""
		})
		if strings.HasPrefix(res, "!") {
			out = append(out, FromGo(" "+res).units...)
		} else {
			out = append(out, unesc(res).units...)
		}
		out = append(out, ' ', '[')
		out = append(out, sb.ToString().units...)
		out = append(out, ']')
	}
	return FromUnits(out)
}

func TestTextMisc(t *testing.T) {
	fails := 0
	check := func(v []string, got, want string) {
		if got != want {
			fails++
			if fails <= 60 {
				t.Errorf("%q: got %q want %q", v, got, want)
			}
		}
	}
	str := func(v any) string { return fmt.Sprint(v) }
	for _, v := range readTextVectors(t, "text_misc.txt") {
		switch v[0] {
		case "pi", "pl", "pui", "pul", "ps", "pb":
			var s *String
			if v[1] != "null" {
				s = unesc(v[1])
			}
			radix := int32(0)
			if r, err := strconv.Atoi(v[2]); err == nil {
				radix = int32(r)
			}
			got := attempt(func() string {
				switch v[0] {
				case "pi":
					return str(Integer_parseInt_String_int(s, radix))
				case "pl":
					return str(Long_parseLong_String_int(s, radix))
				case "pui":
					return str(Integer_parseUnsignedInt_String_int(s, radix))
				case "pul":
					return str(Long_parseUnsignedLong_String_int(s, radix))
				case "ps":
					return str(Short_parseShort_String_int(s, radix))
				}
				return str(Byte_parseByte_String_int(s, radix))
			})
			check(v, got, v[3])
		case "di", "dl", "ds", "db":
			s := unesc(v[1])
			got := attempt(func() string {
				switch v[0] {
				case "di":
					return str(Integer_decode(s))
				case "dl":
					return str(Long_decode(s))
				case "ds":
					return str(Short_decode(s))
				}
				return str(Byte_decode(s))
			})
			check(v, got, v[2])
		case "tsi":
			i, _ := strconv.ParseInt(v[1], 10, 32)
			radix, _ := strconv.Atoi(v[2])
			check(v, esc(Integer_toString_int_int(int32(i), int32(radix)))+"\t"+
				esc(Integer_toUnsignedString_int_int(int32(i), int32(radix))), v[3]+"\t"+v[4])
		case "tsl":
			l, _ := strconv.ParseInt(v[1], 10, 64)
			radix, _ := strconv.Atoi(v[2])
			check(v, esc(Long_toString_long_int(l, int32(radix)))+"\t"+
				esc(Long_toUnsignedString_long_int(l, int32(radix))), v[3]+"\t"+v[4])
		case "bits":
			l, _ := strconv.ParseInt(v[1], 10, 64)
			i := int32(l)
			got := []string{
				v[1], Integer_toBinaryString(i).Go(), Integer_toHexString(i).Go(),
				Integer_toOctalString(i).Go(), Long_toBinaryString(l).Go(), Long_toHexString(l).Go(),
				Long_toOctalString(l).Go(), str(Integer_bitCount(i)), str(Integer_reverse(i)),
				str(Integer_reverseBytes(i)), str(Integer_highestOneBit(i)), str(Integer_lowestOneBit(i)),
				str(Integer_numberOfLeadingZeros(i)), str(Integer_numberOfTrailingZeros(i)), str(Integer_signum(i)),
				str(Integer_rotateLeft(i, int32(l))), str(Integer_rotateRight(i, int32(l))), str(Long_reverse(l)),
				str(Long_reverseBytes(l)), str(Long_highestOneBit(l)), str(Long_signum(l)),
				str(Long_rotateLeft(l, i)), str(Long_hashCode(l)), str(Integer_compareUnsigned(i, int32(l>>7))),
				str(Long_compareUnsigned(l, l>>7)),
			}
			check(v, strings.Join(got, "\t"), strings.Join(v[1:], "\t"))
		case "chr":
			block, _ := strconv.Atoi(v[1])
			h := uint64(fnvOffset)
			for cp := int32(block) << 12; cp < int32(block+1)<<12; cp++ {
				h = fnvString(h, FromGo(charProps(cp)))
			}
			check(v, strconv.FormatUint(h, 16), v[2])
		case "chx":
			cp, _ := strconv.ParseInt(v[1], 16, 32)
			check(v, charProps(int32(cp)), v[2])
		case "case":
			s := unesc(v[1])
			l := Locale_forLanguageTag(FromGo(v[2]))
			check(v, esc(s.ToUpperCase_Locale(l))+"\t"+esc(s.ToLowerCase_Locale(l)), v[3]+"\t"+v[4])
		case "cmpci":
			s, o := unesc(v[1]), unesc(v[2])
			c := s.CompareToIgnoreCase_String(o)
			sign := 0
			if c < 0 {
				sign = -1
			} else if c > 0 {
				sign = 1
			}
			n := s.Length()
			if o.Length() < n {
				n = o.Length()
			}
			check(v, fmt.Sprintf("%d\t%s\t%s", sign, boolDigit(s.EqualsIgnoreCase_String(o)),
				boolDigit(s.RegionMatches_boolean_int_String_int_int(true, 0, o, 0, n))), v[3]+"\t"+v[4]+"\t"+v[5])
		case "rxall", "rxfirst":
			re, s, rep := unesc(v[1]), unesc(v[2]), unesc(v[3])
			got := attempt(func() string {
				if v[0] == "rxall" {
					return esc(s.ReplaceAll_String_String(re, rep))
				}
				return esc(s.ReplaceFirst_String_String(re, rep))
			})
			check(v, got, v[4])
		case "rxmatch":
			re, s := unesc(v[1]), unesc(v[2])
			check(v, attempt(func() string { return str(s.Matches_String(re)) }), v[3])
		case "rxsplit":
			re, s := unesc(v[1]), unesc(v[2])
			limit, _ := strconv.Atoi(v[3])
			check(v, attempt(func() string { return renderSplit(s.Split_String_int(re, int32(limit))) }),
				strings.Join(v[4:], "\t"))
		case "replace":
			s, a, b := unesc(v[1]), unesc(v[2]), unesc(v[3])
			check(v, esc(s.Replace_CharSequence_CharSequence(a, b)), v[4])
		case "sb":
			check(v, esc(runBuilderScript(unesc(v[1]))), v[2])
		case "bd":
			d := parseBits64(t, v[1])
			x := BigDecimal_new_double(d)
			bv := BigDecimal_valueOf_double(d)
			got := []string{
				v[1], x.ToString().Go(), x.ToPlainString().Go(), x.ToEngineeringString().Go(),
				bv.ToString().Go(), str(x.Scale()), str(x.Precision()), dbits(x.DoubleValue()), fbits(x.FloatValue()),
				str(x.LongValue()), str(x.IntValue()), str(x.HashCode()), bv.StripTrailingZeros().ToString().Go(),
				attempt(func() string { return x.SetScale_int_RoundingMode(20, RoundingMode_HALF_UP).ToString().Go() }),
				attempt(func() string { return x.SetScale_int_int(2, rmHalfEven).ToString().Go() }),
				attempt(func() string { return bv.SetScale_int(0).ToString().Go() }),
				x.Round_MathContext(MathContext_new_int(5)).ToString().Go(),
				bv.MovePointLeft_int(3).ToString().Go(), bv.MovePointRight_int(3).ToString().Go(),
			}
			check(v, strings.Join(got, "\t"), strings.Join(v[1:], "\t"))
		case "bdperf":
			a := BigDecimal_new_double(9.0).SetScale_int(20)
			factor := BigDecimal_new_double(10.0).SetScale_int(20)
			aMul := factor.Multiply_BigDecimal(a)
			aDiv := a.Divide_BigDecimal_int(factor, rmHalfUp)
			det := a.Multiply_BigDecimal(a).Subtract_BigDecimal(aMul.Multiply_BigDecimal(aDiv))
			got := []string{
				"bdperf", a.ToString().Go(), factor.ToString().Go(), aMul.ToString().Go(),
				aDiv.ToString().Go(), det.ToString().Go(),
			}
			check(v, strings.Join(got, "\t"), strings.Join(v, "\t"))
		case "bds":
			s := unesc(v[1])
			got := attempt(func() string {
				x := BigDecimal_new_String(s)
				return x.ToString().Go() + "\t" + str(x.Scale()) + "\t" + x.ToPlainString().Go()
			})
			check(v, got, strings.Join(v[2:], "\t"))
		case "bdop":
			x, y := BigDecimal_new_String(unesc(v[1])), BigDecimal_new_String(FromGo(v[2]))
			got := []string{
				x.Add_BigDecimal(y).ToString().Go(), x.Subtract_BigDecimal(y).ToString().Go(),
				x.Multiply_BigDecimal(y).ToString().Go(),
				attempt(func() string { return x.Divide_BigDecimal_int(y, rmHalfUp).ToString().Go() }),
				attempt(func() string { return x.Divide_BigDecimal_int_int(y, 5, rmHalfEven).ToString().Go() }),
				attempt(func() string {
					return x.Divide_BigDecimal_int_RoundingMode(y, 3, RoundingMode_DOWN).ToString().Go()
				}),
				attempt(func() string { return x.Divide_BigDecimal(y).ToString().Go() }),
				str(x.CompareTo_Object(y)), boolDigit(x.Equals_Object(y)), x.Max_BigDecimal(y).ToString().Go(),
				x.Min_BigDecimal(y).ToString().Go(),
			}
			check(v, strings.Join(got, "\t"), strings.Join(v[3:], "\t"))
		case "bdd":
			batch, _ := strconv.Atoi(v[1])
			r := &splitmix64{state: 0x7000000 + uint64(batch)}
			h := uint64(fnvOffset)
			for i := 0; i < 1000; i++ {
				x := genBigDecimal(r)
				h = fnvString(h, FromGo(dbits(x.DoubleValue())))
				h = fnvString(h, FromGo(fbits(x.FloatValue())))
				h = fnvString(h, x.ToString())
			}
			check(v, strconv.FormatUint(h, 16), v[2])
		case "bdround":
			s := FromGo(v[1])
			mode, _ := strconv.Atoi(v[2])
			got := []string{
				attempt(func() string { return BigDecimal_new_String(s).SetScale_int_int(0, int32(mode)).ToString().Go() }),
				attempt(func() string { return BigDecimal_new_String(s).SetScale_int_int(1, int32(mode)).ToString().Go() }),
			}
			check(v, strings.Join(got, "\t"), strings.Join(v[3:], "\t"))
		case "loc":
			l := Locale_forLanguageTag(unesc(v[1]))
			got := []string{
				esc(l.GetLanguage()), esc(l.GetScript()), esc(l.GetCountry()), esc(l.GetVariant()),
				esc(l.ToString()), esc(l.ToLanguageTag()), str(l.HashCode()),
			}
			check(v, strings.Join(got, "\t"), strings.Join(v[2:], "\t"))
		case "locnew":
			l := Locale_new_String_String_String(unesc(v[1]), unesc(v[2]), unesc(v[3]))
			got := []string{
				esc(l.GetLanguage()), esc(l.GetCountry()), esc(l.GetVariant()), esc(l.ToString()),
				esc(l.ToLanguageTag()), str(l.HashCode()),
			}
			check(v, strings.Join(got, "\t"), strings.Join(v[4:], "\t"))
		case "scp":
			s := unesc(v[1])
			i64, _ := strconv.ParseInt(v[2], 10, 32)
			i := int32(i64)
			got := []string{
				attempt(func() string { return str(s.CodePointAt_int(i)) }),
				attempt(func() string { return str(s.CodePointBefore_int(i)) }),
				attempt(func() string { return str(s.OffsetByCodePoints_int_int(0, i)) }),
				attempt(func() string { return str(s.OffsetByCodePoints_int_int(i, -1)) }),
				attempt(func() string { return str(s.CodePointCount_int_int(0, i)) }),
				attempt(func() string { return str(s.CharAt_int(i)) }),
				attempt(func() string { return esc(s.Substring_int(i)) }),
				attempt(func() string { return esc(s.Substring_int_int(1, i)) }),
				attempt(func() string { return str(s.StartsWith_String_int(Lit("b"), i)) }),
				attempt(func() string { return str(s.LastIndexOf_String_int(Lit("bc"), i)) }),
				attempt(func() string { return str(s.IndexOf_String_int(Lit("bc"), i)) }),
			}
			check(v, strings.Join(got, "\t"), strings.Join(v[3:], "\t"))
		case "sidx":
			s := unesc(v[1])
			i, _ := strconv.ParseInt(v[2], 10, 32)
			c, _ := strconv.ParseInt(v[3], 10, 32)
			got := []string{
				str(s.IndexOf_int_int(int32(c), int32(i))), str(s.LastIndexOf_int_int(int32(c), int32(i))),
				str(s.IndexOf_int(int32(c))), str(s.LastIndexOf_int(int32(c))),
			}
			check(v, strings.Join(got, "\t"), strings.Join(v[4:], "\t"))
		case "s2":
			s, o := unesc(v[1]), unesc(v[2])
			got := []string{
				esc(s.Concat_String(o)), boolDigit(s.ContentEquals_CharSequence(o)),
				boolDigit(s.Contains_CharSequence(o)), boolDigit(s.ContentEquals_CharSequence(StringBuilder_new_String(o))),
				boolDigit(s.Contains_CharSequence(StringBuilder_new_String(o))), str(s.CompareTo_Object(o)),
				boolDigit(s.EndsWith_String(o)), boolDigit(s.StartsWith_String(o)), str(s.LastIndexOf_String(o)),
				str(s.IndexOf_String(o)), boolDigit(s.RegionMatches_int_String_int_int(1, o, 0, 2)),
				boolDigit(s.RegionMatches_boolean_int_String_int_int(true, 0, o, 1, 1)),
				esc(String_join_CharSequence_CharSequenceArr(Lit("|"), []any{s, o, s})),
				esc(s.Replace_char_char('a', 'Z')), esc(s.Trim()),
			}
			check(v, strings.Join(got, "\t"), strings.Join(v[3:], "\t"))
		case "sjoin":
			l := ArrayList_new()
			l.Add_Object(Lit("a"))
			l.Add_Object(nil)
			l.Add_Object(Lit("c"))
			check(v, esc(String_join_CharSequence_Iterable(Lit(", "), l)), v[1])
		case "svo":
			a, _ := strconv.ParseInt(v[1], 10, 32)
			b, _ := strconv.ParseInt(v[2], 10, 32)
			got := []string{
				attempt(func() string { return esc(String_valueOf_charArr_int_int([]uint16{'a', 'b', 'c'}, int32(a), int32(b))) }),
				attempt(func() string {
					return esc(String_new_intArr_int_int([]int32{0x61, 0x10400, 0x62}, int32(a), int32(b)))
				}),
			}
			check(v, strings.Join(got, "\t"), strings.Join(v[3:], "\t"))
		case "hex":
			check(v, Double_toHexString(parseBits64(t, v[1])).Go(), v[2])
		case "hexf":
			check(v, Float_toHexString(parseBits32(t, v[1])).Go(), v[2])
		default:
			t.Fatalf("unknown record %q", v[0])
		}
	}
}
