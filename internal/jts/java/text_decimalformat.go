package java

import (
	"math"
	"math/big"
	"strconv"
)

// This file implements java.text.DecimalFormatSymbols, NumberFormat and
// DecimalFormat (formatting only), and java.math.RoundingMode. Double values
// are converted to digits with FloatingDecimal's conversion (fdConvert) and
// then rounded by java.text.DigitList, exactly as the JDK does, including the
// fast path DecimalFormat takes for patterns like "#,##0.###".

// ---------------------------------------------------------------------------
// RoundingMode

// RoundingMode is java.math.RoundingMode.
type RoundingMode struct {
	Enum
}

func (this *RoundingMode) JClass_RoundingMode() *RoundingMode { return this }
func (this *RoundingMode) JClassName() string                 { return "java.math.RoundingMode" }

// RoundingMode_isInstance implements instanceof RoundingMode.
func RoundingMode_isInstance(o any) bool { _, ok := o.(*RoundingMode); return ok }

// RoundingMode_fromAny casts o to RoundingMode.
func RoundingMode_fromAny(o any) *RoundingMode { return AnyTo[*RoundingMode](o) }

func newRoundingMode(name string, ordinal int32) *RoundingMode {
	r := &RoundingMode{}
	r.jself = r
	return EnumConst(r, name, ordinal)
}

// RoundingMode_UP is RoundingMode.UP.
var RoundingMode_UP = newRoundingMode("UP", 0)

// RoundingMode_DOWN is RoundingMode.DOWN.
var RoundingMode_DOWN = newRoundingMode("DOWN", 1)

// RoundingMode_CEILING is RoundingMode.CEILING.
var RoundingMode_CEILING = newRoundingMode("CEILING", 2)

// RoundingMode_FLOOR is RoundingMode.FLOOR.
var RoundingMode_FLOOR = newRoundingMode("FLOOR", 3)

// RoundingMode_HALF_UP is RoundingMode.HALF_UP.
var RoundingMode_HALF_UP = newRoundingMode("HALF_UP", 4)

// RoundingMode_HALF_DOWN is RoundingMode.HALF_DOWN.
var RoundingMode_HALF_DOWN = newRoundingMode("HALF_DOWN", 5)

// RoundingMode_HALF_EVEN is RoundingMode.HALF_EVEN.
var RoundingMode_HALF_EVEN = newRoundingMode("HALF_EVEN", 6)

// RoundingMode_UNNECESSARY is RoundingMode.UNNECESSARY.
var RoundingMode_UNNECESSARY = newRoundingMode("UNNECESSARY", 7)

// RoundingMode_values implements RoundingMode.values.
func RoundingMode_values() []*RoundingMode {
	return []*RoundingMode{
		RoundingMode_UP, RoundingMode_DOWN, RoundingMode_CEILING, RoundingMode_FLOOR,
		RoundingMode_HALF_UP, RoundingMode_HALF_DOWN, RoundingMode_HALF_EVEN, RoundingMode_UNNECESSARY,
	}
}

// RoundingMode_valueOf_int implements RoundingMode.valueOf(int), which maps
// the legacy BigDecimal.ROUND_ constants.
func RoundingMode_valueOf_int(rm int32) *RoundingMode {
	if rm < 0 || rm > 7 {
		panic(IllegalArgumentException_new_String(Lit("argument out of range")))
	}
	return RoundingMode_values()[rm]
}

// RoundingMode_valueOf_String implements RoundingMode.valueOf(String).
func RoundingMode_valueOf_String(name *String) *RoundingMode {
	if name == nil {
		panic(NullPointerException_new_String(Lit("Name is null")))
	}
	for _, r := range RoundingMode_values() {
		if r.name.Equals_Object(name) {
			return r
		}
	}
	panic(IllegalArgumentException_new_String(Concat(Lit("No enum constant java.math.RoundingMode."), name)))
}

// ---------------------------------------------------------------------------
// DecimalFormatSymbols

// DecimalFormatSymbols is java.text.DecimalFormatSymbols, without the
// currency symbols.
type DecimalFormatSymbols struct {
	zero, grouping, decimal, perMill, percent, digit uint16
	patternSeparator, minus                          uint16
	monetaryDecimal, monetaryGrouping                uint16
	exponent, infinity, nan                          *String
	percentText, perMillText, minusText              *String
	locale                                           *Locale
}

func (this *DecimalFormatSymbols) JClass_DecimalFormatSymbols() *DecimalFormatSymbols { return this }
func (this *DecimalFormatSymbols) JClassName() string                                 { return "java.text.DecimalFormatSymbols" }

// DecimalFormatSymbols_isInstance implements instanceof DecimalFormatSymbols.
func DecimalFormatSymbols_isInstance(o any) bool { _, ok := o.(*DecimalFormatSymbols); return ok }

// DecimalFormatSymbols_fromAny casts o to DecimalFormatSymbols.
func DecimalFormatSymbols_fromAny(o any) *DecimalFormatSymbols {
	return AnyTo[*DecimalFormatSymbols](o)
}

// DecimalFormatSymbols_new implements new DecimalFormatSymbols().
func DecimalFormatSymbols_new() *DecimalFormatSymbols {
	return DecimalFormatSymbols_new_Locale(Locale_getDefault())
}

// DecimalFormatSymbols_new_Locale implements new DecimalFormatSymbols(Locale).
func DecimalFormatSymbols_new_Locale(l *Locale) *DecimalFormatSymbols {
	if l == nil {
		panic(NullPointerException_new())
	}
	d := numberDataFor(l)
	return &DecimalFormatSymbols{
		zero: d.zero, grouping: d.grouping, decimal: d.decimal, perMill: d.perMill, percent: d.percent,
		digit: d.digit, patternSeparator: d.patternSeparator, minus: d.minus,
		monetaryDecimal: d.monetaryDecimal, monetaryGrouping: d.monetaryGrouping,
		exponent: FromGo(d.exponent), infinity: FromGo(d.infinity), nan: FromGo(d.nan),
		percentText: FromGo(d.percentText), perMillText: FromGo(d.perMillText), minusText: FromGo(d.minusText),
		locale: l,
	}
}

// DecimalFormatSymbols_getInstance implements DecimalFormatSymbols.getInstance().
func DecimalFormatSymbols_getInstance() *DecimalFormatSymbols { return DecimalFormatSymbols_new() }

// DecimalFormatSymbols_getInstance_Locale implements
// DecimalFormatSymbols.getInstance(Locale).
func DecimalFormatSymbols_getInstance_Locale(l *Locale) *DecimalFormatSymbols {
	return DecimalFormatSymbols_new_Locale(l)
}

func (this *DecimalFormatSymbols) GetZeroDigit() uint16                { return this.zero }
func (this *DecimalFormatSymbols) SetZeroDigit_char(c uint16)          { this.zero = c }
func (this *DecimalFormatSymbols) GetGroupingSeparator() uint16        { return this.grouping }
func (this *DecimalFormatSymbols) GetDecimalSeparator() uint16         { return this.decimal }
func (this *DecimalFormatSymbols) SetDecimalSeparator_char(c uint16)   { this.decimal = c }
func (this *DecimalFormatSymbols) GetPerMill() uint16                  { return this.perMill }
func (this *DecimalFormatSymbols) GetPercent() uint16                  { return this.percent }
func (this *DecimalFormatSymbols) GetDigit() uint16                    { return this.digit }
func (this *DecimalFormatSymbols) SetDigit_char(c uint16)              { this.digit = c }
func (this *DecimalFormatSymbols) GetPatternSeparator() uint16         { return this.patternSeparator }
func (this *DecimalFormatSymbols) SetPatternSeparator_char(c uint16)   { this.patternSeparator = c }
func (this *DecimalFormatSymbols) GetInfinity() *String                { return this.infinity }
func (this *DecimalFormatSymbols) SetInfinity_String(s *String)        { this.infinity = s }
func (this *DecimalFormatSymbols) GetNaN() *String                     { return this.nan }
func (this *DecimalFormatSymbols) SetNaN_String(s *String)             { this.nan = s }
func (this *DecimalFormatSymbols) GetMinusSign() uint16                { return this.minus }
func (this *DecimalFormatSymbols) GetExponentSeparator() *String       { return this.exponent }
func (this *DecimalFormatSymbols) GetMonetaryDecimalSeparator() uint16 { return this.monetaryDecimal }

func (this *DecimalFormatSymbols) SetGroupingSeparator_char(c uint16) { this.grouping = c }

func (this *DecimalFormatSymbols) SetMonetaryDecimalSeparator_char(c uint16) {
	this.monetaryDecimal = c
}

func (this *DecimalFormatSymbols) SetPerMill_char(c uint16) {
	this.perMill = c
	this.perMillText = StrOfChar(c)
}

func (this *DecimalFormatSymbols) SetPercent_char(c uint16) {
	this.percent = c
	this.percentText = StrOfChar(c)
}

func (this *DecimalFormatSymbols) SetMinusSign_char(c uint16) {
	this.minus = c
	this.minusText = StrOfChar(c)
}

func (this *DecimalFormatSymbols) SetExponentSeparator_String(s *String) {
	if s == nil {
		panic(NullPointerException_new())
	}
	this.exponent = s
}

// Clone implements DecimalFormatSymbols.clone.
func (this *DecimalFormatSymbols) Clone() any {
	c := *this
	return &c
}

// ---------------------------------------------------------------------------
// NumberFormat

// NumberFormat is java.text.NumberFormat. Its only concrete class is
// DecimalFormat, to which the methods that DecimalFormat overrides dispatch.
type NumberFormat struct {
	jself            any
	groupingUsed     bool
	maxIntDigits     int32
	minIntDigits     int32
	maxFracDigits    int32
	minFracDigits    int32
	parseIntegerOnly bool
}

func (this *NumberFormat) JSelf() any                         { return this.jself }
func (this *NumberFormat) JSetSelf(self any)                  { this.jself = self }
func (this *NumberFormat) JClass_NumberFormat() *NumberFormat { return this }
func (this *NumberFormat) JClassName() string                 { return "java.text.NumberFormat" }

// NumberFormat_isInstance implements instanceof NumberFormat.
func NumberFormat_isInstance(o any) bool {
	_, ok := o.(interface{ JClass_NumberFormat() *NumberFormat })
	return ok
}

// NumberFormat_fromAny casts o to NumberFormat.
func NumberFormat_fromAny(o any) *NumberFormat {
	if o == nil {
		return nil
	}
	return AnyTo[interface{ JClass_NumberFormat() *NumberFormat }](o).JClass_NumberFormat()
}

func (this *NumberFormat) df() *DecimalFormat { return this.jself.(*DecimalFormat) }

func (this *NumberFormat) init() {
	this.groupingUsed = true
	this.maxIntDigits = 40
	this.minIntDigits = 1
	this.maxFracDigits = 3
}

func (this *NumberFormat) setMaximumIntegerDigits(v int32) {
	this.maxIntDigits = max32(0, v)
	if this.minIntDigits > this.maxIntDigits {
		this.minIntDigits = this.maxIntDigits
	}
}

func (this *NumberFormat) setMinimumIntegerDigits(v int32) {
	this.minIntDigits = max32(0, v)
	if this.minIntDigits > this.maxIntDigits {
		this.maxIntDigits = this.minIntDigits
	}
}

func (this *NumberFormat) setMaximumFractionDigits(v int32) {
	this.maxFracDigits = max32(0, v)
	if this.maxFracDigits < this.minFracDigits {
		this.minFracDigits = this.maxFracDigits
	}
}

func (this *NumberFormat) setMinimumFractionDigits(v int32) {
	this.minFracDigits = max32(0, v)
	if this.maxFracDigits < this.minFracDigits {
		this.maxFracDigits = this.minFracDigits
	}
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

// NumberFormat_getInstance implements NumberFormat.getInstance().
func NumberFormat_getInstance() *NumberFormat {
	return NumberFormat_getInstance_Locale(Locale_getDefault())
}

// NumberFormat_getInstance_Locale implements NumberFormat.getInstance(Locale).
func NumberFormat_getInstance_Locale(l *Locale) *NumberFormat {
	return NumberFormat_getNumberInstance_Locale(l)
}

// NumberFormat_getNumberInstance implements NumberFormat.getNumberInstance().
func NumberFormat_getNumberInstance() *NumberFormat {
	return NumberFormat_getNumberInstance_Locale(Locale_getDefault())
}

// NumberFormat_getNumberInstance_Locale implements
// NumberFormat.getNumberInstance(Locale).
func NumberFormat_getNumberInstance_Locale(l *Locale) *NumberFormat {
	s := DecimalFormatSymbols_new_Locale(l)
	return &DecimalFormat_new_String_DecimalFormatSymbols(FromGo(numberDataFor(l).numberPattern), s).NumberFormat
}

// NumberFormat_getIntegerInstance implements NumberFormat.getIntegerInstance().
func NumberFormat_getIntegerInstance() *NumberFormat {
	return NumberFormat_getIntegerInstance_Locale(Locale_getDefault())
}

// NumberFormat_getIntegerInstance_Locale implements
// NumberFormat.getIntegerInstance(Locale).
func NumberFormat_getIntegerInstance_Locale(l *Locale) *NumberFormat {
	nf := NumberFormat_getNumberInstance_Locale(l)
	f := nf.df()
	f.SetMaximumFractionDigits_int(0)
	f.SetDecimalSeparatorAlwaysShown_boolean(false)
	f.SetParseIntegerOnly_boolean(true)
	return nf
}

// NumberFormat_getPercentInstance implements NumberFormat.getPercentInstance().
func NumberFormat_getPercentInstance() *NumberFormat {
	return NumberFormat_getPercentInstance_Locale(Locale_getDefault())
}

// NumberFormat_getPercentInstance_Locale implements
// NumberFormat.getPercentInstance(Locale).
func NumberFormat_getPercentInstance_Locale(l *Locale) *NumberFormat {
	s := DecimalFormatSymbols_new_Locale(l)
	return &DecimalFormat_new_String_DecimalFormatSymbols(FromGo(numberDataFor(l).percentPattern), s).NumberFormat
}

// Format_double implements NumberFormat.format(double), which tries
// DecimalFormat's fast path first.
func (this *NumberFormat) Format_double(d float64) *String {
	f := this.df()
	if s := f.fastFormat(d); s != nil {
		return FromUnits(s)
	}
	return FromUnits(f.formatDouble(d, nil))
}

// Format_long implements NumberFormat.format(long).
func (this *NumberFormat) Format_long(v int64) *String {
	return FromUnits(this.df().formatLong(v, nil))
}

// Format_Object implements Format.format(Object). Unlike format(double), it
// does not use the fast path.
func (this *NumberFormat) Format_Object(o any) *String {
	f := this.df()
	switch x := o.(type) {
	case int64, int32, int16, int8:
		return FromUnits(f.formatLong(Number_LongValue(x), nil))
	case *BigDecimal:
		return FromUnits(f.formatBigDecimal(x, nil))
	case float64, float32:
		return FromUnits(f.formatDouble(Number_DoubleValue(x), nil))
	}
	panic(IllegalArgumentException_new_String(Lit("Cannot format given Object as a Number")))
}

func (this *NumberFormat) IsGroupingUsed() bool { return this.groupingUsed }

func (this *NumberFormat) SetGroupingUsed_boolean(v bool) { this.df().SetGroupingUsed_boolean(v) }

func (this *NumberFormat) IsParseIntegerOnly() bool { return this.parseIntegerOnly }

func (this *NumberFormat) SetParseIntegerOnly_boolean(v bool) { this.parseIntegerOnly = v }

func (this *NumberFormat) GetMaximumIntegerDigits() int32 { return this.df().GetMaximumIntegerDigits() }

func (this *NumberFormat) SetMaximumIntegerDigits_int(v int32) {
	this.df().SetMaximumIntegerDigits_int(v)
}

func (this *NumberFormat) GetMinimumIntegerDigits() int32 { return this.df().GetMinimumIntegerDigits() }

func (this *NumberFormat) SetMinimumIntegerDigits_int(v int32) {
	this.df().SetMinimumIntegerDigits_int(v)
}

func (this *NumberFormat) GetMaximumFractionDigits() int32 {
	return this.df().GetMaximumFractionDigits()
}

func (this *NumberFormat) SetMaximumFractionDigits_int(v int32) {
	this.df().SetMaximumFractionDigits_int(v)
}

func (this *NumberFormat) GetMinimumFractionDigits() int32 {
	return this.df().GetMinimumFractionDigits()
}

func (this *NumberFormat) SetMinimumFractionDigits_int(v int32) {
	this.df().SetMinimumFractionDigits_int(v)
}

func (this *NumberFormat) GetRoundingMode() *RoundingMode { return this.df().roundingMode }

func (this *NumberFormat) SetRoundingMode_RoundingMode(m *RoundingMode) {
	this.df().SetRoundingMode_RoundingMode(m)
}

// Clone implements NumberFormat.clone.
func (this *NumberFormat) Clone() any { return this.df().Clone() }

// ---------------------------------------------------------------------------
// DecimalFormat

const (
	dfDoubleIntegerDigits   = 309
	dfDoubleFractionDigits  = 340
	dfMaximumIntegerDigits  = math.MaxInt32
	dfMaximumFractionDigits = math.MaxInt32
)

// DecimalFormat is java.text.DecimalFormat.
type DecimalFormat struct {
	NumberFormat
	symbols                     *DecimalFormatSymbols
	posPrefixPattern            []uint16
	posSuffixPattern            []uint16
	negPrefixPattern            []uint16
	negSuffixPattern            []uint16
	positivePrefix              []uint16
	positiveSuffix              []uint16
	negativePrefix              []uint16
	negativeSuffix              []uint16
	multiplier                  int32
	groupingSize                int8
	decimalSeparatorAlwaysShown bool
	isCurrencyFormat            bool
	useExponentialNotation      bool
	minExponentDigits           int8
	maximumIntegerDigits        int32
	minimumIntegerDigits        int32
	maximumFractionDigits       int32
	minimumFractionDigits       int32
	roundingMode                *RoundingMode
}

func (this *DecimalFormat) JClass_DecimalFormat() *DecimalFormat { return this }
func (this *DecimalFormat) JClassName() string                   { return "java.text.DecimalFormat" }

// DecimalFormat_isInstance implements instanceof DecimalFormat.
func DecimalFormat_isInstance(o any) bool { _, ok := o.(*DecimalFormat); return ok }

// DecimalFormat_fromAny casts o to DecimalFormat.
func DecimalFormat_fromAny(o any) *DecimalFormat { return AnyTo[*DecimalFormat](o) }

func newDecimalFormat(pattern *String, symbols *DecimalFormatSymbols) *DecimalFormat {
	this := &DecimalFormat{}
	this.jself = this
	this.NumberFormat.init()
	this.maximumIntegerDigits = 40
	this.minimumIntegerDigits = 1
	this.maximumFractionDigits = 3
	this.multiplier = 1
	this.groupingSize = 3
	this.roundingMode = RoundingMode_HALF_EVEN
	this.symbols = symbols
	this.applyPattern(pattern.units)
	return this
}

// DecimalFormat_new implements new DecimalFormat(), using the number pattern
// of the default locale.
func DecimalFormat_new() *DecimalFormat {
	l := Locale_getDefault()
	return newDecimalFormat(FromGo(numberDataFor(l).numberPattern), DecimalFormatSymbols_new_Locale(l))
}

// DecimalFormat_new_String implements new DecimalFormat(String).
func DecimalFormat_new_String(pattern *String) *DecimalFormat {
	if pattern == nil {
		panic(NullPointerException_new())
	}
	return newDecimalFormat(pattern, DecimalFormatSymbols_new_Locale(Locale_getDefault()))
}

// DecimalFormat_new_String_DecimalFormatSymbols implements new
// DecimalFormat(String, DecimalFormatSymbols).
func DecimalFormat_new_String_DecimalFormatSymbols(pattern *String, symbols *DecimalFormatSymbols) *DecimalFormat {
	if pattern == nil || symbols == nil {
		panic(NullPointerException_new())
	}
	return newDecimalFormat(pattern, symbols.Clone().(*DecimalFormatSymbols))
}

// Clone implements DecimalFormat.clone.
func (this *DecimalFormat) Clone() any {
	c := *this
	c.jself = &c
	c.symbols = this.symbols.Clone().(*DecimalFormatSymbols)
	return &c
}

func (this *DecimalFormat) GetDecimalFormatSymbols() *DecimalFormatSymbols {
	return this.symbols.Clone().(*DecimalFormatSymbols)
}

func (this *DecimalFormat) SetDecimalFormatSymbols_DecimalFormatSymbols(s *DecimalFormatSymbols) {
	if s == nil {
		return
	}
	this.symbols = s.Clone().(*DecimalFormatSymbols)
	this.expandAffixes()
}

func (this *DecimalFormat) GetPositivePrefix() *String {
	return FromUnits(append([]uint16(nil), this.positivePrefix...))
}

func (this *DecimalFormat) GetPositiveSuffix() *String {
	return FromUnits(append([]uint16(nil), this.positiveSuffix...))
}

func (this *DecimalFormat) GetNegativePrefix() *String {
	return FromUnits(append([]uint16(nil), this.negativePrefix...))
}

func (this *DecimalFormat) GetNegativeSuffix() *String {
	return FromUnits(append([]uint16(nil), this.negativeSuffix...))
}

func (this *DecimalFormat) SetPositivePrefix_String(s *String) {
	this.positivePrefix = append([]uint16(nil), s.units...)
	this.posPrefixPattern = nil
}

func (this *DecimalFormat) SetPositiveSuffix_String(s *String) {
	this.positiveSuffix = append([]uint16(nil), s.units...)
	this.posSuffixPattern = nil
}

func (this *DecimalFormat) SetNegativePrefix_String(s *String) {
	this.negativePrefix = append([]uint16(nil), s.units...)
	this.negPrefixPattern = nil
}

func (this *DecimalFormat) SetNegativeSuffix_String(s *String) {
	this.negativeSuffix = append([]uint16(nil), s.units...)
	this.negSuffixPattern = nil
}

func (this *DecimalFormat) GetMultiplier() int32           { return this.multiplier }
func (this *DecimalFormat) SetMultiplier_int(v int32)      { this.multiplier = v }
func (this *DecimalFormat) GetGroupingSize() int32         { return int32(this.groupingSize) }
func (this *DecimalFormat) SetGroupingUsed_boolean(v bool) { this.groupingUsed = v }

func (this *DecimalFormat) SetGroupingSize_int(v int32) {
	if v < 0 || v > math.MaxInt8 {
		panic(IllegalArgumentException_new_String(Concat(Lit("newValue is out of valid range. value: "), StrOfInt(v))))
	}
	this.groupingSize = int8(v)
}

func (this *DecimalFormat) IsDecimalSeparatorAlwaysShown() bool {
	return this.decimalSeparatorAlwaysShown
}

func (this *DecimalFormat) SetDecimalSeparatorAlwaysShown_boolean(v bool) {
	this.decimalSeparatorAlwaysShown = v
}

func (this *DecimalFormat) GetRoundingMode() *RoundingMode { return this.roundingMode }

func (this *DecimalFormat) SetRoundingMode_RoundingMode(m *RoundingMode) {
	if m == nil {
		panic(NullPointerException_new())
	}
	this.roundingMode = m
}

func (this *DecimalFormat) GetMaximumIntegerDigits() int32  { return this.maximumIntegerDigits }
func (this *DecimalFormat) GetMinimumIntegerDigits() int32  { return this.minimumIntegerDigits }
func (this *DecimalFormat) GetMaximumFractionDigits() int32 { return this.maximumFractionDigits }
func (this *DecimalFormat) GetMinimumFractionDigits() int32 { return this.minimumFractionDigits }

func (this *DecimalFormat) SetMaximumIntegerDigits_int(v int32) {
	this.maximumIntegerDigits = min32(max32(v, 0), dfMaximumIntegerDigits)
	this.NumberFormat.setMaximumIntegerDigits(min32(this.maximumIntegerDigits, dfDoubleIntegerDigits))
	if this.minimumIntegerDigits > this.maximumIntegerDigits {
		this.minimumIntegerDigits = this.maximumIntegerDigits
		this.NumberFormat.setMinimumIntegerDigits(min32(this.minimumIntegerDigits, dfDoubleIntegerDigits))
	}
}

func (this *DecimalFormat) SetMinimumIntegerDigits_int(v int32) {
	this.minimumIntegerDigits = min32(max32(v, 0), dfMaximumIntegerDigits)
	this.NumberFormat.setMinimumIntegerDigits(min32(this.minimumIntegerDigits, dfDoubleIntegerDigits))
	if this.minimumIntegerDigits > this.maximumIntegerDigits {
		this.maximumIntegerDigits = this.minimumIntegerDigits
		this.NumberFormat.setMaximumIntegerDigits(min32(this.maximumIntegerDigits, dfDoubleIntegerDigits))
	}
}

func (this *DecimalFormat) SetMaximumFractionDigits_int(v int32) {
	this.maximumFractionDigits = min32(max32(v, 0), dfMaximumFractionDigits)
	this.NumberFormat.setMaximumFractionDigits(min32(this.maximumFractionDigits, dfDoubleFractionDigits))
	if this.minimumFractionDigits > this.maximumFractionDigits {
		this.minimumFractionDigits = this.maximumFractionDigits
		this.NumberFormat.setMinimumFractionDigits(min32(this.minimumFractionDigits, dfDoubleFractionDigits))
	}
}

func (this *DecimalFormat) SetMinimumFractionDigits_int(v int32) {
	this.minimumFractionDigits = min32(max32(v, 0), dfMaximumFractionDigits)
	this.NumberFormat.setMinimumFractionDigits(min32(this.minimumFractionDigits, dfDoubleFractionDigits))
	if this.minimumFractionDigits > this.maximumFractionDigits {
		this.maximumFractionDigits = this.minimumFractionDigits
		this.NumberFormat.setMaximumFractionDigits(min32(this.maximumFractionDigits, dfDoubleFractionDigits))
	}
}

// ApplyPattern_String implements DecimalFormat.applyPattern.
func (this *DecimalFormat) ApplyPattern_String(pattern *String) {
	if pattern == nil {
		panic(NullPointerException_new())
	}
	this.applyPattern(pattern.units)
}

const (
	dfQuote        = '\''
	dfCurrencySign = 0xa4
)

// applyPattern implements DecimalFormat.applyPattern(String, false).
func (this *DecimalFormat) applyPattern(pattern []uint16) {
	const (
		zeroDigit         = '0'
		groupingSeparator = ','
		decimalSeparator  = '.'
		percent           = '%'
		perMill           = 0x2030
		digit             = '#'
		separator         = ';'
		minus             = '-'
	)
	exponent := []uint16{'E'}
	malformed := func(msg string) {
		panic(IllegalArgumentException_new_String(Concat(FromGo(msg), FromUnits(append([]uint16(nil), pattern...)), Lit("\""))))
	}
	gotNegative := false
	this.decimalSeparatorAlwaysShown = false
	this.isCurrencyFormat = false
	this.useExponentialNotation = false
	start := 0
	regionMatches := func(pos int) bool {
		return pos+len(exponent) <= len(pattern) && unitsEqual(pattern[pos:pos+len(exponent)], exponent)
	}
	for j := 1; j >= 0 && start < len(pattern); j-- {
		inQuote := false
		var prefix, suffix []uint16
		decimalPos := -1
		multiplier := int32(1)
		digitLeftCount, zeroDigitCount, digitRightCount := 0, 0, 0
		groupingCount := -1
		phase := 0
		affix := &prefix
		for pos := start; pos < len(pattern); pos++ {
			ch := pattern[pos]
			switch phase {
			case 0, 2:
				if inQuote {
					if ch == dfQuote {
						if pos+1 < len(pattern) && pattern[pos+1] == dfQuote {
							pos++
							*affix = append(*affix, dfQuote, dfQuote)
						} else {
							inQuote = false
						}
						continue
					}
				} else {
					switch {
					case ch == digit || ch == zeroDigit || ch == groupingSeparator || ch == decimalSeparator:
						phase = 1
						pos--
						continue
					case ch == dfCurrencySign:
						doubled := pos+1 < len(pattern) && pattern[pos+1] == dfCurrencySign
						if doubled {
							pos++
						}
						this.isCurrencyFormat = true
						if doubled {
							*affix = append(*affix, dfQuote, dfCurrencySign, dfCurrencySign)
						} else {
							*affix = append(*affix, dfQuote, dfCurrencySign)
						}
						continue
					case ch == dfQuote:
						if pos+1 < len(pattern) && pattern[pos+1] == dfQuote {
							pos++
							*affix = append(*affix, dfQuote, dfQuote)
						} else {
							inQuote = true
						}
						continue
					case ch == separator:
						if phase == 0 || j == 0 {
							panic(IllegalArgumentException_new_String(Concat(Lit("Unquoted special character '"),
								StrOfChar(ch), Lit("' in pattern \""), FromUnits(append([]uint16(nil), pattern...)), Lit("\""))))
						}
						start = pos + 1
						pos = len(pattern)
						continue
					case ch == percent:
						if multiplier != 1 {
							malformed("Too many percent/per mille characters in pattern \"")
						}
						multiplier = 100
						*affix = append(*affix, dfQuote, '%')
						continue
					case ch == perMill:
						if multiplier != 1 {
							malformed("Too many percent/per mille characters in pattern \"")
						}
						multiplier = 1000
						*affix = append(*affix, dfQuote, perMill)
						continue
					case ch == minus:
						*affix = append(*affix, dfQuote, '-')
						continue
					}
				}
				*affix = append(*affix, ch)
			case 1:
				if j == 0 {
					for pos < len(pattern) {
						c := pattern[pos]
						if c == digit || c == zeroDigit || c == groupingSeparator || c == decimalSeparator {
							pos++
						} else if regionMatches(pos) {
							pos += len(exponent)
						} else {
							pos--
							phase = 2
							affix = &suffix
							break
						}
					}
					continue
				}
				switch {
				case ch == digit:
					if zeroDigitCount > 0 {
						digitRightCount++
					} else {
						digitLeftCount++
					}
					if groupingCount >= 0 && decimalPos < 0 {
						groupingCount = int(int8(groupingCount + 1))
					}
				case ch == zeroDigit:
					if digitRightCount > 0 {
						malformed("Unexpected '0' in pattern \"")
					}
					zeroDigitCount++
					if groupingCount >= 0 && decimalPos < 0 {
						groupingCount = int(int8(groupingCount + 1))
					}
				case ch == groupingSeparator:
					groupingCount = 0
				case ch == decimalSeparator:
					if decimalPos >= 0 {
						malformed("Multiple decimal separators in pattern \"")
					}
					decimalPos = digitLeftCount + zeroDigitCount + digitRightCount
				case regionMatches(pos):
					if this.useExponentialNotation {
						malformed("Multiple exponential symbols in pattern \"")
					}
					this.useExponentialNotation = true
					this.minExponentDigits = 0
					pos += len(exponent)
					for pos < len(pattern) && pattern[pos] == zeroDigit {
						this.minExponentDigits++
						pos++
					}
					if digitLeftCount+zeroDigitCount < 1 || this.minExponentDigits < 1 {
						malformed("Malformed exponential pattern \"")
					}
					phase = 2
					affix = &suffix
					pos--
					continue
				default:
					phase = 2
					affix = &suffix
					pos--
					continue
				}
			}
		}
		if zeroDigitCount == 0 && digitLeftCount > 0 && decimalPos >= 0 {
			n := decimalPos
			if n == 0 {
				n++
			}
			digitRightCount = digitLeftCount - n
			digitLeftCount = n - 1
			zeroDigitCount = 1
		}
		if decimalPos < 0 && digitRightCount > 0 ||
			decimalPos >= 0 && (decimalPos < digitLeftCount || decimalPos > digitLeftCount+zeroDigitCount) ||
			groupingCount == 0 || inQuote {
			malformed("Malformed pattern \"")
		}
		if j == 1 {
			this.posPrefixPattern = prefix
			this.posSuffixPattern = suffix
			this.negPrefixPattern = prefix
			this.negSuffixPattern = suffix
			digitTotalCount := digitLeftCount + zeroDigitCount + digitRightCount
			effectiveDecimalPos := digitTotalCount
			if decimalPos >= 0 {
				effectiveDecimalPos = decimalPos
			}
			this.SetMinimumIntegerDigits_int(int32(effectiveDecimalPos - digitLeftCount))
			if this.useExponentialNotation {
				this.SetMaximumIntegerDigits_int(int32(digitLeftCount) + this.GetMinimumIntegerDigits())
			} else {
				this.SetMaximumIntegerDigits_int(dfMaximumIntegerDigits)
			}
			if decimalPos >= 0 {
				this.SetMaximumFractionDigits_int(int32(digitTotalCount - decimalPos))
				this.SetMinimumFractionDigits_int(int32(digitLeftCount + zeroDigitCount - decimalPos))
			} else {
				this.SetMaximumFractionDigits_int(0)
				this.SetMinimumFractionDigits_int(0)
			}
			this.SetGroupingUsed_boolean(groupingCount > 0)
			if groupingCount > 0 {
				this.groupingSize = int8(groupingCount)
			} else {
				this.groupingSize = 0
			}
			this.multiplier = multiplier
			this.SetDecimalSeparatorAlwaysShown_boolean(decimalPos == 0 || decimalPos == digitTotalCount)
		} else {
			this.negPrefixPattern = prefix
			this.negSuffixPattern = suffix
			gotNegative = true
		}
	}
	if len(pattern) == 0 {
		this.posPrefixPattern = []uint16{}
		this.posSuffixPattern = []uint16{}
		this.SetMinimumIntegerDigits_int(0)
		this.SetMaximumIntegerDigits_int(dfMaximumIntegerDigits)
		this.SetMinimumFractionDigits_int(0)
		this.SetMaximumFractionDigits_int(dfMaximumFractionDigits)
	}
	if !gotNegative || unitsEqual(this.negPrefixPattern, this.posPrefixPattern) && unitsEqual(this.negSuffixPattern, this.posSuffixPattern) {
		this.negSuffixPattern = this.posSuffixPattern
		this.negPrefixPattern = append([]uint16{dfQuote, '-'}, this.posPrefixPattern...)
	}
	this.expandAffixes()
}

func (this *DecimalFormat) expandAffixes() {
	if this.posPrefixPattern != nil {
		this.positivePrefix = this.expandAffix(this.posPrefixPattern)
	}
	if this.posSuffixPattern != nil {
		this.positiveSuffix = this.expandAffix(this.posSuffixPattern)
	}
	if this.negPrefixPattern != nil {
		this.negativePrefix = this.expandAffix(this.negPrefixPattern)
	}
	if this.negSuffixPattern != nil {
		this.negativeSuffix = this.expandAffix(this.negSuffixPattern)
	}
}

func (this *DecimalFormat) expandAffix(pattern []uint16) []uint16 {
	out := []uint16{}
	for i := 0; i < len(pattern); {
		c := pattern[i]
		i++
		if c == dfQuote {
			c = pattern[i]
			i++
			switch c {
			case dfCurrencySign:
				panic(Unsupported("DecimalFormat currency patterns"))
			case '%':
				out = append(out, this.symbols.percentText.units...)
				continue
			case 0x2030:
				out = append(out, this.symbols.perMillText.units...)
				continue
			case '-':
				out = append(out, this.symbols.minusText.units...)
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

// formatDouble implements DecimalFormat.format(double, StringBuffer,
// FieldDelegate), appending to out.
func (this *DecimalFormat) formatDouble(number float64, out []uint16) []uint16 {
	if number != number || math.IsInf(number, 0) && this.multiplier == 0 {
		return append(out, this.symbols.nan.units...)
	}
	isNegative := (number < 0 || number == 0 && math.Signbit(number)) != (this.multiplier < 0)
	if this.multiplier != 1 {
		number = float64(number * float64(this.multiplier))
	}
	if math.IsInf(number, 0) {
		if isNegative {
			out = append(out, this.negativePrefix...)
		} else {
			out = append(out, this.positivePrefix...)
		}
		out = append(out, this.symbols.infinity.units...)
		if isNegative {
			return append(out, this.negativeSuffix...)
		}
		return append(out, this.positiveSuffix...)
	}
	if isNegative {
		number = -number
	}
	maxIntDigits := this.NumberFormat.maxIntDigits
	minIntDigits := this.NumberFormat.minIntDigits
	maxFraDigits := this.NumberFormat.maxFracDigits
	minFraDigits := this.NumberFormat.minFracDigits
	var dl digitList
	dl.roundingMode = this.roundingMode.ordinal
	if this.useExponentialNotation {
		dl.setDouble(isNegative, number, int(maxIntDigits+maxFraDigits), false)
	} else {
		dl.setDouble(isNegative, number, int(maxFraDigits), true)
	}
	return this.subformat(out, &dl, isNegative, false, maxIntDigits, minIntDigits, maxFraDigits, minFraDigits)
}

// formatLong implements DecimalFormat.format(long, StringBuffer,
// FieldDelegate), appending to out.
func (this *DecimalFormat) formatLong(number int64, out []uint16) []uint16 {
	isNegative := number < 0
	if isNegative {
		number = -number
	}
	useBig := false
	if number < 0 {
		useBig = this.multiplier != 0
	} else if this.multiplier != 1 && this.multiplier != 0 {
		cutoff := math.MaxInt64 / int64(this.multiplier)
		if cutoff < 0 {
			cutoff = -cutoff
		}
		useBig = number > cutoff
	}
	if useBig {
		if isNegative {
			number = -number
		}
		return this.formatBigInteger(big.NewInt(number), out)
	}
	number *= int64(this.multiplier)
	if number == 0 {
		isNegative = false
	} else if this.multiplier < 0 {
		number = -number
		isNegative = !isNegative
	}
	maxIntDigits := this.NumberFormat.maxIntDigits
	minIntDigits := this.NumberFormat.minIntDigits
	maxFraDigits := this.NumberFormat.maxFracDigits
	minFraDigits := this.NumberFormat.minFracDigits
	var dl digitList
	dl.roundingMode = this.roundingMode.ordinal
	max := 0
	if this.useExponentialNotation {
		max = int(maxIntDigits + maxFraDigits)
	}
	dl.setLong(isNegative, number, max)
	return this.subformat(out, &dl, isNegative, true, maxIntDigits, minIntDigits, maxFraDigits, minFraDigits)
}

// formatBigInteger implements DecimalFormat.format(BigInteger, ..., true)
// as used for long values whose product with the multiplier overflows.
func (this *DecimalFormat) formatBigInteger(number *big.Int, out []uint16) []uint16 {
	n := new(big.Int).Set(number)
	if this.multiplier != 1 {
		n.Mul(n, big.NewInt(int64(this.multiplier)))
	}
	isNegative := n.Sign() < 0
	if isNegative {
		n.Neg(n)
	}
	maxIntDigits := this.NumberFormat.maxIntDigits
	minIntDigits := this.NumberFormat.minIntDigits
	maxFraDigits := this.NumberFormat.maxFracDigits
	minFraDigits := this.NumberFormat.minFracDigits
	var dl digitList
	dl.roundingMode = this.roundingMode.ordinal
	max := 0
	if this.useExponentialNotation {
		max = int(maxIntDigits + maxFraDigits)
	}
	dl.setBigInteger(isNegative, n, max)
	return this.subformat(out, &dl, isNegative, true, maxIntDigits, minIntDigits, maxFraDigits, minFraDigits)
}

// formatBigDecimal implements DecimalFormat.format(BigDecimal, StringBuffer,
// FieldDelegate).
func (this *DecimalFormat) formatBigDecimal(number *BigDecimal, out []uint16) []uint16 {
	if this.multiplier != 1 {
		number = number.Multiply_BigDecimal(BigDecimal_valueOf_long(int64(this.multiplier)))
	}
	isNegative := number.Signum() == -1
	if isNegative {
		number = number.Negate()
	}
	maxIntDigits := this.GetMaximumIntegerDigits()
	minIntDigits := this.GetMinimumIntegerDigits()
	maxFraDigits := this.GetMaximumFractionDigits()
	minFraDigits := this.GetMinimumFractionDigits()
	maximumDigits := maxIntDigits + maxFraDigits
	var dl digitList
	dl.roundingMode = this.roundingMode.ordinal
	if this.useExponentialNotation {
		if maximumDigits < 0 {
			maximumDigits = math.MaxInt32
		}
		dl.setString(isNegative, number.ToString().Go(), false, true, int(maximumDigits), false)
	} else {
		dl.setString(isNegative, number.ToString().Go(), false, true, int(maxFraDigits), true)
	}
	return this.subformat(out, &dl, isNegative, false, maxIntDigits, minIntDigits, maxFraDigits, minFraDigits)
}

func (this *DecimalFormat) subformat(out []uint16, dl *digitList, isNegative, isInteger bool,
	maxIntDigits, minIntDigits, maxFraDigits, minFraDigits int32,
) []uint16 {
	if isNegative {
		out = append(out, this.negativePrefix...)
	} else {
		out = append(out, this.positivePrefix...)
	}
	out = this.subformatNumber(out, dl, isInteger, int(maxIntDigits), int(minIntDigits), int(maxFraDigits), int(minFraDigits))
	if isNegative {
		return append(out, this.negativeSuffix...)
	}
	return append(out, this.positiveSuffix...)
}

func (this *DecimalFormat) subformatNumber(out []uint16, dl *digitList, isInteger bool,
	maxIntDigits, minIntDigits, maxFraDigits, minFraDigits int,
) []uint16 {
	grouping := this.symbols.grouping
	decimal := this.symbols.decimal
	if this.isCurrencyFormat {
		grouping = this.symbols.monetaryGrouping
		decimal = this.symbols.monetaryDecimal
	}
	zero := this.symbols.zero
	zeroDelta := zero - '0'
	if dl.isZero() {
		dl.decimalAt = 0
	}
	if this.useExponentialNotation {
		exponent := dl.decimalAt
		repeat := maxIntDigits
		minimumIntegerDigits := minIntDigits
		if repeat > 1 && repeat > minIntDigits {
			if exponent >= 1 {
				exponent = ((exponent - 1) / repeat) * repeat
			} else {
				exponent = ((exponent - repeat) / repeat) * repeat
			}
			minimumIntegerDigits = 1
		} else {
			exponent -= minimumIntegerDigits
		}
		minimumDigits := minIntDigits + minFraDigits
		if minimumDigits < 0 {
			minimumDigits = math.MaxInt32
		}
		integerDigits := minimumIntegerDigits
		if !dl.isZero() {
			integerDigits = dl.decimalAt - exponent
		}
		if minimumDigits < integerDigits {
			minimumDigits = integerDigits
		}
		totalDigits := dl.count
		if minimumDigits > totalDigits {
			totalDigits = minimumDigits
		}
		for i := 0; i < totalDigits; i++ {
			if i == integerDigits {
				out = append(out, decimal)
			}
			if i < dl.count {
				out = append(out, uint16(dl.digits[i])+zeroDelta)
			} else {
				out = append(out, zero)
			}
		}
		if this.decimalSeparatorAlwaysShown && totalDigits == integerDigits {
			out = append(out, decimal)
		}
		out = append(out, this.symbols.exponent.units...)
		if dl.isZero() {
			exponent = 0
		}
		negativeExponent := exponent < 0
		if negativeExponent {
			exponent = -exponent
			out = append(out, this.symbols.minusText.units...)
		}
		var el digitList
		el.roundingMode = dl.roundingMode
		el.setLong(negativeExponent, int64(exponent), 0)
		for i := el.decimalAt; i < int(this.minExponentDigits); i++ {
			out = append(out, zero)
		}
		for i := 0; i < el.decimalAt; i++ {
			if i < el.count {
				out = append(out, uint16(el.digits[i])+zeroDelta)
			} else {
				out = append(out, zero)
			}
		}
		return out
	}
	count := minIntDigits
	digitIndex := 0
	if dl.decimalAt > 0 && count < dl.decimalAt {
		count = dl.decimalAt
	}
	if count > maxIntDigits {
		count = maxIntDigits
		digitIndex = dl.decimalAt - count
	}
	sizeBeforeIntegerPart := len(out)
	for i := count - 1; i >= 0; i-- {
		if i < dl.decimalAt && digitIndex < dl.count {
			out = append(out, uint16(dl.digits[digitIndex])+zeroDelta)
			digitIndex++
		} else {
			out = append(out, zero)
		}
		if this.groupingUsed && i > 0 && this.groupingSize != 0 && i%int(this.groupingSize) == 0 {
			out = append(out, grouping)
		}
	}
	fractionPresent := minFraDigits > 0 || !isInteger && digitIndex < dl.count
	if !fractionPresent && len(out) == sizeBeforeIntegerPart {
		out = append(out, zero)
	}
	if this.decimalSeparatorAlwaysShown || fractionPresent {
		out = append(out, decimal)
	}
	for i := 0; i < maxFraDigits; i++ {
		if i >= minFraDigits && (isInteger || digitIndex >= dl.count) {
			break
		}
		if -1-i > dl.decimalAt-1 {
			out = append(out, zero)
			continue
		}
		if !isInteger && digitIndex < dl.count {
			out = append(out, uint16(dl.digits[digitIndex])+zeroDelta)
			digitIndex++
		} else {
			out = append(out, zero)
		}
	}
	return out
}

// fastFormat implements DecimalFormat.fastFormat, returning nil when the
// fast path does not apply.
func (this *DecimalFormat) fastFormat(d float64) []uint16 {
	if this.roundingMode != RoundingMode_HALF_EVEN || !this.groupingUsed || this.groupingSize != 3 ||
		this.multiplier != 1 || this.decimalSeparatorAlwaysShown || this.useExponentialNotation {
		return nil
	}
	if this.minimumIntegerDigits != 1 || this.maximumIntegerDigits < 10 {
		return nil
	}
	if this.isCurrencyFormat {
		if this.minimumFractionDigits != 2 || this.maximumFractionDigits != 2 {
			return nil
		}
	} else if this.minimumFractionDigits != 0 || this.maximumFractionDigits != 3 {
		return nil
	}
	if math.IsInf(d, 0) || d != d {
		return nil
	}
	negative := false
	if d < 0 {
		negative = true
		d = -d
	} else if d == 0 {
		negative = math.Signbit(d)
		d = 0
	}
	if d > math.MaxInt32 {
		return nil
	}
	fractionalMaxIntBound := int32(999)
	fractionalScaleFactor := 1000.0
	groupingChar := this.symbols.grouping
	decimalChar := this.symbols.decimal
	if this.isCurrencyFormat {
		fractionalMaxIntBound = 99
		fractionalScaleFactor = 100.0
		groupingChar = this.symbols.monetaryGrouping
		decimalChar = this.symbols.monetaryDecimal
	}
	integralPartAsInt := int32(d)
	exactFractionalPart := d - float64(integralPartAsInt)
	scaledFractional := float64(exactFractionalPart * fractionalScaleFactor)
	fractionalPartAsInt := int32(scaledFractional)
	scaledFractional = scaledFractional - float64(fractionalPartAsInt)
	if scaledFractional >= 0.5 {
		roundItUp := true
		if scaledFractional == 0.5 {
			roundItUp = this.exactRoundUp(exactFractionalPart, fractionalPartAsInt)
		}
		if roundItUp {
			if fractionalPartAsInt < fractionalMaxIntBound {
				fractionalPartAsInt++
			} else {
				fractionalPartAsInt = 0
				integralPartAsInt++
			}
		}
	}
	zeroDelta := this.symbols.zero - '0'
	digit := func(v int32) uint16 { return uint16('0'+v) + zeroDelta }
	// Fractional digits.
	var frac []uint16
	if this.isCurrencyFormat {
		frac = []uint16{digit(fractionalPartAsInt / 10 % 10), digit(fractionalPartAsInt % 10)}
	} else if fractionalPartAsInt != 0 {
		h, t, o := fractionalPartAsInt/100, fractionalPartAsInt/10%10, fractionalPartAsInt%10
		frac = []uint16{digit(h)}
		if o != 0 {
			frac = append(frac, digit(t), digit(o))
		} else if t != 0 {
			frac = append(frac, digit(t))
		}
	}
	// Integral digits, with grouping.
	var rev []uint16
	n := integralPartAsInt
	for n > 999 {
		q := n / 1000
		r := n - q*1000
		n = q
		rev = append(rev, digit(r%10), digit(r/10%10), digit(r/100), groupingChar)
	}
	rev = append(rev, digit(n%10))
	if n > 9 {
		rev = append(rev, digit(n/10%10))
		if n > 99 {
			rev = append(rev, digit(n/100))
		}
	}
	var out []uint16
	if negative {
		out = append(out, this.negativePrefix...)
	} else {
		out = append(out, this.positivePrefix...)
	}
	for i := len(rev) - 1; i >= 0; i-- {
		out = append(out, rev[i])
	}
	if len(frac) > 0 {
		out = append(out, decimalChar)
		out = append(out, frac...)
	}
	if negative {
		return append(out, this.negativeSuffix...)
	}
	return append(out, this.positiveSuffix...)
}

// exactRoundUp implements DecimalFormat.exactRoundUp, which decides a tie
// in the fast path from the exact value of the fractional part.
func (this *DecimalFormat) exactRoundUp(fractionalPart float64, scaledFractionalPartAsInt int32) bool {
	var approxMax, approxMedium, approxMin float64
	if this.isCurrencyFormat {
		approxMax = float64(fractionalPart * 128.00)
		approxMedium = -float64(fractionalPart * 32.00)
		approxMin = float64(fractionalPart * 4.00)
	} else {
		approxMax = float64(fractionalPart * 1024.00)
		approxMedium = -float64(fractionalPart * 16.00)
		approxMin = -float64(fractionalPart * 8.00)
	}
	fastTwoSumApproximation := approxMedium + approxMin
	bVirtual := fastTwoSumApproximation - approxMedium
	fastTwoSumRoundOff := approxMin - bVirtual
	approxS1 := fastTwoSumApproximation
	roundoffS1 := fastTwoSumRoundOff
	fastTwoSumApproximation = approxMax + approxS1
	bVirtual = fastTwoSumApproximation - approxMax
	fastTwoSumRoundOff = approxS1 - bVirtual
	roundoff1000 := fastTwoSumRoundOff
	approx1000 := fastTwoSumApproximation
	roundoffTotal := roundoffS1 + roundoff1000
	fastTwoSumApproximation = approx1000 + roundoffTotal
	bVirtual = fastTwoSumApproximation - approx1000
	scaledFractionalRoundoff := roundoffTotal - bVirtual
	switch {
	case scaledFractionalRoundoff > 0:
		return true
	case scaledFractionalRoundoff < 0:
		return false
	}
	return scaledFractionalPartAsInt&1 != 0
}

// ---------------------------------------------------------------------------
// DigitList

// digitList is java.text.DigitList: the digits of a non-negative value as
// 0.digits[:count] * 10^decimalAt.
type digitList struct {
	decimalAt    int
	count        int
	digits       []byte
	isNegative   bool
	roundingMode int32 // ordinal of a RoundingMode
}

const (
	rmUp = iota
	rmDown
	rmCeiling
	rmFloor
	rmHalfUp
	rmHalfDown
	rmHalfEven
	rmUnnecessary
)

func (d *digitList) isZero() bool {
	for i := 0; i < d.count; i++ {
		if d.digits[i] != '0' {
			return false
		}
	}
	return true
}

// setDouble implements DigitList.set(boolean, double, int, boolean).
func (d *digitList) setDouble(isNegative bool, source float64, maximumDigits int, fixedPoint bool) {
	fd := fdConvert(source)
	d.setString(isNegative, fd.javaFormatString(), fd.roundedUp, fd.exact, maximumDigits, fixedPoint)
}

// setString implements DigitList.set(boolean, String, boolean, boolean, int,
// boolean), which parses digits in the format of fdDecimal.javaFormatString.
func (d *digitList) setString(isNegative bool, s string, roundedUp, valueExactAsDecimal bool, maximumDigits int, fixedPoint bool) {
	d.isNegative = isNegative
	d.digits = d.digits[:0]
	d.decimalAt = -1
	d.count = 0
	exponent := 0
	leadingZerosAfterDecimal := 0
	nonZeroDigitSeen := false
	for i := 0; i < len(s); {
		c := s[i]
		i++
		if c == '.' {
			d.decimalAt = d.count
		} else if c == 'e' || c == 'E' {
			exponent = digitListParseInt(s[i:])
			break
		} else {
			if !nonZeroDigitSeen {
				nonZeroDigitSeen = c != '0'
				if !nonZeroDigitSeen && d.decimalAt != -1 {
					leadingZerosAfterDecimal++
				}
			}
			if nonZeroDigitSeen {
				d.digits = append(d.digits, c)
				d.count++
			}
		}
	}
	if d.decimalAt == -1 {
		d.decimalAt = d.count
	}
	if nonZeroDigitSeen {
		d.decimalAt += exponent - leadingZerosAfterDecimal
	}
	if fixedPoint {
		if -d.decimalAt > maximumDigits {
			d.count = 0
			return
		} else if -d.decimalAt == maximumDigits {
			if d.shouldRoundUp(0, roundedUp, valueExactAsDecimal) {
				d.count = 1
				d.decimalAt++
				d.setDigit(0, '1')
			} else {
				d.count = 0
			}
			return
		}
	}
	for d.count > 1 && d.digits[d.count-1] == '0' {
		d.count--
	}
	if fixedPoint {
		d.round(maximumDigits+d.decimalAt, roundedUp, valueExactAsDecimal)
	} else {
		d.round(maximumDigits, roundedUp, valueExactAsDecimal)
	}
}

func (d *digitList) setDigit(i int, c byte) {
	for len(d.digits) <= i {
		d.digits = append(d.digits, '0')
	}
	d.digits[i] = c
}

// digitListParseInt implements DigitList.parseInt.
func digitListParseInt(s string) int {
	positive := true
	i := 0
	if i < len(s) {
		switch s[i] {
		case '-':
			positive = false
			i++
		case '+':
			i++
		}
	}
	value := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		value = value*10 + int(c-'0')
	}
	if positive {
		return value
	}
	return -value
}

func (d *digitList) round(maximumDigits int, alreadyRounded, valueExactAsDecimal bool) {
	if maximumDigits >= 0 && maximumDigits < d.count {
		if d.shouldRoundUp(maximumDigits, alreadyRounded, valueExactAsDecimal) {
			for {
				maximumDigits--
				if maximumDigits < 0 {
					d.digits[0] = '1'
					d.decimalAt++
					maximumDigits = 0
					break
				}
				d.digits[maximumDigits]++
				if d.digits[maximumDigits] <= '9' {
					break
				}
			}
			maximumDigits++
		}
		d.count = maximumDigits
		for d.count > 1 && d.digits[d.count-1] == '0' {
			d.count--
		}
	}
}

func (d *digitList) shouldRoundUp(maximumDigits int, alreadyRounded, valueExactAsDecimal bool) bool {
	if maximumDigits >= d.count {
		return false
	}
	switch d.roundingMode {
	case rmUp:
		for i := maximumDigits; i < d.count; i++ {
			if d.digits[i] != '0' {
				return true
			}
		}
	case rmDown:
	case rmCeiling:
		for i := maximumDigits; i < d.count; i++ {
			if d.digits[i] != '0' {
				return !d.isNegative
			}
		}
	case rmFloor:
		for i := maximumDigits; i < d.count; i++ {
			if d.digits[i] != '0' {
				return d.isNegative
			}
		}
	case rmHalfUp, rmHalfDown:
		if d.digits[maximumDigits] > '5' {
			return true
		} else if d.digits[maximumDigits] == '5' {
			if maximumDigits != d.count-1 {
				return true
			}
			if valueExactAsDecimal {
				return d.roundingMode == rmHalfUp
			}
			return !alreadyRounded
		}
	case rmHalfEven:
		if d.digits[maximumDigits] > '5' {
			return true
		} else if d.digits[maximumDigits] == '5' {
			if maximumDigits == d.count-1 {
				if alreadyRounded {
					return false
				}
				if !valueExactAsDecimal {
					return true
				}
				return maximumDigits > 0 && d.digits[maximumDigits-1]%2 != 0
			}
			for i := maximumDigits + 1; i < d.count; i++ {
				if d.digits[i] != '0' {
					return true
				}
			}
		}
	case rmUnnecessary:
		for i := maximumDigits; i < d.count; i++ {
			if d.digits[i] != '0' {
				panic(ArithmeticException_new_String(Lit(
					"Rounding needed with the rounding mode being set to RoundingMode.UNNECESSARY")))
			}
		}
	}
	return false
}

// setLong implements DigitList.set(boolean, long, int).
func (d *digitList) setLong(isNegative bool, source int64, maximumDigits int) {
	d.isNegative = isNegative
	if source <= 0 {
		if source == math.MinInt64 {
			d.digits = []byte("9223372036854775808")
			d.count = len(d.digits)
			d.decimalAt = d.count
		} else {
			d.digits = d.digits[:0]
			d.count = 0
			d.decimalAt = 0
		}
	} else {
		s := strconv.FormatInt(source, 10)
		d.decimalAt = len(s)
		n := len(s)
		for s[n-1] == '0' {
			n--
		}
		d.digits = []byte(s[:n])
		d.count = n
	}
	if maximumDigits > 0 {
		d.round(maximumDigits, false, true)
	}
}

// setBigInteger implements DigitList.set(boolean, BigInteger, int).
func (d *digitList) setBigInteger(isNegative bool, source *big.Int, maximumDigits int) {
	d.isNegative = isNegative
	s := source.String()
	d.decimalAt = len(s)
	n := len(s)
	for n > 1 && s[n-1] == '0' {
		n--
	}
	if s == "0" {
		n = 0
	}
	d.digits = []byte(s[:n])
	d.count = n
	if maximumDigits > 0 {
		d.round(maximumDigits, false, true)
	}
}
