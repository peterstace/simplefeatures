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
	case DimXYZ:
		c.Z = s.floats[i*stride+2]
	case DimXYM:
		c.M = s.floats[i*stride+2]
	case DimXYZM:
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

func (s Sequence) Force2D() Sequence {
	// TODO: We could avoid all of this copying by storing both the coordinate
	// type and the stride independently within a sequence. Then all we would
	// need to do is return a shallow copy of the Sequence, but with just the
	// coordinate type change to DimXY (the stride would remain the same).
	flat := make([]float64, 2*s.Length())
	n := s.Length()
	for i := 0; i < n; i++ {
		xy := s.GetXY(i)
		flat[2*i+0] = xy.X
		flat[2*i+1] = xy.Y
	}
	return Sequence{DimXY, flat}
}

// getLine extracts a line segment from a sequence by joining together adjacent
// locations in the sequence. It is designed to be called with i equal to each
// index in the sequence (from 0 to n-1). The flag indicates if the returned
// line is valid.
func getLine(seq Sequence, i int) (Line, bool) {
	if i == 0 {
		return Line{}, false
	}
	ln := Line{
		a: Coordinates{XY: seq.GetXY(i - 1)},
		b: Coordinates{XY: seq.GetXY(i)},
	}
	return ln, ln.a.XY != ln.b.XY
}
