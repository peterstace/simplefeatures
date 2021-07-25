package exact_test

import (
	"fmt"
	"testing"

	. "github.com/peterstace/simplefeatures/internal/exact"
)

func TestLineIntersection(t *testing.T) {
	var (
		p00 = XY{X: 0, Y: 0}
		p10 = XY{X: 1, Y: 0}
		p01 = XY{X: 0, Y: 1}
		p11 = XY{X: 1, Y: 1}
		p21 = XY{X: 2, Y: 1}
		p02 = XY{X: 0, Y: 2}
		p22 = XY{X: 2, Y: 2}
		p20 = XY{X: 2, Y: 0}
		p40 = XY{X: 4, Y: 0}
		p60 = XY{X: 6, Y: 0}
		p42 = XY{X: 4, Y: 2}
		p63 = XY{X: 6, Y: 3}
	)
	ln := func(a, b XY) Line {
		return Line{A: a, B: b}
	}

	e := func(xy XY) XY {
		return xy
	}
	p := func(xy XY) XY {
		return XY{X: -xy.Y, Y: xy.X}
	}
	q := func(xy XY) XY {
		return XY{X: -xy.X, Y: xy.Y}
	}

	for _, tc := range []struct {
		description  string
		lineA, lineB Line
		inter        Intersection
	}{
		{
			description: "parallel same length",
			lineA:       ln(p01, p11),
			lineB:       ln(p00, p10),
			inter:       Intersection{Empty: true},
		},
		{
			description: "parallel slanted",
			lineA:       ln(p00, p21),
			lineB:       ln(p01, p22),
			inter:       Intersection{Empty: true},
		},
		{
			description: "parallel different length",
			lineA:       ln(p01, p21),
			lineB:       ln(p00, p10),
			inter:       Intersection{Empty: true},
		},
		{
			description: "cross",
			lineA:       ln(p00, p22),
			lineB:       ln(p02, p20),
			inter:       Intersection{A: p11, B: p11},
		},
		{
			description: "would cross if longer",
			lineA:       ln(p00, p21),
			lineB:       ln(p02, p11),
			inter:       Intersection{Empty: true},
		},
		{
			description: "just touches in middle",
			lineA:       ln(p00, p22),
			lineB:       ln(p02, p11),
			inter:       Intersection{A: p11, B: p11},
		},
		{
			description: "touches at endpoints",
			lineA:       ln(p00, p20),
			lineB:       ln(p00, p02),
			inter:       Intersection{A: p00, B: p00},
		},

		// On same infinite sloping line.
		{
			description: "on same sloping inf line but no intersection",
			lineA:       ln(p00, p21),
			lineB:       ln(p42, p63),
			inter:       Intersection{Empty: true},
		},
		{
			description: "on same sloping inf line touching",
			lineA:       ln(p00, p21),
			lineB:       ln(p21, p42),
			inter:       Intersection{A: p21, B: p21},
		},
		{
			description: "on same sloping inf line staggered",
			lineA:       ln(p00, p42),
			lineB:       ln(p21, p63),
			inter:       Intersection{A: p21, B: p42},
		},
		{
			description: "on same sloping inf line overlapping",
			lineA:       ln(p00, p63),
			lineB:       ln(p21, p42),
			inter:       Intersection{A: p21, B: p42},
		},
		{
			description: "on same sloping inf line overlapping with one same endpoint",
			lineA:       ln(p00, p42),
			lineB:       ln(p21, p42),
			inter:       Intersection{A: p21, B: p42},
		},
		{
			description: "on same sloping inf line overlapping with two same endpoints",
			lineA:       ln(p00, p42),
			lineB:       ln(p00, p42),
			inter:       Intersection{A: p00, B: p42},
		},

		// On same infinite horizontal line.
		{
			description: "on same horizontal inf line but no intersection",
			lineA:       ln(p00, p20),
			lineB:       ln(p40, p60),
			inter:       Intersection{Empty: true},
		},
		{
			description: "on same horizontal inf line touching",
			lineA:       ln(p00, p20),
			lineB:       ln(p20, p40),
			inter:       Intersection{A: p20, B: p20},
		},
		{
			description: "on same horizontal inf line staggered",
			lineA:       ln(p00, p40),
			lineB:       ln(p20, p60),
			inter:       Intersection{A: p20, B: p40},
		},
		{
			description: "on same horizontal inf line overlapping",
			lineA:       ln(p00, p60),
			lineB:       ln(p20, p40),
			inter:       Intersection{A: p20, B: p40},
		},
		{
			description: "on same horizontal inf line overlapping with one same endpoint",
			lineA:       ln(p00, p40),
			lineB:       ln(p20, p40),
			inter:       Intersection{A: p20, B: p40},
		},
		{
			description: "on same horizontal inf line overlapping with two same endpoints",
			lineA:       ln(p00, p40),
			lineB:       ln(p00, p40),
			inter:       Intersection{A: p00, B: p40},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			for flip := 0; flip < 8; flip++ {
				for groupIdx, group := range [][]func(XY) XY{
					{e},
					{p},
					{p, p},
					{p, p, p},
					{q},
					{q, p},
					{q, p, p},
					{q, p, p, p},
				} {
					t.Run(fmt.Sprintf("flip_%d_group_%d", flip, groupIdx), func(t *testing.T) {
						lineA := tc.lineA
						lineB := tc.lineB
						if flip&1 != 0 {
							lineA, lineB = lineB, lineA
						}
						if flip&2 != 0 {
							lineA.A, lineA.B = lineA.B, lineA.A
						}
						if flip&4 != 0 {
							lineB.A, lineB.B = lineB.B, lineB.A
						}

						want := tc.inter
						for _, fn := range group {
							want.A = fn(want.A)
							want.B = fn(want.B)
							lineA.A = fn(lineA.A)
							lineA.B = fn(lineA.B)
							lineB.A = fn(lineB.A)
							lineB.B = fn(lineB.B)
						}

						got := LineIntersection(lineA, lineB)
						if want.Empty != got.Empty {
							t.Fatalf("got_empty:%v want_empty:%v", got.Empty, want.Empty)
						}
						if !got.Empty {
							if (want.A != got.A || want.B != got.B) && (want.A != got.B || want.B != got.A) {
								t.Fatalf("got:%v want:%v", got, want)
							}
						}
					})
				}
			}
		})
	}
}
