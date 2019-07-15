package simplefeatures

import (
	"encoding/binary"
	"fmt"
	"io"
)

// UnmarshalWKB reads the Well Known Binary (WKB), and returns the
// corresponding Geometry.
func UnmarshalWKB(r io.Reader) (Geometry, error) {
	p := wkbParser{r: r}
	p.parseByteOrder()
	p.parseGeomType()

	return nil, p.err
}

type wkbParser struct {
	r        io.Reader
	err      error
	bo       binary.ByteOrder
	coordLen int // 2 for XY, 3 for XYZ, 3 for XYM, 4 for XYZM
	geomType uint32
}

func (p *wkbParser) parseByteOrder() {
	if p.err != nil {
		return
	}
	var buf [1]byte
	if _, err := io.ReadFull(p.r, buf[:]); err != nil {
		p.err = err
		return
	}
	switch buf[0] {
	case 0:
		p.bo = binary.BigEndian
	case 1:
		p.bo = binary.LittleEndian
	default:
		p.err = fmt.Errorf("invalid byte order: %x", buf[0])
	}
}

func (p *wkbParser) parseGeomType() {
	if p.err != nil {
		return
	}
	if err := binary.Read(p.r, p.bo, &p.geomType); err != nil {
		p.err = err
		return
	}
}

type wkbGeomType struct {
	name     string
	coordLen int
}

var wkbGeomTypes = map[uint32]wkbGeomType{
	1: wkbGeomType{"Point", 2},
}
