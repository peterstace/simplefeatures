package geom

// SyntaxError indicates an error in the structural representation of a
// serialised geometry.
type SyntaxError struct {
	reason string
}

// Error gives the error text of the syntax error.
func (e SyntaxError) Error() string {
	return e.reason
}

//type TopologyError struct {
//}
