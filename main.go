package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var googleDomains = map[string]string{
	"com": "https://www.google.com/search?q=",
	"cn":  "https://www.google.cn/search?q=",
	"fr":  "https://www.google.fr/search?q=",
	"de":  "https://www.google.de/search?q=",
	"ru":  "https://www.google.ru/search?q=",
	"jp":  "https://www.google.co.jp/search?q=",
	"es":  "https://www.google.es/search?q=",
	"it":  "https://www.google.it/search?q=",
	"br":  "https://www.google.com.br/search?q=",
	"mx":  "https://www.google.com.mx/search?q=",
	"ca":  "https://www.google.ca/search?q=",
	"uk":  "https://www.google.co.uk/search?q=",
	"in":  "https://www.google.co.in/search?q=",
	"au":  "https://www.google.com.au/search?q=",
	"ar":  "https://www.google.com.ar/search?q=",
	"ch":  "https://www.google.ch/search?q=",
	"co":  "https://www.google.com.co/search?q=",
	"dk":  "https://www.google.dk/search?q=",
}

type SearchResult struct {
	ResultRank  int
	ResultURL   string
	ResultTitle string
	ResultDesc  string
}

var userAgents = []string{
	"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
}

func randomUserAgent() string {
	randNum := rand.Int() % len(userAgents)
	return userAgents[randNum]
}

func buildGoogleUrls(searchTerm, countryCode, languageCode string, pages, count int) ([]string, error) {
	toScrape := []string{}
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)
	if googleBase, found := googleDomains[countryCode]; found {
		for i := 0; i < pages; i++ {
			start := i * count
			scrapeURL := fmt.Sprintf("%s%s&num=%dhl=%s&start=%d&filter=0", googleBase, searchTerm, count, languageCode, start)
			toScrape = append(toScrape, scrapeURL)
		}
	} else {
		err := fmt.Errorf("no Google domain found for country code: %s", countryCode)
		return nil, err
	}
	return toScrape, nil
}

func GoogleScrape(searchTerm, languageCode, countryCode string, proxyString interface{}, pages int, count int, backoff int) ([]SearchResult, error) {
	results := []SearchResult{}
	resultCounter := 0
	googlePages, err := buildGoogleUrls(searchTerm, countryCode, languageCode, pages, count)
	fmt.Println("googlePages", googlePages)
	if err != nil {
		return nil, err
	}
	for _, page := range googlePages {
		res, err := scrapeClientRequest(page, proxyString)
		if err != nil {
			return nil, err
		}
		data, err := googleResultParsing(res, resultCounter)
		if err != nil {
			return nil, err
		}
		resultCounter += len(data)
		// for _, result := range data {
		results = append(results, data...)
		// }
		time.Sleep(time.Duration(backoff) * time.Second)
	}
	return results, nil
}

func googleResultParsing(response *http.Response, rank int) ([]SearchResult, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	results := []SearchResult{}
	sel := doc.Find("div.g")
	fmt.Println("sel", sel.Nodes)
	rank++
	for i := range sel.Nodes {
		fmt.Println("y3434")
		item := sel.Eq(i)
		linkTag := item.Find("a")
		link, _ := linkTag.Attr("href")
		titleTag := item.Find("h3.r")
		descTag := item.Find("span.st")
		desc := descTag.Text()
		title := titleTag.Text()
		link = strings.Trim(link, " ")
		fmt.Println("link", link)
		if link != "" && link != "#" && strings.HasPrefix(link, "/") {
			result := SearchResult{
				rank,
				link,
				title,
				desc,
			}
			results = append(results, result)
			rank++
		}
	}
	return results, nil
}

func getScrapeClient(proxyString interface{}) *http.Client {
	switch v := proxyString.(type) {
	case string:
		proxyURL, _ := url.Parse(v)
		return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
	default:
		return &http.Client{}
	}
}

func scrapeClientRequest(searchURL string, proxyString interface{}) (*http.Response, error) {
	baseClient := getScrapeClient(proxyString)
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", randomUserAgent())

	res, err := baseClient.Do(req)
	if res.StatusCode != 200 {
		err := fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func main() {
	res, err := GoogleScrape("golang", "en", "com", nil, 1, 30, 10)
	fmt.Println("res", res)
	if err == nil {
		for _, result := range res {
			fmt.Println(result)
		}
	}
}
