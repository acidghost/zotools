// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package storage

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageLoad(t *testing.T) {
	oldFS := defaultFS
	t.Cleanup(func() { defaultFS = oldFS })
	t.Run("Read error", func(t *testing.T) {
		defaultFS = fstest.MapFS(map[string]*fstest.MapFile{})
		s := New("filename.json")
		err := s.Load()
		var e *errReadStorage
		assert.ErrorAs(t, err, &e)
	})
	t.Run("Invalid JSON", func(t *testing.T) {
		f := "filename.json"
		defaultFS = fstest.MapFS(map[string]*fstest.MapFile{
			f: {Data: []byte(`{"key": "value"`)},
		})
		s := New(f)
		err := s.Load()
		var e *errNotJSON
		assert.ErrorAs(t, err, &e)
	})
}

func TestStoragePersist(t *testing.T) {
	t.Run("Persist file", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "filename.json")
		s := New(f)
		s.Data.Lib.Version = 0
		s.Data.Lib.Items = []Item{}
		s.Data.Search = nil
		err := s.Persist()
		require.NoError(t, err)
		bs, err := os.ReadFile(f)
		assert.NoError(t, err)
		exp := `{"Lib":{"Version":0,"Items":[]},"Search":null}`
		assert.Equal(t, string(bs), exp)
	})
	t.Run("Not existent folder", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "somefolder", "filename.json")
		s := New(f)
		err := s.Persist()
		var e *errWrite
		assert.ErrorAs(t, err, &e)
		var pe *fs.PathError
		assert.ErrorAs(t, err, &pe)
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
		s := New(f)
		err = s.Drop()
		require.NoError(t, err)
		assert.NoFileExists(t, f)
	})
	t.Run("Non existent", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "filename.json")
		s := New(f)
		err := s.Drop()
		var e *errDrop
		assert.ErrorAs(t, err, &e)
		var pe *fs.PathError
		assert.ErrorAs(t, err, &pe)
	})
}
