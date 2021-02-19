package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/acidghost/zotools/zotero"
)

type DB struct {
	filename string
	file     *os.File
	Lib      *Library
}

type Library struct {
	Version uint         `json:"version"`
	Items   []StoredItem `json:"items"`
}

type StoredItem struct {
	Key         string           `json:"key"`
	Version     uint             `json:"version"`
	Title       string           `json:"title"`
	Abstract    string           `json:"abstractNote"`
	ItemType    string           `json:"itemType"`
	Creators    []zotero.Creator `json:"creators"`
	Attachments []Attachment     `json:"attachments"`
}

type Attachment struct {
	Key         string
	Version     uint
	ContentType string
	Filename    string
}

func LoadDB(filename string) (*DB, error) {
	newDB := false
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		newDB = true
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		verb := "open"
		if newDB {
			verb = "create"
		}
		return nil, fmt.Errorf("could not %s file %s: %v", verb, filename, err)
	}

	var library Library
	if !newDB {
		fileContents, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read from file %s: %v", filename, err)
		}

		if err := json.Unmarshal(fileContents, &library); err != nil {
			return nil, fmt.Errorf("failed to read JSON from %s: %v", filename, err)
		}
	} else {
		library = Library{}
	}

	return &DB{filename, file, &library}, nil
}

func (db *DB) PersistLibrary() error {
	serialized, err := json.Marshal(db.Lib)
	if err != nil {
		return fmt.Errorf("failed to serialize library as JSON: %v", err)
	}
	_, err = db.file.Write(serialized)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %v", db.filename, err)
	}
	return nil
}

func (db *DB) Drop() (err error) {
	err = os.Remove(db.filename)
	if err != nil {
		err = fmt.Errorf("failed to delete %s: %v", db.filename, err)
	}
	return
}
