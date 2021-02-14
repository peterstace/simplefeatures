package geom

import (
	"strconv"
	"testing"
)

func TestMatrixZeroValue(t *testing.T) {
	var m IntersectionMatrix
	const want = "FFFFFFFFF"
	got := m.StringCode()
	if got != want {
		t.Errorf("want=%v got=%v", want, got)
	}
}

func TestMatrixWith(t *testing.T) {
	for i, tc := range []struct {
		code   string
		matrix func() IntersectionMatrix
	}{
		{"FFFFFFFFF", func() IntersectionMatrix {
			var m IntersectionMatrix
			return m
		}},
		{"2FFFFFFFF", func() IntersectionMatrix {
			var m IntersectionMatrix
			m = m.with(imInterior, imInterior, imEntry2)
			return m
		}},
		{"F1FFFFFFF", func() IntersectionMatrix {
			var m IntersectionMatrix
			m = m.with(imInterior, imBoundary, imEntry1)
			return m
		}},
		{"2121012F2", func() IntersectionMatrix {
			var m IntersectionMatrix
			m = m.with(imInterior, imInterior, imEntry2)
			m = m.with(imInterior, imBoundary, imEntry1)
			m = m.with(imInterior, imExterior, imEntry2)
			m = m.with(imBoundary, imInterior, imEntry1)
			m = m.with(imBoundary, imBoundary, imEntry0)
			m = m.with(imBoundary, imExterior, imEntry1)
			m = m.with(imExterior, imInterior, imEntry2)
			m = m.with(imExterior, imBoundary, imEntryF)
			m = m.with(imExterior, imExterior, imEntry2)
			return m
		}},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			m := tc.matrix()
			got := m.StringCode()
			if got != tc.code {
				t.Errorf("want=%v got=%v", tc.code, got)
			}
		})
	}
}

func TestMatrixFromStringCode(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		const code = "F01F200F1"
		m, err := IntersectionMatrixFromStringCode(code)
		if err != nil {
			t.Fatal(err)
		}
		if c := m.StringCode(); c != code {
			t.Errorf("unexpected StringCode(): %v", c)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		for i, code := range []string{
			"",           // Empty
			"F01F200F",   // 8 length
			"F01F200F10", // 10 length
			"F01F2*0F1",  // * is invalid in Matrix
			"F01F2X0F1",  // X is invalid in Matrix
		} {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				if _, err := IntersectionMatrixFromStringCode(code); err == nil {
					t.Errorf("expected an error but got nil")
				}
			})
		}
	})
}

func TestMatrixGet(t *testing.T) {
	m, err := IntersectionMatrixFromStringCode("2121012F2")
	if err != nil {
		t.Fatal(err)
	}

	checkGet := func(locA, locB imLocation, dim imEntry) {
		got := m.get(locA, locB)
		if got != dim {
			t.Errorf("%v %v want=%v got=%v", locA, locB, dim, got)
		}
	}
	checkGet(imInterior, imInterior, imEntry2)
	checkGet(imInterior, imBoundary, imEntry1)
	checkGet(imInterior, imExterior, imEntry2)
	checkGet(imBoundary, imInterior, imEntry1)
	checkGet(imBoundary, imBoundary, imEntry0)
	checkGet(imBoundary, imExterior, imEntry1)
	checkGet(imExterior, imInterior, imEntry2)
	checkGet(imExterior, imBoundary, imEntryF)
	checkGet(imExterior, imExterior, imEntry2)
}

func TestRelateMatch(t *testing.T) {
	for i, tc := range []struct {
		mat  string
		pat  string
		want bool
	}{
		{"FFFFFFFFF", "FFFFFFFFF", true},
		{"FFFFFFFFF", "000000000", false},
		{"FFFFFFFFF", "111111111", false},
		{"FFFFFFFFF", "222222222", false},
		{"FFFFFFFFF", "TTTTTTTTT", false},
		{"FFFFFFFFF", "*********", true},

		{"000000000", "FFFFFFFFF", false},
		{"000000000", "000000000", true},
		{"000000000", "111111111", false},
		{"000000000", "222222222", false},
		{"000000000", "TTTTTTTTT", true},
		{"000000000", "*********", true},

		{"111111111", "FFFFFFFFF", false},
		{"111111111", "000000000", false},
		{"111111111", "111111111", true},
		{"111111111", "222222222", false},
		{"111111111", "TTTTTTTTT", true},
		{"111111111", "*********", true},

		{"222222222", "FFFFFFFFF", false},
		{"222222222", "000000000", false},
		{"222222222", "111111111", false},
		{"222222222", "222222222", true},
		{"222222222", "TTTTTTTTT", true},
		{"222222222", "*********", true},

		{"F012F012F", "*********", true},
		{"F012F012F", "F*1**T*2*", true},
		{"F012F012F", "F*11*T*2*", false},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := RelateMatches(tc.mat, tc.pat)
			if err != nil {
				t.Error(err)
			}
			if got != tc.want {
				t.Logf("matrix:  %v", tc.mat)
				t.Logf("pattern: %v", tc.pat)
				t.Errorf("want=%t got=%t", tc.want, got)
			}
		})
	}
}

func TestRelateMatchError(t *testing.T) {
	for i, tc := range []struct {
		mat string
		pat string
	}{
		{"FFFFFFFF", "FFFFFFFFF"},
		{"FFFFFFFFF", "FFFFFFFF"},
		{"FFFFXFFFF", "FFFFFFFFF"},
		{"FFFFFFFFF", "FFFFXFFFF"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := RelateMatches(tc.mat, tc.pat)
			t.Log(err)
			if err == nil {
				t.Error("expected error but got nil")
			}
		})
	}
}
