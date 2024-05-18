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

type KLineService struct {
	symbol string
	length int64
	// Container
	container linkedlist.IndexLinkedList[Kline]
	// Dependencies
	client binance.Client
	// Subscriber
	id          int64
	subscribers map[int64]func(*Kline)
	// Dynamic varaible
	currentTime int64
	// pipeline control
	mutex   sync.Mutex
	eventCh chan struct{}
	// init control
	isSetup bool
}

func NewKLineService(symbol string, length int64) *KLineService {
	return &KLineService{
		symbol:      symbol,
		length:      length,
		container:   *linkedlist.NewIndexedLinkedList[Kline](),
		client:      *binance.NewClient("", ""),
		id:          0,
		subscribers: make(map[int64]func(*Kline)),
		currentTime: 0,
		eventCh:     make(chan struct{}, 1),
		isSetup:     false,
	}
}

func (srv *KLineService) Run() error {
	setupCh := make(chan struct{})
	go srv.publishKline(setupCh)
	go srv.requestCurrentKline()
	<-setupCh
	log.Info("Received First Kline")
	go srv.requestHistoricalKline(setupCh)
	<-setupCh
	log.Info("Request target historical data")
	// go srv.popHistoricalKline()
	return nil
}

func (srv *KLineService) Symbol() string {
	return srv.symbol
}

func (srv *KLineService) Length() int64 {
	return srv.length
}

func (srv *KLineService) Head() (Kline, error) {
	return srv.container.Head()
}

func (srv *KLineService) Tail() (Kline, error) {
	return srv.container.Tail()
}

func (srv *KLineService) Size() int64 {
	return srv.container.Size()
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

func (srv *KLineService) ListSubsriber() []int64 {
	result := []int64{}
	for key := range srv.subscribers {
		result = append(result, key)
	}
	return result
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
	ksrv := srv.client.NewKlinesService()
	ksrv.Symbol(strings.ToUpper(srv.symbol))
	ksrv.Limit(5) // The latest recent 5
	ksrv.Interval("1s")
	for range ticker.C {
		// Perform the request
		if srv.isSetup {
			startTime, err := srv.container.Tail()
			if err != nil {
				log.Errorf("Fail to get latest data")
				continue
			}
			ksrv.StartTime(startTime.CloseTime)
			ksrv.EndTime(time.Now().UTC().UnixMilli())
		}

		bKlines, err := ksrv.Do(context.Background())
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
			if err := srv.pushBack(kline); err == nil {
				srv.eventCh <- struct{}{}
				continue
			}
		}
	}
}

func (srv *KLineService) publishKline(setupCh chan<- struct{}) {
	for range srv.eventCh {
		if !srv.isSetup {
			srv.isSetup = true
			setupCh <- struct{}{}
		}
		kline, _ := srv.container.Tail()
		for _, subscriber := range srv.subscribers {
			subscriber(&kline)
		}
	}
}

func (srv *KLineService) requestHistoricalKline(setupCh chan<- struct{}) {
	ksrv := srv.client.NewKlinesService()
	limit := 500
	ksrv.Symbol(strings.ToUpper(srv.symbol))
	ksrv.Interval("1s")
	ksrv.Limit(limit)
	for size := srv.container.Size(); size < srv.length; size = srv.container.Size() {
		// log.Infof("Current Size: %d", size)
		startKline, err := srv.container.Head()
		if err != nil {
			log.Errorf("fail get head kline %s", err.Error())
		}
		endTime := startKline.OpenTime
		startTime := endTime - int64(limit*1000) // back 500 seconds
		ksrv.StartTime(startTime)
		ksrv.EndTime(endTime)
		bklines, err := ksrv.Do(context.Background())
		if err != nil {
			log.Errorf("Fail to retrieve historical klines %s", err.Error())
		}
		for i := len(bklines) - 1; i >= 0; i-- {
			bkline := bklines[i]
			kline, err := convertFromKline(bkline)
			if err != nil {
				log.Errorf("Fail to convert klines %s", err.Error())
				continue
			}
			if err := srv.pushFront(kline); err != nil {
				log.Errorf("Fail to push front kline %+v", kline)
			}
		}
	}
	setupCh <- struct{}{}
}

// func (srv *KLineService) removeOldestKline(ctx context.Context) {
//
// }
//
// func (srv *KLineService) wsKlineHandler(event *binance.WsKlineEvent) {
//
// }

func (srv *KLineService) pushBack(kline *Kline) error {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	closeTime := kline.CloseTime
	return srv.container.PushBack(closeTime, *kline)
}

func (srv *KLineService) pushFront(kline *Kline) error {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	closeTime := kline.CloseTime
	return srv.container.PushFront(closeTime, *kline)
}

// func (srv *KLineService) insertHistoricalKline(kline *Kline) {
//
// }
//
// func (srv *KLineService) handleError(err error) {
// 	log.Errorf("Error: %s", err.Error())
// }
