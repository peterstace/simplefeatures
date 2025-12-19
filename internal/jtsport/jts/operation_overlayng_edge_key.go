package jts

import (
	"fmt"
	"math"
)

// OperationOverlayng_EdgeKey is a key for sorting and comparing edges in a
// noded arrangement. Relies on the fact that in a correctly noded arrangement
// edges are identical (up to direction) if they have their first segment in
// common.
type OperationOverlayng_EdgeKey struct {
	p0x float64
	p0y float64
	p1x float64
	p1y float64
}

// OperationOverlayng_EdgeKey_Create creates an EdgeKey for the given edge.
func OperationOverlayng_EdgeKey_Create(edge *OperationOverlayng_Edge) *OperationOverlayng_EdgeKey {
	return OperationOverlayng_NewEdgeKey(edge)
}

// OperationOverlayng_NewEdgeKey creates a new EdgeKey for the given edge.
func OperationOverlayng_NewEdgeKey(edge *OperationOverlayng_Edge) *OperationOverlayng_EdgeKey {
	ek := &OperationOverlayng_EdgeKey{}
	ek.initPoints(edge)
	return ek
}

func (ek *OperationOverlayng_EdgeKey) initPoints(edge *OperationOverlayng_Edge) {
	direction := edge.Direction()
	if direction {
		ek.init(edge.GetCoordinate(0), edge.GetCoordinate(1))
	} else {
		length := edge.Size()
		ek.init(edge.GetCoordinate(length-1), edge.GetCoordinate(length-2))
	}
}

func (ek *OperationOverlayng_EdgeKey) init(p0, p1 *Geom_Coordinate) {
	ek.p0x = p0.GetX()
	ek.p0y = p0.GetY()
	ek.p1x = p1.GetX()
	ek.p1y = p1.GetY()
}

// CompareTo compares this EdgeKey to another.
func (ek *OperationOverlayng_EdgeKey) CompareTo(other *OperationOverlayng_EdgeKey) int {
	if ek.p0x < other.p0x {
		return -1
	}
	if ek.p0x > other.p0x {
		return 1
	}
	if ek.p0y < other.p0y {
		return -1
	}
	if ek.p0y > other.p0y {
		return 1
	}
	// first points are equal, compare second
	if ek.p1x < other.p1x {
		return -1
	}
	if ek.p1x > other.p1x {
		return 1
	}
	if ek.p1y < other.p1y {
		return -1
	}
	if ek.p1y > other.p1y {
		return 1
	}
	return 0
}

// Equals tests if this EdgeKey is equal to another.
func (ek *OperationOverlayng_EdgeKey) Equals(other *OperationOverlayng_EdgeKey) bool {
	return ek.p0x == other.p0x &&
		ek.p0y == other.p0y &&
		ek.p1x == other.p1x &&
		ek.p1y == other.p1y
}

// HashCode gets a hashcode for this object.
func (ek *OperationOverlayng_EdgeKey) HashCode() int {
	// Algorithm from Effective Java by Joshua Bloch
	result := 17
	result = 37*result + operationOverlayng_EdgeKey_hashCodeFloat64(ek.p0x)
	result = 37*result + operationOverlayng_EdgeKey_hashCodeFloat64(ek.p0y)
	result = 37*result + operationOverlayng_EdgeKey_hashCodeFloat64(ek.p1x)
	result = 37*result + operationOverlayng_EdgeKey_hashCodeFloat64(ek.p1y)
	return result
}

// operationOverlayng_EdgeKey_hashCodeFloat64 computes a hash code for a
// double value, using the algorithm from Joshua Bloch's book "Effective Java".
func operationOverlayng_EdgeKey_hashCodeFloat64(x float64) int {
	f := math.Float64bits(x)
	return int(f ^ (f >> 32))
}

// String returns a string representation of the EdgeKey.
func (ek *OperationOverlayng_EdgeKey) String() string {
	return "EdgeKey(" + ek.format(ek.p0x, ek.p0y) +
		", " + ek.format(ek.p1x, ek.p1y) + ")"
}

func (ek *OperationOverlayng_EdgeKey) format(x, y float64) string {
	return operationOverlayng_EdgeKey_formatOrdinate(x) + " " + operationOverlayng_EdgeKey_formatOrdinate(y)
}

func operationOverlayng_EdgeKey_formatOrdinate(v float64) string {
	return fmt.Sprintf("%g", v)
}
