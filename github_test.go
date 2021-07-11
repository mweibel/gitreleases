package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	log "github.com/inconshreveable/log15"
)

func discardLogger() log.Logger {
	logger := log.New()
	logger.SetHandler(log.DiscardHandler())
	return logger
}

type NoopCache struct{}

func (nc *NoopCache) Put(k, v string, err error) {
}
func (nc *NoopCache) Get(k string) (v string, err error) {
	return "", nil
}

func testingHTTPClient(handler http.Handler) (*httptest.Server, func()) {
	s := httptest.NewServer(handler)

	return s, s.Close
}

var fetchReleaseURLResponsesLatest = map[string]struct {
	AssetName   string
	FileName    string
	ReturnValue string
	ReturnError error
}{
	"owner not found": {
		AssetName:   "testing.zip",
		FileName:    "error_owner_not_found.json",
		ReturnValue: "",
		ReturnError: errors.New("Could not resolve to a User with the username 'testing'."),
	},
	"repo not found": {
		AssetName:   "testing.zip",
		FileName:    "error_repo_not_found.json",
		ReturnValue: "",
		ReturnError: errors.New("Could not resolve to a Repository with the name 'testing'."),
	},
	"release not found": {
		AssetName:   "testing.zip",
		FileName:    "error_release_not_found_latest.json",
		ReturnValue: "",
		ReturnError: errReleaseNotFound,
	},
	"asset not found": {
		AssetName:   "testing.zip",
		FileName:    "error_asset_not_found_latest.json",
		ReturnValue: "",
		ReturnError: errAssetNotFound,
	},
	"asset found": {
		AssetName:   "testing.zip",
		FileName:    "ok_asset_found_latest.json",
		ReturnValue: "https://example.com/testing/testing/releases/download/latest/testing.zip",
		ReturnError: nil,
	},
	"ziparchive found": {
		AssetName:   "ziparchive",
		FileName:    "ok_asset_found_archive.json",
		ReturnValue: "https://github.com/testing/testing/archive/1.1.0.zip",
		ReturnError: nil,
	},
	"targzarchive found": {
		AssetName:   "targzarchive",
		FileName:    "ok_asset_found_archive.json",
		ReturnValue: "https://github.com/testing/testing/archive/1.1.0.tar.gz",
		ReturnError: nil,
	},
}

func TestGithubClient_FetchReleaseURL_Latest(t *testing.T) {
	for name, data := range fetchReleaseURLResponsesLatest {
		t.Run(name, func(t *testing.T) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				file, err := os.Open(filepath.Join("test", "fixtures", data.FileName))
				if err != nil {
					panic(err)
				}
				_, err = io.Copy(w, file)
				if err != nil {
					panic(err)
				}
			})
			httpServer, teardown := testingHTTPClient(h)
			defer teardown()

			cache := NoopCache{}
			gh := NewGitHubClient(httpServer.URL, http.DefaultClient, &cache, discardLogger())

			url, err := gh.FetchReleaseURL(context.Background(), "testing", "testing", "latest", data.AssetName)
			if url != data.ReturnValue {
				t.Errorf("url does not match. Expected: '%s', got '%s'", data.ReturnValue, url)
			}
			if fmt.Sprintf("%s", err) != fmt.Sprintf("%s", data.ReturnError) {
				t.Errorf("err does not match. Expected: '%v', got '%v'", data.ReturnError, err)
			}
		})
	}
}

var fetchReleaseURLResponsesTag = map[string]struct {
	AssetName   string
	FileName    string
	ReturnValue string
	ReturnError error
}{
	"owner not found": {
		AssetName:   "testing.zip",
		FileName:    "error_owner_not_found.json",
		ReturnValue: "",
		ReturnError: errors.New("Could not resolve to a User with the username 'testing'."),
	},
	"repo not found": {
		AssetName:   "testing.zip",
		FileName:    "error_repo_not_found.json",
		ReturnValue: "",
		ReturnError: errors.New("Could not resolve to a Repository with the name 'testing'."),
	},
	"release not found": {
		AssetName:   "testing.zip",
		FileName:    "error_release_not_found_tag.json",
		ReturnValue: "",
		ReturnError: errReleaseNotFound,
	},
	"asset not found": {
		AssetName:   "testing.zip",
		FileName:    "error_asset_not_found_tag.json",
		ReturnValue: "",
		ReturnError: errAssetNotFound,
	},
	"asset found": {
		AssetName:   "testing.zip",
		FileName:    "ok_asset_found_tag.json",
		ReturnValue: "https://example.com/testing/testing/releases/download/sometag/testing.zip",
		ReturnError: nil,
	},
	"ziparchive found": {
		AssetName:   "ziparchive",
		FileName:    "ok_asset_found_tag.json",
		ReturnValue: "https://github.com/testing/testing/archive/sometag.zip",
		ReturnError: nil,
	},
	"targzarchive found": {
		AssetName:   "targzarchive",
		FileName:    "ok_asset_found_tag.json",
		ReturnValue: "https://github.com/testing/testing/archive/sometag.tar.gz",
		ReturnError: nil,
	},
}

func TestGithubClient_FetchReleaseURL_Tag(t *testing.T) {
	for name, data := range fetchReleaseURLResponsesTag {
		t.Run(name, func(t *testing.T) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				file, err := os.Open(filepath.Join("test", "fixtures", data.FileName))
				if err != nil {
					panic(err)
				}
				_, err = io.Copy(w, file)
				if err != nil {
					panic(err)
				}
			})
			httpServer, teardown := testingHTTPClient(h)
			defer teardown()

			cache := NoopCache{}
			gh := NewGitHubClient(httpServer.URL, http.DefaultClient, &cache, discardLogger())

			url, err := gh.FetchReleaseURL(context.Background(), "testing", "testing", "sometag", data.AssetName)
			if url != data.ReturnValue {
				t.Errorf("url does not match. Expected: '%s', got '%s'", data.ReturnValue, url)
			}
			if fmt.Sprintf("%s", err) != fmt.Sprintf("%s", data.ReturnError) {
				t.Errorf("err does not match. Expected: '%v', got '%v'", data.ReturnError, err)
			}
		})
	}
}
