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

func (s *feedServer) GetConfig(ctx context.Context, in *emptypb.Empty) (*pb.ConfigResponse, error) {
	return &pb.ConfigResponse{
		Symbol: s.klineSrv.Symbol(),
		Length: s.klineSrv.Length(),
	}, nil
}

// GetStatus implements feed.FeedServer
func (s *feedServer) GetStatus(ctx context.Context, in *emptypb.Empty) (*pb.StatusResponse, error) {
	// Example response, normally you would query this data from your application logic
	log.Info("GetStatus get called")
	defer log.Info("Leave GetStatus")
	startKline, err := s.klineSrv.Head()
	if err != nil {
		return &pb.StatusResponse{
			Status:    pb.Status_ERROR,
			Start:     0,
			End:       0,
			Timestamp: time.Now().UnixMilli(),
			Size:      0,
		}, err
	}
	endKline, err := s.klineSrv.Tail()
	if err != nil {
		return &pb.StatusResponse{
			Status:    pb.Status_ERROR,
			Start:     startKline.OpenTime,
			End:       0,
			Timestamp: time.Now().UnixMilli(),
			Size:      0,
		}, err
	}
	size := s.klineSrv.Size()
	status := s.klineSrv.Status()

	return &pb.StatusResponse{
		Status:    convertToStatus(status),
		Start:     startKline.OpenTime,
		End:       endKline.CloseTime,
		Timestamp: time.Now().UnixMilli(),
		Size:      size,
	}, nil
}

func (s *feedServer) GetSubscriber(context.Context, *emptypb.Empty) (*pb.SubscriberResponse, error) {
	return &pb.SubscriberResponse{
		Subscribers: s.klineSrv.ListSubsriber(),
	}, nil
}

// StreamData implements feed.FeedServer
func (s *feedServer) StreamKline(in *emptypb.Empty, stream pb.Feed_StreamKlineServer) error {
	log.Info("StreamKline get called")
	defer log.Info("Leave StreamKline")
	klineCh := make(chan *pb.Kline)

	kline_handler := func(srvKline *service.Kline) {
		pbKline := convertToPbKline(srvKline)
		klineCh <- pbKline
	}
	defer log.Info("Leave StreamData")
	id := s.klineSrv.Subscribe(kline_handler)
	defer s.klineSrv.Unsubscribe(id)
	for kline := range klineCh {
		response := pb.KlineResponse{
			Kline: kline,
		}
		if err := stream.Send(&response); err != nil {
			log.Warnf("Error sending data to client: %s", err.Error())
			break // Exit the loop if we fail to send data
		}
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

func convertToStatus(status service.Status) pb.Status {
	switch status {
	case service.StatusCreated:
		return pb.Status_CREATED
	case service.StatusInitializing:
		return pb.Status_INITIALIZING
	case service.StatusRunning:
		return pb.Status_OK
	case service.StatusError:
		return pb.Status_ERROR
	default:
		return pb.Status_UNKNOWN
	}
}
