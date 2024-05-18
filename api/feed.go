package api

import (
	"context"
	"time"

	pb "github.com/BullionBear/crypto-feed/api/gen/feed"
	"github.com/BullionBear/crypto-feed/pkg/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

// server is used to implement feed.FeedServer.
type feedServer struct {
	klineSrv *service.KLineService
	pb.UnimplementedFeedServer
}

func NewFeedServer(klineSrv *service.KLineService) *feedServer {
	return &feedServer{
		klineSrv: klineSrv,
	}
}

// GetStatus implements feed.FeedServer
func (s *feedServer) GetStatus(ctx context.Context, in *emptypb.Empty) (*pb.StatusResponse, error) {
	// Example response, normally you would query this data from your application logic
	log.Info("GetStatus get called")
	defer log.Info("Leave GetStatus")
	return &pb.StatusResponse{
		Status:    "Service is running",
		Timestamp: time.Now().Unix(),
	}, nil
}

// StreamData implements feed.FeedServer
func (s *feedServer) StreamKline(req *pb.StreamRequest, stream pb.Feed_StreamKlineServer) error {
	log.Info("StreamKline get called")
	defer log.Info("Leave StreamKline")
	klineCh := make(chan *pb.Kline)
	log.WithFields(log.Fields{
		"streamID": req.StreamId,
	}).Info("StreamData is subscribed")

	kline_handler := func(srvKline *service.Kline) {
		pbKline := convertToPbKline(srvKline)
		klineCh <- pbKline
	}
	defer log.Info("Leave StreamData")
	s.klineSrv.Subscribe(kline_handler)
	for kline := range klineCh {
		response := pb.DataResponse{
			Kline: kline,
		}
		stream.Send(&response)
	}
	return nil
}

func convertToPbKline(srvKline *service.Kline) *pb.Kline {
	return &pb.Kline{
		OpenTime:                 srvKline.OpenTime,
		Open:                     srvKline.Open,
		High:                     srvKline.High,
		Low:                      srvKline.Low,
		Close:                    srvKline.Close,
		Volume:                   srvKline.Volume,
		CloseTime:                srvKline.CloseTime,
		QuoteAssetVolume:         srvKline.QuoteAssetVolume,
		TradeNum:                 srvKline.TradeNum,
		TakerBuyBaseAssetVolume:  srvKline.TakerBuyBaseAssetVolume,
		TakerBuyQuoteAssetVolume: srvKline.TakerBuyQuoteAssetVolume,
	}
}
