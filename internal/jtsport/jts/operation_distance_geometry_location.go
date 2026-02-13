package jts

import "strconv"

// OperationDistance_GeometryLocation_INSIDE_AREA is a special value of
// segmentIndex used for locations inside area geometries. These locations are
// not located on a segment, and thus do not have an associated segment index.
const OperationDistance_GeometryLocation_INSIDE_AREA = -1

// OperationDistance_GeometryLocation represents the location of a point on a
// Geometry. Maintains both the actual point location (which may not be exact,
// if the point is not a vertex) as well as information about the component and
// segment index where the point occurs. Locations inside area Geometries will
// not have an associated segment index, so in this case the segment index will
// have the sentinel value of INSIDE_AREA.
type OperationDistance_GeometryLocation struct {
	component *Geom_Geometry
	segIndex  int
	pt        *Geom_Coordinate
}

// OperationDistance_NewGeometryLocation constructs a GeometryLocation
// specifying a point on a geometry, as well as the segment that the point is
// on (or INSIDE_AREA if the point is not on a segment).
func OperationDistance_NewGeometryLocation(component *Geom_Geometry, segIndex int, pt *Geom_Coordinate) *OperationDistance_GeometryLocation {
	return &OperationDistance_GeometryLocation{
		component: component,
		segIndex:  segIndex,
		pt:        pt,
	}
}

// OperationDistance_NewGeometryLocationInsideArea constructs a GeometryLocation
// specifying a point inside an area geometry.
func OperationDistance_NewGeometryLocationInsideArea(component *Geom_Geometry, pt *Geom_Coordinate) *OperationDistance_GeometryLocation {
	return OperationDistance_NewGeometryLocation(component, OperationDistance_GeometryLocation_INSIDE_AREA, pt)
}

// GetGeometryComponent returns the geometry component on (or in) which this
// location occurs.
func (gl *OperationDistance_GeometryLocation) GetGeometryComponent() *Geom_Geometry {
	return gl.component
}

// GetSegmentIndex returns the segment index for this location. If the location
// is inside an area, the index will have the value INSIDE_AREA.
func (gl *OperationDistance_GeometryLocation) GetSegmentIndex() int {
	return gl.segIndex
}

// GetCoordinate returns the Coordinate of this location.
func (gl *OperationDistance_GeometryLocation) GetCoordinate() *Geom_Coordinate {
	return gl.pt
}

// IsInsideArea tests whether this location represents a point inside an area
// geometry.
func (gl *OperationDistance_GeometryLocation) IsInsideArea() bool {
	return gl.segIndex == OperationDistance_GeometryLocation_INSIDE_AREA
}

// String returns a string representation of this GeometryLocation.
func (gl *OperationDistance_GeometryLocation) String() string {
	return gl.component.GetGeometryType() +
		"[" + strconv.Itoa(gl.segIndex) + "]" +
		"-" + Io_WKTWriter_ToPoint(gl.pt)
}
