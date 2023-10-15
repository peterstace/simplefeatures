package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestRelateMatch(t *testing.T) {
	for i, tc := range []struct {
		mat  string
		pat  string
		want bool
	}{
		{"FFFFFFFFF", "FFFFFFFFF", true},
		{"FFFFFFFFF", "000000000", false},
		{"FFFFFFFFF", "111111111", false},
		{"FFFFFFFFF", "222222222", false},
		{"FFFFFFFFF", "TTTTTTTTT", false},
		{"FFFFFFFFF", "*********", true},

		{"000000000", "FFFFFFFFF", false},
		{"000000000", "000000000", true},
		{"000000000", "111111111", false},
		{"000000000", "222222222", false},
		{"000000000", "TTTTTTTTT", true},
		{"000000000", "*********", true},

		{"111111111", "FFFFFFFFF", false},
		{"111111111", "000000000", false},
		{"111111111", "111111111", true},
		{"111111111", "222222222", false},
		{"111111111", "TTTTTTTTT", true},
		{"111111111", "*********", true},

		{"222222222", "FFFFFFFFF", false},
		{"222222222", "000000000", false},
		{"222222222", "111111111", false},
		{"222222222", "222222222", true},
		{"222222222", "TTTTTTTTT", true},
		{"222222222", "*********", true},

		{"F012F012F", "*********", true},
		{"F012F012F", "F*1**T*2*", true},
		{"F012F012F", "F*11*T*2*", false},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := geom.RelateMatches(tc.mat, tc.pat)
			if err != nil {
				t.Error(err)
			}
			if got != tc.want {
				t.Logf("matrix:  %v", tc.mat)
				t.Logf("pattern: %v", tc.pat)
				t.Errorf("want=%t got=%t", tc.want, got)
			}
		})
	}
}

func TestRelateMatchError(t *testing.T) {
	for i, tc := range []struct {
		mat string
		pat string
	}{
		{"FFFFFFFF", "FFFFFFFFF"},
		{"FFFFFFFFF", "FFFFFFFF"},
		{"FFFFXFFFF", "FFFFFFFFF"},
		{"FFFFFFFFF", "FFFFXFFFF"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := geom.RelateMatches(tc.mat, tc.pat)
			t.Log(err)
			if err == nil {
				t.Error("expected error but got nil")
			}
		})
	}
}
