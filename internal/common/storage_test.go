// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package common

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestStorageLoad(t *testing.T) {
	oldFS := defaultFS
	t.Cleanup(func() { defaultFS = oldFS })
	t.Run("Read error", func(t *testing.T) {
		defaultFS = fstest.MapFS(map[string]*fstest.MapFile{})
		s := NewStorage("filename.json")
		err := s.Load()
		if err == nil {
			t.Fatalf("Expected an error")
		} else if !strings.Contains(err.Error(), "could not read storage") {
			t.Fatalf("Did not return the expected error: %#v", err)
		}
	})
	t.Run("Invalid JSON", func(t *testing.T) {
		f := "filename.json"
		defaultFS = fstest.MapFS(map[string]*fstest.MapFile{
			f: {Data: []byte(`{"key": "value"`)},
		})
		s := NewStorage(f)
		err := s.Load()
		if err == nil {
			t.Fatalf("Expected an error")
		} else if !strings.Contains(err.Error(), "failed to read JSON") {
			t.Fatalf("Did not return the expected error")
		}
	})
}

func TestStoragePersist(t *testing.T) {
	t.Run("Persist file", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "filename.json")
		s := NewStorage(f)
		s.Data.Lib.Version = 0
		s.Data.Lib.Items = []Item{}
		s.Data.Search = nil
		err := s.Persist()
		if err != nil {
			t.Fatalf("Unexpected error: %#v", err)
		}
		bs, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("Failed to read file back: %v", err)
		}
		exp := `{"Lib":{"Version":0,"Items":[]},"Search":null}`
		if string(bs) != exp {
			t.Fatalf("Persisted wrong content: %v", bs)
		}
	})
}

func TestStorageDrop(t *testing.T) {
	t.Run("Actually drop", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "filename.json")
		file, err := os.Create(f)
		if err != nil {
			t.Fatalf("Cannot create file: %v", err)
		}
		file.Close()
		s := NewStorage(f)
		err = s.Drop()
		if err != nil {
			t.Fatalf("Unexpected error: %#v", err)
		}
		_, err = os.Stat(f)
		var pe *fs.PathError
		if err == nil {
			t.Fatalf("File was not deleted")
		} else if !errors.As(err, &pe) {
			t.Fatalf("Unexpected stat error: %#v", err)
		}
	})
	t.Run("Non existent", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "filename.json")
		s := NewStorage(f)
		err := s.Drop()
		if err == nil {
			t.Fatalf("Expected error")
		}
		var pe *fs.PathError
		if !errors.As(err, &pe) {
			t.Fatalf("Unexpected error: %#v", err)
		}
	})
}
