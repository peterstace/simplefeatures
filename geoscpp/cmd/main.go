package main

import (
	"fmt"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/geoscpp"
)

func main() {
	fmt.Println(geoscpp.MulAdd(2, 3, 4))       // 10
	fmt.Println(geoscpp.MulAddSimple(2, 3, 4)) // 10

	u, err := geoscpp.Union(geom.Geometry{}, geom.Geometry{})
	fmt.Printf("g:%q\n", u)
	fmt.Printf("err:%q\n", err)
}
