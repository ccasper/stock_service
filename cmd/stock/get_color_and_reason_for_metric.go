package main

import "fmt"

func getColorAndReasonForMetric(name string, value float64) (string, string) {
	switch name {
	case "Short Ratio":
		if value < 2 {
			return "green", "Low short interest ratio suggests the stock may be undervalued."
		} else if value < 5 {
			return "yellow", "Moderate short interest ratio — the stock is fairly valued, but depends on industry norms."
		}
		return "red", "High short interest ratio may indicate overvaluation, meaning you're paying a premium for short interest."

	case "Short Percent of Float":
		if value < .1 {
			return "green", "Low short interest ratio suggests the stock may be undervalued."
		} else if value < .5 {
			return "yellow", "Moderate short interest ratio — the stock is fairly valued, but depends on industry norms."
		}
		return "red", "High short interest ratio may indicate overvaluation, meaning you're paying a premium for short interest."
	case "P/E Ratio":
		if value < 15 {
			return "green", "Low P/E suggests the stock may be undervalued relative to earnings."
		} else if value < 25 {
			return "yellow", "Moderate P/E — the stock is fairly valued, but depends on industry norms."
		}
		return "red", "High P/E may indicate overvaluation, meaning you're paying a premium for earnings."
	case "PEG Ratio":
		desc := "<br><br>PEG= (P/E Ratio)/(Earnings Growth Rate) This tells you how much you're paying for each percentage point of expected growth."
		if value < 15 {
			return "green", "Low P/E ratio factoring in future earnings growth suggests the stock may be undervalued relative to earnings and growth." + desc
		} else if value < 25 {
			return "yellow", "Moderate P/E ratio factoring in future earnings growth — the stock is fairly valued, but depends on industry norms." + desc
		}
		return "red", "High P/E ratio factoring in future earnings growth may indicate overvaluation, meaning you're paying a premium for earnings and growth." + desc

	case "Forward P/E":
		if value < 15 {
			return "green", "Low forward P/E suggests earnings are expected to grow, making it potentially undervalued."
		} else if value < 25 {
			return "yellow", "Moderate forward P/E — growth expectations are priced in."
		}
		return "red", "High forward P/E could mean over-optimistic growth assumptions or expensive valuation."

	case "P/B Ratio":
		if value < 1.5 {
			return "green", "Low P/B may indicate the stock is trading below its book value — potentially a bargain."
		} else if value < 3 {
			return "yellow", "Moderate P/B — fairly valued compared to assets."
		}
		return "red", "High P/B may suggest overvaluation or overconfidence in asset efficiency."

	case "P/S Ratio":
		if value < 2 {
			return "green", "Low P/S suggests the stock is reasonably priced relative to revenue."
		} else if value < 5 {
			return "yellow", "Moderate P/S — revenue valuation is acceptable but monitor margins."
		}
		return "red", "High P/S can indicate overvaluation, especially if profits are weak."

	case "Debt/Equity":
		if value < 0.5 {
			return "green", "Low debt levels suggest financial stability and lower risk."
		} else if value < 1.5 {
			return "yellow", "Moderate debt — manageable, but watch interest costs."
		}
		return "red", "High debt increases financial risk, especially if cash flows are weak."

	case "Current Ratio":
		if value > 2 {
			return "green", "Strong liquidity — the company can easily meet short-term liabilities."
		} else if value > 1 {
			return "yellow", "Adequate liquidity, but less buffer in case of financial stress."
		}
		return "red", "Poor liquidity — the company may struggle to cover short-term obligations."

	case "Quick Ratio":
		if value > 1.5 {
			return "green", "Excellent liquidity — even excluding inventory, the company is financially healthy."
		} else if value > 0.8 {
			return "yellow", "Acceptable liquidity — but inventory reliance is higher."
		}
		return "red", "Poor quick ratio — short-term liabilities may not be well-covered."

	case "ROE":
		if value > 20 {
			return "green", "Excellent ROE — the company is using equity efficiently to generate profits."
		} else if value > 10 {
			return "yellow", "Moderate ROE — reasonable returns on equity."
		}
		return "red", "Low ROE — inefficient capital use or declining profitability."

	case "ROA":
		if value > 10 {
			return "green", "High ROA — strong use of assets to generate earnings."
		} else if value > 5 {
			return "yellow", "Moderate ROA — reasonable asset efficiency."
		}
		return "red", "Low ROA — could signal inefficient asset management or low profitability."

	case "Gross Margin":
		if value > 40 {
			return "green", "Strong gross margin — the company has pricing power or cost efficiency."
		} else if value > 20 {
			return "yellow", "Moderate margins — acceptable for many industries."
		}
		return "red", "Low margins — may struggle with profitability or face pricing pressure."

	case "Operating Margin":
		if value > 20 {
			return "green", "Excellent operating efficiency and cost control."
		} else if value > 10 {
			return "yellow", "Decent operating margin — the business model is sustainable."
		}
		return "red", "Low operating margin — profitability may be under pressure."

	case "Net Margin":
		if value > 15 {
			return "green", "Strong net margin — good bottom-line profitability."
		} else if value > 5 {
			return "yellow", "Moderate net margin — acceptable for many industries."
		}
		return "red", "Weak net margin — high costs or low pricing power."

	case "Revenue Growth":
		if value > 15 {
			return "green", "Strong revenue growth — indicates expansion and market demand."
		} else if value > 5 {
			return "yellow", "Moderate growth — stable but not rapid."
		}
		return "red", "Weak revenue growth — may indicate stagnation or competitive pressure."

	case "Earnings Growth":
		if value > 15 {
			return "green", "Strong earnings growth — profit is accelerating."
		} else if value > 5 {
			return "yellow", "Moderate earnings growth — consistent but not spectacular."
		}
		return "red", "Weak or negative earnings growth — could be a red flag for investors."

	case "Free Cash Flow":
		if value > 0 {
			return "green", "Positive FCF — the company generates more cash than it spends, allowing flexibility."
		}
		return "red", "Negative FCF — the company is spending more than it brings in, may need financing."

	case "Beta":
		if value < 0.8 {
			return "green", "Low beta — the stock is less volatile than the market, suitable for risk-averse investors."
		} else if value < 1.2 {
			return "yellow", "Average beta — price movement is roughly in line with the market."
		}
		return "red", "High beta — more volatile, riskier in down markets."

	case "Dividend Yield":
		if value > 3 {
			return "green", "High dividend yield — good income potential for investors."
		} else if value > 1 {
			return "yellow", "Moderate dividend — some income, but not a focus."
		}
		return "red", "Low or no dividend — not ideal for income-focused investors."
	case "ROIC":
		desc := "<br><br>ROIC is a measure of profitability relative to total assets. It is crucial to compare ROIC within the same sector, as industries like technology may achieve higher ROIC due to lower capital requirements, while capital-intensive industries like utilities typically have lower ratios."
		if value > 10 {
			return "green", fmt.Sprintf("High ROIC — %s", desc)
		} else if value > 5 {
			return "yellow", fmt.Sprintf("Moderate ROIC — %s", desc)
		} else {
			return "red", fmt.Sprintf("Low ROIC — %s", desc)
		}

	}

	// Default case
	return "yellow", "No specific evaluation available for this metric."
}
