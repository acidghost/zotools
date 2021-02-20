package main

import (
	"flag"
	"fmt"
	"os"

	. "github.com/acidghost/zotools/internal/cache"
	"github.com/acidghost/zotools/internal/zotero"
)

type SyncCommand struct {
	fs *flag.FlagSet
}

func NewSyncCommand() *SyncCommand {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	return &SyncCommand{fs}
}

func (c *SyncCommand) Name() string {
	return c.fs.Name()
}

func (c *SyncCommand) Run(args []string, config Config) {
	c.fs.Parse(args)
	cache, err := Load(config.Cache)
	if err != nil {
		Dief("Failed to load the local cache:\n - %v\n", err)
	}

	zot, err := zotero.New(config.Key)
	if err != nil {
		Dief("Failed to initialize Zotero API:\n - %v\n", err)
	}

	if cache.Lib.Version == 0 {
		// Initial sync queries all the items

		dropCache := func() {
			err := cache.Drop()
			if err != nil {
				Eprintf("Cache was not successfully deleted:\n - %v\n", err)
			}
			os.Exit(1)
		}

		items, err := zot.AllItems()
		if err != nil {
			Eprintf("Failed to load items:\n - %v\n", err)
			dropCache()
		}

		initSync(cache, items)
		fmt.Printf("Retrived %d top level items\n", len(cache.Lib.Items))

		if err := cache.PersistLibrary(); err != nil {
			Eprintf("Failed to persist library:\n - %v\n", err)
			dropCache()
		}

		println("Library persisted!")
	} else {
		// TODO: code me
		println("Synchronizing an existing library is not supported yet")
	}
}

func initSync(cache *Cache, items zotero.ItemsResult) {
	byKey := make(map[string]StoredItem)
	for _, item := range items.Items {
		if item.Data.ParentKey != "" {
			attach := Attachment{
				Key:         item.Key,
				Version:     item.Version,
				ContentType: item.Data.ContentType,
				Filename:    item.Data.Filename,
			}
			if parent, exists := byKey[item.Data.ParentKey]; exists {
				parent.Attachments = append(parent.Attachments, attach)
				byKey[item.Data.ParentKey] = parent
			} else {
				byKey[item.Data.ParentKey] = StoredItem{
					Attachments: []Attachment{attach},
				}
			}
		} else {
			if existing, exists := byKey[item.Key]; exists {
				// Already present, only attachments
				byKey[item.Key] = StoredItem{
					Key:         item.Key,
					Version:     item.Version,
					Title:       item.Data.Title,
					Abstract:    item.Data.Abstract,
					Creators:    item.Data.Creators,
					Attachments: existing.Attachments,
				}
			} else {
				byKey[item.Key] = StoredItem{
					Key:         item.Key,
					Version:     item.Version,
					Title:       item.Data.Title,
					Abstract:    item.Data.Abstract,
					Creators:    item.Data.Creators,
					Attachments: []Attachment{},
				}
			}
		}
	}

	cache.Lib.Version = items.Version
	cache.Lib.Items = make([]StoredItem, 0, len(byKey))
	for _, item := range byKey {
		cache.Lib.Items = append(cache.Lib.Items, item)
	}
}
