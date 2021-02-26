// (c) Copyright 2021, zotools' Authors.
//
// Licensed under the terms of the GNU AGPL License version 3.

package zotero

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const keyReplyFmt = `{
    "key": "%s",
    "userID": %d,
    "username": "%s",
    "access": {
        "user": {
            "library": true,
            "files": true
        }
    }
}`

func TestNewWithURL(t *testing.T) {
	t.Run("Successful", func(t *testing.T) {
		key, userID, username := "someapikey", uint(1337), "myusername"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			keyHeader := r.Header.Get(authHeader)
			assert.Equal(t, fmt.Sprint(apiVersion), r.Header.Get(apiVersionHeader))
			assert.Equal(t, key, keyHeader)
			fmt.Fprintf(w, keyReplyFmt, keyHeader, userID, username)
		}))
		defer ts.Close()
		z, err := newWithURL(ts.URL, key)
		require.NoError(t, err)
		assert.Equal(t, z.key, key)
		assert.Equal(t, z.userInfo.UserID, userID)
		assert.Equal(t, z.userInfo.Username, username)
	})
	t.Run("Broken URL", func(t *testing.T) {
		z, err := newWithURL("...\x00..somedomain.com", "someapikey")
		require.Error(t, err)
		assert.Nil(t, z)
		checkErrorKind(t, err, ErrWrongURL)
	})
	t.Run("Failed request", func(t *testing.T) {
		ts := httptest.NewUnstartedServer(nil)
		defer ts.Close()
		z, err := newWithURL(ts.URL, "someapikey")
		require.Error(t, err)
		assert.Nil(t, z)
		checkErrorKind(t, err, ErrMakeReq)
	})
	t.Run("Not OK reply", func(t *testing.T) {
		ts := httptest.NewServer(http.NotFoundHandler())
		defer ts.Close()
		z, err := newWithURL(ts.URL, "someapikey")
		require.Error(t, err)
		assert.Nil(t, z)
		var ek *ErrWrongStatus
		checkErrorKindStruct(t, err, &ek)
	})
	t.Run("Invalid JSON reply", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "invalidjson")
		}))
		defer ts.Close()
		z, err := newWithURL(ts.URL, "someapikey")
		require.Error(t, err)
		assert.Nil(t, z)
		checkErrorKind(t, err, ErrJSON)
		var inner *json.SyntaxError
		assert.ErrorAs(t, err, &inner)
	})
}

//go:embed assets/items.json
var itemsReply string

const itemsReplyCount = 10

func TestItems(t *testing.T) {
	itemsReplyCountS := fmt.Sprint(itemsReplyCount)
	t.Run("Successful - no more", func(t *testing.T) {
		const start, version uint = 0, 42
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.Equal(t, fmt.Sprint(start), q.Get("start"))
			assert.Equal(t, fmt.Sprint(MaxLimit), q.Get("limit"))
			w.Header().Add(totalResHeader, itemsReplyCountS)
			w.Header().Add(lastModifiedHeader, fmt.Sprint(version))
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, itemsReply)
		}))
		defer ts.Close()
		res, more, err := zotFromServer(ts).Items(start, MaxLimit)
		require.NoError(t, err)
		assert.Falsef(t, more, "expected no more items")
		assert.Equal(t, res.Version, version)
		assert.Len(t, res.Items, itemsReplyCount)
	})
	t.Run("Successful - more", func(t *testing.T) {
		const start, limit, version uint = 0, itemsReplyCount, 42
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const tot uint = 100
			require.Greater(t, tot, limit)
			w.Header().Add(totalResHeader, fmt.Sprint(tot))
			w.Header().Add(lastModifiedHeader, fmt.Sprint(version))
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, itemsReply)
		}))
		defer ts.Close()
		res, more, err := zotFromServer(ts).Items(start, limit)
		require.NoError(t, err)
		assert.Truef(t, more, "expected more items")
		assert.Equal(t, res.Version, version)
		assert.Len(t, res.Items, int(limit))
	})
	t.Run("Failed request", func(t *testing.T) {
		ts := httptest.NewUnstartedServer(nil)
		defer ts.Close()
		res, _, err := zotFromServer(ts).Items(0, MaxLimit)
		require.Error(t, err)
		assert.Nil(t, res)
		checkErrorKind(t, err, ErrMakeReq)
		var inner *url.Error
		assert.ErrorAs(t, err, &inner)
	})
	t.Run("Broken URL", func(t *testing.T) {
		var client http.Client
		z := Zotero{"someapikey", "http://bad\x00url.com", client, apiKey{}}
		res, _, err := z.Items(0, MaxLimit)
		require.Error(t, err)
		assert.Nil(t, res)
		checkErrorKind(t, err, ErrWrongURL)
	})
	t.Run("Status not OK", func(t *testing.T) {
		ts := httptest.NewServer(http.NotFoundHandler())
		defer ts.Close()
		res, _, err := zotFromServer(ts).Items(0, MaxLimit)
		assert.Nil(t, res)
		assert.Error(t, err)
		var ek *ErrWrongStatus
		checkErrorKindStruct(t, err, &ek)
	})
	t.Run("Wrong total results header", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(totalResHeader, "asd")
			w.Header().Add(lastModifiedHeader, "42")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, itemsReply)
		}))
		defer ts.Close()
		res, _, err := zotFromServer(ts).Items(0, MaxLimit)
		assert.Nil(t, res)
		assert.Error(t, err)
		var ek *ErrParseHeader
		checkErrorKindStruct(t, err, &ek)
		assert.Equal(t, ek.header, totalResHeader)
	})
	t.Run("Invalid JSON reply", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(totalResHeader, "1337")
			w.Header().Add(lastModifiedHeader, "42")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "invalidjson")
		}))
		defer ts.Close()
		res, _, err := zotFromServer(ts).Items(0, MaxLimit)
		require.Error(t, err)
		assert.Nil(t, res)
		checkErrorKind(t, err, ErrJSON)
		var inner *json.SyntaxError
		assert.ErrorAs(t, err, &inner)
	})
	t.Run("Wrong version header", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(totalResHeader, "42")
			w.Header().Add(lastModifiedHeader, "asd")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, itemsReply)
		}))
		defer ts.Close()
		res, _, err := zotFromServer(ts).Items(0, MaxLimit)
		assert.Nil(t, res)
		assert.Error(t, err)
		var ek *ErrParseHeader
		checkErrorKindStruct(t, err, &ek)
		assert.Equal(t, ek.header, lastModifiedHeader)
	})
}

func TestAllItems(t *testing.T) {
	t.Run("Successful single", func(t *testing.T) {
		const version uint = 42
		requests := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(totalResHeader, fmt.Sprint(itemsReplyCount))
			w.Header().Add(lastModifiedHeader, fmt.Sprint(version))
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, itemsReply)
			requests++
		}))
		defer ts.Close()
		res, err := zotFromServer(ts).AllItems()
		assert.NoError(t, err)
		assert.Equal(t, res.Version, version)
		assert.Len(t, res.Items, itemsReplyCount)
		assert.Equal(t, requests, 1)
	})
	t.Run("Successful multiple", func(t *testing.T) {
		const version uint = 42
		requests := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(totalResHeader, fmt.Sprint(MaxLimit*2))
			w.Header().Add(lastModifiedHeader, fmt.Sprint(version))
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, itemsReply)
			requests++
		}))
		defer ts.Close()
		res, err := zotFromServer(ts).AllItems()
		assert.NoError(t, err)
		assert.Equal(t, res.Version, version)
		assert.Len(t, res.Items, itemsReplyCount*2)
		assert.Equal(t, requests, 2)
	})
	t.Run("Error Items", func(t *testing.T) {
		ts := httptest.NewServer(http.NotFoundHandler())
		defer ts.Close()
		_, err := zotFromServer(ts).AllItems()
		var ek *ErrWrongStatus
		checkErrorKindStruct(t, err, &ek)
	})
}

func checkErrorKind(t *testing.T, err, kind error) {
	t.Helper()
	var e *Error
	if assert.ErrorAs(t, err, &e) {
		assert.ErrorIs(t, e.Kind(), kind)
	}
}

func checkErrorKindStruct(t *testing.T, err error, kind interface{}) {
	t.Helper()
	var e *Error
	require.ErrorAs(t, err, &e)
	require.ErrorAs(t, e.Kind(), kind)
}

func zotFromServer(ts *httptest.Server) *Zotero {
	client := ts.Client()
	if client == nil {
		client = http.DefaultClient
	}
	return &Zotero{"someapikey", ts.URL, *client, apiKey{}}
}
