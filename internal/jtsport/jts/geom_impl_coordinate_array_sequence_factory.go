package jts

var _ Geom_CoordinateSequenceFactory = (*GeomImpl_CoordinateArraySequenceFactory)(nil)

// GeomImpl_CoordinateArraySequenceFactory creates CoordinateSequences represented as an
// array of Coordinates.
type GeomImpl_CoordinateArraySequenceFactory struct{}

func (f *GeomImpl_CoordinateArraySequenceFactory) IsGeom_CoordinateSequenceFactory() {}

// geomImpl_CoordinateArraySequenceFactory_instance is the singleton instance.
var geomImpl_CoordinateArraySequenceFactory_instance = func() *GeomImpl_CoordinateArraySequenceFactory {
	casf := &GeomImpl_CoordinateArraySequenceFactory{}
	// Register this as the default factory.
	Geom_SetDefaultCoordinateSequenceFactory(casf)
	return casf
}()

// GeomImpl_CoordinateArraySequenceFactory_Instance returns the singleton instance of CoordinateArraySequenceFactory.
func GeomImpl_CoordinateArraySequenceFactory_Instance() *GeomImpl_CoordinateArraySequenceFactory {
	return geomImpl_CoordinateArraySequenceFactory_instance
}

func (f *GeomImpl_CoordinateArraySequenceFactory) CreateFromCoordinates(coordinates []*Geom_Coordinate) Geom_CoordinateSequence {
	return GeomImpl_NewCoordinateArraySequence(coordinates)
}

func (f *GeomImpl_CoordinateArraySequenceFactory) CreateFromCoordinateSequence(coordSeq Geom_CoordinateSequence) Geom_CoordinateSequence {
	return GeomImpl_NewCoordinateArraySequenceFromCoordinateSequence(coordSeq)
}

func (f *GeomImpl_CoordinateArraySequenceFactory) CreateWithSizeAndDimension(size, dimension int) Geom_CoordinateSequence {
	// Clip dimension to range [2, 3].
	if dimension > 3 {
		dimension = 3
	}
	if dimension < 2 {
		dimension = 2
	}
	return GeomImpl_NewCoordinateArraySequenceWithSizeAndDimension(size, dimension)
}

func (f *GeomImpl_CoordinateArraySequenceFactory) CreateWithSizeAndDimensionAndMeasures(size, dimension, measures int) Geom_CoordinateSequence {
	spatial := dimension - measures

	// Clip measures to max 1.
	if measures > 1 {
		measures = 1
	}
	// Clip spatial dimension to max 3.
	if spatial > 3 {
		spatial = 3
	}
	// Handle bogus spatial dimension.
	if spatial < 2 {
		spatial = 2
	}

	return GeomImpl_NewCoordinateArraySequenceWithSizeDimensionAndMeasures(size, spatial+measures, measures)
}
