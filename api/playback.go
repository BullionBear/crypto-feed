package api

/*
PlaybackServer is used to backtest trading strategies with same interface as FeedServer.
*/

import (
	pb "github.com/BullionBear/crypto-feed/api/gen/feed"
	"github.com/BullionBear/crypto-feed/domain/pgdb"
)

type playbackServer struct {
	pb.UnimplementedFeedServer
	db *pgdb.PgDatabase
}

func NewPlaybackServer(db *pgdb.PgDatabase) *playbackServer {
	return &playbackServer{
		db: db,
	}
}
