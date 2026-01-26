package jts

// Geom_CoordinateSequence_X is the standard ordinate index for X (0).
const Geom_CoordinateSequence_X = 0

// Geom_CoordinateSequence_Y is the standard ordinate index for Y (1).
const Geom_CoordinateSequence_Y = 1

// Geom_CoordinateSequence_Z is the standard ordinate index for Z (2).
// This constant assumes XYZM coordinate sequence definition. Check this
// assumption using GetDimension() and GetMeasures() before use.
const Geom_CoordinateSequence_Z = 2

// Geom_CoordinateSequence_M is the standard ordinate index for M (3).
// This constant assumes XYZM coordinate sequence definition. Check this
// assumption using GetDimension() and GetMeasures() before use.
const Geom_CoordinateSequence_M = 3

// Geom_CoordinateSequence is the internal representation of a list of coordinates
// inside a Geometry.
//
// This allows Geometries to store their points using something other than the
// JTS Geom_Coordinate class. For example, a storage-efficient implementation might
// store coordinate sequences as an array of x's and an array of y's. Or a
// custom coordinate class might support extra attributes like M-values.
//
// Implementing a custom coordinate storage structure requires implementing the
// Geom_CoordinateSequence and Geom_CoordinateSequenceFactory interfaces. To use the
// custom Geom_CoordinateSequence, create a new GeometryFactory parameterized by the
// Geom_CoordinateSequenceFactory. The GeometryFactory can then be used to create
// new Geometrys. The new Geometries will use the custom Geom_CoordinateSequence
// implementation.
type Geom_CoordinateSequence interface {
	IsGeom_CoordinateSequence()

	// GetDimension returns the dimension (number of ordinates in each coordinate)
	// for this sequence.
	//
	// This total includes any measures, indicated by non-zero GetMeasures().
	GetDimension() int

	// GetMeasures returns the number of measures included in GetDimension() for
	// each coordinate for this sequence.
	//
	// For a measured coordinate sequence a non-zero value is returned.
	//   - For XY sequence measures is zero.
	//   - For XYM sequence measure is one.
	//   - For XYZ sequence measure is zero.
	//   - For XYZM sequence measure is one.
	//   - Values greater than one are supported.
	GetMeasures() int

	// HasZ checks GetDimension() and GetMeasures() to determine if GetZ() is
	// supported.
	HasZ() bool

	// HasM tests whether the coordinates in the sequence have measures associated
	// with them. Returns true if GetMeasures() > 0. See GetMeasures() to determine
	// the number of measures present.
	HasM() bool

	// CreateCoordinate creates a coordinate for use in this sequence.
	//
	// The coordinate is created supporting the same number of GetDimension() and
	// GetMeasures() as this sequence and is suitable for use with
	// GetCoordinateInto().
	CreateCoordinate() *Geom_Coordinate

	// GetCoordinate returns (possibly a copy of) the i'th coordinate in this
	// sequence. Whether or not the Geom_Coordinate returned is the actual underlying
	// Geom_Coordinate or merely a copy depends on the implementation.
	//
	// Note: In the future the semantics of this method may change to guarantee
	// that the Geom_Coordinate returned is always a copy. Callers should not assume
	// that they can modify a Geom_CoordinateSequence by modifying the object returned
	// by this method.
	GetCoordinate(i int) *Geom_Coordinate

	// GetCoordinateCopy returns a copy of the i'th coordinate in this sequence.
	// This method optimizes the situation where the caller is going to make a copy
	// anyway - if the implementation has already created a new Geom_Coordinate object,
	// no further copy is needed.
	GetCoordinateCopy(i int) *Geom_Coordinate

	// GetCoordinateInto copies the i'th coordinate in the sequence to the supplied
	// Geom_Coordinate. Only the first two dimensions are copied.
	GetCoordinateInto(index int, coord *Geom_Coordinate)

	// GetX returns ordinate X (0) of the specified coordinate.
	GetX(index int) float64

	// GetY returns ordinate Y (1) of the specified coordinate.
	GetY(index int) float64

	// GetZ returns ordinate Z of the specified coordinate if available.
	// Returns NaN if not defined.
	GetZ(index int) float64

	// GetM returns ordinate M of the specified coordinate if available.
	// Returns NaN if not defined.
	GetM(index int) float64

	// GetOrdinate returns the ordinate of a coordinate in this sequence. Ordinate
	// indices 0 and 1 are assumed to be X and Y.
	//
	// Ordinates indices greater than 1 have user-defined semantics (for instance,
	// they may contain other dimensions or measure values as described by
	// GetDimension() and GetMeasures()).
	GetOrdinate(index, ordinateIndex int) float64

	// Size returns the number of coordinates in this sequence.
	Size() int

	// SetOrdinate sets the value for a given ordinate of a coordinate in this
	// sequence.
	SetOrdinate(index, ordinateIndex int, value float64)

	// ToCoordinateArray returns (possibly copies of) the Coordinates in this
	// collection. Whether or not the Coordinates returned are the actual
	// underlying Coordinates or merely copies depends on the implementation.
	//
	// Note that if this implementation does not store its data as an array of
	// Coordinates, this method will incur a performance penalty because the array
	// needs to be built from scratch.
	ToCoordinateArray() []*Geom_Coordinate

	// ExpandEnvelope expands the given Geom_Envelope to include the coordinates in the
	// sequence. Allows implementing classes to optimize access to coordinate
	// values.
	ExpandEnvelope(env *Geom_Envelope) *Geom_Envelope

	// Copy returns a deep copy of this collection.
	Copy() Geom_CoordinateSequence
}
