package simplefeatures

func isPointInsideOrOnRing(pt XY, ring LinearRing) bool {
	ptg, err := NewPointFromCoords(Coordinates{pt})
	if err != nil {
		panic(err)
	}
	// find max x coordinate
	// TODO: should be able to use envelope for this
	maxX := ring.ls.lines[0].a.X
	for _, ln := range ring.ls.lines {
		maxX = smax(maxX, ln.b.X)
		//maxX = math.Max(maxX, ln.b.X)
		if !ln.Intersection(ptg).IsEmpty() {
			return true
		}
	}
	if sgt(pt.X, maxX) {
		//if pt.X > maxX {
		return false
	}

	ray, err := NewLine(Coordinates{pt}, Coordinates{XY{sadd(maxX, one), pt.Y}})
	if err != nil {
		panic(err)
	}
	var count int
	for _, seg := range ring.ls.lines {
		inter := seg.Intersection(ray)
		if inter.IsEmpty() {
			continue
		}
		if inter.Dimension() == 1 {
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
		if inter.Equals(ep1) || inter.Equals(ep2) {
			otherY := ep1.coords.Y
			if inter.Equals(ep1) {
				otherY = ep2.coords.Y
			}
			if slt(otherY, pt.Y) {
				//if otherY < pt.Y {
				count++
			}
		} else {
			count++
		}
	}
	return count%2 == 1
}
