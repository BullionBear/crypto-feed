package interfaces

type Exchange interface {
	FetchOrderBook(symbol string) (OrderBook, error)
	PlaceOrder(order Order) (OrderResponse, error)
}
