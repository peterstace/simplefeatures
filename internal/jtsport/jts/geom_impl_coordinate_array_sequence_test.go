package jts_test

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

const casTestSize = 100

func casGetFactory() *jts.GeomImpl_CoordinateArraySequenceFactory {
	return jts.GeomImpl_CoordinateArraySequenceFactory_Instance()
}

func TestCoordinateArraySequenceZeroLength(t *testing.T) {
	seq := casGetFactory().CreateWithSizeAndDimension(0, 3)
	if seq.Size() != 0 {
		t.Errorf("expected size 0, got %d", seq.Size())
	}

	seq2 := casGetFactory().CreateFromCoordinates(nil)
	if seq2.Size() != 0 {
		t.Errorf("expected size 0, got %d", seq2.Size())
	}
}

func TestCoordinateArraySequenceCreateBySizeAndModify(t *testing.T) {
	coords := casCreateArray(casTestSize)
	seq := casGetFactory().CreateWithSizeAndDimension(casTestSize, 3)
	for i := 0; i < seq.Size(); i++ {
		seq.SetOrdinate(i, 0, coords[i].X)
		seq.SetOrdinate(i, 1, coords[i].Y)
		seq.SetOrdinate(i, 2, coords[i].GetZ())
	}
	if !casIsEqual(seq, coords) {
		t.Error("sequences should be equal")
	}
}

func TestCoordinateArraySequence2DZOrdinate(t *testing.T) {
	coords := casCreateArray(casTestSize)
	seq := casGetFactory().CreateWithSizeAndDimension(casTestSize, 2)
	for i := 0; i < seq.Size(); i++ {
		seq.SetOrdinate(i, 0, coords[i].X)
		seq.SetOrdinate(i, 1, coords[i].Y)
	}
	for i := 0; i < seq.Size(); i++ {
		p := seq.GetCoordinate(i)
		if !math.IsNaN(p.GetZ()) {
			t.Errorf("expected NaN for Z ordinate at index %d, got %v", i, p.GetZ())
		}
	}
}

func TestCoordinateArraySequenceCreateByInit(t *testing.T) {
	coords := casCreateArray(casTestSize)
	seq := casGetFactory().CreateFromCoordinates(coords)
	if !casIsEqual(seq, coords) {
		t.Error("sequences should be equal")
	}
}

func TestCoordinateArraySequenceCreateByInitAndCopy(t *testing.T) {
	coords := casCreateArray(casTestSize)
	seq := casGetFactory().CreateFromCoordinates(coords)
	seq2 := casGetFactory().CreateFromCoordinateSequence(seq)
	if !casIsEqual(seq2, coords) {
		t.Error("sequences should be equal")
	}
}

func TestCoordinateArraySequenceFactoryLimits(t *testing.T) {
	factory := casGetFactory()
	csFactory := factory

	sequence := csFactory.CreateWithSizeAndDimension(10, 4)
	if sequence.GetDimension() != 3 {
		t.Errorf("expected clipped dimension 3, got %d", sequence.GetDimension())
	}
	if sequence.GetMeasures() != 0 {
		t.Errorf("expected default measure 0, got %d", sequence.GetMeasures())
	}
	if !sequence.HasZ() {
		t.Error("expected hasZ true")
	}
	if sequence.HasM() {
		t.Error("expected hasM false")
	}

	sequence = csFactory.CreateWithSizeAndDimensionAndMeasures(10, 4, 0)
	if sequence.GetDimension() != 3 {
		t.Errorf("expected clipped dimension 3, got %d", sequence.GetDimension())
	}
	if sequence.GetMeasures() != 0 {
		t.Errorf("expected provided measure 0, got %d", sequence.GetMeasures())
	}
	if !sequence.HasZ() {
		t.Error("expected hasZ true")
	}
	if sequence.HasM() {
		t.Error("expected hasM false")
	}

	sequence = csFactory.CreateWithSizeAndDimensionAndMeasures(10, 4, 2)
	if sequence.GetDimension() != 3 {
		t.Errorf("expected clipped dimension 3, got %d", sequence.GetDimension())
	}
	if sequence.GetMeasures() != 1 {
		t.Errorf("expected clipped measure 1, got %d", sequence.GetMeasures())
	}
	if sequence.HasZ() {
		t.Error("expected hasZ false")
	}
	if !sequence.HasM() {
		t.Error("expected hasM true")
	}

	sequence = csFactory.CreateWithSizeAndDimensionAndMeasures(10, 5, 1)
	if sequence.GetDimension() != 4 {
		t.Errorf("expected clipped dimension 4, got %d", sequence.GetDimension())
	}
	if sequence.GetMeasures() != 1 {
		t.Errorf("expected provided measure 1, got %d", sequence.GetMeasures())
	}
	if !sequence.HasZ() {
		t.Error("expected hasZ true")
	}
	if !sequence.HasM() {
		t.Error("expected hasM true")
	}

	sequence = csFactory.CreateWithSizeAndDimension(10, 1)
	if sequence.GetDimension() != 2 {
		t.Errorf("expected clipped dimension 2, got %d", sequence.GetDimension())
	}
	if sequence.GetMeasures() != 0 {
		t.Errorf("expected default measure 0, got %d", sequence.GetMeasures())
	}
	if sequence.HasZ() {
		t.Error("expected hasZ false")
	}
	if sequence.HasM() {
		t.Error("expected hasM false")
	}

	sequence = csFactory.CreateWithSizeAndDimensionAndMeasures(10, 2, 1)
	if sequence.GetDimension() != 3 {
		t.Errorf("expected clipped dimension 3, got %d", sequence.GetDimension())
	}
	if sequence.GetMeasures() != 1 {
		t.Errorf("expected provided measure 1, got %d", sequence.GetMeasures())
	}
	if sequence.HasZ() {
		t.Error("expected hasZ false")
	}
	if !sequence.HasM() {
		t.Error("expected hasM true")
	}
}

func TestCoordinateArraySequenceDimensionAndMeasure(t *testing.T) {
	factory := casGetFactory()

	// Test XY (dimension 2)
	seq := factory.CreateWithSizeAndDimension(5, 2)
	casInitProgression(seq)
	if seq.GetDimension() != 2 {
		t.Errorf("expected dimension 2, got %d", seq.GetDimension())
	}
	if seq.HasZ() {
		t.Error("expected hasZ false for XY")
	}
	if seq.HasM() {
		t.Error("expected hasM false for XY")
	}
	coord := seq.GetCoordinate(4)
	if _, ok := java.GetLeaf(coord).(*jts.Geom_CoordinateXY); !ok {
		t.Error("expected CoordinateXY type")
	}
	if coord.GetX() != 4.0 {
		t.Errorf("expected X=4.0, got %v", coord.GetX())
	}
	if coord.GetY() != 4.0 {
		t.Errorf("expected Y=4.0, got %v", coord.GetY())
	}
	array := seq.ToCoordinateArray()
	if !coord.Equals(array[4]) {
		t.Error("expected coord to equal array[4]")
	}
	if !casIsEqual(seq, array) {
		t.Error("expected seq to equal array")
	}
	copy := factory.CreateFromCoordinates(array)
	if !casIsEqual(copy, array) {
		t.Error("expected copy to equal array")
	}
	copy = factory.CreateFromCoordinateSequence(seq)
	if !casIsEqual(copy, array) {
		t.Error("expected copy to equal array")
	}

	// Test XYZ (dimension 3)
	seq = factory.CreateWithSizeAndDimension(5, 3)
	casInitProgression(seq)
	if seq.GetDimension() != 3 {
		t.Errorf("expected dimension 3, got %d", seq.GetDimension())
	}
	if !seq.HasZ() {
		t.Error("expected hasZ true for XYZ")
	}
	if seq.HasM() {
		t.Error("expected hasM false for XYZ")
	}
	coord = seq.GetCoordinate(4)
	self := java.GetLeaf(coord)
	if _, ok := self.(*jts.Geom_CoordinateXY); ok {
		t.Error("expected plain Coordinate type, not CoordinateXY")
	}
	if _, ok := self.(*jts.Geom_CoordinateXYM); ok {
		t.Error("expected plain Coordinate type, not CoordinateXYM")
	}
	if _, ok := self.(*jts.Geom_CoordinateXYZM); ok {
		t.Error("expected plain Coordinate type, not CoordinateXYZM")
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
	array = seq.ToCoordinateArray()
	if !coord.Equals(array[4]) {
		t.Error("expected coord to equal array[4]")
	}
	if !casIsEqual(seq, array) {
		t.Error("expected seq to equal array")
	}
	copy = factory.CreateFromCoordinates(array)
	if !casIsEqual(copy, array) {
		t.Error("expected copy to equal array")
	}
	copy = factory.CreateFromCoordinateSequence(seq)
	if !casIsEqual(copy, array) {
		t.Error("expected copy to equal array")
	}

	// Test XYM (dimension 3, measure 1)
	seq = factory.CreateWithSizeAndDimensionAndMeasures(5, 3, 1)
	casInitProgression(seq)
	if seq.GetDimension() != 3 {
		t.Errorf("expected dimension 3, got %d", seq.GetDimension())
	}
	if seq.HasZ() {
		t.Error("expected hasZ false for XYM")
	}
	if !seq.HasM() {
		t.Error("expected hasM true for XYM")
	}
	coord = seq.GetCoordinate(4)
	if _, ok := java.GetLeaf(coord).(*jts.Geom_CoordinateXYM); !ok {
		t.Error("expected CoordinateXYM type")
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
	array = seq.ToCoordinateArray()
	if !coord.Equals(array[4]) {
		t.Error("expected coord to equal array[4]")
	}
	if !casIsEqual(seq, array) {
		t.Error("expected seq to equal array")
	}
	copy = factory.CreateFromCoordinates(array)
	if !casIsEqual(copy, array) {
		t.Error("expected copy to equal array")
	}
	copy = factory.CreateFromCoordinateSequence(seq)
	if !casIsEqual(copy, array) {
		t.Error("expected copy to equal array")
	}

	// Test XYZM (dimension 4, measure 1)
	seq = factory.CreateWithSizeAndDimensionAndMeasures(5, 4, 1)
	casInitProgression(seq)
	if seq.GetDimension() != 4 {
		t.Errorf("expected dimension 4, got %d", seq.GetDimension())
	}
	if !seq.HasZ() {
		t.Error("expected hasZ true for XYZM")
	}
	if !seq.HasM() {
		t.Error("expected hasM true for XYZM")
	}
	coord = seq.GetCoordinate(4)
	if _, ok := java.GetLeaf(coord).(*jts.Geom_CoordinateXYZM); !ok {
		t.Error("expected CoordinateXYZM type")
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
	array = seq.ToCoordinateArray()
	if !coord.Equals(array[4]) {
		t.Error("expected coord to equal array[4]")
	}
	if !casIsEqual(seq, array) {
		t.Error("expected seq to equal array")
	}
	copy = factory.CreateFromCoordinates(array)
	if !casIsEqual(copy, array) {
		t.Error("expected copy to equal array")
	}
	copy = factory.CreateFromCoordinateSequence(seq)
	if !casIsEqual(copy, array) {
		t.Error("expected copy to equal array")
	}

	// Test XM clipped to XYM
	seq = factory.CreateWithSizeAndDimensionAndMeasures(5, 2, 1)
	if seq.GetDimension() != 3 {
		t.Errorf("expected dimension 3, got %d", seq.GetDimension())
	}
	if seq.GetMeasures() != 1 {
		t.Errorf("expected measures 1, got %d", seq.GetMeasures())
	}
}

func TestCoordinateArraySequenceMixedCoordinates(t *testing.T) {
	factory := casGetFactory()

	coord1 := jts.Geom_NewCoordinateWithXYZ(1.0, 1.0, 1.0)
	coord2 := jts.Geom_NewCoordinateXY2DWithXY(2.0, 2.0)
	coord3 := jts.Geom_NewCoordinateXYM3DWithXYM(3.0, 3.0, 3.0)

	array := []*jts.Geom_Coordinate{coord1, coord2.Geom_Coordinate, coord3.Geom_Coordinate, nil}
	seq := factory.CreateFromCoordinates(array)

	if seq.GetDimension() != 3 {
		t.Errorf("expected dimension 3, got %d", seq.GetDimension())
	}
	if seq.GetMeasures() != 1 {
		t.Errorf("expected measures 1, got %d", seq.GetMeasures())
	}
	if !coord1.Equals(seq.GetCoordinate(0)) {
		t.Error("expected coord1 to equal seq[0]")
	}
	if !coord2.Equals(seq.GetCoordinate(1)) {
		t.Error("expected coord2 to equal seq[1]")
	}
	if !coord3.Equals(seq.GetCoordinate(2)) {
		t.Error("expected coord3 to equal seq[2]")
	}
	if seq.GetCoordinate(3) != nil {
		t.Error("expected seq[3] to be nil")
	}
}

func casCreateArray(size int) []*jts.Geom_Coordinate {
	coords := make([]*jts.Geom_Coordinate, size)
	for i := 0; i < size; i++ {
		base := float64(2 * 1)
		coords[i] = jts.Geom_NewCoordinateWithXYZ(base, base+1, base+2)
	}
	return coords
}

func casIsEqual(seq jts.Geom_CoordinateSequence, coords []*jts.Geom_Coordinate) bool {
	if seq.Size() != len(coords) {
		return false
	}

	p := seq.CreateCoordinate()
	for i := 0; i < seq.Size(); i++ {
		if !coords[i].Equals(seq.GetCoordinate(i)) {
			return false
		}

		// Ordinate named getters.
		if !casIsEqualFloat(coords[i].X, seq.GetX(i)) {
			return false
		}
		if !casIsEqualFloat(coords[i].Y, seq.GetY(i)) {
			return false
		}
		if seq.HasZ() {
			if !casIsEqualFloat(coords[i].GetZ(), seq.GetZ(i)) {
				return false
			}
		}
		if seq.HasM() {
			if !casIsEqualFloat(coords[i].GetM(), seq.GetM(i)) {
				return false
			}
		}

		// Ordinate indexed getters.
		if !casIsEqualFloat(coords[i].X, seq.GetOrdinate(i, jts.Geom_CoordinateSequence_X)) {
			return false
		}
		if !casIsEqualFloat(coords[i].Y, seq.GetOrdinate(i, jts.Geom_CoordinateSequence_Y)) {
			return false
		}
		if seq.GetDimension() > 2 {
			if !casIsEqualFloat(coords[i].GetOrdinate(2), seq.GetOrdinate(i, 2)) {
				return false
			}
		}
		if seq.GetDimension() > 3 {
			if !casIsEqualFloat(coords[i].GetOrdinate(3), seq.GetOrdinate(i, 3)) {
				return false
			}
		}

		// Coordinate getter.
		seq.GetCoordinateInto(i, p)
		if !casIsEqualFloat(coords[i].X, p.X) {
			return false
		}
		if !casIsEqualFloat(coords[i].Y, p.Y) {
			return false
		}
		if seq.HasZ() {
			if !casIsEqualFloat(coords[i].GetZ(), p.GetZ()) {
				return false
			}
		}
		if seq.HasM() {
			if !casIsEqualFloat(coords[i].GetM(), p.GetM()) {
				return false
			}
		}
	}
	return true
}

func casIsEqualFloat(expected, actual float64) bool {
	return expected == actual || (math.IsNaN(expected) && math.IsNaN(actual))
}

func casInitProgression(seq jts.Geom_CoordinateSequence) {
	for index := 0; index < seq.Size(); index++ {
		for ordinateIndex := 0; ordinateIndex < seq.GetDimension(); ordinateIndex++ {
			seq.SetOrdinate(index, ordinateIndex, float64(index))
		}
	}
}
