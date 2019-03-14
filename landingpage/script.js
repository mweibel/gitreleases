function updateText(elements, text) {
  elements.forEach(function(el) {
    el.innerText = text;
  });
}
function updateGitreleasesLink(elements, organization, repo, filename) {
  elements.forEach(function(el) {
    console.log(el, `/gh/${organization}/${repo}/latest/${filename}`);
    el.setAttribute("href", `/gh/${organization}/${repo}/latest/${filename}`);
  });
}

function clearResults(resultsEl) {
  resultsEl.childNodes.forEach(function(node) {
    node.remove();
  });
}

function filterReleases(response) {
  return response.slice(0, 5).find(function(release) {
    return !!release.assets;
  });
}

function onDocumentLoad() {
  const inputOrganization = document.querySelector(".input-organization");
  const inputRepo = document.querySelector(".input-repo");
  const ghOrganizations = document.querySelectorAll(".gh-organization");
  const ghRepo = document.querySelectorAll(".gh-repo");
  const ghFilename = document.querySelectorAll(".gh-filename");
  const gitreleasesLink = document.querySelectorAll(".gitreleases-link");
  const ghReleasesSearch = document.querySelector(".gh-releases-search");
  const ghReleasesResult = document.querySelector(".gh-releases-results");

  ghReleasesSearch.addEventListener("submit", function(event) {
    event.preventDefault();

    const organization = inputOrganization.value;
    const repo = inputRepo.value;

    const url = `https://api.github.com/repos/${organization}/${repo}/releases`;

    fetch(url, {
      method: "GET",
      headers: {
        Accept: "application/vnd.github.v3+json"
      }
    })
      .then(function(response) {
        if (!response.ok) {
          return Promise.reject(response);
        }
        return response.json();
      })
      .then(function(response) {
        const release = filterReleases(response);
        return release.assets;
      })
      .then(function(assets) {
        if (!assets || !assets.length) {
          return Promise.reject(new Error("No asset found"));
        }
        if (assets.length === 1) {
          updateGitreleasesLink(
            gitreleasesLink,
            organization,
            repo,
            assets[0].name
          );
          return;
        }

        clearResults(ghReleasesResult);

        assets.slice(0, 5).forEach(function(asset) {
          const li = document.createElement("li");
          li.appendChild(document.createTextNode(asset.name));
          ghReleasesResult.appendChild(li);

          li.addEventListener("mouseover", function() {
            updateText(ghFilename, asset.name);
            updateGitreleasesLink(
              gitreleasesLink,
              organization,
              repo,
              asset.name
            );
          });
        });
      })
      .catch(function(error) {
        clearResults(ghReleasesResult);

        const message = error.statusText || error.message;
        const li = document.createElement("li");
        li.appendChild(document.createTextNode(message));
        ghReleasesResult.appendChild(li);
      });
  });

  inputOrganization.addEventListener("keyup", function(event) {
    const { value } = event.target;
    updateText(ghOrganizations, value);
  });
  inputRepo.addEventListener("keyup", function(event) {
    const { value } = event.target;
    updateText(ghRepo, value);
  });
}

document.addEventListener("DOMContentLoaded", onDocumentLoad);
