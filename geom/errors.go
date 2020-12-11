package geom

// SyntaxError indicates an error in the structural representation of a
// serialised geometry.
type SyntaxError struct {
	reason string
}

// Error gives the error text of the syntax error.
func (e SyntaxError) Error() string {
	// TODO: prefix with "syntax error:"
	return e.reason
}

// TopologyError indicates an error with the topological structure of a
// geometry.
type TopologyError struct {
	reason string
}

// Error gives the error text of the topology error.
func (e TopologyError) Error() string {
	return "topological error: " + e.reason
}
