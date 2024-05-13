package interfaces

type Exchange interface {
	FetchKLines(symbol Symbol, interval string, startTime *int64, endTime *int64, limit *int64) ([]KLine, error)
	FetchOrderBook(symbol string) (OrderBook, error)
	PlaceOrder(order Order) (OrderResponse, error)
}
