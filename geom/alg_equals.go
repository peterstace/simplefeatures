package geom

import (
	"fmt"
	"reflect"
)

func equals(g1, g2 Geometry) (bool, error) {
	if g1.IsEmpty() && g2.IsEmpty() {
		return true, nil
	}
	if g1.IsEmpty() || g2.IsEmpty() {
		return false, nil
	}
	if g1.Dimension() != g2.Dimension() {
		return false, nil
	}

	if rank(g1) > rank(g2) {
		g1, g2 = g2, g1
	}
	switch g1 := g1.(type) {
	case Point:
		switch g2 := g2.(type) {
		case Point:
			return g1.coords.XY.Equals(g2.coords.XY), nil
		case MultiPoint:
			g1Set := NewMultiPoint([]Point{g1})
			return equalsMultiPointAndMultiPoint(g1Set, g2), nil
		}
	case MultiPoint:
		switch g2 := g2.(type) {
		case MultiPoint:
			return equalsMultiPointAndMultiPoint(g1, g2), nil
		}
	}
	return false, fmt.Errorf("not implemented: equals with %T and %T", g1, g2)
}

func equalsMultiPointAndMultiPoint(mp1, mp2 MultiPoint) bool {
	s1 := make(map[XY]struct{})
	s2 := make(map[XY]struct{})
	for _, p := range mp1.pts {
		s1[p.coords.XY] = struct{}{}
	}
	for _, p := range mp2.pts {
		s2[p.coords.XY] = struct{}{}
	}
	return reflect.DeepEqual(s1, s2)
}
