package jts

// OperationOverlayng_OverlayLabel is a structure recording the topological
// situation for an edge in a topology graph used during overlay processing.
// A label contains the topological Locations for one or two input geometries
// to an overlay operation. An input geometry may be either a Line or an Area.
// The label locations for each input geometry are populated with the Locations
// for the edge Positions when they are created or once they are computed by
// topological evaluation. A label also records the (effective) dimension of
// each input geometry. For area edges the role (shell or hole) of the
// originating ring is recorded, to allow determination of edge handling in
// collapse cases.
//
// In an OverlayGraph a single label is shared between the two
// oppositely-oriented OverlayEdges of a symmetric pair. Accessors for
// orientation-sensitive information are parameterized by the orientation of
// the containing edge.
//
// For each input geometry (0 and 1), the label records that an edge is in one
// of the following states (identified by the dim field). Each state has
// additional information about the edge topology.
//
//   - A Boundary edge of an Area (polygon)
//   - A Collapsed edge of an input Area (formed by merging two or more parent edges)
//   - A Line edge from an input line
//   - An edge which is Not Part of an input geometry (and thus must be part of
//     the other geometry)
type OperationOverlayng_OverlayLabel struct {
	aDim      int
	aIsHole   bool
	aLocLeft  int
	aLocRight int
	aLocLine  int

	bDim      int
	bIsHole   bool
	bLocLeft  int
	bLocRight int
	bLocLine  int
}

const (
	operationOverlayng_OverlayLabel_SYM_UNKNOWN  = '#'
	operationOverlayng_OverlayLabel_SYM_BOUNDARY = 'B'
	operationOverlayng_OverlayLabel_SYM_COLLAPSE = 'C'
	operationOverlayng_OverlayLabel_SYM_LINE     = 'L'
)

const (
	// OperationOverlayng_OverlayLabel_DIM_UNKNOWN is the dimension of an input
	// geometry which is not known.
	OperationOverlayng_OverlayLabel_DIM_UNKNOWN = -1

	// OperationOverlayng_OverlayLabel_DIM_NOT_PART is the dimension of an edge
	// which is not part of a specified input geometry.
	OperationOverlayng_OverlayLabel_DIM_NOT_PART = OperationOverlayng_OverlayLabel_DIM_UNKNOWN

	// OperationOverlayng_OverlayLabel_DIM_LINE is the dimension of an edge
	// which is a line.
	OperationOverlayng_OverlayLabel_DIM_LINE = 1

	// OperationOverlayng_OverlayLabel_DIM_BOUNDARY is the dimension for an edge
	// which is part of an input Area geometry boundary.
	OperationOverlayng_OverlayLabel_DIM_BOUNDARY = 2

	// OperationOverlayng_OverlayLabel_DIM_COLLAPSE is the dimension for an edge
	// which is a collapsed part of an input Area geometry boundary. A collapsed
	// edge represents two or more line segments which have the same endpoints.
	// They usually are caused by edges in valid polygonal geometries having
	// their endpoints become identical due to precision reduction.
	OperationOverlayng_OverlayLabel_DIM_COLLAPSE = 3
)

// OperationOverlayng_OverlayLabel_LOC_UNKNOWN indicates that the location is
// currently unknown.
var OperationOverlayng_OverlayLabel_LOC_UNKNOWN = Geom_Location_None

// OperationOverlayng_NewOverlayLabel creates an uninitialized label.
func OperationOverlayng_NewOverlayLabel() *OperationOverlayng_OverlayLabel {
	return &OperationOverlayng_OverlayLabel{
		aDim:      OperationOverlayng_OverlayLabel_DIM_NOT_PART,
		aIsHole:   false,
		aLocLeft:  OperationOverlayng_OverlayLabel_LOC_UNKNOWN,
		aLocRight: OperationOverlayng_OverlayLabel_LOC_UNKNOWN,
		aLocLine:  OperationOverlayng_OverlayLabel_LOC_UNKNOWN,
		bDim:      OperationOverlayng_OverlayLabel_DIM_NOT_PART,
		bIsHole:   false,
		bLocLeft:  OperationOverlayng_OverlayLabel_LOC_UNKNOWN,
		bLocRight: OperationOverlayng_OverlayLabel_LOC_UNKNOWN,
		bLocLine:  OperationOverlayng_OverlayLabel_LOC_UNKNOWN,
	}
}

// OperationOverlayng_NewOverlayLabelForBoundary creates a label for an Area edge.
func OperationOverlayng_NewOverlayLabelForBoundary(index, locLeft, locRight int, isHole bool) *OperationOverlayng_OverlayLabel {
	lbl := OperationOverlayng_NewOverlayLabel()
	lbl.InitBoundary(index, locLeft, locRight, isHole)
	return lbl
}

// OperationOverlayng_NewOverlayLabelForLine creates a label for a Line edge.
func OperationOverlayng_NewOverlayLabelForLine(index int) *OperationOverlayng_OverlayLabel {
	lbl := OperationOverlayng_NewOverlayLabel()
	lbl.InitLine(index)
	return lbl
}

// OperationOverlayng_NewOverlayLabelCopy creates a label which is a copy of
// another label.
func OperationOverlayng_NewOverlayLabelCopy(lbl *OperationOverlayng_OverlayLabel) *OperationOverlayng_OverlayLabel {
	return &OperationOverlayng_OverlayLabel{
		aLocLeft:  lbl.aLocLeft,
		aLocRight: lbl.aLocRight,
		aLocLine:  lbl.aLocLine,
		aDim:      lbl.aDim,
		aIsHole:   lbl.aIsHole,
		bLocLeft:  lbl.bLocLeft,
		bLocRight: lbl.bLocRight,
		bLocLine:  lbl.bLocLine,
		bDim:      lbl.bDim,
		bIsHole:   lbl.bIsHole,
	}
}

// Dimension gets the effective dimension of the given input geometry.
func (ol *OperationOverlayng_OverlayLabel) Dimension(index int) int {
	if index == 0 {
		return ol.aDim
	}
	return ol.bDim
}

// InitBoundary initializes the label for an input geometry which is an Area
// boundary.
func (ol *OperationOverlayng_OverlayLabel) InitBoundary(index, locLeft, locRight int, isHole bool) {
	if index == 0 {
		ol.aDim = OperationOverlayng_OverlayLabel_DIM_BOUNDARY
		ol.aIsHole = isHole
		ol.aLocLeft = locLeft
		ol.aLocRight = locRight
		ol.aLocLine = Geom_Location_Interior
	} else {
		ol.bDim = OperationOverlayng_OverlayLabel_DIM_BOUNDARY
		ol.bIsHole = isHole
		ol.bLocLeft = locLeft
		ol.bLocRight = locRight
		ol.bLocLine = Geom_Location_Interior
	}
}

// InitCollapse initializes the label for an edge which is the collapse of part
// of the boundary of an Area input geometry. The location of the collapsed
// edge relative to the parent area geometry is initially unknown. It must be
// determined from the topology of the overlay graph.
func (ol *OperationOverlayng_OverlayLabel) InitCollapse(index int, isHole bool) {
	if index == 0 {
		ol.aDim = OperationOverlayng_OverlayLabel_DIM_COLLAPSE
		ol.aIsHole = isHole
	} else {
		ol.bDim = OperationOverlayng_OverlayLabel_DIM_COLLAPSE
		ol.bIsHole = isHole
	}
}

// InitLine initializes the label for an input geometry which is a Line.
func (ol *OperationOverlayng_OverlayLabel) InitLine(index int) {
	if index == 0 {
		ol.aDim = OperationOverlayng_OverlayLabel_DIM_LINE
		ol.aLocLine = OperationOverlayng_OverlayLabel_LOC_UNKNOWN
	} else {
		ol.bDim = OperationOverlayng_OverlayLabel_DIM_LINE
		ol.bLocLine = OperationOverlayng_OverlayLabel_LOC_UNKNOWN
	}
}

// InitNotPart initializes the label for an edge which is not part of an input
// geometry.
func (ol *OperationOverlayng_OverlayLabel) InitNotPart(index int) {
	if index == 0 {
		ol.aDim = OperationOverlayng_OverlayLabel_DIM_NOT_PART
	} else {
		ol.bDim = OperationOverlayng_OverlayLabel_DIM_NOT_PART
	}
}

// SetLocationLine sets the line location. This is used to set the locations
// for linear edges encountered during area label propagation.
func (ol *OperationOverlayng_OverlayLabel) SetLocationLine(index, loc int) {
	if index == 0 {
		ol.aLocLine = loc
	} else {
		ol.bLocLine = loc
	}
}

// SetLocationAll sets the location of all positions for a given input.
func (ol *OperationOverlayng_OverlayLabel) SetLocationAll(index, loc int) {
	if index == 0 {
		ol.aLocLine = loc
		ol.aLocLeft = loc
		ol.aLocRight = loc
	} else {
		ol.bLocLine = loc
		ol.bLocLeft = loc
		ol.bLocRight = loc
	}
}

// SetLocationCollapse sets the location for a collapsed edge (the Line
// position) for an input geometry, depending on the ring role recorded in the
// label. If the input geometry edge is from a shell, the location is EXTERIOR,
// if it is a hole it is INTERIOR.
func (ol *OperationOverlayng_OverlayLabel) SetLocationCollapse(index int) {
	loc := Geom_Location_Exterior
	if ol.IsHole(index) {
		loc = Geom_Location_Interior
	}
	if index == 0 {
		ol.aLocLine = loc
	} else {
		ol.bLocLine = loc
	}
}

// IsLine tests whether at least one of the sources is a Line.
func (ol *OperationOverlayng_OverlayLabel) IsLine() bool {
	return ol.aDim == OperationOverlayng_OverlayLabel_DIM_LINE || ol.bDim == OperationOverlayng_OverlayLabel_DIM_LINE
}

// IsLineIndex tests whether a source is a Line.
func (ol *OperationOverlayng_OverlayLabel) IsLineIndex(index int) bool {
	if index == 0 {
		return ol.aDim == OperationOverlayng_OverlayLabel_DIM_LINE
	}
	return ol.bDim == OperationOverlayng_OverlayLabel_DIM_LINE
}

// IsLinear tests whether an edge is linear (a Line or a Collapse) in an input
// geometry.
func (ol *OperationOverlayng_OverlayLabel) IsLinear(index int) bool {
	if index == 0 {
		return ol.aDim == OperationOverlayng_OverlayLabel_DIM_LINE || ol.aDim == OperationOverlayng_OverlayLabel_DIM_COLLAPSE
	}
	return ol.bDim == OperationOverlayng_OverlayLabel_DIM_LINE || ol.bDim == OperationOverlayng_OverlayLabel_DIM_COLLAPSE
}

// IsKnown tests whether a the source of a label is known.
func (ol *OperationOverlayng_OverlayLabel) IsKnown(index int) bool {
	if index == 0 {
		return ol.aDim != OperationOverlayng_OverlayLabel_DIM_UNKNOWN
	}
	return ol.bDim != OperationOverlayng_OverlayLabel_DIM_UNKNOWN
}

// IsNotPart tests whether a label is for an edge which is not part of a given
// input geometry.
func (ol *OperationOverlayng_OverlayLabel) IsNotPart(index int) bool {
	if index == 0 {
		return ol.aDim == OperationOverlayng_OverlayLabel_DIM_NOT_PART
	}
	return ol.bDim == OperationOverlayng_OverlayLabel_DIM_NOT_PART
}

// IsBoundaryEither tests if a label is for an edge which is in the boundary of
// either source geometry.
func (ol *OperationOverlayng_OverlayLabel) IsBoundaryEither() bool {
	return ol.aDim == OperationOverlayng_OverlayLabel_DIM_BOUNDARY || ol.bDim == OperationOverlayng_OverlayLabel_DIM_BOUNDARY
}

// IsBoundaryBoth tests if a label is for an edge which is in the boundary of
// both source geometries.
func (ol *OperationOverlayng_OverlayLabel) IsBoundaryBoth() bool {
	return ol.aDim == OperationOverlayng_OverlayLabel_DIM_BOUNDARY && ol.bDim == OperationOverlayng_OverlayLabel_DIM_BOUNDARY
}

// IsBoundaryCollapse tests if the label is a collapsed edge of one area and is
// a (non-collapsed) boundary edge of the other area.
func (ol *OperationOverlayng_OverlayLabel) IsBoundaryCollapse() bool {
	if ol.IsLine() {
		return false
	}
	return !ol.IsBoundaryBoth()
}

// IsBoundaryTouch tests if a label is for an edge where two areas touch along
// their boundary.
func (ol *OperationOverlayng_OverlayLabel) IsBoundaryTouch() bool {
	return ol.IsBoundaryBoth() &&
		ol.GetLocation(0, Geom_Position_Right, true) != ol.GetLocation(1, Geom_Position_Right, true)
}

// IsBoundary tests if a label is for an edge which is in the boundary of a
// source geometry. Collapses are not reported as being in the boundary.
func (ol *OperationOverlayng_OverlayLabel) IsBoundary(index int) bool {
	if index == 0 {
		return ol.aDim == OperationOverlayng_OverlayLabel_DIM_BOUNDARY
	}
	return ol.bDim == OperationOverlayng_OverlayLabel_DIM_BOUNDARY
}

// IsBoundarySingleton tests whether a label is for an edge which is a boundary
// of one geometry and not part of the other.
func (ol *OperationOverlayng_OverlayLabel) IsBoundarySingleton() bool {
	if ol.aDim == OperationOverlayng_OverlayLabel_DIM_BOUNDARY && ol.bDim == OperationOverlayng_OverlayLabel_DIM_NOT_PART {
		return true
	}
	if ol.bDim == OperationOverlayng_OverlayLabel_DIM_BOUNDARY && ol.aDim == OperationOverlayng_OverlayLabel_DIM_NOT_PART {
		return true
	}
	return false
}

// IsLineLocationUnknown tests if the line location for a source is unknown.
func (ol *OperationOverlayng_OverlayLabel) IsLineLocationUnknown(index int) bool {
	if index == 0 {
		return ol.aLocLine == OperationOverlayng_OverlayLabel_LOC_UNKNOWN
	}
	return ol.bLocLine == OperationOverlayng_OverlayLabel_LOC_UNKNOWN
}

// IsLineInArea tests if a line edge is inside a source geometry (i.e. it has
// location INTERIOR).
func (ol *OperationOverlayng_OverlayLabel) IsLineInArea(index int) bool {
	if index == 0 {
		return ol.aLocLine == Geom_Location_Interior
	}
	return ol.bLocLine == Geom_Location_Interior
}

// IsHole tests if the ring role of an edge is a hole.
func (ol *OperationOverlayng_OverlayLabel) IsHole(index int) bool {
	if index == 0 {
		return ol.aIsHole
	}
	return ol.bIsHole
}

// IsCollapse tests if an edge is a Collapse for a source geometry.
func (ol *OperationOverlayng_OverlayLabel) IsCollapse(index int) bool {
	return ol.Dimension(index) == OperationOverlayng_OverlayLabel_DIM_COLLAPSE
}

// IsInteriorCollapse tests if a label is a Collapse has location INTERIOR, to
// at least one source geometry.
func (ol *OperationOverlayng_OverlayLabel) IsInteriorCollapse() bool {
	if ol.aDim == OperationOverlayng_OverlayLabel_DIM_COLLAPSE && ol.aLocLine == Geom_Location_Interior {
		return true
	}
	if ol.bDim == OperationOverlayng_OverlayLabel_DIM_COLLAPSE && ol.bLocLine == Geom_Location_Interior {
		return true
	}
	return false
}

// IsCollapseAndNotPartInterior tests if a label is a Collapse and NotPart with
// location INTERIOR for the other geometry.
func (ol *OperationOverlayng_OverlayLabel) IsCollapseAndNotPartInterior() bool {
	if ol.aDim == OperationOverlayng_OverlayLabel_DIM_COLLAPSE && ol.bDim == OperationOverlayng_OverlayLabel_DIM_NOT_PART && ol.bLocLine == Geom_Location_Interior {
		return true
	}
	if ol.bDim == OperationOverlayng_OverlayLabel_DIM_COLLAPSE && ol.aDim == OperationOverlayng_OverlayLabel_DIM_NOT_PART && ol.aLocLine == Geom_Location_Interior {
		return true
	}
	return false
}

// GetLineLocation gets the line location for a source geometry.
func (ol *OperationOverlayng_OverlayLabel) GetLineLocation(index int) int {
	if index == 0 {
		return ol.aLocLine
	}
	return ol.bLocLine
}

// IsLineInterior tests if a line is in the interior of a source geometry.
func (ol *OperationOverlayng_OverlayLabel) IsLineInterior(index int) bool {
	if index == 0 {
		return ol.aLocLine == Geom_Location_Interior
	}
	return ol.bLocLine == Geom_Location_Interior
}

// GetLocation gets the location for a Position of an edge of a source for an
// edge with given orientation.
func (ol *OperationOverlayng_OverlayLabel) GetLocation(index, position int, isForward bool) int {
	if index == 0 {
		switch position {
		case Geom_Position_Left:
			if isForward {
				return ol.aLocLeft
			}
			return ol.aLocRight
		case Geom_Position_Right:
			if isForward {
				return ol.aLocRight
			}
			return ol.aLocLeft
		case Geom_Position_On:
			return ol.aLocLine
		}
	}
	// index == 1
	switch position {
	case Geom_Position_Left:
		if isForward {
			return ol.bLocLeft
		}
		return ol.bLocRight
	case Geom_Position_Right:
		if isForward {
			return ol.bLocRight
		}
		return ol.bLocLeft
	case Geom_Position_On:
		return ol.bLocLine
	}
	return OperationOverlayng_OverlayLabel_LOC_UNKNOWN
}

// GetLocationBoundaryOrLine gets the location for this label for either a
// Boundary or a Line edge. This supports a simple determination of whether the
// edge should be included as a result edge.
func (ol *OperationOverlayng_OverlayLabel) GetLocationBoundaryOrLine(index, position int, isForward bool) int {
	if ol.IsBoundary(index) {
		return ol.GetLocation(index, position, isForward)
	}
	return ol.GetLineLocation(index)
}

// GetLocationIndex gets the linear location for the given source.
func (ol *OperationOverlayng_OverlayLabel) GetLocationIndex(index int) int {
	if index == 0 {
		return ol.aLocLine
	}
	return ol.bLocLine
}

// HasSides tests whether this label has side position information for a source
// geometry.
func (ol *OperationOverlayng_OverlayLabel) HasSides(index int) bool {
	if index == 0 {
		return ol.aLocLeft != OperationOverlayng_OverlayLabel_LOC_UNKNOWN ||
			ol.aLocRight != OperationOverlayng_OverlayLabel_LOC_UNKNOWN
	}
	return ol.bLocLeft != OperationOverlayng_OverlayLabel_LOC_UNKNOWN ||
		ol.bLocRight != OperationOverlayng_OverlayLabel_LOC_UNKNOWN
}

// Copy creates a copy of this label.
func (ol *OperationOverlayng_OverlayLabel) Copy() *OperationOverlayng_OverlayLabel {
	return OperationOverlayng_NewOverlayLabelCopy(ol)
}

// String returns a string representation of the label.
func (ol *OperationOverlayng_OverlayLabel) String() string {
	return ol.ToStringWithDirection(true)
}

// ToStringWithDirection returns a string representation of the label with the
// given direction.
func (ol *OperationOverlayng_OverlayLabel) ToStringWithDirection(isForward bool) string {
	return "A:" + ol.locationString(0, isForward) + "/B:" + ol.locationString(1, isForward)
}

func (ol *OperationOverlayng_OverlayLabel) locationString(index int, isForward bool) string {
	buf := ""
	if ol.IsBoundary(index) {
		buf += string(Geom_Location_ToLocationSymbol(ol.GetLocation(index, Geom_Position_Left, isForward)))
		buf += string(Geom_Location_ToLocationSymbol(ol.GetLocation(index, Geom_Position_Right, isForward)))
	} else {
		// is a linear edge
		locLine := ol.aLocLine
		if index != 0 {
			locLine = ol.bLocLine
		}
		buf += string(Geom_Location_ToLocationSymbol(locLine))
	}
	if ol.IsKnown(index) {
		dim := ol.aDim
		if index != 0 {
			dim = ol.bDim
		}
		buf += string(OperationOverlayng_OverlayLabel_DimensionSymbol(dim))
	}
	if ol.IsCollapse(index) {
		isHole := ol.aIsHole
		if index != 0 {
			isHole = ol.bIsHole
		}
		buf += string(OperationOverlayng_OverlayLabel_RingRoleSymbol(isHole))
	}
	return buf
}

// OperationOverlayng_OverlayLabel_RingRoleSymbol gets a symbol for the a ring
// role (Shell or Hole).
func OperationOverlayng_OverlayLabel_RingRoleSymbol(isHole bool) byte {
	if isHole {
		return 'h'
	}
	return 's'
}

// OperationOverlayng_OverlayLabel_DimensionSymbol gets the symbol for the
// dimension code of an edge.
func OperationOverlayng_OverlayLabel_DimensionSymbol(dim int) byte {
	switch dim {
	case OperationOverlayng_OverlayLabel_DIM_LINE:
		return operationOverlayng_OverlayLabel_SYM_LINE
	case OperationOverlayng_OverlayLabel_DIM_COLLAPSE:
		return operationOverlayng_OverlayLabel_SYM_COLLAPSE
	case OperationOverlayng_OverlayLabel_DIM_BOUNDARY:
		return operationOverlayng_OverlayLabel_SYM_BOUNDARY
	}
	return operationOverlayng_OverlayLabel_SYM_UNKNOWN
}
