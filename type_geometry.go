package simplefeatures

type Geometry interface {
	// AsText returns the WKT representation of the geometry.
	AsText() []byte

	// AppendWKT appends the WKT representation of the geometry to dst and
	// returns the resultant slice.
	AppendWKT(dst []byte) []byte
}
