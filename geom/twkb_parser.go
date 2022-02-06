package geom

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Tiny Well Known Binary
// See spec https://github.com/TWKB/Specification/blob/master/twkb.md

const (
	twkbTypePoint              = 1
	twkbTypeLineString         = 2
	twkbTypePolygon            = 3
	twkbTypeMultiPoint         = 4
	twkbTypeMultiLineString    = 5
	twkbTypeMultiPolygon       = 6
	twkbTypeGeometryCollection = 7
)

const (
	twkbMaxDimensions = 4
)

// UnmarshalTWKB parses a Tiny Well Known Binary (TWKB), returning the
// corresponding Geometry.
func UnmarshalTWKB(twkb []byte, opts ...ConstructorOption) (Geometry, error) {
	p := newTWKBParser(twkb, opts...)
	return p.nextGeometry()
}

// TWKBParser holds all state information for interpreting TWKB buffers
// including information such as the last reference point used in coord deltas.
type TWKBParser struct {
	twkb []byte
	pos  int
	opts []ConstructorOption

	kind  int
	ctype CoordinatesType

	dimensions int
	precX      int
	precY      int
	hasZ       bool
	hasM       bool
	precZ      int
	precM      int
	scalings   [twkbMaxDimensions]float64

	hasBBox bool
	hasSize bool
	hasIDs  bool
	hasExt  bool
	isEmpty bool

	bbox []int64

	refpoint [twkbMaxDimensions]int64
}

func newTWKBParser(twkb []byte, opts ...ConstructorOption) TWKBParser {
	return TWKBParser{
		twkb:       twkb,
		opts:       opts,
		ctype:      DimXY,
		dimensions: 2,
	}
}

// Parse a geometry and return it, the number of bytes consumed, and any error.
func (p *TWKBParser) parseGeometry() (Geometry, int, error) {
	geom, err := p.nextGeometry()
	return geom, p.pos, err
}

// Parse a geometry and return it and any error.
func (p *TWKBParser) nextGeometry() (Geometry, error) {
	if err := p.parseTypeAndPrecision(); err != nil {
		return Geometry{}, err
	}
	if err := p.parseMetadataHeader(); err != nil {
		return Geometry{}, err
	}
	if p.hasExt {
		if err := p.parseExtendedPrecision(); err != nil {
			return Geometry{}, err
		}
	}
	if p.hasSize {
		if err := p.parseSize(); err != nil {
			return Geometry{}, err
		}
	}
	if p.hasBBox {
		if err := p.parseBBox(); err != nil {
			return Geometry{}, err
		}
	}

	switch p.kind {
	case twkbTypePoint:
		if pt, err := p.parsePoint(); err != nil {
			return Geometry{}, err
		} else {
			return pt.AsGeometry(), nil
		}
	case twkbTypeLineString:
		if ls, err := p.parseLineString(); err != nil {
			return Geometry{}, err
		} else {
			return ls.AsGeometry(), nil
		}
	case twkbTypePolygon:
		if poly, err := p.parsePolygon(); err != nil {
			return Geometry{}, err
		} else {
			return poly.AsGeometry(), nil
		}
	case twkbTypeMultiPoint:
		if mp, err := p.parseMultiPoint(); err != nil {
			return Geometry{}, err
		} else {
			return mp.AsGeometry(), nil
		}
	case twkbTypeMultiLineString:
		if ml, err := p.parseMultiLineString(); err != nil {
			return Geometry{}, err
		} else {
			return ml.AsGeometry(), nil
		}
	case twkbTypeMultiPolygon:
		if mp, err := p.parseMultiPolygon(); err != nil {
			return Geometry{}, err
		} else {
			return mp.AsGeometry(), nil
		}
	case twkbTypeGeometryCollection:
		if gc, err := p.parseGeometryCollection(); err != nil {
			return Geometry{}, err
		} else {
			return gc.AsGeometry(), nil
		}
	}

	return Geometry{}, p.parserError("cannot unmarshal unsupported geometry type")
}

func (p *TWKBParser) parseTypeAndPrecision() error {
	if len(p.twkb) <= p.pos {
		return p.parserError("cannot unmarshal empty buffer")
	}
	typeprec := p.twkb[p.pos]
	p.pos++

	p.kind = int(typeprec & 0x0f)
	prec := int(DecodeZigZagInt32(uint32(typeprec) >> 4))
	p.precX = prec
	p.precY = prec

	p.scalings[0] = math.Pow10(-p.precX)
	p.scalings[1] = math.Pow10(-p.precY)
	return nil
}

func (p *TWKBParser) parseMetadataHeader() error {
	if len(p.twkb) <= p.pos {
		return p.parserError("cannot unmarshal lacking metadata header")
	}
	metaheader := p.twkb[p.pos]
	p.pos++

	p.hasBBox = (metaheader & 1) != 0
	p.hasSize = (metaheader & 2) != 0
	p.hasIDs = (metaheader & 4) != 0
	p.hasExt = (metaheader & 8) != 0
	p.isEmpty = (metaheader & 16) != 0
	return nil
}

func (p *TWKBParser) parseExtendedPrecision() error {
	if len(p.twkb) <= p.pos {
		return p.parserError("cannot unmarshal lacking extended pecision byte")
	}
	extprec := p.twkb[p.pos]
	p.pos++

	p.ctype = DimXY
	if extprec&1 != 0 {
		p.hasZ = true
		p.precZ = int(DecodeZigZagInt32(uint32(extprec>>2) & 0x07))
	}
	if extprec&2 != 0 {
		p.hasM = true
		p.precM = int(DecodeZigZagInt32(uint32(extprec>>5) & 0x07))
	}
	if p.hasZ && p.hasM {
		p.ctype = DimXYZM
		p.dimensions = 4
		p.scalings[2] = math.Pow10(-p.precZ)
		p.scalings[3] = math.Pow10(-p.precM)
	} else if p.hasZ {
		p.ctype = DimXYZ
		p.dimensions = 3
		p.scalings[2] = math.Pow10(-p.precZ)
	} else if p.hasM {
		p.ctype = DimXYM
		p.dimensions = 3
		p.scalings[2] = math.Pow10(-p.precM)
	}
	return nil
}

func (p *TWKBParser) parseSize() error {
	if len(p.twkb) <= p.pos {
		return p.parserError("cannot unmarshal lacking size varint")
	}
	bytesRemaining, err := p.parseUnsignedVarint()
	if err != nil {
		return p.parserError("cannot unmarshal size varint malformed")
	}
	if uint64(p.pos)+bytesRemaining > uint64(len(p.twkb)) {
		return p.parserError("cannot unmarshal remaining input smaller than size varint indicates")
	}
	return nil
}

func (p *TWKBParser) parseBBox() error {
	// Parse the bounding box, but do not otherwise use it.
	for i := 0; i < p.dimensions; i++ {
		minVal, err := p.parseSignedVarint()
		if err != nil {
			return p.parserError("cannot unmarshal BBox min varint malformed")
		}
		p.bbox = append(p.bbox, minVal)

		deltaVal, err := p.parseSignedVarint()
		if err != nil {
			return p.parserError("cannot unmarshal BBox delta varint malformed")
		}
		p.bbox = append(p.bbox, deltaVal)
	}
	return nil
}

func (p *TWKBParser) parsePoint() (Point, error) {
	if p.isEmpty {
		return NewEmptyPoint(p.ctype), nil
	}
	return p.nextPoint()
}

func (p *TWKBParser) nextPoint() (Point, error) {
	var c Coordinates
	c.Type = p.ctype

	coords, err := p.parsePointArray(1)
	if err != nil {
		return Point{}, err
	}

	c.XY.X = coords[0]
	c.XY.Y = coords[1]
	if p.hasZ && p.hasM {
		c.Z = coords[2]
		c.M = coords[3]
	} else if p.hasZ {
		c.Z = coords[2]
	} else if p.hasM {
		c.M = coords[2]
	}

	pt, err := NewPoint(c, p.opts...)
	return pt, p.annotateError(err)
}

func (p *TWKBParser) parseLineString() (LineString, error) {
	if p.isEmpty {
		return NewLineString(NewSequence(nil, p.ctype), p.opts...)
	}
	return p.nextLineString()
}

func (p *TWKBParser) nextLineString() (LineString, error) {
	coords, err := p.parsePointCountAndArray()
	if err != nil {
		return LineString{}, err
	}
	ls, err := NewLineString(NewSequence(coords, p.ctype), p.opts...)
	return ls, p.annotateError(err)
}

func (p *TWKBParser) parsePolygon() (Polygon, error) {
	if p.isEmpty {
		return NewPolygon(nil, p.opts...)
	}
	return p.nextPolygon()
}

func (p *TWKBParser) nextPolygon() (Polygon, error) {
	numRings, err := p.parseUnsignedVarint()
	if err != nil {
		return Polygon{}, p.parserError("cannot unmarshal num rings varint malformed")
	}

	var rings []LineString
	for r := uint64(0); r < numRings; r++ {
		coords, err := p.parsePointCountAndArray()
		if err != nil {
			return Polygon{}, err
		}
		if len(coords) > 0 {
			// Append first point to close the ring, as per the spec.
			// "rings are assumed to be implicitly closed, so the first and
			//  last point should not be the same"
			//
			// Note, the spec is unclear on what to do if the data indicates
			// the ring is already closed, as seen in one of the example tests:
			// See "03031b000400040205000004000004030000030500000002020000010100"
			// from https://github.com/TWKB/twkb.js/blob/master/test/twkb.spec.js
			// which closes each ring with a 5th point at the point (0,0).
			// (Note the inner ring also touches the outer ring at that point,
			// which violates polygon geometry constraints.)
			//
			// Currently, closed rings in the input aren't handled;
			// they seems to violate the spec.
			for d := 0; d < p.dimensions; d++ {
				coords = append(coords, coords[0+d])
			}
		}
		ls, err := NewLineString(NewSequence(coords, p.ctype), p.opts...)
		if err != nil {
			return Polygon{}, p.annotateError(err)
		}
		rings = append(rings, ls)
	}
	poly, err := NewPolygon(rings, p.opts...)
	return poly, p.annotateError(err)
}

func (p *TWKBParser) parseMultiPoint() (MultiPoint, error) {
	if p.isEmpty {
		return NewMultiPoint(nil, p.opts...), nil
	}
	return p.nextMultiPoint()
}

func (p *TWKBParser) nextMultiPoint() (MultiPoint, error) {
	numPoints, err := p.parseUnsignedVarint()
	if err != nil {
		return MultiPoint{}, p.parserError("cannot unmarshal num points varint malformed")
	}
	_, err = p.parseIDList(int(numPoints))
	if err != nil {
		return MultiPoint{}, err
	}
	var pts []Point
	for i := 0; i < int(numPoints); i++ {
		pt, err := p.nextPoint()
		if err != nil {
			return MultiPoint{}, err
		}
		pts = append(pts, pt)
	}
	return NewMultiPoint(pts, p.opts...), nil
}

func (p *TWKBParser) parseMultiLineString() (MultiLineString, error) {
	if p.isEmpty {
		return NewMultiLineString(nil, p.opts...), nil
	}
	return p.nextMultiLineString()
}

func (p *TWKBParser) nextMultiLineString() (MultiLineString, error) {
	numLineStrings, err := p.parseUnsignedVarint()
	if err != nil {
		return MultiLineString{}, p.parserError("cannot unmarshal num linestrings varint malformed")
	}
	_, err = p.parseIDList(int(numLineStrings))
	if err != nil {
		return MultiLineString{}, err
	}
	var lines []LineString
	for i := 0; i < int(numLineStrings); i++ {
		ls, err := p.nextLineString()
		if err != nil {
			return MultiLineString{}, err
		}
		lines = append(lines, ls)
	}
	return NewMultiLineString(lines, p.opts...), nil
}

func (p *TWKBParser) parseMultiPolygon() (MultiPolygon, error) {
	if p.isEmpty {
		return NewMultiPolygon(nil, p.opts...)
	}
	return p.nextMultiPolygon()
}

func (p *TWKBParser) nextMultiPolygon() (MultiPolygon, error) {
	numPolygons, err := p.parseUnsignedVarint()
	if err != nil {
		return MultiPolygon{}, p.parserError("cannot unmarshal num polygons varint malformed")
	}
	_, err = p.parseIDList(int(numPolygons))
	if err != nil {
		return MultiPolygon{}, err
	}
	var polys []Polygon
	for i := 0; i < int(numPolygons); i++ {
		poly, err := p.nextPolygon()
		if err != nil {
			return MultiPolygon{}, err
		}
		polys = append(polys, poly)
	}
	mp, err := NewMultiPolygon(polys, p.opts...)
	return mp, p.annotateError(err)
}

func (p *TWKBParser) parseGeometryCollection() (GeometryCollection, error) {
	if p.isEmpty {
		return NewGeometryCollection(nil, p.opts...), nil
	}
	return p.nextGeometryCollection()
}

func (p *TWKBParser) nextGeometryCollection() (GeometryCollection, error) {
	numGeoms, err := p.parseUnsignedVarint()
	if err != nil {
		return GeometryCollection{}, p.parserError("cannot unmarshal num polygons varint malformed")
	}
	_, err = p.parseIDList(int(numGeoms))
	if err != nil {
		return GeometryCollection{}, err
	}
	var geoms []Geometry
	for i := 0; i < int(numGeoms); i++ {
		subParser := newTWKBParser(p.twkb[p.pos:], p.opts...)
		geom, nbytes, err := subParser.parseGeometry()
		if err != nil {
			return GeometryCollection{}, err
		}
		p.pos += nbytes // Sub-parser's geometry has been read, so ensure it is skipped.
		geoms = append(geoms, geom)
	}
	return NewGeometryCollection(geoms, p.opts...), nil
}

// Read a number of points then convert that many points from int to float coords.
// Utilise and update the running memory of the previous reference point.
func (p *TWKBParser) parsePointCountAndArray() ([]float64, error) {
	numPoints, err := p.parseUnsignedVarint()
	if err != nil {
		return nil, p.parserError("cannot unmarshal num points varint malformed")
	}

	return p.parsePointArray(int(numPoints))
}

// Convert a given number of points from integer to floating point coordinates.
// Utilise and update the running memory of the previous reference point.
// The returned array will contain numPoints * the number of dimensions values.
func (p *TWKBParser) parsePointArray(numPoints int) ([]float64, error) {
	var coords = make([]float64, numPoints*p.dimensions)
	c := 0
	for i := 0; i < numPoints; i++ {
		for d := 0; d < p.dimensions; d++ {
			val, err := p.parseSignedVarint()
			if err != nil {
				return nil, p.parserError("cannot unmarshal point %d of %d coord %d varint malformed", i, numPoints, d)
			}

			p.refpoint[d] += val // Reverse coord differencing to find the true value.
			coords[c] = float64(p.refpoint[d]) * p.scalings[d]
			c++
		}
	}
	return coords, nil
}

func (p *TWKBParser) parseIDList(numIDs int) ([]int, error) {
	if !p.hasIDs {
		return nil, nil
	}
	var ids = make([]int, numIDs)
	for i := 0; i < numIDs; i++ {
		id, err := p.parseSignedVarint()
		if err != nil {
			return nil, p.parserError("cannot unmarshal ID varint in ID list")
		}
		ids = append(ids, int(id))
	}
	return ids, nil
}

func (p *TWKBParser) parseUnsignedVarint() (uint64, error) {
	// LEB128.
	val, n := binary.Uvarint(p.twkb[p.pos:])
	if n <= 0 {
		return 0, p.parserError("problem parsing unsigned varint")
	}
	p.pos += n // Have now read the varint.
	return val, nil
}

func (p *TWKBParser) parseSignedVarint() (int64, error) {
	// LEB128 Zig-Zag encoded.
	uval, n := binary.Uvarint(p.twkb[p.pos:])
	if n <= 0 {
		return 0, p.parserError("problem parsing signed varint")
	}
	p.pos += n // Have now read the varint.
	return DecodeZigZagInt64(uval), nil
}

func (p *TWKBParser) parserError(msg string, vals ...interface{}) error {
	s := fmt.Sprintf("TWKB parser error at byte %d: %s", p.pos, msg)
	return fmt.Errorf(s, vals...)
}

func (p *TWKBParser) annotateError(err error) error {
	if err == nil {
		return nil
	}
	return p.parserError(err.Error())
}

// DecodeZigZagInt32 accepts a uint32 and reverses the zigzag encoding
// to produce the decoded signed int32 value.
func DecodeZigZagInt32(z uint32) int32 {
	return int32(z>>1) ^ -int32(z&1)
}

// DecodeZigZagInt64 accepts a uint64 and reverses the zigzag encoding
// to produce the decoded signed int64 value.
func DecodeZigZagInt64(z uint64) int64 {
	return int64(z>>1) ^ -int64(z&1)
}

// EncodeZigZagInt32 accepts a signed int32 and zigzag encodes
// it to produce an encoded uint32 value.
func EncodeZigZagInt32(n int32) uint32 {
	return uint32((n << 1) ^ (n >> 31))
}

// EncodeZigZagInt64 accepts a signed int64 and zigzag encodes
// it to produce an encoded uint64 value.
func EncodeZigZagInt64(n int64) uint64 {
	return uint64((n << 1) ^ (n >> 63))
}
