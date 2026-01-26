package jts

// Noding_SegmentString is an interface for classes which represent a sequence
// of contiguous line segments. SegmentStrings can carry a context object,
// which is useful for preserving topological or parentage information.
type Noding_SegmentString interface {
	IsNoding_SegmentString()

	// GetData gets the user-defined data for this segment string.
	GetData() any

	// SetData sets the user-defined data for this segment string.
	SetData(data any)

	// Size gets the number of coordinates in this segment string.
	Size() int

	// GetCoordinate gets the segment string coordinate at a given index.
	GetCoordinate(i int) *Geom_Coordinate

	// GetCoordinates gets the coordinates in this segment string.
	GetCoordinates() []*Geom_Coordinate

	// IsClosed tests if a segment string is a closed ring.
	IsClosed() bool

	// PrevInRing gets the previous vertex in a ring from a vertex index.
	PrevInRing(index int) *Geom_Coordinate

	// NextInRing gets the next vertex in a ring from a vertex index.
	NextInRing(index int) *Geom_Coordinate
}
