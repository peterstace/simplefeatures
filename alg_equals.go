package simplefeatures

func equals(g1, g2 Geometry) bool {
	pt1, ok1 := g1.(Point)
	pt2, ok2 := g2.(Point)
	if ok1 && ok2 {
		return pt1.coords.XY == pt2.coords.XY
	}

	panic("not implemented")
}
