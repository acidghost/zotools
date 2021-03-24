// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package storage

import (
	"encoding/json"
	"io/fs"
	"os"

	"github.com/acidghost/zotools/internal/utils"
	"github.com/acidghost/zotools/internal/zotero"
)

var defaultFS fs.FS = &utils.DummyFS{}

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
	Items   ItemsMap
}

type ItemsMap map[string]Item

type Item struct {
	Version     uint
	Title       string
	Abstract    string
	ItemType    string
	Creators    []zotero.Creator
	Attachments AttachmentsMap
}

type AttachmentsMap map[string]Attachment

type Attachment struct {
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

type errSpec string

const (
	errReadStorageSpec = errSpec("wrap:could not read storage from file {{filename string %q}}")
	errNotJSONSpec     = errSpec("wrap:failed to parse JSON from {{filename string %q}}")
	errSerializeSpec   = errSpec("wrap:failed to serialize as JSON")
	errWriteSpec       = errSpec("wrap:failed to write to {{filename string %q}}")
	errDropSpec        = errSpec("wrap:failed to delete {{filename string %q}}")
)

//go:generate gorror -type=errSpec -suffix=Spec

func New(filename string) Storage {
	var data StoredData
	data.Lib = Library{0, make(ItemsMap)}
	return Storage{filename, data}
}

func (s *Storage) Load() error {
	storeBytes, err := fs.ReadFile(defaultFS, s.filename)
	if err != nil {
		return newErrReadStorage(s.filename, err)
	}
	if err := json.Unmarshal(storeBytes, &s.Data); err != nil {
		return newErrNotJSON(s.filename, err)
	}
	return nil
}

func (s *Storage) Persist() error {
	serialized, err := json.Marshal(s.Data)
	if err != nil {
		return newErrSerialize(err)
	}
	err = os.WriteFile(s.filename, serialized, 0644)
	if err != nil {
		return newErrWrite(s.filename, err)
	}
	return nil
}

func (s *Storage) Drop() (err error) {
	err = os.Remove(s.filename)
	if err != nil {
		err = newErrDrop(s.filename, err)
	}
	return
}
