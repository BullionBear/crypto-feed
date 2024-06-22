package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/BullionBear/crypto-feed/api"
	pb "github.com/BullionBear/crypto-feed/api/gen/feed"
	"github.com/BullionBear/crypto-feed/domain/config"
	"github.com/BullionBear/crypto-feed/domain/pgdb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func init() {
	// Set formatter to TextFormatter for human-readable logs
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat:           "2006-01-02 15:04:05", // Customize timestamp format
		FullTimestamp:             true,                  // Show full timestamp instead of elapsed time
		ForceColors:               true,                  // Force colors even if stdout is not a tty
		DisableColors:             false,                 // Set to true to disable colors
		DisableQuote:              true,                  // Disable quoting of values
		EnvironmentOverrideColors: true,                  // Override coloring based on environment settings
	})
}

func main() {
	configPath := flag.String("config", "path/to/config.json", "path to config file")
	flag.Parse()

	// Read and parse the configuration file
	config, err := config.ReadPlaybackConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	lis, err := net.Listen("tcp", ":"+fmt.Sprintf("%d", config.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	// New Resources
	dbConfig := config.Postgres
	db, err := pgdb.NewPgDatabase(dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode, dbConfig.Timezone)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	playbackServer := api.NewPlaybackServer(db, config.StartTime, config.EndTime)

	pb.RegisterFeedServer(s, playbackServer)
	log.Infof("server listening at %s", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
