package geom_test

// This helper lives here rather than in internal/test because it references
// geom.ValidationError and geom.RuleViolation, which are only exported via
// geom/export_test.go (the standard Go pattern for test-only exports). A
// regular (non-test) package like internal/test cannot see those symbols.

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
)

func expectRuleViolation(tb testing.TB, g interface{ Validate() error }, rv geom.RuleViolation) {
	tb.Helper()
	err := g.Validate()
	var ve geom.ValidationError
	test.ErrAs(tb, err, &ve)
	test.Eq(tb, string(ve.RuleViolation), string(rv))
}
