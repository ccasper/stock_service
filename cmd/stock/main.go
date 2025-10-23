package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/coreos/go-systemd/daemon"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var (
	// Path to storing data.
	dataDir string
)

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
		if m := formatMetric(cfg.name, cfg.value, cfg.isPercent); m != nil {
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
	// Enable a 30 second watchdog with systemd:
	// Tell systemd we're ready
	sent, err := daemon.SdNotify(false, "READY=1")
	if err != nil {
		log.Fatalf("Failed to notify systemd: %v", err)
	}
	if !sent {
		log.Println("Not running under systemd or watchdog not enabled")
	}

	// Get watchdog interval from environment variable (set by systemd)
	watchdogInterval, err := daemon.SdWatchdogEnabled(false)
	if err != nil {
		log.Fatalf("Failed to check watchdog: %v", err)
	}

	if watchdogInterval > 0 {
		ticker := time.NewTicker(watchdogInterval / 2) // Send keep-alive at half interval
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				// Send watchdog keep-alive ping
				daemon.SdNotify(false, "WATCHDOG=1")
			}
		}()
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/stock", stockHandler)
	http.HandleFunc("/api/metrics", apiHandler)

	// Get the location of this executable.
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	dataDirDefault := filepath.Join(filepath.Dir(exePath), "../data")

	port := flag.String("port", "8080", "port to listen on")
	ip := flag.String("ip", "", "ip to listen on")
	data := flag.String("data", dataDirDefault, "data directory for storage")

	// Parse command-line flags
	flag.Parse()

	// Set the package wide dataDir path.
	dataDir = *data

	log.Printf("Server starting on http://%s:%s", *ip, *port)
	log.Fatal(http.ListenAndServe(*ip+":"+*port, nil))
}

type Response struct {
	QuoteSummary QuoteSummary `json:"quoteSummary"`
}

type QuoteSummary struct {
	Result []Result    `json:"result"`
	Error  interface{} `json:"error"`
}

type Result struct {
	SummaryDetail        SummaryDetail        `json:"summaryDetail"`
	DefaultKeyStatistics DefaultKeyStatistics `json:"defaultKeyStatistics"`
	Earnings             Earnings             `json:"earnings"`
	FinancialData        FinancialData        `json:"financialData"`
}

type PriceHint struct {
	Raw     int    `json:"raw"`
	Fmt     string `json:"fmt"`
	LongFmt string `json:"longFmt"`
}

type FmtRaw struct {
	Raw     float64 `json:"raw"`
	Fmt     string  `json:"fmt"`
	LongFmt string  `json:"longFmt,omitempty"`
}

type SummaryDetail struct {
	MaxAge                       int         `json:"maxAge"`
	PriceHint                    PriceHint   `json:"priceHint"`
	PreviousClose                FmtRaw      `json:"previousClose"`
	Open                         FmtRaw      `json:"open"`
	DayLow                       FmtRaw      `json:"dayLow"`
	DayHigh                      FmtRaw      `json:"dayHigh"`
	RegularMarketPreviousClose   FmtRaw      `json:"regularMarketPreviousClose"`
	RegularMarketOpen            FmtRaw      `json:"regularMarketOpen"`
	RegularMarketDayLow          FmtRaw      `json:"regularMarketDayLow"`
	RegularMarketDayHigh         FmtRaw      `json:"regularMarketDayHigh"`
	DividendRate                 FmtRaw      `json:"dividendRate"`
	DividendYield                FmtRaw      `json:"dividendYield"`
	ExDividendDate               FmtRaw      `json:"exDividendDate"`
	PayoutRatio                  FmtRaw      `json:"payoutRatio"`
	FiveYearAvgDividendYield     interface{} `json:"fiveYearAvgDividendYield"`
	Beta                         FmtRaw      `json:"beta"`
	TrailingPE                   FmtRaw      `json:"trailingPE"`
	ForwardPE                    FmtRaw      `json:"forwardPE"`
	Volume                       FmtRaw      `json:"volume"`
	RegularMarketVolume          FmtRaw      `json:"regularMarketVolume"`
	AverageVolume                FmtRaw      `json:"averageVolume"`
	AverageVolume10days          FmtRaw      `json:"averageVolume10days"`
	AverageDailyVolume10Day      FmtRaw      `json:"averageDailyVolume10Day"`
	Bid                          FmtRaw      `json:"bid"`
	Ask                          FmtRaw      `json:"ask"`
	BidSize                      FmtRaw      `json:"bidSize"`
	AskSize                      FmtRaw      `json:"askSize"`
	MarketCap                    FmtRaw      `json:"marketCap"`
	Yield                        interface{} `json:"yield"`
	YtdReturn                    interface{} `json:"ytdReturn"`
	QtdReturn                    interface{} `json:"qtdReturn"`
	TotalAssets                  interface{} `json:"totalAssets"`
	ExpireDate                   interface{} `json:"expireDate"`
	StrikePrice                  interface{} `json:"strikePrice"`
	OpenInterest                 interface{} `json:"openInterest"`
	FiftyTwoWeekLow              FmtRaw      `json:"fiftyTwoWeekLow"`
	FiftyTwoWeekHigh             FmtRaw      `json:"fiftyTwoWeekHigh"`
	AllTimeHigh                  FmtRaw      `json:"allTimeHigh"`
	AllTimeLow                   FmtRaw      `json:"allTimeLow"`
	PriceToSalesTrailing12Months FmtRaw      `json:"priceToSalesTrailing12Months"`
	FiftyDayAverage              FmtRaw      `json:"fiftyDayAverage"`
	TwoHundredDayAverage         FmtRaw      `json:"twoHundredDayAverage"`
	TrailingAnnualDividendRate   FmtRaw      `json:"trailingAnnualDividendRate"`
	TrailingAnnualDividendYield  FmtRaw      `json:"trailingAnnualDividendYield"`
	NavPrice                     interface{} `json:"navPrice"`
	Currency                     string      `json:"currency"`
	FromCurrency                 *string     `json:"fromCurrency"`
	ToCurrency                   *string     `json:"toCurrency"`
	LastMarket                   *string     `json:"lastMarket"`
	CoinMarketCapLink            *string     `json:"coinMarketCapLink"`
	Volume24Hr                   interface{} `json:"volume24Hr"`
	VolumeAllCurrencies          interface{} `json:"volumeAllCurrencies"`
	CirculatingSupply            interface{} `json:"circulatingSupply"`
	Algorithm                    *string     `json:"algorithm"`
	MaxSupply                    interface{} `json:"maxSupply"`
	StartDate                    interface{} `json:"startDate"`
	Tradeable                    bool        `json:"tradeable"`
}

type DefaultKeyStatistics struct {
	MaxAge                       int         `json:"maxAge"`
	PriceHint                    PriceHint   `json:"priceHint"`
	EnterpriseValue              FmtRaw      `json:"enterpriseValue"`
	ForwardPE                    FmtRaw      `json:"forwardPE"`
	ProfitMargins                FmtRaw      `json:"profitMargins"`
	FloatShares                  FmtRaw      `json:"floatShares"`
	SharesOutstanding            FmtRaw      `json:"sharesOutstanding"`
	SharesShort                  FmtRaw      `json:"sharesShort"`
	SharesShortPriorMonth        FmtRaw      `json:"sharesShortPriorMonth"`
	SharesShortPreviousMonthDate FmtRaw      `json:"sharesShortPreviousMonthDate"`
	DateShortInterest            FmtRaw      `json:"dateShortInterest"`
	SharesPercentSharesOut       FmtRaw      `json:"sharesPercentSharesOut"`
	HeldPercentInsiders          FmtRaw      `json:"heldPercentInsiders"`
	HeldPercentInstitutions      FmtRaw      `json:"heldPercentInstitutions"`
	ShortRatio                   FmtRaw      `json:"shortRatio"`
	ShortPercentOfFloat          FmtRaw      `json:"shortPercentOfFloat"`
	Beta                         FmtRaw      `json:"beta"`
	ImpliedSharesOutstanding     FmtRaw      `json:"impliedSharesOutstanding"`
	MorningStarOverallRating     interface{} `json:"morningStarOverallRating"`
	MorningStarRiskRating        interface{} `json:"morningStarRiskRating"`
	Category                     *string     `json:"category"`
	BookValue                    FmtRaw      `json:"bookValue"`
	PriceToBook                  FmtRaw      `json:"priceToBook"`
	AnnualReportExpenseRatio     interface{} `json:"annualReportExpenseRatio"`
	YtdReturn                    interface{} `json:"ytdReturn"`
	QtdReturn                    interface{} `json:"qtdReturn"`
	Beta3Year                    interface{} `json:"beta3Year"`
	TotalAssets                  interface{} `json:"totalAssets"`
	Yield                        interface{} `json:"yield"`
	FundFamily                   *string     `json:"fundFamily"`
	FundInceptionDate            interface{} `json:"fundInceptionDate"`
	LegalType                    *string     `json:"legalType"`
	ThreeYearAverageReturn       interface{} `json:"threeYearAverageReturn"`
	FiveYearAverageReturn        interface{} `json:"fiveYearAverageReturn"`
	PriceToSalesTrailing12Months interface{} `json:"priceToSalesTrailing12Months"`
	LastFiscalYearEnd            FmtRaw      `json:"lastFiscalYearEnd"`
	NextFiscalYearEnd            FmtRaw      `json:"nextFiscalYearEnd"`
	MostRecentQuarter            FmtRaw      `json:"mostRecentQuarter"`
	EarningsQuarterlyGrowth      FmtRaw      `json:"earningsQuarterlyGrowth"`
	RevenueQuarterlyGrowth       interface{} `json:"revenueQuarterlyGrowth"`
	NetIncomeToCommon            FmtRaw      `json:"netIncomeToCommon"`
	TrailingEps                  FmtRaw      `json:"trailingEps"`
	ForwardEps                   FmtRaw      `json:"forwardEps"`
	// PEG Ratio = (P/E Ratio) TrailingEps / (Earnings Growth Rate)
	PegRatio               interface{} `json:"pegRatio"`
	LastSplitFactor        string      `json:"lastSplitFactor"`
	LastSplitDate          FmtRaw      `json:"lastSplitDate"`
	EnterpriseToRevenue    FmtRaw      `json:"enterpriseToRevenue"`
	EnterpriseToEbitda     FmtRaw      `json:"enterpriseToEbitda"`
	FiftyTwoWeekChange     FmtRaw      `json:"52WeekChange"`
	SandP52WeekChange      FmtRaw      `json:"SandP52WeekChange"`
	LastDividendValue      FmtRaw      `json:"lastDividendValue"`
	LastDividendDate       FmtRaw      `json:"lastDividendDate"`
	LastCapGain            interface{} `json:"lastCapGain"`
	AnnualHoldingsTurnover interface{} `json:"annualHoldingsTurnover"`
	LatestFundingDate      interface{} `json:"latestFundingDate"`
	LatestAmountRaised     interface{} `json:"latestAmountRaised"`
	LatestImpliedValuation interface{} `json:"latestImpliedValuation"`
	LatestShareClass       *string     `json:"latestShareClass"`
	LeadInvestor           *string     `json:"leadInvestor"`
	FundingToDate          interface{} `json:"fundingToDate"`
	TotalFundingRounds     interface{} `json:"totalFundingRounds"`
}

type Earnings struct {
	MaxAge             int             `json:"maxAge"`
	EarningsChart      EarningsChart   `json:"earningsChart"`
	FinancialsChart    FinancialsChart `json:"financialsChart"`
	FinancialCurrency  string          `json:"financialCurrency"`
	DefaultMethodology string          `json:"defaultMethodology"`
}

type EarningsChart struct {
	Quarterly                  []QuarterlyEarning `json:"quarterly"`
	CurrentQuarterEstimate     FmtRaw             `json:"currentQuarterEstimate"`
	CurrentQuarterEstimateDate string             `json:"currentQuarterEstimateDate"`
	CurrentCalendarQuarter     string             `json:"currentCalendarQuarter"`
	CurrentQuarterEstimateYear int                `json:"currentQuarterEstimateYear"`
	CurrentFiscalQuarter       string             `json:"currentFiscalQuarter"`
	EarningsDate               []EarningsDate     `json:"earningsDate"`
	IsEarningsDateEstimate     bool               `json:"isEarningsDateEstimate"`
}

type QuarterlyEarning struct {
	Date            string `json:"date"`
	Actual          FmtRaw `json:"actual"`
	Estimate        FmtRaw `json:"estimate"`
	FiscalQuarter   string `json:"fiscalQuarter"`
	CalendarQuarter string `json:"calendarQuarter"`
	Difference      string `json:"difference"`
	SurprisePct     string `json:"surprisePct"`
}

type EarningsDate struct {
	Raw int64  `json:"raw"`
	Fmt string `json:"fmt"`
}

type FinancialsChart struct {
	Yearly    []YearlyFinancial    `json:"yearly"`
	Quarterly []QuarterlyFinancial `json:"quarterly"`
}

type YearlyFinancial struct {
	Date     int    `json:"date"`
	Revenue  FmtRaw `json:"revenue"`
	Earnings FmtRaw `json:"earnings"`
}

type QuarterlyFinancial struct {
	Date          string `json:"date"`
	FiscalQuarter string `json:"fiscalQuarter"`
	Revenue       FmtRaw `json:"revenue"`
	Earnings      FmtRaw `json:"earnings"`
}

type FinancialData struct {
	MaxAge                  int    `json:"maxAge"`
	CurrentPrice            FmtRaw `json:"currentPrice"`
	TargetHighPrice         FmtRaw `json:"targetHighPrice"`
	TargetLowPrice          FmtRaw `json:"targetLowPrice"`
	TargetMeanPrice         FmtRaw `json:"targetMeanPrice"`
	TargetMedianPrice       FmtRaw `json:"targetMedianPrice"`
	RecommendationMean      FmtRaw `json:"recommendationMean"`
	RecommendationKey       string `json:"recommendationKey"`
	NumberOfAnalystOpinions FmtRaw `json:"numberOfAnalystOpinions"`
	TotalCash               FmtRaw `json:"totalCash"`
	TotalCashPerShare       FmtRaw `json:"totalCashPerShare"`
	Ebitda                  FmtRaw `json:"ebitda"`
	TotalDebt               FmtRaw `json:"totalDebt"`
	QuickRatio              FmtRaw `json:"quickRatio"`
	CurrentRatio            FmtRaw `json:"currentRatio"`
	TotalRevenue            FmtRaw `json:"totalRevenue"`
	DebtToEquity            FmtRaw `json:"debtToEquity"`
	RevenuePerShare         FmtRaw `json:"revenuePerShare"`
	ReturnOnAssets          FmtRaw `json:"returnOnAssets"`
	ReturnOnEquity          FmtRaw `json:"returnOnEquity"`
	GrossProfits            FmtRaw `json:"grossProfits"`
	FreeCashflow            FmtRaw `json:"freeCashflow"`
	OperatingCashflow       FmtRaw `json:"operatingCashflow"`
	EarningsGrowth          FmtRaw `json:"earningsGrowth"`
	RevenueGrowth           FmtRaw `json:"revenueGrowth"`
	GrossMargins            FmtRaw `json:"grossMargins"`
	EbitdaMargins           FmtRaw `json:"ebitdaMargins"`
	OperatingMargins        FmtRaw `json:"operatingMargins"`
	ProfitMargins           FmtRaw `json:"profitMargins"`
	FinancialCurrency       string `json:"financialCurrency"`
}

func getStockMetrics(ticker string) (*Result, error) {
	// Ensure cache dir exists
	cacheDir := "./data/stockdata"
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("could not create cache dir: %v", err)
	}

	// Format current time for filename (rounded to hour)
	now := time.Now().UTC().Truncate(time.Hour)
	filename := fmt.Sprintf("%s-%s.json", ticker, now.Format("2006-01-02-15"))
	cachePath := filepath.Join(cacheDir, filename)

	var qs Response

	// If file exists, read and return
	if _, err := os.Stat(cachePath); err == nil {
		cachedData, err := ioutil.ReadFile(cachePath)
		if err != nil {
			return nil, fmt.Errorf("could not read cached file: %v", err)
		}
		err = json.Unmarshal(cachedData, &qs)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling cached JSON: %v", err)
		}
		fmt.Println("Loaded from cache:", cachePath)
		return &qs.QuoteSummary.Result[0], nil
	}

	// Create temp directories for Chrome data
	userDataDir := filepath.Join(os.TempDir(), fmt.Sprintf("chrome-user-data-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		log.Fatal(err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("user-data-dir", userDataDir),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		log.Fatal(err)
	}

	var crumb string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(fmt.Sprintf("https://finance.yahoo.com/quote/%s", ticker)),
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(`window.YAHOO && window.YAHOO.context && window.YAHOO.context.user && window.YAHOO.context.user.crumb || ""`, &crumb),
	); err != nil {
		log.Fatal(err)
	}

	if crumb == "" {
		log.Fatal("crumb token not found")
	}
	log.Printf("Crumb token: %s\n", crumb)

	// Fetch cookies properly inside chromedp.Run and ActionFunc
	var cookies []*network.Cookie
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		cookies, err = network.GetCookies().WithURLs([]string{"https://finance.yahoo.com"}).Do(ctx)
		return err
	})); err != nil {
		log.Fatal(err)
	}

	var cookiePairs []string
	for _, c := range cookies {
		cookiePairs = append(cookiePairs, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	cookieHeader := strings.Join(cookiePairs, "; ")
	log.Printf("Cookie header: %s\n", cookieHeader)

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=summaryDetail,financialData,defaultKeyStatistics,earnings&crumb=%s",
		ticker,
		crumb,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Cookie", cookieHeader)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; chromedp)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response status:", resp.Status)
	fmt.Println("Response body:", string(body))

	// Write to cache
	if err := ioutil.WriteFile(cachePath, body, 0644); err != nil {
		return nil, fmt.Errorf("could not write cache file: %v", err)
	}
	fmt.Println("Fetched and cached:", cachePath)

	// Unmarshal JSON into the struct
	err = json.Unmarshal(body, &qs)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// Now you can access the fields, for example:
	fmt.Printf("MaxAge in SummaryDetail: %d\n", qs.QuoteSummary.Result[0].SummaryDetail.MaxAge)
	fmt.Printf("Previous Close: %.2f\n", qs.QuoteSummary.Result[0].SummaryDetail.PreviousClose.Raw)

	if len(qs.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("No data returned for ticker %s", ticker)
	}

	// 7) Print key financial health metrics
	fmt.Printf("=== Financial summary for %s ===\n\n", ticker)
	data := qs.QuoteSummary.Result[0]

	fmt.Printf("Previous Close: %.2f\n", data.SummaryDetail.PreviousClose.Raw)
	fmt.Printf("Open: %.2f\n", data.SummaryDetail.Open.Raw)
	fmt.Printf("Day Range: %.2f - %.2f\n", data.SummaryDetail.DayLow.Raw, data.SummaryDetail.DayHigh.Raw)
	fmt.Printf("52-Week Range: %.2f - %.2f\n", data.SummaryDetail.FiftyTwoWeekLow.Raw, data.SummaryDetail.FiftyTwoWeekHigh.Raw)
	fmt.Printf("Market Cap: %s\n", data.SummaryDetail.MarketCap.Fmt)

	fmt.Printf("Trailing PE: %.2f\n", data.SummaryDetail.TrailingPE.Raw)
	fmt.Printf("Forward PE: %.2f\n", data.SummaryDetail.ForwardPE.Raw)
	fmt.Printf("Beta: %.2f\n", data.SummaryDetail.Beta.Raw)

	fmt.Printf("Volume: %s\n", data.SummaryDetail.Volume.Fmt)
	fmt.Printf("Average Volume: %s\n", data.SummaryDetail.AverageVolume.Fmt)

	fmt.Printf("Dividend Yield: %s\n", data.SummaryDetail.DividendYield.Fmt)
	fmt.Printf("Payout Ratio: %s\n", data.SummaryDetail.PayoutRatio.Fmt)

	fmt.Println()

	fmt.Println("=== Financial Data ===")
	fmt.Printf("Profit Margins: %.2f%%\n", data.FinancialData.ProfitMargins.Raw*100)
	fmt.Printf("Return on Assets: %.2f%%\n", data.FinancialData.ReturnOnAssets.Raw*100)
	fmt.Printf("Return on Equity: %.2f%%\n", data.FinancialData.ReturnOnEquity.Raw*100)
	fmt.Printf("Total Debt: %.2f\n", data.FinancialData.TotalDebt.Raw)
	fmt.Printf("Revenue Growth: %.2f%%\n", data.FinancialData.RevenueGrowth.Raw*100)

	fmt.Println()

	fmt.Println("=== Earnings Estimates ===")
	fmt.Printf("Current Quarter Estimate: %.2f\n", data.Earnings.EarningsChart.CurrentQuarterEstimate.Raw)
	//fmt.Printf("Current Year Estimate: %.2f\n", data.Earnings.EarningsChart.CurrentYearEstimate)

	fmt.Println("\nDone.")
	return &data, nil

}

type Metric struct {
	Name   string
	Value  string
	Color  string
	Reason string
}

func getColorForMetric(name string, value float64) (string, string) {
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

func formatMetric(name string, value *float64, isPercent bool) *Metric {
	if value == nil {
		return nil
	}
	var valueStr string
	if isPercent {
		valueStr = fmt.Sprintf("%.2f%%", *value*100)
	} else {
		// for large numbers (marketcap, ev, fcf) show compact formatting
		valueStr = formatLargeNumber(*value)
	}
	color, reason := getColorForMetric(name, *value)
	return &Metric{Name: name, Value: valueStr, Color: color, Reason: reason}
}

func formatLargeNumber(n float64) string {
	switch {
	case abs(n) >= 1_000_000_000_000:
		return fmt.Sprintf("%.2fT", n/1_000_000_000_000)
	case abs(n) >= 1_000_000_000:
		return fmt.Sprintf("%.2fB", n/1_000_000_000)
	case abs(n) >= 1_000_000:
		return fmt.Sprintf("%.2fM", n/1_000_000)
	case abs(n) >= 1_000:
		return fmt.Sprintf("%.2fK", n/1_000)
	default:
		return fmt.Sprintf("%.2f", n)
	}
}

func abs(n float64) float64 {
	if n > 0 {
		return n
	}
	return n * -1
}

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
