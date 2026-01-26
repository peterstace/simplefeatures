package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geomgraph_EdgeRing represents a ring of DirectedEdges which forms a polygon.
type Geomgraph_EdgeRing struct {
	child java.Polymorphic

	startDe         *Geomgraph_DirectedEdge // The directed edge which starts the list of edges for this EdgeRing.
	maxNodeDegree   int
	edges           []*Geomgraph_DirectedEdge // The DirectedEdges making up this EdgeRing.
	pts             []*Geom_Coordinate
	label           *Geomgraph_Label // Label stores the locations of each geometry on the face surrounded by this ring.
	ring            *Geom_LinearRing // The ring created for this EdgeRing.
	isHole          bool
	shell           *Geomgraph_EdgeRing   // If non-null, the ring is a hole and this EdgeRing is its containing shell.
	holes           []*Geomgraph_EdgeRing // A list of EdgeRings which are holes in this EdgeRing.
	geometryFactory *Geom_GeometryFactory
}

// GetChild returns the immediate child in the type hierarchy chain.
func (er *Geomgraph_EdgeRing) GetChild() java.Polymorphic {
	return er.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (er *Geomgraph_EdgeRing) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewEdgeRing creates a new EdgeRing from a starting DirectedEdge.
// Note: This constructor should not be used directly when creating subtypes
// like MaximalEdgeRing or MinimalEdgeRing. Use geomgraph_NewEdgeRingBase
// followed by geomgraph_InitEdgeRing instead.
func Geomgraph_NewEdgeRing(start *Geomgraph_DirectedEdge, geometryFactory *Geom_GeometryFactory) *Geomgraph_EdgeRing {
	er := geomgraph_NewEdgeRingBase(geometryFactory)
	geomgraph_InitEdgeRing(er, start)
	return er
}

// geomgraph_NewEdgeRingBase creates an uninitialized EdgeRing. The caller must
// set up the child chain and then call geomgraph_InitEdgeRing.
func geomgraph_NewEdgeRingBase(geometryFactory *Geom_GeometryFactory) *Geomgraph_EdgeRing {
	return &Geomgraph_EdgeRing{
		maxNodeDegree:   -1,
		label:           Geomgraph_NewLabelOn(Geom_Location_None),
		geometryFactory: geometryFactory,
		edges:           make([]*Geomgraph_DirectedEdge, 0),
		pts:             make([]*Geom_Coordinate, 0),
		holes:           make([]*Geomgraph_EdgeRing, 0),
	}
}

// geomgraph_InitEdgeRing initializes an EdgeRing by computing points from the
// starting DirectedEdge. This must be called after the child chain is set up.
func geomgraph_InitEdgeRing(er *Geomgraph_EdgeRing, start *Geomgraph_DirectedEdge) {
	er.computePoints(start)
	er.ComputeRing()
}

// IsIsolated returns true if the ring is isolated.
func (er *Geomgraph_EdgeRing) IsIsolated() bool {
	return er.label.GetGeometryCount() == 1
}

// IsHole returns true if this ring is a hole.
func (er *Geomgraph_EdgeRing) IsHole() bool {
	return er.isHole
}

// GetCoordinate returns the coordinate at the given index.
func (er *Geomgraph_EdgeRing) GetCoordinate(i int) *Geom_Coordinate {
	return er.pts[i]
}

// GetLinearRing returns the LinearRing for this EdgeRing.
func (er *Geomgraph_EdgeRing) GetLinearRing() *Geom_LinearRing {
	return er.ring
}

// GetLabel returns the label for this EdgeRing.
func (er *Geomgraph_EdgeRing) GetLabel() *Geomgraph_Label {
	return er.label
}

// IsShell returns true if this ring is a shell (not a hole).
func (er *Geomgraph_EdgeRing) IsShell() bool {
	return er.shell == nil
}

// GetShell returns the shell EdgeRing if this is a hole.
func (er *Geomgraph_EdgeRing) GetShell() *Geomgraph_EdgeRing {
	return er.shell
}

// SetShell sets the shell EdgeRing for this hole.
func (er *Geomgraph_EdgeRing) SetShell(shell *Geomgraph_EdgeRing) {
	er.shell = shell
	if shell != nil {
		shell.AddHole(er)
	}
}

// AddHole adds a hole to this shell.
func (er *Geomgraph_EdgeRing) AddHole(ring *Geomgraph_EdgeRing) {
	er.holes = append(er.holes, ring)
}

// ToPolygon converts this EdgeRing to a Polygon.
func (er *Geomgraph_EdgeRing) ToPolygon(geometryFactory *Geom_GeometryFactory) *Geom_Polygon {
	holeLR := make([]*Geom_LinearRing, len(er.holes))
	for i, hole := range er.holes {
		holeLR[i] = hole.GetLinearRing()
	}
	return geometryFactory.CreatePolygonWithLinearRingAndHoles(er.GetLinearRing(), holeLR)
}

// ComputeRing computes a LinearRing from the point list previously collected.
// Tests if the ring is a hole (i.e. if it is CCW) and sets the hole flag accordingly.
func (er *Geomgraph_EdgeRing) ComputeRing() {
	if er.ring != nil {
		return // Don't compute more than once.
	}
	coord := make([]*Geom_Coordinate, len(er.pts))
	copy(coord, er.pts)
	er.ring = er.geometryFactory.CreateLinearRingFromCoordinates(coord)
	er.isHole = Algorithm_Orientation_IsCCW(er.ring.GetCoordinates())
}

// GetNext returns the next DirectedEdge in the edge ring. This is an abstract
// method that must be implemented by subtypes.
func (er *Geomgraph_EdgeRing) GetNext(de *Geomgraph_DirectedEdge) *Geomgraph_DirectedEdge {
	if impl, ok := java.GetLeaf(er).(interface {
		GetNext_BODY(*Geomgraph_DirectedEdge) *Geomgraph_DirectedEdge
	}); ok {
		return impl.GetNext_BODY(de)
	}
	panic("abstract method Geomgraph_EdgeRing.GetNext called")
}

// SetEdgeRing sets the EdgeRing for a DirectedEdge. This is an abstract method
// that must be implemented by subtypes.
func (er *Geomgraph_EdgeRing) SetEdgeRing(de *Geomgraph_DirectedEdge, ring *Geomgraph_EdgeRing) {
	if impl, ok := java.GetLeaf(er).(interface {
		SetEdgeRing_BODY(*Geomgraph_DirectedEdge, *Geomgraph_EdgeRing)
	}); ok {
		impl.SetEdgeRing_BODY(de, ring)
		return
	}
	panic("abstract method Geomgraph_EdgeRing.SetEdgeRing called")
}

// GetEdges returns the list of DirectedEdges that make up this EdgeRing.
func (er *Geomgraph_EdgeRing) GetEdges() []*Geomgraph_DirectedEdge {
	return er.edges
}

// computePoints collects all the points from the DirectedEdges of this ring
// into a contiguous list.
func (er *Geomgraph_EdgeRing) computePoints(start *Geomgraph_DirectedEdge) {
	er.startDe = start
	de := start
	isFirstEdge := true
	for {
		if de == nil {
			panic(Geom_NewTopologyException("Found null DirectedEdge"))
		}
		if de.GetEdgeRing() == er {
			panic(Geom_NewTopologyExceptionWithCoordinate("Directed Edge visited twice during ring-building at", de.GetCoordinate()))
		}

		er.edges = append(er.edges, de)
		label := de.GetLabel()
		Util_Assert_IsTrueWithMessage(label.IsArea(), "label is not area")
		er.mergeLabel(label)
		er.addPoints(de.GetEdge(), de.IsForward(), isFirstEdge)
		isFirstEdge = false
		er.SetEdgeRing(de, er)
		de = er.GetNext(de)
		if de == start {
			break
		}
	}
}

// GetMaxNodeDegree returns the maximum node degree.
func (er *Geomgraph_EdgeRing) GetMaxNodeDegree() int {
	if er.maxNodeDegree < 0 {
		er.computeMaxNodeDegree()
	}
	return er.maxNodeDegree
}

func (er *Geomgraph_EdgeRing) computeMaxNodeDegree() {
	er.maxNodeDegree = 0
	de := er.startDe
	for {
		node := de.GetNode()
		des := java.GetLeaf(node.GetEdges()).(*Geomgraph_DirectedEdgeStar)
		degree := des.GetOutgoingDegreeForEdgeRing(er)
		if degree > er.maxNodeDegree {
			er.maxNodeDegree = degree
		}
		de = er.GetNext(de)
		if de == er.startDe {
			break
		}
	}
	er.maxNodeDegree *= 2
}

// SetInResult marks all edges in this ring as being in the result.
func (er *Geomgraph_EdgeRing) SetInResult() {
	de := er.startDe
	for {
		de.GetEdge().SetInResult(true)
		de = de.GetNext()
		if de == er.startDe {
			break
		}
	}
}

func (er *Geomgraph_EdgeRing) mergeLabel(deLabel *Geomgraph_Label) {
	er.mergeLabelForGeom(deLabel, 0)
	er.mergeLabelForGeom(deLabel, 1)
}

// mergeLabelForGeom merges the RHS label from a DirectedEdge into the label
// for this EdgeRing. The DirectedEdge label may be null. This is acceptable -
// it results from a node which is NOT an intersection node between the
// Geometries (e.g. the end node of a LinearRing). In this case the DirectedEdge
// label does not contribute any information to the overall labelling, and is
// simply skipped.
func (er *Geomgraph_EdgeRing) mergeLabelForGeom(deLabel *Geomgraph_Label, geomIndex int) {
	loc := deLabel.GetLocation(geomIndex, Geom_Position_Right)
	// No information to be had from this label.
	if loc == Geom_Location_None {
		return
	}
	// If there is no current RHS value, set it.
	if er.label.GetLocationOn(geomIndex) == Geom_Location_None {
		er.label.SetLocationOn(geomIndex, loc)
	}
}

func (er *Geomgraph_EdgeRing) addPoints(edge *Geomgraph_Edge, isForward, isFirstEdge bool) {
	edgePts := edge.GetCoordinates()
	if isForward {
		startIndex := 1
		if isFirstEdge {
			startIndex = 0
		}
		for i := startIndex; i < len(edgePts); i++ {
			er.pts = append(er.pts, edgePts[i])
		}
	} else {
		// Is backward.
		startIndex := len(edgePts) - 2
		if isFirstEdge {
			startIndex = len(edgePts) - 1
		}
		for i := startIndex; i >= 0; i-- {
			er.pts = append(er.pts, edgePts[i])
		}
	}
}

// ContainsPoint tests if this ring contains the given point. This method will
// cause the ring to be computed. It will also check any holes, if they have
// been assigned.
func (er *Geomgraph_EdgeRing) ContainsPoint(p *Geom_Coordinate) bool {
	shell := er.GetLinearRing()
	env := shell.GetEnvelopeInternal()
	if !env.ContainsCoordinate(p) {
		return false
	}
	if !Algorithm_PointLocation_IsInRing(p, shell.GetCoordinates()) {
		return false
	}
	for _, hole := range er.holes {
		if hole.ContainsPoint(p) {
			return false
		}
	}
	return true
}
