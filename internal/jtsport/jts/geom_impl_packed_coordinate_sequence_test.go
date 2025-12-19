package jts_test

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

// Tests ported from PackedCoordinateSequenceTest.java.

const pcsTestSize = 100

func pcsGetFactory() *jts.GeomImpl_PackedCoordinateSequenceFactory {
	return jts.GeomImpl_NewPackedCoordinateSequenceFactory()
}

func TestPackedCoordinateSequenceDouble(t *testing.T) {
	pcsCheckAll(t, jts.GeomImpl_PackedCoordinateSequenceFactory_DOUBLE_FACTORY)
}

func TestPackedCoordinateSequenceFloat(t *testing.T) {
	pcsCheckAll(t, jts.GeomImpl_PackedCoordinateSequenceFactory_FLOAT_FACTORY)
}

func pcsCheckAll(t *testing.T, factory *jts.GeomImpl_PackedCoordinateSequenceFactory) {
	t.Helper()
	t.Run("Dim2_Size1", func(t *testing.T) { pcsCheckDim2(t, 1, factory) })
	t.Run("Dim2_Size5", func(t *testing.T) { pcsCheckDim2(t, 5, factory) })
	t.Run("Dim3", func(t *testing.T) { pcsCheckDim3(t, factory) })
	t.Run("Dim3_M1", func(t *testing.T) { pcsCheckDim3_M1(t, factory) })
	t.Run("Dim4_M1", func(t *testing.T) { pcsCheckDim4_M1(t, factory) })
	t.Run("Dim4", func(t *testing.T) { pcsCheckDim4(t, factory) })
	t.Run("DimInvalid", func(t *testing.T) { pcsCheckDimInvalid(t, factory) })
}

func pcsCheckDim2(t *testing.T, size int, factory *jts.GeomImpl_PackedCoordinateSequenceFactory) {
	t.Helper()
	seq := factory.CreateWithSizeAndDimension(size, 2)
	pcsInitProgression(seq)

	if seq.GetDimension() != 2 {
		t.Errorf("Dimension should be 2, got %d", seq.GetDimension())
	}
	if seq.HasZ() {
		t.Error("Z should not be present")
	}
	if seq.HasM() {
		t.Error("M should not be present")
	}

	indexLast := size - 1
	valLast := float64(indexLast)

	coord := seq.GetCoordinate(indexLast)
	if !java.InstanceOf[*jts.Geom_CoordinateXY](coord) {
		t.Error("Coordinate should be CoordinateXY")
	}
	if coord.GetX() != valLast {
		t.Errorf("expected X=%v, got %v", valLast, coord.GetX())
	}
	if coord.GetY() != valLast {
		t.Errorf("expected Y=%v, got %v", valLast, coord.GetY())
	}

	array := seq.ToCoordinateArray()
	if !coord.Equals(array[indexLast]) {
		t.Error("coord should equal array[indexLast]")
	}
	if coord == array[indexLast] {
		t.Error("coord should not be same instance as array[indexLast]")
	}
	if !pcsIsEqual(seq, array) {
		t.Error("sequence should equal array")
	}

	copy := factory.CreateFromCoordinates(array)
	if !pcsIsEqual(copy, array) {
		t.Error("copy should equal array")
	}

	copy2 := factory.CreateFromCoordinateSequence(seq)
	if !pcsIsEqual(copy2, array) {
		t.Error("copy2 should equal array")
	}
}

func pcsCheckDim3(t *testing.T, factory *jts.GeomImpl_PackedCoordinateSequenceFactory) {
	t.Helper()
	seq := factory.CreateWithSizeAndDimension(5, 3)
	pcsInitProgression(seq)

	if seq.GetDimension() != 3 {
		t.Errorf("Dimension should be 3, got %d", seq.GetDimension())
	}
	if !seq.HasZ() {
		t.Error("Z should be present")
	}
	if seq.HasM() {
		t.Error("M should not be present")
	}

	coord := seq.GetCoordinate(4)
	if java.GetLeaf(coord) != coord {
		// The coordinate should be a base Coordinate (not XY, XYM, or XYZM).
		if java.InstanceOf[*jts.Geom_CoordinateXY](coord) ||
			java.InstanceOf[*jts.Geom_CoordinateXYM](coord) ||
			java.InstanceOf[*jts.Geom_CoordinateXYZM](coord) {
			t.Error("Coordinate should be base Coordinate class")
		}
	}
	if coord.GetX() != 4.0 {
		t.Errorf("expected X=4.0, got %v", coord.GetX())
	}
	if coord.GetY() != 4.0 {
		t.Errorf("expected Y=4.0, got %v", coord.GetY())
	}
	if coord.GetZ() != 4.0 {
		t.Errorf("expected Z=4.0, got %v", coord.GetZ())
	}

	array := seq.ToCoordinateArray()
	if !coord.Equals(array[4]) {
		t.Error("coord should equal array[4]")
	}
	if coord == array[4] {
		t.Error("coord should not be same instance as array[4]")
	}
	if !pcsIsEqual(seq, array) {
		t.Error("sequence should equal array")
	}

	copy := factory.CreateFromCoordinates(array)
	if !pcsIsEqual(copy, array) {
		t.Error("copy should equal array")
	}

	copy2 := factory.CreateFromCoordinateSequence(seq)
	if !pcsIsEqual(copy2, array) {
		t.Error("copy2 should equal array")
	}
}

func pcsCheckDim3_M1(t *testing.T, factory *jts.GeomImpl_PackedCoordinateSequenceFactory) {
	t.Helper()
	seq := factory.CreateWithSizeAndDimensionAndMeasures(5, 3, 1)
	pcsInitProgression(seq)

	if seq.GetDimension() != 3 {
		t.Errorf("Dimension should be 3, got %d", seq.GetDimension())
	}
	if seq.HasZ() {
		t.Error("Z should not be present")
	}
	if !seq.HasM() {
		t.Error("M should be present")
	}

	coord := seq.GetCoordinate(4)
	if !java.InstanceOf[*jts.Geom_CoordinateXYM](coord) {
		t.Error("Coordinate should be CoordinateXYM")
	}
	if coord.GetX() != 4.0 {
		t.Errorf("expected X=4.0, got %v", coord.GetX())
	}
	if coord.GetY() != 4.0 {
		t.Errorf("expected Y=4.0, got %v", coord.GetY())
	}
	if coord.GetM() != 4.0 {
		t.Errorf("expected M=4.0, got %v", coord.GetM())
	}

	array := seq.ToCoordinateArray()
	if !coord.Equals(array[4]) {
		t.Error("coord should equal array[4]")
	}
	if coord == array[4] {
		t.Error("coord should not be same instance as array[4]")
	}
	if !pcsIsEqual(seq, array) {
		t.Error("sequence should equal array")
	}

	copy := factory.CreateFromCoordinates(array)
	if !pcsIsEqual(copy, array) {
		t.Error("copy should equal array")
	}

	copy2 := factory.CreateFromCoordinateSequence(seq)
	if !pcsIsEqual(copy2, array) {
		t.Error("copy2 should equal array")
	}
}

func pcsCheckDim4_M1(t *testing.T, factory *jts.GeomImpl_PackedCoordinateSequenceFactory) {
	t.Helper()
	seq := factory.CreateWithSizeAndDimensionAndMeasures(5, 4, 1)
	pcsInitProgression(seq)

	if seq.GetDimension() != 4 {
		t.Errorf("Dimension should be 4, got %d", seq.GetDimension())
	}
	if !seq.HasZ() {
		t.Error("Z should be present")
	}
	if !seq.HasM() {
		t.Error("M should be present")
	}

	coord := seq.GetCoordinate(4)
	if !java.InstanceOf[*jts.Geom_CoordinateXYZM](coord) {
		t.Error("Coordinate should be CoordinateXYZM")
	}
	if coord.GetX() != 4.0 {
		t.Errorf("expected X=4.0, got %v", coord.GetX())
	}
	if coord.GetY() != 4.0 {
		t.Errorf("expected Y=4.0, got %v", coord.GetY())
	}
	if coord.GetZ() != 4.0 {
		t.Errorf("expected Z=4.0, got %v", coord.GetZ())
	}
	if coord.GetM() != 4.0 {
		t.Errorf("expected M=4.0, got %v", coord.GetM())
	}

	array := seq.ToCoordinateArray()
	if !coord.Equals(array[4]) {
		t.Error("coord should equal array[4]")
	}
	if coord == array[4] {
		t.Error("coord should not be same instance as array[4]")
	}
	if !pcsIsEqual(seq, array) {
		t.Error("sequence should equal array")
	}

	copy := factory.CreateFromCoordinates(array)
	if !pcsIsEqual(copy, array) {
		t.Error("copy should equal array")
	}

	copy2 := factory.CreateFromCoordinateSequence(seq)
	if !pcsIsEqual(copy2, array) {
		t.Error("copy2 should equal array")
	}
}

func pcsCheckDim4(t *testing.T, factory *jts.GeomImpl_PackedCoordinateSequenceFactory) {
	t.Helper()
	seq := factory.CreateWithSizeAndDimension(5, 4)
	pcsInitProgression(seq)

	if seq.GetDimension() != 4 {
		t.Errorf("Dimension should be 4, got %d", seq.GetDimension())
	}
	if !seq.HasZ() {
		t.Error("Z should be present")
	}
	if !seq.HasM() {
		t.Error("M should be present")
	}

	coord := seq.GetCoordinate(4)
	if !java.InstanceOf[*jts.Geom_CoordinateXYZM](coord) {
		t.Error("Coordinate should be CoordinateXYZM")
	}
	if coord.GetX() != 4.0 {
		t.Errorf("expected X=4.0, got %v", coord.GetX())
	}
	if coord.GetY() != 4.0 {
		t.Errorf("expected Y=4.0, got %v", coord.GetY())
	}
	if coord.GetZ() != 4.0 {
		t.Errorf("expected Z=4.0, got %v", coord.GetZ())
	}
	if coord.GetM() != 4.0 {
		t.Errorf("expected M=4.0, got %v", coord.GetM())
	}

	array := seq.ToCoordinateArray()
	if !coord.Equals(array[4]) {
		t.Error("coord should equal array[4]")
	}
	if coord == array[4] {
		t.Error("coord should not be same instance as array[4]")
	}
	if !pcsIsEqual(seq, array) {
		t.Error("sequence should equal array")
	}

	copy := factory.CreateFromCoordinates(array)
	if !pcsIsEqual(copy, array) {
		t.Error("copy should equal array")
	}

	copy2 := factory.CreateFromCoordinateSequence(seq)
	if !pcsIsEqual(copy2, array) {
		t.Error("copy2 should equal array")
	}
}

func pcsCheckDimInvalid(t *testing.T, factory *jts.GeomImpl_PackedCoordinateSequenceFactory) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Dimension=2/Measure=1 (XM) should have panicked")
		}
	}()
	factory.CreateWithSizeAndDimensionAndMeasures(5, 2, 1)
}

func pcsInitProgression(seq jts.Geom_CoordinateSequence) {
	for index := 0; index < seq.Size(); index++ {
		for ordinateIndex := 0; ordinateIndex < seq.GetDimension(); ordinateIndex++ {
			seq.SetOrdinate(index, ordinateIndex, float64(index))
		}
	}
}

func pcsIsEqual(seq jts.Geom_CoordinateSequence, coords []*jts.Geom_Coordinate) bool {
	if seq.Size() != len(coords) {
		return false
	}

	p := seq.CreateCoordinate()
	for i := 0; i < seq.Size(); i++ {
		if !coords[i].Equals(seq.GetCoordinate(i)) {
			return false
		}

		// Ordinate named getters.
		if !pcsIsEqualFloat(coords[i].GetX(), seq.GetX(i)) {
			return false
		}
		if !pcsIsEqualFloat(coords[i].GetY(), seq.GetY(i)) {
			return false
		}
		if seq.HasZ() {
			if !pcsIsEqualFloat(coords[i].GetZ(), seq.GetZ(i)) {
				return false
			}
		}
		if seq.HasM() {
			if !pcsIsEqualFloat(coords[i].GetM(), seq.GetM(i)) {
				return false
			}
		}

		// Ordinate indexed getters.
		if !pcsIsEqualFloat(coords[i].GetX(), seq.GetOrdinate(i, jts.Geom_CoordinateSequence_X)) {
			return false
		}
		if !pcsIsEqualFloat(coords[i].GetY(), seq.GetOrdinate(i, jts.Geom_CoordinateSequence_Y)) {
			return false
		}
		if seq.GetDimension() > 2 {
			if !pcsIsEqualFloat(coords[i].GetOrdinate(2), seq.GetOrdinate(i, 2)) {
				return false
			}
		}
		if seq.GetDimension() > 3 {
			if !pcsIsEqualFloat(coords[i].GetOrdinate(3), seq.GetOrdinate(i, 3)) {
				return false
			}
		}

		// Coordinate getter.
		seq.GetCoordinateInto(i, p)
		if !pcsIsEqualFloat(coords[i].GetX(), p.GetX()) {
			return false
		}
		if !pcsIsEqualFloat(coords[i].GetY(), p.GetY()) {
			return false
		}
		if seq.HasZ() {
			if !pcsIsEqualFloat(coords[i].GetZ(), p.GetZ()) {
				return false
			}
		}
		if seq.HasM() {
			if !pcsIsEqualFloat(coords[i].GetM(), p.GetM()) {
				return false
			}
		}
	}
	return true
}

func pcsIsEqualFloat(expected, actual float64) bool {
	return expected == actual || (math.IsNaN(expected) && math.IsNaN(actual))
}

// Tests inherited from CoordinateSequenceTestBase.java.

func TestPackedCoordinateSequenceZeroLength(t *testing.T) {
	factory := pcsGetFactory()
	seq := factory.CreateWithSizeAndDimension(0, 3)
	if seq.Size() != 0 {
		t.Errorf("expected size 0, got %d", seq.Size())
	}

	seq2 := factory.CreateFromCoordinates(nil)
	if seq2.Size() != 0 {
		t.Errorf("expected size 0, got %d", seq2.Size())
	}
}

func TestPackedCoordinateSequenceCreateBySizeAndModify(t *testing.T) {
	coords := pcsCreateArray(pcsTestSize)

	factory := pcsGetFactory()
	seq := factory.CreateWithSizeAndDimension(pcsTestSize, 3)
	for i := 0; i < seq.Size(); i++ {
		seq.SetOrdinate(i, 0, coords[i].GetX())
		seq.SetOrdinate(i, 1, coords[i].GetY())
		seq.SetOrdinate(i, 2, coords[i].GetZ())
	}

	if !pcsIsEqual(seq, coords) {
		t.Error("sequence should equal coords")
	}
}

func TestPackedCoordinateSequence2DZOrdinate(t *testing.T) {
	coords := pcsCreateArray(pcsTestSize)

	factory := pcsGetFactory()
	seq := factory.CreateWithSizeAndDimension(pcsTestSize, 2)
	for i := 0; i < seq.Size(); i++ {
		seq.SetOrdinate(i, 0, coords[i].GetX())
		seq.SetOrdinate(i, 1, coords[i].GetY())
	}

	for i := 0; i < seq.Size(); i++ {
		p := seq.GetCoordinate(i)
		if !math.IsNaN(p.GetZ()) {
			t.Errorf("expected Z to be NaN, got %v", p.GetZ())
		}
	}
}

func TestPackedCoordinateSequenceCreateByInit(t *testing.T) {
	coords := pcsCreateArray(pcsTestSize)

	factory := pcsGetFactory()
	seq := factory.CreateFromCoordinates(coords)

	if !pcsIsEqual(seq, coords) {
		t.Error("sequence should equal coords")
	}
}

func TestPackedCoordinateSequenceCreateByInitAndCopy(t *testing.T) {
	coords := pcsCreateArray(pcsTestSize)

	factory := pcsGetFactory()
	seq := factory.CreateFromCoordinates(coords)
	seq2 := factory.CreateFromCoordinateSequence(seq)

	if !pcsIsEqual(seq2, coords) {
		t.Error("seq2 should equal coords")
	}
}

func pcsCreateArray(size int) []*jts.Geom_Coordinate {
	coords := make([]*jts.Geom_Coordinate, size)
	for i := range coords {
		base := float64(2 * 1)
		coords[i] = jts.Geom_NewCoordinateWithXYZ(base, base+1, base+2)
	}
	return coords
}
