package config

import (
	"encoding/json"
	"os"
)

func Load(path string) (*Config, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &Config{}

	err = json.NewDecoder(file).Decode(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
