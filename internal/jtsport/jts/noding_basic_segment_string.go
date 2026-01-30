package jts

var _ Noding_SegmentString = (*Noding_BasicSegmentString)(nil)

// Noding_BasicSegmentString represents a read-only list of contiguous line
// segments. This can be used for detection of intersections or nodes.
// SegmentStrings can carry a context object, which is useful for preserving
// topological or parentage information.
//
// If adding nodes is required use NodedSegmentString.
type Noding_BasicSegmentString struct {
	pts  []*Geom_Coordinate
	data any
}

func (ss *Noding_BasicSegmentString) IsNoding_SegmentString() {}

// Noding_NewBasicSegmentString creates a new segment string from a list of
// vertices.
func Noding_NewBasicSegmentString(pts []*Geom_Coordinate, data any) *Noding_BasicSegmentString {
	return &Noding_BasicSegmentString{
		pts:  pts,
		data: data,
	}
}

// GetData gets the user-defined data for this segment string.
func (ss *Noding_BasicSegmentString) GetData() any {
	return ss.data
}

// SetData sets the user-defined data for this segment string.
func (ss *Noding_BasicSegmentString) SetData(data any) {
	ss.data = data
}

// Size returns the number of coordinates in this segment string.
func (ss *Noding_BasicSegmentString) Size() int {
	return len(ss.pts)
}

// GetCoordinate gets the segment string coordinate at a given index.
func (ss *Noding_BasicSegmentString) GetCoordinate(i int) *Geom_Coordinate {
	return ss.pts[i]
}

// GetCoordinates gets the coordinates in this segment string.
func (ss *Noding_BasicSegmentString) GetCoordinates() []*Geom_Coordinate {
	return ss.pts
}

// IsClosed tests if a segment string is a closed ring.
func (ss *Noding_BasicSegmentString) IsClosed() bool {
	return ss.pts[0].Equals(ss.pts[len(ss.pts)-1])
}

// PrevInRing gets the previous vertex in a ring from a vertex index.
func (ss *Noding_BasicSegmentString) PrevInRing(index int) *Geom_Coordinate {
	prevIndex := index - 1
	if prevIndex < 0 {
		prevIndex = ss.Size() - 2
	}
	return ss.GetCoordinate(prevIndex)
}

// NextInRing gets the next vertex in a ring from a vertex index.
func (ss *Noding_BasicSegmentString) NextInRing(index int) *Geom_Coordinate {
	nextIndex := index + 1
	if nextIndex > ss.Size()-1 {
		nextIndex = 1
	}
	return ss.GetCoordinate(nextIndex)
}

// GetSegmentOctant gets the octant of the segment starting at vertex index.
func (ss *Noding_BasicSegmentString) GetSegmentOctant(index int) int {
	if index == len(ss.pts)-1 {
		return -1
	}
	return Noding_Octant_Octant(ss.GetCoordinate(index), ss.GetCoordinate(index+1))
}
