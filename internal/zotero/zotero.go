// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package zotero

import (
	"encoding/json"
	"errors"
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

type Error struct {
	err  error
	kind error
}

func (e *Error) Error() string {
	if e.err == nil {
		return e.kind.Error()
	}
	return fmt.Sprintf("%v: %v", e.kind, e.err)
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Kind() error {
	return e.kind
}

var (
	ErrWrongURL = errors.New("failed to create request")
	ErrMakeReq  = errors.New("failed to execute request")
	ErrReadBody = errors.New("failed to read response body")
	ErrJSON     = errors.New("failed to parse data from JSON reply")
)

// ErrWrongStatus is returned if we receive an unexpected response status code
type ErrWrongStatus struct {
	expected, received int
}

func (e *ErrWrongStatus) Error() string {
	return fmt.Sprintf("received %v status code instead of %v", e.received, e.expected)
}

// ErrParseHeader is returned if we fail to parse an header somewhere
type ErrParseHeader struct {
	header string
}

func (e *ErrParseHeader) Error() string {
	return fmt.Sprintf("failed to parse header %q", e.header)
}

func New(key string) (*Zotero, error) {
	return newWithURL(apiURL, key)
}

func newWithURL(baseURL, key string) (*Zotero, error) {
	req, err := http.NewRequest("GET", baseURL+"/keys/current", nil)
	if err != nil {
		return nil, &Error{err, ErrWrongURL}
	}

	setHeaders(req, key)

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return nil, &Error{err, ErrMakeReq}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &Error{nil, &ErrWrongStatus{http.StatusOK, resp.StatusCode}}
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &Error{err, ErrReadBody}
	}

	var apiKeyInfo apiKey
	if err := json.Unmarshal(respBody, &apiKeyInfo); err != nil {
		return nil, &Error{err, ErrJSON}
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
		return nil, false, &Error{err, ErrWrongURL}
	}

	setHeaders(req, z.key)

	resp, err := z.client.Do(req)
	if err != nil {
		return nil, false, &Error{err, ErrMakeReq}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, &Error{nil, &ErrWrongStatus{http.StatusOK, resp.StatusCode}}
	}

	totalItems, err := strconv.ParseUint(resp.Header.Get(totalResHeader), 10, 64)
	if err != nil {
		return nil, false, &Error{err, &ErrParseHeader{totalResHeader}}
	}

	more := uint64(start+limit) < totalItems
	items := []Item{}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, more, &Error{err, ErrReadBody}
	}

	if err := json.Unmarshal(respBody, &items); err != nil {
		return nil, more, &Error{err, ErrJSON}
	}

	versionH := resp.Header.Get(lastModifiedHeader)
	version, err := strconv.ParseUint(versionH, 10, 64)
	if err != nil {
		return nil, more, &Error{err, &ErrParseHeader{lastModifiedHeader}}
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
