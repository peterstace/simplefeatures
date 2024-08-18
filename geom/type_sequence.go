package geom

import "fmt"

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
// The CoordinatesType indicates what type of coordinates the Sequence will
// store (i.e. just XY, XYZ, XYM, or XYZM).
//
// The coordinates in the passed slice should be interleaved. That is, all of
// the coordinates for the first point come first, followed by all of the
// coordinates for the second point etc. Coordinates must be in the order of X
// followed by Y, then Z (if using XYZ or XYZM), then M (if using XYM or XYZM).
//
// The length of the coordinates slice must be a multiple of the dimensionality
// of the coordinates type. If the length is not a multiple, then this is a
// programming error and the function will panic.
func NewSequence(coordinates []float64, ctype CoordinatesType) Sequence {
	if len(coordinates)%ctype.Dimension() != 0 {
		panic("invalid coordinates length: inconsistent with CoordinatesType")
	}
	return Sequence{ctype, coordinates}
}

// validate checks the X and Y values in the sequence for NaNs and infinities.
func (s Sequence) validate() error {
	n := s.Length()
	for i := 0; i < n; i++ {
		if err := s.GetXY(i).validate(); err != nil {
			return wrap(err, "invalid XY at index %d", i)
		}
	}
	return nil
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

// Slice creates a new Sequence that is a subslice of this Sequence. Indexing
// rules work in the same way as Go Slices.
func (s Sequence) Slice(i, j int) Sequence {
	stride := s.ctype.Dimension()
	return Sequence{s.ctype, s.floats[i*stride : j*stride]}
}

// ForceCoordinatesType returns a new Sequence with a different CoordinatesType. If a
// dimension is added, then its new value is set to zero for each point
// location in the Sequence.
func (s Sequence) ForceCoordinatesType(newCType CoordinatesType) Sequence {
	if s.ctype == newCType {
		return s
	}
	if len(s.floats) == 0 {
		return Sequence{newCType, nil}
	}

	n := s.Length()
	stride := newCType.Dimension()
	flat := make([]float64, stride*n)
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

// appendAllPoints appends the float64 coordinates of all points (in order) to
// dst.
func (s Sequence) appendAllPoints(dst []float64) []float64 {
	return append(dst, s.floats...)
}

// appendPoint appends the float64 coordinates of the ith point to dst.
func (s Sequence) appendPoint(dst []float64, i int) []float64 {
	stride := s.ctype.Dimension()
	return append(dst, s.floats[i*stride:(i+1)*stride]...)
}

// assertNoUnusedCapacity panics if the backing slice contains any unused
// capacity.
func (s Sequence) assertNoUnusedCapacity() {
	if cap(s.floats)-len(s.floats) != 0 {
		panic(fmt.Sprintf("unused capacity assertion "+
			"failure: cap=%d len=%d", cap(s.floats), len(s.floats)))
	}
}

// less gives a lexicographical ordering between sequences, considering only
// the XY parts of each coordinate when they have Z or M components.
func (s Sequence) less(o Sequence) bool {
	oLen := o.Length()
	for i := 0; i < s.Length(); i++ {
		if i >= oLen {
			return true
		}
		sxy, oxy := s.GetXY(i), o.GetXY(i)
		if sxy != oxy {
			return sxy.Less(oxy)
		}
	}
	return false
}

// Envelope returns the axis aligned bounding box that most tightly surrounds
// the XY values in the sequence.
func (s Sequence) Envelope() Envelope {
	n := s.Length()
	if n == 0 {
		return Envelope{}
	}

	xy0 := s.GetXY(0)
	lower, upper := xy0, xy0

	stride := s.ctype.Dimension()
	for i := stride; i < len(s.floats); i += stride {
		x := s.floats[i]
		y := s.floats[i+1]
		lower.X = fastMin(lower.X, x)
		lower.Y = fastMin(lower.Y, y)
		upper.X = fastMax(upper.X, x)
		upper.Y = fastMax(upper.Y, y)
	}
	return newUncheckedEnvelope(lower, upper)
}

// getLine extracts a 2D line segment from a sequence by joining together
// adjacent locations in the sequence. It is designed to be called with i equal
// to each index in the sequence (from 0 to n-1, both inclusive). The flag
// indicates if the returned line is valid.
func getLine(seq Sequence, i int) (line, bool) {
	if i == 0 {
		return line{}, false
	}
	ln := line{
		a: seq.GetXY(i - 1),
		b: seq.GetXY(i),
	}
	return ln, ln.a != ln.b
}

// firstAndLastLines returns the index of the first and last line segments (if
// they exist) in the sequence.
func firstAndLastLines(seq Sequence) (int, int, bool) {
	n := seq.Length()
	first, last := -1, -1
	for i := 1; i < n; i++ {
		if seq.GetXY(i) != seq.GetXY(i-1) {
			first = i
			break
		}
	}
	for i := n - 1; i >= 1; i-- {
		if seq.GetXY(i) != seq.GetXY(i-1) {
			last = i
			break
		}
	}
	return first, last, first != -1 && last != -1
}

// previousLine finds the index of the line segment previous to line segment i.
// This may not be i-1 in the case where there are duplicate points. If there
// is no previous line, then false will be returned.
func previousLine(seq Sequence, i int) (int, bool) {
	i--
	for i >= 0 {
		if _, ok := getLine(seq, i); ok {
			return i, true
		}
		i--
	}
	return 0, false
}

// nextLine finds the index of the line segment after line segment i.  This may
// not be i+1 in the case where there are duplicate points. If there is no next
// line, then false will be returned.
func nextLine(seq Sequence, i int) (int, bool) {
	n := seq.Length()
	i++
	for i < n {
		if _, ok := getLine(seq, i); ok {
			return i, true
		}
		i++
	}
	return 0, false
}
