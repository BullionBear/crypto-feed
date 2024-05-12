package main

import (
	"net"

	"github.com/BullionBear/crypto-feed/api"
	pb "github.com/BullionBear/crypto-feed/api/gen/feed"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	feedServer := api.NewFeedServer()
	pb.RegisterFeedServer(s, feedServer)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
