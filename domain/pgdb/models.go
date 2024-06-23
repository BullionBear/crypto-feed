package pgdb

type PlaybackKline struct {
	OpenTime            int64
	Open                float64
	High                float64
	Low                 float64
	Close               float64
	Volume              float64
	CloseTime           int64
	QuoteVolume         float64
	Count               int64
	TakerBuyVolume      float64
	TakerBuyQuoteVolume float64
	Ignore              int64
}

func (PlaybackKline) TableName() string {
	return "btcusdt_kline_1s"
}
