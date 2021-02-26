// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package common

import (
	"bytes"
	"errors"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testValidConfigJSON = `{"key": "somekey", "storage": "storage.json", "zotero": "zotero"}`
)

func TestLoadConfig(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		r := bytes.NewReader([]byte(testValidConfigJSON))
		c, err := loadConfigReader(r)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assert.Equal(t, c.Key, "somekey")
		assert.Equal(t, c.Storage, "storage.json")
		assert.Equal(t, c.Zotero, "zotero")
	})
	t.Run("Read error", func(t *testing.T) {
		expErr := errors.New("some reader error")
		r := iotest.ErrReader(expErr)
		_, err := loadConfigReader(r)
		require.Error(t, err)
		assert.ErrorIs(t, err, expErr)
	})
	t.Run("Invalid JSON", func(t *testing.T) {
		jsonRaw := `{error}`
		r := bytes.NewReader([]byte(jsonRaw))
		_, err := loadConfigReader(r)
		assert.Error(t, err)
	})
	t.Run("Invalid config", func(t *testing.T) {
		jsonRaw := `{}`
		r := bytes.NewReader([]byte(jsonRaw))
		_, err := loadConfigReader(r)
		require.Error(t, err)
		var ec *ErrConfig
		require.ErrorAs(t, err, &ec)
		found := map[string]bool{"key": false, "zotero": false, "storage": false}
		for _, err := range ec.errors {
			switch {
			case errors.Is(err, ErrConfigEmptyKey):
				found["key"] = true
			case errors.Is(err, ErrConfigEmptyStorage):
				found["storage"] = true
			case errors.Is(err, ErrConfigEmptyZotero):
				found["zotero"] = true
			default:
				t.Fatalf("Unknown type of error: %v", err)
			}
		}
		for k, v := range found {
			if !v {
				t.Fatalf("Error %q not found", k)
			}
		}
	})
}
