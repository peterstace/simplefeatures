package exact_test

import (
	"fmt"
	"testing"

	. "github.com/peterstace/simplefeatures/internal/exact"
)

func TestSegmentIntersection(t *testing.T) {
	var (
		p00 = XY64{X: 0, Y: 0}
		p10 = XY64{X: 1, Y: 0}
		p01 = XY64{X: 0, Y: 1}
		p11 = XY64{X: 1, Y: 1}
		p21 = XY64{X: 2, Y: 1}
		p02 = XY64{X: 0, Y: 2}
		p22 = XY64{X: 2, Y: 2}
		p20 = XY64{X: 2, Y: 0}
		p40 = XY64{X: 4, Y: 0}
		p60 = XY64{X: 6, Y: 0}
		p42 = XY64{X: 4, Y: 2}
		p63 = XY64{X: 6, Y: 3}
	)
	ln := func(a, b XY64) Segment {
		return Segment{A: a, B: b}
	}

	e := func(xy XY64) XY64 {
		return xy
	}
	p := func(xy XY64) XY64 {
		return XY64{X: -xy.Y, Y: xy.X}
	}
	q := func(xy XY64) XY64 {
		return XY64{X: -xy.X, Y: xy.Y}
	}

	for _, tc := range []struct {
		description string
		segA, segB  Segment
		inter       Intersection
	}{
		{
			description: "parallel same length",
			segA:        ln(p01, p11),
			segB:        ln(p00, p10),
			inter:       Intersection{Empty: true},
		},
		{
			description: "parallel slanted",
			segA:        ln(p00, p21),
			segB:        ln(p01, p22),
			inter:       Intersection{Empty: true},
		},
		{
			description: "parallel different length",
			segA:        ln(p01, p21),
			segB:        ln(p00, p10),
			inter:       Intersection{Empty: true},
		},
		{
			description: "cross",
			segA:        ln(p00, p22),
			segB:        ln(p02, p20),
			inter:       Intersection{A: p11, B: p11},
		},
		{
			description: "would cross if longer",
			segA:        ln(p00, p21),
			segB:        ln(p02, p11),
			inter:       Intersection{Empty: true},
		},
		{
			description: "just touches in middle",
			segA:        ln(p00, p22),
			segB:        ln(p02, p11),
			inter:       Intersection{A: p11, B: p11},
		},
		{
			description: "touches at endpoints",
			segA:        ln(p00, p20),
			segB:        ln(p00, p02),
			inter:       Intersection{A: p00, B: p00},
		},

		// On same infinite sloping segment.
		{
			description: "on same sloping inf segment but no intersection",
			segA:        ln(p00, p21),
			segB:        ln(p42, p63),
			inter:       Intersection{Empty: true},
		},
		{
			description: "on same sloping inf segment touching",
			segA:        ln(p00, p21),
			segB:        ln(p21, p42),
			inter:       Intersection{A: p21, B: p21},
		},
		{
			description: "on same sloping inf segment staggered",
			segA:        ln(p00, p42),
			segB:        ln(p21, p63),
			inter:       Intersection{A: p21, B: p42},
		},
		{
			description: "on same sloping inf segment overlapping",
			segA:        ln(p00, p63),
			segB:        ln(p21, p42),
			inter:       Intersection{A: p21, B: p42},
		},
		{
			description: "on same sloping inf segment overlapping with one same endpoint",
			segA:        ln(p00, p42),
			segB:        ln(p21, p42),
			inter:       Intersection{A: p21, B: p42},
		},
		{
			description: "on same sloping inf segment overlapping with two same endpoints",
			segA:        ln(p00, p42),
			segB:        ln(p00, p42),
			inter:       Intersection{A: p00, B: p42},
		},

		// On same infinite horizontal segment.
		{
			description: "on same horizontal inf segment but no intersection",
			segA:        ln(p00, p20),
			segB:        ln(p40, p60),
			inter:       Intersection{Empty: true},
		},
		{
			description: "on same horizontal inf segment touching",
			segA:        ln(p00, p20),
			segB:        ln(p20, p40),
			inter:       Intersection{A: p20, B: p20},
		},
		{
			description: "on same horizontal inf segment staggered",
			segA:        ln(p00, p40),
			segB:        ln(p20, p60),
			inter:       Intersection{A: p20, B: p40},
		},
		{
			description: "on same horizontal inf segment overlapping",
			segA:        ln(p00, p60),
			segB:        ln(p20, p40),
			inter:       Intersection{A: p20, B: p40},
		},
		{
			description: "on same horizontal inf segment overlapping with one same endpoint",
			segA:        ln(p00, p40),
			segB:        ln(p20, p40),
			inter:       Intersection{A: p20, B: p40},
		},
		{
			description: "on same horizontal inf segment overlapping with two same endpoints",
			segA:        ln(p00, p40),
			segB:        ln(p00, p40),
			inter:       Intersection{A: p00, B: p40},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			for flip := 0; flip < 8; flip++ {
				for groupIdx, group := range [][]func(XY64) XY64{
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
						segA := tc.segA
						segB := tc.segB
						if flip&1 != 0 {
							segA, segB = segB, segA
						}
						if flip&2 != 0 {
							segA.A, segA.B = segA.B, segA.A
						}
						if flip&4 != 0 {
							segB.A, segB.B = segB.B, segB.A
						}

						want := tc.inter
						for _, fn := range group {
							want.A = fn(want.A)
							want.B = fn(want.B)
							segA.A = fn(segA.A)
							segA.B = fn(segA.B)
							segB.A = fn(segB.A)
							segB.B = fn(segB.B)
						}

						got := SegmentIntersection(segA, segB)
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
