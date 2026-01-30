package jts

import "testing"

var (
	testA = Geom_Dimension_A
	testL = Geom_Dimension_L
	testP = Geom_Dimension_P
)

func TestIntersectionMatrixToString(t *testing.T) {
	i := Geom_NewIntersectionMatrix()
	i.SetFromString("012*TF012")
	if got := i.String(); got != "012*TF012" {
		t.Errorf("String() = %v, want 012*TF012", got)
	}

	c := Geom_NewIntersectionMatrixFromMatrix(i)
	if got := c.String(); got != "012*TF012" {
		t.Errorf("String() = %v, want 012*TF012", got)
	}
}

func TestIntersectionMatrixTranspose(t *testing.T) {
	x := Geom_NewIntersectionMatrixWithElements("012*TF012")

	i := Geom_NewIntersectionMatrixFromMatrix(x)
	j := i.Transpose()
	if i != j {
		t.Errorf("Transpose() did not return same pointer")
	}

	if got := i.String(); got != "0*01T12F2" {
		t.Errorf("String() = %v, want 0*01T12F2", got)
	}

	if got := x.String(); got != "012*TF012" {
		t.Errorf("Original unchanged: String() = %v, want 012*TF012", got)
	}
}

func TestIntersectionMatrixIsDisjoint(t *testing.T) {
	if !Geom_NewIntersectionMatrixWithElements("FF*FF****").IsDisjoint() {
		t.Errorf("FF*FF**** should be disjoint")
	}
	if !Geom_NewIntersectionMatrixWithElements("FF1FF2T*0").IsDisjoint() {
		t.Errorf("FF1FF2T*0 should be disjoint")
	}
	if Geom_NewIntersectionMatrixWithElements("*F*FF****").IsDisjoint() {
		t.Errorf("*F*FF**** should not be disjoint")
	}
}

func TestIntersectionMatrixIsTouches(t *testing.T) {
	if !Geom_NewIntersectionMatrixWithElements("FT*******").IsTouches(testP, testA) {
		t.Errorf("FT******* should touch for P,A")
	}
	if !Geom_NewIntersectionMatrixWithElements("FT*******").IsTouches(testA, testP) {
		t.Errorf("FT******* should touch for A,P")
	}
	if Geom_NewIntersectionMatrixWithElements("FT*******").IsTouches(testP, testP) {
		t.Errorf("FT******* should not touch for P,P")
	}
}

func TestIntersectionMatrixIsIntersects(t *testing.T) {
	if Geom_NewIntersectionMatrixWithElements("FF*FF****").IsIntersects() {
		t.Errorf("FF*FF**** should not intersect")
	}
	if Geom_NewIntersectionMatrixWithElements("FF1FF2T*0").IsIntersects() {
		t.Errorf("FF1FF2T*0 should not intersect")
	}
	if !Geom_NewIntersectionMatrixWithElements("*F*FF****").IsIntersects() {
		t.Errorf("*F*FF**** should intersect")
	}
}

func TestIntersectionMatrixIsCrosses(t *testing.T) {
	if !Geom_NewIntersectionMatrixWithElements("TFTFFFFFF").IsCrosses(testP, testL) {
		t.Errorf("TFTFFFFFF should cross for P,L")
	}
	if Geom_NewIntersectionMatrixWithElements("TFTFFFFFF").IsCrosses(testL, testP) {
		t.Errorf("TFTFFFFFF should not cross for L,P")
	}
	if Geom_NewIntersectionMatrixWithElements("TFFFFFTFF").IsCrosses(testP, testL) {
		t.Errorf("TFFFFFTFF should not cross for P,L")
	}
	if !Geom_NewIntersectionMatrixWithElements("TFFFFFTFF").IsCrosses(testL, testP) {
		t.Errorf("TFFFFFTFF should cross for L,P")
	}
	if !Geom_NewIntersectionMatrixWithElements("0FFFFFFFF").IsCrosses(testL, testL) {
		t.Errorf("0FFFFFFFF should cross for L,L")
	}
	if Geom_NewIntersectionMatrixWithElements("1FFFFFFFF").IsCrosses(testL, testL) {
		t.Errorf("1FFFFFFFF should not cross for L,L")
	}
}

func TestIntersectionMatrixIsWithin(t *testing.T) {
	if !Geom_NewIntersectionMatrixWithElements("T0F00F000").IsWithin() {
		t.Errorf("T0F00F000 should be within")
	}
	if Geom_NewIntersectionMatrixWithElements("T00000FF0").IsWithin() {
		t.Errorf("T00000FF0 should not be within")
	}
}

func TestIntersectionMatrixIsContains(t *testing.T) {
	if Geom_NewIntersectionMatrixWithElements("T0F00F000").IsContains() {
		t.Errorf("T0F00F000 should not contain")
	}
	if !Geom_NewIntersectionMatrixWithElements("T00000FF0").IsContains() {
		t.Errorf("T00000FF0 should contain")
	}
}

func TestIntersectionMatrixIsOverlaps(t *testing.T) {
	if !Geom_NewIntersectionMatrixWithElements("2*2***2**").IsOverlaps(testP, testP) {
		t.Errorf("2*2***2** should overlap for P,P")
	}
	if !Geom_NewIntersectionMatrixWithElements("2*2***2**").IsOverlaps(testA, testA) {
		t.Errorf("2*2***2** should overlap for A,A")
	}
	if Geom_NewIntersectionMatrixWithElements("2*2***2**").IsOverlaps(testP, testA) {
		t.Errorf("2*2***2** should not overlap for P,A")
	}
	if Geom_NewIntersectionMatrixWithElements("2*2***2**").IsOverlaps(testL, testL) {
		t.Errorf("2*2***2** should not overlap for L,L")
	}
	if !Geom_NewIntersectionMatrixWithElements("1*2***2**").IsOverlaps(testL, testL) {
		t.Errorf("1*2***2** should overlap for L,L")
	}
	if Geom_NewIntersectionMatrixWithElements("0FFFFFFF2").IsOverlaps(testP, testP) {
		t.Errorf("0FFFFFFF2 should not overlap for P,P")
	}
	if Geom_NewIntersectionMatrixWithElements("1FFF0FFF2").IsOverlaps(testL, testL) {
		t.Errorf("1FFF0FFF2 should not overlap for L,L")
	}
	if Geom_NewIntersectionMatrixWithElements("2FFF1FFF2").IsOverlaps(testA, testA) {
		t.Errorf("2FFF1FFF2 should not overlap for A,A")
	}
}

func TestIntersectionMatrixIsEquals(t *testing.T) {
	if !Geom_NewIntersectionMatrixWithElements("0FFFFFFF2").IsEquals(testP, testP) {
		t.Errorf("0FFFFFFF2 should equal for P,P")
	}
	if !Geom_NewIntersectionMatrixWithElements("1FFF0FFF2").IsEquals(testL, testL) {
		t.Errorf("1FFF0FFF2 should equal for L,L")
	}
	if !Geom_NewIntersectionMatrixWithElements("2FFF1FFF2").IsEquals(testA, testA) {
		t.Errorf("2FFF1FFF2 should equal for A,A")
	}
	if Geom_NewIntersectionMatrixWithElements("0F0FFFFF2").IsEquals(testP, testP) {
		t.Errorf("0F0FFFFF2 should not equal for P,P")
	}
	if !Geom_NewIntersectionMatrixWithElements("1FFF1FFF2").IsEquals(testL, testL) {
		t.Errorf("1FFF1FFF2 should equal for L,L")
	}
	if Geom_NewIntersectionMatrixWithElements("2FFF1*FF2").IsEquals(testA, testA) {
		t.Errorf("2FFF1*FF2 should not equal for A,A")
	}
	if Geom_NewIntersectionMatrixWithElements("0FFFFFFF2").IsEquals(testP, testL) {
		t.Errorf("0FFFFFFF2 should not equal for P,L")
	}
	if Geom_NewIntersectionMatrixWithElements("1FFF0FFF2").IsEquals(testL, testA) {
		t.Errorf("1FFF0FFF2 should not equal for L,A")
	}
	if Geom_NewIntersectionMatrixWithElements("2FFF1FFF2").IsEquals(testA, testP) {
		t.Errorf("2FFF1FFF2 should not equal for A,P")
	}
}
