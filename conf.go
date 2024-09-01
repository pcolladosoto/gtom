package main

import (
	"encoding/json"
	"os"
)

func loadConf(path string) (gtomConfig, error) {
	var parsedConf gtomConfig

	rawConf, err := os.ReadFile(path)
	if err != nil {
		return parsedConf, err
	}

	if err := json.Unmarshal(rawConf, &parsedConf); err != nil {
		return parsedConf, err
	}

	return parsedConf, nil
}
