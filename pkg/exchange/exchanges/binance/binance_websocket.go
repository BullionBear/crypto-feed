package binance

import (
	"github.com/BullionBear/crypto-feed/pkg/exchange/interfaces"
)

const (
	wsUrl = "wss://stream.binance.com:9443/ws"
)

func (b *Binance) SubscribeKLine(symbol interfaces.Symbol, interval string, handler interfaces.KLineHandler, errHandler interfaces.ErrHandler) (doneC, stopC chan struct{}, err error) {
	// channel := fmt.Sprint("/%s@kline_%s", strings.ToLower(symbol.Base)+strings.ToLower(symbol.Quote), interval)
	return nil, nil, nil
}
