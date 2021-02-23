// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package common

import (
	"encoding/json"
	"os"
)

type Config struct {
	Key     string
	Zotero  string
	Storage string
}

func LoadConfig(path string) Config {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		Dief("Failed to read config file:\n - %v\n", err)
	}

	var config Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		Dief("Failed to parse config JSON from %s: %v\n", path, err)
	}

	if config.Storage == "" {
		Dief("Storage is empty in %s\n", path)
	}

	return config
}
