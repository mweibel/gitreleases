package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	httpClient *http.Client
	client     *githubv4.Client
}

type releaseAssetNodes []struct {
	DownloadUrl string
}

type fetchSpecificTag struct {
	Repository struct {
		Release *struct {
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

type GitHubErrorType int

const (
	TypeNotFound GitHubErrorType = iota
	TypeServerError
)

type GitHubError struct {
	WrappedError error
	Type         GitHubErrorType
}

func (e GitHubError) Error() string {
	return e.WrappedError.Error()
}

func NewGitHubError(text string, t GitHubErrorType) GitHubError {
	return GitHubError{errors.New(text), t}
}

var (
	errReleaseNotFound = NewGitHubError("github: no release found", TypeNotFound)
	errAssetNotFound   = NewGitHubError("github: asset not found", TypeNotFound)
)

// parseGraphqlError translates between the unfortunately opaque error type of the graphql library and our own.
// Specific to GitHub errors and possible to fail if GitHub changes the error messages.
func parseGraphqlError(err error) GitHubError {
	text := err.Error()
	if strings.Contains(text, "Could not resolve to") {
		return GitHubError{err, TypeNotFound}
	}
	return GitHubError{err, TypeServerError}
}

func (gh *GithubClient) fetchLatestRelease(ctx context.Context, owner, repo, assetName string) (releaseAssetNodes, error) {
	q := fetchLatestRelease{}
	variables := map[string]interface{}{
		"owner":     githubv4.String(owner),
		"repo":      githubv4.String(repo),
		"assetName": githubv4.String(assetName),
	}

	err := gh.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, parseGraphqlError(err)
	}

	releases := q.Repository.Releases.Nodes
	if len(releases) == 0 {
		return nil, errReleaseNotFound
	}

	assets := releases[0].ReleaseAssets.Nodes

	return assets, nil
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
	if err != nil {
		return nil, parseGraphqlError(err)
	}

	release := q.Repository.Release
	if release == nil {
		return nil, errReleaseNotFound
	}
	assets := release.ReleaseAssets.Nodes

	return assets, nil
}

func (gh *GithubClient) FetchReleaseURL(ctx context.Context, owner, repo, tag, assetName string) (string, error) {
	var assets releaseAssetNodes
	var err error
	if tag == "latest" {
		assets, err = gh.fetchLatestRelease(ctx, owner, repo, assetName)
	} else {
		assets, err = gh.fetchSpecificTag(ctx, owner, repo, tag, assetName)
	}
	if err != nil {
		return "", err
	}
	if len(assets) == 0 {
		return "", errAssetNotFound
	}

	return assets[0].DownloadUrl, nil
}

func NewOauthClient(ctx context.Context, token string) *http.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return oauth2.NewClient(ctx, src)
}

func NewGitHubClient(url string, httpClient *http.Client) *GithubClient {
	gc := GithubClient{
		httpClient: httpClient,
	}

	gc.client = githubv4.NewEnterpriseClient(url, gc.httpClient)

	return &gc
}
