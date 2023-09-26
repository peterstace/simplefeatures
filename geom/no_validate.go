package geom

// NoValidate causes functions to skip geometry constraint validation.
// Functions where validation can be skipped accept a variadic list of
// NoValidate values. If at least one NoValidate value is passed in, then the
// function will skip validation, otherwise it will perform validation as its
// default behaviour.
//
// NoValidate is just an empty struct type, so can be passed in as
// NoValidate{}.
//
// Some algorithms implemented in simplefeatures rely on valid geometries to
// operate correctly. If invalid geometries are supplied, then the results may
// not be correct.
type NoValidate struct{}
