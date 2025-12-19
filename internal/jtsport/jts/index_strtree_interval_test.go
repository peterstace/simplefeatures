package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestInterval_IntersectsBasic(t *testing.T) {
	junit.AssertTrue(t, jts.IndexStrtree_NewInterval(5, 10).Intersects(jts.IndexStrtree_NewInterval(7, 12)))
	junit.AssertTrue(t, jts.IndexStrtree_NewInterval(7, 12).Intersects(jts.IndexStrtree_NewInterval(5, 10)))
	junit.AssertTrue(t, !jts.IndexStrtree_NewInterval(5, 10).Intersects(jts.IndexStrtree_NewInterval(11, 12)))
	junit.AssertTrue(t, !jts.IndexStrtree_NewInterval(11, 12).Intersects(jts.IndexStrtree_NewInterval(5, 10)))
	junit.AssertTrue(t, jts.IndexStrtree_NewInterval(5, 10).Intersects(jts.IndexStrtree_NewInterval(10, 12)))
	junit.AssertTrue(t, jts.IndexStrtree_NewInterval(10, 12).Intersects(jts.IndexStrtree_NewInterval(5, 10)))
}

func TestInterval_IntersectsZeroWidthInterval(t *testing.T) {
	junit.AssertTrue(t, jts.IndexStrtree_NewInterval(10, 10).Intersects(jts.IndexStrtree_NewInterval(7, 12)))
	junit.AssertTrue(t, jts.IndexStrtree_NewInterval(7, 12).Intersects(jts.IndexStrtree_NewInterval(10, 10)))
	junit.AssertTrue(t, !jts.IndexStrtree_NewInterval(10, 10).Intersects(jts.IndexStrtree_NewInterval(11, 12)))
	junit.AssertTrue(t, !jts.IndexStrtree_NewInterval(11, 12).Intersects(jts.IndexStrtree_NewInterval(10, 10)))
	junit.AssertTrue(t, jts.IndexStrtree_NewInterval(10, 10).Intersects(jts.IndexStrtree_NewInterval(10, 12)))
	junit.AssertTrue(t, jts.IndexStrtree_NewInterval(10, 12).Intersects(jts.IndexStrtree_NewInterval(10, 10)))
}

func TestInterval_CopyConstructor(t *testing.T) {
	junit.AssertEqualsDeep(t, jts.IndexStrtree_NewInterval(3, 4), jts.IndexStrtree_NewInterval(3, 4))
	junit.AssertEqualsDeep(t, jts.IndexStrtree_NewInterval(3, 4), jts.IndexStrtree_NewIntervalFromInterval(jts.IndexStrtree_NewInterval(3, 4)))
}

func TestInterval_GetCentre(t *testing.T) {
	junit.AssertEquals(t, 6.5, jts.IndexStrtree_NewInterval(4, 9).GetCentre())
}

func TestInterval_ExpandToInclude(t *testing.T) {
	junit.AssertEqualsDeep(t, jts.IndexStrtree_NewInterval(3, 8), jts.IndexStrtree_NewInterval(3, 4).ExpandToInclude(jts.IndexStrtree_NewInterval(7, 8)))
	junit.AssertEqualsDeep(t, jts.IndexStrtree_NewInterval(3, 7), jts.IndexStrtree_NewInterval(3, 7).ExpandToInclude(jts.IndexStrtree_NewInterval(4, 5)))
	junit.AssertEqualsDeep(t, jts.IndexStrtree_NewInterval(3, 8), jts.IndexStrtree_NewInterval(3, 7).ExpandToInclude(jts.IndexStrtree_NewInterval(4, 8)))
}
