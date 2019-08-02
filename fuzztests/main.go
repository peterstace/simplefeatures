package main

import (
	"database/sql"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	path := flag.String("path", "", "path from which to search for geometries in Go code")
	dbURL := flag.String("dburl", "", "connection URL for postgis instance")
	flag.Parse()

	if *path == "" {
		fmt.Println("path not set")
		os.Exit(1)
	}
	log.Printf("path: %v", *path)

	if *dbURL == "" {
		fmt.Println("dbURL not set")
		os.Exit(1)
	}
	log.Printf("dbURL: %v", *dbURL)

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("could not open db using '%s': %v", *dbURL, err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("could not ping database: %v", err)
	}

	packages, err := buildASTsForPackages(*path)
	if err != nil {
		log.Fatalf("building ASTs: %v", err)
	}

	var candidates []string
	for _, pkg := range packages {
		for _, s := range extractStringsFromPackageAST(pkg) {
			candidates = append(candidates, s)
		}
	}
	corpus := newCorpus(db, candidates)
	corpus.loadGeometries()
	corpus.checkProperties()
}

func buildASTsForPackages(path string) (map[string]*ast.Package, error) {
	packages := map[string]*ast.Package{}
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
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
		for pkgName, pkg := range pkgs {
			packages[pkgName] = pkg
		}
		return nil
	})
	return packages, err
}

func extractStringsFromPackageAST(pkg *ast.Package) []string {
	var strs []string
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
	return strs
}
