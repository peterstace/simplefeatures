package simplefeatures

import (
	"errors"
	"strconv"
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
	// TODO: No rings may cross.
	// TODO: Rings may intersect, but only at a point (and only as a tangent).
	// TODO: check linear ring directions?
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
	ring := func(r LinearRing) {
		dst = append(dst, '(')
		for i, pt := range r.ls.pts {
			dst = strconv.AppendFloat(dst, pt.X, 'f', -1, 64)
			dst = append(dst, ' ')
			dst = strconv.AppendFloat(dst, pt.Y, 'f', -1, 64)
			if i != len(r.ls.pts)-1 {
				dst = append(dst, ',')
			}
		}
		dst = append(dst, ')')
	}
	ring(p.outer)
	for _, h := range p.holes {
		dst = append(dst, ',')
		ring(h)
	}
	return append(dst, ')')
}

func (p Polygon) IsSimple() bool {
	panic("not implemented")
}

func (p Polygon) Intersection(Geometry) Geometry {
	panic("not implemented")
}

func (p Polygon) IsEmpty() bool {
	return false
}

func (p Polygon) Dimension() int {
	return 2
}
