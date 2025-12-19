package jts

import (
	"fmt"
	"math"
)

// Io_WKBReader_HexToBytes converts a hexadecimal string to a byte array.
// The hexadecimal digit symbols are case-insensitive.
func Io_WKBReader_HexToBytes(hex string) []byte {
	byteLen := len(hex) / 2
	bytes := make([]byte, byteLen)

	for i := 0; i < len(hex)/2; i++ {
		i2 := 2 * i
		if i2+1 > len(hex) {
			panic("Hex string has odd length")
		}

		nib1 := io_WKBReader_hexToInt(hex[i2])
		nib0 := io_WKBReader_hexToInt(hex[i2+1])
		b := byte((nib1 << 4) + nib0)
		bytes[i] = b
	}
	return bytes
}

func io_WKBReader_hexToInt(hex byte) int {
	switch {
	case hex >= '0' && hex <= '9':
		return int(hex - '0')
	case hex >= 'a' && hex <= 'f':
		return int(hex - 'a' + 10)
	case hex >= 'A' && hex <= 'F':
		return int(hex - 'A' + 10)
	default:
		panic("Invalid hex digit: '" + string(hex) + "'")
	}
}

const (
	io_WKBReader_INVALID_GEOM_TYPE_MSG = "Invalid geometry type encountered in "
	io_WKBReader_FIELD_NUMCOORDS       = "numCoords"
	io_WKBReader_FIELD_NUMRINGS        = "numRings"
	io_WKBReader_FIELD_NUMELEMS        = "numElems"
)

// Io_WKBReader reads a Geometry from a byte stream in Well-Known Binary format.
type Io_WKBReader struct {
	factory          *Geom_GeometryFactory
	csFactory        Geom_CoordinateSequenceFactory
	precisionModel   *Geom_PrecisionModel
	inputDimension   int
	isStrict         bool
	dis              *Io_ByteOrderDataInStream
	ordValues        []float64
	maxNumFieldValue int
}

// Io_NewWKBReader creates a new WKBReader with the default GeometryFactory.
func Io_NewWKBReader() *Io_WKBReader {
	return Io_NewWKBReaderWithFactory(Geom_NewGeometryFactoryDefault())
}

// Io_NewWKBReaderWithFactory creates a new WKBReader with the given GeometryFactory.
func Io_NewWKBReaderWithFactory(geometryFactory *Geom_GeometryFactory) *Io_WKBReader {
	return &Io_WKBReader{
		factory:        geometryFactory,
		precisionModel: geometryFactory.GetPrecisionModel(),
		csFactory:      geometryFactory.GetCoordinateSequenceFactory(),
		inputDimension: 2,
		isStrict:       false,
		dis:            Io_NewByteOrderDataInStream(),
	}
}

// ReadBytes reads a single Geometry in WKB format from a byte array.
func (r *Io_WKBReader) ReadBytes(bytes []byte) (*Geom_Geometry, error) {
	return r.read(Io_NewByteArrayInStream(bytes), len(bytes)/8)
}

// ReadStream reads a Geometry in binary WKB format from an InStream.
func (r *Io_WKBReader) ReadStream(is Io_InStream) (*Geom_Geometry, error) {
	return r.read(is, math.MaxInt32)
}

func (r *Io_WKBReader) read(is Io_InStream, maxCoordNum int) (*Geom_Geometry, error) {
	r.maxNumFieldValue = maxCoordNum
	r.dis.SetInStream(is)
	return r.readGeometry(0)
}

func (r *Io_WKBReader) readNumField(fieldName string) (int, error) {
	num, err := r.dis.ReadInt()
	if err != nil {
		return 0, err
	}
	if num < 0 || int(num) > r.maxNumFieldValue {
		return 0, Io_NewParseException(fieldName + " value is too large")
	}
	return int(num), nil
}

func (r *Io_WKBReader) readGeometry(SRID int) (*Geom_Geometry, error) {
	byteOrderWKB, err := r.dis.ReadByte()
	if err != nil {
		return nil, err
	}

	if byteOrderWKB == Io_WKBConstants_wkbNDR {
		r.dis.SetOrder(Io_ByteOrderValues_LITTLE_ENDIAN)
	} else if byteOrderWKB == Io_WKBConstants_wkbXDR {
		r.dis.SetOrder(Io_ByteOrderValues_BIG_ENDIAN)
	} else if r.isStrict {
		return nil, Io_NewParseException(fmt.Sprintf("Unknown geometry byte order (not NDR or XDR): %d", byteOrderWKB))
	}

	typeInt, err := r.dis.ReadInt()
	if err != nil {
		return nil, err
	}

	// To get geometry type mask out EWKB flag bits, and use only low 3 digits.
	geometryType := (int(typeInt) & 0xffff) % 1000

	// Handle 3D and 4D WKB geometries.
	hasZ := (int(typeInt)&0x80000000) != 0 || (int(typeInt)&0xffff)/1000 == 1 || (int(typeInt)&0xffff)/1000 == 3
	hasM := (int(typeInt)&0x40000000) != 0 || (int(typeInt)&0xffff)/1000 == 2 || (int(typeInt)&0xffff)/1000 == 3
	r.inputDimension = 2
	if hasZ {
		r.inputDimension++
	}
	if hasM {
		r.inputDimension++
	}

	ordinateFlags := Io_Ordinate_CreateXY()
	if hasZ {
		ordinateFlags.Add(Io_Ordinate_Z)
	}
	if hasM {
		ordinateFlags.Add(Io_Ordinate_M)
	}

	// Determine if SRIDs are present (EWKB only).
	hasSRID := (int(typeInt) & 0x20000000) != 0
	if hasSRID {
		sridVal, err := r.dis.ReadInt()
		if err != nil {
			return nil, err
		}
		SRID = int(sridVal)
	}

	// Only allocate ordValues buffer if necessary.
	if r.ordValues == nil || len(r.ordValues) < r.inputDimension {
		r.ordValues = make([]float64, r.inputDimension)
	}

	var geom *Geom_Geometry
	switch geometryType {
	case Io_WKBConstants_wkbPoint:
		geom, err = r.readPoint(ordinateFlags)
	case Io_WKBConstants_wkbLineString:
		geom, err = r.readLineString(ordinateFlags)
	case Io_WKBConstants_wkbPolygon:
		geom, err = r.readPolygon(ordinateFlags)
	case Io_WKBConstants_wkbMultiPoint:
		geom, err = r.readMultiPoint(SRID)
	case Io_WKBConstants_wkbMultiLineString:
		geom, err = r.readMultiLineString(SRID)
	case Io_WKBConstants_wkbMultiPolygon:
		geom, err = r.readMultiPolygon(SRID)
	case Io_WKBConstants_wkbGeometryCollection:
		geom, err = r.readGeometryCollection(SRID)
	default:
		return nil, Io_NewParseException(fmt.Sprintf("Unknown WKB type %d", geometryType))
	}
	if err != nil {
		return nil, err
	}

	r.setSRID(geom, SRID)
	return geom, nil
}

func (r *Io_WKBReader) setSRID(g *Geom_Geometry, SRID int) {
	if SRID != 0 {
		g.SetSRID(SRID)
	}
}

func (r *Io_WKBReader) readPoint(ordinateFlags *Io_OrdinateSet) (*Geom_Geometry, error) {
	pts, err := r.readCoordinateSequence(1, ordinateFlags)
	if err != nil {
		return nil, err
	}
	// If X and Y are NaN create an empty point.
	if math.IsNaN(pts.GetX(0)) || math.IsNaN(pts.GetY(0)) {
		return r.factory.CreatePoint().Geom_Geometry, nil
	}
	return r.factory.CreatePointFromCoordinateSequence(pts).Geom_Geometry, nil
}

func (r *Io_WKBReader) readLineString(ordinateFlags *Io_OrdinateSet) (*Geom_Geometry, error) {
	size, err := r.readNumField(io_WKBReader_FIELD_NUMCOORDS)
	if err != nil {
		return nil, err
	}
	pts, err := r.readCoordinateSequenceLineString(size, ordinateFlags)
	if err != nil {
		return nil, err
	}
	return r.factory.CreateLineStringFromCoordinateSequence(pts).Geom_Geometry, nil
}

func (r *Io_WKBReader) readLinearRing(ordinateFlags *Io_OrdinateSet) (*Geom_LinearRing, error) {
	size, err := r.readNumField(io_WKBReader_FIELD_NUMCOORDS)
	if err != nil {
		return nil, err
	}
	pts, err := r.readCoordinateSequenceRing(size, ordinateFlags)
	if err != nil {
		return nil, err
	}
	return r.factory.CreateLinearRingFromCoordinateSequence(pts), nil
}

func (r *Io_WKBReader) readPolygon(ordinateFlags *Io_OrdinateSet) (*Geom_Geometry, error) {
	numRings, err := r.readNumField(io_WKBReader_FIELD_NUMRINGS)
	if err != nil {
		return nil, err
	}

	var holes []*Geom_LinearRing
	if numRings > 1 {
		holes = make([]*Geom_LinearRing, numRings-1)
	}

	// Empty polygon.
	if numRings <= 0 {
		return r.factory.CreatePolygon().Geom_Geometry, nil
	}

	shell, err := r.readLinearRing(ordinateFlags)
	if err != nil {
		return nil, err
	}
	for i := 0; i < numRings-1; i++ {
		holes[i], err = r.readLinearRing(ordinateFlags)
		if err != nil {
			return nil, err
		}
	}
	return r.factory.CreatePolygonWithLinearRingAndHoles(shell, holes).Geom_Geometry, nil
}

func (r *Io_WKBReader) readMultiPoint(SRID int) (*Geom_Geometry, error) {
	numGeom, err := r.readNumField(io_WKBReader_FIELD_NUMELEMS)
	if err != nil {
		return nil, err
	}
	geoms := make([]*Geom_Point, numGeom)
	for i := 0; i < numGeom; i++ {
		g, err := r.readGeometry(SRID)
		if err != nil {
			return nil, err
		}
		pt, ok := g.GetChild().(*Geom_Point)
		if !ok {
			return nil, Io_NewParseException(io_WKBReader_INVALID_GEOM_TYPE_MSG + "MultiPoint")
		}
		geoms[i] = pt
	}
	return r.factory.CreateMultiPointFromPoints(geoms).Geom_Geometry, nil
}

func (r *Io_WKBReader) readMultiLineString(SRID int) (*Geom_Geometry, error) {
	numGeom, err := r.readNumField(io_WKBReader_FIELD_NUMELEMS)
	if err != nil {
		return nil, err
	}
	geoms := make([]*Geom_LineString, numGeom)
	for i := 0; i < numGeom; i++ {
		g, err := r.readGeometry(SRID)
		if err != nil {
			return nil, err
		}
		ls, ok := g.GetChild().(*Geom_LineString)
		if !ok {
			return nil, Io_NewParseException(io_WKBReader_INVALID_GEOM_TYPE_MSG + "MultiLineString")
		}
		geoms[i] = ls
	}
	return r.factory.CreateMultiLineStringFromLineStrings(geoms).Geom_Geometry, nil
}

func (r *Io_WKBReader) readMultiPolygon(SRID int) (*Geom_Geometry, error) {
	numGeom, err := r.readNumField(io_WKBReader_FIELD_NUMELEMS)
	if err != nil {
		return nil, err
	}
	geoms := make([]*Geom_Polygon, numGeom)
	for i := 0; i < numGeom; i++ {
		g, err := r.readGeometry(SRID)
		if err != nil {
			return nil, err
		}
		poly, ok := g.GetChild().(*Geom_Polygon)
		if !ok {
			return nil, Io_NewParseException(io_WKBReader_INVALID_GEOM_TYPE_MSG + "MultiPolygon")
		}
		geoms[i] = poly
	}
	return r.factory.CreateMultiPolygonFromPolygons(geoms).Geom_Geometry, nil
}

func (r *Io_WKBReader) readGeometryCollection(SRID int) (*Geom_Geometry, error) {
	numGeom, err := r.readNumField(io_WKBReader_FIELD_NUMELEMS)
	if err != nil {
		return nil, err
	}
	geoms := make([]*Geom_Geometry, numGeom)
	for i := 0; i < numGeom; i++ {
		geoms[i], err = r.readGeometry(SRID)
		if err != nil {
			return nil, err
		}
	}
	return r.factory.CreateGeometryCollectionFromGeometries(geoms).Geom_Geometry, nil
}

func (r *Io_WKBReader) readCoordinateSequence(size int, ordinateFlags *Io_OrdinateSet) (Geom_CoordinateSequence, error) {
	measures := 0
	if ordinateFlags.Contains(Io_Ordinate_M) {
		measures = 1
	}
	seq := r.csFactory.CreateWithSizeAndDimensionAndMeasures(size, r.inputDimension, measures)
	targetDim := seq.GetDimension()
	if targetDim > r.inputDimension {
		targetDim = r.inputDimension
	}
	for i := 0; i < size; i++ {
		if err := r.readCoordinate(); err != nil {
			return nil, err
		}
		for j := 0; j < targetDim; j++ {
			seq.SetOrdinate(i, j, r.ordValues[j])
		}
	}
	return seq, nil
}

func (r *Io_WKBReader) readCoordinateSequenceLineString(size int, ordinateFlags *Io_OrdinateSet) (Geom_CoordinateSequence, error) {
	seq, err := r.readCoordinateSequence(size, ordinateFlags)
	if err != nil {
		return nil, err
	}
	if r.isStrict {
		return seq, nil
	}
	if seq.Size() == 0 || seq.Size() >= 2 {
		return seq, nil
	}
	return Geom_CoordinateSequences_Extend(r.csFactory, seq, 2), nil
}

func (r *Io_WKBReader) readCoordinateSequenceRing(size int, ordinateFlags *Io_OrdinateSet) (Geom_CoordinateSequence, error) {
	seq, err := r.readCoordinateSequence(size, ordinateFlags)
	if err != nil {
		return nil, err
	}
	if r.isStrict {
		return seq, nil
	}
	if Geom_CoordinateSequences_IsRing(seq) {
		return seq, nil
	}
	return Geom_CoordinateSequences_EnsureValidRing(r.csFactory, seq), nil
}

func (r *Io_WKBReader) readCoordinate() error {
	for i := 0; i < r.inputDimension; i++ {
		val, err := r.dis.ReadDouble()
		if err != nil {
			return err
		}
		if i <= 1 {
			r.ordValues[i] = r.precisionModel.MakePrecise(val)
		} else {
			r.ordValues[i] = val
		}
	}
	return nil
}
