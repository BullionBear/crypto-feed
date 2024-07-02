package api

/*
PlaybackServer is used to backtest trading strategies with same interface as FeedServer.
*/

import (
	"context"
	"time"

	pb "github.com/BullionBear/crypto-feed/api/gen/feed"
	"github.com/BullionBear/crypto-feed/domain/pgdb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

type playbackServer struct {
	pb.UnimplementedFeedServer
	db        *pgdb.PgDatabase
	startTime int64
	endTime   int64
	sleepMs   int64
}

func NewPlaybackServer(db *pgdb.PgDatabase, startTime, endTime int64) *playbackServer {
	return &playbackServer{
		db:        db,
		startTime: startTime,
		endTime:   endTime,
		sleepMs:   0,
	}
}

func (s *playbackServer) GetConfig(ctx context.Context, in *emptypb.Empty) (*pb.ConfigResponse, error) {
	return &pb.ConfigResponse{
		Symbol: "",
		Length: 0,
	}, nil
}

func (s *playbackServer) GetStatus(ctx context.Context, in *emptypb.Empty) (*pb.StatusResponse, error) {
	return &pb.StatusResponse{
		Status: pb.Status_OK,
		Start:  s.startTime,
		End:    s.endTime,
		Size:   s.endTime - s.startTime + 1,
	}, nil
}

func (s *playbackServer) GetSubscriber(ctx context.Context, in *emptypb.Empty) (*pb.SubscriberResponse, error) {
	return &pb.SubscriberResponse{
		Subscribers: make([]int64, 0),
	}, nil
}

func (s *playbackServer) SubscribeKline(in *emptypb.Empty, stream pb.Feed_SubscribeKlineServer) error {
	log.Info("SubscribeKline get called")
	defer log.Info("Leave SubscribeKline")
	interval := int64(3_600_000) // 1 hour interval (3600 seconds)
	currentTime := s.startTime
	for {
		endTime := currentTime + interval - 1
		if endTime >= s.endTime {
			endTime = s.endTime
		}
		//	Query klines from the database
		log.Infof("Querying klines from %d to %d", currentTime, endTime)
		klines, err := s.db.QueryKlines(currentTime, endTime)
		if err != nil {
			return err
		}

		if len(klines) == 0 {
			break
		}

		for _, kline := range klines {
			// sleep
			if s.sleepMs > 0 {
				time.Sleep(time.Duration(s.sleepMs) * time.Millisecond)
			}
			if err := stream.Send(&pb.KlineResponse{
				Kline: playbackToPbKline(&kline),
			}); err != nil {
				return err
			}
		}

		// Increment the current time for the next batch of records
		currentTime = endTime + 1

		// If we've reached the end of the time range, break the loop
		if currentTime >= s.endTime {
			break
		}
	}

	return nil
}

func (s *playbackServer) ReadHistoricalKline(request *pb.ReadKlineRequest, stream pb.Feed_ReadHistoricalKlineServer) error {
	start := int64(request.Start) / 1000 * 1000
	end := int64(request.End) / 1000 * 1000

	interval := int64(1_000_000) // 1,000,000 ms interval (1000 seconds)
	currentTime := start
	for {
		endTime := currentTime + interval
		if endTime > end {
			endTime = end
		}

		klines, err := s.db.QueryKlines(currentTime, endTime)
		if err != nil {
			return err
		}

		if len(klines) == 0 {
			break
		}

		for _, kline := range klines {
			if err := stream.Send(&pb.KlineResponse{
				Kline: playbackToPbKline(&kline),
			}); err != nil {
				return err
			}
		}

		// Increment the current time for the next batch of records
		currentTime = endTime + 1

		// If we've reached the end of the time range, break the loop
		if currentTime > end {
			break
		}
	}
	return nil
}

func playbackToPbKline(playbackKline *pgdb.PlaybackKline) *pb.Kline {
	return &pb.Kline{
		OpenTime:                 playbackKline.OpenTime,
		Open:                     playbackKline.Open,
		High:                     playbackKline.High,
		Low:                      playbackKline.Low,
		Close:                    playbackKline.Close,
		Volume:                   playbackKline.Volume,
		CloseTime:                playbackKline.CloseTime,
		QuoteAssetVolume:         playbackKline.QuoteVolume,
		TradeNum:                 playbackKline.Count,
		TakerBuyBaseAssetVolume:  playbackKline.TakerBuyVolume,
		TakerBuyQuoteAssetVolume: playbackKline.TakerBuyQuoteVolume,
	}
}
