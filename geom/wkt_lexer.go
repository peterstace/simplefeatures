package geom

import (
	"fmt"
	"strings"
	"text/scanner"
)

type wktLexer struct {
	scn    scanner.Scanner
	peeked string
}

func newWKTLexer(wkt string) wktLexer {
	var scn scanner.Scanner
	scn.Init(strings.NewReader(wkt))
	scn.Mode = scanner.ScanInts | scanner.ScanFloats | scanner.ScanIdents
	return wktLexer{scn: scn}
}

var wktUnexpectedEOF = wktSyntaxError{"unexpected EOF"}

func (w *wktLexer) next() (string, error) {
	if w.peeked != "" {
		tok := w.peeked
		w.peeked = ""
		return tok, nil
	}

	var err error
	w.scn.Error = func(_ *scanner.Scanner, msg string) {
		err = wktSyntaxError{fmt.Sprintf("invalid token '%s' (%s)", w.scn.TokenText(), msg)}
	}
	isEOF := w.scn.Scan() == scanner.EOF
	if err != nil {
		return "", err
	}
	if isEOF {
		return "", wktUnexpectedEOF
	}
	return w.scn.TokenText(), nil
}

func (w *wktLexer) peek() (string, error) {
	if w.peeked != "" {
		return w.peeked, nil
	}
	tok, err := w.next()
	if err != nil {
		return "", err
	}
	w.peeked = tok
	return tok, nil
}
