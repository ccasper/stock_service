package main

import (
	common "app/internal/common"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	g "maragu.dev/gomponents"

	// Importing this as '.' is intentional for cleaner HTML like code.
	. "maragu.dev/gomponents/html"
)

var g_dataDir string

func renderMetricCard(m Metric) g.Node {
	bgColor := map[string]string{
		"green":  "bg-green-900 border-green-500",
		"yellow": "bg-yellow-900 border-yellow-500",
		"red":    "bg-red-900 border-red-500",
	}[m.Color]
	textColor := map[string]string{
		"green":  "text-green-300",
		"yellow": "text-yellow-300",
		"red":    "text-red-300",
	}[m.Color]
	return Div(Class("border-l-4 p-4 rounded-lg shadow "+bgColor),
		Div(Class("flex justify-between items-start mb-2"),
			H3(Class("font-semibold text-lg "+textColor), g.Text(m.Name)),
			Span(Class("text-2xl font-bold "+textColor), g.Text(m.Value)),
		),
		P(Class("text-sm text-gray-400 mt-2"), g.Raw(m.Reason)),
	)
}

func stockPage(symbol string) g.Node {
	result, err := getStockMetrics(symbol)
	if err != nil {
		return errorPage("Error Fetching Data", err.Error(), symbol)
	}

	metricsList := buildMetricsList(result)

	var metricCards []g.Node
	for _, m := range metricsList {
		metricCards = append(metricCards,
			Div(
				g.Attr("x-show", fmt.Sprintf("filter === 'all' || filter === '%s'", m.Color)),
				renderMetricCard(m),
			),
		)
	}

	return HTML(
		Head(
			Meta(Charset("UTF-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1.0")),
			TitleEl(g.Text(fmt.Sprintf("%s - Stock Analysis", symbol))),
			Script(Src("https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"), Defer()),
			Script(Src("https://cdn.tailwindcss.com")),
			Script(g.Raw(`tailwind.config = { theme: { extend: { colors: { darkbg: '#1a1a1a' } } } }`)),
		),
		Body(
			Class("bg-darkbg text-gray-200 min-h-screen"),
			g.Attr("x-data", "{ isNavigating: false }"), Div(Class("container mx-auto px-4 py-8"),
				Div(Class("mb-8"),
					Div(Class("flex items-center justify-between mb-4"),
						Div(
							H1(Class("text-4xl font-bold text-white mb-2"), g.Text(symbol)),
							P(Class("text-gray-400"), g.Text("Real-time Stock Metrics from Yahoo Finance")),
						),
						A(
							Href("/"),
							Class("text-blue-400 hover:underline text-sm flex items-center"),
							g.Attr("@click", "isNavigating = true"),
							// Spinner shown while navigating
							Div(
								g.Attr("x-show", "isNavigating"),
								g.Attr("x-transition"),
								Class("ml-2 w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"),
							),
							g.Text("← New Search"),
						),
					),
					P(Class("text-xs text-gray-500 mt-2"), g.Text(fmt.Sprintf("Showing %d available metrics", len(metricsList)))),
				),

				Div(Class("mb-6 flex gap-4"),
					colorKey("green", "Strong Buy Signal"),
					colorKey("yellow", "Neutral"),
					colorKey("red", "Caution"),
				),

				Div(g.Attr("x-data", "{ filter: 'all' }"),
					filterButtons(),
					Div(Class("grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"),
						g.Group(metricCards),
					),
				),

				Div(Class("mt-8 p-4 bg-gray-800 rounded-lg border border-gray-700"),
					P(Class("text-sm text-gray-400"),
						g.Text("Note: Data is pulled from Yahoo Finance's internal JSON endpoint. The endpoint is unofficial and may change."),
					),
				),
			),
		),
	)
}

func errorPage(title, message, symbol string) g.Node {
	return HTML(
		Head(
			Meta(Charset("UTF-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1.0")),
			TitleEl(g.Text("Error")),
			Script(Src("https://cdn.tailwindcss.com")),
		),
		Body(Class("bg-gray-900 text-gray-100 min-h-screen flex items-center justify-center"),
			Div(Class("container mx-auto px-4"),
				Div(Class("max-w-md mx-auto bg-gray-800 rounded-lg shadow-lg p-8"),
					H1(Class("text-2xl font-bold text-red-400 mb-4"), g.Text(title)),
					P(Class("text-gray-300 mb-4"), g.Text(message)),
					A(Href("/"), Class("text-blue-400 hover:underline"), g.Text("← Back to Home")),
				),
			),
		),
	)
}

func buildMetricsList(result *Result) []Metric {
	tempPegRatio := result.SummaryDetail.TrailingPE.Raw / result.FinancialData.EarningsGrowth.Raw
	tempROIC := CalculateROIC(result.FinancialData, result.DefaultKeyStatistics)
	var metricsList []Metric
	metricConfigs := []struct {
		name         string
		value        *float64
		isPercent    bool
		needsScaling bool
	}{
		{"P/E Ratio", &result.SummaryDetail.TrailingPE.Raw, false, false},
		{"Short Ratio", &result.DefaultKeyStatistics.ShortRatio.Raw, false, false},
		{"Short Percent of Float", &result.DefaultKeyStatistics.ShortPercentOfFloat.Raw, true, false},
		{"Forward P/E", &result.SummaryDetail.ForwardPE.Raw, false, false},
		{"P/B Ratio", &result.DefaultKeyStatistics.PriceToBook.Raw, false, false},
		{"P/S Ratio", &result.SummaryDetail.PriceToSalesTrailing12Months.Raw, false, false},
		{"PEG Ratio", &tempPegRatio, false, false},
		{"Debt/Equity", &result.FinancialData.DebtToEquity.Raw, false, false},
		{"Current Ratio", &result.FinancialData.CurrentRatio.Raw, false, false},
		{"Quick Ratio", &result.FinancialData.QuickRatio.Raw, false, false},
		{"ROE", &result.FinancialData.ReturnOnEquity.Raw, true, true},
		{"ROA", &result.FinancialData.ReturnOnAssets.Raw, true, true},
		{"ROIC", &tempROIC, true, false}, // not present in struct; set to 0 or compute elsewhere
		{"Gross Margin", &result.FinancialData.GrossMargins.Raw, true, true},
		{"Operating Margin", &result.FinancialData.OperatingMargins.Raw, true, true},
		{"Net Margin", &result.DefaultKeyStatistics.ProfitMargins.Raw, true, true},
		{"Revenue Growth", &result.FinancialData.RevenueGrowth.Raw, true, true},
		{"Earnings Growth", &result.FinancialData.EarningsGrowth.Raw, true, true},
		{"Free Cash Flow", &result.FinancialData.FreeCashflow.Raw, false, false},
		{"Beta", &result.SummaryDetail.Beta.Raw, false, false},
		{"Dividend Yield", &result.SummaryDetail.DividendYield.Raw, true, false},
		{"Price", &result.FinancialData.CurrentPrice.Raw, false, false},
		{"Market Cap", &result.SummaryDetail.MarketCap.Raw, false, false},
		{"Enterprise Value", &result.DefaultKeyStatistics.EnterpriseValue.Raw, false, false},
		{"Shares Outstanding", &result.DefaultKeyStatistics.SharesOutstanding.Raw, false, false},
		{"Book Value", &result.DefaultKeyStatistics.BookValue.Raw, false, false},
		{"Return on Equity", &result.FinancialData.ReturnOnEquity.Raw, true, true},
	}

	for _, cfg := range metricConfigs {
		if m := buildMetricCardInformation(cfg.name, cfg.value, cfg.isPercent); m != nil {
			metricsList = append(metricsList, *m)
		}
	}
	return metricsList
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	page := HTML(
		Head(
			Meta(Charset("UTF-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1.0")),
			TitleEl(g.Text("Stock Metrics Analyzer")),
			Script(Src("https://cdn.tailwindcss.com")),
			Script(Src("https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"), Defer()), // 🔧 Add Alpine.js
		),
		Body(Class("bg-gray-50 min-h-screen flex items-center justify-center"),
			Div(Class("container mx-auto px-4"),
				Div(Class("max-w-md mx-auto bg-white rounded-lg shadow-lg p-8"),
					H1(Class("text-3xl font-bold text-gray-900 mb-2"), g.Text("Stock Metrics Analyzer")),
					P(Class("text-gray-600 mb-6 text-sm"), g.Text("Real-time data from Yahoo Finance")),

					// 🔧 Alpine.js loading state
					Div(g.Attr("x-data", "{ isLoading: false }"),
						FormEl(
							Action("/stock"),
							Method("GET"),
							g.Attr("@submit", "isLoading = true"),
							Class("space-y-4"),

							Div(
								Label(For("symbol"), Class("block text-sm font-medium text-gray-700 mb-2"),
									g.Text("Enter Stock Symbol"),
								),
								Input(
									Type("text"),
									Name("symbol"),
									ID("symbol"),
									Placeholder("e.g., AAPL, MSFT, GOOGL"),
									Required(),
									Class("w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent uppercase"),
								),
							),

							Button(
								Type("submit"),
								Class("w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 transition flex items-center justify-center relative"),
								g.Attr(":disabled", "isLoading"),

								// Spinner
								Div(
									g.Attr("x-show", "isLoading"),
									g.Attr("x-transition.opacity"),
									Class("absolute left-4 w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"),
								),

								// Text changes
								Span(
									g.Attr("x-text", "isLoading ? 'Loading...' : 'Analyze Stock'"),
									Class("ml-2"),
								),
							)),
					),

					Div(Class("mt-6 p-4 bg-gray-50 rounded-lg"),
						P(Class("text-xs text-gray-600"),
							g.Text("This tool uses Yahoo Finance's internal JSON endpoint. Use responsibly."),
						),
					),
				),
			),
		),
	)

	w.Header().Set("Content-Type", "text/html")
	page.Render(w)
}

func colorKey(color, label string) g.Node {
	colorClass := map[string]string{
		"green":  "bg-green-500",
		"yellow": "bg-yellow-500",
		"red":    "bg-red-500",
	}[color]
	return Div(Class("flex items-center gap-2"),
		Div(Class("w-4 h-4 "+colorClass+" rounded")),
		Span(Class("text-sm text-gray-300"), g.Text(label)),
	)
}

func filterButtons() g.Node {
	return Div(Class("mb-6"),
		g.Attr(":disabled", "isNavigating"),
		Label(Class("mr-4 text-gray-400 font-medium"), g.Text("Filter by: ")),
		Button(g.Attr("@click", "filter = 'all'"), g.Attr(":class", "filter === 'all' ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-300'"), Class("px-4 py-2 rounded mr-2 transition"), g.Text("All")),
		Button(g.Attr("@click", "filter = 'green'"), g.Attr(":class", "filter === 'green' ? 'bg-green-600 text-white' : 'bg-gray-700 text-gray-300'"), Class("px-4 py-2 rounded mr-2 transition"), g.Text("Strong")),
		Button(g.Attr("@click", "filter = 'yellow'"), g.Attr(":class", "filter === 'yellow' ? 'bg-yellow-600 text-white' : 'bg-gray-700 text-gray-300'"), Class("px-4 py-2 rounded mr-2 transition"), g.Text("Neutral")),
		Button(g.Attr("@click", "filter = 'red'"), g.Attr(":class", "filter === 'red' ? 'bg-red-600 text-white' : 'bg-gray-700 text-gray-300'"), Class("px-4 py-2 rounded transition"), g.Text("Caution")),
	)
}

func stockHandler(w http.ResponseWriter, r *http.Request) {
	symbol := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
	if symbol == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	page := stockPage(symbol)
	w.Header().Set("Content-Type", "text/html")
	page.Render(w)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	symbol := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
	if symbol == "" {
		http.Error(w, "symbol parameter required", http.StatusBadRequest)
		return
	}
	metrics, err := getStockMetrics(symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func main() {
	d := &SystemdDaemon{}
	EnableBackgroundWatchdog(d)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/stock", stockHandler)
	http.HandleFunc("/api/metrics", apiHandler)

	port := flag.Int64("port", 8080, "port to listen on")
	ip := flag.String("ip", "", "ip to listen on")
	flagDataDir := flag.String("data", "", "directory to store data")

	// Parse command-line flags
	flag.Parse()

	g_dataDir = *flagDataDir

	// Run the health check port.
	healthPort := (*port) + 1
	err := common.StartHealthServer(VERSION, fmt.Sprintf("%s:%d", *ip, healthPort))
	if err != nil {
		log.Printf("Error starting health server: %v", err)
		log.Fatal(err)
	}

	log.Printf("Server starting on http://%s:%d", *ip, *port)
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", *ip, *port), nil)
	if err != nil {
		log.Printf("Error starting serving server: %v", err)
		log.Fatal(err)
	}

}

type Metric struct {
	Name   string
	Value  string
	Color  string
	Reason string
}

// Makes the metric presentable for a Metric Card.
func buildMetricCardInformation(name string, value *float64, isPercent bool) *Metric {
	if value == nil {
		return nil
	}
	var valueStr string
	if isPercent {
		valueStr = fmt.Sprintf("%.2f%%", *value*100)
	} else {
		// for large numbers (marketcap, ev, fcf) show compact formatting
		valueStr = common.FormatLargeNumber(*value)
	}
	color, reason := getColorAndReasonForMetric(name, *value)
	return &Metric{Name: name, Value: valueStr, Color: color, Reason: reason}
}
