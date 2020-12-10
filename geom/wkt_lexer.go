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

// errEOF is a sentinel error to indicate the end of token stream. Although the
// text of the error indicates that the EOF is "unexpected", this may not
// necessarily be the case (in which case the caller should substitute the
// error with a different behaviour or error).
var errEOF = SyntaxError{"unexpected EOF"}

func (w *wktLexer) next() (string, error) {
	if w.peeked != "" {
		tok := w.peeked
		w.peeked = ""
		return tok, nil
	}

	var err error
	w.scn.Error = func(_ *scanner.Scanner, msg string) {
		err = SyntaxError{fmt.Sprintf("invalid token: '%v' (%s)", w.scn.TokenText(), msg)}
	}
	isEOF := w.scn.Scan() == scanner.EOF
	if err != nil {
		return "", err
	}
	if isEOF {
		return "", errEOF
	}
	return w.scn.TokenText(), nil
}

func (w *wktLexer) peek() (string, error) {
	if w.peeked != "" {
		return w.peeked, nil
	}
	tok, err := w.next()
	if err != nil {
		// This error is from the original io.Reader, so no need to use a
		// structured error.
		return "", err
	}
	w.peeked = tok
	return tok, nil
}
