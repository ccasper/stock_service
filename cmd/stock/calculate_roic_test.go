package main

import (
	"testing"
)

func TestCalculateROIC(t *testing.T) {
	tests := []struct {
		name          string
		financialData FinancialData
		keyStats      DefaultKeyStatistics
		want          float64
	}{
		{
			name: "normal case",
			financialData: FinancialData{
				Ebitda:    FmtRaw{Raw: 1000},
				TotalDebt: FmtRaw{Raw: 500},
				TotalCash: FmtRaw{Raw: 200},
			},
			keyStats: DefaultKeyStatistics{
				BookValue:         FmtRaw{Raw: 300},
				SharesOutstanding: FmtRaw{Raw: 10},
			},
			want: (1000 * (1 - 0.21)) / (500 + 300*10 - 200), // (1000*0.79) / (500 + 3000 - 200) = 790 / 3300 ≈ 0.2394
		},
		{
			name: "zero invested capital",
			financialData: FinancialData{
				Ebitda:    FmtRaw{Raw: 1000},
				TotalDebt: FmtRaw{Raw: 200},
				TotalCash: FmtRaw{Raw: 1200},
			},
			keyStats: DefaultKeyStatistics{
				BookValue:         FmtRaw{Raw: 50},
				SharesOutstanding: FmtRaw{Raw: 20},
			},
			want: 0, // investedCapital = 200 + 50*20 - 1200 = 200 + 1000 - 1200 = 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateROIC(tt.financialData, tt.keyStats)
			if diff := got - tt.want; diff > 0.0001 || diff < -0.0001 {
				t.Errorf("CalculateROIC() = %v, want %v", got, tt.want)
			}
		})
	}
}
