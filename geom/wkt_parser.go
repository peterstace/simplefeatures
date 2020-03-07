package geom

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// Function names is the parser are chosen to match closely with the BNF
// productions in the WKT grammar.
//
// Convention: functions starting with 'next' consume token(s), and build the
// next production in the grammar.

// UnmarshalWKT parses a Well Known Text (WKT), and returns the corresponding
// Geometry.
func UnmarshalWKT(r io.Reader, opts ...ConstructorOption) (Geometry, error) {
	p := newParser(r, opts)
	geom := p.nextGeometryTaggedText()
	p.checkEOF()
	if p.err != nil {
		return Geometry{}, p.err
	}
	return geom, nil
}

func newParser(r io.Reader, opts []ConstructorOption) *parser {
	return &parser{lexer: newWKTLexer(r), opts: opts}
}

type parser struct {
	lexer *wktLexer
	opts  []ConstructorOption
	err   error
}

func (p *parser) check(err error) {
	if err != nil && p.err == nil {
		p.err = err
	}
}

func (p *parser) errorf(format string, args ...interface{}) {
	p.check(fmt.Errorf(format, args...))
}

func (p *parser) nextToken() string {
	tok, err := p.lexer.next()
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	p.check(err)
	return tok
}

func (p *parser) peekToken() string {
	tok, err := p.lexer.peek()
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	p.check(err)
	return tok
}

func (p *parser) checkEOF() {
	tok, err := p.lexer.next()
	if err != io.EOF {
		p.check(fmt.Errorf("expected EOF but encountered %v", tok))
	}
}

func (p *parser) nextGeometryTaggedText() Geometry {
	geomType, ctype := p.nextGeomTag()
	switch geomType {
	case "POINT":
		c, ok := p.nextPointText(ctype)
		if !ok {
			return NewEmptyPoint(ctype).AsGeometry()
		} else {
			return NewPointC(c, ctype, p.opts...).AsGeometry()
		}
	case "LINESTRING":
		ls := p.nextLineStringText(ctype)
		seq := ls.Coordinates()
		n := seq.Length()
		if n == 2 {
			ln, err := NewLineC(seq.Get(0), seq.Get(1), ctype, p.opts...)
			p.check(err)
			return ln.AsGeometry()
		}
		return ls.AsGeometry()
	case "POLYGON":
		return p.nextPolygonText(ctype).AsGeometry()
	case "MULTIPOINT":
		return p.nextMultiPointText(ctype).AsGeometry()
	case "MULTILINESTRING":
		return p.nextMultiLineString(ctype).AsGeometry()
	case "MULTIPOLYGON":
		return p.nextMultiPolygonText(ctype).AsGeometry()
	case "GEOMETRYCOLLECTION":
		return p.nextGeometryCollectionText(ctype)
	default:
		p.errorf("unexpected token: %v", geomType)
		return Geometry{}
	}
}

func (p *parser) nextGeomTag() (string, CoordinatesType) {
	geomType := strings.ToUpper(p.nextToken())
	switch p.peekToken() {
	case "Z":
		p.nextToken()
		return geomType, XYZ
	case "M":
		p.nextToken()
		return geomType, XYM
	case "ZM":
		p.nextToken()
		return geomType, XYZM
	default:
		return geomType, XYOnly
	}
}

func (p *parser) nextEmptySetOrLeftParen() string {
	tok := p.nextToken()
	if tok != "EMPTY" && tok != "(" {
		p.errorf("expected 'EMPTY' or '(' but encountered %v", tok)
	}
	return tok
}

func (p *parser) nextRightParen() {
	tok := p.nextToken()
	if tok != ")" {
		p.check(fmt.Errorf("expected ')' but encountered %v", tok))
	}
}

func (p *parser) nextCommaOrRightParen() string {
	tok := p.nextToken()
	if tok != ")" && tok != "," {
		p.check(fmt.Errorf("expected ')' or ',' but encountered %v", tok))
	}
	return tok
}

func (p *parser) nextPoint(ctype CoordinatesType) Coordinates {
	var c Coordinates
	c.X = p.nextSignedNumericLiteral()
	c.Y = p.nextSignedNumericLiteral()
	if ctype.Is3D() {
		c.Z = p.nextSignedNumericLiteral()
	}
	if ctype.IsMeasured() {
		c.M = p.nextSignedNumericLiteral()
	}
	return c
}

func (p *parser) nextPointAppend(dst []float64, ctype CoordinatesType) []float64 {
	for i := 0; i < ctype.Dimension(); i++ {
		dst = append(dst, p.nextSignedNumericLiteral())
	}
	return dst
}

func (p *parser) nextSignedNumericLiteral() float64 {
	var negative bool
	tok := p.nextToken()
	if tok == "-" {
		negative = true
		tok = p.nextToken()
	}
	f, err := strconv.ParseFloat(tok, 64)
	p.check(err)
	// NaNs and Infs are not allowed by the WKT grammar.
	if math.IsNaN(f) || math.IsInf(f, 0) {
		p.errorf("invalid signed numeric literal: %s", tok)
	}
	if negative {
		f *= -1
	}
	return f
}

func (p *parser) nextPointText(ctype CoordinatesType) (Coordinates, bool) {
	tok := p.nextEmptySetOrLeftParen()
	if tok == "EMPTY" {
		return Coordinates{}, false
	}
	c := p.nextPoint(ctype)
	p.nextRightParen()
	return c, true
}

func (p *parser) nextLineStringText(ctype CoordinatesType) LineString {
	var floats []float64
	tok := p.nextEmptySetOrLeftParen()
	if tok == "(" {
		floats = p.nextPointAppend(floats, ctype)
		for {
			tok := p.nextCommaOrRightParen()
			if tok == "," {
				floats = p.nextPointAppend(floats, ctype)
			} else {
				break
			}
		}
	}
	seq := NewSequenceNoCopy(floats, ctype)
	ls, err := NewLineStringFromSequence(seq, p.opts...)
	p.check(err)
	return ls
}

func (p *parser) nextPolygonText(ctype CoordinatesType) Polygon {
	rings := p.nextPolygonOrMultiLineStringText(ctype)
	poly, err := NewPolygon(rings, ctype, p.opts...)
	p.check(err)
	return poly
}

func (p *parser) nextMultiLineString(ctype CoordinatesType) MultiLineString {
	lss := p.nextPolygonOrMultiLineStringText(ctype)
	mls, err := NewMultiLineString(lss, ctype, p.opts...)
	p.check(err)
	return mls
}

func (p *parser) nextPolygonOrMultiLineStringText(ctype CoordinatesType) []LineString {
	tok := p.nextEmptySetOrLeftParen()
	if tok == "EMPTY" {
		return nil
	}
	ls := p.nextLineStringText(ctype)
	lss := []LineString{ls}
	for {
		tok := p.nextCommaOrRightParen()
		if tok == "," {
			lss = append(lss, p.nextLineStringText(ctype))
		} else {
			break
		}
	}
	return lss
}

func (p *parser) nextMultiPointText(ctype CoordinatesType) MultiPoint {
	var floats []float64
	var empty BitSet
	tok := p.nextEmptySetOrLeftParen()
	if tok == "(" {
		for i := 0; true; i++ {
			if p.peekToken() == "EMPTY" {
				p.nextToken()
				for j := 0; j < ctype.Dimension(); j++ {
					floats = append(floats, 0)
				}
				empty.Set(i)
			} else {
				floats = p.nextMultiPointStylePointAppend(floats, ctype)
			}
			tok := p.nextCommaOrRightParen()
			if tok != "," {
				break
			}
		}
	}
	seq := NewSequenceNoCopy(floats, ctype)
	return NewMultiPointFromSequence(seq, empty, p.opts...)
}

func (p *parser) nextMultiPointStylePointAppend(dst []float64, ctype CoordinatesType) []float64 {
	// This is an extension of the spec, and is required to handle WKT output
	// from non-complying implementations. In particular, PostGIS doesn't
	// comply to the spec (it outputs points as part of a multipoint without
	// their surrounding parentheses).
	var useParens bool
	if p.peekToken() == "(" {
		p.nextToken()
		useParens = true
	}
	dst = p.nextPointAppend(dst, ctype)
	if useParens {
		p.nextRightParen()
	}
	return dst
}

func (p *parser) nextMultiPolygonText(ctype CoordinatesType) MultiPolygon {
	var polys []Polygon
	tok := p.nextEmptySetOrLeftParen()
	if tok == "(" {
		for {
			poly := p.nextPolygonText(ctype)
			polys = append(polys, poly)
			tok := p.nextCommaOrRightParen()
			if tok == ")" {
				break
			}
		}
	}
	mp, err := NewMultiPolygon(polys, ctype, p.opts...)
	p.check(err)
	return mp
}

func (p *parser) nextGeometryCollectionText(ctype CoordinatesType) Geometry {
	var geoms []Geometry
	tok := p.nextEmptySetOrLeftParen()
	if tok == "(" {
		for {
			g := p.nextGeometryTaggedText()
			geoms = append(geoms, g)
			tok := p.nextCommaOrRightParen()
			if tok == ")" {
				break
			}
		}
	}
	gc, err := NewGeometryCollection(geoms, ctype, p.opts...)
	p.check(err)
	return gc.AsGeometry()
}
