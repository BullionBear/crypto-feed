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
	status      Status
	errCh       chan struct{}
	// pipeline control
	mutex   sync.RWMutex
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
		status:      StatusCreated,
		errCh:       make(chan struct{}),
		eventCh:     make(chan struct{}, 1),
		isSetup:     false,
	}
}

func (srv *KLineService) Run() error {
	srv.status = StatusInitializing
	setupCh := make(chan struct{})
	go srv.publishKline(setupCh)
	go srv.requestCurrentKline()
	<-setupCh
	log.Info("Received First Kline")
	go srv.requestHistoricalKline(setupCh)
	<-setupCh
	log.Info("Finish retrieve historical klines")
	go srv.popHistoricalKline()
	go srv.subscribeCurrentKline()
	srv.status = StatusRunning
	go func() {
		for range srv.errCh {
			srv.status = StatusError
			log.Errorf("some goroutine dead, need to check")
		}
	}()
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

func (srv *KLineService) Status() Status {
	return srv.status
}

func (srv *KLineService) Subscribe(handler func(event *Kline)) int64 {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	id := srv.id
	srv.subscribers[id] = handler
	srv.id++
	return id
}

func (srv *KLineService) Unsubscribe(subscriberID int64) error {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	delete(srv.subscribers, subscriberID)
	return nil
}

func (srv *KLineService) ListSubsriber() []int64 {
	result := []int64{}
	srv.mutex.RLock()
	defer srv.mutex.RUnlock()
	for key := range srv.subscribers {
		result = append(result, key)
	}
	return result
}

func (srv *KLineService) Query(start int64, end int64, handler func(event *Kline)) error {
	key := start
	for {
		kline, err := srv.container.Get(key)
		if err != nil {
			log.Errorf("Failed to query key %d: %s", key, err.Error())
			return err
		}
		handler(&kline)

		if key == end {
			break
		}

		if key, err = srv.container.Next(key); err != nil {
			log.Errorf("Failed to get next key after %d: %s", key, err.Error())
			return err
		}
	}
	return nil
}

func (srv *KLineService) subscribeCurrentKline() {
	log.Info("start subscribe current kline")
	var wsKlineHandler = func(event *binance.WsKlineEvent) {
		kline, err := convertFromWsKline(&event.Kline)
		if err != nil {
			log.Errorf("fail to convert wsKline %s", err.Error())
		}
		srv.pushBack(kline)
	}
	var errHandler = func(err error) {
		log.Errorf("handle error of wsKline %s", err.Error())
	}
	reconnectCh := make(chan struct{}, 1)
	reconnectCh <- struct{}{}
	for range reconnectCh {
		doneC, _, err := binance.WsKlineServe(srv.symbol, "1s", wsKlineHandler, errHandler)
		if err != nil {
			log.Errorf("fail to create ws channel: %s", err.Error())
		}
		time.Sleep(5 * time.Second)
		<-doneC
		reconnectCh <- struct{}{}
	}
	srv.errCh <- struct{}{}
}

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
	srv.errCh <- struct{}{}
}

func (srv *KLineService) publishKline(setupCh chan<- struct{}) {
	for range srv.eventCh {
		if !srv.isSetup {
			currentTime, err := srv.container.HeadKey(0) // Initialize running Key
			if err != nil {
				log.Warnf("unable to retrieve currentTime")
			}
			srv.currentTime = currentTime
			srv.isSetup = true
			setupCh <- struct{}{}
		}
		for {
			nextTime, err := srv.container.Next(srv.currentTime)
			if err != nil {
				break
			}
			kline, _ := srv.container.Get(srv.currentTime)
			srv.mutex.RLock()
			for _, subscriber := range srv.subscribers {
				subscriber(&kline)
			}
			srv.mutex.RUnlock()
			srv.currentTime = nextTime
		}
	}
	srv.errCh <- struct{}{}
}

func (srv *KLineService) requestHistoricalKline(setupCh chan<- struct{}) {

	ksrv := srv.client.NewKlinesService()
	limit := 1000
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
		startTime := endTime - int64(limit*1000) // rollback 500 seconds
		ksrv.StartTime(startTime - 1)
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
		time.Sleep(30 * time.Millisecond) // avoid reach request rate limit
	}
	setupCh <- struct{}{}
}

func (srv *KLineService) popHistoricalKline() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		nPop := srv.Size() - srv.length
		if nPop <= 0 {
			continue
		}
		for i := 0; i < int(nPop); i++ {
			_, err := srv.container.PopFront()
			if err != nil {
				log.Errorf("fail to pop kline %s", err.Error())
			}
		}
	}
	srv.errCh <- struct{}{}
}

func (srv *KLineService) pushBack(kline *Kline) error {
	closeTime := kline.OpenTime
	return srv.container.PushBack(closeTime, *kline)
}

func (srv *KLineService) pushFront(kline *Kline) error {
	closeTime := kline.OpenTime
	return srv.container.PushFront(closeTime, *kline)
}
