package simplefeatures

import (
	"log"
	"math"
)

type foo int

func isPointInsideOrOnRing(pt XY, ring LinearRing) bool {
	// find max x coordinate
	maxX := ring.ls.lines[0].a.X
	for _, ln := range ring.ls.lines {
		maxX = math.Max(maxX, ln.b.X)
	}
	if pt.X > maxX {
		return false
	}

	ray, err := NewLine(Coordinates{pt}, Coordinates{XY{maxX + 1, pt.Y}})
	if err != nil {
		log.Println(Coordinates{pt}, Coordinates{XY{maxX, pt.Y}})
		panic(err)
	}
	var count int
	for _, seg := range ring.ls.lines {
		inter := seg.Intersection(ray)
		if !inter.IsEmpty() && seg.b.Y < pt.Y {
			count++
		}
	}
	return count%2 == 1
}
