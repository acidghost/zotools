// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Config struct {
	Key     string
	Zotero  string
	Storage string
}

const (
	configEmptyMsg = "is empty in config"
)

var (
	ErrConfigEmptyKey     = errors.New("key " + configEmptyMsg)
	ErrConfigEmptyZotero  = errors.New("zotero " + configEmptyMsg)
	ErrConfigEmptyStorage = errors.New("storage " + configEmptyMsg)
)

type ErrConfig struct {
	errors []error
}

func (e *ErrConfig) Error() string {
	ss := make([]string, 0, len(e.errors))
	for _, err := range e.errors {
		ss = append(ss, fmt.Sprintf("- %v", err))
	}
	return strings.Join(ss, "\n")
}

func LoadConfig(path string) Config {
	file, err := os.Open(path)
	if err != nil {
		Die("Failed to open config file %q: %v\n", path, err)
	}
	config, err := loadConfigReader(file)
	if err != nil {
		var ec *ErrConfig
		if errors.As(err, &ec) {
			Die("Wrong config values in %q:\n%v\n", path, err)
		}
		Die("Failed to load config from %q: %v\n", path, err)
	}
	return config
}

func loadConfigReader(r io.Reader) (config Config, err error) {
	configBytes, err := io.ReadAll(r)
	if err != nil {
		return
	}
	if err = json.Unmarshal(configBytes, &config); err != nil {
		return
	}
	var ec *ErrConfig = &ErrConfig{make([]error, 0)}
	if config.Key == "" {
		ec.errors = append(ec.errors, ErrConfigEmptyKey)
	}
	if config.Zotero == "" {
		ec.errors = append(ec.errors, ErrConfigEmptyZotero)
	}
	if config.Storage == "" {
		ec.errors = append(ec.errors, ErrConfigEmptyStorage)
	}
	if len(ec.errors) > 0 {
		err = ec
	}
	return
}
