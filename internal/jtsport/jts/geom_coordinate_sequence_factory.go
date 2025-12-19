package jts

// Geom_CoordinateSequenceFactory is a factory to create concrete instances of
// Geom_CoordinateSequence. Used to configure GeometryFactory to provide specific
// kinds of Geom_CoordinateSequences.
type Geom_CoordinateSequenceFactory interface {
	IsGeom_CoordinateSequenceFactory()

	// CreateFromCoordinates returns a Geom_CoordinateSequence based on the given array.
	// Whether the array is copied or simply referenced is implementation-dependent.
	// This method must handle nil arguments by creating an empty sequence.
	CreateFromCoordinates(coordinates []*Geom_Coordinate) Geom_CoordinateSequence

	// CreateFromCoordinateSequence creates a Geom_CoordinateSequence which is a copy of
	// the given Geom_CoordinateSequence. This method must handle nil arguments by
	// creating an empty sequence.
	CreateFromCoordinateSequence(coordSeq Geom_CoordinateSequence) Geom_CoordinateSequence

	// CreateWithSizeAndDimension creates a Geom_CoordinateSequence of the specified size
	// and dimension. For this to be useful, the Geom_CoordinateSequence implementation
	// must be mutable.
	//
	// If the requested dimension is larger than the Geom_CoordinateSequence
	// implementation can provide, then a sequence of maximum possible dimension
	// should be created. An error should not be thrown.
	CreateWithSizeAndDimension(size, dimension int) Geom_CoordinateSequence

	// CreateWithSizeAndDimensionAndMeasures creates a Geom_CoordinateSequence of the
	// specified size and dimension with measure support. For this to be useful, the
	// Geom_CoordinateSequence implementation must be mutable.
	//
	// If the requested dimension or measures are larger than the Geom_CoordinateSequence
	// implementation can provide, then a sequence of maximum possible dimension
	// should be created. An error should not be thrown.
	CreateWithSizeAndDimensionAndMeasures(size, dimension, measures int) Geom_CoordinateSequence
}
