function updateText(elements, text) {
  elements.forEach(function(el) {
    el.innerText = text;
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
function clearInput(e) {
  e.target.value = '';
}

function onDocumentLoad() {
  const inputOrganization = document.querySelector(".input-organization");
  const inputRepo = document.querySelector(".input-repo");
  const inputSubmit = document.querySelector('.input-submit');
  const ghReleasesSearch = document.querySelector(".gh-releases-search");
  const ghReleasesResult = document.querySelector(".gh-releases-results");

  inputOrganization.addEventListener("click", clearInput, {
    once: true
  })
  inputRepo.addEventListener("click", clearInput, {
    once: true
  })

  ghReleasesSearch.addEventListener("submit", function(event) {
    event.preventDefault();
    event.stopPropagation();

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

        let message = error.statusText || error.message;
        if (message === 'Forbidden') {
          message = `${message} - most likely means that you exceeded the hourly rate limit, sorry. Try constructing the URL on your own please :)`
        }
        const li = document.createElement("li");
        li.appendChild(document.createTextNode(message));
        ghReleasesResult.appendChild(li);
      });

    return false;
  });
}

if(document.readyState === 'loading') {
  document.addEventListener("DOMContentLoaded", onDocumentLoad);
} else {
  onDocumentLoad();
}
