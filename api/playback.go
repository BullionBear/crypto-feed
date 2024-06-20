package api

import (
	pb "github.com/BullionBear/crypto-feed/api/gen/feed"
)

type playbackServer struct {
	pb.UnimplementedFeedServer
}

func NewPlaybackServer() *playbackServer {
	return &playbackServer{}
}
