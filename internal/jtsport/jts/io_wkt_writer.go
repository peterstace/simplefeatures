package jts

import (
	"io"
	"math"
	"strings"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const (
	io_WKTWriter_Indent          = 2
	io_WKTWriter_OutputDimension = 2
)

// Io_WKTWriter_ToPoint generates the WKT for a POINT specified by a Coordinate.
func Io_WKTWriter_ToPoint(p0 *Geom_Coordinate) string {
	return Io_WKTConstants_POINT + " ( " + io_WKTWriter_FormatCoord(p0) + " )"
}

// Io_WKTWriter_ToLineStringFromSeq generates the WKT for a LINESTRING specified
// by a CoordinateSequence.
func Io_WKTWriter_ToLineStringFromSeq(seq Geom_CoordinateSequence) string {
	var buf strings.Builder
	buf.WriteString(Io_WKTConstants_LINESTRING)
	buf.WriteString(" ")
	if seq.Size() == 0 {
		buf.WriteString(Io_WKTConstants_EMPTY)
	} else {
		buf.WriteString("(")
		for i := 0; i < seq.Size(); i++ {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(io_WKTWriter_FormatXY(seq.GetX(i), seq.GetY(i)))
		}
		buf.WriteString(")")
	}
	return buf.String()
}

// Io_WKTWriter_ToLineStringFromCoords generates the WKT for a LINESTRING
// specified by a Coordinate array.
func Io_WKTWriter_ToLineStringFromCoords(coord []*Geom_Coordinate) string {
	var buf strings.Builder
	buf.WriteString(Io_WKTConstants_LINESTRING)
	buf.WriteString(" ")
	if len(coord) == 0 {
		buf.WriteString(Io_WKTConstants_EMPTY)
	} else {
		buf.WriteString("(")
		for i, c := range coord {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(io_WKTWriter_FormatCoord(c))
		}
		buf.WriteString(")")
	}
	return buf.String()
}

// Io_WKTWriter_ToLineStringFromTwoCoords generates the WKT for a LINESTRING
// specified by two Coordinates.
func Io_WKTWriter_ToLineStringFromTwoCoords(p0, p1 *Geom_Coordinate) string {
	return Io_WKTConstants_LINESTRING + " ( " + io_WKTWriter_FormatCoord(p0) + ", " + io_WKTWriter_FormatCoord(p1) + " )"
}

func io_WKTWriter_FormatCoord(p *Geom_Coordinate) string {
	return io_WKTWriter_FormatXY(p.GetX(), p.GetY())
}

func io_WKTWriter_FormatXY(x, y float64) string {
	return Io_OrdinateFormat_Default.Format(x) + " " + Io_OrdinateFormat_Default.Format(y)
}

func io_WKTWriter_CreateFormatter(precisionModel *Geom_PrecisionModel) *Io_OrdinateFormat {
	return Io_OrdinateFormat_Create(precisionModel.GetMaximumSignificantDigits())
}

func io_WKTWriter_StringOfChar(ch byte, count int) string {
	var buf strings.Builder
	buf.Grow(count)
	for i := 0; i < count; i++ {
		buf.WriteByte(ch)
	}
	return buf.String()
}

// io_CheckOrdinatesFilter is a filter implementation to test if a coordinate
// sequence actually has meaningful values for an ordinate bit-pattern.
type io_CheckOrdinatesFilter struct {
	checkOrdinateFlags *Io_OrdinateSet
	outputOrdinates    *Io_OrdinateSet
}

var _ Geom_CoordinateSequenceFilter = (*io_CheckOrdinatesFilter)(nil)

func (f *io_CheckOrdinatesFilter) IsGeom_CoordinateSequenceFilter() {}

func io_NewCheckOrdinatesFilter(checkOrdinateFlags *Io_OrdinateSet) *io_CheckOrdinatesFilter {
	return &io_CheckOrdinatesFilter{
		outputOrdinates:    Io_Ordinate_CreateXY(),
		checkOrdinateFlags: checkOrdinateFlags,
	}
}

func (f *io_CheckOrdinatesFilter) Filter(seq Geom_CoordinateSequence, i int) {
	if f.checkOrdinateFlags.Contains(Io_Ordinate_Z) && !f.outputOrdinates.Contains(Io_Ordinate_Z) {
		if !math.IsNaN(seq.GetZ(i)) {
			f.outputOrdinates.Add(Io_Ordinate_Z)
		}
	}
	if f.checkOrdinateFlags.Contains(Io_Ordinate_M) && !f.outputOrdinates.Contains(Io_Ordinate_M) {
		if !math.IsNaN(seq.GetM(i)) {
			f.outputOrdinates.Add(Io_Ordinate_M)
		}
	}
}

func (f *io_CheckOrdinatesFilter) IsGeometryChanged() bool {
	return false
}

func (f *io_CheckOrdinatesFilter) IsDone() bool {
	return f.outputOrdinates.Equals(f.checkOrdinateFlags)
}

func (f *io_CheckOrdinatesFilter) GetOutputOrdinates() *Io_OrdinateSet {
	return f.outputOrdinates
}

// Io_WKTWriter writes the Well-Known Text representation of a Geometry.
type Io_WKTWriter struct {
	outputOrdinates *Io_OrdinateSet
	outputDimension int
	precisionModel  *Geom_PrecisionModel
	ordinateFormat  *Io_OrdinateFormat
	isFormatted     bool
	coordsPerLine   int
	indentTabStr    string
}

// Io_NewWKTWriter creates a new WKTWriter with default settings.
func Io_NewWKTWriter() *Io_WKTWriter {
	return Io_NewWKTWriterWithDimension(io_WKTWriter_OutputDimension)
}

// Io_NewWKTWriterWithDimension creates a writer that writes Geometries with
// the given output dimension (2 to 4).
func Io_NewWKTWriterWithDimension(outputDimension int) *Io_WKTWriter {
	if outputDimension < 2 || outputDimension > 4 {
		panic("Invalid output dimension (must be 2 to 4)")
	}

	w := &Io_WKTWriter{
		outputDimension: outputDimension,
		outputOrdinates: Io_Ordinate_CreateXY(),
		coordsPerLine:   -1,
	}
	w.SetTab(io_WKTWriter_Indent)

	if outputDimension > 2 {
		w.outputOrdinates.Add(Io_Ordinate_Z)
	}
	if outputDimension > 3 {
		w.outputOrdinates.Add(Io_Ordinate_M)
	}

	return w
}

// SetFormatted sets whether the output will be formatted.
func (w *Io_WKTWriter) SetFormatted(isFormatted bool) {
	w.isFormatted = isFormatted
}

// SetMaxCoordinatesPerLine sets the maximum number of coordinates per line
// written in formatted output.
func (w *Io_WKTWriter) SetMaxCoordinatesPerLine(coordsPerLine int) {
	w.coordsPerLine = coordsPerLine
}

// SetTab sets the tab size to use for indenting.
func (w *Io_WKTWriter) SetTab(size int) {
	if size <= 0 {
		panic("Tab count must be positive")
	}
	w.indentTabStr = io_WKTWriter_StringOfChar(' ', size)
}

// SetOutputOrdinates sets the Ordinates that are to be written.
func (w *Io_WKTWriter) SetOutputOrdinates(outputOrdinates *Io_OrdinateSet) {
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

// GetOutputOrdinates gets a bit-pattern defining which ordinates should be written.
func (w *Io_WKTWriter) GetOutputOrdinates() *Io_OrdinateSet {
	return w.outputOrdinates
}

// SetPrecisionModel sets a PrecisionModel that should be used on the ordinates written.
func (w *Io_WKTWriter) SetPrecisionModel(precisionModel *Geom_PrecisionModel) {
	w.precisionModel = precisionModel
	w.ordinateFormat = Io_OrdinateFormat_Create(precisionModel.GetMaximumSignificantDigits())
}

// Write converts a Geometry to its Well-known Text representation.
func (w *Io_WKTWriter) Write(geometry *Geom_Geometry) string {
	var sb strings.Builder
	w.writeFormatted(geometry, false, &sb)
	return sb.String()
}

// WriteToWriter converts a Geometry to its Well-known Text representation and
// writes it to the given writer.
func (w *Io_WKTWriter) WriteToWriter(geometry *Geom_Geometry, writer io.Writer) error {
	return w.writeFormatted(geometry, w.isFormatted, writer)
}

// WriteFormatted is the same as Write, but with newlines and spaces to make
// the well-known text more readable.
func (w *Io_WKTWriter) WriteFormatted(geometry *Geom_Geometry) string {
	var sb strings.Builder
	w.writeFormatted(geometry, true, &sb)
	return sb.String()
}

// WriteFormattedToWriter is the same as WriteToWriter, but with newlines and
// spaces to make the well-known text more readable.
func (w *Io_WKTWriter) WriteFormattedToWriter(geometry *Geom_Geometry, writer io.Writer) error {
	return w.writeFormatted(geometry, true, writer)
}

func (w *Io_WKTWriter) writeFormatted(geometry *Geom_Geometry, useFormatting bool, writer io.Writer) error {
	formatter := w.getFormatter(geometry)
	return w.appendGeometryTaggedText(geometry, useFormatting, writer, formatter)
}

func (w *Io_WKTWriter) getFormatter(geometry *Geom_Geometry) *Io_OrdinateFormat {
	if w.ordinateFormat != nil {
		return w.ordinateFormat
	}
	pm := geometry.GetPrecisionModel()
	return io_WKTWriter_CreateFormatter(pm)
}

func (w *Io_WKTWriter) appendGeometryTaggedText(geometry *Geom_Geometry, useFormatting bool, writer io.Writer, formatter *Io_OrdinateFormat) error {
	cof := io_NewCheckOrdinatesFilter(w.outputOrdinates)
	geometry.ApplyCoordinateSequenceFilter(cof)
	return w.appendGeometryTaggedTextWithOrdinates(geometry, cof.GetOutputOrdinates(), useFormatting, 0, writer, formatter)
}

func (w *Io_WKTWriter) appendGeometryTaggedTextWithOrdinates(geometry *Geom_Geometry, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if err := w.indent(useFormatting, level, writer); err != nil {
		return err
	}

	self := java.GetLeaf(geometry)

	switch g := self.(type) {
	case *Geom_Point:
		return w.appendPointTaggedText(g, outputOrdinates, useFormatting, level, writer, formatter)
	case *Geom_LinearRing:
		return w.appendLinearRingTaggedText(g, outputOrdinates, useFormatting, level, writer, formatter)
	case *Geom_LineString:
		return w.appendLineStringTaggedText(g, outputOrdinates, useFormatting, level, writer, formatter)
	case *Geom_Polygon:
		return w.appendPolygonTaggedText(g, outputOrdinates, useFormatting, level, writer, formatter)
	case *Geom_MultiPoint:
		return w.appendMultiPointTaggedText(g, outputOrdinates, useFormatting, level, writer, formatter)
	case *Geom_MultiLineString:
		return w.appendMultiLineStringTaggedText(g, outputOrdinates, useFormatting, level, writer, formatter)
	case *Geom_MultiPolygon:
		return w.appendMultiPolygonTaggedText(g, outputOrdinates, useFormatting, level, writer, formatter)
	case *Geom_GeometryCollection:
		return w.appendGeometryCollectionTaggedText(g, outputOrdinates, useFormatting, level, writer, formatter)
	default:
		panic("Unsupported Geometry implementation")
	}
}

func (w *Io_WKTWriter) appendPointTaggedText(point *Geom_Point, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if _, err := io.WriteString(writer, Io_WKTConstants_POINT); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, " "); err != nil {
		return err
	}
	if err := w.appendOrdinateText(outputOrdinates, writer); err != nil {
		return err
	}
	return w.appendSequenceText(point.GetCoordinateSequence(), outputOrdinates, useFormatting, level, false, writer, formatter)
}

func (w *Io_WKTWriter) appendLineStringTaggedText(lineString *Geom_LineString, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if _, err := io.WriteString(writer, Io_WKTConstants_LINESTRING); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, " "); err != nil {
		return err
	}
	if err := w.appendOrdinateText(outputOrdinates, writer); err != nil {
		return err
	}
	return w.appendSequenceText(lineString.GetCoordinateSequence(), outputOrdinates, useFormatting, level, false, writer, formatter)
}

func (w *Io_WKTWriter) appendLinearRingTaggedText(linearRing *Geom_LinearRing, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if _, err := io.WriteString(writer, Io_WKTConstants_LINEARRING); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, " "); err != nil {
		return err
	}
	if err := w.appendOrdinateText(outputOrdinates, writer); err != nil {
		return err
	}
	return w.appendSequenceText(linearRing.GetCoordinateSequence(), outputOrdinates, useFormatting, level, false, writer, formatter)
}

func (w *Io_WKTWriter) appendPolygonTaggedText(polygon *Geom_Polygon, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if _, err := io.WriteString(writer, Io_WKTConstants_POLYGON); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, " "); err != nil {
		return err
	}
	if err := w.appendOrdinateText(outputOrdinates, writer); err != nil {
		return err
	}
	return w.appendPolygonText(polygon, outputOrdinates, useFormatting, level, false, writer, formatter)
}

func (w *Io_WKTWriter) appendMultiPointTaggedText(multipoint *Geom_MultiPoint, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if _, err := io.WriteString(writer, Io_WKTConstants_MULTIPOINT); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, " "); err != nil {
		return err
	}
	if err := w.appendOrdinateText(outputOrdinates, writer); err != nil {
		return err
	}
	return w.appendMultiPointText(multipoint, outputOrdinates, useFormatting, level, writer, formatter)
}

func (w *Io_WKTWriter) appendMultiLineStringTaggedText(multiLineString *Geom_MultiLineString, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if _, err := io.WriteString(writer, Io_WKTConstants_MULTILINESTRING); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, " "); err != nil {
		return err
	}
	if err := w.appendOrdinateText(outputOrdinates, writer); err != nil {
		return err
	}
	return w.appendMultiLineStringText(multiLineString, outputOrdinates, useFormatting, level, writer, formatter)
}

func (w *Io_WKTWriter) appendMultiPolygonTaggedText(multiPolygon *Geom_MultiPolygon, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if _, err := io.WriteString(writer, Io_WKTConstants_MULTIPOLYGON); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, " "); err != nil {
		return err
	}
	if err := w.appendOrdinateText(outputOrdinates, writer); err != nil {
		return err
	}
	return w.appendMultiPolygonText(multiPolygon, outputOrdinates, useFormatting, level, writer, formatter)
}

func (w *Io_WKTWriter) appendGeometryCollectionTaggedText(geometryCollection *Geom_GeometryCollection, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if _, err := io.WriteString(writer, Io_WKTConstants_GEOMETRYCOLLECTION); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, " "); err != nil {
		return err
	}
	if err := w.appendOrdinateText(outputOrdinates, writer); err != nil {
		return err
	}
	return w.appendGeometryCollectionText(geometryCollection, outputOrdinates, useFormatting, level, writer, formatter)
}

func (w *Io_WKTWriter) appendCoordinate(seq Geom_CoordinateSequence, outputOrdinates *Io_OrdinateSet, i int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if _, err := io.WriteString(writer, io_WKTWriter_WriteNumber(seq.GetX(i), formatter)+" "+io_WKTWriter_WriteNumber(seq.GetY(i), formatter)); err != nil {
		return err
	}

	if outputOrdinates.Contains(Io_Ordinate_Z) {
		if _, err := io.WriteString(writer, " "); err != nil {
			return err
		}
		if _, err := io.WriteString(writer, io_WKTWriter_WriteNumber(seq.GetZ(i), formatter)); err != nil {
			return err
		}
	}

	if outputOrdinates.Contains(Io_Ordinate_M) {
		if _, err := io.WriteString(writer, " "); err != nil {
			return err
		}
		if _, err := io.WriteString(writer, io_WKTWriter_WriteNumber(seq.GetM(i), formatter)); err != nil {
			return err
		}
	}

	return nil
}

func io_WKTWriter_WriteNumber(d float64, formatter *Io_OrdinateFormat) string {
	return formatter.Format(d)
}

func (w *Io_WKTWriter) appendOrdinateText(outputOrdinates *Io_OrdinateSet, writer io.Writer) error {
	if outputOrdinates.Contains(Io_Ordinate_Z) {
		if _, err := io.WriteString(writer, Io_WKTConstants_Z); err != nil {
			return err
		}
	}
	if outputOrdinates.Contains(Io_Ordinate_M) {
		if _, err := io.WriteString(writer, Io_WKTConstants_M); err != nil {
			return err
		}
	}
	return nil
}

func (w *Io_WKTWriter) appendSequenceText(seq Geom_CoordinateSequence, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, indentFirst bool, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if seq.Size() == 0 {
		_, err := io.WriteString(writer, Io_WKTConstants_EMPTY)
		return err
	}

	if indentFirst {
		if err := w.indent(useFormatting, level, writer); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(writer, "("); err != nil {
		return err
	}
	for i := 0; i < seq.Size(); i++ {
		if i > 0 {
			if _, err := io.WriteString(writer, ", "); err != nil {
				return err
			}
			if w.coordsPerLine > 0 && i%w.coordsPerLine == 0 {
				if err := w.indent(useFormatting, level+1, writer); err != nil {
					return err
				}
			}
		}
		if err := w.appendCoordinate(seq, outputOrdinates, i, writer, formatter); err != nil {
			return err
		}
	}
	_, err := io.WriteString(writer, ")")
	return err
}

func (w *Io_WKTWriter) appendPolygonText(polygon *Geom_Polygon, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, indentFirst bool, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if polygon.IsEmpty() {
		_, err := io.WriteString(writer, Io_WKTConstants_EMPTY)
		return err
	}

	if indentFirst {
		if err := w.indent(useFormatting, level, writer); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(writer, "("); err != nil {
		return err
	}
	if err := w.appendSequenceText(polygon.GetExteriorRing().GetCoordinateSequence(), outputOrdinates, useFormatting, level, false, writer, formatter); err != nil {
		return err
	}
	for i := 0; i < polygon.GetNumInteriorRing(); i++ {
		if _, err := io.WriteString(writer, ", "); err != nil {
			return err
		}
		if err := w.appendSequenceText(polygon.GetInteriorRingN(i).GetCoordinateSequence(), outputOrdinates, useFormatting, level+1, true, writer, formatter); err != nil {
			return err
		}
	}
	_, err := io.WriteString(writer, ")")
	return err
}

func (w *Io_WKTWriter) appendMultiPointText(multiPoint *Geom_MultiPoint, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if multiPoint.GetNumGeometries() == 0 {
		_, err := io.WriteString(writer, Io_WKTConstants_EMPTY)
		return err
	}

	if _, err := io.WriteString(writer, "("); err != nil {
		return err
	}
	for i := 0; i < multiPoint.GetNumGeometries(); i++ {
		if i > 0 {
			if _, err := io.WriteString(writer, ", "); err != nil {
				return err
			}
			if err := w.indentCoords(useFormatting, i, level+1, writer); err != nil {
				return err
			}
		}
		point := java.Cast[*Geom_Point](multiPoint.GetGeometryN(i))
		if err := w.appendSequenceText(point.GetCoordinateSequence(), outputOrdinates, useFormatting, level, false, writer, formatter); err != nil {
			return err
		}
	}
	_, err := io.WriteString(writer, ")")
	return err
}

func (w *Io_WKTWriter) appendMultiLineStringText(multiLineString *Geom_MultiLineString, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if multiLineString.GetNumGeometries() == 0 {
		_, err := io.WriteString(writer, Io_WKTConstants_EMPTY)
		return err
	}

	level2 := level
	doIndent := false
	if _, err := io.WriteString(writer, "("); err != nil {
		return err
	}
	for i := 0; i < multiLineString.GetNumGeometries(); i++ {
		if i > 0 {
			if _, err := io.WriteString(writer, ", "); err != nil {
				return err
			}
			level2 = level + 1
			doIndent = true
		}
		lineString := java.Cast[*Geom_LineString](multiLineString.GetGeometryN(i))
		if err := w.appendSequenceText(lineString.GetCoordinateSequence(), outputOrdinates, useFormatting, level2, doIndent, writer, formatter); err != nil {
			return err
		}
	}
	_, err := io.WriteString(writer, ")")
	return err
}

func (w *Io_WKTWriter) appendMultiPolygonText(multiPolygon *Geom_MultiPolygon, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if multiPolygon.GetNumGeometries() == 0 {
		_, err := io.WriteString(writer, Io_WKTConstants_EMPTY)
		return err
	}

	level2 := level
	doIndent := false
	if _, err := io.WriteString(writer, "("); err != nil {
		return err
	}
	for i := 0; i < multiPolygon.GetNumGeometries(); i++ {
		if i > 0 {
			if _, err := io.WriteString(writer, ", "); err != nil {
				return err
			}
			level2 = level + 1
			doIndent = true
		}
		polygon := java.Cast[*Geom_Polygon](multiPolygon.GetGeometryN(i))
		if err := w.appendPolygonText(polygon, outputOrdinates, useFormatting, level2, doIndent, writer, formatter); err != nil {
			return err
		}
	}
	_, err := io.WriteString(writer, ")")
	return err
}

func (w *Io_WKTWriter) appendGeometryCollectionText(geometryCollection *Geom_GeometryCollection, outputOrdinates *Io_OrdinateSet, useFormatting bool, level int, writer io.Writer, formatter *Io_OrdinateFormat) error {
	if geometryCollection.GetNumGeometries() == 0 {
		_, err := io.WriteString(writer, Io_WKTConstants_EMPTY)
		return err
	}

	level2 := level
	if _, err := io.WriteString(writer, "("); err != nil {
		return err
	}
	for i := 0; i < geometryCollection.GetNumGeometries(); i++ {
		if i > 0 {
			if _, err := io.WriteString(writer, ", "); err != nil {
				return err
			}
			level2 = level + 1
		}
		if err := w.appendGeometryTaggedTextWithOrdinates(geometryCollection.GetGeometryN(i), outputOrdinates, useFormatting, level2, writer, formatter); err != nil {
			return err
		}
	}
	_, err := io.WriteString(writer, ")")
	return err
}

func (w *Io_WKTWriter) indentCoords(useFormatting bool, coordIndex int, level int, writer io.Writer) error {
	if w.coordsPerLine <= 0 || coordIndex%w.coordsPerLine != 0 {
		return nil
	}
	return w.indent(useFormatting, level, writer)
}

func (w *Io_WKTWriter) indent(useFormatting bool, level int, writer io.Writer) error {
	if !useFormatting || level <= 0 {
		return nil
	}
	if _, err := io.WriteString(writer, "\n"); err != nil {
		return err
	}
	for i := 0; i < level; i++ {
		if _, err := io.WriteString(writer, w.indentTabStr); err != nil {
			return err
		}
	}
	return nil
}
