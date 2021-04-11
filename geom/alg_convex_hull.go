package geom

import (
	"fmt"
	"sort"
)

func convexHull(g Geometry) Geometry {
	if g.IsEmpty() {
		// Any empty geometry could be returned here to to give correct
		// behaviour. However, to replicate PostGIS behaviour, we always return
		// the original geometry.
		return g.Force2D()
	}
	pts := convexHullPointSet(g)

	hull := monotoneChain(pts)
	switch len(hull) {
	case 0:
		return GeometryCollection{}.AsGeometry()
	case 1:
		return NewPointFromXY(hull[0]).AsGeometry()
	case 2:
		if hull[0] == hull[1] {
			panic(fmt.Sprintf("bug in monotoneChain routine - output 2 coincident points: %v", hull))
		}
		ln := line{hull[0], hull[1]}
		return ln.asLineString().AsGeometry()
	default:
		floats := make([]float64, 2*len(hull))
		for i := range hull {
			floats[2*i+0] = hull[i].X
			floats[2*i+1] = hull[i].Y
		}
		seq := NewSequence(floats, DimXY)
		ring, err := NewLineString(seq)
		if err != nil {
			panic(fmt.Errorf("bug in monotoneChain routine - didn't produce a valid ring: %v", err))
		}
		poly, err := NewPolygonFromRings([]LineString{ring})
		if err != nil {
			panic(fmt.Errorf("bug in monotoneChain routine - didn't produce a valid polygon: %v", err))
		}
		return poly.AsGeometry()
	}
}

func convexHullPointSet(g Geometry) []XY {
	switch {
	case g.IsGeometryCollection():
		var points []XY
		c := g.AsGeometryCollection()
		n := c.NumGeometries()
		for i := 0; i < n; i++ {
			points = append(
				points,
				convexHullPointSet(c.GeometryN(i))...,
			)
		}
		return points
	case g.IsPoint():
		xy, ok := g.AsPoint().XY()
		if !ok {
			return nil
		}
		return []XY{xy}
	case g.IsLineString():
		cs := g.AsLineString().Coordinates()
		n := cs.Length()
		points := make([]XY, n)
		for i := 0; i < n; i++ {
			points[i] = cs.GetXY(i)
		}
		return points
	case g.IsPolygon():
		p := g.AsPolygon()
		return convexHullPointSet(p.ExteriorRing().AsGeometry())
	case g.IsMultiPoint():
		m := g.AsMultiPoint()
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
		m := g.AsMultiLineString()
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
		m := g.AsMultiPolygon()
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

func monotoneChain(ps []XY) []XY {
	// TODO precondition: there must be at least 3 points

	sort.Slice(ps, func(i, j int) bool {
		if ps[i].X != ps[j].X {
			return ps[i].X < ps[j].X
		}
		return ps[i].Y < ps[j].Y
	})

	var U, L []XY

	for _, p := range ps {
		for len(L) >= 2 && orientation(L[len(L)-2], L[len(L)-1], p) != leftTurn {
			L = L[:len(L)-1]
		}
		L = append(L, p)
	}

	for i := len(ps) - 1; i >= 0; i-- {
		for len(U) >= 2 && orientation(U[len(U)-2], U[len(U)-1], ps[i]) != leftTurn {
			U = U[:len(U)-1]
		}
		U = append(U, ps[i])
	}

	if len(L) > 0 {
		L = L[:len(L)-1]
	}
	if len(U) > 0 {
		U = U[:len(U)-1]
	}

	return append(L, U...)
}
