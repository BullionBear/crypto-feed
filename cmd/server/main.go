package main

import (
	"net"

	"github.com/BullionBear/crypto-feed/api"
	pb "github.com/BullionBear/crypto-feed/api/gen/feed"
	"github.com/BullionBear/crypto-feed/pkg/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	klineSrv := service.NewKLineService("BTCUSDT", 86400)
	klineSrv.Run()
	feedServer := api.NewFeedServer(klineSrv)

	pb.RegisterFeedServer(s, feedServer)
	log.Infof("server listening at %s", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
