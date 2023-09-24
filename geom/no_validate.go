package geom

// NoValidate instructs geometry constructors to skip geometry constraint
// validations. This allows simplefeatures to work with geometries that are
// invalid.
//
// Some algorithms implemented in simplefeatures rely on valid geometries to
// operate correctly. If invalid geometries are supplied, then the results may
// not be correct.
type NoValidate struct{}
