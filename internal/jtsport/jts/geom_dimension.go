package jts

import "fmt"

// Geom_Dimension_P is the dimension value of a point (0).
const Geom_Dimension_P = 0

// Geom_Dimension_L is the dimension value of a curve (1).
const Geom_Dimension_L = 1

// Geom_Dimension_A is the dimension value of a surface (2).
const Geom_Dimension_A = 2

// Geom_Dimension_False is the dimension value of the empty geometry (-1).
const Geom_Dimension_False = -1

// Geom_Dimension_True is the dimension value of non-empty geometries (= {P, L, A}).
const Geom_Dimension_True = -2

// Geom_Dimension_DontCare is the dimension value for any dimension (= {False, True}).
const Geom_Dimension_DontCare = -3

// Geom_Dimension_SymFalse is the symbol for the FALSE pattern matrix entry.
const Geom_Dimension_SymFalse = 'F'

// Geom_Dimension_SymTrue is the symbol for the TRUE pattern matrix entry.
const Geom_Dimension_SymTrue = 'T'

// Geom_Dimension_SymDontCare is the symbol for the DONTCARE pattern matrix entry.
const Geom_Dimension_SymDontCare = '*'

// Geom_Dimension_SymP is the symbol for the P (dimension 0) pattern matrix entry.
const Geom_Dimension_SymP = '0'

// Geom_Dimension_SymL is the symbol for the L (dimension 1) pattern matrix entry.
const Geom_Dimension_SymL = '1'

// Geom_Dimension_SymA is the symbol for the A (dimension 2) pattern matrix entry.
const Geom_Dimension_SymA = '2'

// Geom_Dimension_ToDimensionSymbol converts the dimension value to a dimension symbol, for
// example, True => 'T'.
//
// dimensionValue is a number that can be stored in the Geom_IntersectionMatrix.
// Possible values are {True, False, DontCare, 0, 1, 2}.
//
// Returns a character for use in the string representation of an
// Geom_IntersectionMatrix. Possible values are {T, F, *, 0, 1, 2}.
func Geom_Dimension_ToDimensionSymbol(dimensionValue int) byte {
	switch dimensionValue {
	case Geom_Dimension_False:
		return Geom_Dimension_SymFalse
	case Geom_Dimension_True:
		return Geom_Dimension_SymTrue
	case Geom_Dimension_DontCare:
		return Geom_Dimension_SymDontCare
	case Geom_Dimension_P:
		return Geom_Dimension_SymP
	case Geom_Dimension_L:
		return Geom_Dimension_SymL
	case Geom_Dimension_A:
		return Geom_Dimension_SymA
	default:
		panic(fmt.Sprintf("unknown dimension value: %d", dimensionValue))
	}
}

// Geom_Dimension_ToDimensionValue converts the dimension symbol to a dimension value, for
// example, '*' => DontCare.
//
// dimensionSymbol is a character for use in the string representation of an
// Geom_IntersectionMatrix. Possible values are {T, F, *, 0, 1, 2}.
//
// Returns a number that can be stored in the Geom_IntersectionMatrix. Possible
// values are {True, False, DontCare, 0, 1, 2}.
func Geom_Dimension_ToDimensionValue(dimensionSymbol byte) int {
	switch geom_toUpperCase(dimensionSymbol) {
	case Geom_Dimension_SymFalse:
		return Geom_Dimension_False
	case Geom_Dimension_SymTrue:
		return Geom_Dimension_True
	case Geom_Dimension_SymDontCare:
		return Geom_Dimension_DontCare
	case Geom_Dimension_SymP:
		return Geom_Dimension_P
	case Geom_Dimension_SymL:
		return Geom_Dimension_L
	case Geom_Dimension_SymA:
		return Geom_Dimension_A
	default:
		panic(fmt.Sprintf("unknown dimension symbol: %c", dimensionSymbol))
	}
}

// geom_toUpperCase converts a byte to uppercase if it's a lowercase ASCII letter.
func geom_toUpperCase(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 'a' + 'A'
	}
	return b
}
