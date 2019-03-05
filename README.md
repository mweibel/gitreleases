# Gitreleases

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
