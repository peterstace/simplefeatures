package geom

// Tiny Well Known Binary
// See spec https://github.com/TWKB/Specification/blob/master/twkb.md

type twkbGeometryType int

const (
	twkbTypePoint              twkbGeometryType = 1
	twkbTypeLineString         twkbGeometryType = 2
	twkbTypePolygon            twkbGeometryType = 3
	twkbTypeMultiPoint         twkbGeometryType = 4
	twkbTypeMultiLineString    twkbGeometryType = 5
	twkbTypeMultiPolygon       twkbGeometryType = 6
	twkbTypeGeometryCollection twkbGeometryType = 7
)

const (
	twkbMaxDimensions = 4
)

type twkbMetadataHeader int

const (
	twkbHasBBox    twkbMetadataHeader = 1
	twkbHasSize    twkbMetadataHeader = 2
	twkbHasIDs     twkbMetadataHeader = 4
	twkbHasExtPrec twkbMetadataHeader = 8
	twkbIsEmpty    twkbMetadataHeader = 16
)

// decodeZigZagInt64 accepts a uint64 and reverses the zigzag encoding
// to produce the decoded signed int64 value.
func decodeZigZagInt64(z uint64) int64 {
	return int64(z>>1) ^ -int64(z&1)
}

// encodeZigZagInt64 accepts a signed int64 and zigzag encodes
// it to produce an encoded uint64 value.
func encodeZigZagInt64(n int64) uint64 {
	return uint64((n << 1) ^ (n >> 63))
}
