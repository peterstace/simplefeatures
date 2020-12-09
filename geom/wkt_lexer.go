package geom

import (
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

func (w *wktLexer) next() string {
	if w.peeked != "" {
		tok := w.peeked
		w.peeked = ""
		return tok
	}

	w.scn.Error = func(_ *scanner.Scanner, msg string) {
		panic(msg)
	}
	if w.scn.Scan() == scanner.EOF {
		// The special "EOF" token indicates the end of stream.
		return "EOF"
	}
	tok := w.scn.TokenText()
	if tok == "EOF" {
		// If for some reason there is a literal "EOF" in the input, we replace
		// it with something else to differentiate it from the true end of
		// stream marker (EOF). EOF doesn't appear naturally within an WKT, so
		// this is okay to do.
		tok = "<EOF>"
	}
	return tok
}

func (w *wktLexer) peek() string {
	if w.peeked != "" {
		return w.peeked
	}
	w.peeked = w.next()
	return w.peeked
}
