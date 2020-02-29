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
	switch tok := p.nextToken(); strings.ToUpper(tok) {
	case "POINT":
		ctype := XYOnly // TODO
		c, ok := p.nextPointText(ctype)
		if !ok {
			return NewEmptyPoint(ctype).AsGeometry()
		} else {
			return NewPointC(c, ctype, p.opts...).AsGeometry()
		}
	case "LINESTRING":
		ctype := XYOnly // TODO
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
		return p.nextPolygonText(XYOnly /*TODO*/).AsGeometry()
	case "MULTIPOINT":
		return p.nextMultiPointText(XYOnly /*TODO*/).AsGeometry()
	case "MULTILINESTRING":
		return p.nextMultiLineString(XYOnly /*TODO*/).AsGeometry()
	case "MULTIPOLYGON":
		return p.nextMultiPolygonText(XYOnly /*TODO*/).AsGeometry()
	case "GEOMETRYCOLLECTION":
		return p.nextGeometryCollectionText(XYOnly /*TODO*/)
	default:
		p.errorf("unexpected token: %v", tok)
		return Geometry{}
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
	// TODO: handle z, m, and zm points.
	x := p.nextSignedNumericLiteral()
	y := p.nextSignedNumericLiteral()
	return Coordinates{XY: XY{x, y}}
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
	tok := p.nextEmptySetOrLeftParen()
	if tok == "EMPTY" {
		return NewEmptyLineString(ctype)
	}
	floats := p.nextPointAppend(nil, ctype)
	for {
		tok := p.nextCommaOrRightParen()
		if tok == "," {
			floats = p.nextPointAppend(floats, ctype)
		} else {
			break
		}
	}
	seq := NewSequenceNoCopy(floats, ctype)
	ls, err := NewLineStringFromSequence(seq, p.opts...)
	p.check(err)
	return ls
}

func (p *parser) nextPolygonText(ctype CoordinatesType) Polygon {
	rings := p.nextPolygonOrMultiLineStringText(ctype)
	if len(rings) == 0 {
		return NewEmptyPolygon(ctype)
	} else {
		poly, err := NewPolygon(rings, p.opts...)
		p.check(err)
		return poly
	}
}

func (p *parser) nextMultiLineString(ctype CoordinatesType) MultiLineString {
	lss := p.nextPolygonOrMultiLineStringText(ctype)
	if len(lss) == 0 {
		return NewEmptyMultiLineString(ctype)
	} else {
		mls, err := NewMultiLineString(lss, p.opts...)
		p.check(err)
		return mls
	}
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
	tok := p.nextEmptySetOrLeftParen()
	if tok == "EMPTY" {
		return NewEmptyMultiPoint(ctype)
	}

	var floats []float64
	var empty BitSet
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
	tok := p.nextEmptySetOrLeftParen()
	if tok == "EMPTY" {
		return NewEmptyMultiPolygon(ctype)
	}
	poly := p.nextPolygonText(ctype)
	polys := []Polygon{poly}
	for {
		tok := p.nextCommaOrRightParen()
		if tok == "," {
			poly := p.nextPolygonText(ctype)
			polys = append(polys, poly)
		} else {
			break
		}
	}
	mp, err := NewMultiPolygon(polys, p.opts...)
	p.check(err)
	return mp
}

func (p *parser) nextGeometryCollectionText(ctype CoordinatesType) Geometry {
	tok := p.nextEmptySetOrLeftParen()
	if tok == "EMPTY" {
		return NewGeometryCollection(nil, p.opts...).AsGeometry()
	}
	geoms := []Geometry{
		p.nextGeometryTaggedText(),
	}
	for {
		tok := p.nextCommaOrRightParen()
		if tok == "," {
			geom := p.nextGeometryTaggedText()
			geoms = append(geoms, geom)
		} else {
			break
		}
	}
	return NewGeometryCollection(geoms, p.opts...).AsGeometry()
}
