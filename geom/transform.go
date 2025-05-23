package geom

func transformSequence(seq Sequence, fn func(XY) XY) Sequence {
	floats := make([]float64, 0, seq.CoordinatesType().Dimension()*seq.Length())
	n := seq.Length()
	ctype := seq.CoordinatesType()
	for i := 0; i < n; i++ {
		c := seq.Get(i)
		c.XY = fn(c.XY)
		switch ctype {
		case DimXY:
			floats = append(floats, c.X, c.Y)
		case DimXYZ:
			floats = append(floats, c.X, c.Y, c.Z)
		case DimXYM:
			floats = append(floats, c.X, c.Y, c.M)
		case DimXYZM:
			floats = append(floats, c.X, c.Y, c.Z, c.M)
		default:
			panic(ctype)
		}
	}
	return NewSequence(floats, ctype)
}

func transformSequenceAllAtOnce(seq Sequence, fn func(CoordinatesType, []float64) error) (Sequence, error) {
	clone := clone1DFloat64s(seq.floats)
	ct := seq.CoordinatesType()
	err := fn(ct, clone)
	if err != nil {
		return Sequence{}, err
	}
	return NewSequence(clone, ct), nil
}
