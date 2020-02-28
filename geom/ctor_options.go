package geom

// ConstructorOption allows the behaviour of Geometry constructor functions to be modified.
type ConstructorOption func(o *ctorOptionSet)

// DisableAllValidations causes geometry constructors to skip all validations.
// This allows invalid geometries to be loaded, but also has several
// implications for using the resultant geometries:
//
// 1. If the geometry is invalid, then any geometric calculations resulting
//    from the geometry may be invalid.
//
// 2. If the geometry is invalid, then invoking geometric calculations may
//    cause a panic or infinite loop.
//
// This option should be used with caution. It is most useful when invalid
// geometries need to be loaded, but no geometric calculations will be
// performed.
func DisableAllValidations(o *ctorOptionSet) {
	o.skipValidations = true
}

type ctorOptionSet struct {
	skipValidations bool
}

func newOptionSet(opts []ConstructorOption) ctorOptionSet {
	// Optimise the case where there are no options. This prevents the `cos`
	// variable escaping to the heap.
	if len(opts) == 0 {
		return ctorOptionSet{}
	}

	var cos ctorOptionSet
	for _, opt := range opts {
		opt(&cos)
	}
	return cos
}

func skipValidations(opts []ConstructorOption) bool {
	os := newOptionSet(opts)
	return os.skipValidations
}
