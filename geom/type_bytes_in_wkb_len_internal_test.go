package geom

import "testing"

// TestBytesInWKBLenMatchesAsBinary checks that bytesInWKB returns the same
// number as the len of the WKB produced by AsBinary.
func TestBytesInWKBLenMatchesAsBinary(t *testing.T) {
	for i, g := range wkbTestGeometries() {
		if got, want := len(g.AsBinary()), g.bytesInWKB(); got != want {
			t.Errorf("bytesInWKB mismatch: geometry %d len(AsBinary)=%d bytesInWKB=%d", i, got, want)
		}
	}
}
