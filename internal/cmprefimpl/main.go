package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/libgeos"
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

	geoms = deduplicateGeometries(geoms)

	h := libgeos.NewHandle()
	defer h.Close()

	for _, g := range geoms {
		fmt.Println(g.AsText())
		err := unaryChecks(h, g)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
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
