// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package zotero

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	ApiUrl     = "https://api.zotero.org"
	ApiVersion = uint(3)
	MaxLimit   = 100
)

const (
	authHeader       = "Zotero-API-Key"
	apiVersionHeader = "Zotero-API-Version"
)

type ApiKey struct {
	UserId   uint   `json:"userId"`
	Username string `json:"username"`
}

type Item struct {
	Key     string   `json:"key"`
	Version uint     `json:"version"`
	Data    ItemData `json:"data"`
}

type ItemData struct {
	Title       string    `json:"title"`
	Abstract    string    `json:"abstractNote"`
	ItemType    string    `json:"itemType"`
	Creators    []Creator `json:"creators"`
	ParentKey   string    `json:"parentItem,omitempty"`
	ContentType string    `json:"contentType,omitempty"`
	Filename    string    `json:"filename,omitempty"`
}

type Creator struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Zotero struct {
	key      string
	client   http.Client
	UserInfo ApiKey
}

func setHeaders(req *http.Request, key string) {
	req.Header.Add(apiVersionHeader, fmt.Sprint(ApiVersion))
	req.Header.Add(authHeader, key)
}

func New(key string) (*Zotero, error) {
	req, err := http.NewRequest("GET", ApiUrl+"/keys/current", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for API key info: %v", err)
	}

	setHeaders(req, key)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request for API key info: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received %v status code while getting API key info", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API key response body: %v", err)
	}

	var apiKeyInfo ApiKey
	if err := json.Unmarshal(respBody, &apiKeyInfo); err != nil {
		return nil, fmt.Errorf("failed to parse API key info from JSON reply: %v", err)
	}

	return &Zotero{key, client, apiKeyInfo}, nil
}

type ItemsResult struct {
	Items   []Item
	Version uint
}

func (z *Zotero) Items(start, limit uint) (*ItemsResult, bool, error) {
	url := fmt.Sprintf("%s/users/%v/items?limit=%v&start=%v",
		ApiUrl, z.UserInfo.UserId, limit, start)

	fmt.Printf("Requesting items %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to create request: %v", err)
	}

	setHeaders(req, z.key)

	resp, err := z.client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to send request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("Received %v status code", resp.StatusCode)
	}

	totalItems, err := strconv.ParseUint(resp.Header.Get("Total-Results"), 10, 64)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to parse number of total results from header: %v", err)
	}

	more := uint64(start+limit) < totalItems
	items := []Item{}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, more, fmt.Errorf("Failed to read reply body: %v", err)
	}

	if err := json.Unmarshal(respBody, &items); err != nil {
		return nil, more, fmt.Errorf("Failed to parse JSON from reply: %v", err)
	}

	versionH := resp.Header.Get("Last-Modified-Version")
	version, err := strconv.ParseUint(versionH, 10, 64)
	if err != nil {
		return nil, more, fmt.Errorf("Failed to parse version from header '%v': %v", versionH, err)
	}

	return &ItemsResult{items, uint(version)}, more, nil
}

func (z *Zotero) AllItems() (ItemsResult, error) {
	ir := ItemsResult{Items: []Item{}}
	var start uint = 0
	for {
		itemsRes, more, err := z.Items(start, MaxLimit)
		if err != nil {
			return ir, err
		}
		ir.Version = itemsRes.Version
		ir.Items = append(ir.Items, itemsRes.Items...)
		if !more {
			return ir, nil
		}
		start += MaxLimit
	}
}
