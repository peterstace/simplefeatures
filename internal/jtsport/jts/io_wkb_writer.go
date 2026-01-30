package jts

import (
	"bytes"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Io_WKBWriter_ToHex converts a byte array to a hexadecimal string.
func Io_WKBWriter_ToHex(b []byte) string {
	var buf bytes.Buffer
	for i := 0; i < len(b); i++ {
		buf.WriteByte(io_WKBWriter_toHexDigit((b[i] >> 4) & 0x0F))
		buf.WriteByte(io_WKBWriter_toHexDigit(b[i] & 0x0F))
	}
	return buf.String()
}

func io_WKBWriter_toHexDigit(n byte) byte {
	if n < 0 || n > 15 {
		panic("Nibble value out of range")
	}
	if n <= 9 {
		return '0' + n
	}
	return 'A' + (n - 10)
}

// Io_WKBWriter writes a Geometry into Well-Known Binary format.
type Io_WKBWriter struct {
	outputOrdinates    *Io_OrdinateSet
	outputDimension    int
	byteOrder          int
	includeSRID        bool
	byteArrayOS        *bytes.Buffer
	byteArrayOutStream Io_OutStream
	buf                []byte
}

// Io_NewWKBWriter creates a writer with output dimension = 2 and BIG_ENDIAN byte order.
func Io_NewWKBWriter() *Io_WKBWriter {
	return Io_NewWKBWriterWithDimensionAndOrder(2, Io_ByteOrderValues_BIG_ENDIAN)
}

// Io_NewWKBWriterWithDimension creates a writer with the given dimension and BIG_ENDIAN byte order.
func Io_NewWKBWriterWithDimension(outputDimension int) *Io_WKBWriter {
	return Io_NewWKBWriterWithDimensionAndOrder(outputDimension, Io_ByteOrderValues_BIG_ENDIAN)
}

// Io_NewWKBWriterWithDimensionAndSRID creates a writer with the given dimension,
// BIG_ENDIAN byte order, and SRID flag.
func Io_NewWKBWriterWithDimensionAndSRID(outputDimension int, includeSRID bool) *Io_WKBWriter {
	return Io_NewWKBWriterWithDimensionOrderAndSRID(outputDimension, Io_ByteOrderValues_BIG_ENDIAN, includeSRID)
}

// Io_NewWKBWriterWithDimensionAndOrder creates a writer with the given dimension and byte order.
func Io_NewWKBWriterWithDimensionAndOrder(outputDimension int, byteOrder int) *Io_WKBWriter {
	return Io_NewWKBWriterWithDimensionOrderAndSRID(outputDimension, byteOrder, false)
}

// Io_NewWKBWriterWithDimensionOrderAndSRID creates a writer with the given dimension,
// byte order, and SRID flag.
func Io_NewWKBWriterWithDimensionOrderAndSRID(outputDimension int, byteOrder int, includeSRID bool) *Io_WKBWriter {
	if outputDimension < 2 || outputDimension > 4 {
		panic("Output dimension must be 2 to 4")
	}

	outputOrdinates := Io_Ordinate_CreateXY()
	if outputDimension > 2 {
		outputOrdinates.Add(Io_Ordinate_Z)
	}
	if outputDimension > 3 {
		outputOrdinates.Add(Io_Ordinate_M)
	}

	byteArrayOS := &bytes.Buffer{}
	return &Io_WKBWriter{
		outputDimension:    outputDimension,
		byteOrder:          byteOrder,
		includeSRID:        includeSRID,
		outputOrdinates:    outputOrdinates,
		byteArrayOS:        byteArrayOS,
		byteArrayOutStream: Io_NewOutputStreamOutStream(byteArrayOS),
		buf:                make([]byte, 8),
	}
}

// SetOutputOrdinates sets the ordinates to be written.
func (w *Io_WKBWriter) SetOutputOrdinates(outputOrdinates *Io_OrdinateSet) {
	w.outputOrdinates.Remove(Io_Ordinate_Z)
	w.outputOrdinates.Remove(Io_Ordinate_M)

	if w.outputDimension == 3 {
		if outputOrdinates.Contains(Io_Ordinate_Z) {
			w.outputOrdinates.Add(Io_Ordinate_Z)
		} else if outputOrdinates.Contains(Io_Ordinate_M) {
			w.outputOrdinates.Add(Io_Ordinate_M)
		}
	}
	if w.outputDimension == 4 {
		if outputOrdinates.Contains(Io_Ordinate_Z) {
			w.outputOrdinates.Add(Io_Ordinate_Z)
		}
		if outputOrdinates.Contains(Io_Ordinate_M) {
			w.outputOrdinates.Add(Io_Ordinate_M)
		}
	}
}

// GetOutputOrdinates returns the ordinates being written.
func (w *Io_WKBWriter) GetOutputOrdinates() *Io_OrdinateSet {
	return w.outputOrdinates
}

// Write writes a Geometry into a byte array.
func (w *Io_WKBWriter) Write(geom *Geom_Geometry) []byte {
	w.byteArrayOS.Reset()
	if err := w.WriteToStream(geom, w.byteArrayOutStream); err != nil {
		panic("Unexpected IO exception: " + err.Error())
	}
	return w.byteArrayOS.Bytes()
}

// WriteToStream writes a Geometry to an OutStream.
func (w *Io_WKBWriter) WriteToStream(geom *Geom_Geometry, os Io_OutStream) error {
	switch g := java.GetLeaf(geom).(type) {
	case *Geom_Point:
		return w.writePoint(g, os)
	case *Geom_LineString:
		return w.writeLineString(g, os)
	case *Geom_LinearRing:
		return w.writeLineString(g.Geom_LineString, os)
	case *Geom_Polygon:
		return w.writePolygon(g, os)
	case *Geom_MultiPoint:
		return w.writeGeometryCollection(Io_WKBConstants_wkbMultiPoint, g.Geom_GeometryCollection, os)
	case *Geom_MultiLineString:
		return w.writeGeometryCollection(Io_WKBConstants_wkbMultiLineString, g.Geom_GeometryCollection, os)
	case *Geom_MultiPolygon:
		return w.writeGeometryCollection(Io_WKBConstants_wkbMultiPolygon, g.Geom_GeometryCollection, os)
	case *Geom_GeometryCollection:
		return w.writeGeometryCollection(Io_WKBConstants_wkbGeometryCollection, g, os)
	default:
		Util_Assert_ShouldNeverReachHereWithMessage("Unknown Geometry type")
		return nil
	}
}

func (w *Io_WKBWriter) writePoint(pt *Geom_Point, os Io_OutStream) error {
	if err := w.writeByteOrder(os); err != nil {
		return err
	}
	if err := w.writeGeometryType(Io_WKBConstants_wkbPoint, pt.Geom_Geometry, os); err != nil {
		return err
	}
	if pt.GetCoordinateSequence().Size() == 0 {
		return w.writeNaNs(w.outputDimension, os)
	}
	return w.writeCoordinateSequence(pt.GetCoordinateSequence(), false, os)
}

func (w *Io_WKBWriter) writeLineString(line *Geom_LineString, os Io_OutStream) error {
	if err := w.writeByteOrder(os); err != nil {
		return err
	}
	if err := w.writeGeometryType(Io_WKBConstants_wkbLineString, line.Geom_Geometry, os); err != nil {
		return err
	}
	return w.writeCoordinateSequence(line.GetCoordinateSequence(), true, os)
}

func (w *Io_WKBWriter) writePolygon(poly *Geom_Polygon, os Io_OutStream) error {
	if err := w.writeByteOrder(os); err != nil {
		return err
	}
	if err := w.writeGeometryType(Io_WKBConstants_wkbPolygon, poly.Geom_Geometry, os); err != nil {
		return err
	}
	if poly.IsEmpty() {
		return w.writeInt(0, os)
	}
	if err := w.writeInt(int32(poly.GetNumInteriorRing()+1), os); err != nil {
		return err
	}
	if err := w.writeCoordinateSequence(poly.GetExteriorRing().GetCoordinateSequence(), true, os); err != nil {
		return err
	}
	for i := 0; i < poly.GetNumInteriorRing(); i++ {
		if err := w.writeCoordinateSequence(poly.GetInteriorRingN(i).GetCoordinateSequence(), true, os); err != nil {
			return err
		}
	}
	return nil
}

func (w *Io_WKBWriter) writeGeometryCollection(geometryType int, gc *Geom_GeometryCollection, os Io_OutStream) error {
	if err := w.writeByteOrder(os); err != nil {
		return err
	}
	if err := w.writeGeometryType(geometryType, gc.Geom_Geometry, os); err != nil {
		return err
	}
	if err := w.writeInt(int32(gc.GetNumGeometries()), os); err != nil {
		return err
	}
	originalIncludeSRID := w.includeSRID
	w.includeSRID = false
	for i := 0; i < gc.GetNumGeometries(); i++ {
		if err := w.WriteToStream(gc.GetGeometryN(i), os); err != nil {
			w.includeSRID = originalIncludeSRID
			return err
		}
	}
	w.includeSRID = originalIncludeSRID
	return nil
}

func (w *Io_WKBWriter) writeByteOrder(os Io_OutStream) error {
	if w.byteOrder == Io_ByteOrderValues_LITTLE_ENDIAN {
		w.buf[0] = Io_WKBConstants_wkbNDR
	} else {
		w.buf[0] = Io_WKBConstants_wkbXDR
	}
	return os.Write(w.buf, 1)
}

func (w *Io_WKBWriter) writeGeometryType(geometryType int, g *Geom_Geometry, os Io_OutStream) error {
	ordinals := 0
	if w.outputOrdinates.Contains(Io_Ordinate_Z) {
		ordinals |= 0x80000000
	}
	if w.outputOrdinates.Contains(Io_Ordinate_M) {
		ordinals |= 0x40000000
	}

	flag3D := 0
	if w.outputDimension > 2 {
		flag3D = ordinals
	}
	typeInt := geometryType | flag3D
	if w.includeSRID {
		typeInt |= 0x20000000
	}
	if err := w.writeInt(int32(typeInt), os); err != nil {
		return err
	}
	if w.includeSRID {
		return w.writeInt(int32(g.GetSRID()), os)
	}
	return nil
}

func (w *Io_WKBWriter) writeInt(intValue int32, os Io_OutStream) error {
	Io_ByteOrderValues_PutInt(intValue, w.buf, w.byteOrder)
	return os.Write(w.buf, 4)
}

func (w *Io_WKBWriter) writeCoordinateSequence(seq Geom_CoordinateSequence, writeSize bool, os Io_OutStream) error {
	if writeSize {
		if err := w.writeInt(int32(seq.Size()), os); err != nil {
			return err
		}
	}
	for i := 0; i < seq.Size(); i++ {
		if err := w.writeCoordinate(seq, i, os); err != nil {
			return err
		}
	}
	return nil
}

func (w *Io_WKBWriter) writeCoordinate(seq Geom_CoordinateSequence, index int, os Io_OutStream) error {
	Io_ByteOrderValues_PutDouble(seq.GetX(index), w.buf, w.byteOrder)
	if err := os.Write(w.buf, 8); err != nil {
		return err
	}
	Io_ByteOrderValues_PutDouble(seq.GetY(index), w.buf, w.byteOrder)
	if err := os.Write(w.buf, 8); err != nil {
		return err
	}

	if w.outputDimension >= 3 {
		ordVal := seq.GetOrdinate(index, 2)
		Io_ByteOrderValues_PutDouble(ordVal, w.buf, w.byteOrder)
		if err := os.Write(w.buf, 8); err != nil {
			return err
		}
	}
	if w.outputDimension == 4 {
		ordVal := seq.GetOrdinate(index, 3)
		Io_ByteOrderValues_PutDouble(ordVal, w.buf, w.byteOrder)
		if err := os.Write(w.buf, 8); err != nil {
			return err
		}
	}
	return nil
}

func (w *Io_WKBWriter) writeNaNs(numNaNs int, os Io_OutStream) error {
	for i := 0; i < numNaNs; i++ {
		Io_ByteOrderValues_PutDouble(java.CanonicalNaN, w.buf, w.byteOrder)
		if err := os.Write(w.buf, 8); err != nil {
			return err
		}
	}
	return nil
}
