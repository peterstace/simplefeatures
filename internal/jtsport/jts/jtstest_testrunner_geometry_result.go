package jts

var _ JtstestTestrunner_Result = (*JtstestTestrunner_GeometryResult)(nil)

type JtstestTestrunner_GeometryResult struct {
	geometry *Geom_Geometry
}

func JtstestTestrunner_NewGeometryResult(geometry *Geom_Geometry) *JtstestTestrunner_GeometryResult {
	return &JtstestTestrunner_GeometryResult{geometry: geometry}
}

func (r *JtstestTestrunner_GeometryResult) IsJtstestTestrunner_Result() {}

func (r *JtstestTestrunner_GeometryResult) GetGeometry() *Geom_Geometry {
	return r.geometry
}

func (r *JtstestTestrunner_GeometryResult) EqualsResult(other JtstestTestrunner_Result, tolerance float64) bool {
	otherGeometryResult, ok := other.(*JtstestTestrunner_GeometryResult)
	if !ok {
		return false
	}
	otherGeometry := otherGeometryResult.geometry

	thisGeometryClone := r.geometry.Copy()
	otherGeometryClone := otherGeometry.Copy()
	thisGeometryClone.Normalize()
	otherGeometryClone.Normalize()
	return thisGeometryClone.EqualsExactWithTolerance(otherGeometryClone, tolerance)
}

func (r *JtstestTestrunner_GeometryResult) ToLongString() string {
	return r.geometry.ToText()
}

func (r *JtstestTestrunner_GeometryResult) ToFormattedString() string {
	writer := Io_NewWKTWriter()
	return writer.WriteFormatted(r.geometry)
}

func (r *JtstestTestrunner_GeometryResult) ToShortString() string {
	// TRANSLITERATION NOTE: Java returns geometry.getClass().getName() which
	// returns the full qualified class name. Go's GetGeometryType() returns
	// just the geometry type name (e.g., "Polygon" not
	// "org.locationtech.jts.geom.Polygon"). For test output compatibility, we
	// prefix with the package structure.
	geomType := r.geometry.GetGeometryType()
	return "org.locationtech.jts.geom." + geomType
}
