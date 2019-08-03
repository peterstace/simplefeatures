package geom

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
		return must(NewMultiPolygon(polys)).(MultiPolygon)
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
	points := map[xyHash]Point{}
	lines := map[xyxyHash]Line{}
	var polys []Polygon

	for _, g := range geoms {
		switch g := g.(type) {
		case EmptySet:
		case Point:
			points[g.coords.XY.hash()] = g
		case Line:
			g = orderLine(g)
			lines[hashXYXY(g.a.XY, g.b.XY)] = g
		case LineString:
			for _, g := range g.lines {
				g = orderLine(g)
				lines[hashXYXY(g.a.XY, g.b.XY)] = g
			}
		case LinearRing:
			for _, g := range g.ls.lines {
				g = orderLine(g)
				lines[hashXYXY(g.a.XY, g.b.XY)] = g
			}
		case Polygon:
			polys = append(polys, g)
		case MultiPoint:
			for _, pt := range g.pts {
				points[pt.coords.XY.hash()] = pt
			}
		case MultiLineString:
			for _, linestr := range g.lines {
				for _, g := range linestr.lines {
					g = orderLine(g)
					lines[hashXYXY(g.a.XY, g.b.XY)] = g
				}
			}
		case MultiPolygon:
			for _, poly := range g.polys {
				polys = append(polys, poly)
			}
		case GeometryCollection:
			pts, lns, pls := flattenGeometries(g.geoms)
			for _, pt := range pts {
				points[pt.coords.XY.hash()] = pt
			}
			for _, ln := range lns {
				lines[hashXYXY(ln.a.XY, ln.b.XY)] = ln
			}
			polys = append(polys, pls...)
		default:
			panic("unknown type")
		}
	}

	for _, pt := range points {
		for _, line := range lines {
			if line.a.XY.Equals(pt.coords.XY) || line.b.XY.Equals(pt.coords.XY) {
				delete(points, pt.coords.XY.hash())
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
	if a.X.LT(b.X) {
		return -1
	} else if a.X.GT(b.X) {
		return +1
	}
	if a.Y.LT(b.Y) {
		return -1
	} else if a.Y.GT(b.Y) {
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
		return must(NewLineC(line.b, line.a)).(Line)
	}
	return line
}
