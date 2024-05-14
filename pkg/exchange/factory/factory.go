package factory

import (
	"github.com/BullionBear/crypto-feed/pkg/exchange/exchanges/binance"
	"github.com/BullionBear/crypto-feed/pkg/exchange/interfaces"
)

type ExchangeFactory struct{}

func (f *ExchangeFactory) GetExchange(name string) interfaces.IExchange {
	switch name {
	case "binance":
		return &binance.Binance{}

	default:
		return nil // or some default implementation
	}
}
