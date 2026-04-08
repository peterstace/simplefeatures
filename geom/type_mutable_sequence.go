package geom

import "fmt"

// mutableSequence is a variable-length sequence of coordinates with
// homogeneous [CoordinatesType]. It is the mutable counterpart of [Sequence],
// using the same flat []float64 storage. It behaves like a "fat slice" — a
// []float64 plus coordinate type information. Like a Go slice, mutating
// operations return the updated value.
type mutableSequence struct {
	floats []float64
	ctype  CoordinatesType
}

func newMutableSequence(ctype CoordinatesType) mutableSequence {
	return mutableSequence{ctype: ctype}
}

func sequenceToMutable(seq Sequence) mutableSequence {
	floats := make([]float64, len(seq.floats))
	copy(floats, seq.floats)
	return mutableSequence{floats: floats, ctype: seq.CoordinatesType()}
}

func (s mutableSequence) Append(c Coordinates) mutableSequence {
	c.Type = s.ctype
	s.floats = c.appendFloat64s(s.floats)
	return s
}

func (s mutableSequence) AppendMutable(other mutableSequence) mutableSequence {
	if s.ctype != other.ctype {
		panic(fmt.Sprintf("AppendMutable: mismatched CoordinatesType: %s vs %s", s.ctype, other.ctype))
	}
	s.floats = append(s.floats, other.floats...)
	return s
}

func (s mutableSequence) Length() int {
	return len(s.floats) / s.ctype.Dimension()
}

func (s mutableSequence) Get(i int) Coordinates {
	dim := s.ctype.Dimension()
	offset := i * dim
	c := Coordinates{
		XY:   XY{X: s.floats[offset], Y: s.floats[offset+1]},
		Type: s.ctype,
	}
	if s.ctype.Is3D() {
		c.Z = s.floats[offset+2]
	}
	if s.ctype.IsMeasured() {
		c.M = s.floats[offset+dim-1]
	}
	return c
}

func (s mutableSequence) GetXY(i int) XY {
	dim := s.ctype.Dimension()
	offset := i * dim
	return XY{X: s.floats[offset], Y: s.floats[offset+1]}
}

func (s mutableSequence) Slice(i, j int) mutableSequence {
	dim := s.ctype.Dimension()
	s.floats = s.floats[i*dim : j*dim]
	return s
}

func (s mutableSequence) ToSequence() Sequence {
	return NewSequence(s.floats, s.ctype)
}
