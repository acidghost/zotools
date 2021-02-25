// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package common

import (
	"bytes"
	"errors"
	"testing"

	"github.com/acidghost/zotools/internal/testutils"
)

func TestLoadConfig(t *testing.T) {
	jsonRaw := `{"key": "somekey", "storage": "storage.json", "zotero": "zotero"}`
	r := bytes.NewReader([]byte(jsonRaw))
	c, err := loadConfigReader(r)
	if err != nil {
		t.Fail()
	}
	testutils.AssertEq(t, c.Key, "somekey")
	testutils.AssertEq(t, c.Storage, "storage.json")
	testutils.AssertEq(t, c.Zotero, "zotero")
}

func TestLoadConfigNotJSON(t *testing.T) {
	jsonRaw := `{error}`
	r := bytes.NewReader([]byte(jsonRaw))
	_, err := loadConfigReader(r)
	if err == nil {
		t.Fail()
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	jsonRaw := `{}`
	r := bytes.NewReader([]byte(jsonRaw))
	_, err := loadConfigReader(r)
	if err == nil {
		t.Fail()
	}
	var ec *ErrConfig
	if errors.As(err, &ec) {
		found := map[string]bool{"key": false, "zotero": false, "storage": false}
		for _, err := range ec.errors {
			if errors.Is(err, ErrConfigEmptyKey) {
				found["key"] = true
			} else if errors.Is(err, ErrConfigEmptyStorage) {
				found["storage"] = true
			} else if errors.Is(err, ErrConfigEmptyZotero) {
				found["zotero"] = true
			} else {
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
}
