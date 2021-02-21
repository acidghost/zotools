package main

import (
	"testing"

	"github.com/acidghost/zotools/internal/cache"
	"github.com/acidghost/zotools/internal/zotero"
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
	c := &cache.Cache{Lib: &cache.Library{}}
	initSync(c, itemsRes)
	assertEqNest(t, c, "cache.Lib.Version", uint(1337))
	assertEqNest(t, c, "cache.Lib.Items[0].Key", "item1")
	assertEqNest(t, c, "cache.Lib.Items[0].Attachments[0].Key", "item2")
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
	c := &cache.Cache{Lib: &cache.Library{}}
	initSync(c, itemsRes)
	assertEqNest(t, c, "cache.Lib.Version", uint(1337))
	assertEqNest(t, c, "cache.Lib.Items[0].Key", "item1")
	assertEqNest(t, c, "cache.Lib.Items[0].Attachments[0].Key", "item2")
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
	c := &cache.Cache{Lib: &cache.Library{}}
	initSync(c, itemsRes)
	assertEqNest(t, c, "cache.Lib.Items[0].Key", "item1")
	assertEqNest(t, c, "cache.Lib.Items[0].Attachments[0].Key", "item2")
	assertEqNest(t, c, "cache.Lib.Items[0].Attachments[1].Key", "item3")
}
