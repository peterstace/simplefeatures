package simplefeatures

import "sort"

// canonicalise converts a list of Geometries into a canonical form. This is is
// a workaround for an equality check, until Equals is implemented.
//
// TODO: once Equals is implemented, remove this code.
func canonicalise(geoms []Geometry) Geometry {
	pts, lns, polys := flattenGeometries(geoms)
	if len(pts) > 0 && len(lns) == 0 && len(polys) == 0 {
		if len(pts) == 1 {
			return pts[0]
		}
		return NewMultiPoint(pts)
	}
	if len(pts) == 0 && len(lns) > 0 && len(polys) == 0 {
		if len(lns) == 1 {
			return lns[0]
		}
	}
	if len(pts) == 0 && len(lns) == 0 && len(polys) > 0 {
		if len(polys) == 1 {
			return polys[0]
		}
		mp, err := NewMultiPolygon(polys)
		if err != nil {
			panic(err)
		}
		return mp
	}
	var allGeoms []Geometry
	for _, pt := range pts {
		allGeoms = append(allGeoms, pt)
	}
	for _, ln := range lns {
		allGeoms = append(allGeoms, ln)
	}
	for _, poly := range polys {
		allGeoms = append(allGeoms, poly)
	}
	return NewGeometryCollection(allGeoms)
}

func flattenGeometries(geoms []Geometry) ([]Point, []Line, []Polygon) {
	type xyxy struct {
		a, b XY
	}
	points := map[XY]Point{}
	lines := map[xyxy]Line{}
	var polys []Polygon

	for _, g := range geoms {
		switch g := g.(type) {
		case EmptySet:
		case Point:
			points[g.coords.XY] = g
		case Line:
			g = orderLine(g)
			lines[xyxy{g.a.XY, g.b.XY}] = g
		case LineString:
			for _, g := range g.lines {
				g = orderLine(g)
				lines[xyxy{g.a.XY, g.b.XY}] = g
			}
		case LinearRing:
			for _, g := range g.ls.lines {
				g = orderLine(g)
				lines[xyxy{g.a.XY, g.b.XY}] = g
			}
		case Polygon:
			polys = append(polys, g)
		case MultiPoint:
			for _, pt := range g.pts {
				points[pt.coords.XY] = pt
			}
		case MultiLineString:
			for _, linestr := range g.lines {
				for _, g := range linestr.lines {
					g = orderLine(g)
					lines[xyxy{g.a.XY, g.b.XY}] = g
				}
			}
		case MultiPolygon:
			for _, poly := range g.polys {
				polys = append(polys, poly)
			}
		case GeometryCollection:
			pts, lns, pls := flattenGeometries(g.geoms)
			for _, pt := range pts {
				points[pt.coords.XY] = pt
			}
			for _, ln := range lns {
				lines[xyxy{ln.a.XY, ln.b.XY}] = ln
			}
			polys = append(polys, pls...)
		default:
			panic("unknown type")
		}
	}

	for pt := range points {
		for _, line := range lines {
			if xyeq(line.a.XY, pt) || xyeq(line.b.XY, pt) {
				delete(points, pt)
				break
			}
		}
	}

	var pointsSlice []Point
	for _, pt := range points {
		pointsSlice = append(pointsSlice, pt)
	}
	sort.Slice(pointsSlice, func(i, j int) bool {
		return xyCmp(
			pointsSlice[i].coords.XY,
			pointsSlice[j].coords.XY,
		) < 0
	})
	var linesSlice []Line
	for _, line := range lines {
		linesSlice = append(linesSlice, line)
	}
	sort.Slice(linesSlice, func(i, j int) bool {
		return lineCmp(linesSlice[i], linesSlice[j]) < 0
	})

	return pointsSlice, linesSlice, polys
}

func xyCmp(a, b XY) int {
	if slt(a.X, b.X) {
		//if a.X < b.X {
		return -1
	} else if sgt(a.X, b.X) {
		//} else if a.X > b.X {
		return +1
	}
	if slt(a.Y, b.Y) {
		//if a.Y < b.Y {
		return -1
	} else if sgt(a.Y, b.Y) {
		//} else if a.Y > b.Y {
		return +1
	}
	return 0
}

func lineCmp(a, b Line) int {
	if cmp := xyCmp(a.a.XY, b.a.XY); cmp != 0 {
		return cmp
	}
	return xyCmp(a.b.XY, b.b.XY)
}

func orderLine(line Line) Line {
	if xyCmp(line.a.XY, line.b.XY) > 0 {
		newLine, err := NewLine(line.b, line.a)
		if err != nil {
			panic(err)
		}
		return newLine
	}
	return line
}
