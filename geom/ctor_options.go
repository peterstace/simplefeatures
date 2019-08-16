package geom

// ConstructorOption allows the behaviour of Geometry constructor functions to be modified.
type ConstructorOption func(o *ctorOptionSet)

// DisableExpensiveValidations gives a hint that geometry constructors may opt
// to skip any expensive validations. There are several implications as a
// result of using this option:
//
// 1. If the geometry is invalid, then any geometric calculations resulting
//    from the geometry may be invalid.
//
// 2. If the geometry is invalid, then invoking geometric calculations may
//    cause a panic or infinite loop (this is a theoretical concern that
//    hasn't yet been observed in practice).
//
// This option should be used with caution, but can be safely used when the
// geometry is known to be valid a priori.
func DisableExpensiveValidations(o *ctorOptionSet) {
	o.skipExpensiveValidations = true
}

type ctorOptionSet struct {
	skipExpensiveValidations bool
}

func newOptionSet(opts []ConstructorOption) ctorOptionSet {
	var cos ctorOptionSet
	for _, opt := range opts {
		opt(&cos)
	}
	return cos
}

func doExpensiveValidations(opts []ConstructorOption) bool {
	return !newOptionSet(opts).skipExpensiveValidations
}
