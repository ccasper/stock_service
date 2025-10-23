package main

import (
	"log"
	"testing"
)

func TestGetStockMetrics_Integration(t *testing.T) {
	log.Printf("Running integration test for getStockMetrics() ... this is a slower test!")
	ticker := "AAPL" // Use a reliable, real ticker

	result, err := getStockMetrics(ticker)
	if err != nil {
		t.Fatalf("getStockMetrics(%q) returned error: %v", ticker, err)
	}

	if result == nil {
		t.Fatalf("getStockMetrics(%q) returned nil result", ticker)
	}

	// Basic sanity checks on returned data
	if result.SummaryDetail.FiftyTwoWeekHigh.Raw < 1.0 {
		t.Errorf("Expected some value for FiftyTwoWeekHigh field %+v", result.SummaryDetail.FiftyTwoWeekHigh)
	}

	if result.FinancialData.Ebitda.Raw == 0 {
		t.Errorf("Expected non-zero EBITDA %+v", result.FinancialData.Ebitda)
	}

	// Add more assertions as needed depending on Result struct fields

	t.Logf("Successfully fetched data for %s", ticker)
}
