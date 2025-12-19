package jts

import (
	"math"
	"strings"
)

var _ Geom_CoordinateSequence = (*GeomImpl_CoordinateArraySequence)(nil)

// GeomImpl_CoordinateArraySequence is a CoordinateSequence backed by an array of
// Coordinates. This is the implementation that Geometrys use by default.
// Coordinates returned by ToCoordinateArray and GetCoordinate are live --
// modifications to them are actually changing the CoordinateSequence's
// underlying data. A dimension may be specified for the coordinates in the
// sequence, which may be 2 or 3. The actual coordinates will always have 3
// ordinates, but the dimension is useful as metadata in some situations.
type GeomImpl_CoordinateArraySequence struct {
	dimension   int
	measures    int
	coordinates []*Geom_Coordinate
}

func (cas *GeomImpl_CoordinateArraySequence) IsGeom_CoordinateSequence() {}

// GeomImpl_NewCoordinateArraySequence constructs a sequence based on the given array of
// Coordinates (the array is not copied). The coordinate dimension defaults to
// 3.
func GeomImpl_NewCoordinateArraySequence(coordinates []*Geom_Coordinate) *GeomImpl_CoordinateArraySequence {
	return GeomImpl_NewCoordinateArraySequenceWithDimensionAndMeasures(
		coordinates,
		Geom_CoordinateArrays_Dimension(coordinates),
		Geom_CoordinateArrays_Measures(coordinates),
	)
}

// GeomImpl_NewCoordinateArraySequenceWithDimension constructs a sequence based on the
// given array of Coordinates (the array is not copied).
func GeomImpl_NewCoordinateArraySequenceWithDimension(coordinates []*Geom_Coordinate, dimension int) *GeomImpl_CoordinateArraySequence {
	return GeomImpl_NewCoordinateArraySequenceWithDimensionAndMeasures(
		coordinates,
		dimension,
		Geom_CoordinateArrays_Measures(coordinates),
	)
}

// GeomImpl_NewCoordinateArraySequenceWithDimensionAndMeasures constructs a sequence
// based on the given array of Coordinates (the array is not copied). It is
// your responsibility to ensure the array contains Coordinates of the
// indicated dimension and measures.
func GeomImpl_NewCoordinateArraySequenceWithDimensionAndMeasures(coordinates []*Geom_Coordinate, dimension, measures int) *GeomImpl_CoordinateArraySequence {
	cas := &GeomImpl_CoordinateArraySequence{
		dimension: dimension,
		measures:  measures,
	}
	if coordinates == nil {
		cas.coordinates = []*Geom_Coordinate{}
	} else {
		cas.coordinates = coordinates
	}
	return cas
}

// GeomImpl_NewCoordinateArraySequenceWithSize constructs a sequence of a given size,
// populated with new Coordinates.
func GeomImpl_NewCoordinateArraySequenceWithSize(size int) *GeomImpl_CoordinateArraySequence {
	cas := &GeomImpl_CoordinateArraySequence{
		dimension:   3,
		measures:    0,
		coordinates: make([]*Geom_Coordinate, size),
	}
	for i := 0; i < size; i++ {
		cas.coordinates[i] = Geom_NewCoordinate()
	}
	return cas
}

// GeomImpl_NewCoordinateArraySequenceWithSizeAndDimension constructs a sequence of a
// given size, populated with new Coordinates.
func GeomImpl_NewCoordinateArraySequenceWithSizeAndDimension(size, dimension int) *GeomImpl_CoordinateArraySequence {
	cas := &GeomImpl_CoordinateArraySequence{
		dimension:   dimension,
		measures:    0,
		coordinates: make([]*Geom_Coordinate, size),
	}
	for i := 0; i < size; i++ {
		cas.coordinates[i] = Geom_Coordinates_Create(dimension)
	}
	return cas
}

// GeomImpl_NewCoordinateArraySequenceWithSizeDimensionAndMeasures constructs a sequence
// of a given size, populated with new Coordinates.
func GeomImpl_NewCoordinateArraySequenceWithSizeDimensionAndMeasures(size, dimension, measures int) *GeomImpl_CoordinateArraySequence {
	cas := &GeomImpl_CoordinateArraySequence{
		dimension:   dimension,
		measures:    measures,
		coordinates: make([]*Geom_Coordinate, size),
	}
	for i := 0; i < size; i++ {
		cas.coordinates[i] = cas.createCoordinate()
	}
	return cas
}

// GeomImpl_NewCoordinateArraySequenceFromCoordinateSequence creates a new sequence based
// on a deep copy of the given CoordinateSequence. The coordinate dimension is
// set to equal the dimension of the input.
func GeomImpl_NewCoordinateArraySequenceFromCoordinateSequence(coordSeq Geom_CoordinateSequence) *GeomImpl_CoordinateArraySequence {
	if coordSeq == nil {
		return &GeomImpl_CoordinateArraySequence{
			coordinates: []*Geom_Coordinate{},
			dimension:   3,
			measures:    0,
		}
	}
	cas := &GeomImpl_CoordinateArraySequence{
		dimension:   coordSeq.GetDimension(),
		measures:    coordSeq.GetMeasures(),
		coordinates: make([]*Geom_Coordinate, coordSeq.Size()),
	}
	for i := range cas.coordinates {
		cas.coordinates[i] = coordSeq.GetCoordinateCopy(i)
	}
	return cas
}

// createCoordinate creates a coordinate for use in this sequence.
func (cas *GeomImpl_CoordinateArraySequence) createCoordinate() *Geom_Coordinate {
	return Geom_Coordinates_CreateWithMeasures(cas.dimension, cas.measures)
}

func (cas *GeomImpl_CoordinateArraySequence) GetDimension() int {
	return cas.dimension
}

func (cas *GeomImpl_CoordinateArraySequence) GetMeasures() int {
	return cas.measures
}

func (cas *GeomImpl_CoordinateArraySequence) HasZ() bool {
	return (cas.GetDimension() - cas.GetMeasures()) > 2
}

func (cas *GeomImpl_CoordinateArraySequence) HasM() bool {
	return cas.GetMeasures() > 0
}

func (cas *GeomImpl_CoordinateArraySequence) CreateCoordinate() *Geom_Coordinate {
	return Geom_Coordinates_CreateWithMeasures(cas.GetDimension(), cas.GetMeasures())
}

func (cas *GeomImpl_CoordinateArraySequence) GetCoordinate(i int) *Geom_Coordinate {
	return cas.coordinates[i]
}

func (cas *GeomImpl_CoordinateArraySequence) GetCoordinateCopy(i int) *Geom_Coordinate {
	copyCoord := cas.createCoordinate()
	copyCoord.SetCoordinate(cas.coordinates[i])
	return copyCoord
}

func (cas *GeomImpl_CoordinateArraySequence) GetCoordinateInto(index int, coord *Geom_Coordinate) {
	coord.SetCoordinate(cas.coordinates[index])
}

func (cas *GeomImpl_CoordinateArraySequence) GetX(index int) float64 {
	return cas.coordinates[index].X
}

func (cas *GeomImpl_CoordinateArraySequence) GetY(index int) float64 {
	return cas.coordinates[index].Y
}

func (cas *GeomImpl_CoordinateArraySequence) GetZ(index int) float64 {
	if cas.HasZ() {
		return cas.coordinates[index].GetZ()
	}
	return math.NaN()
}

func (cas *GeomImpl_CoordinateArraySequence) GetM(index int) float64 {
	if cas.HasM() {
		return cas.coordinates[index].GetM()
	}
	return math.NaN()
}

func (cas *GeomImpl_CoordinateArraySequence) GetOrdinate(index, ordinateIndex int) float64 {
	switch ordinateIndex {
	case Geom_CoordinateSequence_X:
		return cas.coordinates[index].X
	case Geom_CoordinateSequence_Y:
		return cas.coordinates[index].Y
	default:
		return cas.coordinates[index].GetOrdinate(ordinateIndex)
	}
}

func (cas *GeomImpl_CoordinateArraySequence) Size() int {
	return len(cas.coordinates)
}

func (cas *GeomImpl_CoordinateArraySequence) SetOrdinate(index, ordinateIndex int, value float64) {
	switch ordinateIndex {
	case Geom_CoordinateSequence_X:
		cas.coordinates[index].X = value
	case Geom_CoordinateSequence_Y:
		cas.coordinates[index].Y = value
	default:
		cas.coordinates[index].SetOrdinate(ordinateIndex, value)
	}
}

func (cas *GeomImpl_CoordinateArraySequence) ToCoordinateArray() []*Geom_Coordinate {
	return cas.coordinates
}

func (cas *GeomImpl_CoordinateArraySequence) ExpandEnvelope(env *Geom_Envelope) *Geom_Envelope {
	for i := range cas.coordinates {
		env.ExpandToIncludeCoordinate(cas.coordinates[i])
	}
	return env
}

func (cas *GeomImpl_CoordinateArraySequence) Copy() Geom_CoordinateSequence {
	cloneCoordinates := make([]*Geom_Coordinate, cas.Size())
	for i := range cas.coordinates {
		duplicate := cas.createCoordinate()
		duplicate.SetCoordinate(cas.coordinates[i])
		cloneCoordinates[i] = duplicate
	}
	return GeomImpl_NewCoordinateArraySequenceWithDimensionAndMeasures(cloneCoordinates, cas.dimension, cas.measures)
}

// String returns the string representation of the coordinate array.
func (cas *GeomImpl_CoordinateArraySequence) String() string {
	if len(cas.coordinates) > 0 {
		var strBuilder strings.Builder
		strBuilder.WriteString("(")
		strBuilder.WriteString(cas.coordinates[0].String())
		for i := 1; i < len(cas.coordinates); i++ {
			strBuilder.WriteString(", ")
			strBuilder.WriteString(cas.coordinates[i].String())
		}
		strBuilder.WriteString(")")
		return strBuilder.String()
	}
	return "()"
}
