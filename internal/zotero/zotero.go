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
	apiURL     = "https://api.zotero.org"
	apiVersion = uint(3)
	MaxLimit   = 100
)

const (
	authHeader         = "Zotero-API-Key"
	apiVersionHeader   = "Zotero-API-Version"
	lastModifiedHeader = "Last-Modified-Version"
	totalResHeader     = "Total-Results"
)

type apiKey struct {
	UserID   uint   `json:"userId"`
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
	url      string
	client   http.Client
	userInfo apiKey
}

type errSpec string

const (
	errWrongURL    = errSpec("wrap:creating request to {{url string %s}}")
	errMakeReq     = errSpec("wrap:executing request to {{req.URL http.Request %s}}")
	errReadBody    = errSpec("wrap:reading response body")
	errJSON        = errSpec("wrap:parsing JSON from reply")
	errWrongStatus = errSpec("nowrap:received {{recv int %v}} status code instead of {{exp int %v}}")
	errParseHeader = errSpec("wrap:parsing header {{header string %q}}")
)

//go:generate gorror -type=errSpec -P -import=net/http

func New(key string) (*Zotero, error) {
	return newWithURL(apiURL, key)
}

func newWithURL(baseURL, key string) (*Zotero, error) {
	url := baseURL + "/keys/current"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, NewErrWrongURL(url, err)
	}

	setHeaders(req, key)

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return nil, NewErrMakeReq(*req, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewErrWrongStatus(resp.StatusCode, http.StatusOK)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewErrReadBody(err)
	}

	var apiKeyInfo apiKey
	if err := json.Unmarshal(respBody, &apiKeyInfo); err != nil {
		return nil, NewErrJSON(err)
	}

	return &Zotero{key, baseURL, client, apiKeyInfo}, nil
}

func setHeaders(req *http.Request, key string) {
	req.Header.Add(apiVersionHeader, fmt.Sprint(apiVersion))
	req.Header.Add(authHeader, key)
}

type ItemsResult struct {
	Items   []Item
	Version uint
}

func (z *Zotero) Items(start, limit uint) (*ItemsResult, bool, error) {
	url := fmt.Sprintf("%s/users/%d/items?limit=%d&start=%d",
		z.url, z.userInfo.UserID, limit, start)

	fmt.Printf("Requesting items %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, NewErrWrongURL(url, err)
	}

	setHeaders(req, z.key)

	resp, err := z.client.Do(req)
	if err != nil {
		return nil, false, NewErrMakeReq(*req, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, NewErrWrongStatus(resp.StatusCode, http.StatusOK)
	}

	totalItems, err := strconv.ParseUint(resp.Header.Get(totalResHeader), 10, 64)
	if err != nil {
		return nil, false, NewErrParseHeader(totalResHeader, err)
	}

	more := uint64(start+limit) < totalItems
	items := []Item{}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, more, NewErrReadBody(err)
	}

	if err := json.Unmarshal(respBody, &items); err != nil {
		return nil, more, NewErrJSON(err)
	}

	versionH := resp.Header.Get(lastModifiedHeader)
	version, err := strconv.ParseUint(versionH, 10, 64)
	if err != nil {
		return nil, more, NewErrParseHeader(lastModifiedHeader, err)
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
