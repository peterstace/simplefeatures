package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// operationUnion_InputExtracter extracts atomic elements from input geometries
// or collections, recording the dimension found. Empty geometries are discarded
// since they do not contribute to the result of UnaryUnionOp.
type operationUnion_InputExtracter struct {
	geomFactory *Geom_GeometryFactory
	polygons    []*Geom_Polygon
	lines       []*Geom_LineString
	points      []*Geom_Point
	// The default dimension for an empty GeometryCollection.
	dimension int
}

var _ Geom_GeometryFilter = (*operationUnion_InputExtracter)(nil)

func (ie *operationUnion_InputExtracter) IsGeom_GeometryFilter() {}

// operationUnion_InputExtracter_ExtractFromCollection extracts elements from a
// collection of geometries.
func operationUnion_InputExtracter_ExtractFromCollection(geoms []*Geom_Geometry) *operationUnion_InputExtracter {
	extracter := operationUnion_NewInputExtracter()
	extracter.addCollection(geoms)
	return extracter
}

// operationUnion_InputExtracter_Extract extracts elements from a geometry.
func operationUnion_InputExtracter_Extract(geom *Geom_Geometry) *operationUnion_InputExtracter {
	extracter := operationUnion_NewInputExtracter()
	extracter.add(geom)
	return extracter
}

// operationUnion_NewInputExtracter creates a new InputExtracter.
func operationUnion_NewInputExtracter() *operationUnion_InputExtracter {
	return &operationUnion_InputExtracter{
		polygons:  make([]*Geom_Polygon, 0),
		lines:     make([]*Geom_LineString, 0),
		points:    make([]*Geom_Point, 0),
		dimension: Geom_Dimension_False,
	}
}

// IsEmpty tests whether there were any non-empty geometries extracted.
func (ie *operationUnion_InputExtracter) IsEmpty() bool {
	return len(ie.polygons) == 0 &&
		len(ie.lines) == 0 &&
		len(ie.points) == 0
}

// GetDimension gets the maximum dimension extracted.
func (ie *operationUnion_InputExtracter) GetDimension() int {
	return ie.dimension
}

// GetFactory gets the geometry factory from the extracted geometry, if there is
// one. If an empty collection was extracted, will return nil.
func (ie *operationUnion_InputExtracter) GetFactory() *Geom_GeometryFactory {
	return ie.geomFactory
}

// GetExtract gets the extracted atomic geometries of the given dimension dim.
func (ie *operationUnion_InputExtracter) GetExtract(dim int) []*Geom_Geometry {
	switch dim {
	case 0:
		result := make([]*Geom_Geometry, len(ie.points))
		for i, p := range ie.points {
			result[i] = p.Geom_Geometry
		}
		return result
	case 1:
		result := make([]*Geom_Geometry, len(ie.lines))
		for i, l := range ie.lines {
			result[i] = l.Geom_Geometry
		}
		return result
	case 2:
		result := make([]*Geom_Geometry, len(ie.polygons))
		for i, p := range ie.polygons {
			result[i] = p.Geom_Geometry
		}
		return result
	}
	Util_Assert_ShouldNeverReachHereWithMessage("Invalid dimension: " + string(rune('0'+dim)))
	return nil
}

func (ie *operationUnion_InputExtracter) addCollection(geoms []*Geom_Geometry) {
	for _, geom := range geoms {
		ie.add(geom)
	}
}

func (ie *operationUnion_InputExtracter) add(geom *Geom_Geometry) {
	if ie.geomFactory == nil {
		ie.geomFactory = geom.GetFactory()
	}
	geom.ApplyGeometryFilter(ie)
}

// Filter performs an operation with or on geom.
func (ie *operationUnion_InputExtracter) Filter(geom *Geom_Geometry) {
	ie.recordDimension(geom.GetDimension())

	if java.InstanceOf[*Geom_GeometryCollection](geom) {
		return
	}
	// Don't keep empty geometries.
	if geom.IsEmpty() {
		return
	}

	if java.InstanceOf[*Geom_Polygon](geom) {
		ie.polygons = append(ie.polygons, java.Cast[*Geom_Polygon](geom))
		return
	} else if java.InstanceOf[*Geom_LineString](geom) {
		ie.lines = append(ie.lines, java.Cast[*Geom_LineString](geom))
		return
	} else if java.InstanceOf[*Geom_Point](geom) {
		ie.points = append(ie.points, java.Cast[*Geom_Point](geom))
		return
	}
	Util_Assert_ShouldNeverReachHereWithMessage("Unhandled geometry type: " + geom.GetGeometryType())
}

func (ie *operationUnion_InputExtracter) recordDimension(dim int) {
	if dim > ie.dimension {
		ie.dimension = dim
	}
}
