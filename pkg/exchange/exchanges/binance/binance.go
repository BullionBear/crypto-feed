package binance

import (
	"errors"

	"github.com/BullionBear/crypto-feed/pkg/exchange/interfaces"
)

const (
	apiURL      = "https://api.binance.com"
	candleStick = "/api/v3/klines"
)

type Binance struct {
	// Binance-specific attributes
}

func (b *Binance) FetchKLines(symbol interfaces.Symbol, interval string, startTime *int64, endTime *int64, limit *int64) ([]interfaces.KLine, error) {
	return []interfaces.KLine{}, errors.New("not implement yet")
}

func (b *Binance) SubscribeKLine(symbol interfaces.Symbol, interval string, dataHandler func(data interface{})) error {
	return errors.New("not implement yet")
}
