package simplefeatures

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

var unexpectedEOF = errors.New("unexpected EOF")

func UnmarshalWKT(r io.Reader) (Geometry, error) {
	p := newParser(r)

	tok, err := p.nextToken()
	if err != nil {
		return nil, err
	}

	switch tok {
	case "POINT":
		return p.readPointText()
	case "LINESTRING":
		return p.readLineStringText()
	default:
		return nil, fmt.Errorf("unexpected token: %v", tok)
	}
}

func newParser(r io.Reader) *parser {
	return &parser{newWKTLexer(r)}
}

type parser struct {
	lexer *wktLexer
}

func (p *parser) nextToken() (string, error) {
	tok, err := p.lexer.next()
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return tok, err
}

func (p *parser) nextEmptyOrOpener() (string, error) {
	tok, err := p.nextToken()
	if err != nil {
		return "", err
	}

	// Skip the Z, M or ZM of an SF1.2 3/4 dim coordinate.
	if tok == "Z" || tok == "M" || tok == "ZM" {
		tok, err = p.nextToken()
		if err != nil {
			return "", err
		}
	}

	if tok == "EMPTY" || tok == "(" {
		return tok, nil
	}
	return "", fmt.Errorf("expected 'Z', 'M', 'ZM', 'EMPTY' or '(' but encountered %v", tok)
}

func (p *parser) nextCloser() error {
	tok, err := p.nextToken()
	if err != nil {
		return err
	}
	if tok != ")" {
		return fmt.Errorf("expected ')' but encountered %v", tok)
	}
	return nil
}

func (p *parser) nextCloserOrComma() (string, error) {
	tok, err := p.nextToken()
	if err != nil {
		return "", err
	}
	if tok != ")" && tok != "," {
		return "", fmt.Errorf("expected ')' or ',' but encountered %v", tok)
	}
	return tok, nil
}

func (p *parser) preciseCoordinate() (Point, error) {
	xStr, err := p.nextToken()
	if err != nil {
		return Point{}, err
	}
	yStr, err := p.nextToken()
	if err != nil {
		return Point{}, err
	}

	// TODO: consume 0, 1, or 2 numeric tokens if they are next

	x, err := strconv.ParseFloat(xStr, 64)
	if err != nil {
		return Point{}, err
	}
	y, err := strconv.ParseFloat(yStr, 64)
	if err != nil {
		return Point{}, err
	}
	return NewPoint(x, y), nil
}

func (t *parser) coordinates() ([]Point, error) {
	tok, err := t.nextEmptyOrOpener()
	if err != nil {
		return nil, err
	}
	if tok == "EMPTY" {
		return nil, nil
	}

	pt, err := t.preciseCoordinate()
	if err != nil {
		return nil, err
	}
	pts := []Point{pt}

	tok, err = t.nextCloserOrComma()
	if err != nil {
		return nil, err
	}
	for tok == "," {
		pt, err := t.preciseCoordinate()
		if err != nil {
			return nil, err
		}
		pts = append(pts, pt)
		tok, err = t.nextCloserOrComma()
		if err != nil {
			return nil, err
		}
	}
	return pts, nil
}

func (p *parser) readPointText() (Point, error) {
	tok, err := p.nextEmptyOrOpener()
	if err != nil {
		return Point{}, err
	}
	if tok == "EMPTY" {
		return NewEmptyPoint(), nil
	}

	pt, err := p.preciseCoordinate()
	if err != nil {
		return Point{}, err
	}

	if err := p.nextCloser(); err != nil {
		return Point{}, err
	}

	return pt, nil
}

func (p *parser) readLineStringText() (LineString, error) {
	coords, err := p.coordinates()
	if err != nil {
		return LineString{}, err
	}
	return NewLineString(coords)
}
