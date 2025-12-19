package jts

var _ Noding_SegmentString = (*Noding_NodedSegmentString)(nil)

// Noding_NodedSegmentString represents a list of contiguous line segments, and
// supports noding the segments. The line segments are represented by an array
// of Coordinates. Intended to optimize the noding of contiguous segments by
// reducing the number of allocated objects. SegmentStrings can carry a context
// object, which is useful for preserving topological or parentage information.
// All noded substrings are initialized with the same context object.
//
// For read-only applications use BasicSegmentString, which is (slightly) more
// lightweight.
type Noding_NodedSegmentString struct {
	nodeList *Noding_SegmentNodeList
	pts      []*Geom_Coordinate
	data     any
}

func (nss *Noding_NodedSegmentString) IsNoding_SegmentString() {}

// Noding_NodedSegmentString_GetNodedSubstrings gets the SegmentStrings which
// result from splitting the input segment strings at node points.
func Noding_NodedSegmentString_GetNodedSubstrings(segStrings []*Noding_NodedSegmentString) []*Noding_NodedSegmentString {
	var resultEdgelist []*Noding_NodedSegmentString
	Noding_NodedSegmentString_GetNodedSubstringsInto(segStrings, &resultEdgelist)
	return resultEdgelist
}

// Noding_NodedSegmentString_GetNodedSubstringsInto adds the noded SegmentStrings
// which result from splitting the input segment strings at node points.
func Noding_NodedSegmentString_GetNodedSubstringsInto(segStrings []*Noding_NodedSegmentString, resultEdgelist *[]*Noding_NodedSegmentString) {
	for _, ss := range segStrings {
		ss.GetNodeList().AddSplitEdges(resultEdgelist)
	}
}

// Noding_NewNodedSegmentString creates an instance from a list of vertices and
// optional data object.
func Noding_NewNodedSegmentString(pts []*Geom_Coordinate, data any) *Noding_NodedSegmentString {
	nss := &Noding_NodedSegmentString{
		pts:  pts,
		data: data,
	}
	nss.nodeList = Noding_NewSegmentNodeList(nss)
	return nss
}

// Noding_NewNodedSegmentStringFromSegmentString creates a new instance from a
// SegmentString.
func Noding_NewNodedSegmentStringFromSegmentString(ss Noding_SegmentString) *Noding_NodedSegmentString {
	return Noding_NewNodedSegmentString(ss.GetCoordinates(), ss.GetData())
}

// GetData gets the user-defined data for this segment string.
func (nss *Noding_NodedSegmentString) GetData() any {
	return nss.data
}

// SetData sets the user-defined data for this segment string.
func (nss *Noding_NodedSegmentString) SetData(data any) {
	nss.data = data
}

// GetNodeList gets the node list for this segment string.
func (nss *Noding_NodedSegmentString) GetNodeList() *Noding_SegmentNodeList {
	return nss.nodeList
}

// Size returns the number of coordinates in this segment string.
func (nss *Noding_NodedSegmentString) Size() int {
	return len(nss.pts)
}

// GetCoordinate gets the segment string coordinate at a given index.
func (nss *Noding_NodedSegmentString) GetCoordinate(i int) *Geom_Coordinate {
	return nss.pts[i]
}

// GetCoordinates gets the coordinates in this segment string.
func (nss *Noding_NodedSegmentString) GetCoordinates() []*Geom_Coordinate {
	return nss.pts
}

// GetNodedCoordinates gets a list of coordinates with all nodes included.
func (nss *Noding_NodedSegmentString) GetNodedCoordinates() []*Geom_Coordinate {
	return nss.nodeList.GetSplitCoordinates()
}

// IsClosed tests if a segment string is a closed ring.
func (nss *Noding_NodedSegmentString) IsClosed() bool {
	return nss.pts[0].Equals(nss.pts[len(nss.pts)-1])
}

// PrevInRing gets the previous vertex in a ring from a vertex index.
func (nss *Noding_NodedSegmentString) PrevInRing(index int) *Geom_Coordinate {
	prevIndex := index - 1
	if prevIndex < 0 {
		prevIndex = nss.Size() - 2
	}
	return nss.GetCoordinate(prevIndex)
}

// NextInRing gets the next vertex in a ring from a vertex index.
func (nss *Noding_NodedSegmentString) NextInRing(index int) *Geom_Coordinate {
	nextIndex := index + 1
	if nextIndex > nss.Size()-1 {
		nextIndex = 1
	}
	return nss.GetCoordinate(nextIndex)
}

// HasNodes tests whether any nodes have been added.
func (nss *Noding_NodedSegmentString) HasNodes() bool {
	return nss.nodeList.Size() > 0
}

// GetSegmentOctant gets the octant of the segment starting at vertex index.
func (nss *Noding_NodedSegmentString) GetSegmentOctant(index int) int {
	if index == len(nss.pts)-1 {
		return -1
	}
	return nss.safeOctant(nss.GetCoordinate(index), nss.GetCoordinate(index+1))
}

func (nss *Noding_NodedSegmentString) safeOctant(p0, p1 *Geom_Coordinate) int {
	if p0.Equals2D(p1) {
		return 0
	}
	return Noding_Octant_Octant(p0, p1)
}

// AddIntersections adds EdgeIntersections for one or both intersections found
// for a segment of an edge to the edge intersection list.
func (nss *Noding_NodedSegmentString) AddIntersections(li *Algorithm_LineIntersector, segmentIndex, geomIndex int) {
	for i := 0; i < li.GetIntersectionNum(); i++ {
		nss.AddIntersectionFromLineIntersector(li, segmentIndex, geomIndex, i)
	}
}

// AddIntersectionFromLineIntersector adds a SegmentNode for intersection
// intIndex. An intersection that falls exactly on a vertex of the SegmentString
// is normalized to use the higher of the two possible segmentIndexes.
func (nss *Noding_NodedSegmentString) AddIntersectionFromLineIntersector(li *Algorithm_LineIntersector, segmentIndex, geomIndex, intIndex int) {
	intPt := li.GetIntersection(intIndex).Copy()
	nss.AddIntersection(intPt, segmentIndex)
}

// AddIntersection adds an intersection node for a given point and segment to
// this segment string.
func (nss *Noding_NodedSegmentString) AddIntersection(intPt *Geom_Coordinate, segmentIndex int) {
	nss.AddIntersectionNode(intPt, segmentIndex)
}

// AddIntersectionNode adds an intersection node for a given point and segment
// to this segment string. If an intersection already exists for this exact
// location, the existing node will be returned.
func (nss *Noding_NodedSegmentString) AddIntersectionNode(intPt *Geom_Coordinate, segmentIndex int) *Noding_SegmentNode {
	normalizedSegmentIndex := segmentIndex
	// Normalize the intersection point location.
	nextSegIndex := normalizedSegmentIndex + 1
	if nextSegIndex < len(nss.pts) {
		nextPt := nss.pts[nextSegIndex]

		// Normalize segment index if intPt falls on vertex. The check for point
		// equality is 2D only - Z values are ignored.
		if intPt.Equals2D(nextPt) {
			normalizedSegmentIndex = nextSegIndex
		}
	}
	// Add the intersection point to edge intersection list.
	ei := nss.nodeList.Add(intPt, normalizedSegmentIndex)
	return ei
}

// String returns a string representation of this NodedSegmentString.
func (nss *Noding_NodedSegmentString) String() string {
	return Io_WKTWriter_ToLineStringFromCoords(nss.pts)
}
