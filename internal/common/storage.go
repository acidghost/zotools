// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package common

import (
	"encoding/json"
	"fmt"
	"io/fs"
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
	Key         string
	Filename    string
	ContentType string
}

func NewStorage(filename string) Storage {
	var data StoredData
	data.Lib = Library{0, []Item{}}
	return Storage{filename, data}
}

func (s *Storage) Load() error {
	storeBytes, err := fs.ReadFile(defaultFS, s.filename)
	if err != nil {
		return fmt.Errorf("could not read storage file %s: %w", s.filename, err)
	}
	if err := json.Unmarshal(storeBytes, &s.Data); err != nil {
		return fmt.Errorf("failed to read JSON from %s: %w", s.filename, err)
	}
	return nil
}

func (s *Storage) Persist() error {
	serialized, err := json.Marshal(s.Data)
	if err != nil {
		return fmt.Errorf("failed to serialize library as JSON: %w", err)
	}
	err = os.WriteFile(s.filename, serialized, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", s.filename, err)
	}
	return nil
}

func (s *Storage) Drop() (err error) {
	err = os.Remove(s.filename)
	if err != nil {
		err = fmt.Errorf("failed to delete %s: %w", s.filename, err)
	}
	return
}
