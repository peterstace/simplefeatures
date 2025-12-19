package jts

import "strings"

// Geom_IntersectionMatrix_IsTrue tests if the dimension value matches TRUE (i.e. has value 0, 1, 2 or
// TRUE).
func Geom_IntersectionMatrix_IsTrue(actualDimensionValue int) bool {
	return actualDimensionValue >= 0 || actualDimensionValue == Geom_Dimension_True
}

// Geom_IntersectionMatrix_Matches tests if the dimension value satisfies the dimension symbol.
//
// actualDimensionValue is a number that can be stored in the IntersectionMatrix.
// Possible values are {True, False, DontCare, 0, 1, 2}.
//
// requiredDimensionSymbol is a character used in the string representation of
// an IntersectionMatrix. Possible values are {T, F, *, 0, 1, 2}.
//
// Returns true if the dimension symbol matches the dimension value.
func Geom_IntersectionMatrix_Matches(actualDimensionValue int, requiredDimensionSymbol byte) bool {
	if requiredDimensionSymbol == Geom_Dimension_SymDontCare {
		return true
	}
	if requiredDimensionSymbol == Geom_Dimension_SymTrue &&
		(actualDimensionValue >= 0 || actualDimensionValue == Geom_Dimension_True) {
		return true
	}
	if requiredDimensionSymbol == Geom_Dimension_SymFalse && actualDimensionValue == Geom_Dimension_False {
		return true
	}
	if requiredDimensionSymbol == Geom_Dimension_SymP && actualDimensionValue == Geom_Dimension_P {
		return true
	}
	if requiredDimensionSymbol == Geom_Dimension_SymL && actualDimensionValue == Geom_Dimension_L {
		return true
	}
	if requiredDimensionSymbol == Geom_Dimension_SymA && actualDimensionValue == Geom_Dimension_A {
		return true
	}
	return false
}

// Geom_IntersectionMatrix_MatchesStrings tests if each of the actual dimension symbols in a matrix
// string satisfies the corresponding required dimension symbol in a pattern
// string.
//
// actualDimensionSymbols is nine dimension symbols to validate. Possible values
// are {T, F, *, 0, 1, 2}.
//
// requiredDimensionSymbols is nine dimension symbols to validate against.
// Possible values are {T, F, *, 0, 1, 2}.
//
// Returns true if each of the required dimension symbols encompass the
// corresponding actual dimension symbol.
func Geom_IntersectionMatrix_MatchesStrings(actualDimensionSymbols, requiredDimensionSymbols string) bool {
	m := Geom_NewIntersectionMatrixWithElements(actualDimensionSymbols)
	return m.MatchesPattern(requiredDimensionSymbols)
}

// Geom_IntersectionMatrix models a Dimensionally Extended Nine-Intersection Model
// (DE-9IM) matrix. DE-9IM matrix values (such as "212FF1FF2") specify the
// topological relationship between two Geometries. This class can also
// represent matrix patterns (such as "T*T******") which are used for matching
// instances of DE-9IM matrices.
//
// DE-9IM matrices are 3x3 matrices with integer entries. The matrix indices
// {0,1,2} represent the topological locations that occur in a geometry
// (Interior, Boundary, Exterior). These are provided by the constants
// Geom_Location_Interior, Geom_Location_Boundary, and Geom_Location_Exterior.
//
// When used to specify the topological relationship between two geometries,
// the matrix entries represent the possible dimensions of each intersection:
// Geom_Dimension_A = 2, Geom_Dimension_L = 1, Geom_Dimension_P = 0 and Geom_Dimension_False = -1.
// When used to represent a matrix pattern entries can have the additional
// values Geom_Dimension_True ("T") and Geom_Dimension_DontCare ("*").
type Geom_IntersectionMatrix struct {
	matrix [3][3]int
}

// Geom_NewIntersectionMatrix creates an IntersectionMatrix with FALSE dimension
// values.
func Geom_NewIntersectionMatrix() *Geom_IntersectionMatrix {
	im := &Geom_IntersectionMatrix{}
	im.SetAll(Geom_Dimension_False)
	return im
}

// Geom_NewIntersectionMatrixWithElements creates an IntersectionMatrix with the
// given dimension symbols.
//
// elements is a String of nine dimension symbols in row major order.
func Geom_NewIntersectionMatrixWithElements(elements string) *Geom_IntersectionMatrix {
	im := Geom_NewIntersectionMatrix()
	im.SetFromString(elements)
	return im
}

// Geom_NewIntersectionMatrixFromMatrix creates an IntersectionMatrix with the same
// elements as other.
func Geom_NewIntersectionMatrixFromMatrix(other *Geom_IntersectionMatrix) *Geom_IntersectionMatrix {
	im := Geom_NewIntersectionMatrix()
	im.matrix[Geom_Location_Interior][Geom_Location_Interior] = other.matrix[Geom_Location_Interior][Geom_Location_Interior]
	im.matrix[Geom_Location_Interior][Geom_Location_Boundary] = other.matrix[Geom_Location_Interior][Geom_Location_Boundary]
	im.matrix[Geom_Location_Interior][Geom_Location_Exterior] = other.matrix[Geom_Location_Interior][Geom_Location_Exterior]
	im.matrix[Geom_Location_Boundary][Geom_Location_Interior] = other.matrix[Geom_Location_Boundary][Geom_Location_Interior]
	im.matrix[Geom_Location_Boundary][Geom_Location_Boundary] = other.matrix[Geom_Location_Boundary][Geom_Location_Boundary]
	im.matrix[Geom_Location_Boundary][Geom_Location_Exterior] = other.matrix[Geom_Location_Boundary][Geom_Location_Exterior]
	im.matrix[Geom_Location_Exterior][Geom_Location_Interior] = other.matrix[Geom_Location_Exterior][Geom_Location_Interior]
	im.matrix[Geom_Location_Exterior][Geom_Location_Boundary] = other.matrix[Geom_Location_Exterior][Geom_Location_Boundary]
	im.matrix[Geom_Location_Exterior][Geom_Location_Exterior] = other.matrix[Geom_Location_Exterior][Geom_Location_Exterior]
	return im
}

// Add adds one matrix to another. Addition is defined by taking the maximum
// dimension value of each position in the summand matrices.
func (im *Geom_IntersectionMatrix) Add(other *Geom_IntersectionMatrix) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			im.SetAtLeast(i, j, other.Get(i, j))
		}
	}
}

// Set changes the value of one of this IntersectionMatrix's elements.
//
// row is the row of this IntersectionMatrix, indicating the interior, boundary
// or exterior of the first Geometry.
//
// column is the column of this IntersectionMatrix, indicating the interior,
// boundary or exterior of the second Geometry.
//
// dimensionValue is the new value of the element.
func (im *Geom_IntersectionMatrix) Set(row, column, dimensionValue int) {
	im.matrix[row][column] = dimensionValue
}

// SetFromString changes the elements of this IntersectionMatrix to the
// dimension symbols in dimensionSymbols.
//
// dimensionSymbols is nine dimension symbols to which to set this
// IntersectionMatrix's elements. Possible values are {T, F, *, 0, 1, 2}.
func (im *Geom_IntersectionMatrix) SetFromString(dimensionSymbols string) {
	for i := 0; i < len(dimensionSymbols); i++ {
		row := i / 3
		col := i % 3
		im.matrix[row][col] = Geom_Dimension_ToDimensionValue(dimensionSymbols[i])
	}
}

// SetAtLeast changes the specified element to minimumDimensionValue if the
// element is less.
//
// row is the row of this IntersectionMatrix, indicating the interior, boundary
// or exterior of the first Geometry.
//
// column is the column of this IntersectionMatrix, indicating the interior,
// boundary or exterior of the second Geometry.
//
// minimumDimensionValue is the dimension value with which to compare the
// element. The order of dimension values from least to greatest is {DontCare,
// True, False, 0, 1, 2}.
func (im *Geom_IntersectionMatrix) SetAtLeast(row, column, minimumDimensionValue int) {
	if im.matrix[row][column] < minimumDimensionValue {
		im.matrix[row][column] = minimumDimensionValue
	}
}

// SetAtLeastIfValid changes the specified element to minimumDimensionValue if
// row >= 0 and column >= 0 and the element is less. Does nothing if row < 0 or
// column < 0.
func (im *Geom_IntersectionMatrix) SetAtLeastIfValid(row, column, minimumDimensionValue int) {
	if row >= 0 && column >= 0 {
		im.SetAtLeast(row, column, minimumDimensionValue)
	}
}

// SetAtLeastFromString sets each element in this IntersectionMatrix to the
// corresponding minimum dimension symbol if the element is less.
//
// minimumDimensionSymbols is nine dimension symbols with which to compare the
// elements of this IntersectionMatrix. The order of dimension values from
// least to greatest is {DontCare, True, False, 0, 1, 2}.
func (im *Geom_IntersectionMatrix) SetAtLeastFromString(minimumDimensionSymbols string) {
	for i := 0; i < len(minimumDimensionSymbols); i++ {
		row := i / 3
		col := i % 3
		im.SetAtLeast(row, col, Geom_Dimension_ToDimensionValue(minimumDimensionSymbols[i]))
	}
}

// SetAll changes the elements of this IntersectionMatrix to dimensionValue.
//
// dimensionValue is the dimension value to which to set this
// IntersectionMatrix's elements. Possible values {True, False, DontCare, 0, 1, 2}.
func (im *Geom_IntersectionMatrix) SetAll(dimensionValue int) {
	for ai := 0; ai < 3; ai++ {
		for bi := 0; bi < 3; bi++ {
			im.matrix[ai][bi] = dimensionValue
		}
	}
}

// Get returns the value of one of this matrix entries. The value of the
// provided index is one of the values from the Location constants. The value
// returned is a constant from the Dimension constants.
func (im *Geom_IntersectionMatrix) Get(row, column int) int {
	return im.matrix[row][column]
}

// IsDisjoint tests if this matrix matches [FF*FF****].
func (im *Geom_IntersectionMatrix) IsDisjoint() bool {
	return im.matrix[Geom_Location_Interior][Geom_Location_Interior] == Geom_Dimension_False &&
		im.matrix[Geom_Location_Interior][Geom_Location_Boundary] == Geom_Dimension_False &&
		im.matrix[Geom_Location_Boundary][Geom_Location_Interior] == Geom_Dimension_False &&
		im.matrix[Geom_Location_Boundary][Geom_Location_Boundary] == Geom_Dimension_False
}

// IsIntersects tests if IsDisjoint returns false.
func (im *Geom_IntersectionMatrix) IsIntersects() bool {
	return !im.IsDisjoint()
}

// IsTouches tests if this matrix matches [FT*******], [F**T*****] or
// [F***T****].
//
// dimensionOfGeometryA is the dimension of the first Geometry.
// dimensionOfGeometryB is the dimension of the second Geometry.
//
// Returns true if the two Geometries related by this matrix touch; Returns
// false if both Geometries are points.
func (im *Geom_IntersectionMatrix) IsTouches(dimensionOfGeometryA, dimensionOfGeometryB int) bool {
	if dimensionOfGeometryA > dimensionOfGeometryB {
		// No need to get transpose because pattern matrix is symmetrical.
		return im.IsTouches(dimensionOfGeometryB, dimensionOfGeometryA)
	}
	if (dimensionOfGeometryA == Geom_Dimension_A && dimensionOfGeometryB == Geom_Dimension_A) ||
		(dimensionOfGeometryA == Geom_Dimension_L && dimensionOfGeometryB == Geom_Dimension_L) ||
		(dimensionOfGeometryA == Geom_Dimension_L && dimensionOfGeometryB == Geom_Dimension_A) ||
		(dimensionOfGeometryA == Geom_Dimension_P && dimensionOfGeometryB == Geom_Dimension_A) ||
		(dimensionOfGeometryA == Geom_Dimension_P && dimensionOfGeometryB == Geom_Dimension_L) {
		return im.matrix[Geom_Location_Interior][Geom_Location_Interior] == Geom_Dimension_False &&
			(Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Boundary]) ||
				Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Boundary][Geom_Location_Interior]) ||
				Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Boundary][Geom_Location_Boundary]))
	}
	return false
}

// IsCrosses tests whether this geometry crosses the specified geometry.
//
// The crosses predicate has the following equivalent definitions:
//   - The geometries have some but not all interior points in common.
//   - The DE-9IM Intersection Matrix for the two geometries matches:
//   - [T*T******] (for P/L, P/A, and L/A situations)
//   - [T*****T**] (for L/P, L/A, and A/L situations)
//   - [0********] (for L/L situations)
//
// For any other combination of dimensions this predicate returns false.
//
// The SFS defined this predicate only for P/L, P/A, L/L, and L/A situations.
// JTS extends the definition to apply to L/P, A/P and A/L situations as well.
// This makes the relation symmetric.
func (im *Geom_IntersectionMatrix) IsCrosses(dimensionOfGeometryA, dimensionOfGeometryB int) bool {
	if (dimensionOfGeometryA == Geom_Dimension_P && dimensionOfGeometryB == Geom_Dimension_L) ||
		(dimensionOfGeometryA == Geom_Dimension_P && dimensionOfGeometryB == Geom_Dimension_A) ||
		(dimensionOfGeometryA == Geom_Dimension_L && dimensionOfGeometryB == Geom_Dimension_A) {
		return Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Interior]) &&
			Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Exterior])
	}
	if (dimensionOfGeometryA == Geom_Dimension_L && dimensionOfGeometryB == Geom_Dimension_P) ||
		(dimensionOfGeometryA == Geom_Dimension_A && dimensionOfGeometryB == Geom_Dimension_P) ||
		(dimensionOfGeometryA == Geom_Dimension_A && dimensionOfGeometryB == Geom_Dimension_L) {
		return Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Interior]) &&
			Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Exterior][Geom_Location_Interior])
	}
	if dimensionOfGeometryA == Geom_Dimension_L && dimensionOfGeometryB == Geom_Dimension_L {
		return im.matrix[Geom_Location_Interior][Geom_Location_Interior] == 0
	}
	return false
}

// IsWithin tests whether this matrix matches [T*F**F***].
func (im *Geom_IntersectionMatrix) IsWithin() bool {
	return Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Interior]) &&
		im.matrix[Geom_Location_Interior][Geom_Location_Exterior] == Geom_Dimension_False &&
		im.matrix[Geom_Location_Boundary][Geom_Location_Exterior] == Geom_Dimension_False
}

// IsContains tests whether this matrix matches [T*****FF*].
func (im *Geom_IntersectionMatrix) IsContains() bool {
	return Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Interior]) &&
		im.matrix[Geom_Location_Exterior][Geom_Location_Interior] == Geom_Dimension_False &&
		im.matrix[Geom_Location_Exterior][Geom_Location_Boundary] == Geom_Dimension_False
}

// IsCovers tests if this matrix matches [T*****FF*] or [*T****FF*] or
// [***T**FF*] or [****T*FF*].
func (im *Geom_IntersectionMatrix) IsCovers() bool {
	hasPointInCommon := Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Interior]) ||
		Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Boundary]) ||
		Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Boundary][Geom_Location_Interior]) ||
		Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Boundary][Geom_Location_Boundary])

	return hasPointInCommon &&
		im.matrix[Geom_Location_Exterior][Geom_Location_Interior] == Geom_Dimension_False &&
		im.matrix[Geom_Location_Exterior][Geom_Location_Boundary] == Geom_Dimension_False
}

// IsCoveredBy tests if this matrix matches [T*F**F***] or [*TF**F***] or
// [**FT*F***] or [**F*TF***].
func (im *Geom_IntersectionMatrix) IsCoveredBy() bool {
	hasPointInCommon := Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Interior]) ||
		Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Boundary]) ||
		Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Boundary][Geom_Location_Interior]) ||
		Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Boundary][Geom_Location_Boundary])

	return hasPointInCommon &&
		im.matrix[Geom_Location_Interior][Geom_Location_Exterior] == Geom_Dimension_False &&
		im.matrix[Geom_Location_Boundary][Geom_Location_Exterior] == Geom_Dimension_False
}

// IsEquals tests whether the argument dimensions are equal and this matrix
// matches the pattern [T*F**FFF*].
//
// Note: This pattern differs from the one stated in Simple feature access -
// Part 1: Common architecture. That document states the pattern as [TFFFTFFFT].
// This would specify that two identical POINTs are not equal, which is not
// desirable behaviour. The pattern used here has been corrected to compute
// equality in this situation.
func (im *Geom_IntersectionMatrix) IsEquals(dimensionOfGeometryA, dimensionOfGeometryB int) bool {
	if dimensionOfGeometryA != dimensionOfGeometryB {
		return false
	}
	return Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Interior]) &&
		im.matrix[Geom_Location_Interior][Geom_Location_Exterior] == Geom_Dimension_False &&
		im.matrix[Geom_Location_Boundary][Geom_Location_Exterior] == Geom_Dimension_False &&
		im.matrix[Geom_Location_Exterior][Geom_Location_Interior] == Geom_Dimension_False &&
		im.matrix[Geom_Location_Exterior][Geom_Location_Boundary] == Geom_Dimension_False
}

// IsOverlaps tests if this matrix matches [T*T***T**] (for two points or two
// surfaces) or [1*T***T**] (for two curves).
func (im *Geom_IntersectionMatrix) IsOverlaps(dimensionOfGeometryA, dimensionOfGeometryB int) bool {
	if (dimensionOfGeometryA == Geom_Dimension_P && dimensionOfGeometryB == Geom_Dimension_P) ||
		(dimensionOfGeometryA == Geom_Dimension_A && dimensionOfGeometryB == Geom_Dimension_A) {
		return Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Interior]) &&
			Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Exterior]) &&
			Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Exterior][Geom_Location_Interior])
	}
	if dimensionOfGeometryA == Geom_Dimension_L && dimensionOfGeometryB == Geom_Dimension_L {
		return im.matrix[Geom_Location_Interior][Geom_Location_Interior] == 1 &&
			Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Interior][Geom_Location_Exterior]) &&
			Geom_IntersectionMatrix_IsTrue(im.matrix[Geom_Location_Exterior][Geom_Location_Interior])
	}
	return false
}

// MatchesPattern tests whether this matrix matches the given matrix pattern.
//
// pattern is a pattern containing nine dimension symbols with which to compare
// the entries of this matrix. Possible symbol values are {T, F, *, 0, 1, 2}.
//
// Returns true if this matrix matches the pattern.
func (im *Geom_IntersectionMatrix) MatchesPattern(pattern string) bool {
	if len(pattern) != 9 {
		panic("Should be length 9: " + pattern)
	}
	for ai := 0; ai < 3; ai++ {
		for bi := 0; bi < 3; bi++ {
			if !Geom_IntersectionMatrix_Matches(im.matrix[ai][bi], pattern[3*ai+bi]) {
				return false
			}
		}
	}
	return true
}

// Transpose transposes this IntersectionMatrix.
//
// Returns this IntersectionMatrix as a convenience.
func (im *Geom_IntersectionMatrix) Transpose() *Geom_IntersectionMatrix {
	temp := im.matrix[1][0]
	im.matrix[1][0] = im.matrix[0][1]
	im.matrix[0][1] = temp
	temp = im.matrix[2][0]
	im.matrix[2][0] = im.matrix[0][2]
	im.matrix[0][2] = temp
	temp = im.matrix[2][1]
	im.matrix[2][1] = im.matrix[1][2]
	im.matrix[1][2] = temp
	return im
}

// String returns a nine-character String representation of this
// IntersectionMatrix.
func (im *Geom_IntersectionMatrix) String() string {
	var builder strings.Builder
	builder.Grow(9)
	for ai := 0; ai < 3; ai++ {
		for bi := 0; bi < 3; bi++ {
			builder.WriteByte(Geom_Dimension_ToDimensionSymbol(im.matrix[ai][bi]))
		}
	}
	return builder.String()
}
