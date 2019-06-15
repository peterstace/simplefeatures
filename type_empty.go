package simplefeatures

// EmptySet is a 0-dimensional geometry that represents the empty pointset.
type EmptySet struct {
	wkt string
}

func NewEmptyPoint() EmptySet {
	return EmptySet{"POINT EMPTY"}
}

func NewEmptyLineString() EmptySet {
	return EmptySet{"POINTSTRING EMPTY"}
}

func NewEmptyPolygon() EmptySet {
	return EmptySet{"POLYGON EMPTY"}
}

func (e EmptySet) AsText() []byte {
	return []byte(e.wkt)
}

func (e EmptySet) AppendWKT(dst []byte) []byte {
	return append(dst, e.wkt...)
}

func (e EmptySet) IsSimple() bool {
	return true
}

func (e EmptySet) Intersection(Geometry) Geometry {
	// TODO: global intersection dispatch?
	return NewGeometryCollection(nil)
}

func (e EmptySet) IsEmpty() bool {
	return true
}

func (e EmptySet) Dimension() int {
	return 0
}
