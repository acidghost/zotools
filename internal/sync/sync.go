package sync

import (
	"flag"
	"fmt"
	"os"

	"github.com/acidghost/zotools/internal/common"
	. "github.com/acidghost/zotools/internal/common"
	"github.com/acidghost/zotools/internal/zotero"
)

type SyncCommand struct {
	fs *flag.FlagSet
}

func New(cmd string) *SyncCommand {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	return &SyncCommand{fs}
}

func (c *SyncCommand) Run(args []string, config Config) {
	c.fs.Parse(args)

	exists := true
	if _, err := os.Stat(config.Storage); err != nil && os.IsNotExist(err) {
		exists = false
	}

	storage := common.NewStorage(config.Storage)
	if exists {
		if err := storage.Load(); err != nil {
			Dief("Failed to load the local storage:\n - %v\n", err)
		}
	}

	zot, err := zotero.New(config.Key)
	if err != nil {
		Dief("Failed to initialize Zotero API:\n - %v\n", err)
	}

	if storage.Lib.Version == 0 {
		// Initial sync queries all the items
		items, err := zot.AllItems()
		if err != nil {
			Dief("Failed to load items:\n - %v\n", err)
		}

		initSync(&storage, items)
		fmt.Printf("Retrived %d top level items\n", len(storage.Lib.Items))

		if err := storage.PersistLibrary(); err != nil {
			Dief("Failed to persist library:\n - %v\n", err)
		}

		println("Library persisted!")
	} else {
		// TODO: code me
		println("Synchronizing an existing library is not supported yet")
	}
}

func initSync(storage *Storage, items zotero.ItemsResult) {
	byKey := make(map[string]Item)
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
				byKey[item.Data.ParentKey] = Item{
					Attachments: []Attachment{attach},
				}
			}
		} else {
			if existing, exists := byKey[item.Key]; exists {
				// Already present, only attachments
				byKey[item.Key] = Item{
					Key:         item.Key,
					Version:     item.Version,
					Title:       item.Data.Title,
					Abstract:    item.Data.Abstract,
					Creators:    item.Data.Creators,
					Attachments: existing.Attachments,
				}
			} else {
				byKey[item.Key] = Item{
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

	storage.Lib.Version = items.Version
	storage.Lib.Items = make([]Item, 0, len(byKey))
	for _, item := range byKey {
		storage.Lib.Items = append(storage.Lib.Items, item)
	}
}
