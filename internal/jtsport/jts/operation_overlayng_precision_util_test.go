package jts

import "testing"

func TestPrecisionUtilInts(t *testing.T) {
	checkRobustScale(t, "POINT(1 1)", "POINT(10 10)", 1, 1e12, 1)
}

func TestPrecisionUtilBNull(t *testing.T) {
	checkRobustScale(t, "POINT(1 1)", "", 1, 1e13, 1)
}

func TestPrecisionUtilPower10(t *testing.T) {
	checkRobustScale(t, "POINT(100 100)", "POINT(1000 1000)", 1, 1e11, 1)
}

func TestPrecisionUtilDecimalsDifferent(t *testing.T) {
	checkRobustScale(t, "POINT( 1.123 1.12 )", "POINT( 10.123 10.12345 )", 1e5, 1e12, 1e5)
}

func TestPrecisionUtilDecimalsShort(t *testing.T) {
	checkRobustScale(t, "POINT(1 1.12345)", "POINT(10 10)", 1e5, 1e12, 1e5)
}

func TestPrecisionUtilDecimalsMany(t *testing.T) {
	checkRobustScale(t, "POINT(1 1.123451234512345)", "POINT(10 10)", 1e12, 1e12, 1e15)
}

func TestPrecisionUtilDecimalsAllLong(t *testing.T) {
	checkRobustScale(t, "POINT( 1.123451234512345 1.123451234512345 )", "POINT( 10.123451234512345 10.123451234512345 )", 1e12, 1e12, 1e15)
}

func TestPrecisionUtilSafeScaleChosen(t *testing.T) {
	checkRobustScale(t, "POINT( 123123.123451234512345 1 )", "POINT( 10 10 )", 1e8, 1e8, 1e11)
}

func TestPrecisionUtilSafeScaleChosenLargeMagnitude(t *testing.T) {
	checkRobustScale(t, "POINT( 123123123.123451234512345 1 )", "POINT( 10 10 )", 1e5, 1e5, 1e8)
}

func TestPrecisionUtilInherentWithLargeMagnitude(t *testing.T) {
	checkRobustScale(t, "POINT( 123123123.12 1 )", "POINT( 10 10 )", 1e2, 1e5, 1e2)
}

func TestPrecisionUtilMixedMagnitude(t *testing.T) {
	checkRobustScale(t, "POINT( 1.123451234512345 1 )", "POINT( 100000.12345 10 )", 1e8, 1e8, 1e15)
}

func TestPrecisionUtilInherentBelowSafe(t *testing.T) {
	checkRobustScale(t, "POINT( 100000.1234512 1 )", "POINT( 100000.12345 10 )", 1e7, 1e8, 1e7)
}

func checkRobustScale(t *testing.T, wktA, wktB string, scaleExpected, safeScaleExpected, inherentScaleExpected float64) {
	t.Helper()
	reader := Io_NewWKTReader()
	a, err := reader.Read(wktA)
	if err != nil {
		t.Fatalf("failed to read wktA: %v", err)
	}
	var b *Geom_Geometry
	if wktB != "" {
		b, err = reader.Read(wktB)
		if err != nil {
			t.Fatalf("failed to read wktB: %v", err)
		}
	}

	robustScale := OperationOverlayng_PrecisionUtil_RobustScale(a, b)
	if robustScale != scaleExpected {
		t.Errorf("Auto scale: expected %v, got %v", scaleExpected, robustScale)
	}

	inherentScale := OperationOverlayng_PrecisionUtil_InherentScaleGeoms(a, b)
	if inherentScale != inherentScaleExpected {
		t.Errorf("Inherent scale: expected %v, got %v", inherentScaleExpected, inherentScale)
	}

	safeScale := OperationOverlayng_PrecisionUtil_SafeScaleGeoms(a, b)
	if safeScale != safeScaleExpected {
		t.Errorf("Safe scale: expected %v, got %v", safeScaleExpected, safeScale)
	}
}
