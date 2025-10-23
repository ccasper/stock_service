package main

func CalculateROIC(financialData FinancialData, keyStats DefaultKeyStatistics) float64 {
	// Approximate NOPAT
	ebitda := financialData.Ebitda.Raw
	taxRate := 0.21 // Use actual if available
	nopat := ebitda * (1 - taxRate)

	// Calculate Equity = Book Value × Shares Outstanding
	equity := keyStats.BookValue.Raw * keyStats.SharesOutstanding.Raw

	// Invested Capital = Debt + Equity - Cash
	investedCapital := financialData.TotalDebt.Raw + equity - financialData.TotalCash.Raw

	if investedCapital == 0 {
		return 0
	}

	roic := nopat / investedCapital
	return roic
}
