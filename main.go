package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"log"
	"io"
	"sort"
	"strconv"
)

type ApiResponse struct {
    MetaData            MetaData                       `json:"Meta Data"`
    TimeSeriesDaily     map[string]DigitalCurrencyDay  `json:"Time Series (Digital Currency Daily)"`
}

type MetaData struct {
    Information         string `json:"1. Information"`
    DigitalCurrencyCode string `json:"2. Digital Currency Code"`
    DigitalCurrencyName string `json:"3. Digital Currency Name"`
    MarketCode          string `json:"4. Market Code"`
    MarketName          string `json:"5. Market Name"`
    LastRefreshed       string `json:"6. Last Refreshed"`
    TimeZone            string `json:"7. Time Zone"`
}

type DigitalCurrencyDay struct {
    OpenUSD   string `json:"1b. open (USD)"`
    HighUSD   string `json:"2b. high (USD)"`
    LowUSD    string `json:"3b. low (USD)"`
    CloseUSD  string `json:"4b. close (USD)"`
    Volume    string `json:"5. volume"`
    MarketCap string `json:"6. market cap (USD)"`
}

func fetchData(symbol, market, apiKey string) ([]string, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=DIGITAL_CURRENCY_DAILY&symbol=%v&market=%v&apikey=%v", symbol, market, apiKey)

	response, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var apiResponse ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}

	var dates []string
	for date := range apiResponse.TimeSeriesDaily {
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	var closePrices []string
	for _, date := range dates {
		closePrices = append(closePrices, apiResponse.TimeSeriesDaily[date].CloseUSD)
	}

	return closePrices, nil
}

func calculateSMA(closePrices []string, n int, gap int) float64 {
	endIndex := n + gap
	periods := closePrices[gap:endIndex]

	var closingSum float64 = 0

	for _, price := range periods {
		closingValue, err := strconv.ParseFloat(price, 64)
		if err != nil {
			log.Println("Error converting string to int:", err)
			continue
		}
		closingSum += closingValue
	}

	nFloat := float64(n)
	SMA := closingSum / nFloat
	
	return SMA
}

func calculateEMA(closePrices []string, n int) ([]float64, error) {
    if len(closePrices) < n {
        return nil, fmt.Errorf("not enough data points to calculate starting SMA")
    }

    startingSMA := calculateSMA(closePrices, n, 0)
    multiplier := 2.0 / float64(n+1)
    EMAs := make([]float64, len(closePrices))

    EMAs[0] = startingSMA
    for i := 1; i < len(closePrices); i++ {
        price, err := strconv.ParseFloat(closePrices[i], 64)
        if err != nil {
            return nil, fmt.Errorf("error converting closing price to float: %v", err)
        }
        EMAs[i] = (price - EMAs[i-1]) * multiplier + EMAs[i-1]
    }

    return EMAs, nil
}

func main() {
	symbol := "SOL"
	market := "USD"
	apiKey := "KTAWNDLPXLT38BS2"

	closePrices, _ := fetchData(symbol, market, apiKey)
	currentPriceString := closePrices[0]
	currentPrice, _ := strconv.ParseFloat(currentPriceString, 64)
	fmt.Println("Current price: ", currentPrice)

	shortSMA := calculateSMA(closePrices, 50, 0)
	longSMA := calculateSMA(closePrices, 200, 0)

	fmt.Println("Short and long SMAs:", shortSMA, longSMA)

	fmt.Println("SMA crossing strategy (50/200 days data):")
	if shortSMA > longSMA {
		fmt.Println("Signal: buy")
	} else if shortSMA < longSMA {
		fmt.Println("Signal: sell")
	}

	shortEMAs, _ := calculateEMA(closePrices, 50)
	longEMAs, _ := calculateEMA(closePrices, 200)

	fmt.Println("Short and long EMAs:", shortEMAs[len(shortEMAs)-1], longEMAs[len(longEMAs)-1])

	fmt.Println("EMA crossing strategy (50/200 days data):")
	if shortEMAs[len(shortEMAs)-1] > longEMAs[len(longEMAs)-1] {
		fmt.Println("Signal: buy")
	} else if shortEMAs[len(shortEMAs)-1] < longEMAs[len(longEMAs)-1] {
		fmt.Println("Signal: sell")
	}
	
}