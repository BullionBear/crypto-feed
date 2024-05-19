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
	log.Info("Finish retrieve historical klines")
	go srv.popHistoricalKline()
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

func (srv *KLineService) SubscribeFromHead(handler func(event *Kline)) int64 {
	middleCh := make(chan *Kline, 1024)
	isFirst := true
	firstCh := make(chan int64)
	var middleHandler = func(event *Kline) {
		middleCh <- event
		if isFirst {
			isFirst = false
			firstCh <- event.CloseTime // Get Key
			close(firstCh)
		}
	}
	id := srv.id
	srv.Subscribe(middleHandler)
	go func() {
		end := <-firstCh
		start, err := srv.container.HeadKey(2)
		if err != nil {
			for key := start; key != end; key, _ = srv.container.Next(key) {
				kline, err := srv.container.Get(key)
				if err != nil {
					log.Errorf("Fail retrieve kline: %s", err.Error())
				}
				handler(&kline)
			}
		}
		for kline := range middleCh {
			handler(kline)
		}
	}()
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
			for _, subscriber := range srv.subscribers {
				subscriber(&kline)
			}

			srv.currentTime = nextTime
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
}

func (srv *KLineService) pushBack(kline *Kline) error {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	closeTime := kline.OpenTime
	return srv.container.PushBack(closeTime, *kline)
}

func (srv *KLineService) pushFront(kline *Kline) error {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	closeTime := kline.OpenTime
	return srv.container.PushFront(closeTime, *kline)
}
