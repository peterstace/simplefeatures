package simplefeatures

import "reflect"

func equals(g1, g2 Geometry) bool {
	if rank(g1) > rank(g2) {
		g1, g2 = g2, g1
	}
	switch g1 := g1.(type) {
	case Point:
		switch g2 := g2.(type) {
		case Point:
			return xyeq(g1.coords.XY, g2.coords.XY)
		case MultiPoint:
			g1Set := NewMultiPoint([]Point{g1})
			return equalsMultiPointAndMultiPoint(g1Set, g2)
		}
	case MultiPoint:
		switch g2 := g2.(type) {
		case MultiPoint:
			return equalsMultiPointAndMultiPoint(g1, g2)
		}
	}
	panic("not implemented")
}

func equalsMultiPointAndMultiPoint(mp1, mp2 MultiPoint) bool {
	s1 := make(map[XY]bool)
	s2 := make(map[XY]bool)
	for _, p := range mp1.pts {
		s1[p.coords.XY] = true
	}
	for _, p := range mp2.pts {
		s2[p.coords.XY] = true
	}
	return reflect.DeepEqual(s1, s2)
}
