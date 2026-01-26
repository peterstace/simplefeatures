package jts

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestConstructor3D(t *testing.T) {
	c := Geom_NewCoordinateWithXYZ(350.2, 4566.8, 5266.3)
	junit.AssertEquals(t, 350.2, c.X)
	junit.AssertEquals(t, 4566.8, c.Y)
	junit.AssertEquals(t, 5266.3, c.GetZ())
}

func TestConstructor2D(t *testing.T) {
	c := Geom_NewCoordinateWithXY(350.2, 4566.8)
	junit.AssertEquals(t, 350.2, c.X)
	junit.AssertEquals(t, 4566.8, c.Y)
	junit.AssertEqualsNaN(t, Geom_Coordinate_NullOrdinate, c.GetZ())
}

func TestDefaultConstructor(t *testing.T) {
	c := Geom_NewCoordinate()
	junit.AssertEquals(t, 0.0, c.X)
	junit.AssertEquals(t, 0.0, c.Y)
	junit.AssertEqualsNaN(t, Geom_Coordinate_NullOrdinate, c.GetZ())
}

func TestCopyConstructor3D(t *testing.T) {
	orig := Geom_NewCoordinateWithXYZ(350.2, 4566.8, 5266.3)
	c := Geom_NewCoordinateFromCoordinate(orig)
	junit.AssertEquals(t, 350.2, c.X)
	junit.AssertEquals(t, 4566.8, c.Y)
	junit.AssertEquals(t, 5266.3, c.GetZ())
}

func TestSetCoordinate(t *testing.T) {
	orig := Geom_NewCoordinateWithXYZ(350.2, 4566.8, 5266.3)
	c := Geom_NewCoordinate()
	c.SetCoordinate(orig)
	junit.AssertEquals(t, 350.2, c.X)
	junit.AssertEquals(t, 4566.8, c.Y)
	junit.AssertEquals(t, 5266.3, c.GetZ())
}

func TestGetOrdinate(t *testing.T) {
	c := Geom_NewCoordinateWithXYZ(350.2, 4566.8, 5266.3)
	junit.AssertEquals(t, 350.2, c.GetOrdinate(Geom_Coordinate_X))
	junit.AssertEquals(t, 4566.8, c.GetOrdinate(Geom_Coordinate_Y))
	junit.AssertEquals(t, 5266.3, c.GetOrdinate(Geom_Coordinate_Z))
}

func TestSetOrdinate(t *testing.T) {
	c := Geom_NewCoordinate()
	c.SetOrdinate(Geom_Coordinate_X, 111)
	c.SetOrdinate(Geom_Coordinate_Y, 222)
	c.SetOrdinate(Geom_Coordinate_Z, 333)
	junit.AssertEquals(t, 111.0, c.GetOrdinate(Geom_Coordinate_X))
	junit.AssertEquals(t, 222.0, c.GetOrdinate(Geom_Coordinate_Y))
	junit.AssertEquals(t, 333.0, c.GetOrdinate(Geom_Coordinate_Z))
}

func TestEqualsCoordinate(t *testing.T) {
	c1 := Geom_NewCoordinateWithXYZ(1, 2, 3)
	c2 := Geom_NewCoordinateWithXYZ(1, 2, 3)
	junit.AssertTrue(t, c1.Equals2D(c2))

	c3 := Geom_NewCoordinateWithXYZ(1, 22, 3)
	junit.AssertTrue(t, !c1.Equals2D(c3))
}

func TestEquals2D(t *testing.T) {
	c1 := Geom_NewCoordinateWithXYZ(1, 2, 3)
	c2 := Geom_NewCoordinateWithXYZ(1, 2, 3)
	junit.AssertTrue(t, c1.Equals2D(c2))

	c3 := Geom_NewCoordinateWithXYZ(1, 22, 3)
	junit.AssertTrue(t, !c1.Equals2D(c3))
}

func TestEquals3D(t *testing.T) {
	c1 := Geom_NewCoordinateWithXYZ(1, 2, 3)
	c2 := Geom_NewCoordinateWithXYZ(1, 2, 3)
	junit.AssertTrue(t, c1.Equals3D(c2))

	c3 := Geom_NewCoordinateWithXYZ(1, 22, 3)
	junit.AssertTrue(t, !c1.Equals3D(c3))
}

func TestEquals2DWithTolerance(t *testing.T) {
	c := Geom_NewCoordinateWithXYZ(100.0, 200.0, 50.0)
	aBitOff := Geom_NewCoordinateWithXYZ(100.1, 200.1, 50.0)
	junit.AssertTrue(t, c.Equals2DWithTolerance(aBitOff, 0.2))
}

func TestEqualsInZ(t *testing.T) {
	c := Geom_NewCoordinateWithXYZ(100.0, 200.0, 50.0)
	withSameZ := Geom_NewCoordinateWithXYZ(100.1, 200.1, 50.1)
	junit.AssertTrue(t, c.EqualInZ(withSameZ, 0.2))
}

func TestCompareTo(t *testing.T) {
	lowest := Geom_NewCoordinateWithXYZ(10.0, 100.0, 50.0)
	highest := Geom_NewCoordinateWithXYZ(20.0, 100.0, 50.0)
	equalToHighest := Geom_NewCoordinateWithXYZ(20.0, 100.0, 50.0)
	higherStill := Geom_NewCoordinateWithXYZ(20.0, 200.0, 50.0)

	junit.AssertEquals(t, -1, lowest.CompareTo(highest))
	junit.AssertEquals(t, 1, highest.CompareTo(lowest))
	junit.AssertEquals(t, -1, highest.CompareTo(higherStill))
	junit.AssertEquals(t, 0, highest.CompareTo(equalToHighest))
}

func TestToString(t *testing.T) {
	expectedResult := "(100, 200, 50)"
	actualResult := Geom_NewCoordinateWithXYZ(100.0, 200.0, 50.0).String()
	junit.AssertEquals(t, expectedResult, actualResult)
}

func TestCopyCoordinate(t *testing.T) {
	c := Geom_NewCoordinateWithXYZ(100.0, 200.0, 50.0)
	clone := c.Copy()
	junit.AssertTrue(t, c.Equals3D(clone))
}

func TestDistanceCoordinate(t *testing.T) {
	coord1 := Geom_NewCoordinateWithXYZ(0.0, 0.0, 0.0)
	coord2 := Geom_NewCoordinateWithXYZ(100.0, 200.0, 50.0)
	distance := coord1.Distance(coord2)
	junit.AssertEqualsFloat64(t, 223.60679774997897, distance, 0.00001)
}

func TestDistance3D(t *testing.T) {
	coord1 := Geom_NewCoordinateWithXYZ(0.0, 0.0, 0.0)
	coord2 := Geom_NewCoordinateWithXYZ(100.0, 200.0, 50.0)
	distance := coord1.Distance3D(coord2)
	junit.AssertEqualsFloat64(t, 229.128784747792, distance, 0.000001)
}

func TestCoordinateXY(t *testing.T) {
	xy := Geom_NewCoordinateXY2D()
	checkZUnsupported(t, xy.Geom_Coordinate)
	checkMUnsupported(t, xy.Geom_Coordinate)

	xy = Geom_NewCoordinateXY2DWithXY(1.0, 1.0)
	coord := Geom_NewCoordinateFromCoordinate(xy.Geom_Coordinate)
	junit.AssertTrue(t, xy.Equals(coord))
	junit.AssertTrue(t, !xy.EqualInZ(coord, 0.000001))

	coord = Geom_NewCoordinateWithXYZ(1.0, 1.0, 1.0)
	xy = Geom_NewCoordinateXY2DFromCoordinate(coord)
	junit.AssertTrue(t, xy.Equals(coord))
	junit.AssertTrue(t, !xy.EqualInZ(coord, 0.000001))
}

func TestCoordinateXYM(t *testing.T) {
	xym := Geom_NewCoordinateXYM3D()
	checkZUnsupported(t, xym.Geom_Coordinate)

	xym.SetM(1.0)
	junit.AssertEquals(t, 1.0, xym.GetM())

	coord := Geom_NewCoordinateFromCoordinate(xym.Geom_Coordinate)
	junit.AssertTrue(t, xym.Equals(coord))
	junit.AssertTrue(t, !xym.EqualInZ(coord, 0.000001))

	coord = Geom_NewCoordinateWithXYZ(1.0, 1.0, 1.0)
	xym = Geom_NewCoordinateXYM3DFromCoordinate(coord)
	junit.AssertTrue(t, xym.Equals(coord))
	junit.AssertTrue(t, !xym.EqualInZ(coord, 0.000001))
}

func TestCoordinateXYZM(t *testing.T) {
	xyzm := Geom_NewCoordinateXYZM4D()
	xyzm.SetZ(1.0)
	junit.AssertEquals(t, 1.0, xyzm.GetZ())
	xyzm.SetM(1.0)
	junit.AssertEquals(t, 1.0, xyzm.GetM())

	coord := Geom_NewCoordinateFromCoordinate(xyzm.Geom_Coordinate)
	junit.AssertTrue(t, xyzm.Equals(coord))
	junit.AssertTrue(t, xyzm.EqualInZ(coord, 0.000001))
	junit.AssertTrue(t, math.IsNaN(coord.GetM()))

	coord = Geom_NewCoordinateWithXYZ(1.0, 1.0, 1.0)
	xyzm = Geom_NewCoordinateXYZM4DFromCoordinate(coord)
	junit.AssertTrue(t, xyzm.Equals(coord))
	junit.AssertTrue(t, xyzm.EqualInZ(coord, 0.000001))
}

func TestCoordinateHash(t *testing.T) {
	doTestCoordinateHash(t, true, Geom_NewCoordinateWithXY(1, 2), Geom_NewCoordinateWithXY(1, 2))
	doTestCoordinateHash(t, false, Geom_NewCoordinateWithXY(1, 2), Geom_NewCoordinateWithXY(3, 4))
	doTestCoordinateHash(t, false, Geom_NewCoordinateWithXY(1, 2), Geom_NewCoordinateWithXY(1, 4))
	doTestCoordinateHash(t, false, Geom_NewCoordinateWithXY(1, 2), Geom_NewCoordinateWithXY(3, 2))
	doTestCoordinateHash(t, false, Geom_NewCoordinateWithXY(1, 2), Geom_NewCoordinateWithXY(2, 1))
}

func doTestCoordinateHash(t *testing.T, equal bool, a, b *Geom_Coordinate) {
	junit.AssertEquals(t, equal, a.Equals(b))
	junit.AssertEquals(t, equal, a.HashCode() == b.HashCode())
}

// checkZUnsupported confirms the z field is not supported by GetZ and SetZ.
func checkZUnsupported(t *testing.T, coord *Geom_Coordinate) {
	defer func() {
		if r := recover(); r == nil {
			junit.Fail(t, coord.String()+" does not support Z")
		}
	}()
	coord.SetZ(0.0)
	junit.AssertTrue(t, math.IsNaN(coord.Z))
	coord.Z = 0.0
	junit.AssertTrue(t, math.IsNaN(coord.GetZ()))
}

// checkMUnsupported confirms the M measure is not supported by GetM and SetM.
func checkMUnsupported(t *testing.T, coord *Geom_Coordinate) {
	defer func() {
		if r := recover(); r == nil {
			junit.Fail(t, coord.String()+" does not support M")
		}
	}()
	coord.SetM(0.0)
	junit.AssertTrue(t, math.IsNaN(coord.GetM()))
}
