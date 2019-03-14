function updateText(elements, text) {
  elements.forEach(function(el) {
    el.innerText = text;
  });
}

function updateGitreleasesLink(elements, organization, repo, filename) {
  elements.forEach(function(el) {
    el.setAttribute("href", `/gh/${organization}/${repo}/latest/${filename}`);
  });
}

function clearResults(resultsEl) {
  while (resultsEl.firstChild) {
    resultsEl.removeChild(resultsEl.firstChild);
  }
}

function filterReleases(response) {
  return response.slice(0, 5).find(function(release) {
    return !!release.assets;
  });
}

function createGitreleasesLink (path) {
  const a = document.createElement('a')
  a.className = 'primary no-underline underline-hover'
  a.setAttribute('rel', 'noopener')
  a.setAttribute('href', path)
  a.appendChild(document.createTextNode(`gitreleases.dev${path}`))
  return a
}

function setButtonLoading(btn) {
  btn.className += ' gray'
  btn.setAttribute('disabled', true)
  btn.value = 'Loading...'
}
function resetButton(btn) {
  btn.classList.remove('gray')
  btn.removeAttribute('disabled');
  btn.value = 'Search'
}

function onDocumentLoad() {
  const inputOrganization = document.querySelector(".input-organization");
  const inputRepo = document.querySelector(".input-repo");
  const inputSubmit = document.querySelector('.input-submit');
  const ghOrganizations = document.querySelectorAll(".gh-organization");
  const ghRepo = document.querySelectorAll(".gh-repo");
  const ghFilename = document.querySelectorAll(".gh-filename");
  const ghReleasesSearch = document.querySelector(".gh-releases-search");
  const ghReleasesResult = document.querySelector(".gh-releases-results");

  ghReleasesSearch.addEventListener("submit", function(event) {
    event.preventDefault();

    setButtonLoading(inputSubmit)

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

        clearResults(ghReleasesResult);

        assets.slice(0, 5).forEach(function(asset) {
          const path = `/gh/${organization}/${repo}/latest/${asset.name}`
          const li = document.createElement("li");
          li.className = 'pb2';
          li.appendChild(createGitreleasesLink(path))

          ghReleasesResult.appendChild(li);
        });
        resetButton(inputSubmit)
      })
      .catch(function(error) {
        resetButton(inputSubmit)
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
