package geom

import (
	"fmt"
	"sort"
)

func convexHull(g Geometry) Geometry {
	if g.IsEmpty() {
		// Any empty geometry could be returned here to give correct behaviour.
		// However, to replicate PostGIS behaviour, we always return the
		// original geometry.
		return g.Force2D()
	}

	pts := convexHullPointSet(g)

	// Check for point case:
	if !hasAtLeast2DistinctPointsInXYs(pts) {
		return pts[0].AsPoint().AsGeometry()
	}

	hull := monotoneChain(pts)

	// Check for linear case:
	if half, ok := isLinearHull(hull); ok {
		// Since the result is a straight line, we can simplify it by ignoring
		// any interior points.
		return line{half[0], half[len(half)-1]}.asLineString().AsGeometry()
	}

	// Convert hull to polygon.
	floats := make([]float64, 2*len(hull))
	for i := range hull {
		floats[2*i+0] = hull[i].X
		floats[2*i+1] = hull[i].Y
	}
	seq := NewSequence(floats, DimXY)
	ring := NewLineString(seq)
	poly := NewPolygon([]LineString{ring})
	if err := poly.Validate(); err != nil {
		panic(fmt.Errorf("bug in monotoneChain routine - didn't produce a valid polygon: %w", err))
	}
	return poly.AsGeometry()
}

func hasAtLeast2DistinctPointsInXYs(pts []XY) bool {
	if len(pts) <= 1 {
		return false
	}
	for _, pt := range pts[1:] {
		if pt != pts[0] {
			return true
		}
	}
	return false
}

// isLinearHull checks if a hull generated by monotoneChain is entirely linear
// (i.e. it doesn't enclose any area, and represents a straight line).
func isLinearHull(hull []XY) ([]XY, bool) {
	// In the linear case, the upper and lower hulls will be the same. We can
	// just check the points just before and after the middle point in the
	// hull. In the case where the upper and lower hulls are the same, these
	// points will also be the same.
	if len(hull)%2 == 0 {
		return nil, false
	}
	i := len(hull) / 2
	if hull[i-1] != hull[i+1] {
		return nil, false
	}
	return hull[:i+1], true
}

func convexHullPointSet(g Geometry) []XY {
	switch {
	case g.IsGeometryCollection():
		var points []XY
		c := g.MustAsGeometryCollection()
		n := c.NumGeometries()
		for i := 0; i < n; i++ {
			points = append(
				points,
				convexHullPointSet(c.GeometryN(i))...,
			)
		}
		return points
	case g.IsPoint():
		xy, ok := g.MustAsPoint().XY()
		if !ok {
			return nil
		}
		return []XY{xy}
	case g.IsLineString():
		cs := g.MustAsLineString().Coordinates()
		n := cs.Length()
		points := make([]XY, n)
		for i := 0; i < n; i++ {
			points[i] = cs.GetXY(i)
		}
		return points
	case g.IsPolygon():
		p := g.MustAsPolygon()
		return convexHullPointSet(p.ExteriorRing().AsGeometry())
	case g.IsMultiPoint():
		m := g.MustAsMultiPoint()
		n := m.NumPoints()
		points := make([]XY, 0, n)
		for i := 0; i < n; i++ {
			xy, ok := m.PointN(i).XY()
			if ok {
				points = append(points, xy)
			}
		}
		return points
	case g.IsMultiLineString():
		m := g.MustAsMultiLineString()
		var points []XY
		n := m.NumLineStrings()
		for i := 0; i < n; i++ {
			cs := m.LineStringN(i).Coordinates()
			m := cs.Length()
			for j := 0; j < m; j++ {
				points = append(points, cs.GetXY(j))
			}
		}
		return points
	case g.IsMultiPolygon():
		m := g.MustAsMultiPolygon()
		var points []XY
		numPolys := m.NumPolygons()
		for i := 0; i < numPolys; i++ {
			cs := m.PolygonN(i).ExteriorRing().Coordinates()
			m := cs.Length()
			for j := 0; j < m; j++ {
				points = append(points, cs.GetXY(j))
			}
		}
		return points
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

func monotoneChain(pts []XY) []XY {
	sort.Slice(pts, func(i, j int) bool {
		return pts[i].Less(pts[j])
	})

	// Calculate lower hull.
	var lower []XY
	for _, p := range pts {
		for len(lower) >= 2 && orientation(lower[len(lower)-2], lower[len(lower)-1], p) != leftTurn {
			lower = lower[:len(lower)-1]
		}
		lower = append(lower, p)
	}

	// Calculate upper hull.
	var upper []XY
	for i := len(pts) - 1; i >= 0; i-- {
		for len(upper) >= 2 && orientation(upper[len(upper)-2], upper[len(upper)-1], pts[i]) != leftTurn {
			upper = upper[:len(upper)-1]
		}
		upper = append(upper, pts[i])
	}

	// Join the upper and lower hulls, ignoring the first point in the upper
	// hull since it will be same as the last point in the lower hull.
	return append(lower, upper[1:]...)
}
