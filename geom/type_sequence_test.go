package geom_test

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestSequenceZeroValue(t *testing.T) {
	var seq geom.Sequence
	expectIntEq(t, seq.Length(), 0)
	expectCoordinatesTypeEq(t, seq.CoordinatesType(), geom.DimXY)
}

func TestSequenceCoordinatesType(t *testing.T) {
	for _, ct := range []geom.CoordinatesType{
		geom.DimXY,
		geom.DimXYZ,
		geom.DimXYM,
		geom.DimXYZM,
	} {
		t.Run(ct.String(), func(t *testing.T) {
			seq := geom.NewSequence(nil, ct)
			expectCoordinatesTypeEq(t, seq.CoordinatesType(), ct)
		})
	}
}

func checkSequence(t *testing.T, seq geom.Sequence, coords []geom.Coordinates) {
	t.Helper()
	gotLen := seq.Length()
	expectIntEq(t, gotLen, len(coords))
	expectPanics(t, func() { seq.Get(-1) })
	expectPanics(t, func() { seq.GetXY(-1) })
	for j, c := range coords {
		expectCoordsEq(t, c, seq.Get(j))
		expectXYEq(t, c.XY, seq.GetXY(j))
	}
	expectPanics(t, func() { seq.Get(len(coords)) })
	expectPanics(t, func() { seq.GetXY(len(coords)) })
}

func TestSequenceLengthAndGet(t *testing.T) {
	for i, tt := range []struct {
		seq    geom.Sequence
		coords []geom.Coordinates
		rev    []geom.Coordinates
	}{
		{
			geom.NewSequence(nil, geom.DimXY),
			nil,
			nil,
		},
		{
			geom.NewSequence(nil, geom.DimXYZ),
			nil,
			nil,
		},
		{
			geom.NewSequence(nil, geom.DimXYM),
			nil,
			nil,
		},
		{
			geom.NewSequence(nil, geom.DimXYZM),
			nil,
			nil,
		},

		{
			geom.NewSequence([]float64{1, 2}, geom.DimXY),
			[]geom.Coordinates{
				{Type: geom.DimXY, XY: geom.XY{X: 1, Y: 2}},
			},
			[]geom.Coordinates{
				{Type: geom.DimXY, XY: geom.XY{X: 1, Y: 2}},
			},
		},
		{
			geom.NewSequence([]float64{1, 2, 3}, geom.DimXYZ),
			[]geom.Coordinates{
				{Type: geom.DimXYZ, XY: geom.XY{X: 1, Y: 2}, Z: 3},
			},
			[]geom.Coordinates{
				{Type: geom.DimXYZ, XY: geom.XY{X: 1, Y: 2}, Z: 3},
			},
		},
		{
			geom.NewSequence([]float64{1, 2, 3}, geom.DimXYM),
			[]geom.Coordinates{
				{Type: geom.DimXYM, XY: geom.XY{X: 1, Y: 2}, M: 3},
			},
			[]geom.Coordinates{
				{Type: geom.DimXYM, XY: geom.XY{X: 1, Y: 2}, M: 3},
			},
		},
		{
			geom.NewSequence([]float64{1, 2, 3, 4}, geom.DimXYZM),
			[]geom.Coordinates{
				{Type: geom.DimXYZM, XY: geom.XY{X: 1, Y: 2}, Z: 3, M: 4},
			},
			[]geom.Coordinates{
				{Type: geom.DimXYZM, XY: geom.XY{X: 1, Y: 2}, Z: 3, M: 4},
			},
		},

		{
			geom.NewSequence([]float64{1, 2, 3, 4}, geom.DimXY),
			[]geom.Coordinates{
				{Type: geom.DimXY, XY: geom.XY{X: 1, Y: 2}},
				{Type: geom.DimXY, XY: geom.XY{X: 3, Y: 4}},
			},
			[]geom.Coordinates{
				{Type: geom.DimXY, XY: geom.XY{X: 3, Y: 4}},
				{Type: geom.DimXY, XY: geom.XY{X: 1, Y: 2}},
			},
		},
		{
			geom.NewSequence([]float64{1, 2, 3, 4, 5, 6}, geom.DimXYZ),
			[]geom.Coordinates{
				{Type: geom.DimXYZ, XY: geom.XY{X: 1, Y: 2}, Z: 3},
				{Type: geom.DimXYZ, XY: geom.XY{X: 4, Y: 5}, Z: 6},
			},
			[]geom.Coordinates{
				{Type: geom.DimXYZ, XY: geom.XY{X: 4, Y: 5}, Z: 6},
				{Type: geom.DimXYZ, XY: geom.XY{X: 1, Y: 2}, Z: 3},
			},
		},
		{
			geom.NewSequence([]float64{1, 2, 3, 4, 5, 6}, geom.DimXYM),
			[]geom.Coordinates{
				{Type: geom.DimXYM, XY: geom.XY{X: 1, Y: 2}, M: 3},
				{Type: geom.DimXYM, XY: geom.XY{X: 4, Y: 5}, M: 6},
			},
			[]geom.Coordinates{
				{Type: geom.DimXYM, XY: geom.XY{X: 4, Y: 5}, M: 6},
				{Type: geom.DimXYM, XY: geom.XY{X: 1, Y: 2}, M: 3},
			},
		},
		{
			geom.NewSequence([]float64{1, 2, 3, 4, 5, 6, 7, 8}, geom.DimXYZM),
			[]geom.Coordinates{
				{Type: geom.DimXYZM, XY: geom.XY{X: 1, Y: 2}, Z: 3, M: 4},
				{Type: geom.DimXYZM, XY: geom.XY{X: 5, Y: 6}, Z: 7, M: 8},
			},
			[]geom.Coordinates{
				{Type: geom.DimXYZM, XY: geom.XY{X: 5, Y: 6}, Z: 7, M: 8},
				{Type: geom.DimXYZM, XY: geom.XY{X: 1, Y: 2}, Z: 3, M: 4},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			checkSequence(t, tt.seq, tt.coords)
			checkSequence(t, tt.seq.Reverse(), tt.rev)

			for _, ct := range []geom.CoordinatesType{
				geom.DimXY,
				geom.DimXYZ,
				geom.DimXYM,
				geom.DimXYZM,
			} {
				t.Run(ct.String(), func(t *testing.T) {
					var wantCoords []geom.Coordinates
					for _, c := range tt.coords {
						c.Type = ct
						if !ct.Is3D() {
							c.Z = 0
						}
						if !ct.IsMeasured() {
							c.M = 0
						}
						wantCoords = append(wantCoords, c)
					}
					forced := tt.seq.ForceCoordinatesType(ct)
					checkSequence(t, forced, wantCoords)
					expectCoordinatesTypeEq(t, forced.CoordinatesType(), ct)
				})
			}
		})
	}
}

func TestSequencEnvelope(t *testing.T) {
	for i, tc := range []struct {
		floats []float64
		ct     geom.CoordinatesType
		want   geom.Envelope
	}{
		{nil, geom.DimXY, geom.Envelope{}},
		{[]float64{1, 2}, geom.DimXY, onePtEnv(1, 2)},
		{[]float64{3, 2, 1, 4}, geom.DimXY, twoPtEnv(1, 2, 3, 4)},
		{[]float64{3, 6, 1, 4, 5, 2}, geom.DimXY, twoPtEnv(1, 2, 5, 6)},
		{nil, geom.DimXYZ, geom.Envelope{}},
		{[]float64{1, 2, 3}, geom.DimXYZ, onePtEnv(1, 2)},
		{[]float64{3, 2, 0, 1, 4, 0}, geom.DimXYZ, twoPtEnv(1, 2, 3, 4)},
		{[]float64{3, 6, -1, 1, 4, -1, 5, 2, -1}, geom.DimXYZ, twoPtEnv(1, 2, 5, 6)},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			seq := geom.NewSequence(tc.floats, tc.ct)
			expectEnvEq(t, seq.Envelope(), tc.want)
		})
	}
}

func BenchmarkSequenceEnvelope(b *testing.B) {
	for _, sz := range []int{10, 100, 1000, 10_000} {
		b.Run(strconv.Itoa(sz), func(b *testing.B) {
			rnd := rand.New(rand.NewSource(1))
			flts := make([]float64, sz*2)
			for i := range flts {
				flts[i] = rnd.Float64()
			}
			seq := geom.NewSequence(flts, geom.DimXY)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				seq.Envelope()
			}
		})
	}
}
