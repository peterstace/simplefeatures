package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/peterstace/simplefeatures/geom"
)

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
				if !ok {
					// Shouldn't ever happen because we've validated that it's a string literal.
					panic(fmt.Sprintf("could not unquote string '%s'from ast: %v", lit.Value, err))
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

func convertToGeometries(candidates []string) ([]geom.Geometry, error) {
	var geoms []geom.Geometry
	for _, c := range candidates {
		g, err := geom.UnmarshalWKT(c, geom.NoValidate{})
		if err == nil {
			geoms = append(geoms, g)
		}
	}
	if len(geoms) == 0 {
		return nil, errors.New("could not extract any WKT geoms")
	}

	oldCount := len(geoms)
	for _, c := range candidates {
		buf, err := hexStringToBytes(c)
		if err != nil {
			continue
		}
		g, err := geom.UnmarshalWKB(buf, geom.NoValidate{})
		if err == nil {
			geoms = append(geoms, g)
		}
	}
	if oldCount == len(geoms) {
		return nil, errors.New("could not extract any WKB geoms")
	}

	oldCount = len(geoms)
	for _, c := range candidates {
		g, err := geom.UnmarshalGeoJSON([]byte(c), geom.NoValidate{})
		if err == nil {
			geoms = append(geoms, g)
		}
	}
	if oldCount == len(geoms) {
		return nil, errors.New("could not extract any geojson geoms")
	}

	return geoms, nil
}

func hexStringToBytes(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, errors.New("hex string must have even length")
	}
	var buf []byte
	for i := 0; i < len(s); i += 2 {
		x, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			return nil, err
		}
		buf = append(buf, byte(x))
	}
	return buf, nil
}
