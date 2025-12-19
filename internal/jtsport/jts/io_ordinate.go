package jts

// Io_Ordinate represents a Well-Known-Text or Well-Known-Binary ordinate.
type Io_Ordinate int

const (
	Io_Ordinate_X Io_Ordinate = iota + 1
	Io_Ordinate_Y
	Io_Ordinate_Z
	Io_Ordinate_M
)

// Io_OrdinateSet represents a set of ordinates.
// Intended to be used similarly to Java's EnumSet<Ordinate>.
type Io_OrdinateSet struct {
	hasX bool
	hasY bool
	hasZ bool
	hasM bool
}

var io_ordinate_xy = &Io_OrdinateSet{hasX: true, hasY: true}
var io_ordinate_xyz = &Io_OrdinateSet{hasX: true, hasY: true, hasZ: true}
var io_ordinate_xym = &Io_OrdinateSet{hasX: true, hasY: true, hasM: true}
var io_ordinate_xyzm = &Io_OrdinateSet{hasX: true, hasY: true, hasZ: true, hasM: true}

// Io_Ordinate_CreateXY returns an OrdinateSet with X and Y ordinates.
// A copy is returned as Go doesn't have immutable collections.
func Io_Ordinate_CreateXY() *Io_OrdinateSet {
	return io_ordinate_xy.Clone()
}

// Io_Ordinate_CreateXYZ returns an OrdinateSet with X, Y, and Z ordinates.
// A copy is returned as Go doesn't have immutable collections.
func Io_Ordinate_CreateXYZ() *Io_OrdinateSet {
	return io_ordinate_xyz.Clone()
}

// Io_Ordinate_CreateXYM returns an OrdinateSet with X, Y, and M ordinates.
// A copy is returned as Go doesn't have immutable collections.
func Io_Ordinate_CreateXYM() *Io_OrdinateSet {
	return io_ordinate_xym.Clone()
}

// Io_Ordinate_CreateXYZM returns an OrdinateSet with X, Y, Z, and M ordinates.
// A copy is returned as Go doesn't have immutable collections.
func Io_Ordinate_CreateXYZM() *Io_OrdinateSet {
	return io_ordinate_xyzm.Clone()
}

// Contains returns true if the set contains the given ordinate.
func (s *Io_OrdinateSet) Contains(o Io_Ordinate) bool {
	switch o {
	case Io_Ordinate_X:
		return s.hasX
	case Io_Ordinate_Y:
		return s.hasY
	case Io_Ordinate_Z:
		return s.hasZ
	case Io_Ordinate_M:
		return s.hasM
	default:
		return false
	}
}

// Add adds an ordinate to the set.
func (s *Io_OrdinateSet) Add(o Io_Ordinate) {
	switch o {
	case Io_Ordinate_X:
		s.hasX = true
	case Io_Ordinate_Y:
		s.hasY = true
	case Io_Ordinate_Z:
		s.hasZ = true
	case Io_Ordinate_M:
		s.hasM = true
	}
}

// Size returns the number of ordinates in the set.
func (s *Io_OrdinateSet) Size() int {
	count := 0
	if s.hasX {
		count++
	}
	if s.hasY {
		count++
	}
	if s.hasZ {
		count++
	}
	if s.hasM {
		count++
	}
	return count
}

// Clone returns a copy of the ordinate set.
func (s *Io_OrdinateSet) Clone() *Io_OrdinateSet {
	return &Io_OrdinateSet{
		hasX: s.hasX,
		hasY: s.hasY,
		hasZ: s.hasZ,
		hasM: s.hasM,
	}
}

// Remove removes an ordinate from the set.
func (s *Io_OrdinateSet) Remove(o Io_Ordinate) {
	switch o {
	case Io_Ordinate_X:
		s.hasX = false
	case Io_Ordinate_Y:
		s.hasY = false
	case Io_Ordinate_Z:
		s.hasZ = false
	case Io_Ordinate_M:
		s.hasM = false
	}
}

// Equals returns true if the two ordinate sets are equal.
func (s *Io_OrdinateSet) Equals(other *Io_OrdinateSet) bool {
	return s.hasX == other.hasX &&
		s.hasY == other.hasY &&
		s.hasZ == other.hasZ &&
		s.hasM == other.hasM
}
