package jts

// Type codes for PackedCoordinateSequenceFactory.
const (
	// GeomImpl_PackedCoordinateSequenceFactory_DOUBLE is the type code for arrays of type double.
	GeomImpl_PackedCoordinateSequenceFactory_DOUBLE = 0
	// GeomImpl_PackedCoordinateSequenceFactory_FLOAT is the type code for arrays of type float.
	GeomImpl_PackedCoordinateSequenceFactory_FLOAT = 1
)

const (
	geomImpl_PackedCoordinateSequenceFactory_DEFAULT_MEASURES  = 0
	geomImpl_PackedCoordinateSequenceFactory_DEFAULT_DIMENSION = 3
)

// GeomImpl_PackedCoordinateSequenceFactory_DOUBLE_FACTORY is a factory using array type DOUBLE.
var GeomImpl_PackedCoordinateSequenceFactory_DOUBLE_FACTORY = func() *GeomImpl_PackedCoordinateSequenceFactory {
	return GeomImpl_NewPackedCoordinateSequenceFactoryWithType(GeomImpl_PackedCoordinateSequenceFactory_DOUBLE)
}()

// GeomImpl_PackedCoordinateSequenceFactory_FLOAT_FACTORY is a factory using array type FLOAT.
var GeomImpl_PackedCoordinateSequenceFactory_FLOAT_FACTORY = func() *GeomImpl_PackedCoordinateSequenceFactory {
	return GeomImpl_NewPackedCoordinateSequenceFactoryWithType(GeomImpl_PackedCoordinateSequenceFactory_FLOAT)
}()

// GeomImpl_PackedCoordinateSequenceFactory builds packed array coordinate sequences.
// The array data type can be either double or float, and defaults to double.
type GeomImpl_PackedCoordinateSequenceFactory struct {
	typ int
}

var _ Geom_CoordinateSequenceFactory = (*GeomImpl_PackedCoordinateSequenceFactory)(nil)

func (f *GeomImpl_PackedCoordinateSequenceFactory) IsGeom_CoordinateSequenceFactory() {}

// GeomImpl_NewPackedCoordinateSequenceFactory creates a new factory of type DOUBLE.
func GeomImpl_NewPackedCoordinateSequenceFactory() *GeomImpl_PackedCoordinateSequenceFactory {
	return GeomImpl_NewPackedCoordinateSequenceFactoryWithType(GeomImpl_PackedCoordinateSequenceFactory_DOUBLE)
}

// GeomImpl_NewPackedCoordinateSequenceFactoryWithType creates a new factory of the given type.
func GeomImpl_NewPackedCoordinateSequenceFactoryWithType(t int) *GeomImpl_PackedCoordinateSequenceFactory {
	return &GeomImpl_PackedCoordinateSequenceFactory{typ: t}
}

// GetType returns the type of packed coordinate sequence this factory builds.
func (f *GeomImpl_PackedCoordinateSequenceFactory) GetType() int {
	return f.typ
}

// CreateFromCoordinates creates a coordinate sequence from an array of coordinates.
func (f *GeomImpl_PackedCoordinateSequenceFactory) CreateFromCoordinates(coordinates []*Geom_Coordinate) Geom_CoordinateSequence {
	dimension := geomImpl_PackedCoordinateSequenceFactory_DEFAULT_DIMENSION
	measures := geomImpl_PackedCoordinateSequenceFactory_DEFAULT_MEASURES
	if len(coordinates) > 0 && coordinates[0] != nil {
		first := coordinates[0]
		dimension = Geom_Coordinates_Dimension(first)
		measures = Geom_Coordinates_Measures(first)
	}
	if f.typ == GeomImpl_PackedCoordinateSequenceFactory_DOUBLE {
		return GeomImpl_NewPackedCoordinateSequenceDoubleFromCoordinates(coordinates, dimension, measures)
	}
	return GeomImpl_NewPackedCoordinateSequenceFloatFromCoordinates(coordinates, dimension, measures)
}

// CreateFromCoordinateSequence creates a coordinate sequence from an existing sequence.
func (f *GeomImpl_PackedCoordinateSequenceFactory) CreateFromCoordinateSequence(coordSeq Geom_CoordinateSequence) Geom_CoordinateSequence {
	dimension := coordSeq.GetDimension()
	measures := coordSeq.GetMeasures()
	if f.typ == GeomImpl_PackedCoordinateSequenceFactory_DOUBLE {
		return GeomImpl_NewPackedCoordinateSequenceDoubleFromCoordinates(coordSeq.ToCoordinateArray(), dimension, measures)
	}
	return GeomImpl_NewPackedCoordinateSequenceFloatFromCoordinates(coordSeq.ToCoordinateArray(), dimension, measures)
}

// CreateFromDoubles creates a packed coordinate sequence from the provided double array.
func (f *GeomImpl_PackedCoordinateSequenceFactory) CreateFromDoubles(packedCoordinates []float64, dimension int) Geom_CoordinateSequence {
	return f.CreateFromDoublesWithMeasures(packedCoordinates, dimension, geomImpl_PackedCoordinateSequenceFactory_DEFAULT_MEASURES)
}

// CreateFromDoublesWithMeasures creates a packed coordinate sequence from the provided
// double array with explicit measures.
func (f *GeomImpl_PackedCoordinateSequenceFactory) CreateFromDoublesWithMeasures(packedCoordinates []float64, dimension, measures int) Geom_CoordinateSequence {
	if f.typ == GeomImpl_PackedCoordinateSequenceFactory_DOUBLE {
		return GeomImpl_NewPackedCoordinateSequenceDoubleFromDoubles(packedCoordinates, dimension, measures)
	}
	return GeomImpl_NewPackedCoordinateSequenceFloatFromDoubles(packedCoordinates, dimension, measures)
}

// CreateFromFloats creates a packed coordinate sequence from the provided float array.
func (f *GeomImpl_PackedCoordinateSequenceFactory) CreateFromFloats(packedCoordinates []float32, dimension int) Geom_CoordinateSequence {
	measures := geomImpl_PackedCoordinateSequenceFactory_DEFAULT_MEASURES
	if dimension > 3 {
		measures = dimension - 3
	}
	return f.CreateFromFloatsWithMeasures(packedCoordinates, dimension, measures)
}

// CreateFromFloatsWithMeasures creates a packed coordinate sequence from the provided
// float array with explicit measures.
func (f *GeomImpl_PackedCoordinateSequenceFactory) CreateFromFloatsWithMeasures(packedCoordinates []float32, dimension, measures int) Geom_CoordinateSequence {
	if f.typ == GeomImpl_PackedCoordinateSequenceFactory_DOUBLE {
		return GeomImpl_NewPackedCoordinateSequenceDoubleFromFloats(packedCoordinates, dimension, measures)
	}
	return GeomImpl_NewPackedCoordinateSequenceFloatFromFloats(packedCoordinates, dimension, measures)
}

// CreateWithSizeAndDimension creates an empty packed coordinate sequence of a given size and dimension.
func (f *GeomImpl_PackedCoordinateSequenceFactory) CreateWithSizeAndDimension(size, dimension int) Geom_CoordinateSequence {
	measures := geomImpl_PackedCoordinateSequenceFactory_DEFAULT_MEASURES
	if dimension > 3 {
		measures = dimension - 3
	}
	if f.typ == GeomImpl_PackedCoordinateSequenceFactory_DOUBLE {
		return GeomImpl_NewPackedCoordinateSequenceDoubleWithSize(size, dimension, measures)
	}
	return GeomImpl_NewPackedCoordinateSequenceFloatWithSize(size, dimension, measures)
}

// CreateWithSizeAndDimensionAndMeasures creates an empty packed coordinate sequence of
// a given size, dimension, and measures.
func (f *GeomImpl_PackedCoordinateSequenceFactory) CreateWithSizeAndDimensionAndMeasures(size, dimension, measures int) Geom_CoordinateSequence {
	if f.typ == GeomImpl_PackedCoordinateSequenceFactory_DOUBLE {
		return GeomImpl_NewPackedCoordinateSequenceDoubleWithSize(size, dimension, measures)
	}
	return GeomImpl_NewPackedCoordinateSequenceFloatWithSize(size, dimension, measures)
}
