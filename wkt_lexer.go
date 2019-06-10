package simplefeatures

import (
	"errors"
	"io"
	"text/scanner"
)

//TODO: make negative numbers work

type wktLexer struct {
	scn       scanner.Scanner
	peeked    bool
	nextToken string
}

func newWKTLexer(r io.Reader) *wktLexer {
	var scn scanner.Scanner
	scn.Init(r)
	return &wktLexer{scn: scn}
}

func (w *wktLexer) next() (string, error) {
	if w.peeked {
		tok := w.nextToken
		w.peeked = false
		w.nextToken = ""
		return tok, nil
	}

	var err error
	w.scn.Error = func(_ *scanner.Scanner, msg string) {
		err = errors.New(msg)
	}
	isEOF := w.scn.Scan() == scanner.EOF
	if err != nil {
		return "", err
	}
	if isEOF {
		return "", io.EOF
	}
	return w.scn.TokenText(), nil
}

func (w *wktLexer) peek() (string, error) {
	if w.peeked {
		return w.nextToken, nil
	}
	tok, err := w.next()
	if err != nil {
		return "", err
	}
	w.peeked = true
	w.nextToken = tok
	return tok, nil
}
