package geom

import (
	"encoding/binary"
	"math"
)

type wkbMarshaller struct {
	buf []byte
}

var wkbBO = binary.LittleEndian

func newWKBMarshaller(buf []byte) *wkbMarshaller {
	return &wkbMarshaller{buf}
}

func (m *wkbMarshaller) writeByteOrder() {
	const littleEndian byte = 1
	m.buf = append(m.buf, littleEndian)
}

func (m *wkbMarshaller) writeGeomType(geomType GeometryType, ctype CoordinatesType) {
	gt := [...]uint32{7, 1, 2, 3, 4, 5, 6}[geomType]
	var buf [4]byte
	wkbBO.PutUint32(buf[:], uint32(ctype)*1000+gt)
	m.buf = append(m.buf, buf[:]...)
}

func (m *wkbMarshaller) writeFloat64(f float64) {
	var buf [8]byte
	wkbBO.PutUint64(buf[:], math.Float64bits(f))
	m.buf = append(m.buf, buf[:]...)
}

func (m *wkbMarshaller) writeCount(n int) {
	var buf [4]byte
	wkbBO.PutUint32(buf[:], uint32(n))
	m.buf = append(m.buf, buf[:]...)
}

func (m *wkbMarshaller) writeCoordinates(c Coordinates) {
	m.writeFloat64(c.X)
	m.writeFloat64(c.Y)
	if c.Type.Is3D() {
		m.writeFloat64(c.Z)
	}
	if c.Type.IsMeasured() {
		m.writeFloat64(c.M)
	}
}

func (m *wkbMarshaller) writeSequence(seq Sequence) {
	n := seq.Length()
	m.writeCount(n)

	// Iterating over the floats in the sequence and appending them directly
	// rather than using the Get method on the sequence provides a significant
	// performance improvement.
	for _, f := range seq.floats {
		var buf [8]byte
		wkbBO.PutUint64(buf[:], math.Float64bits(f))
		m.buf = append(m.buf, buf[:]...)
	}
}
