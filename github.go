package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/inconshreveable/log15"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	httpClient *http.Client
	cache      Cacher
	client     *githubv4.Client
	logger     log.Logger
}

type rateLimit struct {
	Limit     int
	Cost      int
	Remaining int
	ResetAt   time.Time
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
	RateLimit rateLimit
}
type fetchLatestRelease struct {
	Repository struct {
		Releases struct {
			Nodes []struct {
				ReleaseAssets struct {
					TotalCount int
					Nodes      releaseAssetNodes
				} `graphql:"releaseAssets(name: $assetName, first:1)"`
			}
		} `graphql:"releases(first: 5, orderBy: {direction: DESC, field: CREATED_AT})"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
	RateLimit rateLimit
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

func (gh *GithubClient) fetchLatestRelease(ctx context.Context, owner, repo, assetName string) (releaseAssetNodes, rateLimit, error) {
	q := fetchLatestRelease{}
	variables := map[string]interface{}{
		"owner":     githubv4.String(owner),
		"repo":      githubv4.String(repo),
		"assetName": githubv4.String(assetName),
	}

	err := gh.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, q.RateLimit, parseGraphqlError(err)
	}

	releases := q.Repository.Releases.Nodes
	if len(releases) == 0 {
		return nil, q.RateLimit, errReleaseNotFound
	}
	for _, node := range releases {
		if node.ReleaseAssets.TotalCount > 0 {
			return node.ReleaseAssets.Nodes, q.RateLimit, nil
		}
	}

	return nil, q.RateLimit, errAssetNotFound
}

func (gh *GithubClient) fetchSpecificTag(ctx context.Context, owner, repo, tag, assetName string) (releaseAssetNodes, rateLimit, error) {
	q := fetchSpecificTag{}
	variables := map[string]interface{}{
		"owner":     githubv4.String(owner),
		"repo":      githubv4.String(repo),
		"tag":       githubv4.String(tag),
		"assetName": githubv4.String(assetName),
	}

	err := gh.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, q.RateLimit, parseGraphqlError(err)
	}

	release := q.Repository.Release
	if release == nil {
		return nil, q.RateLimit, errReleaseNotFound
	}
	assets := release.ReleaseAssets.Nodes

	return assets, q.RateLimit, nil
}

// FetchReleaseURL decides based on the supplied `tag` which GraphQL query is executed.
func (gh *GithubClient) FetchReleaseURL(ctx context.Context, owner, repo, tag, assetName string) (string, error) {
	var err error

	cacheKey := fmt.Sprintf("%s/%s/%s/%s", owner, repo, tag, assetName)
	cached, err := gh.cache.Get(cacheKey)
	if cached != "" || err != nil {
		return cached, err
	}

	var assets releaseAssetNodes
	var currLimit rateLimit
	if tag == "latest" {
		assets, currLimit, err = gh.fetchLatestRelease(ctx, owner, repo, assetName)
	} else {
		assets, currLimit, err = gh.fetchSpecificTag(ctx, owner, repo, tag, assetName)
	}

	if currLimit.Limit > 0 && currLimit.Remaining < 50 {
		gh.logger.Crit("almost no points remaining", "limit", currLimit.Limit, "cost", currLimit.Cost, "remaining", currLimit.Remaining, "resetAt", currLimit.ResetAt)
	} else {
		gh.logger.Info("current rate limit points", "limit", currLimit.Limit, "cost", currLimit.Cost, "remaining", currLimit.Remaining, "resetAt", currLimit.ResetAt)
	}

	if err != nil {
		gh.cache.Put(cacheKey, "", err)
		return "", err
	}

	if len(assets) == 0 {
		gh.cache.Put(cacheKey, "", errAssetNotFound)
		return "", errAssetNotFound
	}

	url := assets[0].DownloadUrl
	gh.cache.Put(cacheKey, url, nil)

	return url, nil
}

// NewOauthClient creates an oauth2 client with a static token source to use with GitHub's personal access tokens.
func NewOauthClient(ctx context.Context, token string) *http.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return oauth2.NewClient(ctx, src)
}

// NewGitHubClient creates a GithubClient "enterprise" instance using an established oauth2 HTTP client.
//
// The url and httpClient are parameters mainly for proper testing purposes.
func NewGitHubClient(url string, httpClient *http.Client, cache Cacher, logger log.Logger) *GithubClient {
	gc := GithubClient{
		httpClient: httpClient,
		cache:      cache,
		logger:     logger,
	}

	gc.client = githubv4.NewEnterpriseClient(url, gc.httpClient)

	return &gc
}
