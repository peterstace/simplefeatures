package geom

type Sequence struct {
	ctype  CoordinatesType
	floats []float64
}

func NewSequence(coordinates []float64, ctype CoordinatesType) Sequence {
	tmp := make([]float64, len(coordinates))
	copy(tmp, coordinates)
	return NewSequenceNoCopy(tmp, ctype)
}

func NewSequenceNoCopy(coordinates []float64, ctype CoordinatesType) Sequence {
	if len(coordinates)%ctype.Dimension() != 0 {
		panic("invalid coordinates length: inconsistent with CoordinatesType")
	}
	return Sequence{ctype, coordinates}
}

func (s Sequence) CoordinatesType() CoordinatesType {
	return s.ctype
}

func (s Sequence) Length() int {
	return len(s.floats) / s.ctype.Dimension()
}

func (s Sequence) Get(i int) Coordinates {
	stride := s.ctype.Dimension()
	c := Coordinates{
		XY: XY{
			s.floats[i*stride],
			s.floats[i*stride+1],
		},
	}
	switch s.ctype {
	case XYZ:
		c.Z = s.floats[i*stride+2]
	case XYM:
		c.M = s.floats[i*stride+2]
	case XYZM:
		c.Z = s.floats[i*stride+2]
		c.M = s.floats[i*stride+3]
	}
	return c
}

func (s Sequence) GetXY(i int) XY {
	stride := s.ctype.Dimension()
	return XY{
		s.floats[i*stride],
		s.floats[i*stride+1],
	}
}

func (s Sequence) Reverse() Sequence {
	stride := s.ctype.Dimension()
	n := s.Length()
	reversed := make([]float64, len(s.floats))
	for i := 0; i < n; i++ {
		j := n - i - 1
		copy(
			reversed[i*stride:(i+1)*stride],
			s.floats[j*stride:(j+1)*stride],
		)
	}
	return Sequence{s.ctype, reversed}
}
