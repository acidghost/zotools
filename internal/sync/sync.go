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
	} else {
		// Update existing library. First query versions
		vers, err := zot.VersionsSince(store.Data.Lib.Version)
		if err != nil {
			utils.Die("Failed to query versions:\n - %v\n", err)
		}
		switch {
		case vers.LibVersion > store.Data.Lib.Version:
			toQuery := make([]string, 0)
			for itemKey, v := range vers.Versions {
				if item, ok := store.Data.Lib.Items[itemKey]; !ok || item.Version < v {
					toQuery = append(toQuery, itemKey)
				}
			}
			if len(toQuery) == 0 {
				// No outdated local copies or new items, bail and notify user
				fmt.Printf("Local library already up to date.\n")
				zotItems := make([]string, 0, len(vers.Versions))
				for key := range vers.Versions {
					zotItems = append(zotItems, key)
				}
				fmt.Printf("Zotero suggested new versions for: %v\n", zotItems)
				return
			}
			// Query selected items and update storage
			updatedItems, err := zot.ItemsByKey(toQuery)
			if err != nil {
				utils.Die("Failed to query updated items:\n - %v\n", err)
			}
			if updatedItems.Version != vers.LibVersion {
				utils.Die("Library updated while getting items, try again\n")
			}
			update(&store, vers, *updatedItems)
		case vers.LibVersion < store.Data.Lib.Version:
			utils.Die("Unexpected old library version: %d < %d\n",
				vers.LibVersion, store.Data.Lib.Version)
		default:
			println("Library already up to date")
			return
		}
	}

	if err := store.Persist(); err != nil {
		utils.Die("Failed to persist library:\n - %v\n", err)
	}
	println("Library persisted!")
}

func initSync(store *storage.Storage, items zotero.ItemsResult) {
	byKey := make(storage.ItemsMap)
	for i := range items.Items {
		item := &items.Items[i]
		if item.Data.ParentKey != "" {
			attach := storage.Attachment{
				Version:     item.Version,
				ContentType: item.Data.ContentType,
				Filename:    item.Data.Filename,
			}
			if parent, exists := byKey[item.Data.ParentKey]; exists {
				parent.Attachments[item.Key] = attach
				byKey[item.Data.ParentKey] = parent
			} else {
				byKey[item.Data.ParentKey] = storage.Item{
					Attachments: storage.AttachmentsMap{item.Key: attach},
				}
			}
		} else {
			newItem := storage.Item{
				Version:     item.Version,
				Title:       item.Data.Title,
				Abstract:    item.Data.Abstract,
				Creators:    item.Data.Creators,
				Attachments: nil,
			}
			if existing, exists := byKey[item.Key]; exists {
				// Already present, only attachments
				newItem.Attachments = existing.Attachments
				byKey[item.Key] = newItem
			} else {
				newItem.Attachments = make(storage.AttachmentsMap)
				byKey[item.Key] = newItem
			}
		}
	}

	store.Data.Lib.Version = items.Version
	store.Data.Lib.Items = byKey
}

func update(store *storage.Storage, vers zotero.VersionsReply, updatedItems zotero.ItemsResult) {
	childItems := make(map[string]*zotero.Item)
	for updatedItemIdx := range updatedItems.Items {
		updatedItem := &updatedItems.Items[updatedItemIdx]
		if len(updatedItem.Data.ParentKey) == 0 {
			newItem := storage.Item{
				Version:     updatedItem.Version,
				Title:       updatedItem.Data.Title,
				Abstract:    updatedItem.Data.Abstract,
				Creators:    updatedItem.Data.Creators,
				Attachments: nil,
			}
			if oldItem, ok := store.Data.Lib.Items[updatedItem.Key]; ok {
				newItem.Attachments = oldItem.Attachments
				store.Data.Lib.Items[updatedItem.Key] = newItem
			} else {
				newItem.Attachments = make(storage.AttachmentsMap)
				store.Data.Lib.Items[updatedItem.Key] = newItem
			}
		} else {
			childItems[updatedItem.Data.ParentKey] = updatedItem
		}
	}
	for parentKey, childItem := range childItems {
		if parent, ok := store.Data.Lib.Items[parentKey]; ok {
			parent.Attachments[childItem.Key] = storage.Attachment{
				Version:     childItem.Version,
				ContentType: childItem.Data.ContentType,
				Filename:    childItem.Data.Filename,
			}
		} else {
			utils.Die("Unexpected child w/o a parent: %v\n", childItem)
		}
	}
	store.Data.Lib.Version = vers.LibVersion
	// TODO: update deleted items
}
