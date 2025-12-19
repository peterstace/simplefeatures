package jts

import "math"

// GeomImpl_PackedCoordinateSequenceDouble is a CoordinateSequence implementation
// based on a packed double array. Coordinates returned by ToCoordinateArray and
// GetCoordinate are copies of the internal values. To change the actual values,
// use the provided setters.
type GeomImpl_PackedCoordinateSequenceDouble struct {
	dimension int
	measures  int
	coords    []float64
	coordRef  []*Geom_Coordinate // Cache for toCoordinateArray.
}

var _ Geom_CoordinateSequence = (*GeomImpl_PackedCoordinateSequenceDouble)(nil)

func (s *GeomImpl_PackedCoordinateSequenceDouble) IsGeom_CoordinateSequence() {}

// GeomImpl_NewPackedCoordinateSequenceDoubleFromDoubles builds a new packed
// coordinate sequence from an array of doubles.
func GeomImpl_NewPackedCoordinateSequenceDoubleFromDoubles(coords []float64, dimension, measures int) *GeomImpl_PackedCoordinateSequenceDouble {
	if dimension-measures < 2 {
		panic("Must have at least 2 spatial dimensions")
	}
	if len(coords)%dimension != 0 {
		panic("Packed array does not contain an integral number of coordinates")
	}
	return &GeomImpl_PackedCoordinateSequenceDouble{
		dimension: dimension,
		measures:  measures,
		coords:    coords,
	}
}

// GeomImpl_NewPackedCoordinateSequenceDoubleFromFloats builds a new packed
// coordinate sequence from an array of floats (converting to doubles).
func GeomImpl_NewPackedCoordinateSequenceDoubleFromFloats(coords []float32, dimension, measures int) *GeomImpl_PackedCoordinateSequenceDouble {
	if dimension-measures < 2 {
		panic("Must have at least 2 spatial dimensions")
	}
	doubles := make([]float64, len(coords))
	for i, v := range coords {
		doubles[i] = float64(v)
	}
	return &GeomImpl_PackedCoordinateSequenceDouble{
		dimension: dimension,
		measures:  measures,
		coords:    doubles,
	}
}

// GeomImpl_NewPackedCoordinateSequenceDoubleFromCoordinates builds a new packed
// coordinate sequence from a coordinate array.
func GeomImpl_NewPackedCoordinateSequenceDoubleFromCoordinates(coordinates []*Geom_Coordinate, dimension, measures int) *GeomImpl_PackedCoordinateSequenceDouble {
	if dimension-measures < 2 {
		panic("Must have at least 2 spatial dimensions")
	}
	if coordinates == nil {
		coordinates = []*Geom_Coordinate{}
	}
	coords := make([]float64, len(coordinates)*dimension)
	for i, coord := range coordinates {
		offset := i * dimension
		coords[offset] = coord.GetX()
		coords[offset+1] = coord.GetY()
		if dimension >= 3 {
			coords[offset+2] = coord.GetOrdinate(2) // Z or M.
		}
		if dimension >= 4 {
			coords[offset+3] = coord.GetOrdinate(3) // M.
		}
	}
	return &GeomImpl_PackedCoordinateSequenceDouble{
		dimension: dimension,
		measures:  measures,
		coords:    coords,
	}
}

// GeomImpl_NewPackedCoordinateSequenceDoubleFromCoordinatesInferDimension builds
// a new packed coordinate sequence from a coordinate array, inferring measures
// from dimension.
func GeomImpl_NewPackedCoordinateSequenceDoubleFromCoordinatesInferDimension(coordinates []*Geom_Coordinate, dimension int) *GeomImpl_PackedCoordinateSequenceDouble {
	measures := 0
	if dimension > 3 {
		measures = dimension - 3
	}
	return GeomImpl_NewPackedCoordinateSequenceDoubleFromCoordinates(coordinates, dimension, measures)
}

// GeomImpl_NewPackedCoordinateSequenceDoubleFromCoordinatesDefault builds a new
// packed coordinate sequence from a coordinate array with default dimension 3.
func GeomImpl_NewPackedCoordinateSequenceDoubleFromCoordinatesDefault(coordinates []*Geom_Coordinate) *GeomImpl_PackedCoordinateSequenceDouble {
	return GeomImpl_NewPackedCoordinateSequenceDoubleFromCoordinates(coordinates, 3, 0)
}

// GeomImpl_NewPackedCoordinateSequenceDoubleWithSize builds a new empty packed
// coordinate sequence of a given size.
func GeomImpl_NewPackedCoordinateSequenceDoubleWithSize(size, dimension, measures int) *GeomImpl_PackedCoordinateSequenceDouble {
	if dimension-measures < 2 {
		panic("Must have at least 2 spatial dimensions")
	}
	return &GeomImpl_PackedCoordinateSequenceDouble{
		dimension: dimension,
		measures:  measures,
		coords:    make([]float64, size*dimension),
	}
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) GetDimension() int {
	return s.dimension
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) GetMeasures() int {
	return s.measures
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) HasZ() bool {
	return s.dimension-s.measures > 2
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) HasM() bool {
	return s.measures > 0
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) CreateCoordinate() *Geom_Coordinate {
	return Geom_Coordinates_CreateWithMeasures(s.dimension, s.measures)
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) Size() int {
	return len(s.coords) / s.dimension
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) GetCoordinate(i int) *Geom_Coordinate {
	if s.coordRef != nil {
		return s.coordRef[i]
	}
	return s.getCoordinateInternal(i)
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) GetCoordinateCopy(i int) *Geom_Coordinate {
	return s.getCoordinateInternal(i)
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) GetCoordinateInto(index int, coord *Geom_Coordinate) {
	coord.SetX(s.GetOrdinate(index, 0))
	coord.SetY(s.GetOrdinate(index, 1))
	if s.HasZ() {
		coord.SetZ(s.GetZ(index))
	}
	if s.HasM() {
		coord.SetM(s.GetM(index))
	}
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) getCoordinateInternal(i int) *Geom_Coordinate {
	x := s.coords[i*s.dimension]
	y := s.coords[i*s.dimension+1]
	if s.dimension == 2 && s.measures == 0 {
		return Geom_NewCoordinateXY2DWithXY(x, y).Geom_Coordinate
	} else if s.dimension == 3 && s.measures == 0 {
		z := s.coords[i*s.dimension+2]
		return Geom_NewCoordinateWithXYZ(x, y, z)
	} else if s.dimension == 3 && s.measures == 1 {
		m := s.coords[i*s.dimension+2]
		return Geom_NewCoordinateXYM3DWithXYM(x, y, m).Geom_Coordinate
	} else if s.dimension == 4 {
		z := s.coords[i*s.dimension+2]
		m := s.coords[i*s.dimension+3]
		return Geom_NewCoordinateXYZM4DWithXYZM(x, y, z, m).Geom_Coordinate
	}
	return Geom_NewCoordinateWithXY(x, y)
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) GetX(index int) float64 {
	return s.GetOrdinate(index, 0)
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) GetY(index int) float64 {
	return s.GetOrdinate(index, 1)
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) GetZ(index int) float64 {
	if s.HasZ() {
		return s.GetOrdinate(index, 2)
	}
	return math.NaN()
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) GetM(index int) float64 {
	if s.HasM() {
		mIndex := s.dimension - 1
		return s.GetOrdinate(index, mIndex)
	}
	return math.NaN()
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) GetOrdinate(index, ordinateIndex int) float64 {
	return s.coords[index*s.dimension+ordinateIndex]
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) SetOrdinate(index, ordinate int, value float64) {
	s.coordRef = nil
	s.coords[index*s.dimension+ordinate] = value
}

// GetRawCoordinates returns the underlying array containing the coordinate values.
func (s *GeomImpl_PackedCoordinateSequenceDouble) GetRawCoordinates() []float64 {
	return s.coords
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) ToCoordinateArray() []*Geom_Coordinate {
	if s.coordRef != nil {
		return s.coordRef
	}
	coords := make([]*Geom_Coordinate, s.Size())
	for i := range coords {
		coords[i] = s.getCoordinateInternal(i)
	}
	s.coordRef = coords
	return coords
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) ExpandEnvelope(env *Geom_Envelope) *Geom_Envelope {
	for i := 0; i < len(s.coords); i += s.dimension {
		if i+1 < len(s.coords) {
			env.ExpandToIncludeXY(s.coords[i], s.coords[i+1])
		}
	}
	return env
}

func (s *GeomImpl_PackedCoordinateSequenceDouble) Copy() Geom_CoordinateSequence {
	clone := make([]float64, len(s.coords))
	copy(clone, s.coords)
	return GeomImpl_NewPackedCoordinateSequenceDoubleFromDoubles(clone, s.dimension, s.measures)
}

// GeomImpl_PackedCoordinateSequenceFloat is a CoordinateSequence implementation
// based on a packed float array. Coordinates returned by ToCoordinateArray and
// GetCoordinate are copies of the internal values. To change the actual values,
// use the provided setters.
type GeomImpl_PackedCoordinateSequenceFloat struct {
	dimension int
	measures  int
	coords    []float32
	coordRef  []*Geom_Coordinate // Cache for toCoordinateArray.
}

var _ Geom_CoordinateSequence = (*GeomImpl_PackedCoordinateSequenceFloat)(nil)

func (s *GeomImpl_PackedCoordinateSequenceFloat) IsGeom_CoordinateSequence() {}

// GeomImpl_NewPackedCoordinateSequenceFloatFromFloats constructs a packed
// coordinate sequence from an array of floats.
func GeomImpl_NewPackedCoordinateSequenceFloatFromFloats(coords []float32, dimension, measures int) *GeomImpl_PackedCoordinateSequenceFloat {
	if dimension-measures < 2 {
		panic("Must have at least 2 spatial dimensions")
	}
	if len(coords)%dimension != 0 {
		panic("Packed array does not contain an integral number of coordinates")
	}
	return &GeomImpl_PackedCoordinateSequenceFloat{
		dimension: dimension,
		measures:  measures,
		coords:    coords,
	}
}

// GeomImpl_NewPackedCoordinateSequenceFloatFromDoubles constructs a packed
// coordinate sequence from an array of doubles (converting to floats).
func GeomImpl_NewPackedCoordinateSequenceFloatFromDoubles(coords []float64, dimension, measures int) *GeomImpl_PackedCoordinateSequenceFloat {
	if dimension-measures < 2 {
		panic("Must have at least 2 spatial dimensions")
	}
	floats := make([]float32, len(coords))
	for i, v := range coords {
		floats[i] = float32(v)
	}
	return &GeomImpl_PackedCoordinateSequenceFloat{
		dimension: dimension,
		measures:  measures,
		coords:    floats,
	}
}

// GeomImpl_NewPackedCoordinateSequenceFloatFromCoordinates constructs a packed
// coordinate sequence from a coordinate array.
func GeomImpl_NewPackedCoordinateSequenceFloatFromCoordinates(coordinates []*Geom_Coordinate, dimension, measures int) *GeomImpl_PackedCoordinateSequenceFloat {
	if dimension-measures < 2 {
		panic("Must have at least 2 spatial dimensions")
	}
	if coordinates == nil {
		coordinates = []*Geom_Coordinate{}
	}
	coords := make([]float32, len(coordinates)*dimension)
	for i, coord := range coordinates {
		offset := i * dimension
		coords[offset] = float32(coord.GetX())
		coords[offset+1] = float32(coord.GetY())
		if dimension >= 3 {
			coords[offset+2] = float32(coord.GetOrdinate(2)) // Z or M.
		}
		if dimension >= 4 {
			coords[offset+3] = float32(coord.GetOrdinate(3)) // M.
		}
	}
	return &GeomImpl_PackedCoordinateSequenceFloat{
		dimension: dimension,
		measures:  measures,
		coords:    coords,
	}
}

// GeomImpl_NewPackedCoordinateSequenceFloatFromCoordinatesInferDimension builds
// a new packed coordinate sequence from a coordinate array, inferring measures
// from dimension.
func GeomImpl_NewPackedCoordinateSequenceFloatFromCoordinatesInferDimension(coordinates []*Geom_Coordinate, dimension int) *GeomImpl_PackedCoordinateSequenceFloat {
	measures := 0
	if dimension > 3 {
		measures = dimension - 3
	}
	return GeomImpl_NewPackedCoordinateSequenceFloatFromCoordinates(coordinates, dimension, measures)
}

// GeomImpl_NewPackedCoordinateSequenceFloatWithSize constructs an empty packed
// coordinate sequence of a given size.
func GeomImpl_NewPackedCoordinateSequenceFloatWithSize(size, dimension, measures int) *GeomImpl_PackedCoordinateSequenceFloat {
	if dimension-measures < 2 {
		panic("Must have at least 2 spatial dimensions")
	}
	return &GeomImpl_PackedCoordinateSequenceFloat{
		dimension: dimension,
		measures:  measures,
		coords:    make([]float32, size*dimension),
	}
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) GetDimension() int {
	return s.dimension
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) GetMeasures() int {
	return s.measures
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) HasZ() bool {
	return s.dimension-s.measures > 2
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) HasM() bool {
	return s.measures > 0
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) CreateCoordinate() *Geom_Coordinate {
	return Geom_Coordinates_CreateWithMeasures(s.dimension, s.measures)
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) Size() int {
	return len(s.coords) / s.dimension
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) GetCoordinate(i int) *Geom_Coordinate {
	if s.coordRef != nil {
		return s.coordRef[i]
	}
	return s.getCoordinateInternal(i)
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) GetCoordinateCopy(i int) *Geom_Coordinate {
	return s.getCoordinateInternal(i)
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) GetCoordinateInto(index int, coord *Geom_Coordinate) {
	coord.SetX(s.GetOrdinate(index, 0))
	coord.SetY(s.GetOrdinate(index, 1))
	if s.HasZ() {
		coord.SetZ(s.GetZ(index))
	}
	if s.HasM() {
		coord.SetM(s.GetM(index))
	}
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) getCoordinateInternal(i int) *Geom_Coordinate {
	x := float64(s.coords[i*s.dimension])
	y := float64(s.coords[i*s.dimension+1])
	if s.dimension == 2 && s.measures == 0 {
		return Geom_NewCoordinateXY2DWithXY(x, y).Geom_Coordinate
	} else if s.dimension == 3 && s.measures == 0 {
		z := float64(s.coords[i*s.dimension+2])
		return Geom_NewCoordinateWithXYZ(x, y, z)
	} else if s.dimension == 3 && s.measures == 1 {
		m := float64(s.coords[i*s.dimension+2])
		return Geom_NewCoordinateXYM3DWithXYM(x, y, m).Geom_Coordinate
	} else if s.dimension == 4 {
		z := float64(s.coords[i*s.dimension+2])
		m := float64(s.coords[i*s.dimension+3])
		return Geom_NewCoordinateXYZM4DWithXYZM(x, y, z, m).Geom_Coordinate
	}
	return Geom_NewCoordinateWithXY(x, y)
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) GetX(index int) float64 {
	return s.GetOrdinate(index, 0)
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) GetY(index int) float64 {
	return s.GetOrdinate(index, 1)
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) GetZ(index int) float64 {
	if s.HasZ() {
		return s.GetOrdinate(index, 2)
	}
	return math.NaN()
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) GetM(index int) float64 {
	if s.HasM() {
		mIndex := s.dimension - 1
		return s.GetOrdinate(index, mIndex)
	}
	return math.NaN()
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) GetOrdinate(index, ordinateIndex int) float64 {
	return float64(s.coords[index*s.dimension+ordinateIndex])
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) SetOrdinate(index, ordinate int, value float64) {
	s.coordRef = nil
	s.coords[index*s.dimension+ordinate] = float32(value)
}

// GetRawCoordinates returns the underlying array containing the coordinate values.
func (s *GeomImpl_PackedCoordinateSequenceFloat) GetRawCoordinates() []float32 {
	return s.coords
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) ToCoordinateArray() []*Geom_Coordinate {
	if s.coordRef != nil {
		return s.coordRef
	}
	coords := make([]*Geom_Coordinate, s.Size())
	for i := range coords {
		coords[i] = s.getCoordinateInternal(i)
	}
	s.coordRef = coords
	return coords
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) ExpandEnvelope(env *Geom_Envelope) *Geom_Envelope {
	for i := 0; i < len(s.coords); i += s.dimension {
		if i+1 < len(s.coords) {
			env.ExpandToIncludeXY(float64(s.coords[i]), float64(s.coords[i+1]))
		}
	}
	return env
}

func (s *GeomImpl_PackedCoordinateSequenceFloat) Copy() Geom_CoordinateSequence {
	clone := make([]float32, len(s.coords))
	copy(clone, s.coords)
	return GeomImpl_NewPackedCoordinateSequenceFloatFromFloats(clone, s.dimension, s.measures)
}
