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
	"strings"
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
	var client http.Client
	//nolint:bodyclose // It is actually closed inside the function.
	_, respBody, err := makeReq(url, key, &client)
	if err != nil {
		return nil, err
	}
	var apiKeyInfo apiKey
	if err := json.Unmarshal(respBody, &apiKeyInfo); err != nil {
		return nil, NewErrJSON(err)
	}
	return &Zotero{key, baseURL, client, apiKeyInfo}, nil
}

func makeReq(url, key string, client *http.Client) (*http.Response, []byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, NewErrWrongURL(url, err)
	}
	setHeaders(req, key)
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, NewErrMakeReq(*req, err)
	}
	if err := checkStatusCode(resp, http.StatusOK); err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, NewErrReadBody(err)
	}
	return resp, respBody, nil
}

func setHeaders(req *http.Request, key string) {
	req.Header.Add(apiVersionHeader, fmt.Sprint(apiVersion))
	req.Header.Add(authHeader, key)
}

func checkStatusCode(resp *http.Response, expected int) error {
	if resp.StatusCode != expected {
		return NewErrWrongStatus(resp.StatusCode, expected)
	}
	return nil
}

type ItemsResult struct {
	Items   []Item
	Version uint
}

func (z *Zotero) items(url string) (*ItemsResult, uint64, error) {
	fmt.Printf("Requesting items %s\n", url)
	resp, respBody, err := makeReq(url, z.key, &z.client)
	if err != nil {
		return nil, 0, err
	}
	totalItems, err := strconv.ParseUint(resp.Header.Get(totalResHeader), 10, 64)
	if err != nil {
		return nil, 0, NewErrParseHeader(totalResHeader, err)
	}
	items := []Item{}
	if err := json.Unmarshal(respBody, &items); err != nil {
		return nil, totalItems, NewErrJSON(err)
	}
	version, err := parseVersionHeader(resp)
	if err != nil {
		return nil, totalItems, err
	}
	return &ItemsResult{items, uint(version)}, totalItems, nil
}

func parseVersionHeader(resp *http.Response) (uint, error) {
	versionH := resp.Header.Get(lastModifiedHeader)
	version, err := strconv.ParseUint(versionH, 10, 64)
	if err != nil {
		return 0, NewErrParseHeader(lastModifiedHeader, err)
	}
	return uint(version), nil
}

func (z *Zotero) Items(start, limit uint) (*ItemsResult, bool, error) {
	url := fmt.Sprintf("%s/users/%d/items?limit=%d&start=%d",
		z.url, z.userInfo.UserID, limit, start)
	items, totalItems, err := z.items(url)
	return items, uint64(start+limit) < totalItems, err
}

func (z *Zotero) ItemsByKey(keys []string) (*ItemsResult, error) {
	url := fmt.Sprintf("%s/users/%d/items?itemKey=%s",
		z.url, z.userInfo.UserID, strings.Join(keys, ","))
	items, _, err := z.items(url)
	return items, err
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

type VersionsReply struct {
	Versions   map[string]uint
	LibVersion uint
}

func (z *Zotero) VersionsSince(since uint) (versions VersionsReply, err error) {
	url := fmt.Sprintf("%s/users/%d/items?format=versions&since=%d",
		z.url, z.userInfo.UserID, since)
	resp, respBody, err := makeReq(url, z.key, &z.client)
	if err != nil {
		return
	}
	if err = json.Unmarshal(respBody, &versions.Versions); err != nil {
		return versions, NewErrJSON(err)
	}
	versions.LibVersion, err = parseVersionHeader(resp)
	return
}
