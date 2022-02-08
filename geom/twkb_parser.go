package geom

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// Tiny Well Known Binary
// See spec https://github.com/TWKB/Specification/blob/master/twkb.md

// UnmarshalTWKB parses a Tiny Well Known Binary (TWKB), returning the
// corresponding Geometry.
func UnmarshalTWKB(twkb []byte, opts ...ConstructorOption) (Geometry, error) {
	p := newTWKBParser(twkb, opts...)
	geom, err := p.nextGeometry()
	return geom, p.annotateError(err)
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

// Annotate an error message with the byte offset where it happened.
func (p *TWKBParser) annotateError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("TWKB parsing error at byte %d: %s", p.pos, err.Error())
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
		pt, err := p.parsePoint()
		if err != nil {
			return Geometry{}, err
		}
		return pt.AsGeometry(), nil
	case twkbTypeLineString:
		ls, err := p.parseLineString()
		if err != nil {
			return Geometry{}, err
		}
		return ls.AsGeometry(), nil
	case twkbTypePolygon:
		poly, err := p.parsePolygon()
		if err != nil {
			return Geometry{}, err
		}
		return poly.AsGeometry(), nil
	case twkbTypeMultiPoint:
		mp, err := p.parseMultiPoint()
		if err != nil {
			return Geometry{}, err
		}
		return mp.AsGeometry(), nil
	case twkbTypeMultiLineString:
		ml, err := p.parseMultiLineString()
		if err != nil {
			return Geometry{}, err
		}
		return ml.AsGeometry(), nil
	case twkbTypeMultiPolygon:
		mp, err := p.parseMultiPolygon()
		if err != nil {
			return Geometry{}, err
		}
		return mp.AsGeometry(), nil
	case twkbTypeGeometryCollection:
		gc, err := p.parseGeometryCollection()
		if err != nil {
			return Geometry{}, err
		}
		return gc.AsGeometry(), nil
	}

	return Geometry{}, fmt.Errorf("unsupported geometry type %d", p.kind)
}

func (p *TWKBParser) parseTypeAndPrecision() error {
	if len(p.twkb) <= p.pos {
		return errors.New("empty buffer")
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
		return errors.New("lacking metadata header")
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
		return errors.New("lacking extended pecision byte")
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
		return errors.New("lacking size varint")
	}
	bytesRemaining, err := p.parseUnsignedVarint()
	if err != nil {
		return fmt.Errorf("size varint malformed: %s", err.Error())
	}
	if uint64(p.pos)+bytesRemaining > uint64(len(p.twkb)) {
		return fmt.Errorf("remaining input (%d bytes) smaller than size varint indicates (%d bytes)", len(p.twkb)-p.pos, bytesRemaining)
	}
	return nil
}

func (p *TWKBParser) parseBBox() error {
	// Parse the bounding box, but do not otherwise use it.
	for i := 0; i < p.dimensions; i++ {
		minVal, err := p.parseSignedVarint()
		if err != nil {
			return fmt.Errorf("BBox min varint malformed: %s", err.Error())
		}
		p.bbox = append(p.bbox, minVal)

		deltaVal, err := p.parseSignedVarint()
		if err != nil {
			return fmt.Errorf("BBox delta varint malformed: %s", err.Error())
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

	return NewPoint(c, p.opts...)
}

func (p *TWKBParser) parseLineString() (LineString, error) {
	if p.isEmpty {
		return NewLineString(NewSequence(nil, p.ctype), p.opts...)
	}
	return p.nextLineString()
}

func (p *TWKBParser) nextLineString() (LineString, error) {
	coords, _, err := p.parsePointCountAndArray()
	if err != nil {
		return LineString{}, err
	}
	return NewLineString(NewSequence(coords, p.ctype), p.opts...)
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
		return Polygon{}, fmt.Errorf("num rings varint malformed: %s", err.Error())
	}

	var rings []LineString
	for r := uint64(0); r < numRings; r++ {
		coords, numPoints, err := p.parsePointCountAndArray()
		if err != nil {
			return Polygon{}, err
		}
		if numPoints >= 2 {
			// The spec says rings may need to be closed:
			// "rings are assumed to be implicitly closed, so the first
			//  and last point should not be the same"

			// Note, some implementations can generate TWKB with the rings
			// already closed. We wish to gracefully parse these cases too.
			// So check if any of the first and final point's coords differ.
			finalPointDiffersFromFirst := false
			for d := 0; d < p.dimensions; d++ {
				first := coords[d]
				final := coords[d+(numPoints-1)*p.dimensions]
				if first != final {
					finalPointDiffersFromFirst = true
					break
				}
			}

			if finalPointDiffersFromFirst {
				// Append first point again, to close the ring.
				for d := 0; d < p.dimensions; d++ {
					coords = append(coords, coords[0+d])
				}
			}
		}
		ls, err := NewLineString(NewSequence(coords, p.ctype), p.opts...)
		if err != nil {
			return Polygon{}, err
		}
		rings = append(rings, ls)
	}
	return NewPolygon(rings, p.opts...)
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
		return MultiPoint{}, fmt.Errorf("num points varint malformed: %s", err.Error())
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
		return MultiLineString{}, fmt.Errorf("num linestrings varint malformed: %s", err.Error())
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
		return MultiPolygon{}, fmt.Errorf("num polygons varint malformed: %s", err.Error())
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
	return NewMultiPolygon(polys, p.opts...)
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
		return GeometryCollection{}, fmt.Errorf("num polygons varint malformed: %s", err.Error())
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
			p.pos += nbytes // Add sub-parser's last known position, for error reporting.
			return GeometryCollection{}, err
		}
		p.pos += nbytes // Sub-parser's geometry has been read, so ensure it is skipped.
		geoms = append(geoms, geom)
	}
	return NewGeometryCollection(geoms, p.opts...), nil
}

// Read a number of points then convert that many points from int to float coords.
// Utilise and update the running memory of the previous reference point.
// Return the slice of coords, the number of points, and any error.
func (p *TWKBParser) parsePointCountAndArray() ([]float64, int, error) {
	numPoints, err := p.parseUnsignedVarint()
	if err != nil {
		return nil, 0, fmt.Errorf("num points varint malformed: %s", err.Error())
	}

	coords, err := p.parsePointArray(int(numPoints))
	return coords, int(numPoints), err
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
				return nil, fmt.Errorf("point %d of %d coord %d varint malformed: %s", i, numPoints, d, err.Error())
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
			return nil, fmt.Errorf("ID list varint %d of %d malformed: %s", i, numIDs, err.Error())
		}
		ids = append(ids, int(id))
	}
	return ids, nil
}

func (p *TWKBParser) parseUnsignedVarint() (uint64, error) {
	// LEB128.
	val, n := binary.Uvarint(p.twkb[p.pos:])
	if n == 0 {
		return 0, errors.New("problem parsing unsigned varint: buffer too small")
	}
	if n < 0 {
		return 0, errors.New("problem parsing unsigned varint: numeric overflow")
	}
	p.pos += n // Have now read the varint.
	return val, nil
}

func (p *TWKBParser) parseSignedVarint() (int64, error) {
	// LEB128 Zig-Zag encoded.
	val, n := binary.Varint(p.twkb[p.pos:])
	if n == 0 {
		return 0, errors.New("problem parsing signed varint: buffer too small")
	}
	if n < 0 {
		return 0, errors.New("problem parsing signed varint: numeric overflow")
	}
	p.pos += n // Have now read the varint.
	return val, nil
}
