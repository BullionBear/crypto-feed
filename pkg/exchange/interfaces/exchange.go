package interfaces

type Exchange interface {
	FetchKLines(symbol Symbol, interval string, startTime *int64, endTime *int64, limit *int64) ([]KLine, error)

	SubscribeKLine(symbol Symbol, interval string, dataHandler func(data interface{})) error
}
