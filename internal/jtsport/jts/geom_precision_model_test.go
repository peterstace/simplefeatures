package jts

import "testing"

func TestPrecisionModelParameterlessConstructor(t *testing.T) {
	p := Geom_NewPrecisionModel()
	// Implicit precision model has scale 0.
	if got := p.GetScale(); got != 0 {
		t.Errorf("GetScale() = %v, want 0", got)
	}
}

func TestPrecisionModelGetMaximumSignificantDigits(t *testing.T) {
	tests := []struct {
		name string
		pm   *Geom_PrecisionModel
		want int
	}{
		{
			name: "Floating",
			pm:   Geom_NewPrecisionModelWithType(Geom_PrecisionModel_Floating),
			want: 16,
		},
		{
			name: "FloatingSingle",
			pm:   Geom_NewPrecisionModelWithType(Geom_PrecisionModel_FloatingSingle),
			want: 6,
		},
		{
			name: "Fixed default",
			pm:   Geom_NewPrecisionModelWithType(Geom_PrecisionModel_Fixed),
			want: 1,
		},
		{
			name: "Fixed scale 1000",
			pm:   Geom_NewPrecisionModelWithScale(1000),
			want: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pm.GetMaximumSignificantDigits(); got != tt.want {
				t.Errorf("GetMaximumSignificantDigits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrecisionModelMakePrecise(t *testing.T) {
	pm10 := Geom_NewPrecisionModelWithScale(0.1)

	tests := []struct {
		x1, y1 float64
		x2, y2 float64
	}{
		{1200.4, 1240.4, 1200, 1240},
		{1209.4, 1240.4, 1210, 1240},
	}

	for _, tt := range tests {
		p := Geom_NewCoordinateWithXY(tt.x1, tt.y1)
		pm10.MakePreciseCoordinate(p)
		pPrecise := Geom_NewCoordinateWithXY(tt.x2, tt.y2)
		if !p.Equals2D(pPrecise) {
			t.Errorf("MakePreciseCoordinate(%v, %v) = (%v, %v), want (%v, %v)",
				tt.x1, tt.y1, p.X, p.Y, tt.x2, tt.y2)
		}
	}
}
