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

type errWrapped struct {
	inner error
}

func (e *errWrapped) Unwrap() error {
	return e.inner
}

type errReadStorage struct {
	errWrapped
	filename string
}

func (e *errReadStorage) Error() string {
	return fmt.Sprintf("could not read storage file %q: %v", e.filename, e.inner)
}

type errNotJSON struct {
	errWrapped
	filename string
}

func (e *errNotJSON) Error() string {
	return fmt.Sprintf("failed to parse JSON from %q: %v", e.filename, e.inner)
}

type errSerialize struct {
	errWrapped
}

func (e *errSerialize) Error() string {
	return fmt.Sprintf("failed to serialize as JSON: %v", e.inner)
}

type errWrite struct {
	errWrapped
	filename string
}

func (e *errWrite) Error() string {
	return fmt.Sprintf("failed to write to %q: %v", e.filename, e.inner)
}

type errDrop struct {
	errWrapped
	filename string
}

func (e *errDrop) Error() string {
	return fmt.Sprintf("failed to delete %q: %v", e.filename, e.inner)
}

func NewStorage(filename string) Storage {
	var data StoredData
	data.Lib = Library{0, []Item{}}
	return Storage{filename, data}
}

func (s *Storage) Load() error {
	storeBytes, err := fs.ReadFile(defaultFS, s.filename)
	if err != nil {
		return &errReadStorage{errWrapped{err}, s.filename}
	}
	if err := json.Unmarshal(storeBytes, &s.Data); err != nil {
		return &errNotJSON{errWrapped{err}, s.filename}
	}
	return nil
}

func (s *Storage) Persist() error {
	serialized, err := json.Marshal(s.Data)
	if err != nil {
		return &errSerialize{errWrapped{err}}
	}
	err = os.WriteFile(s.filename, serialized, 0644)
	if err != nil {
		return &errWrite{errWrapped{err}, s.filename}
	}
	return nil
}

func (s *Storage) Drop() (err error) {
	err = os.Remove(s.filename)
	if err != nil {
		err = &errDrop{errWrapped{err}, s.filename}
	}
	return
}
