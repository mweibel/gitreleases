package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var (
	errNoReleaseFound = errors.New("github: no release found")
	errAssetNotFound  = errors.New("github: asset not found")
)

type GithubClient struct {
	oauthClient *http.Client
	client      *githubv4.Client
}

type releaseAssetNodes []struct {
	DownloadUrl string
}

type fetchSpecificTag struct {
	Repository struct {
		Release struct {
			ReleaseAssets struct {
				Nodes releaseAssetNodes
			} `graphql:"releaseAssets(name: $assetName, first:1)"`
		} `graphql:"release(tagName: $tag)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}
type fetchLatestRelease struct {
	Repository struct {
		Releases struct {
			Nodes []struct {
				ReleaseAssets struct {
					Nodes releaseAssetNodes
				} `graphql:"releaseAssets(name: $assetName, first:1)"`
			}
		} `graphql:"releases(last: 1)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

func (gh *GithubClient) fetchLatestRelease(ctx context.Context, owner, repo, assetName string) (releaseAssetNodes, error) {
	q := fetchLatestRelease{}
	variables := map[string]interface{}{
		"owner":     githubv4.String(owner),
		"repo":      githubv4.String(repo),
		"assetName": githubv4.String(assetName),
	}

	err := gh.client.Query(ctx, &q, variables)

	releases := q.Repository.Releases.Nodes
	if len(releases) == 0 {
		return nil, errNoReleaseFound
	}

	assets := releases[0].ReleaseAssets.Nodes

	return assets, err
}

func (gh *GithubClient) fetchSpecificTag(ctx context.Context, owner, repo, tag, assetName string) (releaseAssetNodes, error) {
	q := fetchSpecificTag{}
	variables := map[string]interface{}{
		"owner":     githubv4.String(owner),
		"repo":      githubv4.String(repo),
		"tag":       githubv4.String(tag),
		"assetName": githubv4.String(assetName),
	}

	err := gh.client.Query(ctx, &q, variables)

	release := q.Repository.Release
	assets := release.ReleaseAssets.Nodes

	return assets, err
}

func (gh *GithubClient) FetchReleaseURL(ctx context.Context, owner, repo, tag, assetName string) (string, error) {
	var assets releaseAssetNodes
	var err error
	if tag == "latest" {
		assets, err = gh.fetchLatestRelease(ctx, owner, repo, assetName)
	} else {
		assets, err = gh.fetchSpecificTag(ctx, owner, repo, tag, assetName)
	}
	if len(assets) == 0 {
		return "", errAssetNotFound
	}

	return assets[0].DownloadUrl, err
}

func NewGitHubClient(ctx context.Context, token string) *GithubClient {
	gc := GithubClient{}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	gc.oauthClient = oauth2.NewClient(ctx, src)

	gc.client = githubv4.NewClient(gc.oauthClient)

	return &gc
}
