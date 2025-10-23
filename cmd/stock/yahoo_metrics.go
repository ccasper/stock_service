package main

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