package interfaces

type Symbol struct {
	Base  string
	Quote string
}

type KLine struct {
	Symbol    Symbol
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

type KLineHandler func(event *KLine)

type ErrHandler func(err error)
