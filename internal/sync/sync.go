// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package sync

import (
	"flag"
	"fmt"
	"os"

	"github.com/acidghost/zotools/internal/config"
	"github.com/acidghost/zotools/internal/storage"
	"github.com/acidghost/zotools/internal/utils"
	"github.com/acidghost/zotools/internal/zotero"
)

const syncUsageTop = " " + utils.OptionsUsage

type Command struct {
	fs       *flag.FlagSet
	flagDrop *bool
}

func New(cmd, banner string) *Command {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	flagDrop := fs.Bool("drop", false, "delete storage and start fresh")
	fs.Usage = utils.MakeUsage(fs, cmd, banner, syncUsageTop, "")
	return &Command{fs, flagDrop}
}

func (c *Command) Run(args []string, conf config.Config) {
	//nolint:errcheck
	c.fs.Parse(args)

	exists := true
	if _, err := os.Stat(conf.Storage); err != nil && os.IsNotExist(err) {
		exists = false
	}

	store := storage.New(conf.Storage)
	if *c.flagDrop && exists {
		if err := store.Drop(); err != nil {
			utils.Die("Failed to drop storage:\n - %v\n", err)
		}
	} else if exists {
		if err := store.Load(); err != nil {
			utils.Die("Failed to load the local storage:\n - %v\n", err)
		}
	}

	zot, err := zotero.New(conf.Key)
	if err != nil {
		utils.Die("Failed to initialize Zotero API:\n - %v\n", err)
	}

	if store.Data.Lib.Version == 0 {
		// Initial sync queries all the items
		items, err := zot.AllItems()
		if err != nil {
			utils.Die("Failed to load items:\n - %v\n", err)
		}

		initSync(&store, items)
		fmt.Printf("Retrieved %d top level items\n", len(store.Data.Lib.Items))

		if err := store.Persist(); err != nil {
			utils.Die("Failed to persist library:\n - %v\n", err)
		}

		println("Library persisted!")
	} else {
		// TODO: code me
		println("Synchronizing an existing library is not supported yet")
	}
}

func initSync(store *storage.Storage, items zotero.ItemsResult) {
	byKey := make(map[string]storage.Item)
	for i := range items.Items {
		item := &items.Items[i]
		if item.Data.ParentKey != "" {
			attach := storage.Attachment{
				Key:         item.Key,
				Version:     item.Version,
				ContentType: item.Data.ContentType,
				Filename:    item.Data.Filename,
			}
			if parent, exists := byKey[item.Data.ParentKey]; exists {
				parent.Attachments = append(parent.Attachments, attach)
				byKey[item.Data.ParentKey] = parent
			} else {
				byKey[item.Data.ParentKey] = storage.Item{
					Attachments: []storage.Attachment{attach},
				}
			}
		} else {
			if existing, exists := byKey[item.Key]; exists {
				// Already present, only attachments
				byKey[item.Key] = storage.Item{
					Key:         item.Key,
					Version:     item.Version,
					Title:       item.Data.Title,
					Abstract:    item.Data.Abstract,
					Creators:    item.Data.Creators,
					Attachments: existing.Attachments,
				}
			} else {
				byKey[item.Key] = storage.Item{
					Key:         item.Key,
					Version:     item.Version,
					Title:       item.Data.Title,
					Abstract:    item.Data.Abstract,
					Creators:    item.Data.Creators,
					Attachments: []storage.Attachment{},
				}
			}
		}
	}

	store.Data.Lib.Version = items.Version
	store.Data.Lib.Items = make([]storage.Item, 0, len(byKey))
	for _, item := range byKey {
		store.Data.Lib.Items = append(store.Data.Lib.Items, item)
	}
}
