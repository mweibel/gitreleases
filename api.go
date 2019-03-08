package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const requestTimeout = 2 * time.Second

type apiServer struct {
	server       *http.Server
	router       *mux.Router
	githubClient *GithubClient
	logger       log.Logger
	version      string
}

// Start is starting the HTTP server.
func (as *apiServer) Start() error {
	return as.server.ListenAndServe()
}

// Shutdown stops the HTTP server, possibly gracefully if an according context is provided.
func (as *apiServer) Shutdown(ctx context.Context) error {
	return as.server.Shutdown(ctx)
}

// DownloadRelease fetches a release from GitHub according to parameters specified.
func (as *apiServer) DownloadRelease(w http.ResponseWriter, r *http.Request) {
	reqLogger := as.logger.New("method", r.Method, "url", r.RequestURI)
	reqLogger.Info("fetching release URL")

	vars := mux.Vars(r)
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	url, err := as.githubClient.FetchReleaseURL(ctx, vars["owner"], vars["repo"], vars["tag"], vars["assetName"])
	if ctx.Err() != nil {
		reqLogger.Error("error retrieving release URL", "err", err, "ctx error", ctx.Err())
		writeHTTPError(w, reqLogger, http.StatusBadGateway, "Bad Gateway")
		return
	}
	if err != nil {
		switch t := err.(type) {
		case GitHubError:
			if t.Type == TypeNotFound {
				reqLogger.Info("data not found", "err", t.WrappedError, "vars", vars)
				writeHTTPError(w, reqLogger, http.StatusNotFound, t.WrappedError.Error())
				return
			} else {
				reqLogger.Error("unhandled github error", "err", t.WrappedError, "vars", vars)
				writeHTTPError(w, reqLogger, http.StatusInternalServerError, "Internal Server Error")
				return
			}
		}
		reqLogger.Error("error retrieving release URL", "err", err, "vars", vars)
		writeHTTPError(w, reqLogger, http.StatusInternalServerError, err.Error())
		return
	}

	reqLogger.Info("found release URL", "url", url)

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusMovedPermanently)
}

func (as *apiServer) Status(w http.ResponseWriter, r *http.Request) {
	reqLogger := as.logger.New("method", r.Method, "url", r.RequestURI)

	out := struct {
		Version string `json:"version"`
	}{
		Version: version,
	}
	encoder := json.NewEncoder(w)

	w.WriteHeader(http.StatusOK)

	err := encoder.Encode(&out)
	if err != nil {
		reqLogger.Error("error encoding json", "err", err)
	}
}

func writeHTTPError(w http.ResponseWriter, logger log.Logger, statusCode int, message string) {
	w.WriteHeader(statusCode)
	if _, err := fmt.Fprintln(w, message); err != nil {
		logger.Crit("error writing response", "err", err)
	}
}

func basicAuth(username, password string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()

		if username != user || password != pass {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// NewAPIServer encapsulates the start of the gitreleases HTTP server.
func NewAPIServer(addr, metricsUsername, metricsPassword, version string, client *GithubClient, logger log.Logger) *apiServer {
	r := mux.NewRouter()

	as := apiServer{
		server: &http.Server{
			Addr:           addr,
			Handler:        r,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		githubClient: client,
		logger:       logger,
		version:      version,
	}

	r.Handle("/gh/{owner}/{repo}/{tag}/{assetName}", addRequestMetrics("DownloadRelease",
		http.HandlerFunc(as.DownloadRelease)))
	r.Handle("/metrics", basicAuth(metricsUsername, metricsPassword, promhttp.Handler()))
	r.HandleFunc("/status", as.Status)

	return &as
}
