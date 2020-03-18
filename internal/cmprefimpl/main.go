package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/libgeos"
)

// TODO: These are additional geometries. Needs something a bit more robust...
const (
	_ = "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT EMPTY,POINT(1 2)))"
	_ = "MULTIPOINT((1 2),(2 3),EMPTY)"
	_ = "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY)"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get working dir: %v", err)
	}
	candidates, err := extractStringsFromSource(dir)
	if err != nil {
		log.Fatalf("could not extract strings from src: %v", err)
	}

	geoms, err := convertToGeometries(candidates)
	if err != nil {
		log.Fatalf("could not convert to geometries: %v", err)
	}

	forceTo2D(geoms)
	geoms = deduplicateGeometries(geoms)

	h, err := libgeos.NewHandle()
	if err != nil {
		panic(err)
	}
	defer h.Close()

	var failures int
	for _, g := range geoms {
		var buf bytes.Buffer
		lg := log.New(&buf, "", log.Lshortfile)
		lg.Printf("========================== START ===========================")
		lg.Printf("WKT: %v", g.AsText())
		err := unaryChecks(h, g, lg)
		lg.Printf("=========================== END ============================")
		if err != nil {
			fmt.Printf("Check failed: %v\n", err)
			io.Copy(os.Stdout, &buf)
			fmt.Println()
			failures++
		}
	}
	fmt.Printf("finished unary checks on %d geometries\n", len(geoms))
	fmt.Printf("failures: %d\n", failures)
	if failures > 0 {
		os.Exit(1)
	}

	var skipped, tested int
	for _, g1 := range geoms {
		// Non-empty GeometryCollections are not supported for binary operations by libgeos.
		if g1.IsGeometryCollection() && !g1.IsEmpty() {
			skipped += len(geoms)
			continue
		}
		for _, g2 := range geoms {
			if g2.IsGeometryCollection() && !g2.IsEmpty() {
				skipped++
				continue
			}
			tested++
			var buf bytes.Buffer
			lg := log.New(&buf, "", log.Lshortfile)
			lg.Printf("========================== START ===========================")
			lg.Printf("WKT1: %v", g1.AsText())
			lg.Printf("WKT2: %v", g2.AsText())
			err := binaryChecks(h, g1, g2, lg)
			lg.Printf("=========================== END ============================")
			if err != nil {
				fmt.Printf("Check failed: %v\n", err)
				io.Copy(os.Stdout, &buf)
				fmt.Println()
				failures++
			}
		}
	}
	fmt.Printf("total binary combinations: %d\n", len(geoms)*len(geoms))
	fmt.Printf("tested combinations:       %d\n", tested)
	fmt.Printf("skipped combinations:      %d\n", skipped)
	fmt.Printf("failures:                  %d\n", failures)

	if failures > 0 {
		os.Exit(1)
	}
}

func deduplicateGeometries(geoms []geom.Geometry) []geom.Geometry {
	m := map[string]geom.Geometry{}
	for _, g := range geoms {
		m[g.AsText()] = g
	}
	geoms = geoms[:0]
	for _, g := range m {
		geoms = append(geoms, g)
	}
	sort.Slice(geoms, func(i, j int) bool {
		return geoms[i].AsText() < geoms[j].AsText()
	})
	return geoms
}

// forceTo2D converts all geometries to 2D geometries, dropping any Z and M
// values. This is done because because the C bindings for libgeos don't fully
// support Z and M value.
func forceTo2D(geoms []geom.Geometry) {
	for i := range geoms {
		geoms[i] = geoms[i].Force2D()
	}
}
