package jts

import "math"

// ShapeFractal_HilbertCode_MAX_LEVEL is the maximum curve level that can be
// represented.
const ShapeFractal_HilbertCode_MAX_LEVEL = 16

// ShapeFractal_HilbertCode_Size returns the number of points in the curve for
// the given level. The number of points is 2^(2 * level).
func ShapeFractal_HilbertCode_Size(level int) int {
	shapeFractal_HilbertCode_checkLevel(level)
	return int(math.Pow(2, float64(2*level)))
}

// ShapeFractal_HilbertCode_MaxOrdinate returns the maximum ordinate value for
// points in the curve for the given level. The maximum ordinate is 2^level - 1.
func ShapeFractal_HilbertCode_MaxOrdinate(level int) int {
	shapeFractal_HilbertCode_checkLevel(level)
	return int(math.Pow(2, float64(level))) - 1
}

// ShapeFractal_HilbertCode_Level returns the level of the finite Hilbert curve
// which contains at least the given number of points.
func ShapeFractal_HilbertCode_Level(numPoints int) int {
	pow2 := int(math.Log(float64(numPoints)) / math.Log(2))
	level := pow2 / 2
	size := ShapeFractal_HilbertCode_Size(level)
	if size < numPoints {
		level++
	}
	return level
}

func shapeFractal_HilbertCode_checkLevel(level int) {
	if level > ShapeFractal_HilbertCode_MAX_LEVEL {
		panic("Level must be in range 0 to 16")
	}
}

// ShapeFractal_HilbertCode_Encode encodes a point (x,y) in the range of the
// Hilbert curve at a given level as the index of the point along the curve.
// The index will lie in the range [0, 2^(level + 1)].
func ShapeFractal_HilbertCode_Encode(level, x, y int) int {
	// Fast Hilbert curve algorithm by http://threadlocalmutex.com/
	// Ported from C++ https://github.com/rawrunprotected/hilbert_curves (public domain)

	lvl := shapeFractal_HilbertCode_levelClamp(level)

	x = x << (16 - lvl)
	y = y << (16 - lvl)

	a := int64(x ^ y)
	b := int64(0xFFFF) ^ a
	c := int64(0xFFFF) ^ int64(x|y)
	d := int64(x) & (int64(y) ^ 0xFFFF)

	A := a | (b >> 1)
	B := (a >> 1) ^ a
	C := ((c >> 1) ^ (b & (d >> 1))) ^ c
	D := ((a & (c >> 1)) ^ (d >> 1)) ^ d

	a = A
	b = B
	c = C
	d = D
	A = ((a & (a >> 2)) ^ (b & (b >> 2)))
	B = ((a & (b >> 2)) ^ (b & ((a ^ b) >> 2)))
	C ^= ((a & (c >> 2)) ^ (b & (d >> 2)))
	D ^= ((b & (c >> 2)) ^ ((a ^ b) & (d >> 2)))

	a = A
	b = B
	c = C
	d = D
	A = ((a & (a >> 4)) ^ (b & (b >> 4)))
	B = ((a & (b >> 4)) ^ (b & ((a ^ b) >> 4)))
	C ^= ((a & (c >> 4)) ^ (b & (d >> 4)))
	D ^= ((b & (c >> 4)) ^ ((a ^ b) & (d >> 4)))

	a = A
	b = B
	c = C
	d = D
	C ^= ((a & (c >> 8)) ^ (b & (d >> 8)))
	D ^= ((b & (c >> 8)) ^ ((a ^ b) & (d >> 8)))

	a = C ^ (C >> 1)
	b = D ^ (D >> 1)

	i0 := int64(x ^ y)
	i1 := b | (0xFFFF ^ (i0 | a))

	i0 = (i0 | (i0 << 8)) & 0x00FF00FF
	i0 = (i0 | (i0 << 4)) & 0x0F0F0F0F
	i0 = (i0 | (i0 << 2)) & 0x33333333
	i0 = (i0 | (i0 << 1)) & 0x55555555

	i1 = (i1 | (i1 << 8)) & 0x00FF00FF
	i1 = (i1 | (i1 << 4)) & 0x0F0F0F0F
	i1 = (i1 | (i1 << 2)) & 0x33333333
	i1 = (i1 | (i1 << 1)) & 0x55555555

	index := ((i1 << 1) | i0) >> (32 - 2*lvl)
	return int(index)
}

// shapeFractal_HilbertCode_levelClamp clamps a level to the range valid for
// the index algorithm used.
func shapeFractal_HilbertCode_levelClamp(level int) int {
	// clamp order to [1, 16]
	lvl := level
	if lvl < 1 {
		lvl = 1
	}
	if lvl > ShapeFractal_HilbertCode_MAX_LEVEL {
		lvl = ShapeFractal_HilbertCode_MAX_LEVEL
	}
	return lvl
}

// ShapeFractal_HilbertCode_Decode computes the point on a Hilbert curve of
// given level for a given code index. The point ordinates will lie in the
// range [0, 2^level - 1].
func ShapeFractal_HilbertCode_Decode(level, index int) *Geom_Coordinate {
	shapeFractal_HilbertCode_checkLevel(level)
	lvl := shapeFractal_HilbertCode_levelClamp(level)

	index = index << (32 - 2*lvl)

	i0 := shapeFractal_HilbertCode_deinterleave(index)
	i1 := shapeFractal_HilbertCode_deinterleave(index >> 1)

	t0 := (i0 | i1) ^ 0xFFFF
	t1 := i0 & i1

	prefixT0 := shapeFractal_HilbertCode_prefixScan(t0)
	prefixT1 := shapeFractal_HilbertCode_prefixScan(t1)

	a := (((i0 ^ 0xFFFF) & prefixT1) | (i0 & prefixT0))

	x := (a ^ i1) >> (16 - lvl)
	y := (a ^ i0 ^ i1) >> (16 - lvl)

	return Geom_NewCoordinateWithXY(float64(x), float64(y))
}

func shapeFractal_HilbertCode_prefixScan(x int64) int64 {
	x = (x >> 8) ^ x
	x = (x >> 4) ^ x
	x = (x >> 2) ^ x
	x = (x >> 1) ^ x
	return x
}

func shapeFractal_HilbertCode_deinterleave(x int) int64 {
	x = x & 0x55555555
	x = (x | (x >> 1)) & 0x33333333
	x = (x | (x >> 2)) & 0x0F0F0F0F
	x = (x | (x >> 4)) & 0x00FF00FF
	x = (x | (x >> 8)) & 0x0000FFFF
	return int64(x)
}
