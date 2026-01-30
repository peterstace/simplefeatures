package jts

// Codes which combine a geometry dimension and a location on the geometry.

const OperationRelateng_DimensionLocation_EXTERIOR = Geom_Location_Exterior
const OperationRelateng_DimensionLocation_POINT_INTERIOR = 103
const OperationRelateng_DimensionLocation_LINE_INTERIOR = 110
const OperationRelateng_DimensionLocation_LINE_BOUNDARY = 111
const OperationRelateng_DimensionLocation_AREA_INTERIOR = 120
const OperationRelateng_DimensionLocation_AREA_BOUNDARY = 121

// OperationRelateng_DimensionLocation_LocationArea returns the dimension-location
// code for an area geometry at the given location.
func OperationRelateng_DimensionLocation_LocationArea(loc int) int {
	switch loc {
	case Geom_Location_Interior:
		return OperationRelateng_DimensionLocation_AREA_INTERIOR
	case Geom_Location_Boundary:
		return OperationRelateng_DimensionLocation_AREA_BOUNDARY
	}
	return OperationRelateng_DimensionLocation_EXTERIOR
}

// OperationRelateng_DimensionLocation_LocationLine returns the dimension-location
// code for a line geometry at the given location.
func OperationRelateng_DimensionLocation_LocationLine(loc int) int {
	switch loc {
	case Geom_Location_Interior:
		return OperationRelateng_DimensionLocation_LINE_INTERIOR
	case Geom_Location_Boundary:
		return OperationRelateng_DimensionLocation_LINE_BOUNDARY
	}
	return OperationRelateng_DimensionLocation_EXTERIOR
}

// OperationRelateng_DimensionLocation_LocationPoint returns the dimension-location
// code for a point geometry at the given location.
func OperationRelateng_DimensionLocation_LocationPoint(loc int) int {
	switch loc {
	case Geom_Location_Interior:
		return OperationRelateng_DimensionLocation_POINT_INTERIOR
	}
	return OperationRelateng_DimensionLocation_EXTERIOR
}

// OperationRelateng_DimensionLocation_Location extracts the location from a
// dimension-location code.
func OperationRelateng_DimensionLocation_Location(dimLoc int) int {
	switch dimLoc {
	case OperationRelateng_DimensionLocation_POINT_INTERIOR,
		OperationRelateng_DimensionLocation_LINE_INTERIOR,
		OperationRelateng_DimensionLocation_AREA_INTERIOR:
		return Geom_Location_Interior
	case OperationRelateng_DimensionLocation_LINE_BOUNDARY,
		OperationRelateng_DimensionLocation_AREA_BOUNDARY:
		return Geom_Location_Boundary
	}
	return Geom_Location_Exterior
}

// OperationRelateng_DimensionLocation_Dimension extracts the dimension from a
// dimension-location code.
func OperationRelateng_DimensionLocation_Dimension(dimLoc int) int {
	switch dimLoc {
	case OperationRelateng_DimensionLocation_POINT_INTERIOR:
		return Geom_Dimension_P
	case OperationRelateng_DimensionLocation_LINE_INTERIOR,
		OperationRelateng_DimensionLocation_LINE_BOUNDARY:
		return Geom_Dimension_L
	case OperationRelateng_DimensionLocation_AREA_INTERIOR,
		OperationRelateng_DimensionLocation_AREA_BOUNDARY:
		return Geom_Dimension_A
	}
	return Geom_Dimension_False
}

// OperationRelateng_DimensionLocation_DimensionWithExterior extracts the dimension
// from a dimension-location code, using exteriorDim for EXTERIOR locations.
func OperationRelateng_DimensionLocation_DimensionWithExterior(dimLoc int, exteriorDim int) int {
	if dimLoc == OperationRelateng_DimensionLocation_EXTERIOR {
		return exteriorDim
	}
	return OperationRelateng_DimensionLocation_Dimension(dimLoc)
}
