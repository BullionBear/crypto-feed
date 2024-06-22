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
		sleepMs:   10,
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
	pageSize := 1000
	offset := 0

	for {
		klines, err := s.db.QueryKlines(s.startTime, s.endTime, offset, pageSize)
		if err != nil {
			return err
		}

		if len(klines) == 0 {
			break
		}

		for _, kline := range klines {
			// sleep
			time.Sleep(time.Duration(s.sleepMs) * time.Millisecond)
			if err := stream.Send(&pb.KlineResponse{
				Kline: playbackToPbKline(&kline),
			}); err != nil {
				return err
			}
		}

		// Increment the offset for the next batch of records
		offset += pageSize
	}

	return nil
}

func (s *playbackServer) ReadHistoricalKline(request *pb.ReadKlineRequest, stream pb.Feed_ReadHistoricalKlineServer) error {
	start := int64(request.Start) / 1000 * 1000
	end := int64(request.End) / 1000 * 1000
	pageSize := 1000
	offset := 0
	for {
		klines, err := s.db.QueryKlines(start, end, offset, pageSize)
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

		// Increment the offset for the next batch of records
		offset += pageSize
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
