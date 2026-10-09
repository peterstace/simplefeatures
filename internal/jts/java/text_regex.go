package java

import (
	"strconv"
	"strings"
)

// This file implements the java.util.regex subset behind String.matches,
// replaceAll, replaceFirst and split: a backtracking matcher over UTF-16
// code points with Java's syntax and match semantics for literals, escapes,
// character classes (union, ranges, nesting, intersection), predefined and
// POSIX classes, general category properties, groups (capturing,
// non-capturing, named, lookahead, lookbehind), greedy, lazy and possessive
// quantifiers, alternation, back references, and the ^ $ \b \B \A \z \Z
// anchors, all without flags. Other constructs (inline flags, scripts and
// blocks, \G, \R, \X, \N) panic as unsupported.

type rxKind int

const (
	rxChar    rxKind = iota // a code point predicate
	rxConcat                // children in sequence
	rxAlt                   // alternatives
	rxGroup                 // capturing group (index) or plain grouping (-1)
	rxRepeat                // child repeated min..max times
	rxBegin                 // ^ and \A
	rxDollar                // $ and \Z
	rxEnd                   // \z
	rxBound                 // \b (negate false) and \B (negate true)
	rxLook                  // lookahead or lookbehind
	rxBackRef               // \n or \k<name>
)

type rxNode struct {
	kind     rxKind
	pred     func(cp int32) bool
	children []*rxNode
	index    int // group index, back reference
	min, max int // repetition bounds; max -1 means unbounded
	lazy     bool
	poss     bool
	negate   bool
	behind   bool
}

type regex struct {
	root   *rxNode
	groups int // number of capturing groups
	names  map[string]int
}

// compileRegex implements Pattern.compile(regex) for the supported subset.
func compileRegex(s *String) *regex {
	if s == nil {
		panic(NullPointerException_new())
	}
	var cps []int32
	for i := 0; i < len(s.units); {
		cp := codePointAtUnits(s.units, i, len(s.units))
		cps = append(cps, cp)
		i += int(Character_charCount(cp))
	}
	p := &rxParser{src: s, pat: cps, names: map[string]int{}}
	root := p.expr()
	if p.cursor < len(p.pat) {
		if p.pat[p.cursor] == ')' {
			panic(p.error("Unmatched closing ')'"))
		}
		panic(p.error("Unexpected internal error"))
	}
	return &regex{root: root, groups: p.groups, names: p.names}
}

type rxParser struct {
	src    *String
	pat    []int32
	cursor int
	groups int
	names  map[string]int
}

// error returns a PatternSyntaxException at the previous character, as
// Pattern.error does.
func (p *rxParser) error(desc string) any {
	return PatternSyntaxException_new(FromGo(desc), p.src, int32(p.cursor-1))
}

func (p *rxParser) peek() int32 {
	if p.cursor < len(p.pat) {
		return p.pat[p.cursor]
	}
	return 0
}

func (p *rxParser) next() int32 {
	p.cursor++
	return p.peek()
}

func (p *rxParser) read() int32 {
	c := p.peek()
	p.cursor++
	return c
}

func (p *rxParser) atEnd() bool { return p.cursor >= len(p.pat) }

func (p *rxParser) expr() *rxNode {
	var alts []*rxNode
	for {
		alts = append(alts, p.sequence())
		if p.atEnd() || p.peek() != '|' {
			break
		}
		p.cursor++
	}
	if len(alts) == 1 {
		return alts[0]
	}
	return &rxNode{kind: rxAlt, children: alts}
}

func (p *rxParser) sequence() *rxNode {
	seq := &rxNode{kind: rxConcat}
	for !p.atEnd() {
		ch := p.peek()
		var node *rxNode
		switch ch {
		case '|', ')':
			return seq
		case '(':
			node = p.group()
		case '[':
			p.cursor++
			pred := p.clazz(true)
			node = &rxNode{kind: rxChar, pred: pred}
		case '\\':
			p.cursor++
			if p.atEnd() {
				p.cursor = len(p.pat) + 1
				panic(p.error("Unescaped trailing backslash"))
			}
			if p.peek() == 'p' || p.peek() == 'P' {
				comp := p.peek() == 'P'
				p.cursor++
				oneLetter := true
				if p.peek() == '{' {
					oneLetter = false
					p.cursor++
				}
				node = &rxNode{kind: rxChar, pred: p.family(oneLetter, comp)}
			} else if p.peek() == 'Q' {
				p.cursor++
				for !p.atEnd() {
					if p.peek() == '\\' && p.cursor+1 < len(p.pat) && p.pat[p.cursor+1] == 'E' {
						p.cursor += 2
						break
					}
					c := p.read()
					seq.children = append(seq.children, p.closure(literalNode(c)))
				}
				continue
			} else {
				node = p.escapeNode()
			}
		case '^':
			p.cursor++
			node = &rxNode{kind: rxBegin}
		case '$':
			p.cursor++
			node = &rxNode{kind: rxDollar}
		case '.':
			p.cursor++
			node = &rxNode{kind: rxChar, pred: func(c int32) bool {
				return c != '\n' && c != '\r' && c|1 != 0x2029 && c != 0x85
			}}
		case '?', '*', '+':
			p.cursor++
			panic(p.error("Dangling meta character '" + string(rune(ch)) + "'"))
		case '{':
			// The JDK applies a quantifier here to an empty node.
			node = &rxNode{kind: rxConcat}
		default:
			p.cursor++
			node = literalNode(ch)
		}
		seq.children = append(seq.children, p.closure(node))
	}
	return seq
}

func literalNode(c int32) *rxNode {
	return &rxNode{kind: rxChar, pred: func(x int32) bool { return x == c }}
}

// group parses a group starting at '('.
func (p *rxParser) group() *rxNode {
	p.cursor++
	if p.peek() == '?' {
		p.cursor++
		ch := p.read()
		switch ch {
		case ':':
			return p.groupBody(-1)
		case '=', '!':
			body := p.expr()
			p.closeGroup()
			return &rxNode{kind: rxLook, children: []*rxNode{body}, negate: ch == '!'}
		case '>':
			body := p.expr()
			p.closeGroup()
			return &rxNode{kind: rxRepeat, children: []*rxNode{body}, min: 1, max: 1, poss: true}
		case '<':
			ch = p.read()
			if ch == '=' || ch == '!' {
				body := p.expr()
				p.closeGroup()
				return &rxNode{kind: rxLook, children: []*rxNode{body}, negate: ch == '!', behind: true}
			}
			p.cursor--
			name := p.groupName(p.read())
			if _, ok := p.names[name]; ok {
				panic(p.error("Named capturing group <" + name + "> is already defined"))
			}
			p.groups++
			p.names[name] = p.groups
			return p.groupBody(p.groups)
		case '$', '@':
			panic(p.error("Unknown group type"))
		}
		panic(Unsupported("regex inline modifiers"))
	}
	p.groups++
	return p.groupBody(p.groups)
}

func (p *rxParser) groupBody(index int) *rxNode {
	body := p.expr()
	p.closeGroup()
	return &rxNode{kind: rxGroup, index: index, children: []*rxNode{body}}
}

func (p *rxParser) closeGroup() {
	if p.atEnd() || p.peek() != ')' {
		p.cursor++
		panic(p.error("Unclosed group"))
	}
	p.cursor++
}

func (p *rxParser) groupName(ch int32) string {
	isLetter := func(c int32) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
	isDigit := func(c int32) bool { return c >= '0' && c <= '9' }
	if !isLetter(ch) {
		panic(p.error("capturing group name does not start with a Latin letter"))
	}
	var sb strings.Builder
	for isLetter(ch) || isDigit(ch) {
		sb.WriteRune(rune(ch))
		ch = p.read()
	}
	if ch != '>' {
		panic(p.error("named capturing group is missing trailing '>'"))
	}
	return sb.String()
}

// closure parses a quantifier following node, if any.
func (p *rxParser) closure(node *rxNode) *rxNode {
	if p.atEnd() {
		return node
	}
	var min, max int
	switch p.peek() {
	case '?':
		min, max = 0, 1
		p.cursor++
	case '*':
		min, max = 0, -1
		p.cursor++
	case '+':
		min, max = 1, -1
		p.cursor++
	case '{':
		p.cursor++
		ch := p.peek()
		if ch < '0' || ch > '9' {
			p.cursor++
			panic(p.error("Illegal repetition"))
		}
		readInt := func() int {
			n := 0
			for ch = p.peek(); ch >= '0' && ch <= '9'; ch = p.next() {
				n = n*10 + int(ch-'0')
				if n > 1<<31-1 {
					panic(p.error("Illegal repetition range"))
				}
			}
			return n
		}
		min = readInt()
		max = min
		if ch == ',' {
			p.cursor++
			ch = p.peek()
			if ch == '}' {
				max = -1
			} else {
				max = readInt()
			}
		}
		if p.peek() != '}' {
			p.cursor++
			panic(p.error("Unclosed counted closure"))
		}
		p.cursor++
		if max != -1 && max < min {
			panic(p.error("Illegal repetition range"))
		}
	default:
		return node
	}
	r := &rxNode{kind: rxRepeat, children: []*rxNode{node}, min: min, max: max}
	if !p.atEnd() {
		switch p.peek() {
		case '?':
			r.lazy = true
			p.cursor++
		case '+':
			r.poss = true
			p.cursor++
		}
	}
	return r
}

// escapeNode parses an escape outside a class (after the backslash).
func (p *rxParser) escapeNode() *rxNode {
	ch := p.peek()
	switch ch {
	case 'A':
		p.cursor++
		return &rxNode{kind: rxBegin}
	case 'Z':
		p.cursor++
		return &rxNode{kind: rxDollar}
	case 'z':
		p.cursor++
		return &rxNode{kind: rxEnd}
	case 'b':
		p.cursor++
		if p.peek() == '{' {
			panic(Unsupported("regex \\b{g}"))
		}
		return &rxNode{kind: rxBound}
	case 'B':
		p.cursor++
		return &rxNode{kind: rxBound, negate: true}
	case 'G', 'R', 'X', 'N':
		panic(Unsupported("regex \\" + string(rune(ch))))
	case 'k':
		p.cursor++
		if p.read() != '<' {
			panic(p.error("\\k is not followed by '<' for named capturing group"))
		}
		name := p.groupName(p.read())
		n, ok := p.names[name]
		if !ok {
			panic(p.error("named capturing group <" + name + "> does not exist"))
		}
		return &rxNode{kind: rxBackRef, index: n}
	}
	if ch >= '1' && ch <= '9' {
		// A back reference takes as many digits as name an existing group.
		p.cursor++
		ref := int(ch - '0')
		for !p.atEnd() {
			d := p.peek()
			if d < '0' || d > '9' {
				break
			}
			nr := ref*10 + int(d-'0')
			if nr > p.groups {
				break
			}
			ref = nr
			p.cursor++
		}
		return &rxNode{kind: rxBackRef, index: ref}
	}
	pred, c := p.escape(false)
	if pred != nil {
		return &rxNode{kind: rxChar, pred: pred}
	}
	return literalNode(c)
}

// escape parses an escape (after the backslash) that denotes a character
// or a predefined class; it returns either a predicate or a code point.
func (p *rxParser) escape(inclass bool) (func(int32) bool, int32) {
	ch := p.read()
	switch ch {
	case '0':
		return nil, p.octal()
	case 'd':
		return isASCIIDigit, 0
	case 'D':
		return negatePred(isASCIIDigit), 0
	case 's':
		return isASCIISpace, 0
	case 'S':
		return negatePred(isASCIISpace), 0
	case 'w':
		return isASCIIWord, 0
	case 'W':
		return negatePred(isASCIIWord), 0
	case 'h':
		return isHorizWS, 0
	case 'H':
		return negatePred(isHorizWS), 0
	case 'v':
		return isVertWS, 0
	case 'V':
		return negatePred(isVertWS), 0
	case 'a':
		return nil, 7
	case 'e':
		return nil, 27
	case 'f':
		return nil, '\f'
	case 'n':
		return nil, '\n'
	case 'r':
		return nil, '\r'
	case 't':
		return nil, '\t'
	case 'c':
		if p.atEnd() {
			panic(p.error("Illegal control escape sequence"))
		}
		return nil, p.read() ^ 64
	case 'x':
		return nil, p.hex()
	case 'u':
		return nil, p.unicode()
	}
	if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '1' && ch <= '9' {
		panic(p.error("Illegal/unsupported escape sequence"))
	}
	return nil, ch
}

func (p *rxParser) octal() int32 {
	isOct := func(c int32) bool { return c >= '0' && c <= '7' }
	n := p.peek()
	if !isOct(n) {
		panic(p.error("Illegal octal escape sequence"))
	}
	p.cursor++
	m := p.peek()
	if !isOct(m) {
		return n - '0'
	}
	p.cursor++
	o := p.peek()
	if isOct(o) && n <= '3' {
		p.cursor++
		return (n-'0')*64 + (m-'0')*8 + (o - '0')
	}
	return (n-'0')*8 + (m - '0')
}

func hexValue(c int32) int32 {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return -1
}

func (p *rxParser) hex() int32 {
	n := p.read()
	if hexValue(n) >= 0 {
		m := p.read()
		if hexValue(m) >= 0 {
			return hexValue(n)*16 + hexValue(m)
		}
	} else if n == '{' && hexValue(p.peek()) >= 0 {
		var v int32
		for {
			n = p.read()
			if hexValue(n) < 0 {
				break
			}
			v = v<<4 + hexValue(n)
			if v > 0x10ffff {
				panic(p.error("Hexadecimal codepoint is too big"))
			}
		}
		if n != '}' {
			panic(p.error("Unclosed hexadecimal escape sequence"))
		}
		return v
	}
	panic(p.error("Illegal hexadecimal escape sequence"))
}

// unicode implements Pattern.u: \uXXXX, joining an escaped surrogate pair.
func (p *rxParser) unicode() int32 {
	n := p.uxxxx()
	if Character_isHighSurrogate(uint16(n)) {
		cur := p.cursor
		if p.read() == '\\' && p.read() == 'u' {
			n2 := p.uxxxx()
			if Character_isLowSurrogate(uint16(n2)) {
				return Character_toCodePoint(uint16(n), uint16(n2))
			}
		}
		p.cursor = cur
	}
	return n
}

func (p *rxParser) uxxxx() int32 {
	var v int32
	for i := 0; i < 4; i++ {
		d := hexValue(p.read())
		if d < 0 {
			panic(p.error("Illegal Unicode escape sequence"))
		}
		v = v*16 + d
	}
	return v
}

func negatePred(f func(int32) bool) func(int32) bool { return func(c int32) bool { return !f(c) } }

func unionPred(a, b func(int32) bool) func(int32) bool {
	if a == nil {
		return b
	}
	return func(c int32) bool { return a(c) || b(c) }
}

func isASCIIDigit(c int32) bool { return c >= '0' && c <= '9' }
func isASCIISpace(c int32) bool { return c == ' ' || c >= '\t' && c <= '\r' }
func isASCIIWord(c int32) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_'
}

func isHorizWS(c int32) bool {
	return c == ' ' || c == '\t' || c == 0xa0 || c == 0x1680 || c == 0x180e || c >= 0x2000 && c <= 0x200a ||
		c == 0x202f || c == 0x205f || c == 0x3000
}

func isVertWS(c int32) bool {
	return c >= '\n' && c <= '\r' || c == 0x85 || c == 0x2028 || c == 0x2029
}

// clazz parses a character class after its '[', as Pattern.clazz does.
func (p *rxParser) clazz(consume bool) func(int32) bool {
	var prev func(int32) bool
	var bits func(int32) bool
	isNeg := false
	ch := p.peek()
	if ch == '^' && p.cursor > 0 && p.pat[p.cursor-1] == '[' {
		ch = p.next()
		isNeg = true
	}
	for {
		switch {
		case p.atEnd():
			panic(p.error("Unclosed character class"))
		case ch == '[':
			p.cursor++
			prev = unionPred(prev, p.clazz(true))
			ch = p.peek()
			continue
		case ch == '&' && p.cursor+1 < len(p.pat) && p.pat[p.cursor+1] == '&':
			p.cursor += 2
			ch = p.peek()
			var right func(int32) bool
			for ch != ']' && ch != '&' {
				if p.atEnd() {
					panic(p.error("Unclosed character class"))
				}
				if ch == '[' {
					p.cursor++
					right = unionPred(right, p.clazz(true))
				} else {
					if bits != nil {
						prev = unionPred(prev, bits)
						bits = nil
					}
					right = p.clazz(false)
				}
				ch = p.peek()
			}
			if bits != nil {
				prev = unionPred(prev, bits)
				bits = nil
			}
			if prev == nil {
				if right == nil {
					panic(p.error("Bad class syntax"))
				}
				prev = right
			} else {
				if right == nil {
					panic(p.error("Bad intersection syntax"))
				}
				l, r := prev, right
				prev = func(c int32) bool { return l(c) && r(c) }
			}
			continue
		case ch == ']':
			if prev != nil || bits != nil {
				if consume {
					p.cursor++
				}
				res := unionPred(prev, bits)
				if bits == nil {
					res = prev
				}
				if isNeg {
					return negatePred(res)
				}
				return res
			}
		}
		pred := p.rangeItem()
		bits = unionPred(bits, pred)
		ch = p.peek()
	}
}

// rangeItem parses one class member: a character, a range, an escape or a
// property.
func (p *rxParser) rangeItem() func(int32) bool {
	ch := p.peek()
	var c int32
	if ch == '\\' {
		p.cursor++
		e := p.peek()
		if e == 'p' || e == 'P' {
			p.cursor++
			oneLetter := true
			if p.peek() == '{' {
				oneLetter = false
				p.cursor++
			}
			return p.family(oneLetter, e == 'P')
		}
		if e == 'Q' {
			panic(Unsupported("regex \\Q in a character class"))
		}
		pred, v := p.escape(true)
		if pred != nil {
			return pred
		}
		c = v
	} else {
		p.cursor++
		c = ch
	}
	if p.peek() == '-' && p.cursor+1 < len(p.pat) {
		endRange := p.pat[p.cursor+1]
		if endRange != '[' && endRange != ']' {
			p.cursor++
			m := p.peek()
			if m == '\\' {
				p.cursor++
				pred, v := p.escape(true)
				if pred != nil {
					panic(p.error("Illegal character range"))
				}
				m = v
			} else {
				p.cursor++
			}
			if m < c {
				panic(p.error("Illegal character range"))
			}
			lo, hi := c, m
			return func(x int32) bool { return x >= lo && x <= hi }
		}
	}
	return func(x int32) bool { return x == c }
}

// family parses a \p or \P property name.
func (p *rxParser) family(oneLetter, negate bool) func(int32) bool {
	var name string
	if oneLetter {
		name = string(rune(p.read()))
	} else {
		start := p.cursor
		for !p.atEnd() && p.peek() != '}' {
			p.cursor++
		}
		if p.atEnd() {
			panic(p.error("Unclosed character family"))
		}
		var sb strings.Builder
		for _, c := range p.pat[start:p.cursor] {
			sb.WriteRune(rune(c))
		}
		name = sb.String()
		p.cursor++
		if name == "" {
			panic(p.error("Empty character family"))
		}
	}
	pred := propertyPredicate(name)
	if pred == nil {
		panic(p.error("Unknown character property name {" + name + "}"))
	}
	if negate {
		return negatePred(pred)
	}
	return pred
}

func typeIn(types ...int32) func(int32) bool {
	return func(c int32) bool {
		t := charType(c)
		for _, x := range types {
			if t == x {
				return true
			}
		}
		return false
	}
}

func propertyPredicate(name string) func(int32) bool {
	isPunct := func(c int32) bool {
		return c >= '!' && c <= '/' || c >= ':' && c <= '@' || c >= '[' && c <= '`' || c >= '{' && c <= '~'
	}
	isAlpha := func(c int32) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
	switch name {
	case "Lower":
		return func(c int32) bool { return c >= 'a' && c <= 'z' }
	case "Upper":
		return func(c int32) bool { return c >= 'A' && c <= 'Z' }
	case "ASCII":
		return func(c int32) bool { return c >= 0 && c < 0x80 }
	case "Alpha":
		return isAlpha
	case "Digit":
		return isASCIIDigit
	case "Alnum":
		return func(c int32) bool { return isAlpha(c) || isASCIIDigit(c) }
	case "Punct":
		return isPunct
	case "Graph":
		return func(c int32) bool { return isAlpha(c) || isASCIIDigit(c) || isPunct(c) }
	case "Print":
		return func(c int32) bool { return isAlpha(c) || isASCIIDigit(c) || isPunct(c) || c == ' ' }
	case "Blank":
		return func(c int32) bool { return c == ' ' || c == '\t' }
	case "Cntrl":
		return func(c int32) bool { return c >= 0 && c < 0x20 || c == 0x7f }
	case "XDigit":
		return func(c int32) bool { return hexValue(c) >= 0 }
	case "Space":
		return isASCIISpace
	case "javaLowerCase":
		return Character_isLowerCase_int
	case "javaUpperCase":
		return Character_isUpperCase_int
	case "javaWhitespace":
		return Character_isWhitespace_int
	case "javaLetter":
		return Character_isLetter_int
	case "javaDigit":
		return Character_isDigit_int
	case "javaLetterOrDigit":
		return Character_isLetterOrDigit_int
	case "javaAlphabetic":
		return Character_isAlphabetic
	case "javaSpaceChar":
		return Character_isSpaceChar_int
	case "javaTitleCase":
		return Character_isTitleCase_int
	case "javaDefined":
		return Character_isDefined_int
	case "javaISOControl":
		return Character_isISOControl_int
	case "javaIdeographic":
		return Character_isIdeographic
	}
	n := strings.TrimPrefix(name, "Is")
	n = strings.TrimPrefix(n, "gc=")
	n = strings.TrimPrefix(n, "general_category=")
	cats := map[string][]int32{
		"Cn": {charUnassigned}, "Lu": {charUppercaseLetter}, "Ll": {charLowercaseLetter},
		"Lt": {charTitlecaseLetter}, "Lm": {charModifierLetter}, "Lo": {charOtherLetter},
		"Mn": {charNonSpacingMark}, "Me": {charEnclosingMark}, "Mc": {charCombiningSpacingMark},
		"Nd": {charDecimalDigitNumber}, "Nl": {charLetterNumber}, "No": {charOtherNumber},
		"Zs": {charSpaceSeparator}, "Zl": {charLineSeparator}, "Zp": {charParagraphSeparator},
		"Cc": {charControl}, "Cf": {charFormat}, "Co": {charPrivateUse}, "Cs": {charSurrogate},
		"Pd": {charDashPunctuation}, "Ps": {charStartPunctuation}, "Pe": {charEndPunctuation},
		"Pc": {charConnectorPunctuation}, "Po": {charOtherPunctuation}, "Sm": {charMathSymbol},
		"Sc": {charCurrencySymbol}, "Sk": {charModifierSymbol}, "So": {charOtherSymbol},
		"Pi": {29}, "Pf": {30},
		"L":  {charUppercaseLetter, charLowercaseLetter, charTitlecaseLetter, charModifierLetter, charOtherLetter},
		"LC": {charUppercaseLetter, charLowercaseLetter, charTitlecaseLetter},
		"M":  {charNonSpacingMark, charEnclosingMark, charCombiningSpacingMark},
		"N":  {charDecimalDigitNumber, charLetterNumber, charOtherNumber},
		"Z":  {charSpaceSeparator, charLineSeparator, charParagraphSeparator},
		"C":  {charControl, charFormat, charPrivateUse, charSurrogate, charUnassigned},
		"P": {
			charDashPunctuation, charStartPunctuation, charEndPunctuation, charConnectorPunctuation,
			charOtherPunctuation, 29, 30,
		},
		"S": {charMathSymbol, charCurrencySymbol, charModifierSymbol, charOtherSymbol},
	}
	if types, ok := cats[n]; ok {
		return typeIn(types...)
	}
	switch n {
	case "all":
		return func(int32) bool { return true }
	case "Alphabetic":
		return Character_isAlphabetic
	case "Letter":
		return Character_isLetter_int
	case "Ideographic":
		return Character_isIdeographic
	case "Lowercase":
		return Character_isLowerCase_int
	case "Uppercase":
		return Character_isUpperCase_int
	case "Titlecase":
		return Character_isTitleCase_int
	case "White_Space", "WhiteSpace":
		return func(c int32) bool {
			return c >= '\t' && c <= '\r' || c == 0x85 || Character_isSpaceChar_int(c)
		}
	case "Digit":
		return Character_isDigit_int
	}
	if strings.HasPrefix(name, "In") || strings.HasPrefix(name, "Is") || strings.Contains(name, "=") {
		panic(Unsupported("regex script and block properties"))
	}
	return nil
}

// ---------------------------------------------------------------------------
// Matching

type rxMatcher struct {
	re   *regex
	text []uint16
	caps []int // start and end of each group; -1 when unset
}

func (m *rxMatcher) cpAt(i int) (int32, int) {
	cp := codePointAtUnits(m.text, i, len(m.text))
	return cp, int(Character_charCount(cp))
}

// match matches node at position i, calling k with the end position; it
// backtracks when k returns false.
func (m *rxMatcher) match(n *rxNode, i int, k func(int) bool) bool {
	switch n.kind {
	case rxChar:
		if i >= len(m.text) {
			return false
		}
		cp, w := m.cpAt(i)
		return n.pred(cp) && k(i+w)
	case rxConcat:
		return m.matchSeq(n.children, i, k)
	case rxAlt:
		for _, c := range n.children {
			if m.match(c, i, k) {
				return true
			}
		}
		return false
	case rxGroup:
		if n.index < 0 {
			return m.match(n.children[0], i, k)
		}
		s0, e0 := m.caps[2*n.index], m.caps[2*n.index+1]
		ok := m.match(n.children[0], i, func(j int) bool {
			ps, pe := m.caps[2*n.index], m.caps[2*n.index+1]
			m.caps[2*n.index], m.caps[2*n.index+1] = i, j
			if k(j) {
				return true
			}
			m.caps[2*n.index], m.caps[2*n.index+1] = ps, pe
			return false
		})
		if !ok {
			m.caps[2*n.index], m.caps[2*n.index+1] = s0, e0
		}
		return ok
	case rxRepeat:
		return m.matchRepeat(n, i, 0, k)
	case rxBegin:
		return i == 0 && k(i)
	case rxEnd:
		return i == len(m.text) && k(i)
	case rxDollar:
		end := len(m.text)
		if i < end-2 {
			return false
		}
		if i == end-2 && (m.text[i] != '\r' || m.text[i+1] != '\n') {
			return false
		}
		if i < end {
			ch := m.text[i]
			switch {
			case ch == '\n':
				if i > 0 && m.text[i-1] == '\r' {
					return false
				}
			case ch == '\r' || ch == 0x85 || ch|1 == 0x2029:
			default:
				return false
			}
		}
		return k(i)
	case rxBound:
		isWord := func(j int) bool {
			cp := codePointAtUnits(m.text, j, len(m.text))
			if isASCIIWord(cp) {
				return true
			}
			if charType(cp) != charNonSpacingMark {
				return false
			}
			for x := j; x >= 0; x-- {
				c := codePointAtUnits(m.text, x, len(m.text))
				if Character_isLetterOrDigit_int(c) {
					return true
				}
				if charType(c) != charNonSpacingMark {
					return false
				}
			}
			return false
		}
		left := false
		if i > 0 {
			cp := codePointBeforeUnits(m.text, i, 0)
			left = isWord(i - int(Character_charCount(cp)))
		}
		right := i < len(m.text) && isWord(i)
		return left != right != n.negate && k(i)
	case rxLook:
		saved := append([]int(nil), m.caps...)
		var found bool
		if n.behind {
			for j := i; j >= 0 && !found; j-- {
				found = m.match(n.children[0], j, func(e int) bool { return e == i })
			}
		} else {
			found = m.match(n.children[0], i, func(int) bool { return true })
		}
		if found == n.negate {
			copy(m.caps, saved)
			return false
		}
		if n.negate {
			copy(m.caps, saved)
		}
		if k(i) {
			return true
		}
		copy(m.caps, saved)
		return false
	case rxBackRef:
		s, e := m.caps[2*n.index], m.caps[2*n.index+1]
		if s < 0 {
			return false
		}
		l := e - s
		if i+l > len(m.text) || !unitsEqual(m.text[i:i+l], m.text[s:e]) {
			return false
		}
		return k(i + l)
	}
	return false
}

func (m *rxMatcher) matchSeq(ns []*rxNode, i int, k func(int) bool) bool {
	if len(ns) == 0 {
		return k(i)
	}
	return m.match(ns[0], i, func(j int) bool { return m.matchSeq(ns[1:], j, k) })
}

// matchRepeat matches the remaining iterations of a repetition, count of
// which have matched, ending at i.
func (m *rxMatcher) matchRepeat(n *rxNode, i, count int, k func(int) bool) bool {
	child := n.children[0]
	if n.poss {
		// Match as many iterations as possible without backtracking into
		// them.
		pos := i
		for c := 0; n.max == -1 || c < n.max; c++ {
			end := -1
			if !m.match(child, pos, func(j int) bool { end = j; return true }) || end == pos && c >= n.min {
				if c < n.min {
					return false
				}
				break
			}
			pos = end
		}
		return k(pos)
	}
	canMore := n.max == -1 || count < n.max
	tryMore := func() bool {
		return canMore && m.match(child, i, func(j int) bool {
			if j == i {
				// A zero-length iteration ends the loop, as in Pattern.Loop.
				return k(j)
			}
			return m.matchRepeat(n, j, count+1, k)
		})
	}
	if count < n.min {
		return tryMore()
	}
	if n.lazy {
		return k(i) || tryMore()
	}
	return tryMore() || k(i)
}

// find implements Matcher.find from position from: the leftmost match,
// with its group positions in m.caps.
func (m *rxMatcher) find(from int) (int, int, bool) {
	for start := from; start <= len(m.text); start++ {
		if start > 0 && start < len(m.text) && Character_isLowSurrogate(m.text[start]) &&
			Character_isHighSurrogate(m.text[start-1]) {
			continue
		}
		for i := range m.caps {
			m.caps[i] = -1
		}
		end := -1
		if m.match(m.re.root, start, func(j int) bool { end = j; return true }) {
			m.caps[0], m.caps[1] = start, end
			return start, end, true
		}
	}
	return 0, 0, false
}

// newMatcher returns a matcher for text. Like the JDK's Matcher, it has room
// for at least ten groups, so a back reference to a missing group below ten
// never matches rather than failing.
func (re *regex) newMatcher(text []uint16) *rxMatcher {
	n := re.groups + 1
	if n < 10 {
		n = 10
	}
	return &rxMatcher{re: re, text: text, caps: make([]int, 2*n)}
}

// matches implements Matcher.matches: the whole text must match.
func (re *regex) matches(text []uint16) bool {
	m := re.newMatcher(text)
	return m.match(re.root, 0, func(j int) bool { return j == len(text) })
}

// replace implements Matcher.replaceAll and replaceFirst.
func (re *regex) replace(s *String, replacement *String, all bool) *String {
	if replacement == nil {
		panic(NullPointerException_new_String(Lit("replacement")))
	}
	m := re.newMatcher(s.units)
	start, end, ok := m.find(0)
	if !ok {
		return s
	}
	var out []uint16
	last := 0
	for ok {
		out = append(out, s.units[last:start]...)
		out = m.appendReplacement(out, replacement.units)
		last = end
		if !all {
			break
		}
		from := end
		if end == start {
			from = end + 1
		}
		if from > len(s.units) {
			break
		}
		start, end, ok = m.find(from)
	}
	out = append(out, s.units[last:]...)
	return FromUnits(out)
}

// appendReplacement implements Matcher.appendExpandedReplacement.
func (m *rxMatcher) appendReplacement(out []uint16, r []uint16) []uint16 {
	isDigit := func(c uint16) bool { return c >= '0' && c <= '9' }
	for cursor := 0; cursor < len(r); {
		c := r[cursor]
		switch c {
		case '\\':
			cursor++
			if cursor == len(r) {
				panic(IllegalArgumentException_new_String(Lit("character to be escaped is missing")))
			}
			out = append(out, r[cursor])
			cursor++
		case '$':
			cursor++
			if cursor == len(r) {
				panic(IllegalArgumentException_new_String(Lit("Illegal group reference: group index is missing")))
			}
			var refNum int
			if r[cursor] == '{' {
				cursor++
				var name []uint16
				for cursor < len(r) && (isDigit(r[cursor]) || r[cursor] >= 'a' && r[cursor] <= 'z' ||
					r[cursor] >= 'A' && r[cursor] <= 'Z') {
					name = append(name, r[cursor])
					cursor++
				}
				if len(name) == 0 {
					panic(IllegalArgumentException_new_String(Lit("named capturing group has 0 length name")))
				}
				if cursor >= len(r) || r[cursor] != '}' {
					panic(IllegalArgumentException_new_String(Lit("named capturing group is missing trailing '}'")))
				}
				gname := FromUnits(name).Go()
				if isDigit(name[0]) {
					panic(IllegalArgumentException_new_String(FromGo("capturing group name {" + gname + "} starts with digit character")))
				}
				n, ok := m.re.names[gname]
				if !ok {
					panic(IllegalArgumentException_new_String(FromGo("No group with name {" + gname + "}")))
				}
				refNum = n
				cursor++
			} else {
				refNum = int(r[cursor]) - '0'
				if refNum < 0 || refNum > 9 {
					panic(IllegalArgumentException_new_String(Lit("Illegal group reference")))
				}
				cursor++
				for cursor < len(r) && isDigit(r[cursor]) {
					nr := refNum*10 + int(r[cursor]-'0')
					if m.re.groups < nr {
						break
					}
					refNum = nr
					cursor++
				}
			}
			if refNum > m.re.groups {
				panic(IndexOutOfBoundsException_new_String(FromGo("No group " + strconv.Itoa(refNum))))
			}
			if s, e := m.caps[2*refNum], m.caps[2*refNum+1]; s != -1 && e != -1 {
				out = append(out, m.text[s:e]...)
			}
		default:
			out = append(out, c)
			cursor++
		}
	}
	return out
}

// split implements String.split(String, int) and Pattern.split.
func (re *regex) split(s *String, limit int32) []*String {
	m := re.newMatcher(s.units)
	index := 0
	matchLimited := limit > 0
	var list []*String
	from := 0
	for from <= len(s.units) {
		start, end, ok := m.find(from)
		if !ok {
			break
		}
		if !matchLimited || int32(len(list)) < limit-1 {
			if !(index == 0 && start == 0 && start == end) {
				list = append(list, s.Substring_int_int(int32(index), int32(start)))
				index = end
			}
		} else if int32(len(list)) == limit-1 {
			list = append(list, s.Substring_int_int(int32(index), int32(len(s.units))))
			index = end
		}
		from = end
		if end == start {
			from = end + 1
		}
	}
	if index == 0 {
		return []*String{s}
	}
	if !matchLimited || int32(len(list)) < limit {
		list = append(list, s.Substring_int_int(int32(index), int32(len(s.units))))
	}
	n := len(list)
	if limit == 0 {
		for n > 0 && len(list[n-1].units) == 0 {
			n--
		}
	}
	return list[:n]
}
