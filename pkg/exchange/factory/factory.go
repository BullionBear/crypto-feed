package factory

import (
	"github.com/BullionBear/crypto-feed/pkg/exchange/exchanges"
	"github.com/BullionBear/crypto-feed/pkg/exchange/interfaces"
)

type ExchangeFactory struct{}

func (f *ExchangeFactory) GetExchange(name string) interfaces.Exchange {
	switch name {
	case "binance":
		return &exchanges.Binance{}

	default:
		return nil // or some default implementation
	}
}
