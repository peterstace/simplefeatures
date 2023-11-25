package geom

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestLineLineIntersectionExtra(t *testing.T) {
	for i, tc := range []struct {
		ln1, ln2 line
		wantOK   bool
		wantXY   XY
	}{
		{
			line{XY{1, 2.1}, XY{2.1, 1}},
			line{XY{1.5, 1.5}, XY{8.5, 1.5}},
			true,
			XY{1.5999999999999999, 1.4999999999999998},
		},
		{
			line{XY{1, 0}, XY{0, 1}},
			line{XY{1, 1}, XY{2, 1}},
			false,
			XY{0, 1},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotXY, gotOK := lineLineIntersection(tc.ln1, tc.ln2)
			if gotOK != tc.wantOK {
				t.Errorf("gotOK(%t) != wantOK(%t)", gotOK, tc.wantOK)
			}
			if gotXY != tc.wantXY {
				t.Errorf("gotXY(%v) != wantXY(%v)", gotXY, tc.wantXY)
			}
		})
	}
}

func TestLineLineIntersection(t *testing.T) {
	for i := 0; i < 1000; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			rnd := rand.New(rand.NewSource(int64(i)))
			ln1 := line{
				XY{rnd.Float64(), rnd.Float64()},
				XY{rnd.Float64(), rnd.Float64()},
			}
			ln2 := line{
				XY{rnd.Float64(), rnd.Float64()},
				XY{rnd.Float64(), rnd.Float64()},
			}

			gotXY, gotOK := lineLineIntersection(ln1, ln2)

			t.Run("correctness", func(t *testing.T) {
				ls1 := ln1.asLineString().AsGeometry()
				ls2 := ln2.asLineString().AsGeometry()

				if gotOK {
					gotPt := gotXY.AsPoint().AsGeometry()
					dist1 := mustDistance(t, ls1, gotPt)
					dist2 := mustDistance(t, ls2, gotPt)
					if dist1 > 1e-6 || dist2 > 1e-6 {
						t.Fatalf("distance between line and intersection point "+
							"is %v and %v but got is %v", dist1, dist2, gotXY)
					}
				} else {
					dist := mustDistance(t, ls1, ls2)
					if dist < 1e-6 {
						t.Fatalf("distance between lines is %v but got is %v", dist, gotXY)
					}
				}
			})

			t.Run("symmetry", func(t *testing.T) {
				fLn1, fLn2 := ln1, ln2
				for flags := uint(0); flags < 8; flags++ {
					if (flags & 0b001) != 0 {
						fLn1.a, fLn1.b = fLn1.b, fLn1.a
					}
					if (flags & 0b010) != 0 {
						fLn2.a, fLn2.b = fLn2.b, fLn2.a
					}
					if (flags & 0b100) != 0 {
						fLn1, fLn2 = fLn2, fLn1
					}
					fGotXY, fGotOK := lineLineIntersection(fLn1, fLn2)

					if gotOK != fGotOK {
						t.Fatalf("gotOK(%t) != fGotOK(%t)", gotOK, fGotOK)
					}
					if gotXY != fGotXY {
						t.Fatalf("gotXY(%v) != fGotXY(%v)", gotXY, fGotXY)
					}
				}
			})
		})
	}
}

func mustDistance(t testing.TB, g1, g2 Geometry) float64 {
	dist, ok := Distance(g1, g2)
	if !ok {
		t.Fatal("!ok")
	}
	return dist
}
