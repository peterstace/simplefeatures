package geom

import "math"

// vertexGrid maps from truncated XY values to vertex records. A truncated XY
// value is an XY value where the last few bits in the XYs' mantissas have been
// set to zero. It's used as a way to merge together vertices that are very
// close together. It prevents vertexRecords from being added to adjacent cells
// in the grid, guaranteeing that there is at least 1 full grid cell between
// any 2 entries.
type vertexGrid map[XY]*vertexRecord

// lookup looks for a vertex at (or near) XY and returns it if it exists. It
// has the same semantics as a regular map.
func (g vertexGrid) lookup(xy XY) (*vertexRecord, bool) {
	tx := truncateMantissa(xy.X)
	ty := truncateMantissa(xy.Y)
	if v, ok := g[XY{tx, ty}]; ok {
		return v, true
	}

	txAdd1 := addULPs(tx, 0xff)
	txSub1 := addULPs(tx, -0xff)
	tyAdd1 := addULPs(ty, 0xff)
	tySub1 := addULPs(ty, -0xff)

	for _, adjacent := range [...]XY{
		{txAdd1, tyAdd1},
		{txAdd1, tySub1},
		{txAdd1, ty},
		{txSub1, tyAdd1},
		{txSub1, tySub1},
		{txSub1, ty},
		{ty, tyAdd1},
		{ty, tySub1},
	} {
		if v, ok := g[adjacent]; ok {
			return v, true
		}
	}
	return nil, false
}

// assign sets the vertex record to the grid at xy.
func (g vertexGrid) assign(xy XY, rec *vertexRecord) {
	tx := truncateMantissa(xy.X)
	ty := truncateMantissa(xy.Y)
	g[XY{tx, ty}] = rec
}

// truncateMantissa returns the input float64 but with some of the least
// significant bits of its mantissa set to zero.
func truncateMantissa(f float64) float64 {
	u := math.Float64bits(f)
	u &= ^uint64(0xff) // zero out last byte
	return math.Float64frombits(u)
}

// addULPs returns the float64 that is delta ULPs away from f. Delta may be negative.
func addULPs(f float64, delta int64) float64 {
	u := math.Float64bits(f)
	u += uint64(delta)
	return math.Float64frombits(u)
}
