package geom_test

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
	"github.com/peterstace/simplefeatures/rtree"
)

func TestEnvelopeNew(t *testing.T) {
	for _, tc := range []struct {
		desc string
		xys  []geom.XY
		want geom.Envelope
	}{
		{
			desc: "nil slice",
			xys:  nil,
			want: geom.Envelope{},
		},
		{
			desc: "empty slice",
			xys:  []geom.XY{},
			want: geom.Envelope{},
		},
		{
			desc: "single element",
			xys:  []geom.XY{{1, 2}},
			want: geom.NewEnvelopeXY(1, 2),
		},
		{
			desc: "two same elements",
			xys:  []geom.XY{{1, 2}, {1, 2}},
			want: geom.NewEnvelopeXY(1, 2),
		},
		{
			desc: "two different elements",
			xys:  []geom.XY{{1, 2}, {-1, 3}},
			want: geom.NewEnvelopeXY(-1, 2, 1, 3),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got := geom.NewEnvelope(tc.xys...)
			test.Eq(t, got, tc.want)
		})
	}
}

func TestEnvelopeAttributes(t *testing.T) {
	for _, tc := range []struct {
		description                      string
		env                              geom.Envelope
		isEmpty, isPoint, isLine, isRect bool
		area, width, height              float64
		center, min, max, geom           string
	}{
		{
			description: "empty",
			env:         geom.Envelope{},
			isEmpty:     true,
			isPoint:     false,
			isLine:      false,
			isRect:      false,
			area:        0,
			width:       0,
			height:      0,
			center:      "POINT EMPTY",
			min:         "POINT EMPTY",
			max:         "POINT EMPTY",
			geom:        "GEOMETRYCOLLECTION EMPTY",
		},
		{
			description: "single point",
			env:         geom.NewEnvelopeXY(1, 2),
			isEmpty:     false,
			isPoint:     true,
			isLine:      false,
			isRect:      false,
			area:        0,
			width:       0,
			height:      0,
			center:      "POINT(1 2)",
			min:         "POINT(1 2)",
			max:         "POINT(1 2)",
			geom:        "POINT(1 2)",
		},
		{
			description: "two horizontal points",
			env:         geom.NewEnvelopeXY(1, 4, 3, 4),
			isEmpty:     false,
			isPoint:     false,
			isLine:      true,
			isRect:      false,
			area:        0,
			width:       2,
			height:      0,
			center:      "POINT(2 4)",
			min:         "POINT(1 4)",
			max:         "POINT(3 4)",
			geom:        "LINESTRING(1 4,3 4)",
		},
		{
			description: "two vertical points",
			env:         geom.NewEnvelopeXY(4, 1, 4, 3),
			isEmpty:     false,
			isPoint:     false,
			isLine:      true,
			isRect:      false,
			area:        0,
			width:       0,
			height:      2,
			center:      "POINT(4 2)",
			min:         "POINT(4 1)",
			max:         "POINT(4 3)",
			geom:        "LINESTRING(4 1,4 3)",
		},
		{
			description: "two diagonal points",
			env:         geom.NewEnvelopeXY(1, 4, 3, 7),
			isEmpty:     false,
			isPoint:     false,
			isLine:      false,
			isRect:      true,
			area:        6,
			width:       2,
			height:      3,
			center:      "POINT(2 5.5)",
			min:         "POINT(1 4)",
			max:         "POINT(3 7)",
			geom:        "POLYGON((1 4,3 4,3 7,1 7,1 4))",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			t.Run("IsEmpty", func(t *testing.T) {
				test.Eq(t, tc.env.IsEmpty(), tc.isEmpty)
			})
			t.Run("IsPoint", func(t *testing.T) {
				test.Eq(t, tc.env.IsPoint(), tc.isPoint)
			})
			t.Run("IsLine", func(t *testing.T) {
				test.Eq(t, tc.env.IsLine(), tc.isLine)
			})
			t.Run("IsRectangle", func(t *testing.T) {
				test.Eq(t, tc.env.IsRectangle(), tc.isRect)
			})
			t.Run("Area", func(t *testing.T) {
				test.Eq(t, tc.env.Area(), tc.area)
			})
			t.Run("Width", func(t *testing.T) {
				test.Eq(t, tc.env.Width(), tc.width)
			})
			t.Run("Height", func(t *testing.T) {
				test.Eq(t, tc.env.Height(), tc.height)
			})
			t.Run("Center", func(t *testing.T) {
				test.ExactEqualsWKT(t, tc.env.Center().AsGeometry(), tc.center)
			})
			t.Run("Min", func(t *testing.T) {
				test.ExactEqualsWKT(t, tc.env.Min().AsGeometry(), tc.min)
			})
			t.Run("Max", func(t *testing.T) {
				test.ExactEqualsWKT(t, tc.env.Max().AsGeometry(), tc.max)
			})
			t.Run("MinMaxXYs", func(t *testing.T) {
				gotMin, gotMax, gotOK := tc.env.MinMaxXYs()
				test.Eq(t, gotOK, !tc.isEmpty)
				if gotOK {
					wantMin, minOK := test.FromWKT(t, tc.min).MustAsPoint().XY()
					test.True(t, minOK)
					test.Eq(t, gotMin, wantMin)

					wantMax, maxOK := test.FromWKT(t, tc.max).MustAsPoint().XY()
					test.True(t, maxOK)
					test.Eq(t, gotMax, wantMax)
				}
			})
			t.Run("AsGeometry", func(t *testing.T) {
				test.ExactEqualsWKT(t, tc.env.AsGeometry(), tc.geom, geom.IgnoreOrder)
			})
		})
	}
}

func TestEnvelopeExpandToIncludeXY(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		env := geom.Envelope{}.ExpandToIncludeXY(geom.XY{1, 2})
		test.ExactEqualsWKT(t, env.Min().AsGeometry(), "POINT(1 2)")
		test.ExactEqualsWKT(t, env.Max().AsGeometry(), "POINT(1 2)")
	})
	t.Run("single point extend to same", func(t *testing.T) {
		env := geom.NewEnvelopeXY(1, 2).ExpandToIncludeXY(geom.XY{1, 2})
		test.ExactEqualsWKT(t, env.Min().AsGeometry(), "POINT(1 2)")
		test.ExactEqualsWKT(t, env.Max().AsGeometry(), "POINT(1 2)")
	})
	t.Run("single point extend to different", func(t *testing.T) {
		env := geom.NewEnvelopeXY(1, 2).ExpandToIncludeXY(geom.XY{-1, 3})
		test.ExactEqualsWKT(t, env.Min().AsGeometry(), "POINT(-1 2)")
		test.ExactEqualsWKT(t, env.Max().AsGeometry(), "POINT(1 3)")
	})
	t.Run("area extend within", func(t *testing.T) {
		env := geom.NewEnvelopeXY(1, 2, 3, 4).ExpandToIncludeXY(geom.XY{2, 3})
		test.ExactEqualsWKT(t, env.Min().AsGeometry(), "POINT(1 2)")
		test.ExactEqualsWKT(t, env.Max().AsGeometry(), "POINT(3 4)")
	})
	t.Run("area extend outside", func(t *testing.T) {
		env := geom.NewEnvelopeXY(1, 2, 3, 4).ExpandToIncludeXY(geom.XY{100, 200})
		test.ExactEqualsWKT(t, env.Min().AsGeometry(), "POINT(1 2)")
		test.ExactEqualsWKT(t, env.Max().AsGeometry(), "POINT(100 200)")
	})
}

func TestEnvelopeContains(t *testing.T) {
	for _, tc := range []struct {
		env      geom.Envelope
		subtests map[geom.XY]bool
	}{
		{
			env: geom.Envelope{},
			subtests: map[geom.XY]bool{
				{}:     false,
				{1, 2}: false,
			},
		},
		{
			env: geom.NewEnvelopeXY(1, 2),
			subtests: map[geom.XY]bool{
				{}:     false,
				{1, 2}: true,
				{3, 1}: false,
			},
		},
		{
			env: geom.NewEnvelopeXY(1, 2, 4, 5),
			subtests: func() map[geom.XY]bool {
				m := map[geom.XY]bool{}
				for x := 0; x <= 5; x++ {
					for y := 1; y <= 6; y++ {
						m[geom.XY{float64(x), float64(y)}] = x >= 1 && x <= 4 && y >= 2 && y <= 5
					}
				}
				return m
			}(),
		},
	} {
		t.Run(fmt.Sprintf("env %v", tc.env.AsGeometry().AsText()), func(t *testing.T) {
			for xy, want := range tc.subtests {
				t.Run(fmt.Sprintf("xy %v want %v", xy, want), func(t *testing.T) {
					got := tc.env.Contains(xy)
					test.Eq(t, got, want)
				})
			}
		})
	}
}

func TestEnvelopeExpandToIncludeEnvelope(t *testing.T) {
	for _, tc := range []struct {
		desc   string
		e1, e2 geom.Envelope
		want   geom.Envelope
	}{
		{
			desc: "empty and empty",
			e1:   geom.Envelope{},
			e2:   geom.Envelope{},
			want: geom.Envelope{},
		},
		{
			desc: "point and empty",
			e1:   geom.NewEnvelopeXY(1, 2),
			e2:   geom.Envelope{},
			want: geom.NewEnvelopeXY(1, 2),
		},
		{
			desc: "rect and empty",
			e1:   geom.NewEnvelopeXY(1, 1, 2, 2),
			e2:   geom.Envelope{},
			want: geom.NewEnvelopeXY(1, 1, 2, 2),
		},
		{
			desc: "same point",
			e1:   geom.NewEnvelopeXY(1, 2),
			e2:   geom.NewEnvelopeXY(1, 2),
			want: geom.NewEnvelopeXY(1, 2),
		},
		{
			desc: "same rect",
			e1:   geom.NewEnvelopeXY(1, 1, 2, 2),
			e2:   geom.NewEnvelopeXY(1, 1, 2, 2),
			want: geom.NewEnvelopeXY(1, 1, 2, 2),
		},
		{
			desc: "point and point",
			e1:   geom.NewEnvelopeXY(1, 2),
			e2:   geom.NewEnvelopeXY(-1, 3),
			want: geom.NewEnvelopeXY(-1, 2, 1, 3),
		},
		{
			desc: "point and rect",
			e1:   geom.NewEnvelopeXY(1, 1, 2, 2),
			e2:   geom.NewEnvelopeXY(3, 1),
			want: geom.NewEnvelopeXY(1, 1, 3, 2),
		},
		{
			desc: "rect inside other",
			e1:   geom.NewEnvelopeXY(1, 11, 4, 14),
			e2:   geom.NewEnvelopeXY(2, 12, 3, 13),
			want: geom.NewEnvelopeXY(1, 11, 4, 14),
		},
		{
			desc: "rect overlapping corner",
			e1:   geom.NewEnvelopeXY(1, 11, 3, 13),
			e2:   geom.NewEnvelopeXY(2, 12, 4, 14),
			want: geom.NewEnvelopeXY(1, 11, 4, 14),
		},
	} {
		t.Run(tc.desc+" fwd", func(t *testing.T) {
			got := tc.e1.ExpandToIncludeEnvelope(tc.e2)
			test.Eq(t, got, tc.want)
		})
		t.Run(tc.desc+" rev", func(t *testing.T) {
			got := tc.e2.ExpandToIncludeEnvelope(tc.e1)
			test.Eq(t, got, tc.want)
		})
	}
}

func TestEnvelopeInvalidXYInteractions(t *testing.T) {
	var (
		nan = math.NaN()
		inf = math.Inf(+1)
	)
	for i, tc := range []struct {
		violation geom.RuleViolation
		xy        geom.XY
	}{
		{geom.ViolateNaN, geom.XY{0, nan}},
		{geom.ViolateNaN, geom.XY{nan, 0}},
		{geom.ViolateNaN, geom.XY{nan, nan}},
		{geom.ViolateInf, geom.XY{0, inf}},
		{geom.ViolateInf, geom.XY{inf, 0}},
		{geom.ViolateInf, geom.XY{inf, inf}},
		{geom.ViolateInf, geom.XY{0, -inf}},
		{geom.ViolateInf, geom.XY{-inf, 0}},
		{geom.ViolateInf, geom.XY{-inf, -inf}},
	} {
		t.Run(fmt.Sprintf("new_envelope_with_first_arg_invalid_%d", i), func(t *testing.T) {
			env := geom.NewEnvelope(tc.xy)
			expectRuleViolation(t, env, tc.violation)
		})
		t.Run(fmt.Sprintf("new_envelope_with_second_arg_invalid_%d", i), func(t *testing.T) {
			env := geom.NewEnvelope(geom.XY{}, tc.xy)
			expectRuleViolation(t, env, tc.violation)
		})
		t.Run(fmt.Sprintf("expand_to_include_invalid_xy_%d", i), func(t *testing.T) {
			env := geom.NewEnvelope(geom.XY{-1, -1}, geom.XY{1, 1})
			env = env.ExpandToIncludeXY(tc.xy)
			expectRuleViolation(t, env, tc.violation)
		})
		t.Run(fmt.Sprintf("expand_from_invalid_to_include_env_%d", i), func(t *testing.T) {
			env := geom.NewEnvelope(tc.xy)
			env = env.ExpandToIncludeXY(geom.XY{1, 1})
			expectRuleViolation(t, env, tc.violation)
		})
		t.Run(fmt.Sprintf("expand_to_include_invalid_env_%d", i), func(t *testing.T) {
			base := geom.NewEnvelope(geom.XY{-1, -1}, geom.XY{1, 1})
			other := geom.NewEnvelope(tc.xy)
			env := base.ExpandToIncludeEnvelope(other)
			expectRuleViolation(t, env, tc.violation)
		})
		t.Run(fmt.Sprintf("expand_from_invalid_to_include_env_%d", i), func(t *testing.T) {
			base := geom.NewEnvelope(tc.xy)
			other := geom.NewEnvelope(geom.XY{-1, -1}, geom.XY{1, 1})
			env := base.ExpandToIncludeEnvelope(other)
			expectRuleViolation(t, env, tc.violation)
		})
		t.Run(fmt.Sprintf("contains_invalid_xy_%d", i), func(t *testing.T) {
			env := geom.NewEnvelope(geom.XY{-1, -1}, geom.XY{1, 1})
			test.False(t, env.Contains(tc.xy))
		})
	}
}

func TestEnvelopeIntersects(t *testing.T) {
	for i, tt := range []struct {
		e1, e2 geom.Envelope
		want   bool
	}{
		// Empty vs empty.
		{geom.Envelope{}, geom.Envelope{}, false},

		// Empty vs non-empty.
		{geom.Envelope{}, geom.NewEnvelopeXY(0, 0), false},
		{geom.Envelope{}, geom.NewEnvelopeXY(0, 0, 1, 1), false},

		// Single pt vs single pt.
		{geom.NewEnvelopeXY(0, 0), geom.NewEnvelopeXY(0, 0), true},
		{geom.NewEnvelopeXY(1, 2), geom.NewEnvelopeXY(1, 2), true},
		{geom.NewEnvelopeXY(1, 2), geom.NewEnvelopeXY(1, 3), false},
		{geom.NewEnvelopeXY(1, 2), geom.NewEnvelopeXY(2, 2), false},

		// Single pt vs rect.
		{geom.NewEnvelopeXY(0, 0), geom.NewEnvelopeXY(0, 0, 1, 1), true},
		{geom.NewEnvelopeXY(1, 1), geom.NewEnvelopeXY(0, 0, 1, 1), true},
		{geom.NewEnvelopeXY(0, 1), geom.NewEnvelopeXY(0, 0, 1, 1), true},
		{geom.NewEnvelopeXY(1, 0), geom.NewEnvelopeXY(0, 0, 1, 1), true},
		{geom.NewEnvelopeXY(0.5, 0.5), geom.NewEnvelopeXY(0, 0, 1, 1), true},
		{geom.NewEnvelopeXY(0.5, 1.5), geom.NewEnvelopeXY(0, 0, 1, 1), false},

		// Rect vs Rect.
		{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(2, 2, 3, 3), false},
		{geom.NewEnvelopeXY(0, 2, 1, 3), geom.NewEnvelopeXY(2, 0, 3, 1), false},
		{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(1, 1, 2, 2), true},
		{geom.NewEnvelopeXY(0, 1, 1, 2), geom.NewEnvelopeXY(1, 0, 2, 1), true},
		{geom.NewEnvelopeXY(0, 0, 2, 2), geom.NewEnvelopeXY(1, 1, 3, 3), true},
		{geom.NewEnvelopeXY(0, 1, 2, 3), geom.NewEnvelopeXY(1, 0, 3, 2), true},
		{geom.NewEnvelopeXY(0, 0, 2, 1), geom.NewEnvelopeXY(1, 0, 3, 1), true},
		{geom.NewEnvelopeXY(0, 0, 1, 2), geom.NewEnvelopeXY(0, 1, 1, 3), true},
		{geom.NewEnvelopeXY(0, 0, 2, 2), geom.NewEnvelopeXY(1, -1, 3, 3), true},
		{geom.NewEnvelopeXY(0, 0, 2, 2), geom.NewEnvelopeXY(1, -1, 3, 3), true},
		{geom.NewEnvelopeXY(-1, 0, 2, 1), geom.NewEnvelopeXY(0, -1, 1, 2), true},
		{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(-1, -1, 2, 2), true},
		{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(1, 0, 2, 1), true},
		{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(0, 1, 1, 2), true},
		{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(2, 0, 3, 1), false},
		{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(0, 2, 1, 3), false},
		{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(2, -1, 3, 2), false},
		{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(-1, -2, 2, -1), false},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got1 := tt.e1.Intersects(tt.e2)
			got2 := tt.e2.Intersects(tt.e1)
			if got1 != tt.want || got2 != tt.want {
				t.Logf("env1: %v", tt.e1)
				t.Logf("env2: %v", tt.e2)
				t.Errorf("want=%v got1=%v", tt.want, got1)
				t.Errorf("want=%v got2=%v", tt.want, got2)
			}
		})
	}
}

func TestEnvelopeCovers(t *testing.T) {
	for i, tt := range []struct {
		env1, env2 geom.Envelope
		want       bool
	}{
		// Empty vs empty.
		{geom.Envelope{}, geom.Envelope{}, false},

		// Empty vs single pt.
		{geom.Envelope{}, geom.NewEnvelopeXY(1, 2), false},
		{geom.NewEnvelopeXY(1, 2), geom.Envelope{}, false},
		{geom.Envelope{}, geom.NewEnvelopeXY(0, 0), false},
		{geom.NewEnvelopeXY(0, 0), geom.Envelope{}, false},

		// Empty vs rect.
		{geom.Envelope{}, geom.NewEnvelopeXY(1, 2, 3, 4), false},
		{geom.NewEnvelopeXY(1, 2, 3, 4), geom.Envelope{}, false},

		// Single pt vs single pt.
		{geom.NewEnvelopeXY(1, 2), geom.NewEnvelopeXY(1, 2), true},
		{geom.NewEnvelopeXY(1, 2), geom.NewEnvelopeXY(3, 2), false},
		{geom.NewEnvelopeXY(1, 2), geom.NewEnvelopeXY(1, 3), false},
		{geom.NewEnvelopeXY(1, 2), geom.NewEnvelopeXY(3, 3), false},

		// Single pt vs single rect.
		{geom.NewEnvelopeXY(1, 2), geom.NewEnvelopeXY(1, 2, 3, 4), false},
		{geom.NewEnvelopeXY(1, 2), geom.NewEnvelopeXY(0, 0, 3, 3), false},
		{geom.NewEnvelopeXY(0, 0, 3, 3), geom.NewEnvelopeXY(1, 2), true},
		{geom.NewEnvelopeXY(0, 0, 3, 3), geom.NewEnvelopeXY(0, 0), true},
		{geom.NewEnvelopeXY(0, 0, 3, 3), geom.NewEnvelopeXY(3, 3), true},
		{geom.NewEnvelopeXY(0, 0, 3, 3), geom.NewEnvelopeXY(0, 3), true},
		{geom.NewEnvelopeXY(0, 0, 3, 3), geom.NewEnvelopeXY(3, 4), false},
		{geom.NewEnvelopeXY(0, 0, 3, 3), geom.NewEnvelopeXY(4, 3), false},

		// Rect vs Rect
		{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(2, 0, 3, 1), false},
		{geom.NewEnvelopeXY(0, 0, 2, 2), geom.NewEnvelopeXY(1, 1, 3, 3), false},
		{geom.NewEnvelopeXY(0, 0, 3, 3), geom.NewEnvelopeXY(1, 1, 2, 2), true},
		{geom.NewEnvelopeXY(0, 0, 2, 2), geom.NewEnvelopeXY(1, 1, 2, 2), true},
		{geom.NewEnvelopeXY(1, 1, 2, 2), geom.NewEnvelopeXY(0, 0, 3, 3), false},
		{geom.NewEnvelopeXY(1, 1, 2, 2), geom.NewEnvelopeXY(0, 0, 2, 2), false},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := tt.env1.Covers(tt.env2)
			if got != tt.want {
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestEnvelopeDistance(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		t.Run("both", func(t *testing.T) {
			_, ok := geom.Envelope{}.Distance(geom.Envelope{})
			test.False(t, ok)
		})
		t.Run("only one", func(t *testing.T) {
			_, ok := geom.Envelope{}.Distance(geom.NewEnvelopeXY(1, 2))
			test.False(t, ok)
			_, ok = geom.NewEnvelopeXY(1, 2).Distance(geom.Envelope{})
			test.False(t, ok)
		})
	})
	t.Run("non-empty", func(t *testing.T) {
		for i, tt := range []struct {
			env1, env2 geom.Envelope
			want       float64
		}{
			// Pt vs pt.
			{geom.NewEnvelopeXY(3, 0), geom.NewEnvelopeXY(4, 0), 1},
			{geom.NewEnvelopeXY(3, 0), geom.NewEnvelopeXY(3, 1), 1},
			{geom.NewEnvelopeXY(3, 0), geom.NewEnvelopeXY(4, 1), math.Sqrt(2)},

			// Pt vs rect.
			{geom.NewEnvelopeXY(2, 1), geom.NewEnvelopeXY(1, 2, 3, 4), 1},
			{geom.NewEnvelopeXY(2, 1), geom.NewEnvelopeXY(2, 2, 3, 3), 1},
			{geom.NewEnvelopeXY(2, 1), geom.NewEnvelopeXY(3, 2, 4, 3), math.Sqrt(2)},

			// Rect vs rect.
			{geom.NewEnvelopeXY(0, 0, 2, 2), geom.NewEnvelopeXY(1, 1, 3, 3), 0},
			{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(2, 0, 2, 1), 1},
			{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(0, 3, 1, 4), 2},
			{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(2, 2, 3, 3), math.Sqrt(2)},
			{geom.NewEnvelopeXY(0, 2, 1, 3), geom.NewEnvelopeXY(2, 0, 3, 1), math.Sqrt(2)},
			{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(1, 1, 2, 2), 0},
			{geom.NewEnvelopeXY(0, 1, 1, 2), geom.NewEnvelopeXY(1, 0, 2, 1), 0},
			{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(1, 0, 2, 1), 0},
			{geom.NewEnvelopeXY(0, 0, 1, 1), geom.NewEnvelopeXY(0, 1, 1, 2), 0},
		} {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				got1, ok1 := tt.env1.Distance(tt.env2)
				got2, ok2 := tt.env2.Distance(tt.env1)
				test.True(t, ok1)
				test.True(t, ok2)
				test.Eq(t, got1, got2)
				test.Eq(t, got1, tt.want)
			})
		}
	})
}

func TestEnvelopeTransformXY(t *testing.T) {
	transform := func(in geom.XY) geom.XY {
		return geom.XY{in.X * 1.5, in.Y * 2.5}
	}
	for i, tc := range []struct {
		input geom.Envelope
		want  geom.Envelope
	}{
		{geom.Envelope{}, geom.Envelope{}},
		{geom.NewEnvelopeXY(1, 2), geom.NewEnvelopeXY(1.5, 5)},
		{geom.NewEnvelopeXY(1, 2, 3, 4), geom.NewEnvelopeXY(1.5, 5, 4.5, 10)},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := tc.input.TransformXY(transform)
			test.Eq(t, got, tc.want)
		})
	}
}

func TestEnvelopeTransformBugFix(t *testing.T) {
	// Reproduces a bug where a transform that alters which coordinates are min
	// and max causes a malformed envelope.
	env := geom.NewEnvelopeXY(1, 2, 3, 4)
	got := env.TransformXY(func(in geom.XY) geom.XY {
		return geom.XY{-in.X, -in.Y}
	})
	test.Eq(t, got, geom.NewEnvelopeXY(-3, -4, -1, -2))
}

func BenchmarkEnvelopeTransformXY(b *testing.B) {
	input := geom.NewEnvelopeXY(1, 2, 3, 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input.TransformXY(func(in geom.XY) geom.XY {
			return geom.XY{in.X * 1.5, in.Y * 2.5}
		})
	}
}

func TestBoundingDiagonal(t *testing.T) {
	for i, tc := range []struct {
		env  []geom.XY
		want string
	}{
		{
			nil,
			"GEOMETRYCOLLECTION EMPTY",
		},
		{
			[]geom.XY{{1, 2}},
			"POINT(1 2)",
		},
		{
			[]geom.XY{{3, 2}, {1, 4}},
			"LINESTRING(1 2,3 4)",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			env := geom.NewEnvelope(tc.env...)
			got := env.BoundingDiagonal()
			test.ExactEqualsWKT(t, got, tc.want)
		})
	}
}

func TestEnvelopeEmptyAsBox(t *testing.T) {
	_, ok := geom.Envelope{}.AsBox()
	test.False(t, ok)
}

func TestEnvelopeNonEmptyAsBox(t *testing.T) {
	got, ok := geom.NewEnvelopeXY(1, 2, 3, 4).AsBox()
	test.True(t, ok)
	want := rtree.Box{MinX: 1, MinY: 2, MaxX: 3, MaxY: 4}
	test.True(t, got == want)
}

func TestEnvelopeString(t *testing.T) {
	for i, tc := range []struct {
		env  geom.Envelope
		want string
	}{
		{geom.Envelope{}, "ENVELOPE EMPTY"},
		{geom.NewEnvelopeXY(1.5, 2.5, 3.5, 4.5), "ENVELOPE(1.5 2.5,3.5 4.5)"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := fmt.Sprintf("%v", tc.env)
			t.Log(got)
			test.True(t, got == tc.want)
		})
	}
}
