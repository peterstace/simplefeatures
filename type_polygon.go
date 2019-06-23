package simplefeatures

import (
	"errors"
)

// Polygon is a planar surface, defined by 1 exiterior boundary and 0 or more
// interior boundaries. Each interior boundary defines a hole in the polygon.
type Polygon struct {
	outer LinearRing
	holes []LinearRing
}

// NewPolygon creates a polygon given its outer and inner rings. No rings may
// cross each other, and can only intersect each with each other at a point.
func NewPolygon(outer LinearRing, holes ...LinearRing) (Polygon, error) {
	allRings := append(holes, outer)
	nextInterVert := len(allRings)
	interVerts := make(map[xyString]int)
	graph := newGraph()

	// Rings may intersect, but only at a single point.
	for i := 0; i < len(allRings); i++ {
		for j := i + 1; j < len(allRings); j++ {
			inter := allRings[i].Intersection(allRings[j])
			env, has := inter.Envelope()
			if !has {
				continue // no intersection
			}
			if !xyeq(env.Min(), env.Max()) {
				return Polygon{}, errors.New("polygon rings must not intersect at multiple points")
			}

			interVert, ok := interVerts[xykey(env.Min())]
			if !ok {
				interVert = nextInterVert
				nextInterVert++
				interVerts[xykey(env.Min())] = interVert
			}
			graph.addEdge(interVert, i)
			graph.addEdge(interVert, j)
		}
	}

	// All inner rings must be inside the outer ring.
	for _, hole := range holes {
		for _, line := range hole.ls.lines {
			if !isPointInsideOrOnRing(line.a.XY, outer) {
				return Polygon{}, errors.New("hole must be inside outer ring")
			}
		}
	}

	// Connectedness check: a graph is created where the intersections and
	// rings are modelled as vertices. Edges are added to the graph between an
	// intersection vertex and a ring vertex if the ring participates in that
	// intersection. The interior of the polygon is connected iff the graph
	// does not contain a cycle.
	if graph.hasCycle() {
		return Polygon{}, errors.New("polygon interiors must be connected")
	}

	return Polygon{outer: outer, holes: holes}, nil
}

func NewPolygonFromCoords(coords [][]Coordinates) (Polygon, error) {
	if len(coords) == 0 {
		return Polygon{}, errors.New("Polygon must have an outer ring")
	}
	outer, err := NewLinearRing(coords[0])
	if err != nil {
		return Polygon{}, err
	}
	var holes []LinearRing
	for _, holeCoords := range coords[1:] {
		hole, err := NewLinearRing(holeCoords)
		if err != nil {
			return Polygon{}, err
		}
		holes = append(holes, hole)
	}
	return NewPolygon(outer, holes...)
}

func (p Polygon) AsText() []byte {
	return p.AppendWKT(nil)
}

func (p Polygon) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("POLYGON")...)
	return p.appendWKTBody(dst)
}

func (p Polygon) appendWKTBody(dst []byte) []byte {
	dst = append(dst, '(')
	dst = p.outer.ls.appendWKTBody(dst)
	for _, h := range p.holes {
		dst = append(dst, ',')
		dst = h.ls.appendWKTBody(dst)
	}
	return append(dst, ')')
}

func (p Polygon) IsSimple() bool {
	panic("not implemented")
}

func (p Polygon) Intersection(g Geometry) Geometry {
	return intersection(p, g)
}

func (p Polygon) IsEmpty() bool {
	return false
}

func (p Polygon) Dimension() int {
	return 2
}

func (p Polygon) Equals(other Geometry) bool {
	return equals(p, other)
}

func (p Polygon) Envelope() (Envelope, bool) {
	return p.outer.Envelope()
}
