package service

import (
	"strconv"

	"github.com/adshao/go-binance/v2"
)

type Status string

var (
	StatusCreated      = Status("created")
	StatusInitializing = Status("initializing")
	StatusRunning      = Status("running")
	StatusError        = Status("error")
)

type Kline struct {
	OpenTime                 int64   `json:"openTime"`
	Open                     float64 `json:"open"`
	High                     float64 `json:"high"`
	Low                      float64 `json:"low"`
	Close                    float64 `json:"close"`
	Volume                   float64 `json:"volume"`
	CloseTime                int64   `json:"closeTime"`
	QuoteAssetVolume         float64 `json:"quoteAssetVolume"`
	TradeNum                 int64   `json:"tradeNum"`
	TakerBuyBaseAssetVolume  float64 `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume float64 `json:"takerBuyQuoteAssetVolume"`
}

func convertFromKline(bKline *binance.Kline) (*Kline, error) {
	// Helper function to convert string to float64
	strToFloat64 := func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	}

	open, err := strToFloat64(bKline.Open)
	if err != nil {
		return nil, err
	}

	high, err := strToFloat64(bKline.High)
	if err != nil {
		return nil, err
	}

	low, err := strToFloat64(bKline.Low)
	if err != nil {
		return nil, err
	}

	close, err := strToFloat64(bKline.Close)
	if err != nil {
		return nil, err
	}

	volume, err := strToFloat64(bKline.Volume)
	if err != nil {
		return nil, err
	}

	quoteAssetVolume, err := strToFloat64(bKline.QuoteAssetVolume)
	if err != nil {
		return nil, err
	}

	takerBuyBaseAssetVolume, err := strToFloat64(bKline.TakerBuyBaseAssetVolume)
	if err != nil {
		return nil, err
	}

	takerBuyQuoteAssetVolume, err := strToFloat64(bKline.TakerBuyQuoteAssetVolume)
	if err != nil {
		return nil, err
	}

	return &Kline{
		OpenTime:                 bKline.OpenTime,
		Open:                     open,
		High:                     high,
		Low:                      low,
		Close:                    close,
		Volume:                   volume,
		CloseTime:                bKline.CloseTime,
		QuoteAssetVolume:         quoteAssetVolume,
		TradeNum:                 bKline.TradeNum,
		TakerBuyBaseAssetVolume:  takerBuyBaseAssetVolume,
		TakerBuyQuoteAssetVolume: takerBuyQuoteAssetVolume,
	}, nil
}

func convertFromWsKline(wsKline *binance.WsKline) (*Kline, error) {
	// Helper function to convert string to float64
	strToFloat64 := func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	}

	open, err := strToFloat64(wsKline.Open)
	if err != nil {
		return nil, err
	}

	high, err := strToFloat64(wsKline.High)
	if err != nil {
		return nil, err
	}

	low, err := strToFloat64(wsKline.Low)
	if err != nil {
		return nil, err
	}

	close, err := strToFloat64(wsKline.Close)
	if err != nil {
		return nil, err
	}

	volume, err := strToFloat64(wsKline.Volume)
	if err != nil {
		return nil, err
	}

	quoteVolume, err := strToFloat64(wsKline.QuoteVolume)
	if err != nil {
		return nil, err
	}

	activeBuyVolume, err := strToFloat64(wsKline.ActiveBuyVolume)
	if err != nil {
		return nil, err
	}

	activeBuyQuoteVolume, err := strToFloat64(wsKline.ActiveBuyQuoteVolume)
	if err != nil {
		return nil, err
	}

	return &Kline{
		OpenTime:                 wsKline.StartTime,
		Open:                     open,
		High:                     high,
		Low:                      low,
		Close:                    close,
		Volume:                   volume,
		CloseTime:                wsKline.EndTime,
		QuoteAssetVolume:         quoteVolume,
		TradeNum:                 wsKline.TradeNum,
		TakerBuyBaseAssetVolume:  activeBuyVolume,
		TakerBuyQuoteAssetVolume: activeBuyQuoteVolume,
	}, nil
}
