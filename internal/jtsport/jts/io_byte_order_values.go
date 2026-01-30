package jts

import "math"

// Byte order constants.
const (
	Io_ByteOrderValues_BIG_ENDIAN    = 1
	Io_ByteOrderValues_LITTLE_ENDIAN = 2
)

// Io_ByteOrderValues_GetInt reads an int32 value from the buffer using the
// specified byte order.
func Io_ByteOrderValues_GetInt(buf []byte, byteOrder int) int32 {
	if byteOrder == Io_ByteOrderValues_BIG_ENDIAN {
		return int32(buf[0]&0xff)<<24 |
			int32(buf[1]&0xff)<<16 |
			int32(buf[2]&0xff)<<8 |
			int32(buf[3]&0xff)
	}
	// LITTLE_ENDIAN
	return int32(buf[3]&0xff)<<24 |
		int32(buf[2]&0xff)<<16 |
		int32(buf[1]&0xff)<<8 |
		int32(buf[0]&0xff)
}

// Io_ByteOrderValues_PutInt writes an int32 value to the buffer using the
// specified byte order.
func Io_ByteOrderValues_PutInt(intValue int32, buf []byte, byteOrder int) {
	if byteOrder == Io_ByteOrderValues_BIG_ENDIAN {
		buf[0] = byte(intValue >> 24)
		buf[1] = byte(intValue >> 16)
		buf[2] = byte(intValue >> 8)
		buf[3] = byte(intValue)
	} else {
		// LITTLE_ENDIAN
		buf[0] = byte(intValue)
		buf[1] = byte(intValue >> 8)
		buf[2] = byte(intValue >> 16)
		buf[3] = byte(intValue >> 24)
	}
}

// Io_ByteOrderValues_GetLong reads an int64 value from the buffer using the
// specified byte order.
func Io_ByteOrderValues_GetLong(buf []byte, byteOrder int) int64 {
	if byteOrder == Io_ByteOrderValues_BIG_ENDIAN {
		return int64(buf[0]&0xff)<<56 |
			int64(buf[1]&0xff)<<48 |
			int64(buf[2]&0xff)<<40 |
			int64(buf[3]&0xff)<<32 |
			int64(buf[4]&0xff)<<24 |
			int64(buf[5]&0xff)<<16 |
			int64(buf[6]&0xff)<<8 |
			int64(buf[7]&0xff)
	}
	// LITTLE_ENDIAN
	return int64(buf[7]&0xff)<<56 |
		int64(buf[6]&0xff)<<48 |
		int64(buf[5]&0xff)<<40 |
		int64(buf[4]&0xff)<<32 |
		int64(buf[3]&0xff)<<24 |
		int64(buf[2]&0xff)<<16 |
		int64(buf[1]&0xff)<<8 |
		int64(buf[0]&0xff)
}

// Io_ByteOrderValues_PutLong writes an int64 value to the buffer using the
// specified byte order.
func Io_ByteOrderValues_PutLong(longValue int64, buf []byte, byteOrder int) {
	if byteOrder == Io_ByteOrderValues_BIG_ENDIAN {
		buf[0] = byte(longValue >> 56)
		buf[1] = byte(longValue >> 48)
		buf[2] = byte(longValue >> 40)
		buf[3] = byte(longValue >> 32)
		buf[4] = byte(longValue >> 24)
		buf[5] = byte(longValue >> 16)
		buf[6] = byte(longValue >> 8)
		buf[7] = byte(longValue)
	} else {
		// LITTLE_ENDIAN
		buf[0] = byte(longValue)
		buf[1] = byte(longValue >> 8)
		buf[2] = byte(longValue >> 16)
		buf[3] = byte(longValue >> 24)
		buf[4] = byte(longValue >> 32)
		buf[5] = byte(longValue >> 40)
		buf[6] = byte(longValue >> 48)
		buf[7] = byte(longValue >> 56)
	}
}

// Io_ByteOrderValues_GetDouble reads a float64 value from the buffer using
// the specified byte order.
func Io_ByteOrderValues_GetDouble(buf []byte, byteOrder int) float64 {
	longVal := Io_ByteOrderValues_GetLong(buf, byteOrder)
	return math.Float64frombits(uint64(longVal))
}

// Io_ByteOrderValues_PutDouble writes a float64 value to the buffer using the
// specified byte order.
func Io_ByteOrderValues_PutDouble(doubleValue float64, buf []byte, byteOrder int) {
	longVal := int64(math.Float64bits(doubleValue))
	Io_ByteOrderValues_PutLong(longVal, buf, byteOrder)
}
