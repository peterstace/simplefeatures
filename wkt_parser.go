package simplefeatures

import (
	"fmt"
	"io"
	"strconv"
)

func UnmarshalWKT(r io.Reader) (Geometry, error) {
	p := newParser(r)
	var g Geometry
	switch tok := p.nextToken(); tok {
	case "POINT":
		g = p.nextPointBody()
	case "LINESTRING":
		g = p.nextLineStringBody()
	default:
		return nil, fmt.Errorf("unexpected token: %v", tok)
	}
	p.checkEOF()
	if p.err != nil {
		return nil, p.err
	}
	return g, nil
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

func (p *parser) nextEmptyOrOpener() string {
	tok := p.nextToken()

	// TODO: this doesn't seem quite right...
	// Skip the Z, M or ZM of an SF1.2 3/4 dim coordinate.
	if tok == "Z" || tok == "M" || tok == "ZM" {
		tok = p.nextToken()
	}

	if tok != "EMPTY" && tok != "(" {
		p.check(fmt.Errorf("expected 'Z', 'M', 'ZM', 'EMPTY' or '(' but encountered %v", tok))
	}
	return tok
}

func (p *parser) nextCloser() {
	tok := p.nextToken()
	if tok != ")" {
		p.check(fmt.Errorf("expected ')' but encountered %v", tok))
	}
}

func (p *parser) nextCloserOrComma() string {
	tok := p.nextToken()
	if tok != ")" && tok != "," {
		p.check(fmt.Errorf("expected ')' or ',' but encountered %v", tok))
	}
	return tok
}

func (p *parser) preciseCoordinate() Point {
	xStr := p.nextToken()
	yStr := p.nextToken()

	// TODO: consume 0, 1, or 2 numeric tokens if they are next

	x, err := strconv.ParseFloat(xStr, 64)
	p.check(err)
	y, err := strconv.ParseFloat(yStr, 64)
	p.check(err)
	return NewPoint(x, y)
}

func (t *parser) coordinates() []Point {
	tok := t.nextEmptyOrOpener()
	if tok == "EMPTY" {
		return nil
	}

	pt := t.preciseCoordinate()
	pts := []Point{pt}

	tok = t.nextCloserOrComma()
	for tok == "," {
		pt := t.preciseCoordinate()
		pts = append(pts, pt)
		tok = t.nextCloserOrComma()
	}
	return pts
}

func (p *parser) nextPointBody() Point {
	tok := p.nextEmptyOrOpener()
	if tok == "EMPTY" {
		return NewEmptyPoint()
	}
	pt := p.preciseCoordinate()
	p.nextCloser()
	return pt
}

func (p *parser) nextLineStringBody() LineString {
	coords := p.coordinates()
	ls, err := NewLineString(coords)
	p.check(err)
	return ls
}
