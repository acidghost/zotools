// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package sync

import (
	"testing"

	"github.com/acidghost/zotools/internal/storage"
	"github.com/acidghost/zotools/internal/zotero"
	"github.com/stretchr/testify/assert"
)

func TestInitSync(t *testing.T) {
	itemsRes := zotero.ItemsResult{
		Version: 1337,
		Items: []zotero.Item{
			{
				Key:     "item1",
				Version: 1,
				Data: zotero.ItemData{
					Title:    "title item1",
					Abstract: "abstract item1",
				},
			},
			{
				Key:     "item2",
				Version: 2,
				Data: zotero.ItemData{
					Title:     "item2.pdf",
					ParentKey: "item1",
				},
			},
		},
	}
	s := storage.New("")
	initSync(&s, itemsRes)
	assert.Equal(t, s.Data.Lib.Version, uint(1337))
	assert.Equal(t, s.Data.Lib.Items[0].Key, "item1")
	assert.Equal(t, s.Data.Lib.Items[0].Attachments[0].Key, "item2")
}

func TestInitSyncInv(t *testing.T) {
	itemsRes := zotero.ItemsResult{
		Version: 1337,
		Items: []zotero.Item{
			{
				Key:     "item2",
				Version: 2,
				Data: zotero.ItemData{
					Title:     "item2.pdf",
					ParentKey: "item1",
				},
			},
			{
				Key:     "item1",
				Version: 1,
				Data: zotero.ItemData{
					Title:    "title item1",
					Abstract: "abstract item1",
				},
			},
		},
	}
	s := storage.New("")
	initSync(&s, itemsRes)
	assert.Equal(t, s.Data.Lib.Version, uint(1337))
	assert.Equal(t, s.Data.Lib.Items[0].Key, "item1")
	assert.Equal(t, s.Data.Lib.Items[0].Attachments[0].Key, "item2")
}

func TestInitSyncMultiAttach(t *testing.T) {
	itemsRes := zotero.ItemsResult{
		Version: 1337,
		Items: []zotero.Item{
			{
				Key:     "item2",
				Version: 2,
				Data: zotero.ItemData{
					Title:     "item2.pdf",
					ParentKey: "item1",
				},
			},
			{
				Key:     "item1",
				Version: 1,
				Data: zotero.ItemData{
					Title:    "title item1",
					Abstract: "abstract item1",
				},
			},
			{
				Key:     "item3",
				Version: 2,
				Data: zotero.ItemData{
					Title:     "item2.pdf",
					ParentKey: "item1",
				},
			},
		},
	}
	s := storage.New("")
	initSync(&s, itemsRes)
	assert.Equal(t, s.Data.Lib.Items[0].Key, "item1")
	assert.Equal(t, s.Data.Lib.Items[0].Attachments[0].Key, "item2")
	assert.Equal(t, s.Data.Lib.Items[0].Attachments[1].Key, "item3")
}
