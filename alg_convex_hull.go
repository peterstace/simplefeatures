package simplefeatures

import (
	"fmt"
	"log"
	"sort"
)

func convexHullG(g Geometry) Geometry {
	if g.IsEmpty() {
		// special case to mirror postgis behaviour
		return g
	}
	pts := g.convexHullPointSet()
	hull := grahamScan(pts)
	switch len(hull) {
	case 0:
		return NewGeometryCollection(nil)
	case 1:
		return NewPoint(hull[0])
	case 2:
		ln, err := NewLine(
			Coordinates{hull[0]},
			Coordinates{hull[1]},
		)
		if err != nil {
			panic("bug in grahamScan routine - output 2 coincident points")
		}
		return ln
	default:
		hull = append(hull, hull[0]) // close the polygon
		coords := [][]Coordinates{make([]Coordinates, len(hull))}
		for i := range hull {
			coords[0][i] = Coordinates{XY: hull[i]}
		}
		poly, err := NewPolygonFromCoords(coords)
		if err != nil {
			panic(fmt.Errorf("bug in grahamScan routine - didn't produce a valid polygon: %v", err))
		}
		return poly
	}
}

type pointStack []XY

func (s *pointStack) push(p XY) {
	(*s) = append(*s, p)
}

func (s *pointStack) pop() XY {
	p := s.top()
	(*s) = (*s)[:len(*s)-1]
	return p
}

func (s *pointStack) top() XY {
	return (*s)[len(*s)-1]
}

func (s *pointStack) underTop() XY {
	return (*s)[len(*s)-2]
}

// grahamScan returns the convex hull of the input points.
func grahamScan(pts []XY) []XY {
	log.Println("grahamScan input", len(pts))
	for _, pt := range pts {
		log.Println("\t", pt)
	}
	if len(pts) <= 1 {
		return pts
	}

	sortByPolarAngle(pts)
	log.Println("grahamScan sorted")
	for _, pt := range pts {
		log.Println("\t", pt)
	}
	pts = append(pts, pts[0])

	var resultStack pointStack
	resultStack.push(pts[0])
	pts = pts[1:]
	for len(pts) > 0 && len(resultStack) < 2 {
		if !resultStack.top().Equals(pts[0]) {
			resultStack.push(pts[0])
		}
		pts = pts[1:]
	}

	log.Println("state after initial population")
	for _, pt := range resultStack {
		log.Println("\t stack ", pt)
	}
	for _, pt := range pts {
		log.Println("\t pts   ", pt)
	}

	for len(pts) > 0 {
		log.Println("considering", pts[0])
		ori := orient(resultStack.underTop(), resultStack.top(), pts[0])
		log.Println("\tori:", ori)
		switch ori {
		case leftTurn:
			log.Println("\tnot popping")
			resultStack.push(pts[0])
		case collinear:
			if distanceSq(resultStack.underTop(), pts[0]).GT(distanceSq(resultStack.underTop(), resultStack.top())) {
				resultStack.pop()
				resultStack.push(pts[0])
			}
		default:
			log.Println("\tpopping")
			resultStack.pop()
			if orient(resultStack.underTop(), resultStack.top(), pts[0]) == collinear {
				log.Println("\tdouble popping")
				resultStack.pop()
			}
			resultStack.push(pts[0])
		}
		pts = pts[1:]

		log.Println("\tstack state")
		for _, pt := range resultStack {
			log.Println("\t", pt)
		}
	}

	log.Println("grahamScan output")
	for _, pt := range resultStack {
		log.Println("\t", pt)
	}
	return resultStack
}

//func deduplicate(pts []XY) []XY {
//j := -1 // tracks last deduplicated element
//for i := range pts {
//if j < 0 || !pts[i].Equals(pts[j]) {
//j++
//pts[j] = pts[i]
//}
//}
//return pts[:j+1]
//}

// soryByPolarAngle sorts the points by their polar angle
func sortByPolarAngle(pts []XY) {
	//log.Println("sort")
	ltlp := ltl(pts)

	// swap the ltl point with first point
	pts[ltlp], pts[0] = pts[0], pts[ltlp]

	//for _, pt := range pts[1:] {
	//log.Println("\t", pt)
	//}

	virtualPoint := pts[0]
	//log.Println("\tvirt", virtualPoint)

	pts = pts[1:]
	sort.Slice(pts, func(i, j int) bool {
		//if i == 0 {
		//return false
		//}

		if virtualPoint.Equals(pts[i]) {
			//log.Printf("\tsort %s %s true", pts[i], pts[j])
			return true
		}
		if virtualPoint.Equals(pts[j]) {
			//log.Printf("\tsort %s %s false", pts[i], pts[j])
			return false
		}

		ori := orient(virtualPoint, pts[i], pts[j])

		if ori == collinear {
			//log.Printf("\tsort %s %s %t", pts[i], pts[j], distanceSq(virtualPoint, pts[i]).GT(distanceSq(virtualPoint, pts[j])))
			return distanceSq(virtualPoint, pts[i]).GT(distanceSq(virtualPoint, pts[j]))
		}

		//log.Printf("\tsort %s %s %t", pts[i], pts[j], ori == leftTurn)
		return ori == leftTurn
	})
}

// ltl stands for lowest-then-leftmost points. It returns the index of lowest-then-leftmost point
func ltl(pts []XY) int {
	rpi := 0

	for i := 1; i < len(pts); i++ {
		if pts[i].Y.LT(pts[rpi].Y) || (pts[i].Y.Equals(pts[rpi].Y) && pts[i].X.LT(pts[rpi].X)) {
			rpi = i
		}
	}

	return rpi
}

type orientation int

const (
	// rightTurn indicates the orientation is right turn which is anticlockwise
	rightTurn orientation = iota + 1
	// collinear indicates three points are on the same line
	collinear
	// leftTurn indicates the orientation is left turn which is clockwise
	leftTurn
)

func (o orientation) String() string {
	switch o {
	case rightTurn:
		return "right turn"
	case collinear:
		return "collinear"
	case leftTurn:
		return "left turn"
	default:
		return "invalid orientation"
	}
}

// orient checks if s is on the right hand side or left hand side of the line formed by p and q
// if it returns -1 which means there is an unexpected result.
func orient(p, q, s XY) orientation {
	cp := crossProduct(p, q, s)
	switch {
	case cp.GT(zero):
		return leftTurn
	case cp.Equals(zero):
		return collinear
	case cp.LT(zero):
		return rightTurn
	default:
		return -1
	}
}

// crossProduct implements Heron's formula which returns the 2 times of area formed by p, q and s
//         | p.X p.Y 1 |
// 2 * S = | q.X q.Y 1 |
//         | s.X s.Y 1 |
// when p, q and s are clockwise, the return value is negative
// when p, q and s are anticlockwise, the return value is positive
func crossProduct(p, q, s XY) Scalar {
	return q.Sub(p).Cross(s.Sub(q))
}

// distance give the length of p an q
func distanceSq(p, q XY) Scalar {
	return p.Sub(q).Dot(p.Sub(q))
}
