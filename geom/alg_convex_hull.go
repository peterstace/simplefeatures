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
	hull := grahamScan(pts)
	switch len(hull) {
	case 0:
		return GeometryCollection{}.AsGeometry()
	case 1:
		return NewPointFromXY(hull[0]).AsGeometry()
	case 2:
		if hull[0] == hull[1] {
			panic(fmt.Sprintf("bug in grahamScan routine - output 2 coincident points: %v", hull))
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
			panic(fmt.Errorf("bug in grahamScan routine - didn't produce a valid ring: %v", err))
		}
		poly, err := NewPolygonFromRings([]LineString{ring})
		if err != nil {
			// We could get an error here in the case where there are multiple
			// input points that are collinear within some numerical precision
			// bounds. The Graham scan thinks that the points are not
			// collinear, but polygon validations thinks that they are
			// collinear (i.e. not a valid Polygon). If we hit this case, we
			// need to give a LineString instead. This is a bit of a hack, but
			// would likely be fixed by
			// https://github.com/peterstace/simplefeatures/issues/246

			origin := hull[0]
			destin := hull[0]
			for _, pt := range hull[1:] {
				if origin.squareDistanceTo(pt) > origin.squareDistanceTo(destin) {
					destin = pt
				}
			}
			if origin == destin {
				panic(fmt.Sprintf("bug in collinear point fallback - output 2 coincident points: %v", origin))
			}
			ln := line{origin, destin}
			return ln.asLineString().AsGeometry()
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

// grahamScan returns the convex hull of the input points. It will either
// represent the empty set (zero points), a point (one point), a line (2
// points), or a closed polygon (>= 3 points).
func grahamScan(ps []XY) []XY {
	if len(ps) <= 1 {
		return ps
	}

	sortByPolarAngle(ps)

	// Append the lowest-then-leftmost point so that the polygon will be closed.
	ps = append(ps, ps[0])

	// Populate the stack with the first 2 distict points.
	var i int // tracks progress through the ps slice
	var stack pointStack
	stack.push(ps[0])
	i++
	for i < len(ps) && len(stack) < 2 {
		if stack.top() != ps[i] {
			stack.push(ps[i])
		}
		i++
	}

	for i < len(ps) {
		ori := orientation(stack.underTop(), stack.top(), ps[i])
		switch ori {
		case leftTurn:
			// This point _might_ be part of the convex hull. It will be popped
			// later if it actually isn't part of the convex hull.
			stack.push(ps[i])
		case collinear:
			// This point is part of the convex hull, so long as it extends the
			// current line segment (in which case the preceding point is
			// _not_ part of the convex hull).
			if distanceSq(stack.underTop(), ps[i]) > distanceSq(stack.underTop(), stack.top()) {
				stack.pop()
				stack.push(ps[i])
			}
		default:
			// The preceding point was _not_ part of the convex hull (so it is
			// popped). Potentially the new point reveals that other previous
			// points are also not part of the hull (so pop those as well).
			stack.pop()
			for orientation(stack.underTop(), stack.top(), ps[i]) != leftTurn {
				stack.pop()
			}
			stack.push(ps[i])
		}
		i++
	}
	return stack
}

// sortByPolarAngle sorts the points by their polar angle relative to the
// lowest-then-leftmost anchor point.
func sortByPolarAngle(ps []XY) {
	// the lowest-then-leftmost (anchor) point comes first
	ltlp := lowestThenLeftmost(ps)
	ps[ltlp], ps[0] = ps[0], ps[ltlp]
	anchor := ps[0]

	ps = ps[1:] // only sort the remaining points
	sort.Slice(ps, func(i, j int) bool {
		// If any point is equal to the anchor point, then always put it first.
		// This allows those duplicated points to be removed when the results
		// stack is initiated.
		if anchor == ps[i] {
			return true
		}
		if anchor == ps[j] {
			return false
		}

		// In the normal case, check which order the points are in relative to
		// the anchor.
		switch ori := orientation(anchor, ps[i], ps[j]); ori {
		case leftTurn:
			return true
		case rightTurn:
			return false
		case collinear:
			// When collinear, the point closest to the anchor comes first.
			// This is to prevent some cases where numerical inconsistencies
			// can occur. For example, calls to orientation(A,B,C) being
			// collinear but calls to orientation(A,C,B) NOT being collinear.
			ptI := anchor.Sub(ps[i])
			ptJ := anchor.Sub(ps[j])
			return ptI.Dot(ptI) < ptJ.Dot(ptJ)
		default:
			panic(ori)
		}
	})
}

// lowestThenLeftmost finds the index of the lowest-then-leftmost point.
func lowestThenLeftmost(ps []XY) int {
	rpi := 0
	for i := 1; i < len(ps); i++ {
		if ps[i].Y < ps[rpi].Y || (ps[i].Y == ps[rpi].Y && ps[i].X < ps[rpi].X) {
			rpi = i
		}
	}
	return rpi
}

// distanceSq gives the square of the distance between p and q.
func distanceSq(p, q XY) float64 {
	pSubQ := p.Sub(q)
	return pSubQ.Dot(pSubQ)
}
