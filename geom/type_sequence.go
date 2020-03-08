package geom

// Sequence represents a list of point locations.  It is immutable after
// creation.  All locations in the Sequence are specified using the same
// coordinates type.
//
// The zero value is an empty sequence of coordinate type DimXY.
type Sequence struct {
	ctype  CoordinatesType
	floats []float64
}

// NewSequence creates a new Sequence from a raw slice of float64 coordinates.
// The slice will be retained by the constructed Sequence and must NOT be
// modified by the caller after the function returns.
//
// The CoordiantesType indicates what type of coordinates the Sequence will
// store (i.e. just XY, XYZ, XYM, or XYZM).
//
// The coordinates in the passed slice should be interleaved. That is, all of
// the coordinates for the first point come first, followed by all of the
// coordinates for the second point etc. Coordinates must be in the order of X
// followed by Y, then Z (if using XYZ or XYZM), then M (if using XYM or XYZM).
//
// The length of the coordinates slice must be a multiple of the dimensionality
// of the coordiantes type. If the length is not a multiple, then this is a
// programming error and the function will panic.
func NewSequence(coordinates []float64, ctype CoordinatesType) Sequence {
	if len(coordinates)%ctype.Dimension() != 0 {
		panic("invalid coordinates length: inconsistent with CoordinatesType")
	}
	return Sequence{ctype, coordinates}
}

// CoordinatesType returns the coordinates type used to represent point
// locations in the Sequence.
func (s Sequence) CoordinatesType() CoordinatesType {
	return s.ctype
}

// Length returns the number of point locations represented by the Sequence.
func (s Sequence) Length() int {
	return len(s.floats) / s.ctype.Dimension()
}

// Get returns the Coordinates of the ith point location in the Sequence. It
// panics if i is out of range with respect to the number of points in the
// Sequence.
func (s Sequence) Get(i int) Coordinates {
	stride := s.ctype.Dimension()
	c := Coordinates{
		XY: XY{
			s.floats[i*stride],
			s.floats[i*stride+1],
		},
		Type: s.ctype,
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

// GetXY returns the XY of the ith point location in the Sequence. It panics if
// i is out of range with respect to the number of points in the Sequence.
func (s Sequence) GetXY(i int) XY {
	stride := s.ctype.Dimension()
	return XY{
		s.floats[i*stride],
		s.floats[i*stride+1],
	}
}

// Reverse returns a new Sequence containing the same point locations, but in
// reversed order.
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

// ForceCoordinatesType returns a new Sequence with a different CoordinatesType. If a
// dimension is added, then its new value is set to zero for each point
// location in the Sequence.
func (s Sequence) ForceCoordinatesType(newCType CoordinatesType) Sequence {
	if s.ctype == newCType {
		return s
	}

	stride := newCType.Dimension()
	flat := make([]float64, stride*s.Length())
	n := s.Length()
	for i := 0; i < n; i++ {
		c := s.Get(i)
		flat[stride*i+0] = c.X
		flat[stride*i+1] = c.Y
		switch newCType {
		case DimXYZ:
			flat[stride*i+2] = c.Z
		case DimXYM:
			flat[stride*i+2] = c.M
		case DimXYZM:
			flat[stride*i+2] = c.Z
			flat[stride*i+3] = c.M
		}
	}
	return Sequence{newCType, flat}
}

// Force2D returns a new Sequence with Z and M values removed (if present).
func (s Sequence) Force2D() Sequence {
	return s.ForceCoordinatesType(DimXY)
}

// getLine extracts a 2D line segment from a sequence by joining together
// adjacent locations in the sequence. It is designed to be called with i equal
// to each index in the sequence (from 0 to n-1). The flag indicates if the
// returned line is valid.
func getLine(seq Sequence, i int) (Line, bool) {
	if i == 0 {
		return Line{}, false
	}
	ln := Line{
		a: Coordinates{Type: DimXY, XY: seq.GetXY(i - 1)},
		b: Coordinates{Type: DimXY, XY: seq.GetXY(i)},
	}
	return ln, ln.a.XY != ln.b.XY
}
