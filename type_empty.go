package simplefeatures

// EmptySet is a 0-dimensional geometry that represents the empty pointset.
type EmptySet struct {
	wkt string
}

func NewEmptyPoint() EmptySet {
	return EmptySet{"POINT EMPTY"}
}

func NewEmptyLineString() EmptySet {
	return EmptySet{"LINESTRING EMPTY"}
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

func (e EmptySet) Intersection(g Geometry) Geometry {
	return intersection(e, g)
}

func (e EmptySet) IsEmpty() bool {
	return true
}

func (e EmptySet) Dimension() int {
	return 0
}

func (e EmptySet) Equals(other Geometry) bool {
	return equals(e, other)
}

func (e EmptySet) Envelope() (Envelope, bool) {
	return Envelope{}, false
}
