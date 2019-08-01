package simplefeatures

import (
	"encoding/binary"
	"io"
)

type wkbMarshaller struct {
	w   io.Writer
	err error
}

func newWKBMarshaller(w io.Writer) *wkbMarshaller {
	return &wkbMarshaller{w: w}
}

func (m *wkbMarshaller) setErr(err error) {
	if m.err == nil {
		m.err = err
	}
}

func (m *wkbMarshaller) write(data interface{}) {
	if m.err != nil {
		return
	}
	// Output byte order is an arbitrary choice (either is allowed by the
	// WKB spec). Little endian is chosen because this is the same as
	// Postgres.
	err := binary.Write(m.w, binary.LittleEndian, data)
	m.setErr(err)
}

func (m *wkbMarshaller) writeByteOrder() {
	const littleEndian byte = 1
	m.write(littleEndian)
}

func (m *wkbMarshaller) writeGeomType(geomType uint32) {
	m.write(geomType)
}

func (m *wkbMarshaller) writeFloat64(f float64) {
	m.write(f)
}

func (m *wkbMarshaller) writeCount(n int) {
	m.write(uint32(n))
}
