package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port   int    `json:"port"` // Port as an integer
	Symbol string `json:"symbol"`
	Length int    `json:"length"`
}

func ReadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

type PlaybackConfig struct {
	Port     int            `json:"port"`
	Symbol   string         `json:"symbol"`
	Postgres PostgresConfig `json:"postgres"`
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	SSLMode  string `json:"ssl_mode"`
	Timezone string `json:"timezone"`
}

func ReadPlaybackConfig(path string) (*PlaybackConfig, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config PlaybackConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
