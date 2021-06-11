package extract

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// StringsFromSource parses the Go files (recursively) contained in the given
// dir, and returns any string literals contained therein.
func StringsFromSource(dir string) ([]string, error) {
	var strs []string
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() || strings.Contains(path, ".git") {
			return nil
		}
		pkgs, err := parser.ParseDir(new(token.FileSet), path, nil, 0)
		if err != nil {
			return err
		}
		for _, pkg := range pkgs {
			ast.Inspect(pkg, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				unquoted, err := strconv.Unquote(lit.Value)
				if !ok {
					// Shouldn't ever happen because we've validated that it's a string literal.
					panic(fmt.Sprintf("could not unquote string '%s' from AST: %v", lit.Value, err))
				}
				strs = append(strs, unquoted)
				return true
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}

	strSet := map[string]struct{}{}
	for _, s := range strs {
		strSet[strings.TrimSpace(s)] = struct{}{}
	}
	strs = strs[:0]
	for s := range strSet {
		strs = append(strs, s)
	}
	sort.Strings(strs)
	return strs, nil
}
