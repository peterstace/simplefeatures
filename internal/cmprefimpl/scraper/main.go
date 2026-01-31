// scraper extracts string literals from the simplefeatures codebase and writes
// them to a file for use by the cmprefimpl tests. This decouples the test
// inputs from the unit test source code.
//
// Usage: go run ./internal/cmprefimpl/scraper
package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get working dir: %v", err)
	}

	blacklist, err := loadBlacklist(filepath.Join(dir, "internal/cmprefimpl/testdata/blacklist.txt"))
	if err != nil {
		log.Fatalf("could not load blacklist: %v", err)
	}

	strs, err := extractStringsFromSource(dir)
	if err != nil {
		log.Fatalf("could not extract strings from source: %v", err)
	}

	// Check for stale blacklist entries (blacklisted strings not found in source).
	for bl := range blacklist {
		found := false
		for _, s := range strs {
			if s == bl {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("WARNING: blacklisted string not found in source: %q\n", bl)
		}
	}

	// Filter out blacklisted strings.
	var filtered []string
	for _, s := range strs {
		if _, ok := blacklist[s]; !ok {
			filtered = append(filtered, s)
		}
	}

	outputPath := filepath.Join(dir, "internal/cmprefimpl/testdata/strings.txt")
	if err := writeStringsToFile(outputPath, filtered); err != nil {
		log.Fatalf("could not write output: %v", err)
	}

	fmt.Printf("Wrote %d strings to %s\n", len(filtered), outputPath)
}

func loadBlacklist(path string) (map[string]struct{}, error) {
	blacklist := make(map[string]struct{})

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		blacklist[line] = struct{}{}
	}
	return blacklist, scanner.Err()
}

func extractStringsFromSource(dir string) ([]string, error) {
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
				if err != nil {
					// Shouldn't ever happen because we've validated that it's a string literal.
					panic(fmt.Sprintf("could not unquote string '%s' from ast: %v", lit.Value, err))
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
		// Remove newlines so that strings.txt can use one-string-per-line format.
		// For WKT/WKB/GeoJSON, newlines are just whitespace and don't affect parsing.
		s = strings.ReplaceAll(s, "\n", " ")
		s = strings.ReplaceAll(s, "\r", " ")
		strSet[strings.TrimSpace(s)] = struct{}{}
	}
	strs = strs[:0]
	for s := range strSet {
		strs = append(strs, s)
	}
	sort.Strings(strs)
	return strs, nil
}

func writeStringsToFile(path string, strs []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close() // cleanup on error paths

	w := bufio.NewWriter(f)
	for _, s := range strs {
		if _, err := w.WriteString(s + "\n"); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return f.Close()
}
