package de9im_test

import (
	"strconv"
	"testing"

	. "github.com/peterstace/simplefeatures/de9im"
)

func TestMatrixZeroValue(t *testing.T) {
	var m Matrix
	const want = "FFFFFFFFF"
	got := m.StringCode()
	if got != want {
		t.Errorf("want=%v got=%v", want, got)
	}
}

func TestMatrixWith(t *testing.T) {
	for i, tc := range []struct {
		code   string
		matrix func() Matrix
	}{
		{"FFFFFFFFF", func() Matrix {
			var m Matrix
			return m
		}},
		{"2FFFFFFFF", func() Matrix {
			var m Matrix
			m = m.With(Interior, Interior, Dim2)
			return m
		}},
		{"F1FFFFFFF", func() Matrix {
			var m Matrix
			m = m.With(Interior, Boundary, Dim1)
			return m
		}},
		{"2121012F2", func() Matrix {
			var m Matrix
			m = m.With(Interior, Interior, Dim2)
			m = m.With(Interior, Boundary, Dim1)
			m = m.With(Interior, Exterior, Dim2)
			m = m.With(Boundary, Interior, Dim1)
			m = m.With(Boundary, Boundary, Dim0)
			m = m.With(Boundary, Exterior, Dim1)
			m = m.With(Exterior, Interior, Dim2)
			m = m.With(Exterior, Boundary, Empty)
			m = m.With(Exterior, Exterior, Dim2)
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
		m, err := MatrixFromStringCode(code)
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
				if _, err := MatrixFromStringCode(code); err == nil {
					t.Errorf("expected an error but got nil")
				}
			})
		}
	})
}

func TestMatrixGet(t *testing.T) {
	m, err := MatrixFromStringCode("2121012F2")
	if err != nil {
		t.Fatal(err)
	}

	checkGet := func(locA, locB Location, dim Dimension) {
		got := m.Get(locA, locB)
		if got != dim {
			t.Errorf("%v %v want=%v got=%v", locA, locB, dim, got)
		}
	}
	checkGet(Interior, Interior, Dim2)
	checkGet(Interior, Boundary, Dim1)
	checkGet(Interior, Exterior, Dim2)
	checkGet(Boundary, Interior, Dim1)
	checkGet(Boundary, Boundary, Dim0)
	checkGet(Boundary, Exterior, Dim1)
	checkGet(Exterior, Interior, Dim2)
	checkGet(Exterior, Boundary, Empty)
	checkGet(Exterior, Exterior, Dim2)
}

func TestMinMaxDimension(t *testing.T) {
	for i, tc := range []struct {
		inputA, inputB Dimension
	}{
		{Empty, Empty},
		{Empty, Dim0},
		{Empty, Dim1},
		{Empty, Dim2},
		{Dim0, Dim0},
		{Dim0, Dim1},
		{Dim0, Dim2},
		{Dim1, Dim1},
		{Dim1, Dim2},
		{Dim2, Dim2},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for _, swap := range []struct {
				name string
				flag bool
			}{
				{"fwd", false},
				{"rev", true},
			} {
				for _, minmax := range []struct {
					name string
					want Dimension
					op   func(Dimension, Dimension) Dimension
				}{
					{"min", tc.inputA, MinDimension},
					{"max", tc.inputB, MaxDimension},
				} {
					t.Run(minmax.name, func(t *testing.T) {
						t.Run(swap.name, func(t *testing.T) {
							if swap.flag {
								tc.inputA, tc.inputB = tc.inputA, tc.inputB
							}
							got := minmax.op(tc.inputA, tc.inputB)
							if got != minmax.want {
								t.Errorf("%v %v want=%v got=%v",
									tc.inputA, tc.inputB, minmax.want, got)
							}
						})
					})
				}
			}
		})
	}
}
