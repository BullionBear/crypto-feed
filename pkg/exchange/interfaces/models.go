package interfaces

// Order represents a trading order
type Order struct {
	ID       string
	Symbol   string
	Quantity float64
	Price    float64
	Side     string // "buy" or "sell"
}

// OrderResponse represents a response from an exchange after placing an order
type OrderResponse struct {
	ID     string
	Status string
	Filled float64
}

// OrderBook represents market depth data
type OrderBook struct {
	Bids []OrderBookEntry
	Asks []OrderBookEntry
}

// OrderBookEntry represents a single entry in an OrderBook
type OrderBookEntry struct {
	Price    float64
	Quantity float64
}
