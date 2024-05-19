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
