package java

import (
	"strings"
	"sync"
)

// Locale is java.util.Locale. ext holds the extensions and private use
// subtags in the canonical form of LocaleExtensions.getID ("u-ca-japanese",
// "x-foo"); the Unicode extension's attributes and keywords are kept in the
// order given.
type Locale struct {
	language string
	script   string
	country  string
	variant  string
	ext      string
}

func (this *Locale) JClass_Locale() *Locale { return this }
func (this *Locale) JClassName() string     { return "java.util.Locale" }

// Locale_isInstance implements instanceof Locale.
func Locale_isInstance(o any) bool { _, ok := o.(*Locale); return ok }

// Locale_fromAny casts o to Locale.
func Locale_fromAny(o any) *Locale { return AnyTo[*Locale](o) }

// newLocale normalizes the parts of a locale as java.util.Locale does.
func newLocale(language, script, country, variant string) *Locale {
	language = asciiLower(language)
	switch language {
	case "iw":
		language = "he"
	case "in":
		language = "id"
	case "ji":
		language = "yi"
	}
	return &Locale{language: language, script: script, country: asciiUpper(country), variant: variant}
}

func asciiLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func asciiUpper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
	return string(b)
}

// Locale_ROOT is Locale.ROOT.
var Locale_ROOT = newLocale("", "", "", "")

// Locale_ENGLISH is Locale.ENGLISH.
var Locale_ENGLISH = newLocale("en", "", "", "")

// Locale_FRENCH is Locale.FRENCH.
var Locale_FRENCH = newLocale("fr", "", "", "")

// Locale_GERMAN is Locale.GERMAN.
var Locale_GERMAN = newLocale("de", "", "", "")

// Locale_ITALIAN is Locale.ITALIAN.
var Locale_ITALIAN = newLocale("it", "", "", "")

// Locale_JAPANESE is Locale.JAPANESE.
var Locale_JAPANESE = newLocale("ja", "", "", "")

// Locale_KOREAN is Locale.KOREAN.
var Locale_KOREAN = newLocale("ko", "", "", "")

// Locale_CHINESE is Locale.CHINESE.
var Locale_CHINESE = newLocale("zh", "", "", "")

// Locale_SIMPLIFIED_CHINESE is Locale.SIMPLIFIED_CHINESE.
var Locale_SIMPLIFIED_CHINESE = newLocale("zh", "", "CN", "")

// Locale_TRADITIONAL_CHINESE is Locale.TRADITIONAL_CHINESE.
var Locale_TRADITIONAL_CHINESE = newLocale("zh", "", "TW", "")

// Locale_FRANCE is Locale.FRANCE.
var Locale_FRANCE = newLocale("fr", "", "FR", "")

// Locale_GERMANY is Locale.GERMANY.
var Locale_GERMANY = newLocale("de", "", "DE", "")

// Locale_ITALY is Locale.ITALY.
var Locale_ITALY = newLocale("it", "", "IT", "")

// Locale_JAPAN is Locale.JAPAN.
var Locale_JAPAN = newLocale("ja", "", "JP", "")

// Locale_KOREA is Locale.KOREA.
var Locale_KOREA = newLocale("ko", "", "KR", "")

// Locale_CHINA is Locale.CHINA.
var Locale_CHINA = Locale_SIMPLIFIED_CHINESE

// Locale_PRC is Locale.PRC.
var Locale_PRC = Locale_SIMPLIFIED_CHINESE

// Locale_TAIWAN is Locale.TAIWAN.
var Locale_TAIWAN = Locale_TRADITIONAL_CHINESE

// Locale_UK is Locale.UK.
var Locale_UK = newLocale("en", "", "GB", "")

// Locale_US is Locale.US.
var Locale_US = newLocale("en", "", "US", "")

// Locale_CANADA is Locale.CANADA.
var Locale_CANADA = newLocale("en", "", "CA", "")

// Locale_CANADA_FRENCH is Locale.CANADA_FRENCH.
var Locale_CANADA_FRENCH = newLocale("fr", "", "CA", "")

// localeDefault is the JVM's default locale, which Locale.setDefault
// changes for the whole process. It starts as en-US, the locale the oracle
// JVM is expected to run in.
var localeDefault = struct {
	sync.Mutex
	l *Locale
}{l: Locale_US}

// Locale_getDefault implements Locale.getDefault().
func Locale_getDefault() *Locale {
	localeDefault.Lock()
	defer localeDefault.Unlock()
	return localeDefault.l
}

// Locale_setDefault_Locale implements Locale.setDefault(Locale).
func Locale_setDefault_Locale(l *Locale) {
	if l == nil {
		panic(NullPointerException_new_String(Lit("Can't set default locale to NULL")))
	}
	localeDefault.Lock()
	defer localeDefault.Unlock()
	localeDefault.l = l
}

// Locale_new_String implements new Locale(String).
func Locale_new_String(language *String) *Locale {
	return Locale_new_String_String_String(language, Lit(""), Lit(""))
}

// Locale_new_String_String implements new Locale(String, String).
func Locale_new_String_String(language, country *String) *Locale {
	return Locale_new_String_String_String(language, country, Lit(""))
}

// Locale_new_String_String_String implements new Locale(String, String,
// String).
func Locale_new_String_String_String(language, country, variant *String) *Locale {
	if language == nil || country == nil || variant == nil {
		panic(NullPointerException_new())
	}
	return newLocale(language.Go(), "", country.Go(), variant.Go())
}

// Locale_forLanguageTag implements Locale.forLanguageTag: the language,
// extended language, script, region, variant, extension and private use
// subtags of a BCP 47 tag, stopping at the first subtag that does not fit.
func Locale_forLanguageTag(tag *String) *Locale {
	parts := strings.Split(tag.Go(), "-")
	i := 0
	cur := func() string {
		if i < len(parts) {
			return parts[i]
		}
		return ""
	}
	var language, script, region string
	var variants []string
	if isAlphaLen(cur(), 2, 8) {
		language = asciiLower(cur())
		i++
		for n := 0; n < 3 && isAlphaLen(cur(), 3, 3); n++ {
			if n == 0 {
				language = asciiLower(cur())
			}
			i++
		}
		if isAlphaLen(cur(), 4, 4) {
			script = asciiUpper(cur()[:1]) + asciiLower(cur()[1:])
			i++
		}
		if isAlphaLen(cur(), 2, 2) || len(cur()) == 3 && isDigitString(cur()) {
			region = asciiUpper(cur())
			i++
		}
		for isVariantSubtag(cur()) {
			variants = append(variants, cur())
			i++
		}
	}
	type extension struct{ key, value string }
	var exts []extension
	for i < len(parts) && len(cur()) == 1 && isAlnumLen(cur(), 1, 1) && asciiLower(cur()) != "x" {
		key := asciiLower(cur())
		j := i + 1
		var vals []string
		for j < len(parts) && isAlnumLen(parts[j], 2, 8) {
			vals = append(vals, asciiLower(parts[j]))
			j++
		}
		if len(vals) == 0 {
			break
		}
		exts = append(exts, extension{key, strings.Join(vals, "-")})
		i = j
	}
	privateUse := ""
	if i < len(parts) && asciiLower(cur()) == "x" {
		j := i + 1
		var vals []string
		for j < len(parts) && isAlnumLen(parts[j], 1, 8) {
			vals = append(vals, asciiLower(parts[j]))
			j++
		}
		if len(vals) > 0 {
			privateUse = "x-" + strings.Join(vals, "-")
		}
	}
	if language == "und" {
		language = ""
	}
	l := newLocale(language, script, region, strings.Join(variants, "_"))
	sortExtensions := func() {
		for a := 1; a < len(exts); a++ {
			for b := a; b > 0 && exts[b].key < exts[b-1].key; b-- {
				exts[b], exts[b-1] = exts[b-1], exts[b]
			}
		}
	}
	sortExtensions()
	var ids []string
	for _, e := range exts {
		ids = append(ids, e.key+"-"+e.value)
	}
	if privateUse != "" {
		ids = append(ids, privateUse)
	}
	l.ext = strings.Join(ids, "-")
	return l
}

func isAlphaLen(s string, min, max int) bool {
	if len(s) < min || len(s) > max {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i] | 0x20
		if c < 'a' || c > 'z' {
			return false
		}
	}
	return true
}

func isDigitString(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isAlnumLen(s string, min, max int) bool {
	if len(s) < min || len(s) > max {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i] | 0x20
		if (c < 'a' || c > 'z') && (s[i] < '0' || s[i] > '9') {
			return false
		}
	}
	return true
}

// isVariantSubtag implements LanguageTag.isVariant.
func isVariantSubtag(s string) bool {
	if len(s) >= 5 && len(s) <= 8 {
		return isAlnumLen(s, 5, 8)
	}
	return len(s) == 4 && s[0] >= '0' && s[0] <= '9' && isAlnumLen(s, 4, 4)
}

// GetLanguage implements Locale.getLanguage.
func (this *Locale) GetLanguage() *String { return FromGo(this.language) }

// GetScript implements Locale.getScript.
func (this *Locale) GetScript() *String { return FromGo(this.script) }

// GetCountry implements Locale.getCountry.
func (this *Locale) GetCountry() *String { return FromGo(this.country) }

// GetVariant implements Locale.getVariant.
func (this *Locale) GetVariant() *String { return FromGo(this.variant) }

// HasExtensions implements Locale.hasExtensions.
func (this *Locale) HasExtensions() bool { return this.ext != "" }

// StripExtensions implements Locale.stripExtensions.
func (this *Locale) StripExtensions() *Locale {
	if this.ext == "" {
		return this
	}
	l := *this
	l.ext = ""
	return &l
}

// ToLanguageTag implements Locale.toLanguageTag.
func (this *Locale) ToLanguageTag() *String { return FromGo(this.languageTag()) }

// languageTag implements Locale.toLanguageTag, via LanguageTag.parseLocale.
func (this *Locale) languageTag() string {
	var parts []string
	language := ""
	if isAlphaLen(this.language, 2, 8) {
		language = this.language
	}
	hasSubtag := false
	if isAlphaLen(this.script, 4, 4) {
		parts = append(parts, this.script)
		hasSubtag = true
	}
	if isAlphaLen(this.country, 2, 2) || len(this.country) == 3 && isDigitString(this.country) {
		parts = append(parts, this.country)
		hasSubtag = true
	}
	variant := this.variant
	if language == "no" && this.country == "NO" && variant == "NY" {
		language = "nn"
		variant = ""
	}
	var privuseVar []string
	if variant != "" {
		vs := strings.FieldsFunc(variant, func(r rune) bool { return r == '_' || r == '-' })
		k := 0
		for ; k < len(vs) && isVariantSubtag(vs[k]); k++ {
			parts = append(parts, vs[k])
			hasSubtag = true
		}
		for ; k < len(vs) && isAlnumLen(vs[k], 1, 8); k++ {
			privuseVar = append(privuseVar, vs[k])
		}
	}
	privateUse := ""
	for _, e := range strings.Split(this.ext, "-x-") {
		if strings.HasPrefix(e, "x-") {
			privateUse = e[2:]
		}
	}
	extensions := this.ext
	if i := strings.Index(extensions, "x-"); i == 0 {
		extensions = ""
	} else if i := strings.Index(extensions, "-x-"); i >= 0 {
		privateUse = extensions[i+3:]
		extensions = extensions[:i]
	}
	if extensions != "" {
		parts = append(parts, extensions)
		hasSubtag = true
	}
	if len(privuseVar) > 0 {
		pv := "lvariant-" + strings.Join(privuseVar, "-")
		if privateUse == "" {
			privateUse = pv
		} else {
			privateUse += "-" + pv
		}
	}
	if language == "" && (hasSubtag || privateUse == "") {
		language = "und"
	}
	if language != "" {
		parts = append([]string{language}, parts...)
	}
	if privateUse != "" {
		parts = append(parts, "x", privateUse)
	}
	return strings.Join(parts, "-")
}

// ToString implements Locale.toString.
func (this *Locale) ToString() *String {
	l, s, r, v := this.language != "", this.script != "", this.country != "", this.variant != ""
	e := this.ext != ""
	out := this.language
	if r || l && (v || s || e) {
		out += "_" + this.country
	}
	if v && (l || r) {
		out += "_" + this.variant
	}
	if s && (l || r) {
		out += "_#" + this.script
	}
	if e && (l || r) {
		out += "_"
		if !s {
			out += "#"
		}
		out += this.ext
	}
	return FromGo(out)
}

// Equals_Object implements Locale.equals.
func (this *Locale) Equals_Object(o any) bool {
	l, ok := o.(*Locale)
	return ok && l != nil && *l == *this
}

// HashCode implements Locale.hashCode.
func (this *Locale) HashCode() int32 {
	h := FromGo(this.language).HashCode()
	h = 31*h + FromGo(this.script).HashCode()
	h = 31*h + FromGo(this.country).HashCode()
	h = 31*h + FromGo(this.variant).HashCode()
	if this.ext != "" {
		h ^= FromGo(this.ext).HashCode()
	}
	return h
}

// localeNumberData is the number formatting data of a locale; see
// TextLocaleDataGen.
type localeNumberData struct {
	zero, grouping, decimal, perMill, percent, digit uint16
	patternSeparator, minus                          uint16
	monetaryDecimal, monetaryGrouping                uint16
	exponent, infinity, nan                          string
	percentText, perMillText, minusText              string
	numberPattern, percentPattern                    string
}

// numberDataFor returns the number formatting data the JDK uses for l,
// falling back from the full language tag towards the root locale.
func numberDataFor(l *Locale) *localeNumberData {
	tag := l.languageTag()
	for {
		if i, ok := localeNumberDataIndex[tag]; ok {
			return &localeNumberDataSets[i]
		}
		j := strings.LastIndexByte(tag, '-')
		if j < 0 {
			return &localeNumberDataSets[localeNumberDataIndex["und"]]
		}
		tag = tag[:j]
	}
}
