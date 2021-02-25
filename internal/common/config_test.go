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
	r := bytes.NewReader([]byte(testValidConfigJSON))
	c, err := loadConfigReader(r)
	if err != nil {
		t.Fail()
	}
	testutils.AssertEq(t, c.Key, "somekey")
	testutils.AssertEq(t, c.Storage, "storage.json")
	testutils.AssertEq(t, c.Zotero, "zotero")
}

func TestLoadConfigReadErr(t *testing.T) {
	r := iotest.ErrReader(errors.New("some reader error"))
	_, err := loadConfigReader(r)
	if err == nil {
		t.Fail()
	}
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
}
