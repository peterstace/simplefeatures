package simplefeatures

import (
	"log"
	"math"
)

type foo int

func isPointInsideOrOnRing(pt XY, ring LinearRing) bool {
	ptg, err := NewPointFromCoords(Coordinates{pt})
	if err != nil {
		panic(err)
	}
	// find max x coordinate
	maxX := ring.ls.lines[0].a.X
	for _, ln := range ring.ls.lines {
		maxX = math.Max(maxX, ln.b.X)
		if !ln.Intersection(ptg).IsEmpty() {
			return true
		}
	}
	if pt.X > maxX {
		return false
	}

	ray, err := NewLine(Coordinates{pt}, Coordinates{XY{maxX + 1, pt.Y}})
	if err != nil {
		log.Println(Coordinates{pt}, Coordinates{XY{maxX, pt.Y}})
		panic(err)
	}
	log.Println("RAY", string(ray.AsText()))
	var count int
	for _, seg := range ring.ls.lines {
		log.Println("SEG:", string(seg.AsText()))
		inter := seg.Intersection(ray)
		if inter.IsEmpty() {
			continue
		}
		ep1, err := NewPointFromCoords(seg.a)
		if err != nil {
			panic(err)
		}
		ep2, err := NewPointFromCoords(seg.b)
		if err != nil {
			panic(err)
		}
		if inter.Dimension() == 1 || inter.Equals(ep1) || inter.Equals(ep2) {
			if seg.b.Y < pt.Y {
				log.Println("     ++")
				count++
			}
		} else {
			log.Println("     ++")
			count++
		}

		// Increment if it passes through a point NOT in the endpoints.
		// Only increment on Y condition:
		//  - It passes through as a line
		//  - It passes through an endpoint

		//inter := seg.Intersection(ray)
		//if !inter.IsEmpty() && seg.b.Y < pt.Y {
		//count++
		//}
	}
	log.Println("COUNT", count)
	return count%2 == 1
}
