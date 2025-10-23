package main

import (
	"testing"
)

func TestFormatLargeNumber(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{999, "999.00"},
		{1000, "1.00K"},
		{1234, "1.23K"},
		{999999, "1000.00K"}, // edge case, rounds to 1000K = 1M?
		{1_000_000, "1.00M"},
		{2_504_999, "2.50M"},
		{999_999_999, "1000.00M"}, // edge case again
		{1_000_000_000, "1.00B"},
		{5_432_100_000, "5.43B"},
		{1_000_000_555_000, "1.00T"},
		{3_210_000_000_000, "3.21T"},
		{0, "0.00"},
		{-500, "-500.00"},
		{-1500, "-1.50K"},
		{-2_504_999, "-2.50M"},
		{-1_000_000_555_000, "-1.00T"},
	}

	for _, tt := range tests {
		got := formatLargeNumber(tt.input)
		if got != tt.expected {
			t.Errorf("formatLargeNumber(%v) = %v; want %v", tt.input, got, tt.expected)
		}
	}
}
