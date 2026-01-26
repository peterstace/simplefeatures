package jts

// WKB byte order constants.
const (
	Io_WKBConstants_wkbXDR = 0 // Big Endian
	Io_WKBConstants_wkbNDR = 1 // Little Endian
)

// WKB geometry type constants.
const (
	Io_WKBConstants_wkbPoint              = 1
	Io_WKBConstants_wkbLineString         = 2
	Io_WKBConstants_wkbPolygon            = 3
	Io_WKBConstants_wkbMultiPoint         = 4
	Io_WKBConstants_wkbMultiLineString    = 5
	Io_WKBConstants_wkbMultiPolygon       = 6
	Io_WKBConstants_wkbGeometryCollection = 7
)
