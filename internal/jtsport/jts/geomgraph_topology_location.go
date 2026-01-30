package jts

import (
	"strings"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geomgraph_TopologyLocation is the labelling of a GraphComponent's
// topological relationship to a single Geometry.
//
// If the parent component is an area edge, each side and the edge itself have
// a topological location. These locations are named:
//   - ON: on the edge
//   - LEFT: left-hand side of the edge
//   - RIGHT: right-hand side
//
// If the parent component is a line edge or node, there is a single
// topological relationship attribute, ON.
//
// The possible values of a topological location are {Geom_Location_None,
// Geom_Location_Exterior, Geom_Location_Boundary, Geom_Location_Interior}.
//
// The labelling is stored in an array location[j] where j has the values ON,
// LEFT, RIGHT.
type Geomgraph_TopologyLocation struct {
	child    java.Polymorphic
	location []int
}

// GetChild returns the immediate child in the type hierarchy chain.
func (tl *Geomgraph_TopologyLocation) GetChild() java.Polymorphic {
	return tl.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (tl *Geomgraph_TopologyLocation) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewTopologyLocationFromArray constructs a TopologyLocation from an
// array of location values.
func Geomgraph_NewTopologyLocationFromArray(location []int) *Geomgraph_TopologyLocation {
	tl := &Geomgraph_TopologyLocation{}
	tl.init(len(location))
	return tl
}

// Geomgraph_NewTopologyLocationOnLeftRight constructs a TopologyLocation
// specifying how points on, to the left of, and to the right of some
// GraphComponent relate to some Geometry. Possible values for the parameters
// are Geom_Location_None, Geom_Location_Exterior, Geom_Location_Boundary, and
// Geom_Location_Interior.
func Geomgraph_NewTopologyLocationOnLeftRight(on, left, right int) *Geomgraph_TopologyLocation {
	tl := &Geomgraph_TopologyLocation{}
	tl.init(3)
	tl.location[Geom_Position_On] = on
	tl.location[Geom_Position_Left] = left
	tl.location[Geom_Position_Right] = right
	return tl
}

// Geomgraph_NewTopologyLocationOn constructs a TopologyLocation for an edge or
// node (not an area).
func Geomgraph_NewTopologyLocationOn(on int) *Geomgraph_TopologyLocation {
	tl := &Geomgraph_TopologyLocation{}
	tl.init(1)
	tl.location[Geom_Position_On] = on
	return tl
}

// Geomgraph_NewTopologyLocationFromTopologyLocation constructs a
// TopologyLocation which is a copy of the given TopologyLocation.
func Geomgraph_NewTopologyLocationFromTopologyLocation(gl *Geomgraph_TopologyLocation) *Geomgraph_TopologyLocation {
	tl := &Geomgraph_TopologyLocation{}
	tl.init(len(gl.location))
	for i := range tl.location {
		tl.location[i] = gl.location[i]
	}
	return tl
}

func (tl *Geomgraph_TopologyLocation) init(size int) {
	tl.location = make([]int, size)
	tl.SetAllLocations(Geom_Location_None)
}

// Get returns the location at the given position index.
func (tl *Geomgraph_TopologyLocation) Get(posIndex int) int {
	if posIndex < len(tl.location) {
		return tl.location[posIndex]
	}
	return Geom_Location_None
}

// IsNull returns true if all locations are NONE.
func (tl *Geomgraph_TopologyLocation) IsNull() bool {
	for i := range tl.location {
		if tl.location[i] != Geom_Location_None {
			return false
		}
	}
	return true
}

// IsAnyNull returns true if any locations are NONE.
func (tl *Geomgraph_TopologyLocation) IsAnyNull() bool {
	for i := range tl.location {
		if tl.location[i] == Geom_Location_None {
			return true
		}
	}
	return false
}

// IsEqualOnSide returns true if the location at the given index is equal to
// the location at that index in le.
func (tl *Geomgraph_TopologyLocation) IsEqualOnSide(le *Geomgraph_TopologyLocation, locIndex int) bool {
	return tl.location[locIndex] == le.location[locIndex]
}

// IsArea returns true if this TopologyLocation is for an area (i.e., has
// LEFT/RIGHT as well as ON).
func (tl *Geomgraph_TopologyLocation) IsArea() bool {
	return len(tl.location) > 1
}

// IsLine returns true if this TopologyLocation is for a line or node (i.e.,
// only has ON).
func (tl *Geomgraph_TopologyLocation) IsLine() bool {
	return len(tl.location) == 1
}

// Flip swaps the LEFT and RIGHT locations.
func (tl *Geomgraph_TopologyLocation) Flip() {
	if len(tl.location) <= 1 {
		return
	}
	temp := tl.location[Geom_Position_Left]
	tl.location[Geom_Position_Left] = tl.location[Geom_Position_Right]
	tl.location[Geom_Position_Right] = temp
}

// SetAllLocations sets all locations to the given value.
func (tl *Geomgraph_TopologyLocation) SetAllLocations(locValue int) {
	for i := range tl.location {
		tl.location[i] = locValue
	}
}

// SetAllLocationsIfNull sets all locations that are currently NONE to the
// given value.
func (tl *Geomgraph_TopologyLocation) SetAllLocationsIfNull(locValue int) {
	for i := range tl.location {
		if tl.location[i] == Geom_Location_None {
			tl.location[i] = locValue
		}
	}
}

// SetLocation sets the location at the given index.
func (tl *Geomgraph_TopologyLocation) SetLocation(locIndex, locValue int) {
	tl.location[locIndex] = locValue
}

// SetLocationOn sets the ON location.
func (tl *Geomgraph_TopologyLocation) SetLocationOn(locValue int) {
	tl.SetLocation(Geom_Position_On, locValue)
}

// GetLocations returns the array of location values.
func (tl *Geomgraph_TopologyLocation) GetLocations() []int {
	return tl.location
}

// SetLocations sets the ON, LEFT, and RIGHT locations.
func (tl *Geomgraph_TopologyLocation) SetLocations(on, left, right int) {
	tl.location[Geom_Position_On] = on
	tl.location[Geom_Position_Left] = left
	tl.location[Geom_Position_Right] = right
}

// AllPositionsEqual returns true if all locations are equal to the given value.
func (tl *Geomgraph_TopologyLocation) AllPositionsEqual(loc int) bool {
	for i := range tl.location {
		if tl.location[i] != loc {
			return false
		}
	}
	return true
}

// Merge updates only the NONE attributes of this object with the attributes of
// another.
func (tl *Geomgraph_TopologyLocation) Merge(gl *Geomgraph_TopologyLocation) {
	// If the src is an Area label & and the dest is not, increase the dest to
	// be an Area.
	if len(gl.location) > len(tl.location) {
		newLoc := make([]int, 3)
		newLoc[Geom_Position_On] = tl.location[Geom_Position_On]
		newLoc[Geom_Position_Left] = Geom_Location_None
		newLoc[Geom_Position_Right] = Geom_Location_None
		tl.location = newLoc
	}
	for i := range tl.location {
		if tl.location[i] == Geom_Location_None && i < len(gl.location) {
			tl.location[i] = gl.location[i]
		}
	}
}

// String returns a string representation of this TopologyLocation.
func (tl *Geomgraph_TopologyLocation) String() string {
	var buf strings.Builder
	if len(tl.location) > 1 {
		buf.WriteByte(Geom_Location_ToLocationSymbol(tl.location[Geom_Position_Left]))
	}
	buf.WriteByte(Geom_Location_ToLocationSymbol(tl.location[Geom_Position_On]))
	if len(tl.location) > 1 {
		buf.WriteByte(Geom_Location_ToLocationSymbol(tl.location[Geom_Position_Right]))
	}
	return buf.String()
}
