package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/BullionBear/crypto-feed/pkg/linkedlist"
	"github.com/adshao/go-binance/v2"
	log "github.com/sirupsen/logrus"
)

// var monthInSecond = 30 * 24 * int(time.Hour)

type KLineService struct {
	symbol string
	// Container
	container linkedlist.IndexLinkedList[Kline]
	// Dependencies
	client binance.Client
	// Subscriber
	id          int64
	subscribers map[int64]func(*Kline)
	// pipeline control
	mutex     sync.Mutex
	controlCh chan struct{}
}

func NewKLineService(symbol string, length int64) *KLineService {
	return &KLineService{
		symbol:      symbol,
		container:   *linkedlist.NewIndexedLinkedList[Kline](),
		client:      *binance.NewClient("", ""),
		id:          0,
		subscribers: make(map[int64]func(*Kline)),
		controlCh:   make(chan struct{}, 1),
	}
}

func (srv *KLineService) Run() error {
	go srv.requestCurrentKline()
	go srv.publishKline()
	// go srv.subscribeKline(ctx)
	// go srv.requestCurrentKline(ctx)
	// srv.requestHistoricalKline()
	// go srv.removeOldestKline(ctx)
	return nil
}

func (srv *KLineService) Subscribe(handler func(event *Kline)) int64 {
	id := srv.id
	srv.subscribers[id] = handler
	srv.id++
	return id
}

func (srv *KLineService) Unsubscribe(subscriberID int64) error {
	delete(srv.subscribers, subscriberID)
	return nil
}

// unc (srv *KLineService) subscribeKline(ctx context.Context) error {
// 	for {
// 		doneC, _, err := binance.WsKlineServe(strings.ToLower(srv.symbol), "1s", srv.wsKlineHandler, srv.handleError)
// 		if err != nil {
// 			log.Errorf("Unable to listen to WebSocket: %+v", err)
// 			select {
// 			case <-ctx.Done():
// 				log.Info("Shutdown signal received, stopping reconnection attempts")
// 				return ctx.Err()
// 			case <-time.After(1 * time.Second):
// 				continue
// 			}
// 		}
//
// 		select {
// 		case <-doneC:
// 			log.Info("WebSocket connection closed. Reconnecting...")
// 			select {
// 			case <-ctx.Done():
// 				log.Info("Shutdown signal received, stopping reconnection attempts")
// 				return ctx.Err()
// 			case <-time.After(1 * time.Second):
// 				// Continue to reconnect
// 			}
// 		case <-ctx.Done():
// 			log.Info("Shutdown signal received, closing WebSocket connection")
// 			return ctx.Err()
// 		}
// 	}
//

func (srv *KLineService) requestCurrentKline() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Perform the request
		req := srv.client.NewKlinesService()
		req.Symbol(strings.ToUpper(srv.symbol))
		req.Limit(5) // The latest recent 5
		req.Interval("1s")
		bKlines, err := req.Do(context.Background())
		if err != nil {
			log.Errorf("Fail to retrieve kline: %+v", err)
			continue
		}

		for _, bKline := range bKlines {
			kline, err := convertFromKline(bKline)
			if err != nil {
				log.Errorf("Fail to convert kline: %+v", bKline)
				continue
			}
			if err := srv.insertLastestKline(kline); err == nil {
				srv.controlCh <- struct{}{}
				continue
			}
		}
	}
}

func (srv *KLineService) publishKline() {
	for range srv.controlCh {
		kline, _ := srv.container.Tail()
		for _, subscriber := range srv.subscribers {
			subscriber(&kline)
		}
	}
}

// func (srv *KLineService) requestHistoricalKline() {
// 	now := time.Now().UTC().Second()
// 	start := now - monthInSecond
// 	srv := srv.client.NewKlinesService()
// 	srv.Symbol(strings.ToUpper(srv.symbol))
// 	srv.Interval("1s")
// 	interval := 600
//
// 	for last := now; last >= start; last -= interval {
// 		srv.StartTime(int64(last-interval) * 1000)
// 		srv.EndTime(int64(last))
// 		bklines, err := srv.Do(context.Background())
// 		if err != nil {
// 			log.Errorf("Fail to retrieve historical klines %s", err.Error())
// 		}
// 		for i := len(bklines) - 1; i >= 0; i-- {
// 			bkline := bklines[i]
// 			kline, err := convertFromKline(bkline)
// 			if err != nil {
// 				log.Printf("Fail to convert klines %s", err.Error())
// 				continue
// 			}
// 			srv.insertHistoricalKline(kline)
// 		}
// 	}
// }

// func (srv *KLineService) removeOldestKline(ctx context.Context) {
//
// }
//
// func (srv *KLineService) wsKlineHandler(event *binance.WsKlineEvent) {
//
// }

func (srv *KLineService) insertLastestKline(kline *Kline) error {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	closeTime := kline.CloseTime
	return srv.container.PushBack(closeTime, *kline)
}

// func (srv *KLineService) insertHistoricalKline(kline *Kline) {
//
// }
//
// func (srv *KLineService) handleError(err error) {
// 	log.Errorf("Error: %s", err.Error())
// }
