package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/peterstace/simplefeatures/geom"
)

func main() {
	a0 := `{"type":"Polygon","coordinates":[[
		[-83.58253051,32.73168239],
		[-83.59843118,32.74617142],
		[-83.70048117,32.63984372],
		[-83.58253051,32.73168239]]]
	}`
	a1 := `{"type":"Polygon","coordinates":[[
		[-83.70047745,32.63984661],
		[-83.68891846,32.59896320],
		[-83.58253417,32.73167955],
		[-83.70047745,32.63984661]]]
	}`

	var g0, g1 geom.Geometry

	err := json.NewDecoder(strings.NewReader(a0)).Decode(&g0)
	if err != nil {
		log.Fatal(err)

	}

	err = json.NewDecoder(strings.NewReader(a1)).Decode(&g1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(g0.Summary())
	fmt.Println(g1.Summary())

	u, err := geom.Union(g0, g1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(u.AsText())
}
