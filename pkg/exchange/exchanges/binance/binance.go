package binance

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/BullionBear/crypto-feed/pkg/exchange/interfaces"
	log "github.com/sirupsen/logrus"
)

const (
	// Base Url
	apiUrl = "https://api.binance.com"
	// Endpoint
	candleStick = "/api/v3/klines"
)

type Binance struct {
	// Binance-specific attributes
}

func (b *Binance) FetchKLines(symbol interfaces.Symbol, interval string, startTime *int64, endTime *int64, limit *int64) ([]interfaces.KLine, error) {
	var klines []interfaces.KLine

	// Construct the request URL with query parameters
	u, err := url.Parse(apiUrl + candleStick)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("symbol", symbol.Base+symbol.Quote)
	q.Set("interval", interval)
	if startTime != nil {
		q.Set("startTime", strconv.FormatInt(*startTime, 10))
	}
	if endTime != nil {
		q.Set("endTime", strconv.FormatInt(*endTime, 10))
	}
	if limit != nil {
		q.Set("limit", strconv.FormatInt(*limit, 10))
	}
	u.RawQuery = q.Encode()

	// Send the request
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Handle non-200 status
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("API request failed with status " + resp.Status)
	}

	// Unmarshal JSON data
	var rawKLines []KLineResponse
	if err := json.Unmarshal(body, &rawKLines); err != nil {
		return nil, err
	}

	// Convert raw data to your interface type
	for _, rawKLine := range rawKLines {
		open, err := strconv.ParseFloat(rawKLine.Open, 64)
		if err != nil {
			log.Warn("Error converting Open value: ", err)
			continue // or return, depending on how you want to handle the error
		}

		high, err := strconv.ParseFloat(rawKLine.High, 64)
		if err != nil {
			log.Warn("Error converting High value: ", err)
			continue
		}

		low, err := strconv.ParseFloat(rawKLine.Low, 64)
		if err != nil {
			log.Warn("Error converting Low value: ", err)
			continue
		}

		closeVal, err := strconv.ParseFloat(rawKLine.Close, 64)
		if err != nil {
			log.Warn("Error converting Close value: ", err)
			continue
		}

		volume, err := strconv.ParseFloat(rawKLine.Volume, 64)
		if err != nil {
			log.Warn("Error converting Volume value: ", err)
			continue
		}

		kline := interfaces.KLine{
			Symbol:    symbol,
			Timestamp: rawKLine.CloseTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closeVal,
			Volume:    volume,
		}
		klines = append(klines, kline)
	}
	return klines, nil
}
