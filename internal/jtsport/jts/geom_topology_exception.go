package jts

import "fmt"

// Geom_TopologyException indicates an invalid or inconsistent topological situation
// encountered during processing.
type Geom_TopologyException struct {
	msg string
	pt  *Geom_Coordinate
}

// Geom_NewTopologyException creates a TopologyException with the given message.
func Geom_NewTopologyException(msg string) *Geom_TopologyException {
	return &Geom_TopologyException{msg: msg}
}

// Geom_NewTopologyExceptionWithCoordinate creates a TopologyException with the given
// message and coordinate.
func Geom_NewTopologyExceptionWithCoordinate(msg string, pt *Geom_Coordinate) *Geom_TopologyException {
	return &Geom_TopologyException{
		msg: geom_TopologyException_msgWithCoord(msg, pt),
		pt:  Geom_NewCoordinateFromCoordinate(pt),
	}
}

// geom_TopologyException_msgWithCoord formats a message with a coordinate appended.
func geom_TopologyException_msgWithCoord(msg string, pt *Geom_Coordinate) string {
	if pt != nil {
		return fmt.Sprintf("%s [ %s ]", msg, pt.String())
	}
	return msg
}

// Error implements the error interface.
func (te *Geom_TopologyException) Error() string {
	return te.msg
}

// GetCoordinate returns the coordinate associated with this exception, or nil
// if none was provided.
func (te *Geom_TopologyException) GetCoordinate() *Geom_Coordinate {
	return te.pt
}
