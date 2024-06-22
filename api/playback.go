package api

/*
PlaybackServer is used to backtest trading strategies with same interface as FeedServer.
*/

import (
	pb "github.com/BullionBear/crypto-feed/api/gen/feed"
)

type playbackServer struct {
	pb.UnimplementedFeedServer
}

func NewPlaybackServer() *playbackServer {
	return &playbackServer{}
}
