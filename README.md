# Git Releases

Git Releases allows you to directly link to the assets of your latest release on GitHub.

## Contributions

For any significant change, please open an issue first to discuss it. 

Other than that, anyone is more than welcome to contribute. 

### Instructions

This project requires `go` 1.11.

Run locally:

```bash
# retrieve a personal access token from GitHub on http://github.com/settings/tokens
$ METRICS_USERNAME=gitreleases METRICS_PASSWORD=gitreleases LISTEN_ADDR=":8080" GITHUB_TOKEN="$GITHUB_TOKEN" go run main.go github.go api.go metrics.go cache.go
``` 

Please use `goimports` for formatting the code.

To update anything on the landingpage:
```bash
$ make install
$ make public/style.min.css
# that's a bit convoluted to use, but anyway
$ cp public/style.min.css landingpage/style.min.css

# Then open `landingpage/index.html` and start editing
```

#### Building

```bash
$ make install
$ make
```

#### Deployment

Deployment is done using kubernetes. k8s files are in the `k8s` folder.

```bash
# build, pack, deploy
$ make ship
```

## GitHub API

GitHub API Explorer: https://developer.github.com/v4/explorer/

### GET /gh/{owner}/{repo}/{tag}/{assetName}

```graphql
{
  repository(owner: $owner, name: $repo) {
    release(tagName: $tag) {
      releaseAssets(name: $assetName, first:1) {
        nodes {
          downloadUrl
        }
      }
    }
  }
}
```

### GET /gh/{owner}/{repo}/latest/{assetName}

```graphql
{
  repository(owner: $owner, name: $repo) {
    releases(last: 1) {
      nodes {
        releaseAssets(name: $assetName, first:1) {
          nodes {
            downloadUrl
          }
        }
      }
    }
  }
}
```
