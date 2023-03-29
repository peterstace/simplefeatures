package main

import (
	"fmt"

	"github.com/peterstace/simplefeatures/geoscpp"
)

func main() {
	fmt.Println(geoscpp.MulAdd(2, 3, 4))       // 10
	fmt.Println(geoscpp.MulAddSimple(2, 3, 4)) // 10
}
