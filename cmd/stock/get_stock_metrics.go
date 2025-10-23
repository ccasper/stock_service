package main

import (
	"context"
	"encoding/json"
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
)

func getStockMetrics(ticker string) (*Result, error) {
	// Ensure cache dir exists. This is relative to the CWD, or
	// WorkingDirectory=/opt/stock when launched via systemd.
	cacheDir := "./data/stockdata"
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("could not create cache dir: %v", err)
	}

	// Format current time for cache filename (rounded to hour)
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

	// Create temp directories for Chrome data.
	// Required for the chrome/chromium headless request to work.
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

	// All this is chrome/chromium fiasco is to get a crumb token so we can
	// read the ticker data.
	var crumb string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(fmt.Sprintf("https://finance.yahoo.com/quote/%s", ticker)),
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(`window.YAHOO && window.YAHOO.context && window.YAHOO.context.user && window.YAHOO.context.user.crumb || ""`, &crumb),
	); err != nil {
		log.Fatal(err)
	}

	// Without a crumb token, we can't read the ticker data
	if crumb == "" {
		// Hmm, maybe Yahoo changed the page format?
		log.Fatal("crumb token not found")
	}
	log.Printf("Crumb token: %s\n", crumb)

	// Unfortunately, the crumb is not enough to fetch the data,
	// we need the cookies to pass as well.
	//
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

	// Lets make the request to get all our ticker data.
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
	// Uncomment this for debugging.
	// fmt.Println("Response body:", string(body))

	// Write to JSON blob ticker data to our cache
	if err := ioutil.WriteFile(cachePath, body, 0644); err != nil {
		return nil, fmt.Errorf("could not write cache file: %v", err)
	}
	fmt.Println("Fetched and cached:", cachePath)

	// Unmarshal JSON into the struct
	err = json.Unmarshal(body, &qs)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	if len(qs.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("no data returned for ticker %s", ticker)
	}

	data := qs.QuoteSummary.Result[0]

	return &data, nil
}
