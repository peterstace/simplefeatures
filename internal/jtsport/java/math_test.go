package java_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

func TestRound(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		// Positive numbers - same as Go's math.Round.
		{1.4, 1},
		{1.5, 2},
		{1.6, 2},
		{2.5, 3},
		// Negative numbers - differs from Go's math.Round (which rounds away from zero).
		{-1.4, -1},
		{-1.5, -1}, // Go would give -2.
		{-1.6, -2},
		{-2.5, -2}, // Go would give -3.
		// Edge cases from PrecisionModel tests.
		{-1232.5, -1232}, // Go would give -1233.
		{-1232.4, -1232},
		{-1232.6, -1233},
		{1232.5, 1233},
		// Zero.
		{0, 0},
		{0.5, 1},
		{-0.5, 0}, // Go would give -1.
	}

	for _, tt := range tests {
		result := java.Round(tt.input)
		if result != tt.expected {
			t.Errorf("Round(%v) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}
