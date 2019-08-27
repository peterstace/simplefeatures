package geom

// ConstructorOption allows the behaviour of Geometry constructor functions to be modified.
type ConstructorOption func(o *ctorOptionSet)

// DisableAllValidations causes geometry constructors to skip all validations.
// This allows invalid geometries to be loaded, but also has several
// implications for using the resultant geomerties:
//
// 1. If the geometry is invalid, then any geometric calculations resulting
//    from the geometry may be invalid.
//
// 2. If the geometry is invalid, then invoking geometric calculations may
//    cause a panic or infinite loop (this is a theoretical concern that
//    hasn't yet been observed in practice).
//
// This option should be used with caution, but can be safely used with
// geometries that are known to be valid a priori. It can also be used with
// invalid geometries for cases where no geometric calculations need to be
// performed.
func DisableAllValidations(o *ctorOptionSet) {
	o.skipAllValidations = true
}

// DisableExpensiveValidations gives a hint that geometry constructors may opt
// to skip any expensive validations. All of the caveats that come with the
// DisableAllValidations option also come with this option.
func DisableExpensiveValidations(o *ctorOptionSet) {
	o.skipExpensiveValidations = true
}

type ctorOptionSet struct {
	skipExpensiveValidations bool
	skipAllValidations       bool
}

func newOptionSet(opts []ConstructorOption) ctorOptionSet {
	var cos ctorOptionSet
	for _, opt := range opts {
		opt(&cos)
	}
	return cos
}

func doExpensiveValidations(opts []ConstructorOption) bool {
	os := newOptionSet(opts)
	return !os.skipExpensiveValidations && !os.skipAllValidations
}

func doCheapValidations(opts []ConstructorOption) bool {
	return !newOptionSet(opts).skipAllValidations
}
