package geom

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

// nativeOrder is the byte order of the host. It's safe to share because it's
// only assigned in init.
var nativeOrder binary.AppendByteOrder

func init() {
	var buf [2]byte
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0x1234)

	switch buf[0] {
	case 0x12:
		nativeOrder = binary.BigEndian
	case 0x34:
		nativeOrder = binary.LittleEndian
	default:
		panic(fmt.Sprintf("unexpected buf[0]: %d", buf[0]))
	}
}

// wkbHeaderBytes is the size of the byte order and geometry type fields that
// prefix every WKB geometry.
const wkbHeaderBytes = 1 + 4

// wkbCountBytes is the size of a uint32 count field.
const wkbCountBytes = 4

// wkbSequenceBytes is the number of bytes written by writeSequence for seq.
func wkbSequenceBytes(seq Sequence) int {
	return wkbCountBytes + 8*len(seq.floats)
}

type wkbMarshaler struct {
	buf []byte
}

func newWKBMarshaler(buf []byte) *wkbMarshaler {
	return &wkbMarshaler{buf}
}

func (m *wkbMarshaler) writeByteOrder() {
	if nativeOrder == binary.LittleEndian {
		m.buf = append(m.buf, 1)
	} else {
		m.buf = append(m.buf, 0)
	}
}

func (m *wkbMarshaler) writeGeomType(geomType GeometryType, ctype CoordinatesType) {
	gt := [...]uint32{7, 1, 2, 3, 4, 5, 6}[geomType]
	m.buf = nativeOrder.AppendUint32(m.buf, uint32(ctype)*1000+gt)
}

func (m *wkbMarshaler) writeFloat64(f float64) {
	m.buf = nativeOrder.AppendUint64(m.buf, math.Float64bits(f))
}

func (m *wkbMarshaler) writeCount(n int) {
	m.buf = nativeOrder.AppendUint32(m.buf, uint32(n))
}

func (m *wkbMarshaler) writeCoordinates(c Coordinates) {
	m.writeFloat64(c.X)
	m.writeFloat64(c.Y)
	if c.Type.Is3D() {
		m.writeFloat64(c.Z)
	}
	if c.Type.IsMeasured() {
		m.writeFloat64(c.M)
	}
}

func (m *wkbMarshaler) writeSequence(seq Sequence) {
	n := seq.Length()
	m.writeCount(n)

	// Rather than iterating over the sequence using the Get method, then
	// writing the Coordinates of the point using the writeCoordinates
	// function, we instead directly append the byte representation of the
	// floats. This relies on the assumption that the WKB being produced has
	// native byte order. This hack provides a *significant* performance
	// improvement.
	m.buf = append(m.buf, floatsAsBytes(seq.floats)...)
}

// floatsAsBytes reinterprets the floats slice as a bytes slice in a similar
// manner to reinterpret_cast in C++.
func floatsAsBytes(floats []float64) []byte {
	if len(floats) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&floats[0])), 8*len(floats))
}
