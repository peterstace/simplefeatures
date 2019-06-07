package gatig

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"text/scanner"
)

var unexpectedEOF = errors.New("unexpected EOF")

func UnmarshalWKT(r io.Reader) (Geometry, error) {
	toker := newTokenizer(r)
	tok, ok := toker.nextToken()
	if !ok {
		return nil, unexpectedEOF
	}
	switch tok {
	case "POINT":
		return toker.readPointText()
	case "LINESTRING":
		return toker.readLineStringText()
	default:
		return nil, fmt.Errorf("unexpected token: %v", tok)
	}

	/*
		toker := newTokenizer(r)
		for {
			tok, ok := toker.nextWord()
			if !ok {
				break
			}
			fmt.Println(tok)
			time.Sleep(100 * time.Millisecond)
		}
		return nil, nil
	*/
}

func newTokenizer(r io.Reader) *tokenizer {
	var scanner scanner.Scanner
	scanner.Init(r)
	return &tokenizer{scanner}
}

type tokenizer struct {
	scn scanner.Scanner
}

func (t *tokenizer) nextToken() (string, bool) {
	tok := t.scn.Scan()
	if tok == scanner.EOF {
		return "", false
	}
	return t.scn.TokenText(), true
}

func (t *tokenizer) nextEmptyOrOpener() (string, error) {
	tok, ok := t.nextToken()
	if !ok {
		return "", unexpectedEOF
	}

	// Skip the Z, M or ZM of an SF1.2 3/4 dim coordinate.
	if tok == "Z" || tok == "M" || tok == "ZM" {
		tok, ok = t.nextToken()
		if !ok {
			return "", unexpectedEOF
		}
	}

	if tok == "EMPTY" || tok == "(" {
		return tok, nil
	}
	return "", fmt.Errorf("expected 'Z', 'M', 'ZM', 'EMPTY' or '(' but encountered %v", tok)
}

func (t *tokenizer) nextCloser() error {
	tok, ok := t.nextToken()
	if !ok {
		return unexpectedEOF
	}
	if tok != ")" {
		return fmt.Errorf("expected ')' but encountered %v", tok)
	}
	return nil
}

func (t *tokenizer) preciseCoordinate() (Point, error) {
	xStr, ok := t.nextToken()
	if !ok {
		return Point{}, unexpectedEOF
	}
	yStr, ok := t.nextToken()
	if !ok {
		return Point{}, unexpectedEOF
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

func (t *tokenizer) coordinates() ([]Point, error) {
	// TODO
	return nil, nil
}

func (t *tokenizer) readPointText() (Point, error) {
	tok, err := t.nextEmptyOrOpener()
	if err != nil {
		return Point{}, err
	}
	if tok == "EMPTY" {
		return NewEmptyPoint(), nil
	}

	pt, err := t.preciseCoordinate()
	if err != nil {
		return Point{}, err
	}

	if err := t.nextCloser(); err != nil {
		return Point{}, err
	}

	return pt, nil
}

func (t *tokenizer) readLineStringText() (LineString, error) {
	coords, err := t.coordinates()
	if err != nil {
		return LineString{}, err
	}
	return NewLineString(coords)
}
