package api

import (
	"context"
	"strconv"
	"time"

	pb "github.com/BullionBear/crypto-feed/api/gen/feed"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

// server is used to implement feed.FeedServer.
type feedServer struct {
	pb.UnimplementedFeedServer
}

func NewFeedServer() *feedServer {
	return &feedServer{}
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
func (s *feedServer) StreamData(req *pb.StreamRequest, stream pb.Feed_StreamDataServer) error {
	// Example of streaming data
	log.WithFields(log.Fields{
		"streamID": req.StreamId,
	}).Info("StreamData is subscribed")
	defer log.Info("Leave StreamData")
	for i := 0; i < 10; i++ {
		err := stream.Send(&pb.DataResponse{
			Data:      "Stream Message " + strconv.Itoa(i),
			Timestamp: time.Now().Unix(),
		})
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

// GetHistory implements feed.FeedServer
func (s *feedServer) GetHistory(ctx context.Context, in *pb.HistoryRequest) (*pb.HistoryResponse, error) {
	// Example data, normally you would query this from your storage based on in.FromDate and in.ToDate
	dataPoints := make([]*pb.DataPoint, 0)
	for i := 0; i < 5; i++ {
		dataPoints = append(dataPoints, &pb.DataPoint{
			Data:      "Data point " + strconv.Itoa(i),
			Timestamp: time.Now().Unix() - int64(i)*1000,
		})
	}
	return &pb.HistoryResponse{
		DataPoints: dataPoints,
	}, nil
}
