package jts

import "fmt"

// Noding_SegmentNode represents an intersection point between two
// SegmentStrings.
type Noding_SegmentNode struct {
	segString     *Noding_NodedSegmentString
	Coord          *Geom_Coordinate // The point of intersection.
	SegmentIndex   int              // The index of the containing line segment in the parent edge.
	segmentOctant int
	isInterior    bool
}

// Noding_NewSegmentNode creates a new SegmentNode.
func Noding_NewSegmentNode(segString *Noding_NodedSegmentString, coord *Geom_Coordinate, segmentIndex, segmentOctant int) *Noding_SegmentNode {
	coordCopy := coord.Copy()
	isInterior := !coord.Equals2D(segString.GetCoordinate(segmentIndex))
	return &Noding_SegmentNode{
		segString:     segString,
		Coord:          coordCopy,
		SegmentIndex:   segmentIndex,
		segmentOctant: segmentOctant,
		isInterior:    isInterior,
	}
}

// GetCoordinate gets the Coordinate giving the location of this node.
func (sn *Noding_SegmentNode) GetCoordinate() *Geom_Coordinate {
	return sn.Coord
}

// IsInterior returns whether this node is in the interior of its segment (not
// at a vertex).
func (sn *Noding_SegmentNode) IsInterior() bool {
	return sn.isInterior
}

// IsEndPoint returns whether this node is an endpoint of the segment string.
func (sn *Noding_SegmentNode) IsEndPoint(maxSegmentIndex int) bool {
	if sn.SegmentIndex == 0 && !sn.isInterior {
		return true
	}
	if sn.SegmentIndex == maxSegmentIndex {
		return true
	}
	return false
}

// CompareTo compares this SegmentNode with another.
//
// Returns -1 if this SegmentNode is located before the argument location, 0 if
// this SegmentNode is at the argument location, 1 if this SegmentNode is
// located after the argument location.
func (sn *Noding_SegmentNode) CompareTo(other *Noding_SegmentNode) int {
	if sn.SegmentIndex < other.SegmentIndex {
		return -1
	}
	if sn.SegmentIndex > other.SegmentIndex {
		return 1
	}

	if sn.Coord.Equals2D(other.Coord) {
		return 0
	}

	// An exterior node is the segment start point, so always sorts first. This
	// guards against a robustness problem where the octants are not reliable.
	if !sn.isInterior {
		return -1
	}
	if !other.isInterior {
		return 1
	}

	return Noding_SegmentPointComparator_Compare(sn.segmentOctant, sn.Coord, other.Coord)
}

// String returns a string representation of this SegmentNode.
func (sn *Noding_SegmentNode) String() string {
	return fmt.Sprintf("%d:%s", sn.SegmentIndex, sn.Coord.String())
}
