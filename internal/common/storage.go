// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU GPL License version 3.

package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/acidghost/zotools/internal/zotero"
)

type Storage struct {
	filename string
	Data     StoredData
}

type StoredData struct {
	Lib    Library
	Search *SearchResults
}

type Library struct {
	Version uint
	Items   []Item
}

type Item struct {
	Key         string
	Version     uint
	Title       string
	Abstract    string
	ItemType    string
	Creators    []zotero.Creator
	Attachments []Attachment
}

type Attachment struct {
	Key         string
	Version     uint
	ContentType string
	Filename    string
}

type SearchResults struct {
	Term  string
	Items []SearchResultsItem
}

type SearchResultsItem struct {
	Key      string
	Filename string
}

func NewStorage(filename string) Storage {
	var data StoredData
	data.Lib = Library{0, []Item{}}
	return Storage{filename, data}
}

func (s *Storage) Load() error {
	storeBytes, err := os.ReadFile(s.filename)
	if err != nil {
		return fmt.Errorf("could not read storage file %s: %v\n", s.filename, err)
	}

	if err := json.Unmarshal(storeBytes, &s.Data); err != nil {
		return fmt.Errorf("failed to read JSON from %s: %v", s.filename, err)
	}

	return nil
}

func (s *Storage) Persist() error {
	serialized, err := json.Marshal(s.Data)
	if err != nil {
		return fmt.Errorf("failed to serialize library as JSON: %v", err)
	}
	file, err := os.OpenFile(s.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("could not open storage file %s: %v\n", s.filename, err)
	}
	defer file.Close()
	_, err = file.Write(serialized)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %v", s.filename, err)
	}
	return nil
}

func (s *Storage) Drop() (err error) {
	err = os.Remove(s.filename)
	if err != nil {
		err = fmt.Errorf("failed to delete %s: %v", s.filename, err)
	}
	return
}
