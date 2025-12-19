package jts

// Geom_CoordinateList is a list of Geom_Coordinates, which may be set to prevent
// repeated coordinates from occurring in the list.
type Geom_CoordinateList struct {
	coords []*Geom_Coordinate
}

// Geom_NewCoordinateList constructs a new list without any coordinates.
func Geom_NewCoordinateList() *Geom_CoordinateList {
	return &Geom_CoordinateList{
		coords: make([]*Geom_Coordinate, 0),
	}
}

// Geom_NewCoordinateListFromCoordinates constructs a new list from a slice of
// Geom_Coordinates, allowing repeated points. (I.e. this constructor produces a
// Geom_CoordinateList with exactly the same set of points as the input slice.)
func Geom_NewCoordinateListFromCoordinates(coords []*Geom_Coordinate) *Geom_CoordinateList {
	cl := &Geom_CoordinateList{
		coords: make([]*Geom_Coordinate, 0, len(coords)),
	}
	cl.AddCoordinates(coords, true)
	return cl
}

// Geom_NewCoordinateListFromCoordinatesAllowRepeated constructs a new list from a
// slice of Geom_Coordinates, allowing caller to specify if repeated points are to
// be removed.
func Geom_NewCoordinateListFromCoordinatesAllowRepeated(coords []*Geom_Coordinate, allowRepeated bool) *Geom_CoordinateList {
	cl := &Geom_CoordinateList{
		coords: make([]*Geom_Coordinate, 0, len(coords)),
	}
	cl.AddCoordinates(coords, allowRepeated)
	return cl
}

// Size returns the number of coordinates in the list.
func (cl *Geom_CoordinateList) Size() int {
	return len(cl.coords)
}

// Get returns the coordinate at the specified index.
func (cl *Geom_CoordinateList) Get(i int) *Geom_Coordinate {
	return cl.coords[i]
}

// GetCoordinate returns the coordinate at the specified index.
func (cl *Geom_CoordinateList) GetCoordinate(i int) *Geom_Coordinate {
	return cl.coords[i]
}

// Set replaces the coordinate at the specified index.
func (cl *Geom_CoordinateList) Set(i int, coord *Geom_Coordinate) {
	cl.coords[i] = coord
}

// AddCoordinatesRange adds a section of a slice of coordinates to the list.
func (cl *Geom_CoordinateList) AddCoordinatesRange(coords []*Geom_Coordinate, allowRepeated bool, start, end int) bool {
	inc := 1
	if start > end {
		inc = -1
	}

	for i := start; i != end; i += inc {
		cl.AddCoordinate(coords[i], allowRepeated)
	}
	return true
}

// AddCoordinatesWithDirection adds a slice of coordinates to the list.
func (cl *Geom_CoordinateList) AddCoordinatesWithDirection(coords []*Geom_Coordinate, allowRepeated bool, direction bool) bool {
	if direction {
		for i := 0; i < len(coords); i++ {
			cl.AddCoordinate(coords[i], allowRepeated)
		}
	} else {
		for i := len(coords) - 1; i >= 0; i-- {
			cl.AddCoordinate(coords[i], allowRepeated)
		}
	}
	return true
}

// AddCoordinates adds a slice of coordinates to the list.
func (cl *Geom_CoordinateList) AddCoordinates(coords []*Geom_Coordinate, allowRepeated bool) bool {
	cl.AddCoordinatesWithDirection(coords, allowRepeated, true)
	return true
}

// AddCoordinate adds a coordinate to the end of the list.
func (cl *Geom_CoordinateList) AddCoordinate(coord *Geom_Coordinate, allowRepeated bool) {
	if !allowRepeated {
		if len(cl.coords) >= 1 {
			last := cl.coords[len(cl.coords)-1]
			if last.Equals2D(coord) {
				return
			}
		}
	}
	cl.coords = append(cl.coords, coord)
}

// AddCoordinateAtIndex inserts the specified coordinate at the specified
// position in this list.
func (cl *Geom_CoordinateList) AddCoordinateAtIndex(i int, coord *Geom_Coordinate, allowRepeated bool) {
	if !allowRepeated {
		size := len(cl.coords)
		if size > 0 {
			if i > 0 {
				prev := cl.coords[i-1]
				if prev.Equals2D(coord) {
					return
				}
			}
			if i < size {
				next := cl.coords[i]
				if next.Equals2D(coord) {
					return
				}
			}
		}
	}
	// Insert at position i.
	cl.coords = append(cl.coords[:i], append([]*Geom_Coordinate{coord}, cl.coords[i:]...)...)
}

// AddAll adds all coordinates from a slice to the list.
func (cl *Geom_CoordinateList) AddAll(coll []*Geom_Coordinate, allowRepeated bool) bool {
	isChanged := false
	for _, coord := range coll {
		cl.AddCoordinate(coord, allowRepeated)
		isChanged = true
	}
	return isChanged
}

// CloseRing ensures this Geom_CoordinateList is a ring, by adding the start point
// if necessary.
func (cl *Geom_CoordinateList) CloseRing() {
	if len(cl.coords) > 0 {
		duplicate := cl.coords[0].Copy()
		cl.AddCoordinate(duplicate, false)
	}
}

// ToCoordinateArray returns the Geom_Coordinates in this collection as a slice.
func (cl *Geom_CoordinateList) ToCoordinateArray() []*Geom_Coordinate {
	result := make([]*Geom_Coordinate, len(cl.coords))
	copy(result, cl.coords)
	return result
}

// ToCoordinateArrayWithDirection creates a slice containing the coordinates in
// this list, oriented in the given direction (forward or reverse).
func (cl *Geom_CoordinateList) ToCoordinateArrayWithDirection(isForward bool) []*Geom_Coordinate {
	if isForward {
		result := make([]*Geom_Coordinate, len(cl.coords))
		copy(result, cl.coords)
		return result
	}
	// Construct reversed slice.
	size := len(cl.coords)
	pts := make([]*Geom_Coordinate, size)
	for i := 0; i < size; i++ {
		pts[i] = cl.coords[size-i-1]
	}
	return pts
}

// Clone returns a deep copy of this Geom_CoordinateList instance.
func (cl *Geom_CoordinateList) Clone() *Geom_CoordinateList {
	clone := &Geom_CoordinateList{
		coords: make([]*Geom_Coordinate, len(cl.coords)),
	}
	for i := 0; i < len(cl.coords); i++ {
		clone.coords[i] = cl.coords[i].Copy()
	}
	return clone
}
