package geom

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

var nativeOrder binary.ByteOrder

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

type wkbMarshaller struct {
	buf []byte
}

func newWKBMarshaller(buf []byte) *wkbMarshaller {
	return &wkbMarshaller{buf}
}

func (m *wkbMarshaller) writeByteOrder() {
	if nativeOrder == binary.LittleEndian {
		m.buf = append(m.buf, 1)
	} else {
		m.buf = append(m.buf, 0)
	}
}

func (m *wkbMarshaller) writeGeomType(geomType GeometryType, ctype CoordinatesType) {
	gt := [...]uint32{7, 1, 2, 3, 4, 5, 6}[geomType]
	var buf [4]byte
	nativeOrder.PutUint32(buf[:], uint32(ctype)*1000+gt)
	m.buf = append(m.buf, buf[:]...)
}

func (m *wkbMarshaller) writeFloat64(f float64) {
	var buf [8]byte
	nativeOrder.PutUint64(buf[:], math.Float64bits(f))
	m.buf = append(m.buf, buf[:]...)
}

func (m *wkbMarshaller) writeCount(n int) {
	var buf [4]byte
	nativeOrder.PutUint32(buf[:], uint32(n))
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
	var byts []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&byts))
	hdr.Data = (*reflect.SliceHeader)(unsafe.Pointer(&floats)).Data
	hdr.Len = 8 * len(floats)
	hdr.Cap = 8 * cap(floats)
	return byts
}
