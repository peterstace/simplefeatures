package jts

import (
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geom_Coordinate_NullOrdinate is the value used to indicate a null or missing ordinate
// value. In particular, used for the value of ordinates for dimensions
// greater than the defined dimension of a coordinate.
var Geom_Coordinate_NullOrdinate = math.NaN()

// Standard ordinate index values.
const Geom_Coordinate_X = 0
const Geom_Coordinate_Y = 1
const Geom_Coordinate_Z = 2
const Geom_Coordinate_M = 3

// Geom_Coordinate is a lightweight class used to store coordinates on the
// 2-dimensional Cartesian plane.
//
// It is distinct from Point, which is a subclass of Geometry. Unlike objects
// of type Point (which contain additional information such as an envelope, a
// precision model, and spatial reference system information), a Geom_Coordinate
// only contains ordinate values and accessor methods.
//
// Coordinates are two-dimensional points, with an additional Z-ordinate. If a
// Z-ordinate value is not specified or not defined, constructed coordinates
// have a Z-ordinate of NaN (which is also the value of NullOrdinate). The
// standard comparison functions ignore the Z-ordinate. Apart from the basic
// accessor functions, JTS supports only specific operations involving the
// Z-ordinate.
//
// Implementations may optionally support Z-ordinate and M-measure values as
// appropriate for a CoordinateSequence. Use of GetZ() and GetM() accessors, or
// GetOrdinate(int) are recommended.
type Geom_Coordinate struct {
	child java.Polymorphic
	X      float64
	Y      float64
	Z      float64
}

// Geom_NewCoordinate constructs a Geom_Coordinate at (0,0,NaN).
func Geom_NewCoordinate() *Geom_Coordinate {
	return Geom_NewCoordinateWithXY(0.0, 0.0)
}

// Geom_NewCoordinateWithXY constructs a Geom_Coordinate at (x,y,NaN).
func Geom_NewCoordinateWithXY(x, y float64) *Geom_Coordinate {
	return Geom_NewCoordinateWithXYZ(x, y, Geom_Coordinate_NullOrdinate)
}

// Geom_NewCoordinateWithXYZ constructs a Geom_Coordinate at (x,y,z).
func Geom_NewCoordinateWithXYZ(x, y, z float64) *Geom_Coordinate {
	return &Geom_Coordinate{
		X: x,
		Y: y,
		Z: z,
	}
}

// Geom_NewCoordinateFromCoordinate constructs a Geom_Coordinate having the same (x,y,z)
// values as other.
func Geom_NewCoordinateFromCoordinate(other *Geom_Coordinate) *Geom_Coordinate {
	return Geom_NewCoordinateWithXYZ(other.X, other.Y, other.GetZ())
}

// GetChild returns the immediate child in the type hierarchy chain.
func (c *Geom_Coordinate) GetChild() java.Polymorphic {
	return c.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (c *Geom_Coordinate) GetParent() java.Polymorphic {
	return nil
}

// SetCoordinate sets this Geom_Coordinate's (x,y,z) values to that of other.
func (c *Geom_Coordinate) SetCoordinate(other *Geom_Coordinate) {
	if impl, ok := java.GetLeaf(c).(interface{ SetCoordinate_BODY(*Geom_Coordinate) }); ok {
		impl.SetCoordinate_BODY(other)
		return
	}
	c.SetCoordinate_BODY(other)
}

func (c *Geom_Coordinate) SetCoordinate_BODY(other *Geom_Coordinate) {
	c.X = other.X
	c.Y = other.Y
	c.Z = other.GetZ()
}

// GetX retrieves the value of the X ordinate.
func (c *Geom_Coordinate) GetX() float64 {
	return c.X
}

// SetX sets the X ordinate value.
func (c *Geom_Coordinate) SetX(x float64) {
	c.X = x
}

// GetY retrieves the value of the Y ordinate.
func (c *Geom_Coordinate) GetY() float64 {
	return c.Y
}

// SetY sets the Y ordinate value.
func (c *Geom_Coordinate) SetY(y float64) {
	c.Y = y
}

// GetZ retrieves the value of the Z ordinate, if present. If no Z value is
// present returns NaN.
func (c *Geom_Coordinate) GetZ() float64 {
	if impl, ok := java.GetLeaf(c).(interface{ GetZ_BODY() float64 }); ok {
		return impl.GetZ_BODY()
	}
	return c.GetZ_BODY()
}

func (c *Geom_Coordinate) GetZ_BODY() float64 {
	return c.Z
}

// SetZ sets the Z ordinate value.
func (c *Geom_Coordinate) SetZ(z float64) {
	if impl, ok := java.GetLeaf(c).(interface{ SetZ_BODY(float64) }); ok {
		impl.SetZ_BODY(z)
		return
	}
	c.SetZ_BODY(z)
}

func (c *Geom_Coordinate) SetZ_BODY(z float64) {
	c.Z = z
}

// GetM retrieves the value of the measure, if present. If no measure value is
// present returns NaN.
func (c *Geom_Coordinate) GetM() float64 {
	if impl, ok := java.GetLeaf(c).(interface{ GetM_BODY() float64 }); ok {
		return impl.GetM_BODY()
	}
	return c.GetM_BODY()
}

func (c *Geom_Coordinate) GetM_BODY() float64 {
	return math.NaN()
}

// SetM sets the measure value, if supported.
func (c *Geom_Coordinate) SetM(m float64) {
	if impl, ok := java.GetLeaf(c).(interface{ SetM_BODY(float64) }); ok {
		impl.SetM_BODY(m)
		return
	}
	c.SetM_BODY(m)
}

func (c *Geom_Coordinate) SetM_BODY(m float64) {
	panic(fmt.Sprintf("Invalid ordinate index: %d", Geom_Coordinate_M))
}

// GetOrdinate gets the ordinate value for the given index.
//
// The base implementation supports values for the index are Geom_Coordinate_X,
// Geom_Coordinate_Y, and Geom_Coordinate_Z.
func (c *Geom_Coordinate) GetOrdinate(ordinateIndex int) float64 {
	if impl, ok := java.GetLeaf(c).(interface{ GetOrdinate_BODY(int) float64 }); ok {
		return impl.GetOrdinate_BODY(ordinateIndex)
	}
	return c.GetOrdinate_BODY(ordinateIndex)
}

func (c *Geom_Coordinate) GetOrdinate_BODY(ordinateIndex int) float64 {
	switch ordinateIndex {
	case Geom_Coordinate_X:
		return c.X
	case Geom_Coordinate_Y:
		return c.Y
	case Geom_Coordinate_Z:
		return c.GetZ()
	}
	panic(fmt.Sprintf("Invalid ordinate index: %d", ordinateIndex))
}

// SetOrdinate sets the ordinate for the given index to a given value.
//
// The base implementation supported values for the index are Geom_Coordinate_X,
// Geom_Coordinate_Y, and Geom_Coordinate_Z.
func (c *Geom_Coordinate) SetOrdinate(ordinateIndex int, value float64) {
	if impl, ok := java.GetLeaf(c).(interface{ SetOrdinate_BODY(int, float64) }); ok {
		impl.SetOrdinate_BODY(ordinateIndex, value)
		return
	}
	c.SetOrdinate_BODY(ordinateIndex, value)
}

func (c *Geom_Coordinate) SetOrdinate_BODY(ordinateIndex int, value float64) {
	switch ordinateIndex {
	case Geom_Coordinate_X:
		c.X = value
	case Geom_Coordinate_Y:
		c.Y = value
	case Geom_Coordinate_Z:
		c.SetZ(value)
	default:
		panic(fmt.Sprintf("Invalid ordinate index: %d", ordinateIndex))
	}
}

// IsValid tests if the coordinate has valid X and Y ordinate values. An
// ordinate value is valid iff it is finite.
func (c *Geom_Coordinate) IsValid() bool {
	if !geom_isFinite(c.X) {
		return false
	}
	if !geom_isFinite(c.Y) {
		return false
	}
	return true
}

// Equals2D returns whether the planar projections of the two Coordinates are
// equal.
func (c *Geom_Coordinate) Equals2D(other *Geom_Coordinate) bool {
	if c.X != other.X {
		return false
	}
	if c.Y != other.Y {
		return false
	}
	return true
}

// Equals2DWithTolerance tests if another Geom_Coordinate has the same values for
// the X and Y ordinates, within a specified tolerance value. The Z ordinate is
// ignored.
func (c *Geom_Coordinate) Equals2DWithTolerance(other *Geom_Coordinate, tolerance float64) bool {
	if !geom_equalsWithTolerance(c.X, other.X, tolerance) {
		return false
	}
	if !geom_equalsWithTolerance(c.Y, other.Y, tolerance) {
		return false
	}
	return true
}

// Equals3D tests if another coordinate has the same values for the X, Y and Z
// ordinates.
func (c *Geom_Coordinate) Equals3D(other *Geom_Coordinate) bool {
	return (c.X == other.X) && (c.Y == other.Y) &&
		((c.GetZ() == other.GetZ()) ||
			(math.IsNaN(c.GetZ()) && math.IsNaN(other.GetZ())))
}

// EqualInZ tests if another coordinate has the same value for Z, within a
// tolerance.
func (c *Geom_Coordinate) EqualInZ(other *Geom_Coordinate, tolerance float64) bool {
	return geom_equalsWithTolerance(c.GetZ(), other.GetZ(), tolerance)
}

// Equals returns true if other has the same values for the x and y ordinates.
// Since Coordinates are 2.5D, this routine ignores the z value when making the
// comparison.
func (c *Geom_Coordinate) Equals(other *Geom_Coordinate) bool {
	return c.Equals2D(other)
}

// CompareTo compares this Geom_Coordinate with the specified Geom_Coordinate for order.
// This method ignores the z value when making the comparison. Returns:
//   - -1: this.x < other.x || ((this.x == other.x) && (this.y < other.y))
//   - 0: this.x == other.x && this.y == other.y
//   - 1: this.x > other.x || ((this.x == other.x) && (this.y > other.y))
//
// Note: This method assumes that ordinate values are valid numbers. NaN values
// are not handled correctly.
func (c *Geom_Coordinate) CompareTo(other *Geom_Coordinate) int {
	if c.X < other.X {
		return -1
	}
	if c.X > other.X {
		return 1
	}
	if c.Y < other.Y {
		return -1
	}
	if c.Y > other.Y {
		return 1
	}
	return 0
}

// String returns a string of the form (x, y, z).
func (c *Geom_Coordinate) String() string {
	if impl, ok := java.GetLeaf(c).(interface{ String_BODY() string }); ok {
		return impl.String_BODY()
	}
	return c.String_BODY()
}

func (c *Geom_Coordinate) String_BODY() string {
	return fmt.Sprintf("(%v, %v, %v)", c.X, c.Y, c.GetZ())
}

// Copy creates a copy of this Geom_Coordinate.
func (c *Geom_Coordinate) Copy() *Geom_Coordinate {
	if impl, ok := java.GetLeaf(c).(interface{ Copy_BODY() *Geom_Coordinate }); ok {
		return impl.Copy_BODY()
	}
	return c.Copy_BODY()
}

func (c *Geom_Coordinate) Copy_BODY() *Geom_Coordinate {
	return Geom_NewCoordinateFromCoordinate(c)
}

// Create creates a new Geom_Coordinate of the same type as this Geom_Coordinate, but
// with no values.
func (c *Geom_Coordinate) Create() *Geom_Coordinate {
	if impl, ok := java.GetLeaf(c).(interface{ Create_BODY() *Geom_Coordinate }); ok {
		return impl.Create_BODY()
	}
	return c.Create_BODY()
}

func (c *Geom_Coordinate) Create_BODY() *Geom_Coordinate {
	return Geom_NewCoordinate()
}

// Distance computes the 2-dimensional Euclidean distance to another location.
// The Z-ordinate is ignored.
func (c *Geom_Coordinate) Distance(other *Geom_Coordinate) float64 {
	dx := c.X - other.X
	dy := c.Y - other.Y
	return math.Hypot(dx, dy)
}

// Distance3D computes the 3-dimensional Euclidean distance to another
// location.
func (c *Geom_Coordinate) Distance3D(other *Geom_Coordinate) float64 {
	dx := c.X - other.X
	dy := c.Y - other.Y
	dz := c.GetZ() - other.GetZ()
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// HashCode gets a hashcode for this coordinate.
func (c *Geom_Coordinate) HashCode() int {
	result := 17
	result = 37*result + Geom_Coordinate_HashCodeFloat64(c.X)
	result = 37*result + Geom_Coordinate_HashCodeFloat64(c.Y)
	return result
}

// Geom_Coordinate_HashCodeFloat64 computes a hash code for a double value, using the algorithm
// from Joshua Bloch's book "Effective Java".
func Geom_Coordinate_HashCodeFloat64(x float64) int {
	f := math.Float64bits(x)
	return int(f ^ (f >> 32))
}

// Geom_DimensionalComparator compares two Coordinates, allowing for either a
// 2-dimensional or 3-dimensional comparison, and handling NaN values
// correctly.
type Geom_DimensionalComparator struct {
	dimensionsToTest int
}

// Geom_NewDimensionalComparator creates a comparator for 2 dimensional coordinates.
func Geom_NewDimensionalComparator() *Geom_DimensionalComparator {
	return Geom_NewDimensionalComparatorWithDimensions(2)
}

// Geom_NewDimensionalComparatorWithDimensions creates a comparator for 2 or 3
// dimensional coordinates, depending on the value provided.
func Geom_NewDimensionalComparatorWithDimensions(dimensionsToTest int) *Geom_DimensionalComparator {
	if dimensionsToTest != 2 && dimensionsToTest != 3 {
		panic("only 2 or 3 dimensions may be specified")
	}
	return &Geom_DimensionalComparator{dimensionsToTest: dimensionsToTest}
}

// Compare compares two Coordinates along to the number of dimensions
// specified. Returns -1, 0, or 1 depending on whether c1 is less than, equal
// to, or greater than c2.
func (dc *Geom_DimensionalComparator) Compare(c1, c2 *Geom_Coordinate) int {
	compX := dc.compare(c1.X, c2.X)
	if compX != 0 {
		return compX
	}
	compY := dc.compare(c1.Y, c2.Y)
	if compY != 0 {
		return compY
	}
	if dc.dimensionsToTest <= 2 {
		return 0
	}
	compZ := dc.compare(c1.GetZ(), c2.GetZ())
	return compZ
}

// compare compares two float64s, allowing for NaN values. NaN is treated as
// being less than any valid number. Returns -1, 0, or 1 depending on whether a
// is less than, equal to or greater than b.
func (dc *Geom_DimensionalComparator) compare(a, b float64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	if math.IsNaN(a) {
		if math.IsNaN(b) {
			return 0
		}
		return -1
	}
	if math.IsNaN(b) {
		return 1
	}
	return 0
}

// geom_isFinite returns true if the value is finite (not NaN and not infinite).
func geom_isFinite(x float64) bool {
	return !math.IsNaN(x) && !math.IsInf(x, 0)
}

// geom_equalsWithTolerance tests if two values are equal within a tolerance.
func geom_equalsWithTolerance(x1, x2, tolerance float64) bool {
	return math.Abs(x1-x2) <= tolerance
}
