package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/BullionBear/crypto-feed/api"
	pb "github.com/BullionBear/crypto-feed/api/gen/feed"
	"github.com/BullionBear/crypto-feed/domain/config"
	"github.com/BullionBear/crypto-feed/pkg/service"
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

	// Enable the logging of the source (file and line number)
	log.SetReportCaller(true)

	// Set the log level
	log.SetLevel(log.InfoLevel)
}

func main() {
	configPath := flag.String("config", "path/to/config.json", "path to config file")
	flag.Parse()

	// Read and parse the configuration file
	config, err := config.ReadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	lis, err := net.Listen("tcp", ":"+fmt.Sprintf("%d", config.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	klineSrv := service.NewKLineService(config.Symbol, int64(config.Length))
	go klineSrv.Run()
	feedServer := api.NewFeedServer(klineSrv)

	pb.RegisterFeedServer(s, feedServer)
	log.Infof("server listening at %s", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
