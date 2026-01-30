package jts

import (
	"fmt"
	"io"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geomgraph_EdgeIntersection represents a point on an edge which intersects
// with another edge.
//
// The intersection may either be a single point, or a line segment (in which
// case this point is the start of the line segment). The intersection point
// must be precise.
type Geomgraph_EdgeIntersection struct {
	child java.Polymorphic

	// Coord is the point of intersection.
	Coord *Geom_Coordinate

	// SegmentIndex is the index of the containing line segment in the parent
	// edge.
	SegmentIndex int

	// Dist is the edge distance of this point along the containing line
	// segment.
	Dist float64
}

// GetChild returns the immediate child in the type hierarchy chain.
func (ei *Geomgraph_EdgeIntersection) GetChild() java.Polymorphic {
	return ei.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (ei *Geomgraph_EdgeIntersection) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewEdgeIntersection creates a new EdgeIntersection.
func Geomgraph_NewEdgeIntersection(coord *Geom_Coordinate, segmentIndex int, dist float64) *Geomgraph_EdgeIntersection {
	return &Geomgraph_EdgeIntersection{
		Coord:        coord.Copy(),
		SegmentIndex: segmentIndex,
		Dist:         dist,
	}
}

// GetCoordinate returns the point of intersection.
func (ei *Geomgraph_EdgeIntersection) GetCoordinate() *Geom_Coordinate {
	return ei.Coord
}

// GetSegmentIndex returns the index of the containing line segment.
func (ei *Geomgraph_EdgeIntersection) GetSegmentIndex() int {
	return ei.SegmentIndex
}

// GetDistance returns the edge distance of this point along the containing
// line segment.
func (ei *Geomgraph_EdgeIntersection) GetDistance() float64 {
	return ei.Dist
}

// CompareTo compares this EdgeIntersection to another.
func (ei *Geomgraph_EdgeIntersection) CompareTo(other *Geomgraph_EdgeIntersection) int {
	return ei.Compare(other.SegmentIndex, other.Dist)
}

// Compare compares this EdgeIntersection to a segment index and distance.
// Returns -1 if this EdgeIntersection is located before the argument location,
// 0 if at the argument location, or 1 if located after the argument location.
func (ei *Geomgraph_EdgeIntersection) Compare(segmentIndex int, dist float64) int {
	if ei.SegmentIndex < segmentIndex {
		return -1
	}
	if ei.SegmentIndex > segmentIndex {
		return 1
	}
	if ei.Dist < dist {
		return -1
	}
	if ei.Dist > dist {
		return 1
	}
	return 0
}

// IsEndPoint returns true if this intersection is at an endpoint of the edge.
func (ei *Geomgraph_EdgeIntersection) IsEndPoint(maxSegmentIndex int) bool {
	if ei.SegmentIndex == 0 && ei.Dist == 0.0 {
		return true
	}
	if ei.SegmentIndex == maxSegmentIndex {
		return true
	}
	return false
}

// String returns a string representation of this EdgeIntersection.
func (ei *Geomgraph_EdgeIntersection) String() string {
	return fmt.Sprintf("%v seg # = %d dist = %v", ei.Coord, ei.SegmentIndex, ei.Dist)
}

// Print writes a representation to the given writer.
func (ei *Geomgraph_EdgeIntersection) Print(out io.Writer) {
	io.WriteString(out, ei.String()+"\n")
}
