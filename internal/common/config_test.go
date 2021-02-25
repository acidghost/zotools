// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package common

import (
	"bytes"
	"errors"
	"testing"
	"testing/iotest"

	"github.com/acidghost/zotools/internal/testutils"
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
		testutils.AssertEq(t, c.Key, "somekey")
		testutils.AssertEq(t, c.Storage, "storage.json")
		testutils.AssertEq(t, c.Zotero, "zotero")
	})
	t.Run("Read error", func(t *testing.T) {
		expErr := errors.New("some reader error")
		r := iotest.ErrReader(expErr)
		_, err := loadConfigReader(r)
		if err == nil {
			t.Fatalf("Expected error")
		} else if !errors.Is(err, expErr) {
			t.Fatalf("Got wrong error")
		}
	})
	t.Run("Invalid JSON", func(t *testing.T) {
		jsonRaw := `{error}`
		r := bytes.NewReader([]byte(jsonRaw))
		_, err := loadConfigReader(r)
		if err == nil {
			t.Fatalf("Expected error")
		}
	})
	t.Run("Invalid config", func(t *testing.T) {
		jsonRaw := `{}`
		r := bytes.NewReader([]byte(jsonRaw))
		_, err := loadConfigReader(r)
		if err == nil {
			t.Fatalf("Expected error")
		}
		var ec *ErrConfig
		if errors.As(err, &ec) {
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
					t.Fail()
				}
			}
			for k, v := range found {
				if !v {
					t.Fatalf("Error %q not found", k)
				}
			}
		} else {
			t.Fail()
		}
	})
}
