package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// operationBuffer_RightmostEdgeFinder finds the DirectedEdge in a list which has the highest coordinate,
// and which is oriented L to R at that point. (I.e. the right side is on the RHS of the edge.)
type operationBuffer_RightmostEdgeFinder struct {
	minIndex   int
	minCoord   *Geom_Coordinate
	minDe      *Geomgraph_DirectedEdge
	orientedDe *Geomgraph_DirectedEdge
}

// operationBuffer_newRightmostEdgeFinder creates a RightmostEdgeFinder.
// A RightmostEdgeFinder finds the DirectedEdge with the rightmost coordinate.
// The DirectedEdge returned is guaranteed to have the R of the world on its RHS.
func operationBuffer_newRightmostEdgeFinder() *operationBuffer_RightmostEdgeFinder {
	return &operationBuffer_RightmostEdgeFinder{
		minIndex: -1,
	}
}

// GetEdge returns the rightmost edge.
func (ref *operationBuffer_RightmostEdgeFinder) GetEdge() *Geomgraph_DirectedEdge {
	return ref.orientedDe
}

// GetCoordinate returns the coordinate of the rightmost edge.
func (ref *operationBuffer_RightmostEdgeFinder) GetCoordinate() *Geom_Coordinate {
	return ref.minCoord
}

// FindEdge finds the rightmost edge in the list.
func (ref *operationBuffer_RightmostEdgeFinder) FindEdge(dirEdgeList []*Geomgraph_DirectedEdge) {
	// Check all forward DirectedEdges only. This is still general,
	// because each edge has a forward DirectedEdge.
	for _, de := range dirEdgeList {
		if !de.IsForward() {
			continue
		}
		ref.checkForRightmostCoordinate(de)
	}

	// If the rightmost point is a node, we need to identify which of
	// the incident edges is rightmost.
	Util_Assert_IsTrueWithMessage(ref.minIndex != 0 || ref.minCoord.Equals(ref.minDe.GetCoordinate()), "inconsistency in rightmost processing")
	if ref.minIndex == 0 {
		ref.findRightmostEdgeAtNode()
	} else {
		ref.findRightmostEdgeAtVertex()
	}
	// now check that the extreme side is the R side.
	// If not, use the sym instead.
	ref.orientedDe = ref.minDe
	rightmostSide := ref.getRightmostSide(ref.minDe, ref.minIndex)
	if rightmostSide == Geom_Position_Left {
		ref.orientedDe = ref.minDe.GetSym()
	}
}

func (ref *operationBuffer_RightmostEdgeFinder) findRightmostEdgeAtNode() {
	node := ref.minDe.GetNode()
	star := java.Cast[*Geomgraph_DirectedEdgeStar](node.GetEdges())
	ref.minDe = star.GetRightmostEdge()
	// the DirectedEdge returned by the previous call is not
	// necessarily in the forward direction. Use the sym edge if it isn't.
	if !ref.minDe.IsForward() {
		ref.minDe = ref.minDe.GetSym()
		ref.minIndex = len(ref.minDe.GetEdge().GetCoordinates()) - 1
	}
}

func (ref *operationBuffer_RightmostEdgeFinder) findRightmostEdgeAtVertex() {
	// The rightmost point is an interior vertex, so it has a segment on either side of it.
	// If these segments are both above or below the rightmost point, we need to
	// determine their relative orientation to decide which is rightmost.
	pts := ref.minDe.GetEdge().GetCoordinates()
	if ref.minIndex <= 0 || ref.minIndex >= len(pts) {
		panic("rightmost point expected to be interior vertex of edge")
	}
	pPrev := pts[ref.minIndex-1]
	pNext := pts[ref.minIndex+1]
	orientation := Algorithm_Orientation_Index(ref.minCoord, pNext, pPrev)
	usePrev := false
	// both segments are below min point
	if pPrev.Y < ref.minCoord.Y && pNext.Y < ref.minCoord.Y &&
		orientation == Algorithm_Orientation_Counterclockwise {
		usePrev = true
	} else if pPrev.Y > ref.minCoord.Y && pNext.Y > ref.minCoord.Y &&
		orientation == Algorithm_Orientation_Clockwise {
		usePrev = true
	}
	// if both segments are on the same side, do nothing - either is safe
	// to select as a rightmost segment
	if usePrev {
		ref.minIndex = ref.minIndex - 1
	}
}

func (ref *operationBuffer_RightmostEdgeFinder) checkForRightmostCoordinate(de *Geomgraph_DirectedEdge) {
	coord := de.GetEdge().GetCoordinates()
	for i := 0; i < len(coord)-1; i++ {
		// only check vertices which are the start or end point of a non-horizontal segment
		// <FIX> MD 19 Sep 03 - NO! we can test all vertices, since the rightmost must have a non-horiz segment adjacent to it
		if ref.minCoord == nil || coord[i].X > ref.minCoord.X {
			ref.minDe = de
			ref.minIndex = i
			ref.minCoord = coord[i]
		}
	}
}

func (ref *operationBuffer_RightmostEdgeFinder) getRightmostSide(de *Geomgraph_DirectedEdge, index int) int {
	side := ref.getRightmostSideOfSegment(de, index)
	if side < 0 {
		side = ref.getRightmostSideOfSegment(de, index-1)
	}
	if side < 0 {
		// reaching here can indicate that segment is horizontal
		// testing only
		ref.minCoord = nil
		ref.checkForRightmostCoordinate(de)
	}
	return side
}

func (ref *operationBuffer_RightmostEdgeFinder) getRightmostSideOfSegment(de *Geomgraph_DirectedEdge, i int) int {
	e := de.GetEdge()
	coord := e.GetCoordinates()

	if i < 0 || i+1 >= len(coord) {
		return -1
	}
	if coord[i].Y == coord[i+1].Y {
		return -1 // indicates edge is parallel to x-axis
	}

	pos := Geom_Position_Left
	if coord[i].Y < coord[i+1].Y {
		pos = Geom_Position_Right
	}
	return pos
}
