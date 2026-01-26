package jts

import (
	"strings"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geomgraph_Label_ToLineLabel converts a Label to a Line label (that is, one
// with no side Locations).
func Geomgraph_Label_ToLineLabel(label *Geomgraph_Label) *Geomgraph_Label {
	lineLabel := Geomgraph_NewLabelOn(Geom_Location_None)
	for i := 0; i < 2; i++ {
		lineLabel.SetLocationOn(i, label.GetLocationOn(i))
	}
	return lineLabel
}

// Geomgraph_Label indicates the topological relationship of a component of a
// topology graph to a given Geometry.
//
// This class supports labels for relationships to two Geometries, which is
// sufficient for algorithms for binary operations.
//
// Topology graphs support the concept of labeling nodes and edges in the
// graph. The label of a node or edge specifies its topological relationship to
// one or more geometries. (In fact, since JTS operations have only two
// arguments labels are required for only two geometries). A label for a node
// or edge has one or two elements, depending on whether the node or edge
// occurs in one or both of the input Geometries. Elements contain attributes
// which categorize the topological location of the node or edge relative to
// the parent Geometry; that is, whether the node or edge is in the interior,
// boundary or exterior of the Geometry. Attributes have a value from the set
// {Interior, Boundary, Exterior}. In a node each element has a single
// attribute <On>. For an edge each element has a triplet of attributes <Left,
// On, Right>.
//
// It is up to the client code to associate the 0 and 1 TopologyLocations with
// specific geometries.
type Geomgraph_Label struct {
	child java.Polymorphic
	elt   [2]*Geomgraph_TopologyLocation
}

// GetChild returns the immediate child in the type hierarchy chain.
func (l *Geomgraph_Label) GetChild() java.Polymorphic {
	return l.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (l *Geomgraph_Label) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewLabelOn constructs a Label with a single location for both
// Geometries. Initialize the locations to the given value.
func Geomgraph_NewLabelOn(onLoc int) *Geomgraph_Label {
	return &Geomgraph_Label{
		elt: [2]*Geomgraph_TopologyLocation{
			Geomgraph_NewTopologyLocationOn(onLoc),
			Geomgraph_NewTopologyLocationOn(onLoc),
		},
	}
}

// Geomgraph_NewLabelGeomOn constructs a Label with a single location for both
// Geometries. Initialize the location for the Geometry index.
func Geomgraph_NewLabelGeomOn(geomIndex, onLoc int) *Geomgraph_Label {
	l := &Geomgraph_Label{
		elt: [2]*Geomgraph_TopologyLocation{
			Geomgraph_NewTopologyLocationOn(Geom_Location_None),
			Geomgraph_NewTopologyLocationOn(Geom_Location_None),
		},
	}
	l.elt[geomIndex].SetLocationOn(onLoc)
	return l
}

// Geomgraph_NewLabelOnLeftRight constructs a Label with On, Left and Right
// locations for both Geometries. Initialize the locations for both Geometries
// to the given values.
func Geomgraph_NewLabelOnLeftRight(onLoc, leftLoc, rightLoc int) *Geomgraph_Label {
	return &Geomgraph_Label{
		elt: [2]*Geomgraph_TopologyLocation{
			Geomgraph_NewTopologyLocationOnLeftRight(onLoc, leftLoc, rightLoc),
			Geomgraph_NewTopologyLocationOnLeftRight(onLoc, leftLoc, rightLoc),
		},
	}
}

// Geomgraph_NewLabelGeomOnLeftRight constructs a Label with On, Left and Right
// locations for both Geometries. Initialize the locations for the given
// Geometry index.
func Geomgraph_NewLabelGeomOnLeftRight(geomIndex, onLoc, leftLoc, rightLoc int) *Geomgraph_Label {
	l := &Geomgraph_Label{
		elt: [2]*Geomgraph_TopologyLocation{
			Geomgraph_NewTopologyLocationOnLeftRight(Geom_Location_None, Geom_Location_None, Geom_Location_None),
			Geomgraph_NewTopologyLocationOnLeftRight(Geom_Location_None, Geom_Location_None, Geom_Location_None),
		},
	}
	l.elt[geomIndex].SetLocations(onLoc, leftLoc, rightLoc)
	return l
}

// Geomgraph_NewLabelFromLabel constructs a Label with the same values as the
// argument Label.
func Geomgraph_NewLabelFromLabel(lbl *Geomgraph_Label) *Geomgraph_Label {
	return &Geomgraph_Label{
		elt: [2]*Geomgraph_TopologyLocation{
			Geomgraph_NewTopologyLocationFromTopologyLocation(lbl.elt[0]),
			Geomgraph_NewTopologyLocationFromTopologyLocation(lbl.elt[1]),
		},
	}
}

// Flip flips the Left and Right locations for both elements.
func (l *Geomgraph_Label) Flip() {
	l.elt[0].Flip()
	l.elt[1].Flip()
}

// GetLocation returns the location for the given geometry index and position
// index.
func (l *Geomgraph_Label) GetLocation(geomIndex, posIndex int) int {
	return l.elt[geomIndex].Get(posIndex)
}

// GetLocationOn returns the ON location for the given geometry index.
func (l *Geomgraph_Label) GetLocationOn(geomIndex int) int {
	return l.elt[geomIndex].Get(Geom_Position_On)
}

// SetLocation sets the location for the given geometry index and position
// index.
func (l *Geomgraph_Label) SetLocation(geomIndex, posIndex, location int) {
	l.elt[geomIndex].SetLocation(posIndex, location)
}

// SetLocationOn sets the ON location for the given geometry index.
func (l *Geomgraph_Label) SetLocationOn(geomIndex, location int) {
	l.elt[geomIndex].SetLocation(Geom_Position_On, location)
}

// SetAllLocations sets all locations for the given geometry index to the given
// value.
func (l *Geomgraph_Label) SetAllLocations(geomIndex, location int) {
	l.elt[geomIndex].SetAllLocations(location)
}

// SetAllLocationsIfNull sets all NONE locations for the given geometry index
// to the given value.
func (l *Geomgraph_Label) SetAllLocationsIfNull(geomIndex, location int) {
	l.elt[geomIndex].SetAllLocationsIfNull(location)
}

// SetAllLocationsIfNullBoth sets all NONE locations for both geometry indices
// to the given value.
func (l *Geomgraph_Label) SetAllLocationsIfNullBoth(location int) {
	l.SetAllLocationsIfNull(0, location)
	l.SetAllLocationsIfNull(1, location)
}

// Merge merges this label with another one. Merging updates any null
// attributes of this label with the attributes from lbl.
func (l *Geomgraph_Label) Merge(lbl *Geomgraph_Label) {
	for i := 0; i < 2; i++ {
		if l.elt[i] == nil && lbl.elt[i] != nil {
			l.elt[i] = Geomgraph_NewTopologyLocationFromTopologyLocation(lbl.elt[i])
		} else {
			l.elt[i].Merge(lbl.elt[i])
		}
	}
}

// GetGeometryCount returns the number of non-null geometry elements.
func (l *Geomgraph_Label) GetGeometryCount() int {
	count := 0
	if !l.elt[0].IsNull() {
		count++
	}
	if !l.elt[1].IsNull() {
		count++
	}
	return count
}

// IsNull returns true if all locations for the given geometry index are NONE.
func (l *Geomgraph_Label) IsNull(geomIndex int) bool {
	return l.elt[geomIndex].IsNull()
}

// IsAnyNull returns true if any locations for the given geometry index are
// NONE.
func (l *Geomgraph_Label) IsAnyNull(geomIndex int) bool {
	return l.elt[geomIndex].IsAnyNull()
}

// IsArea returns true if either element is an area.
func (l *Geomgraph_Label) IsArea() bool {
	return l.elt[0].IsArea() || l.elt[1].IsArea()
}

// IsAreaAt returns true if the element at the given geometry index is an area.
func (l *Geomgraph_Label) IsAreaAt(geomIndex int) bool {
	return l.elt[geomIndex].IsArea()
}

// IsLine returns true if the element at the given geometry index is a line.
func (l *Geomgraph_Label) IsLine(geomIndex int) bool {
	return l.elt[geomIndex].IsLine()
}

// IsEqualOnSide returns true if the label is equal to lbl on the given side.
func (l *Geomgraph_Label) IsEqualOnSide(lbl *Geomgraph_Label, side int) bool {
	return l.elt[0].IsEqualOnSide(lbl.elt[0], side) &&
		l.elt[1].IsEqualOnSide(lbl.elt[1], side)
}

// AllPositionsEqual returns true if all positions for the given geometry index
// equal the given location.
func (l *Geomgraph_Label) AllPositionsEqual(geomIndex, loc int) bool {
	return l.elt[geomIndex].AllPositionsEqual(loc)
}

// ToLine converts one GeometryLocation to a Line location.
func (l *Geomgraph_Label) ToLine(geomIndex int) {
	if l.elt[geomIndex].IsArea() {
		l.elt[geomIndex] = Geomgraph_NewTopologyLocationOn(l.elt[geomIndex].location[0])
	}
}

// String returns a string representation of this Label.
func (l *Geomgraph_Label) String() string {
	var buf strings.Builder
	if l.elt[0] != nil {
		buf.WriteString("A:")
		buf.WriteString(l.elt[0].String())
	}
	if l.elt[1] != nil {
		buf.WriteString(" B:")
		buf.WriteString(l.elt[1].String())
	}
	return buf.String()
}
