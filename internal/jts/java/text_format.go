package java

import (
	"math"
	"math/big"
	"math/bits"
	"reflect"
	"strconv"
)

// This file implements java.util.Formatter's format string handling.
// Date/time conversions (%t) accept only null arguments.

// String_format_String_ObjectArr implements String.format(String,
// Object...), formatting with the default locale.
func String_format_String_ObjectArr(format *String, args any) *String {
	return Format(Locale_getDefault(), format, varargs(args))
}

// String_format_Locale_String_ObjectArr implements String.format(Locale,
// String, Object...). A null locale means no localization.
func String_format_Locale_String_ObjectArr(l *Locale, format *String, args any) *String {
	return Format(l, format, varargs(args))
}

// varargs converts a Java Object[] (a Go slice of any element type, or nil
// for null) to a []any, keeping nil for a null array.
func varargs(args any) []any {
	switch a := args.(type) {
	case nil:
		return nil
	case []any:
		if a == nil {
			return nil
		}
		return a
	}
	v := reflect.ValueOf(args)
	if v.Kind() != reflect.Slice {
		panic(ClassCastError(args, "[Ljava.lang.Object;"))
	}
	if v.IsNil() {
		return nil
	}
	out := make([]any, v.Len())
	for i := range out {
		e := v.Index(i)
		switch e.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
			if e.IsNil() {
				continue
			}
		}
		out[i] = e.Interface()
	}
	return out
}

// Format implements java.util.Formatter.format(Locale, String, Object...)
// into a new String. A nil locale means no localization, as for a Formatter
// with a null locale. A nil args (a null Object[]) gives every conversion a
// null argument. It throws the JDK's IllegalFormatException subclasses for
// malformed format strings.
func Format(locale *Locale, format *String, args []any) *String {
	if format == nil {
		panic(NullPointerException_new())
	}
	f := &formatter{locale: locale}
	last, lasto := -1, -1
	for _, part := range parseFormat(format.units) {
		if part.spec == nil {
			f.out = append(f.out, part.text...)
			continue
		}
		fs := part.spec
		var arg any
		switch index := fs.index; {
		case index == -2:
		case index == -1:
			if last < 0 || args != nil && last > len(args)-1 {
				panic(newMissingFormatArgumentException(fs.String()))
			}
			if args != nil {
				arg = args[last]
			}
		case index == 0:
			lasto++
			last = lasto
			if args != nil && lasto > len(args)-1 {
				panic(newMissingFormatArgumentException(fs.String()))
			}
			if args != nil {
				arg = args[lasto]
			}
		default:
			last = index - 1
			if args != nil && last > len(args)-1 {
				panic(newMissingFormatArgumentException(fs.String()))
			}
			if args != nil {
				arg = args[last]
			}
		}
		f.print(fs, arg)
	}
	return FromUnits(f.out)
}

type formatter struct {
	locale *Locale
	out    []uint16
}

// Formatter flags.
const (
	ffLeftJustify  = 1 << 0
	ffUppercase    = 1 << 1
	ffAlternate    = 1 << 2
	ffPlus         = 1 << 3
	ffLeadingSpace = 1 << 4
	ffZeroPad      = 1 << 5
	ffGroup        = 1 << 6
	ffParentheses  = 1 << 7
	ffPrevious     = 1 << 8
)

func flagsString(f int) string {
	s := ""
	for _, x := range []struct {
		f int
		c string
	}{
		{ffLeftJustify, "-"},
		{ffUppercase, "^"},
		{ffAlternate, "#"},
		{ffPlus, "+"},
		{ffLeadingSpace, " "},
		{ffZeroPad, "0"},
		{ffGroup, ","},
		{ffParentheses, "("},
		{ffPrevious, "<"},
	} {
		if f&x.f != 0 {
			s += x.c
		}
	}
	return s
}

func hasFlags(flags, f int) bool { return flags&f == f }

type formatPart struct {
	text []uint16
	spec *formatSpec
}

type formatSpec struct {
	index     int
	flags     int
	width     int
	precision int
	dt        bool
	c         uint16
}

// String implements FormatSpecifier.toString.
func (fs *formatSpec) String() string {
	s := "%" + flagsString(fs.flags&^ffUppercase)
	if fs.index > 0 {
		s += strconv.Itoa(fs.index) + "$"
	}
	if fs.width != -1 {
		s += strconv.Itoa(fs.width)
	}
	if fs.precision != -1 {
		s += "." + strconv.Itoa(fs.precision)
	}
	if fs.dt {
		if hasFlags(fs.flags, ffUppercase) {
			s += "T"
		} else {
			s += "t"
		}
	}
	c := fs.c
	if hasFlags(fs.flags, ffUppercase) {
		c = Character_toUpperCase_char(c)
	}
	return s + string(rune(c))
}

func isConversion(c uint16) bool {
	switch c {
	case 'b', 'B', 's', 'S', 'h', 'H', 'c', 'C', 'd', 'o', 'x', 'X', 'e', 'E', 'g', 'G', 'f', 'a', 'A', 'n', '%':
		return true
	}
	return false
}

// parseFormat implements Formatter.parse. Format specifiers have the form
// %[index$][flags][width][.precision][t]conversion.
func parseFormat(s []uint16) []formatPart {
	var parts []formatPart
	isDigit := func(c uint16) bool { return c >= '0' && c <= '9' }
	isFlag := func(c uint16) bool {
		switch c {
		case '-', '#', '+', ' ', '0', ',', '(', '<':
			return true
		}
		return false
	}
	isLetter := func(c uint16) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
	i := 0
	for i < len(s) {
		n := unitsIndexOf(s, []uint16{'%'}, int32(i))
		if n < 0 {
			parts = append(parts, formatPart{text: s[i:]})
			break
		}
		if i != int(n) {
			parts = append(parts, formatPart{text: s[i:n]})
		}
		i = int(n) + 1
		if i >= len(s) {
			panic(newUnknownFormatConversionException("%"))
		}
		c := s[i]
		if isConversion(c) {
			parts = append(parts, formatPart{spec: newSimpleSpec(c)})
			i++
			continue
		}
		// Match the full specifier pattern at i.
		j := i
		indexStart, indexEnd := -1, -1
		k := j
		for k < len(s) && isDigit(s[k]) {
			k++
		}
		if k > j && k < len(s) && s[k] == '$' {
			indexStart, indexEnd = j, k
			j = k + 1
		}
		flagsStart := j
		for j < len(s) && isFlag(s[j]) {
			j++
		}
		flagsEnd := j
		widthStart := j
		for j < len(s) && isDigit(s[j]) {
			j++
		}
		widthEnd := j
		precStart, precEnd := -1, -1
		if j+1 < len(s) && s[j] == '.' && isDigit(s[j+1]) {
			precStart = j + 1
			j += 2
			for j < len(s) && isDigit(s[j]) {
				j++
			}
			precEnd = j
		}
		dt := false
		var dtUpper bool
		if j+1 < len(s) && (s[j] == 't' || s[j] == 'T') && (isLetter(s[j+1]) || s[j+1] == '%') {
			dt = true
			dtUpper = s[j] == 'T'
			j++
		}
		if j >= len(s) || !(isLetter(s[j]) || s[j] == '%') {
			panic(newUnknownFormatConversionException(string(rune(c))))
		}
		fs := &formatSpec{width: -1, precision: -1}
		if indexStart >= 0 {
			v, err := strconv.ParseInt(string(utf16ToASCII(s[indexStart:indexEnd])), 10, 32)
			if err != nil {
				panic(newIllegalFormatArgumentIndexException(math.MinInt32))
			}
			if v <= 0 {
				panic(newIllegalFormatArgumentIndexException(int32(v)))
			}
			fs.index = int(v)
		}
		for _, fc := range s[flagsStart:flagsEnd] {
			v := 0
			switch fc {
			case '-':
				v = ffLeftJustify
			case '#':
				v = ffAlternate
			case '+':
				v = ffPlus
			case ' ':
				v = ffLeadingSpace
			case '0':
				v = ffZeroPad
			case ',':
				v = ffGroup
			case '(':
				v = ffParentheses
			case '<':
				v = ffPrevious
			}
			if hasFlags(fs.flags, v) {
				panic(newDuplicateFormatFlagsException(flagsString(v)))
			}
			fs.flags |= v
		}
		if hasFlags(fs.flags, ffPrevious) {
			fs.index = -1
		}
		if widthEnd > widthStart {
			v, err := strconv.ParseInt(string(utf16ToASCII(s[widthStart:widthEnd])), 10, 32)
			if err != nil {
				panic(newIllegalFormatWidthException(math.MinInt32))
			}
			fs.width = int(v)
		}
		if precStart >= 0 {
			v, err := strconv.ParseInt(string(utf16ToASCII(s[precStart:precEnd])), 10, 32)
			if err != nil {
				panic(newIllegalFormatPrecisionException(math.MinInt32))
			}
			fs.precision = int(v)
		}
		if dt {
			fs.dt = true
			if dtUpper {
				fs.flags |= ffUppercase
			}
		}
		fs.conversion(s[j])
		i = j + 1
		if fs.dt {
			fs.checkDateTime()
		} else {
			fs.check()
		}
		parts = append(parts, formatPart{spec: fs})
	}
	return parts
}

func utf16ToASCII(u []uint16) []byte {
	b := make([]byte, len(u))
	for i, c := range u {
		b[i] = byte(c)
	}
	return b
}

func newSimpleSpec(c uint16) *formatSpec {
	fs := &formatSpec{width: -1, precision: -1, c: c}
	if Character_isUpperCase_char(c) {
		fs.flags = ffUppercase
		fs.c = Character_toLowerCase_char(c)
	}
	if c == 'n' || c == '%' {
		fs.index = -2
	}
	return fs
}

func (fs *formatSpec) conversion(c uint16) {
	fs.c = c
	if !fs.dt {
		if !isConversion(c) {
			panic(newUnknownFormatConversionException(string(rune(c))))
		}
		if Character_isUpperCase_char(c) {
			fs.flags |= ffUppercase
			fs.c = Character_toLowerCase_char(c)
		}
		if c == 'n' || c == '%' {
			fs.index = -2
		}
	}
}

func (fs *formatSpec) failMismatch(f int) {
	panic(newFormatFlagsConversionMismatchException(flagsString(f), fs.c))
}

func (fs *formatSpec) checkBadFlags(bad int) {
	if fs.flags&bad != 0 {
		fs.failMismatch(fs.flags & bad)
	}
}

// check implements the FormatSpecifier constructor's checks.
func (fs *formatSpec) check() {
	switch fs.c {
	case 'b', 's', 'h':
		if (fs.c == 'b' || fs.c == 'h') && hasFlags(fs.flags, ffAlternate) {
			fs.failMismatch(ffAlternate)
		}
		if fs.width == -1 && hasFlags(fs.flags, ffLeftJustify) {
			panic(newMissingFormatWidthException(fs.String()))
		}
		fs.checkBadFlags(ffPlus | ffLeadingSpace | ffZeroPad | ffGroup | ffParentheses)
	case 'c':
		if fs.precision != -1 {
			panic(newIllegalFormatPrecisionException(int32(fs.precision)))
		}
		fs.checkBadFlags(ffAlternate | ffPlus | ffLeadingSpace | ffZeroPad | ffGroup | ffParentheses)
		if fs.width == -1 && hasFlags(fs.flags, ffLeftJustify) {
			panic(newMissingFormatWidthException(fs.String()))
		}
	case 'd', 'o', 'x':
		fs.checkNumeric()
		if fs.precision != -1 {
			panic(newIllegalFormatPrecisionException(int32(fs.precision)))
		}
		if fs.c == 'd' {
			fs.checkBadFlags(ffAlternate)
		} else {
			fs.checkBadFlags(ffGroup)
		}
	case 'e', 'f', 'g', 'a':
		fs.checkNumeric()
		switch fs.c {
		case 'a':
			fs.checkBadFlags(ffParentheses | ffGroup)
		case 'e':
			fs.checkBadFlags(ffGroup)
		case 'g':
			fs.checkBadFlags(ffAlternate)
		}
	case '%', 'n':
		if fs.precision != -1 {
			panic(newIllegalFormatPrecisionException(int32(fs.precision)))
		}
		if fs.c == '%' {
			if fs.flags != ffLeftJustify && fs.flags != 0 {
				panic(newIllegalFormatFlagsException(flagsString(fs.flags)))
			}
			if fs.width == -1 && hasFlags(fs.flags, ffLeftJustify) {
				panic(newMissingFormatWidthException(fs.String()))
			}
		} else {
			if fs.width != -1 {
				panic(newIllegalFormatWidthException(int32(fs.width)))
			}
			if fs.flags != 0 {
				panic(newIllegalFormatFlagsException(flagsString(fs.flags)))
			}
		}
	}
}

// checkDateTime implements FormatSpecifier.checkDateTime. Only null
// arguments can be printed with a date/time conversion.
func (fs *formatSpec) checkDateTime() {
	if fs.precision != -1 {
		panic(newIllegalFormatPrecisionException(int32(fs.precision)))
	}
	switch fs.c {
	case 'H', 'I', 'k', 'l', 'M', 'S', 'L', 'N', 'p', 'z', 'Z', 's', 'Q', 'B', 'b', 'h', 'A', 'a', 'C', 'Y',
		'y', 'j', 'm', 'd', 'e', 'R', 'T', 'r', 'D', 'F', 'c':
	default:
		panic(newUnknownFormatConversionException("t" + string(rune(fs.c))))
	}
	fs.checkBadFlags(ffAlternate | ffPlus | ffLeadingSpace | ffZeroPad | ffGroup | ffParentheses)
	if fs.width == -1 && hasFlags(fs.flags, ffLeftJustify) {
		panic(newMissingFormatWidthException(fs.String()))
	}
}

func (fs *formatSpec) checkNumeric() {
	if fs.width == -1 && fs.flags&(ffLeftJustify|ffZeroPad) != 0 {
		panic(newMissingFormatWidthException(fs.String()))
	}
	if hasFlags(fs.flags, ffPlus|ffLeadingSpace) || hasFlags(fs.flags, ffLeftJustify|ffZeroPad) {
		panic(newIllegalFormatFlagsException(flagsString(fs.flags)))
	}
}

func (f *formatter) failConversion(fs *formatSpec, arg any) {
	panic(newIllegalFormatConversionException(fs.c, className(arg)))
}

func (f *formatter) print(fs *formatSpec, arg any) {
	if fs.dt {
		if arg != nil {
			panic(Unsupported("Formatter date/time conversions"))
		}
		f.printString(fs, Lit("null"))
		return
	}
	switch fs.c {
	case 'd', 'o', 'x':
		f.printInteger(fs, arg)
	case 'e', 'g', 'f', 'a':
		f.printFloat(fs, arg)
	case 'c':
		f.printCharacter(fs, arg)
	case 'b':
		switch {
		case arg == nil:
			f.printString(fs, Lit("false"))
		case Boolean_isInstance(arg):
			f.printString(fs, StrOfBool(arg.(bool)))
		default:
			f.printString(fs, Lit("true"))
		}
	case 's':
		if hasFlags(fs.flags, ffAlternate) {
			fs.failMismatch(ffAlternate)
		}
		f.printString(fs, StrOfObject(arg))
	case 'h':
		if arg == nil {
			f.printString(fs, Lit("null"))
		} else {
			f.printString(fs, Integer_toHexString(Object_HashCode(arg)))
		}
	case 'n':
		f.out = append(f.out, '\n')
	case '%':
		f.printString(fs, Lit("%"))
	}
}

// printString implements FormatSpecifier.print(Formatter, String, Locale).
func (f *formatter) printString(fs *formatSpec, s *String) {
	u := s.units
	if fs.precision != -1 && fs.precision < len(u) {
		u = u[:fs.precision]
	}
	if hasFlags(fs.flags, ffUppercase) {
		u = f.toUpperCase(u)
	}
	f.appendJustified(fs, u)
}

func (f *formatter) toUpperCase(u []uint16) []uint16 {
	l := f.locale
	if l == nil {
		l = Locale_getDefault()
	}
	return javaToUpperCase(u, l)
}

func (f *formatter) appendJustified(fs *formatSpec, u []uint16) {
	if fs.width == -1 {
		f.out = append(f.out, u...)
		return
	}
	padRight := hasFlags(fs.flags, ffLeftJustify)
	if padRight {
		f.out = append(f.out, u...)
	}
	for i := len(u); i < fs.width; i++ {
		f.out = append(f.out, ' ')
	}
	if !padRight {
		f.out = append(f.out, u...)
	}
}

func (f *formatter) printCharacter(fs *formatSpec, arg any) {
	var s *String
	codePoint := func(i int32) *String {
		if !Character_isValidCodePoint(i) {
			panic(newIllegalFormatCodePointException(i))
		}
		return FromUnits(Character_toChars_int(i))
	}
	switch x := arg.(type) {
	case nil:
		s = Lit("null")
	case uint16:
		s = StrOfChar(x)
	case int8:
		s = codePoint(int32(x))
	case int16:
		s = codePoint(int32(x))
	case int32:
		s = codePoint(x)
	default:
		f.failConversion(fs, arg)
	}
	f.printString(fs, s)
}

func (f *formatter) printInteger(fs *formatSpec, arg any) {
	var v int64
	switch x := arg.(type) {
	case nil:
		f.printString(fs, Lit("null"))
		return
	case int8:
		v = int64(x)
		if x < 0 && fs.c != 'd' {
			v += 1 << 8
		}
	case int16:
		v = int64(x)
		if x < 0 && fs.c != 'd' {
			v += 1 << 16
		}
	case int32:
		v = int64(x)
		if x < 0 && fs.c != 'd' {
			v += 1 << 32
		}
	case int64:
		v = x
	default:
		f.failConversion(fs, arg)
	}
	var sb []uint16
	switch fs.c {
	case 'd':
		neg := v < 0
		valueStr := strconv.FormatInt(v, 10)
		sb = f.leadingSign(fs, sb, neg)
		offset := 0
		if neg {
			offset = 1
		}
		sb = f.localizedMagnitude(sb, []byte(valueStr), offset, fs.flags, adjustWidth(fs, neg))
		sb = trailingSign(fs, sb, neg)
	case 'o', 'x':
		fs.checkBadFlags(ffParentheses | ffLeadingSpace | ffPlus)
		var s string
		var prefix string
		if fs.c == 'o' {
			s = strconv.FormatUint(uint64(v), 8)
			if hasFlags(fs.flags, ffAlternate) {
				prefix = "0"
			}
		} else {
			s = strconv.FormatUint(uint64(v), 16)
			if hasFlags(fs.flags, ffAlternate) {
				prefix = "0x"
				if hasFlags(fs.flags, ffUppercase) {
					prefix = "0X"
				}
			}
		}
		for i := 0; i < len(prefix); i++ {
			sb = append(sb, uint16(prefix[i]))
		}
		if hasFlags(fs.flags, ffZeroPad) {
			for i := len(prefix) + len(s); i < fs.width; i++ {
				sb = append(sb, '0')
			}
		}
		su := FromGo(s).units
		if fs.c == 'x' && hasFlags(fs.flags, ffUppercase) {
			su = f.toUpperCase(su)
		}
		sb = append(sb, su...)
	}
	f.appendJustified(fs, sb)
}

func (f *formatter) leadingSign(fs *formatSpec, sb []uint16, neg bool) []uint16 {
	switch {
	case !neg:
		if hasFlags(fs.flags, ffPlus) {
			sb = append(sb, '+')
		} else if hasFlags(fs.flags, ffLeadingSpace) {
			sb = append(sb, ' ')
		}
	case hasFlags(fs.flags, ffParentheses):
		sb = append(sb, '(')
	default:
		sb = append(sb, '-')
	}
	return sb
}

func trailingSign(fs *formatSpec, sb []uint16, neg bool) []uint16 {
	if neg && hasFlags(fs.flags, ffParentheses) {
		sb = append(sb, ')')
	}
	return sb
}

func adjustWidth(fs *formatSpec, neg bool) int {
	w := fs.width
	if w != -1 && neg && hasFlags(fs.flags, ffParentheses) {
		w--
	}
	return w
}

// localizedMagnitude implements Formatter's localizedMagnitude: it appends
// the ASCII digits value[offset:] with the locale's zero digit, decimal
// separator and, for the ',' flag, grouping, zero padding to width.
func (f *formatter) localizedMagnitude(sb []uint16, value []byte, offset int, flags int, width int) []uint16 {
	begin := len(sb)
	zero := uint16('0')
	decSep := uint16('.')
	grpSep := uint16(0)
	grpSize := -1
	var data *localeNumberData
	if f.locale != nil {
		data = numberDataFor(f.locale)
		zero = data.zero
		decSep = data.decimal
	}
	dot := len(value)
	for j := offset; j < len(value); j++ {
		if value[j] == '.' {
			dot = j
			break
		}
	}
	if hasFlags(flags, ffGroup) {
		if f.locale == nil {
			grpSep = ','
			grpSize = 3
		} else {
			grpSep = data.grouping
			if f.locale.Equals_Object(Locale_US) {
				grpSize = 3
			} else {
				df := newDecimalFormat(FromGo(data.numberPattern), DecimalFormatSymbols_new_Locale(f.locale))
				grpSize = int(df.groupingSize)
				if !df.groupingUsed || grpSize == 0 {
					grpSep = 0
				}
			}
		}
	}
	for j := offset; j < len(value); j++ {
		if j == dot {
			sb = append(sb, decSep)
			grpSep = 0
			continue
		}
		sb = append(sb, uint16(value[j]-'0')+zero)
		if grpSep != 0 && j != dot-1 && (dot-j)%grpSize == 1 {
			sb = append(sb, grpSep)
		}
	}
	if width > len(sb) && hasFlags(flags, ffZeroPad) {
		padded := make([]uint16, 0, width)
		padded = append(padded, sb[:begin]...)
		for i := len(sb); i < width; i++ {
			padded = append(padded, zero)
		}
		sb = append(padded, sb[begin:]...)
	}
	return sb
}

func (f *formatter) printFloat(fs *formatSpec, arg any) {
	switch x := arg.(type) {
	case nil:
		f.printString(fs, Lit("null"))
	case float32:
		f.printDouble(fs, float64(x))
	case float64:
		f.printDouble(fs, x)
	case *BigDecimal:
		f.printBigDecimal(fs, x)
	default:
		f.failConversion(fs, arg)
	}
}

func (f *formatter) printDouble(fs *formatSpec, value float64) {
	var sb []uint16
	neg := Double_compare(value, 0) == -1
	switch {
	case value == value:
		v := math.Abs(value)
		sb = f.leadingSign(fs, sb, neg)
		switch {
		case !math.IsInf(v, 0):
			sb = f.printFinite(fs, sb, v, neg)
		case hasFlags(fs.flags, ffUppercase):
			sb = append(sb, FromGo("INFINITY").units...)
		default:
			sb = append(sb, FromGo("Infinity").units...)
		}
		sb = trailingSign(fs, sb, neg)
	case hasFlags(fs.flags, ffUppercase):
		sb = append(sb, 'N', 'A', 'N')
	default:
		sb = append(sb, 'N', 'a', 'N')
	}
	f.appendJustified(fs, sb)
}

// printFinite formats a finite non-negative double for %e, %f, %g and %a.
func (f *formatter) printFinite(fs *formatSpec, sb []uint16, value float64, neg bool) []uint16 {
	switch fs.c {
	case 'e':
		prec := fs.precision
		if prec == -1 {
			prec = 6
		}
		fd := newFormattedFPDecimal(value)
		fd.round(int64(prec) + 1)
		fd.scientificChars(prec)
		mant := addZeros(fd.digits, prec)
		if hasFlags(fs.flags, ffAlternate) && prec == 0 {
			mant = append(mant, '.')
		}
		exp := fd.exp
		if value == 0 {
			exp = []byte("+00")
		}
		newW := fs.width
		if fs.width != -1 {
			newW = adjustWidthFor(fs, fs.width-len(exp)-1, neg)
		}
		sb = f.localizedMagnitude(sb, mant, 0, fs.flags, newW)
		if hasFlags(fs.flags, ffUppercase) {
			sb = append(sb, 'E')
		} else {
			sb = append(sb, 'e')
		}
		sb = append(sb, uint16(exp[0]))
		return f.localizedMagnitude(sb, exp[1:], 0, 0, -1)
	case 'f':
		prec := fs.precision
		if prec == -1 {
			prec = 6
		}
		fd := newFormattedFPDecimal(value)
		fd.round(int64(fd.n) + int64(fd.e) + int64(prec))
		fd.plainChars()
		mant := addZeros(fd.digits, prec)
		if hasFlags(fs.flags, ffAlternate) && prec == 0 {
			mant = append(mant, '.')
		}
		newW := fs.width
		if fs.width != -1 {
			newW = adjustWidth(fs, neg)
		}
		return f.localizedMagnitude(sb, mant, 0, fs.flags, newW)
	case 'g':
		prec := fs.precision
		if prec == -1 {
			prec = 6
		} else if prec == 0 {
			prec = 1
		}
		var exp []byte
		var mant []byte
		var expRounded int
		if value == 0 {
			mant = []byte{'0'}
		} else {
			fd := newFormattedFPDecimal(value)
			fd.round(int64(prec))
			er := fd.n + fd.e - 1
			if -4 <= er && er < prec {
				fd.plainChars()
			} else {
				fd.scientificChars(prec - 1)
			}
			exp = fd.exp
			mant = fd.digits
			expRounded = fd.n + fd.e - 1
		}
		if exp != nil {
			prec--
		} else {
			prec -= expRounded + 1
		}
		mant = addZeros(mant, prec)
		if hasFlags(fs.flags, ffAlternate) && prec == 0 {
			mant = append(mant, '.')
		}
		newW := fs.width
		if fs.width != -1 {
			if exp != nil {
				newW = adjustWidthFor(fs, fs.width-len(exp)-1, neg)
			} else {
				newW = adjustWidth(fs, neg)
			}
		}
		sb = f.localizedMagnitude(sb, mant, 0, fs.flags, newW)
		if exp != nil {
			if hasFlags(fs.flags, ffUppercase) {
				sb = append(sb, 'E')
			} else {
				sb = append(sb, 'e')
			}
			sb = append(sb, uint16(exp[0]))
			sb = f.localizedMagnitude(sb, exp[1:], 0, 0, -1)
		}
		return sb
	case 'a':
		prec := fs.precision
		if prec == -1 {
			prec = 0
		} else if prec == 0 {
			prec = 1
		}
		s := hexDouble(value, prec)
		upper := hasFlags(fs.flags, ffUppercase)
		if upper {
			sb = append(sb, '0', 'X')
		} else {
			sb = append(sb, '0', 'x')
		}
		if hasFlags(fs.flags, ffZeroPad) {
			leading := 2
			if hasFlags(fs.flags, ffLeadingSpace) || hasFlags(fs.flags, ffPlus) || neg {
				leading = 3
			}
			for i := 0; i < fs.width-len(s)-leading; i++ {
				sb = append(sb, '0')
			}
		}
		idx := 0
		for idx < len(s) && s[idx] != 'p' {
			idx++
		}
		va := []byte(s[:idx])
		if upper {
			va = []byte(asciiUpper(string(va)))
		}
		if prec != 0 {
			va = addZeros(va, prec)
		}
		for _, c := range va {
			sb = append(sb, uint16(c))
		}
		if upper {
			sb = append(sb, 'P')
		} else {
			sb = append(sb, 'p')
		}
		for i := idx + 1; i < len(s); i++ {
			sb = append(sb, uint16(s[i]))
		}
		return sb
	}
	return sb
}

func adjustWidthFor(fs *formatSpec, width int, neg bool) int {
	if width != -1 && neg && hasFlags(fs.flags, ffParentheses) {
		width--
	}
	return width
}

// addZeros implements Formatter's addZeros: pad the fraction of mant with
// zeros to prec digits, adding a point if needed.
func addZeros(mant []byte, prec int) []byte {
	i := 0
	for i < len(mant) && mant[i] != '.' {
		i++
	}
	needDot := i == len(mant)
	outPrec := len(mant) - i - 1
	if needDot {
		outPrec = 0
	}
	if outPrec == prec {
		return mant
	}
	out := append([]byte(nil), mant...)
	if needDot {
		out = append(out, '.')
	}
	for j := outPrec; j < prec; j++ {
		out = append(out, '0')
	}
	return out
}

// hexDouble implements Formatter's hexDouble for a finite non-negative d.
func hexDouble(d float64, prec int) string {
	if d == 0 || prec == 0 || prec >= 13 {
		return doubleToHexString(d)[2:]
	}
	exponent := mathGetExponent(d)
	subnormal := exponent == -1023
	if subnormal {
		d *= 0x1p54
	}
	precision := 1 + prec*4
	shiftDistance := uint(53 - precision)
	doppel := int64(math.Float64bits(d))
	newSignif := (doppel & (0x7ff0000000000000 | (1<<52 - 1))) >> shiftDistance
	roundingBits := doppel & ^(^int64(0) << shiftDistance)
	leastZero := newSignif&1 == 0
	round := (int64(1)<<(shiftDistance-1))&roundingBits != 0
	sticky := shiftDistance > 1 && (^(int64(1)<<(shiftDistance-1)))&roundingBits != 0
	if leastZero && round && sticky || !leastZero && round {
		newSignif++
	}
	signBit := doppel & math.MinInt64
	newSignif = signBit | newSignif<<shiftDistance
	result := math.Float64frombits(uint64(newSignif))
	if math.IsInf(result, 0) {
		return "1.0p1024"
	}
	res := doubleToHexString(result)[2:]
	if !subnormal {
		return res
	}
	idx := 0
	for res[idx] != 'p' {
		idx++
	}
	e, _ := strconv.Atoi(res[idx+1:])
	return res[:idx] + "p" + strconv.Itoa(e-54)
}

// mathGetExponent implements Math.getExponent(double).
func mathGetExponent(d float64) int {
	return int((math.Float64bits(d)&0x7ff0000000000000)>>52) - 1023
}

// formattedFPDecimal implements jdk.internal.math.FormattedFPDecimal: the
// decimal f*10^e (n digits) selected by Double.toString, rounded half up to
// the precision of a %e, %f or %g conversion.
type formattedFPDecimal struct {
	f      int64
	e      int
	n      int
	digits []byte
	exp    []byte
}

func newFormattedFPDecimal(v float64) *formattedFPDecimal {
	fd := &formattedFPDecimal{}
	b := math.Float64bits(v)
	if b&^(1<<63) == 0 {
		return fd
	}
	f, e := doubleToDecimal(b)
	n := flog10pow2(64 - bits.LeadingZeros64(uint64(f)))
	if f >= pow10Int64(n) {
		n++
	}
	fd.f, fd.e, fd.n = f, e, n
	return fd
}

func pow10Int64(n int) int64 {
	p := int64(1)
	for ; n > 0; n-- {
		p *= 10
	}
	return p
}

func (fd *formattedFPDecimal) round(pp int64) {
	if fd.n == 0 || pp < 0 {
		fd.f, fd.e, fd.n = 0, 0, 1
		return
	}
	if pp >= int64(fd.n) {
		return
	}
	p := int(pp)
	fd.e += fd.n - p
	pow10 := pow10Int64(fd.n - p)
	fd.f = (fd.f + pow10>>1) / pow10
	if p == 0 {
		fd.n = 1
		if fd.f == 0 {
			fd.e = 0
		}
		return
	}
	fd.n = p
	if fd.f == pow10Int64(p) {
		fd.f /= 10
		fd.e++
	}
}

func (fd *formattedFPDecimal) plainChars() {
	switch {
	case fd.e >= 0:
		fd.digits = make([]byte, fd.n+fd.e)
		fd.fillWithZeros(fd.n, fd.n+fd.e)
		fd.fillWithDigits(fd.f, 0, fd.n)
	case fd.n+fd.e > 0:
		fd.digits = make([]byte, fd.n+1)
		x := fd.fillWithDigits(fd.f, fd.n+1+fd.e, fd.n+1)
		fd.digits[fd.n+fd.e] = '.'
		fd.fillWithDigits(x, 0, fd.n+fd.e)
	default:
		fd.digits = make([]byte, 2-fd.e)
		fd.fillWithDigits(fd.f, 2-fd.e-fd.n, 2-fd.e)
		fd.fillWithZeros(0, 2-fd.e-fd.n)
		fd.digits[1] = '.'
	}
}

func (fd *formattedFPDecimal) scientificChars(prec int) {
	if prec != 0 {
		fd.digits = make([]byte, 1+fd.n)
		x := fd.fillWithDigits(fd.f, 2, 1+fd.n)
		fd.digits[1] = '.'
		fd.digits[0] = byte('0' + x)
	} else {
		fd.digits = []byte{byte('0' + fd.f)}
	}
	er := fd.n + fd.e - 1
	aer := er
	if aer < 0 {
		aer = -aer
	}
	sign := byte('+')
	if er < 0 {
		sign = '-'
	}
	if aer >= 100 {
		fd.exp = []byte{sign, byte('0' + aer/100), byte('0' + aer/10%10), byte('0' + aer%10)}
	} else {
		fd.exp = []byte{sign, byte('0' + aer/10), byte('0' + aer%10)}
	}
}

func (fd *formattedFPDecimal) fillWithDigits(x int64, from, to int) int64 {
	for to > from {
		q := x / 10
		to--
		fd.digits[to] = byte('0' + x - q*10)
		x = q
	}
	return x
}

func (fd *formattedFPDecimal) fillWithZeros(from, to int) {
	for to > from {
		to--
		fd.digits[to] = '0'
	}
}

// printBigDecimal implements Formatter's BigDecimal printing.
func (f *formatter) printBigDecimal(fs *formatSpec, value *BigDecimal) {
	if fs.c == 'a' {
		f.failConversion(fs, value)
	}
	var sb []uint16
	neg := value.Signum() == -1
	v := value.Abs()
	sb = f.leadingSign(fs, sb, neg)
	sb = f.printBigDecimalMagnitude(fs, sb, v, fs.c, fs.precision, neg)
	sb = trailingSign(fs, sb, neg)
	f.appendJustified(fs, sb)
}

func (f *formatter) printBigDecimalMagnitude(fs *formatSpec, sb []uint16, value *BigDecimal, c uint16, precision int, neg bool) []uint16 {
	switch c {
	case 'e':
		prec := precision
		if prec == -1 {
			prec = 6
		}
		origPrec := int(value.Precision())
		nzeros := 0
		var compPrec int
		if prec > origPrec-1 {
			compPrec = origPrec
			nzeros = prec - (origPrec - 1)
		} else {
			compPrec = prec + 1
		}
		v := value.roundToPrecision(int32(compPrec), rmHalfUp)
		mant, exp, hasDot := bigDecimalLayout(v.intVal, v.scale, true)
		if (origPrec == 1 || !hasDot) && (nzeros > 0 || hasFlags(fs.flags, ffAlternate)) {
			mant = append(mant, '.')
		}
		for i := 0; i < nzeros; i++ {
			mant = append(mant, '0')
		}
		newW := fs.width
		if fs.width != -1 {
			newW = adjustWidthFor(fs, fs.width-len(exp)-1, neg)
		}
		sb = f.localizedMagnitude(sb, mant, 0, fs.flags, newW)
		if hasFlags(fs.flags, ffUppercase) {
			sb = append(sb, 'E')
		} else {
			sb = append(sb, 'e')
		}
		sb = append(sb, uint16(exp[0]))
		return f.localizedMagnitude(sb, exp, 1, fs.flags&^ffGroup, -1)
	case 'f':
		prec := precision
		if prec == -1 {
			prec = 6
		}
		scale := int(value.scale)
		if scale > prec {
			compPrec := int(value.Precision())
			if compPrec <= scale {
				value = value.SetScale_int_int(int32(prec), rmHalfUp)
			} else {
				compPrec -= scale - prec
				value = value.roundToPrecision(int32(compPrec), rmHalfUp)
			}
		}
		mant, _, _ := bigDecimalLayout(value.intVal, value.scale, false)
		layoutScale := int(value.scale)
		if layoutScale < 0 {
			layoutScale = 0
		}
		nzeros := 0
		if layoutScale < prec {
			nzeros = prec - layoutScale
		}
		if layoutScale == 0 && (hasFlags(fs.flags, ffAlternate) || nzeros > 0) {
			mant = append(mant, '.')
		}
		for i := 0; i < nzeros; i++ {
			mant = append(mant, '0')
		}
		return f.localizedMagnitude(sb, mant, 0, fs.flags, adjustWidth(fs, neg))
	case 'g':
		prec := precision
		if precision == -1 {
			prec = 6
		} else if precision == 0 {
			prec = 1
		}
		value = value.roundToPrecision(int32(prec), rmHalfUp)
		lower := newBigDecimal(big.NewInt(1), 4)
		upper := newBigDecimal(big.NewInt(1), int32(-prec))
		if value.Equals_Object(BigDecimal_ZERO) || value.compare(lower) != -1 && value.compare(upper) == -1 {
			e := -int(value.scale) + (len(value.intVal.String()) - 1)
			prec = prec - e - 1
			return f.printBigDecimalMagnitude(fs, sb, value, 'f', prec, neg)
		}
		return f.printBigDecimalMagnitude(fs, sb, value, 'e', prec-1, neg)
	}
	return sb
}

// bigDecimalLayout implements Formatter.BigDecimalLayout for a non-negative
// unscaled value.
func bigDecimalLayout(intVal *big.Int, scale int32, scientific bool) (mant, exp []byte, dot bool) {
	coeff := intVal.String()
	n := len(coeff)
	if scale == 0 {
		if n > 1 {
			mant = append(mant, coeff[0])
			if scientific {
				mant = append(mant, '.')
				dot = true
				mant = append(mant, coeff[1:]...)
				exp = []byte("+")
				if n < 10 {
					exp = append(exp, '0')
				}
				exp = strconv.AppendInt(exp, int64(n-1), 10)
			} else {
				mant = append(mant, coeff[1:]...)
			}
		} else {
			mant = append(mant, coeff...)
			if scientific {
				exp = []byte("+00")
			}
		}
		return mant, exp, dot
	}
	if !scientific {
		switch {
		case int(scale) >= n:
			mant = append(mant, '0', '.')
			dot = true
			for i := n; i < int(scale); i++ {
				mant = append(mant, '0')
			}
			mant = append(mant, coeff...)
		case scale > 0:
			pad := n - int(scale)
			mant = append(mant, coeff[:pad]...)
			mant = append(mant, '.')
			dot = true
			mant = append(mant, coeff[pad:]...)
		default:
			mant = append(mant, coeff...)
			if intVal.Sign() != 0 {
				for i := 0; i < int(-scale); i++ {
					mant = append(mant, '0')
				}
			}
		}
		return mant, nil, dot
	}
	mant = append(mant, coeff[0])
	if n > 1 {
		mant = append(mant, '.')
		dot = true
		mant = append(mant, coeff[1:]...)
	}
	adjusted := -int64(scale) + int64(n-1)
	if adjusted != 0 {
		abs := adjusted
		if abs < 0 {
			abs = -abs
			exp = []byte{'-'}
		} else {
			exp = []byte{'+'}
		}
		if abs < 10 {
			exp = append(exp, '0')
		}
		exp = strconv.AppendInt(exp, abs, 10)
	} else {
		exp = []byte("+00")
	}
	return mant, exp, dot
}

func newUnknownFormatConversionException(s string) any {
	return UnknownFormatConversionException_new(FromGo(s))
}

func newMissingFormatArgumentException(s string) any {
	return MissingFormatArgumentException_new(FromGo(s))
}

func newDuplicateFormatFlagsException(f string) any {
	return DuplicateFormatFlagsException_new(FromGo(f))
}

func newIllegalFormatFlagsException(f string) any { return IllegalFormatFlagsException_new(FromGo(f)) }

func newMissingFormatWidthException(s string) any { return MissingFormatWidthException_new(FromGo(s)) }

func newIllegalFormatWidthException(w int32) any { return IllegalFormatWidthException_new(w) }

func newIllegalFormatPrecisionException(p int32) any { return IllegalFormatPrecisionException_new(p) }

func newIllegalFormatCodePointException(c int32) any { return IllegalFormatCodePointException_new(c) }

func newFormatFlagsConversionMismatchException(f string, c uint16) any {
	return FormatFlagsConversionMismatchException_new(FromGo(f), c)
}

func newIllegalFormatConversionException(c uint16, class string) any {
	return IllegalFormatConversionException_new(c, &Class{name: class})
}
