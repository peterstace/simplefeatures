package java

// StreamTokenizer is java.io.StreamTokenizer, ported from JDK 21 including
// its number parsing, which accumulates digits in a double and divides by a
// power of ten rather than rounding correctly.
type StreamTokenizer struct {
	jhash               int32
	reader              readerImpl
	input               inputStreamImpl
	buf                 []uint16
	peekc               int32
	pushedBack          bool
	forceLower          bool
	lineno              int32
	eolIsSignificantP   bool
	slashSlashCommentsP bool
	slashStarCommentsP  bool
	ctype               [256]uint8
	F_ttype             int32
	F_sval              *String
	F_nval              float64
}

// StreamTokenizer_TT_EOF is the static StreamTokenizer.TT_EOF.
const StreamTokenizer_TT_EOF int32 = -1

// StreamTokenizer_TT_EOL is the static StreamTokenizer.TT_EOL.
const StreamTokenizer_TT_EOL int32 = '\n'

// StreamTokenizer_TT_NUMBER is the static StreamTokenizer.TT_NUMBER.
const StreamTokenizer_TT_NUMBER int32 = -2

// StreamTokenizer_TT_WORD is the static StreamTokenizer.TT_WORD.
const StreamTokenizer_TT_WORD int32 = -3

const (
	stNothing    = -4
	stNeedChar   = 1<<31 - 1
	stSkipLF     = 1<<31 - 2
	ctWhitespace = 1
	ctDigit      = 2
	ctAlpha      = 4
	ctQuote      = 8
	ctComment    = 16
)

func (this *StreamTokenizer) JClass_StreamTokenizer() *StreamTokenizer { return this }
func (this *StreamTokenizer) JClassName() string                       { return "java.io.StreamTokenizer" }
func (this *StreamTokenizer) JIdentityHash() *int32                    { return &this.jhash }

func StreamTokenizer_isInstance(o any) bool { _, ok := o.(*StreamTokenizer); return ok }

func StreamTokenizer_fromAny(o any) *StreamTokenizer { return AnyTo[*StreamTokenizer](o) }

func newStreamTokenizer() *StreamTokenizer {
	this := &StreamTokenizer{
		buf:     make([]uint16, 20),
		peekc:   stNeedChar,
		lineno:  1,
		F_ttype: stNothing,
	}
	this.WordChars_int_int('a', 'z')
	this.WordChars_int_int('A', 'Z')
	this.WordChars_int_int(128+32, 255)
	this.WhitespaceChars_int_int(0, ' ')
	this.CommentChar_int('/')
	this.QuoteChar_int('"')
	this.QuoteChar_int('\'')
	this.ParseNumbers()
	return this
}

// StreamTokenizer_new_InputStream implements the deprecated
// StreamTokenizer(InputStream), which treats each byte as a character.
func StreamTokenizer_new_InputStream(is *InputStream) *StreamTokenizer {
	this := newStreamTokenizer()
	if is == nil {
		panic(NullPointerException_new())
	}
	this.input = is.impl()
	return this
}

func StreamTokenizer_new_Reader(r *Reader) *StreamTokenizer {
	this := newStreamTokenizer()
	if r == nil {
		panic(NullPointerException_new())
	}
	this.reader = r.impl()
	return this
}

func (this *StreamTokenizer) ResetSyntax() {
	this.ctype = [256]uint8{}
}

func clampCharRange(low, hi int32) (int32, int32) {
	if low < 0 {
		low = 0
	}
	if hi >= 256 {
		hi = 255
	}
	return low, hi
}

func (this *StreamTokenizer) WordChars_int_int(low, hi int32) {
	low, hi = clampCharRange(low, hi)
	for ; low <= hi; low++ {
		this.ctype[low] |= ctAlpha
	}
}

func (this *StreamTokenizer) WhitespaceChars_int_int(low, hi int32) {
	low, hi = clampCharRange(low, hi)
	for ; low <= hi; low++ {
		this.ctype[low] = ctWhitespace
	}
}

func (this *StreamTokenizer) OrdinaryChars_int_int(low, hi int32) {
	low, hi = clampCharRange(low, hi)
	for ; low <= hi; low++ {
		this.ctype[low] = 0
	}
}

func (this *StreamTokenizer) OrdinaryChar_int(ch int32) {
	if ch >= 0 && ch < 256 {
		this.ctype[ch] = 0
	}
}

func (this *StreamTokenizer) CommentChar_int(ch int32) {
	if ch >= 0 && ch < 256 {
		this.ctype[ch] = ctComment
	}
}

func (this *StreamTokenizer) QuoteChar_int(ch int32) {
	if ch >= 0 && ch < 256 {
		this.ctype[ch] = ctQuote
	}
}

func (this *StreamTokenizer) ParseNumbers() {
	for i := '0'; i <= '9'; i++ {
		this.ctype[i] |= ctDigit
	}
	this.ctype['.'] |= ctDigit
	this.ctype['-'] |= ctDigit
}

func (this *StreamTokenizer) EolIsSignificant_boolean(flag bool)   { this.eolIsSignificantP = flag }
func (this *StreamTokenizer) SlashStarComments_boolean(flag bool)  { this.slashStarCommentsP = flag }
func (this *StreamTokenizer) SlashSlashComments_boolean(flag bool) { this.slashSlashCommentsP = flag }
func (this *StreamTokenizer) LowerCaseMode_boolean(fl bool)        { this.forceLower = fl }

func (this *StreamTokenizer) read() int32 {
	switch {
	case this.reader != nil:
		return this.reader.read()
	case this.input != nil:
		return this.input.read()
	}
	panic(IllegalStateException_new())
}

func (this *StreamTokenizer) charType(c int32) uint8 {
	if c < 256 {
		return this.ctype[c]
	}
	return ctAlpha
}

// NextToken implements StreamTokenizer.nextToken.
func (this *StreamTokenizer) NextToken() int32 {
	if this.pushedBack {
		this.pushedBack = false
		return this.F_ttype
	}
	this.F_sval = nil

	c := this.peekc
	if c < 0 {
		c = stNeedChar
	}
	if c == stSkipLF {
		c = this.read()
		if c < 0 {
			this.F_ttype = StreamTokenizer_TT_EOF
			return this.F_ttype
		}
		if c == '\n' {
			c = stNeedChar
		}
	}
	if c == stNeedChar {
		c = this.read()
		if c < 0 {
			this.F_ttype = StreamTokenizer_TT_EOF
			return this.F_ttype
		}
	}
	this.F_ttype = c
	this.peekc = stNeedChar

	ctype := this.charType(c)
	for ctype&ctWhitespace != 0 {
		if c == '\r' {
			this.lineno++
			if this.eolIsSignificantP {
				this.peekc = stSkipLF
				this.F_ttype = StreamTokenizer_TT_EOL
				return this.F_ttype
			}
			c = this.read()
			if c == '\n' {
				c = this.read()
			}
		} else {
			if c == '\n' {
				this.lineno++
				if this.eolIsSignificantP {
					this.F_ttype = StreamTokenizer_TT_EOL
					return this.F_ttype
				}
			}
			c = this.read()
		}
		if c < 0 {
			this.F_ttype = StreamTokenizer_TT_EOF
			return this.F_ttype
		}
		ctype = this.charType(c)
	}

	if ctype&ctDigit != 0 {
		neg := false
		if c == '-' {
			c = this.read()
			if c != '.' && (c < '0' || c > '9') {
				this.peekc = c
				this.F_ttype = '-'
				return this.F_ttype
			}
			neg = true
		}
		v := 0.0
		decexp := int32(0)
		seendot := int32(0)
		for {
			if c == '.' && seendot == 0 {
				seendot = 1
			} else if '0' <= c && c <= '9' {
				v = float64(v*10) + float64(c-'0')
				decexp += seendot
			} else {
				break
			}
			c = this.read()
		}
		this.peekc = c
		if decexp != 0 {
			denom := 10.0
			decexp--
			for decexp > 0 {
				denom *= 10
				decexp--
			}
			v = v / denom
		}
		if neg {
			v = -v
		}
		this.F_nval = v
		this.F_ttype = StreamTokenizer_TT_NUMBER
		return this.F_ttype
	}

	if ctype&ctAlpha != 0 {
		i := 0
		for {
			if i >= len(this.buf) {
				nb := make([]uint16, len(this.buf)*2)
				copy(nb, this.buf)
				this.buf = nb
			}
			this.buf[i] = uint16(c)
			i++
			c = this.read()
			if c < 0 {
				ctype = ctWhitespace
			} else {
				ctype = this.charType(c)
			}
			if ctype&(ctAlpha|ctDigit) == 0 {
				break
			}
		}
		this.peekc = c
		this.F_sval = FromUnits(append([]uint16(nil), this.buf[:i]...))
		if this.forceLower {
			this.F_sval = this.F_sval.ToLowerCase()
		}
		this.F_ttype = StreamTokenizer_TT_WORD
		return this.F_ttype
	}

	if ctype&ctQuote != 0 {
		this.F_ttype = c
		i := 0
		d := this.read()
		for d >= 0 && d != this.F_ttype && d != '\n' && d != '\r' {
			if d == '\\' {
				c = this.read()
				first := c
				if c >= '0' && c <= '7' {
					c = c - '0'
					c2 := this.read()
					if '0' <= c2 && c2 <= '7' {
						c = c<<3 + (c2 - '0')
						c2 = this.read()
						if '0' <= c2 && c2 <= '7' && first <= '3' {
							c = c<<3 + (c2 - '0')
							d = this.read()
						} else {
							d = c2
						}
					} else {
						d = c2
					}
				} else {
					switch c {
					case 'a':
						c = 0x7
					case 'b':
						c = '\b'
					case 'f':
						c = 0xC
					case 'n':
						c = '\n'
					case 'r':
						c = '\r'
					case 't':
						c = '\t'
					case 'v':
						c = 0xB
					}
					d = this.read()
				}
			} else {
				c = d
				d = this.read()
			}
			if i >= len(this.buf) {
				nb := make([]uint16, len(this.buf)*2)
				copy(nb, this.buf)
				this.buf = nb
			}
			this.buf[i] = uint16(c)
			i++
		}
		if d == this.F_ttype {
			this.peekc = stNeedChar
		} else {
			this.peekc = d
		}
		this.F_sval = FromUnits(append([]uint16(nil), this.buf[:i]...))
		return this.F_ttype
	}

	if c == '/' && (this.slashSlashCommentsP || this.slashStarCommentsP) {
		c = this.read()
		if c == '*' && this.slashStarCommentsP {
			prevc := int32(0)
			for {
				c = this.read()
				if c == '/' && prevc == '*' {
					break
				}
				if c == '\r' {
					this.lineno++
					c = this.read()
					if c == '\n' {
						c = this.read()
					}
				} else if c == '\n' {
					this.lineno++
					c = this.read()
				}
				if c < 0 {
					this.F_ttype = StreamTokenizer_TT_EOF
					return this.F_ttype
				}
				prevc = c
			}
			return this.NextToken()
		}
		if c == '/' && this.slashSlashCommentsP {
			this.peekc = this.skipLine()
			return this.NextToken()
		}
		if this.ctype['/']&ctComment != 0 {
			this.peekc = this.skipLine()
			return this.NextToken()
		}
		this.peekc = c
		this.F_ttype = '/'
		return this.F_ttype
	}

	if ctype&ctComment != 0 {
		this.peekc = this.skipLine()
		return this.NextToken()
	}

	this.F_ttype = c
	return this.F_ttype
}

// skipLine reads up to the end of the line and returns the character that
// ended it ('\n', '\r' or a negative value at the end of the input).
func (this *StreamTokenizer) skipLine() int32 {
	for {
		c := this.read()
		if c == '\n' || c == '\r' || c < 0 {
			return c
		}
	}
}

func (this *StreamTokenizer) PushBack() {
	if this.F_ttype != stNothing {
		this.pushedBack = true
	}
}

func (this *StreamTokenizer) Lineno() int32 { return this.lineno }

func (this *StreamTokenizer) ToString() *String {
	var ret *String
	switch this.F_ttype {
	case StreamTokenizer_TT_EOF:
		ret = Lit("EOF")
	case StreamTokenizer_TT_EOL:
		ret = Lit("EOL")
	case StreamTokenizer_TT_WORD:
		ret = this.F_sval
	case StreamTokenizer_TT_NUMBER:
		ret = Concat(Lit("n="), StrOfDouble(this.F_nval))
	case stNothing:
		ret = Lit("NOTHING")
	default:
		if this.F_ttype < 256 && this.ctype[this.F_ttype]&ctQuote != 0 {
			ret = this.F_sval
		} else {
			ret = FromUnits([]uint16{'\'', uint16(this.F_ttype), '\''})
		}
	}
	return Concat(Lit("Token["), ret, Lit("], line "), StrOfInt(this.lineno))
}
