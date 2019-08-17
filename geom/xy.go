package geom

type XY struct {
	X, Y float64
}

func (w XY) Equals(o XY) bool {
	return w.X == o.X && w.Y == o.Y
}

func (w XY) Sub(o XY) XY {
	return XY{
		w.X - o.X,
		w.Y - o.Y,
	}
}

func (w XY) Add(o XY) XY {
	return XY{
		w.X + o.X,
		w.Y + o.Y,
	}
}

func (w XY) Scale(s float64) XY {
	return XY{
		w.X * s,
		w.Y * s,
	}
}

func (w XY) Cross(o XY) float64 {
	return (w.X * o.Y) - (w.Y * o.X)
}

func (w XY) Midpoint(o XY) XY {
	return w.Add(o).Scale(0.5)
}

func (w XY) Dot(o XY) float64 {
	return w.X*o.X + w.Y*o.Y
}

// Less gives an ordering on XYs. If two XYs have different X values, then the
// one with the lower X value is ordered before the one with the higher X
// value. If the X values are then same, then the Y values are used (the lower
// Y value comes first).
func (w XY) Less(o XY) bool {
	if w.X != o.X {
		return w.X < o.X
	}
	return w.Y < o.Y
}
