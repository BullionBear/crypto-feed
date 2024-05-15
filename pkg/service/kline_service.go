package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	log "github.com/sirupsen/logrus"
)

var monthInSecond = 30 * 24 * int(time.Hour)

type KLineService struct {
	symbol string
	// Container
	container map[int64]*Kline
	first     int64
	last      int64
	// Dependencies
	client binance.Client
	mutex  sync.Mutex
}

func NewKLineService(symbol string, length int64) *KLineService {
	return &KLineService{
		symbol:    symbol,
		container: make(map[int64]*Kline),
		first:     0,
		last:      0,
		client:    *binance.NewClient("", ""),
	}
}

func (k *KLineService) Run(ctx context.Context) error {
	go k.subscribeKline(ctx)
	go k.requestCurrentKline(ctx)
	k.requestHistoricalKline()
	go k.removeOldestKline()
	return nil
}

func (k *KLineService) subscribeKline(ctx context.Context) error {
	for {
		doneC, _, err := binance.WsKlineServe(strings.ToLower(k.symbol), "1s", k.wsKlineHandler, k.handleError)
		if err != nil {
			log.Errorf("Unable to listen to WebSocket: %+v", err)
			select {
			case <-ctx.Done():
				log.Info("Shutdown signal received, stopping reconnection attempts")
				return ctx.Err()
			case <-time.After(1 * time.Second):
				continue
			}
		}

		select {
		case <-doneC:
			log.Info("WebSocket connection closed. Reconnecting...")
			select {
			case <-ctx.Done():
				log.Info("Shutdown signal received, stopping reconnection attempts")
				return ctx.Err()
			case <-time.After(1 * time.Second):
				// Continue to reconnect
			}
		case <-ctx.Done():
			log.Info("Shutdown signal received, closing WebSocket connection")
			return ctx.Err()
		}
	}
}

func (k *KLineService) requestCurrentKline(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			// Exit the goroutine when the context is cancelled
			log.Info("Stopping periodic requests")
			return
		case <-ticker.C:
			// Perform the request
			srv := k.client.NewKlinesService()
			srv.Symbol(strings.ToUpper(k.symbol))
			srv.Limit(5) // The latest recent 5
			srv.Interval("1s")
			bKlines, err := srv.Do(context.Background())
			if err != nil {
				log.Errorf("Fail retrieve kline: %+v", err)
			}
			for _, bKline := range bKlines {
				kline, err := convertFromKline(bKline)
				if err != nil {
					log.Errorf("Fail convert kline: %+v", bKline)
				}
				k.insertLastestKline(kline)
			}
		}
	}
}

func (k *KLineService) requestHistoricalKline() {
	now := time.Now().UTC().Second()
	start := now - monthInSecond
	srv := k.client.NewKlinesService()
	srv.Symbol(strings.ToUpper(k.symbol))
	srv.Interval("1s")
	interval := 600

	for last := now; last >= start; last -= interval {
		srv.StartTime(int64(last-interval) * 1000)
		srv.EndTime(int64(last))
		bklines, err := srv.Do(context.Background())
		if err != nil {
			log.Errorf("Fail to retrieve historical klines %s", err.Error())
		}
		for i := len(bklines) - 1; i >= 0; i-- {
			bkline := bklines[i]
			kline, err := convertFromKline(bkline)
			if err != nil {
				log.Printf("Fail to convert klines %s", err.Error())
				continue
			}
			k.insertHistoricalKline(kline)
		}
	}
}

func (k *KLineService) removeOldestKline(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			// Exit the goroutine when the context is cancelled
			log.Info("Stopping periodic remove kline")
			return
		case <-ticker.C:
			cutline := k.last - int64(monthInSecond)*1000
			for ts := k.first; ts < cutline; ts += 1000 {
				delete(k.container, ts)
			}
		}
	}
}

func (k *KLineService) wsKlineHandler(event *binance.WsKlineEvent) {
	wsKline := event.Kline
	kline, err := convertFromWsKline(&wsKline)
	if err != nil {
		log.Errorf("Unable to convert wsKline %+v", wsKline)
	}
	k.insertLastestKline(kline)
}

func (k *KLineService) insertLastestKline(kline *Kline) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	closeTime := kline.CloseTime
	k.container[closeTime] = kline
	k.last = max(closeTime, k.last)
}

func (k *KLineService) insertHistoricalKline(kline *Kline) {
	closeTime := kline.CloseTime
	k.container[closeTime] = kline
	k.first = min(closeTime, k.first)
}

func (k *KLineService) handleError(err error) {
	log.Errorf("Error: %s", err.Error())
}
