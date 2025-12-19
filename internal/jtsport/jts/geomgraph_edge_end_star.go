package jts

import (
	"sort"
	"strings"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geomgraph_EdgeEndStar is an ordered list of EdgeEnds around a node. They are
// maintained in CCW order (starting with the positive x-axis) around the node
// for efficient lookup and topology building.
type Geomgraph_EdgeEndStar struct {
	child java.Polymorphic

	// edgeMap maintains the edges in sorted order around the node.
	edgeMap []*Geomgraph_EdgeEnd
	// edgeList is a cached copy of the edge map values.
	edgeList []*Geomgraph_EdgeEnd
	// ptInAreaLocation is the location of the point for this star in Geometry
	// i Areas.
	ptInAreaLocation [2]int
}

// GetChild returns the immediate child in the type hierarchy chain.
func (ees *Geomgraph_EdgeEndStar) GetChild() java.Polymorphic {
	return ees.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (ees *Geomgraph_EdgeEndStar) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewEdgeEndStar creates a new EdgeEndStar.
func Geomgraph_NewEdgeEndStar() *Geomgraph_EdgeEndStar {
	return &Geomgraph_EdgeEndStar{
		ptInAreaLocation: [2]int{Geom_Location_None, Geom_Location_None},
	}
}

// Insert inserts an EdgeEnd into this EdgeEndStar. This is an abstract method
// that must be implemented by subtypes.
func (ees *Geomgraph_EdgeEndStar) Insert(e *Geomgraph_EdgeEnd) {
	if impl, ok := java.GetLeaf(ees).(interface {
		Insert_BODY(*Geomgraph_EdgeEnd)
	}); ok {
		impl.Insert_BODY(e)
		return
	}
	panic("abstract method Geomgraph_EdgeEndStar.Insert called")
}

// InsertEdgeEnd inserts an EdgeEnd into the map, and clears the edgeList
// cache, since the list of edges has now changed.
func (ees *Geomgraph_EdgeEndStar) InsertEdgeEnd(e *Geomgraph_EdgeEnd) {
	ees.edgeMap = append(ees.edgeMap, e)
	// Keep sorted by EdgeEnd comparison.
	sort.Slice(ees.edgeMap, func(i, j int) bool {
		return ees.edgeMap[i].CompareTo(ees.edgeMap[j]) < 0
	})
	ees.edgeList = nil // Edge list has changed - clear the cache.
}

// GetCoordinate returns the coordinate for the node this star is based at.
func (ees *Geomgraph_EdgeEndStar) GetCoordinate() *Geom_Coordinate {
	edges := ees.GetEdges()
	if len(edges) == 0 {
		return nil
	}
	return edges[0].GetCoordinate()
}

// GetDegree returns the number of EdgeEnds around this node.
func (ees *Geomgraph_EdgeEndStar) GetDegree() int {
	return len(ees.edgeMap)
}

// GetEdges returns the ordered list of edges. Iterator access to the ordered
// list of edges is optimized by copying the map collection to a list. (This
// assumes that once an iterator is requested, it is likely that insertion into
// the map is complete).
func (ees *Geomgraph_EdgeEndStar) GetEdges() []*Geomgraph_EdgeEnd {
	if ees.edgeList == nil {
		ees.edgeList = make([]*Geomgraph_EdgeEnd, len(ees.edgeMap))
		copy(ees.edgeList, ees.edgeMap)
	}
	return ees.edgeList
}

// GetNextCW returns the next EdgeEnd clockwise from the given EdgeEnd.
func (ees *Geomgraph_EdgeEndStar) GetNextCW(ee *Geomgraph_EdgeEnd) *Geomgraph_EdgeEnd {
	edges := ees.GetEdges()
	i := -1
	for idx, e := range edges {
		if e == ee {
			i = idx
			break
		}
	}
	if i < 0 {
		return nil
	}
	iNextCW := i - 1
	if i == 0 {
		iNextCW = len(edges) - 1
	}
	return edges[iNextCW]
}

// ComputeLabelling computes the labelling for all edges in this star.
func (ees *Geomgraph_EdgeEndStar) ComputeLabelling(geomGraph []*Geomgraph_GeometryGraph) {
	if impl, ok := java.GetLeaf(ees).(interface {
		ComputeLabelling_BODY([]*Geomgraph_GeometryGraph)
	}); ok {
		impl.ComputeLabelling_BODY(geomGraph)
		return
	}
	ees.ComputeLabelling_BODY(geomGraph)
}

// ComputeLabelling_BODY provides the default implementation.
func (ees *Geomgraph_EdgeEndStar) ComputeLabelling_BODY(geomGraph []*Geomgraph_GeometryGraph) {
	ees.computeEdgeEndLabels(geomGraph[0].GetBoundaryNodeRule())
	// Propagate side labels around the edges in the star for each parent
	// Geometry.
	ees.propagateSideLabels(0)
	ees.propagateSideLabels(1)

	// If there are edges that still have null labels for a geometry this must
	// be because there are no area edges for that geometry incident on this
	// node. In this case, to label the edge for that geometry we must test
	// whether the edge is in the interior of the geometry. To do this it
	// suffices to determine whether the node for the edge is in the interior
	// of an area. If so, the edge has location INTERIOR for the geometry. In
	// all other cases (e.g. the node is on a line, on a point, or not on the
	// geometry at all) the edge has the location EXTERIOR for the geometry.
	hasDimensionalCollapseEdge := [2]bool{false, false}
	for _, e := range ees.GetEdges() {
		label := e.GetLabel()
		for geomi := 0; geomi < 2; geomi++ {
			if label.IsLine(geomi) && label.GetLocationOn(geomi) == Geom_Location_Boundary {
				hasDimensionalCollapseEdge[geomi] = true
			}
		}
	}

	for _, e := range ees.GetEdges() {
		label := e.GetLabel()
		for geomi := 0; geomi < 2; geomi++ {
			if label.IsAnyNull(geomi) {
				var loc int
				if hasDimensionalCollapseEdge[geomi] {
					loc = Geom_Location_Exterior
				} else {
					p := e.GetCoordinate()
					loc = ees.getLocation(geomi, p, geomGraph)
				}
				label.SetAllLocationsIfNull(geomi, loc)
			}
		}
	}
}

func (ees *Geomgraph_EdgeEndStar) computeEdgeEndLabels(boundaryNodeRule Algorithm_BoundaryNodeRule) {
	// Compute edge label for each EdgeEnd.
	for _, ee := range ees.GetEdges() {
		ee.ComputeLabel(boundaryNodeRule)
	}
}

func (ees *Geomgraph_EdgeEndStar) getLocation(geomIndex int, p *Geom_Coordinate, geom []*Geomgraph_GeometryGraph) int {
	// Compute location only on demand.
	if ees.ptInAreaLocation[geomIndex] == Geom_Location_None {
		ees.ptInAreaLocation[geomIndex] = AlgorithmLocate_SimplePointInAreaLocator_Locate(p, geom[geomIndex].GetGeometry())
	}
	return ees.ptInAreaLocation[geomIndex]
}

// IsAreaLabelsConsistent checks if the area labels are consistent.
func (ees *Geomgraph_EdgeEndStar) IsAreaLabelsConsistent(geomGraph *Geomgraph_GeometryGraph) bool {
	ees.computeEdgeEndLabels(geomGraph.GetBoundaryNodeRule())
	return ees.checkAreaLabelsConsistent(0)
}

func (ees *Geomgraph_EdgeEndStar) checkAreaLabelsConsistent(geomIndex int) bool {
	// Since edges are stored in CCW order around the node, as we move around
	// the ring we move from the right to the left side of the edge.
	edges := ees.GetEdges()
	// If no edges, trivially consistent.
	if len(edges) <= 0 {
		return true
	}
	// Initialize startLoc to location of last L side (if any).
	lastEdgeIndex := len(edges) - 1
	startLabel := edges[lastEdgeIndex].GetLabel()
	startLoc := startLabel.GetLocation(geomIndex, Geom_Position_Left)
	Util_Assert_IsTrueWithMessage(startLoc != Geom_Location_None, "Found unlabelled area edge")

	currLoc := startLoc
	for _, e := range edges {
		label := e.GetLabel()
		// We assume that we are only checking an area.
		Util_Assert_IsTrueWithMessage(label.IsAreaAt(geomIndex), "Found non-area edge")
		leftLoc := label.GetLocation(geomIndex, Geom_Position_Left)
		rightLoc := label.GetLocation(geomIndex, Geom_Position_Right)
		// Check that edge is really a boundary between inside and outside!
		if leftLoc == rightLoc {
			return false
		}
		// Check side location conflict.
		if rightLoc != currLoc {
			return false
		}
		currLoc = leftLoc
	}
	return true
}

func (ees *Geomgraph_EdgeEndStar) propagateSideLabels(geomIndex int) {
	// Since edges are stored in CCW order around the node, as we move around
	// the ring we move from the right to the left side of the edge.
	startLoc := Geom_Location_None

	// Initialize loc to location of last L side (if any).
	for _, e := range ees.GetEdges() {
		label := e.GetLabel()
		if label.IsAreaAt(geomIndex) && label.GetLocation(geomIndex, Geom_Position_Left) != Geom_Location_None {
			startLoc = label.GetLocation(geomIndex, Geom_Position_Left)
		}
	}

	// No labelled sides found, so no labels to propagate.
	if startLoc == Geom_Location_None {
		return
	}

	currLoc := startLoc
	for _, e := range ees.GetEdges() {
		label := e.GetLabel()
		// Set null ON values to be in current location.
		if label.GetLocation(geomIndex, Geom_Position_On) == Geom_Location_None {
			label.SetLocation(geomIndex, Geom_Position_On, currLoc)
		}
		// Set side labels (if any).
		if label.IsAreaAt(geomIndex) {
			leftLoc := label.GetLocation(geomIndex, Geom_Position_Left)
			rightLoc := label.GetLocation(geomIndex, Geom_Position_Right)
			// If there is a right location, that is the next location to
			// propagate.
			if rightLoc != Geom_Location_None {
				if rightLoc != currLoc {
					panic(Geom_NewTopologyExceptionWithCoordinate("side location conflict", e.GetCoordinate()))
				}
				if leftLoc == Geom_Location_None {
					Util_Assert_ShouldNeverReachHereWithMessage("found single null side (at " + e.GetCoordinate().String() + ")")
				}
				currLoc = leftLoc
			} else {
				// RHS is null - LHS must be null too. This must be an edge from
				// the other geometry, which has no location labelling for this
				// geometry. This edge must lie wholly inside or outside the
				// other geometry (which is determined by the current location).
				// Assign both sides to be the current location.
				Util_Assert_IsTrueWithMessage(label.GetLocation(geomIndex, Geom_Position_Left) == Geom_Location_None, "found single null side")
				label.SetLocation(geomIndex, Geom_Position_Right, currLoc)
				label.SetLocation(geomIndex, Geom_Position_Left, currLoc)
			}
		}
	}
}

// FindIndex returns the index of the given EdgeEnd, or -1 if not found.
func (ees *Geomgraph_EdgeEndStar) FindIndex(eSearch *Geomgraph_EdgeEnd) int {
	ees.GetEdges() // Force edgelist to be computed.
	for i, e := range ees.edgeList {
		if e == eSearch {
			return i
		}
	}
	return -1
}

// String returns a string representation of this EdgeEndStar.
func (ees *Geomgraph_EdgeEndStar) String() string {
	var buf strings.Builder
	buf.WriteString("EdgeEndStar:   ")
	if coord := ees.GetCoordinate(); coord != nil {
		buf.WriteString(coord.String())
	}
	buf.WriteString("\n")
	for _, e := range ees.GetEdges() {
		buf.WriteString(e.String())
		buf.WriteString("\n")
	}
	return buf.String()
}
