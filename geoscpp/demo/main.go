package main

import (
	"fmt"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/geoscpp"
)

func main() {
	fmt.Printf("%s\n", geoscpp.Version())
	u, err := geoscpp.Union(
		mustWKT("POLYGON((0 0,0 1,1 0,0 0))"),
		mustWKT("POLYGON((0 1,1 1,1 0,0 1))"),
	)
	fmt.Printf("g: %v\n", u.AsText())
	fmt.Printf("err: %v\n", err)
}

func mustWKT(wkt string) geom.Geometry {
	g, err := geom.UnmarshalWKT(wkt)
	if err != nil {
		panic(err)
	}
	return g
}
