package geom

import (
	"encoding/binary"
	"fmt"
	"math"
)

// MarshalTWKB accepts a geometry and generates the corresponding TWKB byte slice.
func MarshalTWKB(geom Geometry,
	hasZ, hasM bool,
	precXY, precZ, precM int,
	hasSize, hasBBox, closeRings bool,
	idList []int64,
) ([]byte, error) {
	w := newTWKBWriter(hasZ, hasM, precXY, precZ, precM, hasSize, hasBBox, closeRings, idList)
	if err := w.writeGeometry(geom); err != nil {
		return nil, err
	}
	return w.twkb, nil
}

// TWKBWriter holds all state information needed for generating TWKB data
// including information such as the last reference point used in coord deltas.
type TWKBWriter struct {
	twkb []byte

	kind  twkbGeometryType
	ctype CoordinatesType

	dimensions int
	precXY     int
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

	refpoint [twkbMaxDimensions]int64

	bboxValid bool
	bboxMin   [twkbMaxDimensions]int64
	bboxMax   [twkbMaxDimensions]int64

	idList []int64

	closeRings bool
}

func newTWKBWriter(
	hasZ, hasM bool,
	precXY, precZ, precM int,
	hasSize, hasBBox, closeRings bool,
	idList []int64,
) *TWKBWriter {
	w := TWKBWriter{
		hasSize:    hasSize,
		hasBBox:    hasBBox,
		closeRings: closeRings,
	}

	w.precXY = precXY
	w.scalings[0] = math.Pow10(precXY)
	w.scalings[1] = w.scalings[0]
	w.dimensions = 2

	if hasZ {
		w.hasZ = true
		w.precZ = precZ
		w.scalings[w.dimensions] = math.Pow10(precZ)
		w.dimensions++
	}

	if hasM {
		w.hasM = true
		w.precM = precM
		w.scalings[w.dimensions] = math.Pow10(precM)
		w.dimensions++
	}

	if hasZ && hasM {
		w.ctype = DimXYZM
	} else if hasZ {
		w.ctype = DimXYZ
	} else if hasM {
		w.ctype = DimXYM
	} else {
		w.ctype = DimXY
	}

	if hasZ || hasM {
		w.hasExt = true
	}

	if len(idList) > 0 {
		w.hasIDs = true
		w.idList = idList
	}

	return &w
}

func copyTWKBWriter(other *TWKBWriter) *TWKBWriter {
	return newTWKBWriter(
		other.hasZ,       // Assume child has same dimensionality as parent.
		other.hasM,       // Assume child has same dimensionality as parent.
		other.precXY,     // Same precision as in parent.
		other.precZ,      // Same precision as in parent.
		other.precM,      // Same precision as in parent.
		other.hasSize,    // If parent is using a size header, child should too.
		false,            // No bbox in sub-geometries.
		other.closeRings, // If parent requires closed rings, child should too.
		nil,              // No ID list in sub-geometries.
	)
}

func (w *TWKBWriter) writeGeometry(geom Geometry) error {
	if err := w.writeGeometryByType(geom); err != nil {
		return err
	}
	w.writeAdditionalHeaders()
	return nil
}

func (w *TWKBWriter) writeGeometryByType(geom Geometry) error {
	switch geom.gtype {
	case TypePoint:
		return w.writePoint(geom.MustAsPoint())
	case TypeLineString:
		return w.writeLineString(geom.MustAsLineString())
	case TypePolygon:
		return w.writePolygon(geom.MustAsPolygon())
	case TypeMultiPoint:
		return w.writeMultiPoint(geom.MustAsMultiPoint())
	case TypeMultiLineString:
		return w.writeMultiLineString(geom.MustAsMultiLineString())
	case TypeMultiPolygon:
		return w.writeMultiPolygon(geom.MustAsMultiPolygon())
	case TypeGeometryCollection:
		return w.writeGeometryCollection(geom.MustAsGeometryCollection())
	default:
		return fmt.Errorf("geometry has unsupported type %s", geom.gtype)
	}
}

func (w *TWKBWriter) writePoint(pt Point) error {
	w.writeTypeAndPrecision(twkbTypePoint)

	if ctype := pt.CoordinatesType(); ctype != w.ctype {
		return fmt.Errorf("mismatched coordinate dimensions got %s expected %s", ctype, w.ctype)
	}

	if pt.IsEmpty() {
		w.writeIsEmptyHeader()
		return nil
	}
	w.writeInitialHeaders()

	return w.writePointCoords(pt)
}

func (w *TWKBWriter) writePointCoords(pt Point) error {
	switch pt.CoordinatesType() {
	case DimXY:
		w.writePoints(1, pt.coords.XY.X, pt.coords.XY.Y)
	case DimXYZ:
		w.writePoints(1, pt.coords.XY.X, pt.coords.XY.Y, pt.coords.Z)
	case DimXYM:
		w.writePoints(1, pt.coords.XY.X, pt.coords.XY.Y, pt.coords.M)
	case DimXYZM:
		w.writePoints(1, pt.coords.XY.X, pt.coords.XY.Y, pt.coords.Z, pt.coords.M)
	default:
		return fmt.Errorf("point has unsupported type %s", pt.CoordinatesType())
	}
	return nil
}

func (w *TWKBWriter) writeLineString(ls LineString) error {
	w.writeTypeAndPrecision(twkbTypeLineString)

	if ctype := ls.CoordinatesType(); ctype != w.ctype {
		return fmt.Errorf("mismatched coordinate dimensions got %s expected %s", ctype, w.ctype)
	}

	if ls.IsEmpty() {
		w.writeIsEmptyHeader()
		return nil
	}
	w.writeInitialHeaders()

	return w.writeLineStringCoords(ls)
}

func (w *TWKBWriter) writeLineStringCoords(ls LineString) error {
	coords := ls.Coordinates()
	numPoints := coords.Length()
	w.writeUnsignedVarint(uint64(numPoints))
	w.writePointArray(numPoints, coords.floats)
	return nil
}

func (w *TWKBWriter) writeRing(ls LineString) error {
	coords := ls.Coordinates()
	numPoints := coords.Length()
	if !w.closeRings && numPoints >= 2 {
		numPoints-- // Omit the final point in the ring.
	}
	w.writeUnsignedVarint(uint64(numPoints))
	w.writePointArray(numPoints, coords.floats)
	return nil
}

func (w *TWKBWriter) writePolygon(poly Polygon) error {
	w.writeTypeAndPrecision(twkbTypePolygon)

	if ctype := poly.CoordinatesType(); ctype != w.ctype {
		return fmt.Errorf("mismatched coordinate dimensions got %s expected %s", ctype, w.ctype)
	}

	if poly.IsEmpty() {
		w.writeIsEmptyHeader()
		return nil
	}
	w.writeInitialHeaders()

	return w.writePolygonRings(poly)
}

func (w *TWKBWriter) writePolygonRings(poly Polygon) error {
	w.writeUnsignedVarint(uint64(poly.NumRings()))

	if poly.NumRings() == 0 {
		return nil
	}

	ls := poly.ExteriorRing()
	w.writeRing(ls)

	numRings := poly.NumInteriorRings()
	for i := 0; i < numRings; i++ {
		ls = poly.InteriorRingN(i)
		w.writeRing(ls)
	}
	return nil
}

func (w *TWKBWriter) writeMultiPoint(mp MultiPoint) error {
	w.writeTypeAndPrecision(twkbTypeMultiPoint)

	if ctype := mp.CoordinatesType(); ctype != w.ctype {
		return fmt.Errorf("mismatched coordinate dimensions got %s expected %s", ctype, w.ctype)
	}

	if mp.IsEmpty() {
		w.writeIsEmptyHeader()
		return nil
	}
	w.writeInitialHeaders()

	numPoints := mp.NumPoints()
	w.writeUnsignedVarint(uint64(numPoints))

	if err := w.writeIDList(numPoints); err != nil {
		return err
	}

	for i := 0; i < numPoints; i++ {
		pt := mp.PointN(i)
		w.writePointCoords(pt)
	}
	return nil
}

func (w *TWKBWriter) writeMultiLineString(ml MultiLineString) error {
	w.writeTypeAndPrecision(twkbTypeMultiLineString)

	if ctype := ml.CoordinatesType(); ctype != w.ctype {
		return fmt.Errorf("mismatched coordinate dimensions got %s expected %s", ctype, w.ctype)
	}

	if ml.IsEmpty() {
		w.writeIsEmptyHeader()
		return nil
	}
	w.writeInitialHeaders()

	numLineStrings := ml.NumLineStrings()
	w.writeUnsignedVarint(uint64(numLineStrings))

	if err := w.writeIDList(numLineStrings); err != nil {
		return err
	}

	for i := 0; i < numLineStrings; i++ {
		ls := ml.LineStringN(i)
		w.writeLineStringCoords(ls)
	}
	return nil
}

func (w *TWKBWriter) writeMultiPolygon(mp MultiPolygon) error {
	w.writeTypeAndPrecision(twkbTypeMultiPolygon)

	if ctype := mp.CoordinatesType(); ctype != w.ctype {
		return fmt.Errorf("mismatched coordinate dimensions got %s expected %s", ctype, w.ctype)
	}

	if mp.IsEmpty() {
		w.writeIsEmptyHeader()
		return nil
	}
	w.writeInitialHeaders()

	numPolygons := mp.NumPolygons()
	w.writeUnsignedVarint(uint64(numPolygons))

	if err := w.writeIDList(numPolygons); err != nil {
		return err
	}

	for i := 0; i < numPolygons; i++ {
		poly := mp.PolygonN(i)
		w.writePolygonRings(poly)
	}
	return nil
}

func (w *TWKBWriter) writeGeometryCollection(gc GeometryCollection) error {
	w.writeTypeAndPrecision(twkbTypeGeometryCollection)

	if ctype := gc.CoordinatesType(); ctype != w.ctype {
		return fmt.Errorf("mismatched coordinate dimensions got %s expected %s", ctype, w.ctype)
	}

	if gc.IsEmpty() {
		w.writeIsEmptyHeader()
		return nil
	}
	w.writeInitialHeaders()

	numGeometries := gc.NumGeometries()
	w.writeUnsignedVarint(uint64(numGeometries))

	if err := w.writeIDList(numGeometries); err != nil {
		return err
	}

	for i := 0; i < numGeometries; i++ {
		subWriter := copyTWKBWriter(w)
		geom := gc.GeometryN(i)
		subWriter.writeGeometry(geom)
		w.twkb = append(w.twkb, subWriter.twkb...)
	}
	return nil
}

func (w *TWKBWriter) writeTypeAndPrecision(kind twkbGeometryType) {
	w.kind = kind
	w.writeByte(byte(EncodeZigZagInt32(int32(w.precXY))<<4) | byte(w.kind))
}

func (w *TWKBWriter) writeIsEmptyHeader() {
	w.isEmpty = true
	w.writeMetadataHeader(twkbIsEmpty)
	// Do not write any extended info, bbox, size, or ids.
}

func (w *TWKBWriter) writeInitialHeaders() {
	var metaheader twkbMetadataHeader
	if w.hasExt {
		metaheader |= twkbHasExtPrec
	}
	if w.hasSize {
		metaheader |= twkbHasSize
	}
	if w.hasBBox {
		metaheader |= twkbHasBBox
	}
	if w.hasIDs {
		metaheader |= twkbHasIDs
	}
	w.writeMetadataHeader(metaheader)

	if w.hasExt {
		w.writeExtendedPrecision()
	}
}

func (w *TWKBWriter) writeMetadataHeader(metaheader twkbMetadataHeader) {
	if metaheader&twkbIsEmpty != 0 {
		w.isEmpty = true
	}
	w.writeByte(byte(metaheader))
}

func (w *TWKBWriter) writeExtendedPrecision() {
	var ext byte
	if w.hasZ {
		ext |= 0x01
		ext |= byte(EncodeZigZagInt32(int32(w.precZ)) << 2)
	}
	if w.hasM {
		ext |= 0x02
		ext |= byte(EncodeZigZagInt32(int32(w.precM)) << 5)
	}
	w.writeByte(ext)
}

func (w *TWKBWriter) writePoints(numPoints int, coords ...float64) {
	w.writePointArray(numPoints, coords)
}

// Convert a given number of points from floating point to integer coordinates.
// Utilise and update the running memory of the previous reference point.
// The input coords must contain numPoints * the number of dimensions values.
func (w *TWKBWriter) writePointArray(numPoints int, coords []float64) {
	var buf [binary.MaxVarintLen64]byte
	c := 0
	for i := 0; i < numPoints; i++ {
		for d := 0; d < w.dimensions; d++ {
			fval := coords[c]
			ival := int64(fval * w.scalings[d])
			// Compute bounding box.
			if !w.bboxValid {
				w.bboxMin[d] = ival
				w.bboxMax[d] = ival
			} else if ival < w.bboxMin[d] {
				w.bboxMin[d] = ival
			} else if ival > w.bboxMax[d] {
				w.bboxMax[d] = ival
			}
			// Perform coord differencing to find the int value.
			ival -= w.refpoint[d]
			n := binary.PutVarint(buf[:], ival)

			w.twkb = append(w.twkb, buf[:n]...)
			w.refpoint[d] += ival
			c++
		}
		if !w.bboxValid {
			w.bboxValid = true
		}
	}
}

func (w *TWKBWriter) writeAdditionalHeaders() {
	// These are written in this order so that the size of the
	// bbox is included in the size computation.
	if w.hasBBox {
		w.writeBBox()
	}
	if w.hasSize {
		w.writeSize()
	}
}

func (w *TWKBWriter) writeBBox() {
	// Store bbox min and delta for each dimension.
	var buf [twkbMaxDimensions * 2 * binary.MaxVarintLen64]byte
	n := 0
	for d := 0; d < w.dimensions; d++ {
		n += binary.PutVarint(buf[n:], w.bboxMin[d])
		n += binary.PutVarint(buf[n:], w.bboxMax[d]-w.bboxMin[d])
	}
	// Build a new TWKB buffer with the bbox inserted after the header bytes.
	start := 2
	if w.hasExt {
		start++ // Ensure extended precision byte remains before the bbox values.
	}
	// Insert the bbox data.
	var temp []byte
	temp = append(temp, w.twkb[:start]...)
	temp = append(temp, buf[:n]...)
	temp = append(temp, w.twkb[start:]...)
	w.twkb = temp
}

func (w *TWKBWriter) writeSize() {
	// Compute where to store the size data.
	start := 2
	if w.hasExt {
		start++ // Ensure extended precision byte remains before the size value.
	}
	numBytes := len(w.twkb) - start
	if numBytes < 0 {
		panic("attempt to add size value to buffer lacking TWKB header bytes")
	}

	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], uint64(numBytes))

	// Insert the size data.
	var temp []byte
	temp = append(temp, w.twkb[:start]...)
	temp = append(temp, buf[:n]...)
	temp = append(temp, w.twkb[start:]...)
	w.twkb = temp
}

func (w *TWKBWriter) writeIDList(num int) error {
	if !w.hasIDs {
		return nil
	}
	if num != len(w.idList) {
		return fmt.Errorf("unexpected ID list length %d, expected %d", len(w.idList), num)
	}
	for i := 0; i < num; i++ {
		w.writeSignedVarint(w.idList[i])
	}
	return nil
}

func (w *TWKBWriter) writeSignedVarint(val int64) int {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], val)
	w.twkb = append(w.twkb, buf[:n]...)
	return n
}

func (w *TWKBWriter) writeUnsignedVarint(val uint64) int {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], val)
	w.twkb = append(w.twkb, buf[:n]...)
	return n
}

func (w *TWKBWriter) writeByte(b byte) {
	w.twkb = append(w.twkb, b)
}
