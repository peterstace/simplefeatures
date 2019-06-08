package simplefeatures

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

func UnmarshalWKT(r io.Reader) (Geometry, error) {
	p := newParser(r)
	geom := p.nextGeometryTaggedText()
	p.checkEOF()
	if p.err != nil {
		return nil, p.err
	}
	return geom, nil
}

func newParser(r io.Reader) *parser {
	return &parser{lexer: newWKTLexer(r)}
}

type parser struct {
	lexer *wktLexer
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

func (p *parser) checkEOF() {
	tok, err := p.lexer.next()
	if err != io.EOF {
		p.check(fmt.Errorf("expected EOF but encountered %v", tok))
	}
}

func (p *parser) nextGeometryTaggedText() Geometry {
	switch tok := p.nextToken(); strings.ToLower(tok) {
	case "point":
		return p.nextPointText()
	case "linestring":
		return p.nextLineStringText()
	default:
		p.errorf("unexpected token: %v", tok)
		return nil
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

func (p *parser) nextPoint() Point {
	// TODO: handle z, m, and zm points.
	x := p.nextSignedNumericLiteral()
	y := p.nextSignedNumericLiteral()
	return NewPoint(x, y)
}

func (p *parser) nextSignedNumericLiteral() float64 {
	tok := p.nextToken()
	f, err := strconv.ParseFloat(tok, 64)
	p.check(err)
	// NaNs are not allowed by the WKT grammar.
	if math.IsNaN(f) {
		p.errorf("invalid signed numeric literal: %s", tok)
	}
	return f
}

func (p *parser) nextPointText() Point {
	tok := p.nextEmptySetOrLeftParen()
	if tok == "EMPTY" {
		return NewEmptyPoint()
	}
	pt := p.nextPoint()
	p.nextRightParen()
	return pt
}

func (p *parser) nextLineStringText() LineString {
	tok := p.nextEmptySetOrLeftParen()
	if tok == "EMPTY" {
		ls, err := NewLineString(nil)
		p.check(err)
		return ls
	}
	pt := p.nextPoint()
	pts := []Point{pt}
	for {
		tok := p.nextCommaOrRightParen()
		if tok == ")" {
			break
		}
		pts = append(pts, p.nextPoint())
	}
	ls, err := NewLineString(pts)
	p.check(err)
	return ls
}
