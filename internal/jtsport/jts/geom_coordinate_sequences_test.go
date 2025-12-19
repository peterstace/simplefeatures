package jts_test

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

var coordSeqsOrdinateValues = [][]float64{
	{75.76, 77.43}, {41.35, 90.75}, {73.74, 41.67}, {20.87, 86.49}, {17.49, 93.59}, {67.75, 80.63},
	{63.01, 52.57}, {32.9, 44.44}, {79.36, 29.8}, {38.17, 88.0}, {19.31, 49.71}, {57.03, 19.28},
	{63.76, 77.35}, {45.26, 85.15}, {51.71, 50.38}, {92.16, 19.85}, {64.18, 27.7}, {64.74, 65.1},
	{80.07, 13.55}, {55.54, 94.07},
}

func coordSeqsGetFactory() jts.Geom_CoordinateSequenceFactory {
	return jts.GeomImpl_CoordinateArraySequenceFactory_Instance()
}

func TestCoordinateSequencesCopyToLargerDim(t *testing.T) {
	csFactory := coordSeqsGetFactory()
	cs2D := coordSeqsCreateTestSequence(csFactory, 10, 2)
	cs3D := csFactory.CreateWithSizeAndDimension(10, 3)
	jts.Geom_CoordinateSequences_Copy(cs2D, 0, cs3D, 0, cs3D.Size())
	if !jts.Geom_CoordinateSequences_IsEqual(cs2D, cs3D) {
		t.Error("expected sequences to be equal")
	}
}

func TestCoordinateSequencesCopyToSmallerDim(t *testing.T) {
	csFactory := coordSeqsGetFactory()
	cs3D := coordSeqsCreateTestSequence(csFactory, 10, 3)
	cs2D := csFactory.CreateWithSizeAndDimension(10, 2)
	jts.Geom_CoordinateSequences_Copy(cs3D, 0, cs2D, 0, cs2D.Size())
	if !jts.Geom_CoordinateSequences_IsEqual(cs2D, cs3D) {
		t.Error("expected sequences to be equal")
	}
}

func TestCoordinateSequencesScrollRing(t *testing.T) {
	coordSeqsDoTestScrollRing(t, coordSeqsGetFactory(), 2)
	coordSeqsDoTestScrollRing(t, coordSeqsGetFactory(), 3)
}

func TestCoordinateSequencesScroll(t *testing.T) {
	coordSeqsDoTestScroll(t, coordSeqsGetFactory(), 2)
	coordSeqsDoTestScroll(t, coordSeqsGetFactory(), 3)
}

func TestCoordinateSequencesIndexOf(t *testing.T) {
	coordSeqsDoTestIndexOf(t, coordSeqsGetFactory(), 2)
}

func TestCoordinateSequencesMinCoordinateIndex(t *testing.T) {
	coordSeqsDoTestMinCoordinateIndex(t, coordSeqsGetFactory(), 2)
}

func TestCoordinateSequencesIsRing(t *testing.T) {
	coordSeqsDoTestIsRing(t, coordSeqsGetFactory(), 2)
}

func TestCoordinateSequencesCopy(t *testing.T) {
	coordSeqsDoTestCopy(t, coordSeqsGetFactory(), 2)
}

func TestCoordinateSequencesReverse(t *testing.T) {
	coordSeqsDoTestReverse(t, coordSeqsGetFactory(), 2)
}

func coordSeqsCreateSequenceFromOrdinates(csFactory jts.Geom_CoordinateSequenceFactory, dim int) jts.Geom_CoordinateSequence {
	sequence := csFactory.CreateWithSizeAndDimension(len(coordSeqsOrdinateValues), dim)
	for i := 0; i < len(coordSeqsOrdinateValues); i++ {
		sequence.SetOrdinate(i, 0, coordSeqsOrdinateValues[i][0])
		sequence.SetOrdinate(i, 1, coordSeqsOrdinateValues[i][1])
	}
	return coordSeqsFillNonPlanarDimensions(sequence)
}

func coordSeqsCreateTestSequence(csFactory jts.Geom_CoordinateSequenceFactory, size, dim int) jts.Geom_CoordinateSequence {
	cs := csFactory.CreateWithSizeAndDimension(size, dim)
	for i := 0; i < size; i++ {
		for d := 0; d < dim; d++ {
			cs.SetOrdinate(i, d, float64(i)*math.Pow(10, float64(d)))
		}
	}
	return cs
}

func coordSeqsFillNonPlanarDimensions(seq jts.Geom_CoordinateSequence) jts.Geom_CoordinateSequence {
	if seq.GetDimension() < 3 {
		return seq
	}
	for i := 0; i < seq.Size(); i++ {
		for j := 2; j < seq.GetDimension(); j++ {
			seq.SetOrdinate(i, j, float64(i)*math.Pow(10, float64(j-1)))
		}
	}
	return seq
}

func coordSeqsDoTestReverse(t *testing.T, factory jts.Geom_CoordinateSequenceFactory, dimension int) {
	t.Helper()
	sequence := coordSeqsCreateSequenceFromOrdinates(factory, dimension)
	reversed := sequence.Copy()
	jts.Geom_CoordinateSequences_Reverse(reversed)
	for i := 0; i < sequence.Size(); i++ {
		coordSeqsCheckCoordinateAt(t, sequence, i, reversed, sequence.Size()-i-1, dimension)
	}
}

func coordSeqsDoTestCopy(t *testing.T, factory jts.Geom_CoordinateSequenceFactory, dimension int) {
	t.Helper()
	sequence := coordSeqsCreateSequenceFromOrdinates(factory, dimension)
	if sequence.Size() <= 7 {
		t.Skip("sequence has insufficient size")
	}

	fullCopy := factory.CreateWithSizeAndDimension(sequence.Size(), dimension)
	partialCopy := factory.CreateWithSizeAndDimension(sequence.Size()-5, dimension)

	jts.Geom_CoordinateSequences_Copy(sequence, 0, fullCopy, 0, sequence.Size())
	jts.Geom_CoordinateSequences_Copy(sequence, 2, partialCopy, 0, partialCopy.Size())

	for i := 0; i < fullCopy.Size(); i++ {
		coordSeqsCheckCoordinateAt(t, sequence, i, fullCopy, i, dimension)
	}
	for i := 0; i < partialCopy.Size(); i++ {
		coordSeqsCheckCoordinateAt(t, sequence, 2+i, partialCopy, i, dimension)
	}
}

func coordSeqsDoTestIsRing(t *testing.T, factory jts.Geom_CoordinateSequenceFactory, dimension int) {
	t.Helper()
	ring := coordSeqsCreateCircle(factory, dimension, jts.Geom_NewCoordinate(), 5)
	noRing := coordSeqsCreateCircularString(factory, dimension, jts.Geom_NewCoordinate(), 5, 0.1, 22)
	empty := coordSeqsCreateAlmostRing(factory, dimension, 0)
	incomplete1 := coordSeqsCreateAlmostRing(factory, dimension, 1)
	incomplete2 := coordSeqsCreateAlmostRing(factory, dimension, 2)
	incomplete3 := coordSeqsCreateAlmostRing(factory, dimension, 3)
	incomplete4a := coordSeqsCreateAlmostRing(factory, dimension, 4)
	incomplete4b := jts.Geom_CoordinateSequences_EnsureValidRing(factory, incomplete4a)

	if !jts.Geom_CoordinateSequences_IsRing(ring) {
		t.Error("expected ring to be a ring")
	}
	if jts.Geom_CoordinateSequences_IsRing(noRing) {
		t.Error("expected noRing to not be a ring")
	}
	if !jts.Geom_CoordinateSequences_IsRing(empty) {
		t.Error("expected empty to be a ring")
	}
	if jts.Geom_CoordinateSequences_IsRing(incomplete1) {
		t.Error("expected incomplete1 to not be a ring")
	}
	if jts.Geom_CoordinateSequences_IsRing(incomplete2) {
		t.Error("expected incomplete2 to not be a ring")
	}
	if jts.Geom_CoordinateSequences_IsRing(incomplete3) {
		t.Error("expected incomplete3 to not be a ring")
	}
	if jts.Geom_CoordinateSequences_IsRing(incomplete4a) {
		t.Error("expected incomplete4a to not be a ring")
	}
	if !jts.Geom_CoordinateSequences_IsRing(incomplete4b) {
		t.Error("expected incomplete4b to be a ring")
	}
}

func coordSeqsDoTestIndexOf(t *testing.T, factory jts.Geom_CoordinateSequenceFactory, dimension int) {
	t.Helper()
	sequence := coordSeqsCreateSequenceFromOrdinates(factory, dimension)
	coordinates := sequence.ToCoordinateArray()
	for i := 0; i < sequence.Size(); i++ {
		idx := jts.Geom_CoordinateSequences_IndexOf(coordinates[i], sequence)
		if idx != i {
			t.Errorf("expected indexOf to return %d, got %d", i, idx)
		}
	}
}

func coordSeqsDoTestMinCoordinateIndex(t *testing.T, factory jts.Geom_CoordinateSequenceFactory, dimension int) {
	t.Helper()
	sequence := coordSeqsCreateSequenceFromOrdinates(factory, dimension)
	if sequence.Size() <= 6 {
		t.Skip("sequence has insufficient size")
	}

	minIndex := sequence.Size() / 2
	sequence.SetOrdinate(minIndex, 0, 5)
	sequence.SetOrdinate(minIndex, 1, 5)

	if jts.Geom_CoordinateSequences_MinCoordinateIndex(sequence) != minIndex {
		t.Errorf("expected minCoordinateIndex to return %d", minIndex)
	}
	if jts.Geom_CoordinateSequences_MinCoordinateIndexInRange(sequence, 2, sequence.Size()-2) != minIndex {
		t.Errorf("expected minCoordinateIndex in range to return %d", minIndex)
	}
}

func coordSeqsDoTestScroll(t *testing.T, factory jts.Geom_CoordinateSequenceFactory, dimension int) {
	t.Helper()
	sequence := coordSeqsCreateCircularString(factory, dimension, jts.Geom_NewCoordinateWithXY(20, 20), 7.0, 0.1, 22)
	scrolled := sequence.Copy()

	jts.Geom_CoordinateSequences_ScrollToIndex(scrolled, 12)

	io := 12
	for is := 0; is < scrolled.Size()-1; is++ {
		coordSeqsCheckCoordinateAt(t, sequence, io, scrolled, is, dimension)
		io++
		io %= scrolled.Size()
	}
}

func coordSeqsDoTestScrollRing(t *testing.T, factory jts.Geom_CoordinateSequenceFactory, dimension int) {
	t.Helper()
	sequence := coordSeqsCreateCircle(factory, dimension, jts.Geom_NewCoordinateWithXY(10, 10), 9.0)
	scrolled := sequence.Copy()

	jts.Geom_CoordinateSequences_ScrollToIndex(scrolled, 12)

	io := 12
	for is := 0; is < scrolled.Size()-1; is++ {
		coordSeqsCheckCoordinateAt(t, sequence, io, scrolled, is, dimension)
		io++
		io %= scrolled.Size() - 1
	}
	coordSeqsCheckCoordinateAt(t, scrolled, 0, scrolled, scrolled.Size()-1, dimension)
}

func coordSeqsCheckCoordinateAt(t *testing.T, seq1 jts.Geom_CoordinateSequence, pos1 int, seq2 jts.Geom_CoordinateSequence, pos2 int, dim int) {
	t.Helper()
	if seq1.GetOrdinate(pos1, 0) != seq2.GetOrdinate(pos2, 0) {
		t.Errorf("unexpected x-ordinate at pos %d: expected %v, got %v", pos2, seq1.GetOrdinate(pos1, 0), seq2.GetOrdinate(pos2, 0))
	}
	if seq1.GetOrdinate(pos1, 1) != seq2.GetOrdinate(pos2, 1) {
		t.Errorf("unexpected y-ordinate at pos %d: expected %v, got %v", pos2, seq1.GetOrdinate(pos1, 1), seq2.GetOrdinate(pos2, 1))
	}
	for j := 2; j < dim; j++ {
		if seq1.GetOrdinate(pos1, j) != seq2.GetOrdinate(pos2, j) {
			t.Errorf("unexpected %d-ordinate at pos %d: expected %v, got %v", j, pos2, seq1.GetOrdinate(pos1, j), seq2.GetOrdinate(pos2, j))
		}
	}
}

func coordSeqsCreateAlmostRing(factory jts.Geom_CoordinateSequenceFactory, dimension int, num int) jts.Geom_CoordinateSequence {
	if num > 4 {
		num = 4
	}
	sequence := factory.CreateWithSizeAndDimension(num, dimension)
	if num == 0 {
		return coordSeqsFillNonPlanarDimensions(sequence)
	}

	sequence.SetOrdinate(0, 0, 10)
	sequence.SetOrdinate(0, 1, 10)
	if num == 1 {
		return coordSeqsFillNonPlanarDimensions(sequence)
	}

	sequence.SetOrdinate(1, 0, 20)
	sequence.SetOrdinate(1, 1, 10)
	if num == 2 {
		return coordSeqsFillNonPlanarDimensions(sequence)
	}

	sequence.SetOrdinate(2, 0, 20)
	sequence.SetOrdinate(2, 1, 20)
	if num == 3 {
		return coordSeqsFillNonPlanarDimensions(sequence)
	}

	sequence.SetOrdinate(3, 0, 10.0000000000001)
	sequence.SetOrdinate(3, 1, 9.9999999999999)
	return coordSeqsFillNonPlanarDimensions(sequence)
}

func coordSeqsCreateCircle(factory jts.Geom_CoordinateSequenceFactory, dimension int, center *jts.Geom_Coordinate, radius float64) jts.Geom_CoordinateSequence {
	res := coordSeqsCreateCircularString(factory, dimension, center, radius, 0, 49)
	for i := 0; i < dimension; i++ {
		res.SetOrdinate(48, i, res.GetOrdinate(0, i))
	}
	return res
}

func coordSeqsCreateCircularString(factory jts.Geom_CoordinateSequenceFactory, dimension int, center *jts.Geom_Coordinate, radius float64, startAngle float64, numPoints int) jts.Geom_CoordinateSequence {
	const numSegmentsCircle = 48
	const angleCircle = 2 * math.Pi
	const angleStep = angleCircle / numSegmentsCircle

	sequence := factory.CreateWithSizeAndDimension(numPoints, dimension)
	pm := jts.Geom_NewPrecisionModelWithScale(100)
	angle := startAngle
	for i := 0; i < numPoints; i++ {
		dx := math.Cos(angle) * radius
		sequence.SetOrdinate(i, 0, pm.MakePrecise(center.X+dx))
		dy := math.Sin(angle) * radius
		sequence.SetOrdinate(i, 1, pm.MakePrecise(center.Y+dy))

		for j := 2; j < dimension; j++ {
			sequence.SetOrdinate(i, j, math.Pow(10, float64(j-1))*float64(i))
		}

		angle += angleStep
		angle = math.Mod(angle, angleCircle)
	}
	return sequence
}
