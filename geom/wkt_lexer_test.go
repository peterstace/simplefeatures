package geom

import (
	"io"
	"reflect"
	"strconv"
	"testing"
)

func TestWKTLexer(t *testing.T) {
	for i, tc := range []struct {
		wkt  string
		toks []string
	}{
		{
			"POINT(1 2)",
			[]string{"POINT", "(", "1", "2", ")"},
		},
		{
			// If for some reason the user puts a literal "EOF" in the input
			// WKT, it's replaced with an "<EOF>" in the token stream to
			// differentiate it with "EOF" which is emitted at the end of
			// stream.
			"POINT EOF",
			[]string{"POINT", "EOF"},
		},
		{
			`"hello`,
			[]string{`"`, "hello"},
		},
		{
			`/*hello*/ foo`,
			[]string{`/`, `*`, `hello`, `*`, `/`, `foo`},
		},
		{
			`3.14`,
			[]string{`3.14`},
		},
		{
			`3.`,
			[]string{`3.`},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			lexer := newWKTLexer(tc.wkt)
			var got []string
			for {
				tok, err := lexer.next()
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Fatal(err)
				}
				got = append(got, tok)
			}
			if !reflect.DeepEqual(got, tc.toks) {
				t.Logf("want: %v", tc.toks)
				t.Logf("got:  %v", got)
				t.Error("mismatch")
			}
		})
	}
}
