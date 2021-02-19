package main

import (
	"flag"
	"fmt"
	"os"

	. "github.com/acidghost/zotools/cache"
	"github.com/acidghost/zotools/zotero"
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
	db, err := LoadDB(config.Cache)
	if err != nil {
		Dief("Failed to load the local database:\n - %v\n", err)
	}

	zot, err := zotero.New(config.Key)
	if err != nil {
		Dief("Failed to initialize Zotero API:\n - %v\n", err)
	}

	if db.Lib.Version == 0 {
		// Initial sync queries all the items

		dropDB := func() {
			err := db.Drop()
			if err != nil {
				Eprintf("Database was not successfully deleted:\n - %v\n", err)
			}
			os.Exit(1)
		}

		items, err := zot.AllItems()
		if err != nil {
			Eprintf("Failed to load items:\n - %v\n", err)
			dropDB()
		}

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

		fmt.Printf("Retrived %d top level items\n", len(byKey))
		db.Lib.Version = items.Version
		db.Lib.Items = make([]StoredItem, 0, len(byKey))
		for _, item := range byKey {
			db.Lib.Items = append(db.Lib.Items, item)
		}

		if err := db.PersistLibrary(); err != nil {
			Eprintf("Failed to persist library:\n - %v\n", err)
			dropDB()
		}

		println("Library persisted!")
	} else {
		// TODO: code me
		println("Synchronizing an existing library is not supported yet")
	}
}
