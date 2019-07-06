package simplefeatures

import (
	"errors"
)

// Polygon is a planar surface, defined by 1 exiterior boundary and 0 or more
// interior boundaries. Each interior boundary defines a hole in the polygon.
//
// Its assertions are:
//
// 1. The out ring and holes must be valid LinearRings.
//
// 2. Each pair of rings must only intersect at a single point.
//
// 3. The interior of the polygon is connected.
//
// 4. The holes must be fully inside the outer ring.
//
type Polygon struct {
	outer LinearRing
	holes []LinearRing
}

// NewPolygon creates a polygon given its outer and inner rings. No rings may
// cross each other, and can only intersect each with each other at a point.
func NewPolygon(outer LinearRing, holes ...LinearRing) (Polygon, error) {
	allRings := append(holes, outer)
	nextInterVert := len(allRings)
	interVerts := make(map[xyHash]int)
	graph := newGraph()

	// Rings may intersect, but only at a single point.
	for i := 0; i < len(allRings); i++ {
		for j := i + 1; j < len(allRings); j++ {
			inter := allRings[i].Intersection(allRings[j])
			env, has := inter.Envelope()
			if !has {
				continue // no intersection
			}
			if !env.Min().Equals(env.Max()) {
				return Polygon{}, errors.New("polygon rings must not intersect at multiple points")
			}

			interVert, ok := interVerts[env.Min().hash()]
			if !ok {
				interVert = nextInterVert
				nextInterVert++
				interVerts[env.Min().hash()] = interVert
			}
			graph.addEdge(interVert, i)
			graph.addEdge(interVert, j)
		}
	}

	// All inner rings must be inside the outer ring.
	for _, hole := range holes {
		for _, line := range hole.ls.lines {
			if pointRingSide(line.a.XY, outer) == exterior {
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

func (p Polygon) AsText() string {
	return string(p.AppendWKT(nil))
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

// IsSimple always returns true. All Polygons are simple.
func (p Polygon) IsSimple() bool {
	return true
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

func (p Polygon) rings() []LinearRing {
	rings := make([]LinearRing, 1+len(p.holes))
	rings[0] = p.outer
	for i, h := range p.holes {
		rings[1+i] = h
	}
	return rings
}

func (p Polygon) Boundary() Geometry {
	if len(p.holes) == 0 {
		return p.outer.ls
	}
	bounds := make([]LineString, 1+len(p.holes))
	bounds[0] = p.outer.ls
	for i, h := range p.holes {
		bounds[1+i] = h.ls
	}
	return NewMultiLineString(bounds)
}
