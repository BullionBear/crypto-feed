package interfaces

type IExchange interface {
	FetchKLines(symbol Symbol, interval string, startTime *int64, endTime *int64, limit *int64) ([]KLine, error)

	SubscribeKLine(symbol Symbol, interval string, dataHandler KLineHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error)
}
