package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// coordinates provides utility functions for handling Geom_Coordinate objects.

// Geom_Coordinates_Create is a factory method providing access to common Geom_Coordinate
// implementations.
func Geom_Coordinates_Create(dimension int) *Geom_Coordinate {
	return Geom_Coordinates_CreateWithMeasures(dimension, 0)
}

// Geom_Coordinates_CreateWithMeasures is a factory method providing access to common Geom_Coordinate
// implementations.
func Geom_Coordinates_CreateWithMeasures(dimension, measures int) *Geom_Coordinate {
	if dimension == 2 {
		return Geom_NewCoordinateXY2D().Geom_Coordinate
	} else if dimension == 3 && measures == 0 {
		return Geom_NewCoordinate()
	} else if dimension == 3 && measures == 1 {
		return Geom_NewCoordinateXYM3D().Geom_Coordinate
	} else if dimension == 4 && measures == 1 {
		return Geom_NewCoordinateXYZM4D().Geom_Coordinate
	}
	return Geom_NewCoordinate()
}

// Geom_Coordinates_Dimension determines the dimension based on the concrete type of Geom_Coordinate.
func Geom_Coordinates_Dimension(coordinate *Geom_Coordinate) int {
	if java.InstanceOf[*Geom_CoordinateXY](coordinate) {
		return 2
	} else if java.InstanceOf[*Geom_CoordinateXYM](coordinate) {
		return 3
	} else if java.InstanceOf[*Geom_CoordinateXYZM](coordinate) {
		return 4
	} else if java.InstanceOf[*Geom_Coordinate](coordinate) {
		return 3
	}
	return 3
}

// Geom_Coordinates_HasZ checks if the coordinate can store a Z value, based on the concrete
// type of Geom_Coordinate.
func Geom_Coordinates_HasZ(coordinate *Geom_Coordinate) bool {
	if java.InstanceOf[*Geom_CoordinateXY](coordinate) {
		return false
	} else if java.InstanceOf[*Geom_CoordinateXYM](coordinate) {
		return false
	} else if java.InstanceOf[*Geom_CoordinateXYZM](coordinate) {
		return true
	} else if java.InstanceOf[*Geom_Coordinate](coordinate) {
		return true
	}
	return true
}

// Geom_Coordinates_Measures determines the number of measures based on the concrete type of
// Geom_Coordinate.
func Geom_Coordinates_Measures(coordinate *Geom_Coordinate) int {
	if java.InstanceOf[*Geom_CoordinateXY](coordinate) {
		return 0
	} else if java.InstanceOf[*Geom_CoordinateXYM](coordinate) {
		return 1
	} else if java.InstanceOf[*Geom_CoordinateXYZM](coordinate) {
		return 1
	} else if java.InstanceOf[*Geom_Coordinate](coordinate) {
		return 0
	}
	return 0
}
