package main

import (
	"testing"
)

func TestGetColorAndReasonForMetric(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		wantColor string
	}{
		{"Short Ratio", 1.5, "green"},
		{"Short Ratio", 3.0, "yellow"},
		{"Short Ratio", 6.0, "red"},

		{"P/E Ratio", 10, "green"},
		{"P/E Ratio", 20, "yellow"},
		{"P/E Ratio", 30, "red"},

		{"Unknown Metric", 0, "yellow"},
	}

	for _, tt := range tests {
		gotColor, gotReason := getColorAndReasonForMetric(tt.name, tt.value)

		if gotColor != tt.wantColor {
			t.Errorf("getColorAndReasonForMetric(%q, %v) color = %q; want %q", tt.name, tt.value, gotColor, tt.wantColor)
		}

		if len(gotReason) <= 30 {
			t.Errorf("getColorAndReasonForMetric(%q, %v) reason length = %d; want > 30", tt.name, tt.value, len(gotReason))
		}
	}
}
