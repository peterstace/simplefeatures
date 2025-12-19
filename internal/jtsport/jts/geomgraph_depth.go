package jts

import (
	"fmt"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const geomgraph_Depth_NullValue = -1

// Geomgraph_Depth_DepthAtLocation returns the depth value for a given
// location.
func Geomgraph_Depth_DepthAtLocation(location int) int {
	if location == Geom_Location_Exterior {
		return 0
	}
	if location == Geom_Location_Interior {
		return 1
	}
	return geomgraph_Depth_NullValue
}

// Geomgraph_Depth records the topological depth of the sides of an Edge for up
// to two Geometries.
type Geomgraph_Depth struct {
	child java.Polymorphic
	depth [2][3]int
}

// GetChild returns the immediate child in the type hierarchy chain.
func (d *Geomgraph_Depth) GetChild() java.Polymorphic {
	return d.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (d *Geomgraph_Depth) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewDepth creates a new Depth object.
func Geomgraph_NewDepth() *Geomgraph_Depth {
	d := &Geomgraph_Depth{}
	// Initialize depth array to a sentinel value.
	for i := 0; i < 2; i++ {
		for j := 0; j < 3; j++ {
			d.depth[i][j] = geomgraph_Depth_NullValue
		}
	}
	return d
}

// GetDepth returns the depth value at the given geometry and position indices.
func (d *Geomgraph_Depth) GetDepth(geomIndex, posIndex int) int {
	return d.depth[geomIndex][posIndex]
}

// SetDepth sets the depth value at the given geometry and position indices.
func (d *Geomgraph_Depth) SetDepth(geomIndex, posIndex, depthValue int) {
	d.depth[geomIndex][posIndex] = depthValue
}

// GetLocation returns the location (INTERIOR or EXTERIOR) based on depth.
func (d *Geomgraph_Depth) GetLocation(geomIndex, posIndex int) int {
	if d.depth[geomIndex][posIndex] <= 0 {
		return Geom_Location_Exterior
	}
	return Geom_Location_Interior
}

// Add increments the depth at the given position if the location is INTERIOR.
func (d *Geomgraph_Depth) Add(geomIndex, posIndex, location int) {
	if location == Geom_Location_Interior {
		d.depth[geomIndex][posIndex]++
	}
}

// IsNull returns true if this Depth object has never been initialized (all
// depths are null).
func (d *Geomgraph_Depth) IsNull() bool {
	for i := 0; i < 2; i++ {
		for j := 0; j < 3; j++ {
			if d.depth[i][j] != geomgraph_Depth_NullValue {
				return false
			}
		}
	}
	return true
}

// IsNullAt returns true if the depth for the given geometry index is null.
func (d *Geomgraph_Depth) IsNullAt(geomIndex int) bool {
	return d.depth[geomIndex][1] == geomgraph_Depth_NullValue
}

// IsNullAtPos returns true if the depth at the given geometry and position
// indices is null.
func (d *Geomgraph_Depth) IsNullAtPos(geomIndex, posIndex int) bool {
	return d.depth[geomIndex][posIndex] == geomgraph_Depth_NullValue
}

// AddLabel adds the depths from a Label.
func (d *Geomgraph_Depth) AddLabel(lbl *Geomgraph_Label) {
	for i := 0; i < 2; i++ {
		for j := 1; j < 3; j++ {
			loc := lbl.GetLocation(i, j)
			if loc == Geom_Location_Exterior || loc == Geom_Location_Interior {
				// Initialize depth if it is null, otherwise add this location
				// value.
				if d.IsNullAtPos(i, j) {
					d.depth[i][j] = Geomgraph_Depth_DepthAtLocation(loc)
				} else {
					d.depth[i][j] += Geomgraph_Depth_DepthAtLocation(loc)
				}
			}
		}
	}
}

// GetDelta returns the difference between the right and left depths for the
// given geometry index.
func (d *Geomgraph_Depth) GetDelta(geomIndex int) int {
	return d.depth[geomIndex][Geom_Position_Right] - d.depth[geomIndex][Geom_Position_Left]
}

// Normalize normalizes the depths for each geometry, if they are non-null. A
// normalized depth has depth values in the set { 0, 1 }. Normalizing the
// depths involves reducing the depths by the same amount so that at least one
// of them is 0. If the remaining value is > 0, it is set to 1.
func (d *Geomgraph_Depth) Normalize() {
	for i := 0; i < 2; i++ {
		if !d.IsNullAt(i) {
			minDepth := d.depth[i][1]
			if d.depth[i][2] < minDepth {
				minDepth = d.depth[i][2]
			}

			if minDepth < 0 {
				minDepth = 0
			}
			for j := 1; j < 3; j++ {
				newValue := 0
				if d.depth[i][j] > minDepth {
					newValue = 1
				}
				d.depth[i][j] = newValue
			}
		}
	}
}

// String returns a string representation of this Depth.
func (d *Geomgraph_Depth) String() string {
	return fmt.Sprintf("A: %d,%d B: %d,%d",
		d.depth[0][1], d.depth[0][2],
		d.depth[1][1], d.depth[1][2])
}
