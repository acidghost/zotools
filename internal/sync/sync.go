// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package sync

import (
	"flag"
	"fmt"
	"os"

	"github.com/acidghost/zotools/internal/common"
	"github.com/acidghost/zotools/internal/zotero"
)

const syncUsageTop = " " + common.OptionsUsage

type Command struct {
	fs       *flag.FlagSet
	flagDrop *bool
}

func New(cmd, banner string) *Command {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	flagDrop := fs.Bool("drop", false, "delete storage and start fresh")
	fs.Usage = common.MakeUsage(fs, cmd, banner, syncUsageTop, "")
	return &Command{fs, flagDrop}
}

func (c *Command) Run(args []string, config common.Config) {
	//nolint:errcheck
	c.fs.Parse(args)

	exists := true
	if _, err := os.Stat(config.Storage); err != nil && os.IsNotExist(err) {
		exists = false
	}

	storage := common.NewStorage(config.Storage)
	if *c.flagDrop && exists {
		if err := storage.Drop(); err != nil {
			common.Die("Failed to drop storage:\n - %v\n", err)
		}
	} else if exists {
		if err := storage.Load(); err != nil {
			common.Die("Failed to load the local storage:\n - %v\n", err)
		}
	}

	zot, err := zotero.New(config.Key)
	if err != nil {
		common.Die("Failed to initialize Zotero API:\n - %v\n", err)
	}

	if storage.Data.Lib.Version == 0 {
		// Initial sync queries all the items
		items, err := zot.AllItems()
		if err != nil {
			common.Die("Failed to load items:\n - %v\n", err)
		}

		initSync(&storage, items)
		fmt.Printf("Retrieved %d top level items\n", len(storage.Data.Lib.Items))

		if err := storage.Persist(); err != nil {
			common.Die("Failed to persist library:\n - %v\n", err)
		}

		println("Library persisted!")
	} else {
		// TODO: code me
		println("Synchronizing an existing library is not supported yet")
	}
}

func initSync(storage *common.Storage, items zotero.ItemsResult) {
	byKey := make(map[string]common.Item)
	for i := range items.Items {
		item := &items.Items[i]
		if item.Data.ParentKey != "" {
			attach := common.Attachment{
				Key:         item.Key,
				Version:     item.Version,
				ContentType: item.Data.ContentType,
				Filename:    item.Data.Filename,
			}
			if parent, exists := byKey[item.Data.ParentKey]; exists {
				parent.Attachments = append(parent.Attachments, attach)
				byKey[item.Data.ParentKey] = parent
			} else {
				byKey[item.Data.ParentKey] = common.Item{
					Attachments: []common.Attachment{attach},
				}
			}
		} else {
			if existing, exists := byKey[item.Key]; exists {
				// Already present, only attachments
				byKey[item.Key] = common.Item{
					Key:         item.Key,
					Version:     item.Version,
					Title:       item.Data.Title,
					Abstract:    item.Data.Abstract,
					Creators:    item.Data.Creators,
					Attachments: existing.Attachments,
				}
			} else {
				byKey[item.Key] = common.Item{
					Key:         item.Key,
					Version:     item.Version,
					Title:       item.Data.Title,
					Abstract:    item.Data.Abstract,
					Creators:    item.Data.Creators,
					Attachments: []common.Attachment{},
				}
			}
		}
	}

	storage.Data.Lib.Version = items.Version
	storage.Data.Lib.Items = make([]common.Item, 0, len(byKey))
	for _, item := range byKey {
		storage.Data.Lib.Items = append(storage.Data.Lib.Items, item)
	}
}
