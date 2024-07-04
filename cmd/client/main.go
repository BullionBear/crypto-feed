package main

import (
	"context"
	"io"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/BullionBear/crypto-feed/api/gen/feed"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	serverAddr = "localhost:50051"
)

func main() {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := feed.NewFeedClient(conn)

	// Call GetConfig
	getConfig(c)

	// Call GetStatus
	getStatus(c)

	// Call GetSubscriber
	getSubscriber(c)

	// Subscribe to Kline stream
	// subscribeKline(c)

	// Read historical Kline
	readHistoricalKline(c, 1682899000000, 1682899200000) // Example timestamps
}

func getConfig(c feed.FeedClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetConfig(ctx, &emptypb.Empty{})
	if err != nil {
		log.Printf("could not get config: %v", status.Convert(err).Message())
		return
	}
	log.Printf("Config: %v", r)
}

func getStatus(c feed.FeedClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetStatus(ctx, &emptypb.Empty{})
	if err != nil {
		log.Printf("could not get status: %v", status.Convert(err).Message())
		return
	}
	log.Printf("Status: %v", r)
}

func getSubscriber(c feed.FeedClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetSubscriber(ctx, &emptypb.Empty{})
	if err != nil {
		log.Printf("could not get subscriber: %v", status.Convert(err).Message())
		return
	}
	log.Printf("Subscribers: %v", r.Subscribers)
}

func subscribeKline(c feed.FeedClient) {
	stream, err := c.SubscribeKline(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Printf("could not subscribe to kline: %v", status.Convert(err).Message())
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var dataCount int
	var mu sync.Mutex
	done := make(chan bool)
	defer close(done)

	go func() {
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				log.Printf("Received %d Klines in the last second", dataCount)
				dataCount = 0 // reset the count
				mu.Unlock()
			case <-done:
				return
			}
		}
	}()

	for {
		kline, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Printf("Stream closed by server")
				return
			} else {
				log.Printf("Error receiving from kline stream: %v", status.Convert(err).Message())
			}
		}
		mu.Lock()
		if dataCount == 0 {
			log.Printf("Received %d Kline: %v", dataCount, kline.Kline)
		}
		dataCount++
		mu.Unlock()
	}
}

func readHistoricalKline(c feed.FeedClient, start, end int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := c.ReadHistoricalKline(ctx, &feed.ReadKlineRequest{Start: start, End: end})
	if err != nil {
		log.Printf("could not read historical kline: %v", status.Convert(err).Message())
		return
	}
	for {
		kline, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving from historical kline stream: %v", status.Convert(err).Message())
			break
		}
		log.Printf("Received Historical Kline: %v", kline.Kline)
	}
}

//
